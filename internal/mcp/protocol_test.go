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
	"sync/atomic"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/coach"
	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	promptscatalog "github.com/ricardocabral/icuvisor/internal/prompts"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
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

type protocolTransportKind string

const (
	protocolTransportInMemory       protocolTransportKind = "in_memory"
	protocolTransportStreamableHTTP protocolTransportKind = "streamable_http"
)

var protocolTransportKinds = []protocolTransportKind{protocolTransportInMemory, protocolTransportStreamableHTTP}

func TestStreamableHTTPJSONRPCInitializeAndPingWireEnvelopes(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server, err := NewServer(ctx, Options{Version: "test", Registry: testEchoRegistry{}})
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

	client := &http.Client{Timeout: 2 * time.Second}
	endpoint := "http://" + listener.Addr().String() + StreamableHTTPPath
	protocolVersion := "2025-06-18"

	initialize := codexStreamableHTTPPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"codex-jsonrpc-smoke","version":"test"}}}`, "", "")
	if initialize.status != http.StatusOK {
		t.Fatalf("initialize status = %d body = %q, want 200", initialize.status, initialize.body)
	}
	sessionID := initialize.headers.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatalf("initialize response missing Mcp-Session-Id header; headers = %#v", initialize.headers)
	}
	initializeResult := assertStreamableHTTPJSONRPCEnvelope(t, "initialize", initialize, 1)
	if _, ok := initializeResult["serverInfo"]; !ok {
		t.Fatalf("initialize result = %#v, want serverInfo", initializeResult)
	}

	initialized := codexStreamableHTTPPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`, sessionID, protocolVersion)
	if initialized.status != http.StatusAccepted {
		t.Fatalf("notifications/initialized status = %d body = %q, want 202", initialized.status, initialized.body)
	}

	ping := codexStreamableHTTPPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":2,"method":"ping","params":{}}`, sessionID, protocolVersion)
	if ping.status != http.StatusOK {
		t.Fatalf("ping status = %d body = %q, want 200", ping.status, ping.body)
	}
	assertStreamableHTTPJSONRPCEnvelope(t, "ping", ping, 2)
}

type streamableHTTPWireResponse struct {
	status  int
	headers http.Header
	body    []byte
}

func codexStreamableHTTPPost(t *testing.T, ctx context.Context, client *http.Client, endpoint, payload, sessionID, protocolVersion string) streamableHTTPWireResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		t.Fatalf("NewRequestWithContext() error = %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		request.Header.Set("Mcp-Session-Id", sessionID)
	}
	if protocolVersion != "" {
		request.Header.Set("Mcp-Protocol-Version", protocolVersion)
	}

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("POST %s error = %v", endpoint, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("reading streamable HTTP response body: %v", err)
	}
	return streamableHTTPWireResponse{status: response.StatusCode, headers: response.Header.Clone(), body: body}
}

func assertStreamableHTTPJSONRPCEnvelope(t *testing.T, label string, response streamableHTTPWireResponse, wantID int) map[string]json.RawMessage {
	t.Helper()

	envelope := decodeStreamableHTTPJSONRPCEnvelope(t, label, response)
	var version string
	if raw, ok := envelope["jsonrpc"]; !ok {
		t.Fatalf("%s response %q missing top-level jsonrpc field", label, response.body)
	} else if err := json.Unmarshal(raw, &version); err != nil || version != "2.0" {
		t.Fatalf("%s response jsonrpc = %s, err = %v, want \"2.0\"", label, raw, err)
	}
	var gotID int
	if raw, ok := envelope["id"]; !ok {
		t.Fatalf("%s response %q missing top-level id field", label, response.body)
	} else if err := json.Unmarshal(raw, &gotID); err != nil || gotID != wantID {
		t.Fatalf("%s response id = %s, err = %v, want %d", label, raw, err, wantID)
	}
	if raw, ok := envelope["error"]; ok {
		t.Fatalf("%s response has top-level error member %s", label, raw)
	}
	rawResult, ok := envelope["result"]
	if !ok {
		t.Fatalf("%s response %q missing top-level result field", label, response.body)
	}
	if trimmed := strings.TrimSpace(string(rawResult)); !strings.HasPrefix(trimmed, "{") {
		t.Fatalf("%s response result = %s, want JSON object", label, rawResult)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(rawResult, &result); err != nil {
		t.Fatalf("%s response result decode error = %v; result = %s", label, err, rawResult)
	}
	return result
}

func decodeStreamableHTTPJSONRPCEnvelope(t *testing.T, label string, response streamableHTTPWireResponse) map[string]json.RawMessage {
	t.Helper()

	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(response.headers.Get("Content-Type"), ";")[0]))
	payload := response.body
	if mediaType == "text/event-stream" {
		payload = firstJSONSSEDataPayload(t, label, response.body)
	} else if mediaType != "application/json" {
		t.Fatalf("%s response content type = %q body = %q, want application/json or text/event-stream", label, response.headers.Get("Content-Type"), response.body)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatalf("%s response is not a JSON-RPC object: %v; payload = %q; raw body = %q", label, err, payload, response.body)
	}
	if len(envelope) == 0 {
		t.Fatalf("%s response decoded as empty object; raw body = %q", label, response.body)
	}
	return envelope
}

