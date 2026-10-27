package mcp

import (
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

func TestCommunityToolsRespectSafetyAndToolsets(t *testing.T) {
	for _, tc := range []struct {
		name                                                       string
		mode                                                       safety.Mode
		toolset                                                    safety.Toolset
		wantUpdate, wantDelete, wantCompare, wantSearch, wantShare bool
	}{
		{name: "core safe", mode: safety.ModeSafe, toolset: safety.ToolsetCore, wantUpdate: true},
		{name: "core read only", mode: safety.ModeNone, toolset: safety.ToolsetCore},
		{name: "full safe", mode: safety.ModeSafe, toolset: safety.ToolsetFull, wantUpdate: true, wantCompare: true, wantSearch: true, wantShare: true},
		{name: "full delete enabled", mode: safety.ModeFull, toolset: safety.ToolsetFull, wantUpdate: true, wantDelete: true, wantCompare: true, wantSearch: true, wantShare: true},
		{name: "full read only", mode: safety.ModeNone, toolset: safety.ToolsetFull, wantCompare: true, wantSearch: true},
		{name: "compact", mode: safety.ModeFull, toolset: safety.ToolsetCompact},
	} {
		t.Run(tc.name, func(t *testing.T) {
			capability := safety.NewCapability(tc.mode)
			registry := tools.NewRegistryWithOptions(newNoNetworkProtocolClient(t), tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: capability, Toolset: tc.toolset})
			ctx, session, cleanup := connectTestClientWithOptions(t, Options{Registry: registry, Capability: capability, Toolset: tc.toolset})
			defer cleanup()
			listed, err := session.ListTools(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]bool{}
			for _, tool := range listed.Tools {
				seen[tool.Name] = true
			}
			for name, want := range map[string]bool{"update_activity_message": tc.wantUpdate, "delete_activity_message": tc.wantDelete, "compare_power_curves": tc.wantCompare, "search_workout_library": tc.wantSearch, "add_shared_workout_folder": tc.wantShare} {
				if seen[name] != want {
					t.Errorf("%s registered=%t want=%t", name, seen[name], want)
				}
			}
		})
	}
}
