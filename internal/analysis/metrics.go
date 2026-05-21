package analysis

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Metric is a canonical analysis_metric enum value shared by analyzer tools.
type Metric string

// SourceFamily identifies the read surface or derived family that can supply a metric.
type SourceFamily string

const (
	// SourceFitnessDaily identifies daily rows returned by get_fitness.
	SourceFitnessDaily SourceFamily = "fitness_daily"
	// SourceWellnessDaily identifies daily rows returned by get_wellness_data.
	SourceWellnessDaily SourceFamily = "wellness_daily"
	// SourceActivityRow identifies activity rows returned by get_activities/get_activity_details.
	SourceActivityRow SourceFamily = "activity_row"
	// SourceActivityInterval identifies interval rows returned by get_activity_intervals.
	SourceActivityInterval SourceFamily = "activity_interval"
	// SourceTrainingSummary identifies aggregate rows returned by get_training_summary.
	SourceTrainingSummary SourceFamily = "training_summary"
	// SourceExtendedActivity identifies activity metrics returned by get_extended_metrics.
	SourceExtendedActivity SourceFamily = "extended_activity"
	// SourceExtendedInterval identifies interval metrics returned by get_extended_metrics.
	SourceExtendedInterval SourceFamily = "extended_interval"
	// SourceDerivedWeekly identifies analyzer-level weekly aggregates derived from read tools.
	SourceDerivedWeekly SourceFamily = "derived_weekly"
)

// Grain describes the row/window grain for a metric source.
type Grain string

const (
	// GrainDaily is one value per athlete-local day.
	GrainDaily Grain = "daily"
	// GrainActivity is one value per activity.
	GrainActivity Grain = "activity"
	// GrainInterval is one value per activity interval.
	GrainInterval Grain = "interval"
	// GrainSummaryWindow is one value per summary query window.
	GrainSummaryWindow Grain = "summary_window"
	// GrainDerivedWeekly is one value per analyzer-created week bucket.
	GrainDerivedWeekly Grain = "derived_weekly"
)

// MetricKind classifies how analyzer tools should treat a metric source.
type MetricKind string

const (
	// KindScalar is a numeric scalar series.
	KindScalar MetricKind = "scalar"
	// KindSubjectiveScale is a numeric athlete-reported scale with scale metadata.
	KindSubjectiveScale MetricKind = "subjective_scale"
	// KindDerived is an analyzer-level value derived from source fields.
	KindDerived MetricKind = "derived"
)

// MetricSource describes one read-tool source that can supply a canonical metric.
type MetricSource struct {
	Family     SourceFamily
	Tool       string
	Field      string
	Grain      Grain
	Kind       MetricKind
	UnitLabel  string
	ScaleLabel string
	Method     string
}

// InvalidMetricError is returned when analysis_metric parsing rejects an input.
type InvalidMetricError struct {
	Input string
	Hint  string
}

// Error returns a short, user-facing invalid-argument message.
func (e InvalidMetricError) Error() string {
	if strings.TrimSpace(e.Hint) == "" {
		return "invalid analysis_metric"
	}
	return "invalid analysis_metric: " + e.Hint
}

// ParseMetric parses a canonical metric or safe alias into the closed analysis_metric enum.
func ParseMetric(input string) (Metric, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", InvalidMetricError{Input: input, Hint: supportedMetricHint()}
	}
	normalized := normalizeAliasKey(trimmed)
	if metric, ok := aliasToMetric[normalized]; ok {
		return metric, nil
	}
	if looksLikeExpression(normalized) {
		return "", InvalidMetricError{Input: input, Hint: "expressions are not supported; choose a supported metric"}
	}
	return "", InvalidMetricError{Input: input, Hint: unknownMetricHint(normalized)}
}

// MetricValues returns canonical analysis_metric values in deterministic schema order.
func MetricValues() []string {
	values := make([]string, 0, len(metricCatalog))
	for _, entry := range metricCatalog {
		values = append(values, string(entry.metric))
	}
	sort.Strings(values)
	return values
}

