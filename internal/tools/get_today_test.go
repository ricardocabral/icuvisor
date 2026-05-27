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

type fakeTodayClient struct {
	fakeProfileClient
	fitnessRows   []intervals.SummaryWithCats
	wellnessRows  []intervals.Wellness
	activities    []intervals.Activity
	events        []intervals.Event
	fitnessCalls  []intervals.AthleteSummaryParams
	wellnessCalls []intervals.WellnessParams
	activityCalls []intervals.ListActivitiesParams
	eventCalls    []intervals.ListEventsParams
}

func (f *fakeTodayClient) ListAthleteSummary(_ context.Context, params intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	f.fitnessCalls = append(f.fitnessCalls, params)
	return append([]intervals.SummaryWithCats(nil), f.fitnessRows...), nil
}

func (f *fakeTodayClient) ListWellness(_ context.Context, params intervals.WellnessParams) ([]intervals.Wellness, error) {
	f.wellnessCalls = append(f.wellnessCalls, params)
	return append([]intervals.Wellness(nil), f.wellnessRows...), nil
}

func (f *fakeTodayClient) ListActivities(_ context.Context, params intervals.ListActivitiesParams) ([]intervals.Activity, error) {
	f.activityCalls = append(f.activityCalls, params)
	return append([]intervals.Activity(nil), f.activities...), nil
}

func (f *fakeTodayClient) ListEvents(_ context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.eventCalls = append(f.eventCalls, params)
	return append([]intervals.Event(nil), f.events...), nil
}

