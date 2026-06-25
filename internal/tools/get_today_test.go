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

func TestGetTodayFreshDigestUsesFixedAthleteLocalToday(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		fitnessRows:       decodeSummaries(t, `[{"date":"2026-05-23","fitness":70,"fatigue":79,"form":-9},{"date":"2026-05-24","fitness":72,"fatigue":81,"form":-9}]`),
		wellnessRows:      []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-24","feel":4}`)},
		activities: decodeActivityPage(t,
			`{"id":"yesterday-activity","name":"Stale Run","type":"Run","start_date_local":"2026-05-23T08:15:00","distance":5000}`,
			`{"id":"today-activity","name":"Morning Ride","type":"Ride","start_date_local":"2026-05-24T08:15:00","distance":30000}`,
		),
		events: decodeToolEvents(t,
			`{"id":"yesterday-event","category":"WORKOUT","name":"Stale Workout","start_date_local":"2026-05-23"}`,
			`{"id":"today-event","category":"WORKOUT","name":"Endurance","start_date_local":"2026-05-24"}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)

	assertTodayCalls(t, client, "2026-05-24")
	meta := out["_meta"].(map[string]any)
	if meta["date"] != "2026-05-24" || meta["as_of"] != "2026-05-24T23:30:00-03:00" || meta["as_of_date"] != "2026-05-24" || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("meta = %#v, want fixed athlete-local today", meta)
	}
	fitness := out["fitness"].([]any)
	if len(fitness) != 1 || fitness[0].(map[string]any)["date"] != "2026-05-24" {
		t.Fatalf("fitness = %#v, want only today row", fitness)
	}
	if got := out["wellness"].([]any)[0].(map[string]any)["date"]; got != "2026-05-24" {
		t.Fatalf("wellness date = %#v, want today", got)
	}
	activities := out["completed_activities"].([]any)
	if len(activities) != 1 || activities[0].(map[string]any)["activity_id"] != "today-activity" || !strings.HasPrefix(activities[0].(map[string]any)["start_date_local"].(string), "2026-05-24") {
		t.Fatalf("completed_activities = %#v, want only today row", activities)
	}
	planned := out["planned_events"].([]any)
	if len(planned) != 1 || planned[0].(map[string]any)["event_id"] != "today-event" || planned[0].(map[string]any)["start_date_local"] != "2026-05-24" {
		t.Fatalf("planned_events = %#v, want only today row", planned)
	}
}

func TestGetTodayUsesAthleteLocalMidnightWhenUTCDateDiffers(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		fitnessRows: decodeSummaries(t, `[
			{"date":"2026-05-24","fitness":72,"fatigue":81,"form":-9},
			{"date":"2026-05-25","fitness":73,"fatigue":82,"form":-9}
		]`),
		wellnessRows: []intervals.Wellness{
			decodeWellnessRow(t, `{"id":"2026-05-24","sleepQuality":3}`),
			decodeWellnessRow(t, `{"id":"2026-05-25","sleepQuality":1}`),
		},
		activities: decodeActivityPage(t,
			`{"id":"previous-local","name":"Late yesterday","type":"Run","start_date_local":"2026-05-23T23:55:00","start_date":"2026-05-24T02:55:00Z","timezone":"America/Sao_Paulo","distance":5000}`,
			`{"id":"today-local","name":"Late today","type":"Ride","start_date_local":"2026-05-24T23:20:00","start_date":"2026-05-25T02:20:00Z","timezone":"America/Sao_Paulo","distance":30000}`,
		),
		events: decodeToolEvents(t,
			`{"id":"previous-event","category":"WORKOUT","name":"Yesterday local","start_date_local":"2026-05-23T23:45:00"}`,
			`{"id":"today-event","category":"WORKOUT","name":"Today local","start_date_local":"2026-05-24T23:30:00"}`,
			`{"id":"tomorrow-event","category":"NOTE","name":"UTC tomorrow but local tomorrow too","start_date_local":"2026-05-25T00:05:00"}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)

	assertTodayCalls(t, client, "2026-05-24")
	meta := out["_meta"].(map[string]any)
	if meta["date"] != "2026-05-24" || meta["as_of_date"] != "2026-05-24" || meta["as_of_weekday"] != "Sunday" || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("meta = %#v, want athlete-local 2026-05-24 despite UTC 2026-05-25", meta)
	}
	if got := out["fitness"].([]any)[0].(map[string]any)["date"]; got != "2026-05-24" {
		t.Fatalf("fitness date = %#v, want athlete-local today row", got)
	}
	wellness := out["wellness"].([]any)
	if len(wellness) != 1 || wellness[0].(map[string]any)["date"] != "2026-05-24" {
		t.Fatalf("wellness = %#v, want only athlete-local today row", wellness)
	}
	activities := out["completed_activities"].([]any)
	if len(activities) != 1 || activities[0].(map[string]any)["activity_id"] != "today-local" || activities[0].(map[string]any)["start_date_utc"] != "2026-05-25T02:20:00Z" {
		t.Fatalf("completed_activities = %#v, want local-today activity even when UTC date is tomorrow", activities)
	}
	planned := out["planned_events"].([]any)
	if len(planned) != 1 || planned[0].(map[string]any)["event_id"] != "today-event" || planned[0].(map[string]any)["start_date_local"] != "2026-05-24T23:30:00" {
		t.Fatalf("planned_events = %#v, want only athlete-local today event", planned)
	}
}

