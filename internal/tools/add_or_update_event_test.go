package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

type fakeEventWriterClient struct {
	fakeProfileClient
	event intervals.Event
	calls []intervals.WriteEventParams
	err   error
}

func (f *fakeEventWriterClient) AddOrUpdateEvent(ctx context.Context, params intervals.WriteEventParams) (intervals.Event, error) {
	f.calls = append(f.calls, params)
	return f.event, f.err
}

func TestAddOrUpdateEventCreatePreservesFreeTextTagsAndReadShape(t *testing.T) {
	t.Parallel()

	description := "  Coach note\nKeep this verbatim.  "
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		event:             decodeToolEvents(t, `{"id":"evt-1","category":"WORKOUT","name":"Tempo","start_date_local":"2026-06-01","description":"  Coach note\nKeep this verbatim.  ","tags":["tempo","coach"],"load_target":75,"distance_target":30000,"time_target":3600,"updated":"2026-06-01T12:00:00Z"}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","name":"Tempo","description":"  Coach note\nKeep this verbatim.  ","tags":["tempo","coach"],"target_load":75,"distance_meters":30000,"moving_time_seconds":3600}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.EventID != "" || call.Date != "2026-06-01" || call.Category != "WORKOUT" || call.Name != "Tempo" {
		t.Fatalf("write params = %#v, want create params", call)
	}
	if call.Description == nil || *call.Description != description {
		t.Fatalf("description = %#v, want verbatim", call.Description)
	}
	if !reflect.DeepEqual(call.Tags, []string{"tempo", "coach"}) {
		t.Fatalf("tags = %#v, want preserved order", call.Tags)
	}
	if call.TargetLoad == nil || *call.TargetLoad != 75 || call.DistanceMeters == nil || *call.DistanceMeters != 30000 || call.MovingTimeSeconds == nil || *call.MovingTimeSeconds != 3600 {
		t.Fatalf("planned metrics = %#v, want target load/distance/moving time", call)
	}

	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "evt-1" || row["name"] != "Tempo" || row["description"] != description || row["updated_local"] != "2026-06-01T09:00:00-03:00" {
		t.Fatalf("event row = %#v, want get_event_by_id-compatible row", row)
	}
	if row["load_target"] != float64(75) || row["distance_target_meters"] != float64(30000) || row["time_target_seconds"] != float64(3600) {
		t.Fatalf("planned target row fields = %#v, want load/distance/time targets", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["date"] != "2026-06-01" || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("meta = %#v, want create metadata", meta)
	}
}

func TestAddOrUpdateEventUpdateUsesEventID(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-2","category":"RACE","name":"Updated race","start_date_local":"2026-07-01"}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":" evt-2 ","date":"2026-07-01","category":"RACE","name":"Updated race"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].EventID != "evt-2" {
		t.Fatalf("write calls = %#v, want trimmed event_id update", client.calls)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "update" {
		t.Fatalf("meta operation = %#v, want update", meta["operation"])
	}
}

func TestAddOrUpdateEventSerializesWorkoutDocGoldenFixture(t *testing.T) {
	t.Parallel()

	structured := readWorkoutDocFixture(t, "01-steady-power-cadence-structured.json")
	wantDSL := strings.TrimRight(readTextFixture(t, "01-steady-power-cadence-dsl.txt"), "\n")
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-3","category":"WORKOUT","name":"Golden","start_date_local":"2026-08-01","workout_doc":{"steps":[{"duration":600}]}}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	rawDoc, err := json.Marshal(structured)
	if err != nil {
		t.Fatalf("marshal structured fixture: %v", err)
	}
	rawArgs := json.RawMessage(`{"date":"2026-08-01","category":"WORKOUT","name":"Golden","workout_doc":` + string(rawDoc) + `}`)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: rawArgs})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.Description == nil || *call.Description != wantDSL {
		t.Fatalf("Description = %#v, want golden DSL %q", call.Description, wantDSL)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["workout_doc_uploaded"] != "description_dsl" {
		t.Fatalf("meta = %#v, want description_dsl upload marker", meta)
	}
}

func TestAddOrUpdateEventRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{"date":"2026-01-01T00:00:00Z","category":"WORKOUT"}`,
		`{"date":"2026-01-01","category":""}`,
		`{"date":"2026-01-01","category":"WORKOUT","moving_time_seconds":-1}`,
		`{"date":"2026-01-01","category":"WORKOUT","description":"note","workout_doc":{"steps":[{"duration":600}]}}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestAddOrUpdateEventRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, addOrUpdateEventName)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "non-destructive") {
		t.Fatalf("description = %q, want non-destructive language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"date", "event_id", "category", "name", "description", "workout_doc", "tags", "target_load", "distance_meters", "moving_time_seconds", "elapsed_time_seconds"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
}

func readWorkoutDocFixture(t *testing.T, name string) workoutdoc.WorkoutDoc {
	t.Helper()
	var doc workoutdoc.WorkoutDoc
	if err := json.Unmarshal([]byte(readTextFixture(t, name)), &doc); err != nil {
		t.Fatalf("decode workoutdoc fixture %s: %v", name, err)
	}
	return doc
}

func readTextFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "workoutdoc", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(data)
}
