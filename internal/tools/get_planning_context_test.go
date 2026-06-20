package tools

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakePlanningContextClient struct {
	fakeProfileClient
	weekEvents          []intervals.Event
	raceEvents          []intervals.Event
	seasonEvents        []intervals.Event
	seasonEventsByStart map[string][]intervals.Event
	trainingPlan        intervals.TrainingPlan
	fitnessRows         []intervals.SummaryWithCats
	listCalls           []intervals.ListEventsParams
	trainingCalls       int
	fitnessCalls        []intervals.AthleteSummaryParams
	listErr             error
	trainingPlanErr     error
	fitnessErr          error
}

func (f *fakePlanningContextClient) ListEvents(_ context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	if f.listErr != nil {
		return nil, f.listErr
	}
	if params.Category == planningContextSeasonCategory {
		if f.seasonEventsByStart != nil {
			return append([]intervals.Event(nil), f.seasonEventsByStart[params.Oldest]...), nil
		}
		if len(f.listCalls) == 3 {
			return append([]intervals.Event(nil), f.seasonEvents...), nil
		}
		return nil, nil
	}
	if len(f.listCalls) == 1 {
		return append([]intervals.Event(nil), f.weekEvents...), nil
	}
	return append([]intervals.Event(nil), f.raceEvents...), nil
}

func (f *fakePlanningContextClient) GetTrainingPlan(context.Context) (intervals.TrainingPlan, error) {
	f.trainingCalls++
	return f.trainingPlan, f.trainingPlanErr
}

func (f *fakePlanningContextClient) ListAthleteSummary(_ context.Context, params intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	f.fitnessCalls = append(f.fitnessCalls, params)
	return append([]intervals.SummaryWithCats(nil), f.fitnessRows...), f.fitnessErr
}

func TestGetPlanningContextRegistrationMetadata(t *testing.T) {
	t.Parallel()

	tool := newGetPlanningContextToolWithClock(&fakePlanningContextClient{}, &fakePlanningContextClient{}, "test", "UTC", false, fixedTodayClock())
	if tool.Name != getPlanningContextName || !strings.Contains(tool.Description, "read-only weekly planning context") || !strings.Contains(tool.Description, "without creating an ATP") {
		t.Fatalf("tool metadata = %#v, want read-only planning description", tool)
	}
	if tool.EffectiveToolset() != safety.ToolsetFull {
		t.Fatalf("toolset = %q, want full", tool.EffectiveToolset())
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"week_start", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s: %#v", name, props)
		}
	}
	out := tool.OutputSchema.(map[string]any)
	if !strings.Contains(out["description"].(string), "read_only=true") || !strings.Contains(out["description"].(string), "never creates ATP") {
		t.Fatalf("output schema = %#v, want read-only/no-ATP language", out)
	}
}

