package analysis

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// IntervalSource identifies whether activity intervals look structured, manual, or device-generated.
type IntervalSource string

const (
	// IntervalSourceStructuredWorkout identifies intervals with strong structured-workout signals.
	IntervalSourceStructuredWorkout IntervalSource = "structured_workout"
	// IntervalSourceDeviceLaps identifies intervals that look like generic device auto-laps.
	IntervalSourceDeviceLaps IntervalSource = "device_laps"
	// IntervalSourceManualAdded identifies intervals that expose raw upstream rows without auto-detected group markers.
	IntervalSourceManualAdded IntervalSource = "manual_added"
	// IntervalSourceMixed identifies rows containing both grouped auto-detected and ungrouped manual interval evidence.
	IntervalSourceMixed IntervalSource = "mixed"
	// IntervalSourceUnknown identifies intervals without enough source evidence.
	IntervalSourceUnknown IntervalSource = "unknown"
)

// IntervalSourceInput contains the fields needed to classify activity interval rows.
type IntervalSourceInput struct {
	Raw       map[string]any
	Intervals []IntervalSourceInterval
	Groups    []IntervalSourceGroup
}

// IntervalSourceInterval contains classifier-relevant interval fields.
type IntervalSourceInterval struct {
	Name          string
	Type          string
	Label         string
	Raw           map[string]any
	StartIndex    *int
	EndIndex      *int
	StartDistance *float64
	EndDistance   *float64
	Distance      *float64
	Duration      *float64
}

// IntervalSourceGroup contains classifier-relevant interval group fields.
type IntervalSourceGroup struct {
	Name       string
	Type       string
	Raw        map[string]any
	StartIndex *int
	EndIndex   *int
}

// IntervalSourceResult is the interval-source classifier output.
type IntervalSourceResult struct {
	Source           IntervalSource
	AutoLapSuspected bool
}

type intervalSourceMetric int

const (
	intervalMetricDistance intervalSourceMetric = iota
	intervalMetricDuration
)

type intervalSourceTarget struct {
	value     float64
	tolerance float64
	metric    intervalSourceMetric
}

type intervalSourceSample struct {
	value         float64
	startDistance *float64
	endDistance   *float64
	startIndex    *int
	endIndex      *int
}

var genericLapNamePattern = regexp.MustCompile(`^(lap|split|km|kilometer|mile)\s*#?\d*$`)

// InferIntervalSource classifies interval rows using explicit markers, grouped-row evidence, then uniform-lap heuristics.
func InferIntervalSource(input IntervalSourceInput) IntervalSourceResult {
	if hasStructuredSignal(input) {
		return IntervalSourceResult{Source: IntervalSourceStructuredWorkout}
	}
	if hasExplicitDeviceLapSignal(input) {
		return IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true}
	}
	if source, ok := inferGroupedManualSource(input.Intervals); ok {
		return IntervalSourceResult{Source: source}
	}
	if nearUniformAutoLaps(input) {
		return IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true}
	}
	return IntervalSourceResult{Source: IntervalSourceUnknown}
}

func hasStructuredSignal(input IntervalSourceInput) bool {
	if len(input.Groups) > 0 {
		return true
	}
	if rawHasStructuredMarker(input.Raw) {
		return true
	}
	for _, interval := range input.Intervals {
		if rawHasStructuredMarker(interval.Raw) {
			return true
		}
		if isStructuredIntervalText(interval.Name) || isStructuredIntervalText(interval.Type) || isStructuredIntervalText(interval.Label) {
			return true
		}
	}
	for _, group := range input.Groups {
		if rawHasStructuredMarker(group.Raw) || isStructuredIntervalText(group.Name) || isStructuredIntervalText(group.Type) {
			return true
		}
	}
	return false
}

func rawHasStructuredMarker(raw map[string]any) bool {
	for key, value := range raw {
		normalizedKey := normalizeMarkerText(key)
		if normalizedKey == "workoutdoc" || strings.Contains(normalizedKey, "workoutstep") || strings.Contains(normalizedKey, "structuredworkout") {
			return true
		}
		if (strings.Contains(normalizedKey, "intervalsource") || strings.Contains(normalizedKey, "lapsource") || normalizedKey == "source" || normalizedKey == "origin") && isStructuredIntervalText(anyMarkerString(value)) {
			return true
		}
	}
	return false
}

