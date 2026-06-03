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
		getAthleteProfileName:           safety.ToolsetCore,
		getActivitiesName:               safety.ToolsetCore,
		getActivityDetailsName:          safety.ToolsetCore,
		getActivityIntervalsName:        safety.ToolsetCore,
		getActivitySplitsName:           safety.ToolsetCore,
		getActivityMessagesName:         safety.ToolsetCore,
		getFitnessName:                  safety.ToolsetCore,
		getTodayName:                    safety.ToolsetCore,
		resolveCalendarDatesName:        safety.ToolsetCore,
		getTrainingSummaryName:          safety.ToolsetCore,
		getBestEffortsName:              safety.ToolsetCore,
		getWellnessDataName:             safety.ToolsetCore,
		getEventsName:                   safety.ToolsetCore,
		getEventByIDName:                safety.ToolsetCore,
		addOrUpdateEventName:            safety.ToolsetCore,
		updateWellnessName:              safety.ToolsetCore,
		addActivityMessageName:          safety.ToolsetCore,
		linkActivityToEventName:         safety.ToolsetCore,
		updateActivityName:              safety.ToolsetCore,
		setActivityIntervalsName:        safety.ToolsetFull,
		listAdvancedCapabilitiesName:    safety.ToolsetCore,
		getPowerCurvesName:              safety.ToolsetFull,
		analyzeTrendName:                safety.ToolsetCore,
		analyzeDistributionName:         safety.ToolsetFull,
		analyzeCorrelationName:          safety.ToolsetFull,
		analyzeEffortsDeltaName:         safety.ToolsetFull,
		getFitnessProjectionName:        safety.ToolsetFull,
		getHRCurvesName:                 safety.ToolsetFull,
		getPaceCurvesName:               safety.ToolsetFull,
		getExtendedMetricsName:          safety.ToolsetFull,
		getGearListName:                 safety.ToolsetFull,
		getActivityStreamsName:          safety.ToolsetFull,
		getActivityHistogramName:        safety.ToolsetFull,
		computeActivitySegmentStatsName: safety.ToolsetFull,
		computeZoneTimeName:             safety.ToolsetCore,
		computeLoadBalanceName:          safety.ToolsetFull,
		computeBaselineName:             safety.ToolsetCore,
		computeComplianceRateName:       safety.ToolsetFull,
		getPlanningContextName:          safety.ToolsetFull,
		getTrainingPlanName:             safety.ToolsetFull,
		applyTrainingPlanName:           safety.ToolsetFull,
		getWorkoutLibraryName:           safety.ToolsetFull,
		getWorkoutsInFolderName:         safety.ToolsetFull,
		createWorkoutName:               safety.ToolsetFull,
		updateWorkoutName:               safety.ToolsetFull,
		deleteWorkoutName:               safety.ToolsetFull,
		updateSportSettingsName:         safety.ToolsetFull,
		deleteSportSettingsName:         safety.ToolsetFull,
		getCustomItemsName:              safety.ToolsetFull,
		getCustomItemByIDName:           safety.ToolsetFull,
		createCustomItemName:            safety.ToolsetFull,
		updateCustomItemName:            safety.ToolsetFull,
		deleteCustomItemName:            safety.ToolsetFull,
		deleteEventName:                 safety.ToolsetFull,
		deleteEventsByDateRangeName:     safety.ToolsetFull,
		deleteActivityName:              safety.ToolsetFull,
		deleteGearName:                  safety.ToolsetFull,
		validateWorkoutName:             safety.ToolsetCore,
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

func TestNonCandidateAnalyzerFamilyRemainsFullToolset(t *testing.T) {
	t.Parallel()

	tiers := registeredToolsetsByName(t)
	candidates := analyzerCorePromotionCandidateSet()
	for _, name := range analyzerFamilyCatalogNames() {
		got, exists := tiers[name]
		if !exists {
			t.Fatalf("analyzer-family tool %q was not registered", name)
		}
		if _, candidate := candidates[name]; candidate {
			continue
		}
		if got != safety.ToolsetFull {
			t.Fatalf("non-candidate analyzer-family tool %q tier = %q, want full", name, got)
		}
	}
}

func TestAnalyzerCorePromotionCandidatesAreBenchmarkGated(t *testing.T) {
	t.Parallel()

	// docs/kr5-benchmark.md TP-098 evidence currently gates only these analyzer-family tools.
	candidates := analyzerCorePromotionCandidateSet()
	wantCandidates := []string{analyzeTrendName, computeZoneTimeName, computeBaselineName}
	if len(candidates) != len(wantCandidates) {
		t.Fatalf("promotion candidates = %#v, want exactly %v", candidates, wantCandidates)
	}
	family := make(map[string]struct{}, len(analyzerFamilyCatalogNames()))
	for _, name := range analyzerFamilyCatalogNames() {
		family[name] = struct{}{}
	}
	for _, name := range wantCandidates {
		if _, ok := candidates[name]; !ok {
			t.Fatalf("promotion candidates missing %q", name)
		}
	}
	for name := range candidates {
		if _, ok := family[name]; !ok {
			t.Fatalf("promotion candidate %q is not in analyzer-family catalog names", name)
		}
	}
	tiers := registeredToolsetsByName(t)
	for _, name := range analyzerFamilyCatalogNames() {
		got, exists := tiers[name]
		if !exists {
			t.Fatalf("analyzer-family tool %q was not registered", name)
		}
		if _, candidate := candidates[name]; candidate {
			if got != safety.ToolsetCore {
				t.Fatalf("benchmark-gated candidate %q tier = %q, want core", name, got)
			}
			continue
		}
		if got != safety.ToolsetFull {
			t.Fatalf("non-candidate analyzer-family tool %q tier = %q, want full", name, got)
		}
	}
}

func registeredToolsetsByName(t *testing.T) map[string]safety.Toolset {
	t.Helper()

	registrar := &collectingRegistrar{}
	if err := NewRegistryWithOptions(newNoNetworkIntervalsClient(t), RegistryOptions{Version: "test", TimezoneFallback: "UTC", Capability: safety.NewCapability(safety.ModeFull)}).Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tiers := make(map[string]safety.Toolset, len(registrar.tools))
	for _, tool := range registrar.tools {
		tiers[tool.Name] = tool.EffectiveToolset()
	}
	return tiers
}

func analyzerCorePromotionCandidateSet() map[string]struct{} {
	return map[string]struct{}{
		analyzeTrendName:    {},
		computeZoneTimeName: {},
		computeBaselineName: {},
	}
}
