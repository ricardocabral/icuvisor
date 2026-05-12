package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeWorkoutLibraryClient struct {
	fakeProfileClient
	folders      []intervals.WorkoutFolder
	workouts     []intervals.Workout
	folderCalls  int
	workoutCalls int
}

func (f *fakeWorkoutLibraryClient) ListWorkoutFolders(context.Context) ([]intervals.WorkoutFolder, error) {
	f.folderCalls++
	return append([]intervals.WorkoutFolder(nil), f.folders...), nil
}

func (f *fakeWorkoutLibraryClient) ListLibraryWorkouts(context.Context) ([]intervals.Workout, error) {
	f.workoutCalls++
	return append([]intervals.Workout(nil), f.workouts...), nil
}

func TestWorkoutLibraryRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if len(registrar.tools) != 3 {
		t.Fatalf("registered tool count = %d, want profile + workout library tools", len(registrar.tools))
	}
	libraryTool := findTool(t, registrar.tools, getWorkoutLibraryName)
	if !strings.Contains(libraryTool.Description, "workout-library folders") {
		t.Fatalf("library description = %q, want workout-library language", libraryTool.Description)
	}
	folderTool := findTool(t, registrar.tools, getWorkoutsInFolderName)
	if !strings.Contains(folderTool.Description, "raw workout_doc") {
		t.Fatalf("folder description = %q, want raw workout_doc language", folderTool.Description)
	}
	props := folderTool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"folder_id", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("get_workouts_in_folder schema missing %s", name)
		}
	}
}

func TestGetWorkoutLibraryFoldersAndTopLevelWorkouts(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		folders: decodeToolWorkoutFolders(t,
			`{"id":20,"type":"FOLDER","name":"Threshold","visibility":"PRIVATE","num_workouts":2,"children":[{"id":3,"name":"FTP","type":"Ride"}]}`,
			`{"id":10,"type":"PLAN","name":"Base","visibility":"PUBLIC","children":[{"id":4,"name":"Long Run","type":"Run"},{"id":5,"name":"Endurance","type":"Ride"}]}`,
		),
		workouts: decodeToolWorkouts(t,
			`{"id":1,"name":"Top Tempo","type":"Ride","icu_training_load":60,"moving_time":3600,"workout_doc":{"steps":[{"duration":600}]}}`,
			`{"id":2,"name":"Folder Tempo","type":"Ride","folder_id":20,"workout_doc":{"steps":[{"duration":300}]}}`,
		),
	}
	tool := newGetWorkoutLibraryTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_top_level_workouts":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.folderCalls != 1 || client.workoutCalls != 1 {
		t.Fatalf("calls = folders %d workouts %d, want 1/1", client.folderCalls, client.workoutCalls)
	}
	out := resultMap(t, result)
	folders := out["folders"].([]any)
	if len(folders) != 2 {
		t.Fatalf("folders = %d, want 2", len(folders))
	}
	base := folders[0].(map[string]any)
	if base["folder_id"] != "10" || base["child_count"] != float64(2) {
		t.Fatalf("base row = %#v, want sorted plan with child_count 2", base)
	}
	workouts := out["workouts"].([]any)
	if len(workouts) != 1 {
		t.Fatalf("top-level workouts = %d, want 1", len(workouts))
	}
	row := workouts[0].(map[string]any)
	if row["workout_id"] != "1" || row["name"] != "Top Tempo" {
		t.Fatalf("workout row = %#v, want top-level workout only", row)
	}
	if _, ok := row["workout_doc"]; ok {
		t.Fatalf("get_workout_library exposed raw workout_doc: %#v", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["folder_count"] != float64(2) || meta["top_level_workout_count"] != float64(1) {
		t.Fatalf("meta = %#v, want folder/top-level counts", meta)
	}
}

func TestGetWorkoutLibraryEmptyDoesNotFetchWorkoutsByDefault(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newGetWorkoutLibraryTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.workoutCalls != 0 {
		t.Fatalf("workout calls = %d, want 0 unless top-level workouts requested", client.workoutCalls)
	}
	out := resultMap(t, result)
	if len(out["folders"].([]any)) != 0 {
		t.Fatalf("folders = %#v, want empty", out["folders"])
	}
}

func TestGetWorkoutsInFolderFiltersAndPreservesWorkoutDocWithIncludeFull(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workouts: decodeToolWorkouts(t,
			`{"id":2,"name":"Sweet Spot","type":"Ride","folder_id":20,"icu_training_load":70,"moving_time":3600,"target":"POWER","tags":["sweet-spot"],"workout_doc":{"steps":[{"duration":600},{"duration":300}],"name":"raw doc"}}`,
			`{"id":1,"name":"Other Folder","type":"Run","folder_id":10,"workout_doc":{"steps":[{"duration":100}]}}`,
		),
	}
	tool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	rows := out["workouts"].([]any)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want only folder 20", len(rows))
	}
	row := rows[0].(map[string]any)
	if row["workout_id"] != "2" || row["sport"] != "Ride" || row["icu_training_load"] != float64(70) {
		t.Fatalf("row = %#v, want terse workout fields", row)
	}
	summary := row["workout_doc_summary"].(map[string]any)
	if summary["step_count"] != float64(2) {
		t.Fatalf("summary = %#v, want step_count 2", summary)
	}
	doc := row["workout_doc"].(map[string]any)
	if doc["name"] != "raw doc" {
		t.Fatalf("workout_doc = %#v, want verbatim raw doc", doc)
	}
	meta := out["_meta"].(map[string]any)
	if meta["folder_id"] != "20" || meta["include_full"] != true || meta["count"] != float64(1) {
		t.Fatalf("meta = %#v, want folder/include_full/count", meta)
	}
}

func TestGetWorkoutsInFolderHidesWorkoutDocByDefault(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workouts:          decodeToolWorkouts(t, `{"id":2,"name":"Sweet Spot","type":"Ride","folder_id":20,"workout_doc":{"steps":[{"duration":600}]}}`),
	}
	tool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["workouts"].([]any)[0].(map[string]any)
	if _, ok := row["workout_doc"]; ok {
		t.Fatalf("workout_doc present by default: %#v", row)
	}
	if _, ok := row["workout_doc_summary"]; !ok {
		t.Fatalf("workout_doc_summary missing: %#v", row)
	}
}

func decodeToolWorkoutFolders(t *testing.T, raws ...string) []intervals.WorkoutFolder {
	t.Helper()
	folders := make([]intervals.WorkoutFolder, 0, len(raws))
	for _, raw := range raws {
		var folder intervals.WorkoutFolder
		if err := json.Unmarshal([]byte(raw), &folder); err != nil {
			t.Fatalf("decode workout folder: %v", err)
		}
		folders = append(folders, folder)
	}
	return folders
}

func decodeToolWorkouts(t *testing.T, raws ...string) []intervals.Workout {
	t.Helper()
	workouts := make([]intervals.Workout, 0, len(raws))
	for _, raw := range raws {
		var workout intervals.Workout
		if err := json.Unmarshal([]byte(raw), &workout); err != nil {
			t.Fatalf("decode workout: %v", err)
		}
		workouts = append(workouts, workout)
	}
	return workouts
}
