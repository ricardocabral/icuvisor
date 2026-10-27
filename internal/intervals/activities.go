package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

const maxActivityCarbsIngested = 1<<31 - 1

// ListActivitiesParams contains query parameters for listing athlete activities.
type ListActivitiesParams struct {
	Oldest  string
	Newest  string
	RouteID int64
	Limit   int
	Fields  []string
}

// ActivitiesAroundParams contains query parameters for finding activities around a reference activity.
type ActivitiesAroundParams struct {
	ActivityID string
	Limit      int
}

// LinkActivityToEventParams contains the IDs needed to pair an activity with a planned event.
type LinkActivityToEventParams struct {
	ActivityID string
	EventID    string
}

type linkActivityToEventPayload struct {
	PairedEventID int `json:"paired_event_id"`
}

// UpdateActivityParams contains sparse activity metadata fields to update.
// Set the *Set fields to true to mark a field as explicitly provided so callers
// can distinguish "leave unchanged" from "clear".
type UpdateActivityParams struct {
	ActivityID       string
	Name             string
	NameSet          bool
	Description      string
	DescriptionSet   bool
	CarbsIngested    int
	CarbsIngestedSet bool
	Tags             *[]string
	Commute          *bool
	Feel             *int
	RPE              *int
	SubType          *string
	DistanceM        *float64
}

// Validate checks sparse fields without making any upstream calls.
func (p UpdateActivityParams) Validate() error {
	if strings.TrimSpace(p.ActivityID) == "" {
		return fmt.Errorf("activity ID is required")
	}
	if !p.NameSet && !p.DescriptionSet && !p.CarbsIngestedSet && p.Tags == nil && p.Commute == nil && p.Feel == nil && p.RPE == nil && p.SubType == nil && p.DistanceM == nil {
		return fmt.Errorf("at least one writable field is required")
	}
	if p.CarbsIngestedSet && (p.CarbsIngested < 0 || p.CarbsIngested > maxActivityCarbsIngested) {
		return fmt.Errorf("carbs_ingested must be from 0 through %d", maxActivityCarbsIngested)
	}
	if p.Feel != nil && (*p.Feel < 1 || *p.Feel > 5) {
		return fmt.Errorf("feel must be from 1 through 5")
	}
	if p.RPE != nil && (*p.RPE < 1 || *p.RPE > 10) {
		return fmt.Errorf("rpe must be from 1 through 10")
	}
	if p.SubType != nil {
		switch *p.SubType {
		case "NONE", "COMMUTE", "WARMUP", "COOLDOWN", "RACE":
		default:
			return fmt.Errorf("sub_type must be NONE, COMMUTE, WARMUP, COOLDOWN, or RACE")
		}
	}
	if p.DistanceM != nil && (math.IsNaN(*p.DistanceM) || math.IsInf(*p.DistanceM, 0) || *p.DistanceM < 0 || *p.DistanceM > math.MaxFloat32) {
		return fmt.Errorf("distance_m must be non-negative finite metres representable upstream")
	}
	if p.Tags != nil {
		if len(*p.Tags) > 100 {
			return fmt.Errorf("tags must contain at most 100 entries")
		}
		for _, tag := range *p.Tags {
			if strings.TrimSpace(tag) == "" || len(tag) > 128 {
				return fmt.Errorf("tags must be non-empty and at most 128 bytes each")
			}
		}
	}
	return nil
}

type updateActivityPayload struct {
	Tags          *[]string `json:"tags,omitempty"`
	Commute       *bool     `json:"commute,omitempty"`
	Feel          *int      `json:"feel,omitempty"`
	RPE           *int      `json:"icu_rpe,omitempty"`
	SubType       *string   `json:"sub_type,omitempty"`
	DistanceM     *float64  `json:"distance,omitempty"`
	Name          *string   `json:"name,omitempty"`
	Description   *string   `json:"description,omitempty"`
	CarbsIngested *int      `json:"carbs_ingested,omitempty"`
}

