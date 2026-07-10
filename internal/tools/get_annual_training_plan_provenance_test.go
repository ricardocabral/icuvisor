package tools

import (
	"strings"
	"testing"
)

func TestAnnualTrainingPlanClassifiesNotesByPlanAppliedProvenance(t *testing.T) {
	t.Parallel()

	const planApplied = "2025-12-15T09:30:00Z"
	events := decodeToolEvents(t,
		`{"id":"context-later","category":"NOTE","name":"Travel — Rest","description":"Personal calendar note","start_date_local":"2026-01-12","end_date_local":"2026-01-18","updated":"2026-01-02T10:00:00Z","plan_applied":"   "}`,
		`{"id":"atp-later","category":"NOTE","name":"Semana de recuperação","description":"Reduzir o volume","start_date_local":"2026-01-11","end_date_local":"2026-01-20","updated":"2026-01-02T10:00:00Z","plan_applied":"2025-12-15T09:30:00Z"}`,
		`{"id":"context-first","category":"NOTE","name":"Rest before flight","start_date_local":"2026-01-12","updated":"2026-01-02T09:00:00Z","plan_applied":null}`,
		`{"id":"atp-first","category":"NOTE","name":"Semana de descarga","description":"Bajar el volumen","start_date_local":"2026-01-11","updated":"2026-01-02T09:00:00Z","plan_applied":"2025-12-15T09:30:00Z"}`,
		`{"id":"context-empty","category":"NOTE","name":"Family day","start_date_local":"2026-01-13","plan_applied":""}`,
	)

	response := shapeAnnualTrainingPlanResponse(events, annualTrainingPlanRequest{Oldest: "2026-01-05", Newest: "2026-01-25", Limit: annualTrainingPlanEventLimit}, "UTC", "2026-01-10", false)
	if response.Summary.ATPNoteCount != 2 || response.Summary.ContextNoteCount != 3 {
		t.Fatalf("summary = %#v, want 2 ATP notes and 3 personal context notes", response.Summary)
	}
	if response.Meta.ATPNoteEventCount != 2 || response.Meta.ContextNoteEventCount != 3 || response.Meta.PeriodizationEventCount != 2 {
		t.Fatalf("meta = %#v, want provenance-separated event counts", response.Meta)
	}
	if got := noteSourceIDs(response.Notes); strings.Join(got, ",") != "atp-first,atp-later" {
		t.Fatalf("ATP note order = %#v, want start/updated/ID order", got)
	}
	if got := noteSourceIDs(response.ContextNotes); strings.Join(got, ",") != "context-first,context-later,context-empty" {
		t.Fatalf("context note order = %#v, want start/updated/ID order", got)
	}
	for _, note := range response.Notes {
		if note.Status != "atp_generated" || note.PlanApplied != planApplied || note.Full != nil {
			t.Fatalf("ATP note = %#v, want terse visible provenance", note)
		}
	}
	for _, note := range response.ContextNotes {
		if note.Status != "personal_context" || note.PlanApplied != "" || note.Full != nil {
			t.Fatalf("context note = %#v, want neutral terse context", note)
		}
	}

	weekByStart := annualTrainingPlanWeeksByStart(response.Weeks)
	if week := weekByStart["2026-01-12"]; week.ATPNoteCount != 1 || week.ContextNoteCount != 3 || strings.Join(week.ATPNoteIDs, ",") != "note_atp-later" || strings.Join(week.ContextNoteIDs, ",") != "context_note_context-empty,context_note_context-first,context_note_context-later" {
		t.Fatalf("2026-01-12 week = %#v, want separate sorted ATP/context associations", week)
	}
	if week := weekByStart["2026-01-19"]; week.ATPNoteCount != 1 || week.ContextNoteCount != 0 {
		t.Fatalf("2026-01-19 week = %#v, want multi-day ATP note association only", week)
	}
}

