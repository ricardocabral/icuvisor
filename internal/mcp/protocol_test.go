package mcp

import (
	"bufio"
	"context"
	"errors"
	"net"
	"slices"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type testProfileClient struct {
	profile intervals.AthleteWithSportSettings
}

func (c testProfileClient) GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error) {
	return c.profile, nil
}

type capabilityRegistry struct{}

func (capabilityRegistry) Register(ctx context.Context, registrar tools.Registrar) error {
	for _, tool := range []tools.Tool{
		capabilityTestTool("test_read", tools.RequirementRead),
		capabilityTestTool("test_write", tools.RequirementWrite),
		capabilityTestTool("test_delete", tools.RequirementDelete),
	} {
		if err := registrar.AddTool(tool); err != nil {
			return err
		}
	}
	return nil
}

func capabilityTestTool(name string, requirement tools.Requirement) tools.Tool {
	return tools.Tool{
		Name:        name,
		Description: "Capability filtering test tool.",
		InputSchema: map[string]any{"type": "object"},
		Requirement: requirement,
		Toolset:     safety.ToolsetCore,
		Handler: func(context.Context, tools.Request) (tools.Result, error) {
			return tools.Result{}, nil
		},
	}
}

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

func TestProtocolFiltersToolsByCapability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode safety.Mode
		want []string
	}{
		{name: "safe", mode: safety.ModeSafe, want: []string{"test_read", "test_write"}},
		{name: "full", mode: safety.ModeFull, want: []string{"test_read", "test_write", "test_delete"}},
		{name: "none", mode: safety.ModeNone, want: []string{"test_read"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, session, cleanup := connectTestClientWithOptions(t, Options{Registry: capabilityRegistry{}, Capability: safety.NewCapability(tc.mode)})
			defer cleanup()

			result, err := session.ListTools(ctx, nil)
			if err != nil {
				t.Fatalf("ListTools() error = %v", err)
			}
			got := make([]string, 0, len(result.Tools))
			for _, tool := range result.Tools {
				got = append(got, tool.Name)
			}
			slices.Sort(got)
			slices.Sort(tc.want)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("tools/list = %v, want %v", got, tc.want)
			}
		})
	}
}

type deleteToolsRegistry struct{}

var deleteToolNames = []string{"delete_event", "delete_events_by_date_range", "delete_activity", "delete_custom_item", "delete_sport_settings", "delete_gear", "delete_workout"}

func (deleteToolsRegistry) Register(ctx context.Context, registrar tools.Registrar) error {
	for _, name := range deleteToolNames {
		if err := registrar.AddTool(capabilityTestTool(name, tools.RequirementDelete)); err != nil {
			return err
		}
	}
	return nil
}

type toolsetCapabilityRegistry struct{}

func (toolsetCapabilityRegistry) Register(ctx context.Context, registrar tools.Registrar) error {
	for _, tool := range []tools.Tool{
		toolsetCapabilityTestTool("core_read", safety.ToolsetCore, tools.RequirementRead),
		toolsetCapabilityTestTool("core_write", safety.ToolsetCore, tools.RequirementWrite),
		toolsetCapabilityTestTool("full_read", safety.ToolsetFull, tools.RequirementRead),
		toolsetCapabilityTestTool("full_write", safety.ToolsetFull, tools.RequirementWrite),
		toolsetCapabilityTestTool("full_delete", safety.ToolsetFull, tools.RequirementDelete),
	} {
		if err := registrar.AddTool(tool); err != nil {
			return err
		}
	}
	return nil
}

func toolsetCapabilityTestTool(name string, toolset safety.Toolset, requirement tools.Requirement) tools.Tool {
	tool := capabilityTestTool(name, requirement)
	tool.Toolset = toolset
	return tool
}

func TestProtocolFiltersToolsByToolsetAndCapability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		toolset safety.Toolset
		mode    safety.Mode
		want    []string
	}{
		{name: "core none", toolset: safety.ToolsetCore, mode: safety.ModeNone, want: []string{"core_read"}},
		{name: "core safe", toolset: safety.ToolsetCore, mode: safety.ModeSafe, want: []string{"core_read", "core_write"}},
		{name: "core full", toolset: safety.ToolsetCore, mode: safety.ModeFull, want: []string{"core_read", "core_write"}},
		{name: "zero toolset defaults core", toolset: "", mode: safety.ModeFull, want: []string{"core_read", "core_write"}},
		{name: "full none", toolset: safety.ToolsetFull, mode: safety.ModeNone, want: []string{"core_read", "full_read"}},
		{name: "full safe", toolset: safety.ToolsetFull, mode: safety.ModeSafe, want: []string{"core_read", "core_write", "full_read", "full_write"}},
		{name: "full full", toolset: safety.ToolsetFull, mode: safety.ModeFull, want: []string{"core_read", "core_write", "full_read", "full_write", "full_delete"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, session, cleanup := connectTestClientWithOptions(t, Options{Registry: toolsetCapabilityRegistry{}, Capability: safety.NewCapability(tc.mode), Toolset: tc.toolset})
			defer cleanup()

			result, err := session.ListTools(ctx, nil)
			if err != nil {
				t.Fatalf("ListTools() error = %v", err)
			}
			got := make([]string, 0, len(result.Tools))
			for _, tool := range result.Tools {
				got = append(got, tool.Name)
			}
			slices.Sort(got)
			slices.Sort(tc.want)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("tools/list = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestProtocolHiddenFullToolIsAbsentAndUnknown(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Registry: toolsetCapabilityRegistry{}, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetCore})
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	for _, tool := range result.Tools {
		if tool.Name == "full_read" {
			t.Fatalf("full-only tool appeared in core tools/list: %#v", result.Tools)
		}
	}
	if _, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "full_read", Arguments: map[string]any{}}); err == nil {
		t.Fatal("CallTool(full_read) error = nil, want unknown tool protocol error")
	} else if !strings.Contains(err.Error(), "unknown tool") {
		t.Fatalf("CallTool(full_read) error = %q, want unknown tool", err.Error())
	}
}

func TestProtocolFiltersDeleteToolsByCapability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      safety.Mode
		wantNames []string
	}{
		{name: "full", mode: safety.ModeFull, wantNames: deleteToolNames},
		{name: "safe", mode: safety.ModeSafe, wantNames: nil},
		{name: "none", mode: safety.ModeNone, wantNames: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, session, cleanup := connectTestClientWithOptions(t, Options{Registry: deleteToolsRegistry{}, Capability: safety.NewCapability(tc.mode)})
			defer cleanup()

			result, err := session.ListTools(ctx, nil)
			if err != nil {
				t.Fatalf("ListTools() error = %v", err)
			}
			got := make([]string, 0, len(result.Tools))
			for _, tool := range result.Tools {
				got = append(got, tool.Name)
			}
			want := append([]string(nil), tc.wantNames...)
			slices.Sort(got)
			slices.Sort(want)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Fatalf("tools/list = %v, want %v", got, want)
			}
		})
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