func firstJSONSSEDataPayload(t *testing.T, label string, body []byte) []byte {
	t.Helper()

	var candidates []string
	var current strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			if current.Len() > 0 {
				candidates = append(candidates, current.String())
				current.Reset()
			}
			continue
		}
		if data, ok := strings.CutPrefix(line, "data:"); ok {
			if current.Len() > 0 {
				current.WriteByte('\n')
			}
			current.WriteString(strings.TrimSpace(data))
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("%s response SSE scan error = %v; raw body = %q", label, err, body)
	}
	if current.Len() > 0 {
		candidates = append(candidates, current.String())
	}
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if strings.HasPrefix(trimmed, "{") && json.Valid([]byte(trimmed)) {
			return []byte(trimmed)
		}
	}
	t.Fatalf("%s response SSE body has no JSON object data event: %q", label, body)
	return nil
}

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
				if len(result.Prompts) != 13 {
					t.Fatalf("prompts/list length = %d, want 13: %#v", len(result.Prompts), result.Prompts)
				}
				wantNames := []string{"coach_athlete_onboarding", "coach_roster_triage", "coaching_handoff", "fueling_review", "masters_plan_review", "plan_health_review", "race_week_taper", "recovery_check", "ride_analysis", "shareable_training_report", "training_analysis", "weekly_planning", "weekly_review"}
				for i, want := range wantNames {
					if result.Prompts[i].Name != want || result.Prompts[i].Description == "" {
						t.Fatalf("prompts[%d] = %#v, want name %q with description", i, result.Prompts[i], want)
					}
				}
				got, err := session.GetPrompt(ctx, &sdkmcp.GetPromptParams{Name: promptscatalog.FuelingReviewName, Arguments: map[string]string{"start_date": "2026-05-01", "end_date": "2026-05-14", "race_date": "2026-06-07", "race_name": "A Race"}})
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
				for _, want := range []string{"Scope: start_date=2026-05-01, end_date=2026-05-14, race_date=2026-06-07, race_name=A Race", "include_unnamed:true", "get_wellness_data", "limit:100"} {
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
	if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
		t.Fatalf("default-read tool annotations = %#v, want readOnlyHint true", tool.Annotations)
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
		resources.AnalysisFormulasURI:  {name: "analysis_formulas", mimeType: resources.AnalysisFormulasMIMEType},
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
		resources.AnalysisFormulasURI:  {mimeType: resources.AnalysisFormulasMIMEType, contains: "# Analysis formulas"},
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

type coachACLTestRegistry struct{}

func (coachACLTestRegistry) Register(ctx context.Context, registrar tools.Registrar) error {
	registered := []tools.Tool{
		coachACLTestTool(toolcatalog.GetAthleteProfile, safety.ToolsetCore, tools.RequirementRead),
		coachACLTestTool(toolcatalog.AddOrUpdateEvent, safety.ToolsetCore, tools.RequirementWrite),
		coachACLTestTool(toolcatalog.DeleteEvent, safety.ToolsetCore, tools.RequirementDelete),
		coachACLTestTool(toolcatalog.GetPowerCurves, safety.ToolsetFull, tools.RequirementRead),
	}
	for _, tool := range registered {
		if err := registrar.AddTool(tool); err != nil {
			return err
		}
	}
	return nil
}

func coachACLTestTool(name string, toolset safety.Toolset, requirement tools.Requirement) tools.Tool {
	return tools.Tool{
		Name:        name,
		Description: "Coach ACL test tool.",
		InputSchema: map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{}},
		Toolset:     toolset,
		Requirement: requirement,
		Handler: func(ctx context.Context, req tools.Request) (tools.Result, error) {
			var args map[string]any
			if err := json.Unmarshal(req.Arguments, &args); err != nil {
				return tools.Result{}, err
			}
			if _, ok := args["athlete_id"]; ok {
				return tools.Result{}, errors.New("athlete_id reached strict handler")
			}
			target, _ := intervals.TargetAthleteIDFromContext(ctx)
			return tools.TextResult(map[string]any{"target_athlete_id": target}), nil
		},
	}
}

func coachACLTestConfig() config.Config {
	return config.Config{
		AthleteID: "i111",
		CoachMode: coach.ModeOn,
		Coach: coach.Config{
			DefaultAthleteID: "i111",
			Athletes: []coach.Athlete{
				{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile}, DeniedTools: []string{toolcatalog.GetPowerCurves}},
				{ID: "i222", AllowedTools: []string{"get_*"}},
			},
		},
	}
}

func TestProtocolCoachACLFiltersCatalogAndResolvesAthleteID(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: coachACLTestConfig(), Registry: coachACLTestRegistry{}, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	names := make([]string, 0, len(result.Tools))
	var profileSchema map[string]any
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
		if tool.Name == toolcatalog.GetAthleteProfile {
			profileSchema, _ = tool.InputSchema.(map[string]any)
		}
	}
	if !slices.Contains(names, toolcatalog.GetAthleteProfile) {
		t.Fatalf("tools/list = %v, missing allowed profile tool", names)
	}
	for _, denied := range []string{toolcatalog.AddOrUpdateEvent, toolcatalog.DeleteEvent, toolcatalog.GetPowerCurves} {
		if slices.Contains(names, denied) {
			t.Fatalf("tools/list = %v, leaked denied tool %s", names, denied)
		}
	}
	props, _ := profileSchema["properties"].(map[string]any)
	if _, ok := props["athlete_id"]; !ok {
		t.Fatalf("get_athlete_profile schema = %#v, missing athlete_id", profileSchema)
	}

	call, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: map[string]any{"athlete_id": "i222"}})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if call.IsError {
		t.Fatalf("CallTool() IsError = true, content = %#v", call.Content)
	}
	text := call.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(text, `"target_athlete_id":"i222"`) {
		t.Fatalf("CallTool() text = %s, want target i222", text)
	}
}

