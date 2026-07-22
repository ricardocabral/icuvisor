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
	getActivityStreamsDescription = "Get canonical activity stream channels by activity_id. For a described or date-based activity, resolve it with get_activities first and pass the returned activity_id. Streams are heavy: default returns only available stream metadata; raw samples require include_full:true and can be uniformly bounded per channel with max_points."
	getActivitySplitsDescription  = "Get manual or virtual per-km/per-mile activity splits by activity_id. For split/lap requests on a described or date-based activity, resolve it with get_activities over the athlete-local date window first. Uses manual intervals when present, otherwise derives virtual splits from distance/time streams and honors preferred_units."
)

// ActivityStreamsClient retrieves activity streams.
type ActivityStreamsClient interface {
	GetActivityStreams(context.Context, intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error)
}

type getActivityStreamsRequest struct {
	ActivityID  string   `json:"activity_id"`
	Keys        []string `json:"keys,omitempty"`
	IncludeFull bool     `json:"include_full,omitempty"`
	MaxPoints   *int     `json:"max_points,omitempty"`
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

type getActivityStreamsUnavailableResponse struct {
	ActivityID     string              `json:"activity_id"`
	StravaImported bool                `json:"strava_imported,omitempty"`
	Unavailable    *unavailableReason  `json:"unavailable"`
	Full           map[string]any      `json:"full,omitempty"`
	Meta           activityStreamsMeta `json:"_meta"`
}

type activityStreamRow struct {
	Type                string         `json:"type,omitempty"`
	Name                string         `json:"name,omitempty"`
	Samples             []float64      `json:"samples,omitempty"`
	Data2               []float64      `json:"data2,omitempty"`
	SampleCount         int            `json:"sample_count,omitempty"`
	ReturnedSampleCount int            `json:"returned_sample_count,omitempty"`
	SamplingMethod      string         `json:"sampling_method,omitempty"`
	AllNull             bool           `json:"all_null,omitempty"`
	Custom              bool           `json:"custom,omitempty"`
	Full                map[string]any `json:"full,omitempty"`
}

type activityStreamsMeta struct {
	ServerVersion     string                       `json:"server_version"`
	IncludeFull       bool                         `json:"include_full"`
	SamplesIncluded   bool                         `json:"samples_included"`
	UnknownStreamKeys []string                     `json:"unknown_stream_keys,omitempty"`
	DataAvailability  []dataAvailabilityDiagnostic `json:"data_availability,omitempty"`
}

type getActivitySplitsResponse struct {
	ActivityID string             `json:"activity_id"`
	SplitUnit  string             `json:"split_unit"`
	Source     string             `json:"source"`
	Splits     []activitySplitRow `json:"splits"`
	Meta       activitySplitsMeta `json:"_meta"`
}

type getActivitySplitsUnavailableResponse struct {
	ActivityID     string             `json:"activity_id"`
	StravaImported bool               `json:"strava_imported,omitempty"`
	Unavailable    *unavailableReason `json:"unavailable"`
	Full           map[string]any     `json:"full,omitempty"`
	Meta           activitySplitsMeta `json:"_meta"`
}

type activitySplitRow struct {
	Index           int      `json:"index"`
	DistanceKM      *float64 `json:"distance_km,omitempty"`
	DistanceMI      *float64 `json:"distance_mi,omitempty"`
	DurationSeconds float64  `json:"duration_seconds"`
	PaceSeconds     float64  `json:"pace_seconds"`
}

type activitySplitsMeta struct {
	ServerVersion    string                       `json:"server_version"`
	IncludeFull      bool                         `json:"include_full"`
	Algorithm        string                       `json:"algorithm"`
	Units            map[string]string            `json:"units,omitempty"`
	DataAvailability []dataAvailabilityDiagnostic `json:"data_availability,omitempty"`
}

func newGetActivityStreamsTool(client ActivityStreamsClient, detailsClient ActivityDetailsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getActivityStreamsName, Description: getActivityStreamsDescription, InputSchema: activityStreamsInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityStreamsHandler(client, detailsClient, version, debugMetadata, shapeCfg)})
}

func newGetActivitySplitsTool(streamsClient ActivityStreamsClient, intervalsClient ActivityIntervalsClient, detailsClient ActivityDetailsClient, profileClient ProfileClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getActivitySplitsName, Description: getActivitySplitsDescription, InputSchema: activitySplitsInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivitySplitsHandler(streamsClient, intervalsClient, detailsClient, profileClient, version, debugMetadata, shapeCfg)})
}

