package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	deleteEventName                    = "delete_event"
	deleteEventDescription             = "Delete one calendar event by event_id. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteEventArgumentsMessage = "invalid delete_event arguments; provide event_id only"
	deleteEventMessage                 = "could not delete event; check intervals.icu credentials, athlete ID, event ID, and delete-mode configuration"
	deleteEventEndpoint                = "/athlete/{id}/events/{eventId}"
)

// EventDeleterClient retrieves and deletes calendar events for tools.
type EventDeleterClient interface {
	GetEvent(context.Context, string) (intervals.Event, error)
	DeleteEvent(context.Context, string) error
}

type deleteEventRequest struct {
	EventID string `json:"event_id"`
}

func newDeleteEventTool(client EventDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: deleteEventName, Description: deleteEventDescription, InputSchema: deleteEventInputSchema(), OutputSchema: deleteEventOutputSchema(), Requirement: RequirementDelete, Handler: deleteEventHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func deleteEventHandler(client EventDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteEventRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteEventArgumentsMessage, err)
		}
		profile, unitSystem, timezoneName, err := toolProfileDetails(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(deleteEventMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(deleteEventMessage, errors.New("missing event deleter client"))
		}
		event, err := client.GetEvent(ctx, args.EventID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteEventMessage, err)
		}
		before, err := eventDeleteEcho(event, args.EventID, timezoneName, workoutPreviewContextForEvent(event, profile, unitSystem))
		if err != nil {
			return Result{}, fmt.Errorf("shaping delete_event before echo: %w", err)
		}
		if err := client.DeleteEvent(ctx, args.EventID); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteEventMessage, err)
		}
		payload := newDeleteResourceResponse(args.EventID, "event", deleteEventEndpoint, before)
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteEventName, unitSystem, shapeCfg)
	}
}

func decodeDeleteEventRequest(raw json.RawMessage) (deleteEventRequest, error) {
	var args deleteEventRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[deleteEventRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.EventID = strings.TrimSpace(args.EventID)
	if args.EventID == "" {
		return args, errors.New("event_id is required")
	}
	return args, nil
}

func deleteEventInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"event_id"}, "properties": map[string]any{
		"event_id": map[string]any{"type": "string", "description": "Required opaque upstream event ID to delete. This destructive operation has no confirm argument; the tool is registered only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func deleteEventOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete result with deleted_id, status, _meta.confirmation_status, and _meta.deleted containing a terse echo of the event before deletion. A successful upstream delete call is reported as upstream_delete_succeeded_post_delete_unverified because the tool does not perform a post-delete read."}
}
