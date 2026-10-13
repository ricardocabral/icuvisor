package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	event     intervals.Event
	events    []intervals.Event
	calls     []intervals.WriteEventParams
	listCalls []intervals.ListEventsParams
	err       error
}

func (f *fakeEventWriterClient) AddOrUpdateEvent(ctx context.Context, params intervals.WriteEventParams) (intervals.Event, error) {
	f.calls = append(f.calls, params)
	return f.event, f.err
}

func (f *fakeEventWriterClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	return append([]intervals.Event(nil), f.events...), nil
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
	rowTags := row["tags"].([]any)
	if len(rowTags) != 2 || rowTags[0] != "tempo" || rowTags[1] != "coach" {
		t.Fatalf("row tags = %#v, want returned event tags", rowTags)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["date"] != "2026-06-01" || meta["timezone"] != "America/Sao_Paulo" || meta["confirmation_status"] != writeReturnedByUpstreamStatus {
		t.Fatalf("meta = %#v, want create metadata with returned-write confirmation", meta)
	}
}

func TestAddOrUpdateEventAcceptsPresentationFields(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-presentation","category":"NOTE","name":"Travel","start_date_local":"2026-06-02T00:00:00","color":"#ff8800","not_on_fitness_chart":true}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-02","category":"NOTE","name":"Travel","color":" #ff8800 ","not_on_fitness_chart":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.Color == nil || *call.Color != "#ff8800" {
		t.Fatalf("color = %#v, want trimmed color", call.Color)
	}
	if call.NotOnFitnessChart == nil || !*call.NotOnFitnessChart {
		t.Fatalf("not_on_fitness_chart = %#v, want true", call.NotOnFitnessChart)
	}

	row := resultMap(t, result)["event"].(map[string]any)
	if row["color"] != "#ff8800" || row["not_on_fitness_chart"] != true {
		t.Fatalf("event row = %#v, want presentation fields", row)
	}
}

func TestAddOrUpdateEventAcceptsWeightTrainingAsFreeTextCalendarEvent(t *testing.T) {
	t.Parallel()

	description := "## Strength A\n\n- Back squat: 3 × 5 @ 80 kg total.\n- Split squat: 3 × 8/side @ 16 kg per hand."
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"gym-1","category":"WORKOUT","type":"WeightTraining","name":"Gym strength block","start_date_local":"2026-07-08T00:00:00","description":"## Strength A\n\n- Back squat: 3 × 5 @ 80 kg total.\n- Split squat: 3 × 8/side @ 16 kg per hand.","tags":["gym","strength"],"time_target":2700}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-07-08","category":"WORKOUT","type":"WeightTraining","name":"Gym strength block","description":"## Strength A\n\n- Back squat: 3 × 5 @ 80 kg total.\n- Split squat: 3 × 8/side @ 16 kg per hand.","tags":["gym","strength"],"moving_time_seconds":2700}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if call.Type != "WeightTraining" || call.Description == nil || *call.Description != description {
		t.Fatalf("write params = %#v, want WeightTraining free-text event", call)
	}
	row := resultMap(t, result)["event"].(map[string]any)
	if row["type"] != "WeightTraining" || row["description"] != description || row["time_target_seconds"] != float64(2700) {
		t.Fatalf("event row = %#v, want generic WeightTraining calendar event", row)
	}
	for _, unsupported := range []string{"workout_doc_summary", "exercises", "sets", "reps", "load_kg"} {
		if _, ok := row[unsupported]; ok {
			t.Fatalf("row includes unsupported structured strength key %q: %#v", unsupported, row)
		}
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if _, ok := meta["workout_doc_uploaded"]; ok {
		t.Fatalf("meta includes workout_doc_uploaded for free-text WeightTraining event: %#v", meta)
	}
}

