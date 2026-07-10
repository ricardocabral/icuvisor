package prompts

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMastersPlanReviewEvidenceBoundaries(t *testing.T) {
	t.Parallel()

	prompt := MastersPlanReviewPrompt()
	text := renderPromptText(t, prompt, map[string]string{
		"planned_start":          "2026-05-18",
		"planned_end":            "2026-06-01",
		"history_lookback_days":  "28",
		"baseline_lookback_days": "56",
		"race_date":              "2026-06-07",
		"race_name":              "A Race",
	})

	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "athlete-local sequence", want: "Establish the athlete-local timezone"},
		{name: "non-overlapping windows", want: "Partition non-overlapping personal-baseline/history, completed, planned, and race windows"},
		{name: "partial coverage", want: "label the coverage partial and do not treat it as complete"},
		{name: "race scenario", want: "scenario anchor rather than observed race evidence"},
		{name: "separate observed evidence", want: "Observed tool evidence (tool, athlete-local window, freshness/coverage)"},
		{name: "separate preferences", want: "Athlete-stated preferences (availability and requested duration only)"},
		{name: "separate interpretation", want: "Cautious interpretation"},
		{name: "separate questions", want: "Insufficient evidence and focused questions"},
		{name: "separate proposals", want: "Reviewable proposals"},
		{name: "baseline metadata", want: "status, n_baseline, n_current, min_samples, missing-day counts, freshness_status, caveats, `_meta.method`, and `_meta.formula_ref`"},
		{name: "hard session fallback", want: "Titles, aggregate load, calendar proximity, zones that are absent or invalid, and age cannot classify a session as hard"},
		{name: "projection default boundary", want: "never present default weekly_ramp_pct or recovery_week_cadence values as plan evidence or a masters recommendation"},
		{name: "no inferred constraints", want: "not inferred hard constraints or an implied session count"},
		{name: "ambiguous hard sessions", want: "ambiguous or unavailable hard-session or plan detail"},
		{name: "absent zones", want: "absent or invalid zones"},
		{name: "incomplete history", want: "short, partial, truncated, or missing historical coverage"},
		{name: "wellness freshness", want: "missing, stale, or partial wellness"},
		{name: "provider-native readiness", want: "missing or provider-native readiness"},
		{name: "race gap", want: "missing race context"},
		{name: "projection gap", want: "insufficient explicit projection targets"},
		{name: "gap behavior", want: "name the missing evidence, make no comparison or conclusion for that affected dimension, and ask one focused question"},
		{name: "absolute read-only", want: "This workflow is absolutely read-only: never call write or delete tools, including after approval"},
		{name: "unapplied proposal", want: "conditional, unapplied proposal"},
		{name: "no age policy", want: "Masters is an audience label only, never an age-derived policy or universal cutoff"},
		{name: "no medical claims", want: "Do not make medical, diagnostic, treatment, or injury-risk claims"},
		{name: "no score", want: "do not invent a black-box readiness or risk score"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(text, tc.want) {
				t.Fatalf("masters plan review missing %q:\n%s", tc.want, text)
			}
		})
	}

	wantArguments := []string{"planned_start", "planned_end", "history_lookback_days", "baseline_lookback_days", "race_date", "race_name"}
	if len(prompt.Arguments) != len(wantArguments) {
		t.Fatalf("masters plan review has %d arguments, want %d", len(prompt.Arguments), len(wantArguments))
	}
	for i, want := range wantArguments {
		if prompt.Arguments[i].Name != want {
			t.Fatalf("argument[%d] = %q, want %q", i, prompt.Arguments[i].Name, want)
		}
	}
	for _, tool := range renderedPromptTools(t, prompt, nil) {
		if strings.Contains(tool, "write") || strings.Contains(tool, "delete") {
			t.Fatalf("masters plan review exposes prohibited tool %q", tool)
		}
	}
	for _, want := range []string{"compute_baseline", "get_fitness_projection", "icuvisor_list_advanced_capabilities"} {
		if !strings.Contains(text, want) {
			t.Fatalf("masters plan review missing deterministic/fallback route %q:\n%s", want, text)
		}
	}
}

func TestMastersPlanReviewPortablePackContract(t *testing.T) {
	t.Parallel()

	packBytes, err := os.ReadFile(filepath.Join("..", "..", "docs", "prompts", "client-prompt-packs", "masters-plan-review.md"))
	if err != nil {
		t.Fatalf("read pack: %v", err)
	}
	pack := string(packBytes)
	for _, want := range []string{
		"Registry prompt: `masters_plan_review`",
		"14-day athlete-local planned review from today through day 13",
		"28 completed-history days immediately before planned_start",
		"56 personal-baseline days immediately before that history",
		"strict, paired YYYY-MM-DD window",
		"cannot exceed 90 athlete-local days",
		"defaults to 28, and accepts integers from 1 to 90",
		"defaults to 56, and accepts integers from 1 to 180",
		"race_name is optional but requires race_date",
	} {
		if !strings.Contains(pack, want) {
			t.Fatalf("masters prompt pack missing %q", want)
		}
	}
}

