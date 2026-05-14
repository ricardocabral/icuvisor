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
	getWorkoutLibraryName                    = "get_workout_library"
	getWorkoutLibraryDescription             = "List workout-library folders and plans, not calendar events or the active training-plan assignment. Returns terse folder rows by default; set include_top_level_workouts:true to also include uncategorized top-level workout templates."
	invalidGetWorkoutLibraryArgumentsMessage = "invalid get_workout_library arguments; only include_top_level_workouts is supported"
	fetchWorkoutLibraryMessage               = "could not fetch workout library; check intervals.icu credentials and athlete ID"
	workoutLibraryFoldersEndpoint            = "/athlete/{id}/folders"
	workoutLibraryWorkoutsEndpoint           = "/athlete/{id}/workouts"
)

// WorkoutLibraryClient retrieves workout-library folders and templates for tools.
type WorkoutLibraryClient interface {
	ListWorkoutFolders(context.Context) ([]intervals.WorkoutFolder, error)
	ListLibraryWorkouts(context.Context) ([]intervals.Workout, error)
}

type getWorkoutLibraryRequest struct {
	IncludeTopLevelWorkouts bool `json:"include_top_level_workouts,omitempty"`
}

type getWorkoutLibraryResponse struct {
	Folders  []workoutFolderRow    `json:"folders"`
	Workouts []workoutTemplateRow  `json:"workouts,omitempty"`
	Meta     getWorkoutLibraryMeta `json:"_meta"`
}

type workoutFolderRow struct {
	FolderID    string   `json:"folder_id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Type        string   `json:"type,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
	Description string   `json:"description,omitempty"`
	NumWorkouts int      `json:"num_workouts,omitempty"`
	ChildCount  int      `json:"child_count,omitempty"`
	Sports      []string `json:"sports,omitempty"`
}

type workoutTemplateRow struct {
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
	Full              map[string]any `json:"full,omitempty"`
}

type getWorkoutLibraryMeta struct {
	SourceEndpoints         []string `json:"source_endpoints"`
	FolderCount             int      `json:"folder_count"`
	TopLevelWorkoutCount    int      `json:"top_level_workout_count"`
	IncludeTopLevelWorkouts bool     `json:"include_top_level_workouts"`
	DefaultPayloadScope     string   `json:"default_payload_scope"`
}

func newGetWorkoutLibraryTool(client WorkoutLibraryClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: getWorkoutLibraryName, Description: getWorkoutLibraryDescription, InputSchema: getWorkoutLibraryInputSchema(), OutputSchema: getWorkoutLibraryOutputSchema(), Handler: getWorkoutLibraryHandler(client, profileClient, version, timezoneFallback, debugMetadata)})
}

func getWorkoutLibraryHandler(client WorkoutLibraryClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetWorkoutLibraryRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetWorkoutLibraryArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchWorkoutLibraryMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchWorkoutLibraryMessage, errors.New("missing workout library client"))
		}
		folders, err := client.ListWorkoutFolders(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchWorkoutLibraryMessage, err)
		}
		var workouts []intervals.Workout
		if args.IncludeTopLevelWorkouts {
			workouts, err = client.ListLibraryWorkouts(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(fetchWorkoutLibraryMessage, err)
			}
		}
		payload := shapeGetWorkoutLibraryResponse(folders, workouts, args.IncludeTopLevelWorkouts)
		return encodeShaped(payload, false, []string{"folders", "workouts"}, version, debugMetadata, getWorkoutLibraryName, unitSystem)
	}
}

func decodeGetWorkoutLibraryRequest(raw json.RawMessage) (getWorkoutLibraryRequest, error) {
	var args getWorkoutLibraryRequest
	if len(strings.TrimSpace(string(raw))) == 0 {
		return args, nil
	}
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	return args, nil
}

func shapeGetWorkoutLibraryResponse(folders []intervals.WorkoutFolder, workouts []intervals.Workout, includeTopLevelWorkouts bool) getWorkoutLibraryResponse {
	folderRows := make([]workoutFolderRow, 0, len(folders))
	for _, folder := range folders {
		folderRows = append(folderRows, workoutFolderToRow(folder))
	}
	sort.SliceStable(folderRows, func(i, j int) bool {
		if folderRows[i].Name != folderRows[j].Name {
			return folderRows[i].Name < folderRows[j].Name
		}
		return folderRows[i].FolderID < folderRows[j].FolderID
	})

	workoutRows := []workoutTemplateRow(nil)
	if includeTopLevelWorkouts {
		workoutRows = make([]workoutTemplateRow, 0)
		for _, workout := range workouts {
			if workoutFolderID(workout) != "" {
				continue
			}
			workoutRows = append(workoutRows, workoutToRow(workout, false))
		}
		sort.SliceStable(workoutRows, func(i, j int) bool {
			if workoutRows[i].Name != workoutRows[j].Name {
				return workoutRows[i].Name < workoutRows[j].Name
			}
			return workoutRows[i].WorkoutID < workoutRows[j].WorkoutID
		})
	}

	return getWorkoutLibraryResponse{Folders: folderRows, Workouts: workoutRows, Meta: getWorkoutLibraryMeta{SourceEndpoints: []string{workoutLibraryFoldersEndpoint, workoutLibraryWorkoutsEndpoint}, FolderCount: len(folderRows), TopLevelWorkoutCount: len(workoutRows), IncludeTopLevelWorkouts: includeTopLevelWorkouts, DefaultPayloadScope: "folders/plans plus terse top-level workouts only when include_top_level_workouts:true"}}
}

func workoutFolderToRow(folder intervals.WorkoutFolder) workoutFolderRow {
	row := workoutFolderRow{FolderID: folder.ID, Name: stringValue(folder.Name), Type: stringValue(folder.Type), Visibility: stringValue(folder.Visibility), Description: stringValue(folder.Description), NumWorkouts: intValue(folder.NumWorkouts), ChildCount: len(folder.Children)}
	row.Sports = folderSports(folder)
	return row
}

func folderSports(folder intervals.WorkoutFolder) []string {
	seen := map[string]bool{}
	var sports []string
	for _, child := range folder.Children {
		sport := stringValue(child.Type)
		if sport != "" && !seen[sport] {
			seen[sport] = true
			sports = append(sports, sport)
		}
	}
	sort.Strings(sports)
	return sports
}

func workoutToRow(workout intervals.Workout, includeFull bool) workoutTemplateRow {
	row := workoutTemplateRow{WorkoutID: workout.ID, Name: stringValue(workout.Name), Sport: stringValue(workout.Type), FolderID: workoutFolderID(workout), TrainingLoad: intValue(workout.TrainingLoad), MovingTimeSeconds: intValue(workout.MovingTime), DistanceMeters: workout.Distance, Target: stringValue(workout.Target), Targets: workout.Targets, Tags: workout.Tags, Indoor: workout.Indoor, Description: stringValue(workout.Description)}
	if workout.WorkoutDoc != nil {
		row.WorkoutDocSummary = workoutDocSummary(workout.WorkoutDoc)
	}
	if includeFull {
		row.Full = cloneJSONMap(workout.Raw)
	}
	return row
}

func workoutFolderID(workout intervals.Workout) string {
	return anyString(firstRaw(workout.Raw, "folder_id", "folderId"))
}

func getWorkoutLibraryInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"include_top_level_workouts": map[string]any{"type": "boolean", "default": false, "description": "When true, also fetch all library workouts and include only those without a folder_id as top-level workout templates."},
	}}
}

func getWorkoutLibraryOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Workout-library folders/plans with counts and optional top-level workout template rows. Raw workout_doc is summarized, not returned, in this tool."}
}
