package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/analysis"
)

const (
	analyzeDistributionName        = "analyze_distribution"
	analyzeDistributionDescription = "Use when the prompt asks for an analysis metric's distribution, histogram, quantiles, or outliers; do not fetch rows and reduce them in chat. Computes deterministic stats and buckets with analyzer missing-sample metadata."
	invalidAnalyzeDistributionArgs = "invalid analyze_distribution arguments; provide metric, window, optional bucket_count or buckets, quantiles, sport, and include_full"
	fetchAnalyzeDistributionMsg    = "could not analyze distribution; check credentials, date range, metric, and sport filter"
)

type analyzeDistributionRequest struct {
	Metric      string                `json:"metric"`
	Window      analyzerWindowRequest `json:"window"`
	BucketCount *int                  `json:"bucket_count,omitempty"`
	Buckets     []float64             `json:"buckets,omitempty"`
	Quantiles   []float64             `json:"quantiles,omitempty"`
	Sport       string                `json:"sport,omitempty"`
	IncludeFull bool                  `json:"include_full,omitempty"`
}

func newAnalyzeDistributionTool(fitness FitnessClient, wellness WellnessClient, activities ActivitiesClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	clients := analyzerClients{fitness: fitness, wellness: wellness, activities: activities}
	return fullTool(Tool{Name: analyzeDistributionName, Description: analyzeDistributionDescription, InputSchema: analyzeDistributionInputSchema(), OutputSchema: genericOutputSchema("Analyzer distribution stats, quantiles, histogram buckets, and analyzer _meta."), Handler: analyzeDistributionHandler(clients, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func analyzeDistributionHandler(clients analyzerClients, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAnalyzerStrict[analyzeDistributionRequest](req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeDistributionArgs, err)
		}
		bucketCount := 0
		if len(args.Buckets) > 0 {
			if args.BucketCount != nil {
				return Result{}, NewUserError(invalidAnalyzeDistributionArgs, fmt.Errorf("bucket_count and buckets are mutually exclusive"))
			}
		} else if args.BucketCount == nil {
			bucketCount = 10
		} else {
			bucketCount = *args.BucketCount
		}
		if len(args.Buckets) == 0 && (bucketCount < 2 || bucketCount > 50) {
			return Result{}, NewUserError(invalidAnalyzeDistributionArgs, fmt.Errorf("bucket_count must be 2..50"))
		}
		if len(args.Quantiles) == 0 {
			args.Quantiles = []float64{0.25, 0.5, 0.75}
		}
		for _, quantile := range args.Quantiles {
			if quantile < 0 || quantile > 1 {
				return Result{}, NewUserError(invalidAnalyzeDistributionArgs, fmt.Errorf("quantiles must be between 0 and 1"))
			}
		}
		metric, err := parseMetricArgument(args.Metric)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeDistributionArgs, err)
		}
		window, err := analysis.ParseWindow(args.Window.analysisWindow(), maxAnalyzerWindowDays)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeDistributionArgs, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchAnalyzeDistributionMsg, err)
		}
		grain := analysis.SampleGrainDaily
		selection, err := selectAnalyzerMetricSource(metric, analysis.SampleGrainActivity, true)
		if err == nil && selection.Source.Family == analysis.SourceActivityRow {
			grain = analysis.SampleGrainActivity
		}
		series, err := loadAnalyzerSeries(ctx, clients, metric, window, grain, args.Sport, unitSystem, true)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnalyzeDistributionMsg, err)
		}
		actualGrain, _ := series.Assumptions["sample_grain"].(string)
		if actualGrain == string(analysis.SampleGrainWeekly) {
			grain = analysis.SampleGrainWeekly
		}
		result := analysis.ComputeDistribution(analysis.DistributionInput{Metric: string(metric), Unit: series.Unit, Samples: series.Samples, BucketCount: bucketCount, Buckets: sortedFloatCopy(args.Buckets), Quantiles: args.Quantiles, SampleGrain: grain})
		assumptions := analyzerMetaAssumptions(series.Assumptions, window.Window, args.IncludeFull)
		if len(args.Buckets) == 0 {
			assumptions["bucket_count"] = bucketCount
		} else {
			assumptions["bucket_boundaries"] = sortedFloatCopy(args.Buckets)
		}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: series.Samples, Meta: analysis.AnalyzerMetaInput{Method: "distribution_histogram_quantiles", SourceTools: series.SourceTools, N: result.Stats.N, MissingDays: series.MissingDays, MinSamples: 3, Assumptions: assumptions}}, args.IncludeFull, version, debugMetadata, analyzeDistributionName, unitSystem, shapeCfg)
	}
}

func analyzeDistributionInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"metric", "window"}, "properties": map[string]any{"metric": analyzerMetricProperty("Metric to summarize."), "window": analyzerWindowSchema("Inclusive athlete-local window."), "bucket_count": map[string]any{"type": "integer", "default": 10, "minimum": 2, "maximum": 50, "description": "Number of equal-width histogram buckets when buckets is omitted."}, "buckets": map[string]any{"type": "array", "items": map[string]any{"type": "number"}, "description": "Optional numeric bucket boundaries; values outside are counted below_range/above_range."}, "quantiles": map[string]any{"type": "array", "items": map[string]any{"type": "number", "minimum": 0, "maximum": 1}, "description": "Quantiles to compute; defaults to 0.25, 0.5, 0.75."}, "sport": map[string]any{"type": "string", "description": "Optional sport/category filter for activity-backed metrics."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include sampled values; terse mode omits raw rows."}}}
}
