package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getActivitiesName                    = "get_activities"
	getActivitiesDescription             = "Scan an athlete-local date window and return a paginated activity index: terse unit-disambiguated summary rows with IDs, source/device hints, tags, explicitly requested custom_fields, and Strava-unavailable markers. Rows include calories_burned as active/exercise calories (distinct from wellness kcal_consumed), carbs_ingested_g for athlete-logged carb intake, carbs_used_g for upstream carbs-burned estimate, selected athlete-defined activity custom fields, historical activity weather when Intervals.icu provides it, and opaque pagination. Weather values are completed-activity historical context with provenance, not forecast conditions. Use this before details, intervals, streams, splits, or messages when a prompt describes an activity by date, relative date, or recent training."
	invalidGetActivitiesArgumentsMessage = "invalid get_activities arguments; provide oldest/newest dates or a valid next_page_token"
	fetchActivitiesMessage               = "could not fetch activities; check intervals.icu credentials, athlete ID, and date range"
	activitiesPaginationBoundaryMessage  = "activity pagination hit too many same-timestamp filtered rows; narrow the date range or set include_unnamed true"
	defaultActivitiesPageSize            = 50
	maxActivitiesPageSize                = 200
	maxActivityPageFetches               = 5
	maxActivityFetchLimit                = 201
	stravaUnknownProviderWorkaround      = "Open the intervals.icu Connections page for the activity's original device provider and click Download old data so historical activities are re-imported directly from that provider instead of through Strava's restricted API."
)

var terseActivityFields = []string{
	"id", "name", "type", "sub_type", "start_date_local", "start_date", "timezone",
	"source", "_note", "icu_athlete_id", "external_id", "stream_types",
	"distance", "icu_distance", "moving_time", "elapsed_time", "average_speed", "max_speed",
	"has_weather", "average_weather_temp", "min_weather_temp", "max_weather_temp", "average_wind_speed", "average_wind_gust", "prevailing_wind_deg", "headwind_percent", "tailwind_percent",
	"total_elevation_gain", "total_elevation_loss", "icu_training_load", "average_heartrate",
	"max_heartrate", "average_cadence", "calories", "carbs_ingested", "carbs_used",
	"device_name", "gear_id", "tags",
}

// terseActivityFieldsWithCustom returns the terse upstream field set extended
// with athlete-defined activity custom field codes so intervals.icu includes
// them in field-limited terse list responses.
func terseActivityFieldsWithCustom(customFieldCodes []string) []string {
	fields := append([]string(nil), terseActivityFields...)
	if len(customFieldCodes) == 0 {
		return fields
	}
	seen := make(map[string]bool, len(fields))
	for _, field := range fields {
		seen[field] = true
	}
	for _, code := range customFieldCodes {
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		fields = append(fields, code)
	}
	return fields
}

// ActivitiesClient lists intervals.icu activities for tools.
type ActivitiesClient interface {
	ListActivities(context.Context, intervals.ListActivitiesParams) ([]intervals.Activity, error)
}

// GetActivitiesRequest contains get_activities arguments.
type GetActivitiesRequest struct {
	Oldest         string   `json:"oldest,omitempty"`
	Newest         string   `json:"newest,omitempty"`
	RouteID        int64    `json:"route_id,omitempty"`
	IncludeUnnamed bool     `json:"include_unnamed,omitempty"`
	PageSize       int      `json:"page_size,omitempty"`
	NextPageToken  string   `json:"next_page_token,omitempty"`
	CustomFields   []string `json:"custom_fields,omitempty"`
	IncludeFull    bool     `json:"include_full,omitempty"`
}

type getActivitiesResponse struct {
	Activities []getActivitiesRow `json:"activities"`
	Meta       getActivitiesMeta  `json:"_meta"`
}

