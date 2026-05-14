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
	deleteActivityName                    = "delete_activity"
	deleteActivityDescription             = "Delete one activity by activity_id. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteActivityArgumentsMessage = "invalid delete_activity arguments; provide activity_id only"
	deleteActivityMessage                 = "could not delete activity; check intervals.icu credentials, activity ID, and delete-mode configuration"
	deleteActivityEndpoint                = "/activity/{activityId}"
)

// ActivityDeleterClient retrieves and deletes activities for tools.
type ActivityDeleterClient interface {
	GetActivity(context.Context, string) (intervals.Activity, error)
	DeleteActivity(context.Context, string) error
}

type deleteActivityRequest struct {
	ActivityID string `json:"activity_id"`
}

func newDeleteActivityTool(client ActivityDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: deleteActivityName, Description: deleteActivityDescription, InputSchema: deleteActivityInputSchema(), OutputSchema: deleteActivityOutputSchema(), Requirement: RequirementDelete, Handler: deleteActivityHandler(client, profileClient, version, timezoneFallback, debugMetadata)})
}

func deleteActivityHandler(client ActivityDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteActivityRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteActivityArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(deleteActivityMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(deleteActivityMessage, errors.New("missing activity deleter client"))
		}
		activity, err := client.GetActivity(ctx, args.ActivityID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteActivityMessage, err)
		}
		before, err := activityDeleteEcho(activity, args.ActivityID, timezoneName, unitSystem)
		if err != nil {
			return Result{}, fmt.Errorf("shaping delete_activity before echo: %w", err)
		}
		if err := client.DeleteActivity(ctx, args.ActivityID); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteActivityMessage, err)
		}
		payload := newDeleteResourceResponse(args.ActivityID, "activity", deleteActivityEndpoint, before)
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteActivityName, unitSystem)
	}
}

func decodeDeleteActivityRequest(raw json.RawMessage) (deleteActivityRequest, error) {
	var args deleteActivityRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	if args.ActivityID == "" {
		return args, errors.New("activity_id is required")
	}
	return args, nil
}

func deleteActivityInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id": map[string]any{"type": "string", "description": "Required opaque upstream activity ID to delete. This destructive operation has no confirm argument; the tool is registered only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func deleteActivityOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete confirmation with deleted_id, status, and _meta.deleted containing a terse echo of the activity before deletion."}
}
