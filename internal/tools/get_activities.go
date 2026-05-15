package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	getActivitiesName                    = "get_activities"
	getActivitiesDescription             = "List activities for a date range with terse unit-disambiguated rows, Strava-unavailable detection, and opaque pagination. Use this before details, intervals, streams, splits, or messages when a prompt asks about recent training."
	invalidGetActivitiesArgumentsMessage = "invalid get_activities arguments; provide oldest/newest dates or a valid next_page_token"
	fetchActivitiesMessage               = "could not fetch activities; check intervals.icu credentials, athlete ID, and date range"
	activitiesPaginationBoundaryMessage  = "activity pagination hit too many same-timestamp filtered rows; narrow the date range or set include_unnamed true"
	defaultActivitiesPageSize            = 50
	maxActivitiesPageSize                = 200
	maxActivityPageFetches               = 5
	maxActivityFetchLimit                = 201
	stravaWorkaround                     = "connect device directly to intervals.icu (Garmin, Wahoo, Coros, Suunto, Polar)"
)

var terseActivityFields = []string{
	"id", "name", "type", "sub_type", "start_date_local", "start_date", "timezone",
	"source", "_note", "icu_athlete_id", "external_id", "stream_types",
	"distance", "icu_distance", "moving_time", "elapsed_time", "average_speed", "max_speed",
	"total_elevation_gain", "total_elevation_loss", "icu_training_load", "average_heartrate",
	"max_heartrate", "average_cadence", "calories", "device_name",
}

// ActivitiesClient lists intervals.icu activities for tools.
type ActivitiesClient interface {
	ListActivities(context.Context, intervals.ListActivitiesParams) ([]intervals.Activity, error)
}

// GetActivitiesRequest contains get_activities arguments.
type GetActivitiesRequest struct {
	Oldest         string `json:"oldest,omitempty"`
	Newest         string `json:"newest,omitempty"`
	RouteID        int64  `json:"route_id,omitempty"`
	IncludeUnnamed bool   `json:"include_unnamed,omitempty"`
	PageSize       int    `json:"page_size,omitempty"`
	NextPageToken  string `json:"next_page_token,omitempty"`
	IncludeFull    bool   `json:"include_full,omitempty"`
}

type getActivitiesResponse struct {
	Activities []getActivitiesRow `json:"activities"`
	Meta       getActivitiesMeta  `json:"_meta"`
}

