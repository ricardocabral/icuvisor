package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakeApplyTrainingPlanClient struct {
	fakeProfileClient
	folders     []intervals.WorkoutFolder
	workouts    []intervals.Workout
	events      []intervals.Event
	eventPages  [][]intervals.Event
	created     []intervals.Event
	listCalls   []intervals.ListEventsParams
	writeCalls  []intervals.WriteEventParams
	deleteCalls []string
}

func (f *fakeApplyTrainingPlanClient) ListWorkoutFolders(ctx context.Context) ([]intervals.WorkoutFolder, error) {
	return append([]intervals.WorkoutFolder(nil), f.folders...), nil
}

func (f *fakeApplyTrainingPlanClient) ListLibraryWorkouts(ctx context.Context) ([]intervals.Workout, error) {
	return append([]intervals.Workout(nil), f.workouts...), nil
}

func (f *fakeApplyTrainingPlanClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	idx := len(f.listCalls) - 1
	if idx < len(f.eventPages) {
		return append([]intervals.Event(nil), f.eventPages[idx]...), nil
	}
	return append([]intervals.Event(nil), f.events...), nil
}

func (f *fakeApplyTrainingPlanClient) GetEvent(ctx context.Context, eventID string) (intervals.Event, error) {
	return intervals.Event{}, nil
}

func (f *fakeApplyTrainingPlanClient) DeleteEvent(ctx context.Context, eventID string) error {
	f.deleteCalls = append(f.deleteCalls, eventID)
	return nil
}

func (f *fakeApplyTrainingPlanClient) AddOrUpdateEvent(ctx context.Context, params intervals.WriteEventParams) (intervals.Event, error) {
	f.writeCalls = append(f.writeCalls, params)
	idx := len(f.writeCalls) - 1
	if idx < len(f.created) {
		return f.created[idx], nil
	}
	return intervals.Event{ID: "created"}, nil
}

func TestApplyTrainingPlanDryRunProposesEventsWithConflictMarkersAndNoWrites(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.events = decodeToolEvents(t, `{"id":"evt-conflict","category":"WORKOUT","start_date_local":"2026-06-02T00:00:00"}`)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.writeCalls) != 0 || len(client.deleteCalls) != 0 {
		t.Fatalf("writes=%#v deletes=%#v, want no upstream mutations for default dry_run", client.writeCalls, client.deleteCalls)
	}
	if len(client.listCalls) != 1 || client.listCalls[0].Oldest != "2026-06-01" || client.listCalls[0].Newest != "2026-06-02" {
		t.Fatalf("ListEvents calls = %#v, want proposed date range", client.listCalls)
	}
	out := resultMap(t, result)
	rows := out["proposed_events"].([]any)
	if len(rows) != 2 {
		t.Fatalf("proposed_events count = %d, want 2", len(rows))
	}
	first := rows[0].(map[string]any)
	if first["date"] != "2026-06-01" || first["workout_id"] != "w-1" || len(first["conflicts"].([]any)) != 0 {
		t.Fatalf("first proposed event = %#v, want conflict-free day 1", first)
	}
	second := rows[1].(map[string]any)
	conflicts := second["conflicts"].([]any)
	if second["date"] != "2026-06-02" || second["workout_id"] != "w-2" || len(conflicts) != 1 {
		t.Fatalf("second proposed event = %#v, want one conflict", second)
	}
	if conflict := conflicts[0].(map[string]any); conflict["event_id"] != "evt-conflict" || conflict["reason"] != "existing_event_on_date" {
		t.Fatalf("conflict = %#v, want event_id/reason", conflict)
	}
	meta := out["_meta"].(map[string]any)
	if meta["dry_run"] != true || meta["created_count"] != float64(0) || meta["delete_mode"] != "safe" {
		t.Fatalf("meta = %#v, want dry_run safety metadata", meta)
	}
}