// Activity contains stable activity fields used by read tools and preserves raw upstream fields.
type Activity struct {
	Raw                  map[string]any                `json:"-"`
	MetadataVerification *ActivityMetadataVerification `json:"-"`

	ChatID             *int     `json:"icu_chat_id"`
	ID                 string   `json:"id"`
	Name               *string  `json:"name"`
	Type               *string  `json:"type"`
	SubType            *string  `json:"sub_type"`
	StartDateLocal     *string  `json:"start_date_local"`
	StartDate          *string  `json:"start_date"`
	Timezone           *string  `json:"timezone"`
	Source             *string  `json:"source"`
	Note               *string  `json:"_note"`
	ICUAthleteID       *string  `json:"icu_athlete_id"`
	ExternalID         *string  `json:"external_id"`
	Distance           *float64 `json:"distance"`
	ICUDistance        *float64 `json:"icu_distance"`
	MovingTime         *int     `json:"moving_time"`
	ElapsedTime        *int     `json:"elapsed_time"`
	TotalElevationGain *float64 `json:"total_elevation_gain"`
	TotalElevationLoss *float64 `json:"total_elevation_loss"`
	AverageSpeed       *float64 `json:"average_speed"`
	MaxSpeed           *float64 `json:"max_speed"`
	HasWeather         *bool    `json:"has_weather"`
	AverageWeatherTemp *float64 `json:"average_weather_temp"`
	MinWeatherTemp     *float64 `json:"min_weather_temp"`
	MaxWeatherTemp     *float64 `json:"max_weather_temp"`
	AverageWindSpeed   *float64 `json:"average_wind_speed"`
	AverageWindGust    *float64 `json:"average_wind_gust"`
	PrevailingWindDeg  *int     `json:"prevailing_wind_deg"`
	HeadwindPercent    *float64 `json:"headwind_percent"`
	TailwindPercent    *float64 `json:"tailwind_percent"`
	TrainingLoad       *int     `json:"icu_training_load"`
	AverageHeartRate   *int     `json:"average_heartrate"`
	MaxHeartRate       *int     `json:"max_heartrate"`
	AverageCadence     *float64 `json:"average_cadence"`
	AverageWatts       *int     `json:"icu_average_watts"`
	Calories           *int     `json:"calories"`
	CarbsIngested      *int     `json:"carbs_ingested"`
	CarbsUsed          *int     `json:"carbs_used"`
	DeviceName         *string  `json:"device_name"`
	Commute            *bool    `json:"commute"`
	Feel               *int     `json:"feel"`
	RPE                *int     `json:"icu_rpe"`
	GearID             string   `json:"-"`
	StreamTypes        []string `json:"stream_types"`
}

// UnmarshalJSON decodes Activity while retaining the original object for full responses.
func (a *Activity) UnmarshalJSON(data []byte) error {
	type activityAlias Activity
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// Keep the raw payload lossless, but do not let a malformed optional
	// subjective field make the whole activity unreadable. The normalized
	// fields are populated below only when their upstream JSON types are exact.
	normalizedData := make(map[string]any, len(raw))
	for key, value := range raw {
		switch key {
		case "commute", "feel", "icu_rpe":
			continue
		default:
			normalizedData[key] = value
		}
	}
	dataWithoutNormalizedFields, err := json.Marshal(normalizedData)
	if err != nil {
		return err
	}
	var decoded activityAlias
	if err := json.Unmarshal(dataWithoutNormalizedFields, &decoded); err != nil {
		return err
	}
	*a = Activity(decoded)
	a.Raw = raw
	a.Commute = rawActivityBool(raw["commute"])
	a.Feel = rawActivityScaleInt(raw["feel"])
	a.RPE = rawActivityScaleInt(raw["icu_rpe"])
	a.GearID = rawIDString(raw["gear_id"])
	return nil
}

func rawActivityBool(value any) *bool {
	parsed, ok := value.(bool)
	if !ok {
		return nil
	}
	return &parsed
}

func rawActivityScaleInt(value any) *int {
	parsed, ok := value.(float64)
	if !ok || parsed != float64(int(parsed)) {
		return nil
	}
	out := int(parsed)
	if float64(out) != parsed {
		return nil
	}
	return &out
}

