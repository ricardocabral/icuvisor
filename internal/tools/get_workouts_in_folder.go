package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getWorkoutsInFolderName                    = "get_workouts_in_folder"
	getWorkoutsInFolderDescription             = "List workout-library templates inside one folder or plan by folder_id. Returns terse workout rows with structured-step summaries by default; raw workout_doc is returned only with include_full:true."
	invalidGetWorkoutsInFolderArgumentsMessage = "invalid get_workouts_in_folder arguments; provide folder_id and optional include_full"
	fetchWorkoutsInFolderMessage               = "could not fetch workouts in folder; check intervals.icu credentials, athlete ID, and folder ID"
)

type getWorkoutsInFolderRequest struct {
	FolderID    string `json:"folder_id"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getWorkoutsInFolderResponse struct {
	Workouts []workoutInFolderRow    `json:"workouts"`
	Meta     getWorkoutsInFolderMeta `json:"_meta"`
}

type workoutInFolderRow struct {
	WorkoutID         string         `json:"workout_id,omitempty"`
	Name              string         `json:"name,omitempty"`
	Sport             string         `json:"sport,omitempty"`
	FolderID          string         `json:"folder_id,omitempty"`
	TrainingLoad      int            `json:"icu_training_load,omitempty"`
	MovingTimeSeconds int            `json:"moving_time_seconds,omitempty"`
	DistanceMeters    *float64       `json:"distance_meters,omitempty"`
	Target            string         `json:"target,omitempty"`
	Targets           []string       `json:"targets,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Indoor            *bool          `json:"indoor,omitempty"`
	Description       string         `json:"description,omitempty"`
	WorkoutDocSummary map[string]any `json:"workout_doc_summary,omitempty"`
	WorkoutDoc        any            `json:"workout_doc,omitempty"`
}

type getWorkoutsInFolderMeta struct {
	SourceEndpoint      string `json:"source_endpoint"`
	FolderID            string `json:"folder_id"`
	Count               int    `json:"count"`
	IncludeFull         bool   `json:"include_full"`
	DefaultPayloadScope string `json:"default_payload_scope"`
}

func newGetWorkoutsInFolderTool(client WorkoutLibraryClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: getWorkoutsInFolderName, Description: getWorkoutsInFolderDescription, InputSchema: getWorkoutsInFolderInputSchema(), OutputSchema: getWorkoutsInFolderOutputSchema(), Handler: getWorkoutsInFolderHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func getWorkoutsInFolderHandler(client WorkoutLibraryClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetWorkoutsInFolderRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetWorkoutsInFolderArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchWorkoutsInFolderMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchWorkoutsInFolderMessage, errors.New("missing workout library client"))
		}
		workouts, err := client.ListLibraryWorkouts(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchWorkoutsInFolderMessage, err)
		}
		payload := shapeGetWorkoutsInFolderResponse(workouts, args)
		return encodeShaped(payload, args.IncludeFull, []string{"workouts"}, version, debugMetadata, getWorkoutsInFolderName, unitSystem)
	}
}

func decodeGetWorkoutsInFolderRequest(raw json.RawMessage) (getWorkoutsInFolderRequest, error) {
	var args getWorkoutsInFolderRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.FolderID = strings.TrimSpace(args.FolderID)
	if args.FolderID == "" {
		return args, errors.New("folder_id is required")
	}
	return args, nil
}

func shapeGetWorkoutsInFolderResponse(workouts []intervals.Workout, args getWorkoutsInFolderRequest) getWorkoutsInFolderResponse {
	rows := make([]workoutInFolderRow, 0)
	for _, workout := range workouts {
		if workoutFolderID(workout) != args.FolderID {
			continue
		}
		rows = append(rows, workoutInFolderToRow(workout, args.IncludeFull))
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Name != rows[j].Name {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].WorkoutID < rows[j].WorkoutID
	})
	return getWorkoutsInFolderResponse{Workouts: rows, Meta: getWorkoutsInFolderMeta{SourceEndpoint: workoutLibraryWorkoutsEndpoint, FolderID: args.FolderID, Count: len(rows), IncludeFull: args.IncludeFull, DefaultPayloadScope: "terse workout rows with structured-step summaries; raw workout_doc requires include_full:true"}}
}

func workoutInFolderToRow(workout intervals.Workout, includeFull bool) workoutInFolderRow {
	row := workoutInFolderRow{WorkoutID: workout.ID, Name: stringValue(workout.Name), Sport: stringValue(workout.Type), FolderID: workoutFolderID(workout), TrainingLoad: intValue(workout.TrainingLoad), MovingTimeSeconds: intValue(workout.MovingTime), DistanceMeters: workout.Distance, Target: stringValue(workout.Target), Targets: workout.Targets, Tags: workout.Tags, Indoor: workout.Indoor, Description: stringValue(workout.Description)}
	if workout.WorkoutDoc != nil {
		row.WorkoutDocSummary = workoutDocSummary(workout.WorkoutDoc)
		if includeFull {
			row.WorkoutDoc = workout.WorkoutDoc
		}
	}
	return row
}

func getWorkoutsInFolderInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"folder_id"}, "properties": map[string]any{
		"folder_id":    map[string]any{"type": "string", "description": "Required intervals.icu workout-library folder or plan ID. The public API lists all workouts, and icuvisor filters by this folder_id."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include each raw upstream workout_doc object verbatim. Default mode returns only workout_doc_summary."},
	}}
}

func getWorkoutsInFolderOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Workout-library template rows for one folder ID. Terse rows summarize structured steps; include_full preserves raw workout_doc on reads."}
}
