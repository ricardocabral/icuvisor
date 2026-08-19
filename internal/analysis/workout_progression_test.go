package analysis

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAnalyzeWorkoutProgressionScalarAndRangeTargets(t *testing.T) {
	scalar := 100.0
	rangeMin, rangeMax := 100.0, 110.0
	makeActivity := func(id string, target WorkoutTarget, observed float64) WorkoutProgressionActivity {
		return WorkoutProgressionActivity{
			ID: id, Sport: "ride", DurationSeconds: wpFloatPtr(600), DurationSource: "moving_time", Feel: wpFloatPtr(3), RPE: wpFloatPtr(6),
			Prescription: &WorkoutPrescription{Source: "activity_workout_doc", TotalCount: 1, Intervals: []PrescriptionInterval{{DurationSeconds: wpFloatPtr(600), Target: target}}},
			Completed:    &WorkoutCompleted{IntervalSource: IntervalSourceStructuredWorkout, Intervals: []CompletedInterval{{DurationSeconds: wpFloatPtr(600), Kind: "work", ObservedPowerWatts: wpFloatPtr(observed)}}},
		}
	}

	tests := []struct {
		name       string
		target     WorkoutTarget
		observed   float64
		wantAbs    float64
		wantRel    *float64
		wantWithin int
	}{
		{name: "scalar uses direct value", target: WorkoutTarget{Kind: "power", Unit: "W", Value: &scalar}, observed: 110, wantAbs: 10, wantRel: wpFloatPtr(10), wantWithin: 1},
		{name: "range uses nearest bound", target: WorkoutTarget{Kind: "power", Unit: "W", Min: &rangeMin, Max: &rangeMax}, observed: 120, wantAbs: 10, wantRel: wpFloatPtr(9.090909), wantWithin: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{makeActivity("a1", tc.target, tc.observed)}, false)
			family := got.Rows[0].Adherence.Families["power_watts"]
			if family.MeanAbsoluteError != tc.wantAbs || family.WithinToleranceCount != tc.wantWithin {
				t.Fatalf("family = %#v, want absolute %v/within %d", family, tc.wantAbs, tc.wantWithin)
			}
			if tc.wantRel == nil && family.MeanRelativeErrorPercent != nil {
				t.Fatalf("relative error = %v, want omitted", *family.MeanRelativeErrorPercent)
			}
			if tc.wantRel != nil && (family.MeanRelativeErrorPercent == nil || *family.MeanRelativeErrorPercent != *tc.wantRel) {
				t.Fatalf("relative error = %v, want %v", family.MeanRelativeErrorPercent, *tc.wantRel)
			}
		})
	}
}

func TestAnalyzeWorkoutProgressionZeroDenominatorsRemainIndependent(t *testing.T) {
	zero := 0.0
	got := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{
		{ID: "a1", Sport: "run", DurationSeconds: &zero, RecoverySeconds: &zero, Feel: wpFloatPtr(2), RPE: wpFloatPtr(3), Prescription: &WorkoutPrescription{TotalCount: 1, Intervals: []PrescriptionInterval{{Target: WorkoutTarget{Kind: "power", Unit: "W", Value: &zero}}}}, Completed: &WorkoutCompleted{IntervalSource: IntervalSourceStructuredWorkout, Intervals: []CompletedInterval{{Kind: "work", ObservedPowerWatts: wpFloatPtr(5)}}}},
		{ID: "a2", Sport: "run", DurationSeconds: wpFloatPtr(10), RecoverySeconds: wpFloatPtr(5), Feel: wpFloatPtr(3), RPE: wpFloatPtr(4), Prescription: &WorkoutPrescription{TotalCount: 1, Intervals: []PrescriptionInterval{{Target: WorkoutTarget{Kind: "power", Unit: "W", Value: &zero}}}}, Completed: &WorkoutCompleted{IntervalSource: IntervalSourceStructuredWorkout, Intervals: []CompletedInterval{{Kind: "work", ObservedPowerWatts: wpFloatPtr(5)}}}},
	}, false)
	family := got.Rows[0].Adherence.Families["power_watts"]
	if family.MeanAbsoluteError != 5 || family.MeanRelativeErrorPercent != nil || !testContainsReason(family.Reasons, "zero_denominator") {
		t.Fatalf("zero family = %#v, want absolute-only zero-denominator evidence", family)
	}
	delta := got.Deltas[0]
	if delta.DurationSecondsDelta == nil || *delta.DurationSecondsDelta != 10 || delta.FieldStatus["duration_percent"] != "zero_denominator" || delta.DurationPercentDelta != nil {
		t.Fatalf("duration delta = %#v, want absolute value and omitted percent", delta)
	}
	if delta.RecoverySecondsDelta == nil || *delta.RecoverySecondsDelta != 5 || delta.FieldStatus["recovery_percent"] != "zero_denominator" {
		t.Fatalf("recovery delta = %#v, want absolute value and zero denominator percent", delta)
	}
}