func TestProtocolComputeZoneEnergyIsFullOnlyReadAnalyzer(t *testing.T) {
	t.Parallel()

	client := newNoNetworkProtocolClient(t)
	coreRegistry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetCore})
	coreCtx, coreSession, coreCleanup := connectTestClientWithOptions(t, Options{Registry: coreRegistry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetCore})
	defer coreCleanup()
	core, err := coreSession.ListTools(coreCtx, nil)
	if err != nil {
		t.Fatalf("core ListTools() error = %v", err)
	}
	if len(core.Tools) != 28 {
		t.Fatalf("core tools/list count = %d, want 28", len(core.Tools))
	}
	for _, fullOnly := range []string{toolcatalog.ComputeZoneEnergy, toolcatalog.CreateSportSettings} {
		if hasTool(core.Tools, fullOnly) {
			t.Fatalf("core tools/list exposed full-only %s", fullOnly)
		}
	}

	fullRegistry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	fullCtx, fullSession, fullCleanup := connectTestClientWithOptions(t, Options{Registry: fullRegistry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer fullCleanup()
	full, err := fullSession.ListTools(fullCtx, nil)
	if err != nil {
		t.Fatalf("full ListTools() error = %v", err)
	}
	if len(full.Tools) != 68 {
		t.Fatalf("full tools/list count = %d, want 68", len(full.Tools))
	}
	var zoneEnergy *sdkmcp.Tool
	var createSportSettings *sdkmcp.Tool
	for _, tool := range full.Tools {
		switch tool.Name {
		case toolcatalog.ComputeZoneEnergy:
			zoneEnergy = tool
		case toolcatalog.CreateSportSettings:
			createSportSettings = tool
		}
	}
	if zoneEnergy == nil {
		t.Fatalf("full tools/list missing %s", toolcatalog.ComputeZoneEnergy)
	}
	if createSportSettings == nil {
		t.Fatalf("full tools/list missing %s", toolcatalog.CreateSportSettings)
	}
	if zoneEnergy.Annotations == nil || !zoneEnergy.Annotations.ReadOnlyHint {
		t.Fatalf("%s annotations = %#v, want readOnlyHint true", toolcatalog.ComputeZoneEnergy, zoneEnergy.Annotations)
	}
}

func TestProtocolAthleteScopedSchemasExposeUniformAthleteID(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	cfg := config.Config{AthleteID: "i111", CoachMode: coach.ModeOn, Coach: coach.Config{DefaultAthleteID: "i111", Athletes: []coach.Athlete{{ID: "i111", AllowedTools: []string{"*"}}}}}
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	seen := map[string]struct{}{}
	for _, tool := range result.Tools {
		schema, _ := tool.InputSchema.(map[string]any)
		props, _ := schema["properties"].(map[string]any)
		_, hasAthleteID := props["athlete_id"]
		if toolcatalog.IsAthleteScopedTool(tool.Name) {
			seen[tool.Name] = struct{}{}
			if !hasAthleteID {
				t.Fatalf("%s schema missing athlete_id: %#v", tool.Name, schema)
			}
			arg, _ := props["athlete_id"].(map[string]any)
			if arg["description"] != athleteIDArgumentDescription {
				t.Fatalf("%s athlete_id description = %#v, want %q", tool.Name, arg["description"], athleteIDArgumentDescription)
			}
			if tool.Name == toolcatalog.CreateSportSettings {
				assertCoachCreateSportSettingsSchemaSafe(t, schema)
			}
		} else if hasAthleteID {
			t.Fatalf("non-athlete tool %s unexpectedly has athlete_id", tool.Name)
		}
	}
	for _, name := range toolcatalog.AthleteScopedToolNames() {
		if _, ok := seen[name]; !ok {
			t.Fatalf("athlete-scoped tool %s was not registered in full catalog", name)
		}
	}
}

func assertCoachCreateSportSettingsSchemaSafe(t *testing.T, schema map[string]any) {
	t.Helper()

	if schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("create_sport_settings schema = %#v, want closed object", schema)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("create_sport_settings properties = %#v, want object", schema["properties"])
	}
	allowed := map[string]struct{}{
		"sport": {}, "ftp": {}, "indoor_ftp": {}, "threshold_hr": {}, "threshold_pace": {}, "athlete_id": {},
	}
	if len(properties) != len(allowed) {
		t.Fatalf("create_sport_settings properties = %#v, want only documented arguments plus athlete_id", properties)
	}
	for name := range allowed {
		if _, ok := properties[name]; !ok {
			t.Fatalf("create_sport_settings properties = %#v, missing %q", properties, name)
		}
	}
	for _, forbidden := range []string{
		"api_key", "apikey", "apiKey", "token", "credential", "credential_ref", "credentials",
		"confirm", "recalc_hr_zones", "recalcHrZones", "zones", "power_zones", "power_zone_names",
		"hr_zones", "hr_zone_names", "pace_zones", "pace_zone_names", "apply", "apply_to_activities",
	} {
		if _, ok := properties[forbidden]; ok {
			t.Fatalf("create_sport_settings coach schema exposes forbidden %q: %#v", forbidden, properties)
		}
	}
}

func TestProtocolCoachToolsAbsentWhenCoachModeOff(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC"})
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Registry: registry})
	defer cleanup()

	result, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	for _, forbidden := range []string{toolcatalog.ListAthletes, toolcatalog.SelectAthlete} {
		if hasTool(result.Tools, forbidden) {
			t.Fatalf("coach mode off tools/list exposed %s", forbidden)
		}
		if _, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: forbidden, Arguments: map[string]any{}}); err == nil || !strings.Contains(err.Error(), "unknown tool") {
			t.Fatalf("CallTool(%s) error = %v, want unknown tool", forbidden, err)
		}
	}
}

