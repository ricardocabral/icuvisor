package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func multisportProposalTool() Tool {
	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	return newProposeAnnualTrainingPlanToolWithClock(client, "test", "UTC", false, fixedSeasonPlanClock())
}

func TestProposeAnnualTrainingPlanTriathlonAllocationReconcilesEveryParent(t *testing.T) {
	t.Parallel()

	tool := multisportProposalTool()
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{
		"start_date":"2026-07-13","goal_date":"2026-08-10","current_weekly_load":500,"target_weekly_load":500,
		"current_weekly_hours":10,"target_hours_per_week":10,"sports":["Ride","Run","Swim"],
		"sport_allocations":[{"sport":" Ride ","load_share_pct":50,"weekly_session_count":4},{"sport":"RUN","load_share_pct":30,"weekly_session_count":3},{"sport":"Swim","load_share_pct":20,"weekly_session_count":2}]
	}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	weeks := out["weekly_targets"].([]any)
	bridge := out["_meta"].(map[string]any)["projection_bridge"].(map[string]any)["weekly_plan_targets"].([]any)
	for i, raw := range weeks {
		week := raw.(map[string]any)
		targets, ok := week["sport_targets"].([]any)
		if !ok || len(targets) != 3 {
			t.Fatalf("week %d sport_targets = %#v, want 3 rows", i, week["sport_targets"])
		}
		wantSports := []string{"ride", "run", "swim"}
		var load, hours float64
		for j, rawTarget := range targets {
			target := rawTarget.(map[string]any)
			if target["sport"] != wantSports[j] {
				t.Fatalf("week %d target order = %#v, want %v", i, targets, wantSports)
			}
			load += target["allocated_load"].(float64)
			hours += target["allocated_hours"].(float64)
			wantSessions := []float64{4, 3, 2}[j]
			if target["requested_weekly_session_count"] != wantSessions {
				t.Fatalf("week %d target %d session count = %v, want %v", i, j, target["requested_weekly_session_count"], wantSessions)
			}
		}
		if load != week["training_load"].(float64) || hours != week["target_hours"].(float64) {
			t.Fatalf("week %d sums = %.12g/%.12g, want parent %.12g/%.12g", i, load, hours, week["training_load"], week["target_hours"])
		}
		bridgeWeek := bridge[i].(map[string]any)
		if bridgeWeek["training_load"] != week["training_load"] {
			t.Fatalf("week %d bridge = %#v, want aggregate parent", i, bridgeWeek)
		}
	}
	assumptions := noticeCodes(out["assumptions"].([]any))
	warnings := noticeCodes(out["warnings"].([]any))
	if !assumptions["sport_allocations_input_only"] || !assumptions["sports_context_input_only"] {
		t.Fatalf("assumptions = %#v, want explicit and legacy context assumptions", assumptions)
	}
	if warnings["multi_sport_not_allocated"] {
		t.Fatalf("warnings = %#v, explicit allocation must replace multi-sport warning", warnings)
	}
}

func TestProposeAnnualTrainingPlanAllocationLargestRemainderAndCaps(t *testing.T) {
	t.Parallel()

	tool := multisportProposalTool()
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{
		"start_date":"2026-07-13","goal_date":"2026-07-13","target_weekly_load":10,"current_weekly_load":10,
		"target_hours_per_week":20,"current_weekly_hours":20,"max_hours_per_week":1,"taper_target_load_pct":100,
		"sport_allocations":[{"sport":"A","load_share_pct":33.333334,"weekly_session_count":0},{"sport":"B","load_share_pct":33.333333,"weekly_session_count":1},{"sport":"C","load_share_pct":33.333333,"weekly_session_count":2}]
	}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	week := resultMap(t, result)["weekly_targets"].([]any)[0].(map[string]any)
	targets := week["sport_targets"].([]any)
	wantLoad := []float64{3.4, 3.3, 3.3}
	wantHours := []float64{0.34, 0.33, 0.33}
	for i, raw := range targets {
		target := raw.(map[string]any)
		if target["allocated_load"] != wantLoad[i] || target["allocated_hours"] != wantHours[i] {
			t.Fatalf("target %d = %#v, want load/hours %.1f/%.2f", i, target, wantLoad[i], wantHours[i])
		}
	}
	if week["target_hours"] != 1.0 {
		t.Fatalf("parent hours = %v, want cap 1", week["target_hours"])
	}
}

