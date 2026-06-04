package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestCheckServerVersionRegisteredAsCoreReadOnlyMetaTool(t *testing.T) {
	registrar := &collectingRegistrar{}
	if err := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{Version: "v0.6.0", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeSafe), Toolset: safety.ToolsetCore}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tool := findTool(t, registrar.tools, checkServerVersionName)
	if tool.Requirement.effective() != RequirementRead {
		t.Fatalf("requirement = %q, want read", tool.Requirement.effective())
	}
	if tool.EffectiveToolset() != safety.ToolsetCore {
		t.Fatalf("toolset = %q, want core", tool.EffectiveToolset())
	}
	if !strings.Contains(tool.Description, "description_server_version=v0.6.0") || !strings.Contains(tool.Description, "description_catalog_fingerprint=") || !strings.Contains(tool.Description, "description_toolset=core") || !strings.Contains(tool.Description, "description_delete_mode=safe") {
		t.Fatalf("description missing visible diagnostic fields: %q", tool.Description)
	}
	if schema, ok := tool.InputSchema.(map[string]any); !ok || schema["additionalProperties"] != false {
		t.Fatalf("input schema = %#v, want no-argument closed object", tool.InputSchema)
	}
	for _, descriptor := range Catalog() {
		if descriptor.Name == checkServerVersionName {
			if descriptor.Group != "meta" || descriptor.Tier != string(safety.ToolsetCore) || descriptor.Safety != string(RequirementRead) {
				t.Fatalf("descriptor = %#v, want meta/core/read", descriptor)
			}
			return
		}
	}
	t.Fatalf("Catalog() missing %s", checkServerVersionName)
}

func TestCheckServerVersionOutputShapeAndNoLeakage(t *testing.T) {
	response.SetRuntimeCatalogMetadata("v0.6.1", "live-catalog-hash")
	t.Cleanup(func() { response.SetRuntimeCatalogMetadata("dev", "dev-catalog-hash") })

	tool, err := newCheckServerVersionTool("v0.6.1", []Tool{{Name: "catalog_alpha", Description: "Alpha.", InputSchema: noArgsSchema(), OutputSchema: genericOutputSchema("alpha output"), Requirement: RequirementRead, Toolset: safety.ToolsetCore}}, safety.ModeSafe, safety.ToolsetCore)
	if err != nil {
		t.Fatalf("newCheckServerVersionTool() error = %v", err)
	}
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}

	payload := checkServerVersionResult(t, result)
	if payload.ServerVersion != "v0.6.1" || payload.CatalogHash != "live-catalog-hash" {
		t.Fatalf("runtime fields = version %q hash %q, want live metadata", payload.ServerVersion, payload.CatalogHash)
	}
	if payload.DescriptionServerVersion != "v0.6.1" || payload.DescriptionCatalogFingerprint == "" || payload.Toolset != "core" || payload.DeleteMode != "safe" {
		t.Fatalf("description fields = %#v", payload)
	}
	if payload.Status != checkServerVersionStatus || !strings.Contains(payload.Action, "Compare the visible tool description fields") || !strings.Contains(payload.Action, "start a new conversation") {
		t.Fatalf("status/action = %q/%q", payload.Status, payload.Action)
	}
	if !payload.Meta.NoNetwork || payload.Meta.Source == "" || !strings.Contains(payload.Meta.FingerprintScope, "dynamic coach") {
		t.Fatalf("_meta = %#v", payload.Meta)
	}
	text := resultText(t, result)
	for _, forbidden := range []string{"i12345", "athlete_id", "api_key", "credential", "/Users/", "ICUVISOR_API_KEY"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("diagnostic output leaked forbidden %q: %s", forbidden, text)
		}
	}
}

func TestCheckServerVersionRejectsArguments(t *testing.T) {
	t.Parallel()

	tool, err := newCheckServerVersionTool("test", nil, safety.ModeSafe, safety.ToolsetCore)
	if err != nil {
		t.Fatalf("newCheckServerVersionTool() error = %v", err)
	}
	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"athlete_id":"i12345"}`)})
	if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "no arguments") {
		t.Fatalf("Handler() error = %v, public=%q ok=%v", err, message, ok)
	}
}

func TestDescriptionCatalogFingerprintCapturesSameVersionDrift(t *testing.T) {
	t.Parallel()

	base := []Tool{{Name: "catalog_alpha", Description: "Alpha.", InputSchema: noArgsSchema(), OutputSchema: genericOutputSchema("alpha output"), Requirement: RequirementRead, Toolset: safety.ToolsetCore}}
	baseTool, err := newCheckServerVersionTool("v0.6.1", base, safety.ModeSafe, safety.ToolsetCore)
	if err != nil {
		t.Fatalf("newCheckServerVersionTool(base) error = %v", err)
	}
	changedDescriptionTool, err := newCheckServerVersionTool("v0.6.1", checkVersionMutateTool(base, 0, func(tool *Tool) { tool.Description = "Alpha changed." }), safety.ModeSafe, safety.ToolsetCore)
	if err != nil {
		t.Fatalf("newCheckServerVersionTool(description drift) error = %v", err)
	}
	changedSchemaTool, err := newCheckServerVersionTool("v0.6.1", checkVersionMutateTool(base, 0, func(tool *Tool) { tool.InputSchema = checkVersionSchemaWithArgument("renamed", "alpha input") }), safety.ModeSafe, safety.ToolsetCore)
	if err != nil {
		t.Fatalf("newCheckServerVersionTool(schema drift) error = %v", err)
	}

	baseFingerprint := mustDescriptionFingerprint(t, baseTool)
	if baseFingerprint == mustDescriptionFingerprint(t, changedDescriptionTool) {
		t.Fatalf("description fingerprint did not change after same-version description drift: %s", baseFingerprint)
	}
	if baseFingerprint == mustDescriptionFingerprint(t, changedSchemaTool) {
		t.Fatalf("description fingerprint did not change after same-version schema drift: %s", baseFingerprint)
	}
	rebuiltTool, err := newCheckServerVersionTool("v0.6.1", base, safety.ModeSafe, safety.ToolsetCore)
	if err != nil {
		t.Fatalf("newCheckServerVersionTool(rebuilt) error = %v", err)
	}
	if got := mustDescriptionFingerprint(t, rebuiltTool); got != baseFingerprint {
		t.Fatalf("rebuilt fingerprint = %s, want stable %s", got, baseFingerprint)
	}
}

func checkVersionMutateTool(in []Tool, index int, mutate func(*Tool)) []Tool {
	out := make([]Tool, len(in))
	copy(out, in)
	mutate(&out[index])
	return out
}

func checkVersionSchemaWithArgument(name, description string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			name: map[string]any{
				"type":        "string",
				"description": description,
			},
		},
	}
}

func checkServerVersionResult(t *testing.T, result Result) checkServerVersionResponse {
	t.Helper()
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("marshal StructuredContent: %v", err)
	}
	var payload checkServerVersionResponse
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode StructuredContent: %v", err)
	}
	return payload
}

func mustDescriptionFingerprint(t *testing.T, tool Tool) string {
	t.Helper()
	payload, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	fingerprint := checkServerVersionResult(t, payload).DescriptionCatalogFingerprint
	if fingerprint == "" {
		t.Fatal("description catalog fingerprint is empty")
	}
	return fingerprint
}
