package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

type fakeApplyAnnualTrainingPlanClient struct {
	fakeProfileClient
	events     []intervals.Event
	eventPages [][]intervals.Event
	listCalls  []intervals.ListEventsParams
	writeCalls []intervals.WriteEventParams
	writeErrAt int
}

func (f *fakeApplyAnnualTrainingPlanClient) ListEvents(ctx context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.listCalls = append(f.listCalls, params)
	idx := len(f.listCalls) - 1
	if idx < len(f.eventPages) {
		return append([]intervals.Event(nil), f.eventPages[idx]...), nil
	}
	return append([]intervals.Event(nil), f.events...), nil
}

func (f *fakeApplyAnnualTrainingPlanClient) AddOrUpdateEvent(ctx context.Context, params intervals.WriteEventParams) (intervals.Event, error) {
	f.writeCalls = append(f.writeCalls, params)
	if f.writeErrAt > 0 && len(f.writeCalls) == f.writeErrAt {
		return intervals.Event{}, errors.New("synthetic write failure")
	}
	id := fmt.Sprintf("written-%d", len(f.writeCalls))
	return intervals.Event{ID: id, ExternalID: &params.ExternalID, Category: &params.Category, Name: &params.Name, StartDateLocal: strPtr(params.Date + "T00:00:00"), Description: params.Description}, nil
}

func TestApplyAnnualTrainingPlanDryRunReportsExternalIDsConflictsAndNoWrites(t *testing.T) {
	t.Parallel()

	proposal := sampleApplyAnnualTrainingPlanProposal()
	client := newApplyAnnualTrainingPlanTestClient()
	client.events = decodeToolEvents(t,
		`{"id":"race-1","category":"RACE_A","name":"Tune-up race","start_date_local":"2026-07-13T00:00:00"}`,
		`{"id":"note-1","category":"NOTE","name":"Coach note","start_date_local":"2026-07-20T00:00:00"}`,
	)
	tool := newApplyAnnualTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, nil, "", "")})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.writeCalls) != 0 {
		t.Fatalf("write calls = %#v, want dry-run no writes", client.writeCalls)
	}
	if len(client.listCalls) != 1 || client.listCalls[0].Oldest != "2026-07-13" || client.listCalls[0].Newest != "2026-08-09" {
		t.Fatalf("ListEvents calls = %#v, want proposal date range", client.listCalls)
	}
	out := resultMap(t, result)
	notes := out["proposed_notes"].([]any)
	if len(notes) != 2 {
		t.Fatalf("proposed_notes = %#v, want 2", notes)
	}
	first := notes[0].(map[string]any)
	if first["phase_id"] != "phase_01_base" || !strings.HasPrefix(first["external_id"].(string), applyAnnualTrainingPlanExternalIDPrefix) || first["operation"] != "blocked" {
		t.Fatalf("first proposed note = %#v, want blocked base note with deterministic external_id", first)
	}
	conflicts := first["conflicts"].([]any)
	if len(conflicts) != 1 || conflicts[0].(map[string]any)["category"] != "RACE_A" || conflicts[0].(map[string]any)["protected"] != true {
		t.Fatalf("conflicts = %#v, want protected race", conflicts)
	}
	protected := out["protected_events"].([]any)
	if len(protected) != 2 {
		t.Fatalf("protected_events = %#v, want race and note reported", protected)
	}
	meta := out["_meta"].(map[string]any)
	if meta["dry_run"] != true || meta["writes_performed"] != false || !strings.HasPrefix(meta["preview_token"].(string), applyAnnualTrainingPlanPreviewTokenPrefix) {
		t.Fatalf("meta = %#v, want dry-run preview token and no writes", meta)
	}
}

