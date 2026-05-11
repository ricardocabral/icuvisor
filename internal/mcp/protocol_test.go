package mcp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/tools"
)

func TestProtocolInitialize(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClient(t, testEchoRegistry{})
	defer cleanup()

	initResult := session.InitializeResult()
	if initResult == nil {
		t.Fatal("InitializeResult is nil")
	}
	if initResult.ServerInfo == nil || initResult.ServerInfo.Name != "icuvisor" {
		t.Fatalf("server info = %#v, want icuvisor", initResult.ServerInfo)
	}
	if initResult.ProtocolVersion == "" {
		t.Fatal("protocol version is empty")
	}
	if _, err := session.ListTools(ctx, nil); err != nil {
		t.Fatalf("ListTools() after initialize error = %v", err)
	}
}

func TestProtocolListTools(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClient(t, testEchoRegistry{})
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(result.Tools) != 1 {
		t.Fatalf("tool count = %d, want 1", len(result.Tools))
	}
	tool := result.Tools[0]
	if tool.Name != "test_echo" {
		t.Fatalf("tool name = %q, want test_echo", tool.Name)
	}
	if tool.Description == "" {
		t.Fatal("tool description is empty")
	}
}

func TestProtocolCallToolDispatch(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClient(t, testEchoRegistry{})
	defer cleanup()

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "test_echo",
		Arguments: map[string]any{"message": "hello"},
	})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool() IsError = true, content = %#v", result.Content)
	}
	if len(result.Content) != 1 {
		t.Fatalf("content count = %d, want 1", len(result.Content))
	}
	text, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want TextContent", result.Content[0])
	}
	if !strings.Contains(text.Text, "hello") {
		t.Fatalf("text content = %q, want echoed argument", text.Text)
	}
}

func TestProtocolMalformedRequestsAndHandlerErrors(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClient(t, registryFunc(func(_ context.Context, registrar tools.Registrar) error {
		return registrar.AddTool(tools.Tool{
			Name:        "test_failure",
			Description: "Returns a sanitized test failure.",
			InputSchema: map[string]any{"type": "object"},
			Handler: func(context.Context, tools.Request) (tools.Result, error) {
				return tools.Result{}, errors.New("secret upstream stack detail")
			},
		})
	}))
	defer cleanup()

	if _, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "missing_tool"}); err == nil {
		t.Fatal("CallTool(missing_tool) error = nil, want protocol error")
	} else if !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("CallTool(missing_tool) error = %q, want unknown tool", err.Error())
	}

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "test_failure", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool(test_failure) protocol error = %v", err)
	}
	if !result.IsError {
		t.Fatal("CallTool(test_failure) IsError = false, want true")
	}
	text, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want TextContent", result.Content[0])
	}
	if text.Text != genericToolErrorMessage {
		t.Fatalf("handler error text = %q, want generic message", text.Text)
	}
	if strings.Contains(text.Text, "secret") {
		t.Fatalf("handler error leaked internal detail: %q", text.Text)
	}
}

func connectTestClient(t *testing.T, registry tools.Registry) (context.Context, *sdkmcp.ClientSession, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server, err := NewServer(ctx, Options{Version: "test", Registry: registry})
	if err != nil {
		cancel()
		t.Fatalf("NewServer() error = %v", err)
	}

	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	serverSession, err := server.server.Connect(ctx, serverTransport, nil)
	if err != nil {
		cancel()
		t.Fatalf("server Connect() error = %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "icuvisor-test-client", Version: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		serverSession.Close()
		cancel()
		t.Fatalf("client Connect() error = %v", err)
	}

	cleanup := func() {
		clientSession.Close()
		serverSession.Close()
		cancel()
	}
	return ctx, clientSession, cleanup
}