func hasExplicitDeviceLapSignal(input IntervalSourceInput) bool {
	if rawHasDeviceLapMarker(input.Raw) {
		return true
	}
	for _, interval := range input.Intervals {
		if rawHasDeviceLapMarker(interval.Raw) {
			return true
		}
	}
	for _, group := range input.Groups {
		if rawHasDeviceLapMarker(group.Raw) {
			return true
		}
	}
	return false
}

func rawHasDeviceLapMarker(raw map[string]any) bool {
	for key, value := range raw {
		normalizedKey := normalizeMarkerText(key)
		if !isDeviceLapMarkerKey(normalizedKey) {
			continue
		}
		text := normalizeMarkerText(anyMarkerString(value))
		if strings.Contains(text, "autolap") || strings.Contains(text, "devicelap") || strings.Contains(text, "device") || strings.Contains(text, "lap") || text == "auto" || text == "true" {
			return true
		}
	}
	return false
}

func inferGroupedManualSource(intervals []IntervalSourceInterval) (IntervalSource, bool) {
	grouped := 0
	ungrouped := 0
	for _, interval := range intervals {
		if len(interval.Raw) == 0 {
			continue
		}
		if rawHasGroupIDMarker(interval.Raw) {
			grouped++
			continue
		}
		ungrouped++
	}
	if grouped == 0 && ungrouped == 0 {
		return "", false
	}
	if grouped > 0 && ungrouped > 0 {
		return IntervalSourceMixed, true
	}
	if ungrouped > 0 {
		return IntervalSourceManualAdded, true
	}
	return "", false
}

func rawHasGroupIDMarker(raw map[string]any) bool {
	for key, value := range raw {
		if normalizeMarkerText(key) != "groupid" {
			continue
		}
		if rawMarkerPresent(value) {
			return true
		}
	}
	return false
}

func rawMarkerPresent(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case bool:
		return typed
	case int:
		return typed != 0
	case int8:
		return typed != 0
	case int16:
		return typed != 0
	case int32:
		return typed != 0
	case int64:
		return typed != 0
	case uint:
		return typed != 0
	case uint8:
		return typed != 0
	case uint16:
		return typed != 0
	case uint32:
		return typed != 0
	case uint64:
		return typed != 0
	case float32:
		return typed != 0
	case float64:
		return typed != 0
	default:
		return true
	}
}

func isDeviceLapMarkerKey(normalizedKey string) bool {
	return strings.Contains(normalizedKey, "intervalsource") || strings.Contains(normalizedKey, "lapsource") || strings.Contains(normalizedKey, "autolap") || strings.Contains(normalizedKey, "laptype") || normalizedKey == "source" || normalizedKey == "origin"
}

func isStructuredIntervalText(text string) bool {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" || isGenericLapText(text) {
		return false
	}
	structuredTokens := []string{"warm", "cool", "work", "rest", "recover", "tempo", "threshold", "interval", "repeat", "rep", "zone", "z1", "z2", "z3", "z4", "z5", "z6", "z7", "vo2", "endurance", "easy", "hard"}
	for _, token := range structuredTokens {
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func hasNonGenericIntervalText(values ...string) bool {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if !isGenericLapText(trimmed) {
			return true
		}
	}
	return false
}

func isGenericLapText(text string) bool {
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" || text == "lap" || text == "split" {
		return true
	}
	if _, err := strconv.Atoi(text); err == nil {
		return true
	}
	return genericLapNamePattern.MatchString(text)
}

func nearUniformAutoLaps(input IntervalSourceInput) bool {
	for _, interval := range input.Intervals {
		if hasNonGenericIntervalText(interval.Name, interval.Type, interval.Label) {
			return false
		}
	}
	distanceTargets := []intervalSourceTarget{
		{value: 1000, tolerance: math.Max(25, 1000*0.025), metric: intervalMetricDistance},
		{value: 1609.344, tolerance: math.Max(40, 1609.344*0.025), metric: intervalMetricDistance},
	}
	for _, target := range distanceTargets {
		if matchesUniformTarget(samplesForMetric(input.Intervals, intervalMetricDistance), target) {
			return true
		}
	}
	durationTargets := []intervalSourceTarget{
		{value: 60, tolerance: math.Max(5, 60*0.02), metric: intervalMetricDuration},
		{value: 300, tolerance: math.Max(5, 300*0.02), metric: intervalMetricDuration},
		{value: 600, tolerance: math.Max(5, 600*0.02), metric: intervalMetricDuration},
		{value: 900, tolerance: math.Max(5, 900*0.02), metric: intervalMetricDuration},
		{value: 1800, tolerance: math.Max(5, 1800*0.02), metric: intervalMetricDuration},
		{value: 3600, tolerance: math.Max(5, 3600*0.02), metric: intervalMetricDuration},
	}
	for _, target := range durationTargets {
		if matchesUniformTarget(samplesForMetric(input.Intervals, intervalMetricDuration), target) {
			return true
		}
	}
	return false
}

