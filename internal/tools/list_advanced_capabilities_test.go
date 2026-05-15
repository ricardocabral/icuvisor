package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestListAdvancedCapabilitiesOutputFromCatalog(t *testing.T) {
	registrar := &collectingRegistrar{}
	client := newNoNetworkIntervalsClient(t)
	if err := NewRegistryWithOptions(client, RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull)}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, listAdvancedCapabilitiesName)
	if tool.EffectiveToolset() != safety.ToolsetCore {
		t.Fatalf("toolset = %q, want core", tool.EffectiveToolset())
	}

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := advancedCapabilitiesResult(t, result)
	if payload.CurrentToolset != "core" {
		t.Fatalf("current_toolset = %q, want core", payload.CurrentToolset)
	}
	if !strings.Contains(resultText(t, result), "ICUVISOR_TOOLSET=full") || !strings.Contains(payload.EnableInstruction, "ICUVISOR_TOOLSET=full") {
		t.Fatalf("enable instruction missing ICUVISOR_TOOLSET=full: text=%q structured=%q", resultText(t, result), payload.EnableInstruction)
	}
	if !strings.Contains(payload.EnableInstruction, "restart icuvisor") {
		t.Fatalf("enable instruction = %q, want restart guidance", payload.EnableInstruction)
	}
	if payload.Meta.Toolset != "core" {
		t.Fatalf("_meta.toolset = %q, want core", payload.Meta.Toolset)
	}
	if !strings.Contains(resultText(t, result), `"toolset":"core"`) {
		t.Fatalf("text JSON missing _meta.toolset core: %s", resultText(t, result))
	}
	if payload.Meta.Count != len(payload.AdvancedCapabilities) || payload.Meta.Source == "" || payload.Meta.DeleteModeNote == "" {
		t.Fatalf("_meta = %#v for %d capabilities", payload.Meta, len(payload.AdvancedCapabilities))
	}

	rows := map[string]advancedCapabilityRow{}
	for _, row := range payload.AdvancedCapabilities {
		rows[row.Name] = row
		if strings.Contains(row.Summary, "\n") {
			t.Fatalf("summary for %s is not one line: %q", row.Name, row.Summary)
		}
	}
	if _, ok := rows[getPowerCurvesName]; !ok {
		t.Fatalf("advanced capabilities missing %s: %#v", getPowerCurvesName, rows)
	}
	if rows[getPowerCurvesName].Requirement != string(RequirementRead) {
		t.Fatalf("%s requirement = %q, want read", getPowerCurvesName, rows[getPowerCurvesName].Requirement)
	}
	for _, excluded := range []string{getAthleteProfileName, listAdvancedCapabilitiesName} {
		if _, ok := rows[excluded]; ok {
			t.Fatalf("advanced capabilities included %s: %#v", excluded, rows[excluded])
		}
	}
}

func TestListAdvancedCapabilitiesFullModeStatus(t *testing.T) {
	registrar := &collectingRegistrar{}
	if err := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull), Toolset: safety.ToolsetFull}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, listAdvancedCapabilitiesName)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := advancedCapabilitiesResult(t, result)
	if payload.CurrentToolset != "full" {
		t.Fatalf("current_toolset = %q, want full", payload.CurrentToolset)
	}
	if !strings.Contains(payload.Status, "already enabled") {
		t.Fatalf("status = %q, want already enabled", payload.Status)
	}
	if payload.Meta.Toolset != "full" {
		t.Fatalf("_meta.toolset = %q, want full", payload.Meta.Toolset)
	}
	if !strings.Contains(resultText(t, result), `"toolset":"full"`) {
		t.Fatalf("text JSON missing _meta.toolset full: %s", resultText(t, result))
	}
	if len(payload.AdvancedCapabilities) == 0 {
		t.Fatal("advanced capabilities empty in full mode")
	}
}

func TestListAdvancedCapabilitiesRejectsArguments(t *testing.T) {
	t.Parallel()

	tool := newListAdvancedCapabilitiesTool(nil, safety.ToolsetCore)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":true}`)})
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "no arguments") {
		t.Fatalf("Handler() error = %v, public=%q ok=%v", err, message, ok)
	}
}

func advancedCapabilitiesResult(t *testing.T, result Result) listAdvancedCapabilitiesResponse {
	t.Helper()
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	var payload listAdvancedCapabilitiesResponse
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode StructuredContent: %v", err)
	}
	return payload
}