// ListActivities lists activities in descending date order for the configured athlete.
func (c *Client) ListActivities(ctx context.Context, params ListActivitiesParams) ([]Activity, error) {
	query := url.Values{}
	oldest := strings.TrimSpace(params.Oldest)
	if oldest == "" {
		return nil, fmt.Errorf("listing activities: oldest is required")
	}
	query.Set("oldest", oldest)
	if newest := strings.TrimSpace(params.Newest); newest != "" {
		query.Set("newest", newest)
	}
	if params.RouteID > 0 {
		query.Set("route_id", strconv.FormatInt(params.RouteID, 10))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if len(params.Fields) > 0 {
		fields := compactStrings(params.Fields)
		if len(fields) > 0 {
			query.Set("fields", strings.Join(fields, ","))
		}
	}

	var activities []Activity
	if err := c.doJSONQuery(ctx, &activities, query, "athlete", c.athleteID, "activities"); err != nil {
		return nil, fmt.Errorf("listing activities: %w", err)
	}
	return activities, nil
}

// ListActivitiesAround lists activities near a reference activity for the configured athlete.
func (c *Client) ListActivitiesAround(ctx context.Context, params ActivitiesAroundParams) ([]Activity, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if activityID == "" {
		return nil, fmt.Errorf("listing activities around activity: activity ID is required")
	}
	query := url.Values{}
	query.Set("activity_id", activityID)
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}

	var activities []Activity
	if err := c.doJSONQuery(ctx, &activities, query, "athlete", c.athleteID, "activities-around"); err != nil {
		return nil, fmt.Errorf("listing activities around activity %s: %w", activityID, err)
	}
	return activities, nil
}

// LinkActivityToEvent sets an activity's paired_event_id to a planned event ID.
func (c *Client) LinkActivityToEvent(ctx context.Context, params LinkActivityToEventParams) (Activity, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if activityID == "" {
		return Activity{}, fmt.Errorf("linking activity to event: activity ID is required")
	}
	eventID, err := ParseEventID(params.EventID)
	if err != nil {
		return Activity{}, fmt.Errorf("linking activity %s to event: %w", activityID, err)
	}
	if err := c.ensureActivityIDTarget(ctx, activityID); err != nil {
		return Activity{}, fmt.Errorf("linking activity %s to event %s: %w", activityID, params.EventID, err)
	}
	if _, ok := targetAthleteIDFromContext(ctx); ok {
		if _, err := c.GetEvent(ctx, params.EventID); err != nil {
			return Activity{}, fmt.Errorf("linking activity %s to event %s: %w", activityID, params.EventID, err)
		}
	}
	var activity Activity
	if err := c.doJSONBody(ctx, http.MethodPut, linkActivityToEventPayload{PairedEventID: eventID}, &activity, "activity", activityID); err != nil {
		return Activity{}, fmt.Errorf("linking activity %s to event %s: %w", activityID, params.EventID, err)
	}
	if err := c.ensureActivityTarget(ctx, activity); err != nil {
		return Activity{}, fmt.Errorf("linking activity %s to event %s: %w", activityID, params.EventID, err)
	}
	return activity, nil
}

// UpdateActivity applies a sparse PUT to /activity/{id} for activity metadata.
// Only fields with their corresponding *Set flag are sent; other fields are
// left untouched upstream. Returns the upstream activity after the update.
func (c *Client) UpdateActivity(ctx context.Context, params UpdateActivityParams) (Activity, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if err := params.Validate(); err != nil {
		return Activity{}, fmt.Errorf("updating activity: %w", err)
	}
	payload := updateActivityPayload{Tags: params.Tags, Commute: params.Commute, Feel: params.Feel, RPE: params.RPE, SubType: params.SubType, DistanceM: params.DistanceM}
	if params.DistanceM != nil {
		activity, err := c.GetActivity(ctx, activityID)
		if err != nil {
			return Activity{}, fmt.Errorf("checking manual distance update: %w", err)
		}
		if activity.Source == nil || *activity.Source != "MANUAL" || strings.TrimSpace(rawIDString(activity.Raw["file_type"])) != "" {
			return Activity{}, fmt.Errorf("distance_m corrections require a manual activity without a recorded file")
		}
	}
	if params.NameSet {
		name := params.Name
		payload.Name = &name
	}
	if params.DescriptionSet {
		description := params.Description
		payload.Description = &description
	}
	if params.CarbsIngestedSet {
		carbsIngested := params.CarbsIngested
		payload.CarbsIngested = &carbsIngested
	}
	if err := c.ensureActivityIDTarget(ctx, activityID); err != nil {
		return Activity{}, fmt.Errorf("updating activity %s: %w", activityID, err)
	}
	var activity Activity
	if err := c.doJSONBody(ctx, http.MethodPut, payload, &activity, "activity", activityID); err != nil {
		return Activity{}, fmt.Errorf("updating activity %s: %w", activityID, err)
	}
	if err := c.ensureActivityTarget(ctx, activity); err != nil {
		return Activity{}, fmt.Errorf("updating activity %s: %w", activityID, err)
	}
	if verification := VerifyActivityMetadataUpdate(activity, params); verification != nil {
		if verification.Status != "verified" {
			readback, readErr := c.GetActivity(ctx, activityID)
			if readErr == nil {
				activity = readback
				verification = VerifyActivityMetadataUpdate(activity, params)
				verification.Source = "readback"
			} else {
				if ctx.Err() != nil {
					return Activity{}, ctx.Err()
				}
				verification.Status = "verification_unavailable"
				verification.Caveat = "write may have succeeded; read the activity before retrying because readback failed"
			}
		}
		activity.MetadataVerification = verification
	}
	return activity, nil
}