func TestProposeAnnualTrainingPlanAllocationReconcilesRecoveryAndTaper(t *testing.T) {
	t.Parallel()

	tool := multisportProposalTool()
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{
		"start_date":"2026-07-13","goal_date":"2026-08-10","target_weekly_load":500,"current_weekly_load":500,
		"target_hours_per_week":10,"current_weekly_hours":10,
		"sport_allocations":[{"sport":"ride","load_share_pct":50,"weekly_session_count":4},{"sport":"swim","load_share_pct":50,"weekly_session_count":2}]
	}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	weeks := resultMap(t, result)["weekly_targets"].([]any)
	for _, raw := range weeks {
		week := raw.(map[string]any)
		targets := week["sport_targets"].([]any)
		var load float64
		for _, target := range targets {
			load += target.(map[string]any)["allocated_load"].(float64)
		}
		if load != week["training_load"].(float64) {
			t.Fatalf("week %v load = %v, want %v", week["week_index"], load, week["training_load"])
		}
		if week["week_index"] == float64(4) && (!week["is_recovery_week"].(bool) || week["training_load"] != 300.0) {
			t.Fatalf("recovery week = %#v, want 300 load and recovery flag", week)
		}
		if week["week_index"] == float64(5) && (!week["is_taper_week"].(bool) || week["training_load"] != 300.0) {
			t.Fatalf("taper week = %#v, want 300 load and taper flag", week)
		}
	}
}

func TestProposeAnnualTrainingPlanAllocationValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		args string
		want string
	}{
		{name: "null", args: `{"goal_date":"2026-07-13","sport_allocations":null}`, want: "must be an array"},
		{name: "empty", args: `{"goal_date":"2026-07-13","sport_allocations":[]}`, want: "at least one"},
		{name: "missing field", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":100}]}`, want: "requires"},
		{name: "blank sport", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":" ","load_share_pct":100,"weekly_session_count":1}]}`, want: "sport must be non-empty"},
		{name: "long sport", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"12345678901234567890123456789012345678901","load_share_pct":100,"weekly_session_count":1}]}`, want: "at most 40"},
		{name: "duplicate normalized", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"Ride","load_share_pct":50,"weekly_session_count":1},{"sport":" ride ","load_share_pct":50,"weekly_session_count":1}]}`, want: "duplicate"},
		{name: "negative share", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":-1,"weekly_session_count":1}]}`, want: "load_share_pct"},
		{name: "too many sessions", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":100,"weekly_session_count":15}]}`, want: "weekly_session_count"},
		{name: "fractional sessions", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":100,"weekly_session_count":1.5}]}`, want: "strict objects"},
		{name: "unknown field", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":100,"weekly_session_count":1,"extra":true}]}`, want: "strict objects"},
		{name: "sum outside tolerance", args: `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":99.999,"weekly_session_count":1}]}`, want: "total 100"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeSeasonPlanProposalRequest(json.RawMessage(tc.args))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("decode() error = %v, want %q", err, tc.want)
			}
		})
	}
	for _, share := range []string{"99.999998", "100.000002"} {
		args := `{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":` + share + `,"weekly_session_count":1}]}`
		if _, err := decodeSeasonPlanProposalRequest(json.RawMessage(args)); err == nil {
			t.Fatalf("share %s accepted, want outside tolerance rejection", share)
		}
	}
	if _, err := decodeSeasonPlanProposalRequest(json.RawMessage(`{"goal_date":"2026-07-13","sport_allocations":[{"sport":"other-discipline","load_share_pct":100,"weekly_session_count":1}]}`)); err != nil {
		t.Fatalf("arbitrary named discipline rejected: %v", err)
	}
}

func TestProposeAnnualTrainingPlanAllocationAbsentPreservesLegacyShape(t *testing.T) {
	t.Parallel()

	tool := multisportProposalTool()
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-07-13","goal_date":"2026-07-20","target_weekly_load":500,"sports":["Ride","Run"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	week := out["weekly_targets"].([]any)[0].(map[string]any)
	if _, ok := week["sport_targets"]; ok {
		t.Fatalf("legacy week = %#v, sport_targets should be omitted", week)
	}
	if !noticeCodes(out["warnings"].([]any))["multi_sport_not_allocated"] {
		t.Fatalf("warnings = %#v, want legacy multi-sport warning", out["warnings"])
	}
}

func TestProposeAnnualTrainingPlanAllocationApplyIgnoresSportTargets(t *testing.T) {
	t.Parallel()

	request, err := decodeSeasonPlanProposalRequest(json.RawMessage(`{"goal_date":"2026-07-13","sport_allocations":[{"sport":"ride","load_share_pct":100,"weekly_session_count":3}]}`))
	if err != nil {
		t.Fatalf("decode() error = %v", err)
	}
	resolved := seasonPlanResolvedInputs{startDate: mustDate(t, "2026-07-13"), goalDate: mustDate(t, "2026-07-13"), goalWeekStart: mustDate(t, "2026-07-13"), currentWeeklyLoad: 100, targetWeeklyLoad: 100, currentWeeklyHours: 2, targetWeeklyHours: 2, taperWeeks: 1, taperTargetLoadPct: 60, hoursSource: "test", loadSource: "test"}
	proposal := buildSeasonPlanProposal(request, resolved, "UTC")
	if err := validateSeasonPlanProposalForApply(proposal); err != nil {
		t.Fatalf("proposal with sport targets rejected by apply boundary: %v", err)
	}
	notes := prepareAnnualTrainingPlanNotes(proposal)
	if len(notes) != len(proposal.Phases) || strings.Contains(notes[0].description, "sport_targets") || strings.Contains(notes[0].description, "workout") {
		t.Fatalf("apply notes = %#v, want phase notes only", notes)
	}
}
