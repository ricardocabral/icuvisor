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
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	computeBaselineName                    = "compute_baseline"
	computeBaselineDescription             = "Use when the prompt asks whether current metric values are unusual relative to a historical baseline via z-score/status bands; do not fetch get_* rows or streams and reduce them in chat. Computes z-scores from existing read outputs with wellness interpretation and explicit insufficient-data signals. HRV comparisons with explicit provider or device changes return insufficient_comparable_data; choose windows from one measurement source rather than mixing baselines."
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
	Status                      string                    `json:"status"`
	Metric                      string                    `json:"metric"`
	MetricSource                baselineMetricSource      `json:"metric_source"`
	BaselineWindow              dateWindow                `json:"baseline_window"`
	CurrentWindow               dateWindow                `json:"current_window"`
	CurrentValue                *float64                  `json:"current_value,omitempty"`
	BaselineMean                *float64                  `json:"baseline_mean,omitempty"`
	BaselineStdDev              *float64                  `json:"baseline_stddev,omitempty"`
	ZScore                      *float64                  `json:"z_score,omitempty"`
	Interpretation              string                    `json:"interpretation"`
	NBaseline                   int                       `json:"n_baseline"`
	NCurrent                    int                       `json:"n_current"`
	MinSamples                  int                       `json:"min_samples"`
	MissingBaselineDays         *int                      `json:"missing_baseline_days,omitempty"`
	MissingCurrentDays          *int                      `json:"missing_current_days,omitempty"`
	MissingBaselineAnchors      *int                      `json:"missing_baseline_anchors,omitempty"`
	MissingCurrentAnchors       *int                      `json:"missing_current_anchors,omitempty"`
	FreshnessStatus             string                    `json:"freshness_status,omitempty"`
	CurrentLatestSampleDate     string                    `json:"current_latest_sample_date,omitempty"`
	CurrentWindowEndDate        string                    `json:"current_window_end_date,omitempty"`
	Caveats                     []string                  `json:"caveats,omitempty"`
	SourceContinuity            *baselineSourceContinuity `json:"source_continuity,omitempty"`
	TruncatedActivityCandidates bool                      `json:"truncated_activity_candidates,omitempty"`
	InsufficientReason          string                    `json:"insufficient_reason,omitempty"`
}

type baselineSourceContinuity struct {
	Status               string   `json:"status"`
	Sources              []string `json:"sources,omitempty"`
	UnknownSourceSamples int      `json:"unknown_source_samples"`
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
	Baseline               []float64
	Current                []float64
	Series                 []baselineSample
	Source                 analysis.MetricSource
	MissingBaselineDays    int
	MissingCurrentDays     int
	MissingBaselineAnchors int
	MissingCurrentAnchors  int
	SourceTools            []string
	Truncated              bool
	UnsupportedReason      string
	WellnessFreshness      *wellnessFreshnessSummary
	SourceContinuity       *baselineSourceContinuity
}

