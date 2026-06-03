package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeEventsTrainingPlanClient struct {
	fakeProfileClient
	events          []intervals.Event
	listCalls       []intervals.ListEventsParams
	eventDetail     intervals.Event
	eventDetailErr  error
	detailCalls     []string
	trainingPlan    intervals.TrainingPlan
	trainingPlanErr error
}

func (f *fakeEventsTrainingPlanClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	return append([]intervals.Event(nil), f.events...), nil
}

func (f *fakeEventsTrainingPlanClient) GetEvent(ctx context.Context, eventID string) (intervals.Event, error) {
	f.detailCalls = append(f.detailCalls, eventID)
	return f.eventDetail, f.eventDetailErr
}

func (f *fakeEventsTrainingPlanClient) GetTrainingPlan(ctx context.Context) (intervals.TrainingPlan, error) {
	return f.trainingPlan, f.trainingPlanErr
}

func TestEventsAndTrainingPlanRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	eventsTool := newGetEventsTool(client, client, "test", "UTC", false)
	if !strings.Contains(eventsTool.Description, "calendar events") {
		t.Fatalf("events description = %q, want calendar events", eventsTool.Description)
	}
	eventsProps := eventsTool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"oldest", "newest", "category", "calendar_id", "limit", "resolve", "include_full"} {
		if _, ok := eventsProps[name]; !ok {
			t.Fatalf("get_events schema missing %s", name)
		}
	}
	eventByIDTool := newGetEventByIDTool(client, client, "test", "UTC", false)
	if !strings.Contains(eventByIDTool.Description, "structured unavailable") {
		t.Fatalf("event by ID description = %q, want structured unavailable language", eventByIDTool.Description)
	}
	eventByIDProps := eventByIDTool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"event_id", "date", "oldest", "newest", "resolve", "include_full"} {
		if _, ok := eventByIDProps[name]; !ok {
			t.Fatalf("get_event_by_id schema missing %s", name)
		}
	}
	planTool := newGetTrainingPlanTool(client, client, "test", "UTC", false)
	if !strings.Contains(planTool.Description, "training-plan assignment") {
		t.Fatalf("training plan description = %q, want assignment language", planTool.Description)
	}
}

func TestGetEventsTerseRowsTimezoneAndCategory(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events:            decodeToolEvents(t, `{"id":123,"name":"Tempo","category":"WORKOUT","type":"Ride","start_date_local":"2026-01-03","end_date_local":"2026-01-03","description":"3x tempo","tags":["tempo","coach"],"indoor":true,"updated":"2026-01-03T12:00:00Z","plan_applied":"2026-01-02T12:00:00Z","calendar_id":"cal-1","training_plan_id":456,"icu_training_load":75,"distance":30000,"moving_time":3600,"workout_doc":{"steps":[{"duration":600}]}}`),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-01-31","category":"WORKOUT","limit":10,"resolve":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.listCalls) != 1 {
		t.Fatalf("list calls = %d, want 1", len(client.listCalls))
	}
	call := client.listCalls[0]
	if call.Oldest != "2026-01-01" || call.Newest != "2026-01-31" || call.Category != "WORKOUT" || call.Limit != 10 || call.Resolve == nil || !*call.Resolve {
		t.Fatalf("ListEvents params = %#v, want decoded request params", call)
	}
	out := resultMap(t, result)
	rows := out["events"].([]any)
	row := rows[0].(map[string]any)
	if row["event_id"] != "123" || row["category"] != "WORKOUT" {
		t.Fatalf("row id/category = %#v/%#v, want string id and raw category", row["event_id"], row["category"])
	}
	if row["indoor"] != true {
		t.Fatalf("indoor = %#v, want true", row["indoor"])
	}
	tags := row["tags"].([]any)
	if len(tags) != 2 || tags[0] != "tempo" || tags[1] != "coach" {
		t.Fatalf("tags = %#v, want upstream order", tags)
	}
	if row["updated_local"] != "2026-01-03T09:00:00-03:00" {
		t.Fatalf("updated_local = %#v, want Sao Paulo rendering", row["updated_local"])
	}
	if _, ok := row["full"]; ok {
		t.Fatalf("full present in terse row: %#v", row["full"])
	}
	summary := row["workout_doc_summary"].(map[string]any)
	if summary["step_count"] != float64(1) {
		t.Fatalf("workout_doc_summary = %#v, want step_count 1", summary)
	}
	meta := out["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" || meta["count"] != float64(1) || meta["truncated"] != false {
		t.Fatalf("meta = %#v, want timezone/count/truncated", meta)
	}
}