func TestProtocolLocalModeRejectsAthleteIDOverrideAndUsesConfiguredAthlete(t *testing.T) {
	t.Parallel()

	cfg := config.Config{AthleteID: "i111"}
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: coachACLTestRegistry{}, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer cleanup()

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatalf("local default CallTool() = %#v err=%v, want configured athlete fallback", result, err)
	}
	if text := result.Content[0].(*sdkmcp.TextContent).Text; !strings.Contains(text, `"target_athlete_id":"i111"`) {
		t.Fatalf("local default CallTool() text = %s, want configured target i111", text)
	}

	rejected, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: map[string]any{"athlete_id": "i222"}})
	if err != nil {
		t.Fatalf("local override CallTool() protocol error = %v", err)
	}
	if !rejected.IsError {
		t.Fatalf("local override IsError = false, want true")
	}
	if text := rejected.Content[0].(*sdkmcp.TextContent).Text; text != localModeAthleteTargetMessage {
		t.Fatalf("local override error text = %q, want %q", text, localModeAthleteTargetMessage)
	}
}

func TestProtocolSelectAthleteUpdatesVisibleCatalogAndListAthletes(t *testing.T) {
	t.Parallel()

	cfg := coachACLTestConfig()
	registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull, CoachModeEnabled: true, CoachConfig: cfg.Coach})
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer cleanup()

	initial, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("initial ListTools() error = %v", err)
	}
	if hasTool(initial.Tools, toolcatalog.GetPowerCurves) {
		t.Fatalf("initial tools/list leaked second-athlete tool %s", toolcatalog.GetPowerCurves)
	}
	if !hasTool(initial.Tools, toolcatalog.ListAthletes) || !hasTool(initial.Tools, toolcatalog.SelectAthlete) {
		t.Fatalf("initial tools/list missing coach tools: %#v", initial.Tools)
	}

	listResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.ListAthletes, Arguments: map[string]any{}})
	if err != nil || listResult.IsError {
		t.Fatalf("list_athletes result = %#v err=%v", listResult, err)
	}
	var listParsed struct {
		Meta struct {
			Source          string `json:"source"`
			ActiveAthleteID string `json:"active_athlete_id"`
		} `json:"_meta"`
	}
	if err := json.Unmarshal([]byte(listResult.Content[0].(*sdkmcp.TextContent).Text), &listParsed); err != nil {
		t.Fatalf("unmarshal list_athletes: %v", err)
	}
	if listParsed.Meta.Source != "config" || listParsed.Meta.ActiveAthleteID != "i111" {
		t.Fatalf("list_athletes _meta = %#v, want config source and active i111", listParsed.Meta)
	}

	selectResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.SelectAthlete, Arguments: map[string]any{"athlete_id": "i222"}})
	if err != nil || selectResult.IsError {
		t.Fatalf("select_athlete result = %#v err=%v", selectResult, err)
	}
	var selectParsed struct {
		PreviousAthleteID string   `json:"previous_athlete_id"`
		NewAthleteID      string   `json:"new_athlete_id"`
		AllowedTools      []string `json:"allowed_tools"`
		Meta              struct {
			RequiresNewConversation bool `json:"requires_new_conversation"`
		} `json:"_meta"`
	}
	if err := json.Unmarshal([]byte(selectResult.Content[0].(*sdkmcp.TextContent).Text), &selectParsed); err != nil {
		t.Fatalf("unmarshal select_athlete: %v", err)
	}
	if selectParsed.PreviousAthleteID != "i111" || selectParsed.NewAthleteID != "i222" || !selectParsed.Meta.RequiresNewConversation {
		t.Fatalf("select_athlete response = %#v, want previous/new and requires_new_conversation", selectParsed)
	}

	after, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("after ListTools() error = %v", err)
	}
	requireSameStrings(t, "select_athlete.allowed_tools", selectParsed.AllowedTools, toolNamesFromList(after.Tools))
	if !hasTool(after.Tools, toolcatalog.GetPowerCurves) {
		t.Fatalf("after tools/list = %#v, missing selected-athlete full read tool", after.Tools)
	}
}

