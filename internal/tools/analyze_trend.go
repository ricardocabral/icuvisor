package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/resources"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	analyzeTrendName        = "analyze_trend"
	analyzeTrendDescription = "Use when the prompt asks whether an analysis metric is trending up, trending down, or changing versus baseline; do not fetch get_* rows or streams and reduce them in chat. Computes rolling means, OLS slope, and current-vs-baseline deltas with skipped-missing-day metadata."
	invalidAnalyzeTrendArgs = "invalid analyze_trend arguments; provide metric, window dates, optional baseline_window, rolling_window_days, sport, and include_full"
	fetchAnalyzeTrendMsg    = "could not analyze trend; check credentials, date range, metric, and sport filter"
)

type analyzeTrendRequest struct {
	Metric            string                 `json:"metric"`
	Window            analyzerWindowRequest  `json:"window"`
	BaselineWindow    *analyzerWindowRequest `json:"baseline_window,omitempty"`
	RollingWindowDays int                    `json:"rolling_window_days,omitempty"`
	Sport             string                 `json:"sport,omitempty"`
	IncludeFull       bool                   `json:"include_full,omitempty"`
}

func newAnalyzeTrendTool(fitness FitnessClient, wellness WellnessClient, activities ActivitiesClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	clients := analyzerClients{fitness: fitness, wellness: wellness, activities: activities}
	return fullTool(Tool{Name: analyzeTrendName, Description: analyzeTrendDescription, InputSchema: analyzeTrendInputSchema(), OutputSchema: genericOutputSchema("Analyzer trend result with rolling means, slope, baseline deltas, and analyzer _meta."), Handler: analyzeTrendHandler(clients, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func analyzeTrendHandler(clients analyzerClients, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAnalyzerStrict[analyzeTrendRequest](req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeTrendArgs, err)
		}
		metric, err := parseMetricArgument(args.Metric)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeTrendArgs, err)
		}
		window, err := analysis.ParseWindow(args.Window.analysisWindow(), maxAnalyzerWindowDays)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeTrendArgs, err)
		}
		rolling := args.RollingWindowDays
		if rolling == 0 {
			rolling = 7
		}
		if rolling < 2 || rolling > 90 || rolling > window.Days {
			return Result{}, NewUserError(invalidAnalyzeTrendArgs, fmt.Errorf("rolling_window_days must be 2..90 and no larger than the current window"))
		}
		baselineWindow := analysis.DefaultBaselineWindow(window)
		if args.BaselineWindow != nil {
			baselineWindow = args.BaselineWindow.analysisWindow()
		}
		baseline, err := analysis.ParseWindow(baselineWindow, maxAnalyzerWindowDays)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeTrendArgs, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchAnalyzeTrendMsg, err)
		}
		currentSeries, err := loadAnalyzerSeries(ctx, clients, metric, window, analysis.SampleGrainDaily, args.Sport, unitSystem, true)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnalyzeTrendMsg, err)
		}
		grain := analysis.SampleGrain(currentSeries.Assumptions["sample_grain"].(string))
		effectiveRollingDays := rolling
		if grain == analysis.SampleGrainWeekly {
			if rolling%7 != 0 {
				return Result{}, NewUserError(invalidAnalyzeTrendArgs, fmt.Errorf("rolling_window_days must be a multiple of 7 for weekly metrics"))
			}
			rolling = rolling / 7
		}
		baselineSeries, err := loadAnalyzerSeries(ctx, clients, metric, baseline, grain, args.Sport, unitSystem, true)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnalyzeTrendMsg, err)
		}
		minSamples := analysis.MinBaselineSamples
		if grain == analysis.SampleGrainWeekly {
			minSamples = 4
		}
		trend, points := analysis.ComputeTrend(analysis.TrendInput{Metric: string(metric), Unit: currentSeries.Unit, Samples: currentSeries.Samples, BaselineSamples: baselineSeries.Samples, RollingWindow: rolling, MinSamples: minSamples, BaselineMinSamples: minSamples, SampleGrain: grain})
		assumptions := analyzerMetaAssumptions(currentSeries.Assumptions, window.Window, args.IncludeFull)
		assumptions["baseline_window"] = baseline.Window
		assumptions["rolling_window_days"] = effectiveRollingDays
		if grain == analysis.SampleGrainWeekly {
			assumptions["rolling_bucket_count"] = rolling
		}
		formulaRef := ""
		if trend.ZScore != nil {
			formulaRef = resources.AnalysisFormulaRefZScore
		}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: trend, Series: points, Meta: analysis.AnalyzerMetaInput{Method: "ols_trend_with_baseline", SourceTools: mergeSourceTools(currentSeries, baselineSeries), N: trend.N, MissingDays: currentSeries.MissingDays, MinSamples: minSamples, FormulaRef: formulaRef, Assumptions: assumptions, Boundaries: trend.Boundaries}}, args.IncludeFull, version, debugMetadata, analyzeTrendName, unitSystem, shapeCfg)
	}
}

func analyzeTrendInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"metric", "window"}, "properties": map[string]any{"metric": analyzerMetricProperty("Metric to trend."), "window": analyzerWindowSchema("Current inclusive athlete-local window."), "baseline_window": analyzerWindowSchema("Optional baseline inclusive athlete-local window; defaults to same length immediately before window."), "rolling_window_days": map[string]any{"type": "integer", "default": 7, "minimum": 2, "maximum": 90, "description": "Trailing usable-sample rolling mean window. Must not exceed current window; weekly metrics require multiples of 7."}, "sport": map[string]any{"type": "string", "description": "Optional sport/category filter for activity-backed metrics."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include sampled series and rolling means; terse mode omits raw rows."}}}
}

func analyzerWindowSchema(description string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "description": description, "properties": map[string]any{"start_date": map[string]any{"type": "string", "description": "Inclusive athlete-local start date YYYY-MM-DD."}, "end_date": map[string]any{"type": "string", "description": "Inclusive athlete-local end date YYYY-MM-DD."}}}
}

var _ = response.UnitSystemMetric