func TestAnalyzeWorkoutProgressionStabilityRequiresTwoSamplesPerHalf(t *testing.T) {
	values := []float64{100, 100, 110, 110}
	intervals := make([]CompletedInterval, len(values))
	for i, value := range values {
		intervals[i] = CompletedInterval{Kind: "work", ObservedPowerWatts: wpFloatPtr(value)}
	}
	got := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{{ID: "a1", Sport: "ride", Completed: &WorkoutCompleted{IntervalSource: IntervalSourceStructuredWorkout, Intervals: intervals}, DurationSeconds: wpFloatPtr(600), Feel: wpFloatPtr(3), RPE: wpFloatPtr(5)}}, false)
	metric := got.Rows[0].Stability.Metrics["power_watts"]
	if metric.DriftPercent == nil || *metric.DriftPercent != 10 || metric.SampleCount != 4 {
		t.Fatalf("stability = %#v, want 10%% drift over four samples", metric)
	}
}

func TestAnalyzeWorkoutProgressionPairMatrix(t *testing.T) {
	prescription := func(seconds float64) *WorkoutPrescription {
		return &WorkoutPrescription{TotalCount: 1, Intervals: []PrescriptionInterval{{DurationSeconds: wpFloatPtr(seconds)}}}
	}
	base := func(id, sport string, p *WorkoutPrescription, count int) WorkoutProgressionActivity {
		completed := make([]CompletedInterval, count)
		for i := range completed {
			completed[i] = CompletedInterval{Kind: "work", DurationSeconds: wpFloatPtr(300)}
		}
		return WorkoutProgressionActivity{ID: id, Sport: sport, Prescription: p, Completed: &WorkoutCompleted{IntervalSource: IntervalSourceStructuredWorkout, Intervals: completed}, DurationSeconds: wpFloatPtr(300), Feel: wpFloatPtr(3), RPE: wpFloatPtr(5)}
	}

	missingCompleted := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{base("a1", "run", prescription(300), 1), base("a2", "run", prescription(300), 0)}, false)
	if missingCompleted.Comparison.ComparablePairCount != 1 || missingCompleted.Deltas[0].Status != "insufficient_evidence" || !testContainsReason(missingCompleted.Deltas[0].Reasons, "missing_intervals") {
		t.Fatalf("missing completed pair = %#v, want counted insufficient pair", missingCompleted.Deltas[0])
	}
	mismatch := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{base("a1", "run", prescription(300), 1), base("a2", "run", prescription(600), 1)}, false)
	if mismatch.Comparison.ComparablePairCount != 0 || mismatch.Deltas[0].Status != "not_comparable" || mismatch.Deltas[0].FieldStatus["duration_seconds"] != "not_comparable" {
		t.Fatalf("signature mismatch = %#v, want not-comparable pair", mismatch.Deltas[0])
	}
	mixedSport := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{base("a1", "run", prescription(300), 1), base("a2", "ride", prescription(300), 1)}, false)
	if mixedSport.Deltas[0].Status != "not_comparable" || mixedSport.Comparison.ComparablePairCount != 0 {
		t.Fatalf("mixed sport = %#v, want excluded not-comparable pair", mixedSport.Deltas[0])
	}
}

func TestAnalyzeWorkoutProgressionEnvelopeOmitsRawAuditByDefault(t *testing.T) {
	got := AnalyzeWorkoutProgression([]WorkoutProgressionActivity{{ID: "a1", Sport: "run", InitialReasons: []string{"missing_source"}}}, false)
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "audit") || strings.Contains(string(encoded), "stream") {
		t.Fatalf("terse envelope leaked full evidence: %s", encoded)
	}
}

func testContainsReason(reasons []string, want string) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}

func wpFloatPtr(value float64) *float64 { return &value }