func TestAnnualTrainingPlanPersonalContextOnlyRetainsContextAndUnavailable(t *testing.T) {
	t.Parallel()

	events := decodeToolEvents(t,
		`{"id":"travel","category":"NOTE","name":"Travel — Rest","start_date_local":"2026-01-12","plan_applied":null}`,
	)
	response := shapeAnnualTrainingPlanResponse(events, annualTrainingPlanRequest{Oldest: "2026-01-12", Newest: "2026-01-18", Limit: annualTrainingPlanEventLimit}, "UTC", "2026-01-12", false)

	if len(response.Notes) != 0 || len(response.ContextNotes) != 1 || response.ContextNotes[0].Status != "personal_context" {
		t.Fatalf("notes = %#v context_notes = %#v, want retained personal context only", response.Notes, response.ContextNotes)
	}
	if response.Unavailable == nil || response.Unavailable.Reason != "no_periodization_events" || !strings.Contains(response.Unavailable.Detail, "context_notes") {
		t.Fatalf("unavailable = %#v, want truthful personal-context-only detail", response.Unavailable)
	}
	if response.Meta.PeriodizationEventCount != 0 || response.Meta.ContextNoteEventCount != 1 || response.Weeks[0].ATPNoteCount != 0 || response.Weeks[0].ContextNoteCount != 1 {
		t.Fatalf("response = %#v, personal context must not contribute to ATP counts", response)
	}
}

func TestAnnualTrainingPlanNoteFullPayloadRespectsIncludeFull(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		includeFull bool
		wantFull    bool
	}{
		{name: "terse", includeFull: false, wantFull: false},
		{name: "full", includeFull: true, wantFull: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			events := decodeToolEvents(t,
				`{"id":"atp","category":"NOTE","name":"Recuperação","start_date_local":"2026-01-12","plan_applied":"2025-12-15T09:30:00Z","raw_extra":"atp"}`,
				`{"id":"context","category":"NOTE","name":"Travel — Rest","start_date_local":"2026-01-13","plan_applied":null,"raw_extra":"context"}`,
			)
			response := shapeAnnualTrainingPlanResponse(events, annualTrainingPlanRequest{Oldest: "2026-01-12", Newest: "2026-01-18", Limit: annualTrainingPlanEventLimit, IncludeFull: tc.includeFull}, "UTC", "2026-01-12", false)
			for _, note := range append(append([]annualTrainingPlanNote{}, response.Notes...), response.ContextNotes...) {
				if (note.Full != nil) != tc.wantFull {
					t.Fatalf("note = %#v, full presence want %t", note, tc.wantFull)
				}
			}
			if tc.wantFull && (response.Notes[0].Full["raw_extra"] != "atp" || response.ContextNotes[0].Full["raw_extra"] != "context") {
				t.Fatalf("full rows = %#v %#v, want unchanged raw payloads", response.Notes[0].Full, response.ContextNotes[0].Full)
			}
		})
	}
}

func TestAnnualTrainingPlanOneDayTargetKeepsISOWeekAndProjectionBridge(t *testing.T) {
	t.Parallel()

	events := decodeToolEvents(t,
		`{"id":"target","category":"TARGET","name":"Carga semanal","start_date_local":"2026-01-14","end_date_local":"2026-01-14","load_target":420}`,
		`{"id":"context","category":"NOTE","name":"Travel — Rest","start_date_local":"2026-01-15","plan_applied":null}`,
	)
	response := shapeAnnualTrainingPlanResponse(events, annualTrainingPlanRequest{Oldest: "2026-01-12", Newest: "2026-01-18", Limit: annualTrainingPlanEventLimit}, "UTC", "2026-01-12", false)

	if len(response.Weeks) != 1 || response.Weeks[0].WeekStartDate != "2026-01-12" || response.Weeks[0].WeekEndDate != "2026-01-18" || response.Weeks[0].PartialWeek {
		t.Fatalf("weeks = %#v, want Monday-through-Sunday boundary for one-day TARGET", response.Weeks)
	}
	bridge := response.Meta.ProjectionBridge
	if len(bridge.WeeklyPlanTargets) != 1 || bridge.WeeklyPlanTargets[0].WeekStartDate != "2026-01-12" || bridge.WeeklyPlanTargets[0].TrainingLoad != 420 || bridge.IncludedWeekCount != 1 {
		t.Fatalf("bridge = %#v, want unchanged explicit TARGET-only row", bridge)
	}
	if response.Weeks[0].ContextNoteCount != 1 || response.Weeks[0].ATPNoteCount != 0 {
		t.Fatalf("week = %#v, personal context must stay outside ATP note counts", response.Weeks[0])
	}
}

func noteSourceIDs(notes []annualTrainingPlanNote) []string {
	ids := make([]string, 0, len(notes))
	for _, note := range notes {
		ids = append(ids, note.SourceEventID)
	}
	return ids
}

func annualTrainingPlanWeeksByStart(weeks []annualTrainingPlanWeek) map[string]annualTrainingPlanWeek {
	byStart := make(map[string]annualTrainingPlanWeek, len(weeks))
	for _, week := range weeks {
		byStart[week.WeekStartDate] = week
	}
	return byStart
}