// MetricSources returns the source descriptors for a canonical metric or alias.
func MetricSources(metric Metric) []MetricSource {
	parsed, err := ParseMetric(string(metric))
	if err != nil {
		return nil
	}
	for _, entry := range metricCatalog {
		if entry.metric != parsed {
			continue
		}
		sources := make([]MetricSource, len(entry.sources))
		copy(sources, entry.sources)
		return sources
	}
	return nil
}

// MetricSchemaDescription returns concise schema prose for analysis_metric inputs.
func MetricSchemaDescription() string {
	return "analysis_metric enum; choose one canonical value such as ctl, atl, tsb, weekly_tss, hrv, sleep_secs, if, or vi. Safe aliases are accepted by the server, but schemas enumerate canonical values only. Expressions such as ctl/atl are not supported."
}

// MetricSchemaProperty returns a JSON Schema property for analysis_metric.
func MetricSchemaProperty() map[string]any {
	return map[string]any{
		"type":        "string",
		"description": MetricSchemaDescription(),
		"enum":        MetricValues(),
	}
}

// IsInvalidMetric reports whether err is an InvalidMetricError.
func IsInvalidMetric(err error) bool {
	var metricErr InvalidMetricError
	return errors.As(err, &metricErr)
}

type metricEntry struct {
	metric  Metric
	sources []MetricSource
}

