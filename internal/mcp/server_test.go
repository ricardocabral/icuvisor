package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/resources"
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

type resourceRegistryFunc func(context.Context, resources.Registrar) error

func (f resourceRegistryFunc) Register(ctx context.Context, registrar resources.Registrar) error {
	return f(ctx, registrar)
}

type testEchoRegistry struct{}

type testResourceRegistry struct{}

func (testResourceRegistry) Register(ctx context.Context, registrar resources.Registrar) error {
	return registrar.AddResource(validTestResource("icuvisor://test-resource"))
}

func validTestResource(uri string) resources.Resource {
	return resources.Resource{
		URI:         uri,
		Name:        "test_resource",
		Title:       "Test Resource",
		Description: "Resource registration test fixture.",
		MIMEType:    "text/markdown",
		Handler:     testResourceHandler("# Test Resource\n"),
	}
}

func testResourceHandler(text string) resources.Handler {
	return func(ctx context.Context, req resources.Request) (resources.Result, error) {
		if err := ctx.Err(); err != nil {
			return resources.Result{}, err
		}
		return resources.Result{URI: req.URI, MIMEType: "text/markdown", Text: text}, nil
	}
}

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

func TestNewServerReturnsResourceRegistrationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resource resources.Resource
		want     string
	}{
		{name: "bad uri", resource: validTestResource("not-absolute"), want: "invalid resource URI"},
		{name: "missing name", resource: resources.Resource{URI: "icuvisor://missing-name", Title: "Missing Name", Description: "bad resource", MIMEType: "text/markdown", Handler: testResourceHandler("bad")}, want: "missing a name"},
		{name: "missing title", resource: resources.Resource{URI: "icuvisor://missing-title", Name: "missing_title", Description: "bad resource", MIMEType: "text/markdown", Handler: testResourceHandler("bad")}, want: "missing a title"},
		{name: "missing description", resource: resources.Resource{URI: "icuvisor://missing-description", Name: "missing_description", Title: "Missing Description", MIMEType: "text/markdown", Handler: testResourceHandler("bad")}, want: "missing a description"},
		{name: "missing mime", resource: resources.Resource{URI: "icuvisor://missing-mime", Name: "missing_mime", Title: "Missing MIME", Description: "bad resource", Handler: testResourceHandler("bad")}, want: "missing a MIME type"},
		{name: "missing handler", resource: resources.Resource{URI: "icuvisor://missing-handler", Name: "missing_handler", Title: "Missing Handler", Description: "bad resource", MIMEType: "text/markdown"}, want: "missing a handler"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewServer(context.Background(), Options{
				ResourceRegistry: resourceRegistryFunc(func(_ context.Context, registrar resources.Registrar) error {
					return registrar.AddResource(tc.resource)
				}),
			})
			if err == nil {
				t.Fatal("NewServer() error = nil, want resource registration error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("NewServer() error = %q, want %q", err.Error(), tc.want)
			}
		})
	}
}

func TestNewServerRejectsDuplicateResourceURIs(t *testing.T) {
	t.Parallel()

	resource := validTestResource("icuvisor://duplicate-resource")
	_, err := NewServer(context.Background(), Options{
		ResourceRegistry: resourceRegistryFunc(func(_ context.Context, registrar resources.Registrar) error {
			if err := registrar.AddResource(resource); err != nil {
				return err
			}
			return registrar.AddResource(resource)
		}),
	})
	if err == nil {
		t.Fatal("NewServer() error = nil, want duplicate resource error")
	}
	if !strings.Contains(err.Error(), "duplicate resource URI") {
		t.Fatalf("NewServer() error = %q, want duplicate resource URI", err.Error())
	}
}

func TestNewServerLogsResourceRegistrationCountsOnly(t *testing.T) {
	t.Parallel()

	var log bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&log, &slog.HandlerOptions{Level: slog.LevelInfo}))
	_, err := NewServer(context.Background(), Options{
		ResourceRegistry: testResourceRegistry{},
		Logger:           logger,
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	out := log.String()
	for _, want := range []string{"resource registration complete", "registered_count=1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("log %q missing %q", out, want)
		}
	}
	if strings.Contains(out, "icuvisor://test-resource") {
		t.Fatalf("startup registration log leaked resource URI in %q", out)
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

func TestServeStreamableHTTPInitializesClient(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server, err := NewServer(ctx, Options{Version: "v1.2.3", Registry: testEchoRegistry{}})
	if err != nil {
		cancel()
		listener.Close()
		t.Fatalf("NewServer() error = %v", err)
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- server.ServeStreamableHTTP(ctx, listener)
	}()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "icuvisor-http-test-client", Version: "test"}, nil)
	session, err := client.Connect(ctx, &sdkmcp.StreamableClientTransport{
		Endpoint:             "http://" + listener.Addr().String() + StreamableHTTPPath,
		HTTPClient:           &http.Client{Timeout: 2 * time.Second},
		MaxRetries:           -1,
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		cancel()
		waitForServerRun(t, runDone)
		t.Fatalf("client Connect() error = %v", err)
	}

	if initResult := session.InitializeResult(); initResult == nil || initResult.ServerInfo == nil || initResult.ServerInfo.Name != "icuvisor" {
		t.Fatalf("initialize result = %#v, want icuvisor server info", initResult)
	}
	toolsResult, err := session.ListTools(ctx, nil)
	if err != nil {
		session.Close()
		cancel()
		waitForServerRun(t, runDone)
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(toolsResult.Tools) != 1 || toolsResult.Tools[0].Name != "test_echo" {
		t.Fatalf("tools = %#v, want test_echo", toolsResult.Tools)
	}

	session.Close()
	cancel()
	waitForServerRun(t, runDone)
}

func TestServeStreamableHTTPLogsDoNotLeakConfigOrPayload(t *testing.T) {
	t.Parallel()

	logs := &safeLogBuffer{}
	logger := slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelInfo}))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server, err := NewServer(ctx, Options{
		Version: "v1.2.3",
		Logger:  logger,
		Config: config.Config{
			APIKey:    "sentinel-api-key",
			AthleteID: "i12345",
		},
	})
	if err != nil {
		cancel()
		listener.Close()
		t.Fatalf("NewServer() error = %v", err)
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- server.ServeStreamableHTTP(ctx, listener)
	}()
	waitForLog(t, logs, "server started listening")

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+listener.Addr().String()+StreamableHTTPPath, strings.NewReader("not json sentinel-api-key i12345"))
	if err != nil {
		cancel()
		waitForServerRun(t, runDone)
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	if err != nil {
		cancel()
		waitForServerRun(t, runDone)
		t.Fatalf("malformed HTTP request error = %v", err)
	}
	response.Body.Close()

	cancel()
	waitForServerRun(t, runDone)
	out := logs.String()
	for _, want := range []string{"server started listening", "server session ended"} {
		if !strings.Contains(out, want) {
			t.Fatalf("logs %q missing %q", out, want)
		}
	}
	for _, forbidden := range []string{"sentinel-api-key", "i12345"} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("logs leaked %q: %q", forbidden, out)
		}
	}
}

func TestServeStreamableHTTPCancelClosesListener(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	address := listener.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	server, err := NewServer(ctx, Options{})
	if err != nil {
		cancel()
		listener.Close()
		t.Fatalf("NewServer() error = %v", err)
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- server.ServeStreamableHTTP(ctx, listener)
	}()

	cancel()
	waitForServerRun(t, runDone)
	conn, err := net.DialTimeout("tcp", address, 100*time.Millisecond)
	if err == nil {
		conn.Close()
		t.Fatalf("listener %s still accepts connections after cancellation", address)
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