func TestGetPlanningContextTerseDefaultShapeAndCalls(t *testing.T) {
	t.Parallel()

	client := &fakePlanningContextClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		weekEvents: decodeToolEvents(t,
			`{"id":"w1","category":"WORKOUT","type":"Run","name":"Easy run","start_date_local":"2026-05-25","icu_training_load":35,"raw_extra":"hidden"}`,
			`{"id":"r1","category":"RACE_B","name":"Tune-up race","start_date_local":"2026-05-31"}`,
			`{"id":"n1","category":"NOTE","name":"Travel","start_date_local":"2026-05-26"}`,
			`{"id":"o1","category":"OTHER","name":"Bike fit","start_date_local":"2026-05-27"}`,
		),
		raceEvents:   decodeToolEvents(t, `{"id":"future-race","category":"RACE_A","name":"A race","start_date_local":"2026-06-21"}`, `{"id":"not-race","category":"WORKOUT","name":"Workout","start_date_local":"2026-06-01"}`),
		seasonEvents: decodeToolEvents(t, `{"id":"season-2026","category":"SEASON_START","name":"2026 Build","description":"Base starts","start_date_local":"2026-01-01"}`),
		trainingPlan: decodeTrainingPlan(t, `{"id":"tp1","plan_id":"p1","name":"Base","training_plan_start_date":"2026-05-01","training_plan":{"id":"p1","name":"Base Plan","description":"Build"}}`),
		fitnessRows:  decodeSummaries(t, `[{"date":"2026-05-18","fitness":70,"fatigue":75,"form":-5},{"date":"2026-05-24","fitness":71.2345,"fatigue":80.2,"form":-8.9,"raw_extra":"hidden"}]`),
	}
	tool := newGetPlanningContextToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)

	assertPlanningCalls(t, client, planningCallWant{weekStart: "2026-05-25", weekEnd: "2026-05-31", raceStart: "2026-05-24", raceEnd: "2026-08-16", fitnessStart: "2026-05-18", fitnessEnd: "2026-05-24"})
	week := out["week"].(map[string]any)
	if week["start_date"] != "2026-05-25" || week["end_date"] != "2026-05-31" || week["anchor"] != "upcoming_monday" || week["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("week = %#v, want upcoming athlete-local Monday week", week)
	}
	weekEvents := out["week_events"].(map[string]any)
	for section, want := range map[string]int{"planned_workouts": 1, "races": 1, "notes": 1, "other_events": 1} {
		rows := weekEvents[section].([]any)
		if len(rows) != want {
			t.Fatalf("%s count = %d, want %d: %#v", section, len(rows), want, rows)
		}
		if _, ok := rows[0].(map[string]any)["full"]; ok {
			t.Fatalf("%s included full in terse response: %#v", section, rows[0])
		}
	}
	fitness := out["fitness_context"].(map[string]any)
	current := fitness["current"].(map[string]any)
	if current["date"] != "2026-05-24" || current["ctl"] != 71.235 || current["atl"] != 80.2 || current["tsb"] != -8.9 {
		t.Fatalf("current fitness = %#v, want latest rounded row", current)
	}
	if _, ok := current["full"]; ok {
		t.Fatalf("current fitness included full in terse response: %#v", current)
	}
	upcoming := out["upcoming_races"].([]any)
	if len(upcoming) != 1 || upcoming[0].(map[string]any)["event_id"] != "future-race" {
		t.Fatalf("upcoming_races = %#v, want only race-scan race", upcoming)
	}
	seasonContext := out["season_context"].(map[string]any)
	seasons := seasonContext["seasons"].([]any)
	if len(seasons) != 1 || seasons[0].(map[string]any)["event_id"] != "season-2026" || seasons[0].(map[string]any)["start_date"] != "2026-01-01" {
		t.Fatalf("season_context.seasons = %#v, want terse season row", seasons)
	}
	if _, ok := seasons[0].(map[string]any)["full"]; ok {
		t.Fatalf("season row included full in terse response: %#v", seasons[0])
	}
	seasonMeta := seasonContext["_meta"].(map[string]any)
	if seasonMeta["category"] != planningContextSeasonCategory || seasonMeta["count"] != float64(1) || seasonMeta["chunk_count"] != float64(4) || seasonMeta["seasons_may_be_truncated"] != false {
		t.Fatalf("season meta = %#v, want category/count/chunks", seasonMeta)
	}
	meta := out["_meta"].(map[string]any)
	assertPlanningMetaBasics(t, meta)
	counts := meta["section_counts"].(map[string]any)
	if counts["planned_workouts"] != float64(1) || counts["upcoming_races"] != float64(1) || counts["fitness_rows"] != float64(2) {
		t.Fatalf("section_counts = %#v, want planned/upcoming/fitness counts", counts)
	}
	if got := planningStringSlice(meta["caveat_codes"]); !slices.Equal(got, []string{"read_only_no_atp"}) {
		t.Fatalf("caveat_codes = %#v, want only read_only_no_atp", got)
	}
}

