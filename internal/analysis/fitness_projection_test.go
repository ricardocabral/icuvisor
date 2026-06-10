package analysis

import (
	"reflect"
	"testing"
)

func TestFitnessProjectionWeeklyPlanTargetsFillPartialWeekAndFallback(t *testing.T) {
	t.Parallel()

	input := FitnessProjectionInput{
		StartDate:           "2026-06-03",
		StartCTL:            50,
		StartATL:            55,
		StartTSB:            -5,
		HorizonDays:         7,
		WeeklyRampPct:       0,
		RecoveryWeekCadence: 0,
	}
	setWeeklyPlanTargets(t, &input, weeklyTargetFixture{weekStartDate: "2026-06-01", trainingLoad: 700})

	projection, err := ProjectFitness(input)
	if err != nil {
		t.Fatalf("ProjectFitness() error = %v", err)
	}

	assertProjectionPoint(t, projection, "2026-06-03", "current_fitness", 0)
	assertProjectionPoint(t, projection, "2026-06-04", "weekly_plan_targets", 100)
	assertProjectionPoint(t, projection, "2026-06-07", "weekly_plan_targets", 100)
	assertProjectionPoint(t, projection, "2026-06-08", "modeled_ramp", 50)
}

func TestFitnessProjectionExplicitDailyLoadsOverrideWeeklyTargetsWithoutRedistribution(t *testing.T) {
	t.Parallel()

	input := FitnessProjectionInput{
		StartDate:           "2026-06-03",
		StartCTL:            50,
		StartATL:            55,
		StartTSB:            -5,
		HorizonDays:         4,
		WeeklyRampPct:       0,
		RecoveryWeekCadence: 0,
		PlannedDailyLoads: []FitnessProjectionPlannedLoad{
			{Date: "2026-06-05", TrainingLoad: 42},
		},
	}
	setWeeklyPlanTargets(t, &input, weeklyTargetFixture{weekStartDate: "2026-06-01", trainingLoad: 700})

	projection, err := ProjectFitness(input)
	if err != nil {
		t.Fatalf("ProjectFitness() error = %v", err)
	}

	assertProjectionPoint(t, projection, "2026-06-04", "weekly_plan_targets", 100)
	assertProjectionPoint(t, projection, "2026-06-05", "planned_daily_loads", 42)
	assertProjectionPoint(t, projection, "2026-06-06", "weekly_plan_targets", 100)
}

func TestFitnessProjectionUncoveredDatesRetainRecoveryWeekSource(t *testing.T) {
	t.Parallel()

	input := FitnessProjectionInput{
		StartDate:           "2026-06-01",
		StartCTL:            100,
		StartATL:            100,
		StartTSB:            0,
		HorizonDays:         14,
		WeeklyRampPct:       0,
		RecoveryWeekCadence: 2,
		RecoveryWeekLoadPct: 50,
	}
	setWeeklyPlanTargets(t, &input, weeklyTargetFixture{weekStartDate: "2026-06-01", trainingLoad: 700})

	projection, err := ProjectFitness(input)
	if err != nil {
		t.Fatalf("ProjectFitness() error = %v", err)
	}

	assertProjectionPoint(t, projection, "2026-06-02", "weekly_plan_targets", 100)
	assertProjectionPoint(t, projection, "2026-06-08", "modeled_recovery_week", 50)
}

type weeklyTargetFixture struct {
	weekStartDate string
	trainingLoad  float64
}

func setWeeklyPlanTargets(t *testing.T, input *FitnessProjectionInput, targets ...weeklyTargetFixture) {
	t.Helper()

	field := reflect.ValueOf(input).Elem().FieldByName("WeeklyPlanTargets")
	if !field.IsValid() {
		t.Fatal("FitnessProjectionInput missing WeeklyPlanTargets field")
	}
	if !field.CanSet() || field.Kind() != reflect.Slice {
		t.Fatalf("WeeklyPlanTargets field = %s, want settable slice", field.Kind())
	}
	slice := reflect.MakeSlice(field.Type(), len(targets), len(targets))
	for i, target := range targets {
		entry := slice.Index(i)
		weekStart := entry.FieldByName("WeekStartDate")
		if !weekStart.IsValid() || !weekStart.CanSet() || weekStart.Kind() != reflect.String {
			t.Fatal("weekly target entry missing settable string WeekStartDate field")
		}
		weekStart.SetString(target.weekStartDate)
		trainingLoad := entry.FieldByName("TrainingLoad")
		if !trainingLoad.IsValid() || !trainingLoad.CanSet() || trainingLoad.Kind() != reflect.Float64 {
			t.Fatal("weekly target entry missing settable float64 TrainingLoad field")
		}
		trainingLoad.SetFloat(target.trainingLoad)
	}
	field.Set(slice)
}

func assertProjectionPoint(t *testing.T, projection FitnessProjectionResult, date string, source string, load float64) {
	t.Helper()

	for _, point := range projection.Points {
		if point.Date != date {
			continue
		}
		if point.TrainingLoadSource != source || point.TrainingLoad != load {
			t.Fatalf("point %s = source %q load %.3f, want source %q load %.3f", date, point.TrainingLoadSource, point.TrainingLoad, source, load)
		}
		return
	}
	t.Fatalf("point %s not found in %#v", date, projection.Points)
}
