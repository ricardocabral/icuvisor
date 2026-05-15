package tools

import (
	"context"
	"slices"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/safety"
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

	wantNames := []string{
		addActivityMessageName,
		addOrUpdateEventName,
		applyTrainingPlanName,
		createCustomItemName,
		createWorkoutName,
		deleteActivityName,
		deleteCustomItemName,
		deleteEventName,
		deleteEventsByDateRangeName,
		deleteGearName,
		deleteSportSettingsName,
		deleteWorkoutName,
		getActivitiesName,
		getActivityDetailsName,
		getActivityIntervalsName,
		getActivityMessagesName,
		getActivitySplitsName,
		getActivityStreamsName,
		getAthleteProfileName,
		getBestEffortsName,
		getCustomItemByIDName,
		getCustomItemsName,
		getEventByIDName,
		getEventsName,
		getExtendedMetricsName,
		getFitnessName,
		getPowerCurvesName,
		getTrainingPlanName,
		getTrainingSummaryName,
		getWellnessDataName,
		getWorkoutLibraryName,
		getWorkoutsInFolderName,
		linkActivityToEventName,
		listAdvancedCapabilitiesName,
		updateCustomItemName,
		updateSportSettingsName,
		updateWellnessName,
		updateWorkoutName,
	}
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