type getActivitiesRow struct {
	ActivityID           string                 `json:"activity_id,omitempty"`
	Name                 string                 `json:"name,omitempty"`
	Sport                string                 `json:"sport,omitempty"`
	SubType              string                 `json:"sub_type,omitempty"`
	StartDateLocal       string                 `json:"start_date_local,omitempty"`
	StartDateUTC         string                 `json:"start_date_utc,omitempty"`
	Timezone             string                 `json:"timezone,omitempty"`
	MovingTimeSeconds    int                    `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds   int                    `json:"elapsed_time_seconds,omitempty"`
	DistanceKM           *float64               `json:"distance_km,omitempty"`
	DistanceMI           *float64               `json:"distance_mi,omitempty"`
	PaceSecondsPerKM     *float64               `json:"pace_seconds_per_km,omitempty"`
	PaceSecondsPerMile   *float64               `json:"pace_seconds_per_mile,omitempty"`
	AverageSpeedKMH      *float64               `json:"average_speed_kmh,omitempty"`
	AverageSpeedMPH      *float64               `json:"average_speed_mph,omitempty"`
	MaxSpeedKMH          *float64               `json:"max_speed_kmh,omitempty"`
	MaxSpeedMPH          *float64               `json:"max_speed_mph,omitempty"`
	ElevationGainM       *float64               `json:"elevation_gain_m,omitempty"`
	ElevationLossM       *float64               `json:"elevation_loss_m,omitempty"`
	TrainingLoad         int                    `json:"training_load,omitempty"`
	AverageHeartRateBPM  int                    `json:"average_heart_rate_bpm,omitempty"`
	MaxHeartRateBPM      int                    `json:"max_heart_rate_bpm,omitempty"`
	AverageCadenceRPM    *float64               `json:"average_cadence_rpm,omitempty"`
	CaloriesBurned       *int                   `json:"calories_burned,omitempty"`
	CarbsIngestedG       *int                   `json:"carbs_ingested_g,omitempty"`
	CarbsUsedG           *int                   `json:"carbs_used_g,omitempty"`
	Weather              *activityWeather       `json:"weather,omitempty"`
	DeviceName           string                 `json:"device_name,omitempty"`
	GearID               string                 `json:"gear_id,omitempty"`
	GearName             string                 `json:"gear_name,omitempty"`
	GearResolution       string                 `json:"gear_resolution,omitempty"`
	WorkoutStatus        string                 `json:"workout_status,omitempty"`
	WorkoutStatusCaveats []string               `json:"workout_status_caveats,omitempty"`
	PairedEventID        string                 `json:"paired_event_id,omitempty"`
	HasStreams           bool                   `json:"has_streams,omitempty"`
	StravaImported       bool                   `json:"strava_imported,omitempty"`
	Unavailable          *unavailableReason     `json:"unavailable,omitempty"`
	Tags                 *[]string              `json:"tags,omitempty"`
	CustomFields         map[string]any         `json:"custom_fields,omitempty"`
	HypoxicLoadCaveat    *hypoxicTrainingCaveat `json:"hypoxic_training_caveat,omitempty"`
	Full                 map[string]any         `json:"full,omitempty"`
}

type activityWeather struct {
	Status             string   `json:"status"`
	Provenance         string   `json:"provenance"`
	AverageTempC       *float64 `json:"average_temp_c,omitempty"`
	MinTempC           *float64 `json:"min_temp_c,omitempty"`
	MaxTempC           *float64 `json:"max_temp_c,omitempty"`
	AverageWindSpeedMS *float64 `json:"average_wind_speed_m_s,omitempty"`
	AverageWindGustMS  *float64 `json:"average_wind_gust_m_s,omitempty"`
	PrevailingWindDeg  *int     `json:"prevailing_wind_deg,omitempty"`
	HeadwindPercent    *float64 `json:"headwind_percent,omitempty"`
	TailwindPercent    *float64 `json:"tailwind_percent,omitempty"`
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
	Timezone       string            `json:"timezone,omitempty"`
	FieldSemantics map[string]string `json:"field_semantics,omitempty"`
	AsOf           string            `json:"as_of,omitempty"`
	AsOfDate       string            `json:"as_of_date,omitempty"`
	AsOfWeekday    string            `json:"as_of_weekday,omitempty"`
}

var errActivitiesPaginationBoundary = errors.New("activity pagination boundary exceeded")

func newGetActivitiesToolWithGear(activityClient ActivitiesClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	return newGetActivitiesToolWithGearAndClock(activityClient, profileClient, gearClient, gearCache, customFieldClient, customFieldCache, version, timezoneFallback, debugMetadata, time.Now, shaping...)
}

func newGetActivitiesToolWithGearAndClock(activityClient ActivitiesClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{
		Name:         getActivitiesName,
		Description:  getActivitiesDescription,
		InputSchema:  getActivitiesInputSchema(),
		OutputSchema: getActivitiesOutputSchema(),
		Handler:      getActivitiesHandler(activityClient, profileClient, gearClient, gearCache, customFieldClient, customFieldCache, version, timezoneFallback, debugMetadata, now, shapeCfg),
	})
}

func getActivitiesHandler(activityClient ActivitiesClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
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
		asOfMeta, err := currentDayAsOfMetadata(now, activityTimezoneFallback, args.Oldest, args.Newest)
		if err != nil {
			return Result{}, NewUserError(fetchActivitiesMessage, err)
		}
		targetAthleteID, _ := intervals.TargetAthleteIDFromContext(ctx)
		if token != nil && token.AthleteID != targetAthleteID {
			return Result{}, NewUserError(invalidGetActivitiesArgumentsMessage, errors.New("next_page_token athlete does not match resolved athlete"))
		}
		customFieldCodes, err := selectedActivityCustomFieldCodes(ctx, customFieldClient, customFieldCache, args.CustomFields)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(activityCustomFieldSelectionMessage(err, invalidGetActivitiesArgumentsMessage), err)
		}
		activities, nextToken, err := fetchActivitiesPage(ctx, activityClient, args, token, targetAthleteID, customFieldCodes)
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
		shaped, err := shapeGetActivitiesResponse(activities, gearResolutions, args, nextToken, version, activityTimezoneFallback, debugMetadata, unitSystem, customFieldCodes, asOfMeta, shapeCfg)
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
		"custom_fields":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": maxSelectedActivityCustomFields, "description": "Optional athlete-defined activity custom field codes to fetch and expose under custom_fields; defaults to none to keep terse responses small."},
		"include_full":    map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream activity fields and preserve upstream nulls; default terse rows are unit-disambiguated and null-stripped."},
	}}
}

func getActivitiesOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Paginated activities with unit-disambiguated terse rows, upstream tags when intervals.icu returns a string-array tags field, calories_burned for active/exercise calories (distinct from wellness kcal_consumed intake), carbs_ingested_g for athlete-logged carb intake during activity, carbs_used_g for upstream carbs-burned estimate, Strava unavailable markers, gear_id/gear_name when upstream permits, and gear_resolution values resolved/name_missing/unresolved/lookup_unavailable so unresolved IDs are never guessed. custom_fields holds explicitly requested athlete-defined activity custom field values keyed by the upstream field code when intervals.icu returns them. activities[].weather is emitted only when Intervals.icu returns has_weather=true for a completed activity; it contains historical activity weather with provenance, temperatures in degrees C, wind speed/gust in m/s, wind direction in degrees, and headwind/tailwind percentages. Do not treat activities[].weather as a forecast for planned events or future workouts. Each row's timezone is the IANA zone its start_date_local is in, and _meta.timezone is the athlete's configured timezone; start_date_utc is UTC. When the requested range includes the athlete-local current day, _meta also includes as_of, as_of_date, and as_of_weekday. Derive calendar dates from these timezones so activities are not reported on the wrong day."}
}
