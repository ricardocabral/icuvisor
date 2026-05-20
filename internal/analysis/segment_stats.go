package analysis

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/resources"
)

const (
	SegmentStatMean       = "mean"
	SegmentStatMedian     = "median"
	SegmentStatP90        = "p90"
	SegmentStatDecoupling = "decoupling"
	SegmentStatDrift      = "drift"
	SegmentStatNP         = "np"
	SegmentStatIF         = "if"

	SegmentMetricWatts          = "watts"
	SegmentMetricHeartRate      = "heart_rate"
	SegmentMetricCadence        = "cadence"
	SegmentMetricVelocitySmooth = "velocity_smooth"
	SegmentMetricDistance       = "distance"
	SegmentMetricTime           = "time"

	SegmentAxisTimeSeconds   = "time"
	SegmentAxisDistanceMeter = "distance"
)

const (
	minScalarSamples    = 1
	minSplitHalfSamples = 2
	minNPWindowSeconds  = 30.0
)

var (
	ErrInvalidSegmentStatsInput = errors.New("invalid activity segment stats input")
	ErrMissingSegmentStream     = errors.New("missing required activity stream")
	ErrSegmentOutOfRange        = errors.New("activity segment range outside stream coverage")
)

type MissingSegmentStreamError struct {
	Key string
}

func (e MissingSegmentStreamError) Error() string {
	return "missing required activity stream: " + e.Key
}

func (e MissingSegmentStreamError) Is(target error) bool {
	return target == ErrMissingSegmentStream
}

func MissingSegmentStreamKey(err error) string {
	var missing MissingSegmentStreamError
	if errors.As(err, &missing) {
		return missing.Key
	}
	return ""
}