func TestApplyTrainingPlanApplySkipExistingCreatesOnlyConflictFreeDays(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.events = decodeToolEvents(t, `{"id":"evt-conflict","category":"WORKOUT","start_date_local":"2026-06-02"}`)
	client.created = decodeToolEvents(t, `{"id":"evt-created","category":"WORKOUT","type":"Ride","name":"Endurance","start_date_local":"2026-06-01","load_target":0,"distance":0,"time_target":3600,"description":null,"calendar_id":null}`)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"skip_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.deleteCalls) != 0 {
		t.Fatalf("delete calls = %#v, want none for skip_existing", client.deleteCalls)
	}
	if len(client.writeCalls) != 1 {
		t.Fatalf("write calls = %d, want only conflict-free day", len(client.writeCalls))
	}
	call := client.writeCalls[0]
	if call.Date != "2026-06-01" || call.Category != "WORKOUT" || call.Type != "Ride" || call.Name != "Endurance" || call.Indoor == nil || !*call.Indoor || call.TargetLoad == nil || *call.TargetLoad != 45 || call.MovingTimeSeconds == nil || *call.MovingTimeSeconds != 3600 {
		t.Fatalf("write call = %#v, want event params from add_or_update_event internals", call)
	}
	out := resultMap(t, result)
	created := out["created_events"].([]any)[0].(map[string]any)
	assertKeyAbsent(t, created, "description")
	assertKeyAbsent(t, created, "calendar_id")
	assertKeyAbsent(t, created, "full")
	if created["load_target"] != float64(0) || created["distance_meters"] != float64(0) {
		t.Fatalf("created event = %#v, want zero target load and distance preserved", created)
	}
	meta := out["_meta"].(map[string]any)
	if meta["created_count"] != float64(1) {
		t.Fatalf("meta = %#v, want created_count 1", meta)
	}
	skipped := meta["skipped"].([]any)
	if len(skipped) != 1 || skipped[0].(map[string]any)["date"] != "2026-06-02" {
		t.Fatalf("skipped = %#v, want conflicted day listed", skipped)
	}
}

