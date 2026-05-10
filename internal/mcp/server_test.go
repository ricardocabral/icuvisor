package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

type registryFunc func(context.Context, tools.Registrar) error

func (f registryFunc) Register(ctx context.Context, registrar tools.Registrar) error {
	return f(ctx, registrar)
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
