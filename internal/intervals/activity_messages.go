package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ActivityMessagesParams contains activity message query parameters.
type ActivityMessagesParams struct {
	ActivityID string
	SinceID    int64
	Limit      int
}

// AddActivityMessageParams contains the content for a new activity comment/message.
type AddActivityMessageParams struct {
	ActivityID string
	Content    string
}

type addActivityMessagePayload struct {
	Content string `json:"content"`
}

// NewActivityMessage contains the add-message response and preserves raw upstream fields.
type NewActivityMessage struct {
	Raw     map[string]any `json:"-"`
	ID      int64          `json:"id"`
	NewChat any            `json:"new_chat"`
}

// UnmarshalJSON decodes NewActivityMessage while retaining the original object for full responses.
func (m *NewActivityMessage) UnmarshalJSON(data []byte) error {
	type messageAlias NewActivityMessage
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded messageAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = NewActivityMessage(decoded)
	m.Raw = raw
	return nil
}

// ActivityMessage contains a comment/message on an activity and preserves raw fields.
type ActivityMessage struct {
	Raw map[string]any `json:"-"`

	ID         int64  `json:"id"`
	AthleteID  string `json:"athlete_id"`
	Name       string `json:"name"`
	Created    string `json:"created"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	ActivityID string `json:"activity_id"`
	Deleted    *bool  `json:"-"`
	DeletedAt  string `json:"deleted_at,omitempty"`
	Seen       *bool  `json:"seen"`
}

// UnmarshalJSON decodes ActivityMessage while retaining the original object for full responses.
func (m *ActivityMessage) UnmarshalJSON(data []byte) error {
	type messageAlias ActivityMessage
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	var decoded messageAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if value := raw["deleted"]; value != nil {
		switch value := value.(type) {
		case bool:
			decoded.Deleted = &value
		case string:
			if _, err := time.Parse(time.RFC3339, value); err != nil {
				return fmt.Errorf("decoding message deletion timestamp: %w", err)
			}
			deleted := true
			decoded.Deleted = &deleted
			decoded.DeletedAt = value
		default:
			return fmt.Errorf("decoding message deletion: expected timestamp or boolean")
		}
	}
	*m = ActivityMessage(decoded)
	m.Raw = raw
	return nil
}

// AddActivityMessage appends a message/comment to one activity.
func (c *Client) AddActivityMessage(ctx context.Context, params AddActivityMessageParams) (NewActivityMessage, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if activityID == "" {
		return NewActivityMessage{}, fmt.Errorf("adding activity message: activity ID is required")
	}
	content := strings.TrimSpace(params.Content)
	if content == "" {
		return NewActivityMessage{}, fmt.Errorf("adding activity %s message: content is required", activityID)
	}
	if err := c.ensureActivityIDTarget(ctx, activityID); err != nil {
		return NewActivityMessage{}, fmt.Errorf("adding activity %s message: %w", activityID, err)
	}
	var message NewActivityMessage
	if err := c.doJSONBody(ctx, http.MethodPost, addActivityMessagePayload{Content: params.Content}, &message, "activity", activityID, "messages"); err != nil {
		return NewActivityMessage{}, fmt.Errorf("adding activity %s message: %w", activityID, err)
	}
	return message, nil
}

// GetActivityMessages retrieves messages for one activity.
func (c *Client) GetActivityMessages(ctx context.Context, params ActivityMessagesParams) ([]ActivityMessage, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if activityID == "" {
		return nil, fmt.Errorf("getting activity messages: activity ID is required")
	}
	query := url.Values{}
	if params.SinceID > 0 {
		query.Set("sinceId", strconv.FormatInt(params.SinceID, 10))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if err := c.ensureActivityIDTarget(ctx, activityID); err != nil {
		return nil, fmt.Errorf("getting activity %s messages: %w", activityID, err)
	}
	var messages []ActivityMessage
	if err := c.doJSONQuery(ctx, &messages, query, "activity", activityID, "messages"); err != nil {
		return nil, fmt.Errorf("getting activity %s messages: %w", activityID, err)
	}
	return messages, nil
}
