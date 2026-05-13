package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeWorkoutUpdaterClient struct {
	fakeProfileClient
	workout intervals.Workout
	calls   []intervals.WriteWorkoutParams
	err     error
}

func (f *fakeWorkoutUpdaterClient) UpdateLibraryWorkout(ctx context.Context, params intervals.WriteWorkoutParams) (intervals.Workout, error) {
	f.calls = append(f.calls, params)
	return f.workout, f.err
}

func TestUpdateWorkoutRenameUsesSparseNameOnly(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-1","name":"Renamed Tempo","type":"Ride","folder_id":"f-20","tags":["base"]}`)[0],
	}
	tool := newUpdateWorkoutTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"workout_id":" w-1 ","name":" Renamed Tempo "}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.WorkoutID != "w-1" || !call.NameSet || call.Name != "Renamed Tempo" {
		t.Fatalf("write params = %#v, want sparse rename", call)
	}
	if call.SportSet || call.FolderIDSet || call.DescriptionSet || call.TagsSet {
		t.Fatalf("write params = %#v, want omitted fields untouched", call)
	}
	out := resultMap(t, result)
	row := out["workout"].(map[string]any)
	if row["workout_id"] != "w-1" || row["name"] != "Renamed Tempo" || row["sport"] != "Ride" {
		t.Fatalf("workout row = %#v, want read-shape update", row)
	}
	meta := out["_meta"].(map[string]any)
	fields := meta["fields_updated"].([]any)
	if len(fields) != 1 || fields[0] != "name" {
		t.Fatalf("fields_updated = %#v, want name only", fields)
	}
}

func TestUpdateWorkoutSwapWorkoutDocSerializesGoldenDSL(t *testing.T) {
	t.Parallel()

	structured := readWorkoutDocFixture(t, "03-ramp-freeride-structured.json")
	wantDSL := strings.TrimRight(readTextFixture(t, "03-ramp-freeride-dsl.txt"), "\n")
	rawDoc, err := json.Marshal(structured)
	if err != nil {
		t.Fatalf("marshal structured fixture: %v", err)
	}
	client := &fakeWorkoutUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-2","name":"Ramp","type":"Ride","workout_doc":`+string(rawDoc)+`}`)[0],
	}
	tool := newUpdateWorkoutTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"workout_id":"w-2","workout_doc":` + string(rawDoc) + `}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if !call.DescriptionSet || call.Description == nil || *call.Description != wantDSL {
		t.Fatalf("description = %#v, want golden DSL %q", call.Description, wantDSL)
	}
	if call.NameSet || call.SportSet || call.FolderIDSet || call.TagsSet {
		t.Fatalf("write params = %#v, want workout_doc-only sparse update", call)
	}
	out := resultMap(t, result)
	summary := out["workout"].(map[string]any)["workout_doc_summary"].(map[string]any)
	if summary["step_count"] != float64(len(structured.Steps)) {
		t.Fatalf("workout_doc_summary = %#v, want response read-shape summary", summary)
	}
	meta := out["_meta"].(map[string]any)
	if meta["workout_doc_uploaded"] != "description_dsl" {
		t.Fatalf("meta = %#v, want DSL upload marker", meta)
	}
}

func TestUpdateWorkoutAppendTagSendsReplacementTagList(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutUpdaterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-3","name":"Tagged","type":"Run","tags":["base","new"]}`)[0],
	}
	tool := newUpdateWorkoutTool(client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"workout_id":"w-3","tags":["base","new"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || !client.calls[0].TagsSet || !reflect.DeepEqual(client.calls[0].Tags, []string{"base", "new"}) {
		t.Fatalf("write calls = %#v, want explicit replacement tag list", client.calls)
	}
}

func TestUpdateWorkoutRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutUpdaterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newUpdateWorkoutTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{"name":"Renamed"}`,
		`{"workout_id":"w-1"}`,
		`{"workout_id":"w-1","name":" "}`,
		`{"workout_id":"w-1","sport":" "}`,
		`{"workout_id":"w-1","description":"note","workout_doc":{"steps":[{"duration":600}]}}`,
		`{"workout_id":"w-1","description":""}`,
		`{"workout_id":"w-1","unknown":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestUpdateWorkoutRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutUpdaterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, updateWorkoutName)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "sparse fields") {
		t.Fatalf("description = %q, want sparse update language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"workout_id", "name", "folder_id", "description", "workout_doc", "tags", "sport"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
}