func TestGetEventsResolvesPercentFTPWorkoutTargetPreview(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC", SportSettings: []intervals.SportSettings{{Types: []string{"Ride"}, FTP: 250}}}},
		events:            decodeToolEvents(t, `{"id":"ftp","name":"Sweet Spot","category":"WORKOUT","type":"Ride","start_date_local":"2026-01-03","workout_doc":{"steps":[{"description":"Sweet spot","duration":600,"power":{"min":88,"max":94,"units":"PERCENT_FTP"}}]}}`),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-03","newest":"2026-01-03"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("profile calls = %d, want one call reused for units and target previews", client.calls)
	}
	row := resultMap(t, result)["events"].([]any)[0].(map[string]any)
	summary := row["workout_doc_summary"].(map[string]any)
	previews := summary["target_previews"].([]any)
	if len(previews) != 1 {
		t.Fatalf("target_previews = %#v, want one resolved FTP preview", previews)
	}
	preview := previews[0].(map[string]any)
	if preview["path"] != "1" || preview["family"] != "power" || preview["target"] != "88-94% FTP" || preview["preview"] != "220-235 W" || preview["basis"] != "ftp 250 W" {
		t.Fatalf("preview = %#v, want compact FTP watts resolution", preview)
	}
	if _, ok := row["workout_doc"]; ok {
		t.Fatalf("raw workout_doc leaked in terse row: %#v", row)
	}
}

func TestGetEventsPreservesLongDistanceRaceMeters(t *testing.T) {
	t.Parallel()

	const brevetDistanceMeters = 1_200_000.0
	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events:            decodeToolEvents(t, `{"id":"evt-1200","name":"1200 km brevet","category":"RACE","start_date_local":"2026-08-01","distance":1200000,"distance_target":1200000}`),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-08-01","newest":"2026-08-01","category":"RACE"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	rows := out["events"].([]any)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want one long-distance race", len(rows))
	}
	row := rows[0].(map[string]any)
	if row["distance_meters"] != brevetDistanceMeters || row["distance_target_meters"] != brevetDistanceMeters {
		t.Fatalf("row = %#v, want untruncated 1200 km distance and target distance", row)
	}
	assertKeyAbsent(t, row, "icu_training_load")
	assertKeyAbsent(t, row, "load_target")
	lowerText := strings.ToLower(resultText(t, result))
	for _, forbidden := range []string{"auto-load", "autocalc", "auto calculated", "auto-calculated", "calculated load"} {
		if strings.Contains(lowerText, forbidden) {
			t.Fatalf("result text contains false auto-load wording %q: %s", forbidden, lowerText)
		}
	}
}

