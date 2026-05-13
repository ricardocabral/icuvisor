package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	addActivityMessageName                    = "add_activity_message"
	addActivityMessageDescription             = "Append a non-destructive comment/message to one activity when write tools are enabled. This tool only adds a new message and never overwrites prior messages."
	invalidAddActivityMessageArgumentsMessage = "invalid add_activity_message arguments; provide activity_id and a non-empty message"
	addActivityMessageMessage                 = "could not add activity message; check intervals.icu credentials and activity ID"
)

// ActivityMessageWriterClient appends activity messages.
type ActivityMessageWriterClient interface {
	AddActivityMessage(context.Context, intervals.AddActivityMessageParams) (intervals.NewActivityMessage, error)
}

type addActivityMessageRequest struct {
	ActivityID  string `json:"activity_id"`
	Message     string `json:"message"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type addActivityMessageResponse struct {
	ActivityID string                 `json:"activity_id"`
	MessageID  int64                  `json:"message_id,omitempty"`
	Status     string                 `json:"status"`
	Full       map[string]any         `json:"full,omitempty"`
	Meta       addActivityMessageMeta `json:"_meta"`
}

type addActivityMessageMeta struct {
	AthleteID   string `json:"athlete_id,omitempty"`
	AppendOnly  bool   `json:"append_only"`
	IncludeFull bool   `json:"include_full"`
}

func newAddActivityMessageTool(client ActivityMessageWriterClient, profileClient ProfileClient, version string, debugMetadata bool) Tool {
	return Tool{Name: addActivityMessageName, Description: addActivityMessageDescription, InputSchema: addActivityMessageInputSchema(), OutputSchema: addActivityMessageOutputSchema(), Requirement: RequirementWrite, Handler: addActivityMessageHandler(client, profileClient, version, debugMetadata)}
}

func addActivityMessageHandler(client ActivityMessageWriterClient, profileClient ProfileClient, version string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAddActivityMessageRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAddActivityMessageArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(addActivityMessageMessage, errors.New("missing activity message writer client"))
		}
		message, err := client.AddActivityMessage(ctx, intervals.AddActivityMessageParams{ActivityID: args.ActivityID, Content: args.Message})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(addActivityMessageMessage, err)
		}
		athleteID := ""
		if profileClient != nil {
			profile, profileErr := profileClient.GetAthleteProfile(ctx)
			if profileErr != nil {
				if errors.Is(profileErr, context.Canceled) || errors.Is(profileErr, context.DeadlineExceeded) {
					return Result{}, profileErr
				}
			} else {
				athleteID = config.NormalizeAthleteIDForDisplay(profile.ID)
			}
		}
		payload := addActivityMessageResponse{ActivityID: args.ActivityID, MessageID: message.ID, Status: "appended", Meta: addActivityMessageMeta{AthleteID: athleteID, AppendOnly: true, IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = message.Raw
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addActivityMessageName, "")
	}
}

func decodeAddActivityMessageRequest(raw json.RawMessage) (addActivityMessageRequest, error) {
	var args addActivityMessageRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	if args.ActivityID == "" {
		return args, errors.New("activity_id is required")
	}
	if strings.TrimSpace(args.Message) == "" {
		return args, errors.New("message is required")
	}
	return args, nil
}

func addActivityMessageInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id", "message"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu activity ID to append a message/comment to. Surrounding whitespace is trimmed."},
		"message":      map[string]any{"type": "string", "description": "Required free-text message/comment content. Must be non-empty after trimming; the body is sent as content and is not logged by icuvisor."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream add-message response under full; default returns a terse append confirmation."},
	}}
}

func addActivityMessageOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Append-only activity message confirmation with activity_id, message_id, status, and normalized athlete_id metadata when available."}
}
