package analysis

import (
	"math"
	"testing"
)

func TestComputeTrainingLoadMonotony(t *testing.T) {
	got, err := ComputeTrainingLoadMonotony([]float64{0, 40, 60, 0, 80, 20, 0})
	if err != nil {
		t.Fatalf("ComputeTrainingLoadMonotony() error = %v", err)
	}
	if math.Abs(got.Mean-28.571428571428573) > 1e-12 || math.Abs(got.PopulationStandardDev-29.965967090575756) > 1e-12 || math.Abs(got.Monotony-0.9534625892455925) > 1e-12 || got.ZeroVariance {
		t.Fatalf("result = %#v, want hand-checkable mixed-load vector", got)
	}
}

func TestComputeTrainingLoadMonotonyZeroVariance(t *testing.T) {
	for _, loads := range [][]float64{{0, 0}, {40, 40, 40}} {
		got, err := ComputeTrainingLoadMonotony(loads)
		if err != nil {
			t.Fatalf("loads %v error = %v", loads, err)
		}
		if !got.ZeroVariance || got.Monotony != 0 || math.IsNaN(got.Monotony) || math.IsInf(got.Monotony, 0) {
			t.Fatalf("loads %v result = %#v, want finite zero-variance result", loads, got)
		}
	}
}

func TestComputeTrainingLoadMonotonyScalesFiniteLargeLoads(t *testing.T) {
	got, err := ComputeTrainingLoadMonotony([]float64{math.MaxFloat64, math.MaxFloat64 / 2})
	if err != nil {
		t.Fatalf("large loads error = %v", err)
	}
	if math.IsNaN(got.Mean) || math.IsInf(got.Mean, 0) || math.IsNaN(got.PopulationStandardDev) || math.IsInf(got.PopulationStandardDev, 0) || math.IsNaN(got.Monotony) || math.IsInf(got.Monotony, 0) {
		t.Fatalf("large loads result = %#v, want finite values", got)
	}
}

func TestComputeTrainingLoadMonotonyPreservesScaledVarianceWhenOutputUnderflows(t *testing.T) {
	got, err := ComputeTrainingLoadMonotony([]float64{0, math.SmallestNonzeroFloat64})
	if err != nil {
		t.Fatalf("underflow loads error = %v", err)
	}
	if got.ZeroVariance || got.Monotony != 1 || got.Mean != 0 || got.PopulationStandardDev != 0 {
		t.Fatalf("underflow loads result = %#v, want nonzero scaled variance with finite underflowed fields", got)
	}
}

func TestComputeTrainingLoadMonotonyRejectsInvalidInput(t *testing.T) {
	for _, loads := range [][]float64{nil, {-1}, {math.NaN()}, {math.Inf(1)}} {
		if _, err := ComputeTrainingLoadMonotony(loads); err == nil {
			t.Fatalf("loads %v error = nil, want rejection", loads)
		}
	}
}