func samplesForMetric(intervals []IntervalSourceInterval, metric intervalSourceMetric) []intervalSourceSample {
	samples := make([]intervalSourceSample, 0, len(intervals))
	for _, interval := range intervals {
		var value float64
		switch metric {
		case intervalMetricDistance:
			value = positiveFloat(interval.Distance)
			if value == 0 && interval.StartDistance != nil && interval.EndDistance != nil {
				delta := *interval.EndDistance - *interval.StartDistance
				if delta > 0 {
					value = delta
				}
			}
		case intervalMetricDuration:
			value = positiveFloat(interval.Duration)
		}
		if value == 0 {
			continue
		}
		samples = append(samples, intervalSourceSample{value: value, startDistance: interval.StartDistance, endDistance: interval.EndDistance, startIndex: interval.StartIndex, endIndex: interval.EndIndex})
	}
	return samples
}

func matchesUniformTarget(samples []intervalSourceSample, target intervalSourceTarget) bool {
	if len(samples) < 4 {
		return false
	}
	for dropFirst := 0; dropFirst <= 1; dropFirst++ {
		for dropLast := 0; dropLast <= 1; dropLast++ {
			if dropFirst+dropLast > 2 || len(samples)-dropFirst-dropLast < 4 {
				continue
			}
			core := samples[dropFirst : len(samples)-dropLast]
			matches := 0
			for _, sample := range core {
				if math.Abs(sample.value-target.value) <= target.tolerance {
					matches++
				}
			}
			if matches < 4 || float64(matches)/float64(len(core)) < 0.8 {
				continue
			}
			if sequenceIsMonotonic(core) {
				return true
			}
		}
	}
	return false
}

func sequenceIsMonotonic(samples []intervalSourceSample) bool {
	if allDistanceRanges(samples) {
		for i := 1; i < len(samples); i++ {
			prevEnd := *samples[i-1].endDistance
			currentStart := *samples[i].startDistance
			currentEnd := *samples[i].endDistance
			if currentEnd <= currentStart || currentStart < prevEnd-5 {
				return false
			}
		}
		return true
	}
	if allIndexRanges(samples) {
		for i := 1; i < len(samples); i++ {
			prevEnd := *samples[i-1].endIndex
			currentStart := *samples[i].startIndex
			currentEnd := *samples[i].endIndex
			if currentEnd <= currentStart || currentStart < prevEnd {
				return false
			}
		}
		return true
	}
	return true
}

func allDistanceRanges(samples []intervalSourceSample) bool {
	for _, sample := range samples {
		if sample.startDistance == nil || sample.endDistance == nil || *sample.endDistance <= *sample.startDistance {
			return false
		}
	}
	return len(samples) > 0
}

func allIndexRanges(samples []intervalSourceSample) bool {
	for _, sample := range samples {
		if sample.startIndex == nil || sample.endIndex == nil || *sample.endIndex <= *sample.startIndex {
			return false
		}
	}
	return len(samples) > 0
}

func positiveFloat(value *float64) float64 {
	if value == nil || *value <= 0 {
		return 0
	}
	return *value
}

func normalizeMarkerText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	replacer := strings.NewReplacer("_", "", "-", "", " ", "")
	return replacer.Replace(text)
}

func anyMarkerString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, anyMarkerString(item))
		}
		return strings.Join(parts, " ")
	case bool:
		return strconv.FormatBool(typed)
	case map[string]any:
		parts := make([]string, 0, len(typed))
		for key, item := range typed {
			parts = append(parts, key, anyMarkerString(item))
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}