func hasTool(tools []*sdkmcp.Tool, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func toolNamesFromList(tools []*sdkmcp.Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	slices.Sort(names)
	return names
}

func requireSameStrings(t *testing.T, label string, got []string, want []string) {
	t.Helper()
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	slices.Sort(got)
	slices.Sort(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func advancedCapabilityNames(t *testing.T, result *sdkmcp.CallToolResult) []string {
	t.Helper()
	var parsed struct {
		AdvancedCapabilities []struct {
			Name string `json:"name"`
		} `json:"advanced_capabilities"`
		Meta struct {
			Count int `json:"count"`
		} `json:"_meta"`
	}
	if err := json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &parsed); err != nil {
		t.Fatalf("unmarshal advanced capabilities: %v", err)
	}
	if parsed.Meta.Count != len(parsed.AdvancedCapabilities) {
		t.Fatalf("advanced capabilities count = %d, want %d rows", parsed.Meta.Count, len(parsed.AdvancedCapabilities))
	}
	names := make([]string, 0, len(parsed.AdvancedCapabilities))
	for _, row := range parsed.AdvancedCapabilities {
		names = append(names, row.Name)
	}
	slices.Sort(names)
	return names
}

func TestProtocolGateCompositionTruthTable(t *testing.T) {
	t.Parallel()

	type representativeTool struct {
		name        string
		toolset     safety.Toolset
		requirement tools.Requirement
	}
	representatives := []representativeTool{
		{name: toolcatalog.DeleteEvent, toolset: safety.ToolsetCore, requirement: tools.RequirementDelete},
		{name: toolcatalog.GetAthleteProfile, toolset: safety.ToolsetCore, requirement: tools.RequirementRead},
		{name: toolcatalog.AddOrUpdateEvent, toolset: safety.ToolsetCore, requirement: tools.RequirementWrite},
		{name: toolcatalog.GetPowerCurves, toolset: safety.ToolsetFull, requirement: tools.RequirementRead},
	}

	for _, representative := range representatives {
		for _, mode := range []safety.Mode{safety.ModeNone, safety.ModeSafe, safety.ModeFull} {
			for _, toolset := range []safety.Toolset{safety.ToolsetCore, safety.ToolsetFull} {
				for _, coachAllows := range []bool{false, true} {
					coachLabel := "coach-deny"
					if coachAllows {
						coachLabel = "coach-allow"
					}
					t.Run(representative.name+"/delete-mode-"+mode.String()+"/toolset-"+toolset.String()+"/"+coachLabel, func(t *testing.T) {
						t.Parallel()

						activeAthlete := coach.Athlete{ID: "i111", AllowedTools: []string{representative.name}}
						if !coachAllows {
							activeAthlete.DeniedTools = []string{representative.name}
						}
						cfg := config.Config{
							AthleteID: "i111",
							CoachMode: coach.ModeOn,
							Coach:     coach.Config{DefaultAthleteID: "i111", Athletes: []coach.Athlete{activeAthlete, {ID: "i222", AllowedTools: []string{representative.name}}}},
						}
						registry := registryFunc(func(_ context.Context, registrar tools.Registrar) error {
							return registrar.AddTool(coachACLTestTool(representative.name, representative.toolset, representative.requirement))
						})
						ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: safety.NewCapability(mode), Toolset: toolset})
						defer cleanup()

						result, err := session.ListTools(ctx, nil)
						if err != nil {
							t.Fatalf("ListTools() error = %v", err)
						}
						capabilityAllows := representative.requirement == tools.RequirementRead || representative.requirement == tools.RequirementWrite && mode != safety.ModeNone || representative.requirement == tools.RequirementDelete && mode == safety.ModeFull
						toolsetAllows := representative.toolset == safety.ToolsetCore || toolset == safety.ToolsetFull
						wantVisible := capabilityAllows && toolsetAllows && coachAllows
						if got := hasTool(result.Tools, representative.name); got != wantVisible {
							t.Fatalf("tools/list visibility for %s = %t, want %t", representative.name, got, wantVisible)
						}

						call, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: representative.name, Arguments: map[string]any{}})
						switch {
						case wantVisible:
							if err != nil || call.IsError {
								t.Fatalf("CallTool(%s) = %#v err=%v, want success", representative.name, call, err)
							}
							if text := call.Content[0].(*sdkmcp.TextContent).Text; !strings.Contains(text, `"target_athlete_id":"i111"`) {
								t.Fatalf("CallTool(%s) text = %s, want default target i111", representative.name, text)
							}
						case !capabilityAllows || !toolsetAllows:
							if err == nil || !strings.Contains(err.Error(), "unknown tool") {
								t.Fatalf("CallTool(%s) error = %v, want unknown tool from registration gate", representative.name, err)
							}
						default:
							if err != nil {
								t.Fatalf("CallTool(%s) protocol error = %v", representative.name, err)
							}
							if !call.IsError {
								t.Fatalf("CallTool(%s) IsError = false, want coach ACL veto", representative.name)
							}
							if text := call.Content[0].(*sdkmcp.TextContent).Text; text != toolNotAllowedForAthleteMessage {
								t.Fatalf("CallTool(%s) error text = %q, want %q", representative.name, text, toolNotAllowedForAthleteMessage)
							}
						}
					})
				}
			}
		}
	}
}

func TestProtocolVisibleCatalogMetadataMatchesToolsListAndSessionsAreIsolated(t *testing.T) {
	cfg := coachACLTestConfig()
	registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull, CoachModeEnabled: true, CoachConfig: cfg.Coach})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server, err := NewServer(ctx, Options{Version: "test", Config: cfg, Registry: registry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	if err != nil {
		listener.Close()
		t.Fatalf("NewServer() error = %v", err)
	}
	runDone := make(chan error, 1)
	go func() { runDone <- server.ServeStreamableHTTP(ctx, listener) }()
	endpoint := "http://" + listener.Addr().String() + StreamableHTTPPath
	connect := func(name string) *sdkmcp.ClientSession {
		t.Helper()
		client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: name, Version: "test"}, nil)
		session, err := client.Connect(ctx, &sdkmcp.StreamableClientTransport{Endpoint: endpoint, HTTPClient: &http.Client{Timeout: 2 * time.Second}, MaxRetries: -1, DisableStandaloneSSE: true}, nil)
		if err != nil {
			t.Fatalf("%s Connect() error = %v", name, err)
		}
		return session
	}
	sessionA := connect("icuvisor-test-client-a")
	sessionB := connect("icuvisor-test-client-b")
	defer func() {
		sessionB.Close()
		sessionA.Close()
		cancel()
		waitForServerRun(t, runDone)
	}()

	initialB, err := sessionB.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("session B initial ListTools() error = %v", err)
	}
	if hasTool(initialB.Tools, toolcatalog.GetPowerCurves) {
		t.Fatalf("session B initial tools/list leaked %s", toolcatalog.GetPowerCurves)
	}

	selectedA, err := sessionA.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.SelectAthlete, Arguments: map[string]any{"athlete_id": "i222"}})
	if err != nil || selectedA.IsError {
		t.Fatalf("session A select_athlete = %#v err=%v", selectedA, err)
	}
	var selected struct {
		AllowedTools []string `json:"allowed_tools"`
	}
	if err := json.Unmarshal([]byte(selectedA.Content[0].(*sdkmcp.TextContent).Text), &selected); err != nil {
		t.Fatalf("unmarshal select_athlete: %v", err)
	}
	afterA, err := sessionA.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("session A after ListTools() error = %v", err)
	}
	requireSameStrings(t, "select_athlete.allowed_tools", selected.AllowedTools, toolNamesFromList(afterA.Tools))
	if !hasTool(afterA.Tools, toolcatalog.GetPowerCurves) {
		t.Fatalf("session A tools/list missing %s after selecting i222", toolcatalog.GetPowerCurves)
	}

	afterB, err := sessionB.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("session B after ListTools() error = %v", err)
	}
	requireSameStrings(t, "session B tools/list", toolNamesFromList(afterB.Tools), toolNamesFromList(initialB.Tools))
	if hasTool(afterB.Tools, toolcatalog.GetPowerCurves) {
		t.Fatalf("session B tools/list changed after session A selection: %#v", afterB.Tools)
	}
}

