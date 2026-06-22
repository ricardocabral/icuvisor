package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/streams"
)

const (
	computeActivitySegmentStatsName        = "compute_activity_segment_stats"
	computeActivitySegmentStatsDescription = "Use when the prompt asks for mean, median, p90, normalized power, intensity factor, drift, or decoupling over an explicit activity segment, including first-vs-last distance comparisons, as the analyzer-family raw-stream exception, or when get_activity_intervals reports a single averaged/collapsed lap caveat and the user needs stream-based evidence before judging interval execution; do not fetch get_activity_streams samples and reduce them in chat. Computes deterministic stats from canonical raw streams."
	invalidActivitySegmentStatsMessage     = "invalid compute_activity_segment_stats arguments; provide activity_id, one stat, exactly one time or distance range, and required metric or ftp_watts only when applicable"
	computeActivitySegmentStatsMessage     = "could not compute activity segment stats"
)

type computeActivitySegmentStatsRequest struct {
	ActivityID     string   `json:"activity_id"`
	Stat           string   `json:"stat"`
	Metric         string   `json:"metric,omitempty"`
	StartSeconds   *float64 `json:"start_seconds,omitempty"`
	EndSeconds     *float64 `json:"end_seconds,omitempty"`
	StartDistanceM *float64 `json:"start_distance_m,omitempty"`
	EndDistanceM   *float64 `json:"end_distance_m,omitempty"`
	FTPWatts       *float64 `json:"ftp_watts,omitempty"`
	IncludeFull    bool     `json:"include_full,omitempty"`
}

type activitySegmentStatsResult struct {
	ActivityID         string                 `json:"activity_id"`
	Stat               string                 `json:"stat"`
	Metric             string                 `json:"metric,omitempty"`
	Value              *float64               `json:"value,omitempty"`
	Unit               string                 `json:"unit,omitempty"`
	Segment            analysis.SegmentBounds `json:"segment"`
	MinSamples         int                    `json:"min_samples"`
	InsufficientSample bool                   `json:"insufficient_sample"`
	StreamsUsed        []string               `json:"streams_used"`
	Details            map[string]float64     `json:"details,omitempty"`
}

func newComputeActivitySegmentStatsTool(client ActivityStreamsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: computeActivitySegmentStatsName, Description: computeActivitySegmentStatsDescription, InputSchema: computeActivitySegmentStatsInputSchema(), OutputSchema: genericOutputSchema("Activity segment statistic with analyzer metadata."), Handler: computeActivitySegmentStatsHandler(client, version, debugMetadata, shapeCfg)})
}

func computeActivitySegmentStatsHandler(client ActivityStreamsClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args computeActivitySegmentStatsRequest
		if err := decodeJSONArgs(req.Arguments, &args); err != nil || strings.TrimSpace(args.ActivityID) == "" {
			return Result{}, NewUserError(invalidActivitySegmentStatsMessage, err)
		}
		input, requiredKeys, err := activitySegmentStatsInput(args)
		if err != nil {
			return Result{}, NewUserError(invalidActivitySegmentStatsMessage, err)
		}
		rows, err := client.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: args.ActivityID, Types: upstreamSegmentStreamTypes(requiredKeys), IncludeDefaults: false})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(computeActivitySegmentStatsMessage, err)
		}
		input.Streams, err = canonicalActivitySegmentStreams(rows, requiredKeys)
		if err != nil {
			if errors.Is(err, analysis.ErrMissingSegmentStream) {
				return Result{}, NewUserError(missingActivitySegmentStreamMessage(err), err)
			}
			return Result{}, NewUserError(computeActivitySegmentStatsMessage, err)
		}
		computed, err := analysis.ComputeActivitySegmentStats(input)
		if err != nil {
			message := computeActivitySegmentStatsMessage
			if errors.Is(err, analysis.ErrSegmentOutOfRange) {
				message = "activity segment range is outside available stream coverage"
			} else if errors.Is(err, analysis.ErrInvalidSegmentStatsInput) || errors.Is(err, analysis.ErrMissingSegmentStream) {
				message = invalidActivitySegmentStatsMessage
			}
			return Result{}, NewUserError(message, err)
		}
		result := activitySegmentStatsResult{ActivityID: args.ActivityID, Stat: computed.Stat, Metric: computed.Metric, Value: computed.Value, Unit: computed.Unit, Segment: computed.Segment, MinSamples: computed.MinSamples, InsufficientSample: computed.InsufficientSample, StreamsUsed: computed.StreamsUsed, Details: computed.Details}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: computed.Audit, Meta: analysis.AnalyzerMetaInput{Method: computed.Method, SourceTools: []string{getActivityStreamsName}, N: computed.N, MinSamples: computed.MinSamples, FormulaRef: computed.FormulaRef, InsufficientSample: &computed.InsufficientSample}}, args.IncludeFull, version, debugMetadata, computeActivitySegmentStatsName, response.UnitSystemMetric, shapeCfg)
	}
}