func TestApplyAnnualTrainingPlanIdempotentRetrySkipsExistingMatchingNote(t *testing.T) {
	t.Parallel()

	proposal := sampleApplyAnnualTrainingPlanProposal()
	existing := existingApplyAnnualPlanEvent(t, proposal, proposal.Phases[0], "owned-1", true)
	client := newApplyAnnualTrainingPlanTestClient()
	client.events = []intervals.Event{existing}
	tool := newApplyAnnualTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))

	preview, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(true), "", "")})
	if err != nil {
		t.Fatalf("preview error = %v", err)
	}
	token := resultMap(t, preview)["_meta"].(map[string]any)["preview_token"].(string)
	commit, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(false), "", token)})
	if err != nil {
		t.Fatalf("commit error = %v", err)
	}
	if len(client.writeCalls) != 1 {
		t.Fatalf("write calls = %#v, want only second phase created and matching external ID skipped", client.writeCalls)
	}
	out := resultMap(t, commit)
	meta := out["_meta"].(map[string]any)
	if meta["idempotent_count"] != float64(1) || meta["writes_performed"] != true {
		t.Fatalf("meta = %#v, want idempotent success plus one create", meta)
	}
}

func TestApplyAnnualTrainingPlanProtectsOwnedNotesUnlessReplacePolicyPermitsUpdate(t *testing.T) {
	t.Parallel()

	proposal := sampleApplyAnnualTrainingPlanProposal()
	existing := existingApplyAnnualPlanEvent(t, proposal, proposal.Phases[0], "owned-1", false)

	safeClient := newApplyAnnualTrainingPlanTestClient()
	safeClient.events = []intervals.Event{existing}
	safeTool := newApplyAnnualTrainingPlanTool(safeClient, safeClient, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	safePreview, err := safeTool.Handler(context.Background(), Request{Name: safeTool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(true), applyAnnualTrainingPlanConflictSkip, "")})
	if err != nil {
		t.Fatalf("safe preview error = %v", err)
	}
	safeFirst := resultMap(t, safePreview)["proposed_notes"].([]any)[0].(map[string]any)
	if safeFirst["operation"] != "blocked" || safeFirst["conflicts"].([]any)[0].(map[string]any)["protected"] != true {
		t.Fatalf("safe first note = %#v, want owned note protected under skip_existing", safeFirst)
	}

	fullClient := newApplyAnnualTrainingPlanTestClient()
	fullClient.events = []intervals.Event{existing}
	fullTool := newApplyAnnualTrainingPlanTool(fullClient, fullClient, "test", "UTC", false, safety.NewCapability(safety.ModeFull))
	preview, err := fullTool.Handler(context.Background(), Request{Name: fullTool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(true), applyAnnualTrainingPlanConflictReplaceOwned, "")})
	if err != nil {
		t.Fatalf("full preview error = %v", err)
	}
	token := resultMap(t, preview)["_meta"].(map[string]any)["preview_token"].(string)
	_, err = fullTool.Handler(context.Background(), Request{Name: fullTool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(false), applyAnnualTrainingPlanConflictReplaceOwned, token)})
	if err != nil {
		t.Fatalf("full commit error = %v", err)
	}
	if len(fullClient.writeCalls) != 2 || fullClient.writeCalls[0].EventID != "owned-1" {
		t.Fatalf("write calls = %#v, want first phase updated by preflighted event ID and second created", fullClient.writeCalls)
	}
}