func TestGetTodayRecomputesFreshAthleteLocalAnchorAcrossCalls(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 25, 2, 30, 0, 0, time.UTC)
	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		fitnessRows: decodeSummaries(t, `[
			{"date":"2026-05-24","fitness":72,"fatigue":81,"form":-9},
			{"date":"2026-05-25","fitness":73,"fatigue":82,"form":-9}
		]`),
		wellnessRows: []intervals.Wellness{
			decodeWellnessRow(t, `{"id":"2026-05-24","feel":4}`),
			decodeWellnessRow(t, `{"id":"2026-05-25","feel":3}`),
		},
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, func() time.Time { return now })

	first, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("first Handler() error = %v", err)
	}
	firstMeta := resultMap(t, first)["_meta"].(map[string]any)
	if firstMeta["date"] != "2026-05-24" || firstMeta["as_of"] != "2026-05-24T23:30:00-03:00" {
		t.Fatalf("first meta = %#v, want pre-midnight athlete-local anchor", firstMeta)
	}

	now = time.Date(2026, 5, 25, 3, 10, 0, 0, time.UTC)
	second, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("second Handler() error = %v", err)
	}
	secondMeta := resultMap(t, second)["_meta"].(map[string]any)
	if secondMeta["date"] != "2026-05-25" || secondMeta["as_of"] != "2026-05-25T00:10:00-03:00" || secondMeta["as_of_weekday"] != "Monday" {
		t.Fatalf("second meta = %#v, want fresh post-midnight athlete-local anchor", secondMeta)
	}
	if len(client.fitnessCalls) != 2 || client.fitnessCalls[0].Start != "2026-05-24" || client.fitnessCalls[1].Start != "2026-05-25" {
		t.Fatalf("fitness calls = %#v, want each invocation to fetch its fresh local today", client.fitnessCalls)
	}
	if len(client.wellnessCalls) != 2 || client.wellnessCalls[0].Oldest != "2026-05-24" || client.wellnessCalls[1].Oldest != "2026-05-25" {
		t.Fatalf("wellness calls = %#v, want each invocation to fetch its fresh local today", client.wellnessCalls)
	}
}

