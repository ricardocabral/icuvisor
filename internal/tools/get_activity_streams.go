package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/streams"
)

const (
	getActivityStreamsName        = "get_activity_streams"
	getActivitySplitsName         = "get_activity_splits"
	getActivityStreamsDescription = "Get canonical activity stream channels by activity_id. For a described or date-based activity, resolve it with get_activities first and pass the returned activity_id. Streams are heavy: default returns only available stream metadata; raw samples require include_full:true. Optional time_window (elapsed seconds) or distance_window (meters) selects an inclusive local window; max_points uniformly bounds the selected samples."
	getActivitySplitsDescription  = "Get manual or virtual per-km/per-mile activity splits by activity_id. For split/lap requests on a described or date-based activity, resolve it with get_activities over the athlete-local date window first. Uses manual intervals when present, otherwise derives virtual splits from distance/time streams and honors preferred_units."
)

// ActivityStreamsClient retrieves activity streams.
type ActivityStreamsClient interface {
	GetActivityStreams(context.Context, intervals.ActivityStreamsParams) ([]intervals.ActivityStream, error)
}

type activityStreamWindowBounds struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type activityStreamWindowRequest struct {
	Start float64
	End   float64
}

func (w *activityStreamWindowRequest) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return errors.New("window must be an object")
	}
	for key := range fields {
		if key != "start" && key != "end" {
			return fmt.Errorf("window contains unsupported property %q", key)
		}
	}
	start, ok := fields["start"]
	if !ok || string(start) == "null" {
		return errors.New("window.start is required and must be a number")
	}
	end, ok := fields["end"]
	if !ok || string(end) == "null" {
		return errors.New("window.end is required and must be a number")
	}
	if err := json.Unmarshal(start, &w.Start); err != nil || math.IsNaN(w.Start) || math.IsInf(w.Start, 0) {
		return errors.New("window.start must be a finite number")
	}
	if err := json.Unmarshal(end, &w.End); err != nil || math.IsNaN(w.End) || math.IsInf(w.End, 0) {
		return errors.New("window.end must be a finite number")
	}
	return nil
}

type getActivityStreamsRequest struct {
	ActivityID     string                       `json:"activity_id"`
	Keys           []string                     `json:"keys,omitempty"`
	IncludeFull    bool                         `json:"include_full,omitempty"`
	MaxPoints      *int                         `json:"max_points,omitempty"`
	TimeWindow     *activityStreamWindowRequest `json:"time_window,omitempty"`
	DistanceWindow *activityStreamWindowRequest `json:"distance_window,omitempty"`
}

func (r *getActivityStreamsRequest) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return errors.New("arguments must be a JSON object")
	}
	for key := range fields {
		switch key {
		case "activity_id", "keys", "include_full", "max_points", "time_window", "distance_window":
		default:
			return fmt.Errorf("unsupported property %q", key)
		}
	}
	activityID, ok := fields["activity_id"]
	if !ok {
		return errors.New("activity_id is required")
	}
	if err := json.Unmarshal(activityID, &r.ActivityID); err != nil {
		return errors.New("activity_id must be a string")
	}
	if value, ok := fields["keys"]; ok {
		if string(value) == "null" || json.Unmarshal(value, &r.Keys) != nil {
			return errors.New("keys must be an array of strings")
		}
	}
	if value, ok := fields["include_full"]; ok {
		if string(value) == "null" || json.Unmarshal(value, &r.IncludeFull) != nil {
			return errors.New("include_full must be a boolean")
		}
	}
	if value, ok := fields["max_points"]; ok {
		if string(value) == "null" || json.Unmarshal(value, &r.MaxPoints) != nil || r.MaxPoints == nil {
			return errors.New("max_points must be an integer")
		}
	}
	if value, ok := fields["time_window"]; ok {
		if string(value) == "null" {
			return errors.New("time_window must be an object")
		}
		var window activityStreamWindowRequest
		if err := json.Unmarshal(value, &window); err != nil {
			return fmt.Errorf("invalid time_window: %w", err)
		}
		r.TimeWindow = &window
	}
	if value, ok := fields["distance_window"]; ok {
		if string(value) == "null" {
			return errors.New("distance_window must be an object")
		}
		var window activityStreamWindowRequest
		if err := json.Unmarshal(value, &window); err != nil {
			return fmt.Errorf("invalid distance_window: %w", err)
		}
		r.DistanceWindow = &window
	}
	return nil
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
	Type                string                          `json:"type,omitempty"`
	Name                string                          `json:"name,omitempty"`
	Samples             []float64                       `json:"samples,omitempty"`
	Data2               []float64                       `json:"data2,omitempty"`
	SampleCount         int                             `json:"sample_count,omitempty"`
	SourceSampleCount   *int                            `json:"source_sample_count,omitempty"`
	SelectedSampleCount *int                            `json:"selected_sample_count,omitempty"`
	ReturnedSampleCount *int                            `json:"returned_sample_count,omitempty"`
	SamplingMethod      string                          `json:"sampling_method,omitempty"`
	Window              *activityStreamWindowProvenance `json:"window,omitempty"`
	AllNull             bool                            `json:"all_null,omitempty"`
	Custom              bool                            `json:"custom,omitempty"`
	Full                map[string]any                  `json:"full,omitempty"`
}

