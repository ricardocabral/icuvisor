package toolrouting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

func TestEnvProviderRequiresExplicitProvider(t *testing.T) {
	provider, name, configured, err := EnvProvider(func(string) string { return "" }, nil)
	if err != nil {
		t.Fatalf("EnvProvider() error = %v", err)
	}
	if provider != nil || name != "" || configured {
		t.Fatalf("EnvProvider() = provider %T name %q configured %t, want no provider", provider, name, configured)
	}
}

func TestEnvProviderRequiresAPIKeyWhenOptedIn(t *testing.T) {
	_, _, configured, err := EnvProvider(func(key string) string {
		if key == EnvRoutingEvalProvider {
			return "anthropic"
		}
		return ""
	}, nil)
	if !configured || err == nil || !strings.Contains(err.Error(), EnvAnthropicAPIKey) {
		t.Fatalf("EnvProvider() configured=%t error=%v, want missing key", configured, err)
	}
}

func TestLoadToolDefinitionsFiltersCatalog(t *testing.T) {
	coreSafe, err := LoadToolDefinitions(context.Background(), safety.ToolsetCore, safety.ModeSafe)
	if err != nil {
		t.Fatalf("LoadToolDefinitions(core/safe) error = %v", err)
	}
	if hasTool(coreSafe, toolcatalog.DeleteEvent) {
		t.Fatalf("core/safe catalog includes %s, want delete tools hidden", toolcatalog.DeleteEvent)
	}
	if !hasTool(coreSafe, toolcatalog.GetActivityDetails) {
		t.Fatalf("core/safe catalog missing %s", toolcatalog.GetActivityDetails)
	}
	fullDelete, err := LoadToolDefinitions(context.Background(), safety.ToolsetFull, safety.ModeFull)
	if err != nil {
		t.Fatalf("LoadToolDefinitions(full/full) error = %v", err)
	}
	if !hasTool(fullDelete, toolcatalog.DeleteEvent) || !hasTool(fullDelete, toolcatalog.CreateWorkout) {
		t.Fatalf("full/full catalog missing expected full/delete tools")
	}
}

func TestRunWithoutProviderSkipsWithoutNetwork(t *testing.T) {
	tool := toolcatalog.GetActivityDetails
	fixture := Fixture{Version: 1, Cases: []Case{{ID: "case", Prompt: "show details", ExpectedFirstTool: &tool, CatalogMode: "core_safe", Toolset: "core", DeleteMode: "safe"}}}
	var out strings.Builder
	summary, err := Run(context.Background(), fixture, nil, &out)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if summary.Total != 1 || summary.Skipped != 1 || summary.Passed != 0 || !strings.Contains(out.String(), "provider not configured") {
		t.Fatalf("Run() summary=%#v output=%q, want skipped dry run", summary, out.String())
	}
}

func TestAnthropicProviderFirstTool(t *testing.T) {
	apiKey := strings.Repeat("k", 8)
	var sawAPIKey bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		sawAPIKey = r.Header.Get("x-api-key") == apiKey
		var payload anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(payload.Tools) != 1 || payload.Tools[0].Name != toolcatalog.GetActivities {
			t.Fatalf("request tools = %#v, want get_activities", payload.Tools)
		}
		if payload.Temperature != 0 {
			t.Fatalf("temperature = %v, want deterministic zero", payload.Temperature)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"tool_use","name":"get_activities"}]}`))
	}))
	defer server.Close()

	provider := AnthropicProvider{APIKey: apiKey, Model: "model", Endpoint: server.URL, Client: server.Client()}
	selection, err := provider.FirstTool(context.Background(), PromptRequest{Case: Case{Prompt: "list rides"}, Tools: []ToolDefinition{{Name: toolcatalog.GetActivities, Description: "List activities.", InputSchema: json.RawMessage(`{"type":"object"}`)}}})
	if err != nil {
		t.Fatalf("FirstTool() error = %v", err)
	}
	if !sawAPIKey || selection.ToolName != toolcatalog.GetActivities || selection.RawMessage == "" {
		t.Fatalf("selection=%#v sawAPIKey=%t, want get_activities and API key header", selection, sawAPIKey)
	}
}

func hasTool(defs []ToolDefinition, name string) bool {
	for _, def := range defs {
		if def.Name == name {
			return true
		}
	}
	return false
}
