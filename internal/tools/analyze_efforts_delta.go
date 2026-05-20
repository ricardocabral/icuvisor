package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	analyzeEffortsDeltaName        = "analyze_efforts_delta"
	analyzeEffortsDeltaDescription = "Use when the prompt asks whether best-effort power, heart-rate, or pace buckets changed versus baseline; do not fetch rows and reduce them in chat. Compares current and baseline curve buckets with unit-explicit deltas."
	invalidAnalyzeEffortsDeltaArgs = "invalid analyze_efforts_delta arguments; provide sport, effort_family, current_window, optional baseline_window, and matching duration/distance buckets"
	fetchAnalyzeEffortsDeltaMsg    = "could not analyze efforts delta; check credentials, sport, date range, and buckets"
)

type analyzeEffortsDeltaRequest struct {
	Sport           string                 `json:"sport"`
	EffortFamily    string                 `json:"effort_family"`
	DurationSeconds []int                  `json:"duration_seconds,omitempty"`
	DistanceMeters  []int                  `json:"distance_meters,omitempty"`
	CurrentWindow   analyzerWindowRequest  `json:"current_window"`
	BaselineWindow  *analyzerWindowRequest `json:"baseline_window,omitempty"`
	IncludeFull     bool                   `json:"include_full,omitempty"`
}

func newAnalyzeEffortsDeltaTool(client BestEffortsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: analyzeEffortsDeltaName, Description: analyzeEffortsDeltaDescription, InputSchema: analyzeEffortsDeltaInputSchema(), OutputSchema: genericOutputSchema("Best-efforts current-vs-baseline delta buckets with analyzer _meta."), Handler: analyzeEffortsDeltaHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func analyzeEffortsDeltaHandler(client BestEffortsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAnalyzerStrict[analyzeEffortsDeltaRequest](req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeEffortsDeltaArgs, err)
		}
		args.Sport = strings.TrimSpace(args.Sport)
		if args.Sport == "" {
			return Result{}, NewUserError(invalidAnalyzeEffortsDeltaArgs, fmt.Errorf("sport is required"))
		}
		if err := normalizeEffortBuckets(&args); err != nil {
			return Result{}, NewUserError(invalidAnalyzeEffortsDeltaArgs, err)
		}
		currentWindow, err := analysis.ParseWindow(args.CurrentWindow.analysisWindow(), maxAnalyzerWindowDays)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeEffortsDeltaArgs, err)
		}
		baselineWindow := analysis.DefaultBaselineWindow(currentWindow)
		if args.BaselineWindow != nil {
			baselineWindow = args.BaselineWindow.analysisWindow()
		}
		baseline, err := analysis.ParseWindow(baselineWindow, maxAnalyzerWindowDays)
		if err != nil {
			return Result{}, NewUserError(invalidAnalyzeEffortsDeltaArgs, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchAnalyzeEffortsDeltaMsg, err)
		}
		current, baselineValues, sourceTool, err := loadEffortBucketValues(ctx, client, args, currentWindow, baseline)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnalyzeEffortsDeltaMsg, err)
		}
		result := analysis.ComputeEffortsDelta(analysis.EffortsDeltaInput{Sport: args.Sport, Family: args.EffortFamily, UnitSystem: string(unitSystem), Current: current, Baseline: baselineValues})
		assumptions := map[string]any{"current_window": currentWindow.Window, "baseline_window": baseline.Window, "missing_days_applicable": false, "unit_system": string(unitSystem), "better_direction": result.BetterDirection}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: map[string]any{"current": current, "baseline": baselineValues}, Meta: analysis.AnalyzerMetaInput{Method: "best_efforts_current_vs_baseline", SourceTools: []string{sourceTool}, N: result.N, MissingDays: 0, MinSamples: 1, Assumptions: assumptions}}, args.IncludeFull, version, debugMetadata, analyzeEffortsDeltaName, unitSystem, shapeCfg)
	}
}

