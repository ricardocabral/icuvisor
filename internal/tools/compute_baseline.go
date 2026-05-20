package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/resources"
)

const (
	computeBaselineName                    = "compute_baseline"
	computeBaselineDescription             = "Compute a deterministic baseline z-score from existing read-tool outputs; do not fetch rows or streams and reduce manually. Returns wellness interpretation and explicit insufficient-data signals."
	invalidComputeBaselineArgumentsMessage = "invalid compute_baseline arguments; provide metric, baseline/current date windows, optional sport, and min_samples >= 2"
	fetchComputeBaselineMessage            = "could not compute baseline; check intervals.icu credentials, athlete ID, and date range"
)

type computeBaselineRequest struct {
	Metric            string `json:"metric"`
	BaselineStartDate string `json:"baseline_start_date"`
	BaselineEndDate   string `json:"baseline_end_date"`
	CurrentStartDate  string `json:"current_start_date"`
	CurrentEndDate    string `json:"current_end_date"`
	Sport             string `json:"sport,omitempty"`
	MinSamples        int    `json:"min_samples,omitempty"`
	IncludeFull       bool   `json:"include_full,omitempty"`
}

type baselineSample struct {
	Date          string   `json:"date,omitempty"`
	ActivityID    string   `json:"activity_id,omitempty"`
	Window        string   `json:"window"`
	Value         *float64 `json:"value,omitempty"`
	SourceTool    string   `json:"source_tool,omitempty"`
	MissingReason string   `json:"missing_reason,omitempty"`
}

type computeBaselineResult struct {
	Status                      string               `json:"status"`
	Metric                      string               `json:"metric"`
	MetricSource                baselineMetricSource `json:"metric_source"`
	BaselineWindow              dateWindow           `json:"baseline_window"`
	CurrentWindow               dateWindow           `json:"current_window"`
	CurrentValue                *float64             `json:"current_value,omitempty"`
	BaselineMean                *float64             `json:"baseline_mean,omitempty"`
	BaselineStdDev              *float64             `json:"baseline_stddev,omitempty"`
	ZScore                      *float64             `json:"z_score,omitempty"`
	Interpretation              string               `json:"interpretation"`
	NBaseline                   int                  `json:"n_baseline"`
	NCurrent                    int                  `json:"n_current"`
	MinSamples                  int                  `json:"min_samples"`
	MissingBaselineDays         int                  `json:"missing_baseline_days"`
	MissingCurrentDays          int                  `json:"missing_current_days"`
	TruncatedActivityCandidates bool                 `json:"truncated_activity_candidates,omitempty"`
	InsufficientReason          string               `json:"insufficient_reason,omitempty"`
}

type baselineMetricSource struct {
	Family string `json:"family"`
	Tool   string `json:"tool"`
	Field  string `json:"field"`
	Grain  string `json:"grain"`
}
type dateWindow struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type baselineCollected struct {
	Baseline            []float64
	Current             []float64
	Series              []baselineSample
	Source              analysis.MetricSource
	MissingBaselineDays int
	MissingCurrentDays  int
	SourceTools         []string
	Truncated           bool
	UnsupportedReason   string
}