type SegmentBounds struct {
	Axis  string  `json:"axis"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type SegmentStatsInput struct {
	Stat     string
	Metric   string
	Bounds   SegmentBounds
	Streams  map[string][]float64
	FTPWatts float64
}

type SegmentStatsResult struct {
	Stat               string               `json:"stat"`
	Metric             string               `json:"metric,omitempty"`
	Value              *float64             `json:"value,omitempty"`
	Unit               string               `json:"unit,omitempty"`
	Segment            SegmentBounds        `json:"segment"`
	N                  int                  `json:"n"`
	MinSamples         int                  `json:"min_samples"`
	Method             string               `json:"method"`
	FormulaRef         string               `json:"formula_ref,omitempty"`
	InsufficientSample bool                 `json:"insufficient_sample"`
	StreamsUsed        []string             `json:"streams_used"`
	Details            map[string]float64   `json:"details,omitempty"`
	Audit              map[string][]float64 `json:"audit,omitempty"`
}

func SegmentStatValues() []string {
	return []string{SegmentStatDecoupling, SegmentStatDrift, SegmentStatIF, SegmentStatMean, SegmentStatMedian, SegmentStatNP, SegmentStatP90}
}

func SegmentMetricValues() []string {
	return []string{SegmentMetricCadence, SegmentMetricDistance, SegmentMetricHeartRate, SegmentMetricTime, SegmentMetricVelocitySmooth, SegmentMetricWatts}
}

func RequiredSegmentStreamKeys(stat string, metric string, axis string) ([]string, error) {
	stat = strings.TrimSpace(stat)
	metric = strings.TrimSpace(metric)
	axis = strings.TrimSpace(axis)
	if axis != SegmentAxisTimeSeconds && axis != SegmentAxisDistanceMeter {
		return nil, fmt.Errorf("%w: segment axis must be time or distance", ErrInvalidSegmentStatsInput)
	}
	keys := []string{axis}
	switch stat {
	case SegmentStatMean, SegmentStatMedian, SegmentStatP90:
		if !validSegmentMetric(metric) {
			return nil, fmt.Errorf("%w: unsupported segment metric %q", ErrInvalidSegmentStatsInput, metric)
		}
		keys = append(keys, metric)
	case SegmentStatDrift:
		if metric != "" {
			return nil, fmt.Errorf("%w: metric is not accepted for drift", ErrInvalidSegmentStatsInput)
		}
		keys = append(keys, SegmentAxisTimeSeconds, SegmentMetricHeartRate)
	case SegmentStatDecoupling:
		if metric != "" {
			return nil, fmt.Errorf("%w: metric is not accepted for decoupling", ErrInvalidSegmentStatsInput)
		}
		keys = append(keys, SegmentAxisTimeSeconds, SegmentMetricHeartRate, SegmentMetricWatts)
	case SegmentStatNP, SegmentStatIF:
		if metric != "" {
			return nil, fmt.Errorf("%w: metric is not accepted for %s", ErrInvalidSegmentStatsInput, stat)
		}
		keys = append(keys, SegmentAxisTimeSeconds, SegmentMetricWatts)
	default:
		return nil, fmt.Errorf("%w: unsupported segment stat %q", ErrInvalidSegmentStatsInput, stat)
	}
	return dedupeStrings(keys), nil
}

func ValidateSegmentStatsInput(input SegmentStatsInput) ([]string, error) {
	stat := strings.TrimSpace(input.Stat)
	metric := strings.TrimSpace(input.Metric)
	bounds := input.Bounds
	bounds.Axis = strings.TrimSpace(bounds.Axis)
	keys, err := RequiredSegmentStreamKeys(stat, metric, bounds.Axis)
	if err != nil {
		return nil, err
	}
	if err := validateBounds(bounds); err != nil {
		return nil, err
	}
	if stat == SegmentStatIF && (!finite(input.FTPWatts) || input.FTPWatts <= 0) {
		return nil, fmt.Errorf("%w: ftp_watts is required and must be positive for if", ErrInvalidSegmentStatsInput)
	}
	if stat != SegmentStatIF && input.FTPWatts != 0 {
		return nil, fmt.Errorf("%w: ftp_watts is accepted only for if", ErrInvalidSegmentStatsInput)
	}
	return keys, nil
}

func ComputeActivitySegmentStats(input SegmentStatsInput) (SegmentStatsResult, error) {
	stat := strings.TrimSpace(input.Stat)
	metric := strings.TrimSpace(input.Metric)
	bounds := input.Bounds
	bounds.Axis = strings.TrimSpace(bounds.Axis)
	keys, err := ValidateSegmentStatsInput(SegmentStatsInput{Stat: stat, Metric: metric, Bounds: bounds, FTPWatts: input.FTPWatts})
	if err != nil {
		return SegmentStatsResult{}, err
	}
	streams, err := requiredStreams(input.Streams, keys)
	if err != nil {
		return SegmentStatsResult{}, err
	}
	indices, err := segmentIndices(streams[bounds.Axis], bounds)
	if err != nil {
		return SegmentStatsResult{}, err
	}
	result := SegmentStatsResult{Stat: stat, Metric: metric, Segment: bounds, StreamsUsed: keys}
	switch stat {
	case SegmentStatMean, SegmentStatMedian, SegmentStatP90:
		return computeScalarSegmentStat(result, stat, metric, streams[metric], indices), nil
	case SegmentStatDrift:
		return computeDrift(result, streams[SegmentAxisTimeSeconds], streams[SegmentMetricHeartRate], indices, bounds), nil
	case SegmentStatDecoupling:
		return computeDecoupling(result, streams[SegmentAxisTimeSeconds], streams[SegmentMetricHeartRate], streams[SegmentMetricWatts], indices, bounds), nil
	case SegmentStatNP:
		return computeNP(result, streams[SegmentAxisTimeSeconds], streams[SegmentMetricWatts], indices), nil
	case SegmentStatIF:
		return computeIF(result, streams[SegmentAxisTimeSeconds], streams[SegmentMetricWatts], indices, input.FTPWatts), nil
	default:
		return SegmentStatsResult{}, fmt.Errorf("%w: unsupported segment stat %q", ErrInvalidSegmentStatsInput, stat)
	}
}

func computeScalarSegmentStat(result SegmentStatsResult, stat string, metric string, values []float64, indices []int) SegmentStatsResult {
	finiteValues := make([]float64, 0, len(indices))
	for _, idx := range indices {
		if finite(values[idx]) {
			finiteValues = append(finiteValues, values[idx])
		}
	}
	result.N = len(finiteValues)
	result.MinSamples = minScalarSamples
	result.Unit = segmentMetricUnit(metric)
	result.Method = stat + " of finite sliced " + metric + " samples"
	result.InsufficientSample = InsufficientSample(result.N, result.MinSamples)
	if result.InsufficientSample {
		return result
	}
	var value float64
	switch stat {
	case SegmentStatMean:
		value = mean(finiteValues)
	case SegmentStatMedian:
		value = median(finiteValues)
	case SegmentStatP90:
		value = nearestRankPercentile(finiteValues, 0.90)
	}
	result.Value = floatPtr(round6(value))
	result.Audit = map[string][]float64{metric: append([]float64(nil), finiteValues...)}
	return result
}

func computeDrift(result SegmentStatsResult, times []float64, heartRate []float64, indices []int, bounds SegmentBounds) SegmentStatsResult {
	first, second := splitHalfPairs(times, indices, bounds, func(idx int) (float64, bool) {
		value := heartRate[idx]
		return value, finite(value)
	})
	result.N = len(first) + len(second)
	result.MinSamples = minSplitHalfSamples * 2
	result.Unit = "%"
	result.Method = "100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half over elapsed-time halves"
	result.FormulaRef = resources.AnalysisFormulaRefHRDrift
	result.InsufficientSample = result.N < result.MinSamples || len(first) < minSplitHalfSamples || len(second) < minSplitHalfSamples
	if result.InsufficientSample {
		return result
	}
	firstAvg := mean(first)
	secondAvg := mean(second)
	if firstAvg <= 0 || secondAvg <= 0 {
		result.InsufficientSample = true
		return result
	}
	value := 100 * (secondAvg - firstAvg) / firstAvg
	result.Value = floatPtr(round6(value))
	result.Details = map[string]float64{"avg_hr_first_half": round6(firstAvg), "avg_hr_second_half": round6(secondAvg)}
	result.Audit = map[string][]float64{"heart_rate_first_half": first, "heart_rate_second_half": second}
	return result
}

func computeDecoupling(result SegmentStatsResult, times []float64, heartRate []float64, watts []float64, indices []int, bounds SegmentBounds) SegmentStatsResult {
	first, second := splitHalfIndices(times, indices, bounds, func(idx int) bool {
		return finite(heartRate[idx]) && finite(watts[idx])
	})
	result.N = len(first) + len(second)
	result.MinSamples = minSplitHalfSamples * 2
	result.Unit = "%"
	result.Method = "100 * ((avg_power_first_half / avg_hr_first_half) - (avg_power_second_half / avg_hr_second_half)) / (avg_power_first_half / avg_hr_first_half) over elapsed-time halves"
	result.FormulaRef = resources.AnalysisFormulaRefPwHRDecoupling
	result.InsufficientSample = result.N < result.MinSamples || len(first) < minSplitHalfSamples || len(second) < minSplitHalfSamples
	if result.InsufficientSample {
		return result
	}
	firstHR, firstPower := pairedMeans(first, heartRate, watts)
	secondHR, secondPower := pairedMeans(second, heartRate, watts)
	if firstHR <= 0 || secondHR <= 0 || firstPower <= 0 {
		result.InsufficientSample = true
		return result
	}
	firstRatio := firstPower / firstHR
	secondRatio := secondPower / secondHR
	if firstRatio <= 0 {
		result.InsufficientSample = true
		return result
	}
	value := 100 * (firstRatio - secondRatio) / firstRatio
	result.Value = floatPtr(round6(value))
	result.Details = map[string]float64{"avg_hr_first_half": round6(firstHR), "avg_hr_second_half": round6(secondHR), "avg_power_first_half": round6(firstPower), "avg_power_second_half": round6(secondPower), "ratio_first": round6(firstRatio), "ratio_second": round6(secondRatio)}
	result.Audit = map[string][]float64{"heart_rate_first_half": valuesAt(first, heartRate), "heart_rate_second_half": valuesAt(second, heartRate), "watts_first_half": valuesAt(first, watts), "watts_second_half": valuesAt(second, watts)}
	return result
}

func computeNP(result SegmentStatsResult, times []float64, watts []float64, indices []int) SegmentStatsResult {
	windows := rollingPowerWindows(times, watts, indices)
	result.N = len(windows)
	result.MinSamples = 1
	result.Unit = "W"
	result.Method = "fourth root of mean fourth power of simple 30-second elapsed-time rolling average watts"
	result.InsufficientSample = InsufficientSample(result.N, result.MinSamples)
	if result.InsufficientSample {
		return result
	}
	value := normalizedPower(windows)
	result.Value = floatPtr(round6(value))
	result.Audit = map[string][]float64{"rolling_30s_avg_watts": windows}
	return result
}

func computeIF(result SegmentStatsResult, times []float64, watts []float64, indices []int, ftpWatts float64) SegmentStatsResult {
	result = computeNP(result, times, watts, indices)
	result.Stat = SegmentStatIF
	result.Unit = "unitless"
	result.Method = "normalized_power / ftp_watts, where normalized_power is fourth root of mean fourth power of simple 30-second elapsed-time rolling average watts"
	if result.InsufficientSample || result.Value == nil {
		return result
	}
	np := *result.Value
	result.Details = map[string]float64{"normalized_power_watts": round6(np), "ftp_watts": round6(ftpWatts)}
	value := np / ftpWatts
	result.Value = floatPtr(round6(value))
	return result
}

func validateBounds(bounds SegmentBounds) error {
	if bounds.Axis != SegmentAxisTimeSeconds && bounds.Axis != SegmentAxisDistanceMeter {
		return fmt.Errorf("%w: segment axis must be time or distance", ErrInvalidSegmentStatsInput)
	}
	if !finite(bounds.Start) || !finite(bounds.End) || bounds.Start < 0 || bounds.End < 0 || bounds.End <= bounds.Start {
		return fmt.Errorf("%w: segment bounds must be finite non-negative values with end greater than start", ErrInvalidSegmentStatsInput)
	}
	return nil
}

func requiredStreams(streams map[string][]float64, keys []string) (map[string][]float64, error) {
	out := make(map[string][]float64, len(keys))
	var expectedLen int
	for _, key := range keys {
		values, ok := streams[key]
		if !ok || len(values) == 0 {
			return nil, MissingSegmentStreamError{Key: key}
		}
		if expectedLen == 0 {
			expectedLen = len(values)
		} else if len(values) != expectedLen {
			return nil, fmt.Errorf("%w: required stream lengths must match", ErrInvalidSegmentStatsInput)
		}
		out[key] = values
	}
	return out, nil
}

func segmentIndices(axis []float64, bounds SegmentBounds) ([]int, error) {
	if len(axis) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingSegmentStream, bounds.Axis)
	}
	minAxis := math.Inf(1)
	maxAxis := math.Inf(-1)
	indices := make([]int, 0, len(axis))
	for i, value := range axis {
		if !finite(value) {
			continue
		}
		if value < minAxis {
			minAxis = value
		}
		if value > maxAxis {
			maxAxis = value
		}
		if value >= bounds.Start && value <= bounds.End {
			indices = append(indices, i)
		}
	}
	if math.IsInf(minAxis, 0) || bounds.Start < minAxis || bounds.End > maxAxis {
		return nil, fmt.Errorf("%w: %.3f..%.3f outside %.3f..%.3f", ErrSegmentOutOfRange, bounds.Start, bounds.End, minAxis, maxAxis)
	}
	return indices, nil
}

func splitHalfPairs(times []float64, indices []int, bounds SegmentBounds, valueAt func(int) (float64, bool)) ([]float64, []float64) {
	if len(indices) == 0 {
		return nil, nil
	}
	start, end := splitHalfTimeEndpoints(times, indices, bounds)
	mid := start + (end-start)/2
	first := []float64{}
	second := []float64{}
	for _, idx := range indices {
		value, ok := valueAt(idx)
		if !ok {
			continue
		}
		if times[idx] < mid {
			first = append(first, value)
		} else {
			second = append(second, value)
		}
	}
	return first, second
}

func splitHalfIndices(times []float64, indices []int, bounds SegmentBounds, valid func(int) bool) ([]int, []int) {
	if len(indices) == 0 {
		return nil, nil
	}
	start, end := splitHalfTimeEndpoints(times, indices, bounds)
	mid := start + (end-start)/2
	first := []int{}
	second := []int{}
	for _, idx := range indices {
		if !valid(idx) {
			continue
		}
		if times[idx] < mid {
			first = append(first, idx)
		} else {
			second = append(second, idx)
		}
	}
	return first, second
}

func splitHalfTimeEndpoints(times []float64, indices []int, bounds SegmentBounds) (float64, float64) {
	if bounds.Axis == SegmentAxisTimeSeconds {
		return bounds.Start, bounds.End
	}
	return times[indices[0]], times[indices[len(indices)-1]]
}

func pairedMeans(indices []int, heartRate []float64, watts []float64) (float64, float64) {
	hrs := make([]float64, 0, len(indices))
	powers := make([]float64, 0, len(indices))
	for _, idx := range indices {
		hrs = append(hrs, heartRate[idx])
		powers = append(powers, watts[idx])
	}
	return mean(hrs), mean(powers)
}

func valuesAt(indices []int, values []float64) []float64 {
	out := make([]float64, 0, len(indices))
	for _, idx := range indices {
		out = append(out, values[idx])
	}
	return out
}

func rollingPowerWindows(times []float64, watts []float64, indices []int) []float64 {
	windows := []float64{}
	if len(indices) == 0 {
		return windows
	}
	for _, endIdx := range indices {
		endTime := times[endIdx]
		if !finite(endTime) || endTime-times[indices[0]] < minNPWindowSeconds {
			continue
		}
		values := []float64{}
		for _, idx := range indices {
			t := times[idx]
			if t <= endTime-minNPWindowSeconds || t > endTime {
				continue
			}
			power := watts[idx]
			if finite(power) && power >= 0 {
				values = append(values, power)
			}
		}
		if len(values) > 0 {
			windows = append(windows, mean(values))
		}
	}
	return windows
}

func normalizedPower(windows []float64) float64 {
	var sum float64
	for _, value := range windows {
		sum += math.Pow(value, 4)
	}
	return math.Pow(sum/float64(len(windows)), 0.25)
}

func mean(values []float64) float64 {
	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func median(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func nearestRankPercentile(values []float64, p float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	rank := int(math.Ceil(p * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

func validSegmentMetric(metric string) bool {
	switch metric {
	case SegmentMetricWatts, SegmentMetricHeartRate, SegmentMetricCadence, SegmentMetricVelocitySmooth, SegmentMetricDistance, SegmentMetricTime:
		return true
	default:
		return false
	}
}

func segmentMetricUnit(metric string) string {
	switch metric {
	case SegmentMetricWatts:
		return "W"
	case SegmentMetricHeartRate:
		return "bpm"
	case SegmentMetricCadence:
		return "rpm"
	case SegmentMetricVelocitySmooth:
		return "m/s"
	case SegmentMetricDistance:
		return "m"
	case SegmentMetricTime:
		return "s"
	default:
		return ""
	}
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func floatPtr(value float64) *float64 {
	return &value
}

func round6(value float64) float64 {
	const factor = 1_000_000.0
	return math.Round(value*factor) / factor
}