func TestAddOrUpdateEventMapsExternalIDArgument(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-ext","external_id":"icuvisor-test-ext","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"2026-06-01"}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","type":"Ride","name":"Tempo","external_id":" icuvisor-test-ext "}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].ExternalID != "icuvisor-test-ext" {
		t.Fatalf("write calls = %#v, want trimmed external_id mapped", client.calls)
	}
	row := resultMap(t, result)["event"].(map[string]any)
	if row["external_id"] != "icuvisor-test-ext" {
		t.Fatalf("event row = %#v, want external_id exposed", row)
	}
}

func TestAddOrUpdateEventCreateSkipsSameDayMatchingExternalID(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            decodeToolEvents(t, `{"id":"evt-existing-ext","external_id":"icuvisor-ext-1","category":"WORKOUT","type":"Ride","name":"Older body","start_date_local":"2026-06-01T00:00:00","load_target":42}`),
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","type":"Ride","name":"Retried body","external_id":"icuvisor-ext-1","target_load":75}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want matching external_id create skipped", client.calls)
	}
	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "evt-existing-ext" || row["external_id"] != "icuvisor-ext-1" {
		t.Fatalf("event row = %#v, want existing external_id duplicate", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "skip_duplicate" || meta["duplicate_event_id"] != "evt-existing-ext" || meta["duplicate_warning"] != duplicateExternalIDSkippedWarning || meta["confirmation_status"] != skippedExistingEventStatus {
		t.Fatalf("meta = %#v, want external_id duplicate skip metadata", meta)
	}
	conflicts := meta["same_day_conflicts"].([]any)
	if len(conflicts) != 1 || conflicts[0].(map[string]any)["reason"] != "matching_external_id" || conflicts[0].(map[string]any)["date"] != "2026-06-01" {
		t.Fatalf("same_day_conflicts = %#v, want matching_external_id conflict", conflicts)
	}
}

func TestAddOrUpdateEventCreateSkipsExactSameDayDuplicate(t *testing.T) {
	t.Parallel()

	description := "Tempo prescription"
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            decodeToolEvents(t, `{"id":"evt-existing","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"2026-06-01T00:00:00","description":"Tempo prescription","tags":["tempo"],"color":"#ff8800","not_on_fitness_chart":true,"indoor":true,"load_target":75,"distance_target":30000,"time_target":3600}`),
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","type":"Ride","name":"Tempo","description":"Tempo prescription","tags":["tempo"],"color":"#ff8800","not_on_fitness_chart":true,"indoor":true,"target_load":75,"distance_meters":30000,"moving_time_seconds":3600}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want duplicate create skipped", client.calls)
	}
	if len(client.listCalls) != 1 || client.listCalls[0].Oldest != "2026-06-01" || client.listCalls[0].Newest != "2026-06-01" {
		t.Fatalf("ListEvents calls = %#v, want same-day duplicate preflight", client.listCalls)
	}
	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "evt-existing" {
		t.Fatalf("event row = %#v, want existing duplicate", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "skip_duplicate" || meta["duplicate_event_id"] != "evt-existing" {
		t.Fatalf("meta = %#v, want duplicate skip metadata", meta)
	}
	_ = description
}

func TestAddOrUpdateEventCreateWarnsInsteadOfSkippingNonExactSameDayEvent(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            decodeToolEvents(t, `{"id":"evt-other","category":"WORKOUT","type":"Ride","name":"Different workout","start_date_local":"2026-06-01T00:00:00","load_target":80}`),
		event:             decodeToolEvents(t, `{"id":"evt-created","category":"WORKOUT","type":"Ride","start_date_local":"2026-06-01T00:00:00"}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","type":"Ride"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %#v, want non-exact same-day event to warn and create", client.calls)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["duplicate_warning"] == nil {
		t.Fatalf("meta = %#v, want create with duplicate warning", meta)
	}
	conflicts := meta["same_day_conflicts"].([]any)
	if len(conflicts) != 1 || conflicts[0].(map[string]any)["event_id"] != "evt-other" {
		t.Fatalf("same_day_conflicts = %#v, want existing non-exact event", conflicts)
	}
}

func TestAddOrUpdateEventCreateDoesNotTreatActualMetricsAsTargetDuplicate(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            decodeToolEvents(t, `{"id":"evt-actual","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"2026-06-01T00:00:00","icu_training_load":75,"distance":30000,"moving_time":3600}`),
		event:             decodeToolEvents(t, `{"id":"evt-created","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"2026-06-01T00:00:00","load_target":75,"distance_target":30000,"time_target":3600}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-06-01","category":"WORKOUT","type":"Ride","name":"Tempo","target_load":75,"distance_meters":30000,"moving_time_seconds":3600}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %#v, want actual metrics without target fields to warn and create", client.calls)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["operation"] != "create" || meta["duplicate_event_id"] != nil {
		t.Fatalf("meta = %#v, want create rather than duplicate skip", meta)
	}
	conflicts := meta["same_day_conflicts"].([]any)
	if len(conflicts) != 1 || conflicts[0].(map[string]any)["event_id"] != "evt-actual" {
		t.Fatalf("same_day_conflicts = %#v, want actual-metric event conflict", conflicts)
	}
}

func TestAddOrUpdateEventAcceptsLongDistanceRaceMeters(t *testing.T) {
	t.Parallel()

	const brevetDistanceMeters = 1_200_000.0
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-1200","category":"RACE","name":"1200 km brevet","start_date_local":"2026-08-01","distance_target":1200000}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	createResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-08-01","category":"RACE","name":"1200 km brevet","distance_meters":1200000}`)})
	if err != nil {
		t.Fatalf("Handler(create) error = %v", err)
	}
	updateResult, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-08-01","event_id":"evt-1200","category":"RACE","name":"1200 km brevet","distance_meters":1200000}`)})
	if err != nil {
		t.Fatalf("Handler(update) error = %v", err)
	}
	if len(client.calls) != 2 {
		t.Fatalf("write calls = %d, want create and update", len(client.calls))
	}
	if client.calls[0].EventID != "" || client.calls[1].EventID != "evt-1200" {
		t.Fatalf("event IDs = %q/%q, want create then update", client.calls[0].EventID, client.calls[1].EventID)
	}
	for idx, call := range client.calls {
		if call.DistanceMeters == nil || *call.DistanceMeters != brevetDistanceMeters {
			t.Fatalf("call %d distance_meters = %#v, want 1200 km in meters", idx, call.DistanceMeters)
		}
		if call.TargetLoad != nil {
			t.Fatalf("call %d target_load = %#v, want omitted rather than auto-calculated", idx, call.TargetLoad)
		}
	}
	for _, result := range []Result{createResult, updateResult} {
		out := resultMap(t, result)
		row := out["event"].(map[string]any)
		if row["distance_target_meters"] != brevetDistanceMeters {
			t.Fatalf("event row = %#v, want untruncated 1200 km target distance", row)
		}
		assertKeyAbsent(t, row, "load_target")
		lowerText := strings.ToLower(resultText(t, result))
		for _, forbidden := range []string{"auto-load", "autocalc", "auto calculated", "auto-calculated", "calculated load"} {
			if strings.Contains(lowerText, forbidden) {
				t.Fatalf("result text contains false auto-load wording %q: %s", forbidden, lowerText)
			}
		}
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
	assertKeyAbsent(t, row, "tags")
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

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":" evt-2 ","external_id":" ext-updated ","date":"2026-07-01","category":"RACE","name":"Updated race"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].EventID != "evt-2" || client.calls[0].ExternalID != "ext-updated" {
		t.Fatalf("write calls = %#v, want trimmed event_id and external_id update", client.calls)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "update" || meta["confirmation_status"] != writeReturnedByUpstreamStatus {
		t.Fatalf("meta = %#v, want update with returned-write confirmation", meta)
	}
}

func TestAddOrUpdateEventDescriptionOnlyWorkoutWarning(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		args        string
		wantWarning bool
	}{
		{
			name:        "workout update with description and no workout_doc warns",
			args:        `{"event_id":"evt-2","date":"2026-07-01","category":"WORKOUT","type":"Ride","description":"Add coach note only."}`,
			wantWarning: true,
		},
		{
			name: "workout update with workout_doc does not warn",
			args: `{"event_id":"evt-2","date":"2026-07-01","category":"WORKOUT","type":"Ride","description":"Keep structure.","workout_doc":{"steps":[{"duration":600,"power":{"value":60,"units":"PERCENT_FTP"}}]}}`,
		},
		{
			name: "workout create with description does not warn",
			args: `{"date":"2026-07-01","category":"WORKOUT","type":"Ride","description":"Strength session prose."}`,
		},
		{
			name: "non-workout update with description does not warn",
			args: `{"event_id":"evt-race","date":"2026-07-01","category":"RACE","description":"Race plan."}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeEventWriterClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
				event:             decodeToolEvents(t, `{"id":"evt-2","category":"WORKOUT","name":"Updated","start_date_local":"2026-07-01","workout_doc":{"steps":[{"duration":600}]}}`)[0],
			}
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			meta := resultMap(t, result)["_meta"].(map[string]any)
			warning, ok := meta["description_only_workout_warning"].(string)
			if tc.wantWarning {
				if !ok || warning != descriptionOnlyWorkoutWarning {
					t.Fatalf("description_only_workout_warning = %#v, want %q", meta["description_only_workout_warning"], descriptionOnlyWorkoutWarning)
				}
				return
			}
			if ok {
				t.Fatalf("description_only_workout_warning present = %q", warning)
			}
		})
	}
}

func TestAddOrUpdateEventCanClearTags(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-2","category":"WORKOUT","name":"Untagged","start_date_local":"2026-07-01","tags":[]}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"evt-2","date":"2026-07-01","category":"WORKOUT","type":"Ride","tags":[]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("write calls = %d, want 1", len(client.calls))
	}
	call := client.calls[0]
	if !call.TagsSet {
		t.Fatalf("TagsSet = false, want true for explicit empty tags")
	}
	if len(call.Tags) != 0 {
		t.Fatalf("Tags = %#v, want empty replacement list", call.Tags)
	}
}

func TestAddOrUpdateEventSerializesRepeatWorkoutDocGoldenFixture(t *testing.T) {
	t.Parallel()

	structured := readWorkoutDocFixture(t, "02-repeat-recovery-structured.json")
	wantDSL := strings.TrimRight(readTextFixture(t, "02-repeat-recovery-dsl.txt"), "\n")
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-3","category":"WORKOUT","name":"Golden","start_date_local":"2026-08-01","workout_doc":{"steps":[{"description":"Main Set","reps":3,"steps":[{"duration":300,"power":{"value":110,"units":"PERCENT_FTP"}},{"description":"Recovery","duration":180,"power":{"value":50,"units":"PERCENT_FTP"}}]}]}}`)[0],
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
	firstLine, _, _ := strings.Cut(*call.Description, "\n")
	if firstLine != "Main Set 3x" || strings.HasPrefix(firstLine, "-") {
		t.Fatalf("repeat header = %q, want canonical header without leading dash", firstLine)
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

func TestAddOrUpdateEventMergesDescriptionAndWorkoutDoc(t *testing.T) {
	t.Parallel()

	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":"evt-merge","category":"WORKOUT","name":"Merged","start_date_local":"2026-08-02","workout_doc":{"steps":[{"duration":600}]}}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	prose := "Coach note before.\n" + workoutdoc.StepsSentinel + "\nFuel after."
	rawArgs := mustMarshalArgs(t, map[string]any{
		"date":        "2026-08-02",
		"category":    "WORKOUT",
		"type":        "Ride",
		"name":        "Merged",
		"description": prose,
		"workout_doc": map[string]any{"steps": []any{map[string]any{"description": "Warmup", "duration": 600, "power": map[string]any{"value": 60, "units": "PERCENT_FTP"}}}},
	})

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(rawArgs)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	want := "Coach note before.\n- Warmup 10m 60%\nFuel after."
	if len(client.calls) != 1 || client.calls[0].Description == nil || *client.calls[0].Description != want {
		t.Fatalf("description = %#v, want merged DSL %q", client.calls, want)
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
		`{"date":"2026-01-01","category":"NOTE","name":"Blank color","color":"   "}`,
		`{"date":"2026-01-01","category":"WORKOUT","type":"Ride","workout_doc":{"steps":[{"description":"10m warmup","duration":600}]}}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestAddOrUpdateEventRaceInputExamplesIncludePlanningFields(t *testing.T) {
	t.Parallel()

	examples := addOrUpdateEventInputExamples()
	seen := map[string]bool{}
	for _, example := range examples {
		category, _ := example["category"].(string)
		if category != "RACE_A" && category != "RACE_B" && category != "RACE_C" {
			continue
		}
		seen[category] = true
		for _, field := range []string{"date", "type", "name", "distance_meters", "target_load"} {
			if _, ok := example[field]; !ok {
				t.Fatalf("%s example missing %s: %#v", category, field, example)
			}
		}
		if _, ok := example["moving_time_seconds"]; !ok {
			if _, ok := example["elapsed_time_seconds"]; !ok {
				t.Fatalf("%s example missing expected duration: %#v", category, example)
			}
		}
	}
	for _, category := range []string{"RACE_A", "RACE_B", "RACE_C"} {
		if !seen[category] {
			t.Fatalf("missing %s race input example in %#v", category, examples)
		}
	}
}

func TestStrengthDescriptionExamplesAreVisibleInToolCatalog(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		schema map[string]any
	}{
		{name: addOrUpdateEventName, schema: addOrUpdateEventInputSchema()},
		{name: updateActivityName, schema: updateActivityInputSchema()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			examples := schemaCatalogExamples(tc.schema)
			found := false
			for _, raw := range examples {
				example, ok := raw.(map[string]any)
				if !ok {
					t.Fatalf("catalog example type = %T", raw)
				}
				description, _ := example["description"].(string)
				if !strings.Contains(description, "- Back squat:") {
					continue
				}
				found = true
				if !strings.Contains(description, "\n\n- Back squat:") || !strings.Contains(description, "\n- Split squat:") || strings.Contains(description, "•") {
					t.Fatalf("strength example description = %q, want Markdown list with one exercise per line", description)
				}
				if err := validateExampleAgainstSchema(tc.schema, jsonRoundTripExample(t, example), tc.name); err != nil {
					t.Fatalf("strength example invalid: %v", err)
				}
				if tc.name == addOrUpdateEventName {
					if example["type"] != "WeightTraining" || example["category"] != "WORKOUT" {
						t.Fatalf("strength event example = %#v, want WeightTraining WORKOUT", example)
					}
					if _, ok := example["workout_doc"]; ok {
						t.Fatalf("strength event example must not claim structured workout_doc support: %#v", example)
					}
				}
			}
			if !found {
				t.Fatalf("%s catalog examples missing visible Markdown strength session: %#v", tc.name, examples)
			}
		})
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
	for _, name := range []string{"date", "start_time", "end_date_local", "event_id", "external_id", "category", "type", "name", "description", "workout_doc", "tags", "color", "not_on_fitness_chart", "indoor", "hide_from_athlete", "athlete_cannot_edit", "target_load", "distance_meters", "moving_time_seconds", "elapsed_time_seconds"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
}

func TestAddOrUpdateEventCalendarFieldsAreSparseAndReversible(t *testing.T) {
	t.Parallel()
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":123,"category":"NOTE","name":"Camp","start_date_local":"2026-07-10T00:00:00","end_date_local":"2026-07-12T16:00:00","hide_from_athlete":false,"athlete_cannot_edit":false}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	for _, tc := range []struct {
		name string
		args string
		end  *string
		hide *bool
		lock *bool
	}{
		{name: "set", args: `{"event_id":"123","date":"2026-07-10","category":"NOTE","end_date_local":"2026-07-12T16:00:00","hide_from_athlete":true,"athlete_cannot_edit":true}`, end: calendarStringPtr("2026-07-12T16:00:00"), hide: calendarBoolPtr(true), lock: calendarBoolPtr(true)},
		{name: "reverse", args: `{"event_id":"123","date":"2026-07-10","category":"NOTE","hide_from_athlete":false,"athlete_cannot_edit":false}`, hide: calendarBoolPtr(false), lock: calendarBoolPtr(false)},
		{name: "omit", args: `{"event_id":"123","date":"2026-07-10","category":"NOTE"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatal(err)
			}
			call := client.calls[len(client.calls)-1]
			if !reflect.DeepEqual(call.EndDateLocal, tc.end) || !reflect.DeepEqual(call.HideFromAthlete, tc.hide) || !reflect.DeepEqual(call.AthleteCannotEdit, tc.lock) {
				t.Fatalf("params = %+v, want end=%v hide=%v lock=%v", call, tc.end, tc.hide, tc.lock)
			}
			row := resultMap(t, result)["event"].(map[string]any)
			if row["event_id"] != "123" || row["end_date_local"] != "2026-07-12T16:00:00" || row["hide_from_athlete"] != false || row["athlete_cannot_edit"] != false {
				t.Fatalf("row = %#v, want numeric ID, span, explicit false flags", row)
			}
		})
	}
}

func TestAddOrUpdateEventRejectsInvalidEndDateBeforeWrite(t *testing.T) {
	t.Parallel()
	client := &fakeEventWriterClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}}}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	for _, tc := range []struct{ name, end string }{
		{name: "before start", end: "2026-07-09"},
		{name: "invalid date", end: "2026-02-30"},
		{name: "utc timestamp", end: "2026-07-12T16:00:00Z"},
		{name: "offset timestamp", end: "2026-07-12T16:00:00-03:00"},
		{name: "blank", end: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args, _ := json.Marshal(map[string]any{"date": "2026-07-10", "category": "NOTE", "name": "Camp", "end_date_local": tc.end})
			if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: args}); err == nil {
				t.Fatalf("end %q accepted", tc.end)
			}
		})
	}
	if len(client.calls) != 0 {
		t.Fatalf("writes = %#v, want none", client.calls)
	}
}

