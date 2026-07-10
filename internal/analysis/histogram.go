package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	HistogramMetricPowerWatts   HistogramMetric = "power_watts"
	HistogramMetricHeartRateBPM HistogramMetric = "heart_rate_bpm"
	HistogramMetricPaceSeconds  HistogramMetric = "pace_seconds_per_km"

	BucketMethodConfiguredZones = "configured_zones"
	BucketMethodFixedWidth      = "fixed_width"
)

type HistogramMetric string

type HistogramSample struct {
	Value   float64
	Seconds float64
}

type HistogramBucket struct {
	Label          string   `json:"label"`
	Lower          *float64 `json:"lower"`
	Upper          *float64 `json:"upper"`
	LowerInclusive bool     `json:"lower_inclusive"`
	UpperExclusive bool     `json:"upper_exclusive"`
	Seconds        float64  `json:"seconds"`
	Percentage     float64  `json:"percentage"`
	Unit           string   `json:"unit"`
}

type HistogramZoneConfig struct {
	Sport          string
	SportSettingID int
	Metric         string
	PaceUnits      string
	Boundaries     []float64
	Names          []string
	Unit           string
}

type HistogramZoneSource struct {
	Sport          string    `json:"sport,omitempty"`
	SportSettingID int       `json:"sport_setting_id,omitempty"`
	Metric         string    `json:"metric"`
	PaceUnits      string    `json:"pace_units,omitempty"`
	Boundaries     []float64 `json:"boundaries,omitempty"`
	BoundaryUnit   string    `json:"boundary_unit"`
}

type HistogramFixedWidth struct {
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	BucketCount int     `json:"bucket_count"`
	Width       float64 `json:"width"`
	Unit        string  `json:"unit"`
}

type HistogramResult struct {
	Buckets      []HistogramBucket
	BucketMethod string
	N            int
	Unit         string
	ZoneSource   *HistogramZoneSource
	FixedWidth   *HistogramFixedWidth
}

func ParseHistogramMetric(input string) (HistogramMetric, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "power_watts", "power", "watts":
		return HistogramMetricPowerWatts, nil
	case "heart_rate_bpm", "heart_rate", "heartrate", "hr":
		return HistogramMetricHeartRateBPM, nil
	case "pace_seconds_per_km", "pace", "run_pace":
		return HistogramMetricPaceSeconds, nil
	default:
		return "", InvalidMetricError{Input: input, Hint: "use one of: power_watts, heart_rate_bpm, pace_seconds_per_km"}
	}
}

func HistogramMetricValues() []string {
	return []string{string(HistogramMetricHeartRateBPM), string(HistogramMetricPaceSeconds), string(HistogramMetricPowerWatts)}
}

func HistogramMetricSchemaProperty() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": "Histogram metric enum for a single activity; schemas enumerate canonical values only. Safe aliases such as power, hr, heart_rate, and pace are accepted by the server.",
		"enum":        HistogramMetricValues(),
	}
}

func BuildHistogram(samples []HistogramSample, unit string, zones *HistogramZoneConfig) HistogramResult {
	valid := validHistogramSamples(samples)
	result := HistogramResult{N: len(valid), Unit: strings.TrimSpace(unit)}
	if len(valid) == 0 {
		return result
	}
	if zones != nil {
		buckets, source, ok := configuredZoneBuckets(*zones, result.Unit)
		if ok {
			assignBuckets(buckets, valid)
			finalizeBuckets(buckets, valid)
			result.Buckets = buckets
			result.BucketMethod = BucketMethodConfiguredZones
			result.ZoneSource = source
			return result
		}
	}
	buckets, fixed := fixedWidthBuckets(valid, result.Unit)
	assignBuckets(buckets, valid)
	finalizeBuckets(buckets, valid)
	result.Buckets = buckets
	result.BucketMethod = BucketMethodFixedWidth
	result.FixedWidth = fixed
	return result
}

// ConvertPaceZonePercentage derives a pace-duration boundary from an upstream
// percent-of-threshold pace zone and the stored m/s threshold.
func ConvertPaceZonePercentage(percent float64, thresholdMetersPerSecond float64, emittedUnit string) (float64, bool) {
	if percent <= 0 || thresholdMetersPerSecond <= 0 || !finite(percent) || !finite(thresholdMetersPerSecond) {
		return 0, false
	}
	var distanceMeters float64
	switch emittedUnit {
	case "seconds_per_km":
		distanceMeters = 1000
	case "seconds_per_mile":
		distanceMeters = 1609.344
	default:
		return 0, false
	}
	seconds := distanceMeters * 100 / (thresholdMetersPerSecond * percent)
	if seconds <= 0 || !finite(seconds) {
		return 0, false
	}
	return seconds, true
}

func validHistogramSamples(samples []HistogramSample) []HistogramSample {
	valid := make([]HistogramSample, 0, len(samples))
	for _, sample := range samples {
		if finite(sample.Value) && finite(sample.Seconds) && sample.Seconds > 0 {
			valid = append(valid, sample)
		}
	}
	return valid
}

