package analysis

import (
	"math"
	"sort"
)

// DistributionInput contains numeric samples and histogram options.
type DistributionInput struct {
	Metric      string
	Unit        string
	Samples     []NumericSample
	BucketCount int
	Buckets     []float64
	Quantiles   []float64
	SampleGrain SampleGrain
}

// DistributionBucket is a half-open bucket except the final bucket includes its upper bound.
type DistributionBucket struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Count int     `json:"count"`
}

// QuantileValue is one requested quantile.
type QuantileValue struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

// DistributionResult is the terse distribution result.
type DistributionResult struct {
	Metric      string               `json:"metric"`
	Unit        string               `json:"unit,omitempty"`
	SampleGrain string               `json:"sample_grain"`
	Stats       BasicStats           `json:"stats"`
	Quantiles   []QuantileValue      `json:"quantiles,omitempty"`
	Histogram   []DistributionBucket `json:"histogram_buckets,omitempty"`
	BelowRange  int                  `json:"below_range,omitempty"`
	AboveRange  int                  `json:"above_range,omitempty"`
}

// ComputeDistribution computes summary statistics, quantiles, and histogram buckets.
func ComputeDistribution(input DistributionInput) DistributionResult {
	values := Values(input.Samples)
	sort.Float64s(values)
	result := DistributionResult{Metric: input.Metric, Unit: input.Unit, SampleGrain: string(input.SampleGrain), Stats: Stats(values)}
	for _, q := range input.Quantiles {
		if q < 0 || q > 1 || len(values) == 0 {
			continue
		}
		result.Quantiles = append(result.Quantiles, QuantileValue{Quantile: Round(q), Value: Round(quantileR7(values, q))})
	}
	buckets, below, above := histogram(values, input.BucketCount, input.Buckets)
	result.Histogram = buckets
	result.BelowRange = below
	result.AboveRange = above
	return result
}

func quantileR7(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	h := 1 + (float64(len(sorted))-1)*q
	lo := int(math.Floor(h)) - 1
	hi := int(math.Ceil(h)) - 1
	if lo == hi {
		return sorted[lo]
	}
	frac := h - math.Floor(h)
	return sorted[lo] + frac*(sorted[hi]-sorted[lo])
}

func histogram(sorted []float64, bucketCount int, explicit []float64) ([]DistributionBucket, int, int) {
	if len(sorted) == 0 {
		return nil, 0, 0
	}
	if len(explicit) > 0 {
		bounds := append([]float64(nil), explicit...)
		sort.Float64s(bounds)
		if len(bounds) < 2 {
			return nil, 0, 0
		}
		buckets := make([]DistributionBucket, 0, len(bounds)-1)
		for i := 0; i < len(bounds)-1; i++ {
			buckets = append(buckets, DistributionBucket{Lower: Round(bounds[i]), Upper: Round(bounds[i+1])})
		}
		var below, above int
		for _, value := range sorted {
			switch {
			case value < bounds[0]:
				below++
			case value > bounds[len(bounds)-1]:
				above++
			default:
				for i := range buckets {
					last := i == len(buckets)-1
					if (value >= bounds[i] && value < bounds[i+1]) || (last && value == bounds[i+1]) {
						buckets[i].Count++
						break
					}
				}
			}
		}
		return buckets, below, above
	}
	if bucketCount <= 0 {
		bucketCount = 10
	}
	minValue, maxValue := sorted[0], sorted[len(sorted)-1]
	if minValue == maxValue {
		return []DistributionBucket{{Lower: Round(minValue - 0.5), Upper: Round(maxValue + 0.5), Count: len(sorted)}}, 0, 0
	}
	width := (maxValue - minValue) / float64(bucketCount)
	buckets := make([]DistributionBucket, bucketCount)
	for i := range buckets {
		lower := minValue + float64(i)*width
		buckets[i] = DistributionBucket{Lower: Round(lower), Upper: Round(lower + width)}
	}
	for _, value := range sorted {
		idx := int(math.Floor((value - minValue) / width))
		if idx >= bucketCount {
			idx = bucketCount - 1
		}
		buckets[idx].Count++
	}
	return buckets, 0, 0
}
