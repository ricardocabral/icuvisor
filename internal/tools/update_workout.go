package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	updateWorkoutName                    = "update_workout"
	updateWorkoutDescription             = "Update one reusable workout-library template by workout_id with sparse fields only. Omitted fields are left untouched; structured workout_doc steps are serialized to the intervals.icu workout-description DSL before upload."
	invalidUpdateWorkoutArgumentsMessage = "invalid update_workout arguments; provide workout_id plus at least one sparse workout field"
	updateWorkoutMessage                 = "could not update workout; check intervals.icu credentials, athlete ID, workout ID, folder ID, and writable workout fields"
)

// WorkoutUpdaterClient sparsely updates workout-library templates for tools.
type WorkoutUpdaterClient interface {
	UpdateLibraryWorkout(context.Context, intervals.WriteWorkoutParams) (intervals.Workout, error)
}

type updateWorkoutRequest struct {
	WorkoutID   string                 `json:"workout_id"`
	Name        string                 `json:"name,omitempty"`
	FolderID    string                 `json:"folder_id,omitempty"`
	Description *string                `json:"description,omitempty"`
	WorkoutDoc  *workoutdoc.WorkoutDoc `json:"workout_doc,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Sport       string                 `json:"sport,omitempty"`

	nameProvided        bool
	folderIDProvided    bool
	descriptionProvided bool
	workoutDocProvided  bool
	tagsProvided        bool
	sportProvided       bool
}

type updateWorkoutResponse struct {
	Workout workoutTemplateRow `json:"workout"`
	Meta    updateWorkoutMeta  `json:"_meta"`
}

type updateWorkoutMeta struct {
	Operation           string   `json:"operation"`
	SourceEndpoint      string   `json:"source_endpoint"`
	WorkoutID           string   `json:"workout_id"`
	FieldsUpdated       []string `json:"fields_updated"`
	WorkoutDocUploaded  string   `json:"workout_doc_uploaded,omitempty"`
	DefaultPayloadScope string   `json:"default_payload_scope"`
}

func newUpdateWorkoutTool(client WorkoutUpdaterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: updateWorkoutName, Description: updateWorkoutDescription, InputSchema: updateWorkoutInputSchema(), OutputSchema: updateWorkoutOutputSchema(), Requirement: RequirementWrite, Handler: updateWorkoutHandler(client, profileClient, version, timezoneFallback, debugMetadata)})
}

func updateWorkoutHandler(client WorkoutUpdaterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeUpdateWorkoutRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateWorkoutArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(updateWorkoutMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(updateWorkoutMessage, errors.New("missing workout updater client"))
		}
		params, uploaded, err := updateWorkoutParams(args)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateWorkoutArgumentsMessage, err)
		}
		workout, err := client.UpdateLibraryWorkout(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(updateWorkoutMessage, err)
		}
		payload := shapeUpdateWorkoutResponse(workout, args, uploaded)
		return encodeShaped(payload, false, nil, version, debugMetadata, updateWorkoutName, unitSystem)
	}
}

func decodeUpdateWorkoutRequest(raw json.RawMessage) (updateWorkoutRequest, error) {
	fields, err := rawObjectFields(raw)
	if err != nil {
		return updateWorkoutRequest{}, err
	}
	var args updateWorkoutRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.WorkoutID = strings.TrimSpace(args.WorkoutID)
	args.Name = strings.TrimSpace(args.Name)
	args.FolderID = strings.TrimSpace(args.FolderID)
	args.Sport = strings.TrimSpace(args.Sport)
	args.nameProvided = fields["name"]
	args.folderIDProvided = fields["folder_id"]
	args.descriptionProvided = fields["description"]
	args.workoutDocProvided = fields["workout_doc"]
	args.tagsProvided = fields["tags"]
	args.sportProvided = fields["sport"]
	if args.WorkoutID == "" {
		return args, errors.New("workout_id is required")
	}
	if args.nameProvided && args.Name == "" {
		return args, errors.New("name cannot be empty when supplied")
	}
	if args.sportProvided && args.Sport == "" {
		return args, errors.New("sport cannot be empty when supplied")
	}
	if args.descriptionProvided && args.WorkoutDoc != nil {
		return args, errors.New("provide either free-text description or structured workout_doc, not both")
	}
	if args.descriptionProvided && args.Description != nil && strings.TrimSpace(*args.Description) == "" {
		return args, errors.New("use workout_doc with an explicit empty steps list to clear structured workout content")
	}
	if len(updateWorkoutFieldsUpdated(args)) == 0 {
		return args, errors.New("at least one sparse field is required")
	}
	return args, nil
}

func rawObjectFields(raw json.RawMessage) (map[string]bool, error) {
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(decoded))
	for key := range decoded {
		out[key] = true
	}
	return out, nil
}

func updateWorkoutParams(args updateWorkoutRequest) (intervals.WriteWorkoutParams, string, error) {
	params := intervals.WriteWorkoutParams{WorkoutID: args.WorkoutID, Name: args.Name, NameSet: args.nameProvided, FolderID: args.FolderID, FolderIDSet: args.folderIDProvided, Description: args.Description, DescriptionSet: args.descriptionProvided, Tags: append([]string(nil), args.Tags...), TagsSet: args.tagsProvided, Sport: args.Sport, SportSet: args.sportProvided}
	if !args.workoutDocProvided {
		return params, "", nil
	}
	if args.WorkoutDoc == nil {
		return intervals.WriteWorkoutParams{}, "", errors.New("workout_doc cannot be null")
	}
	dsl, err := workoutdoc.Serialize(*args.WorkoutDoc)
	if err != nil {
		return intervals.WriteWorkoutParams{}, "", fmt.Errorf("serializing workout_doc: %w", err)
	}
	params.Description = &dsl
	params.DescriptionSet = true
	return params, "description_dsl", nil
}

func shapeUpdateWorkoutResponse(workout intervals.Workout, args updateWorkoutRequest, workoutDocUploaded string) updateWorkoutResponse {
	return updateWorkoutResponse{Workout: workoutToRow(workout, false), Meta: updateWorkoutMeta{Operation: "update", SourceEndpoint: workoutLibraryWorkoutsEndpoint, WorkoutID: args.WorkoutID, FieldsUpdated: updateWorkoutFieldsUpdated(args), WorkoutDocUploaded: workoutDocUploaded, DefaultPayloadScope: "same terse workout row shape used by get_workout_library/get_workouts_in_folder; raw workout_doc remains summarized"}}
}

func updateWorkoutFieldsUpdated(args updateWorkoutRequest) []string {
	fields := []string{}
	if args.nameProvided {
		fields = append(fields, "name")
	}
	if args.folderIDProvided {
		fields = append(fields, "folder_id")
	}
	if args.descriptionProvided {
		fields = append(fields, "description")
	}
	if args.workoutDocProvided {
		fields = append(fields, "workout_doc")
	}
	if args.tagsProvided {
		fields = append(fields, "tags")
	}
	if args.sportProvided {
		fields = append(fields, "sport")
	}
	sort.Strings(fields)
	return fields
}

func updateWorkoutInputSchema() map[string]any {
	examples := updateWorkoutInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"workout_id"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"workout_id":  map[string]any{"type": "string", "description": "Required upstream workout-library template ID to update. Surrounding whitespace is trimmed."},
		"name":        map[string]any{"type": "string", "description": "Optional replacement workout-library template name/title. Omit to leave unchanged."},
		"folder_id":   map[string]any{"type": "string", "description": "Optional replacement intervals.icu workout-library folder or plan ID. Omit to leave unchanged; an explicit empty string moves the workout to the top level when upstream supports it."},
		"description": map[string]any{"type": "string", "description": "Optional replacement free-text workout description. Omit to leave unchanged; mutually exclusive with workout_doc. Empty strings are rejected to avoid accidentally clearing structured workout content."},
		"workout_doc": map[string]any{"type": "object", "description": "Optional replacement structured WorkoutDoc. The server serializes this to the intervals.icu workout-description DSL string and never sends workout_doc upstream; an explicit empty steps list serializes to an empty DSL to clear structured content intentionally."},
		"tags":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional replacement workout-library tags. Omit to leave unchanged; provide the full desired tag list when appending a tag."},
		"sport":       map[string]any{"type": "string", "description": "Optional replacement upstream sport/activity type such as Ride, Run, Swim, or the athlete account's configured activity type. Omit to leave unchanged."},
	}}
}

func updateWorkoutInputExamples() []map[string]any {
	return []map[string]any{
		{
			"workout_id": "workout-example-7",
			"name":       "Endurance aerobic ride - revised",
		},
		{
			"workout_id": "workout-example-8",
			"workout_doc": map[string]any{
				"steps": []any{
					map[string]any{"description": "Warm up", "duration": 600, "power": map[string]any{"value": 55, "units": "PERCENT_FTP"}},
					map[string]any{"description": "Tempo", "duration": 1800, "power": map[string]any{"min": 80, "max": 85, "units": "PERCENT_FTP"}},
				},
			},
		},
		{
			"workout_id":  "workout-example-9",
			"folder_id":   "folder-race-prep",
			"sport":       "Ride",
			"tags":        []any{"race-prep", "indoor"},
			"description": "Sharpening workout with short openers. Keep recoveries honest.",
		},
	}
}

func updateWorkoutOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Update confirmation containing the same terse workout row shape used by workout-library read tools plus operation, source endpoint, workout ID, fields-updated, and workout_doc upload metadata."}
}