func newComputeBaselineTool(fitnessClient FitnessClient, wellnessClient WellnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: computeBaselineName, Description: computeBaselineDescription, InputSchema: computeBaselineInputSchema(), OutputSchema: genericOutputSchema("Baseline z-score with analyzer metadata."), Handler: computeBaselineHandler(fitnessClient, wellnessClient, activitiesClient, extendedClient, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
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
		collected, err := collectBaselineSamples(ctx, args, metric, unitSystem, fitnessClient, wellnessClient, activitiesClient, extendedClient)
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
		stats := analysis.ComputeBaselineStats(collected.Baseline, collected.Current, minSamples, false)
		if collected.UnsupportedReason != "" {
			stats.Status = "unsupported_metric_source"
			stats.Reason = collected.UnsupportedReason
		}
		if continuity := collected.SourceContinuity; continuity != nil && continuity.Status == "changed" {
			stats = analysis.BaselineStats{Status: "insufficient_comparable_data", Reason: "hrv_source_changed"}
		}
		if collected.Truncated {
			if stats.Reason == "" && stats.Status != "ok" {
				stats.Reason = stats.Status
			}
			stats.Status = "partial"
		}
		interpretation := analysis.InterpretBaselineZScore(metric, stats.ZScore)
		result := computeBaselineResult{Status: stats.Status, Metric: string(metric), MetricSource: sourceDTO(collected.Source), BaselineWindow: dateWindow{args.BaselineStartDate, args.BaselineEndDate}, CurrentWindow: dateWindow{args.CurrentStartDate, args.CurrentEndDate}, CurrentValue: roundOptional(stats.CurrentValue), BaselineMean: roundOptional(stats.BaselineMean), BaselineStdDev: roundOptional(stats.BaselineStdDev), ZScore: roundOptional(stats.ZScore), Interpretation: interpretation, NBaseline: len(collected.Baseline), NCurrent: len(collected.Current), MinSamples: minSamples, TruncatedActivityCandidates: collected.Truncated, InsufficientReason: stats.Reason}
		assumptions := map[string]any{"metric": string(metric), "interpretation": interpretation, "interpretation_direction": baselineInterpretationDirection(metric), "activity_candidates_truncated": collected.Truncated}
		if collected.Source.Grain == analysis.GrainWeekly {
			result.MissingBaselineAnchors = baselineIntPtr(collected.MissingBaselineAnchors)
			result.MissingCurrentAnchors = baselineIntPtr(collected.MissingCurrentAnchors)
			assumptions["sample_grain"] = string(analysis.SampleGrainWeekly)
			assumptions["expected_baseline_anchors"] = len(expectedWeeklyAnchors(args.BaselineStartDate, args.BaselineEndDate))
			assumptions["expected_current_anchors"] = len(expectedWeeklyAnchors(args.CurrentStartDate, args.CurrentEndDate))
			assumptions["missing_baseline_anchors"] = collected.MissingBaselineAnchors
			assumptions["missing_current_anchors"] = collected.MissingCurrentAnchors
			assumptions["missing_days_applicable"] = false
		} else {
			result.MissingBaselineDays = baselineIntPtr(collected.MissingBaselineDays)
			result.MissingCurrentDays = baselineIntPtr(collected.MissingCurrentDays)
		}
		if freshness := collected.WellnessFreshness; freshness != nil {
			result.FreshnessStatus = freshness.Status
			result.CurrentLatestSampleDate = freshness.LatestSampleDate
			result.CurrentWindowEndDate = freshness.WindowEndDate
			result.Caveats = append([]string(nil), freshness.Caveats...)
			assumptions["wellness_freshness"] = freshness.computeAssumptions()
			if result.Status == "ok" && (freshness.Status == "stale_current_window" || freshness.Status == "stale_provenance") {
				result.Status = "partial"
				if result.InsufficientReason == "" {
					result.InsufficientReason = freshness.Status
				}
			}
		}
		result.SourceContinuity = collected.SourceContinuity
		if continuity := collected.SourceContinuity; continuity != nil {
			assumptions["hrv_source_continuity"] = continuity.Status
			switch continuity.Status {
			case "changed":
				result.Caveats = append(result.Caveats, "Explicit HRV provider or device provenance changed within the requested windows. Choose baseline and current windows from one measurement source; no pooled z-score or interpretation is returned.")
			case "unknown":
				result.Caveats = append(result.Caveats, "Some HRV samples have no explicit source provenance; measurement-source continuity cannot be verified.")
			}
		}
		meta := analysis.AnalyzerMetaInput{Method: "baseline_z_score", SourceTools: collected.SourceTools, N: len(collected.Baseline), MissingDays: collected.MissingBaselineDays + collected.MissingCurrentDays, MissingAction: analysis.MissingActionSkip, MinSamples: minSamples, FormulaRef: resources.AnalysisFormulaRefZScore, Assumptions: assumptions, Boundaries: baselineBoundaries(collected.Truncated)}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: collected.Series, Meta: meta}, args.IncludeFull, version, debugMetadata, computeBaselineName, unitSystem, shapeCfg)
	}
}

