package tools

import (
	"context"
	"slices"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

func TestRegistryWithIntervalsClientRegistersFullCatalog(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	registry := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{
		Version:          "test",
		TimezoneFallback: "UTC",
		Capability:       safety.NewCapability(safety.ModeFull),
		Toolset:          safety.ToolsetFull,
	})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	wantNames := append(toolcatalog.AthleteScopedToolNames(), toolcatalog.ICUvisorCheckServerVersion, toolcatalog.ICUvisorListAdvancedCapabilities, toolcatalog.ValidateWorkout)
	slices.Sort(wantNames)

	gotNames := make([]string, 0, len(registrar.tools))
	seen := make(map[string]struct{}, len(registrar.tools))
	for _, tool := range registrar.tools {
		gotNames = append(gotNames, tool.Name)
		if _, exists := seen[tool.Name]; exists {
			t.Fatalf("duplicate registered tool %q", tool.Name)
		}
		seen[tool.Name] = struct{}{}
	}
	slices.Sort(gotNames)
	if !slices.Equal(gotNames, wantNames) {
		t.Fatalf("registered tools = %v, want %v", gotNames, wantNames)
	}
}

func TestRegisteredCreateSportSettingsSchemaRemainsClosed(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	registry := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{
		Version:          "test",
		TimezoneFallback: "UTC",
		Capability:       safety.NewCapability(safety.ModeFull),
		Toolset:          safety.ToolsetFull,
	})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for _, tool := range registrar.tools {
		if tool.Name == createSportSettingsName {
			schema, ok := tool.InputSchema.(map[string]any)
			if !ok {
				t.Fatalf("InputSchema type = %T, want map", tool.InputSchema)
			}
			assertCreateSportSettingsSchemaSafe(t, schema, false)
			return
		}
	}
	t.Fatalf("full registry omitted %s", createSportSettingsName)
}

func TestRegisteredToolSchemasDoNotExposeCredentialParameters(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	registry := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{
		Version:          "test",
		TimezoneFallback: "UTC",
		Capability:       safety.NewCapability(safety.ModeFull),
		Toolset:          safety.ToolsetFull,
	})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	for _, tool := range registrar.tools {
		schema, _ := tool.InputSchema.(map[string]any)
		properties, _ := schema["properties"].(map[string]any)
		for _, forbidden := range []string{"api_key", "apikey", "token", "credential", "credential_ref"} {
			if _, ok := properties[forbidden]; ok {
				t.Fatalf("tool %s exposes forbidden credential parameter %q", tool.Name, forbidden)
			}
		}
	}
}

func TestRegisteredAthleteScopedToolsMatchSharedCatalog(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	registry := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{
		Version:          "test",
		TimezoneFallback: "UTC",
		Capability:       safety.NewCapability(safety.ModeFull),
		Toolset:          safety.ToolsetFull,
	})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	gotNames := make([]string, 0, len(registrar.tools))
	for _, tool := range registrar.tools {
		if toolcatalog.IsAthleteScopedTool(tool.Name) {
			gotNames = append(gotNames, tool.Name)
		}
	}
	slices.Sort(gotNames)

	wantNames := toolcatalog.AthleteScopedToolNames()
	if !slices.Equal(gotNames, wantNames) {
		t.Fatalf("registered athlete-scoped tools = %v, want shared catalog %v", gotNames, wantNames)
	}
}
