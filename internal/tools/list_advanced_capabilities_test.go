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
	for _, name := range fullOnlyAnalyzerFamilyNames(registrar.tools) {
		row, ok := rows[name]
		if !ok {
			t.Fatalf("advanced capabilities missing full-only analyzer-family tool %s: %#v", name, rows)
		}
		if row.Summary == "" || !strings.HasPrefix(row.Summary, "Use when the prompt asks ") || !strings.Contains(row.Summary, "do not fetch") {
			t.Fatalf("%s summary = %q, want clear analyzer activation hint", name, row.Summary)
		}
	}
	if _, ok := rows[getPowerCurvesName]; !ok {
		t.Fatalf("advanced capabilities missing %s: %#v", getPowerCurvesName, rows)
	}
	segmentStats, ok := rows[computeActivitySegmentStatsName]
	if !ok {
		t.Fatalf("advanced capabilities missing %s: %#v", computeActivitySegmentStatsName, rows)
	}
	if !strings.Contains(segmentStats.Summary, "raw-stream exception") {
		t.Fatalf("%s summary = %q, want raw-stream exception", computeActivitySegmentStatsName, segmentStats.Summary)
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

func fullOnlyAnalyzerFamilyNames(catalog []Tool) []string {
	analyzerFamily := make(map[string]struct{}, len(analyzerFamilyCatalogNames()))
	for _, name := range analyzerFamilyCatalogNames() {
		analyzerFamily[name] = struct{}{}
	}
	var names []string
	for _, tool := range catalog {
		if _, ok := analyzerFamily[tool.Name]; ok && tool.EffectiveToolset() == safety.ToolsetFull {
			names = append(names, tool.Name)
		}
	}
	return names
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

func TestFilteredAdvancedCapabilitiesHandlerMatchesBaseContract(t *testing.T) {
	t.Parallel()

	catalog := []Tool{
		{Name: "z_full", Description: "Zeta. Details.", Toolset: safety.ToolsetFull, Requirement: RequirementRead},
		{Name: "a_full", Description: "Alpha. Details.", Toolset: safety.ToolsetFull, Requirement: RequirementWrite},
		{Name: "core_tool", Description: "Core.", Toolset: safety.ToolsetCore, Requirement: RequirementRead},
		{Name: listAdvancedCapabilitiesName, Description: "Self.", Toolset: safety.ToolsetCore, Requirement: RequirementRead},
	}
	handler := NewFilteredAdvancedCapabilitiesHandler(catalog, safety.ToolsetFull, func(_ context.Context, tool Tool) bool {
		return tool.Name != "z_full"
	})

	result, err := handler(context.Background(), Request{Name: listAdvancedCapabilitiesName, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := advancedCapabilitiesResult(t, result)
	if payload.CurrentToolset != "full" || !strings.Contains(payload.Status, "already enabled") {
		t.Fatalf("payload status/toolset = %#v", payload)
	}
	if payload.Meta.Toolset != "full" {
		t.Fatalf("_meta.toolset = %q, want full", payload.Meta.Toolset)
	}
	if !strings.Contains(payload.EnableInstruction, "full icuvisor toolset") {
		t.Fatalf("enable_instruction = %q, want legacy coach wording", payload.EnableInstruction)
	}
	if strings.Contains(resultText(t, result), `"delete_mode"`) {
		t.Fatalf("filtered text JSON invented delete_mode metadata: %s", resultText(t, result))
	}
	if payload.Meta.Count != 1 || payload.Meta.Count != len(payload.AdvancedCapabilities) {
		t.Fatalf("_meta.count = %d for rows %#v", payload.Meta.Count, payload.AdvancedCapabilities)
	}
	if got, want := payload.AdvancedCapabilities[0].Name, "a_full"; got != want {
		t.Fatalf("filtered row name = %q, want %q", got, want)
	}
	if got, want := payload.AdvancedCapabilities[0].Summary, "Alpha."; got != want {
		t.Fatalf("summary = %q, want %q", got, want)
	}
	if got, want := payload.AdvancedCapabilities[0].Requirement, string(RequirementWrite); got != want {
		t.Fatalf("requirement = %q, want %q", got, want)
	}
	if !strings.Contains(resultText(t, result), `"count":1`) {
		t.Fatalf("text JSON missing count: %s", resultText(t, result))
	}

	_, err = handler(context.Background(), Request{Name: listAdvancedCapabilitiesName, Arguments: json.RawMessage(`{"include_full":true}`)})
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "no arguments") {
		t.Fatalf("filtered Handler() error = %v, public=%q ok=%v", err, message, ok)
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
