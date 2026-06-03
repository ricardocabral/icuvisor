package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	getActivityDetailsName              = "get_activity_details"
	getActivityIntervalsName            = "get_activity_intervals"
	getActivityDetailsDescription       = "Get one activity's terse metadata and metrics by activity_id, including upstream activity tags when returned, calories_burned as active/exercise calories (distinct from wellness kcal_consumed), carbs_ingested_g for athlete-logged carb intake, carbs_used_g for upstream carbs-burned estimate, and custom_fields for athlete-defined activity custom fields when upstream provides them. This detail payload does not prove lap/rep execution source; before analyzing laps, reps, or interval execution, call get_activity_intervals and check _meta.interval_source/_meta.auto_lap_suspected. If the user described an activity by date/name instead of ID, resolve it with get_activities over the athlete-local date window first. Use include_full only when raw upstream fields are needed; Strava-blocked activities return an unavailable marker instead of sparse N/A rows."
	getActivityIntervalsDescription     = "Get analyzed intervals for one activity by activity_id, including scalar custom interval fields such as lactate under custom_fields when upstream includes them. For reps/laps in a described or date-based activity, resolve the activity with get_activities first and pass the returned activity_id. Check _meta.interval_source (structured_workout, device_laps, or unknown) and _meta.auto_lap_suspected before making lap/rep execution claims; device_laps means the rows look like device/Garmin laps or auto-laps, not a structured workout. Interval units are normalized to the canonical intervals.icu unit enum and raw interval payloads require include_full."
	invalidActivityReadArgumentsMessage = "invalid activity read arguments; provide activity_id and optional include_full"
	fetchActivityDetailsMessage         = "could not fetch activity details; check activity_id and intervals.icu credentials"
)

// ActivityDetailsClient retrieves a single intervals.icu activity.
type ActivityDetailsClient interface {
	GetActivity(context.Context, string) (intervals.Activity, error)
}

// ActivityIntervalsClient retrieves intervals for a single intervals.icu activity.
type ActivityIntervalsClient interface {
	GetActivityIntervals(context.Context, string) (intervals.IntervalsDTO, error)
}