func TestAddOrUpdateEventMultiDayCreateRetryMatchesNormalizedEndAndFlags(t *testing.T) {
	t.Parallel()
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
		events:            decodeToolEvents(t, `{"id":123,"category":"RACE_A","type":"Ride","name":"Stage race","start_date_local":"2026-07-10T00:00:00","end_date_local":"2026-07-12T00:00:00","hide_from_athlete":true,"athlete_cannot_edit":false}`),
		event:             decodeToolEvents(t, `{"id":124,"category":"RACE_A","type":"Ride","name":"Stage race","start_date_local":"2026-07-10T00:00:00","end_date_local":"2026-07-13T00:00:00","hide_from_athlete":true,"athlete_cannot_edit":false}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	request := json.RawMessage(`{"date":"2026-07-10","end_date_local":"2026-07-12","category":"RACE_A","type":"Ride","name":"Stage race","hide_from_athlete":true,"athlete_cannot_edit":false}`)
	for attempt := 0; attempt < 2; attempt++ {
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: request})
		if err != nil {
			t.Fatal(err)
		}
		if got := resultMap(t, result)["_meta"].(map[string]any)["confirmation_status"]; got != skippedExistingEventStatus {
			t.Fatalf("retry %d confirmation = %#v, want skipped duplicate", attempt, got)
		}
	}
	if len(client.calls) != 0 {
		t.Fatalf("writes = %#v, want normalized multi-day duplicate skipped", client.calls)
	}
	request = json.RawMessage(`{"date":"2026-07-10","end_date_local":"2026-07-13","category":"RACE_A","type":"Ride","name":"Stage race","hide_from_athlete":true,"athlete_cannot_edit":false}`)
	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: request}); err != nil {
		t.Fatal(err)
	}
	if len(client.calls) != 1 {
		t.Fatalf("writes = %#v, want changed end date to create separate event", client.calls)
	}
}

func TestAddOrUpdateEventWritesAbsolutePowerAsWatts(t *testing.T) {
	t.Parallel()
	client := &fakeEventWriterClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
		event:             decodeToolEvents(t, `{"id":123,"category":"WORKOUT","type":"Ride","start_date_local":"2026-07-10T00:00:00","workout_doc":{"steps":[{"duration":300,"power":{"value":220,"units":"WATTS"}}]}}`)[0],
	}
	tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-07-10","category":"WORKOUT","type":"Ride","workout_doc":{"steps":[{"duration":300,"power":{"value":220,"units":"WATTS"}}]}}`)})
	if err != nil {
		t.Fatal(err)
	}
	if len(client.calls) != 1 || client.calls[0].Description == nil || !strings.Contains(*client.calls[0].Description, "220w") || strings.Contains(*client.calls[0].Description, "%") {
		t.Fatalf("workout DSL = %#v, want absolute 220w with no percent", client.calls)
	}
	row := resultMap(t, result)["event"].(map[string]any)
	if summary := row["workout_doc_summary"].(map[string]any); summary["target_previews"] != nil {
		t.Fatalf("absolute watt step should not be projected as FTP percentage: %#v", summary)
	}
}