func TestGetTodayWorkoutStatusDoesNotInferCompletionFromSameDayActivity(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		activities: decodeActivityPage(t,
			`{"id":"unlinked-activity","name":"Easy jog","type":"Run","start_date_local":"2026-05-24T07:00:00","moving_time":1800}`,
			`{"id":"linked-activity","name":"Paired ride","type":"Ride","start_date_local":"2026-05-24T09:00:00","paired_event_id":"linked-event","moving_time":3600}`,
		),
		events: decodeToolEvents(t,
			`{"id":"planned-event","category":"WORKOUT","type":"Run","name":"Planned run","start_date_local":"2026-05-24","time_target":1800}`,
			`{"id":"linked-event","category":"WORKOUT","type":"Ride","name":"Linked ride","start_date_local":"2026-05-24","activity_id":"linked-activity"}`,
			`{"id":"note","category":"NOTE","name":"Travel","start_date_local":"2026-05-24"}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	plannedRows := rowsByEventID(out["planned_events"].([]any))
	if plannedRows["planned-event"]["workout_status"] != workoutStatusPlanned {
		t.Fatalf("planned event = %#v, want planned despite same-day unlinked activity", plannedRows["planned-event"])
	}
	if plannedRows["linked-event"]["workout_status"] != workoutStatusCompletedLinked || plannedRows["linked-event"]["paired_activity_id"] != "linked-activity" {
		t.Fatalf("linked event = %#v, want completed_linked", plannedRows["linked-event"])
	}
	activities := out["completed_activities"].([]any)
	byActivityID := map[string]map[string]any{}
	for _, item := range activities {
		row := item.(map[string]any)
		byActivityID[row["activity_id"].(string)] = row
	}
	if byActivityID["unlinked-activity"]["workout_status"] != workoutStatusCompletedUnlinked {
		t.Fatalf("unlinked activity = %#v, want completed_unlinked", byActivityID["unlinked-activity"])
	}
	if !stringSliceContains(stringSliceFromAny(byActivityID["unlinked-activity"]["workout_status_caveats"]), workoutCaveatUnlinkedActivity) {
		t.Fatalf("unlinked activity caveats = %#v, want completion caveat", byActivityID["unlinked-activity"])
	}
	if byActivityID["linked-activity"]["workout_status"] != workoutStatusCompletedLinked || byActivityID["linked-activity"]["paired_event_id"] != "linked-event" {
		t.Fatalf("linked activity = %#v, want completed_linked", byActivityID["linked-activity"])
	}
	annotation := out["annotations"].([]any)[0].(map[string]any)
	if _, ok := annotation["workout_status"]; ok {
		t.Fatalf("annotation carried workout_status: %#v", annotation)
	}
}

func TestGetTodayWeatherContextUsesUpstreamActivityWeatherOnly(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		activities: decodeActivityPage(t,
			`{"id":"today-weather","name":"Lunch Ride","type":"Ride","start_date_local":"2026-05-24T12:00:00","has_weather":true,"average_weather_temp":24.44,"average_wind_speed":3.789}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	weather := out["weather"].(map[string]any)
	if weather["status"] != "completed_activity_weather_available" || weather["completed_activity_weather_count"] != float64(1) || !strings.Contains(weather["provenance"].(string), "get_activities") {
		t.Fatalf("weather context = %#v, want completed activity weather provenance", weather)
	}
	activityWeather := out["completed_activities"].([]any)[0].(map[string]any)["weather"].(map[string]any)
	if activityWeather["average_temp_c"] != 24.4 || activityWeather["average_wind_speed_m_s"] != 3.79 {
		t.Fatalf("activity weather = %#v, want rounded upstream weather fields", activityWeather)
	}
}

func TestGetTodayWeatherContextDoesNotInventForecast(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events:            decodeToolEvents(t, `{"id":"today-event","category":"WORKOUT","name":"Endurance","start_date_local":"2026-05-24","indoor":false}`),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	weather := out["weather"].(map[string]any)
	if weather["status"] != "forecast_unavailable" || !strings.Contains(weather["summary"].(string), "do not infer or invent") || !strings.Contains(weather["provenance"].(string), "upstream gap") {
		t.Fatalf("weather context = %#v, want explicit unavailable no-hallucination wording", weather)
	}
	if strings.Contains(strings.ToLower(weather["summary"].(string)), "sunny") || strings.Contains(strings.ToLower(weather["summary"].(string)), "rain") {
		t.Fatalf("weather context invented conditions: %#v", weather)
	}
	planned := out["planned_events"].([]any)[0].(map[string]any)
	if planned["indoor"] != false {
		t.Fatalf("planned event = %#v, want existing indoor/outdoor flag preserved", planned)
	}
}

func TestGetTodayDoesNotBackfillYesterdayWellness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		rows         []intervals.Wellness
		wantCount    int
		wantFeel     any
		wantNoWeight bool
	}{
		{
			name:         "absent today excludes yesterday",
			rows:         []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-23","feel":5,"sleepQuality":4,"weight":70}`)},
			wantCount:    0,
			wantNoWeight: true,
		},
		{
			name: "partial today excludes richer yesterday",
			rows: []intervals.Wellness{
				decodeWellnessRow(t, `{"id":"2026-05-23","feel":5,"sleepQuality":4,"weight":70}`),
				decodeWellnessRow(t, `{"id":"2026-05-24","feel":2}`),
			},
			wantCount:    1,
			wantFeel:     float64(2),
			wantNoWeight: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakeTodayClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
				wellnessRows:      tc.rows,
			}
			tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			out := resultMap(t, result)
			wellness := out["wellness"].([]any)
			if len(wellness) != tc.wantCount {
				t.Fatalf("wellness = %#v, want %d today rows", wellness, tc.wantCount)
			}
			if tc.wantCount == 0 {
				return
			}
			row := wellness[0].(map[string]any)
			if row["date"] != "2026-05-24" || row["feel"] != tc.wantFeel {
				t.Fatalf("wellness row = %#v, want partial today row only", row)
			}
			if tc.wantNoWeight {
				if _, ok := row["weight"]; ok {
					t.Fatalf("wellness row backfilled yesterday weight: %#v", row)
				}
			}
		})
	}
}

func TestGetTodayPreservesTodayWellnessStaleMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		wellnessRows: []intervals.Wellness{
			decodeWellnessRow(t, `{"id":"2026-05-23","sleepScore":90,"source":"garmin","bridge_fetched_at":"2026-05-20T00:00:00Z"}`),
			decodeWellnessRow(t, `{"id":"2026-05-24","sleepScore":75,"source":"garmin","bridge_fetched_at":"2026-05-22T00:00:00Z"}`),
		},
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	wellness := resultMap(t, result)["wellness"].([]any)
	if len(wellness) != 1 {
		t.Fatalf("wellness = %#v, want only today row", wellness)
	}
	meta := wellness[0].(map[string]any)["_meta"].(map[string]any)
	if meta["stale"] != true || !strings.Contains(meta["stale_reason"].(string), "garmin bridge data is older than 24h") {
		t.Fatalf("wellness _meta = %#v, want stale provenance preserved", meta)
	}
}

