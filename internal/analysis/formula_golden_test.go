package analysis

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/resources"
)

type formulaGolden struct {
	Drift            segmentFormulaGolden      `json:"drift"`
	Decoupling       segmentFormulaGolden      `json:"decoupling"`
	Polarization     polarizationFormulaGolden `json:"polarization"`
	EfficiencyFactor formulaStatusGolden       `json:"efficiency_factor"`
	VariabilityIndex variabilityFormulaGolden  `json:"variability_index"`
	ZScore           zScoreFormulaGolden       `json:"z_score"`
}

type segmentFormulaGolden struct {
	FormulaRef                  string  `json:"formula_ref"`
	MethodContains              string  `json:"method_contains"`
	Value                       float64 `json:"value"`
	AvgHRFirstHalf              float64 `json:"avg_hr_first_half"`
	AvgHRSecondHalf             float64 `json:"avg_hr_second_half"`
	RatioFirst                  float64 `json:"ratio_first"`
	RatioSecond                 float64 `json:"ratio_second"`
	ZeroDenominatorInsufficient bool    `json:"zero_denominator_insufficient"`
}

type polarizationFormulaGolden struct {
	FormulaRef        string    `json:"formula_ref"`
	InputZones        []float64 `json:"input_zones"`
	LowShare          float64   `json:"low_share"`
	ModerateShare     float64   `json:"moderate_share"`
	HighShare         float64   `json:"high_share"`
	Index             float64   `json:"index"`
	State             string    `json:"state"`
	Classification    string    `json:"classification"`
	ModerateZeroState string    `json:"moderate_zero_state"`
	HighZeroState     string    `json:"high_zero_state"`
}

type formulaStatusGolden struct {
	FormulaRef  string `json:"formula_ref"`
	Status      string `json:"status"`
	LocalOutput bool   `json:"local_output"`
}

type variabilityFormulaGolden struct {
	FormulaRef   string  `json:"formula_ref"`
	Status       string  `json:"status"`
	SourceField  string  `json:"source_field"`
	OutputField  string  `json:"output_field"`
	FixtureValue float64 `json:"fixture_value"`
}

type zScoreFormulaGolden struct {
	FormulaRef         string    `json:"formula_ref"`
	Baseline           []float64 `json:"baseline"`
	Current            []float64 `json:"current"`
	BaselineMean       float64   `json:"baseline_mean"`
	SampleStdDev       float64   `json:"sample_stddev"`
	ZScore             float64   `json:"z_score"`
	ZeroVarianceStatus string    `json:"zero_variance_status"`
}

func loadFormulaGolden(t *testing.T) formulaGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "analysis", "formula_golden.json"))
	if err != nil {
		t.Fatalf("read formula golden: %v", err)
	}
	var golden formulaGolden
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("decode formula golden: %v", err)
	}
	return golden
}

func TestFormulaGoldenSegmentStats(t *testing.T) {
	t.Parallel()
	golden := loadFormulaGolden(t)

	drift, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDrift,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 30, End: 90},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 40, 50, 55, 60, 65, 100},
			SegmentMetricHeartRate: {90, 100, 100, 100, 200, 200, 210},
		},
	})
	if err != nil {
		t.Fatalf("drift golden computation error: %v", err)
	}
	assertSegmentGolden(t, "drift", drift, golden.Drift)
	assertGoldenFloat(t, "drift.avg_hr_first_half", drift.Details["avg_hr_first_half"], golden.Drift.AvgHRFirstHalf, 1e-9)
	assertGoldenFloat(t, "drift.avg_hr_second_half", drift.Details["avg_hr_second_half"], golden.Drift.AvgHRSecondHalf, 1e-9)

	decoupling, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDecoupling,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 50},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40, 50},
			SegmentMetricHeartRate: {100, 100, 100, 100, 100, 100},
			SegmentMetricWatts:     {200, 200, 200, 180, 180, 180},
		},
	})
	if err != nil {
		t.Fatalf("decoupling golden computation error: %v", err)
	}
	assertSegmentGolden(t, "decoupling", decoupling, golden.Decoupling)
	assertGoldenFloat(t, "decoupling.ratio_first", decoupling.Details["ratio_first"], golden.Decoupling.RatioFirst, 1e-9)
	assertGoldenFloat(t, "decoupling.ratio_second", decoupling.Details["ratio_second"], golden.Decoupling.RatioSecond, 1e-9)
}

func TestFormulaGoldenSegmentStatsZeroDenominators(t *testing.T) {
	t.Parallel()
	golden := loadFormulaGolden(t)

	drift, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDrift,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 50},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40, 50},
			SegmentMetricHeartRate: {0, 0, 0, 120, 120, 120},
		},
	})
	if err != nil {
		t.Fatalf("drift zero denominator computation error: %v", err)
	}
	if drift.InsufficientSample != golden.Drift.ZeroDenominatorInsufficient || drift.Value != nil || drift.FormulaRef != golden.Drift.FormulaRef {
		t.Fatalf("drift zero-denominator result = %#v, want insufficient=%v, nil value, ref %s", drift, golden.Drift.ZeroDenominatorInsufficient, golden.Drift.FormulaRef)
	}

	decoupling, err := ComputeActivitySegmentStats(SegmentStatsInput{
		Stat:   SegmentStatDecoupling,
		Bounds: SegmentBounds{Axis: SegmentAxisTimeSeconds, Start: 0, End: 50},
		Streams: map[string][]float64{
			SegmentAxisTimeSeconds: {0, 10, 20, 30, 40, 50},
			SegmentMetricHeartRate: {100, 100, 100, 100, 100, 100},
			SegmentMetricWatts:     {0, 0, 0, 180, 180, 180},
		},
	})
	if err != nil {
		t.Fatalf("decoupling zero denominator computation error: %v", err)
	}
	if decoupling.InsufficientSample != golden.Decoupling.ZeroDenominatorInsufficient || decoupling.Value != nil || decoupling.FormulaRef != golden.Decoupling.FormulaRef {
		t.Fatalf("decoupling zero-denominator result = %#v, want insufficient=%v, nil value, ref %s", decoupling, golden.Decoupling.ZeroDenominatorInsufficient, golden.Decoupling.FormulaRef)
	}
}

