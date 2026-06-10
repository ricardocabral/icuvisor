package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakeUnavailableDateRangeClient struct {
	fakeProfileClient
	events     []intervals.Event
	created    []intervals.Event
	calls      []intervals.WriteEventParams
	listCalls  []intervals.ListEventsParams
	writeError error
}

func (f *fakeUnavailableDateRangeClient) AddOrUpdateEvent(ctx context.Context, params intervals.WriteEventParams) (intervals.Event, error) {
	f.calls = append(f.calls, params)
	if f.writeError != nil {
		return intervals.Event{}, f.writeError
	}
	if len(f.created) == 0 {
		return intervals.Event{ID: "evt-created", Category: ptrString(params.Category), Type: ptrString(params.Type), Name: ptrString(params.Name), StartDateLocal: ptrString(params.Date), Description: params.Description, ExternalID: ptrString(params.ExternalID), Raw: map[string]any{"id": "evt-created", "category": params.Category, "type": params.Type, "name": params.Name, "start_date_local": params.Date, "external_id": params.ExternalID}}, nil
	}
	event := f.created[0]
	f.created = f.created[1:]
	return event, nil
}

func (f *fakeUnavailableDateRangeClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	return append([]intervals.Event(nil), f.events...), nil
}

func TestAddUnavailableDateRangeRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeUnavailableDateRangeClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newAddUnavailableDateRangeTool(client, client, "test", "UTC", false)

	if tool.Name != addUnavailableDateRangeName || tool.Requirement != RequirementWrite || tool.EffectiveToolset() != safety.ToolsetCore {
		t.Fatalf("tool metadata = %#v, want core write %s", tool, addUnavailableDateRangeName)
	}
	schema := tool.InputSchema.(map[string]any)
	if _, ok := schema["examples"]; !ok {
		t.Fatalf("schema missing examples: %#v", schema)
	}
	if _, ok := schema["input_examples"]; !ok {
		t.Fatalf("schema missing input_examples: %#v", schema)
	}
}

func TestAddUnavailableDateRangeCreatesInclusivePerDayEvents(t *testing.T) {
	t.Parallel()

	description := "Doctor advised no training."
	client := &fakeUnavailableDateRangeClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		created: decodeToolEvents(t,
			`{"id":"evt-1","category":"HOLIDAY","type":"Unavailable","name":"Holiday","start_date_local":"2026-07-01","description":"Doctor advised no training."}`,
			`{"id":"evt-2","category":"HOLIDAY","type":"Unavailable","name":"Holiday","start_date_local":"2026-07-02","description":"Doctor advised no training."}`,
			`{"id":"evt-3","category":"HOLIDAY","type":"Unavailable","name":"Holiday","start_date_local":"2026-07-03","description":"Doctor advised no training."}`,
		),
	}
	tool := newAddUnavailableDateRangeTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-07-01","end_date":"2026-07-03","category":"vacation","description":"Doctor advised no training."}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 3 {
		t.Fatalf("write calls = %#v, want 3 per-day creates", client.calls)
	}
	for i, call := range client.calls {
		wantDate := []string{"2026-07-01", "2026-07-02", "2026-07-03"}[i]
		if call.Date != wantDate || call.Category != "HOLIDAY" || call.Type != "Unavailable" || call.Name != "Holiday" {
			t.Fatalf("call[%d] = %#v, want holiday unavailable on %s", i, call, wantDate)
		}
		if call.Description == nil || *call.Description != description {
			t.Fatalf("call[%d].Description = %#v, want %q", i, call.Description, description)
		}
		if !strings.HasPrefix(call.ExternalID, "icuvisor-unavailable-v1-") {
			t.Fatalf("call[%d].ExternalID = %q, want generated unavailable id", i, call.ExternalID)
		}
	}
	if client.calls[0].ExternalID == client.calls[1].ExternalID || client.calls[1].ExternalID == client.calls[2].ExternalID {
		t.Fatalf("external IDs = %#v, want per-day idempotency keys", client.calls)
	}

	out := resultMap(t, result)
	if out["status"] != "created" {
		t.Fatalf("status = %#v, want created", out["status"])
	}
	events := out["events"].([]any)
	if len(events) != 3 {
		t.Fatalf("events = %#v, want 3 rows", events)
	}
	meta := out["_meta"].(map[string]any)
	if meta["operation"] != "create_range" || meta["category"] != "HOLIDAY" || meta["timezone"] != "America/Sao_Paulo" || meta["requested_days"] != float64(3) || meta["created_count"] != float64(3) || meta["skipped_count"] != float64(0) {
		t.Fatalf("meta = %#v, want created range counts", meta)
	}
}

