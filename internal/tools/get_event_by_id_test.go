package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetEventByIDDetailSuccessUsesEventEnvelope(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		eventDetail:       decodeToolEvents(t, `{"id":123,"name":"Tempo","category":"WORKOUT","start_date_local":"2026-01-03","updated":"2026-01-03T12:00:00Z","tags":["detail","tempo"]}`)[0],
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
	tags := row["tags"].([]any)
	if len(tags) != 2 || tags[0] != "detail" || tags[1] != "tempo" {
		t.Fatalf("tags = %#v, want detail tags", tags)
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
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		eventDetailErr:    fmt.Errorf("detail: %w", intervals.ErrNotFound),
		events:            decodeToolEvents(t, `{"id":"target","name":"Recovered","category":"WORKOUT","start_date_local":"2026-03-15","tags":[]}`),
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
	if tags, ok := row["tags"].([]any); !ok || len(tags) != 0 {
		t.Fatalf("recovered tags = %#v, want explicit empty array", row["tags"])
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

func TestGetEventByIDNoHintFallbackUsesAthleteLocalTodayWhenUTCDateDiffers(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		eventDetailErr:    fmt.Errorf("detail: %w", intervals.ErrNotFound),
		events:            decodeToolEvents(t, `{"id":"target","name":"Recovered local event","category":"WORKOUT","start_date_local":"2026-05-24T23:30:00"}`),
	}
	tool := newGetEventByIDToolWithClock(client, client, "test", "UTC", false, fixedNow("2026-05-25T02:30:00Z"))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"target"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.listCalls) != 1 {
		t.Fatalf("list calls = %d, want fallback scan", len(client.listCalls))
	}
	call := client.listCalls[0]
	if call.Oldest != "2026-04-24" || call.Newest != "2026-06-23" || call.Limit != fallbackEventByIDLimit || call.Resolve == nil || !*call.Resolve {
		t.Fatalf("fallback ListEvents params = %#v, want ±30 days around athlete-local 2026-05-24", call)
	}
	out := resultMap(t, result)
	row := out["event"].(map[string]any)
	if row["event_id"] != "target" || row["start_date_local"] != "2026-05-24T23:30:00" {
		t.Fatalf("event row = %#v, want recovered athlete-local event", row)
	}
	meta := out["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" || meta["source"] != "list_scan" || meta["recovered"] != true {
		t.Fatalf("meta = %#v, want local timezone list-scan recovery", meta)
	}
	scanned := meta["scanned_range"].(map[string]any)
	if scanned["oldest"] != "2026-04-24" || scanned["newest"] != "2026-06-23" {
		t.Fatalf("scanned_range = %#v, want athlete-local no-hint window", scanned)
	}
}

func TestGetEventByIDMissReturnsStructuredUnavailable(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
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

func TestGetEventByIDListedFixtureDetail404ReturnsInconsistencyWhenRescanMisses(t *testing.T) {
	t.Parallel()

	listedEvents := loadEventListFixture(t, "../intervals/testdata/events/inconsistent/synthetic_list.json")
	if len(listedEvents) != 1 {
		t.Fatalf("listed fixture events = %d, want one synthetic listed event", len(listedEvents))
	}
	listedID := normalizedEventID(listedEvents[0])
	detailNote, err := os.ReadFile("../intervals/testdata/events/inconsistent/synthetic_detail_404.txt")
	if err != nil {
		t.Fatalf("read detail 404 note: %v", err)
	}
	if !strings.Contains(string(detailNote), listedID) || !strings.Contains(string(detailNote), "404 Not Found") {
		t.Fatalf("detail note = %q, want listed ID and 404 marker", detailNote)
	}
	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		eventDetailErr:    fmt.Errorf("detail: %w", intervals.ErrNotFound),
		events:            nil,
	}
	tool := newGetEventByIDToolWithClock(client, client, "test", "UTC", false, fixedNow("2026-05-01T12:00:00Z"))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"event_id":"` + listedID + `","oldest":"2026-03-01","newest":"2026-03-31","resolve":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v, want structured upstream_inconsistency", err)
	}
	out := resultMap(t, result)
	if _, ok := out["event"]; ok {
		t.Fatalf("event present on fixture mismatch: %#v", out["event"])
	}
	unavailable := out["unavailable"].(map[string]any)
	if unavailable["reason"] != "upstream_inconsistency" {
		t.Fatalf("unavailable = %#v, want upstream_inconsistency", unavailable)
	}
	retried := unavailable["retried"].([]any)
	if len(retried) != 2 || retried[0] != "detail" || retried[1] != "list_scan" {
		t.Fatalf("retried = %#v, want detail/list_scan", retried)
	}
	call := client.listCalls[0]
	if call.Oldest != "2026-03-01" || call.Newest != "2026-03-31" || call.Resolve == nil || !*call.Resolve {
		t.Fatalf("fallback list call = %#v, want fixture date range with resolve=true", call)
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

func loadEventListFixture(t *testing.T, path string) []intervals.Event {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read event fixture %s: %v", path, err)
	}
	var events []intervals.Event
	if err := json.Unmarshal(data, &events); err != nil {
		t.Fatalf("decode event fixture %s: %v", path, err)
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