func TestApplyAnnualTrainingPlanPreviewTokenGateRejectsMissingMismatchAndStaleBeforeWrites(t *testing.T) {
	t.Parallel()

	proposal := sampleApplyAnnualTrainingPlanProposal()
	for _, tc := range []struct {
		name  string
		token string
	}{
		{name: "missing"},
		{name: "mismatch", token: "season-plan-preview-v1-not-the-token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newApplyAnnualTrainingPlanTestClient()
			tool := newApplyAnnualTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(false), "", tc.token)})
			if err == nil {
				t.Fatal("commit error = nil, want token rejection")
			}
			if len(client.writeCalls) != 0 {
				t.Fatalf("write calls = %#v, want no writes", client.writeCalls)
			}
		})
	}

	staleClient := newApplyAnnualTrainingPlanTestClient()
	staleClient.eventPages = [][]intervals.Event{nil, decodeToolEvents(t, `{"id":"race-1","category":"RACE_A","name":"Race","start_date_local":"2026-07-13T00:00:00"}`)}
	staleTool := newApplyAnnualTrainingPlanTool(staleClient, staleClient, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	preview, err := staleTool.Handler(context.Background(), Request{Name: staleTool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(true), "", "")})
	if err != nil {
		t.Fatalf("preview error = %v", err)
	}
	staleToken := resultMap(t, preview)["_meta"].(map[string]any)["preview_token"].(string)
	_, err = staleTool.Handler(context.Background(), Request{Name: staleTool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(false), "", staleToken)})
	if err == nil {
		t.Fatal("stale commit error = nil, want token rejection after changed conflicts")
	}
	if len(staleClient.writeCalls) != 0 {
		t.Fatalf("stale write calls = %#v, want no writes", staleClient.writeCalls)
	}
}

func TestApplyAnnualTrainingPlanPartialWriteFailureReturnsStructuredRetrySafeResult(t *testing.T) {
	t.Parallel()

	proposal := sampleApplyAnnualTrainingPlanProposal()
	client := newApplyAnnualTrainingPlanTestClient()
	client.writeErrAt = 2
	tool := newApplyAnnualTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
	preview, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(true), "", "")})
	if err != nil {
		t.Fatalf("preview error = %v", err)
	}
	token := resultMap(t, preview)["_meta"].(map[string]any)["preview_token"].(string)
	commit, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(false), "", token)})
	if err != nil {
		t.Fatalf("commit error = %v", err)
	}
	if !commit.IsError {
		t.Fatalf("commit IsError = false, want structured partial failure")
	}
	out := resultMapAllowError(t, commit)
	meta := out["_meta"].(map[string]any)
	if meta["retry_safe"] != true || meta["failed_external_id"] == "" || len(meta["applied_external_ids"].([]any)) != 1 {
		t.Fatalf("meta = %#v, want retry-safe partial failure metadata", meta)
	}
}

func TestApplyAnnualTrainingPlanRejectsInvalidProposalBeforeCalendarIO(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		mutate func(*seasonPlanProposalResponse)
	}{
		{name: "not read only", mutate: func(p *seasonPlanProposalResponse) { p.Meta.ReadOnly = false }},
		{name: "writes performed", mutate: func(p *seasonPlanProposalResponse) { p.Meta.WritesPerformed = true }},
		{name: "phase count mismatch", mutate: func(p *seasonPlanProposalResponse) { p.Summary.PhaseCount = 99 }},
		{name: "unknown weekly phase", mutate: func(p *seasonPlanProposalResponse) {
			p.WeeklyTargets = []seasonPlanProposalWeeklyTarget{{WeekStartDate: "2026-07-13", WeekEndDate: "2026-07-19", PhaseID: "missing"}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proposal := sampleApplyAnnualTrainingPlanProposal()
			tc.mutate(&proposal)
			client := newApplyAnnualTrainingPlanTestClient()
			tool := newApplyAnnualTrainingPlanTool(client, client, "test", "UTC", false, safety.NewCapability(safety.ModeSafe))
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: applyAnnualTrainingPlanArgs(t, proposal, boolPtr(true), "", "")})
			if err == nil {
				t.Fatal("Handler() error = nil, want invalid proposal rejection")
			}
			if len(client.listCalls) != 0 || len(client.writeCalls) != 0 {
				t.Fatalf("list=%#v writes=%#v, want validation before calendar I/O", client.listCalls, client.writeCalls)
			}
		})
	}
}

func newApplyAnnualTrainingPlanTestClient() *fakeApplyAnnualTrainingPlanClient {
	return &fakeApplyAnnualTrainingPlanClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
}