type zoneBoundary struct {
	value float64
	name  string
}

func configuredZoneBuckets(zones HistogramZoneConfig, unit string) ([]HistogramBucket, *HistogramZoneSource, bool) {
	pairs := make([]zoneBoundary, 0, len(zones.Boundaries))
	for i, boundary := range zones.Boundaries {
		if !finite(boundary) {
			continue
		}
		name := ""
		if i < len(zones.Names) {
			name = strings.TrimSpace(zones.Names[i])
		}
		pairs = append(pairs, zoneBoundary{value: boundary, name: name})
	}
	if len(pairs) == 0 {
		return nil, nil, false
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].value < pairs[j].value })
	buckets := make([]HistogramBucket, 0, len(pairs)+1)
	if pairs[0].value > 0 {
		upper := pairs[0].value
		label := firstNonEmptyString(pairs[0].name, "Zone 1")
		buckets = append(buckets, HistogramBucket{Label: "Below " + label, Upper: &upper, UpperExclusive: true, Unit: unit})
	}
	boundaries := make([]float64, 0, len(pairs))
	for i, pair := range pairs {
		lower := pair.value
		boundaries = append(boundaries, lower)
		bucket := HistogramBucket{Label: firstNonEmptyString(pair.name, fmt.Sprintf("Zone %d", i+1)), Lower: &lower, LowerInclusive: true, Unit: unit}
		if i+1 < len(pairs) {
			upper := pairs[i+1].value
			bucket.Upper = &upper
			bucket.UpperExclusive = true
		}
		buckets = append(buckets, bucket)
	}
	return buckets, &HistogramZoneSource{Sport: strings.TrimSpace(zones.Sport), SportSettingID: zones.SportSettingID, Metric: strings.TrimSpace(zones.Metric), PaceUnits: strings.TrimSpace(zones.PaceUnits), Boundaries: boundaries, BoundaryUnit: strings.TrimSpace(zones.Unit)}, true
}

func fixedWidthBuckets(samples []HistogramSample, unit string) ([]HistogramBucket, *HistogramFixedWidth) {
	minValue, maxValue := samples[0].Value, samples[0].Value
	for _, sample := range samples[1:] {
		if sample.Value < minValue {
			minValue = sample.Value
		}
		if sample.Value > maxValue {
			maxValue = sample.Value
		}
	}
	if minValue == maxValue {
		lower := minValue
		return []HistogramBucket{{Label: formatBucketValue(lower) + " " + unit, Lower: &lower, LowerInclusive: true, Unit: unit}}, &HistogramFixedWidth{Min: round1(minValue), Max: round1(maxValue), BucketCount: 1, Unit: unit}
	}
	const bucketCount = 10
	width := (maxValue - minValue) / bucketCount
	buckets := make([]HistogramBucket, 0, bucketCount)
	for i := 0; i < bucketCount; i++ {
		lower := minValue + float64(i)*width
		bucket := HistogramBucket{Lower: &lower, LowerInclusive: true, Unit: unit}
		if i+1 < bucketCount {
			upper := minValue + float64(i+1)*width
			bucket.Upper = &upper
			bucket.UpperExclusive = true
			bucket.Label = fmt.Sprintf("%s-%s %s", formatBucketValue(lower), formatBucketValue(upper), unit)
		} else {
			bucket.Label = fmt.Sprintf(">= %s %s", formatBucketValue(lower), unit)
		}
		buckets = append(buckets, bucket)
	}
	return buckets, &HistogramFixedWidth{Min: round1(minValue), Max: round1(maxValue), BucketCount: bucketCount, Width: round1(width), Unit: unit}
}

func assignBuckets(buckets []HistogramBucket, samples []HistogramSample) {
	for _, sample := range samples {
		for i := range buckets {
			if bucketContains(buckets[i], sample.Value) {
				buckets[i].Seconds += sample.Seconds
				break
			}
		}
	}
}

func bucketContains(bucket HistogramBucket, value float64) bool {
	if bucket.Lower != nil && value < *bucket.Lower {
		return false
	}
	if bucket.Upper != nil && value >= *bucket.Upper {
		return false
	}
	return true
}

func finalizeBuckets(buckets []HistogramBucket, samples []HistogramSample) {
	total := 0.0
	for _, sample := range samples {
		total += sample.Seconds
	}
	if total <= 0 {
		return
	}
	for i := range buckets {
		seconds := round1(buckets[i].Seconds)
		buckets[i].Seconds = seconds
		buckets[i].Percentage = round1(seconds / total * 100)
		if buckets[i].Lower != nil {
			*buckets[i].Lower = round1(*buckets[i].Lower)
		}
		if buckets[i].Upper != nil {
			*buckets[i].Upper = round1(*buckets[i].Upper)
		}
	}
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func formatBucketValue(value float64) string {
	return fmt.Sprintf("%.1f", round1(value))
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
