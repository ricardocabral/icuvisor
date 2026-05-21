package analysis

import (
	"errors"
	"math"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/resources"
)

func TestRequiredSegmentStreamKeysUsesCanonicalKeys(t *testing.T) {
	keys, err := RequiredSegmentStreamKeys(SegmentStatDecoupling, "", SegmentAxisDistanceMeter)
	if err != nil {
		t.Fatalf("RequiredSegmentStreamKeys() error = %v", err)
	}
	want := []string{"distance", "heart_rate", "time", "watts"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %#v, want %#v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %#v, want %#v", keys, want)
		}
	}
}

func TestRequiredSegmentStreamKeysScalarDistanceOmitsTime(t *testing.T) {
	keys, err := RequiredSegmentStreamKeys(SegmentStatMean, SegmentMetricWatts, SegmentAxisDistanceMeter)
	if err != nil {
		t.Fatalf("RequiredSegmentStreamKeys() error = %v", err)
	}
	want := []string{"distance", "watts"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %#v, want %#v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %#v, want %#v", keys, want)
		}
	}
}

func TestRequiredSegmentStreamKeysRejectsIncompatibleMetric(t *testing.T) {
	_, err := RequiredSegmentStreamKeys(SegmentStatDrift, SegmentMetricHeartRate, SegmentAxisTimeSeconds)
	if !errors.Is(err, ErrInvalidSegmentStatsInput) {
		t.Fatalf("RequiredSegmentStreamKeys() error = %v, want ErrInvalidSegmentStatsInput", err)
	}
}

func TestComputeActivitySegmentStatsScalarTimeSlice(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatMedian,
		Metric: SegmentMetricWatts,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 10, End: 40},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40, 50},
			SegmentMetricWatts:     {100, 200, 400, 800, 1600, 3200},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.Value == nil || *got.Value != 600 {
		t.Fatalf("Value = %v, want 600", got.Value)
	}
	if got.N != 4 || got.InsufficientSample {
		t.Fatalf("n/insufficient = %d/%v, want 4/false", got.N, got.InsufficientSample)
	}
}

func TestComputeActivitySegmentStatsDistanceSliceP90NearestRank(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatP90,
		Metric: SegmentMetricHeartRate,
		Bounds: SegmentBounds{Axis: SegmentAxisDistanceMeter, Start: 100, End: 500},
		Streams: map[string][]float64{
			SegmentAxisDistanceMeter: {0, 100, 200, 300, 400, 500, 600},
			SegmentAxisTimeSeconds:   {0, 10, 20, 30, 40, 50, 60},
			SegmentMetricHeartRate:   {100, 120, 130, 140, 150, 160, 170},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.Value == nil || *got.Value != 160 {
		t.Fatalf("Value = %v, want nearest-rank p90 160", got.Value)
	}
	if got.Segment.Axis != SegmentAxisDistanceMeter || got.N != 5 {
		t.Fatalf("segment/n = %#v/%d, want distance/5", got.Segment, got.N)
	}
}

func TestComputeActivitySegmentStatsInCoverageNoSamplesIsInsufficient(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatMean,
		Metric: SegmentMetricWatts,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 5, End: 6},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20},
			SegmentMetricWatts:     {100, 200, 300},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if !got.InsufficientSample || got.N != 0 || got.Value != nil {
		t.Fatalf("result = %#v, want insufficient n=0 without value", got)
	}
}

func TestComputeActivitySegmentStatsOutOfRangeErrors(t *testing.T) {
	_, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatMean,
		Metric: SegmentMetricWatts,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 10, End: 70},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30},
			SegmentMetricWatts:     {100, 200, 300, 400},
		},
	})
	if !errors.Is(err, ErrSegmentOutOfRange) {
		t.Fatalf("ComputeActivitySegmentStats() error = %v, want ErrSegmentOutOfRange", err)
	}
}

func TestComputeActivitySegmentStatsDriftInsufficientSample(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDrift,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 20},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30},
			SegmentMetricHeartRate: {120, 122, 130, 132},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if !got.InsufficientSample || got.Value != nil || got.N != 3 || got.FormulaRef != resources.AnalysisFormulaRefHRDrift {
		t.Fatalf("result = %#v, want insufficient drift with n=3 and formula ref", got)
	}
}

func TestComputeActivitySegmentStatsDriftSplitsByRequestedTimeBounds(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDrift,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 30, End: 90},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 40, 50, 55, 60, 65, 100},
			SegmentMetricHeartRate: {90, 100, 100, 100, 200, 200, 210},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.Value == nil || *got.Value != 100 {
		t.Fatalf("Value = %v, want 100 when split at requested midpoint 60s", got.Value)
	}
}

func TestComputeActivitySegmentStatsDecoupling(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDecoupling,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 50},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40, 50},
			SegmentMetricHeartRate: {100, 100, 100, 100, 100, 100},
			SegmentMetricWatts:     {200, 200, 200, 180, 180, 180},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.Value == nil || *got.Value != 10 {
		t.Fatalf("Value = %v, want 10", got.Value)
	}
	if got.FormulaRef != resources.AnalysisFormulaRefPwHRDecoupling || got.N != 6 {
		t.Fatalf("formula/n = %q/%d, want Pw:HR ref/6", got.FormulaRef, got.N)
	}
}

func TestComputeActivitySegmentStatsNPAnchorsWindowToRequestedStart(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatNP,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 5, End: 60},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 40, 60, 70},
			SegmentMetricWatts:     {100, 200, 300, 400},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.N != 2 || got.InsufficientSample || got.Value == nil {
		t.Fatalf("result = %#v, want requested-start anchored NP windows", got)
	}
}

func TestComputeActivitySegmentStatsDistanceDerivedSkipsNonFiniteTimes(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDrift,
		Bounds: SegmentBounds{Axis: SegmentAxisDistanceMeter, Start: 0, End: 400},
		Streams: map[string][]float64{
			SegmentAxisDistanceMeter: {0, 100, 200, 300, 400},
			SegmentAxisTimeSeconds:   {0, 10, math.NaN(), 30, 40},
			SegmentMetricHeartRate:   {100, 100, 200, 200, 200},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.N != 4 || got.Value == nil || *got.Value != 100 {
		t.Fatalf("result = %#v, want non-finite time skipped with four paired samples", got)
	}
}

func TestComputeActivitySegmentStatsNPAllowsZeroWatts(t *testing.T) {
	got, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatNP,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 40},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40},
			SegmentMetricWatts:     {0, 0, 300, 300, 300},
		},
	})
	if err != nil {
		t.Fatalf("ComputeActivitySegmentStats() error = %v", err)
	}
	if got.Value == nil || *got.Value <= 0 {
		t.Fatalf("Value = %v, want positive NP with zero-watt samples included", got.Value)
	}
	if got.N != 2 || got.InsufficientSample {
		t.Fatalf("n/insufficient = %d/%v, want 2/false", got.N, got.InsufficientSample)
	}
}

func TestComputeActivitySegmentStatsIFRequiresFTP(t *testing.T) {
	_, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatIF,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 40},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40},
			SegmentMetricWatts:     {0, 0, 300, 300, 300},
		},
	})
	if !errors.Is(err, ErrInvalidSegmentStatsInput) {
		t.Fatalf("ComputeActivitySegmentStats() error = %v, want ErrInvalidSegmentStatsInput", err)
	}
}