func TestApplyTrainingPlanReplaceExistingRequiresFullAndDeletesBeforeCreate(t *testing.T) {
	safeClient := newApplyTrainingPlanTestClient(t)
	safeTool := newApplyTrainingPlanTool(safeClient, safeClient, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	if _, err := safeTool.Handler(context.Background(), Request{Name: safeTool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","conflict_policy":"replace_existing"}`)}); err == nil {
		t.Fatal("Handler() error = nil, want replace_existing rejected outside full delete mode")
	}

	fullClient := newApplyTrainingPlanTestClient(t)
	fullClient.events = decodeToolEvents(t, `{"id":"evt-old","category":"WORKOUT","start_date_local":"2026-06-01"}`)
	fullClient.created = decodeToolEvents(t, `{"id":"evt-new","category":"WORKOUT","type":"Ride","name":"Endurance","start_date_local":"2026-06-01"}`, `{"id":"evt-new-2","category":"WORKOUT","type":"Run","name":"Run","start_date_local":"2026-06-02"}`)
	fullTool := newApplyTrainingPlanTool(fullClient, fullClient, "test", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := fullTool.Handler(context.Background(), Request{Name: fullTool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"replace_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if !reflect.DeepEqual(fullClient.deleteCalls, []string{"evt-old"}) {
		t.Fatalf("delete calls = %#v, want conflicting event deleted", fullClient.deleteCalls)
	}
	if len(fullClient.writeCalls) != 2 {
		t.Fatalf("write calls = %d, want both days created after replacement", len(fullClient.writeCalls))
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["created_count"] != float64(2) || meta["delete_mode"] != "full" {
		t.Fatalf("meta = %#v, want full-mode created count", meta)
	}
	replaced := meta["replaced"].([]any)
	if len(replaced) != 1 || replaced[0].(map[string]any)["date"] != "2026-06-01" {
		t.Fatalf("replaced = %#v, want replaced day metadata", replaced)
	}
}

func TestApplyTrainingPlanReplaceExistingProtectsMixedNonWorkoutConflicts(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t, `{"id":"w-1","name":"Endurance","type":"Ride","folder_id":"plan-1","day":1}`)
	client.events = decodeToolEvents(t,
		`{"id":"evt-workout","category":"WORKOUT","type":"Ride","name":"Old workout","start_date_local":"2026-06-01T00:00:00"}`,
		`{"id":"evt-note","category":"NOTE","type":"Note","name":"Travel note","start_date_local":"2026-06-01T00:00:00"}`,
		`{"id":"evt-race","category":"RACE_A","type":"Ride","name":"A race","start_date_local":"2026-06-01"}`,
		`{"id":"evt-holiday","category":"HOLIDAY","type":"Unavailable","name":"Out of office","start_date_local":"2026-06-01"}`,
	)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"replace_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.deleteCalls) != 0 || len(client.writeCalls) != 0 {
		t.Fatalf("deletes=%#v writes=%#v, want mixed protected day untouched", client.deleteCalls, client.writeCalls)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["created_count"] != float64(0) {
		t.Fatalf("meta = %#v, want no created events", meta)
	}
	skipped := meta["skipped"].([]any)
	if len(skipped) != 1 {
		t.Fatalf("skipped = %#v, want one skipped protected day", skipped)
	}
	conflicts := skipped[0].(map[string]any)["conflicts"].([]any)
	if len(conflicts) != 4 {
		t.Fatalf("conflicts = %#v, want workout plus three protected conflicts", conflicts)
	}
	byID := conflictsByEventID(conflicts)
	if byID["evt-workout"]["protected"] != false || byID["evt-workout"]["category"] != "WORKOUT" {
		t.Fatalf("workout conflict = %#v, want replaceable WORKOUT detail", byID["evt-workout"])
	}
	for _, id := range []string{"evt-note", "evt-race", "evt-holiday"} {
		conflict := byID[id]
		if conflict["protected"] != true || conflict["date"] != "2026-06-01" || conflict["category"] == "" || conflict["name"] == "" || conflict["type"] == "" {
			t.Fatalf("conflict %s = %#v, want protected category/type/name/date details", id, conflict)
		}
	}
}

func TestApplyTrainingPlanRepeatedApplyReportsExactDuplicateAndProtectedRows(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t, `{"id":"w-repeat","name":"Tempo","type":"Ride","folder_id":"plan-1","day":1,"description":"Tempo prescription","tags":["tempo"],"indoor":true,"icu_training_load":75,"moving_time":3600}`)
	client.events = decodeToolEvents(t,
		`{"id":"evt-repeat","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"2026-06-01T00:00:00","description":"Tempo prescription","tags":["tempo"],"indoor":true,"load_target":75,"time_target":3600}`,
		`{"id":"evt-note","category":"NOTE","type":"Note","name":"Do not delete","start_date_local":"2026-06-01T00:00:00"}`,
	)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"skip_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.writeCalls) != 0 {
		t.Fatalf("write calls = %#v, want exact duplicate plus protected row skipped", client.writeCalls)
	}
	out := resultMap(t, result)
	skipped := out["_meta"].(map[string]any)["skipped"].([]any)
	conflicts := skipped[0].(map[string]any)["conflicts"].([]any)
	if len(conflicts) != 2 {
		t.Fatalf("conflicts = %#v, want exact duplicate and note", conflicts)
	}
	byID := conflictsByEventID(conflicts)
	if byID["evt-repeat"]["reason"] != "duplicate_existing_event" || byID["evt-repeat"]["protected"] != true {
		t.Fatalf("duplicate conflict = %#v, want protected duplicate", byID["evt-repeat"])
	}
	if byID["evt-note"]["category"] != "NOTE" || byID["evt-note"]["protected"] != true {
		t.Fatalf("note conflict = %#v, want protected NOTE", byID["evt-note"])
	}
}

func TestApplyTrainingPlanReplaceExistingProtectsRepreflightOnlyConflicts(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t, `{"id":"w-1","name":"Endurance","type":"Ride","folder_id":"plan-1","day":1}`)
	client.eventPages = [][]intervals.Event{
		{},
		decodeToolEvents(t, `{"id":"evt-sick","category":"SICK","type":"Unavailable","name":"Sick","start_date_local":"2026-06-01"}`),
	}
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"replace_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.listCalls) != 2 {
		t.Fatalf("ListEvents calls = %#v, want initial range and per-day re-preflight", client.listCalls)
	}
	if len(client.deleteCalls) != 0 || len(client.writeCalls) != 0 {
		t.Fatalf("deletes=%#v writes=%#v, want re-preflight protected conflict untouched", client.deleteCalls, client.writeCalls)
	}
	conflicts := resultMap(t, result)["_meta"].(map[string]any)["skipped"].([]any)[0].(map[string]any)["conflicts"].([]any)
	if conflict := conflicts[0].(map[string]any); conflict["category"] != "SICK" || conflict["protected"] != true {
		t.Fatalf("conflict = %#v, want protected SICK conflict", conflict)
	}
}

