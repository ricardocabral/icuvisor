package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getActivityMessagesName        = "get_activity_messages"
	getActivityMessagesDescription = "List comments and notes on one activity by activity_id. Responses are terse by default and render created timestamps in the athlete's configured timezone."
	defaultActivityMessagesLimit   = 100
	maxActivityMessagesLimit       = 200
)

// ActivityMessagesClient retrieves activity messages.
type ActivityMessagesClient interface {
	GetActivityMessages(context.Context, intervals.ActivityMessagesParams) ([]intervals.ActivityMessage, error)
}

type getActivityMessagesRequest struct {
	ActivityID  string `json:"activity_id"`
	SinceID     int64  `json:"since_id,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getActivityMessagesResponse struct {
	ActivityID     string               `json:"activity_id"`
	StravaImported bool                 `json:"strava_imported,omitempty"`
	Unavailable    *unavailableReason   `json:"unavailable,omitempty"`
	Messages       []activityMessageRow `json:"messages,omitempty"`
	Full           map[string]any       `json:"full,omitempty"`
	Meta           activityReadMeta     `json:"_meta"`
}

type activityMessageRow struct {
	ID        int64          `json:"id"`
	AthleteID string         `json:"athlete_id,omitempty"`
	Name      string         `json:"name,omitempty"`
	Created   string         `json:"created,omitempty"`
	Type      string         `json:"type,omitempty"`
	Content   string         `json:"content,omitempty"`
	Deleted   *bool          `json:"deleted,omitempty"`
	Seen      *bool          `json:"seen,omitempty"`
	Full      map[string]any `json:"full,omitempty"`
}

func newGetActivityMessagesTool(client ActivityMessagesClient, profileClient ProfileClient, detailsClient ActivityDetailsClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return coreTool(Tool{Name: getActivityMessagesName, Description: getActivityMessagesDescription, InputSchema: activityMessagesInputSchema(), OutputSchema: activityReadOutputSchema(), Handler: getActivityMessagesHandler(client, profileClient, detailsClient, version, timezoneFallback, debugMetadata)})
}

func getActivityMessagesHandler(client ActivityMessagesClient, profileClient ProfileClient, detailsClient ActivityDetailsClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args getActivityMessagesRequest
		if err := decodeJSONArgs(req.Arguments, &args); err != nil || strings.TrimSpace(args.ActivityID) == "" {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		if args.SinceID < 0 {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, errors.New("since_id must be non-negative"))
		}
		limit, err := normalizeActivityMessagesLimit(args.Limit)
		if err != nil {
			return Result{}, NewUserError(invalidActivityReadArgumentsMessage, err)
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return Result{}, ctxErr
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAthleteProfileMessage, err)
		}
		messages, err := client.GetActivityMessages(ctx, intervals.ActivityMessagesParams{ActivityID: args.ActivityID, SinceID: args.SinceID, Limit: limit})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			if isActivityReadFallbackCandidate(err) && detailsClient != nil {
				activity, activityErr := detailsClient.GetActivity(ctx, args.ActivityID)
				if ctxErr := ctx.Err(); ctxErr != nil {
					return Result{}, ctxErr
				}
				if errors.Is(activityErr, context.Canceled) || errors.Is(activityErr, context.DeadlineExceeded) {
					return Result{}, activityErr
				}
				if activityErr == nil && isStravaBlocked(activity) {
					payload := stravaUnavailableMessagesResponse(args.ActivityID, args.IncludeFull, version, limit, args.SinceID, activity.Raw)
					return encodeActivityMessagesResponse(payload, args.IncludeFull, version, debugMetadata)
				}
			}
			return Result{}, NewUserError(fetchActivityDetailsMessage, err)
		}
		payload := shapeActivityMessages(args.ActivityID, messages, profileTimezone(profile.Timezone, timezoneFallback), args.IncludeFull, version, limit, args.SinceID)
		return encodeActivityMessagesResponse(payload, args.IncludeFull, version, debugMetadata)
	}
}

func shapeActivityMessages(activityID string, messages []intervals.ActivityMessage, timezone string, includeFull bool, version string, limit int, sinceID int64) getActivityMessagesResponse {
	out := getActivityMessagesResponse{ActivityID: activityID, Messages: make([]activityMessageRow, 0, len(messages)), Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, Limit: limit, SinceID: sinceID}}
	for _, message := range messages {
		row := activityMessageRow{ID: message.ID, AthleteID: message.AthleteID, Name: message.Name, Created: renderActivityMessageTime(message.Created, timezone), Type: message.Type, Content: message.Content, Deleted: message.Deleted, Seen: message.Seen}
		if includeFull {
			row.Full = message.Raw
		}
		out.Messages = append(out.Messages, row)
	}
	return out
}

func stravaUnavailableMessagesResponse(activityID string, includeFull bool, version string, limit int, sinceID int64, raw map[string]any) getActivityMessagesResponse {
	out := getActivityMessagesResponse{ActivityID: activityID, StravaImported: true, Unavailable: &unavailableReason{Reason: "strava_tos", Workaround: stravaWorkaround}, Meta: activityReadMeta{ServerVersion: normalizeVersion(version), IncludeFull: includeFull, Limit: limit, SinceID: sinceID}}
	if includeFull {
		out.Full = raw
	}
	return out
}

func encodeActivityMessagesResponse(payload getActivityMessagesResponse, includeFull bool, version string, debugMetadata bool) (Result, error) {
	shaped, err := response.Shape(payload, response.Options{IncludeFull: includeFull, RowCollections: []string{"messages"}, ServerVersion: version, DebugMetadata: debugMetadata, QueryType: getActivityMessagesName})
	if err != nil {
		return Result{}, err
	}
	text, _ := json.Marshal(shaped)
	return Result{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: shaped}, nil
}

func normalizeActivityMessagesLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultActivityMessagesLimit, nil
	}
	if limit < 0 || limit > maxActivityMessagesLimit {
		return 0, errors.New("limit out of range")
	}
	return limit, nil
}

func renderActivityMessageTime(value string, timezone string) string {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return value
	}
	rendered, err := response.RenderTimeInTimezone(parsed, timezone)
	if err != nil {
		return value
	}
	return rendered
}

func activityMessagesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{"activity_id": map[string]any{"type": "string", "description": "intervals.icu activity ID whose messages/comments should be listed."}, "since_id": map[string]any{"type": "integer", "minimum": 0, "description": "Return messages after this upstream message ID; zero means no cursor."}, "limit": map[string]any{"type": "integer", "default": defaultActivityMessagesLimit, "minimum": 1, "maximum": maxActivityMessagesLimit, "description": "Maximum messages to return; defaults to 100 and is capped at 200."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "Include raw upstream message fields and preserve nulls."}}}
}