func TestProtocolSelectAthleteMetadataUsesPostGateCatalog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mode      safety.Mode
		toolset   safety.Toolset
		athletes  []coach.Athlete
		wantNew   bool
		forbidden string
	}{
		{name: "delete hidden by mode", mode: safety.ModeSafe, toolset: safety.ToolsetFull, athletes: []coach.Athlete{{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile, toolcatalog.DeleteEvent}}, {ID: "i222", AllowedTools: []string{toolcatalog.GetAthleteProfile}}}, wantNew: false, forbidden: toolcatalog.DeleteEvent},
		{name: "full hidden by core", mode: safety.ModeFull, toolset: safety.ToolsetCore, athletes: []coach.Athlete{{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile, toolcatalog.GetPowerCurves}}, {ID: "i222", AllowedTools: []string{toolcatalog.GetAthleteProfile}}}, wantNew: false, forbidden: toolcatalog.GetPowerCurves},
		{name: "visible core read changes", mode: safety.ModeFull, toolset: safety.ToolsetFull, athletes: []coach.Athlete{{ID: "i111", AllowedTools: []string{toolcatalog.GetAthleteProfile}}, {ID: "i222", AllowedTools: []string{"get_*"}}}, wantNew: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Config{AthleteID: "i111", CoachMode: coach.ModeOn, Coach: coach.Config{DefaultAthleteID: "i111", Athletes: tc.athletes}}
			registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(tc.mode), Toolset: tc.toolset, CoachModeEnabled: true, CoachConfig: cfg.Coach})
			ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: safety.NewCapability(tc.mode), Toolset: tc.toolset})
			defer cleanup()

			result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.SelectAthlete, Arguments: map[string]any{"athlete_id": "i222"}})
			if err != nil || result.IsError {
				t.Fatalf("select_athlete result = %#v err=%v", result, err)
			}
			var parsed struct {
				AllowedTools []string `json:"allowed_tools"`
				Meta         struct {
					RequiresNewConversation bool `json:"requires_new_conversation"`
				} `json:"_meta"`
			}
			if err := json.Unmarshal([]byte(result.Content[0].(*sdkmcp.TextContent).Text), &parsed); err != nil {
				t.Fatalf("unmarshal select_athlete: %v", err)
			}
			if parsed.Meta.RequiresNewConversation != tc.wantNew {
				t.Fatalf("requires_new_conversation = %t, want %t; tools=%v", parsed.Meta.RequiresNewConversation, tc.wantNew, parsed.AllowedTools)
			}
			if tc.forbidden != "" && slices.Contains(parsed.AllowedTools, tc.forbidden) {
				t.Fatalf("allowed_tools = %v, should not include hidden %s", parsed.AllowedTools, tc.forbidden)
			}
		})
	}
}