func calendarStringPtr(value string) *string { return &value }
func calendarBoolPtr(value bool) *bool       { return &value }

func TestAddOrUpdateEventStartTimeAndDuplicateDetection(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		startTime     *string
		existingStart string
		wantWrites    int
	}{
		{name: "timed create", startTime: calendarStringPtr("07:30"), wantWrites: 1},
		{name: "same time retry", startTime: calendarStringPtr("07:30"), existingStart: "2026-07-10T07:30:00"},
		{name: "equivalent fractional timestamp retry", startTime: calendarStringPtr("07:30"), existingStart: "2026-07-10T07:30:00.000"},
		{name: "different time is separate", startTime: calendarStringPtr("17:30"), existingStart: "2026-07-10T07:30:00", wantWrites: 1},
		{name: "date only is separate from timed", existingStart: "2026-07-10T07:30:00", wantWrites: 1},
		{name: "date only retry", existingStart: "2026-07-10"},
		{name: "explicit midnight retry", startTime: calendarStringPtr("00:00"), existingStart: "2026-07-10T00:00:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantStart := "2026-07-10T00:00:00"
			if tc.startTime != nil {
				wantStart = "2026-07-10T" + *tc.startTime + ":00"
			}
			client := &fakeEventWriterClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
				event:             decodeToolEvents(t, `{"id":"new","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"`+wantStart+`"}`)[0],
			}
			if tc.existingStart != "" {
				client.events = decodeToolEvents(t, `{"id":"existing","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"`+tc.existingStart+`"}`)
			}
			args := map[string]any{"date": "2026-07-10", "category": "WORKOUT", "type": "Ride", "name": "Tempo"}
			if tc.startTime != nil {
				args["start_time"] = *tc.startTime
			}
			raw, _ := json.Marshal(args)
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: raw})
			if err != nil {
				t.Fatal(err)
			}
			if len(client.calls) != tc.wantWrites {
				t.Fatalf("writes = %d, want %d", len(client.calls), tc.wantWrites)
			}
			if tc.wantWrites == 1 && !reflect.DeepEqual(client.calls[0].StartTime, tc.startTime) {
				t.Fatalf("start_time = %#v, want %#v", client.calls[0].StartTime, tc.startTime)
			}
			if len(client.listCalls) != 1 || client.listCalls[0].Oldest != "2026-07-10" || client.listCalls[0].Newest != "2026-07-10" {
				t.Fatalf("preflight = %#v, want date-only bounds", client.listCalls)
			}
			out := resultMap(t, result)
			if tc.wantWrites == 1 && out["event"].(map[string]any)["start_date_local"] != wantStart {
				t.Fatalf("event = %#v, want start %s", out["event"], wantStart)
			}
			if out["_meta"].(map[string]any)["timezone"] != "America/Sao_Paulo" {
				t.Fatalf("meta = %#v, want athlete timezone", out["_meta"])
			}
		})
	}
}

