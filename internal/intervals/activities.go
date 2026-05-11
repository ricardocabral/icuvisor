package intervals

import (
	"context"
	"encoding/json"
	"fmt"
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
	DeviceName         *string  `json:"device_name"`
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
