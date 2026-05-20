package tools

import (
	"context"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestToolEffectiveToolsetDefaultsEmptyToFull(t *testing.T) {
	t.Parallel()

	if got := (Tool{}).EffectiveToolset(); got != safety.ToolsetFull {
		t.Fatalf("empty Tool effective toolset = %q, want full", got)
	}
	if got := (Tool{Toolset: safety.Toolset("advanced")}).EffectiveToolset(); got != safety.ToolsetFull {
		t.Fatalf("invalid Tool effective toolset = %q, want full", got)
	}
}

func TestRegisteredToolTierMembership(t *testing.T) {
	t.Parallel()

	registrar := &collectingRegistrar{}
	if err := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull)}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	expected := map[string]safety.Toolset{
		getAthleteProfileName:        safety.ToolsetCore,
		getActivitiesName:            safety.ToolsetCore,
		getActivityDetailsName:       safety.ToolsetCore,
		getActivityIntervalsName:     safety.ToolsetCore,
		getActivitySplitsName:        safety.ToolsetCore,
		getActivityMessagesName:      safety.ToolsetCore,
		getFitnessName:               safety.ToolsetCore,
		getTrainingSummaryName:       safety.ToolsetCore,
		getBestEffortsName:           safety.ToolsetCore,
		getWellnessDataName:          safety.ToolsetCore,
		getEventsName:                safety.ToolsetCore,
		getEventByIDName:             safety.ToolsetCore,
		addOrUpdateEventName:         safety.ToolsetCore,
		updateWellnessName:           safety.ToolsetCore,
		addActivityMessageName:       safety.ToolsetCore,
		linkActivityToEventName:      safety.ToolsetCore,
		listAdvancedCapabilitiesName: safety.ToolsetCore,
		getPowerCurvesName:           safety.ToolsetFull,
		getHRCurvesName:              safety.ToolsetFull,
		getPaceCurvesName:            safety.ToolsetFull,
		getExtendedMetricsName:       safety.ToolsetFull,
		getGearListName:              safety.ToolsetFull,
		getActivityStreamsName:       safety.ToolsetFull,
		getActivityHistogramName:     safety.ToolsetFull,
		getTrainingPlanName:          safety.ToolsetFull,
		applyTrainingPlanName:        safety.ToolsetFull,
		getWorkoutLibraryName:        safety.ToolsetFull,
		getWorkoutsInFolderName:      safety.ToolsetFull,
		createWorkoutName:            safety.ToolsetFull,
		updateWorkoutName:            safety.ToolsetFull,
		deleteWorkoutName:            safety.ToolsetFull,
		updateSportSettingsName:      safety.ToolsetFull,
		deleteSportSettingsName:      safety.ToolsetFull,
		getCustomItemsName:           safety.ToolsetFull,
		getCustomItemByIDName:        safety.ToolsetFull,
		createCustomItemName:         safety.ToolsetFull,
		updateCustomItemName:         safety.ToolsetFull,
		deleteCustomItemName:         safety.ToolsetFull,
		deleteEventName:              safety.ToolsetFull,
		deleteEventsByDateRangeName:  safety.ToolsetFull,
		deleteActivityName:           safety.ToolsetFull,
		deleteGearName:               safety.ToolsetFull,
	}

	seen := make(map[string]safety.Toolset, len(registrar.tools))
	for _, tool := range registrar.tools {
		if _, exists := expected[tool.Name]; !exists {
			t.Fatalf("unexpected registered tool %q with tier %q", tool.Name, tool.EffectiveToolset())
		}
		if _, exists := seen[tool.Name]; exists {
			t.Fatalf("duplicate registered tool %q", tool.Name)
		}
		seen[tool.Name] = tool.EffectiveToolset()
	}
	for name, want := range expected {
		got, exists := seen[name]
		if !exists {
			t.Fatalf("expected tool %q was not registered", name)
		}
		if got != want {
			t.Fatalf("tool %q tier = %q, want %q", name, got, want)
		}
	}
}