func TestMastersPlanReviewHandlerBoundsAndDefaults(t *testing.T) {
	t.Parallel()

	prompt := MastersPlanReviewPrompt()
	tests := []struct {
		name      string
		arguments map[string]string
		wantScope string
		wantError string
	}{
		{
			name:      "defaults",
			arguments: nil,
			wantScope: "resolve a 14-day athlete-local planned review from today through day 13; use the preceding 28 completed-history days and preceding 56 personal-baseline days without overlap",
		},
		{
			name: "normalizes whitespace",
			arguments: map[string]string{
				"planned_start":          " 2026-05-18 ",
				"planned_end":            " 2026-06-01 ",
				"history_lookback_days":  " 28 ",
				"baseline_lookback_days": " 56 ",
				"race_date":              " 2026-06-07 ",
				"race_name":              " A Race ",
			},
			wantScope: "planned_start=2026-05-18, planned_end=2026-06-01, history_lookback_days=28, baseline_lookback_days=56, race_date=2026-06-07, race_name=A Race",
		},
		{
			name:      "same-day planned window",
			arguments: map[string]string{"planned_start": "2026-05-18", "planned_end": "2026-05-18"},
			wantScope: "planned_start=2026-05-18, planned_end=2026-05-18",
		},
		{
			name:      "ninety-day planned window",
			arguments: map[string]string{"planned_start": "2026-01-01", "planned_end": "2026-03-31"},
			wantScope: "planned_start=2026-01-01, planned_end=2026-03-31",
		},
		{name: "invalid history integer", arguments: map[string]string{"history_lookback_days": "bad"}, wantError: "invalid history_lookback_days; provide an integer from 1 to 90"},
		{name: "zero history", arguments: map[string]string{"history_lookback_days": "0"}, wantError: "invalid history_lookback_days; provide an integer from 1 to 90"},
		{name: "out-of-range history", arguments: map[string]string{"history_lookback_days": "91"}, wantError: "invalid history_lookback_days; provide an integer from 1 to 90"},
		{name: "invalid baseline integer", arguments: map[string]string{"baseline_lookback_days": "bad"}, wantError: "invalid baseline_lookback_days; provide an integer from 1 to 180"},
		{name: "zero baseline", arguments: map[string]string{"baseline_lookback_days": "0"}, wantError: "invalid baseline_lookback_days; provide an integer from 1 to 180"},
		{name: "out-of-range baseline", arguments: map[string]string{"baseline_lookback_days": "181"}, wantError: "invalid baseline_lookback_days; provide an integer from 1 to 180"},
		{name: "start-only planned window", arguments: map[string]string{"planned_start": "2026-05-18"}, wantError: "invalid planned window; provide both planned_start and planned_end as YYYY-MM-DD"},
		{name: "end-only planned window", arguments: map[string]string{"planned_end": "2026-05-18"}, wantError: "invalid planned window; provide both planned_start and planned_end as YYYY-MM-DD"},
		{name: "malformed planned start", arguments: map[string]string{"planned_start": "bad", "planned_end": "2026-05-18"}, wantError: "invalid planned_start; provide YYYY-MM-DD"},
		{name: "malformed planned end", arguments: map[string]string{"planned_start": "2026-05-18", "planned_end": "bad"}, wantError: "invalid planned_end; provide YYYY-MM-DD"},
		{name: "reversed planned window", arguments: map[string]string{"planned_start": "2026-05-19", "planned_end": "2026-05-18"}, wantError: "invalid planned window; planned_end must be on or after planned_start"},
		{name: "overlong planned window", arguments: map[string]string{"planned_start": "2026-01-01", "planned_end": "2026-04-01"}, wantError: "invalid planned window; provide 90 athlete-local days or fewer"},
		{name: "malformed race date", arguments: map[string]string{"race_date": "bad"}, wantError: "invalid race_date; provide YYYY-MM-DD"},
		{name: "name-only race", arguments: map[string]string{"race_name": "A Race"}, wantError: "missing race_date; provide YYYY-MM-DD"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := prompt.Handler(context.Background(), Request{Arguments: tc.arguments})
			if tc.wantError != "" {
				var userErr *UserError
				if !errors.As(err, &userErr) {
					t.Fatalf("Handler() error = %v, want UserError", err)
				}
				if userErr.Message != tc.wantError {
					t.Fatalf("UserError.Message = %q, want %q", userErr.Message, tc.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(result.Messages) != 1 || !strings.Contains(result.Messages[0].Text, "Scope: "+tc.wantScope) {
				t.Fatalf("Handler() scope = %q, want %q", result.Messages, tc.wantScope)
			}
		})
	}
}
