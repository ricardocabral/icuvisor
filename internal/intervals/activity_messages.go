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

	"github.com/ricardocabral/icuvisor/internal/config"
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

// ChangeActivityMessageParams identifies one activity comment and its replacement text.
// DeleteActivityMessage ignores Content.
type ChangeActivityMessageParams struct {
	ActivityID string
	MessageID  int64
	Content    string
}

// ActivityMessageChange carries a recovery copy only after a successful scoped change.
type ActivityMessageChange struct {
	Timezone string
	Before   ActivityMessage
	Status   string
}

// GetAuthenticatedAthleteProfile retrieves the credential owner's identity, never the coach target.
func (c *Client) GetAuthenticatedAthleteProfile(ctx context.Context) (AthleteWithSportSettings, error) {
	principalContext := context.WithValue(ctx, targetAthleteIDContextKey{}, "")
	var profile AthleteWithSportSettings
	if err := c.doJSON(principalContext, &profile, "athlete", "0"); err != nil {
		return AthleteWithSportSettings{}, fmt.Errorf("getting authenticated athlete profile: %w", err)
	}
	return profile, nil
}

// UpdateActivityMessage replaces the credential owner's comment after checking activity membership.
func (c *Client) UpdateActivityMessage(ctx context.Context, params ChangeActivityMessageParams) (ActivityMessageChange, error) {
	if strings.TrimSpace(params.Content) == "" {
		return ActivityMessageChange{}, fmt.Errorf("updating activity message: content is required")
	}
	return c.changeActivityMessage(ctx, params, false)
}

// DeleteActivityMessage deletes the credential owner's comment after checking activity membership.
func (c *Client) DeleteActivityMessage(ctx context.Context, params ChangeActivityMessageParams) (ActivityMessageChange, error) {
	return c.changeActivityMessage(ctx, params, true)
}

func (c *Client) changeActivityMessage(ctx context.Context, params ChangeActivityMessageParams, remove bool) (ActivityMessageChange, error) {
	activityID := strings.TrimSpace(params.ActivityID)
	if activityID == "" || params.MessageID <= 0 {
		return ActivityMessageChange{}, fmt.Errorf("changing activity message: activity ID and positive message ID are required")
	}
	activity, err := c.GetActivity(ctx, activityID)
	if err != nil {
		return ActivityMessageChange{}, fmt.Errorf("checking activity message target: %w", err)
	}
	profile, err := c.GetAthleteProfile(ctx)
	if err != nil {
		return ActivityMessageChange{}, fmt.Errorf("checking activity message athlete: %w", err)
	}
	target, err := config.NormalizeAthleteID(profile.ID)
	if err != nil || activity.ICUAthleteID == nil {
		return ActivityMessageChange{}, ErrTargetAthleteMismatch
	}
	owner, err := config.NormalizeAthleteID(*activity.ICUAthleteID)
	if err != nil || owner != target {
		return ActivityMessageChange{}, ErrTargetAthleteMismatch
	}
	if activity.ChatID == nil || *activity.ChatID <= 0 {
		return ActivityMessageChange{}, fmt.Errorf("activity has no comment chat")
	}
	messages, err := c.GetActivityMessages(ctx, ActivityMessagesParams{ActivityID: activityID, SinceID: params.MessageID - 1, Limit: 200})
	if err != nil {
		return ActivityMessageChange{}, fmt.Errorf("checking activity comment membership: %w", err)
	}
	var message *ActivityMessage
	for i := range messages {
		if messages[i].ID == params.MessageID {
			message = &messages[i]
			break
		}
	}
	if message == nil || (message.ActivityID != "" && message.ActivityID != activityID) {
		return ActivityMessageChange{}, fmt.Errorf("message not found on the requested activity")
	}
	principal, err := c.GetAuthenticatedAthleteProfile(ctx)
	if err != nil {
		return ActivityMessageChange{}, err
	}
	author, err := config.NormalizeAthleteID(message.AthleteID)
	if err != nil {
		return ActivityMessageChange{}, fmt.Errorf("comment author unavailable")
	}
	principalID, err := config.NormalizeAthleteID(principal.ID)
	if err != nil || principalID != author {
		return ActivityMessageChange{}, fmt.Errorf("only the authenticated athlete's own comments can be changed")
	}
	if message.Deleted != nil && *message.Deleted {
		if remove {
			return ActivityMessageChange{Before: *message, Status: "already_deleted", Timezone: profile.Timezone}, nil
		}
		return ActivityMessageChange{}, fmt.Errorf("deleted comments cannot be edited")
	}
	method, status := http.MethodPut, "updated"
	var body any = addActivityMessagePayload{Content: params.Content}
	if remove {
		method, status, body = http.MethodDelete, "deleted", nil
	}
	var receipt json.RawMessage
	path := []string{"chats", strconv.Itoa(*activity.ChatID), "messages", strconv.FormatInt(params.MessageID, 10)}
	if remove {
		err = c.doJSONNoBody(ctx, method, &receipt, path...)
	} else {
		err = c.doJSONBody(ctx, method, body, &receipt, path...)
	}
	if err != nil {
		return ActivityMessageChange{}, fmt.Errorf("changing activity comment: %w", err)
	}
	return ActivityMessageChange{Before: *message, Status: status, Timezone: profile.Timezone}, nil
}