func sampleApplyAnnualTrainingPlanProposal() seasonPlanProposalResponse {
	return seasonPlanProposalResponse{
		Summary: seasonPlanProposalSummary{StartDate: "2026-07-13", EndDate: "2026-08-09", GoalDate: "2026-08-03", TotalWeeks: 4, PhaseCount: 2, RecoveryWeekCount: 0, RaceAnchorCount: 1, TargetWeeklyLoad: 420, TargetHoursPerWeek: 8.4},
		Phases: []seasonPlanProposalPhase{
			{PhaseID: "phase_01_base", PhaseType: "base", Name: "Base", StartDate: "2026-07-13", EndDate: "2026-07-26", WeekCount: 2, StartWeekIndex: 1, EndWeekIndex: 2},
			{PhaseID: "phase_02_taper", PhaseType: "race_taper", Name: "Race taper", StartDate: "2026-07-27", EndDate: "2026-08-09", WeekCount: 2, StartWeekIndex: 3, EndWeekIndex: 4},
		},
		WeeklyTargets: []seasonPlanProposalWeeklyTarget{
			{WeekStartDate: "2026-07-13", WeekEndDate: "2026-07-19", WeekIndex: 1, PhaseID: "phase_01_base", PhaseType: "base", TrainingLoad: 300, TargetHours: 6, LoadSource: "test", HoursSource: "test"},
			{WeekStartDate: "2026-07-27", WeekEndDate: "2026-08-02", WeekIndex: 3, PhaseID: "phase_02_taper", PhaseType: "race_taper", TrainingLoad: 240, TargetHours: 4.8, LoadSource: "test", HoursSource: "test"},
		},
		RaceAnchors: []seasonPlanProposalRaceAnchor{{Date: "2026-08-03", Type: "race", Source: "input", WeekStartDate: "2026-08-03"}},
		Assumptions: []seasonPlanProposalNotice{},
		Warnings:    []seasonPlanProposalNotice{},
		Meta:        seasonPlanProposalMeta{SchemaVersion: seasonPlanProposalSchemaVersion, ReadOnly: true, WritesPerformed: false, Timezone: "UTC"},
	}
}

func applyAnnualTrainingPlanArgs(t *testing.T, proposal seasonPlanProposalResponse, dryRun *bool, conflictPolicy string, previewToken string) json.RawMessage {
	t.Helper()
	payload := map[string]any{"proposal": proposal}
	if dryRun != nil {
		payload["dry_run"] = *dryRun
	}
	if conflictPolicy != "" {
		payload["conflict_policy"] = conflictPolicy
	}
	if previewToken != "" {
		payload["preview_token"] = previewToken
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}
	return data
}

func existingApplyAnnualPlanEvent(t *testing.T, proposal seasonPlanProposalResponse, phase seasonPlanProposalPhase, eventID string, exact bool) intervals.Event {
	t.Helper()
	notes := prepareAnnualTrainingPlanNotes(proposal)
	var note applyAnnualTrainingPlanPreparedNote
	for _, candidate := range notes {
		if candidate.phase.PhaseID == phase.PhaseID {
			note = candidate
			break
		}
	}
	description := note.description
	name := note.params.Name
	if !exact {
		description = applyAnnualTrainingPlanNoteMarker + "\nold body"
		name = name + " old"
	}
	return intervals.Event{ID: eventID, ExternalID: &note.params.ExternalID, Category: strPtr("NOTE"), Name: &name, StartDateLocal: strPtr(note.params.Date + "T00:00:00"), Description: &description}
}

func resultMapAllowError(t *testing.T, result Result) map[string]any {
	t.Helper()
	if len(result.Content) != 1 {
		t.Fatalf("content count = %d, want 1", len(result.Content))
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].Text), &out); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return out
}

func strPtr(value string) *string { return &value }
