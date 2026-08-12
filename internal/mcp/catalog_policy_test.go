package mcp

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

func TestServerCatalogPolicyIsIndependentFromExecutionPolicy(t *testing.T) {
	t.Parallel()

	client, err := intervals.NewClient(intervals.Options{
		Config: config.Config{
			APIKey:      "not-a-real-key",
			AthleteID:   "i12345",
			APIBaseURL:  "http://127.0.0.1",
			HTTPTimeout: time.Second,
		},
		Version: "test",
		HTTPClient: &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("destructive safe-mode call reached the upstream HTTP client")
			return nil, nil
		})},
	})
	if err != nil {
		t.Fatalf("intervals.NewClient() error = %v", err)
	}
	executionCapability := safety.NewCapability(safety.ModeSafe)
	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{
		Version:           "test",
		TimezoneFallback:  "UTC",
		Capability:        executionCapability,
		Toolset:           safety.ToolsetFull,
		CatalogCapability: safety.NewCapability(safety.ModeFull),
		CatalogToolset:    safety.ToolsetFull,
	})
	ctx, session, cleanup := connectTestClientWithOptions(t, Options{
		Config:            config.Config{AthleteID: "i12345", Timezone: "UTC"},
		Registry:          registry,
		Capability:        executionCapability,
		Toolset:           safety.ToolsetFull,
		CatalogCapability: safety.NewCapability(safety.ModeFull),
		CatalogToolset:    safety.ToolsetFull,
	})
	defer cleanup()

	listed, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	descriptions := make(map[string]string, len(listed.Tools))
	inputSchemas := make(map[string]string, len(listed.Tools))
	for _, tool := range listed.Tools {
		descriptions[tool.Name] = tool.Description
		schema, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatalf("json.Marshal(%s input schema) error = %v", tool.Name, err)
		}
		inputSchemas[tool.Name] = string(schema)
	}
	if got := descriptions["update_sport_settings"]; !strings.Contains(got, "supplying zones overwrites prior zone definitions") {
		t.Fatalf("update_sport_settings description = %q, want full catalog description", got)
	}
	if got := descriptions["apply_training_plan"]; !strings.Contains(got, "only replaces existing workout events") {
		t.Fatalf("apply_training_plan description = %q, want replacement policy description", got)
	}
	if got := inputSchemas["apply_training_plan"]; !strings.Contains(got, `"replace_existing"`) {
		t.Fatalf("apply_training_plan input schema omits full-catalog replacement option: %s", got)
	}
	if got := inputSchemas["apply_annual_training_plan"]; !strings.Contains(got, `"replace_icuvisor_notes"`) {
		t.Fatalf("apply_annual_training_plan input schema omits full-catalog replacement option: %s", got)
	}

	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{
			name: "update_sport_settings",
			args: map[string]any{"sport": "Run", "zones": []any{map[string]any{"kind": "hr", "boundaries": []any{0, 120}}}},
			want: "zones overwrite prior sport-setting zone definitions",
		},
		{
			name: "apply_training_plan",
			args: map[string]any{"plan_id": "plan-1", "start_date": "2026-06-01", "conflict_policy": "replace_existing"},
			want: "invalid apply_training_plan arguments",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: tc.name, Arguments: tc.args})
			if err != nil {
				t.Fatalf("CallTool() protocol error = %v", err)
			}
			if !result.IsError {
				t.Fatalf("CallTool() IsError = false, want safe execution denial")
			}
			if text := callToolText(result); !strings.Contains(text, tc.want) {
				t.Fatalf("CallTool() text = %q, want %q", text, tc.want)
			}
		})
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func callToolText(result *sdkmcp.CallToolResult) string {
	var text strings.Builder
	for _, content := range result.Content {
		if item, ok := content.(*sdkmcp.TextContent); ok {
			text.WriteString(item.Text)
		}
	}
	return text.String()
}
