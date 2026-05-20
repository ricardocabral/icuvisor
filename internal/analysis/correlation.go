package analysis

import (
	"math"
	"sort"
)

const (
	CorrelationPearson  = "pearson"
	CorrelationSpearman = "spearman"
)

// PairedSample is one x/y pair used by correlation.
type PairedSample struct {
	Key  string  `json:"key,omitempty"`
	Date string  `json:"date,omitempty"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

// CorrelationInput contains paired samples and method selection.
type CorrelationInput struct {
	MetricX string
	MetricY string
	Method  string
	LagDays int
	Pairs   []PairedSample
}

// CorrelationResult is the terse correlation result.
type CorrelationResult struct {
	MetricX          string   `json:"metric_x"`
	MetricY          string   `json:"metric_y"`
	Method           string   `json:"method"`
	LagDays          int      `json:"lag_days"`
	N                int      `json:"n"`
	Coefficient      *float64 `json:"coefficient,omitempty"`
	Slope            *float64 `json:"slope,omitempty"`
	Intercept        *float64 `json:"intercept,omitempty"`
	Direction        string   `json:"direction,omitempty"`
	Strength         string   `json:"strength,omitempty"`
	RegressionMethod string   `json:"regression_method,omitempty"`
	Boundaries       []string `json:"-"`
}

// ComputeCorrelation computes Pearson or Spearman correlation plus raw OLS regression.
func ComputeCorrelation(input CorrelationInput) CorrelationResult {
	method := input.Method
	if method == "" {
		method = CorrelationPearson
	}
	xs, ys := pairValues(input.Pairs)
	result := CorrelationResult{MetricX: input.MetricX, MetricY: input.MetricY, Method: method, LagDays: input.LagDays, N: len(xs), RegressionMethod: "raw_ols"}
	if slope, intercept, ok := ols(xs, ys); ok {
		result.Slope = roundPtr(slope)
		result.Intercept = roundPtr(intercept)
	} else {
		result.Boundaries = append(result.Boundaries, "regression omitted because x variance is zero")
	}
	corrX, corrY := xs, ys
	if method == CorrelationSpearman {
		corrX = averageRanks(xs)
		corrY = averageRanks(ys)
	}
	if coefficient, ok := pearson(corrX, corrY); ok {
		result.Coefficient = roundPtr(coefficient)
		result.Direction = coefficientDirection(coefficient)
		result.Strength = coefficientStrength(coefficient)
	} else {
		result.Boundaries = append(result.Boundaries, "coefficient omitted because variance is zero")
	}
	return result
}

func pairValues(pairs []PairedSample) ([]float64, []float64) {
	xs, ys := make([]float64, 0, len(pairs)), make([]float64, 0, len(pairs))
	for _, pair := range pairs {
		if math.IsNaN(pair.X) || math.IsInf(pair.X, 0) || math.IsNaN(pair.Y) || math.IsInf(pair.Y, 0) {
			continue
		}
		xs = append(xs, pair.X)
		ys = append(ys, pair.Y)
	}
	return xs, ys
}

func pearson(xs, ys []float64) (float64, bool) {
	if len(xs) != len(ys) || len(xs) < 2 {
		return 0, false
	}
	meanX, _ := numericMean(xs)
	meanY, _ := numericMean(ys)
	var cov, vx, vy float64
	for i := range xs {
		dx := xs[i] - meanX
		dy := ys[i] - meanY
		cov += dx * dy
		vx += dx * dx
		vy += dy * dy
	}
	if vx == 0 || vy == 0 {
		return 0, false
	}
	return cov / math.Sqrt(vx*vy), true
}

func ols(xs, ys []float64) (float64, float64, bool) {
	if len(xs) != len(ys) || len(xs) < 2 {
		return 0, 0, false
	}
	meanX, _ := numericMean(xs)
	meanY, _ := numericMean(ys)
	var num, denom float64
	for i := range xs {
		dx := xs[i] - meanX
		num += dx * (ys[i] - meanY)
		denom += dx * dx
	}
	if denom == 0 {
		return 0, 0, false
	}
	slope := num / denom
	return slope, meanY - slope*meanX, true
}

type rankValue struct {
	value float64
	idx   int
}

func averageRanks(values []float64) []float64 {
	items := make([]rankValue, len(values))
	for i, value := range values {
		items[i] = rankValue{value: value, idx: i}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].value < items[j].value })
	ranks := make([]float64, len(values))
	for i := 0; i < len(items); {
		j := i + 1
		for j < len(items) && items[j].value == items[i].value {
			j++
		}
		rank := (float64(i+1) + float64(j)) / 2
		for k := i; k < j; k++ {
			ranks[items[k].idx] = rank
		}
		i = j
	}
	return ranks
}

func coefficientDirection(value float64) string {
	switch {
	case value > 0:
		return "positive"
	case value < 0:
		return "negative"
	default:
		return "none"
	}
}

func coefficientStrength(value float64) string {
	abs := math.Abs(value)
	switch {
	case abs < 0.1:
		return "negligible"
	case abs < 0.3:
		return "weak"
	case abs < 0.5:
		return "moderate"
	case abs < 0.7:
		return "strong"
	default:
		return "very_strong"
	}
}
