package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

type fakeWorkoutCreatorClient struct {
	fakeProfileClient
	workout intervals.Workout
	calls   []intervals.WriteWorkoutParams
	err     error
}

func (f *fakeWorkoutCreatorClient) CreateLibraryWorkout(ctx context.Context, params intervals.WriteWorkoutParams) (intervals.Workout, error) {
	f.calls = append(f.calls, params)
	return f.workout, f.err
}

func TestCreateWorkoutWithStructuredStepsSerializesDSLAndReturnsReadShape(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutCreatorClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-1","name":"Sweet Spot","type":"Ride","folder_id":"f-20","tags":["sweet-spot"],"workout_doc":{"steps":[{"duration":600},{"duration":300}],"name":"Sweet Spot"}}`)[0],
	}
	tool := newCreateWorkoutTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"name":"Sweet Spot","folder_id":"f-20","sport":"Ride","tags":["sweet-spot"],"workout_doc":{"steps":[{"description":"Warmup","duration":600,"power":{"value":65,"units":"PERCENT_FTP"}},{"duration":300,"freeride":true}]}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.Name != "Sweet Spot" || call.FolderID != "f-20" || call.Sport != "Ride" {
		t.Fatalf("write params = %#v, want create inputs", call)
	}
	if call.Description == nil || *call.Description != "- Warmup 10m 65%\n- 5m freeride" {
		t.Fatalf("description DSL = %#v, want serialized workout_doc", call.Description)
	}
	if !reflect.DeepEqual(call.Tags, []string{"sweet-spot"}) {
		t.Fatalf("tags = %#v, want preserved order", call.Tags)
	}

	out := resultMap(t, result)
	row := out["workout"].(map[string]any)
	if row["workout_id"] != "w-1" || row["name"] != "Sweet Spot" || row["sport"] != "Ride" || row["folder_id"] != "f-20" {
		t.Fatalf("workout row = %#v, want get_workout_library-compatible row", row)
	}
	summary := row["workout_doc_summary"].(map[string]any)
	if summary["step_count"] != float64(2) || summary["name"] != "Sweet Spot" {
		t.Fatalf("workout_doc_summary = %#v, want returned read-shape summary", summary)
	}
	if _, ok := row["workout_doc"]; ok {
		t.Fatalf("raw workout_doc present by default: %#v", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["workout_doc_uploaded"] != "description_dsl" || meta["sport"] != "Ride" {
		t.Fatalf("meta = %#v, want create metadata", meta)
	}
}

func TestCreateWorkoutWithFreeTextOnlyPreservesDescription(t *testing.T) {
	t.Parallel()

	description := "  Coach note\nKeep this verbatim.  "
	client := &fakeWorkoutCreatorClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-2","name":"Free Text","type":"Run","description":"  Coach note\nKeep this verbatim.  ","tags":["coach"]}`)[0],
	}
	tool := newCreateWorkoutTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"name":"Free Text","folder_id":"f-test-folder","sport":"Run","description":"  Coach note\nKeep this verbatim.  ","tags":["coach"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	if client.calls[0].FolderID != "f-test-folder" {
		t.Fatalf("folder ID = %q, want sanitized test folder", client.calls[0].FolderID)
	}
	if client.calls[0].Description == nil || *client.calls[0].Description != description {
		t.Fatalf("description = %#v, want verbatim", client.calls[0].Description)
	}
	out := resultMap(t, result)
	row := out["workout"].(map[string]any)
	if row["description"] != description || row["sport"] != "Run" {
		t.Fatalf("workout row = %#v, want free-text read shape", row)
	}
	meta := out["_meta"].(map[string]any)
	if _, ok := meta["workout_doc_uploaded"]; ok {
		t.Fatalf("meta = %#v, want no workout_doc upload marker for free text", meta)
	}
}

func TestCreateWorkoutStripsSparseNullsAndPreservesFalseZero(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutCreatorClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-sparse","name":"Sparse","type":"Ride","folder_id":"f-20","description":null,"target":null,"indoor":false,"distance":0}`)[0],
	}
	tool := newCreateWorkoutTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"name":"Sparse","folder_id":"f-20","sport":"Ride"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["workout"].(map[string]any)
	assertKeyAbsent(t, row, "description")
	assertKeyAbsent(t, row, "target")
	assertKeyAbsent(t, row, "full")
	if row["indoor"] != false || row["distance_meters"] != float64(0) {
		t.Fatalf("workout row = %#v, want indoor=false and distance_meters=0 preserved", row)
	}
}

