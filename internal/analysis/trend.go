package analysis

import "math"

// TrendInput contains analyzer-ready samples for trend computation.
type TrendInput struct {
	Metric             string
	Unit               string
	Samples            []NumericSample
	BaselineSamples    []NumericSample
	RollingWindow      int
	MinSamples         int
	BaselineMinSamples int
	SampleGrain        SampleGrain
}

// TrendPoint is emitted only in include_full series responses by tool adapters.
type TrendPoint struct {
	Key         string   `json:"key,omitempty"`
	Date        string   `json:"date,omitempty"`
	Value       float64  `json:"value"`
	RollingMean *float64 `json:"rolling_mean,omitempty"`
}

// TrendResult is the terse trend computation result.
type TrendResult struct {
	Metric             string   `json:"metric"`
	Unit               string   `json:"unit,omitempty"`
	SampleGrain        string   `json:"sample_grain"`
	N                  int      `json:"n"`
	WindowMean         *float64 `json:"window_mean,omitempty"`
	RollingLatestMean  *float64 `json:"rolling_latest_mean,omitempty"`
	Slope              *float64 `json:"slope,omitempty"`
	BaselineMean       *float64 `json:"baseline_mean,omitempty"`
	AbsoluteDelta      *float64 `json:"absolute_delta,omitempty"`
	PercentDelta       *float64 `json:"percent_delta,omitempty"`
	ZScore             *float64 `json:"z_score,omitempty"`
	TrendDirection     string   `json:"trend_direction,omitempty"`
	BaselineSampleSize int      `json:"baseline_n,omitempty"`
	Boundaries         []string `json:"-"`
}

// ComputeTrend computes rolling mean, OLS slope, and optional baseline deltas.
func ComputeTrend(input TrendInput) (TrendResult, []TrendPoint) {
	values := Values(input.Samples)
	result := TrendResult{Metric: input.Metric, Unit: input.Unit, SampleGrain: string(input.SampleGrain), N: len(values), BaselineSampleSize: len(Values(input.BaselineSamples))}
	if windowMean, ok := numericMean(values); ok {
		result.WindowMean = roundPtr(windowMean)
	}
	series := rollingSeries(input.Samples, input.RollingWindow)
	for i := len(series) - 1; i >= 0; i-- {
		if series[i].RollingMean != nil {
			result.RollingLatestMean = series[i].RollingMean
			break
		}
	}
	if len(values) >= input.MinSamples {
		if slope, ok := trendSlope(input.Samples, input.SampleGrain); ok {
			result.Slope = roundPtr(slope)
			result.TrendDirection = trendDirection(slope)
		} else {
			result.Boundaries = append(result.Boundaries, "trend slope unavailable because x variance is zero")
		}
	}
	baselineValues := Values(input.BaselineSamples)
	if len(baselineValues) >= input.BaselineMinSamples {
		baselineStats := Stats(baselineValues)
		result.BaselineMean = baselineStats.Mean
		if result.WindowMean != nil && baselineStats.Mean != nil {
			delta := *result.WindowMean - *baselineStats.Mean
			result.AbsoluteDelta = roundPtr(delta)
			if *baselineStats.Mean != 0 {
				result.PercentDelta = roundPtr(delta / math.Abs(*baselineStats.Mean) * 100)
			} else {
				result.Boundaries = append(result.Boundaries, "percent_delta omitted because baseline mean is zero")
			}
			if baselineStats.StdDev != nil && *baselineStats.StdDev != 0 {
				result.ZScore = roundPtr(delta / *baselineStats.StdDev)
			} else {
				result.Boundaries = append(result.Boundaries, "z_score omitted because baseline standard deviation is zero")
			}
		}
	}
	return result, series
}

func rollingSeries(samples []NumericSample, window int) []TrendPoint {
	series := make([]TrendPoint, len(samples))
	if window <= 0 {
		window = 1
	}
	values := Values(samples)
	for i, sample := range samples {
		point := TrendPoint{Key: sample.Key, Date: sample.Date, Value: Round(sample.Value)}
		if i+1 >= window && i < len(values) {
			if avg, ok := numericMean(values[i+1-window : i+1]); ok {
				point.RollingMean = roundPtr(avg)
			}
		}
		series[i] = point
	}
	return series
}

func trendSlope(samples []NumericSample, grain SampleGrain) (float64, bool) {
	values := Values(samples)
	xs := make([]float64, len(values))
	for i := range values {
		xs[i] = float64(i)
	}
	if grain == SampleGrainWeekly {
		for i, sample := range samples {
			xs[i] = float64(sample.Bucket)
		}
	}
	return olsSlopeXY(xs, values)
}

func olsSlopeXY(xs []float64, values []float64) (float64, bool) {
	if len(xs) != len(values) || len(values) < 2 {
		return 0, false
	}
	n := float64(len(values))
	var sumX, sumY, sumXX, sumXY float64
	for i, y := range values {
		x := xs[i]
		sumX += x
		sumY += y
		sumXX += x * x
		sumXY += x * y
	}
	denom := n*sumXX - sumX*sumX
	if denom == 0 {
		return 0, false
	}
	return (n*sumXY - sumX*sumY) / denom, true
}

func trendDirection(slope float64) string {
	switch {
	case slope > 0:
		return "increasing"
	case slope < 0:
		return "decreasing"
	default:
		return "flat"
	}
}