var metricCatalog = []metricEntry{
	metric("ctl", src(SourceFitnessDaily, "get_fitness", "ctl", GrainDaily, "CTL"), src(SourceWellnessDaily, "get_wellness_data", "ctl", GrainDaily, "CTL")),
	metric("atl", src(SourceFitnessDaily, "get_fitness", "atl", GrainDaily, "ATL"), src(SourceWellnessDaily, "get_wellness_data", "atl", GrainDaily, "ATL")),
	metric("tsb", src(SourceFitnessDaily, "get_fitness", "tsb", GrainDaily, "TSB")),
	metric("ramp", src(SourceWellnessDaily, "get_wellness_data", "rampRate", GrainDaily, "CTL/week")),
	metric("ctl_load", src(SourceWellnessDaily, "get_wellness_data", "ctlLoad", GrainDaily, "load")),
	metric("atl_load", src(SourceWellnessDaily, "get_wellness_data", "atlLoad", GrainDaily, "load")),
	metric("rhr", src(SourceWellnessDaily, "get_wellness_data", "restingHR", GrainDaily, "bpm")),
	metric("hrv", src(SourceWellnessDaily, "get_wellness_data", "hrv", GrainDaily, "ms rMSSD")),
	metric("hrv_sdnn", src(SourceWellnessDaily, "get_wellness_data", "hrvSDNN", GrainDaily, "ms SDNN")),
	metric("weight", src(SourceWellnessDaily, "get_wellness_data", "weight", GrainDaily, "athlete preferred mass")),
	metric("kcal_consumed", src(SourceWellnessDaily, "get_wellness_data", "kcalConsumed", GrainDaily, "kcal")),
	metric("sleep_secs", src(SourceWellnessDaily, "get_wellness_data", "sleepSecs", GrainDaily, "seconds")),
	metric("sleep_score", src(SourceWellnessDaily, "get_wellness_data", "sleepScore", GrainDaily, "device score")),
	metric("sleep_quality", scaleSrc("sleepQuality", "1-4 subjective sleep quality")),
	metric("avg_sleeping_hr", src(SourceWellnessDaily, "get_wellness_data", "avgSleepingHR", GrainDaily, "bpm")),
	metric("feel", scaleSrc("feel", "1-5 subjective feel"), MetricSource{Family: SourceExtendedActivity, Tool: "get_extended_metrics", Field: "feel", Grain: GrainActivity, Kind: KindSubjectiveScale, UnitLabel: "1-5", ScaleLabel: "1-5 subjective feel"}),
	metric("soreness", scaleSrc("soreness", "1-5 subjective soreness")),
	metric("fatigue", scaleSrc("fatigue", "1-5 subjective fatigue")),
	metric("stress", scaleSrc("stress", "1-5 subjective stress")),
	metric("mood", scaleSrc("mood", "1-5 subjective mood")),
	metric("motivation", scaleSrc("motivation", "1-5 subjective motivation")),
	metric("sp_o2", src(SourceWellnessDaily, "get_wellness_data", "spO2", GrainDaily, "%")),
	metric("systolic", src(SourceWellnessDaily, "get_wellness_data", "systolic", GrainDaily, "mmHg")),
	metric("diastolic", src(SourceWellnessDaily, "get_wellness_data", "diastolic", GrainDaily, "mmHg")),
	metric("hydration", src(SourceWellnessDaily, "get_wellness_data", "hydration", GrainDaily, "unknown")),
	metric("hydration_volume", src(SourceWellnessDaily, "get_wellness_data", "hydrationVolume", GrainDaily, "volume")),
	metric("readiness", src(SourceWellnessDaily, "get_wellness_data", "readiness", GrainDaily, "provider scale")),
	metric("baevsky_si", src(SourceWellnessDaily, "get_wellness_data", "baevskySI", GrainDaily, "index")),
	metric("blood_glucose", src(SourceWellnessDaily, "get_wellness_data", "bloodGlucose", GrainDaily, "glucose")),
	metric("lactate", src(SourceWellnessDaily, "get_wellness_data", "lactate", GrainDaily, "mmol/L")),
	metric("body_fat", src(SourceWellnessDaily, "get_wellness_data", "bodyFat", GrainDaily, "%")),
	metric("abdomen", src(SourceWellnessDaily, "get_wellness_data", "abdomen", GrainDaily, "athlete preferred length")),
	metric("vo2max", src(SourceWellnessDaily, "get_wellness_data", "vo2max", GrainDaily, "ml/kg/min")),
	metric("steps", src(SourceWellnessDaily, "get_wellness_data", "steps", GrainDaily, "steps")),
	metric("respiration", src(SourceWellnessDaily, "get_wellness_data", "respiration", GrainDaily, "breaths/min")),
	metric("carbohydrates", src(SourceWellnessDaily, "get_wellness_data", "carbohydrates", GrainDaily, "g")),
	metric("protein", src(SourceWellnessDaily, "get_wellness_data", "protein", GrainDaily, "g")),
	metric("fat_total", src(SourceWellnessDaily, "get_wellness_data", "fatTotal", GrainDaily, "g")),
	metric("moving_time_seconds", src(SourceActivityRow, "get_activities", "moving_time_seconds", GrainActivity, "seconds"), src(SourceTrainingSummary, "get_training_summary", "moving_time_seconds", GrainSummaryWindow, "seconds")),
	metric("elapsed_time_seconds", src(SourceActivityRow, "get_activities", "elapsed_time_seconds", GrainActivity, "seconds"), src(SourceTrainingSummary, "get_training_summary", "elapsed_time_seconds", GrainSummaryWindow, "seconds")),
	metric("duration_seconds", src(SourceActivityInterval, "get_activity_intervals", "duration_seconds", GrainInterval, "seconds")),
	metric("time_seconds", src(SourceTrainingSummary, "get_training_summary", "time_seconds", GrainSummaryWindow, "seconds")),
	metric("distance_km", src(SourceActivityRow, "get_activities", "distance_km", GrainActivity, "km"), src(SourceTrainingSummary, "get_training_summary", "distance_km", GrainSummaryWindow, "km")),
	metric("distance_mi", src(SourceActivityRow, "get_activities", "distance_mi", GrainActivity, "mi"), src(SourceTrainingSummary, "get_training_summary", "distance_mi", GrainSummaryWindow, "mi")),
	metric("distance_m", src(SourceActivityInterval, "get_activity_intervals", "distance_m", GrainInterval, "m")),
	metric("pace_seconds_per_km", src(SourceActivityRow, "get_activities", "pace_seconds_per_km", GrainActivity, "s/km")),
	metric("pace_seconds_per_mile", src(SourceActivityRow, "get_activities", "pace_seconds_per_mile", GrainActivity, "s/mi")),
	metric("average_speed_kmh", src(SourceActivityRow, "get_activities", "average_speed_kmh", GrainActivity, "km/h")),
	metric("average_speed_mph", src(SourceActivityRow, "get_activities", "average_speed_mph", GrainActivity, "mph")),
	metric("max_speed_kmh", src(SourceActivityRow, "get_activities", "max_speed_kmh", GrainActivity, "km/h")),
	metric("max_speed_mph", src(SourceActivityRow, "get_activities", "max_speed_mph", GrainActivity, "mph")),
	metric("elevation_gain_m", src(SourceActivityRow, "get_activities", "elevation_gain_m", GrainActivity, "m"), src(SourceTrainingSummary, "get_training_summary", "elevation_gain_m", GrainSummaryWindow, "m")),
	metric("elevation_loss_m", src(SourceActivityRow, "get_activities", "elevation_loss_m", GrainActivity, "m")),
	metric("training_load", src(SourceActivityRow, "get_activities", "training_load", GrainActivity, "load"), src(SourceTrainingSummary, "get_training_summary", "training_load", GrainSummaryWindow, "load"), src(SourceExtendedActivity, "get_extended_metrics", "training_load", GrainActivity, "load"), src(SourceExtendedInterval, "get_extended_metrics", "training_load", GrainInterval, "load")),
	metric("average_heart_rate_bpm", src(SourceActivityRow, "get_activities", "average_heart_rate_bpm", GrainActivity, "bpm"), src(SourceActivityInterval, "get_activity_intervals", "average_heart_rate_bpm", GrainInterval, "bpm")),
	metric("max_heart_rate_bpm", src(SourceActivityRow, "get_activities", "max_heart_rate_bpm", GrainActivity, "bpm")),
	metric("average_cadence_rpm", src(SourceActivityRow, "get_activities", "average_cadence_rpm", GrainActivity, "rpm")),
	metric("calories_burned", src(SourceActivityRow, "get_activities", "calories_burned", GrainActivity, "kcal"), src(SourceTrainingSummary, "get_training_summary", "calories_burned", GrainSummaryWindow, "kcal")),
	metric("average_power_watts", src(SourceActivityInterval, "get_activity_intervals", "average_power_watts", GrainInterval, "W")),
	metric("session_rpe", src(SourceTrainingSummary, "get_training_summary", "session_rpe", GrainSummaryWindow, "RPE"), src(SourceExtendedActivity, "get_extended_metrics", "session_rpe", GrainActivity, "RPE")),
	metric("time_in_zones_total_seconds", src(SourceTrainingSummary, "get_training_summary", "time_in_zones_total_seconds", GrainSummaryWindow, "seconds")),
	derivedMetric("weekly_tss", "get_training_summary", "training_load", "TSS-equivalent", "weekly bucketed sum of training_load"),
	derivedMetric("weekly_hours", "get_training_summary", "time_seconds", "hours", "weekly bucketed time_seconds / 3600"),
	metric("stride_length_m", src(SourceExtendedActivity, "get_extended_metrics", "stride_length_m", GrainActivity, "m"), src(SourceExtendedInterval, "get_extended_metrics", "stride_length_m", GrainInterval, "m")),
	metric("cardiac_decoupling_percent", src(SourceExtendedActivity, "get_extended_metrics", "cardiac_decoupling_percent", GrainActivity, "%")),
	metric("pw_hr", src(SourceExtendedActivity, "get_extended_metrics", "pw_hr", GrainActivity, "%")),
	metric("aerobic_decoupling_percent", src(SourceExtendedActivity, "get_extended_metrics", "aerobic_decoupling_percent", GrainActivity, "%"), src(SourceExtendedInterval, "get_extended_metrics", "aerobic_decoupling_percent", GrainInterval, "%")),
	metric("joules_above_ftp_kj", src(SourceExtendedActivity, "get_extended_metrics", "joules_above_ftp_kj", GrainActivity, "kJ"), src(SourceExtendedInterval, "get_extended_metrics", "joules_above_ftp_kj", GrainInterval, "kJ")),
	metric("if", src(SourceExtendedActivity, "get_extended_metrics", "intensity_factor", GrainActivity, "unitless")),
	metric("vi", src(SourceExtendedActivity, "get_extended_metrics", "variability_index", GrainActivity, "unitless")),
	metric("polarization_index", src(SourceExtendedActivity, "get_extended_metrics", "polarization_index", GrainActivity, "unitless")),
	metric("trimp", src(SourceExtendedActivity, "get_extended_metrics", "trimp", GrainActivity, "TRIMP")),
	metric("strain_score", src(SourceExtendedActivity, "get_extended_metrics", "strain_score", GrainActivity, "score"), src(SourceExtendedInterval, "get_extended_metrics", "strain_score", GrainInterval, "score")),
	metric("hr_load", src(SourceExtendedActivity, "get_extended_metrics", "hr_load", GrainActivity, "load")),
	metric("pace_load", src(SourceExtendedActivity, "get_extended_metrics", "pace_load", GrainActivity, "load")),
	metric("power_load", src(SourceExtendedActivity, "get_extended_metrics", "power_load", GrainActivity, "load")),
	metric("left_right_balance_percent", src(SourceExtendedActivity, "get_extended_metrics", "left_right_balance_percent", GrainActivity, "%"), src(SourceExtendedInterval, "get_extended_metrics", "left_right_balance_percent", GrainInterval, "%")),
	metric("rpe", src(SourceExtendedActivity, "get_extended_metrics", "rpe", GrainActivity, "RPE")),
	metric("compliance_pct", src(SourceExtendedActivity, "get_extended_metrics", "compliance_percent", GrainActivity, "%")),
	metric("dfa_alpha1", src(SourceExtendedInterval, "get_extended_metrics", "dfa_alpha1", GrainInterval, "unitless")),
	metric("w_prime_balance_start_kj", src(SourceExtendedInterval, "get_extended_metrics", "w_prime_balance_start_kj", GrainInterval, "kJ")),
	metric("w_prime_balance_end_kj", src(SourceExtendedInterval, "get_extended_metrics", "w_prime_balance_end_kj", GrainInterval, "kJ")),
}

