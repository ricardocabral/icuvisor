package tools

import (
	"context"
	"errors"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"strings"
)

const (
	updateActivityMessageName  = "update_activity_message"
	changeActivityMessageError = "could not change activity comment; check the activity and message IDs, ownership, credentials, and write/delete mode"
)

// ActivityMessageEditorClient changes only scoped, authenticated-athlete-owned comments.
type ActivityMessageEditorClient interface {
	UpdateActivityMessage(context.Context, intervals.ChangeActivityMessageParams) (intervals.ActivityMessageChange, error)
	DeleteActivityMessage(context.Context, intervals.ChangeActivityMessageParams) (intervals.ActivityMessageChange, error)
}

type changeActivityMessageRequest struct {
	ActivityID  string  `json:"activity_id"`
	MessageID   int64   `json:"message_id"`
	Message     *string `json:"message,omitempty"`
	IncludeFull bool    `json:"include_full,omitempty"`
}

type changeActivityMessageResponse struct {
	ActivityID string             `json:"activity_id"`
	MessageID  int64              `json:"message_id"`
	Status     string             `json:"status"`
	Before     activityMessageRow `json:"before"`
}

func newUpdateActivityMessageTool(client ActivityMessageEditorClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	return coreTool(Tool{Name: updateActivityMessageName, Description: "Replace the text of your own comment on one activity using activity_id and message_id. Returns the prior text for recovery; upstream permissions still apply. Read the existing comment first and preserve user edits. Omitted fields are never passed through to the chat API.", InputSchema: changeActivityMessageInputSchema(false), OutputSchema: changeActivityMessageOutputSchema(), Requirement: RequirementWrite, Handler: changeActivityMessageHandler(client, false, version, debugMetadata, responseShapingOrDefault(shaping))})
}

func changeActivityMessageHandler(client ActivityMessageEditorClient, remove bool, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := DecodeStrict[changeActivityMessageRequest](req.Arguments)
		if err != nil {
			return Result{}, NewUserError("invalid activity comment arguments; provide activity_id, positive message_id, and non-empty message for edits", err)
		}
		args.ActivityID = strings.TrimSpace(args.ActivityID)
		if args.ActivityID == "" || args.MessageID <= 0 || (!remove && (args.Message == nil || strings.TrimSpace(*args.Message) == "")) || (remove && args.Message != nil) {
			return Result{}, NewUserError("invalid activity comment arguments; provide activity_id, positive message_id, and non-empty message for edits", nil)
		}
		if client == nil {
			return Result{}, NewUserError(changeActivityMessageError, errors.New("missing activity message editor"))
		}
		params := intervals.ChangeActivityMessageParams{ActivityID: args.ActivityID, MessageID: args.MessageID}
		if args.Message != nil {
			params.Content = *args.Message
		}
		var changed intervals.ActivityMessageChange
		if remove {
			changed, err = client.DeleteActivityMessage(ctx, params)
		} else {
			changed, err = client.UpdateActivityMessage(ctx, params)
		}
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(changeActivityMessageError, err)
		}
		before := activityMessageRow{ID: changed.Before.ID, Content: changed.Before.Content, Created: changed.Before.Created, Deleted: changed.Before.Deleted, DeletedAt: changed.Before.DeletedAt}
		if args.IncludeFull {
			before.Full = changed.Before.Raw
		}
		name := updateActivityMessageName
		if remove {
			name = deleteActivityMessageName
		}
		return encodeShaped(changeActivityMessageResponse{ActivityID: args.ActivityID, MessageID: args.MessageID, Status: changed.Status, Before: before}, args.IncludeFull, nil, version, debugMetadata, name, "", shapeCfg)
	}
}

func changeActivityMessageInputSchema(remove bool) map[string]any {
	props := map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required activity ID; membership and selected athlete ownership are checked before changing a comment."},
		"message_id":   map[string]any{"type": "integer", "minimum": 1, "description": "Existing comment ID from get_activity_messages. The credential owner must be the comment author; chat IDs are resolved by the server."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "Include the original upstream comment under before.full for recovery; default returns prior text and identifiers only."},
	}
	required := []string{"activity_id", "message_id"}
	if !remove {
		props["message"] = map[string]any{"type": "string", "description": "Replacement comment text, non-empty after trimming. Read and preserve current user edits before replacement. The previous content is returned under before."}
		required = append(required, "message")
	}
	return map[string]any{"type": "object", "additionalProperties": false, "required": required, "properties": props}
}
func changeActivityMessageOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Scoped own-comment change result: activity_id, message_id, status (updated/deleted/already_deleted), and before with prior text for manual recovery. No other athlete's comment content is returned on failure."}
}