func TestProtocolCoachModeEndToEndRoutesSelectedDefaultAndOverrideTargets(t *testing.T) {
	cfg := config.Config{
		AthleteID: "i111",
		CoachMode: coach.ModeOn,
		Coach: coach.Config{
			DefaultAthleteID: "i111",
			Athletes: []coach.Athlete{
				{ID: "i111", Label: "Full access athlete", AllowedTools: []string{"*"}},
				{ID: "i222", Label: "Read only athlete", AllowedTools: []string{"get_*"}},
			},
		},
	}
	var upstreamRequests atomic.Int64
	client, closeServer := newProtocolIntervalsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamRequests.Add(1)
		if r.Method != http.MethodGet || (r.URL.Path != "/athlete/i111" && r.URL.Path != "/athlete/i222") {
			t.Fatalf("unexpected intervals request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/athlete/i111":
			_, _ = io.WriteString(w, `{"id":"i111","name":"Full Athlete","measurement_preference":"METRIC","timezone":"UTC","sportSettings":[{"types":["Ride"],"ftp":300}]}`)
		case "/athlete/i222":
			_, _ = io.WriteString(w, `{"id":"i222","name":"Read Only Athlete","measurement_preference":"METRIC","timezone":"UTC","sportSettings":[{"types":["Ride"],"ftp":200}]}`)
		}
	}))
	defer closeServer()

	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull, CoachModeEnabled: true, CoachConfig: cfg.Coach})
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer cleanup()

	initial, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("initial ListTools() error = %v", err)
	}
	for _, want := range []string{toolcatalog.GetAthleteProfile, toolcatalog.AddOrUpdateEvent, toolcatalog.DeleteEvent, toolcatalog.SelectAthlete, toolcatalog.ListAthletes} {
		if !hasTool(initial.Tools, want) {
			t.Fatalf("initial tools/list missing %s", want)
		}
	}

	profile, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: map[string]any{}})
	if err != nil || profile.IsError {
		t.Fatalf("default get_athlete_profile = %#v err=%v", profile, err)
	}
	if text := profile.Content[0].(*sdkmcp.TextContent).Text; !strings.Contains(text, `"athlete_id":"i111"`) || !strings.Contains(text, `"ftp_watts":300`) {
		t.Fatalf("default profile text = %s, want i111 route", text)
	}

	selected, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.SelectAthlete, Arguments: map[string]any{"athlete_id": "i222"}})
	if err != nil || selected.IsError {
		t.Fatalf("select_athlete = %#v err=%v", selected, err)
	}
	selectedText := selected.Content[0].(*sdkmcp.TextContent).Text
	if !strings.Contains(selectedText, `"new_athlete_id":"i222"`) || !strings.Contains(selectedText, `"requires_new_conversation":true`) {
		t.Fatalf("select_athlete text = %s, want selected i222 with catalog caveat", selectedText)
	}

	afterSelect, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("after-select ListTools() error = %v", err)
	}
	if !hasTool(afterSelect.Tools, toolcatalog.GetAthleteProfile) {
		t.Fatalf("read-only tools/list missing %s", toolcatalog.GetAthleteProfile)
	}
	for _, forbidden := range []string{toolcatalog.AddOrUpdateEvent, toolcatalog.DeleteEvent} {
		if hasTool(afterSelect.Tools, forbidden) {
			t.Fatalf("read-only tools/list exposed %s", forbidden)
		}
	}

	readOnlyProfile, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: map[string]any{}})
	if err != nil || readOnlyProfile.IsError {
		t.Fatalf("selected get_athlete_profile = %#v err=%v", readOnlyProfile, err)
	}
	if text := readOnlyProfile.Content[0].(*sdkmcp.TextContent).Text; !strings.Contains(text, `"athlete_id":"i222"`) || !strings.Contains(text, `"ftp_watts":200`) {
		t.Fatalf("selected profile text = %s, want i222 route", text)
	}

	overrideProfile, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: map[string]any{"athlete_id": "i111"}})
	if err != nil || overrideProfile.IsError {
		t.Fatalf("override get_athlete_profile = %#v err=%v", overrideProfile, err)
	}
	if text := overrideProfile.Content[0].(*sdkmcp.TextContent).Text; !strings.Contains(text, `"athlete_id":"i111"`) || !strings.Contains(text, `"ftp_watts":300`) {
		t.Fatalf("override profile text = %s, want i111 route", text)
	}

	beforeDenied := upstreamRequests.Load()
	for _, denied := range []struct {
		name string
		args map[string]any
		want string
	}{
		{name: toolcatalog.AddOrUpdateEvent, args: map[string]any{"athlete_id": "i222", "date": "2026-05-15", "category": "NOTE", "name": "Denied write"}, want: toolNotAllowedForAthleteMessage},
		{name: toolcatalog.DeleteEvent, args: map[string]any{"athlete_id": "i222", "event_id": "e-denied"}, want: toolNotAllowedForAthleteMessage},
		{name: toolcatalog.GetAthleteProfile, args: map[string]any{"athlete_id": "i999"}, want: unauthorizedTargetAthleteMessage},
	} {
		result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: denied.name, Arguments: denied.args})
		if err != nil {
			t.Fatalf("CallTool(%s) protocol error = %v", denied.name, err)
		}
		if !result.IsError {
			t.Fatalf("CallTool(%s) IsError = false, want request-time coach authorization denial", denied.name)
		}
		if text := result.Content[0].(*sdkmcp.TextContent).Text; text != denied.want {
			t.Fatalf("CallTool(%s) error text = %q, want %q", denied.name, text, denied.want)
		}
	}
	if afterDenied := upstreamRequests.Load(); afterDenied != beforeDenied {
		t.Fatalf("denied write/delete upstream requests = %d after denial, want unchanged %d", afterDenied, beforeDenied)
	}
}

func TestProtocolAthleteIDRejectionMessageIsEnumerationSafe(t *testing.T) {
	t.Parallel()

	ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: coachACLTestConfig(), Registry: coachACLTestRegistry{}, Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull})
	defer cleanup()

	for _, tc := range []struct {
		name string
		args map[string]any
		want string
	}{
		{name: "invalid format", args: map[string]any{"athlete_id": "not-an-id"}, want: invalidTargetAthleteFormatMessage},
		{name: "unauthorized roster target", args: map[string]any{"athlete_id": "i999"}, want: unauthorizedTargetAthleteMessage},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.GetAthleteProfile, Arguments: tc.args})
			if err != nil {
				t.Fatalf("CallTool() protocol error = %v", err)
			}
			if !result.IsError {
				t.Fatalf("CallTool(%v) IsError = false, want true", tc.args)
			}
			text := result.Content[0].(*sdkmcp.TextContent).Text
			if text != tc.want {
				t.Fatalf("CallTool(%v) error text = %q, want %q", tc.args, text, tc.want)
			}
		})
	}
}

