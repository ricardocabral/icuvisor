package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListEventsParams contains query parameters for listing athlete calendar events.
type ListEventsParams struct {
	Oldest     string
	Newest     string
	Category   string
	CalendarID string
	Limit      int
	Resolve    *bool
}

// Event contains stable calendar event fields used by read tools and preserves raw upstream fields.
type Event struct {
	Raw map[string]any `json:"-"`

	ID             string   `json:"-"`
	Name           *string  `json:"name"`
	Type           *string  `json:"type"`
	Category       *string  `json:"category"`
	StartDateLocal *string  `json:"start_date_local"`
	EndDateLocal   *string  `json:"end_date_local"`
	Updated        *string  `json:"updated"`
	PlanApplied    *string  `json:"plan_applied"`
	Description    *string  `json:"description"`
	TrainingLoad   *float64 `json:"icu_training_load"`
	Distance       *float64 `json:"distance"`
	MovingTime     *int     `json:"moving_time"`
	ElapsedTime    *int     `json:"elapsed_time"`
	WorkoutDoc     any      `json:"workout_doc"`
	TrainingPlanID any      `json:"training_plan_id"`
	CalendarID     any      `json:"calendar_id"`
}

// UnmarshalJSON decodes Event while retaining the original object for full responses.
func (e *Event) UnmarshalJSON(data []byte) error {
	type eventAlias Event
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded eventAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*e = Event(decoded)
	e.Raw = raw
	e.ID = rawIDString(raw["id"])
	return nil
}

// ListEvents lists calendar events for the configured athlete and local date range.
func (c *Client) ListEvents(ctx context.Context, params ListEventsParams) ([]Event, error) {
	query := url.Values{}
	oldest := strings.TrimSpace(params.Oldest)
	newest := strings.TrimSpace(params.Newest)
	if oldest == "" {
		return nil, fmt.Errorf("listing events: oldest is required")
	}
	if newest == "" {
		return nil, fmt.Errorf("listing events: newest is required")
	}
	query.Set("oldest", oldest)
	query.Set("newest", newest)
	if category := strings.TrimSpace(params.Category); category != "" {
		query.Set("category", category)
	}
	if calendarID := strings.TrimSpace(params.CalendarID); calendarID != "" {
		query.Set("calendar_id", calendarID)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Resolve != nil {
		query.Set("resolve", strconv.FormatBool(*params.Resolve))
	}

	var events []Event
	if err := c.doJSONQuery(ctx, &events, query, "athlete", c.athleteID, "events"); err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}
	return events, nil
}

// GetEvent retrieves one calendar event for the configured athlete.
func (c *Client) GetEvent(ctx context.Context, eventID string) (Event, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return Event{}, fmt.Errorf("getting event: event ID is required")
	}
	var event Event
	if err := c.doJSON(ctx, &event, "athlete", c.athleteID, "events", eventID); err != nil {
		return Event{}, fmt.Errorf("getting event %s: %w", eventID, err)
	}
	return event, nil
}

func rawIDString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}
