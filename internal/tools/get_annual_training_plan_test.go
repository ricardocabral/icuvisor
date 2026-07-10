package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakeAnnualTrainingPlanClient struct {
	fakeProfileClient
	events []intervals.Event
	calls  []intervals.ListEventsParams
	err    error
}

func (f *fakeAnnualTrainingPlanClient) ListEvents(_ context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.calls = append(f.calls, params)
	if f.err != nil {
		return nil, f.err
	}
	return append([]intervals.Event(nil), f.events...), nil
}

func TestGetAnnualTrainingPlanRegistrationMetadata(t *testing.T) {
	t.Parallel()

	tool := newGetAnnualTrainingPlanToolWithClock(&fakeAnnualTrainingPlanClient{}, &fakeAnnualTrainingPlanClient{}, "test", "UTC", false, fixedTodayClock())
	if tool.Name != getAnnualTrainingPlanName || !strings.Contains(tool.Description, "annual training plan") || !strings.Contains(tool.Description, "do not manually join raw get_events") {
		t.Fatalf("tool metadata = %#v, want ATP activation hint", tool)
	}
	if tool.EffectiveToolset() != safety.ToolsetFull {
		t.Fatalf("toolset = %q, want full", tool.EffectiveToolset())
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"oldest", "newest", "calendar_id", "limit", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s: %#v", name, props)
		}
	}
	if _, ok := props["resolve"]; ok {
		t.Fatalf("schema includes resolve despite deterministic v1 contract: %#v", props)
	}
}

func TestGetAnnualTrainingPlanExtractsPhasesTargetsNotesAndBridge(t *testing.T) {
	t.Parallel()

	client := &fakeAnnualTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"p1","category":"PLAN","type":"Base","name":"Base phase","description":"Aerobic base","start_date_local":"2026-01-01","end_date_local":"2026-01-28","tags":["base"],"load_target":100}`,
			`{"id":"p2","category":"PLAN","type":"Build","name":"Build phase","start_date_local":"2026-01-29"}`,
			`{"id":"t1","category":"TARGET","name":"Week load","start_date_local":"2026-01-05","load_target":300,"time_target":36000,"distance_target":100000}`,
			`{"id":"t2","category":"TARGET","name":"Extra load","start_date_local":"2026-01-06","load_target":50}`,
			`{"id":"n1","category":"NOTE","name":"Recovery week","description":"Deload and rest","start_date_local":"2026-01-12","end_date_local":"2026-01-18","tags":["recovery"],"plan_applied":"2025-12-15T09:30:00Z"}`,
			`{"id":"w1","category":"WORKOUT","name":"Ignored workout","start_date_local":"2026-01-05","load_target":999}`,
		),
	}
	tool := newGetAnnualTrainingPlanToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-02-15"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	if len(client.calls) != 1 || client.calls[0].Oldest != "2026-01-01" || client.calls[0].Newest != "2026-02-15" || client.calls[0].Limit != annualTrainingPlanEventLimit {
		t.Fatalf("ListEvents calls = %#v, want bounded ATP scan with limit %d", client.calls, annualTrainingPlanEventLimit)
	}
	summary := out["summary"].(map[string]any)
	if summary["phase_count"] != float64(2) || summary["target_event_count"] != float64(2) || summary["atp_note_count"] != float64(1) || summary["context_note_count"] != float64(0) || summary["total_load_target"] != float64(350) {
		t.Fatalf("summary = %#v, want phase/target/ATP-note counts and total load", summary)
	}
	phases := out["phases"].([]any)
	if len(phases) != 2 || phases[0].(map[string]any)["phase_id"] != "phase_p1" || phases[1].(map[string]any)["end_date_source"] != "range_end" {
		t.Fatalf("phases = %#v, want stable phase rows", phases)
	}
	weeks := out["weeks"].([]any)
	var targetWeek map[string]any
	var noteWeek map[string]any
	for _, item := range weeks {
		week := item.(map[string]any)
		switch week["week_start_date"] {
		case "2026-01-05":
			targetWeek = week
		case "2026-01-12":
			noteWeek = week
		}
	}
	if targetWeek == nil || targetWeek["target_event_count"] != float64(2) || targetWeek["load_target"] != float64(350) || targetWeek["time_target_seconds"] != float64(36000) || targetWeek["distance_target_meters"] != float64(100000) {
		t.Fatalf("target week = %#v, want summed TARGET load/time/distance", targetWeek)
	}
	if noteWeek == nil || noteWeek["atp_note_count"] != float64(1) || noteWeek["context_note_count"] != float64(0) {
		t.Fatalf("note week = %#v, want ATP note count without personal context", noteWeek)
	}
	notes := out["notes"].([]any)
	if len(notes) != 1 || notes[0].(map[string]any)["status"] != "atp_generated" || notes[0].(map[string]any)["plan_applied"] != "2025-12-15T09:30:00Z" {
		t.Fatalf("notes = %#v, want terse ATP provenance", notes)
	}
	if _, ok := notes[0].(map[string]any)["recovery_hint"]; ok {
		t.Fatalf("notes = %#v, must not infer recovery semantics from English keywords", notes)
	}
	bridge := out["_meta"].(map[string]any)["projection_bridge"].(map[string]any)
	bridgeRows := bridge["weekly_plan_targets"].([]any)
	if bridge["target_tool"] != getFitnessProjectionName || len(bridgeRows) != 1 || bridgeRows[0].(map[string]any)["week_start_date"] != "2026-01-05" || bridgeRows[0].(map[string]any)["training_load"] != float64(350) {
		t.Fatalf("projection bridge = %#v, want copyable weekly target", bridge)
	}
	if _, ok := phases[0].(map[string]any)["full"]; ok {
		t.Fatalf("terse phase included full: %#v", phases[0])
	}
}

func TestGetAnnualTrainingPlanCurrentPhaseUsesAthleteLocalTodayWhenUTCDateDiffers(t *testing.T) {
	t.Parallel()

	client := &fakeAnnualTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"p1","category":"PLAN","type":"Base","name":"Base phase","start_date_local":"2026-05-01","end_date_local":"2026-05-24"}`,
			`{"id":"p2","category":"PLAN","type":"Build","name":"Build phase","start_date_local":"2026-05-25","end_date_local":"2026-06-30"}`,
		),
	}
	tool := newGetAnnualTrainingPlanToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-06-30"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	summary := out["summary"].(map[string]any)
	if summary["current_phase_id"] != "phase_p1" {
		t.Fatalf("summary = %#v, want current phase from athlete-local 2026-05-24 not UTC 2026-05-25", summary)
	}
	meta := out["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("meta = %#v, want athlete timezone", meta)
	}
}

