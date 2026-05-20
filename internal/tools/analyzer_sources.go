package tools

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const analyzerActivityWindowTooLargeMessage = "activity window too large; narrow date range"

type analyzerMetricSelection struct {
	Metric analysis.Metric
	Source analysis.MetricSource
}

type analyzerSampleSeries struct {
	Metric      analysis.Metric
	Unit        string
	ScaleLabel  string
	Samples     []analysis.NumericSample
	MissingDays int
	SourceTools []string
	Assumptions map[string]any
}

func selectAnalyzerMetricSource(metric analysis.Metric, grain analysis.SampleGrain, allowWeekly bool) (analyzerMetricSelection, error) {
	sources := analysis.MetricSources(metric)
	if len(sources) == 0 {
		return analyzerMetricSelection{}, fmt.Errorf("unsupported analysis metric %q", metric)
	}
	families := sourcePreference(grain, allowWeekly)
	for _, family := range families {
		for _, source := range sources {
			if source.Family == family {
				return analyzerMetricSelection{Metric: metric, Source: source}, nil
			}
		}
	}
	return analyzerMetricSelection{}, analyzerUnsupportedMetricError(sources)
}

func sourcePreference(grain analysis.SampleGrain, allowWeekly bool) []analysis.SourceFamily {
	if grain == analysis.SampleGrainActivity {
		return []analysis.SourceFamily{analysis.SourceActivityRow}
	}
	families := []analysis.SourceFamily{analysis.SourceFitnessDaily, analysis.SourceWellnessDaily, analysis.SourceTrainingSummary, analysis.SourceActivityRow}
	if allowWeekly {
		families = append(families, analysis.SourceDerivedWeekly)
	}
	return families
}

func analyzerUnsupportedMetricError(sources []analysis.MetricSource) error {
	for _, source := range sources {
		if source.Family == analysis.SourceActivityInterval || source.Family == analysis.SourceExtendedInterval {
			return errors.New("metric is interval-only; use get_activity_intervals or compute_activity_segment_stats")
		}
		if source.Family == analysis.SourceExtendedActivity {
			return errors.New("metric requires per-activity extended metrics; use get_extended_metrics")
		}
	}
	return errors.New("metric is not supported by this analyzer")
}

func loadAllAnalyzerActivities(ctx context.Context, client ActivitiesClient, oldest string, newest string) ([]intervals.Activity, error) {
	args := GetActivitiesRequest{Oldest: oldest, Newest: newest, IncludeUnnamed: true, PageSize: maxActivitiesPageSize}
	var token *activitiesPageToken
	var out []intervals.Activity
	seenTokens := map[string]bool{}
	for pages := 0; pages < 100; pages++ {
		rows, nextToken, err := fetchActivitiesPage(ctx, client, args, token, "")
		if err != nil {
			if errors.Is(err, errActivitiesPaginationBoundary) {
				return nil, errors.New(analyzerActivityWindowTooLargeMessage)
			}
			return nil, err
		}
		out = append(out, rows...)
		trimmedToken := strings.TrimSpace(nextToken)
		if trimmedToken == "" {
			return out, nil
		}
		if seenTokens[trimmedToken] {
			return nil, errors.New(analyzerActivityWindowTooLargeMessage)
		}
		seenTokens[trimmedToken] = true
		token, err = parseActivitiesPageToken(trimmedToken)
		if err != nil {
			return nil, err
		}
		if token == nil {
			return nil, errors.New(analyzerActivityWindowTooLargeMessage)
		}
	}
	return nil, errors.New(analyzerActivityWindowTooLargeMessage)
}

