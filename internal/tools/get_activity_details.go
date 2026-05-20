package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	getActivityDetailsName              = "get_activity_details"
	getActivityIntervalsName            = "get_activity_intervals"
	getActivityDetailsDescription       = "Get one activity's terse metadata and metrics by activity_id. Use include_full only when raw upstream fields are needed; Strava-blocked activities return an unavailable marker instead of sparse N/A rows."
	getActivityIntervalsDescription     = "Get analyzed intervals for one activity by activity_id. Interval units are normalized to the canonical intervals.icu unit enum and raw interval payloads require include_full."
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
	ServerVersion string `json:"server_version"`
	IncludeFull   bool   `json:"include_full"`
	Limit         int    `json:"limit,omitempty"`
	SinceID       int64  `json:"since_id,omitempty"`
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

func newGetActivityDetailsTool(client ActivityDetailsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getActivityDetailsName, Description: getActivityDetailsDescription, InputSchema: activityReadInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityDetailsHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func newGetActivityIntervalsTool(client ActivityIntervalsClient, detailsClient ActivityDetailsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getActivityIntervalsName, Description: getActivityIntervalsDescription, InputSchema: activityReadInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityIntervalsHandler(client, detailsClient, version, debugMetadata, shapeCfg)})
}

func getActivityDetailsHandler(client ActivityDetailsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
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
		payload := getActivityDetailsResponse{Activity: activityRow(activity, args.IncludeFull, profileTimezone(profile.Timezone, timezoneFallback), unitSystem), Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: args.IncludeFull}}
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
	out := getActivityIntervalsResponse{ActivityID: firstNonEmpty(dto.ID, activityID), Analyzed: dto.Analyzed, Intervals: make([]activityIntervalRow, 0, len(dto.ICUIntervals)), Groups: make([]activityIntervalGroup, 0, len(dto.ICUGroups)), Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull}}
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

func stravaUnavailableIntervalsResponse(activityID string, includeFull bool, version string, raw map[string]any) getActivityIntervalsUnavailableResponse {
	out := getActivityIntervalsUnavailableResponse{ActivityID: activityID, StravaImported: true, Unavailable: &unavailableReason{Reason: "strava_blocked", Workaround: stravaWorkaround}, Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull}}
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
	row := activityIntervalRow{IntervalID: anyString(interval.ID), Name: stringValue(interval.Name), Type: stringValue(interval.Type), Unit: unit, UnknownUnit: unknown, StartIndex: intValue(interval.StartIndex), EndIndex: intValue(interval.EndIndex), StartTime: stringValue(interval.StartTime), EndTime: stringValue(interval.EndTime), StartDistance: interval.StartDistance, EndDistance: interval.EndDistance, Distance: interval.Distance, Duration: interval.Duration, AveragePower: interval.AveragePower, AverageHR: interval.AverageHR, Pace: interval.Pace}
	if includeFull {
		row.Full = interval.Raw
	}
	return row
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
	return map[string]any{"type": "object", "additionalProperties": true}
}
