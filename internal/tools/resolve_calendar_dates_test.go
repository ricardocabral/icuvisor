package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func TestResolveCalendarDatesRegistrationMetadata(t *testing.T) {
	t.Parallel()

	tool := newResolveCalendarDatesToolWithClock(&fakeProfileClient{}, "test", "UTC", false, fixedCalendarClock(time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC)))
	if tool.Name != resolveCalendarDatesName || !strings.Contains(tool.Description, "athlete-local calendar anchors") {
		t.Fatalf("tool metadata = %#v, want resolve_calendar_dates anchor description", tool)
	}
	if tool.EffectiveToolset() != safety.ToolsetCore {
		t.Fatalf("toolset = %q, want core", tool.EffectiveToolset())
	}
	schema := tool.InputSchema.(map[string]any)
	if schema["additionalProperties"] != false {
		t.Fatalf("additionalProperties = %#v, want false", schema["additionalProperties"])
	}
	props := schema["properties"].(map[string]any)
	if _, ok := props["base_date"]; !ok {
		t.Fatalf("schema missing base_date: %#v", props)
	}
	offsets := props["offsets"].(map[string]any)
	if offsets["maxItems"] != maxCalendarOffsets || offsets["uniqueItems"] != true {
		t.Fatalf("offsets schema = %#v, want maxItems/uniqueItems", offsets)
	}
}

func TestResolveCalendarDatesDefaultsBaseDateFromAthleteTimezone(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}}
	tool := newResolveCalendarDatesToolWithClock(client, "test", "UTC", false, fixedCalendarClock(time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC)))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	dates := out["dates"].([]any)
	if len(dates) != 1 {
		t.Fatalf("dates = %#v, want one default anchor", dates)
	}
	anchor := dates[0].(map[string]any)
	if anchor["offset_days"] != float64(0) || anchor["date"] != "2026-05-24" || anchor["weekday"] != "Sunday" {
		t.Fatalf("anchor = %#v, want athlete-local Sunday 2026-05-24", anchor)
	}
	meta := out["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" || meta["base_date"] != "2026-05-24" || meta["base_weekday"] != "Sunday" || meta["server_version"] != "test" || meta["count"] != float64(1) {
		t.Fatalf("meta = %#v, want local base metadata", meta)
	}
}

func TestResolveCalendarDatesAppliesOffsetsInLocalCalendar(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/New_York"}}
	tool := newResolveCalendarDatesToolWithClock(client, "test", "UTC", false, fixedCalendarClock(time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC)))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"base_date":"2026-03-07","offsets":[0,1,2,7,-1]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)["dates"].([]any)
	want := []struct {
		offset  float64
		date    string
		weekday string
	}{
		{0, "2026-03-07", "Saturday"},
		{1, "2026-03-08", "Sunday"},
		{2, "2026-03-09", "Monday"},
		{7, "2026-03-14", "Saturday"},
		{-1, "2026-03-06", "Friday"},
	}
	if len(got) != len(want) {
		t.Fatalf("dates = %#v, want %d rows", got, len(want))
	}
	for i, wantRow := range want {
		row := got[i].(map[string]any)
		if row["offset_days"] != wantRow.offset || row["date"] != wantRow.date || row["weekday"] != wantRow.weekday {
			t.Fatalf("row %d = %#v, want %#v", i, row, wantRow)
		}
	}
}

func TestResolveCalendarDatesRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "UTC"}}
	tool := newResolveCalendarDatesToolWithClock(client, "test", "UTC", false, fixedCalendarClock(time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC)))
	tests := []struct {
		name string
		args string
	}{
		{name: "non object", args: `[]`},
		{name: "unknown field", args: `{"date":"2026-05-24"}`},
		{name: "bad base date", args: `{"base_date":"2026-02-30"}`},
		{name: "duplicate offsets", args: `{"offsets":[0,1,1]}`},
		{name: "offset too large", args: `{"offsets":[367]}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatalf("Handler() nil error, want invalid input for %s", tc.args)
			}
			if !strings.Contains(err.Error(), "invalid resolve_calendar_dates arguments") {
				t.Fatalf("error = %v, want invalid arguments", err)
			}
		})
	}
}

func TestResolveCalendarDatesReportsInvalidTimezone(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{Timezone: "Mars/Olympus_Mons"}}
	tool := newResolveCalendarDatesToolWithClock(client, "test", "UTC", false, fixedCalendarClock(time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC)))

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want invalid timezone")
	}
	if !strings.Contains(err.Error(), fetchResolveCalendarDatesMessage) {
		t.Fatalf("error = %v, want timezone guidance", err)
	}
	if strings.Contains(err.Error(), invalidResolveCalendarDatesMessage) {
		t.Fatalf("error = %v, want timezone guidance instead of invalid arguments", err)
	}
}

func fixedCalendarClock(now time.Time) func() time.Time {
	return func() time.Time { return now }
}