var aliasToMetric = buildAliasMap()

func metric(value string, sources ...MetricSource) metricEntry {
	return metricEntry{metric: Metric(value), sources: sources}
}

func src(family SourceFamily, tool string, field string, grain Grain, unit string) MetricSource {
	return MetricSource{Family: family, Tool: tool, Field: field, Grain: grain, Kind: KindScalar, UnitLabel: unit}
}

func scaleSrc(field string, scale string) MetricSource {
	return MetricSource{Family: SourceWellnessDaily, Tool: "get_wellness_data", Field: field, Grain: GrainDaily, Kind: KindSubjectiveScale, UnitLabel: scale, ScaleLabel: scale}
}

func derivedMetric(value string, tool string, field string, unit string, method string) metricEntry {
	return metric(value, MetricSource{Family: SourceDerivedWeekly, Tool: tool, Field: field, Grain: GrainDerivedWeekly, Kind: KindDerived, UnitLabel: unit, Method: method})
}

func buildAliasMap() map[string]Metric {
	aliases := make(map[string]Metric, len(metricCatalog)*2)
	for _, entry := range metricCatalog {
		aliases[normalizeAliasKey(string(entry.metric))] = entry.metric
	}
	for alias, metric := range explicitAliases() {
		aliases[normalizeAliasKey(alias)] = metric
	}
	return aliases
}

