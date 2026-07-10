package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
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

	client := &fakeWorkoutLibraryClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	libraryTool := newGetWorkoutLibraryTool(client, client, "test", "UTC", false)
	if !strings.Contains(libraryTool.Description, "workout-library folders") {
		t.Fatalf("library description = %q, want workout-library language", libraryTool.Description)
	}
	folderTool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)
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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
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
	summary := row["workout_doc_summary"].(map[string]any)
	if summary["present"] != true || summary["step_count"] != float64(1) {
		t.Fatalf("workout_doc_summary = %#v, want typed summary preserving step count", summary)
	}
	keys := summary["top_level_keys"].([]any)
	if len(keys) != 1 || keys[0] != "steps" {
		t.Fatalf("top_level_keys = %#v, want typed summary keys", keys)
	}
	meta := out["_meta"].(map[string]any)
	if meta["folder_count"] != float64(2) || meta["top_level_workout_count"] != float64(1) {
		t.Fatalf("meta = %#v, want folder/top-level counts", meta)
	}
}

func TestGetWorkoutLibraryResolvesPercentFTPTopLevelWorkoutTargetPreview(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{Types: []string{"Ride"}, FTP: 300}}}},
		workouts:          decodeToolWorkouts(t, `{"id":1,"name":"FTP Blocks","type":"Ride","workout_doc":{"steps":[{"description":"Threshold","duration":600,"power":{"value":105,"units":"PERCENT_FTP"}}]}}`),
	}
	tool := newGetWorkoutLibraryTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_top_level_workouts":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("profile calls = %d, want one call reused for units and target previews", client.calls)
	}
	row := resultMap(t, result)["workouts"].([]any)[0].(map[string]any)
	summary := row["workout_doc_summary"].(map[string]any)
	previews := summary["target_previews"].([]any)
	if len(previews) != 1 {
		t.Fatalf("target_previews = %#v, want one resolved FTP preview", previews)
	}
	preview := previews[0].(map[string]any)
	if preview["target"] != "105% FTP" || preview["preview"] != "315 W" || preview["basis"] != "ftp 300 W" {
		t.Fatalf("preview = %#v, want compact FTP watts resolution", preview)
	}
	if _, ok := row["workout_doc"]; ok {
		t.Fatalf("raw workout_doc leaked in top-level workout row: %#v", row)
	}
}

func TestGetWorkoutLibraryEmptyDoesNotFetchWorkoutsByDefault(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workouts: decodeToolWorkouts(t,
			`{"id":2,"name":"Sweet Spot","description":"multi-paragraph coach notes","type":"Ride","folder_id":20,"icu_training_load":70,"moving_time":3600,"target":"POWER","tags":["sweet-spot"],"workout_doc":{"steps":[{"duration":600},{"duration":300}],"name":"raw doc"}}`,
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
	if row["description"] != "multi-paragraph coach notes" {
		t.Fatalf("description = %#v, want include_full preserved description", row["description"])
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

func TestPaceTargetPreviewHonorsSportDisplayUnits(t *testing.T) {
	tests := []struct {
		name       string
		paceUnits  string
		unitSystem response.UnitSystem
		preview    string
		basis      string
	}{
		{name: "metric athlete kilometers", paceUnits: "MINS_KM", unitSystem: response.UnitSystemMetric, preview: "4:55/km", basis: "threshold pace 4:40/km"},
		{name: "metric athlete miles", paceUnits: "MINS_MILE", unitSystem: response.UnitSystemMetric, preview: "7:54/mi", basis: "threshold pace 7:31/mi"},
		{name: "400 meter track display", paceUnits: "SECS_400M", unitSystem: response.UnitSystemImperial, preview: "1:58/400m", basis: "threshold pace 1:52/400m"},
		{name: "250 meter display", paceUnits: "SECS_250M", unitSystem: response.UnitSystemMetric, preview: "1:14/250m", basis: "threshold pace 1:10/250m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview, basis, ok := paceTargetPreview(workoutTargetBounds{Values: []float64{95}}, 3.5714285, tt.paceUnits, tt.unitSystem)
			if !ok || preview != tt.preview || basis != tt.basis {
				t.Fatalf("paceTargetPreview() = %q, %q, %t; want %q, %q, true", preview, basis, ok, tt.preview, tt.basis)
			}
		})
	}
}

func TestPaceTargetPreviewRejectsUnformattableMPS(t *testing.T) {
	if _, _, ok := paceTargetPreview(workoutTargetBounds{Values: []float64{100}}, 1e-306, "MINS_KM", response.UnitSystemMetric); ok {
		t.Fatal("paceTargetPreview accepted an overflowing m/s-to-duration conversion")
	}
}

func TestGetWorkoutsInFolderResolvesHRAndPaceTargetPreviews(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{Types: []string{"Run"}, LTHR: 170, MaxHR: 190, ThresholdPace: 3.3333333333333335, PaceUnits: "MINS_KM"}}}},
		workouts: decodeToolWorkouts(t,
			`{"id":2,"name":"Run Targets","type":"Run","folder_id":20,"workout_doc":{"steps":[{"description":"Tempo HR","duration":600,"hr":{"min":95,"max":99,"units":"PERCENT_LTHR"}},{"description":"Cruise","duration":600,"pace":{"value":95,"units":"PERCENT_THRESHOLD"}}]}}`,
		),
	}
	tool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	summary := resultMap(t, result)["workouts"].([]any)[0].(map[string]any)["workout_doc_summary"].(map[string]any)
	previews := summary["target_previews"].([]any)
	if len(previews) != 2 {
		t.Fatalf("target_previews = %#v, want HR and pace previews", previews)
	}
	hr := previews[0].(map[string]any)
	if hr["family"] != "hr" || hr["target"] != "95-99% LTHR" || hr["preview"] != "162-168 bpm" || hr["basis"] != "lthr 170 bpm" {
		t.Fatalf("hr preview = %#v, want LTHR bpm resolution", hr)
	}
	pace := previews[1].(map[string]any)
	if pace["family"] != "pace" || pace["target"] != "95% Pace" || pace["preview"] != "5:16/km" || pace["basis"] != "threshold pace 5:00/km" {
		t.Fatalf("pace preview = %#v, want threshold pace resolution", pace)
	}
}