func newComputeBaselineTool(fitnessClient FitnessClient, wellnessClient WellnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: computeBaselineName, Description: computeBaselineDescription, InputSchema: computeBaselineInputSchema(), OutputSchema: genericOutputSchema("Baseline z-score with analyzer metadata."), Handler: computeBaselineHandler(fitnessClient, wellnessClient, activitiesClient, extendedClient, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func computeBaselineHandler(fitnessClient FitnessClient, wellnessClient WellnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, metric, err := decodeComputeBaselineRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidComputeBaselineArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchComputeBaselineMessage, err)
		}
		collected, err := collectBaselineSamples(ctx, args, metric, fitnessClient, wellnessClient, activitiesClient, extendedClient)
		if err != nil {
			if contextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchComputeBaselineMessage, err)
		}
		minSamples := args.MinSamples
		if minSamples == 0 {
			minSamples = analysis.MinBaselineSamples
		}
		stats := analysis.ComputeBaselineStats(collected.Baseline, collected.Current, minSamples, metric == "weekly_tss" || metric == "weekly_hours")
		if collected.UnsupportedReason != "" {
			stats.Status = "unsupported_metric_source"
			stats.Reason = collected.UnsupportedReason
		}
		interpretation := analysis.InterpretBaselineZScore(metric, stats.ZScore)
		result := computeBaselineResult{Status: stats.Status, Metric: string(metric), MetricSource: sourceDTO(collected.Source), BaselineWindow: dateWindow{args.BaselineStartDate, args.BaselineEndDate}, CurrentWindow: dateWindow{args.CurrentStartDate, args.CurrentEndDate}, CurrentValue: roundOptional(stats.CurrentValue), BaselineMean: roundOptional(stats.BaselineMean), BaselineStdDev: roundOptional(stats.BaselineStdDev), ZScore: roundOptional(stats.ZScore), Interpretation: interpretation, NBaseline: len(collected.Baseline), NCurrent: len(collected.Current), MinSamples: minSamples, MissingBaselineDays: collected.MissingBaselineDays, MissingCurrentDays: collected.MissingCurrentDays, TruncatedActivityCandidates: collected.Truncated, InsufficientReason: stats.Reason}
		assumptions := map[string]any{"metric": string(metric), "interpretation": interpretation, "interpretation_direction": baselineInterpretationDirection(metric), "activity_candidates_truncated": collected.Truncated}
		meta := analysis.AnalyzerMetaInput{Method: "baseline_z_score", SourceTools: collected.SourceTools, N: len(collected.Baseline), MissingDays: collected.MissingBaselineDays + collected.MissingCurrentDays, MissingAction: analysis.MissingActionSkip, MinSamples: minSamples, FormulaRef: resources.AnalysisFormulaRefZScore, Assumptions: assumptions, Boundaries: []string{"missing samples are skipped; no imputation", "raw streams are not used for baseline metrics"}}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: collected.Series, Meta: meta}, args.IncludeFull, version, debugMetadata, computeBaselineName, unitSystem, shapeCfg)
	}
}

func decodeComputeBaselineRequest(raw json.RawMessage) (computeBaselineRequest, analysis.Metric, error) {
	var args computeBaselineRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, "", errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[computeBaselineRequest](raw)
	if err != nil {
		return args, "", err
	}
	args = decoded
	args.Metric = strings.TrimSpace(args.Metric)
	args.Sport = strings.TrimSpace(args.Sport)
	for _, value := range []string{args.BaselineStartDate, args.BaselineEndDate, args.CurrentStartDate, args.CurrentEndDate} {
		if !validDate(value) {
			return args, "", errors.New("all dates must be YYYY-MM-DD")
		}
	}
	if args.BaselineEndDate < args.BaselineStartDate || args.CurrentEndDate < args.CurrentStartDate {
		return args, "", errors.New("window end dates must be on or after start dates")
	}
	if args.MinSamples != 0 && args.MinSamples < 2 {
		return args, "", errors.New("min_samples must be at least 2")
	}
	metric, err := analysis.ParseMetric(args.Metric)
	if err != nil {
		return args, "", err
	}
	return args, metric, nil
}

func collectBaselineSamples(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, fitnessClient FitnessClient, wellnessClient WellnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient) (baselineCollected, error) {
	sources := analysis.MetricSources(metric)
	for _, source := range sources {
		switch source.Family {
		case analysis.SourceFitnessDaily, analysis.SourceTrainingSummary, analysis.SourceDerivedWeekly:
			if fitnessClient == nil {
				continue
			}
			return collectSummaryBaseline(ctx, args, metric, source, fitnessClient)
		case analysis.SourceWellnessDaily:
			if wellnessClient == nil {
				continue
			}
			return collectWellnessBaseline(ctx, args, metric, source, wellnessClient)
		case analysis.SourceActivityRow, analysis.SourceExtendedActivity:
			if activitiesClient == nil {
				continue
			}
			return collectActivityBaseline(ctx, args, metric, source, activitiesClient, extendedClient)
		}
	}
	return baselineCollected{UnsupportedReason: "interval_grain_not_supported_for_baseline"}, nil
}

func collectSummaryBaseline(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, source analysis.MetricSource, client FitnessClient) (baselineCollected, error) {
	rows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.BaselineStartDate, End: args.CurrentEndDate})
	if err != nil {
		return baselineCollected{}, err
	}
	out := baselineCollected{Source: source, SourceTools: []string{source.Tool}}
	seenB, seenC := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		value, ok := summaryMetricValue(row, metric, source.Field)
		date := row.Date
		window := sampleWindow(date, args)
		if window == "" {
			continue
		}
		if ok {
			addBaselineSample(&out, window, date, "", value, source.Tool)
			if window == "baseline" {
				seenB[date] = true
			} else {
				seenC[date] = true
			}
		} else {
			out.Series = append(out.Series, baselineSample{Date: date, Window: window, SourceTool: source.Tool, MissingReason: "missing_metric"})
		}
	}
	out.MissingBaselineDays = dateCount(args.BaselineStartDate, args.BaselineEndDate) - len(seenB)
	out.MissingCurrentDays = dateCount(args.CurrentStartDate, args.CurrentEndDate) - len(seenC)
	return out, nil
}

