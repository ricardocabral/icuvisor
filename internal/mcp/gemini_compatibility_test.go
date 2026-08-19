package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type geminiCompatibilityRegistry struct{}

func (geminiCompatibilityRegistry) Register(_ context.Context, registrar tools.Registrar) error {
	if err := registrar.AddTool(tools.Tool{
		Name:        "get_athlete_profile",
		Description: "Returns a synthetic read-only athlete profile for MCP compatibility tests.",
		InputSchema: map[string]any{"type": "object", "additionalProperties": false},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"athlete_id": map[string]any{"type": "string"},
				"name":       map[string]any{"type": "string"},
			},
		},
		Toolset:     safety.ToolsetCore,
		Requirement: tools.RequirementRead,
		Handler: func(ctx context.Context, req tools.Request) (tools.Result, error) {
			if err := ctx.Err(); err != nil {
				return tools.Result{}, err
			}
			result := map[string]any{"athlete_id": "i00001", "name": "Synthetic Athlete"}
			text, err := json.Marshal(result)
			if err != nil {
				return tools.Result{}, err
			}
			return tools.Result{
				Content:           []tools.Content{{Type: tools.ContentTypeText, Text: string(text)}},
				StructuredContent: result,
			}, nil
		},
	}); err != nil {
		return err
	}
	return registrar.AddTool(tools.Tool{
		Name:        "compatibility_error",
		Description: "Returns a sanitized compatibility-test error.",
		InputSchema: map[string]any{"type": "object", "additionalProperties": true},
		Toolset:     safety.ToolsetCore,
		Requirement: tools.RequirementRead,
		Handler: func(context.Context, tools.Request) (tools.Result, error) {
			return tools.Result{}, errors.New("upstream failure contains synthetic-credential and request-payload")
		},
	})
}

type geminiCompatibilityResponse struct {
	status  int
	headers http.Header
	body    []byte
}