func TestGetPlanningContextIncludeFullWidensOnlySourceRows(t *testing.T) {
	t.Parallel()

	client := &fakePlanningContextClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		weekEvents:        decodeToolEvents(t, `{"id":"w1","category":"WORKOUT","name":"Workout","start_date_local":"2026-05-25","raw_extra":"kept"}`),
		raceEvents:        decodeToolEvents(t, `{"id":"r1","category":"RACE","name":"Race","start_date_local":"2026-06-21","raw_extra":"kept"}`),
		seasonEvents:      decodeToolEvents(t, `{"id":"s1","category":"SEASON_START","name":"Season","start_date_local":"2026-01-01","raw_extra":"kept"}`),
		trainingPlan:      decodeTrainingPlan(t, `{"id":"tp1","plan_id":"p1","name":"Base","raw_extra":"kept","training_plan":{"id":"p1","name":"Base Plan"}}`),
		fitnessRows:       decodeSummaries(t, `[{"date":"2026-05-24","fitness":71,"fatigue":80,"form":-9,"raw_extra":"kept"}]`),
	}
	tool := newGetPlanningContextToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	event := out["week_events"].(map[string]any)["planned_workouts"].([]any)[0].(map[string]any)
	if event["full"].(map[string]any)["raw_extra"] != "kept" {
		t.Fatalf("event full = %#v, want raw_extra", event["full"])
	}
	fitness := out["fitness_context"].(map[string]any)["rows"].([]any)[0].(map[string]any)
	if fitness["full"].(map[string]any)["raw_extra"] != "kept" {
		t.Fatalf("fitness full = %#v, want raw_extra", fitness["full"])
	}
	plan := out["training_plan"].(map[string]any)["training_plan"].(map[string]any)
	if plan["full"].(map[string]any)["raw_extra"] != "kept" {
		t.Fatalf("training plan full = %#v, want raw_extra", plan["full"])
	}
	season := out["season_context"].(map[string]any)["seasons"].([]any)[0].(map[string]any)
	if season["full"].(map[string]any)["raw_extra"] != "kept" {
		t.Fatalf("season full = %#v, want raw_extra", season["full"])
	}
	if out["_meta"].(map[string]any)["include_full"] != true {
		t.Fatalf("_meta.include_full = %#v, want true", out["_meta"])
	}
}

func TestGetPlanningContextWeekStartValidationAndFitnessWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          string
		wantWeekStart string
		wantWeekEnd   string
		wantAnchor    string
		wantErr       bool
	}{
		{name: "supplied mid-week normalizes to Monday", args: `{"week_start":"2026-06-10"}`, wantWeekStart: "2026-06-08", wantWeekEnd: "2026-06-14", wantAnchor: "supplied_week_start_normalized_to_monday"},
		{name: "future week keeps current fitness window", args: `{"week_start":"2026-07-15"}`, wantWeekStart: "2026-07-13", wantWeekEnd: "2026-07-19", wantAnchor: "supplied_week_start_normalized_to_monday"},
		{name: "invalid week_start returns user error", args: `{"week_start":"next monday"}`, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := &fakePlanningContextClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}}}
			tool := newGetPlanningContextToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "week_start must be YYYY-MM-DD") {
					t.Fatalf("Handler() error = %v, want week_start validation error", err)
				}
				if len(client.listCalls) != 0 || len(client.fitnessCalls) != 0 || client.trainingCalls != 0 {
					t.Fatalf("client calls after invalid input = lists %#v fitness %#v training %d, want none", client.listCalls, client.fitnessCalls, client.trainingCalls)
				}
				return
			}
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			out := resultMap(t, result)
			week := out["week"].(map[string]any)
			if week["start_date"] != tc.wantWeekStart || week["end_date"] != tc.wantWeekEnd || week["anchor"] != tc.wantAnchor {
				t.Fatalf("week = %#v, want %s..%s anchor %s", week, tc.wantWeekStart, tc.wantWeekEnd, tc.wantAnchor)
			}
			assertPlanningCalls(t, client, planningCallWant{weekStart: tc.wantWeekStart, weekEnd: tc.wantWeekEnd, raceStart: "2026-05-24", raceEnd: "2026-08-16", fitnessStart: "2026-05-18", fitnessEnd: "2026-05-24"})
		})
	}
}