type activityReadRequest struct {
	ActivityID  string `json:"activity_id"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getActivityDetailsResponse struct {
	Activity getActivitiesRow `json:"activity"`
	Meta     activityReadMeta `json:"_meta"`
}

type getActivityIntervalsResponse struct {
	ActivityID string                  `json:"activity_id,omitempty"`
	Analyzed   bool                    `json:"analyzed"`
	Intervals  []activityIntervalRow   `json:"intervals,omitempty"`
	Groups     []activityIntervalGroup `json:"groups,omitempty"`
	Full       map[string]any          `json:"full,omitempty"`
	Meta       activityReadMeta        `json:"_meta"`
}

type getActivityIntervalsUnavailableResponse struct {
	ActivityID     string             `json:"activity_id,omitempty"`
	StravaImported bool               `json:"strava_imported,omitempty"`
	Unavailable    *unavailableReason `json:"unavailable"`
	Full           map[string]any     `json:"full,omitempty"`
	Meta           activityReadMeta   `json:"_meta"`
}

type activityReadMeta struct {
	ServerVersion    string                  `json:"server_version"`
	IncludeFull      bool                    `json:"include_full"`
	Timezone         string                  `json:"timezone,omitempty"`
	Limit            int                     `json:"limit,omitempty"`
	SinceID          int64                   `json:"since_id,omitempty"`
	FieldSemantics   map[string]string       `json:"field_semantics,omitempty"`
	IntervalSource   analysis.IntervalSource `json:"interval_source,omitempty"`
	AutoLapSuspected *bool                   `json:"auto_lap_suspected,omitempty"`
}

type activityIntervalRow struct {
	IntervalID    string         `json:"interval_id,omitempty"`
	Name          string         `json:"name,omitempty"`
	Type          string         `json:"type,omitempty"`
	Unit          units.Unit     `json:"unit,omitempty"`
	UnknownUnit   string         `json:"unknown_unit,omitempty"`
	StartIndex    int            `json:"start_index,omitempty"`
	EndIndex      int            `json:"end_index,omitempty"`
	StartTime     string         `json:"start_time,omitempty"`
	EndTime       string         `json:"end_time,omitempty"`
	StartDistance *float64       `json:"start_distance_m,omitempty"`
	EndDistance   *float64       `json:"end_distance_m,omitempty"`
	Distance      *float64       `json:"distance_m,omitempty"`
	Duration      *float64       `json:"duration_seconds,omitempty"`
	AveragePower  *float64       `json:"average_power_watts,omitempty"`
	AverageHR     *float64       `json:"average_heart_rate_bpm,omitempty"`
	Pace          *float64       `json:"pace,omitempty"`
	CustomFields  map[string]any `json:"custom_fields,omitempty"`
	Full          map[string]any `json:"full,omitempty"`
}

type activityIntervalGroup struct {
	GroupID    string         `json:"group_id,omitempty"`
	Name       string         `json:"name,omitempty"`
	Type       string         `json:"type,omitempty"`
	StartIndex int            `json:"start_index,omitempty"`
	EndIndex   int            `json:"end_index,omitempty"`
	Full       map[string]any `json:"full,omitempty"`
}

func newGetActivityDetailsToolWithGear(client ActivityDetailsClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getActivityDetailsName, Description: getActivityDetailsDescription, InputSchema: activityReadInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityDetailsHandler(client, profileClient, gearClient, gearCache, customFieldClient, customFieldCache, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func newGetActivityIntervalsTool(client ActivityIntervalsClient, detailsClient ActivityDetailsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getActivityIntervalsName, Description: getActivityIntervalsDescription, InputSchema: activityReadInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityIntervalsHandler(client, detailsClient, version, debugMetadata, shapeCfg)})
}

func getActivityDetailsHandler(client ActivityDetailsClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeActivityReadRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return Result{}, ctxErr
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchActivityDetailsMessage, err)
		}
		activity, err := client.GetActivity(ctx, args.ActivityID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchActivityDetailsMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		gearResolutions, err := resolveActivityGear(ctx, gearClient, gearCache, []intervals.Activity{activity})
		if err != nil {
			return Result{}, err
		}
		customFieldCodes := customFieldCache.activityFieldCodes(ctx, customFieldClient)
		activityTimezone := profileTimezone(profile.Timezone, timezoneFallback)
		row := activityRow(activity, args.IncludeFull, activityTimezone, unitSystem, gearResolutions[activity.ID], customFieldCodes)
		payload := getActivityDetailsResponse{Activity: row, Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: args.IncludeFull, Timezone: activityTimezone, FieldSemantics: activityFieldSemantics([]getActivitiesRow{row})}}
		shaped, err := response.Shape(payload, shapeCfg.options(args.IncludeFull, nil, version, debugMetadata, getActivityDetailsName, unitSystem))
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_activity_details response: %w", err)
		}
		if _, err := json.Marshal(shaped); err != nil {
			return Result{}, fmt.Errorf("encoding get_activity_details response: %w", err)
		}
		return TextResult(shaped), nil
	}
}

func getActivityIntervalsHandler(client ActivityIntervalsClient, detailsClient ActivityDetailsClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeActivityReadRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		dto, err := client.GetActivityIntervals(ctx, args.ActivityID)
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			unavailable, unavailableErr := detectActivityUnavailable(ctx, detailsClient, args.ActivityID, err)
			if unavailableErr != nil {
				return Result{}, unavailableErr
			}
			return encodeActivityIntervalsResponse(unavailableActivityIntervalsResponse(unavailable, args.IncludeFull, version), args.IncludeFull, version, debugMetadata, shapeCfg)
		}
		payload := shapeActivityIntervalsDTO(args.ActivityID, dto, args.IncludeFull, version)
		return encodeActivityIntervalsResponse(payload, args.IncludeFull, version, debugMetadata, shapeCfg)
	}
}

func decodeActivityReadRequest(raw json.RawMessage) (activityReadRequest, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return activityReadRequest{}, errors.New("arguments must be a JSON object")
	}
	args, err := DecodeStrict[activityReadRequest](trimmed)
	if err != nil {
		return activityReadRequest{}, err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	if args.ActivityID == "" {
		return activityReadRequest{}, errors.New("activity_id is required")
	}
	return args, nil
}

func shapeActivityIntervalsDTO(activityID string, dto intervals.IntervalsDTO, includeFull bool, version string) any {
	if len(dto.ICUIntervals) == 0 && len(dto.ICUGroups) == 0 {
		activity := activityFromRaw(dto.Raw)
		if isStravaBlocked(activity) {
			return stravaUnavailableIntervalsResponse(firstNonEmpty(dto.ID, activityID), includeFull, version, dto.Raw)
		}
	}
	classification := classifyActivityIntervalsDTO(dto)
	out := getActivityIntervalsResponse{ActivityID: firstNonEmpty(dto.ID, activityID), Analyzed: dto.Analyzed, Intervals: make([]activityIntervalRow, 0, len(dto.ICUIntervals)), Groups: make([]activityIntervalGroup, 0, len(dto.ICUGroups)), Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, IntervalSource: classification.Source, AutoLapSuspected: boolPtr(classification.AutoLapSuspected)}}
	if includeFull {
		out.Full = dto.Raw
	}
	for _, interval := range dto.ICUIntervals {
		out.Intervals = append(out.Intervals, shapeActivityInterval(interval, includeFull))
	}
	for _, group := range dto.ICUGroups {
		out.Groups = append(out.Groups, shapeActivityIntervalGroup(group, includeFull))
	}
	return out
}

func classifyActivityIntervalsDTO(dto intervals.IntervalsDTO) analysis.IntervalSourceResult {
	input := analysis.IntervalSourceInput{Raw: dto.Raw, Intervals: make([]analysis.IntervalSourceInterval, 0, len(dto.ICUIntervals)), Groups: make([]analysis.IntervalSourceGroup, 0, len(dto.ICUGroups))}
	for _, interval := range dto.ICUIntervals {
		input.Intervals = append(input.Intervals, analysis.IntervalSourceInterval{Name: stringValue(interval.Name), Type: stringValue(interval.Type), Label: anyString(interval.Raw["label"]), Raw: interval.Raw, StartIndex: interval.StartIndex, EndIndex: interval.EndIndex, StartDistance: interval.StartDistance, EndDistance: interval.EndDistance, Distance: interval.Distance, Duration: interval.Duration})
	}
	for _, group := range dto.ICUGroups {
		input.Groups = append(input.Groups, analysis.IntervalSourceGroup{Name: stringValue(group.Name), Type: stringValue(group.Type), Raw: group.Raw, StartIndex: group.StartIndex, EndIndex: group.EndIndex})
	}
	return analysis.InferIntervalSource(input)
}

func boolPtr(value bool) *bool {
	return &value
}

func stravaUnavailableIntervalsResponse(activityID string, includeFull bool, version string, raw map[string]any) getActivityIntervalsUnavailableResponse {
	out := getActivityIntervalsUnavailableResponse{ActivityID: activityID, StravaImported: true, Unavailable: &unavailableReason{Reason: "strava_blocked", Workaround: stravaBlockedWorkaround(raw)}, Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull}}
	if includeFull {
		out.Full = raw
	}
	return out
}

func unavailableActivityIntervalsResponse(unavailable activityUnavailable, includeFull bool, version string) getActivityIntervalsUnavailableResponse {
	out := getActivityIntervalsUnavailableResponse{ActivityID: unavailable.ActivityID, StravaImported: unavailable.StravaImported, Unavailable: unavailable.Unavailable, Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull}}
	if includeFull {
		out.Full = unavailable.Full
	}
	return out
}

func encodeActivityIntervalsResponse(payload any, includeFull bool, version string, debugMetadata bool, shaping ...responseShaping) (Result, error) {
	shapeCfg := responseShapingOrDefault(shaping)
	shaped, err := response.Shape(payload, shapeCfg.options(includeFull, []string{"intervals", "groups"}, version, debugMetadata, getActivityIntervalsName, ""))
	if err != nil {
		return Result{}, fmt.Errorf("shaping get_activity_intervals response: %w", err)
	}
	if _, err := json.Marshal(shaped); err != nil {
		return Result{}, fmt.Errorf("encoding get_activity_intervals response: %w", err)
	}
	return TextResult(shaped), nil
}

func isActivityReadFallbackCandidate(err error) bool {
	return isActivityReadLegacyFallbackCandidate(err) || errors.Is(err, intervals.ErrRateLimited) || errors.Is(err, intervals.ErrUpstream)
}

func isActivityReadLegacyFallbackCandidate(err error) bool {
	return errors.Is(err, intervals.ErrNotFound) || errors.Is(err, intervals.ErrUnauthorized)
}

func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func activityFromRaw(raw map[string]any) intervals.Activity {
	if len(raw) == 0 {
		return intervals.Activity{Raw: raw}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return intervals.Activity{Raw: raw}
	}
	var activity intervals.Activity
	if err := json.Unmarshal(data, &activity); err != nil {
		return intervals.Activity{Raw: raw}
	}
	return activity
}

func shapeActivityInterval(interval intervals.ActivityInterval, includeFull bool) activityIntervalRow {
	var unit units.Unit
	var unknown string
	if rawUnit := stringValue(interval.Unit); rawUnit != "" {
		unit, unknown = units.ParseUnit(rawUnit)
	}
	row := activityIntervalRow{IntervalID: anyString(interval.ID), Name: stringValue(interval.Name), Type: stringValue(interval.Type), Unit: unit, UnknownUnit: unknown, StartIndex: intValue(interval.StartIndex), EndIndex: intValue(interval.EndIndex), StartTime: stringValue(interval.StartTime), EndTime: stringValue(interval.EndTime), StartDistance: interval.StartDistance, EndDistance: interval.EndDistance, Distance: interval.Distance, Duration: interval.Duration, AveragePower: interval.AveragePower, AverageHR: interval.AverageHR, Pace: interval.Pace, CustomFields: intervalCustomFields(interval.Raw)}
	if includeFull {
		row.Full = interval.Raw
	}
	return row
}

func intervalCustomFields(raw map[string]any) map[string]any {
	fields := make(map[string]any)
	for key, value := range raw {
		if !isCustomIntervalFieldKey(key) || !isCustomIntervalFieldValue(value) {
			continue
		}
		fields[key] = value
	}
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func isCustomIntervalFieldKey(key string) bool {
	_, known := knownIntervalRawFields[key]
	return !known
}

func isCustomIntervalFieldValue(value any) bool {
	if value == nil {
		return false
	}
	switch value.(type) {
	case string, bool, float64, int, json.Number:
		return true
	default:
		return false
	}
}

var knownIntervalRawFields = map[string]struct{}{
	"id": {}, "name": {}, "label": {}, "type": {}, "unit": {}, "group_id": {},
	"start_index": {}, "end_index": {}, "start_time": {}, "end_time": {},
	"start_distance": {}, "end_distance": {}, "distance": {}, "duration": {},
	"moving_time": {}, "elapsed_time": {}, "elapsed_time_excluding_pauses": {}, "recording_time": {},
	"average_power": {}, "average_watts": {}, "average_watts_kg": {}, "weighted_average_watts": {},
	"min_watts": {}, "max_watts": {}, "normalized_power": {}, "intensity": {},
	"average_hr": {}, "average_heartrate": {}, "max_heartrate": {}, "min_heartrate": {},
	"average_cadence": {}, "max_cadence": {}, "average_speed": {}, "max_speed": {},
	"average_pace": {}, "pace": {}, "gap": {}, "total_elevation_gain": {}, "total_elevation_loss": {},
	"average_stride": {}, "average_dfa_a1": {}, "wbal_start": {}, "wbal_end": {},
	"joules_above_ftp": {}, "decoupling": {}, "avg_lr_balance": {}, "strain_score": {}, "training_load": {},
	"w5s_variability": {}, "power_zone": {}, "hr_zone": {}, "pace_zone": {},
}

func shapeActivityIntervalGroup(group intervals.IntervalGroup, includeFull bool) activityIntervalGroup {
	row := activityIntervalGroup{GroupID: group.ID, Name: stringValue(group.Name), Type: stringValue(group.Type), StartIndex: intValue(group.StartIndex), EndIndex: intValue(group.EndIndex)}
	if includeFull {
		row.Full = group.Raw
	}
	return row
}

func activityReadInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{"activity_id": map[string]any{"type": "string", "description": "intervals.icu activity ID."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "Include raw upstream fields; default terse mode strips nulls and returns normalized fields."}}}
}

func activityReadOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Activity detail or interval response. Activity detail rows include upstream tags when intervals.icu returns a string-array tags field, calories_burned (active/exercise calories, distinct from wellness kcal_consumed intake), carbs_ingested_g (athlete-logged carb intake in grams), carbs_used_g (upstream carbs-burned estimate in grams), gear_id/gear_name when upstream permits, and gear_resolution values resolved/name_missing/unresolved/lookup_unavailable so unresolved IDs are never guessed. custom_fields holds athlete-defined activity custom field values keyed by the upstream field code when intervals.icu returns them. On activity detail responses the activity's timezone field and _meta.timezone give the IANA zone start_date_local is in; start_date_utc is UTC. Derive the calendar date from that timezone so the activity is not reported on the wrong day. For lap/rep or interval-execution analysis, use get_activity_intervals and inspect _meta.interval_source plus _meta.auto_lap_suspected; get_activity_details alone does not identify whether intervals are structured_workout rows or device_laps/auto-laps."}
}
