package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ricardocabral/icuvisor/internal/intervals"
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
	"max_heartrate", "average_cadence", "calories", "device_name", "gear_id",
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
	CaloriesBurned      *int               `json:"calories_burned,omitempty"`
	DeviceName          string             `json:"device_name,omitempty"`
	GearID              string             `json:"gear_id,omitempty"`
	GearName            string             `json:"gear_name,omitempty"`
	GearResolution      string             `json:"gear_resolution,omitempty"`
	HasStreams          bool               `json:"has_streams,omitempty"`
	StravaImported      bool               `json:"strava_imported,omitempty"`
	Unavailable         *unavailableReason `json:"unavailable,omitempty"`
	Full                map[string]any     `json:"full,omitempty"`
}

type unavailableReason struct {
	Reason     string `json:"reason"`
	Workaround string `json:"workaround,omitempty"`
}

type getActivitiesMeta struct {
	PageSize       int               `json:"page_size"`
	NextPageToken  string            `json:"next_page_token,omitempty"`
	MoreAvailable  bool              `json:"more_available"`
	IncludeFull    bool              `json:"include_full"`
	FieldSemantics map[string]string `json:"field_semantics,omitempty"`
}

var errActivitiesPaginationBoundary = errors.New("activity pagination boundary exceeded")

func newGetActivitiesToolWithGear(activityClient ActivitiesClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{
		Name:         getActivitiesName,
		Description:  getActivitiesDescription,
		InputSchema:  getActivitiesInputSchema(),
		OutputSchema: getActivitiesOutputSchema(),
		Handler:      getActivitiesHandler(activityClient, profileClient, gearClient, gearCache, version, timezoneFallback, debugMetadata, shapeCfg),
	})
}

func getActivitiesHandler(activityClient ActivitiesClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
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
		targetAthleteID, _ := intervals.TargetAthleteIDFromContext(ctx)
		if token != nil && token.AthleteID != targetAthleteID {
			return Result{}, NewUserError(invalidGetActivitiesArgumentsMessage, errors.New("next_page_token athlete does not match resolved athlete"))
		}
		activities, nextToken, err := fetchActivitiesPage(ctx, activityClient, args, token, targetAthleteID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			if errors.Is(err, errActivitiesPaginationBoundary) {
				return Result{}, NewUserError(activitiesPaginationBoundaryMessage, err)
			}
			return Result{}, NewUserError(fetchActivitiesMessage, err)
		}
		gearResolutions, err := resolveActivityGear(ctx, gearClient, gearCache, activities)
		if err != nil {
			return Result{}, err
		}
		shaped, err := shapeGetActivitiesResponse(activities, gearResolutions, args, nextToken, version, activityTimezoneFallback, debugMetadata, unitSystem, shapeCfg)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_activities response: %w", err)
		}
		if _, err := json.Marshal(shaped); err != nil {
			return Result{}, fmt.Errorf("encoding get_activities response: %w", err)
		}
		return TextResult(shaped), nil
	}
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
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Paginated activities with unit-disambiguated terse rows, Strava unavailable markers, gear_id/gear_name when upstream permits, and gear_resolution values resolved/name_missing/unresolved/lookup_unavailable so unresolved IDs are never guessed."}
}
