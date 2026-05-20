package analysis

import "math"

// SampleGrain is the analyzer sample grain echoed in _meta assumptions.
type SampleGrain string

const (
	SampleGrainDaily    SampleGrain = "daily"
	SampleGrainActivity SampleGrain = "activity"
	SampleGrainWeekly   SampleGrain = "weekly"
)

// NumericSample is one analyzer-ready numeric observation.
type NumericSample struct {
	Key        string  `json:"key,omitempty"`
	Date       string  `json:"date,omitempty"`
	ActivityID string  `json:"activity_id,omitempty"`
	Bucket     int     `json:"bucket,omitempty"`
	Value      float64 `json:"value"`
}

// BasicStats summarizes numeric samples using sample standard deviation.
type BasicStats struct {
	N      int      `json:"n"`
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Mean   *float64 `json:"mean,omitempty"`
	StdDev *float64 `json:"stddev,omitempty"`
}

// Round rounds values to three decimal places for public analyzer fields.
func Round(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return value
	}
	return math.Round(value*1000) / 1000
}

func roundPtr(value float64) *float64 { rounded := Round(value); return &rounded }

// Values extracts sample values in order.
func Values(samples []NumericSample) []float64 {
	values := make([]float64, 0, len(samples))
	for _, sample := range samples {
		if !math.IsNaN(sample.Value) && !math.IsInf(sample.Value, 0) {
			values = append(values, sample.Value)
		}
	}
	return values
}

// Stats returns count/min/max/mean/sample-stddev for values.
func Stats(values []float64) BasicStats {
	clean := make([]float64, 0, len(values))
	for _, value := range values {
		if !math.IsNaN(value) && !math.IsInf(value, 0) {
			clean = append(clean, value)
		}
	}
	stats := BasicStats{N: len(clean)}
	if len(clean) == 0 {
		return stats
	}
	minValue, maxValue, sum := clean[0], clean[0], 0.0
	for _, value := range clean {
		minValue = math.Min(minValue, value)
		maxValue = math.Max(maxValue, value)
		sum += value
	}
	mean := sum / float64(len(clean))
	stats.Min = roundPtr(minValue)
	stats.Max = roundPtr(maxValue)
	stats.Mean = roundPtr(mean)
	if len(clean) > 1 {
		var ss float64
		for _, value := range clean {
			delta := value - mean
			ss += delta * delta
		}
		stats.StdDev = roundPtr(math.Sqrt(ss / float64(len(clean)-1)))
	}
	return stats
}

// MissingSamples returns expected minus usable with a floor at zero.
func MissingSamples(expected int, usable int) int {
	if expected <= usable || expected <= 0 {
		return 0
	}
	if usable < 0 {
		usable = 0
	}
	return expected - usable
}

// AnalysisUsable reports whether a sample count satisfies a minimum.
func AnalysisUsable(n int, minSamples int) bool { return !InsufficientSample(n, minSamples) }

func numericMean(values []float64) (float64, bool) {
	if len(values) == 0 {
		return 0, false
	}
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values)), true
}