func normalizeEffortBuckets(args *analyzeEffortsDeltaRequest) error {
	args.EffortFamily = strings.TrimSpace(args.EffortFamily)
	if args.EffortFamily == "" {
		args.EffortFamily = analysis.EffortFamilyPower
	}
	if len(args.DurationSeconds) > 24 || len(args.DistanceMeters) > 24 {
		return fmt.Errorf("at most 24 buckets are allowed")
	}
	switch args.EffortFamily {
	case analysis.EffortFamilyPower, analysis.EffortFamilyHeartRate:
		if len(args.DistanceMeters) > 0 {
			return fmt.Errorf("distance_meters is only valid for pace")
		}
		args.DurationSeconds = normalizePositiveInts(args.DurationSeconds, defaultDurationBuckets)
		for _, value := range args.DurationSeconds {
			if value > 86400 {
				return fmt.Errorf("duration_seconds values must be <= 86400")
			}
		}
	case analysis.EffortFamilyPace:
		if len(args.DurationSeconds) > 0 {
			return fmt.Errorf("duration_seconds is only valid for power or heart_rate")
		}
		if len(args.DistanceMeters) == 0 {
			return fmt.Errorf("distance_meters is required for pace")
		}
		args.DistanceMeters = normalizePositiveInts(args.DistanceMeters, nil)
		for _, value := range args.DistanceMeters {
			if value > 100000 {
				return fmt.Errorf("distance_meters values must be <= 100000")
			}
		}
	default:
		return fmt.Errorf("effort_family must be power, heart_rate, or pace")
	}
	return nil
}

func loadEffortBucketValues(ctx context.Context, client BestEffortsClient, args analyzeEffortsDeltaRequest, current analysis.ParsedWindow, baseline analysis.ParsedWindow) ([]analysis.EffortBucketValue, []analysis.EffortBucketValue, string, error) {
	currentSpec := rangeCurveSpec(current.StartDate, current.EndDate)
	baselineSpec := rangeCurveSpec(baseline.StartDate, baseline.EndDate)
	switch args.EffortFamily {
	case analysis.EffortFamilyPower:
		cur, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: currentSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			return nil, nil, "", err
		}
		base, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: baselineSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			return nil, nil, "", err
		}
		return effortBucketValues(durationCurveBucketValues(firstCurve(cur), args.DurationSeconds)), effortBucketValues(durationCurveBucketValues(firstCurve(base), args.DurationSeconds)), getPowerCurvesName, nil
	case analysis.EffortFamilyHeartRate:
		cur, err := client.ListAthleteHRCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: currentSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			return nil, nil, "", err
		}
		base, err := client.ListAthleteHRCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: baselineSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			return nil, nil, "", err
		}
		return effortBucketValues(durationCurveBucketValues(firstCurve(cur), args.DurationSeconds)), effortBucketValues(durationCurveBucketValues(firstCurve(base), args.DurationSeconds)), getHRCurvesName, nil
	default:
		cur, err := client.ListAthletePaceCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: currentSpec, DistanceMeters: args.DistanceMeters})
		if err != nil {
			return nil, nil, "", err
		}
		base, err := client.ListAthletePaceCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: baselineSpec, DistanceMeters: args.DistanceMeters})
		if err != nil {
			return nil, nil, "", err
		}
		return effortBucketValues(distanceCurveBucketValues(firstCurve(cur), args.DistanceMeters)), effortBucketValues(distanceCurveBucketValues(firstCurve(base), args.DistanceMeters)), getPaceCurvesName, nil
	}
}

func effortBucketValues(values []curveBucketValue, _ []int) []analysis.EffortBucketValue {
	out := make([]analysis.EffortBucketValue, 0, len(values))
	for _, value := range values {
		out = append(out, analysis.EffortBucketValue{Bucket: value.Bucket, Value: value.Value, ActivityID: value.ActivityID})
	}
	return out
}

func analyzeEffortsDeltaInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"sport", "effort_family", "current_window"}, "properties": map[string]any{"sport": map[string]any{"type": "string", "description": "Intervals.icu sport/type to compare, e.g. Ride or Run."}, "effort_family": map[string]any{"type": "string", "enum": []string{analysis.EffortFamilyPower, analysis.EffortFamilyHeartRate, analysis.EffortFamilyPace}}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1, "maximum": 86400}, "description": "Required/defaulted for power or heart_rate; invalid for pace."}, "distance_meters": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1, "maximum": 100000}, "description": "Required for pace; invalid for power or heart_rate."}, "current_window": analyzerWindowSchema("Current inclusive athlete-local best-efforts window."), "baseline_window": analyzerWindowSchema("Optional baseline inclusive athlete-local window; defaults to same length immediately before current_window."), "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include current/baseline bucket series; terse mode omits raw curve arrays."}}}
}
