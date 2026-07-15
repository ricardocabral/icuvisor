package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetTodayCompletionLoadEvidenceUsesExactLinksOnly(t *testing.T) {
	t.Parallel()

	client := &fakeTodayClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}},
		activities: decodeActivityPage(t,
			`{"id":"under-target-activity","name":"Activity","start_date_local":"2026-05-24T06:00:00","paired_event_id":"under-target-event","icu_training_load":60}`,
			`{"id":"matching-activity","name":"Activity","start_date_local":"2026-05-24T07:00:00","icu_training_load":80}`,
			`{"id":"zero-load-activity","name":"Activity","start_date_local":"2026-05-24T08:00:00","icu_training_load":0}`,
			`{"id":"missing-target-activity","name":"Activity","start_date_local":"2026-05-24T09:00:00","icu_training_load":30}`,
			`{"id":"missing-load-activity","name":"Activity","start_date_local":"2026-05-24T10:00:00"}`,
			`{"id":"zero-target-activity","name":"Activity","start_date_local":"2026-05-24T11:00:00","icu_training_load":30}`,
			`{"id":"ambiguous-activity-one","name":"Activity","start_date_local":"2026-05-24T12:00:00","paired_event_id":"ambiguous-event","icu_training_load":40}`,
			`{"id":"ambiguous-activity-two","name":"Activity","start_date_local":"2026-05-24T13:00:00","paired_event_id":"ambiguous-event","icu_training_load":41}`,
			`{"id":"unlinked-activity","name":"Activity","start_date_local":"2026-05-24T14:00:00","icu_training_load":90}`,
		),
		events: decodeToolEvents(t,
			`{"id":"under-target-event","category":"WORKOUT","start_date_local":"2026-05-24","load_target":100.5}`,
			`{"id":"matching-event","category":"WORKOUT","start_date_local":"2026-05-24","activity_id":"matching-activity","load_target":80}`,
			`{"id":"zero-load-event","category":"WORKOUT","start_date_local":"2026-05-24","activity_id":"zero-load-activity","load_target":10}`,
			`{"id":"missing-target-event","category":"WORKOUT","start_date_local":"2026-05-24","activity_id":"missing-target-activity"}`,
			`{"id":"missing-load-event","category":"WORKOUT","start_date_local":"2026-05-24","activity_id":"missing-load-activity","load_target":30}`,
			`{"id":"zero-target-event","category":"WORKOUT","start_date_local":"2026-05-24","activity_id":"zero-target-activity","load_target":0}`,
			`{"id":"stale-link-event","category":"WORKOUT","start_date_local":"2026-05-24","activity_id":"activity-not-in-digest","load_target":20}`,
			`{"id":"ambiguous-event","category":"WORKOUT","start_date_local":"2026-05-24","load_target":40}`,
			`{"id":"unlinked-event","category":"WORKOUT","start_date_local":"2026-05-24","load_target":90}`,
		),
	}
	tool := newGetTodayToolWithClock(client, client, nil, nil, nil, nil, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	planned := rowsByEventID(out["planned_events"].([]any))

	assertCompletionLoadEvidence(t, planned["under-target-event"], 100.5, 60, -40.5)
	assertCompletionLoadEvidence(t, planned["matching-event"], 80, 80, 0)
	assertCompletionLoadEvidence(t, planned["zero-load-event"], 10, 0, -10)
	for _, eventID := range []string{"missing-target-event", "missing-load-event", "zero-target-event", "stale-link-event", "ambiguous-event", "unlinked-event"} {
		if _, ok := planned[eventID]["completion_load_evidence"]; ok {
			t.Fatalf("%s comparison = %#v, want omitted", eventID, planned[eventID])
		}
	}

	if got := planned["under-target-event"]["workout_status"]; got != workoutStatusPlanned {
		t.Fatalf("inverse-linked event status = %#v, want unchanged planned status", got)
	}
	if got := planned["matching-event"]["workout_status"]; got != workoutStatusCompletedLinked {
		t.Fatalf("direct-linked event status = %#v, want existing completed_linked status", got)
	}
	if got := planned["stale-link-event"]["workout_status"]; got != workoutStatusCompletedLinked {
		t.Fatalf("stale-link event status = %#v, want existing completed_linked status", got)
	}
	if got := planned["ambiguous-event"]["workout_status"]; got != workoutStatusPlanned {
		t.Fatalf("ambiguous event status = %#v, want unchanged planned status", got)
	}
	if got := planned["unlinked-event"]["workout_status"]; got != workoutStatusPlanned {
		t.Fatalf("same-day unlinked event status = %#v, want planned", got)
	}

	activities := map[string]map[string]any{}
	for _, item := range out["completed_activities"].([]any) {
		row := item.(map[string]any)
		activities[row["activity_id"].(string)] = row
	}
	unlinked := activities["unlinked-activity"]
	if got := unlinked["workout_status"]; got != workoutStatusCompletedUnlinked {
		t.Fatalf("same-day unlinked activity status = %#v, want completed_unlinked", got)
	}
	if !stringSliceContains(stringSliceFromAny(unlinked["workout_status_caveats"]), workoutCaveatUnlinkedActivity) {
		t.Fatalf("same-day unlinked activity caveats = %#v, want %q", unlinked["workout_status_caveats"], workoutCaveatUnlinkedActivity)
	}

	var legacy struct {
		PlannedEvents []struct {
			EventID       string `json:"event_id"`
			WorkoutStatus string `json:"workout_status"`
		} `json:"planned_events"`
	}
	if err := json.Unmarshal([]byte(result.Content[0].Text), &legacy); err != nil {
		t.Fatalf("legacy decode error = %v", err)
	}
	if len(legacy.PlannedEvents) != len(planned) {
		t.Fatalf("legacy planned events = %#v, want %d rows", legacy.PlannedEvents, len(planned))
	}
	if legacy.PlannedEvents[1].EventID != "matching-event" || legacy.PlannedEvents[1].WorkoutStatus != workoutStatusCompletedLinked {
		t.Fatalf("legacy matching event = %#v, want existing fields despite additive evidence", legacy.PlannedEvents[1])
	}
}

func TestGetTodayCompletionLoadEvidenceIsDocumentedInOutputSchema(t *testing.T) {
	t.Parallel()

	description := getTodayOutputSchema()["description"].(string)
	for _, field := range []string{"completion_load_evidence", "planned_load_target", "completed_activity_load", "actual_minus_target_load"} {
		if !strings.Contains(description, field) {
			t.Fatalf("output schema description = %q, want %q", description, field)
		}
	}
}

func assertCompletionLoadEvidence(t *testing.T, row map[string]any, target float64, actual int, delta float64) {
	t.Helper()
	evidence, ok := row["completion_load_evidence"].(map[string]any)
	if !ok {
		t.Fatalf("row = %#v, want completion_load_evidence", row)
	}
	if got := evidence["planned_load_target"]; got != target {
		t.Fatalf("planned_load_target = %#v, want %v", got, target)
	}
	if got := evidence["completed_activity_load"]; got != float64(actual) {
		t.Fatalf("completed_activity_load = %#v, want %d", got, actual)
	}
	if got := evidence["actual_minus_target_load"]; got != delta {
		t.Fatalf("actual_minus_target_load = %#v, want %v", got, delta)
	}
}
