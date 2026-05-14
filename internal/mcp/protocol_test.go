package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	promptscatalog "github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type testProfileClient struct {
	profile intervals.AthleteWithSportSettings
}

func (c testProfileClient) GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error) {
	return c.profile, nil
}

type advancedProtocolClient struct {
	testProfileClient
}

func (advancedProtocolClient) ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
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

type protocolTransportKind string

const (
	protocolTransportInMemory       protocolTransportKind = "in_memory"
	protocolTransportStreamableHTTP protocolTransportKind = "streamable_http"
)

var protocolTransportKinds = []protocolTransportKind{protocolTransportInMemory, protocolTransportStreamableHTTP}

func TestProtocolSharedTransportSuite(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name string
		opts Options
		run  func(*testing.T, context.Context, *sdkmcp.ClientSession)
	}{
		{
			name: "initialize",
			opts: Options{Registry: testEchoRegistry{}},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
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
			},
		},
		{
			name: "tools_list",
			opts: Options{Registry: testEchoRegistry{}},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				result, err := session.ListTools(ctx, nil)
				if err != nil {
					t.Fatalf("ListTools() error = %v", err)
				}
				if len(result.Tools) != 1 || result.Tools[0].Name != "test_echo" || result.Tools[0].Description == "" {
					t.Fatalf("tools/list = %#v, want populated test_echo", result.Tools)
				}
			},
		},
		{
			name: "tool_call_success",
			opts: Options{Registry: testEchoRegistry{}},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "test_echo", Arguments: map[string]any{"message": "hello"}})
				if err != nil {
					t.Fatalf("CallTool() error = %v", err)
				}
				assertEchoToolResult(t, result, "hello")
			},
		},
		{
			name: "missing_tool_and_sanitized_error",
			opts: Options{Registry: failingToolRegistry()},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				if _, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "missing_tool"}); err == nil {
					t.Fatal("CallTool(missing_tool) error = nil, want protocol error")
				} else if !strings.Contains(err.Error(), "unknown tool") {
					t.Fatalf("CallTool(missing_tool) error = %q, want unknown tool", err.Error())
				}
				result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "test_failure", Arguments: map[string]any{}})
				if err != nil {
					t.Fatalf("CallTool(test_failure) protocol error = %v", err)
				}
				assertSanitizedToolError(t, result)
			},
		},
		{
			name: "resources_list_and_read",
			opts: Options{ResourceRegistry: testResourceRegistry{}},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				list, err := session.ListResources(ctx, nil)
				if err != nil {
					t.Fatalf("ListResources() error = %v", err)
				}
				if len(list.Resources) != 1 || list.Resources[0].URI != "icuvisor://test-resource" {
					t.Fatalf("resources/list = %#v, want test resource", list.Resources)
				}
				read, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://test-resource"})
				if err != nil {
					t.Fatalf("ReadResource() error = %v", err)
				}
				assertTestResourceRead(t, read)
			},
		},
		{
			name: "missing_resource",
			opts: Options{ResourceRegistry: testResourceRegistry{}},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				_, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://missing-resource"})
				if err == nil {
					t.Fatal("ReadResource(missing) error = nil, want not-found protocol error")
				}
				if !strings.Contains(err.Error(), "Resource not found") {
					t.Fatalf("ReadResource(missing) error = %q, want Resource not found", err.Error())
				}
			},
		},
		{
			name: "sanitized_resource_error",
			opts: Options{ResourceRegistry: failingResourceRegistry()},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				_, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://failing-resource"})
				if err == nil {
					t.Fatal("ReadResource(failing) error = nil, want sanitized protocol error")
				}
				if !strings.Contains(err.Error(), genericResourceErrorMessage) {
					t.Fatalf("ReadResource(failing) error = %q, want generic resource message", err.Error())
				}
				if strings.Contains(err.Error(), "secret") {
					t.Fatalf("ReadResource(failing) error leaked internal detail: %q", err.Error())
				}
			},
		},
		{
			name: "prompts_list_empty",
			opts: Options{},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				result, err := session.ListPrompts(ctx, nil)
				if err != nil {
					t.Fatalf("ListPrompts() error = %v", err)
				}
				if len(result.Prompts) != 0 {
					t.Fatalf("prompts/list = %#v, want empty current prompt catalog", result.Prompts)
				}
			},
		},
		{
			name: "prompts_list_and_get",
			opts: Options{PromptRegistry: promptscatalog.NewRegistry()},
			run: func(t *testing.T, ctx context.Context, session *sdkmcp.ClientSession) {
				t.Helper()
				result, err := session.ListPrompts(ctx, nil)
				if err != nil {
					t.Fatalf("ListPrompts() error = %v", err)
				}
				if len(result.Prompts) != 5 {
					t.Fatalf("prompts/list length = %d, want 5: %#v", len(result.Prompts), result.Prompts)
				}
				wantNames := []string{"coach_roster_triage", "race_week_taper", "recovery_check", "training_analysis", "weekly_planning"}
				for i, want := range wantNames {
					if result.Prompts[i].Name != want || result.Prompts[i].Description == "" {
						t.Fatalf("prompts[%d] = %#v, want name %q with description", i, result.Prompts[i], want)
					}
				}
				got, err := session.GetPrompt(ctx, &sdkmcp.GetPromptParams{Name: promptscatalog.CoachRosterTriageName, Arguments: map[string]string{"athlete_id": "12345"}})
				if err != nil {
					t.Fatalf("GetPrompt() error = %v", err)
				}
				if len(got.Messages) != 1 {
					t.Fatalf("GetPrompt() messages = %#v, want one message", got.Messages)
				}
				text, ok := got.Messages[0].Content.(*sdkmcp.TextContent)
				if !ok {
					t.Fatalf("GetPrompt() content = %T, want TextContent", got.Messages[0].Content)
				}
				for _, want := range []string{"Scope: athlete_id=i12345", "athlete_id as a coach-mode selector", "get_wellness_data"} {
					if !strings.Contains(text.Text, want) {
						t.Fatalf("GetPrompt() text missing %q:\n%s", want, text.Text)
					}
				}
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			for _, kind := range protocolTransportKinds {
				t.Run(string(kind), func(t *testing.T) {
					ctx, session, cleanup := connectProtocolClient(t, kind, scenario.opts)
					defer cleanup()
					scenario.run(t, ctx, session)
				})
			}
		})
	}
}