func getActivityStreamsHandler(client ActivityStreamsClient, detailsClient ActivityDetailsClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args getActivityStreamsRequest
		if err := decodeJSONArgs(req.Arguments, &args); err != nil || strings.TrimSpace(args.ActivityID) == "" {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		if args.MaxPoints != nil && !args.IncludeFull {
			return Result{}, NewUserError("max_points requires include_full:true", errors.New("max_points was provided without include_full"))
		}
		if args.MaxPoints != nil && (*args.MaxPoints < 2 || *args.MaxPoints > 5000) {
			return Result{}, NewUserError("max_points must be between 2 and 5000", errors.New("max_points is outside the supported range"))
		}
		maxPoints := 0
		if args.MaxPoints != nil {
			maxPoints = *args.MaxPoints
		}
		canonicalKeys, unknown := canonicalStreamKeys(args.Keys)
		upstreamTypes := append([]string(nil), args.Keys...)
		streamsRows, err := client.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: args.ActivityID, Types: upstreamTypes, IncludeDefaults: true})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			unavailable, unavailableErr := detectActivityUnavailable(ctx, detailsClient, args.ActivityID, err)
			if unavailableErr != nil {
				return Result{}, unavailableErr
			}
			payload := unavailableActivityStreamsResponse(unavailable, args.IncludeFull, version)
			return encodeActivityStreamsPayload(payload, args.IncludeFull, version, debugMetadata, shapeCfg)
		}
		payload := shapeActivityStreams(args.ActivityID, streamsRows, canonicalKeys, args.IncludeFull, args.IncludeFull, maxPoints, version, unknown)
		return encodeActivityStreamsPayload(payload, args.IncludeFull, version, debugMetadata, shapeCfg)
	}
}

func getActivitySplitsHandler(streamsClient ActivityStreamsClient, intervalsClient ActivityIntervalsClient, detailsClient ActivityDetailsClient, profileClient ProfileClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
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
				if isContextError(err) {
					return Result{}, err
				}
				unavailable, unavailableErr := detectActivityUnavailable(ctx, detailsClient, args.ActivityID, err)
				if unavailableErr != nil {
					return Result{}, unavailableErr
				}
				payload := unavailableActivitySplitsResponse(unavailable, args.IncludeFull, version, unitSystem)
				return encodeActivitySplitsPayload(payload, args.IncludeFull, version, debugMetadata, shapeCfg, unitSystem)
			}
			splits = virtualSplits(streamRows, splitUnit)
			source = "virtual_streams"
		}
		payload := getActivitySplitsResponse{ActivityID: args.ActivityID, SplitUnit: splitUnit, Source: source, Splits: splits, Meta: activitySplitsMeta{ServerVersion: normalizeVersion(version), IncludeFull: args.IncludeFull, Algorithm: "manual intervals when available; otherwise interpolate distance/time stream samples, ignoring paused-segment semantics when moving samples are absent", Units: unitSystem.Metadata()}}
		return encodeActivitySplitsPayload(payload, args.IncludeFull, version, debugMetadata, shapeCfg, unitSystem)
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

func uniformlySampleStreamSeries(values []float64, maxPoints int) ([]float64, bool) {
	if maxPoints == 0 || maxPoints >= len(values) {
		return values, false
	}
	sampled := make([]float64, maxPoints)
	lastIndex := len(values) - 1
	for i := range sampled {
		index := int(math.Round(float64(i) * float64(lastIndex) / float64(maxPoints-1)))
		sampled[i] = values[index]
	}
	return sampled, true
}

func sampledActivityStreamRaw(raw map[string]any, samples []float64, data2 []float64, samplesReduced bool, data2Reduced bool) map[string]any {
	if raw == nil || (!samplesReduced && !data2Reduced) {
		return raw
	}
	rawCopy := make(map[string]any, len(raw))
	for key, value := range raw {
		rawCopy[key] = value
	}
	if samplesReduced {
		rawCopy["data"] = samples
	}
	if data2Reduced {
		rawCopy["data2"] = data2
	}
	return rawCopy
}