func TestApplyTrainingPlanConflictsProtectRawAndMissingCategories(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t, `{"id":"w-1","name":"Endurance","type":"Ride","folder_id":"plan-1","day":1}`)
	client.events = []intervals.Event{
		{ID: "evt-raw-note", Raw: map[string]any{"category": "NOTE", "type": "Note", "name": "Raw note", "start_date_local": "2026-06-01T00:00:00"}},
		{ID: "evt-missing-category", Name: ptrString("Mystery"), Type: ptrString("Other"), StartDateLocal: ptrString("2026-06-01T00:00:00"), Raw: map[string]any{}},
	}
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeFull), responseShaping{deleteMode: safety.ModeFull, toolset: safety.ToolsetCore})

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"replace_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.deleteCalls) != 0 || len(client.writeCalls) != 0 {
		t.Fatalf("deletes=%#v writes=%#v, want raw/missing category conflicts protected", client.deleteCalls, client.writeCalls)
	}
	conflicts := resultMap(t, result)["_meta"].(map[string]any)["skipped"].([]any)[0].(map[string]any)["conflicts"].([]any)
	byID := conflictsByEventID(conflicts)
	if byID["evt-raw-note"]["category"] != "NOTE" || byID["evt-raw-note"]["protected"] != true {
		t.Fatalf("raw note conflict = %#v, want raw category fallback", byID["evt-raw-note"])
	}
	if _, ok := byID["evt-missing-category"]["category"]; ok {
		t.Fatalf("missing category conflict = %#v, want category omitted", byID["evt-missing-category"])
	}
	if byID["evt-missing-category"]["protected"] != true {
		t.Fatalf("missing category conflict = %#v, want protected", byID["evt-missing-category"])
	}
}

func TestApplyTrainingPlanRepeatedApplySkipsExactExistingPlanEvents(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t, `{"id":"w-repeat","name":"Tempo","type":"Ride","folder_id":"plan-1","day":1,"description":"Tempo prescription","tags":["tempo"],"indoor":true,"icu_training_load":75,"moving_time":3600}`)
	client.events = decodeToolEvents(t, `{"id":"evt-repeat","category":"WORKOUT","type":"Ride","name":"Tempo","start_date_local":"2026-06-01T00:00:00","description":"Tempo prescription","tags":["tempo"],"indoor":true,"load_target":75,"time_target":3600}`)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"skip_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.writeCalls) != 0 {
		t.Fatalf("write calls = %#v, want repeated apply to skip exact duplicate", client.writeCalls)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["created_count"] != float64(0) {
		t.Fatalf("meta = %#v, want no created duplicates", meta)
	}
	skipped := meta["skipped"].([]any)
	if len(skipped) != 1 {
		t.Fatalf("skipped = %#v, want exact duplicate skipped", skipped)
	}
	conflicts := skipped[0].(map[string]any)["conflicts"].([]any)
	if len(conflicts) != 1 || conflicts[0].(map[string]any)["reason"] != "duplicate_existing_event" {
		t.Fatalf("conflicts = %#v, want duplicate_existing_event", conflicts)
	}
}