// ActivityMetadataVerification reports evidence for the explicitly requested new metadata fields.
type ActivityMetadataVerification struct {
	Status     string   `json:"status"`
	Source     string   `json:"source"`
	Confirmed  []string `json:"confirmed_fields,omitempty"`
	Missing    []string `json:"missing_fields,omitempty"`
	Mismatched []string `json:"mismatched_fields,omitempty"`
	Caveat     string   `json:"caveat,omitempty"`
}

// VerifyActivityMetadataUpdate checks native upstream fields, preserving false/zero/empty-list semantics.
func VerifyActivityMetadataUpdate(activity Activity, p UpdateActivityParams) *ActivityMetadataVerification {
	v := &ActivityMetadataVerification{Status: "verified", Source: "write_response"}
	check := func(name, key string, want any) {
		actual, present := activity.Raw[key]
		if !present || actual == nil {
			v.Missing = append(v.Missing, name)
			return
		}
		if activityMetadataValueMatches(actual, want) {
			v.Confirmed = append(v.Confirmed, name)
		} else {
			v.Mismatched = append(v.Mismatched, name)
		}
	}
	if p.Tags != nil {
		check("tags", "tags", *p.Tags)
	}
	if p.Commute != nil {
		check("commute", "commute", *p.Commute)
	}
	if p.Feel != nil {
		check("feel", "feel", *p.Feel)
	}
	if p.RPE != nil {
		check("rpe", "icu_rpe", *p.RPE)
	}
	if p.SubType != nil {
		check("sub_type", "sub_type", *p.SubType)
	}
	if p.DistanceM != nil {
		check("distance_m", "distance", *p.DistanceM)
	}
	if len(v.Confirmed)+len(v.Missing)+len(v.Mismatched) == 0 {
		return nil
	}
	slices.Sort(v.Confirmed)
	slices.Sort(v.Missing)
	slices.Sort(v.Mismatched)
	if len(v.Missing) > 0 {
		v.Status = "verification_unavailable"
	}
	if len(v.Mismatched) > 0 {
		v.Status = "verification_mismatch"
	}
	if v.Status != "verified" {
		v.Caveat = "write may have succeeded; inspect the activity before retrying; requested values are not confirmed"
	}
	return v
}

func activityMetadataValueMatches(actual, want any) bool {
	switch desired := want.(type) {
	case bool:
		got, ok := actual.(bool)
		return ok && got == desired
	case int:
		got, ok := actual.(float64)
		return ok && got == float64(desired)
	case string:
		got, ok := actual.(string)
		return ok && got == desired
	case float64:
		got, ok := actual.(float64)
		return ok && float32(got) == float32(desired)
	case []string:
		values, ok := actual.([]any)
		if !ok || len(values) != len(desired) {
			return false
		}
		got := make([]string, len(values))
		for i, value := range values {
			tag, ok := value.(string)
			if !ok {
				return false
			}
			got[i] = tag
		}
		expected := slices.Clone(desired)
		slices.Sort(got)
		slices.Sort(expected)
		return slices.Equal(got, expected)
	default:
		return false
	}
}

// ParseEventID parses an upstream numeric event ID used by activity paired_event_id.
func ParseEventID(value string) (int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("event ID is required")
	}
	parsed, err := strconv.Atoi(trimmed)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("event ID must be a positive integer")
	}
	return parsed, nil
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