func TestGeminiSparkCompatibilityThroughCoreHTTPHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server, err := NewServer(ctx, Options{Version: "test", Registry: geminiCompatibilityRegistry{}})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	handler := NewStreamableHTTPHandler(func(*http.Request) (*Server, error) {
		return server, nil
	}, StreamableHTTPHandlerOptions{JSONResponse: true})
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)

	client := httpServer.Client()
	endpoint := httpServer.URL + StreamableHTTPPath
	initialize := geminiCompatibilityPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"gemini-spark-compatibility-test","version":"test"}}}`, "", "")
	if initialize.status != http.StatusOK {
		t.Fatalf("initialize status = %d body = %q, want 200", initialize.status, initialize.body)
	}
	assertGeminiJSONContentType(t, "initialize", initialize)
	sessionID := initialize.headers.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatalf("initialize response missing Mcp-Session-Id header; headers = %#v", initialize.headers)
	}
	initializeResult := assertGeminiJSONRPCResult(t, "initialize", initialize, 1)
	var initializeFields struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
	}
	if err := json.Unmarshal(initializeResult, &initializeFields); err != nil {
		t.Fatalf("initialize result decode error = %v; result = %s", err, initializeResult)
	}
	if initializeFields.ProtocolVersion != "2025-11-25" || initializeFields.ServerInfo.Name != "icuvisor" {
		t.Fatalf("initialize result = %s, want protocol version and icuvisor server info", initializeResult)
	}

	initialized := geminiCompatibilityPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`, sessionID, initializeFields.ProtocolVersion)
	if initialized.status != http.StatusAccepted || len(initialized.body) != 0 {
		t.Fatalf("notifications/initialized response = status %d body %q, want 202 with no body", initialized.status, initialized.body)
	}
	if initialized.headers.Get("Content-Type") != "" {
		t.Fatalf("notifications/initialized content type = %q, want empty", initialized.headers.Get("Content-Type"))
	}

	ping := geminiCompatibilityPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":2,"method":"ping","params":{}}`, sessionID, initializeFields.ProtocolVersion)
	if ping.status != http.StatusOK {
		t.Fatalf("ping status = %d body = %q, want 200", ping.status, ping.body)
	}
	assertGeminiJSONContentType(t, "ping", ping)
	assertGeminiJSONRPCResult(t, "ping", ping, 2)

	toolsList := geminiCompatibilityPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":3,"method":"tools/list","params":{}}`, sessionID, initializeFields.ProtocolVersion)
	if toolsList.status != http.StatusOK {
		t.Fatalf("tools/list status = %d body = %q, want 200", toolsList.status, toolsList.body)
	}
	assertGeminiJSONContentType(t, "tools/list", toolsList)
	toolsResult := assertGeminiJSONRPCResult(t, "tools/list", toolsList, 3)
	var listed struct {
		Tools []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(toolsResult, &listed); err != nil {
		t.Fatalf("tools/list result decode error = %v; result = %s", err, toolsResult)
	}
	var foundProfile bool
	for _, tool := range listed.Tools {
		if tool.Name == "get_athlete_profile" {
			foundProfile = true
			if tool.Description == "" {
				t.Fatal("tools/list get_athlete_profile description is empty")
			}
		}
	}
	if !foundProfile {
		t.Fatalf("tools/list = %s, missing get_athlete_profile", toolsResult)
	}

	call := geminiCompatibilityPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_athlete_profile","arguments":{}}}`, sessionID, initializeFields.ProtocolVersion)
	if call.status != http.StatusOK {
		t.Fatalf("tools/call status = %d body = %q, want 200", call.status, call.body)
	}
	assertGeminiJSONContentType(t, "tools/call", call)
	callResult := assertGeminiJSONRPCResult(t, "tools/call", call, 4)
	var called struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(callResult, &called); err != nil {
		t.Fatalf("tools/call result decode error = %v; result = %s", err, callResult)
	}
	if called.IsError || len(called.Content) != 1 || called.Content[0].Type != "text" || !strings.Contains(called.Content[0].Text, `"athlete_id":"i00001"`) {
		t.Fatalf("tools/call result = %s, want successful read-only profile content", callResult)
	}

	safeError := geminiCompatibilityPost(t, ctx, client, endpoint, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"compatibility_error","arguments":{"payload":"synthetic-credential"}}}`, sessionID, initializeFields.ProtocolVersion)
	if safeError.status != http.StatusOK {
		t.Fatalf("sanitized tools/call status = %d body = %q, want 200", safeError.status, safeError.body)
	}
	assertGeminiJSONContentType(t, "sanitized tools/call", safeError)
	safeErrorResult := assertGeminiJSONRPCResult(t, "sanitized tools/call", safeError, 5)
	if !strings.Contains(string(safeErrorResult), genericToolErrorMessage) || strings.Contains(string(safeErrorResult), "synthetic-credential") || strings.Contains(string(safeErrorResult), "request-payload") {
		t.Fatalf("sanitized tools/call result = %s, want generic error without credential or payload", safeErrorResult)
	}
}

func geminiCompatibilityPost(t *testing.T, ctx context.Context, client *http.Client, endpoint, payload, sessionID, protocolVersion string) geminiCompatibilityResponse {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(payload))
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
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
		t.Fatalf("reading POST response: %v", err)
	}
	return geminiCompatibilityResponse{status: response.StatusCode, headers: response.Header.Clone(), body: body}
}

func assertGeminiJSONContentType(t *testing.T, label string, response geminiCompatibilityResponse) {
	t.Helper()
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(response.headers.Get("Content-Type"), ";")[0]))
	if mediaType != "application/json" {
		t.Fatalf("%s content type = %q, want application/json", label, response.headers.Get("Content-Type"))
	}
}

func assertGeminiJSONRPCResult(t *testing.T, label string, response geminiCompatibilityResponse, wantID int) json.RawMessage {
	t.Helper()
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(response.body, &envelope); err != nil {
		t.Fatalf("%s response is not JSON-RPC JSON: %v; body = %q", label, err, response.body)
	}
	if got := string(envelope["jsonrpc"]); got != `"2.0"` {
		t.Fatalf("%s jsonrpc = %s, want \"2.0\"", label, got)
	}
	wantIDJSON, _ := json.Marshal(wantID)
	if got := string(envelope["id"]); got != string(wantIDJSON) {
		t.Fatalf("%s id = %s, want %d", label, got, wantID)
	}
	if _, ok := envelope["error"]; ok {
		t.Fatalf("%s response contains top-level error: %s", label, response.body)
	}
	result, ok := envelope["result"]
	if !ok || len(result) == 0 {
		t.Fatalf("%s response missing result: %s", label, response.body)
	}
	if !json.Valid(result) || !strings.HasPrefix(strings.TrimSpace(string(result)), "{") {
		t.Fatalf("%s result = %s, want JSON object", label, result)
	}
	return result
}
