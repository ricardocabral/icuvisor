package safety_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
	"github.com/ricardocabral/icuvisor/internal/tools"
)

type catalogToolSpec struct {
	name        string
	requirement tools.Requirement
}

var v03ToolCatalog = []catalogToolSpec{
	{name: "add_activity_message", requirement: tools.RequirementWrite},
	{name: "add_or_update_event", requirement: tools.RequirementWrite},
	{name: "add_unavailable_date_range", requirement: tools.RequirementWrite},
	{name: "analyze_correlation", requirement: tools.RequirementRead},
	{name: "analyze_distribution", requirement: tools.RequirementRead},
	{name: "analyze_efforts_delta", requirement: tools.RequirementRead},
	{name: "analyze_trend", requirement: tools.RequirementRead},
	{name: "apply_training_plan", requirement: tools.RequirementWrite},
	{name: "compute_activity_segment_stats", requirement: tools.RequirementRead},
	{name: "compute_baseline", requirement: tools.RequirementRead},
	{name: "compute_compliance_rate", requirement: tools.RequirementRead},
	{name: "compute_load_balance", requirement: tools.RequirementRead},
	{name: "compute_zone_time", requirement: tools.RequirementRead},
	{name: "create_custom_item", requirement: tools.RequirementWrite},
	{name: "create_workout", requirement: tools.RequirementWrite},
	{name: "delete_activity", requirement: tools.RequirementDelete},
	{name: "delete_custom_item", requirement: tools.RequirementDelete},
	{name: "delete_event", requirement: tools.RequirementDelete},
	{name: "delete_events_by_date_range", requirement: tools.RequirementDelete},
	{name: "delete_gear", requirement: tools.RequirementDelete},
	{name: "delete_sport_settings", requirement: tools.RequirementDelete},
	{name: "delete_workout", requirement: tools.RequirementDelete},
	{name: "get_activities", requirement: tools.RequirementRead},
	{name: "get_activities_around", requirement: tools.RequirementRead},
	{name: "get_activity_details", requirement: tools.RequirementRead},
	{name: "get_activity_histogram", requirement: tools.RequirementRead},
	{name: "get_activity_intervals", requirement: tools.RequirementRead},
	{name: "get_activity_messages", requirement: tools.RequirementRead},
	{name: "get_activity_splits", requirement: tools.RequirementRead},
	{name: "get_activity_streams", requirement: tools.RequirementRead},
	{name: "get_annual_training_plan", requirement: tools.RequirementRead},
	{name: "get_athlete_profile", requirement: tools.RequirementRead},
	{name: "get_best_efforts", requirement: tools.RequirementRead},
	{name: "get_custom_item_by_id", requirement: tools.RequirementRead},
	{name: "get_custom_items", requirement: tools.RequirementRead},
	{name: "get_event_by_id", requirement: tools.RequirementRead},
	{name: "get_events", requirement: tools.RequirementRead},
	{name: "get_extended_metrics", requirement: tools.RequirementRead},
	{name: "get_fitness", requirement: tools.RequirementRead},
	{name: "get_fitness_projection", requirement: tools.RequirementRead},
	{name: "get_planning_context", requirement: tools.RequirementRead},
	{name: "get_today", requirement: tools.RequirementRead},
	{name: "get_gear_list", requirement: tools.RequirementRead},
	{name: "get_hr_curves", requirement: tools.RequirementRead},
	{name: "get_pace_curves", requirement: tools.RequirementRead},
	{name: "get_performance_potential", requirement: tools.RequirementRead},
	{name: "get_power_curves", requirement: tools.RequirementRead},
	{name: "resolve_calendar_dates", requirement: tools.RequirementRead},
	{name: "get_training_plan", requirement: tools.RequirementRead},
	{name: "get_training_summary", requirement: tools.RequirementRead},
	{name: "get_wellness_data", requirement: tools.RequirementRead},
	{name: "get_workout_library", requirement: tools.RequirementRead},
	{name: "get_workouts_in_folder", requirement: tools.RequirementRead},
	{name: "icuvisor_check_server_version", requirement: tools.RequirementRead},
	{name: "icuvisor_list_advanced_capabilities", requirement: tools.RequirementRead},
	{name: "link_activity_to_event", requirement: tools.RequirementWrite},
	{name: "set_activity_intervals", requirement: tools.RequirementDelete},
	{name: "update_activity", requirement: tools.RequirementWrite},
	{name: "update_custom_item", requirement: tools.RequirementWrite},
	{name: "update_sport_settings", requirement: tools.RequirementWrite},
	{name: "update_wellness", requirement: tools.RequirementWrite},
	{name: "update_workout", requirement: tools.RequirementWrite},
	{name: "validate_workout", requirement: tools.RequirementRead},
}