func TestProtocolTransportParity(t *testing.T) {
	t.Parallel()

	snapshots := make(map[protocolTransportKind][]byte, len(protocolTransportKinds))
	for _, kind := range protocolTransportKinds {
		ctx, session, cleanup := connectProtocolClient(t, kind, Options{Registry: testEchoRegistry{}, ResourceRegistry: testResourceRegistry{}})
		snapshot, err := protocolParitySnapshot(ctx, session)
		cleanup()
		if err != nil {
			t.Fatalf("%s parity snapshot error = %v", kind, err)
		}
		snapshots[kind] = snapshot
	}

	if string(snapshots[protocolTransportInMemory]) != string(snapshots[protocolTransportStreamableHTTP]) {
		t.Fatalf("protocol responses differ across transports\nin_memory: %s\nstreamable_http: %s", snapshots[protocolTransportInMemory], snapshots[protocolTransportStreamableHTTP])
	}
}

func TestProtocolMalformedHTTPPost(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server, err := NewServer(ctx, Options{Version: "test"})
	if err != nil {
		cancel()
		listener.Close()
		t.Fatalf("NewServer() error = %v", err)
	}
	runDone := make(chan error, 1)
	go func() {
		runDone <- server.ServeStreamableHTTP(ctx, listener)
	}()
	defer func() {
		cancel()
		waitForServerRun(t, runDone)
	}()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+listener.Addr().String()+StreamableHTTPPath, strings.NewReader("not json sentinel-api-key i12345"))
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	if err != nil {
		t.Fatalf("malformed HTTP request error = %v", err)
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read malformed HTTP response: %v", err)
	}
	body := string(bodyBytes)
	if response.StatusCode < http.StatusBadRequest {
		t.Fatalf("malformed HTTP status = %d, body = %q; want client/server error", response.StatusCode, body)
	}
	lowerBody := strings.ToLower(body)
	if strings.Contains(lowerBody, "panic") || strings.Contains(body, "sentinel-api-key") || strings.Contains(body, "i12345") {
		t.Fatalf("malformed HTTP response leaked internal detail: %q", body)
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

func TestProtocolResourceInitializeListAndRead(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClientWithOptions(t, Options{ResourceRegistry: testResourceRegistry{}})
	defer cleanup()

	initResult := session.InitializeResult()
	if initResult == nil || initResult.Capabilities == nil || initResult.Capabilities.Resources == nil {
		t.Fatalf("initialize capabilities = %#v, want resources capability", initResult)
	}

	list, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	if len(list.Resources) != 1 {
		t.Fatalf("resource count = %d, want 1", len(list.Resources))
	}
	resource := list.Resources[0]
	if resource.URI != "icuvisor://test-resource" {
		t.Fatalf("resource URI = %q, want icuvisor://test-resource", resource.URI)
	}
	if resource.Name != "test_resource" || resource.Title != "Test Resource" || resource.Description == "" || resource.MIMEType != "text/markdown" {
		t.Fatalf("resource metadata = %#v, want populated metadata", resource)
	}

	read, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://test-resource"})
	if err != nil {
		t.Fatalf("ReadResource() error = %v", err)
	}
	if len(read.Contents) != 1 {
		t.Fatalf("content count = %d, want 1", len(read.Contents))
	}
	content := read.Contents[0]
	if content.URI != "icuvisor://test-resource" || content.MIMEType != "text/markdown" || !strings.Contains(content.Text, "Test Resource") {
		t.Fatalf("resource content = %#v, want URI/MIME/text", content)
	}
}

func TestProtocolUnknownResourceReturnsNotFound(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClientWithOptions(t, Options{ResourceRegistry: testResourceRegistry{}})
	defer cleanup()

	_, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://missing-resource"})
	if err == nil {
		t.Fatal("ReadResource(missing) error = nil, want not-found protocol error")
	}
	if !strings.Contains(err.Error(), "Resource not found") {
		t.Fatalf("ReadResource(missing) error = %q, want Resource not found", err.Error())
	}
}

func TestProtocolDefaultResourceRegistryIncludesAllResources(t *testing.T) {
	t.Parallel()

	profileClient := testProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Name: "Example Athlete", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{Types: []string{"Ride"}, FTP: 250}}}}
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{ResourceRegistry: resources.NewRegistryWithOptions(profileClient, resources.ResourceOptions{Version: "v0.1-test", TimezoneFallback: "UTC"})})
	defer cleanup()

	list, err := session.ListResources(ctx, nil)
	if err != nil {
		t.Fatalf("ListResources() error = %v", err)
	}
	want := map[string]struct {
		name     string
		mimeType string
	}{
		resources.WorkoutSyntaxURI:     {name: "workout_syntax", mimeType: resources.WorkoutSyntaxMIMEType},
		resources.EventCategoriesURI:   {name: "event_categories", mimeType: resources.EventCategoriesMIMEType},
		resources.CustomItemSchemasURI: {name: "custom_item_schemas", mimeType: resources.CustomItemSchemasMIMEType},
		resources.AthleteProfileURI:    {name: "athlete_profile", mimeType: resources.AthleteProfileMIMEType},
	}
	for _, resource := range list.Resources {
		if expected, ok := want[resource.URI]; ok {
			if resource.MIMEType != expected.mimeType || resource.Name != expected.name {
				t.Fatalf("resource metadata = %#v, want %#v", resource, expected)
			}
			delete(want, resource.URI)
		}
	}
	if len(want) > 0 {
		t.Fatalf("resources/list = %#v, missing %v", list.Resources, want)
	}

	for uri, expected := range map[string]struct {
		mimeType string
		contains string
	}{
		resources.WorkoutSyntaxURI:     {mimeType: resources.WorkoutSyntaxMIMEType, contains: "# Workout syntax"},
		resources.EventCategoriesURI:   {mimeType: resources.EventCategoriesMIMEType, contains: "# Event categories"},
		resources.CustomItemSchemasURI: {mimeType: resources.CustomItemSchemasMIMEType, contains: "# Custom item content schemas"},
		resources.AthleteProfileURI:    {mimeType: resources.AthleteProfileMIMEType, contains: "\"athlete_id\":\"i12345\""},
	} {
		read, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: uri})
		if err != nil {
			t.Fatalf("ReadResource(%s) error = %v", uri, err)
		}
		if len(read.Contents) != 1 || read.Contents[0].URI != uri || read.Contents[0].MIMEType != expected.mimeType || !strings.Contains(read.Contents[0].Text, expected.contains) {
			t.Fatalf("resource %s read = %#v", uri, read.Contents)
		}
	}
}