func TestAddUnavailableDateRangeSkipsRepeatedRangeByGeneratedExternalID(t *testing.T) {
	t.Parallel()

	firstID := addUnavailableDateRangeExternalID("SICK", "2026-08-10", "Sick", "Flu")
	secondID := addUnavailableDateRangeExternalID("SICK", "2026-08-11", "Sick", "Flu")
	client := &fakeUnavailableDateRangeClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events: decodeToolEvents(t,
			`{"id":"evt-sick-1","external_id":"`+firstID+`","category":"SICK","type":"Unavailable","name":"Sick","start_date_local":"2026-08-10","description":"Flu"}`,
			`{"id":"evt-sick-2","external_id":"`+secondID+`","category":"SICK","type":"Unavailable","name":"Sick","start_date_local":"2026-08-11","description":"Flu"}`,
		),
	}
	tool := newAddUnavailableDateRangeTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-08-10","end_date":"2026-08-11","category":"sickness","description":"Flu"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want repeated range skipped", client.calls)
	}
	out := resultMap(t, result)
	if out["status"] != "skipped" {
		t.Fatalf("status = %#v, want skipped", out["status"])
	}
	meta := out["_meta"].(map[string]any)
	if meta["created_count"] != float64(0) || meta["skipped_count"] != float64(2) {
		t.Fatalf("meta = %#v, want all skipped", meta)
	}
}

func TestAddUnavailableDateRangeCreatesMissingDaysAndReportsConflicts(t *testing.T) {
	t.Parallel()

	existingID := addUnavailableDateRangeExternalID("INJURED", "2026-09-01", "Injured", "")
	client := &fakeUnavailableDateRangeClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events: decodeToolEvents(t,
			`{"id":"evt-injured","external_id":"`+existingID+`","category":"INJURED","type":"Unavailable","name":"Injured","start_date_local":"2026-09-01"}`,
			`{"id":"evt-workout","category":"WORKOUT","type":"Run","name":"Workout to review","start_date_local":"2026-09-02"}`,
		),
		created: decodeToolEvents(t, `{"id":"evt-created","category":"INJURED","type":"Unavailable","name":"Injured","start_date_local":"2026-09-02"}`),
	}
	tool := newAddUnavailableDateRangeTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-09-01","end_date":"2026-09-02","category":"injury"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].Date != "2026-09-02" || client.calls[0].Category != "INJURED" {
		t.Fatalf("write calls = %#v, want only missing injured day", client.calls)
	}
	out := resultMap(t, result)
	if out["status"] != "partial" {
		t.Fatalf("status = %#v, want partial", out["status"])
	}
	meta := out["_meta"].(map[string]any)
	if meta["created_count"] != float64(1) || meta["skipped_count"] != float64(1) || meta["conflict_count"] != float64(1) {
		t.Fatalf("meta = %#v, want mixed counts", meta)
	}
}

func TestAddUnavailableDateRangeRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	client := &fakeUnavailableDateRangeClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newAddUnavailableDateRangeTool(client, client, "test", "UTC", false)

	tests := []struct {
		name string
		args string
	}{
		{name: "unsupported category", args: `{"start_date":"2026-07-01","end_date":"2026-07-01","category":"NOTE"}`},
		{name: "broad travel alias rejected", args: `{"start_date":"2026-07-01","end_date":"2026-07-01","category":"travel"}`},
		{name: "reversed range", args: `{"start_date":"2026-07-03","end_date":"2026-07-01","category":"HOLIDAY"}`},
		{name: "excessive range", args: `{"start_date":"2026-07-01","end_date":"2026-08-01","category":"HOLIDAY"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatal("Handler() error = nil, want invalid input error")
			}
			if message, ok := PublicErrorMessage(err); !ok || !strings.Contains(message, "invalid add_unavailable_date_range arguments") {
				t.Fatalf("PublicErrorMessage(%v) = %q/%v, want invalid add_unavailable_date_range arguments", err, message, ok)
			}
		})
	}
	if len(client.calls) != 0 {
		t.Fatalf("write calls = %#v, want no writes for invalid inputs", client.calls)
	}
}
