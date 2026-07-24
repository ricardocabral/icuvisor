package analysis

import (
	"errors"
	"math"
)

// MonotonyResult contains the Foster-style daily-load monotony calculation.
type MonotonyResult struct {
	Mean                  float64
	PopulationStandardDev float64
	Monotony              float64
	ZeroVariance          bool
}

// ComputeTrainingLoadMonotony computes mean load divided by population standard deviation.
// It scales inputs before accumulation so finite accepted loads do not overflow intermediate sums.
func ComputeTrainingLoadMonotony(loads []float64) (MonotonyResult, error) {
	if len(loads) == 0 {
		return MonotonyResult{}, errors.New("training load sample is empty")
	}
	maxLoad := 0.0
	for _, load := range loads {
		if math.IsNaN(load) || math.IsInf(load, 0) || load < 0 {
			return MonotonyResult{}, errors.New("training load must be finite and non-negative")
		}
		if load > maxLoad {
			maxLoad = load
		}
	}
	if maxLoad == 0 {
		return MonotonyResult{ZeroVariance: true}, nil
	}

	meanScaled := compensatedMean(loads, maxLoad)
	varianceScaled := 0.0
	compensation := 0.0
	for _, load := range loads {
		delta := load/maxLoad - meanScaled
		term := delta * delta
		y := term - compensation
		t := varianceScaled + y
		compensation = (t - varianceScaled) - y
		varianceScaled = t
	}
	varianceScaled /= float64(len(loads))
	if varianceScaled <= 0 || math.IsNaN(varianceScaled) || math.IsInf(varianceScaled, 0) {
		return MonotonyResult{ZeroVariance: true}, nil
	}

	sdScaled := math.Sqrt(varianceScaled)
	mean := finiteScaled(meanScaled, maxLoad)
	standardDeviation := finiteScaled(sdScaled, maxLoad)
	monotony := meanScaled / sdScaled
	if math.IsNaN(monotony) || math.IsInf(monotony, 0) {
		return MonotonyResult{}, errors.New("training load monotony is not finite")
	}
	return MonotonyResult{Mean: mean, PopulationStandardDev: standardDeviation, Monotony: monotony}, nil
}

func compensatedMean(loads []float64, scale float64) float64 {
	sum := 0.0
	compensation := 0.0
	for _, load := range loads {
		y := load/scale - compensation
		t := sum + y
		compensation = (t - sum) - y
		sum = t
	}
	return sum / float64(len(loads))
}

func finiteScaled(value, scale float64) float64 {
	if value == 0 {
		return 0
	}
	if value > math.MaxFloat64/scale {
		return math.MaxFloat64
	}
	return value * scale
}
