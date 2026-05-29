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
	updateWorkoutDescription             = "Update one reusable workout-library template by workout_id with sparse fields only. Omitted fields stay untouched. Before updating, present a human-readable before/after preview for user approval that covers total duration, key steps, target intensities, load/distance/time changes, and exactly which title/prose/tags/folder/structured steps are preserved. A supplied `description` and/or `workout_doc` replaces the template's upstream description/DSL; `description` is not append-only notes. To preserve structured steps while changing prose, provide the desired `workout_doc` plus prose and use the `<!-- icuvisor:steps -->` sentinel to position serialized steps. Prefer `workout_doc` when the structure is known, and call `validate_workout` first if uncertain about the DSL syntax (see icuvisor://workout-syntax for the cheat sheet and common mistakes)."
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
	Operation                     string   `json:"operation"`
	SourceEndpoint                string   `json:"source_endpoint"`
	WorkoutID                     string   `json:"workout_id"`
	FieldsUpdated                 []string `json:"fields_updated"`
	WorkoutDocUploaded            string   `json:"workout_doc_uploaded,omitempty"`
	WorkoutDocWarning             string   `json:"workout_doc_warning,omitempty"`
	DescriptionOnlyWorkoutWarning string   `json:"description_only_workout_warning,omitempty"`
	DefaultPayloadScope           string   `json:"default_payload_scope"`
}

func newUpdateWorkoutTool(client WorkoutUpdaterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: updateWorkoutName, Description: updateWorkoutDescription, InputSchema: updateWorkoutInputSchema(), OutputSchema: updateWorkoutOutputSchema(), Requirement: RequirementWrite, Handler: updateWorkoutHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func updateWorkoutHandler(client WorkoutUpdaterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
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
		return encodeShaped(payload, false, nil, version, debugMetadata, updateWorkoutName, unitSystem, shapeCfg)
	}
}

func decodeUpdateWorkoutRequest(raw json.RawMessage) (updateWorkoutRequest, error) {
	fields, err := rawObjectFields(raw)
	if err != nil {
		return updateWorkoutRequest{}, err
	}
	var args updateWorkoutRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[updateWorkoutRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
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
	if args.descriptionProvided && args.WorkoutDoc == nil && args.Description != nil && strings.TrimSpace(*args.Description) == "" {
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
	prose := ""
	if args.Description != nil {
		prose = *args.Description
	}
	dsl, err := workoutdoc.MergeDescription(prose, *args.WorkoutDoc)
	if err != nil {
		return intervals.WriteWorkoutParams{}, "", fmt.Errorf("merging workout_doc with description: %w", err)
	}
	params.Description = &dsl
	params.DescriptionSet = true
	return params, "description_dsl", nil
}

func shapeUpdateWorkoutResponse(workout intervals.Workout, args updateWorkoutRequest, workoutDocUploaded string) updateWorkoutResponse {
	uploadedSteps := args.WorkoutDoc != nil && len(args.WorkoutDoc.Steps) > 0
	return updateWorkoutResponse{Workout: workoutToRow(workout, false), Meta: updateWorkoutMeta{Operation: "update", SourceEndpoint: workoutLibraryWorkoutsEndpoint, WorkoutID: args.WorkoutID, FieldsUpdated: updateWorkoutFieldsUpdated(args), WorkoutDocUploaded: workoutDocUploaded, WorkoutDocWarning: workoutDocRenderWarning(uploadedSteps, workout.WorkoutDoc), DescriptionOnlyWorkoutWarning: updateWorkoutDescriptionOnlyWorkoutWarning(args), DefaultPayloadScope: "same terse workout row shape used by get_workout_library/get_workouts_in_folder; raw workout_doc remains summarized"}}
}

func updateWorkoutDescriptionOnlyWorkoutWarning(args updateWorkoutRequest) string {
	if !args.descriptionProvided || args.Description == nil || args.WorkoutDoc != nil {
		return ""
	}
	return descriptionOnlyWorkoutWarning
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
		"description": map[string]any{"type": "string", "description": "Optional replacement for the upstream workout description/DSL, not append-only notes. Omit to leave unchanged. Supplying description without workout_doc can replace existing structured steps; provide the desired workout_doc to preserve or merge structure. May be supplied with workout_doc; use the " + workoutdoc.StepsSentinel + " sentinel on its own line to choose where serialized steps are inserted. Empty strings without workout_doc are rejected to avoid accidentally clearing structured workout content."},
		"workout_doc": map[string]any{"type": "object", "description": "Optional replacement structured WorkoutDoc for the upstream workout description/DSL. Serialized and merged with description when both are supplied; empty steps clear structured content. Include the desired steps when changing prose so the replacement description/DSL preserves workout structure. In each structured step, description is a label/comment only: do not include duration or distance tokens there; use duration seconds or distance instead. Syntax reference: icuvisor://workout-syntax."},
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
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Update confirmation containing the same terse workout row shape used by workout-library read tools plus operation, source endpoint, workout ID, fields-updated, and workout_doc upload metadata. _meta.workout_doc_warning is set when intervals.icu stored the workout but did not parse the uploaded workout_doc into a graphically rendered structured workout. _meta.description_only_workout_warning is set when description is supplied without workout_doc."}
}