func TestApplyTrainingPlanSkipsDuplicateSameDayPlannedEvents(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t,
		`{"id":"w-am","name":"AM Ride","type":"Ride","folder_id":"plan-1","day":1,"description":"Endurance"}`,
		`{"id":"w-pm","name":"PM Ride","type":"Ride","folder_id":"plan-1","day":1,"description":"More endurance"}`,
	)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01","dry_run":false,"conflict_policy":"skip_existing"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.writeCalls) != 0 {
		t.Fatalf("write calls = %#v, want duplicate same-day plan entries skipped", client.writeCalls)
	}
	out := resultMap(t, result)
	rows := out["proposed_events"].([]any)
	for _, raw := range rows {
		conflicts := raw.(map[string]any)["conflicts"].([]any)
		if len(conflicts) != 1 || conflicts[0].(map[string]any)["reason"] != "duplicate_plan_date" {
			t.Fatalf("proposed row = %#v, want duplicate_plan_date conflict", raw)
		}
	}
	meta := out["_meta"].(map[string]any)
	skipped := meta["skipped"].([]any)
	if len(skipped) != 2 {
		t.Fatalf("skipped = %#v, want both duplicate same-day planned events skipped", skipped)
	}
}

func TestApplyTrainingPlanRejectsPlanWithoutRelativeDayMetadata(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	client.workouts = decodeToolWorkouts(t, `{"id":"w-no-day","name":"No day","type":"Ride","folder_id":"plan-1"}`)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"plan_id":"plan-1","start_date":"2026-06-01"}`)})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Handler() error = %v, want ErrInvalidInput", err)
	}
}

func TestApplyTrainingPlanRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := newApplyTrainingPlanTestClient(t)
	tool := newApplyTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	if tool.Requirement != RequirementWrite {
		t.Fatalf("requirement = %q, want write", tool.Requirement)
	}
	if strings.Contains(strings.ToLower(tool.Description), "confirm") || !strings.Contains(tool.Description, "dry_run:true") {
		t.Fatalf("description = %q, want safety default language without confirm", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"plan_id", "start_date", "dry_run", "conflict_policy"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s", name)
		}
	}
	enum := props["conflict_policy"].(map[string]any)["enum"].([]string)
	if !reflect.DeepEqual(enum, []string{applyTrainingPlanConflictSkip}) {
		t.Fatalf("safe conflict enum = %#v, want skip_existing only", enum)
	}
}

func conflictsByEventID(conflicts []any) map[string]map[string]any {
	byID := map[string]map[string]any{}
	for _, raw := range conflicts {
		conflict := raw.(map[string]any)
		if eventID, ok := conflict["event_id"].(string); ok {
			byID[eventID] = conflict
		}
	}
	return byID
}

func newApplyTrainingPlanTestClient(t *testing.T) *fakeApplyTrainingPlanClient {
	t.Helper()
	return &fakeApplyTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		folders:           decodeToolWorkoutFolders(t, `{"id":"plan-1","type":"PLAN","name":"Base plan"}`),
		workouts: decodeToolWorkouts(t,
			`{"id":"w-1","name":"Endurance","type":"Ride","folder_id":"plan-1","day":1,"indoor":true,"icu_training_load":45,"moving_time":3600,"workout_doc":{"steps":[{"duration":600}]}}`,
			`{"id":"w-2","name":"Run","type":"Run","folder_id":"plan-1","day":2,"description":"Easy run"}`,
		),
	}
}