func TestGetEventsPreservesMultipleSameDayEvents(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"201","name":"AM aerobic run","category":"WORKOUT","type":"Run","start_date_local":"2026-05-25","icu_training_load":35}`,
			`{"id":"202","name":"PM endurance ride","category":"WORKOUT","type":"Ride","start_date_local":"2026-05-25","icu_training_load":52}`,
			`{"id":"203","name":"Travel window","category":"NOTE","start_date_local":"2026-05-25"}`,
			`{"id":"204","name":"Goal race marker","category":"RACE_B","start_date_local":"2026-05-25"}`,
		),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-25","newest":"2026-05-25","limit":10}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.listCalls) != 1 || client.listCalls[0].Oldest != "2026-05-25" || client.listCalls[0].Newest != "2026-05-25" {
		t.Fatalf("ListEvents params = %#v, want single athlete-local date range", client.listCalls)
	}
	rows := resultMap(t, result)["events"].([]any)
	if len(rows) != 4 {
		t.Fatalf("events count = %d, want all same-day events: %#v", len(rows), rows)
	}
	want := []struct {
		id       string
		name     string
		category string
	}{
		{id: "201", name: "AM aerobic run", category: "WORKOUT"},
		{id: "202", name: "PM endurance ride", category: "WORKOUT"},
		{id: "203", name: "Travel window", category: "NOTE"},
		{id: "204", name: "Goal race marker", category: "RACE_B"},
	}
	for i, wantRow := range want {
		row := rows[i].(map[string]any)
		if row["event_id"] != wantRow.id || row["name"] != wantRow.name || row["category"] != wantRow.category || row["start_date_local"] != "2026-05-25" {
			t.Fatalf("events[%d] = %#v, want separately identified same-day event %#v", i, row, wantRow)
		}
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["count"] != float64(4) || meta["truncated"] != false {
		t.Fatalf("meta = %#v, want all same-day events counted without truncation", meta)
	}
}

func TestGetEventsCurrentDayRangeAddsAsOfMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events:            decodeToolEvents(t, `{"id":123,"name":"Easy","category":"WORKOUT","start_date_local":"2026-05-24"}`),
	}
	tool := newGetEventsToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-24","newest":"2026-05-24","limit":10}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	assertSaoPauloAsOfMeta(t, meta)
	if meta["count"] != float64(1) || meta["limit"] != float64(10) || meta["truncated"] != false {
		t.Fatalf("events meta = %#v, want preserved count/limit/truncated", meta)
	}
}

func TestGetEventsPastRangeOmitsAsOfMetadata(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events:            decodeToolEvents(t, `{"id":123,"name":"Easy","category":"WORKOUT","start_date_local":"2026-05-23"}`),
	}
	tool := newGetEventsToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-23","limit":10}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	assertNoAsOfMeta(t, meta)
	if meta["timezone"] != "America/Sao_Paulo" || meta["count"] != float64(1) || meta["limit"] != float64(10) || meta["truncated"] != false {
		t.Fatalf("events meta = %#v, want preserved timezone/count/limit/truncated", meta)
	}
	rangeMeta := meta["date_range"].(map[string]any)
	if rangeMeta["oldest"] != "2026-05-01" || rangeMeta["newest"] != "2026-05-23" {
		t.Fatalf("date_range = %#v, want requested past range", rangeMeta)
	}
}

func TestGetEventsCapsRowsAndReportsTruncation(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events: decodeToolEvents(t,
			`{"id":3,"name":"Suppressed","category":"WORKOUT","start_date_local":"2026-01-03"}`,
			`{"id":1,"name":"First","category":"WORKOUT","start_date_local":"2026-01-01"}`,
			`{"id":2,"name":"Second","category":"WORKOUT","start_date_local":"2026-01-02"}`,
		),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-01-31","limit":2}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	rows := out["events"].([]any)
	if len(rows) != 2 {
		t.Fatalf("row count = %d, want cap of 2", len(rows))
	}
	meta := out["_meta"].(map[string]any)
	if meta["limit"] != float64(2) || meta["count"] != float64(2) || meta["truncated"] != true {
		t.Fatalf("meta = %#v, want capped count and truncated=true", meta)
	}
	if rows[0].(map[string]any)["event_id"] != "1" || rows[1].(map[string]any)["event_id"] != "3" {
		t.Fatalf("rows = %#v, want first two upstream rows sorted at response boundary", rows)
	}
}

func TestGetEventsTagEdgeCases(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		events: decodeToolEvents(t,
			`{"id":"present","category":"WORKOUT","start_date_local":"2026-01-01","tags":["alpha","beta"]}`,
			`{"id":"empty","category":"WORKOUT","start_date_local":"2026-01-02","tags":[]}`,
			`{"id":"missing","category":"WORKOUT","start_date_local":"2026-01-03"}`,
			`{"id":"null","category":"WORKOUT","start_date_local":"2026-01-04","tags":null}`,
			`{"id":"object","category":"WORKOUT","start_date_local":"2026-01-05","tags":{"name":"alpha"}}`,
			`{"id":"mixed","category":"WORKOUT","start_date_local":"2026-01-06","tags":["alpha",7]}`,
		),
	}
	tool := newGetEventsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-01-31","limit":10,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	rows := resultMap(t, result)["events"].([]any)
	byID := map[string]map[string]any{}
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		byID[row["event_id"].(string)] = row
	}

	present := byID["present"]
	presentTags := present["tags"].([]any)
	if len(presentTags) != 2 || presentTags[0] != "alpha" || presentTags[1] != "beta" {
		t.Fatalf("present tags = %#v, want preserved order", presentTags)
	}
	fullTags := present["full"].(map[string]any)["tags"].([]any)
	if len(fullTags) != 2 || fullTags[0] != "alpha" || fullTags[1] != "beta" {
		t.Fatalf("full tags = %#v, want raw payload preserved", fullTags)
	}

	emptyTags, ok := byID["empty"]["tags"].([]any)
	if !ok || len(emptyTags) != 0 {
		t.Fatalf("empty tags = %#v, want explicit empty array", byID["empty"]["tags"])
	}
	for _, id := range []string{"missing", "null", "object", "mixed"} {
		assertKeyAbsent(t, byID[id], "tags")
	}
	assertKeyPresentNil(t, byID["null"]["full"].(map[string]any), "tags")
}

func TestGetEventsRejectsDateTimesAndTooLargeRange(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newGetEventsTool(client, client, "test", "UTC", false)
	cases := []string{
		`{"oldest":"2026-01-01T00:00:00","newest":"2026-01-31"}`,
		`{"oldest":"2026-01-01","newest":"2027-01-02"}`,
	}
	for _, raw := range cases {
		if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)}); err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
	}
}

func TestGetTrainingPlanSummarizesNestedPlan(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		trainingPlan:      decodeToolTrainingPlan(t, `{"training_plan_id":456,"training_plan_start_date":"2026-02-01","alias":"Base alias","training_plan":{"id":456,"name":"Base plan","children":[{"workout_doc":{"steps":[1]}}],"workout_doc":{"steps":[2]}}}`),
	}
	tool := newGetTrainingPlanTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	row := out["training_plan"].(map[string]any)
	if row["training_plan_id"] != "456" || row["alias"] != "Base alias" {
		t.Fatalf("training_plan row = %#v, want assignment IDs/alias", row)
	}
	if _, ok := row["full"]; ok {
		t.Fatalf("full present in default training plan row")
	}
	summary := row["plan_summary"].(map[string]any)
	if summary["child_count"] != float64(1) || summary["name"] != "Base plan" {
		t.Fatalf("plan_summary = %#v, want lightweight child count/name", summary)
	}
	keys := summary["top_level_keys"].([]any)
	if len(keys) != 2 || keys[0] != "id" || keys[1] != "name" {
		t.Fatalf("top_level_keys = %#v, want typed summary keys without nested payloads", keys)
	}
	if _, ok := summary["workout_doc"]; ok {
		t.Fatalf("plan_summary includes workout_doc: %#v", summary)
	}
}

func TestGetTrainingPlanIncludeFullPreservesRawPayload(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		trainingPlan:      decodeToolTrainingPlan(t, `{"training_plan_id":456,"training_plan_start_date":"2026-02-01","alias":"Base alias","training_plan":{"id":456,"name":"Base plan","children":[],"workouts":[],"custom":"kept"}}`),
	}
	tool := newGetTrainingPlanTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["training_plan"].(map[string]any)
	full := row["full"].(map[string]any)
	nested := full["training_plan"].(map[string]any)
	if nested["custom"] != "kept" || nested["id"] != float64(456) {
		t.Fatalf("full = %#v, want raw nested training plan preserved", full)
	}
	summary := row["plan_summary"].(map[string]any)
	if summary["child_count"] != float64(0) || summary["workout_count"] != float64(0) {
		t.Fatalf("plan_summary = %#v, want explicit empty nested counts preserved", summary)
	}
}

func TestGetTrainingPlanNoActivePlanShape(t *testing.T) {
	t.Parallel()

	client := &fakeEventsTrainingPlanClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
	tool := newGetTrainingPlanTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	unavailable := out["unavailable"].(map[string]any)
	if unavailable["reason"] != "no_active_training_plan" {
		t.Fatalf("unavailable = %#v, want no_active_training_plan", unavailable)
	}
	if _, ok := out["training_plan"]; ok {
		t.Fatalf("training_plan present for no active plan: %#v", out["training_plan"])
	}
}

func decodeToolEvents(t *testing.T, raws ...string) []intervals.Event {
	t.Helper()
	events := make([]intervals.Event, 0, len(raws))
	for _, raw := range raws {
		var event intervals.Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			t.Fatalf("decode event: %v", err)
		}
		events = append(events, event)
	}
	return events
}

func decodeToolTrainingPlan(t *testing.T, raw string) intervals.TrainingPlan {
	t.Helper()
	var plan intervals.TrainingPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		t.Fatalf("decode training plan: %v", err)
	}
	return plan
}
