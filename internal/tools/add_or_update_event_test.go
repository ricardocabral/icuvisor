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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		event:             decodeToolEvents(t, `{"id":"evt-1","category":"WORKOUT","name":"Tempo","start_date_local":"2026-06-01","description":"  Coach note\nKeep this verbatim.  ","tags":["tempo","coach"],"indoor":true,"load_target":75,"distance_target":30000,"time_target":3600,"updated":"2026-06-01T12:00:00Z"}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","type":"VirtualRide","name":"Tempo","description":"  Coach note\nKeep this verbatim.  ","tags":["tempo","coach"],"indoor":true,"target_load":75,"distance_meters":30000,"moving_time_seconds":3600}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.EventID != "" || call.Date != "2026-06-01" || call.Category != "WORKOUT" || call.Type != "VirtualRide" || call.Name != "Tempo" {
		t.Fatalf("write params = %#v, want create params", call)
	}
	if call.Description == nil || *call.Description != description {
		t.Fatalf("description = %#v, want verbatim", call.Description)
	}
	if !reflect.DeepEqual(call.Tags, []string{"tempo", "coach"}) {
		t.Fatalf("tags = %#v, want preserved order", call.Tags)
	}
	if call.Indoor == nil || !*call.Indoor {
		t.Fatalf("indoor = %#v, want true", call.Indoor)
	}
	if call.TargetLoad == nil || *call.TargetLoad != 75 || call.DistanceMeters == nil || *call.DistanceMeters != 30000 || call.MovingTimeSeconds == nil || *call.MovingTimeSeconds != 3600 {
		t.Fatalf("planned metrics = %#v, want target load/distance/moving time", call)
	}

	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "evt-1" || row["name"] != "Tempo" || row["description"] != description || row["indoor"] != true || row["updated_local"] != "2026-06-01T09:00:00-03:00" {
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

func TestAddOrUpdateEventStripsSparseNullsAndPreservesRawFull(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-sparse","category":"WORKOUT","type":"Ride","name":"Sparse","start_date_local":"2026-06-03","indoor":false,"load_target":0,"distance":0,"notes":null}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-03","category":"WORKOUT","type":"Ride","name":"Sparse"}`)})
	if err != nil {
		t.Fatalf("Handler() default error = %v", err)
	}
	row := resultMap(t, result)["event"].(map[string]any)
	assertKeyAbsent(t, row, "notes")
	assertKeyAbsent(t, row, "full")
	if row["indoor"] != false || row["load_target"] != float64(0) || row["distance_meters"] != float64(0) {
		t.Fatalf("event row = %#v, want false indoor plus zero load_target and distance_meters preserved", row)
	}

	fullResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-03","category":"WORKOUT","type":"Ride","name":"Sparse","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() include_full error = %v", err)
	}
	fullRow := resultMap(t, fullResult)["event"].(map[string]any)
	full := fullRow["full"].(map[string]any)
	assertKeyPresentNil(t, full, "notes")
}

func TestAddOrUpdateEventNoteCreateAcceptsDateOnlyInput(t *testing.T) {
	t.Parallel()

	responseBytes, err := os.ReadFile(filepath.Join("..", "intervals", "testdata", "events", "note_create_response.json"))
	if err != nil {
		t.Fatalf("read NOTE response fixture: %v", err)
	}
	var noteEvents []intervals.Event
	if err := json.Unmarshal(responseBytes, &noteEvents); err != nil {
		t.Fatalf("decode NOTE response fixture: %v", err)
	}
	if len(noteEvents) != 1 {
		t.Fatalf("NOTE response fixture events = %d, want one", len(noteEvents))
	}
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		event:             noteEvents[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-05-25","category":"NOTE","name":"tp-075 fixture note","description":"tp-075 captured note fixture"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.Date != "2026-05-25" || call.Category != "NOTE" || call.Type != "" || call.Name != "tp-075 fixture note" {
		t.Fatalf("write params = %#v, want date-only NOTE create without type", call)
	}
	if call.Description == nil || *call.Description != "tp-075 captured note fixture" {
		t.Fatalf("description = %#v, want NOTE fixture description", call.Description)
	}
	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["category"] != "NOTE" || row["start_date_local"] != "2026-05-25T00:00:00" {
		t.Fatalf("event row = %#v, want NOTE response fixture with local datetime", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["date"] != "2026-05-25" {
		t.Fatalf("meta = %#v, want date-only create metadata", meta)
	}
}

func TestAddOrUpdateEventUpdateUsesEventID(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-3","category":"WORKOUT","name":"Golden","start_date_local":"2026-08-01","workout_doc":{"steps":[{"duration":600}]}}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	rawDoc, err := json.Marshal(structured)
	if err != nil {
		t.Fatalf("marshal structured fixture: %v", err)
	}
	rawArgs := json.RawMessage(`{"date":"2026-08-01","category":"WORKOUT","type":"Ride","name":"Golden","workout_doc":` + string(rawDoc) + `}`)

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
	if _, ok := meta["workout_doc_warning"]; ok {
		t.Fatalf("workout_doc_warning present when upstream rendered workout_doc: %#v", meta)
	}
}

func TestAddOrUpdateEventWarnsWhenUpstreamDoesNotRenderWorkoutDoc(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-9","category":"WORKOUT","name":"Unrendered","start_date_local":"2026-08-01"}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-08-01","category":"WORKOUT","type":"Ride","name":"Unrendered","workout_doc":{"steps":[{"description":"Warmup","duration":600,"power":{"value":65,"units":"PERCENT_FTP"}}]}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if warning, _ := meta["workout_doc_warning"].(string); warning == "" {
		t.Fatalf("workout_doc_warning = %#v, want non-empty render warning when upstream returns no workout_doc", meta["workout_doc_warning"])
	}
}

func TestAddOrUpdateEventRejectsBadArguments(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	for _, raw := range []string{
		`{"date":"2026-01-01T00:00:00Z","category":"WORKOUT"}`,
		`{"date":"2026-01-01","category":""}`,
		`{"date":"2026-01-01","category":"WORKOUT"}`,
		`{"date":"2026-01-01","category":"NOTE"}`,
		`{"date":"2026-01-01","category":"WORKOUT","type":"Ride","moving_time_seconds":-1}`,
		`{"date":"2026-01-01","category":"WORKOUT","type":"Ride","description":"note","workout_doc":{"steps":[{"duration":600}]}}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestAddOrUpdateEventRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "non-destructive") {
		t.Fatalf("description = %q, want non-destructive language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"date", "event_id", "category", "type", "name", "description", "workout_doc", "tags", "indoor", "target_load", "distance_meters", "moving_time_seconds", "elapsed_time_seconds"} {
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