func explicitAliases() map[string]Metric {
	return map[string]Metric{
		"resting_hr":            "rhr",
		"restingHR":             "rhr",
		"resting_heart_rate":    "rhr",
		"restingHeartRate":      "rhr",
		"sleepSecs":             "sleep_secs",
		"sleep_seconds":         "sleep_secs",
		"sleepScore":            "sleep_score",
		"sleepQuality":          "sleep_quality",
		"hrvSDNN":               "hrv_sdnn",
		"hrv_sdnn_ms":           "hrv_sdnn",
		"intensity_factor":      "if",
		"variability_index":     "vi",
		"compliance_percent":    "compliance_pct",
		"compliance_percentage": "compliance_pct",
		"rampRate":              "ramp",
		"ramp_rate":             "ramp",
		"spO2":                  "sp_o2",
		"baevskySI":             "baevsky_si",
		"kcalConsumed":          "kcal_consumed",
		"avgSleepingHR":         "avg_sleeping_hr",
		"hydrationVolume":       "hydration_volume",
		"bloodGlucose":          "blood_glucose",
		"bodyFat":               "body_fat",
		"fatTotal":              "fat_total",
		"wPrimeBalanceStartKJ":  "w_prime_balance_start_kj",
		"wPrimeBalanceEndKJ":    "w_prime_balance_end_kj",
		"w_prime_balance_start": "w_prime_balance_start_kj",
		"w_prime_balance_end":   "w_prime_balance_end_kj",
		"left_right_balance":    "left_right_balance_percent",
		"cardiac_decoupling":    "cardiac_decoupling_percent",
		"aerobic_decoupling":    "aerobic_decoupling_percent",
		"compliance":            "compliance_pct",
	}
}