func collectWellnessBaseline(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, source analysis.MetricSource, client WellnessClient) (baselineCollected, error) {
	rows, err := client.ListWellness(ctx, intervals.WellnessParams{Oldest: args.BaselineStartDate, Newest: args.CurrentEndDate, Fields: []string{source.Field}})
	if err != nil {
		return baselineCollected{}, err
	}
	out := baselineCollected{Source: source, SourceTools: []string{source.Tool}}
	seenB, seenC := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		date := wellnessDate(row)
		window := sampleWindow(date, args)
		if window == "" {
			continue
		}
		value, ok := wellnessMetricValue(row, source.Field)
		if ok {
			addBaselineSample(&out, window, date, "", value, source.Tool)
			if window == "baseline" {
				seenB[date] = true
			} else {
				seenC[date] = true
			}
		} else {
			out.Series = append(out.Series, baselineSample{Date: date, Window: window, SourceTool: source.Tool, MissingReason: "missing_metric"})
		}
	}
	out.MissingBaselineDays = dateCount(args.BaselineStartDate, args.BaselineEndDate) - len(seenB)
	out.MissingCurrentDays = dateCount(args.CurrentStartDate, args.CurrentEndDate) - len(seenC)
	return out, nil
}

func collectActivityBaseline(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, source analysis.MetricSource, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient) (baselineCollected, error) {
	activities, err := activitiesClient.ListActivities(ctx, intervals.ListActivitiesParams{Oldest: args.BaselineStartDate, Newest: args.CurrentEndDate, Limit: maxComputeActivityCandidates})
	if err != nil {
		return baselineCollected{}, err
	}
	out := baselineCollected{Source: source, SourceTools: []string{getActivitiesName}, Truncated: len(activities) >= maxComputeActivityCandidates}
	if source.Tool == getExtendedMetricsName {
		out.SourceTools = append(out.SourceTools, getExtendedMetricsName)
	}
	sort.SliceStable(activities, func(i, j int) bool {
		if stringValue(activities[i].StartDateLocal) != stringValue(activities[j].StartDateLocal) {
			return stringValue(activities[i].StartDateLocal) < stringValue(activities[j].StartDateLocal)
		}
		return activities[i].ID < activities[j].ID
	})
	seenB, seenC := map[string]bool{}, map[string]bool{}
	for _, activity := range activities {
		if args.Sport != "" && !sameFold(args.Sport, stringValue(activity.Type)) {
			continue
		}
		date := localDatePrefix(stringValue(activity.StartDateLocal))
		window := sampleWindow(date, args)
		if window == "" {
			continue
		}
		value, ok := activityMetricValue(activity, source.Field)
		if !ok && source.Tool == getExtendedMetricsName && extendedClient != nil && activity.ID != "" {
			detail, derr := extendedClient.GetActivity(ctx, activity.ID)
			if derr == nil {
				value, ok = extendedActivityMetricValue(detail.Raw, source.Field)
			}
		}
		if ok {
			addBaselineSample(&out, window, date, activity.ID, value, source.Tool)
			if window == "baseline" {
				seenB[date] = true
			} else {
				seenC[date] = true
			}
		} else {
			out.Series = append(out.Series, baselineSample{Date: date, ActivityID: activity.ID, Window: window, SourceTool: source.Tool, MissingReason: "missing_metric"})
		}
	}
	out.MissingBaselineDays = dateCount(args.BaselineStartDate, args.BaselineEndDate) - len(seenB)
	out.MissingCurrentDays = dateCount(args.CurrentStartDate, args.CurrentEndDate) - len(seenC)
	return out, nil
}

func addBaselineSample(out *baselineCollected, window, date, activityID string, value float64, tool string) {
	v := round(value, 4)
	if window == "baseline" {
		out.Baseline = append(out.Baseline, value)
	} else {
		out.Current = append(out.Current, value)
	}
	out.Series = append(out.Series, baselineSample{Date: date, ActivityID: activityID, Window: window, Value: &v, SourceTool: tool})
}
func sampleWindow(date string, args computeBaselineRequest) string {
	if date >= args.BaselineStartDate && date <= args.BaselineEndDate {
		return "baseline"
	}
	if date >= args.CurrentStartDate && date <= args.CurrentEndDate {
		return "current"
	}
	return ""
}

