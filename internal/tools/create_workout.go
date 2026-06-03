package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	createWorkoutName                    = "create_workout"
	createWorkoutDescription             = "Create a reusable workout-library template, not a calendar event. Before creating, present a human-readable preview for user approval that covers total duration, key steps, target intensities, planned load/distance/time, and what prose/tags/folder are preserved or newly set. Creation builds the initial upstream template description/DSL from `workout_doc` (structured steps), `description` (free-text prose), or both; `description` is part of that single upstream description/DSL field, not a separate append-only notes channel. When both are supplied, icuvisor merges them and the `<!-- icuvisor:steps -->` sentinel controls serialized-step placement. Structured zone targets use the athlete's sport settings so ambiguous zones are emitted with metric suffixes when needed, such as `Z2 Power`, `Z2 HR`, or `Z2 Pace`. Prefer `workout_doc` when the structure is known, and call `validate_workout` first if uncertain about the DSL syntax (see icuvisor://workout-syntax)."
	invalidCreateWorkoutArgumentsMessage = "invalid create_workout arguments; provide name, sport, folder_id, optional tags, and optional description and/or workout_doc"
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
	WorkoutDocWarning   string   `json:"workout_doc_warning,omitempty"`
	DefaultPayloadScope string   `json:"default_payload_scope"`
}

func newCreateWorkoutTool(client WorkoutCreatorClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: createWorkoutName, Description: createWorkoutDescription, InputSchema: createWorkoutInputSchema(), OutputSchema: createWorkoutOutputSchema(), Requirement: RequirementWrite, Handler: createWorkoutHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func createWorkoutHandler(client WorkoutCreatorClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeCreateWorkoutRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCreateWorkoutArgumentsMessage, err)
		}
		profile, unitSystem, _, err := toolProfileDetails(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(createWorkoutMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(createWorkoutMessage, errors.New("missing workout creator client"))
		}
		params, uploaded, err := createWorkoutParams(args, workoutDocSerializeOptionsForSport(profile, args.Sport))
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
		payload := shapeCreateWorkoutResponse(workout, args, uploaded, profile, unitSystem)
		return encodeShaped(payload, false, nil, version, debugMetadata, createWorkoutName, unitSystem, shapeCfg)
	}
}