func TestGetTodayDigestUsesAthleteLocalDateAndSourceShapes(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "imperial", Timezone: "America/Sao_Paulo"}},
		fitnessRows:       decodeSummaries(t, `[{"date":"2026-05-24","fitness":71.2345,"fatigue":80.2,"form":-8.9}]`),
		wellnessRows:      []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-05-24","sleepQuality":3,"feel":4,"weight":null}`)},
		activities: decodeActivityPage(t,
			`{"id":"a1","name":"Morning Run","type":"Run","start_date_local":"2026-05-24T07:30:00","distance":1609.344,"moving_time":480,"calories":120,"stream_types":["heartrate"],"tags":["commute","easy"]}`,
		),
		events: decodeToolEvents(t,
			`{"id":"11","category":"WORKOUT","type":"Run","name":"Easy run","start_date_local":"2026-05-24","icu_training_load":35,"tags":["plan","run"]}`,
			`{"id":"12","category":"NOTE","name":"Travel","description":"Pack race shoes","start_date_local":"2026-05-24","tags":["logistics"]}`,
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
	activityTags := activity["tags"].([]any)
	if len(activityTags) != 2 || activityTags[0] != "commute" || activityTags[1] != "easy" {
		t.Fatalf("activity tags = %#v, want completed activity tags", activityTags)
	}
	if _, ok := activity["full"]; ok {
		t.Fatalf("activity included full in terse digest: %#v", activity)
	}
	planned := out["planned_events"].([]any)
	annotations := out["annotations"].([]any)
	if len(planned) != 1 || planned[0].(map[string]any)["category"] != "WORKOUT" {
		t.Fatalf("planned_events = %#v, want only workout event", planned)
	}
	plannedTags := planned[0].(map[string]any)["tags"].([]any)
	if len(plannedTags) != 2 || plannedTags[0] != "plan" || plannedTags[1] != "run" {
		t.Fatalf("planned tags = %#v, want event tags", plannedTags)
	}
	if len(annotations) != 2 || annotations[0].(map[string]any)["category"] != "NOTE" || annotations[1].(map[string]any)["category"] != "RACE_B" {
		t.Fatalf("annotations = %#v, want NOTE and race categories", annotations)
	}
	annotationTags := annotations[0].(map[string]any)["tags"].([]any)
	if len(annotationTags) != 1 || annotationTags[0] != "logistics" {
		t.Fatalf("annotation tags = %#v, want note tags", annotationTags)
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

func TestGetTodayPreservesMultipleSameDayPlannedEventsAndAnnotations(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"101","category":"WORKOUT","type":"Run","name":"AM aerobic run","start_date_local":"2026-05-24","icu_training_load":35}`,
			`{"id":"102","category":"WORKOUT","type":"Ride","name":"PM endurance ride","start_date_local":"2026-05-24","icu_training_load":52}`,
			`{"id":"103","category":"NOTE","name":"Travel window","description":"Pack race shoes","start_date_local":"2026-05-24"}`,
			`{"id":"104","category":"RACE_A","name":"Goal race marker","start_date_local":"2026-05-24"}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)

	assertTodayCalls(t, client, "2026-05-24")
	planned := out["planned_events"].([]any)
	if len(planned) != 2 {
		t.Fatalf("planned_events count = %d, want both same-day workouts: %#v", len(planned), planned)
	}
	wantPlanned := []struct {
		id       string
		name     string
		category string
	}{
		{id: "101", name: "AM aerobic run", category: "WORKOUT"},
		{id: "102", name: "PM endurance ride", category: "WORKOUT"},
	}
	for i, want := range wantPlanned {
		row := planned[i].(map[string]any)
		if row["event_id"] != want.id || row["name"] != want.name || row["category"] != want.category || row["start_date_local"] != "2026-05-24" {
			t.Fatalf("planned_events[%d] = %#v, want separately identified same-day workout %#v", i, row, want)
		}
	}
	annotations := out["annotations"].([]any)
	if len(annotations) != 2 || annotations[0].(map[string]any)["event_id"] != "103" || annotations[1].(map[string]any)["event_id"] != "104" {
		t.Fatalf("annotations = %#v, want NOTE and race rows preserved separately", annotations)
	}
	counts := out["_meta"].(map[string]any)["section_counts"].(map[string]any)
	if counts["planned_events"] != float64(2) || counts["annotations"] != float64(2) {
		t.Fatalf("section_counts = %#v, want all same-day events counted", counts)
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