func TestGetWorkoutsInFolderResolvesYardSwimPaceTargetPreviews(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{Types: []string{"Swim"}, ThresholdPace: 1.016, PaceUnits: "SECS_100Y"}}}},
		workouts: decodeToolWorkouts(t,
			`{"id":3,"name":"Pool Targets","type":"Swim","folder_id":20,"workout_doc":{"steps":[{"description":"Cruise","duration":600,"pace":{"value":95,"units":"PERCENT_THRESHOLD"}}]}}`,
		),
	}
	tool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	summary := resultMap(t, result)["workouts"].([]any)[0].(map[string]any)["workout_doc_summary"].(map[string]any)
	previews := summary["target_previews"].([]any)
	if len(previews) != 1 {
		t.Fatalf("target_previews = %#v, want one pace preview", previews)
	}
	pace := previews[0].(map[string]any)
	if pace["family"] != "pace" || pace["target"] != "95% Pace" || pace["preview"] != "1:35/100y" || pace["basis"] != "threshold pace 1:30/100y" {
		t.Fatalf("pace preview = %#v, want threshold pace resolved as /100y", pace)
	}
}

func TestGetWorkoutsInFolderOmitsUnsupportedTargetPreviews(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{Types: []string{"Ride"}}}}},
		workouts: decodeToolWorkouts(t,
			`{"id":2,"name":"Unsupported","type":"Ride","folder_id":20,"workout_doc":{"steps":[{"description":"No FTP","duration":600,"power":{"value":90,"units":"PERCENT_FTP"}},{"description":"Absolute","duration":300,"power":{"value":220,"units":"WATTS"}},{"description":"Text","duration":300,"pace":{"text":"Marathon Pace"}}]}}`,
		),
	}
	tool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	summary := resultMap(t, result)["workouts"].([]any)[0].(map[string]any)["workout_doc_summary"].(map[string]any)
	assertKeyAbsent(t, summary, "target_previews")
}

func TestGetWorkoutsInFolderHidesWorkoutDocByDefault(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workouts:          decodeToolWorkouts(t, `{"id":2,"name":"Sweet Spot","description":"multi-paragraph coach notes","type":"Ride","folder_id":20,"workout_doc":{"steps":[{"duration":600}]}}`),
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
	if _, ok := row["description"]; ok {
		t.Fatalf("description present by default: %#v", row)
	}
	if _, ok := row["workout_doc_summary"]; !ok {
		t.Fatalf("workout_doc_summary missing: %#v", row)
	}
}

func TestGetWorkoutsInFolderDefaultOmitsLargeRawPayloads(t *testing.T) {
	t.Parallel()

	largeDescription := strings.Repeat("coach note ", 600)
	largeStepNote := strings.Repeat("hold form ", 600)
	client := &fakeWorkoutLibraryClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workouts: decodeToolWorkouts(t,
			`{"id":1,"name":"Big Tempo","description":`+quoteJSON(t, largeDescription)+`,"type":"Ride","folder_id":20,"workout_doc":{"steps":[{"duration":600,"text":`+quoteJSON(t, largeStepNote)+`}]}}`,
			`{"id":2,"name":"Big VO2","description":`+quoteJSON(t, largeDescription)+`,"type":"Ride","folder_id":20,"workout_doc":{"steps":[{"duration":300,"text":`+quoteJSON(t, largeStepNote)+`}]}}`,
		),
	}
	tool := newGetWorkoutsInFolderTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	rows := out["workouts"].([]any)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal default output: %v", err)
	}
	if strings.Contains(string(encoded), largeDescription) || strings.Contains(string(encoded), largeStepNote) {
		t.Fatalf("default response leaked large raw payload: %s", encoded)
	}
	for _, got := range rows {
		row := got.(map[string]any)
		if _, ok := row["description"]; ok {
			t.Fatalf("description present by default: %#v", row)
		}
		if _, ok := row["workout_doc"]; ok {
			t.Fatalf("workout_doc present by default: %#v", row)
		}
		if _, ok := row["workout_doc_summary"]; !ok {
			t.Fatalf("workout_doc_summary missing: %#v", row)
		}
	}

	fullResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"folder_id":"20","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler(include_full) error = %v", err)
	}
	fullEncoded, err := json.Marshal(resultMap(t, fullResult))
	if err != nil {
		t.Fatalf("marshal full output: %v", err)
	}
	if !strings.Contains(string(fullEncoded), largeDescription) || !strings.Contains(string(fullEncoded), largeStepNote) {
		t.Fatalf("include_full response omitted raw payload detail: %s", fullEncoded)
	}
}

func quoteJSON(t *testing.T, value string) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("quote json: %v", err)
	}
	return string(encoded)
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