func normalizeAliasKey(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func looksLikeExpression(input string) bool {
	trimmed := strings.TrimSpace(input)
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "_per_") || strings.Contains(lower, " per ") {
		return true
	}
	return strings.ContainsAny(trimmed, "+*/^=<>():,|[]{}") || strings.Contains(trimmed, " - ")
}

func unknownMetricHint(normalized string) string {
	switch {
	case containsAny(normalized, "5min", "5_min", "20min", "20_min", "5k", "best_effort", "best-effort", "power_curve", "pace_curve"):
		return "try analyze_efforts_delta for best-effort durations/distances"
	case containsAny(normalized, "zone", "polarized", "polarization_zone", "load_balance"):
		return "try compute_zone_time or compute_load_balance for zone distributions"
	case containsAny(normalized, "segment", "stream", "hr_drift", "mean_power", "median_power", "p90"):
		return "try compute_activity_segment_stats for within-activity stream stats"
	case containsAny(normalized, "adherence", "compliance_rate", "completed_vs_scheduled"):
		return "try compute_compliance_rate for scheduled-vs-completed analysis"
	default:
		return supportedMetricHint()
	}
}

func supportedMetricHint() string {
	return "use one of: ctl, atl, tsb, weekly_tss, hrv, sleep_secs, if, vi"
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

// ValidateMetricCatalog returns an error if the static metric catalog is internally inconsistent.
func ValidateMetricCatalog() error {
	seen := map[Metric]struct{}{}
	for _, entry := range metricCatalog {
		if strings.TrimSpace(string(entry.metric)) == "" {
			return fmt.Errorf("analysis metric catalog contains an empty metric")
		}
		if _, ok := seen[entry.metric]; ok {
			return fmt.Errorf("analysis metric catalog contains duplicate metric %q", entry.metric)
		}
		seen[entry.metric] = struct{}{}
		if len(entry.sources) == 0 {
			return fmt.Errorf("analysis metric %q has no sources", entry.metric)
		}
	}
	values := MetricValues()
	if !sort.StringsAreSorted(append([]string(nil), values...)) {
		return fmt.Errorf("analysis metric catalog must be sorted for stable schemas")
	}
	return nil
}