type activityStreamWindowProvenance struct {
	Time     *activityStreamWindowDimension `json:"time,omitempty"`
	Distance *activityStreamWindowDimension `json:"distance,omitempty"`
	Empty    bool                           `json:"empty"`
	Status   string                         `json:"status"`
}

type activityStreamWindowDimension struct {
	Requested    activityStreamWindowBounds  `json:"requested"`
	Effective    *activityStreamWindowBounds `json:"effective,omitempty"`
	BoundaryKey  string                      `json:"boundary_key"`
	BoundaryUnit string                      `json:"boundary_unit"`
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
		if err := validateActivityStreamWindow(args.TimeWindow, "time", 86400); err != nil {
			return Result{}, NewUserError(err.Error(), err)
		}
		if err := validateActivityStreamWindow(args.DistanceWindow, "distance", 1000000); err != nil {
			return Result{}, NewUserError(err.Error(), err)
		}
		maxPoints := 0
		if args.MaxPoints != nil {
			maxPoints = *args.MaxPoints
		}
		canonicalKeys, unknown := canonicalStreamKeys(args.Keys)
		upstreamTypes := activityStreamUpstreamTypes(args.Keys, canonicalKeys, args.TimeWindow != nil, args.DistanceWindow != nil)
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
		payload := shapeActivityStreams(args.ActivityID, streamsRows, canonicalKeys, args.IncludeFull, args.IncludeFull, maxPoints, version, unknown, &activityStreamWindowRequestSet{Time: args.TimeWindow, Distance: args.DistanceWindow})
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

type activityStreamWindowRequestSet struct {
	Time     *activityStreamWindowRequest
	Distance *activityStreamWindowRequest
}

type activityStreamWindowSelection struct {
	Indexes        []int
	BoundaryLength int
	Provenance     *activityStreamWindowProvenance
	Diagnostic     *dataAvailabilityDiagnostic
}

func validateActivityStreamWindow(window *activityStreamWindowRequest, name string, maxWidth float64) error {
	if window == nil {
		return nil
	}
	if window.Start < 0 || window.End < 0 {
		return fmt.Errorf("%s_window bounds must be non-negative", name)
	}
	if window.Start > window.End {
		return fmt.Errorf("%s_window start must be less than or equal to end", name)
	}
	if window.End-window.Start > maxWidth {
		return fmt.Errorf("%s_window width must be no more than %v", name, maxWidth)
	}
	return nil
}

func activityStreamUpstreamTypes(keys, canonical []string, timeWindow, distanceWindow bool) []string {
	upstream := append([]string(nil), keys...)
	if len(keys) == 0 {
		return upstream
	}
	requested := make(map[string]bool, len(canonical))
	for _, key := range canonical {
		requested[key] = true
	}
	if timeWindow && !requested["time"] {
		upstream = append(upstream, "time")
	}
	if distanceWindow && !requested["distance"] {
		upstream = append(upstream, "distance")
	}
	return upstream
}

func buildActivityStreamWindowSelection(rows []intervals.ActivityStream, request *activityStreamWindowRequestSet) *activityStreamWindowSelection {
	if request == nil || (request.Time == nil && request.Distance == nil) {
		return nil
	}
	selection := &activityStreamWindowSelection{Provenance: &activityStreamWindowProvenance{Empty: true, Status: "invalid"}}
	var timeValues, distanceValues []float64
	if request.Time != nil {
		selection.Provenance.Time = activityStreamWindowDimensionForRequest(request.Time, "time", "seconds")
		row, ok := findActivityBoundaryStream(rows, "time")
		if !ok {
			selection.Diagnostic = windowDiagnostic("window_boundary_unavailable", "time", "The time boundary stream is unavailable; the requested window was not applied.")
			return selection
		}
		var reason string
		timeValues, reason = validActivityBoundaryValues(row)
		if reason != "" {
			selection.Diagnostic = windowDiagnostic("window_boundary_invalid", "time", "The time boundary stream is null, non-finite, empty, or non-monotonic; the requested window was not applied.")
			return selection
		}
		selection.Provenance.Time.Effective = effectiveActivityWindowBounds(timeValues, request.Time)
	}
	if request.Distance != nil {
		selection.Provenance.Distance = activityStreamWindowDimensionForRequest(request.Distance, "distance", "meters")
		row, ok := findActivityBoundaryStream(rows, "distance")
		if !ok {
			selection.Diagnostic = windowDiagnostic("window_boundary_unavailable", "distance", "The distance boundary stream is unavailable; the requested window was not applied.")
			return selection
		}
		var reason string
		distanceValues, reason = validActivityBoundaryValues(row)
		if reason != "" {
			selection.Diagnostic = windowDiagnostic("window_boundary_invalid", "distance", "The distance boundary stream is null, non-finite, empty, or non-monotonic; the requested window was not applied.")
			return selection
		}
		selection.Provenance.Distance.Effective = effectiveActivityWindowBounds(distanceValues, request.Distance)
	}
	if len(timeValues) > 0 && len(distanceValues) > 0 && len(timeValues) != len(distanceValues) {
		selection.Diagnostic = windowDiagnostic("window_boundary_length_mismatch", "time,distance", "The time and distance boundary streams have different lengths; no intersection window was applied.")
		return selection
	}
	if len(timeValues) > 0 {
		selection.BoundaryLength = len(timeValues)
	} else {
		selection.BoundaryLength = len(distanceValues)
	}
	for index := 0; index < selection.BoundaryLength; index++ {
		selected := true
		if request.Time != nil {
			selected = selected && timeValues[index] >= request.Time.Start && timeValues[index] <= request.Time.End
		}
		if request.Distance != nil {
			selected = selected && distanceValues[index] >= request.Distance.Start && distanceValues[index] <= request.Distance.End
		}
		if selected {
			selection.Indexes = append(selection.Indexes, index)
		}
	}
	selection.Provenance.Empty = len(selection.Indexes) == 0
	if selection.Provenance.Empty {
		selection.Provenance.Status = "empty"
	} else {
		selection.Provenance.Status = "selected"
	}
	return selection
}

func activityStreamWindowDimensionForRequest(request *activityStreamWindowRequest, key, unit string) *activityStreamWindowDimension {
	return &activityStreamWindowDimension{Requested: activityStreamWindowBounds{Start: request.Start, End: request.End}, BoundaryKey: key, BoundaryUnit: unit}
}

func effectiveActivityWindowBounds(values []float64, request *activityStreamWindowRequest) *activityStreamWindowBounds {
	if len(values) == 0 || request.End < values[0] || request.Start > values[len(values)-1] {
		return nil
	}
	start, end := request.Start, request.End
	if start < values[0] {
		start = values[0]
	}
	if end > values[len(values)-1] {
		end = values[len(values)-1]
	}
	return &activityStreamWindowBounds{Start: start, End: end}
}

func findActivityBoundaryStream(rows []intervals.ActivityStream, key string) (intervals.ActivityStream, bool) {
	for _, row := range rows {
		canonical, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if canonical == key {
			return row, true
		}
	}
	return intervals.ActivityStream{}, false
}

func validActivityBoundaryValues(row intervals.ActivityStream) ([]float64, string) {
	if row.AllNull || len(row.Data) == 0 || rawArrayHasNull(row.Raw, "data") {
		return nil, "invalid"
	}
	values := append([]float64(nil), row.Data...)
	for index, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || (index > 0 && value < values[index-1]) {
			return nil, "invalid"
		}
	}
	return values, ""
}