func activityMetricValue(activity intervals.Activity, metric analysis.Metric, unitSystem response.UnitSystem) (float64, bool) {
	switch metric {
	case "moving_time_seconds":
		return intPointerValue(activity.MovingTime)
	case "elapsed_time_seconds":
		return intPointerValue(activity.ElapsedTime)
	case "distance_km":
		return convertedDistanceValue(activity, unitSystem, units.UnitKM)
	case "distance_mi":
		return convertedDistanceValue(activity, unitSystem, units.UnitMI)
	case "pace_seconds_per_km":
		return activityPace(activity, units.UnitKM)
	case "pace_seconds_per_mile":
		return activityPace(activity, units.UnitMI)
	case "average_speed_kmh":
		return activitySpeed(activity.AverageSpeed, units.UnitKMH)
	case "average_speed_mph":
		return activitySpeed(activity.AverageSpeed, units.UnitMPH)
	case "max_speed_kmh":
		return activitySpeed(activity.MaxSpeed, units.UnitKMH)
	case "max_speed_mph":
		return activitySpeed(activity.MaxSpeed, units.UnitMPH)
	case "elevation_gain_m":
		return floatPointerValue(activity.TotalElevationGain)
	case "elevation_loss_m":
		return floatPointerValue(activity.TotalElevationLoss)
	case "training_load":
		return intPointerValue(activity.TrainingLoad)
	case "average_heart_rate_bpm":
		return intPointerValue(activity.AverageHeartRate)
	case "max_heart_rate_bpm":
		return intPointerValue(activity.MaxHeartRate)
	case "average_cadence_rpm":
		return floatPointerValue(activity.AverageCadence)
	case "calories_burned":
		return intPointerValue(activity.Calories)
	default:
		return 0, false
	}
}

func summaryMetricValue(row intervals.SummaryWithCats, metric analysis.Metric, unitSystem response.UnitSystem) (float64, bool) {
	switch metric {
	case "ctl":
		return row.Fitness, true
	case "atl":
		return row.Fatigue, true
	case "tsb":
		return row.Form, true
	case "moving_time_seconds":
		return float64(row.MovingTime), row.MovingTime != 0
	case "elapsed_time_seconds":
		return float64(row.ElapsedTime), row.ElapsedTime != 0
	case "time_seconds":
		return float64(row.Time), row.Time != 0
	case "distance_km", "distance_mi":
		converted := response.ToPreferred(row.Distance, units.UnitM, unitSystem)
		if metric == "distance_km" {
			return response.ToPreferred(row.Distance, units.UnitM, response.UnitSystemMetric).Value, row.Distance != 0
		}
		return response.ToPreferred(row.Distance, units.UnitM, response.UnitSystemImperial).Value, converted.Value != 0
	case "elevation_gain_m":
		return row.TotalElevationGain, row.TotalElevationGain != 0
	case "training_load", "weekly_tss":
		return float64(row.TrainingLoad), row.TrainingLoad != 0
	case "weekly_hours":
		return float64(row.Time) / 3600, row.Time != 0
	case "calories_burned":
		return float64(row.Calories), row.Calories != 0
	case "session_rpe":
		return float64(row.SRPE), row.SRPE != 0
	case "time_in_zones_total_seconds":
		return float64(row.TimeInZonesTot), row.TimeInZonesTot != 0
	default:
		return 0, false
	}
}

func wellnessMetricValue(row intervals.Wellness, metric analysis.Metric) (float64, bool) {
	switch metric {
	case "ctl":
		return floatPointerValue(row.CTL)
	case "atl":
		return floatPointerValue(row.ATL)
	case "ramp":
		return floatPointerValue(row.RampRate)
	case "ctl_load":
		return floatPointerValue(row.CTLLoad)
	case "atl_load":
		return floatPointerValue(row.ATLLoad)
	case "rhr":
		return intPointerValue(row.RestingHR)
	case "hrv":
		return floatPointerValue(row.HRV)
	case "hrv_sdnn":
		return floatPointerValue(row.HRVSDNN)
	case "weight":
		return floatPointerValue(row.Weight)
	case "kcal_consumed":
		return intPointerValue(row.KcalConsumed)
	case "sleep_secs":
		return intPointerValue(row.SleepSecs)
	case "sleep_score":
		return floatPointerValue(row.SleepScore)
	case "sleep_quality":
		return intPointerValue(row.SleepQuality)
	case "avg_sleeping_hr":
		return floatPointerValue(row.AvgSleepingHR)
	case "feel":
		return intPointerValue(row.Feel)
	case "soreness":
		return intPointerValue(row.Soreness)
	case "fatigue":
		return intPointerValue(row.Fatigue)
	case "stress":
		return intPointerValue(row.Stress)
	case "mood":
		return intPointerValue(row.Mood)
	case "motivation":
		return intPointerValue(row.Motivation)
	case "sp_o2":
		return floatPointerValue(row.SpO2)
	case "systolic":
		return intPointerValue(row.Systolic)
	case "diastolic":
		return intPointerValue(row.Diastolic)
	case "hydration":
		return intPointerValue(row.Hydration)
	case "hydration_volume":
		return floatPointerValue(row.HydrationVolume)
	case "readiness":
		return floatPointerValue(row.Readiness)
	case "baevsky_si":
		return floatPointerValue(row.BaevskySI)
	case "blood_glucose":
		return floatPointerValue(row.BloodGlucose)
	case "lactate":
		return floatPointerValue(row.Lactate)
	case "body_fat":
		return floatPointerValue(row.BodyFat)
	case "abdomen":
		return floatPointerValue(row.Abdomen)
	case "vo2max":
		return floatPointerValue(row.VO2Max)
	case "steps":
		return intPointerValue(row.Steps)
	case "respiration":
		return floatPointerValue(row.Respiration)
	case "carbohydrates":
		return floatPointerValue(row.Carbohydrates)
	case "protein":
		return floatPointerValue(row.Protein)
	case "fat_total":
		return floatPointerValue(row.FatTotal)
	default:
		return 0, false
	}
}

