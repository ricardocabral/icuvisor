package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

const (
	deleteWorkoutName                    = "delete_workout"
	deleteWorkoutDescription             = "Delete one reusable workout-library template by workout_id. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteWorkoutArgumentsMessage = "invalid delete_workout arguments; provide workout_id only"
	deleteWorkoutMessage                 = "could not delete workout; check intervals.icu credentials, athlete ID, workout ID, and delete-mode configuration"
)

// WorkoutDeleterClient deletes workout-library templates for tools.
type WorkoutDeleterClient interface {
	DeleteLibraryWorkout(context.Context, string) error
}

type deleteWorkoutRequest struct {
	WorkoutID string `json:"workout_id"`
}

type deleteWorkoutResponse struct {
	Deleted deleteWorkoutDeleted `json:"deleted"`
	Meta    deleteWorkoutMeta    `json:"_meta"`
}

type deleteWorkoutDeleted struct {
	WorkoutID string `json:"workout_id"`
	Status    string `json:"status"`
}

type deleteWorkoutMeta struct {
	Operation      string `json:"operation"`
	SourceEndpoint string `json:"source_endpoint"`
}

func newDeleteWorkoutTool(client WorkoutDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: deleteWorkoutName, Description: deleteWorkoutDescription, InputSchema: deleteWorkoutInputSchema(), OutputSchema: deleteWorkoutOutputSchema(), Requirement: RequirementDelete, Handler: deleteWorkoutHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func deleteWorkoutHandler(client WorkoutDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteWorkoutRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteWorkoutArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(deleteWorkoutMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(deleteWorkoutMessage, errors.New("missing workout deleter client"))
		}
		if err := client.DeleteLibraryWorkout(ctx, args.WorkoutID); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteWorkoutMessage, err)
		}
		payload := deleteWorkoutResponse{Deleted: deleteWorkoutDeleted{WorkoutID: args.WorkoutID, Status: "deleted"}, Meta: deleteWorkoutMeta{Operation: "delete", SourceEndpoint: workoutLibraryWorkoutsEndpoint}}
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteWorkoutName, unitSystem)
	}
}

func decodeDeleteWorkoutRequest(raw json.RawMessage) (deleteWorkoutRequest, error) {
	var args deleteWorkoutRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.WorkoutID = strings.TrimSpace(args.WorkoutID)
	if args.WorkoutID == "" {
		return args, errors.New("workout_id is required")
	}
	return args, nil
}

func deleteWorkoutInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"workout_id"}, "properties": map[string]any{
		"workout_id": map[string]any{"type": "string", "description": "Required upstream workout-library template ID to delete. This destructive operation has no confirm argument; the tool is registered only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func deleteWorkoutOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete confirmation containing workout_id, status, source endpoint, and common server/delete-mode metadata."}
}