type getActivitiesRow struct {
	ActivityID          string             `json:"activity_id,omitempty"`
	Name                string             `json:"name,omitempty"`
	Sport               string             `json:"sport,omitempty"`
	SubType             string             `json:"sub_type,omitempty"`
	StartDateLocal      string             `json:"start_date_local,omitempty"`
	StartDateUTC        string             `json:"start_date_utc,omitempty"`
	Timezone            string             `json:"timezone,omitempty"`
	MovingTimeSeconds   int                `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds  int                `json:"elapsed_time_seconds,omitempty"`
	DistanceKM          *float64           `json:"distance_km,omitempty"`
	DistanceMI          *float64           `json:"distance_mi,omitempty"`
	PaceSecondsPerKM    *float64           `json:"pace_seconds_per_km,omitempty"`
	PaceSecondsPerMile  *float64           `json:"pace_seconds_per_mile,omitempty"`
	AverageSpeedKMH     *float64           `json:"average_speed_kmh,omitempty"`
	AverageSpeedMPH     *float64           `json:"average_speed_mph,omitempty"`
	MaxSpeedKMH         *float64           `json:"max_speed_kmh,omitempty"`
	MaxSpeedMPH         *float64           `json:"max_speed_mph,omitempty"`
	ElevationGainM      *float64           `json:"elevation_gain_m,omitempty"`
	ElevationLossM      *float64           `json:"elevation_loss_m,omitempty"`
	TrainingLoad        int                `json:"training_load,omitempty"`
	AverageHeartRateBPM int                `json:"average_heart_rate_bpm,omitempty"`
	MaxHeartRateBPM     int                `json:"max_heart_rate_bpm,omitempty"`
	AverageCadenceRPM   *float64           `json:"average_cadence_rpm,omitempty"`
	CaloriesBurned      int                `json:"calories_burned,omitempty"`
	DeviceName          string             `json:"device_name,omitempty"`
	HasStreams          bool               `json:"has_streams,omitempty"`
	StravaImported      bool               `json:"strava_imported,omitempty"`
	Unavailable         *unavailableReason `json:"unavailable,omitempty"`
	Full                map[string]any     `json:"full,omitempty"`
}

type unavailableReason struct {
	Reason     string `json:"reason"`
	Workaround string `json:"workaround"`
}

type getActivitiesMeta struct {
	PageSize      int    `json:"page_size"`
	NextPageToken string `json:"next_page_token,omitempty"`
	MoreAvailable bool   `json:"more_available"`
	IncludeFull   bool   `json:"include_full"`
}

var errActivitiesPaginationBoundary = errors.New("activity pagination boundary exceeded")

type activitiesPageToken struct {
	Version              int      `json:"v"`
	Oldest               string   `json:"oldest"`
	Newest               string   `json:"newest,omitempty"`
	RouteID              int64    `json:"route_id,omitempty"`
	IncludeUnnamed       bool     `json:"include_unnamed"`
	IncludeFull          bool     `json:"include_full"`
	PageSize             int      `json:"page_size"`
	Fields               []string `json:"fields,omitempty"`
	BeforeStartDateLocal string   `json:"before_start_date_local,omitempty"`
	BeforeID             string   `json:"before_id,omitempty"`
	SkipIDsAtBoundary    []string `json:"skip_ids_at_boundary,omitempty"`
}

func newGetActivitiesTool(activityClient ActivitiesClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return coreTool(Tool{
		Name:         getActivitiesName,
		Description:  getActivitiesDescription,
		InputSchema:  getActivitiesInputSchema(),
		OutputSchema: getActivitiesOutputSchema(),
		Handler:      getActivitiesHandler(activityClient, profileClient, version, timezoneFallback, debugMetadata),
	})
}

func getActivitiesHandler(activityClient ActivitiesClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		args, token, err := decodeGetActivitiesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetActivitiesArgumentsMessage, err)
		}
		if activityClient == nil || profileClient == nil {
			return Result{}, NewUserError(fetchActivitiesMessage, errors.New("missing activities or profile client"))
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return Result{}, ctxErr
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchActivitiesMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		activityTimezoneFallback := profileTimezone(profile.Timezone, timezoneFallback)
		activities, nextToken, err := fetchActivitiesPage(ctx, activityClient, args, token)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			if errors.Is(err, errActivitiesPaginationBoundary) {
				return Result{}, NewUserError(activitiesPaginationBoundaryMessage, err)
			}
			return Result{}, NewUserError(fetchActivitiesMessage, err)
		}
		shaped, err := shapeGetActivitiesResponse(activities, args, nextToken, version, activityTimezoneFallback, debugMetadata, unitSystem)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_activities response: %w", err)
		}
		if _, err := json.Marshal(shaped); err != nil {
			return Result{}, fmt.Errorf("encoding get_activities response: %w", err)
		}
		return TextResult(shaped), nil
	}
}

func decodeGetActivitiesRequest(raw json.RawMessage) (GetActivitiesRequest, *activitiesPageToken, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return GetActivitiesRequest{}, nil, errors.New("oldest is required unless next_page_token is supplied")
	}
	if trimmed[0] != '{' {
		return GetActivitiesRequest{}, nil, errors.New("arguments must be a JSON object")
	}
	args, err := DecodeStrict[GetActivitiesRequest](trimmed)
	if err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	var supplied activitiesTokenArgs
	if err := json.Unmarshal(trimmed, &supplied); err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	args.PageSize = normalizeActivitiesPageSize(args.PageSize)
	if strings.TrimSpace(args.NextPageToken) == "" {
		if strings.TrimSpace(args.Oldest) == "" {
			return GetActivitiesRequest{}, nil, errors.New("oldest is required")
		}
		return args, nil, nil
	}
	token, err := parseActivitiesPageToken(args.NextPageToken)
	if err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	if err := validateActivitiesTokenArgs(args, token, supplied); err != nil {
		return GetActivitiesRequest{}, nil, err
	}
	args.Oldest = token.Oldest
	args.Newest = token.Newest
	args.RouteID = token.RouteID
	args.IncludeUnnamed = token.IncludeUnnamed
	args.IncludeFull = token.IncludeFull
	args.PageSize = token.PageSize
	return args, token, nil
}

func normalizeActivitiesPageSize(pageSize int) int {
	if pageSize <= 0 {
		return defaultActivitiesPageSize
	}
	if pageSize > maxActivitiesPageSize {
		return maxActivitiesPageSize
	}
	return pageSize
}

type activitiesTokenArgs struct {
	Oldest         *string `json:"oldest"`
	Newest         *string `json:"newest"`
	RouteID        *int64  `json:"route_id"`
	IncludeUnnamed *bool   `json:"include_unnamed"`
	PageSize       *int    `json:"page_size"`
	IncludeFull    *bool   `json:"include_full"`
}

func validateActivitiesTokenArgs(args GetActivitiesRequest, token *activitiesPageToken, supplied activitiesTokenArgs) error {
	if token.Version != 1 {
		return fmt.Errorf("unsupported token version %d", token.Version)
	}
	if token.PageSize <= 0 || token.PageSize > maxActivitiesPageSize || strings.TrimSpace(token.Oldest) == "" {
		return errors.New("invalid token payload")
	}
	if supplied.Oldest != nil && strings.TrimSpace(args.Oldest) != strings.TrimSpace(token.Oldest) {
		return errors.New("oldest does not match next_page_token")
	}
	if supplied.Newest != nil && strings.TrimSpace(args.Newest) != strings.TrimSpace(token.Newest) {
		return errors.New("newest does not match next_page_token")
	}
	if supplied.RouteID != nil && args.RouteID != token.RouteID {
		return errors.New("route_id does not match next_page_token")
	}
	if supplied.PageSize != nil && args.PageSize != token.PageSize {
		return errors.New("page_size does not match next_page_token")
	}
	if supplied.IncludeUnnamed != nil && args.IncludeUnnamed != token.IncludeUnnamed {
		return errors.New("include_unnamed does not match next_page_token")
	}
	if supplied.IncludeFull != nil && args.IncludeFull != token.IncludeFull {
		return errors.New("include_full does not match next_page_token")
	}
	return nil
}

func parseActivitiesPageToken(value string) (*activitiesPageToken, error) {
	data, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("decoding next_page_token: %w", err)
	}
	var token activitiesPageToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("decoding next_page_token JSON: %w", err)
	}
	return &token, nil
}

func encodeActivitiesPageToken(token activitiesPageToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func fetchActivitiesPage(ctx context.Context, client ActivitiesClient, args GetActivitiesRequest, token *activitiesPageToken) ([]intervals.Activity, string, error) {
	cursor := activitiesPageToken{Version: 1, Oldest: args.Oldest, Newest: args.Newest, RouteID: args.RouteID, IncludeUnnamed: args.IncludeUnnamed, IncludeFull: args.IncludeFull, PageSize: args.PageSize}
	if !args.IncludeFull {
		cursor.Fields = append([]string(nil), terseActivityFields...)
	}
	if token != nil {
		cursor.BeforeStartDateLocal = token.BeforeStartDateLocal
		cursor.BeforeID = token.BeforeID
		cursor.SkipIDsAtBoundary = append([]string(nil), token.SkipIDsAtBoundary...)
		cursor.Fields = append([]string(nil), token.Fields...)
	}
	page := make([]intervals.Activity, 0, args.PageSize)
	fetchLimit := min(args.PageSize*2+1, maxActivityFetchLimit)
	lastFullWindow := false
	cursorAdvanced := false
	for fetches := 0; fetches < maxActivityPageFetches; fetches++ {
		newest := effectiveNewest(args.Newest, cursor)
		if activityBoundaryBefore(newest, args.Oldest) {
			lastFullWindow = false
			break
		}
		params := intervals.ListActivitiesParams{Oldest: args.Oldest, Newest: newest, RouteID: args.RouteID, Limit: fetchLimit}
		if !args.IncludeFull {
			params.Fields = cursor.Fields
		}
		activities, err := client.ListActivities(ctx, params)
		if err != nil {
			return nil, "", err
		}
		if len(activities) == 0 {
			lastFullWindow = false
			break
		}
		lastFullWindow = len(activities) == fetchLimit
		sortActivities(activities)
		candidates := activitiesAfterCursor(activities, cursor)
		if len(candidates) == 0 {
			advanced := advanceCursorPast(&cursor, activities[len(activities)-1])
			if !advanced && lastFullWindow && fetchLimit < maxActivityFetchLimit {
				fetchLimit = maxActivityFetchLimit
				continue
			}
			if !advanced && lastFullWindow && fetchLimit >= maxActivityFetchLimit {
				return nil, "", errActivitiesPaginationBoundary
			}
			if !advanced {
				advanced = advanceCursorBeforeBoundary(&cursor)
			}
			cursorAdvanced = advanced || cursorAdvanced
			if !advanced {
				lastFullWindow = false
				break
			}
			continue
		}
		for _, activity := range candidates {
			if !args.IncludeUnnamed && strings.TrimSpace(stringValue(activity.Name)) == "" && !isStravaBlocked(activity) {
				cursorAdvanced = advanceCursorPast(&cursor, activity) || cursorAdvanced
				continue
			}
			if len(page) == args.PageSize {
				if !cursorAdvanced {
					return page, "", nil
				}
				nextToken, err := encodeActivitiesPageToken(cursor)
				if err != nil {
					return nil, "", err
				}
				return page, nextToken, nil
			}
			page = append(page, activity)
			cursorAdvanced = advanceCursorPast(&cursor, activity) || cursorAdvanced
		}
		if len(activities) < fetchLimit {
			break
		}
		if len(page) == args.PageSize {
			if !cursorAdvanced {
				return page, "", nil
			}
			nextToken, err := encodeActivitiesPageToken(cursor)
			if err != nil {
				return nil, "", err
			}
			return page, nextToken, nil
		}
	}
	if lastFullWindow && cursorAdvanced {
		nextToken, err := encodeActivitiesPageToken(cursor)
		if err != nil {
			return nil, "", err
		}
		return page, nextToken, nil
	}
	return page, "", nil
}

func effectiveNewest(newest string, cursor activitiesPageToken) string {
	if cursor.BeforeStartDateLocal != "" {
		return cursor.BeforeStartDateLocal
	}
	return newest
}

func sortActivities(activities []intervals.Activity) {
	sort.SliceStable(activities, func(i, j int) bool {
		leftDate := activitySortDate(activities[i])
		rightDate := activitySortDate(activities[j])
		if leftDate != rightDate {
			return leftDate > rightDate
		}
		return activities[i].ID > activities[j].ID
	})
}

func activitiesAfterCursor(activities []intervals.Activity, cursor activitiesPageToken) []intervals.Activity {
	if cursor.BeforeStartDateLocal == "" {
		return activities
	}
	out := make([]intervals.Activity, 0, len(activities))
	skips := make(map[string]bool, len(cursor.SkipIDsAtBoundary))
	for _, id := range cursor.SkipIDsAtBoundary {
		skips[id] = true
	}
	for _, activity := range activities {
		date := activitySortDate(activity)
		if date > cursor.BeforeStartDateLocal {
			continue
		}
		if date == cursor.BeforeStartDateLocal {
			if skips[activity.ID] || (cursor.BeforeID != "" && activity.ID >= cursor.BeforeID) {
				continue
			}
		}
		out = append(out, activity)
	}
	return out
}

func advanceCursorBeforeBoundary(cursor *activitiesPageToken) bool {
	if cursor.BeforeStartDateLocal == "" {
		return false
	}
	before := justBeforeActivityTimestamp(cursor.BeforeStartDateLocal)
	if before == "" || before == cursor.BeforeStartDateLocal {
		return false
	}
	cursor.BeforeStartDateLocal = before
	cursor.BeforeID = ""
	cursor.SkipIDsAtBoundary = nil
	return true
}

func activityBoundaryBefore(left string, right string) bool {
	leftTime, leftOK := parseActivityBoundary(left)
	rightTime, rightOK := parseActivityBoundary(right)
	return leftOK && rightOK && leftTime.Before(rightTime)
}

func parseActivityBoundary(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	layouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func justBeforeActivityTimestamp(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	layouts := []string{"2006-01-02T15:04:05", "2006-01-02T15:04", time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" {
			return parsed.AddDate(0, 0, -1).Format(layout)
		}
		return parsed.Add(-time.Second).Format(layout)
	}
	return ""
}

func advanceCursorPast(cursor *activitiesPageToken, activity intervals.Activity) bool {
	date := activitySortDate(activity)
	if date == "" {
		return false
	}
	if cursor.BeforeStartDateLocal != date {
		cursor.BeforeStartDateLocal = date
		cursor.BeforeID = activity.ID
		cursor.SkipIDsAtBoundary = []string{activity.ID}
		return true
	}
	if slices.Contains(cursor.SkipIDsAtBoundary, activity.ID) {
		return false
	}
	cursor.BeforeID = activity.ID
	cursor.SkipIDsAtBoundary = append(cursor.SkipIDsAtBoundary, activity.ID)
	return true
}

func activitySortDate(activity intervals.Activity) string {
	if value := stringValue(activity.StartDateLocal); value != "" {
		return value
	}
	return stringValue(activity.StartDate)
}

func shapeGetActivitiesResponse(activities []intervals.Activity, args GetActivitiesRequest, nextToken string, version string, timezoneFallback string, debugMetadata bool, unitSystem response.UnitSystem) (any, error) {
	rows := make([]getActivitiesRow, 0, len(activities))
	for _, activity := range activities {
		rows = append(rows, activityRow(activity, args.IncludeFull, timezoneFallback, unitSystem))
	}
	payload := getActivitiesResponse{Activities: rows, Meta: getActivitiesMeta{PageSize: args.PageSize, NextPageToken: nextToken, MoreAvailable: nextToken != "", IncludeFull: args.IncludeFull}}
	return response.Shape(payload, response.Options{IncludeFull: args.IncludeFull, RowCollections: []string{"activities"}, ServerVersion: version, DebugMetadata: debugMetadata, QueryType: getActivitiesName, UnitSystem: unitSystem})
}

func activityRow(activity intervals.Activity, includeFull bool, timezoneFallback string, unitSystem response.UnitSystem) getActivitiesRow {
	row := getActivitiesRow{ActivityID: activity.ID, Name: strings.TrimSpace(stringValue(activity.Name)), Sport: stringValue(activity.Type), SubType: stringValue(activity.SubType), StartDateLocal: stringValue(activity.StartDateLocal), StartDateUTC: stringValue(activity.StartDate), Timezone: firstNonEmpty(stringValue(activity.Timezone), timezoneFallback)}
	if isStravaBlocked(activity) {
		row.StravaImported = true
		row.Unavailable = &unavailableReason{Reason: "strava_tos", Workaround: stravaWorkaround}
		if includeFull {
			row.Full = activity.Raw
		}
		return row
	}
	row.MovingTimeSeconds = intValue(activity.MovingTime)
	row.ElapsedTimeSeconds = intValue(activity.ElapsedTime)
	row.TrainingLoad = intValue(activity.TrainingLoad)
	row.AverageHeartRateBPM = intValue(activity.AverageHeartRate)
	row.MaxHeartRateBPM = intValue(activity.MaxHeartRate)
	row.AverageCadenceRPM = activity.AverageCadence
	row.CaloriesBurned = intValue(activity.Calories)
	row.DeviceName = stringValue(activity.DeviceName)
	row.HasStreams = len(activity.StreamTypes) > 0
	if activity.TotalElevationGain != nil {
		value := *activity.TotalElevationGain
		row.ElevationGainM = &value
	}
	if activity.TotalElevationLoss != nil {
		value := *activity.TotalElevationLoss
		row.ElevationLossM = &value
	}
	applyActivityDistanceAndPace(&row, activity, unitSystem)
	applyActivitySpeed(&row, activity.AverageSpeed, true, unitSystem)
	applyActivitySpeed(&row, activity.MaxSpeed, false, unitSystem)
	if includeFull {
		row.Full = activity.Raw
	}
	return row
}

func applyActivityDistanceAndPace(row *getActivitiesRow, activity intervals.Activity, unitSystem response.UnitSystem) {
	distanceMeters := firstFloat(activity.ICUDistance, activity.Distance)
	if distanceMeters == nil || *distanceMeters <= 0 {
		return
	}
	converted := response.ToPreferred(*distanceMeters, units.UnitM, unitSystem)
	value := round(converted.Value, 3)
	if converted.Unit == units.UnitMI {
		row.DistanceMI = &value
	} else {
		row.DistanceKM = &value
	}
	if row.MovingTimeSeconds > 0 && isRunLikeActivity(activity) {
		pace := float64(row.MovingTimeSeconds) / converted.Value
		pace = round(pace, 1)
		if converted.Unit == units.UnitMI {
			row.PaceSecondsPerMile = &pace
		} else {
			row.PaceSecondsPerKM = &pace
		}
	}
}

func applyActivitySpeed(row *getActivitiesRow, speed *float64, average bool, unitSystem response.UnitSystem) {
	if speed == nil || *speed <= 0 {
		return
	}
	converted := response.ToPreferred(*speed, units.UnitMS, unitSystem)
	value := round(converted.Value, 3)
	if average {
		if converted.Unit == units.UnitMPH {
			row.AverageSpeedMPH = &value
		} else {
			row.AverageSpeedKMH = &value
		}
		return
	}
	if converted.Unit == units.UnitMPH {
		row.MaxSpeedMPH = &value
	} else {
		row.MaxSpeedKMH = &value
	}
}

func isRunLikeActivity(activity intervals.Activity) bool {
	sport := strings.ToLower(strings.TrimSpace(stringValue(activity.Type) + " " + stringValue(activity.SubType)))
	return strings.Contains(sport, "run") || strings.Contains(sport, "jog")
}

func isStravaBlocked(activity intervals.Activity) bool {
	source := strings.ToLower(stringValue(activity.Source))
	note := strings.ToLower(stringValue(activity.Note))
	if strings.Contains(source, "strava") || strings.Contains(note, "strava") {
		return true
	}
	if source != "" {
		return false
	}
	meaningful := 0
	for _, key := range []string{"name", "type", "distance", "icu_distance", "moving_time", "elapsed_time"} {
		if value, ok := activity.Raw[key]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
			meaningful++
		}
	}
	if meaningful > 0 {
		return false
	}
	if len(activity.Raw) == 0 {
		return true
	}
	_, hasHiddenNote := activity.Raw["_note"]
	if hasHiddenNote {
		return true
	}
	stubKeys := map[string]bool{"id": true, "icu_athlete_id": true, "athlete_id": true, "start_date_local": true, "start_date": true, "external_id": true}
	nullableStubKeys := map[string]bool{"name": true, "type": true, "sub_type": true, "distance": true, "icu_distance": true, "moving_time": true, "elapsed_time": true, "source": true, "_note": true}
	for key, value := range activity.Raw {
		if stubKeys[key] {
			continue
		}
		if value == nil && nullableStubKeys[key] {
			continue
		}
		return false
	}
	return true
}

func firstFloat(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func anyString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func round(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}

func getActivitiesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"oldest":          map[string]any{"type": "string", "description": "Required local ISO-8601 start date/date-time unless next_page_token is supplied."},
		"newest":          map[string]any{"type": "string", "description": "Optional local ISO-8601 end date/date-time; defaults upstream to now."},
		"route_id":        map[string]any{"type": "integer", "description": "Optional intervals.icu route ID filter."},
		"include_unnamed": map[string]any{"type": "boolean", "default": false, "description": "When false, drop rows with an empty activity name after bounded pagination."},
		"page_size":       map[string]any{"type": "integer", "default": defaultActivitiesPageSize, "minimum": 1, "maximum": maxActivitiesPageSize, "description": "Number of terse rows to return per page; values above 200 are capped."},
		"next_page_token": map[string]any{"type": "string", "description": "Opaque token from _meta.next_page_token for the next page. Do not edit."},
		"include_full":    map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream activity fields and preserve upstream nulls; default terse rows are unit-disambiguated and null-stripped."},
	}}
}

func getActivitiesOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Paginated activities with unit-disambiguated terse rows, Strava unavailable markers, and _meta.next_page_token when more data may be available."}
}