func activitySegmentStatsInput(args computeActivitySegmentStatsRequest) (analysis.SegmentStatsInput, []string, error) {
	bounds, err := activitySegmentBounds(args)
	if err != nil {
		return analysis.SegmentStatsInput{}, nil, err
	}
	stat := strings.TrimSpace(args.Stat)
	if stat == analysis.SegmentStatIF && args.FTPWatts == nil {
		return analysis.SegmentStatsInput{}, nil, fmt.Errorf("%w: ftp_watts is required for if", analysis.ErrInvalidSegmentStatsInput)
	}
	if stat != analysis.SegmentStatIF && args.FTPWatts != nil {
		return analysis.SegmentStatsInput{}, nil, fmt.Errorf("%w: ftp_watts is accepted only for if", analysis.ErrInvalidSegmentStatsInput)
	}
	ftpWatts := 0.0
	if args.FTPWatts != nil {
		ftpWatts = *args.FTPWatts
	}
	input := analysis.SegmentStatsInput{Stat: stat, Metric: strings.TrimSpace(args.Metric), Bounds: bounds, FTPWatts: ftpWatts}
	required, err := analysis.ValidateSegmentStatsInput(input)
	if err != nil {
		return analysis.SegmentStatsInput{}, nil, err
	}
	return input, required, nil
}

func activitySegmentBounds(args computeActivitySegmentStatsRequest) (analysis.SegmentBounds, error) {
	hasTimeStart := args.StartSeconds != nil
	hasTimeEnd := args.EndSeconds != nil
	hasDistanceStart := args.StartDistanceM != nil
	hasDistanceEnd := args.EndDistanceM != nil
	if hasTimeStart != hasTimeEnd || hasDistanceStart != hasDistanceEnd {
		return analysis.SegmentBounds{}, fmt.Errorf("%w: segment ranges require both start and end", analysis.ErrInvalidSegmentStatsInput)
	}
	if hasTimeStart == hasDistanceStart {
		return analysis.SegmentBounds{}, fmt.Errorf("%w: provide exactly one time or distance range", analysis.ErrInvalidSegmentStatsInput)
	}
	if hasTimeStart {
		return analysis.SegmentBounds{Axis: analysis.SegmentAxisTimeSeconds, Start: *args.StartSeconds, End: *args.EndSeconds}, nil
	}
	return analysis.SegmentBounds{Axis: analysis.SegmentAxisDistanceMeter, Start: *args.StartDistanceM, End: *args.EndDistanceM}, nil
}

func missingActivitySegmentStreamMessage(err error) string {
	key := analysis.MissingSegmentStreamKey(err)
	if key == "" {
		return "missing required activity stream for compute_activity_segment_stats"
	}
	return "missing required activity stream for compute_activity_segment_stats: " + key
}

func canonicalActivitySegmentStreams(rows []intervals.ActivityStream, required []string) (map[string][]float64, error) {
	requiredSet := make(map[string]struct{}, len(required))
	for _, key := range required {
		requiredSet[key] = struct{}{}
	}
	out := make(map[string][]float64, len(required))
	for _, row := range rows {
		key, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if _, ok := requiredSet[key]; !ok {
			continue
		}
		out[key] = append([]float64(nil), row.Data...)
	}
	for _, key := range required {
		if len(out[key]) == 0 {
			return nil, analysis.MissingSegmentStreamError{Key: key}
		}
	}
	return out, nil
}

func upstreamSegmentStreamTypes(required []string) []string {
	out := make([]string, 0, len(required))
	seen := map[string]struct{}{}
	for _, key := range required {
		streamType := upstreamSegmentStreamType(key)
		if streamType == "" {
			continue
		}
		if _, ok := seen[streamType]; ok {
			continue
		}
		seen[streamType] = struct{}{}
		out = append(out, streamType)
	}
	return out
}

func upstreamSegmentStreamType(canonical string) string {
	switch canonical {
	case analysis.SegmentAxisTimeSeconds:
		return "time"
	case analysis.SegmentAxisDistanceMeter:
		return "distance"
	case analysis.SegmentMetricWatts:
		return "watts"
	case analysis.SegmentMetricHeartRate:
		return "heartrate"
	case analysis.SegmentMetricCadence:
		return "cadence"
	case analysis.SegmentMetricVelocitySmooth:
		return "velocity_smooth"
	default:
		return ""
	}
}

func computeActivitySegmentStatsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id", "stat"}, "properties": map[string]any{
		"activity_id":      map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose canonical raw streams should be analyzed."},
		"stat":             map[string]any{"type": "string", "enum": analysis.SegmentStatValues(), "description": "Single segment statistic to compute. Use one of mean, median, p90, decoupling, drift, np, or if."},
		"metric":           map[string]any{"type": "string", "enum": analysis.SegmentMetricValues(), "description": "Canonical stream metric for mean/median/p90 only: watts, heart_rate, cadence, velocity_smooth, distance, or time. Use velocity_smooth for pace comparisons and convert m/s to pace in the final answer. Omit for derived stats."},
		"start_seconds":    map[string]any{"type": "number", "minimum": 0, "description": "Inclusive segment start in elapsed activity seconds. Provide with end_seconds and without distance bounds."},
		"end_seconds":      map[string]any{"type": "number", "minimum": 0, "description": "Inclusive segment end in elapsed activity seconds; must be greater than start_seconds."},
		"start_distance_m": map[string]any{"type": "number", "minimum": 0, "description": "Inclusive segment start distance in meters. Provide with end_distance_m and without time bounds."},
		"end_distance_m":   map[string]any{"type": "number", "minimum": 0, "description": "Inclusive segment end distance in meters; must be greater than start_distance_m."},
		"ftp_watts":        map[string]any{"type": "number", "exclusiveMinimum": 0, "description": "Required for stat=if only; positive athlete FTP in watts used as NP / FTP denominator."},
		"include_full":     map[string]any{"type": "boolean", "default": false, "description": "When true, include only sliced/calculation inputs needed to audit this stat. Terse mode never returns raw stream samples."},
	}}
}