func TestGetAnnualTrainingPlanEmptyResponseUsesUnavailableAndEmptyArrays(t *testing.T) {
	t.Parallel()

	client := &fakeAnnualTrainingPlanClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}}}
	tool := newGetAnnualTrainingPlanToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-02-15"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	if len(out["phases"].([]any)) != 0 || len(out["weeks"].([]any)) != 0 || len(out["notes"].([]any)) != 0 || len(out["context_notes"].([]any)) != 0 {
		t.Fatalf("empty ATP arrays = phases %#v weeks %#v notes %#v context_notes %#v, want all empty", out["phases"], out["weeks"], out["notes"], out["context_notes"])
	}
	summary := out["summary"].(map[string]any)
	if summary["phase_count"] != float64(0) || summary["week_count"] != float64(0) || summary["target_event_count"] != float64(0) {
		t.Fatalf("empty summary = %#v, want zero counts", summary)
	}
	unavailable := out["unavailable"].(map[string]any)
	if unavailable["reason"] != "no_periodization_events" {
		t.Fatalf("unavailable = %#v, want no_periodization_events", unavailable)
	}
}

func TestGetAnnualTrainingPlanIncludeFullWidensOnlySourceRows(t *testing.T) {
	t.Parallel()

	client := &fakeAnnualTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"p1","category":"PLAN","name":"Base","start_date_local":"2026-01-01","raw_extra":"phase"}`,
			`{"id":"t1","category":"TARGET","name":"Target","start_date_local":"2026-01-05","load_target":100,"raw_extra":"target"}`,
			`{"id":"n1","category":"NOTE","name":"Note","start_date_local":"2026-01-06","plan_applied":"2025-12-15T09:30:00Z","raw_extra":"note"}`,
		),
	}
	tool := newGetAnnualTrainingPlanToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-01-31","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	phase := out["phases"].([]any)[0].(map[string]any)
	if phase["full"].(map[string]any)["raw_extra"] != "phase" {
		t.Fatalf("phase full = %#v, want raw_extra", phase["full"])
	}
	note := out["notes"].([]any)[0].(map[string]any)
	if note["full"].(map[string]any)["raw_extra"] != "note" {
		t.Fatalf("note full = %#v, want raw_extra", note["full"])
	}
	weeks := out["weeks"].([]any)
	var targetEvents []any
	for _, item := range weeks {
		week := item.(map[string]any)
		if week["week_start_date"] == "2026-01-05" {
			targetEvents = week["target_events"].([]any)
		}
	}
	if len(targetEvents) != 1 || targetEvents[0].(map[string]any)["full"].(map[string]any)["raw_extra"] != "target" {
		t.Fatalf("target_events = %#v, want raw target full payload", targetEvents)
	}
}

func TestGetAnnualTrainingPlanSharedBoundaryAndOverlappingNotes(t *testing.T) {
	t.Parallel()

	client := &fakeAnnualTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"p1","category":"PLAN","name":"Base","start_date_local":"2026-01-01","end_date_local":"2026-01-28"}`,
			`{"id":"p2","category":"PLAN","name":"Build","start_date_local":"2026-01-28","end_date_local":"2026-02-28"}`,
			`{"id":"n1","category":"NOTE","name":"Boundary recovery","description":"Recovery day before build","start_date_local":"2026-01-28","plan_applied":"2025-12-15T09:30:00Z"}`,
		),
	}
	tool := newGetAnnualTrainingPlanToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-02-28"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	phases := out["phases"].([]any)
	if phases[0].(map[string]any)["end_date"] != "2026-01-28" || phases[0].(map[string]any)["end_date_source"] != "shared_boundary" {
		t.Fatalf("phase boundary = %#v, want explicit shared boundary preserved", phases[0])
	}
	note := out["notes"].([]any)[0].(map[string]any)
	phaseIDs := planningStringSlice(note["phase_ids"])
	if len(phaseIDs) != 2 || phaseIDs[0] != "phase_p1" || phaseIDs[1] != "phase_p2" {
		t.Fatalf("note phase_ids = %#v, want overlap with both boundary phases", phaseIDs)
	}
	caveats := planningStringSlice(out["_meta"].(map[string]any)["caveats"])
	found := false
	for _, caveat := range caveats {
		if strings.Contains(caveat, "same date the next phase starts") {
			found = true
		}
	}
	if !found {
		t.Fatalf("caveats = %#v, want shared-boundary caveat", caveats)
	}
}

