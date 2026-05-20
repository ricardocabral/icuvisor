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
	if looksLikeExpression(trimmed) {
		return "", InvalidMetricError{Input: input, Hint: "expressions are not supported; choose a supported metric"}
	}
	normalized := normalizeAliasKey(trimmed)
	if metric, ok := aliasToMetric[normalized]; ok {
		return metric, nil
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
	return "analysis_metric enum; choose one canonical value such as ctl, atl, tsb, weekly_tss, hrv, sleep_secs, if, or vi. Aliases are accepted by the server, but schemas enumerate canonical values only. Expressions such as ctl/atl are not supported."
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
	metric("ctl", src(SourceFitnessDaily, "get_fitness", "ctl", GrainDaily, KindScalar, "CTL"), src(SourceWellnessDaily, "get_wellness_data", "ctl", GrainDaily, KindScalar, "CTL")),
	metric("atl", src(SourceFitnessDaily, "get_fitness", "atl", GrainDaily, KindScalar, "ATL"), src(SourceWellnessDaily, "get_wellness_data", "atl", GrainDaily, KindScalar, "ATL")),
	metric("tsb", src(SourceFitnessDaily, "get_fitness", "tsb", GrainDaily, KindScalar, "TSB")),
	metric("ramp", src(SourceWellnessDaily, "get_wellness_data", "rampRate", GrainDaily, KindScalar, "CTL/week")),
	metric("ctl_load", src(SourceWellnessDaily, "get_wellness_data", "ctlLoad", GrainDaily, KindScalar, "load")),
	metric("atl_load", src(SourceWellnessDaily, "get_wellness_data", "atlLoad", GrainDaily, KindScalar, "load")),
	metric("rhr", src(SourceWellnessDaily, "get_wellness_data", "restingHR", GrainDaily, KindScalar, "bpm")),
	metric("hrv", src(SourceWellnessDaily, "get_wellness_data", "hrv", GrainDaily, KindScalar, "ms rMSSD")),
	metric("hrv_sdnn", src(SourceWellnessDaily, "get_wellness_data", "hrvSDNN", GrainDaily, KindScalar, "ms SDNN")),
	metric("weight", src(SourceWellnessDaily, "get_wellness_data", "weight", GrainDaily, KindScalar, "athlete preferred mass")),
	metric("kcal_consumed", src(SourceWellnessDaily, "get_wellness_data", "kcalConsumed", GrainDaily, KindScalar, "kcal")),
	metric("sleep_secs", src(SourceWellnessDaily, "get_wellness_data", "sleepSecs", GrainDaily, KindScalar, "seconds")),
	metric("sleep_score", src(SourceWellnessDaily, "get_wellness_data", "sleepScore", GrainDaily, KindScalar, "device score")),
	metric("sleep_quality", scaleSrc("get_wellness_data", "sleepQuality", "1-4 subjective sleep quality")),
	metric("avg_sleeping_hr", src(SourceWellnessDaily, "get_wellness_data", "avgSleepingHR", GrainDaily, KindScalar, "bpm")),
	metric("feel", scaleSrc("get_wellness_data", "feel", "1-5 subjective feel"), MetricSource{Family: SourceExtendedActivity, Tool: "get_extended_metrics", Field: "feel", Grain: GrainActivity, Kind: KindSubjectiveScale, UnitLabel: "1-5", ScaleLabel: "1-5 subjective feel"}),
	metric("soreness", scaleSrc("get_wellness_data", "soreness", "1-5 subjective soreness")),
	metric("fatigue", scaleSrc("get_wellness_data", "fatigue", "1-5 subjective fatigue")),
	metric("stress", scaleSrc("get_wellness_data", "stress", "1-5 subjective stress")),
	metric("mood", scaleSrc("get_wellness_data", "mood", "1-5 subjective mood")),
	metric("motivation", scaleSrc("get_wellness_data", "motivation", "1-5 subjective motivation")),
	metric("sp_o2", src(SourceWellnessDaily, "get_wellness_data", "spO2", GrainDaily, KindScalar, "%")),
	metric("systolic", src(SourceWellnessDaily, "get_wellness_data", "systolic", GrainDaily, KindScalar, "mmHg")),
	metric("diastolic", src(SourceWellnessDaily, "get_wellness_data", "diastolic", GrainDaily, KindScalar, "mmHg")),
	metric("hydration", src(SourceWellnessDaily, "get_wellness_data", "hydration", GrainDaily, KindScalar, "unknown")),
	metric("hydration_volume", src(SourceWellnessDaily, "get_wellness_data", "hydrationVolume", GrainDaily, KindScalar, "volume")),
	metric("readiness", src(SourceWellnessDaily, "get_wellness_data", "readiness", GrainDaily, KindScalar, "provider scale")),
	metric("baevsky_si", src(SourceWellnessDaily, "get_wellness_data", "baevskySI", GrainDaily, KindScalar, "index")),
	metric("blood_glucose", src(SourceWellnessDaily, "get_wellness_data", "bloodGlucose", GrainDaily, KindScalar, "glucose")),
	metric("lactate", src(SourceWellnessDaily, "get_wellness_data", "lactate", GrainDaily, KindScalar, "mmol/L")),
	metric("body_fat", src(SourceWellnessDaily, "get_wellness_data", "bodyFat", GrainDaily, KindScalar, "%")),
	metric("abdomen", src(SourceWellnessDaily, "get_wellness_data", "abdomen", GrainDaily, KindScalar, "athlete preferred length")),
	metric("vo2max", src(SourceWellnessDaily, "get_wellness_data", "vo2max", GrainDaily, KindScalar, "ml/kg/min")),
	metric("steps", src(SourceWellnessDaily, "get_wellness_data", "steps", GrainDaily, KindScalar, "steps")),
	metric("respiration", src(SourceWellnessDaily, "get_wellness_data", "respiration", GrainDaily, KindScalar, "breaths/min")),
	metric("carbohydrates", src(SourceWellnessDaily, "get_wellness_data", "carbohydrates", GrainDaily, KindScalar, "g")),
	metric("protein", src(SourceWellnessDaily, "get_wellness_data", "protein", GrainDaily, KindScalar, "g")),
	metric("fat_total", src(SourceWellnessDaily, "get_wellness_data", "fatTotal", GrainDaily, KindScalar, "g")),
	metric("moving_time_seconds", src(SourceActivityRow, "get_activities", "moving_time_seconds", GrainActivity, KindScalar, "seconds"), src(SourceTrainingSummary, "get_training_summary", "moving_time_seconds", GrainSummaryWindow, KindScalar, "seconds")),
	metric("elapsed_time_seconds", src(SourceActivityRow, "get_activities", "elapsed_time_seconds", GrainActivity, KindScalar, "seconds"), src(SourceTrainingSummary, "get_training_summary", "elapsed_time_seconds", GrainSummaryWindow, KindScalar, "seconds")),
	metric("duration_seconds", src(SourceActivityInterval, "get_activity_intervals", "duration_seconds", GrainInterval, KindScalar, "seconds")),
	metric("time_seconds", src(SourceTrainingSummary, "get_training_summary", "time_seconds", GrainSummaryWindow, KindScalar, "seconds")),
	metric("distance_km", src(SourceActivityRow, "get_activities", "distance_km", GrainActivity, KindScalar, "km"), src(SourceTrainingSummary, "get_training_summary", "distance_km", GrainSummaryWindow, KindScalar, "km")),
	metric("distance_mi", src(SourceActivityRow, "get_activities", "distance_mi", GrainActivity, KindScalar, "mi"), src(SourceTrainingSummary, "get_training_summary", "distance_mi", GrainSummaryWindow, KindScalar, "mi")),
	metric("distance_m", src(SourceActivityInterval, "get_activity_intervals", "distance_m", GrainInterval, KindScalar, "m")),
	metric("pace_seconds_per_km", src(SourceActivityRow, "get_activities", "pace_seconds_per_km", GrainActivity, KindScalar, "s/km")),
	metric("pace_seconds_per_mile", src(SourceActivityRow, "get_activities", "pace_seconds_per_mile", GrainActivity, KindScalar, "s/mi")),
	metric("average_speed_kmh", src(SourceActivityRow, "get_activities", "average_speed_kmh", GrainActivity, KindScalar, "km/h")),
	metric("average_speed_mph", src(SourceActivityRow, "get_activities", "average_speed_mph", GrainActivity, KindScalar, "mph")),
	metric("max_speed_kmh", src(SourceActivityRow, "get_activities", "max_speed_kmh", GrainActivity, KindScalar, "km/h")),
	metric("max_speed_mph", src(SourceActivityRow, "get_activities", "max_speed_mph", GrainActivity, KindScalar, "mph")),
	metric("elevation_gain_m", src(SourceActivityRow, "get_activities", "elevation_gain_m", GrainActivity, KindScalar, "m"), src(SourceTrainingSummary, "get_training_summary", "elevation_gain_m", GrainSummaryWindow, KindScalar, "m")),
	metric("elevation_loss_m", src(SourceActivityRow, "get_activities", "elevation_loss_m", GrainActivity, KindScalar, "m")),
	metric("training_load", src(SourceActivityRow, "get_activities", "training_load", GrainActivity, KindScalar, "load"), src(SourceTrainingSummary, "get_training_summary", "training_load", GrainSummaryWindow, KindScalar, "load"), src(SourceExtendedActivity, "get_extended_metrics", "training_load", GrainActivity, KindScalar, "load"), src(SourceExtendedInterval, "get_extended_metrics", "training_load", GrainInterval, KindScalar, "load")),
	metric("average_heart_rate_bpm", src(SourceActivityRow, "get_activities", "average_heart_rate_bpm", GrainActivity, KindScalar, "bpm"), src(SourceActivityInterval, "get_activity_intervals", "average_heart_rate_bpm", GrainInterval, KindScalar, "bpm")),
	metric("max_heart_rate_bpm", src(SourceActivityRow, "get_activities", "max_heart_rate_bpm", GrainActivity, KindScalar, "bpm")),
	metric("average_cadence_rpm", src(SourceActivityRow, "get_activities", "average_cadence_rpm", GrainActivity, KindScalar, "rpm")),
	metric("calories_burned", src(SourceActivityRow, "get_activities", "calories_burned", GrainActivity, KindScalar, "kcal"), src(SourceTrainingSummary, "get_training_summary", "calories_burned", GrainSummaryWindow, KindScalar, "kcal")),
	metric("average_power_watts", src(SourceActivityInterval, "get_activity_intervals", "average_power_watts", GrainInterval, KindScalar, "W")),
	metric("session_rpe", src(SourceTrainingSummary, "get_training_summary", "session_rpe", GrainSummaryWindow, KindScalar, "RPE"), src(SourceExtendedActivity, "get_extended_metrics", "session_rpe", GrainActivity, KindScalar, "RPE")),
	metric("time_in_zones_total_seconds", src(SourceTrainingSummary, "get_training_summary", "time_in_zones_total_seconds", GrainSummaryWindow, KindScalar, "seconds")),
	derivedMetric("weekly_tss", "get_training_summary", "training_load", "TSS-equivalent", "weekly bucketed sum of training_load"),
	derivedMetric("weekly_hours", "get_training_summary", "time_seconds", "hours", "weekly bucketed time_seconds / 3600"),
	metric("stride_length_m", src(SourceExtendedActivity, "get_extended_metrics", "stride_length_m", GrainActivity, KindScalar, "m"), src(SourceExtendedInterval, "get_extended_metrics", "stride_length_m", GrainInterval, KindScalar, "m")),
	metric("cardiac_decoupling_percent", src(SourceExtendedActivity, "get_extended_metrics", "cardiac_decoupling_percent", GrainActivity, KindScalar, "%")),
	metric("pw_hr", src(SourceExtendedActivity, "get_extended_metrics", "pw_hr", GrainActivity, KindScalar, "%")),
	metric("aerobic_decoupling_percent", src(SourceExtendedActivity, "get_extended_metrics", "aerobic_decoupling_percent", GrainActivity, KindScalar, "%"), src(SourceExtendedInterval, "get_extended_metrics", "aerobic_decoupling_percent", GrainInterval, KindScalar, "%")),
	metric("joules_above_ftp_kj", src(SourceExtendedActivity, "get_extended_metrics", "joules_above_ftp_kj", GrainActivity, KindScalar, "kJ"), src(SourceExtendedInterval, "get_extended_metrics", "joules_above_ftp_kj", GrainInterval, KindScalar, "kJ")),
	metric("if", src(SourceExtendedActivity, "get_extended_metrics", "intensity_factor", GrainActivity, KindScalar, "unitless")),
	metric("vi", src(SourceExtendedActivity, "get_extended_metrics", "variability_index", GrainActivity, KindScalar, "unitless")),
	metric("polarization_index", src(SourceExtendedActivity, "get_extended_metrics", "polarization_index", GrainActivity, KindScalar, "unitless")),
	metric("trimp", src(SourceExtendedActivity, "get_extended_metrics", "trimp", GrainActivity, KindScalar, "TRIMP")),
	metric("strain_score", src(SourceExtendedActivity, "get_extended_metrics", "strain_score", GrainActivity, KindScalar, "score"), src(SourceExtendedInterval, "get_extended_metrics", "strain_score", GrainInterval, KindScalar, "score")),
	metric("hr_load", src(SourceExtendedActivity, "get_extended_metrics", "hr_load", GrainActivity, KindScalar, "load")),
	metric("pace_load", src(SourceExtendedActivity, "get_extended_metrics", "pace_load", GrainActivity, KindScalar, "load")),
	metric("power_load", src(SourceExtendedActivity, "get_extended_metrics", "power_load", GrainActivity, KindScalar, "load")),
	metric("left_right_balance_percent", src(SourceExtendedActivity, "get_extended_metrics", "left_right_balance_percent", GrainActivity, KindScalar, "%"), src(SourceExtendedInterval, "get_extended_metrics", "left_right_balance_percent", GrainInterval, KindScalar, "%")),
	metric("rpe", src(SourceExtendedActivity, "get_extended_metrics", "rpe", GrainActivity, KindScalar, "RPE")),
	metric("compliance_pct", src(SourceExtendedActivity, "get_extended_metrics", "compliance_percent", GrainActivity, KindScalar, "%")),
	metric("dfa_alpha1", src(SourceExtendedInterval, "get_extended_metrics", "dfa_alpha1", GrainInterval, KindScalar, "unitless")),
	metric("w_prime_balance_start_kj", src(SourceExtendedInterval, "get_extended_metrics", "w_prime_balance_start_kj", GrainInterval, KindScalar, "kJ")),
	metric("w_prime_balance_end_kj", src(SourceExtendedInterval, "get_extended_metrics", "w_prime_balance_end_kj", GrainInterval, KindScalar, "kJ")),
}

var aliasToMetric = buildAliasMap()

func metric(value string, sources ...MetricSource) metricEntry {
	return metricEntry{metric: Metric(value), sources: sources}
}

func src(family SourceFamily, tool string, field string, grain Grain, kind MetricKind, unit string) MetricSource {
	return MetricSource{Family: family, Tool: tool, Field: field, Grain: grain, Kind: kind, UnitLabel: unit}
}

func scaleSrc(tool string, field string, scale string) MetricSource {
	return MetricSource{Family: SourceWellnessDaily, Tool: tool, Field: field, Grain: GrainDaily, Kind: KindSubjectiveScale, UnitLabel: scale, ScaleLabel: scale}
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
	if strings.Contains(trimmed, "_per_") || strings.Contains(strings.ToLower(trimmed), " per ") {
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