func decodeCreateWorkoutRequest(raw json.RawMessage) (createWorkoutRequest, error) {
	var args createWorkoutRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[createWorkoutRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Name = strings.TrimSpace(args.Name)
	args.FolderID = strings.TrimSpace(args.FolderID)
	args.Sport = strings.TrimSpace(args.Sport)
	if args.Name == "" {
		return args, errors.New("name is required")
	}
	if args.Sport == "" {
		return args, errors.New("sport is required")
	}
	if args.FolderID == "" {
		return args, errors.New("folder_id is required")
	}
	return args, nil
}

func createWorkoutParams(args createWorkoutRequest, options workoutdoc.SerializeOptions) (intervals.WriteWorkoutParams, string, error) {
	params := intervals.WriteWorkoutParams{Name: args.Name, FolderID: args.FolderID, Description: args.Description, Tags: append([]string(nil), args.Tags...), Sport: args.Sport}
	if args.WorkoutDoc == nil {
		return params, "", nil
	}
	prose := ""
	if args.Description != nil {
		prose = *args.Description
	}
	dsl, err := workoutdoc.MergeDescriptionWithOptions(prose, *args.WorkoutDoc, options)
	if err != nil {
		return intervals.WriteWorkoutParams{}, "", fmt.Errorf("merging workout_doc with description: %w", err)
	}
	params.Description = &dsl
	return params, "description_dsl", nil
}

func shapeCreateWorkoutResponse(workout intervals.Workout, args createWorkoutRequest, workoutDocUploaded string, profile intervals.AthleteWithSportSettings, unitSystem response.UnitSystem) createWorkoutResponse {
	uploadedSteps := args.WorkoutDoc != nil && len(args.WorkoutDoc.Steps) > 0
	return createWorkoutResponse{Workout: workoutToRow(workout, false, workoutPreviewContextForWorkout(workout, profile, unitSystem)), Meta: createWorkoutMeta{Operation: "create", SourceEndpoint: workoutLibraryWorkoutsEndpoint, FolderID: args.FolderID, Sport: args.Sport, Tags: append([]string(nil), args.Tags...), WorkoutDocUploaded: workoutDocUploaded, WorkoutDocWarning: workoutDocRenderWarning(uploadedSteps, workout.WorkoutDoc), DefaultPayloadScope: "same terse workout row shape used by get_workout_library/get_workouts_in_folder; raw workout_doc remains summarized"}}
}

func createWorkoutInputSchema() map[string]any {
	examples := createWorkoutInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"name", "folder_id", "sport"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"name":        map[string]any{"type": "string", "description": "Required workout-library template name/title. Surrounding whitespace is trimmed."},
		"folder_id":   map[string]any{"type": "string", "description": "Required intervals.icu workout-library folder ID. Must identify an existing folder owned by the athlete; top-level workout creates are refused upstream."},
		"description": map[string]any{"type": "string", "description": "Optional free-text prose for the initial upstream workout description/DSL. This is part of the single upstream description/DSL field, not separate append-only notes. Preserved verbatim. May be supplied with workout_doc; use the " + workoutdoc.StepsSentinel + " sentinel on its own line to choose where serialized steps are inserted."},
		"workout_doc": map[string]any{"type": "object", "description": "Optional structured WorkoutDoc for the initial upstream workout description/DSL. Serialized and merged with description when both are supplied. In each structured step, description is a label/comment only: do not include duration or distance tokens there; use duration seconds or distance instead. Zone targets are serialized using the supplied sport and the athlete's sport settings, adding metric suffixes such as Z2 Power, Z2 HR, or Z2 Pace when needed. Syntax reference: icuvisor://workout-syntax."},
		"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional workout-library tags to preserve on the upstream template, in caller-provided order."},
		"sport":       map[string]any{"type": "string", "description": "Required upstream sport/activity type for the workout template, such as Ride, Run, Swim, VirtualRide, or the athlete account's configured activity type. Used with athlete sport settings to disambiguate structured workout_doc zone targets."},
	}}
}

func createWorkoutInputExamples() []map[string]any {
	return []map[string]any{
		{
			"name":        "Endurance aerobic ride",
			"sport":       "Ride",
			"folder_id":   "folder-ride-1",
			"description": "60m easy aerobic endurance. Keep it conversational.",
		},
		{
			"name":      "Threshold builder",
			"sport":     "Ride",
			"folder_id": "folder-build-1",
			"workout_doc": map[string]any{
				"steps": []any{
					map[string]any{"description": "Warm up", "duration": 900, "power": map[string]any{"value": 60, "units": "PERCENT_FTP"}},
					map[string]any{"description": "Threshold", "duration": 1200, "power": map[string]any{"min": 95, "max": 100, "units": "PERCENT_FTP"}},
					map[string]any{"description": "Cool down", "duration": 600, "power": map[string]any{"value": 50, "units": "PERCENT_FTP"}},
				},
			},
			"tags": []any{"threshold", "build"},
		},
		{
			"name":      "Progressive run",
			"sport":     "Run",
			"folder_id": "folder-run-1",
			"workout_doc": map[string]any{
				"steps": []any{
					map[string]any{"duration": 900, "pace": map[string]any{"text": "easy"}},
					map[string]any{"duration": 1200, "pace": map[string]any{"text": "steady"}, "rpe": map[string]any{"value": 6}},
					map[string]any{"duration": 600, "pace": map[string]any{"text": "easy"}},
				},
			},
			"tags": []any{"run", "progression"},
		},
	}
}

func createWorkoutOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Create confirmation containing the same terse workout row shape used by workout-library read tools plus operation, source endpoint, sport, and workout_doc upload metadata. _meta.workout_doc_warning is set when intervals.icu stored the workout but did not parse the uploaded workout_doc into a graphically rendered structured workout."}
}
