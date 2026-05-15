package tools

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/streams"
)

const (
	getActivityStreamsName        = "get_activity_streams"
	getActivitySplitsName         = "get_activity_splits"
	getActivityStreamsDescription = "Get canonical activity stream channels by activity_id. Streams are heavy: default returns only available stream metadata; include samples only with include_full:true or explicit keys."
	getActivitySplitsDescription  = "Get manual or virtual per-km/per-mile activity splits. Uses manual intervals when present, otherwise derives virtual splits from distance/time streams and honors preferred_units."
)

// ActivityStreamsClient retrieves activity streams.
type ActivityStreamsClient interface {
	GetActivityStreams(context.Context, intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error)
}

type getActivityStreamsRequest struct {
	ActivityID  string   `json:"activity_id"`
	Keys        []string `json:"keys,omitempty"`
	IncludeFull bool     `json:"include_full,omitempty"`
}

type getActivitySplitsRequest struct {
	ActivityID  string `json:"activity_id"`
	SplitUnit   string `json:"split_unit,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getActivityStreamsResponse struct {
	ActivityID string                       `json:"activity_id"`
	Streams    map[string]activityStreamRow `json:"streams"`
	Meta       activityStreamsMeta          `json:"_meta"`
}

type activityStreamRow struct {
	Type    string         `json:"type,omitempty"`
	Name    string         `json:"name,omitempty"`
	Samples []float64      `json:"samples,omitempty"`
	Data2   []float64      `json:"data2,omitempty"`
	AllNull bool           `json:"all_null,omitempty"`
	Custom  bool           `json:"custom,omitempty"`
	Full    map[string]any `json:"full,omitempty"`
}

type activityStreamsMeta struct {
	ServerVersion     string   `json:"server_version"`
	IncludeFull       bool     `json:"include_full"`
	SamplesIncluded   bool     `json:"samples_included"`
	UnknownStreamKeys []string `json:"unknown_stream_keys,omitempty"`
}

type getActivitySplitsResponse struct {
	ActivityID string             `json:"activity_id"`
	SplitUnit  string             `json:"split_unit"`
	Source     string             `json:"source"`
	Splits     []activitySplitRow `json:"splits"`
	Meta       activitySplitsMeta `json:"_meta"`
}

type activitySplitRow struct {
	Index           int      `json:"index"`
	DistanceKM      *float64 `json:"distance_km,omitempty"`
	DistanceMI      *float64 `json:"distance_mi,omitempty"`
	DurationSeconds float64  `json:"duration_seconds"`
	PaceSeconds     float64  `json:"pace_seconds"`
}

type activitySplitsMeta struct {
	ServerVersion string            `json:"server_version"`
	IncludeFull   bool              `json:"include_full"`
	Algorithm     string            `json:"algorithm"`
	Units         map[string]string `json:"units,omitempty"`
}

func newGetActivityStreamsTool(client ActivityStreamsClient, version string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: getActivityStreamsName, Description: getActivityStreamsDescription, InputSchema: activityStreamsInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityStreamsHandler(client, version, debugMetadata)})
}

func newGetActivitySplitsTool(streamsClient ActivityStreamsClient, intervalsClient ActivityIntervalsClient, profileClient ProfileClient, version string, debugMetadata bool) Tool {
	return coreTool(Tool{Name: getActivitySplitsName, Description: getActivitySplitsDescription, InputSchema: activitySplitsInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivitySplitsHandler(streamsClient, intervalsClient, profileClient, version, debugMetadata)})
}

func getActivityStreamsHandler(client ActivityStreamsClient, version string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args getActivityStreamsRequest
		if err := decodeJSONArgs(req.Arguments, &args); err != nil || strings.TrimSpace(args.ActivityID) == "" {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		canonicalKeys, unknown := canonicalStreamKeys(args.Keys)
		upstreamTypes := append([]string(nil), args.Keys...)
		streamsRows, err := client.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: args.ActivityID, Types: upstreamTypes, IncludeDefaults: true})
		if err != nil {
			return Result{}, NewUserError(fetchActivityDetailsMessage, err)
		}
		samples := args.IncludeFull || len(args.Keys) > 0
		payload := shapeActivityStreams(args.ActivityID, streamsRows, canonicalKeys, samples, args.IncludeFull, version, unknown)
		shaped, err := response.Shape(payload, response.Options{IncludeFull: args.IncludeFull, ServerVersion: version, DebugMetadata: debugMetadata, QueryType: getActivityStreamsName})
		if err != nil {
			return Result{}, err
		}
		return TextResult(shaped), nil
	}
}

func getActivitySplitsHandler(streamsClient ActivityStreamsClient, intervalsClient ActivityIntervalsClient, profileClient ProfileClient, version string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args getActivitySplitsRequest
		if err := decodeJSONArgs(req.Arguments, &args); err != nil || strings.TrimSpace(args.ActivityID) == "" {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			return Result{}, NewUserError(fetchAthleteProfileMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		splitUnit := normalizeSplitUnit(args.SplitUnit, unitSystem)
		dto, _ := intervalsClient.GetActivityIntervals(ctx, args.ActivityID)
		splits, source := splitsFromIntervals(dto, splitUnit)
		if len(splits) == 0 {
			streamRows, err := streamsClient.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: args.ActivityID, Types: []string{"distance", "time"}, IncludeDefaults: true})
			if err != nil {
				return Result{}, NewUserError(fetchActivityDetailsMessage, err)
			}
			splits = virtualSplits(streamRows, splitUnit)
			source = "virtual_streams"
		}
		payload := getActivitySplitsResponse{ActivityID: args.ActivityID, SplitUnit: splitUnit, Source: source, Splits: splits, Meta: activitySplitsMeta{ServerVersion: normalizeVersion(version), IncludeFull: args.IncludeFull, Algorithm: "manual intervals when available; otherwise interpolate distance/time stream samples, ignoring paused-segment semantics when moving samples are absent", Units: unitSystem.Metadata()}}
		shaped, err := response.Shape(payload, response.Options{IncludeFull: args.IncludeFull, RowCollections: []string{"splits"}, ServerVersion: version, DebugMetadata: debugMetadata, QueryType: getActivitySplitsName, UnitSystem: unitSystem})
		if err != nil {
			return Result{}, err
		}
		return TextResult(shaped), nil
	}
}

func decodeJSONArgs(raw json.RawMessage, out any) error {
	if len(raw) == 0 {
		return errors.New("arguments must be a JSON object")
	}
	return json.Unmarshal(raw, out)
}

func canonicalStreamKeys(keys []string) ([]string, []string) {
	canonical := make([]string, 0, len(keys))
	unknown := []string{}
	for _, key := range keys {
		c, known := streams.CanonicalKey(key)
		if c != "" {
			canonical = append(canonical, c)
		}
		if !known {
			unknown = append(unknown, key)
		}
	}
	return canonical, unknown
}

func shapeActivityStreams(activityID string, rows []intervals.ActivityStream, requested []string, samples bool, includeFull bool, version string, unknown []string) getActivityStreamsResponse {
	requestedSet := stringSet(requested)
	out := getActivityStreamsResponse{ActivityID: activityID, Streams: map[string]activityStreamRow{}, Meta: activityStreamsMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, SamplesIncluded: samples, UnknownStreamKeys: unknown}}
	for _, streamRow := range rows {
		key, known := streams.CanonicalKey(firstNonEmpty(streamRow.Type, streamRow.Name))
		if !known {
			out.Meta.UnknownStreamKeys = append(out.Meta.UnknownStreamKeys, firstNonEmpty(streamRow.Type, streamRow.Name))
		}
		if len(requestedSet) > 0 && !requestedSet[key] {
			continue
		}
		row := activityStreamRow{Type: streamRow.Type, Name: streamRow.Name, AllNull: streamRow.AllNull, Custom: streamRow.Custom}
		if samples {
			row.Samples = streamRow.Data
			row.Data2 = streamRow.Data2
		}
		if includeFull {
			row.Full = streamRow.Raw
		}
		out.Streams[key] = row
	}
	return out
}

func normalizeSplitUnit(requested string, unitSystem response.UnitSystem) string {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "mi" || requested == "mile" || requested == "miles" {
		return "mi"
	}
	if requested == "km" {
		return "km"
	}
	if unitSystem == response.UnitSystemImperial {
		return "mi"
	}
	return "km"
}

func splitDistanceMeters(splitUnit string) float64 {
	if splitUnit == "mi" {
		return 1609.344
	}
	return 1000
}

func splitsFromIntervals(dto intervals.IntervalsDTO, splitUnit string) ([]activitySplitRow, string) {
	rows := []activitySplitRow{}
	for _, interval := range dto.ICUIntervals {
		if interval.Distance != nil && interval.Duration != nil && *interval.Distance > 0 && *interval.Duration > 0 {
			rows = append(rows, newSplitRow(len(rows)+1, *interval.Distance, *interval.Duration, splitUnit))
		}
	}
	if len(rows) > 0 {
		return rows, "manual_intervals"
	}
	return nil, ""
}

func virtualSplits(rows []intervals.ActivityStream, splitUnit string) []activitySplitRow {
	var distance, times []float64
	for _, row := range rows {
		key, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if key == "distance" {
			distance = row.Data
		}
		if key == "time" {
			times = row.Data
		}
	}
	if len(distance) == 0 || len(times) == 0 || len(distance) != len(times) {
		return nil
	}
	step := splitDistanceMeters(splitUnit)
	out := []activitySplitRow{}
	previousTime := 0.0
	for target := step; target <= distance[len(distance)-1]+0.001; target += step {
		t := interpolateTime(distance, times, target)
		duration := t - previousTime
		if duration > 0 {
			out = append(out, newSplitRow(len(out)+1, step, duration, splitUnit))
			previousTime = t
		}
	}
	return out
}

func interpolateTime(distance []float64, times []float64, target float64) float64 {
	for i := 1; i < len(distance); i++ {
		if distance[i] >= target {
			span := distance[i] - distance[i-1]
			if span <= 0 {
				return times[i]
			}
			ratio := (target - distance[i-1]) / span
			return times[i-1] + ratio*(times[i]-times[i-1])
		}
	}
	return times[len(times)-1]
}

func newSplitRow(index int, meters float64, duration float64, splitUnit string) activitySplitRow {
	pace := duration
	row := activitySplitRow{Index: index, DurationSeconds: math.Round(duration*10) / 10, PaceSeconds: math.Round(pace*10) / 10}
	if splitUnit == "mi" {
		value := math.Round((meters/1609.344)*1000) / 1000
		row.DistanceMI = &value
	} else {
		value := math.Round((meters/1000)*1000) / 1000
		row.DistanceKM = &value
	}
	return row
}

func activityStreamsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose stream channels should be listed."},
		"keys":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional stream keys to include as samples. Values are canonicalized to snake_case when known; unknown keys are reported in _meta. Supplying keys includes samples even when include_full is false."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream stream payloads and samples for available stream channels; default returns metadata only unless keys is supplied."},
	}}
}
func activitySplitsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose manual or virtual splits should be returned."},
		"split_unit":   map[string]any{"type": "string", "enum": []string{"km", "mi"}, "description": "Optional split distance unit. Defaults to the athlete's preferred_units when omitted, falling back to km."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, preserve full response metadata during shaping; split rows remain terse and unit-disambiguated by default."},
	}}
}