func TestGetPlanningContextEmptyDataCaveats(t *testing.T) {
	t.Parallel()

	client := &fakePlanningContextClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}}}
	tool := newGetPlanningContextToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	codes := planningStringSlice(resultMap(t, result)["_meta"].(map[string]any)["caveat_codes"])
	for _, want := range []string{"read_only_no_atp", "no_week_events", "no_week_workouts", "no_active_training_plan", "no_fitness_rows", "no_upcoming_races"} {
		if !slices.Contains(codes, want) {
			t.Fatalf("caveat_codes = %#v, missing %s", codes, want)
		}
	}
}

func TestGetPlanningContextPartialAndTruncationCaveats(t *testing.T) {
	t.Parallel()

	client := &fakePlanningContextClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		weekEvents:        repeatedPlanningEvents(t, planningContextEventLimit, "week", "OTHER"),
		raceEvents:        repeatedPlanningEvents(t, planningContextEventLimit, "race-scan", "WORKOUT"),
		trainingPlan:      decodeTrainingPlan(t, `{"id":"tp1","plan_id":"p1","name":"Partial"}`),
		fitnessRows:       decodeSummaries(t, `[{"date":"2026-05-24","fitness":71,"fatigue":80,"form":-9}]`),
	}
	tool := newGetPlanningContextToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	truncation := meta["truncation"].(map[string]any)
	if truncation["week_events_may_be_truncated"] != true || truncation["race_scan_may_be_truncated"] != true {
		t.Fatalf("truncation = %#v, want both true", truncation)
	}
	codes := planningStringSlice(meta["caveat_codes"])
	for _, want := range []string{"read_only_no_atp", "no_week_workouts", "partial_training_plan_summary", "no_upcoming_races", "week_events_may_be_truncated", "upcoming_races_may_be_truncated"} {
		if !slices.Contains(codes, want) {
			t.Fatalf("caveat_codes = %#v, missing %s", codes, want)
		}
	}
	if slices.Contains(codes, "no_week_events") || slices.Contains(codes, "no_fitness_rows") || slices.Contains(codes, "no_active_training_plan") {
		t.Fatalf("caveat_codes = %#v, included empty/inactive caveat despite data", codes)
	}
}

func decodeTrainingPlan(t *testing.T, raw string) intervals.TrainingPlan {
	t.Helper()
	var plan intervals.TrainingPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		t.Fatalf("decode training plan: %v", err)
	}
	return plan
}

type planningCallWant struct {
	weekStart    string
	weekEnd      string
	raceStart    string
	raceEnd      string
	fitnessStart string
	fitnessEnd   string
}

