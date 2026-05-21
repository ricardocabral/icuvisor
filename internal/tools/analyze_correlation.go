package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/analysis"
)

const (
	analyzeCorrelationName        = "analyze_correlation"
	analyzeCorrelationDescription = "Use when the prompt asks whether two analysis metrics are correlated or lagged together; do not fetch get_* rows or streams and reduce them in chat. Computes Pearson or Spearman correlation plus OLS slope/intercept with paired-sample metadata."
	invalidAnalyzeCorrelationArgs = "invalid analyze_correlation arguments; provide metric_x, metric_y, window, optional method, pairing_grain, lag_days, sport, and include_full"
	fetchAnalyzeCorrelationMsg    = "could not analyze correlation; check credentials, date range, metrics, and sport filter"
)

type analyzeCorrelationRequest struct {
	MetricX      string                `json:"metric_x"`
	MetricY      string                `json:"metric_y"`
	Window       analyzerWindowRequest `json:"window"`
	Method       string                `json:"method,omitempty"`
	PairingGrain string                `json:"pairing_grain,omitempty"`
	LagDays      int                   `json:"lag_days,omitempty"`
	Sport        string                `json:"sport,omitempty"`
	IncludeFull  bool                  `json:"include_full,omitempty"`
}

func newAnalyzeCorrelationTool(fitness FitnessClient, wellness WellnessClient, activities ActivitiesClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	clients := analyzerClients{fitness: fitness, wellness: wellness, activities: activities}
	return fullTool(Tool{Name: analyzeCorrelationName, Description: analyzeCorrelationDescription, InputSchema: analyzeCorrelationInputSchema(), OutputSchema: genericOutputSchema("Analyzer correlation result with coefficient, slope, intercept, paired series, and analyzer _meta."), Handler: analyzeCorrelationHandler(clients, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func analyzeCorrelationHandler(clients analyzerClients, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAnalyzerStrict[analyzeCorrelationRequest](req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, err)
		}
		if args.Method == "" {
			args.Method = analysis.CorrelationPearson
		}
		if args.Method != analysis.CorrelationPearson && args.Method != analysis.CorrelationSpearman {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, fmt.Errorf("method must be pearson or spearman"))
		}
		if args.LagDays < -30 || args.LagDays > 30 {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, fmt.Errorf("lag_days must be -30..30"))
		}
		metricX, err := parseMetricArgument(args.MetricX)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, err)
		}
		metricY, err := parseMetricArgument(args.MetricY)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, err)
		}
		window, err := analysis.ParseWindow(args.Window.analysisWindow(), maxAnalyzerWindowDays)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, err)
		}
		grain := analysis.SampleGrainDaily
		if args.PairingGrain == "activity" {
			if args.LagDays != 0 {
				return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, fmt.Errorf("activity pairing requires lag_days=0"))
			}
			grain = analysis.SampleGrainActivity
		} else if args.PairingGrain != "" && args.PairingGrain != "daily" {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, fmt.Errorf("pairing_grain must be daily or activity"))
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchAnalyzeCorrelationMsg, err)
		}
		yWindow := window
		if grain == analysis.SampleGrainDaily && args.LagDays != 0 {
			shifted, err := analysis.ParseWindow(shiftedLookupWindow(window, args.LagDays), maxAnalyzerWindowDays+30)
			if err != nil {
				return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, err)
			}
			yWindow = shifted
		}
		xSeries, err := loadAnalyzerSeries(ctx, clients, metricX, window, grain, args.Sport, unitSystem, false)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnalyzeCorrelationMsg, err)
		}
		ySeries, err := loadAnalyzerSeries(ctx, clients, metricY, yWindow, grain, args.Sport, unitSystem, false)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnalyzeCorrelationMsg, err)
		}
		var pairs []analysis.PairedSample
		if grain == analysis.SampleGrainActivity {
			pairs = pairActivitySamples(xSeries, ySeries)
		} else {
			pairs = pairDailySamples(xSeries, ySeries, window, args.LagDays)
		}
		result := analysis.ComputeCorrelation(analysis.CorrelationInput{MetricX: string(metricX), MetricY: string(metricY), Method: args.Method, LagDays: args.LagDays, Pairs: pairs})
		assumptions := analyzerMetaAssumptions(map[string]any{"pairing_grain": string(grain), "lag_days": args.LagDays, "lookup_window_y": yWindow.Window, "anchor_metric": "metric_x"}, window.Window, args.IncludeFull)
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: pairs, Meta: analysis.AnalyzerMetaInput{Method: args.Method + "_correlation", SourceTools: mergeSourceTools(xSeries, ySeries), N: result.N, MissingDays: analysis.MissingSamples(window.Days, result.N), MinSamples: analysis.MinCorrelationSamples, Assumptions: assumptions, Boundaries: result.Boundaries}}, args.IncludeFull, version, debugMetadata, analyzeCorrelationName, unitSystem, shapeCfg)
	}
}

func analyzeCorrelationInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"metric_x", "metric_y", "window"}, "properties": map[string]any{"metric_x": analyzerMetricProperty("X metric."), "metric_y": analyzerMetricProperty("Y metric."), "window": analyzerWindowSchema("Inclusive athlete-local anchor window for metric_x."), "method": map[string]any{"type": "string", "enum": []string{analysis.CorrelationPearson, analysis.CorrelationSpearman}, "default": analysis.CorrelationPearson}, "pairing_grain": map[string]any{"type": "string", "enum": []string{"daily", "activity"}, "default": "daily"}, "lag_days": map[string]any{"type": "integer", "minimum": -30, "maximum": 30, "default": 0, "description": "Positive lag pairs x on D with y on D+lag."}, "sport": map[string]any{"type": "string", "description": "Optional sport/category filter for activity-backed metrics."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include paired samples; terse mode omits raw rows."}}}
}