func summaryMetricValue(row intervals.SummaryWithCats, metric analysis.Metric, field string) (float64, bool) {
	switch metric {
	case "ctl":
		return row.Fitness, true
	case "atl":
		return row.Fatigue, true
	case "tsb":
		return row.Form, true
	case "weekly_tss", "training_load":
		return float64(row.TrainingLoad), row.TrainingLoad != 0
	case "weekly_hours":
		return float64(row.Time) / 3600, row.Time != 0
	case "moving_time_seconds":
		return float64(row.MovingTime), row.MovingTime != 0
	case "elapsed_time_seconds", "time_seconds":
		return float64(row.ElapsedTime), row.ElapsedTime != 0
	case "calories_burned":
		return float64(row.Calories), row.Calories != 0
	case "elevation_gain_m":
		return row.TotalElevationGain, row.TotalElevationGain != 0
	case "time_in_zones_total_seconds":
		return float64(row.TimeInZonesTot), row.TimeInZonesTot != 0
	}
	return rawNumber(row.Raw, field)
}
func wellnessMetricValue(row intervals.Wellness, field string) (float64, bool) {
	return rawNumber(row.Raw, field)
}
func activityMetricValue(activity intervals.Activity, field string) (float64, bool) {
	switch field {
	case "moving_time_seconds":
		if activity.MovingTime != nil {
			return float64(*activity.MovingTime), true
		}
	case "elapsed_time_seconds":
		if activity.ElapsedTime != nil {
			return float64(*activity.ElapsedTime), true
		}
	case "training_load":
		if activity.TrainingLoad != nil {
			return float64(*activity.TrainingLoad), true
		}
	case "average_heart_rate_bpm":
		if activity.AverageHeartRate != nil {
			return float64(*activity.AverageHeartRate), true
		}
	case "max_heart_rate_bpm":
		if activity.MaxHeartRate != nil {
			return float64(*activity.MaxHeartRate), true
		}
	case "average_cadence_rpm":
		if activity.AverageCadence != nil {
			return *activity.AverageCadence, true
		}
	case "calories_burned":
		if activity.Calories != nil {
			return float64(*activity.Calories), true
		}
	case "elevation_gain_m":
		if activity.TotalElevationGain != nil {
			return *activity.TotalElevationGain, true
		}
	case "elevation_loss_m":
		if activity.TotalElevationLoss != nil {
			return *activity.TotalElevationLoss, true
		}
	}
	return rawNumber(activity.Raw, field)
}
func extendedActivityMetricValue(raw map[string]any, field string) (float64, bool) {
	metrics := extendedMetricsFromActivity(raw, intervals.PowerVsHR{}, false)
	data, _ := json.Marshal(metrics)
	var m map[string]any
	_ = json.Unmarshal(data, &m)
	return numberFromAny(m[field])
}
func wellnessDate(row intervals.Wellness) string {
	for _, key := range []string{"id", "date", "day"} {
		if value := anyString(row.Raw[key]); len(value) >= len("2006-01-02") {
			return value[:len("2006-01-02")]
		}
	}
	return ""
}
func sourceDTO(source analysis.MetricSource) baselineMetricSource {
	return baselineMetricSource{Family: string(source.Family), Tool: source.Tool, Field: source.Field, Grain: string(source.Grain)}
}
func baselineInterpretationDirection(metric analysis.Metric) string {
	interp := analysis.InterpretBaselineZScore(metric, func() *float64 { v := 2.0; return &v }())
	if interp == "elevated" {
		return "adverse_high"
	}
	if interp == "elevated_beneficial" {
		return "beneficial_high"
	}
	return "not_directional"
}

func computeBaselineInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"metric", "baseline_start_date", "baseline_end_date", "current_start_date", "current_end_date"}, "properties": map[string]any{"metric": analysis.MetricSchemaProperty(), "baseline_start_date": map[string]any{"type": "string", "description": "Baseline inclusive athlete-local start date YYYY-MM-DD."}, "baseline_end_date": map[string]any{"type": "string", "description": "Baseline inclusive athlete-local end date YYYY-MM-DD."}, "current_start_date": map[string]any{"type": "string", "description": "Current inclusive athlete-local start date YYYY-MM-DD."}, "current_end_date": map[string]any{"type": "string", "description": "Current inclusive athlete-local end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "description": "Optional exact case-insensitive activity sport/type filter where the selected source supports sport filtering."}, "min_samples": map[string]any{"type": "integer", "minimum": 2, "default": analysis.MinBaselineSamples, "description": "Minimum usable baseline samples required before z-score calculation."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include baseline/current audit samples; default returns aggregate statistics only."}}}
}
