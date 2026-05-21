package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/streams"
)

const (
	getActivityHistogramName        = "get_activity_histogram"
	getActivityHistogramDescription = "Use when the prompt asks for a single activity's power, heart-rate, or pace distribution; do not fetch get_activity_streams samples and bin them in chat. Summarizes terse time-in-bucket histogram rows."
	invalidActivityHistogramMessage = "invalid get_activity_histogram arguments; provide activity_id and metric"
)

type getActivityHistogramRequest struct {
	ActivityID  string `json:"activity_id"`
	Metric      string `json:"metric"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getActivityHistogramResponse struct {
	ActivityID  string                     `json:"activity_id"`
	Metric      string                     `json:"metric"`
	Buckets     []analysis.HistogramBucket `json:"buckets"`
	Unavailable *histogramUnavailable      `json:"unavailable,omitempty"`
	Meta        activityHistogramMeta      `json:"_meta"`
}

type histogramUnavailable struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type activityHistogramMeta struct {
	analysis.AnalyzerMeta
	BucketMethod string                        `json:"bucket_method,omitempty"`
	EmittedUnit  string                        `json:"emitted_unit,omitempty"`
	ZoneSource   *analysis.HistogramZoneSource `json:"zone_source,omitempty"`
	FixedWidth   *analysis.HistogramFixedWidth `json:"fixed_width,omitempty"`
	Warnings     []string                      `json:"warnings,omitempty"`
}

func newGetActivityHistogramTool(streamsClient ActivityStreamsClient, detailsClient ActivityDetailsClient, profileClient ProfileClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getActivityHistogramName, Description: getActivityHistogramDescription, InputSchema: activityHistogramInputSchema(), OutputSchema: genericOutputSchema("Single-activity histogram buckets with analyzer metadata."), Handler: getActivityHistogramHandler(streamsClient, detailsClient, profileClient, version, debugMetadata, shapeCfg)})
}

func getActivityHistogramHandler(streamsClient ActivityStreamsClient, detailsClient ActivityDetailsClient, profileClient ProfileClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, metric, err := decodeActivityHistogramRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidActivityHistogramMessage, err)
		}
		if streamsClient == nil {
			return Result{}, NewUserError("could not get activity histogram", errors.New("missing activity streams client"))
		}

		sourceTools := []string{getActivityStreamsName}
		var warnings []string
		var activity intervals.Activity
		var profile intervals.AthleteWithSportSettings
		unitSystem := response.UnitSystemMetric
		profileAvailable := false
		if detailsClient != nil {
			sourceTools = append(sourceTools, getActivityDetailsName)
			activity, err = detailsClient.GetActivity(ctx, args.ActivityID)
			if err != nil {
				if isContextError(err) {
					return Result{}, err
				}
				warnings = append(warnings, "activity details unavailable; fixed-width fallback used when zones cannot be selected")
			}
		}
		if profileClient != nil {
			sourceTools = append(sourceTools, getAthleteProfileName)
			profile, err = profileClient.GetAthleteProfile(ctx)
			if err != nil {
				if isContextError(err) {
					return Result{}, err
				}
				warnings = append(warnings, "athlete profile unavailable; fixed-width fallback used and pace defaults to seconds_per_km")
			} else {
				profileAvailable = true
				unitSystem = profileUnitSystem(profile)
			}
		}

		streamRows, err := streamsClient.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: args.ActivityID, Types: histogramStreamTypes(metric), IncludeDefaults: false})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			payload := unavailableActivityHistogramResponse(args.ActivityID, metric, "stream_fetch_failed", "could not fetch required activity streams", sourceTools, 0, warnings)
			return encodeActivityHistogramResponse(payload, args.IncludeFull, version, debugMetadata, unitSystem, shapeCfg)
		}
		streamMap := canonicalActivityStreamData(streamRows)
		emittedUnit := histogramEmittedUnit(metric, unitSystem)
		samples, unavailable := histogramSamples(metric, streamMap, emittedUnit)
		if unavailable != nil {
			payload := unavailableActivityHistogramResponse(args.ActivityID, metric, unavailable.Reason, unavailable.Message, sourceTools, 0, warnings)
			payload.Meta.EmittedUnit = emittedUnit
			return encodeActivityHistogramResponse(payload, args.IncludeFull, version, debugMetadata, unitSystem, shapeCfg)
		}

		zoneConfig := histogramZoneConfig(metric, emittedUnit, activity, profile, profileAvailable)
		result := analysis.BuildHistogram(samples, histogramUnitLabel(metric, emittedUnit), zoneConfig)
		if result.N == 0 || len(result.Buckets) == 0 {
			payload := unavailableActivityHistogramResponse(args.ActivityID, metric, "insufficient_sample", "required streams did not contain valid positive-duration intervals", sourceTools, result.N, warnings)
			payload.Meta.EmittedUnit = emittedUnit
			return encodeActivityHistogramResponse(payload, args.IncludeFull, version, debugMetadata, unitSystem, shapeCfg)
		}

		payload := getActivityHistogramResponse{
			ActivityID: args.ActivityID,
			Metric:     string(metric),
			Buckets:    result.Buckets,
			Meta: activityHistogramMeta{
				AnalyzerMeta: analysis.NewAnalyzerMeta(analysis.AnalyzerMetaInput{Method: "activity_stream_histogram", SourceTools: sourceTools, N: result.N, MissingDays: 0, MissingAction: analysis.MissingActionSkip, MinSamples: 1}),
				BucketMethod: result.BucketMethod,
				EmittedUnit:  emittedUnit,
				ZoneSource:   result.ZoneSource,
				FixedWidth:   result.FixedWidth,
				Warnings:     warnings,
			},
		}
		return encodeActivityHistogramResponse(payload, args.IncludeFull, version, debugMetadata, unitSystem, shapeCfg)
	}
}

func decodeActivityHistogramRequest(raw json.RawMessage) (getActivityHistogramRequest, analysis.HistogramMetric, error) {
	args, err := DecodeStrict[getActivityHistogramRequest](raw)
	if err != nil {
		return args, "", err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	if args.ActivityID == "" {
		return args, "", errors.New("activity_id is required")
	}
	metric, err := analysis.ParseHistogramMetric(args.Metric)
	if err != nil {
		return args, "", err
	}
	return args, metric, nil
}

func activityHistogramInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id", "metric"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose stream distribution should be summarized."},
		"metric":       analysis.HistogramMetricSchemaProperty(),
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, preserve diagnostic metadata. Raw stream samples are never returned by this histogram tool."},
	}}
}

func histogramStreamTypes(metric analysis.HistogramMetric) []string {
	switch metric {
	case analysis.HistogramMetricPowerWatts:
		return []string{"watts", "time"}
	case analysis.HistogramMetricHeartRateBPM:
		return []string{"heart_rate", "time"}
	case analysis.HistogramMetricPaceSeconds:
		return []string{"distance", "time"}
	default:
		return nil
	}
}

func canonicalActivityStreamData(rows []intervals.ActivityStream) map[string][]float64 {
	out := map[string][]float64{}
	for _, row := range rows {
		key, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if key == "" || len(row.Data) == 0 {
			continue
		}
		out[key] = row.Data
	}
	return out
}

func histogramSamples(metric analysis.HistogramMetric, streamMap map[string][]float64, emittedUnit string) ([]analysis.HistogramSample, *histogramUnavailable) {
	switch metric {
	case analysis.HistogramMetricPowerWatts:
		return valueTimeSamples(streamMap["watts"], streamMap["time"], "power stream is missing or not aligned with time")
	case analysis.HistogramMetricHeartRateBPM:
		return valueTimeSamples(streamMap["heart_rate"], streamMap["time"], "heart-rate stream is missing or not aligned with time")
	case analysis.HistogramMetricPaceSeconds:
		return paceSamples(streamMap["distance"], streamMap["time"], emittedUnit)
	default:
		return nil, &histogramUnavailable{Reason: "invalid_metric", Message: "unsupported histogram metric"}
	}
}

func valueTimeSamples(values []float64, times []float64, message string) ([]analysis.HistogramSample, *histogramUnavailable) {
	if len(values) < 2 || len(times) < 2 || len(values) != len(times) {
		return nil, &histogramUnavailable{Reason: "missing_stream", Message: message}
	}
	samples := make([]analysis.HistogramSample, 0, len(values)-1)
	for i := 0; i+1 < len(values); i++ {
		samples = append(samples, analysis.HistogramSample{Value: values[i], Seconds: times[i+1] - times[i]})
	}
	return samples, nil
}

func paceSamples(distance []float64, times []float64, emittedUnit string) ([]analysis.HistogramSample, *histogramUnavailable) {
	if len(distance) < 2 || len(times) < 2 || len(distance) != len(times) {
		return nil, &histogramUnavailable{Reason: "missing_stream", Message: "distance/time streams are missing or not aligned for pace"}
	}
	samples := make([]analysis.HistogramSample, 0, len(distance)-1)
	for i := 0; i+1 < len(distance); i++ {
		dt := times[i+1] - times[i]
		dd := distance[i+1] - distance[i]
		if emittedUnit == "seconds_per_mile" {
			samples = append(samples, analysis.HistogramSample{Value: dt / (dd / 1609.344), Seconds: dt})
		} else {
			samples = append(samples, analysis.HistogramSample{Value: dt / (dd / 1000), Seconds: dt})
		}
	}
	return samples, nil
}

func histogramEmittedUnit(metric analysis.HistogramMetric, unitSystem response.UnitSystem) string {
	if metric == analysis.HistogramMetricPaceSeconds {
		if unitSystem == response.UnitSystemImperial {
			return "seconds_per_mile"
		}
		return "seconds_per_km"
	}
	return histogramUnitLabel(metric, "")
}

func histogramUnitLabel(metric analysis.HistogramMetric, emittedUnit string) string {
	switch metric {
	case analysis.HistogramMetricPowerWatts:
		return "W"
	case analysis.HistogramMetricHeartRateBPM:
		return "bpm"
	case analysis.HistogramMetricPaceSeconds:
		return emittedUnit
	default:
		return ""
	}
}

func histogramZoneConfig(metric analysis.HistogramMetric, emittedUnit string, activity intervals.Activity, profile intervals.AthleteWithSportSettings, profileAvailable bool) *analysis.HistogramZoneConfig {
	if !profileAvailable {
		return nil
	}
	setting, ok := selectHistogramSportSetting(activity, profile.SportSettings)
	if !ok {
		return nil
	}
	config := analysis.HistogramZoneConfig{Sport: strings.TrimSpace(setting.Type), SportSettingID: setting.ID, Metric: string(metric), Unit: histogramUnitLabel(metric, emittedUnit)}
	switch metric {
	case analysis.HistogramMetricPowerWatts:
		for _, boundary := range setting.PowerZones {
			config.Boundaries = append(config.Boundaries, float64(boundary))
		}
		config.Names = append([]string(nil), setting.PowerZoneNames...)
	case analysis.HistogramMetricHeartRateBPM:
		for _, boundary := range setting.HRZones {
			config.Boundaries = append(config.Boundaries, float64(boundary))
		}
		config.Names = append([]string(nil), setting.HRZoneNames...)
	case analysis.HistogramMetricPaceSeconds:
		if strings.TrimSpace(setting.PaceUnits) == "" {
			return nil
		}
		for _, boundary := range setting.PaceZones {
			converted, ok := analysis.ConvertPaceZoneBoundary(boundary, setting.PaceUnits, emittedUnit)
			if !ok {
				return nil
			}
			config.Boundaries = append(config.Boundaries, converted)
		}
		config.Names = append([]string(nil), setting.PaceZoneNames...)
		config.PaceUnits = setting.PaceUnits
	}
	if len(config.Boundaries) == 0 {
		return nil
	}
	return &config
}

func selectHistogramSportSetting(activity intervals.Activity, settings []intervals.SportSettings) (intervals.SportSettings, bool) {
	candidates := []string{}
	if activity.Type != nil {
		candidates = append(candidates, *activity.Type)
	}
	if activity.SubType != nil {
		candidates = append(candidates, *activity.SubType)
	}
	for _, candidate := range candidates {
		needle := strings.ToLower(strings.TrimSpace(candidate))
		if needle == "" {
			continue
		}
		for _, setting := range settings {
			if strings.ToLower(strings.TrimSpace(setting.Type)) == needle {
				return setting, true
			}
		}
		for _, setting := range settings {
			for _, value := range setting.Types {
				if strings.ToLower(strings.TrimSpace(value)) == needle {
					return setting, true
				}
			}
		}
	}
	return intervals.SportSettings{}, false
}

func unavailableActivityHistogramResponse(activityID string, metric analysis.HistogramMetric, reason string, message string, sourceTools []string, n int, warnings []string) getActivityHistogramResponse {
	return getActivityHistogramResponse{
		ActivityID:  activityID,
		Metric:      string(metric),
		Buckets:     []analysis.HistogramBucket{},
		Unavailable: &histogramUnavailable{Reason: reason, Message: message},
		Meta:        activityHistogramMeta{AnalyzerMeta: analysis.NewAnalyzerMeta(analysis.AnalyzerMetaInput{Method: "activity_stream_histogram", SourceTools: sourceTools, N: n, MissingDays: 0, MissingAction: analysis.MissingActionSkip, MinSamples: 1}), Warnings: warnings},
	}
}

func encodeActivityHistogramResponse(payload getActivityHistogramResponse, includeFull bool, version string, debugMetadata bool, unitSystem response.UnitSystem, shapeCfg responseShaping) (Result, error) {
	shaped, err := response.Shape(payload, shapeCfg.options(includeFull, []string{"buckets"}, version, debugMetadata, getActivityHistogramName, unitSystem))
	if err != nil {
		return Result{}, fmt.Errorf("shaping %s response: %w", getActivityHistogramName, err)
	}
	return TextResult(shaped), nil
}