func TestProtocolGetAthleteProfileDispatch(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistry(testProfileClient{profile: intervals.AthleteWithSportSettings{
		ID:                    "12345",
		Name:                  "Example Athlete",
		MeasurementPreference: "METRIC",
		Timezone:              "America/Sao_Paulo",
		SportSettings: []intervals.SportSettings{{
			Types: []string{"Ride"},
			FTP:   250,
		}},
	}}, "v0.1-test", "UTC")
	ctx, session, cleanup := connectTestClient(t, registry)
	defer cleanup()

	toolsResult, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(toolsResult.Tools) != 1 || toolsResult.Tools[0].Name != "get_athlete_profile" {
		t.Fatalf("tools/list = %#v, want get_athlete_profile", toolsResult.Tools)
	}

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "get_athlete_profile", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool(get_athlete_profile) error = %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool(get_athlete_profile) IsError = true, content = %#v", result.Content)
	}
	text, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want TextContent", result.Content[0])
	}
	for _, want := range []string{"\"athlete_id\":\"i12345\"", "\"server_version\":\"v0.1-test\"", "\"ftp_watts\":250"} {
		if !strings.Contains(text.Text, want) {
			t.Fatalf("get_athlete_profile text = %s, missing %s", text.Text, want)
		}
	}
}

func TestProtocolMalformedRequestsAndHandlerErrors(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClient(t, registryFunc(func(_ context.Context, registrar tools.Registrar) error {
		return registrar.AddTool(tools.Tool{
			Name:        "test_failure",
			Description: "Returns a sanitized test failure.",
			InputSchema: map[string]any{"type": "object"},
			Toolset:     safety.ToolsetCore,
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

func TestProtocolMalformedRawRequest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	server, err := NewServer(ctx, Options{
		Version: "test",
		Transport: &sdkmcp.IOTransport{
			Reader: serverConn,
			Writer: serverConn,
		},
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- server.Run(ctx)
	}()

	if _, err := clientConn.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")); err != nil {
		t.Fatalf("write malformed request: %v", err)
	}
	if err := clientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	line, err := bufio.NewReader(clientConn).ReadString('\n')
	if err != nil {
		t.Fatalf("read malformed response: %v", err)
	}
	if !strings.Contains(line, "error") {
		t.Fatalf("malformed response = %q, want JSON-RPC error", line)
	}
	if strings.Contains(strings.ToLower(line), "panic") || strings.Contains(line, "secret") {
		t.Fatalf("malformed response leaked internal detail: %q", line)
	}

	cancel()
	clientConn.Close()
	waitForServerRun(t, runDone)
}

func connectTestClient(t *testing.T, registry tools.Registry) (context.Context, *sdkmcp.ClientSession, func()) {
	t.Helper()

	return connectTestClientWithOptions(t, Options{Registry: registry})
}

func connectTestClientWithOptions(t *testing.T, opts Options) (context.Context, *sdkmcp.ClientSession, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	serverTransport, clientTransport := sdkmcp.NewInMemoryTransports()
	opts.Version = "test"
	opts.Transport = serverTransport
	server, err := NewServer(ctx, opts)
	if err != nil {
		cancel()
		t.Fatalf("NewServer() error = %v", err)
	}

	runDone := make(chan error, 1)
	go func() {
		runDone <- server.Run(ctx)
	}()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "icuvisor-test-client", Version: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		cancel()
		waitForServerRun(t, runDone)
		t.Fatalf("client Connect() error = %v", err)
	}

	cleanup := func() {
		clientSession.Close()
		cancel()
		waitForServerRun(t, runDone)
	}
	return ctx, clientSession, cleanup
}

func waitForServerRun(t *testing.T, runDone <-chan error) {
	t.Helper()

	select {
	case err := <-runDone:
		if err != nil && !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "closed") {
			t.Fatalf("server Run() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server Run() did not stop")
	}
}