func rawArrayHasNull(raw map[string]any, key string) bool {
	value, ok := raw[key]
	if !ok {
		return false
	}
	if value == nil {
		return true
	}
	values, ok := value.([]any)
	if !ok {
		return false
	}
	for _, value := range values {
		if value == nil {
			return true
		}
	}
	return false
}

func rawArrayHasNullAt(raw map[string]any, key string, indexes []int) bool {
	value, ok := raw[key]
	if !ok {
		return false
	}
	if value == nil {
		return true
	}
	values, ok := value.([]any)
	if !ok {
		return false
	}
	for _, index := range indexes {
		if index >= 0 && index < len(values) && values[index] == nil {
			return true
		}
	}
	return false
}

func streamHasData2(row intervals.ActivityStream) bool {
	if row.Raw != nil {
		_, ok := row.Raw["data2"]
		return ok
	}
	return row.Data2 != nil
}

func windowDiagnostic(reason, requested, message string) *dataAvailabilityDiagnostic {
	return &dataAvailabilityDiagnostic{Reason: reason, Requested: []string{requested}, Message: message}
}

func streamCountPointer(value int) *int {
	return &value
}

func selectActivityStreamValues(values []float64, indexes []int) []float64 {
	selected := make([]float64, 0, len(indexes))
	for _, index := range indexes {
		selected = append(selected, values[index])
	}
	return selected
}