func baselineIntPtr(value int) *int { return &value }

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
	if args.BaselineEndDate >= args.CurrentStartDate {
		return args, "", errors.New("baseline_end_date must be before current_start_date")
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

func collectBaselineSamples(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, unitSystem response.UnitSystem, fitnessClient FitnessClient, wellnessClient WellnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient) (baselineCollected, error) {
	sources := analysis.MetricSources(metric)
	for _, source := range sources {
		switch source.Family {
		case analysis.SourceFitnessWeekly, analysis.SourceTrainingSummary:
			if fitnessClient == nil {
				continue
			}
			return collectSummaryBaseline(ctx, args, metric, unitSystem, source, fitnessClient)
		case analysis.SourceWellnessDaily:
			if wellnessClient == nil {
				continue
			}
			return collectWellnessBaseline(ctx, args, metric, source, wellnessClient)
		case analysis.SourceActivityRow, analysis.SourceExtendedActivity:
			if activitiesClient == nil {
				continue
			}
			return collectActivityBaseline(ctx, args, metric, unitSystem, source, activitiesClient, extendedClient)
		}
	}
	return baselineCollected{UnsupportedReason: "interval_grain_not_supported_for_baseline"}, nil
}

func collectSummaryBaseline(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, unitSystem response.UnitSystem, source analysis.MetricSource, client FitnessClient) (baselineCollected, error) {
	rows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.BaselineStartDate, End: args.CurrentEndDate})
	if err != nil {
		return baselineCollected{}, err
	}
	out := baselineCollected{Source: source, SourceTools: []string{source.Tool}}
	seenB, seenC := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		value, ok := summaryMetricValueForSport(row, metric, unitSystem, source.Field, args.Sport)
		date := row.Date
		window := sampleWindow(date, args)
		if window == "" {
			continue
		}
		if !ok {
			out.Series = append(out.Series, baselineSample{Date: date, Window: window, SourceTool: source.Tool, MissingReason: "missing_metric"})
			continue
		}
		addBaselineSample(&out, window, date, "", value, source.Tool)
		if window == "baseline" {
			seenB[date] = true
		} else {
			seenC[date] = true
		}
	}
	out.MissingBaselineAnchors = missingWeeklyAnchorCount(args.BaselineStartDate, args.BaselineEndDate, seenB)
	out.MissingCurrentAnchors = missingWeeklyAnchorCount(args.CurrentStartDate, args.CurrentEndDate, seenC)
	return out, nil
}

func expectedWeeklyAnchors(startDate, endDate string) []string {
	window, err := analysis.ParseWindow(analysis.Window{StartDate: startDate, EndDate: endDate}, maxAnalyzerWindowDays)
	if err != nil {
		return nil
	}
	return expectedWeeklyAnchorDates(window)
}

func missingWeeklyAnchorCount(startDate, endDate string, seen map[string]bool) int {
	missing := 0
	for _, anchor := range expectedWeeklyAnchors(startDate, endDate) {
		if !seen[anchor] {
			missing++
		}
	}
	return missing
}

func collectWellnessBaseline(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, source analysis.MetricSource, client WellnessClient) (baselineCollected, error) {
	rows, err := client.ListWellness(ctx, intervals.WellnessParams{Oldest: args.BaselineStartDate, Newest: args.CurrentEndDate, Fields: analyzerWellnessFields(metric, source.Field)})
	if err != nil {
		return baselineCollected{}, err
	}
	out := baselineCollected{Source: source, SourceTools: []string{source.Tool}}
	seenB, seenC := map[string]bool{}, map[string]bool{}
	freshness := newWellnessFreshnessTracker(metric, args.CurrentStartDate, args.CurrentEndDate)
	for _, row := range rows {
		date := wellnessDate(row)
		window := sampleWindow(date, args)
		if window == "" {
			continue
		}
		value, ok := wellnessMetricValue(row, metric)
		if !ok {
			value, ok = rawNumber(row.Raw, source.Field)
		}
		if ok {
			addBaselineSample(&out, window, date, "", value, source.Tool)
			if window == "baseline" {
				seenB[date] = true
			} else {
				seenC[date] = true
				freshness.record(row, date)
			}
		} else {
			out.Series = append(out.Series, baselineSample{Date: date, Window: window, SourceTool: source.Tool, MissingReason: "missing_metric"})
		}
	}
	out.MissingBaselineDays = dateCount(args.BaselineStartDate, args.BaselineEndDate) - len(seenB)
	out.MissingCurrentDays = dateCount(args.CurrentStartDate, args.CurrentEndDate) - len(seenC)
	out.WellnessFreshness = freshness.baselineSummary(out.MissingCurrentDays)
	out.SourceContinuity = baselineHRVContinuity(rows, args, metric)
	return out, nil
}

