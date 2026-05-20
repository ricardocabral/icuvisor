package analysis

import (
	"math"
	"testing"
)

func TestComputeTrendBaselineAndInsufficient(t *testing.T) {
	samples := []NumericSample{
		{Date: "2026-05-01", Value: 10},
		{Date: "2026-05-02", Value: 12},
		{Date: "2026-05-04", Value: 14},
		{Date: "2026-05-05", Value: 16},
		{Date: "2026-05-06", Value: 18},
		{Date: "2026-05-07", Value: 20},
		{Date: "2026-05-08", Value: 22},
	}
	baseline := []NumericSample{
		{Value: 8}, {Value: 9}, {Value: 10}, {Value: 11}, {Value: 12}, {Value: 13}, {Value: 14},
	}
	got, series := ComputeTrend(TrendInput{Metric: "training_load", Unit: "load", Samples: samples, BaselineSamples: baseline, RollingWindow: 3, MinSamples: 7, BaselineMinSamples: 7, SampleGrain: SampleGrainDaily})
	if got.N != 7 || got.WindowMean == nil || *got.WindowMean != 16 || got.RollingLatestMean == nil || *got.RollingLatestMean != 20 {
		t.Fatalf("trend result = %#v, series=%#v", got, series)
	}
	if got.Slope == nil || *got.Slope != 2 || got.TrendDirection != "increasing" {
		t.Fatalf("slope/direction = %#v/%q, want 2/increasing", got.Slope, got.TrendDirection)
	}
	if got.AbsoluteDelta == nil || *got.AbsoluteDelta != 5 || got.PercentDelta == nil || *got.PercentDelta != 45.455 || got.ZScore == nil {
		t.Fatalf("baseline deltas = %#v", got)
	}
	short, _ := ComputeTrend(TrendInput{Metric: "ctl", Samples: samples[:3], RollingWindow: 2, MinSamples: 7, BaselineMinSamples: 7, SampleGrain: SampleGrainDaily})
	if !InsufficientSample(short.N, 7) || short.Slope != nil {
		t.Fatalf("short trend = %#v, want insufficient with nil slope", short)
	}
}

func TestComputeTrendWeeklySlopeUsesBucketIndexes(t *testing.T) {
	samples := []NumericSample{{Key: "w1", Bucket: 0, Value: 100}, {Key: "w3", Bucket: 2, Value: 120}, {Key: "w4", Bucket: 3, Value: 130}, {Key: "w5", Bucket: 4, Value: 140}}
	got, _ := ComputeTrend(TrendInput{Metric: "weekly_tss", Samples: samples, RollingWindow: 2, MinSamples: 4, BaselineMinSamples: 4, SampleGrain: SampleGrainWeekly})
	if got.Slope == nil || *got.Slope != 10 {
		t.Fatalf("weekly slope = %#v, want 10 per bucket despite missing week", got.Slope)
	}
}

func TestComputeTrendWeeklySlopeSkipsInvalidValues(t *testing.T) {
	samples := []NumericSample{{Key: "w1", Bucket: 0, Value: 100}, {Key: "w2", Bucket: 1, Value: math.NaN()}, {Key: "w3", Bucket: 2, Value: 120}, {Key: "w4", Bucket: 3, Value: math.Inf(1)}, {Key: "w5", Bucket: 4, Value: 140}}
	got, _ := ComputeTrend(TrendInput{Metric: "weekly_tss", Samples: samples, RollingWindow: 2, MinSamples: 3, BaselineMinSamples: 4, SampleGrain: SampleGrainWeekly})
	if got.N != 3 || got.Slope == nil || *got.Slope != 10 {
		t.Fatalf("weekly trend = %#v, want n=3 and slope 10 after invalid samples are skipped", got)
	}
}

func TestComputeTrendDailySlopeUsesDenseIndexesAfterInvalidValues(t *testing.T) {
	samples := []NumericSample{{Date: "2026-05-01", Value: 100}, {Date: "2026-05-02", Value: math.NaN()}, {Date: "2026-05-03", Value: 120}, {Date: "2026-05-04", Value: 140}}
	got, _ := ComputeTrend(TrendInput{Metric: "ctl", Samples: samples, RollingWindow: 2, MinSamples: 3, BaselineMinSamples: 7, SampleGrain: SampleGrainDaily})
	if got.N != 3 || got.Slope == nil || *got.Slope != 20 {
		t.Fatalf("daily trend = %#v, want n=3 and dense-index slope 20 after invalid samples are skipped", got)
	}
}