func TestGetAnnualTrainingPlanMalformedAndMissingFieldsDegradeGracefully(t *testing.T) {
	t.Parallel()

	client := &fakeAnnualTrainingPlanClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		events: decodeToolEvents(t,
			`{"id":"bad-plan","category":"PLAN","name":"Bad phase","start_date_local":"not-a-date"}`,
			`{"id":"bad-target","category":"TARGET","name":"Missing date"}`,
			`{"id":"target-missing-load","category":"TARGET","name":"Target without load","start_date_local":"2026-01-12","time_target":18000}`,
			`{"id":"note-ok","category":"NOTE","name":"Coach note","start_date_local":"2026-01-10","plan_applied":"2025-12-15T09:30:00Z"}`,
		),
	}
	tool := newGetAnnualTrainingPlanToolWithClock(client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","newest":"2026-01-31"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	meta := out["_meta"].(map[string]any)
	if meta["malformed_event_count"] != float64(2) || meta["periodization_event_count"] != float64(2) {
		t.Fatalf("meta = %#v, want malformed skipped and valid target/note counted", meta)
	}
	if phases := out["phases"].([]any); len(phases) != 0 {
		t.Fatalf("phases = %#v, want malformed phase skipped", phases)
	}
	weeks := out["weeks"].([]any)
	var targetWeek map[string]any
	for _, item := range weeks {
		week := item.(map[string]any)
		if week["week_start_date"] == "2026-01-12" {
			targetWeek = week
		}
	}
	if targetWeek == nil || targetWeek["target_event_count"] != float64(1) || targetWeek["missing_load_target_count"] != float64(1) || targetWeek["time_target_seconds"] != float64(18000) {
		t.Fatalf("target week = %#v, want missing load counted without dropping other targets", targetWeek)
	}
	bridge := meta["projection_bridge"].(map[string]any)
	if len(bridge["weekly_plan_targets"].([]any)) != 0 || bridge["excluded_week_count"] != float64(1) {
		t.Fatalf("projection bridge = %#v, want missing-load week excluded", bridge)
	}
}