func TestCreateWorkoutGoldenFixtureRoundTripFromWorkoutDocSerializer(t *testing.T) {
	t.Parallel()

	structured := readWorkoutDocFixture(t, "02-repeat-recovery-structured.json")
	wantDSL := strings.TrimRight(readTextFixture(t, "02-repeat-recovery-dsl.txt"), "\n")
	rawDoc, err := json.Marshal(structured)
	if err != nil {
		t.Fatalf("marshal structured fixture: %v", err)
	}
	parsed, err := workoutdoc.Parse(wantDSL)
	if err != nil {
		t.Fatalf("parse golden DSL: %v", err)
	}
	if !reflect.DeepEqual(parsed.Steps, structured.Steps) {
		t.Fatalf("golden fixture parse/serialize mismatch: parsed=%#v structured=%#v", parsed.Steps, structured.Steps)
	}
	client := &fakeWorkoutCreatorClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		workout:           decodeToolWorkouts(t, `{"id":"w-3","name":"Golden","type":"Ride","workout_doc":`+string(rawDoc)+`}`)[0],
	}
	tool := newCreateWorkoutTool(client, client, "test", "UTC", false)

	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"name":"Golden","folder_id":"f-test-folder","sport":"Ride","workout_doc":` + string(rawDoc) + `}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].FolderID != "f-test-folder" || client.calls[0].Description == nil || *client.calls[0].Description != wantDSL {
		t.Fatalf("description call = %#v, want folder ID and golden DSL %q", client.calls, wantDSL)
	}
}

func TestCreateWorkoutRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutCreatorClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newCreateWorkoutTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{"sport":"Ride"}`,
		`{"name":"Tempo"}`,
		`{"name":"Tempo","sport":"Ride"}`,
		`{"name":"Tempo","folder_id":"   ","sport":"Ride"}`,
		`{"name":"Tempo","folder_id":"f-test-folder","sport":"Ride","description":"note","workout_doc":{"steps":[{"duration":600}]}}`,
		`{"name":"Tempo","folder_id":"f-test-folder","sport":"Ride","unknown":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestCreateWorkoutRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeWorkoutCreatorClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newCreateWorkoutTool(client, client, "test", "UTC", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "workout-library template") {
		t.Fatalf("description = %q, want workout-library language without confirm", tool.Description)
	}
	schema := tool.InputSchema.(map[string]any)
	props := schema["properties"].(map[string]any)
	for _, name := range []string{"name", "folder_id", "description", "workout_doc", "tags", "sport"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	required := schema["required"].([]string)
	if !reflect.DeepEqual(required, []string{"name", "folder_id", "sport"}) {
		t.Fatalf("schema required = %#v, want name/folder_id/sport", required)
	}
	folderDescription := props["folder_id"].(map[string]any)["description"].(string)
	if !strings.Contains(folderDescription, "existing folder") || !strings.Contains(folderDescription, "owned by the athlete") {
		t.Fatalf("folder_id description = %q, want existing-folder contract", folderDescription)
	}
	for _, field := range []string{"examples", "input_examples"} {
		examples := schema[field].([]map[string]any)
		for i, example := range examples {
			folderID, ok := example["folder_id"].(string)
			if !ok || strings.TrimSpace(folderID) == "" {
				t.Fatalf("schema %s[%d] folder_id = %#v, want non-blank example folder", field, i, example["folder_id"])
			}
		}
	}
}
