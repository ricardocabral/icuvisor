package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

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