func TestProtocolResourceHandlerErrorsAreSanitized(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClientWithOptions(t, Options{
		ResourceRegistry: resourceRegistryFunc(func(_ context.Context, registrar resources.Registrar) error {
			return registrar.AddResource(resources.Resource{
				URI:         "icuvisor://failing-resource",
				Name:        "failing_resource",
				Title:       "Failing Resource",
				Description: "Fails for protocol error sanitization tests.",
				MIMEType:    "text/markdown",
				Handler: func(context.Context, resources.Request) (resources.Result, error) {
					return resources.Result{}, errors.New("secret upstream stack detail")
				},
			})
		}),
	})
	defer cleanup()

	_, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://failing-resource"})
	if err == nil {
		t.Fatal("ReadResource(failing) error = nil, want sanitized protocol error")
	}
	if !strings.Contains(err.Error(), genericResourceErrorMessage) {
		t.Fatalf("ReadResource(failing) error = %q, want generic resource message", err.Error())
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("ReadResource(failing) error leaked internal detail: %q", err.Error())
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

func TestProtocolListAdvancedCapabilitiesVisibilityWithRealRegistry(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistryWithOptions(advancedProtocolClient{testProfileClient: testProfileClient{}}, tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC"})
	ctx, session, cleanup := connectTestClient(t, registry)
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
	want := []string{"get_athlete_profile", "icuvisor_list_advanced_capabilities"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("core tools/list = %v, want %v", got, want)
	}

	fullRegistry := tools.NewRegistryWithOptions(advancedProtocolClient{testProfileClient: testProfileClient{}}, tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Toolset: safety.ToolsetFull})
	fullCtx, fullSession, fullCleanup := connectTestClientWithOptions(t, Options{Registry: fullRegistry, Toolset: safety.ToolsetFull})
	defer fullCleanup()
	fullResult, err := fullSession.ListTools(fullCtx, nil)
	if err != nil {
		t.Fatalf("full ListTools() error = %v", err)
	}
	fullNames := make([]string, 0, len(fullResult.Tools))
	for _, tool := range fullResult.Tools {
		fullNames = append(fullNames, tool.Name)
	}
	for _, wantName := range []string{"get_power_curves", "icuvisor_list_advanced_capabilities"} {
		if !slices.Contains(fullNames, wantName) {
			t.Fatalf("full tools/list = %v, missing %s", fullNames, wantName)
		}
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
	gotNames := make([]string, 0, len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		gotNames = append(gotNames, tool.Name)
	}
	slices.Sort(gotNames)
	wantNames := []string{"get_athlete_profile", "icuvisor_list_advanced_capabilities"}
	if strings.Join(gotNames, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("tools/list = %v, want %v", gotNames, wantNames)
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

	return connectProtocolClient(t, protocolTransportInMemory, opts)
}

func connectProtocolClient(t *testing.T, kind protocolTransportKind, opts Options) (context.Context, *sdkmcp.ClientSession, func()) {
	t.Helper()

	switch kind {
	case protocolTransportInMemory:
		return connectInMemoryProtocolClient(t, opts)
	case protocolTransportStreamableHTTP:
		return connectStreamableHTTPProtocolClient(t, opts)
	default:
		t.Fatalf("unknown protocol transport kind %q", kind)
		return nil, nil, nil
	}
}

func connectInMemoryProtocolClient(t *testing.T, opts Options) (context.Context, *sdkmcp.ClientSession, func()) {
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

func connectStreamableHTTPProtocolClient(t *testing.T, opts Options) (context.Context, *sdkmcp.ClientSession, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	opts.Version = "test"
	opts.Transport = nil
	server, err := NewServer(ctx, opts)
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
	clientSession, err := client.Connect(ctx, &sdkmcp.StreamableClientTransport{
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

	cleanup := func() {
		clientSession.Close()
		cancel()
		waitForServerRun(t, runDone)
	}
	return ctx, clientSession, cleanup
}

func protocolParitySnapshot(ctx context.Context, session *sdkmcp.ClientSession) ([]byte, error) {
	toolsResult, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	callResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: "test_echo", Arguments: map[string]any{"message": "parity"}})
	if err != nil {
		return nil, err
	}
	resourcesResult, err := session.ListResources(ctx, nil)
	if err != nil {
		return nil, err
	}
	readResult, err := session.ReadResource(ctx, &sdkmcp.ReadResourceParams{URI: "icuvisor://test-resource"})
	if err != nil {
		return nil, err
	}
	promptsResult, err := session.ListPrompts(ctx, nil)
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Initialize *sdkmcp.InitializeResult    `json:"initialize"`
		Tools      *sdkmcp.ListToolsResult     `json:"tools"`
		Call       *sdkmcp.CallToolResult      `json:"call"`
		Resources  *sdkmcp.ListResourcesResult `json:"resources"`
		Read       *sdkmcp.ReadResourceResult  `json:"read"`
		Prompts    *sdkmcp.ListPromptsResult   `json:"prompts"`
	}{
		Initialize: session.InitializeResult(),
		Tools:      toolsResult,
		Call:       callResult,
		Resources:  resourcesResult,
		Read:       readResult,
		Prompts:    promptsResult,
	})
}

func failingToolRegistry() tools.Registry {
	return registryFunc(func(_ context.Context, registrar tools.Registrar) error {
		return registrar.AddTool(tools.Tool{
			Name:        "test_failure",
			Description: "Returns a sanitized test failure.",
			InputSchema: map[string]any{"type": "object"},
			Toolset:     safety.ToolsetCore,
			Handler: func(context.Context, tools.Request) (tools.Result, error) {
				return tools.Result{}, errors.New("secret upstream stack detail")
			},
		})
	})
}

func failingResourceRegistry() resources.Registry {
	return resourceRegistryFunc(func(_ context.Context, registrar resources.Registrar) error {
		return registrar.AddResource(resources.Resource{
			URI:         "icuvisor://failing-resource",
			Name:        "failing_resource",
			Title:       "Failing Resource",
			Description: "Fails for protocol error sanitization tests.",
			MIMEType:    "text/markdown",
			Handler: func(context.Context, resources.Request) (resources.Result, error) {
				return resources.Result{}, errors.New("secret upstream stack detail")
			},
		})
	})
}

func assertEchoToolResult(t *testing.T, result *sdkmcp.CallToolResult, wantText string) {
	t.Helper()

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
	if !strings.Contains(text.Text, wantText) {
		t.Fatalf("text content = %q, want %q", text.Text, wantText)
	}
}

func assertSanitizedToolError(t *testing.T, result *sdkmcp.CallToolResult) {
	t.Helper()

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

func assertTestResourceRead(t *testing.T, read *sdkmcp.ReadResourceResult) {
	t.Helper()

	if len(read.Contents) != 1 {
		t.Fatalf("content count = %d, want 1", len(read.Contents))
	}
	content := read.Contents[0]
	if content.URI != "icuvisor://test-resource" || content.MIMEType != "text/markdown" || !strings.Contains(content.Text, "Test Resource") {
		t.Fatalf("resource content = %#v, want URI/MIME/text", content)
	}
}

func waitForServerRun(t *testing.T, runDone <-chan error) {
	t.Helper()

	select {
	case err := <-runDone:
		if err != nil && !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "closed") {
			t.Fatalf("server Run() error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server Run() did not stop")
	}
}
