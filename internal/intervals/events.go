package intervals

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

// WriteEventParams contains writable calendar event fields.
type WriteEventParams struct {
	EventID            string
	Date               string
	Category           string
	Type               string
	Name               string
	Description        *string
	Tags               []string
	TargetLoad         *float64
	DistanceMeters     *float64
	MovingTimeSeconds  *int
	ElapsedTimeSeconds *int
}

// Event contains stable calendar event fields used by read tools and preserves raw upstream fields.
type Event struct {
	Raw map[string]any `json:"-"`

	ID                string   `json:"-"`
	Name              *string  `json:"name"`
	Type              *string  `json:"type"`
	Category          *string  `json:"category"`
	StartDateLocal    *string  `json:"start_date_local"`
	EndDateLocal      *string  `json:"end_date_local"`
	Updated           *string  `json:"updated"`
	PlanApplied       *string  `json:"plan_applied"`
	Description       *string  `json:"description"`
	TrainingLoad      *float64 `json:"icu_training_load"`
	LoadTarget        *float64 `json:"load_target"`
	Distance          *float64 `json:"distance"`
	DistanceTarget    *float64 `json:"distance_target"`
	MovingTime        *int     `json:"moving_time"`
	TimeTarget        *int     `json:"time_target"`
	ElapsedTime       *int     `json:"elapsed_time"`
	ElapsedTimeTarget *int     `json:"elapsed_time_target"`
	WorkoutDoc        any      `json:"workout_doc"`
	TrainingPlanID    any      `json:"training_plan_id"`
	CalendarID        any      `json:"calendar_id"`
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

// AddOrUpdateEvent creates a new calendar event or updates an existing event for the configured athlete.
func (c *Client) AddOrUpdateEvent(ctx context.Context, params WriteEventParams) (Event, error) {
	body, err := writeEventBody(params)
	if err != nil {
		return Event{}, err
	}
	if eventID := strings.TrimSpace(params.EventID); eventID != "" {
		var event Event
		if err := c.doJSONBody(ctx, http.MethodPut, body, &event, "athlete", c.athleteID, "events", eventID); err != nil {
			return Event{}, fmt.Errorf("updating event %s: %w", params.EventID, err)
		}
		return event, nil
	}
	var events []Event
	if err := c.doJSONBody(ctx, http.MethodPost, []writeEventPayload{body}, &events, "athlete", c.athleteID, "events", "bulk"); err != nil {
		return Event{}, fmt.Errorf("creating event: %w", err)
	}
	if len(events) == 0 {
		return Event{}, fmt.Errorf("creating event: empty upstream response")
	}
	return events[0], nil
}

type writeEventPayload struct {
	StartDateLocal    string   `json:"start_date_local"`
	Category          string   `json:"category"`
	Type              string   `json:"type,omitempty"`
	Name              string   `json:"name,omitempty"`
	Description       *string  `json:"description,omitempty"`
	Tags              []string `json:"tags,omitempty"`
	LoadTarget        *float64 `json:"load_target,omitempty"`
	DistanceTarget    *float64 `json:"distance_target,omitempty"`
	TimeTarget        *int     `json:"time_target,omitempty"`
	ElapsedTimeTarget *int     `json:"elapsed_time_target,omitempty"`
}

func writeEventBody(params WriteEventParams) (writeEventPayload, error) {
	date := strings.TrimSpace(params.Date)
	if date == "" {
		return writeEventPayload{}, fmt.Errorf("writing event: date is required")
	}
	category := strings.TrimSpace(params.Category)
	if category == "" {
		return writeEventPayload{}, fmt.Errorf("writing event: category is required")
	}
	return writeEventPayload{
		StartDateLocal:    writeEventStartDateLocal(date, category),
		Category:          category,
		Type:              strings.TrimSpace(params.Type),
		Name:              strings.TrimSpace(params.Name),
		Description:       params.Description,
		Tags:              append([]string(nil), params.Tags...),
		LoadTarget:        params.TargetLoad,
		DistanceTarget:    params.DistanceMeters,
		TimeTarget:        params.MovingTimeSeconds,
		ElapsedTimeTarget: params.ElapsedTimeSeconds,
	}, nil
}

func writeEventStartDateLocal(date string, category string) string {
	if strings.EqualFold(strings.TrimSpace(category), "WORKOUT") && len(date) == len("2006-01-02") {
		return date + "T00:00:00"
	}
	return date
}

func (c *Client) doJSONBody(ctx context.Context, method string, body any, out any, pathParts ...string) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding intervals.icu request: %w", err)
	}
	for attempt := 1; ; attempt++ {
		req, err := c.newRequest(ctx, method, pathParts...)
		if err != nil {
			return err
		}
		req.Body = io.NopCloser(bytes.NewReader(payload))
		req.ContentLength = int64(len(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if c.shouldRetryWrite(method, ctx, attempt) {
				if sleepErr := c.sleepBeforeRetry(ctx, attempt, 0); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			return fmt.Errorf("calling intervals.icu: %w", err)
		}

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
			apiErr := errorForStatus(resp.StatusCode, retryAfter)
			_, _ = io.Copy(io.Discard, resp.Body)
			closeErr := resp.Body.Close()
			if c.shouldRetryWriteStatus(method, resp.StatusCode, attempt) {
				if sleepErr := c.sleepBeforeRetry(ctx, attempt, retryAfter); sleepErr != nil {
					return sleepErr
				}
				continue
			}
			if closeErr != nil {
				return fmt.Errorf("closing intervals.icu response: %w", closeErr)
			}
			return fmt.Errorf("calling intervals.icu: %w", apiErr)
		}

		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decoding intervals.icu response: %w", err)
		}
		return nil
	}
}

func (c *Client) shouldRetryWrite(method string, ctx context.Context, attempt int) bool {
	if method == http.MethodPost {
		return false
	}
	return c.shouldRetryTransport(ctx, attempt)
}

func (c *Client) shouldRetryWriteStatus(method string, statusCode int, attempt int) bool {
	if method == http.MethodPost {
		return false
	}
	return c.shouldRetryStatus(statusCode, attempt)
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
