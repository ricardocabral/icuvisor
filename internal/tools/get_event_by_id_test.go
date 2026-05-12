package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetEventByIDDetailSuccessUsesEventEnvelope(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		eventDetail:       decodeToolEvents(t, `{"id":123,"name":"Tempo","category":"WORKOUT","start_date_local":"2026-01-03","updated":"2026-01-03T12:00:00Z"}`)[0],
	}
	tool := newGetEventByIDToolWithClock(client, client, "test", "UTC", false, fixedNow("2026-05-01T12:00:00Z"))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"123"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.listCalls) != 0 {
		t.Fatalf("list calls = %d, want no fallback on detail success", len(client.listCalls))
	}
	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "123" || row["updated_local"] != "2026-01-03T09:00:00-03:00" {
		t.Fatalf("event row = %#v, want detail row with local timestamp", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["source"] != "detail" || meta["recovered"] != false || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("meta = %#v, want detail metadata", meta)
	}
	if _, ok := out["unavailable"]; ok {
		t.Fatalf("unavailable present on detail success: %#v", out["unavailable"])
	}
}

func TestGetEventByIDFallbackScansDateWindowWithResolveAndCap(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		eventDetailErr:    fmt.Errorf("detail: %w", intervals.ErrNotFound),
		events:            decodeToolEvents(t, `{"id":"target","name":"Recovered","category":"WORKOUT","start_date_local":"2026-03-15"}`),
	}
	tool := newGetEventByIDToolWithClock(client, client, "test", "UTC", false, fixedNow("2026-05-01T12:00:00Z"))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"target","date":"2026-03-15"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.detailCalls) != 1 || client.detailCalls[0] != "target" {
		t.Fatalf("detail calls = %#v, want target", client.detailCalls)
	}
	if len(client.listCalls) != 1 {
		t.Fatalf("list calls = %d, want fallback scan", len(client.listCalls))
	}
	call := client.listCalls[0]
	if call.Oldest != "2026-02-13" || call.Newest != "2026-04-14" || call.Limit != fallbackEventByIDLimit || call.Resolve == nil || !*call.Resolve {
		t.Fatalf("fallback ListEvents params = %#v, want date window, cap, resolve=true", call)
	}
	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "target" || row["name"] != "Recovered" {
		t.Fatalf("event row = %#v, want recovered target", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["source"] != "list_scan" || meta["recovered"] != true || meta["limit"] != float64(fallbackEventByIDLimit) || meta["count"] != float64(1) {
		t.Fatalf("meta = %#v, want recovered list_scan metadata", meta)
	}
	scanned := meta["scanned_range"].(map[string]any)
	if scanned["oldest"] != "2026-02-13" || scanned["newest"] != "2026-04-14" {
		t.Fatalf("scanned_range = %#v, want date-derived window", scanned)
	}
}

func TestGetEventByIDMissReturnsStructuredUnavailable(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "12345", PreferredUnits: "metric", Timezone: "UTC"}},
		eventDetailErr:    fmt.Errorf("detail: %w", intervals.ErrNotFound),
		events:            manyToolEvents(t, fallbackEventByIDLimit+1),
	}
	tool := newGetEventByIDToolWithClock(client, client, "test", "UTC", false, fixedNow("2026-05-01T12:00:00Z"))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"missing","oldest":"2026-01-01","newest":"2026-01-31","resolve":false}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v, want structured non-error result", err)
	}
	call := client.listCalls[0]
	if call.Resolve == nil || *call.Resolve {
		t.Fatalf("fallback resolve = %#v, want explicit false honored", call.Resolve)
	}
	out := resultMap(t, result)
	if _, ok := out["event"]; ok {
		t.Fatalf("event present on miss: %#v", out["event"])
	}
	unavailable := out["unavailable"].(map[string]any)
	if unavailable["reason"] != "upstream_inconsistency" {
		t.Fatalf("unavailable = %#v, want upstream_inconsistency", unavailable)
	}
	retried := unavailable["retried"].([]any)
	if len(retried) != 2 || retried[0] != "detail" || retried[1] != "list_scan" {
		t.Fatalf("retried = %#v, want detail/list_scan", retried)
	}
	meta := out["_meta"].(map[string]any)
	if meta["count"] != float64(fallbackEventByIDLimit) || meta["truncated"] != true {
		t.Fatalf("meta = %#v, want capped count and truncated=true", meta)
	}
	scanned := meta["scanned_range"].(map[string]any)
	if scanned["oldest"] != "2026-01-01" || scanned["newest"] != "2026-01-31" {
		t.Fatalf("scanned_range = %#v, want supplied bounds", scanned)
	}
}

func manyToolEvents(t *testing.T, count int) []intervals.Event {
	t.Helper()
	events := make([]intervals.Event, 0, count)
	for i := range count {
		events = append(events, decodeToolEvents(t, fmt.Sprintf(`{"id":"evt-%03d","category":"WORKOUT","start_date_local":"2026-01-01"}`, i))...)
	}
	return events
}

func fixedNow(value string) func() time.Time {
	return func() time.Time {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			panic(err)
		}
		return parsed
	}
}
