package tools

import (
	"context"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
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
	if err := NewRegistryWithOptions(fullCatalogTierClient{}, RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull)}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	expected := map[string]safety.Toolset{
		getAthleteProfileName:       safety.ToolsetCore,
		getActivitiesName:           safety.ToolsetCore,
		getActivityDetailsName:      safety.ToolsetCore,
		getActivityIntervalsName:    safety.ToolsetCore,
		getActivitySplitsName:       safety.ToolsetCore,
		getActivityMessagesName:     safety.ToolsetCore,
		getFitnessName:              safety.ToolsetCore,
		getTrainingSummaryName:      safety.ToolsetCore,
		getBestEffortsName:          safety.ToolsetCore,
		getWellnessDataName:         safety.ToolsetCore,
		getEventsName:               safety.ToolsetCore,
		getEventByIDName:            safety.ToolsetCore,
		addOrUpdateEventName:        safety.ToolsetCore,
		updateWellnessName:          safety.ToolsetCore,
		addActivityMessageName:      safety.ToolsetCore,
		linkActivityToEventName:     safety.ToolsetCore,
		getPowerCurvesName:          safety.ToolsetFull,
		getExtendedMetricsName:      safety.ToolsetFull,
		getActivityStreamsName:      safety.ToolsetFull,
		getTrainingPlanName:         safety.ToolsetFull,
		applyTrainingPlanName:       safety.ToolsetFull,
		getWorkoutLibraryName:       safety.ToolsetFull,
		getWorkoutsInFolderName:     safety.ToolsetFull,
		createWorkoutName:           safety.ToolsetFull,
		updateWorkoutName:           safety.ToolsetFull,
		deleteWorkoutName:           safety.ToolsetFull,
		updateSportSettingsName:     safety.ToolsetFull,
		deleteSportSettingsName:     safety.ToolsetFull,
		getCustomItemsName:          safety.ToolsetFull,
		getCustomItemByIDName:       safety.ToolsetFull,
		createCustomItemName:        safety.ToolsetFull,
		updateCustomItemName:        safety.ToolsetFull,
		deleteCustomItemName:        safety.ToolsetFull,
		deleteEventName:             safety.ToolsetFull,
		deleteEventsByDateRangeName: safety.ToolsetFull,
		deleteActivityName:          safety.ToolsetFull,
		deleteGearName:              safety.ToolsetFull,
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

type fullCatalogTierClient struct{}

func (fullCatalogTierClient) GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error) {
	return intervals.AthleteWithSportSettings{}, nil
}

func (fullCatalogTierClient) ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	return nil, nil
}

func (fullCatalogTierClient) ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
}

func (fullCatalogTierClient) ListAthleteHRCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
}

func (fullCatalogTierClient) ListAthletePaceCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error) {
	return intervals.DataCurveSet{}, nil
}

func (fullCatalogTierClient) ListWellness(context.Context, intervals.WellnessParams) ([]intervals.Wellness, error) {
	return nil, nil
}

func (fullCatalogTierClient) UpdateWellness(context.Context, intervals.WriteWellnessParams) (intervals.Wellness, error) {
	return intervals.Wellness{}, nil
}

func (fullCatalogTierClient) UpdateSportSettings(context.Context, intervals.WriteSportSettingsParams) (intervals.SportSettings, error) {
	return intervals.SportSettings{}, nil
}

func (fullCatalogTierClient) ListActivities(context.Context, intervals.ListActivitiesParams) ([]intervals.Activity, error) {
	return nil, nil
}

func (fullCatalogTierClient) GetActivity(context.Context, string) (intervals.Activity, error) {
	return intervals.Activity{}, nil
}

func (fullCatalogTierClient) DeleteActivity(context.Context, string) error { return nil }

func (fullCatalogTierClient) GetActivityIntervals(context.Context, string) (intervals.IntervalsDTO, error) {
	return intervals.IntervalsDTO{}, nil
}

func (fullCatalogTierClient) GetActivityStreams(context.Context, intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error) {
	return nil, nil
}

func (fullCatalogTierClient) GetActivityMessages(context.Context, intervals.ActivityMessagesParams) ([]intervals.ActivityMessage, error) {
	return nil, nil
}

func (fullCatalogTierClient) AddActivityMessage(context.Context, intervals.AddActivityMessageParams) (intervals.NewActivityMessage, error) {
	return intervals.NewActivityMessage{}, nil
}

func (fullCatalogTierClient) GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error) {
	return intervals.PowerVsHR{}, nil
}

func (fullCatalogTierClient) ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error) {
	return nil, nil
}

func (fullCatalogTierClient) GetEvent(context.Context, string) (intervals.Event, error) {
	return intervals.Event{}, nil
}

func (fullCatalogTierClient) AddOrUpdateEvent(context.Context, intervals.WriteEventParams) (intervals.Event, error) {
	return intervals.Event{}, nil
}

func (fullCatalogTierClient) DeleteEvent(context.Context, string) error { return nil }

func (fullCatalogTierClient) LinkActivityToEvent(context.Context, intervals.LinkActivityToEventParams) (intervals.Activity, error) {
	return intervals.Activity{}, nil
}

func (fullCatalogTierClient) GetTrainingPlan(context.Context) (intervals.TrainingPlan, error) {
	return intervals.TrainingPlan{}, nil
}

func (fullCatalogTierClient) ListWorkoutFolders(context.Context) ([]intervals.WorkoutFolder, error) {
	return nil, nil
}

func (fullCatalogTierClient) ListLibraryWorkouts(context.Context) ([]intervals.Workout, error) {
	return nil, nil
}

func (fullCatalogTierClient) CreateLibraryWorkout(context.Context, intervals.WriteWorkoutParams) (intervals.Workout, error) {
	return intervals.Workout{}, nil
}

func (fullCatalogTierClient) UpdateLibraryWorkout(context.Context, intervals.WriteWorkoutParams) (intervals.Workout, error) {
	return intervals.Workout{}, nil
}

func (fullCatalogTierClient) DeleteLibraryWorkout(context.Context, string) error { return nil }

func (fullCatalogTierClient) DeleteSportSettings(context.Context, string) error { return nil }

func (fullCatalogTierClient) ListCustomItems(context.Context) ([]intervals.CustomItem, error) {
	return nil, nil
}

func (fullCatalogTierClient) GetCustomItem(context.Context, string) (intervals.CustomItem, error) {
	return intervals.CustomItem{}, nil
}

func (fullCatalogTierClient) CreateCustomItem(context.Context, intervals.WriteCustomItemParams) (intervals.CustomItem, error) {
	return intervals.CustomItem{}, nil
}

func (fullCatalogTierClient) UpdateCustomItem(context.Context, intervals.WriteCustomItemParams) (intervals.CustomItem, error) {
	return intervals.CustomItem{}, nil
}

func (fullCatalogTierClient) DeleteCustomItem(context.Context, string) error { return nil }

func (fullCatalogTierClient) GetGear(context.Context, string) (intervals.Gear, error) {
	return intervals.Gear{}, nil
}

func (fullCatalogTierClient) DeleteGear(context.Context, string) error { return nil }