func TestFormulaGoldenPolarization(t *testing.T) {
	t.Parallel()
	golden := loadFormulaGolden(t)

	got := ComputeZoneBalance(golden.Polarization.InputZones)
	assertGoldenFloat(t, "polarization.low_share", got.LowShare, golden.Polarization.LowShare, 1e-6)
	assertGoldenFloat(t, "polarization.moderate_share", got.ModerateShare, golden.Polarization.ModerateShare, 1e-6)
	assertGoldenFloat(t, "polarization.high_share", got.HighShare, golden.Polarization.HighShare, 1e-6)
	if got.Index == nil {
		t.Fatalf("polarization index = nil, want %v", golden.Polarization.Index)
	}
	assertGoldenFloat(t, "polarization.index", *got.Index, golden.Polarization.Index, 1e-5)
	if got.State != golden.Polarization.State || got.Classification != golden.Polarization.Classification {
		t.Fatalf("polarization state/classification = %s/%s, want %s/%s", got.State, got.Classification, golden.Polarization.State, golden.Polarization.Classification)
	}

	moderateZero := ComputeZoneBalance([]float64{700, 100, 0, 200})
	if moderateZero.State != golden.Polarization.ModerateZeroState || moderateZero.Index != nil {
		t.Fatalf("moderate-zero polarization = %#v, want state %s and nil index", moderateZero, golden.Polarization.ModerateZeroState)
	}
	highZero := ComputeZoneBalance([]float64{700, 100, 100, 0})
	if highZero.State != golden.Polarization.HighZeroState || highZero.Index != nil {
		t.Fatalf("high-zero polarization = %#v, want state %s and nil index", highZero, golden.Polarization.HighZeroState)
	}
}

func TestFormulaGoldenZScoreAndResourceOnlyStatuses(t *testing.T) {
	t.Parallel()
	golden := loadFormulaGolden(t)

	got := ComputeBaselineStats(golden.ZScore.Baseline, golden.ZScore.Current, 2, false)
	assertGoldenPtr(t, "z_score.baseline_mean", got.BaselineMean, golden.ZScore.BaselineMean, 1e-9)
	assertGoldenPtr(t, "z_score.sample_stddev", got.BaselineStdDev, golden.ZScore.SampleStdDev, 1e-6)
	assertGoldenPtr(t, "z_score.z_score", got.ZScore, golden.ZScore.ZScore, 1e-5)
	zeroVariance := ComputeBaselineStats([]float64{50, 50}, []float64{60}, 2, false)
	if zeroVariance.Status != golden.ZScore.ZeroVarianceStatus || zeroVariance.ZScore != nil {
		t.Fatalf("zero-variance z-score = %#v, want status %s and nil z-score", zeroVariance, golden.ZScore.ZeroVarianceStatus)
	}

	if golden.EfficiencyFactor.FormulaRef != resources.AnalysisFormulaRefEfficiencyFactor || golden.EfficiencyFactor.Status != "resource_only" || golden.EfficiencyFactor.LocalOutput {
		t.Fatalf("efficiency factor golden = %#v, want resource-only canonical ref with no local output", golden.EfficiencyFactor)
	}
	if _, err := ParseMetric("efficiency_factor"); err == nil {
		t.Fatal("ParseMetric(efficiency_factor) succeeded; EF is currently resource-only and must not gain analyzer output without updating formula guards")
	}
	sources := MetricSources(Metric("vi"))
	if len(sources) != 1 || sources[0].Tool != "get_extended_metrics" || sources[0].Field != golden.VariabilityIndex.OutputField || golden.VariabilityIndex.Status != "upstream_derived" {
		t.Fatalf("VI sources/golden = %#v/%#v, want upstream-derived get_extended_metrics.%s", sources, golden.VariabilityIndex, golden.VariabilityIndex.OutputField)
	}
}

func assertSegmentGolden(t *testing.T, name string, got SegmentStatsResult, want segmentFormulaGolden) {
	t.Helper()
	if got.Value == nil {
		t.Fatalf("%s value = nil, want %v", name, want.Value)
	}
	assertGoldenFloat(t, name+".value", *got.Value, want.Value, 1e-9)
	if got.FormulaRef != want.FormulaRef {
		t.Fatalf("%s formula_ref = %q, want %q", name, got.FormulaRef, want.FormulaRef)
	}
	if !strings.Contains(got.Method, want.MethodContains) {
		t.Fatalf("%s method = %q, want it to contain %q", name, got.Method, want.MethodContains)
	}
}

func assertGoldenPtr(t *testing.T, name string, got *float64, want float64, tolerance float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s = nil, want %v", name, want)
	}
	assertGoldenFloat(t, name, *got, want, tolerance)
}

func assertGoldenFloat(t *testing.T, name string, got float64, want float64, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %v, want %v (tolerance %g); formula golden drifted", name, got, want, tolerance)
	}
}
