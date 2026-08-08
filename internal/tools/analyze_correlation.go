package tools

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	analyzeCorrelationName        = "analyze_correlation"
	analyzeCorrelationDescription = "Use when the prompt asks whether two analysis metrics are correlated or lagged together, including explicitly requested activity custom fields via metric custom:<field_code> plus custom_fields; do not fetch get_* rows or streams and reduce them in chat. Computes Pearson or Spearman correlation plus OLS slope/intercept with paired-sample metadata."
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
	CustomFields []string              `json:"custom_fields,omitempty"`
	IncludeFull  bool                  `json:"include_full,omitempty"`
}

func newAnalyzeCorrelationTool(fitness FitnessClient, wellness WellnessClient, activities ActivitiesClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	clients := newAnalyzerClients(fitness, wellness, activities)
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
		customFieldCodes, err := selectedActivityCustomFieldCodes(ctx, clients.customFields, clients.customFieldCache, args.CustomFields)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(activityCustomFieldSelectionMessage(err, invalidAnalyzeCorrelationArgs), err)
		}
		metricX, err := parseCorrelationMetricArgument(args.MetricX, customFieldCodes)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeCorrelationArgs, err)
		}
		metricY, err := parseCorrelationMetricArgument(args.MetricY, customFieldCodes)
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
		xSeries, err := loadCorrelationMetricSeries(ctx, clients, metricX, window, grain, args.Sport, unitSystem, customFieldCodes)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, analyzerFetchUserError(fetchAnalyzeCorrelationMsg, err)
		}
		ySeries, err := loadCorrelationMetricSeries(ctx, clients, metricY, yWindow, grain, args.Sport, unitSystem, customFieldCodes)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, analyzerFetchUserError(fetchAnalyzeCorrelationMsg, err)
		}
		var pairs []analysis.PairedSample
		if grain == analysis.SampleGrainActivity {
			pairs = pairActivitySamples(xSeries, ySeries)
		} else {
			pairs = pairDailySamples(xSeries, ySeries, window, args.LagDays)
		}
		result := analysis.ComputeCorrelation(analysis.CorrelationInput{MetricX: metricX.Name, MetricY: metricY.Name, Method: args.Method, LagDays: args.LagDays, Pairs: pairs})
		assumptions := analyzerMetaAssumptions(map[string]any{"pairing_grain": string(grain), "lag_days": args.LagDays, "lookup_window_y": yWindow.Window, "anchor_metric": "metric_x", "series": map[string]any{metricX.Name: xSeries.Assumptions, metricY.Name: ySeries.Assumptions}}, window.Window, args.IncludeFull)
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: pairs, Meta: analysis.AnalyzerMetaInput{Method: args.Method + "_correlation", SourceTools: mergeSourceTools(xSeries, ySeries), N: result.N, MissingDays: analysis.MissingSamples(window.Days, result.N), MinSamples: analysis.MinCorrelationSamples, Assumptions: assumptions, Boundaries: result.Boundaries}}, args.IncludeFull, version, debugMetadata, analyzeCorrelationName, unitSystem, shapeCfg)
	}
}

type correlationMetricRef struct {
	Name        string
	Standard    analysis.Metric
	CustomField string
}

func parseCorrelationMetricArgument(value string, selectedCustomFields []string) (correlationMetricRef, error) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(trimmed), "custom:") {
		field := strings.TrimSpace(trimmed[len("custom:"):])
		if field == "" {
			return correlationMetricRef{}, fmt.Errorf("custom metric must use custom:<field_code>")
		}
		if !slices.Contains(selectedCustomFields, field) {
			return correlationMetricRef{}, fmt.Errorf("custom metric %q requires listing %q in custom_fields", "custom:"+field, field)
		}
		return correlationMetricRef{Name: "custom:" + field, CustomField: field}, nil
	}
	metric, err := parseMetricArgument(trimmed)
	if err != nil {
		return correlationMetricRef{}, err
	}
	return correlationMetricRef{Name: string(metric), Standard: metric}, nil
}

func loadCorrelationMetricSeries(ctx context.Context, clients analyzerClients, metric correlationMetricRef, window analysis.ParsedWindow, grain analysis.SampleGrain, sport string, unitSystem response.UnitSystem, customFieldCodes []string) (analyzerSampleSeries, error) {
	if metric.CustomField != "" {
		return loadCustomFieldAnalyzerSeries(ctx, clients, metric.CustomField, window, grain, sport, customFieldCodes)
	}
	return loadAnalyzerSeries(ctx, clients, metric.Standard, window, grain, sport, unitSystem, customFieldCodes, false)
}

func analyzeCorrelationInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"metric_x", "metric_y", "window"}, "properties": map[string]any{"metric_x": correlationMetricProperty("X metric."), "metric_y": correlationMetricProperty("Y metric."), "window": analyzerWindowSchema("Inclusive athlete-local anchor window for metric_x."), "method": map[string]any{"type": "string", "enum": []string{analysis.CorrelationPearson, analysis.CorrelationSpearman}, "default": analysis.CorrelationPearson}, "pairing_grain": map[string]any{"type": "string", "enum": []string{"daily", "activity"}, "default": "daily"}, "lag_days": map[string]any{"type": "integer", "minimum": -30, "maximum": 30, "default": 0, "description": "Positive lag pairs x on D with y on D+lag."}, "sport": map[string]any{"type": "string", "description": "Optional sport/category filter for activity-backed metrics."}, "custom_fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": maxSelectedActivityCustomFields, "description": "Optional athlete-defined activity custom field codes to fetch for custom:<field> correlation metrics; defaults to none."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include paired samples; terse mode omits raw rows."}}}
}

func correlationMetricProperty(description string) map[string]any {
	return map[string]any{"type": "string", "description": description + " Use a canonical analysis_metric enum value or custom:<field_code> when that code is also listed in custom_fields."}
}
