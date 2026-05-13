package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	createWorkoutName                    = "create_workout"
	createWorkoutDescription             = "Create a reusable workout-library template, not a calendar event. Accepts either verbatim free-text description or structured workout_doc steps that icuvisor serializes to the intervals.icu workout-description DSL before upload."
	invalidCreateWorkoutArgumentsMessage = "invalid create_workout arguments; provide name, sport, optional folder_id/tags, and either description or workout_doc"
	createWorkoutMessage                 = "could not create workout; check intervals.icu credentials, athlete ID, folder ID, and writable workout fields"
)

// WorkoutCreatorClient creates workout-library templates for tools.
type WorkoutCreatorClient interface {
	CreateLibraryWorkout(context.Context, intervals.WriteWorkoutParams) (intervals.Workout, error)
}

type createWorkoutRequest struct {
	Name        string                 `json:"name"`
	FolderID    string                 `json:"folder_id,omitempty"`
	Description *string                `json:"description,omitempty"`
	WorkoutDoc  *workoutdoc.WorkoutDoc `json:"workout_doc,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Sport       string                 `json:"sport"`
}

type createWorkoutResponse struct {
	Workout workoutTemplateRow `json:"workout"`
	Meta    createWorkoutMeta  `json:"_meta"`
}

type createWorkoutMeta struct {
	Operation           string   `json:"operation"`
	SourceEndpoint      string   `json:"source_endpoint"`
	FolderID            string   `json:"folder_id,omitempty"`
	Sport               string   `json:"sport"`
	Tags                []string `json:"tags,omitempty"`
	WorkoutDocUploaded  string   `json:"workout_doc_uploaded,omitempty"`
	DefaultPayloadScope string   `json:"default_payload_scope"`
}

func newCreateWorkoutTool(client WorkoutCreatorClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: createWorkoutName, Description: createWorkoutDescription, InputSchema: createWorkoutInputSchema(), OutputSchema: createWorkoutOutputSchema(), Requirement: RequirementWrite, Handler: createWorkoutHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func createWorkoutHandler(client WorkoutCreatorClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeCreateWorkoutRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCreateWorkoutArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(createWorkoutMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(createWorkoutMessage, errors.New("missing workout creator client"))
		}
		params, uploaded, err := createWorkoutParams(args)
		if err != nil {
			return Result{}, NewUserError(invalidCreateWorkoutArgumentsMessage, err)
		}
		workout, err := client.CreateLibraryWorkout(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(createWorkoutMessage, err)
		}
		payload := shapeCreateWorkoutResponse(workout, args, uploaded)
		return encodeShaped(payload, false, nil, version, debugMetadata, createWorkoutName, unitSystem)
	}
}

func decodeCreateWorkoutRequest(raw json.RawMessage) (createWorkoutRequest, error) {
	var args createWorkoutRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.Name = strings.TrimSpace(args.Name)
	args.FolderID = strings.TrimSpace(args.FolderID)
	args.Sport = strings.TrimSpace(args.Sport)
	if args.Name == "" {
		return args, errors.New("name is required")
	}
	if args.Sport == "" {
		return args, errors.New("sport is required")
	}
	if args.Description != nil && args.WorkoutDoc != nil {
		return args, errors.New("provide either free-text description or structured workout_doc, not both")
	}
	return args, nil
}

func createWorkoutParams(args createWorkoutRequest) (intervals.WriteWorkoutParams, string, error) {
	params := intervals.WriteWorkoutParams{Name: args.Name, FolderID: args.FolderID, Description: args.Description, Tags: append([]string(nil), args.Tags...), Sport: args.Sport}
	if args.WorkoutDoc == nil {
		return params, "", nil
	}
	dsl, err := workoutdoc.Serialize(*args.WorkoutDoc)
	if err != nil {
		return intervals.WriteWorkoutParams{}, "", fmt.Errorf("serializing workout_doc: %w", err)
	}
	params.Description = &dsl
	return params, "description_dsl", nil
}

func shapeCreateWorkoutResponse(workout intervals.Workout, args createWorkoutRequest, workoutDocUploaded string) createWorkoutResponse {
	return createWorkoutResponse{Workout: workoutToRow(workout, false), Meta: createWorkoutMeta{Operation: "create", SourceEndpoint: workoutLibraryWorkoutsEndpoint, FolderID: args.FolderID, Sport: args.Sport, Tags: append([]string(nil), args.Tags...), WorkoutDocUploaded: workoutDocUploaded, DefaultPayloadScope: "same terse workout row shape used by get_workout_library/get_workouts_in_folder; raw workout_doc remains summarized"}}
}

func createWorkoutInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"name", "sport"}, "properties": map[string]any{
		"name":        map[string]any{"type": "string", "description": "Required workout-library template name/title. Surrounding whitespace is trimmed."},
		"folder_id":   map[string]any{"type": "string", "description": "Optional intervals.icu workout-library folder or plan ID to place the new template in. Omit to create a top-level library workout."},
		"description": map[string]any{"type": "string", "description": "Optional free-text workout description. Preserved verbatim when workout_doc is omitted; mutually exclusive with workout_doc because intervals.icu accepts one description DSL string on writes."},
		"workout_doc": map[string]any{"type": "object", "description": "Optional structured workout steps using icuvisor's WorkoutDoc shape. Mutually exclusive with description; the server serializes this to the intervals.icu workout-description DSL string and never sends the structured workout_doc object upstream."},
		"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional workout-library tags to preserve on the upstream template, in caller-provided order."},
		"sport":       map[string]any{"type": "string", "description": "Required upstream sport/activity type for the workout template, such as Ride, Run, Swim, VirtualRide, or the athlete account's configured activity type."},
	}}
}

func createWorkoutOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Create confirmation containing the same terse workout row shape used by workout-library read tools plus operation, source endpoint, sport, and workout_doc upload metadata."}
}