func TestAddOrUpdateEventStartTimeValidation(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		time any
		end  string
		ok   bool
	}{
		{name: "start before end", time: "07:30", end: "2026-07-10T08:30:00", ok: true},
		{name: "end before start", time: "07:30", end: "2026-07-10T07:29:59"},
		{name: "empty", time: ""},
		{name: "unpadded hour", time: "7:30"},
		{name: "hour out of range", time: "24:00"},
		{name: "minute out of range", time: "07:60"},
		{name: "seconds", time: "07:30:00"},
		{name: "offset", time: "07:30-03:00"},
		{name: "UTC suffix", time: "07:30Z"},
		{name: "numeric", time: 730},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := &fakeEventWriterClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
				event:             decodeToolEvents(t, `{"id":"1","category":"NOTE","start_date_local":"2026-07-10T07:30:00"}`)[0],
			}
			args := map[string]any{"date": "2026-07-10", "start_time": tc.time, "category": "NOTE", "name": "Reminder"}
			if tc.end != "" {
				args["end_date_local"] = tc.end
			}
			raw, _ := json.Marshal(args)
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: raw})
			if (err == nil) != tc.ok {
				t.Fatalf("error = %v, want valid = %t", err, tc.ok)
			}
			if !tc.ok && (len(client.calls) != 0 || len(client.listCalls) != 0) {
				t.Fatalf("invalid timing caused I/O: writes %#v, reads %#v", client.calls, client.listCalls)
			}
		})
	}
}