func baselineHRVContinuity(rows []intervals.Wellness, args computeBaselineRequest, metric analysis.Metric) *baselineSourceContinuity {
	_, _, ok := wellnessFreshnessField(metric)
	if !ok {
		return nil
	}
	out := &baselineSourceContinuity{Status: "unknown"}
	sources, devices := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		if sampleWindow(wellnessDate(row), args) == "" {
			continue
		}
		if _, ok := wellnessMetricValue(row, metric); !ok {
			continue
		}
		// Native sleep/readiness fields do not identify the HRV measurement source.
		source := ""
		for _, key := range []string{"source", "provider", "wellnessSource", "wellness_source", "integration"} {
			value, _ := row.Raw[key].(string)
			if value = strings.ToLower(strings.TrimSpace(value)); value != "" && value != "unknown" {
				source = value
				break
			}
		}
		if source != "" {
			sources[source] = true
		} else {
			out.UnknownSourceSamples++
		}
		device, _ := row.Raw["device"].(string)
		device = strings.ToLower(strings.TrimSpace(device))
		if device == "unknown" {
			device = ""
		}
		if device != "" {
			devices[device] = true
		}
	}
	for source := range sources {
		out.Sources = append(out.Sources, source)
	}
	sort.Strings(out.Sources)
	switch {
	case len(sources) > 1 || len(devices) > 1:
		out.Status = "changed"
	case len(sources) == 1 && out.UnknownSourceSamples == 0:
		out.Status = "consistent"
	}
	return out
}

func collectActivityBaseline(ctx context.Context, args computeBaselineRequest, metric analysis.Metric, unitSystem response.UnitSystem, source analysis.MetricSource, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient) (baselineCollected, error) {
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
		value, ok := activityMetricValue(activity, metric, unitSystem)
		if !ok {
			value, ok = rawNumber(activity.Raw, source.Field)
		}
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

func summaryMetricValueForSport(row intervals.SummaryWithCats, metric analysis.Metric, unitSystem response.UnitSystem, field string, sport string) (float64, bool) {
	if strings.TrimSpace(sport) == "" {
		value, ok := summaryMetricValue(row, metric, unitSystem)
		if ok {
			return value, true
		}
		return rawNumber(row.Raw, field)
	}
	for _, category := range row.ByCategory {
		if !sameFold(sport, category.Category) {
			continue
		}
		switch metric {
		case "weekly_tss", "training_load":
			return float64(category.TrainingLoad), category.TrainingLoad != 0
		case "weekly_hours":
			return float64(category.Time) / 3600, category.Time != 0
		case "moving_time_seconds":
			return float64(category.MovingTime), category.MovingTime != 0
		case "elapsed_time_seconds", "time_seconds":
			return float64(category.ElapsedTime), category.ElapsedTime != 0
		case "calories_burned":
			return float64(category.Calories), category.Calories != 0
		case "distance_km":
			return category.Distance / 1000, category.Distance != 0
		case "distance_mi":
			return category.Distance / 1609.344, category.Distance != 0
		case "session_rpe":
			return float64(category.SRPE), category.SRPE != 0
		case "elevation_gain_m":
			return category.TotalElevationGain, category.TotalElevationGain != 0
		}
	}
	return 0, false
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
func baselineBoundaries(truncated bool) []string {
	boundaries := []string{"missing samples are skipped; no imputation", "raw streams are not used for baseline metrics"}
	if truncated {
		boundaries = append(boundaries, "activity candidates truncated at deterministic cap; baseline/current activity samples may be incomplete")
	}
	return boundaries
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
