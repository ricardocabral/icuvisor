package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeLinkActivityToEventClient struct {
	fakeProfileClient
	activity intervals.Activity
	event    intervals.Event
	linked   intervals.Activity
	calls    []intervals.LinkActivityToEventParams
	linkErr  error
	actErr   error
	eventErr error
}

func (f *fakeLinkActivityToEventClient) LinkActivityToEvent(ctx context.Context, params intervals.LinkActivityToEventParams) (intervals.Activity, error) {
	f.calls = append(f.calls, params)
	return f.linked, f.linkErr
}

func (f *fakeLinkActivityToEventClient) GetActivity(ctx context.Context, activityID string) (intervals.Activity, error) {
	return f.activity, f.actErr
}

func (f *fakeLinkActivityToEventClient) GetEvent(ctx context.Context, eventID string) (intervals.Event, error) {
	return f.event, f.eventErr
}

func (f *fakeLinkActivityToEventClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	return nil, nil
}

func TestLinkActivityToEventSuccessNoWarning(t *testing.T) {
	t.Parallel()

	client := &fakeLinkActivityToEventClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", Timezone: "UTC"}},
		activity:          decodeActivityFixture(t, `{"id":"a1","start_date_local":"2026-05-10T07:00:00"}`),
		event:             decodeToolEvents(t, `{"id":1001,"start_date_local":"2026-05-10"}`)[0],
		linked:            decodeActivityFixture(t, `{"id":"a1","paired_event_id":1001,"start_date_local":"2026-05-10T07:00:00","extra":null}`),
	}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":" a1 ","event_id":" 1001 ","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.calls) != 1 || client.calls[0].ActivityID != "a1" || client.calls[0].EventID != "1001" {
		t.Fatalf("link calls = %#v, want trimmed IDs", client.calls)
	}
	out := resultMap(t, result)
	if out["activity_id"] != "a1" || out["event_id"] != "1001" || out["status"] != "linked" {
		t.Fatalf("response = %#v, want linked IDs/status", out)
	}
	meta := out["_meta"].(map[string]any)
	if _, ok := meta["warnings"]; ok {
		t.Fatalf("warnings present for same date: %#v", meta["warnings"])
	}
	full := out["full"].(map[string]any)
	if full["paired_event_id"] != float64(1001) {
		t.Fatalf("full = %#v, want raw linked activity", full)
	}
}

func TestLinkActivityToEventIdempotentRelink(t *testing.T) {
	t.Parallel()

	client := &fakeLinkActivityToEventClient{
		activity: decodeActivityFixture(t, `{"id":"a1","start_date_local":"2026-05-10T07:00:00"}`),
		event:    decodeToolEvents(t, `{"id":1001,"start_date_local":"2026-05-10"}`)[0],
		linked:   decodeActivityFixture(t, `{"id":"a1","paired_event_id":1001}`),
	}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)
	for range 2 {
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","event_id":"1001"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		if resultMap(t, result)["status"] != "linked" {
			t.Fatalf("result = %#v, want linked status", resultMap(t, result))
		}
	}
	if len(client.calls) != 2 || client.calls[0] != client.calls[1] {
		t.Fatalf("calls = %#v, want same idempotent relink call", client.calls)
	}
}

func TestLinkActivityToEventMismatchedDateWarning(t *testing.T) {
	t.Parallel()

	client := &fakeLinkActivityToEventClient{
		activity: decodeActivityFixture(t, `{"id":"a1","start_date_local":"2026-05-10T07:00:00"}`),
		event:    decodeToolEvents(t, `{"id":1001,"start_date_local":"2026-05-11"}`)[0],
		linked:   decodeActivityFixture(t, `{"id":"a1","paired_event_id":1001}`),
	}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","event_id":"1001"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	warnings := resultMap(t, result)["_meta"].(map[string]any)["warnings"].([]any)
	warning := warnings[0].(map[string]any)
	if warning["code"] != "date_mismatch" || warning["activity_date"] != "2026-05-10" || warning["event_date"] != "2026-05-11" {
		t.Fatalf("warning = %#v, want date_mismatch with dates", warning)
	}
}

func TestLinkActivityToEventValidationAndErrors(t *testing.T) {
	t.Parallel()

	client := &fakeLinkActivityToEventClient{linkErr: errors.New("upstream detail")}
	tool := newLinkActivityToEventTool(client, client, client, "test", false)
	for _, raw := range []string{
		`{"activity_id":"","event_id":"1001"}`,
		`{"activity_id":"a1","event_id":""}`,
		`{"activity_id":"a1","event_id":"evt-1"}`,
		`{"activity_id":"a1","event_id":"1001","confirm":true}`,
	} {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"a1","event_id":"1001"}`)})
	if message, ok := PublicErrorMessage(err); !ok || message != linkActivityToEventMessage {
		t.Fatalf("PublicErrorMessage = %q, %v; err = %v", message, ok, err)
	}
}

func TestLinkActivityToEventRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeLinkActivityToEventClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", Timezone: "UTC"}}}
	registrar := &collectingRegistrar{}
	if err := NewRegistry(client, "test", "UTC").Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	tool := findTool(t, registrar.tools, linkActivityToEventName)
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if !strings.Contains(tool.Description, "auto-pairing misses") || strings.Contains(strings.ToLower(tool.Description), "confirm") {
		t.Fatalf("description = %q, want forum #97/manual escape hatch language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"activity_id", "event_id", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	if _, ok := props["confirm"]; ok {
		t.Fatalf("schema includes forbidden confirm property")
	}
}