func TestComputeDistributionQuantilesHistogramAndMissing(t *testing.T) {
	samples := []NumericSample{{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4}, {Value: 5}}
	got := ComputeDistribution(DistributionInput{Metric: "hrv", Unit: "ms", Samples: samples, Buckets: []float64{2, 4}, Quantiles: []float64{0.25, 0.5, 0.75}, SampleGrain: SampleGrainDaily})
	if got.Stats.N != 5 || got.Stats.Mean == nil || *got.Stats.Mean != 3 || got.Stats.StdDev == nil || *got.Stats.StdDev != 1.581 {
		t.Fatalf("stats = %#v", got.Stats)
	}
	if len(got.Quantiles) != 3 || got.Quantiles[1].Value != 3 {
		t.Fatalf("quantiles = %#v", got.Quantiles)
	}
	if got.BelowRange != 1 || got.AboveRange != 1 || len(got.Histogram) != 1 || got.Histogram[0].Count != 3 {
		t.Fatalf("histogram = %#v below=%d above=%d", got.Histogram, got.BelowRange, got.AboveRange)
	}
	if missing := MissingSamples(7, got.Stats.N); missing != 2 {
		t.Fatalf("MissingSamples = %d, want 2", missing)
	}
}

func TestComputeCorrelationPearsonSpearmanAndZeroVariance(t *testing.T) {
	pairs := []PairedSample{{X: 1, Y: 2}, {X: 2, Y: 4}, {X: 3, Y: 6}, {X: 4, Y: 8}}
	got := ComputeCorrelation(CorrelationInput{MetricX: "ctl", MetricY: "training_load", Method: CorrelationPearson, Pairs: pairs})
	if got.Coefficient == nil || *got.Coefficient != 1 || got.Slope == nil || *got.Slope != 2 || got.Intercept == nil || *got.Intercept != 0 || got.Strength != "very_strong" {
		t.Fatalf("pearson = %#v", got)
	}
	spearman := ComputeCorrelation(CorrelationInput{Method: CorrelationSpearman, Pairs: []PairedSample{{X: 1, Y: 10}, {X: 1, Y: 10}, {X: 3, Y: 30}}})
	if spearman.Coefficient == nil || *spearman.Coefficient != 1 {
		t.Fatalf("spearman ties = %#v", spearman)
	}
	zero := ComputeCorrelation(CorrelationInput{Method: CorrelationPearson, Pairs: []PairedSample{{X: 1, Y: 2}, {X: 1, Y: 3}}})
	if zero.Coefficient != nil || zero.Slope != nil || len(zero.Boundaries) == 0 {
		t.Fatalf("zero variance = %#v", zero)
	}
	invalid := ComputeCorrelation(CorrelationInput{Method: CorrelationPearson, Pairs: []PairedSample{{X: 1, Y: 2}, {X: 2, Y: 4}, {X: 3, Y: math.NaN()}}})
	if invalid.N != 2 || invalid.Coefficient == nil || *invalid.Coefficient != 1 {
		t.Fatalf("invalid pair filtering = %#v, want n=2 with usable pair coefficient", invalid)
	}
}

func TestComputeEffortsDeltaUnitExplicitRows(t *testing.T) {
	cur, base := 1200.0, 1260.0
	got := ComputeEffortsDelta(EffortsDeltaInput{Sport: "Run", Family: EffortFamilyPace, UnitSystem: "metric", Current: []EffortBucketValue{{Bucket: 5000, Value: &cur, ActivityID: "a1"}}, Baseline: []EffortBucketValue{{Bucket: 5000, Value: &base, ActivityID: "b1"}}})
	if got.N != 1 || got.BetterDirection != "lower" || len(got.Buckets) != 1 {
		t.Fatalf("efforts result = %#v", got)
	}
	row := got.Buckets[0]
	if row.CurrentElapsedSeconds == nil || *row.CurrentElapsedSeconds != 1200 || row.AbsoluteDeltaSeconds == nil || *row.AbsoluteDeltaSeconds != -60 || row.CurrentPaceSecondsPerKM == nil || *row.CurrentPaceSecondsPerKM != 240 || row.AbsoluteDeltaPaceSecondsPerKM == nil || *row.AbsoluteDeltaPaceSecondsPerKM != -12 {
		t.Fatalf("pace row = %#v", row)
	}
	missing := ComputeEffortsDelta(EffortsDeltaInput{Sport: "Ride", Family: EffortFamilyPower, Current: []EffortBucketValue{{Bucket: 60, Value: &cur}}, Baseline: nil})
	if missing.N != 0 || !missing.Buckets[0].BaselineMissing {
		t.Fatalf("missing efforts = %#v", missing)
	}
}