func boundedActivityStreamRaw(raw map[string]any, samples, data2 []float64, data2Present bool) map[string]any {
	out := make(map[string]any, len(raw)+2)
	for key, value := range raw {
		out[key] = value
	}
	out["data"] = samples
	if data2Present {
		out["data2"] = data2
	}
	return out
}

func shapeActivityStreams(activityID string, rows []intervals.ActivityStream, requested []string, samples bool, includeFull bool, maxPoints int, version string, unknown []string, windowRequest *activityStreamWindowRequestSet) getActivityStreamsResponse {
	requestedSet := make(map[string]bool, len(requested))
	for _, key := range requested {
		requestedSet[key] = true
	}
	selection := buildActivityStreamWindowSelection(rows, windowRequest)
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
		if selection == nil {
			var samplesReduced, data2Reduced bool
			if samples {
				row.Samples, samplesReduced = uniformlySampleStreamSeries(streamRow.Data, maxPoints)
				row.Data2, data2Reduced = uniformlySampleStreamSeries(streamRow.Data2, maxPoints)
				if maxPoints != 0 && (samplesReduced || data2Reduced) {
					row.SampleCount = len(streamRow.Data)
					row.ReturnedSampleCount = streamCountPointer(len(row.Samples))
					if len(streamRow.Data2) > len(streamRow.Data) {
						row.SampleCount = len(streamRow.Data2)
						row.ReturnedSampleCount = streamCountPointer(len(row.Data2))
					}
					row.SamplingMethod = "uniform_index"
				}
			}
			if includeFull {
				row.Full = sampledActivityStreamRaw(streamRow.Raw, row.Samples, row.Data2, samplesReduced, data2Reduced)
			}
		} else {
			diagnostic := shapeWindowedActivityStream(&row, streamRow, selection, samples, includeFull, maxPoints)
			if diagnostic != nil {
				out.Meta.DataAvailability = append(out.Meta.DataAvailability, *diagnostic)
			}
		}
		out.Streams[key] = row
	}
	out.Meta.DataAvailability = append(out.Meta.DataAvailability, activityStreamMissingDiagnostics(activityID, requested, out.Streams)...)
	if selection != nil && selection.Diagnostic != nil {
		out.Meta.DataAvailability = append(out.Meta.DataAvailability, *selection.Diagnostic)
	}
	return out
}