func aggregateActivityDay(date string, rows []intervals.Activity, metric analysis.Metric, unitSystem response.UnitSystem) (analysis.NumericSample, bool, map[string]any) {
	assumptions := map[string]any{"sample_grain": string(analysis.SampleGrainDaily)}
	if isActivityAdditiveMetric(metric) {
		var sum float64
		var n int
		for _, row := range rows {
			if value, ok := activityMetricValue(row, metric, unitSystem); ok {
				sum += value
				n++
			}
		}
		assumptions["aggregation"] = "sum"
		return analysis.NumericSample{Key: date, Date: date, Value: sum}, n > 0, assumptions
	}
	if metric == "pace_seconds_per_km" || metric == "pace_seconds_per_mile" {
		return aggregateActivityPace(date, rows, metric)
	}
	if metric == "average_speed_kmh" || metric == "average_speed_mph" {
		return aggregateActivitySpeed(date, rows, metric)
	}
	if metric == "max_speed_kmh" || metric == "max_speed_mph" || metric == "max_heart_rate_bpm" {
		var maxValue float64
		var ok bool
		for _, row := range rows {
			if value, valueOK := activityMetricValue(row, metric, unitSystem); valueOK && (!ok || value > maxValue) {
				maxValue = value
				ok = true
			}
		}
		assumptions["aggregation"] = "max"
		return analysis.NumericSample{Key: date, Date: date, Value: maxValue}, ok, assumptions
	}
	return weightedActivityMean(date, rows, metric, unitSystem)
}

func aggregateActivityPace(date string, rows []intervals.Activity, metric analysis.Metric) (analysis.NumericSample, bool, map[string]any) {
	assumptions := map[string]any{"sample_grain": string(analysis.SampleGrainDaily), "aggregation": "total_moving_time_per_total_distance"}
	var movingSeconds float64
	var distanceMeters float64
	for _, row := range rows {
		distance := firstFloat(row.ICUDistance, row.Distance)
		if distance == nil || *distance <= 0 || row.MovingTime == nil || *row.MovingTime <= 0 {
			continue
		}
		distanceMeters += *distance
		movingSeconds += float64(*row.MovingTime)
	}
	if movingSeconds <= 0 || distanceMeters <= 0 {
		return analysis.NumericSample{}, false, assumptions
	}
	metersPerUnit := 1000.0
	if metric == "pace_seconds_per_mile" {
		metersPerUnit = 1609.344
	}
	return analysis.NumericSample{Key: date, Date: date, Value: movingSeconds / (distanceMeters / metersPerUnit)}, true, assumptions
}

func aggregateActivitySpeed(date string, rows []intervals.Activity, metric analysis.Metric) (analysis.NumericSample, bool, map[string]any) {
	assumptions := map[string]any{"sample_grain": string(analysis.SampleGrainDaily), "aggregation": "total_distance_per_total_moving_time"}
	var movingSeconds float64
	var distanceMeters float64
	for _, row := range rows {
		distance := firstFloat(row.ICUDistance, row.Distance)
		if distance == nil || *distance <= 0 || row.MovingTime == nil || *row.MovingTime <= 0 {
			continue
		}
		distanceMeters += *distance
		movingSeconds += float64(*row.MovingTime)
	}
	if movingSeconds <= 0 || distanceMeters <= 0 {
		return analysis.NumericSample{}, false, assumptions
	}
	speedMS := distanceMeters / movingSeconds
	unitSystem := response.UnitSystemMetric
	if metric == "average_speed_mph" {
		unitSystem = response.UnitSystemImperial
	}
	return analysis.NumericSample{Key: date, Date: date, Value: response.ToPreferred(speedMS, units.UnitMS, unitSystem).Value}, true, assumptions
}