func TestAddOrUpdateEventTimingErrorsAreActionable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		cause   error
		message string
	}{
		{name: "unavailable timing", cause: intervals.ErrEventTimingUnavailable, message: "could not read existing event timing; retry or provide start_time and end_date_local explicitly"},
		{name: "end before start", cause: intervals.ErrEventEndBeforeStart, message: "event end would precede its start; provide end_date_local on or after the new start"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := &fakeEventWriterClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
				err:               fmt.Errorf("internal upstream detail: %w", tc.cause),
			}
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"123","date":"2026-07-10","category":"NOTE"}`)})
			if err == nil || err.Error() != tc.message || !errors.Is(err, tc.cause) {
				t.Fatalf("error = %v, want public message %q with retained cause", err, tc.message)
			}
		})
	}
}

func TestAddOrUpdateEventTimedCreateRecoveryRequiresMatchingTime(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		start string
		ok    bool
	}{
		{name: "same time", start: "2026-07-10T07:30:00", ok: true},
		{name: "different time", start: "2026-07-10T17:30:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			client := &interruptedEventWriter{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
				listResults:       [][]intervals.Event{nil, decodeToolEvents(t, `{"id":"committed","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"`+tc.start+`"}`)},
			}
			tool := newAddOrUpdateEventTool(client, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"date":"2026-07-10","start_time":"07:30","category":"WORKOUT","type":"Ride","name":"Tempo"}`)})
			if (err == nil) != tc.ok {
				t.Fatalf("error = %v, want recovery = %t", err, tc.ok)
			}
			if len(client.writeCalls) != 1 || len(client.listCalls) != 2 {
				t.Fatalf("writes/reads = %d/%d, want 1/2", len(client.writeCalls), len(client.listCalls))
			}
			if tc.ok && resultMap(t, result)["_meta"].(map[string]any)["confirmation_status"] != verifiedAfterInterruptedWriteStatus {
				t.Fatal("matching time must return verified recovery status")
			}
		})
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
