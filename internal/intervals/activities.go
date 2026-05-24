package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ListActivitiesParams contains query parameters for listing athlete activities.
type ListActivitiesParams struct {
	Oldest  string
	Newest  string
	RouteID int64
	Limit   int
	Fields  []string
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
	ActivityID     string
	Name           string
	NameSet        bool
	Description    string
	DescriptionSet bool
}

type updateActivityPayload struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Activity contains stable activity fields used by read tools and preserves raw upstream fields.
type Activity struct {
	Raw map[string]any `json:"-"`

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
	TrainingLoad       *int     `json:"icu_training_load"`
	AverageHeartRate   *int     `json:"average_heartrate"`
	MaxHeartRate       *int     `json:"max_heartrate"`
	AverageCadence     *float64 `json:"average_cadence"`
	Calories           *int     `json:"calories"`
	CarbsIngested      *int     `json:"carbs_ingested"`
	CarbsUsed          *int     `json:"carbs_used"`
	DeviceName         *string  `json:"device_name"`
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
	var decoded activityAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*a = Activity(decoded)
	a.Raw = raw
	a.GearID = rawIDString(raw["gear_id"])
	return nil
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
	if activityID == "" {
		return Activity{}, fmt.Errorf("updating activity: activity ID is required")
	}
	if !params.NameSet && !params.DescriptionSet {
		return Activity{}, fmt.Errorf("updating activity %s: at least one of name or description is required", activityID)
	}
	payload := updateActivityPayload{}
	if params.NameSet {
		name := params.Name
		payload.Name = &name
	}
	if params.DescriptionSet {
		description := params.Description
		payload.Description = &description
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
	return activity, nil
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