func assertPlanningCalls(t *testing.T, client *fakePlanningContextClient, want planningCallWant) {
	t.Helper()
	if len(client.listCalls) != 6 {
		t.Fatalf("ListEvents calls = %#v, want week, race, and 4 season scans", client.listCalls)
	}
	weekCall := client.listCalls[0]
	if weekCall.Oldest != want.weekStart || weekCall.Newest != want.weekEnd || weekCall.Limit != planningContextEventLimit {
		t.Fatalf("week ListEvents = %#v, want %s..%s limit %d", weekCall, want.weekStart, want.weekEnd, planningContextEventLimit)
	}
	raceCall := client.listCalls[1]
	if raceCall.Oldest != want.raceStart || raceCall.Newest != want.raceEnd || raceCall.Limit != planningContextEventLimit {
		t.Fatalf("race ListEvents = %#v, want %s..%s limit %d", raceCall, want.raceStart, want.raceEnd, planningContextEventLimit)
	}
	wantSeasonCalls := []dateRangeMeta{{Oldest: "2023-05-24", Newest: "2024-05-23"}, {Oldest: "2024-05-24", Newest: "2025-05-24"}, {Oldest: "2025-05-25", Newest: "2026-05-25"}, {Oldest: "2026-05-26", Newest: "2027-05-24"}}
	for i, wantChunk := range wantSeasonCalls {
		seasonCall := client.listCalls[i+2]
		if seasonCall.Oldest != wantChunk.Oldest || seasonCall.Newest != wantChunk.Newest || seasonCall.Category != planningContextSeasonCategory || seasonCall.Limit != planningContextEventLimit {
			t.Fatalf("season ListEvents[%d] = %#v, want %#v category %s limit %d", i, seasonCall, wantChunk, planningContextSeasonCategory, planningContextEventLimit)
		}
	}
	if client.trainingCalls != 1 {
		t.Fatalf("training calls = %d, want 1", client.trainingCalls)
	}
	if len(client.fitnessCalls) != 1 || client.fitnessCalls[0].Start != want.fitnessStart || client.fitnessCalls[0].End != want.fitnessEnd {
		t.Fatalf("fitness calls = %#v, want %s..%s", client.fitnessCalls, want.fitnessStart, want.fitnessEnd)
	}
}

func assertPlanningMetaBasics(t *testing.T, meta map[string]any) {
	t.Helper()
	if meta["as_of"] != "2026-05-24T23:30:00-03:00" || meta["as_of_date"] != "2026-05-24" || meta["as_of_weekday"] != "Sunday" || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("as-of meta = %#v, want Sao Paulo fixed clock", meta)
	}
	if meta["read_only"] != true || meta["writes_performed"] != false || meta["planning_scope"] != planningContextScopeContext {
		t.Fatalf("read-only meta = %#v, want context-only no writes", meta)
	}
	if got := planningStringSlice(meta["source_tools"]); !slices.Equal(got, []string{getAthleteProfileName, getEventsName, getTrainingPlanName, getFitnessName}) {
		t.Fatalf("source_tools = %#v, want planning source tools", got)
	}
	weekWindow := meta["week_window"].(map[string]any)
	fitnessWindow := meta["fitness_window"].(map[string]any)
	raceWindow := meta["race_window"].(map[string]any)
	if weekWindow["oldest"] != "2026-05-25" || weekWindow["newest"] != "2026-05-31" || fitnessWindow["oldest"] != "2026-05-18" || fitnessWindow["newest"] != "2026-05-24" || raceWindow["oldest"] != "2026-05-24" || raceWindow["newest"] != "2026-08-16" {
		t.Fatalf("windows = week %#v fitness %#v race %#v, want contract windows", weekWindow, fitnessWindow, raceWindow)
	}
	limits := meta["event_limits"].(map[string]any)
	if limits["week_events"] != float64(planningContextEventLimit) || limits["race_scan"] != float64(planningContextEventLimit) {
		t.Fatalf("event_limits = %#v, want 500/500", limits)
	}
}

func planningStringSlice(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return nil
		}
		out = append(out, text)
	}
	return out
}

func repeatedPlanningEvents(t *testing.T, count int, prefix string, category string) []intervals.Event {
	t.Helper()
	events := make([]intervals.Event, 0, count)
	for i := 0; i < count; i++ {
		events = append(events, decodeToolEvents(t, `{"id":"`+prefix+`-`+string(rune('a'+(i%26)))+`","category":"`+category+`","name":"Event","start_date_local":"2026-05-25"}`)[0])
	}
	return events
}

func TestGetPlanningContextPropagatesReadErrors(t *testing.T) {
	t.Parallel()

	client := &fakePlanningContextClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}}, listErr: errors.New("boom")}
	tool := newGetPlanningContextToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err == nil || !strings.Contains(err.Error(), fetchPlanningContextMessage) {
		t.Fatalf("Handler() error = %v, want planning fetch user error", err)
	}
}
