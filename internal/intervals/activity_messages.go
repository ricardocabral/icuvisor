package intervals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ActivityMessagesParams contains activity message query parameters.
type ActivityMessagesParams struct {
	ActivityID string
	SinceID    int64
	Limit      int
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
	Deleted    *bool  `json:"deleted"`
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
	*m = ActivityMessage(decoded)
	m.Raw = raw
	return nil
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
	var messages []ActivityMessage
	if err := c.doJSONQuery(ctx, &messages, query, "activity", activityID, "messages"); err != nil {
		return nil, fmt.Errorf("getting activity %s messages: %w", activityID, err)
	}
	return messages, nil
}