func shapeActivityStreams(activityID string, rows []intervals.ActivityStream, requested []string, samples bool, includeFull bool, maxPoints int, version string, unknown []string) getActivityStreamsResponse {
	requestedSet := make(map[string]bool, len(requested))
	for _, key := range requested {
		requestedSet[key] = true
	}
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
		var samplesReduced, data2Reduced bool
		if samples {
			row.Samples, samplesReduced = uniformlySampleStreamSeries(streamRow.Data, maxPoints)
			row.Data2, data2Reduced = uniformlySampleStreamSeries(streamRow.Data2, maxPoints)
			if maxPoints != 0 && (samplesReduced || data2Reduced) {
				row.SampleCount = len(streamRow.Data)
				row.ReturnedSampleCount = len(row.Samples)
				if len(streamRow.Data2) > len(streamRow.Data) {
					row.SampleCount = len(streamRow.Data2)
					row.ReturnedSampleCount = len(row.Data2)
				}
				row.SamplingMethod = "uniform_index"
			}
		}
		if includeFull {
			row.Full = sampledActivityStreamRaw(streamRow.Raw, row.Samples, row.Data2, samplesReduced, data2Reduced)
		}
		out.Streams[key] = row
	}
	out.Meta.DataAvailability = activityStreamMissingDiagnostics(activityID, requested, out.Streams)
	return out
}

func unavailableActivityStreamsResponse(unavailable activityUnavailable, includeFull bool, version string) getActivityStreamsUnavailableResponse {
	meta := activityStreamsMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull}
	if diagnostic := restrictedSourceDiagnostic(unavailable.ActivityID, unavailable.Unavailable); diagnostic != nil {
		meta.DataAvailability = []dataAvailabilityDiagnostic{*diagnostic}
	}
	out := getActivityStreamsUnavailableResponse{ActivityID: unavailable.ActivityID, StravaImported: unavailable.StravaImported, Unavailable: unavailable.Unavailable, Meta: meta}
	if includeFull {
		out.Full = unavailable.Full
	}
	return out
}

func encodeActivityStreamsPayload(payload any, includeFull bool, version string, debugMetadata bool, shapeCfg responseShaping) (Result, error) {
	shaped, err := response.Shape(payload, shapeCfg.options(includeFull, nil, version, debugMetadata, getActivityStreamsName, ""))
	if err != nil {
		return Result{}, err
	}
	return TextResult(shaped), nil
}

func unavailableActivitySplitsResponse(unavailable activityUnavailable, includeFull bool, version string, unitSystem response.UnitSystem) getActivitySplitsUnavailableResponse {
	meta := activitySplitsMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, Algorithm: "manual intervals when available; otherwise interpolate distance/time stream samples, ignoring paused-segment semantics when moving samples are absent", Units: unitSystem.Metadata()}
	if diagnostic := restrictedSourceDiagnostic(unavailable.ActivityID, unavailable.Unavailable); diagnostic != nil {
		meta.DataAvailability = []dataAvailabilityDiagnostic{*diagnostic}
	}
	out := getActivitySplitsUnavailableResponse{ActivityID: unavailable.ActivityID, StravaImported: unavailable.StravaImported, Unavailable: unavailable.Unavailable, Meta: meta}
	if includeFull {
		out.Full = unavailable.Full
	}
	return out
}

func encodeActivitySplitsPayload(payload any, includeFull bool, version string, debugMetadata bool, shapeCfg responseShaping, unitSystem response.UnitSystem) (Result, error) {
	shaped, err := response.Shape(payload, shapeCfg.options(includeFull, []string{"splits"}, version, debugMetadata, getActivitySplitsName, unitSystem))
	if err != nil {
		return Result{}, err
	}
	return TextResult(shaped), nil
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
		"keys":         map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional stream channels to return. Values are canonicalized to snake_case when known; unknown keys are reported in _meta. Keys filter channels only and never opt in to raw samples."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream stream payloads and samples for available stream channels. Raw samples are otherwise omitted."},
		"max_points":   map[string]any{"type": "integer", "minimum": 2, "maximum": 5000, "description": "Optional per-channel cap for raw sample arrays. Requires include_full:true; uniformly samples indices while preserving first and last samples, and reports sampling provenance."},
	}}
}
func activitySplitsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose manual or virtual splits should be returned."},
		"split_unit":   map[string]any{"type": "string", "enum": []string{"km", "mi"}, "description": "Optional split distance unit. Defaults to the athlete's preferred_units when omitted, falling back to km."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, preserve full response metadata during shaping; split rows remain terse and unit-disambiguated by default."},
	}}
}