func TestProtocolListAdvancedCapabilitiesVisibilityWithRealRegistry(t *testing.T) {
	t.Parallel()

	registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC"})
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
	for _, wantName := range []string{"get_athlete_profile", "get_activities", "icuvisor_list_advanced_capabilities"} {
		if !slices.Contains(got, wantName) {
			t.Fatalf("core tools/list = %v, missing %s", got, wantName)
		}
	}
	for _, hiddenName := range []string{"get_power_curves", "get_gear_list"} {
		if slices.Contains(got, hiddenName) {
			t.Fatalf("core tools/list = %v, should hide %s", got, hiddenName)
		}
	}

	fullRegistry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Toolset: safety.ToolsetFull})
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
	for _, wantName := range []string{"get_power_curves", "get_gear_list", "icuvisor_list_advanced_capabilities"} {
		if !slices.Contains(fullNames, wantName) {
			t.Fatalf("full tools/list = %v, missing %s", fullNames, wantName)
		}
	}
	if slices.Contains(fullNames, "delete_gear") {
		t.Fatalf("safe full tools/list = %v, should keep delete_gear hidden while get_gear_list remains visible", fullNames)
	}
}

func TestProtocolAdvancedCapabilitiesUsesCoachFilteredCatalog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mode          safety.Mode
		toolset       safety.Toolset
		activeAllowed []string
		wantTools     []string
		wantAdvanced  []string
	}{
		{
			name:          "safe core hides delete rows but advertises full read rows",
			mode:          safety.ModeSafe,
			toolset:       safety.ToolsetCore,
			activeAllowed: []string{"*"},
			wantTools:     []string{toolcatalog.GetAthleteProfile, toolcatalog.ICUvisorListAdvancedCapabilities},
			wantAdvanced:  []string{toolcatalog.GetPowerCurves},
		},
		{
			name:          "full core advertises post-capability full rows hidden by toolset",
			mode:          safety.ModeFull,
			toolset:       safety.ToolsetCore,
			activeAllowed: []string{"*"},
			wantTools:     []string{toolcatalog.GetAthleteProfile, toolcatalog.ICUvisorListAdvancedCapabilities},
			wantAdvanced:  []string{toolcatalog.DeleteEvent, toolcatalog.GetPowerCurves},
		},
		{
			name:          "full full keeps advanced rows aligned with active athlete ACL",
			mode:          safety.ModeFull,
			toolset:       safety.ToolsetFull,
			activeAllowed: []string{toolcatalog.GetAthleteProfile, toolcatalog.GetPowerCurves},
			wantTools:     []string{toolcatalog.GetAthleteProfile, toolcatalog.GetPowerCurves, toolcatalog.ICUvisorListAdvancedCapabilities},
			wantAdvanced:  []string{toolcatalog.GetPowerCurves},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := config.Config{AthleteID: "i111", CoachMode: coach.ModeOn, Coach: coach.Config{DefaultAthleteID: "i111", Athletes: []coach.Athlete{{ID: "i111", AllowedTools: tc.activeAllowed}, {ID: "i222", AllowedTools: []string{"*"}}}}}
			registry := registryFunc(func(_ context.Context, registrar tools.Registrar) error {
				for _, tool := range []tools.Tool{
					coachACLTestTool(toolcatalog.GetAthleteProfile, safety.ToolsetCore, tools.RequirementRead),
					coachACLTestTool(toolcatalog.GetPowerCurves, safety.ToolsetFull, tools.RequirementRead),
					coachACLTestTool(toolcatalog.DeleteEvent, safety.ToolsetFull, tools.RequirementDelete),
					coachACLTestTool(toolcatalog.ICUvisorListAdvancedCapabilities, safety.ToolsetCore, tools.RequirementRead),
				} {
					if err := registrar.AddTool(tool); err != nil {
						return err
					}
				}
				return nil
			})
			ctx, session, cleanup := connectTestClientWithOptions(t, Options{Config: cfg, Registry: registry, Capability: safety.NewCapability(tc.mode), Toolset: tc.toolset})
			defer cleanup()

			listed, err := session.ListTools(ctx, nil)
			if err != nil {
				t.Fatalf("ListTools() error = %v", err)
			}
			requireSameStrings(t, "tools/list", toolNamesFromList(listed.Tools), tc.wantTools)

			result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: toolcatalog.ICUvisorListAdvancedCapabilities, Arguments: map[string]any{}})
			if err != nil {
				t.Fatalf("CallTool(advanced) error = %v", err)
			}
			if result.IsError {
				t.Fatalf("CallTool(advanced) IsError = true, content = %#v", result.Content)
			}
			requireSameStrings(t, "advanced_capabilities names", advancedCapabilityNames(t, result), tc.wantAdvanced)
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

	client, closeServer := newProtocolIntervalsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/athlete/i12345" {
			t.Fatalf("request path = %s, want /athlete/i12345", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"i12345","name":"Example Athlete","measurement_preference":"METRIC","timezone":"America/Sao_Paulo","sportSettings":[{"types":["Ride"],"ftp":250}]}`)
	}))
	defer closeServer()
	registry := tools.NewRegistry(client, "v0.1-test", "UTC")
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
	for _, wantName := range []string{"get_athlete_profile", "get_activities", "icuvisor_list_advanced_capabilities"} {
		if !slices.Contains(gotNames, wantName) {
			t.Fatalf("tools/list = %v, missing %s", gotNames, wantName)
		}
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
	case <-time.After(10 * time.Second):
		t.Fatal("server Run() did not stop")
	}
}