func TestGetTodayRegistrationMetadata(t *testing.T) {
	t.Parallel()

	tool := newGetTodayToolWithClock(&fakeTodayClient{}, &fakeTodayClient{}, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())
	if tool.Name != getTodayName || !strings.Contains(tool.Description, "how's today looking") {
		t.Fatalf("tool metadata = %#v, want get_today daily digest description", tool)
	}
	if tool.EffectiveToolset() != safety.ToolsetCore {
		t.Fatalf("toolset = %q, want core", tool.EffectiveToolset())
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	if len(props) != 1 {
		t.Fatalf("schema properties = %#v, want only include_full", props)
	}
	if _, ok := props["include_full"]; !ok {
		t.Fatalf("schema missing include_full: %#v", props)
	}
}

func TestGetTodayDigestUsesAthleteLocalDateAndSourceShapes(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "imperial", Timezone: "America/Sao_Paulo"}},
		fitnessRows:       decodeSummaries(t, `[{"date":"2026-05-24","fitness":71.2345,"fatigue":80.2,"form":-8.9}]`),
		wellnessRows:      []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-24","sleepQuality":3,"feel":4,"weight":null}`)},
		activities: decodeActivityPage(t,
			`{"id":"a1","name":"Morning Run","type":"Run","start_date_local":"2026-05-24T07:30:00","distance":1609.344,"moving_time":480,"calories":120,"stream_types":["heartrate"]}`,
		),
		events: decodeToolEvents(t,
			`{"id":"11","category":"WORKOUT","type":"Run","name":"Easy run","start_date_local":"2026-05-24","icu_training_load":35}`,
			`{"id":"12","category":"NOTE","name":"Travel","description":"Pack race shoes","start_date_local":"2026-05-24"}`,
			`{"id":"13","category":"RACE_B","name":"Tune-up 5k","start_date_local":"2026-05-24"}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)

	assertTodayCalls(t, client, "2026-05-24")
	fitness := out["fitness"].([]any)[0].(map[string]any)
	if fitness["ctl"] != 71.235 || fitness["atl"] != 80.2 || fitness["tsb"] != -8.9 {
		t.Fatalf("fitness = %#v, want rounded CTL/ATL/TSB", fitness)
	}
	wellness := out["wellness"].([]any)[0].(map[string]any)
	if wellness["sleepQuality"] != float64(3) || wellness["feel"] != float64(4) {
		t.Fatalf("wellness = %#v, want source wellness fields", wellness)
	}
	if _, ok := wellness["weight"]; ok {
		t.Fatalf("wellness kept null weight in terse digest: %#v", wellness)
	}
	if scales := wellness["_meta"].(map[string]any)["scales"].(map[string]any); !strings.Contains(scales["sleepQuality"].(string), "1-4") || !strings.Contains(scales["feel"].(string), "1-5") {
		t.Fatalf("wellness scales = %#v, want source scale labels", scales)
	}
	activity := out["completed_activities"].([]any)[0].(map[string]any)
	if activity["distance_mi"] != 1.0 || activity["pace_seconds_per_mile"] != 480.0 {
		t.Fatalf("activity = %#v, want imperial unit-normalized source shape", activity)
	}
	if _, ok := activity["full"]; ok {
		t.Fatalf("activity included full in terse digest: %#v", activity)
	}
	planned := out["planned_events"].([]any)
	annotations := out["annotations"].([]any)
	if len(planned) != 1 || planned[0].(map[string]any)["category"] != "WORKOUT" {
		t.Fatalf("planned_events = %#v, want only workout event", planned)
	}
	if len(annotations) != 2 || annotations[0].(map[string]any)["category"] != "NOTE" || annotations[1].(map[string]any)["category"] != "RACE_B" {
		t.Fatalf("annotations = %#v, want NOTE and race categories", annotations)
	}
	meta := out["_meta"].(map[string]any)
	if meta["date"] != "2026-05-24" || meta["as_of"] != "2026-05-24T23:30:00-03:00" || meta["as_of_date"] != "2026-05-24" || meta["as_of_weekday"] != "Sunday" || meta["timezone"] != "America/Sao_Paulo" || meta["include_full"] != false {
		t.Fatalf("meta = %#v, want today/as-of/timezone/include_full", meta)
	}
	if counts := meta["section_counts"].(map[string]any); counts["completed_activities"] != float64(1) || counts["annotations"] != float64(2) {
		t.Fatalf("section_counts = %#v, want activity and annotation counts", counts)
	}
	if units := meta["units"].(map[string]any); units["distance"] != "mi" {
		t.Fatalf("units = %#v, want imperial metadata", units)
	}
}

func TestGetTodayIncludeFullWidensEverySection(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		fitnessRows:       decodeSummaries(t, `[{"date":"2026-05-24","fitness":71,"fatigue":80,"form":-9,"extra":"kept"}]`),
		wellnessRows:      []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-24","sleepQuality":3,"weight":null}`)},
		activities:        decodeActivityPage(t, `{"id":"a1","name":"Ride","type":"Ride","start_date_local":"2026-05-24T07:30:00","distance":1000,"raw_extra":"kept"}`),
		events:            decodeToolEvents(t, `{"id":"11","category":"WORKOUT","name":"Ride","start_date_local":"2026-05-24","raw_extra":"kept"}`),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	for section, rowIndex := range map[string]int{"fitness": 0, "wellness": 0, "completed_activities": 0, "planned_events": 0} {
		row := out[section].([]any)[rowIndex].(map[string]any)
		if _, ok := row["full"].(map[string]any); !ok {
			t.Fatalf("%s row missing full payload: %#v", section, row)
		}
	}
	wellness := out["wellness"].([]any)[0].(map[string]any)
	if _, ok := wellness["full"].(map[string]any)["weight"]; !ok {
		t.Fatalf("wellness full did not preserve raw null weight: %#v", wellness["full"])
	}
	if out["_meta"].(map[string]any)["include_full"] != true {
		t.Fatalf("_meta.include_full = %#v, want true", out["_meta"])
	}
}

func fixedTodayClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC) }
}

func assertTodayCalls(t *testing.T, client *fakeTodayClient, wantDate string) {
	t.Helper()
	if len(client.fitnessCalls) != 1 || client.fitnessCalls[0].Start != wantDate || client.fitnessCalls[0].End != wantDate {
		t.Fatalf("fitness calls = %#v, want single today range", client.fitnessCalls)
	}
	if len(client.wellnessCalls) != 1 || client.wellnessCalls[0].Oldest != wantDate || client.wellnessCalls[0].Newest != wantDate {
		t.Fatalf("wellness calls = %#v, want single today range", client.wellnessCalls)
	}
	if len(client.activityCalls) != 1 || client.activityCalls[0].Oldest != wantDate || client.activityCalls[0].Newest != "" {
		t.Fatalf("activity calls = %#v, want oldest today and upstream now newest", client.activityCalls)
	}
	if len(client.eventCalls) != 1 || client.eventCalls[0].Oldest != wantDate || client.eventCalls[0].Newest != wantDate || client.eventCalls[0].Limit != defaultTodayEventsLimit {
		t.Fatalf("event calls = %#v, want single today range", client.eventCalls)
	}
}