func weightedActivityMean(date string, rows []intervals.Activity, metric analysis.Metric, unitSystem response.UnitSystem) (analysis.NumericSample, bool, map[string]any) {
	assumptions := map[string]any{"sample_grain": string(analysis.SampleGrainDaily), "aggregation": "moving_time_weighted_mean"}
	var weighted, weights, plainSum float64
	var plainN, dropped int
	for _, row := range rows {
		value, ok := activityMetricValue(row, metric, unitSystem)
		if !ok {
			continue
		}
		plainSum += value
		plainN++
		if row.MovingTime != nil && *row.MovingTime > 0 {
			weight := float64(*row.MovingTime)
			weighted += value * weight
			weights += weight
		} else {
			dropped++
		}
	}
	if weights > 0 {
		if dropped > 0 {
			assumptions["aggregation_dropped_samples"] = dropped
		}
		return analysis.NumericSample{Key: date, Date: date, Value: weighted / weights}, true, assumptions
	}
	if plainN > 0 {
		assumptions["aggregation"] = "unweighted_mean"
		return analysis.NumericSample{Key: date, Date: date, Value: plainSum / float64(plainN)}, true, assumptions
	}
	return analysis.NumericSample{}, false, assumptions
}

func groupActivitiesByLocalDate(rows []intervals.Activity) map[string][]intervals.Activity {
	out := map[string][]intervals.Activity{}
	for _, row := range rows {
		date := localActivityDate(row)
		if date == "" {
			continue
		}
		out[date] = append(out[date], row)
	}
	return out
}

func localActivityDate(row intervals.Activity) string {
	value := strings.TrimSpace(stringValue(row.StartDateLocal))
	if len(value) >= len(time.DateOnly) {
		return value[:len(time.DateOnly)]
	}
	return ""
}

func sortedSamples(samples []analysis.NumericSample) []analysis.NumericSample {
	out := append([]analysis.NumericSample(nil), samples...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Date == out[j].Date {
			return out[i].Key < out[j].Key
		}
		return out[i].Date < out[j].Date
	})
	return out
}

func isActivityAdditiveMetric(metric analysis.Metric) bool {
	switch metric {
	case "moving_time_seconds", "elapsed_time_seconds", "distance_km", "distance_mi", "elevation_gain_m", "elevation_loss_m", "training_load", "calories_burned":
		return true
	default:
		return false
	}
}

func convertedDistanceValue(activity intervals.Activity, _ response.UnitSystem, unit units.Unit) (float64, bool) {
	distance := firstFloat(activity.ICUDistance, activity.Distance)
	if distance == nil || *distance <= 0 {
		return 0, false
	}
	converted := response.ToPreferred(*distance, units.UnitM, response.UnitSystemMetric)
	if unit == units.UnitMI {
		converted = response.ToPreferred(*distance, units.UnitM, response.UnitSystemImperial)
	}
	return converted.Value, true
}

func activityPace(activity intervals.Activity, unit units.Unit) (float64, bool) {
	distance := firstFloat(activity.ICUDistance, activity.Distance)
	if distance == nil || *distance <= 0 || activity.MovingTime == nil || *activity.MovingTime <= 0 {
		return 0, false
	}
	metersPerUnit := 1000.0
	if unit == units.UnitMI {
		metersPerUnit = 1609.344
	}
	return float64(*activity.MovingTime) / (*distance / metersPerUnit), true
}

func activitySpeed(speed *float64, unit units.Unit) (float64, bool) {
	if speed == nil || *speed <= 0 {
		return 0, false
	}
	converted := response.ToPreferred(*speed, units.UnitMS, response.UnitSystemMetric)
	if unit == units.UnitMPH {
		converted = response.ToPreferred(*speed, units.UnitMS, response.UnitSystemImperial)
	}
	return converted.Value, true
}

func intPointerValue(value *int) (float64, bool) {
	if value == nil {
		return 0, false
	}
	return float64(*value), true
}

func floatPointerValue(value *float64) (float64, bool) {
	if value == nil {
		return 0, false
	}
	return *value, true
}