type gatedCatalogRegistrar struct {
	capability safety.Capability
	tools      map[string]tools.Tool
}

func (r *gatedCatalogRegistrar) AddTool(tool tools.Tool) error {
	if tool.RequiresDelete() && !r.capability.CanDelete() {
		return nil
	}
	if tool.RequiresWrite() && !r.capability.CanWrite() {
		return nil
	}
	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("duplicate tool %q", tool.Name)
	}
	r.tools[tool.Name] = tool
	return nil
}

func TestAdversarialStaticCatalogMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode safety.Mode
	}{
		{mode: safety.ModeSafe},
		{mode: safety.ModeFull},
		{mode: safety.ModeNone},
	}
	for _, tc := range tests {
		t.Run(tc.mode.String(), func(t *testing.T) {
			t.Parallel()

			registered := registerCatalogForMode(t, tc.mode)
			for _, spec := range v03ToolCatalog {
				_, gotRegistered := registered[spec.name]
				wantRegistered := wantToolRegistered(tc.mode, spec.requirement)
				if gotRegistered != wantRegistered {
					t.Fatalf("%s registration in mode %s = %v, want %v", spec.name, tc.mode, gotRegistered, wantRegistered)
				}
			}
			if len(registered) != expectedRegisteredCount(tc.mode) {
				t.Fatalf("registered tool count in mode %s = %d, want %d; tools=%v", tc.mode, len(registered), expectedRegisteredCount(tc.mode), sortedToolNames(registered))
			}
		})
	}
}

func TestAdversarialRegisteredSchemasDoNotExposeConfirm(t *testing.T) {
	t.Parallel()

	for _, mode := range []safety.Mode{safety.ModeSafe, safety.ModeFull, safety.ModeNone} {
		t.Run(mode.String(), func(t *testing.T) {
			t.Parallel()

			registered := registerCatalogForMode(t, mode)
			for name, tool := range registered {
				if schemaContainsConfirmArgument(tool.InputSchema) {
					t.Fatalf("%s input schema in mode %s contains a confirm argument: %#v", name, mode, tool.InputSchema)
				}
			}
		})
	}
}

func TestAdversarialDeleteEventsByDateRangeCapAppliesInFullMode(t *testing.T) {
	t.Parallel()

	registered := registerCatalogForMode(t, safety.ModeFull)
	tool, ok := registered["delete_events_by_date_range"]
	if !ok {
		t.Fatal("delete_events_by_date_range is not registered in full mode")
	}

	_, err := tool.Handler(context.Background(), tools.Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-06-01"}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want range cap rejection")
	}
}

func registerCatalogForMode(t *testing.T, mode safety.Mode) map[string]tools.Tool {
	t.Helper()

	client := newCatalogClient(t)
	registrar := &gatedCatalogRegistrar{capability: safety.NewCapability(mode), tools: make(map[string]tools.Tool)}
	registry := tools.NewRegistryWithOptions(client, tools.RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(mode)})
	if err := registry.Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	return registrar.tools
}

func newCatalogClient(t *testing.T) *intervals.Client {
	t.Helper()

	client, err := intervals.NewClient(intervals.Options{Config: config.Config{APIKey: "x", AthleteID: "i12345", APIBaseURL: "https://example.invalid", HTTPTimeout: time.Second}, Version: "test"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return client
}

func wantToolRegistered(mode safety.Mode, requirement tools.Requirement) bool {
	switch requirement {
	case tools.RequirementDelete:
		return mode == safety.ModeFull
	case tools.RequirementWrite:
		return mode == safety.ModeSafe || mode == safety.ModeFull
	default:
		return true
	}
}

func expectedRegisteredCount(mode safety.Mode) int {
	count := 0
	for _, spec := range v03ToolCatalog {
		if wantToolRegistered(mode, spec.requirement) {
			count++
		}
	}
	return count
}

func schemaContainsConfirmArgument(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if strings.EqualFold(key, "confirm") || schemaContainsConfirmArgument(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if schemaContainsConfirmArgument(child) {
				return true
			}
		}
	case []string:
		for _, item := range typed {
			if strings.EqualFold(item, "confirm") {
				return true
			}
		}
	}
	return false
}

func sortedToolNames(catalog map[string]tools.Tool) []string {
	names := make([]string, 0, len(catalog))
	for name := range catalog {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