func shapeWindowedActivityStream(row *activityStreamRow, stream intervals.ActivityStream, selection *activityStreamWindowSelection, samples, includeFull bool, maxPoints int) *dataAvailabilityDiagnostic {
	row.Window = selection.Provenance
	row.SourceSampleCount = streamCountPointer(len(stream.Data))
	row.SelectedSampleCount = streamCountPointer(0)
	row.ReturnedSampleCount = streamCountPointer(0)
	if selection.Diagnostic != nil {
		return nil
	}
	if stream.AllNull {
		return windowDiagnostic("window_channel_all_null", firstNonEmpty(stream.Type, stream.Name), "The requested stream channel is marked all-null and was withheld from the bounded response.")
	}
	if len(stream.Data) != selection.BoundaryLength {
		return windowDiagnostic("window_channel_length_mismatch", firstNonEmpty(stream.Type, stream.Name), "The requested stream channel length does not match the boundary stream; it was withheld from the bounded response.")
	}
	data2Present := streamHasData2(stream)
	if data2Present && stream.Raw != nil && rawArrayHasNull(stream.Raw, "data2") {
		return windowDiagnostic("window_channel_null", firstNonEmpty(stream.Type, stream.Name), "The requested data2 channel contains null values and was withheld to preserve alignment.")
	}
	if data2Present && len(stream.Data2) != selection.BoundaryLength {
		return windowDiagnostic("window_channel_length_mismatch", firstNonEmpty(stream.Type, stream.Name), "The requested data2 channel length does not match the boundary stream; the channel was withheld to preserve alignment.")
	}
	if rawArrayHasNullAt(stream.Raw, "data", selection.Indexes) || rawArrayHasNullAt(stream.Raw, "data2", selection.Indexes) {
		return windowDiagnostic("window_channel_null", firstNonEmpty(stream.Type, stream.Name), "The requested stream contains null samples and was withheld to avoid converting nulls into zeros.")
	}
	selected := selectActivityStreamValues(stream.Data, selection.Indexes)
	var selectedData2 []float64
	if data2Present {
		selectedData2 = selectActivityStreamValues(stream.Data2, selection.Indexes)
	}
	row.SelectedSampleCount = streamCountPointer(len(selected))
	returned := selected
	returnedData2 := selectedData2
	if maxPoints > 0 {
		returned, _ = uniformlySampleStreamSeries(selected, maxPoints)
		returnedData2, _ = uniformlySampleStreamSeries(selectedData2, maxPoints)
	}
	if len(selected) > maxPoints && maxPoints > 0 {
		row.SamplingMethod = "uniform_index"
	} else {
		row.SamplingMethod = "window"
	}
	if samples {
		row.Samples = returned
		if data2Present {
			row.Data2 = returnedData2
		}
		row.ReturnedSampleCount = streamCountPointer(len(returned))
	}
	if includeFull {
		row.Full = boundedActivityStreamRaw(stream.Raw, returned, returnedData2, data2Present)
	}
	return nil
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

func activityStreamWindowSchema(unit, description string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start", "end"}, "description": description, "properties": map[string]any{
		"start": map[string]any{"type": "number", "minimum": 0, "description": "Inclusive window start in " + unit + "."},
		"end":   map[string]any{"type": "number", "minimum": 0, "description": "Inclusive window end in " + unit + "."},
	}}
}

func activityStreamsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":     map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose stream channels should be listed."},
		"keys":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional stream channels to return. Values are canonicalized to snake_case when known; unknown keys are reported in _meta. Keys filter channels only and never opt in to raw samples."},
		"include_full":    map[string]any{"type": "boolean", "default": false, "description": "When true, include bounded raw upstream stream payloads and samples for available stream channels. Raw samples are otherwise omitted."},
		"max_points":      map[string]any{"type": "integer", "minimum": 2, "maximum": 5000, "description": "Optional per-channel cap for raw sample arrays. Requires include_full:true; uniformly samples selected indices while preserving first and last samples, and reports sampling provenance."},
		"time_window":     activityStreamWindowSchema("elapsed seconds", "Optional inclusive elapsed-time window. Local slicing fetches the complete upstream stream first because no verified upstream window query exists."),
		"distance_window": activityStreamWindowSchema("meters", "Optional inclusive distance window. Local slicing fetches the complete upstream stream first because no verified upstream window query exists."),
	}}
}
func activitySplitsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose manual or virtual splits should be returned."},
		"split_unit":   map[string]any{"type": "string", "enum": []string{"km", "mi"}, "description": "Optional split distance unit. Defaults to the athlete's preferred_units when omitted, falling back to km."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, preserve full response metadata during shaping; split rows remain terse and unit-disambiguated by default."},
	}}
}
