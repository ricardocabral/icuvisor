package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type safeLogBuffer struct {
	mu sync.Mutex
	bytes.Buffer
}

func (b *safeLogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(p)
}

func (b *safeLogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.String()
}

type blockingTransport struct {
	connected chan struct{}
}

func (t *blockingTransport) Connect(context.Context) (sdkmcp.Connection, error) {
	conn := &blockingConn{done: make(chan struct{})}
	close(t.connected)
	return conn, nil
}

type blockingConn struct {
	done chan struct{}
	once sync.Once
}

func (c *blockingConn) Read(ctx context.Context) (jsonrpc.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, io.EOF
	}
}

func (c *blockingConn) Write(context.Context, jsonrpc.Message) error {
	select {
	case <-c.done:
		return io.ErrClosedPipe
	default:
		return nil
	}
}

func (c *blockingConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return nil
}

func (c *blockingConn) SessionID() string { return "" }

type registryFunc func(context.Context, tools.Registrar) error

func (f registryFunc) Register(ctx context.Context, registrar tools.Registrar) error {
	return f(ctx, registrar)
}

type testEchoRegistry struct{}

func (testEchoRegistry) Register(ctx context.Context, registrar tools.Registrar) error {
	return registrar.AddTool(tools.Tool{
		Name:        "test_echo",
		Description: "Echoes raw test input for MCP protocol tests.",
		InputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": true,
		},
		OutputSchema: map[string]any{"type": "object"},
		Toolset:      safety.ToolsetCore,
		Handler: func(ctx context.Context, req tools.Request) (tools.Result, error) {
			if err := ctx.Err(); err != nil {
				return tools.Result{}, err
			}
			return tools.Result{
				Content: []tools.Content{{Type: tools.ContentTypeText, Text: string(req.Arguments)}},
				StructuredContent: map[string]any{
					"tool": req.Name,
				},
			}, nil
		},
	})
}

func TestNewServerHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewServer(ctx, Options{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("NewServer() error = %v, want context.Canceled", err)
	}
}

func TestNewServerAcceptsTestEchoRegistry(t *testing.T) {
	t.Parallel()

	if _, err := NewServer(context.Background(), Options{Registry: testEchoRegistry{}}); err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
}

func TestNewServerReturnsToolRegistrationErrors(t *testing.T) {
	t.Parallel()

	_, err := NewServer(context.Background(), Options{
		Registry: registryFunc(func(_ context.Context, registrar tools.Registrar) error {
			return registrar.AddTool(tools.Tool{
				Name:        "BadName",
				Description: "bad name",
				InputSchema: map[string]any{"type": "object"},
				Handler: func(context.Context, tools.Request) (tools.Result, error) {
					return tools.Result{}, nil
				},
			})
		}),
	})
	if err == nil {
		t.Fatal("NewServer() error = nil, want invalid tool error")
	}
	if !strings.Contains(err.Error(), "invalid tool name") {
		t.Fatalf("NewServer() error = %q, want invalid tool name", err.Error())
	}
}

func TestNewServerLogsRegistrationCountsOnly(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&log, &slog.HandlerOptions{Level: slog.LevelInfo}))
	_, err := NewServer(context.Background(), Options{
		Registry:   capabilityRegistry{},
		Capability: safety.NewCapability(safety.ModeSafe),
		Logger:     logger,
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	out := log.String()
	for _, want := range []string{"tool registration complete", "registered_count=2", "skipped_toolset_count=0", "skipped_capability_count=1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("log %q missing %q", out, want)
		}
	}
	for _, unwanted := range []string{"test_read", "test_write", "test_delete"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("startup registration log leaked tool name %q in %q", unwanted, out)
		}
	}
}

func TestNewServerRejectsInvalidToolset(t *testing.T) {
	t.Parallel()

	_, err := NewServer(context.Background(), Options{
		Registry: registryFunc(func(_ context.Context, registrar tools.Registrar) error {
			return registrar.AddTool(tools.Tool{
				Name:        "test_invalid_toolset",
				Description: "bad tier",
				Toolset:     safety.Toolset("advanced"),
				InputSchema: map[string]any{"type": "object"},
				Handler: func(context.Context, tools.Request) (tools.Result, error) {
					return tools.Result{}, nil
				},
			})
		}),
	})
	if err == nil {
		t.Fatal("NewServer() error = nil, want invalid toolset error")
	}
	if !strings.Contains(err.Error(), "invalid toolset") {
		t.Fatalf("NewServer() error = %q, want invalid toolset", err.Error())
	}
}

func TestNewServerRejectsDuplicateToolNames(t *testing.T) {
	t.Parallel()

	duplicate := tools.Tool{
		Name:        "test_duplicate",
		Description: "duplicate test tool",
		InputSchema: map[string]any{"type": "object"},
		Handler: func(context.Context, tools.Request) (tools.Result, error) {
			return tools.Result{}, nil
		},
	}
	_, err := NewServer(context.Background(), Options{
		Registry: registryFunc(func(_ context.Context, registrar tools.Registrar) error {
			if err := registrar.AddTool(duplicate); err != nil {
				return err
			}
			return registrar.AddTool(duplicate)
		}),
	})
	if err == nil {
		t.Fatal("NewServer() error = nil, want duplicate tool error")
	}
	if !strings.Contains(err.Error(), "duplicate tool name") {
		t.Fatalf("NewServer() error = %q, want duplicate tool name", err.Error())
	}
}

func TestPublicToolErrorMessageSanitizesUnknownErrors(t *testing.T) {
	t.Parallel()

	if got := publicToolErrorMessage(fmt.Errorf("upstream secret detail")); got != genericToolErrorMessage {
		t.Fatalf("publicToolErrorMessage() = %q, want %q", got, genericToolErrorMessage)
	}
	if got := publicToolErrorMessage(tools.NewUserError("try a valid test input", fmt.Errorf("secret cause"))); got != "try a valid test input" {
		t.Fatalf("publicToolErrorMessage() = %q, want public message", got)
	}
}

func TestRunHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	server, err := NewServer(context.Background(), Options{})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := server.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
}

func TestRunLogsStartedListening(t *testing.T) {
	t.Parallel()

	logs := &safeLogBuffer{}
	logger := slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	transport := &blockingTransport{connected: make(chan struct{})}
	server, err := NewServer(context.Background(), Options{
		Version:   "v1.2.3",
		Logger:    logger,
		Transport: transport,
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- server.Run(ctx)
	}()

	select {
	case <-transport.connected:
	case <-time.After(time.Second):
		cancel()
		t.Fatal("server transport did not connect")
	}
	waitForLog(t, logs, "server started listening")

	cancel()
	if err := <-runDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}

	out := logs.String()
	for _, want := range []string{"server started listening", "version=v1.2.3"} {
		if !strings.Contains(out, want) {
			t.Fatalf("listen log %q missing %q", out, want)
		}
	}
}

func waitForLog(t *testing.T, logs *safeLogBuffer, want string) {
	t.Helper()
	deadline := time.After(time.Second)
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		if strings.Contains(logs.String(), want) {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("log %q missing %q", logs.String(), want)
		case <-tick.C:
		}
	}
}
