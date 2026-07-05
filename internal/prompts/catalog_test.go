package prompts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type captureRegistrar struct {
	prompts []Prompt
}

func (r *captureRegistrar) AddPrompt(prompt Prompt) error {
	r.prompts = append(r.prompts, prompt)
	return nil
}

func TestNewRegistryRegistersNinePrompts(t *testing.T) {
	t.Parallel()

	registrar := &captureRegistrar{}
	if err := NewRegistry().Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	wantNames := []string{TrainingAnalysisName, RecoveryCheckName, WeeklyPlanningName, WeeklyReviewName, ShareableTrainingReportName, PlanHealthReviewName, RaceWeekTaperName, CoachRosterTriageName, CoachAthleteOnboardingName}
	if len(registrar.prompts) != len(wantNames) {
		t.Fatalf("registered %d prompts, want %d", len(registrar.prompts), len(wantNames))
	}
	for i, want := range wantNames {
		prompt := registrar.prompts[i]
		if prompt.Name != want {
			t.Fatalf("prompt[%d].Name = %q, want %q", i, prompt.Name, want)
		}
		if prompt.Title == "" || prompt.Description == "" || prompt.Handler == nil {
			t.Fatalf("prompt[%d] incomplete metadata: %#v", i, prompt)
		}
		for _, arg := range prompt.Arguments {
			if arg.Description == "" {
				t.Fatalf("prompt %s argument %s missing description", prompt.Name, arg.Name)
			}
		}
	}
}

func TestRenderedPromptsGolden(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		prompt     Prompt
		arguments  map[string]string
		goldenFile string
	}{
		{name: "training_analysis", prompt: TrainingAnalysisPrompt(), arguments: map[string]string{"start_date": "2026-04-01", "end_date": "2026-04-30"}, goldenFile: "training_analysis.md"},
		{name: "recovery_check", prompt: RecoveryCheckPrompt(), arguments: map[string]string{"date": "2026-05-14", "lookback_days": "10"}, goldenFile: "recovery_check.md"},
		{name: "weekly_planning", prompt: WeeklyPlanningPrompt(), arguments: map[string]string{"week_start": "2026-05-18"}, goldenFile: "weekly_planning.md"},
		{name: "weekly_review", prompt: WeeklyReviewPrompt(), arguments: nil, goldenFile: "weekly_review.md"},
		{name: "shareable_training_report", prompt: ShareableTrainingReportPrompt(), arguments: map[string]string{"report_type": "race_prep", "start_date": "2026-05-01", "end_date": "2026-06-07", "race_date": "2026-06-07", "audience": "family"}, goldenFile: "shareable_training_report.md"},
		{name: "plan_health_review", prompt: PlanHealthReviewPrompt(), arguments: map[string]string{"planned_start": "2026-05-18", "planned_end": "2026-06-01", "completed_lookback_days": "21", "race_date": "2026-06-07", "race_name": "A Race"}, goldenFile: "plan_health_review.md"},
		{name: "race_week_taper", prompt: RaceWeekTaperPrompt(), arguments: map[string]string{"race_date": "2026-06-07", "race_name": "A Race"}, goldenFile: "race_week_taper.md"},
		{name: "coach_roster_triage", prompt: CoachRosterTriagePrompt(), arguments: map[string]string{"athlete_id": "i12345", "start_date": "2026-05-01", "end_date": "2026-05-14"}, goldenFile: "coach_roster_triage.md"},
		{name: "coach_athlete_onboarding", prompt: CoachAthleteOnboardingPrompt(), arguments: map[string]string{"athlete_id": "i12345", "start_date": "2026-05-01", "end_date": "2026-05-28"}, goldenFile: "coach_athlete_onboarding.md"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := renderPromptText(t, tc.prompt, tc.arguments)
			want, err := os.ReadFile(filepath.Join("testdata", tc.goldenFile))
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if got != strings.TrimRight(string(want), "\n") {
				t.Fatalf("rendered prompt mismatch with %s\n--- got ---\n%s\n--- want ---\n%s", tc.goldenFile, got, string(want))
			}
		})
	}
}

func TestTrainingAnalysisPromptIncludesHypoxicContextGuardrail(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, TrainingAnalysisPrompt(), nil)
	for _, want := range []string{
		"If the user explicitly mentions hypoxic training",
		"power-based load may under-represent extra hypoxic strain",
		"HR/RPE/feel/recovery can be supporting context",
		"must not apply a hypoxia multiplier",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("training analysis prompt missing %q:\n%s", want, text)
		}
	}
}

func TestReadinessPromptsRequireProviderNativeLabels(t *testing.T) {
	t.Parallel()

	for _, prompt := range []Prompt{RecoveryCheckPrompt(), WeeklyReviewPrompt()} {
		text := renderPromptText(t, prompt, nil)
		for _, want := range []string{
			"_meta.provenance.readiness.source",
			"native_scale",
			"Garmin Body Battery",
			"Oura readiness",
			"Polar nightly recharge/ANS charge",
			"WHOOP recovery",
			"not a universal recovery score",
			"do not invent",
			"readiness score",
		} {
			if !strings.Contains(text, want) {
				t.Fatalf("%s prompt missing %q:\n%s", prompt.Name, want, text)
			}
		}
	}
}

func TestRecoveryCheckIncludesWeatherAndIndoorOutdoorGuardrails(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, RecoveryCheckPrompt(), nil)
	for _, want := range []string{
		"get_today",
		"weather.status/provenance",
		"forecast_unavailable",
		"do not invent conditions",
		"planned_events[].indoor",
		"Do not infer separate indoor/outdoor FTP",
		"sport_settings[].indoor_ftp_watts only when present",
		"otherwise ask or confirm how to adjust the workout",
		"zone boundaries",
		"do not create a second active workout",
		"Do not call write or delete tools for indoor/outdoor adaptation",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("recovery prompt missing %q:\n%s", want, text)
		}
	}
}

func TestWeeklyReviewRendersExplicitArguments(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, WeeklyReviewPrompt(), map[string]string{
		"week_start":        "2026-05-18",
		"lookback_days":     "14",
		"include_next_week": "true",
	})
	want := "Scope: week_start=2026-05-18, lookback_days=14, include_next_week=true."
	if !strings.Contains(text, want) {
		t.Fatalf("weekly review prompt text = %q, want %q", text, want)
	}
}

func TestWeeklyReviewIncludesFallbackAndSafetyGuidance(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, WeeklyReviewPrompt(), nil)
	for _, want := range []string{
		"athlete-local timezone",
		"do not include wellness, activities, or summary rows after that end date",
		"current-day `_meta.as_of` as partial-day context only",
		"icuvisor_list_advanced_capabilities",
		"workout_status",
		"status counts",
		"_meta.stale",
		"provenance warnings",
		"Do not call write or delete tools unless the user explicitly approves the exact change first.",
		"Do not auto-fill calendars or create ATP notes from the review",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("weekly review prompt missing %q:\n%s", want, text)
		}
	}
}

func TestShareableTrainingReportIncludesPrivacyAndReviewGuidance(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, ShareableTrainingReportPrompt(), map[string]string{
		"report_type": "training_journey",
		"start_date":  "2026-01-01",
		"end_date":    "2026-06-01",
		"audience":    "newsletter",
	})
	for _, want := range []string{
		"Scope: report_type=training_journey, start_date=2026-01-01, end_date=2026-06-01, audience=newsletter.",
		"Markdown first",
		"simple static HTML in chat",
		"does not generate, publish, upload, or host HTML",
		"review and redact private health, location, notes, identifiers, and race logistics",
		"Do not request or accept intervals.icu API keys in chat.",
		"do not use include_full, raw streams, or heavy payloads unless the user explicitly asks or evidence is missing",
		"Do not publish, host, upload, auto-share, or connect to social platforms",
		"athlete manually shares only after review",
		"Do not invent missing metrics",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("shareable training report prompt missing %q:\n%s", want, text)
		}
	}
}

func TestPlanHealthReviewIncludesTransparentRiskGuidance(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, PlanHealthReviewPrompt(), nil)
	for _, want := range []string{
		"icuvisor://analysis-formulas",
		"compute_compliance_rate",
		"get_fitness_projection",
		"completed-lookback, planned-window, and race-scenario dates",
		"resolve_calendar_dates",
		"instead of UTC, client-time, or model arithmetic",
		"do not mix current-day or post-window wellness into completed adherence evidence",
		"current-day `_meta.as_of` as partial-day context only",
		"_meta.method",
		"_meta.assumptions",
		"missed/planned/future/completed status counts",
		"before calling anything skipped, missed, or completed",
		"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
		"Do not invent a black-box plan-health score",
		"no race event is found",
		"reviewed and approved the exact proposal first",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("plan health review prompt missing %q:\n%s", want, text)
		}
	}
}

func TestPlanningPromptsIncludeSeasonContextAndWriteGuardrails(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		prompt Prompt
		args   map[string]string
	}{
		{name: "weekly planning", prompt: WeeklyPlanningPrompt(), args: map[string]string{"week_start": "2026-05-18"}},
		{name: "race-week taper", prompt: RaceWeekTaperPrompt(), args: map[string]string{"race_date": "2026-06-07"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			text := renderPromptText(t, tc.prompt, tc.args)
			for _, want := range []string{
				"priority/category",
				"get_training_plan",
				"resolve_calendar_dates",
				"compute_compliance_rate",
				"icuvisor_list_advanced_capabilities",
				"workout_status",
				"Do not automatically fill the calendar, create ATP notes, or call write/delete tools",
				"approval of exact changes",
				"instead of UTC, client-time, or model arithmetic",
			} {
				if !strings.Contains(text, want) {
					t.Fatalf("%s prompt missing %q:\n%s", tc.name, want, text)
				}
			}
			if tc.name == "weekly planning" {
				for _, want := range []string{
					"total duration, key steps, target intensities, load/distance/time changes",
					"what existing title/prose/tags/structured steps are preserved",
				} {
					if !strings.Contains(text, want) {
						t.Fatalf("%s prompt missing workout preview guidance %q:\n%s", tc.name, want, text)
					}
				}
			}
			if tc.name == "race-week taper" && !strings.Contains(text, "get_fitness_projection") {
				t.Fatalf("race-week taper prompt missing get_fitness_projection:\n%s", text)
			}
		})
	}
}

func TestCoachAthleteOnboardingIncludesAuthorizationAndChecklistGuidance(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, CoachAthleteOnboardingPrompt(), map[string]string{"athlete_id": " I12345 "})
	for _, want := range []string{
		"athlete_id=i12345",
		"confirm the selected athlete's canonical ID/label",
		"authorization and athlete consent",
		"thresholds/zones",
		"wellness/HRV baseline",
		"devices/sources/sync gaps",
		"athlete_id selects a configured athlete; it is not a credential",
		"Do not request or accept intervals.icu API keys",
		"Do not run live account tests",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("coach onboarding prompt missing %q:\n%s", want, text)
		}
	}
}

func TestCoachAthleteOnboardingRejectsInvalidAthleteID(t *testing.T) {
	t.Parallel()

	_, err := CoachAthleteOnboardingPrompt().Handler(context.Background(), Request{Arguments: map[string]string{"athlete_id": "api-key-not-allowed"}})
	if err == nil {
		t.Fatal("Handler() error = nil, want invalid athlete_id")
	}
	if !strings.Contains(err.Error(), "invalid athlete_id") {
		t.Fatalf("Handler() error = %q, want invalid athlete_id", err.Error())
	}
}

func TestCoachRosterTriageNormalizesAthleteID(t *testing.T) {
	t.Parallel()

	text := renderPromptText(t, CoachRosterTriagePrompt(), map[string]string{"athlete_id": "i12345"})
	if !strings.Contains(text, "athlete_id=i12345") {
		t.Fatalf("coach prompt text = %q, want normalized athlete ID", text)
	}
}

func TestCoachRosterTriageRejectsInvalidAthleteID(t *testing.T) {
	t.Parallel()

	_, err := CoachRosterTriagePrompt().Handler(context.Background(), Request{Arguments: map[string]string{"athlete_id": "api-key-not-allowed"}})
	if err == nil {
		t.Fatal("Handler() error = nil, want invalid athlete_id")
	}
	if !strings.Contains(err.Error(), "invalid athlete_id") {
		t.Fatalf("Handler() error = %q, want invalid athlete_id", err.Error())
	}
}

func TestRaceWeekTaperRequiresRaceDate(t *testing.T) {
	t.Parallel()

	_, err := RaceWeekTaperPrompt().Handler(context.Background(), Request{Arguments: map[string]string{}})
	if err == nil {
		t.Fatal("Handler() error = nil, want missing race_date")
	}
	if !strings.Contains(err.Error(), "missing race_date") {
		t.Fatalf("Handler() error = %q, want missing race_date", err.Error())
	}
}

func TestPromptResourceCitationsStayTerse(t *testing.T) {
	t.Parallel()

	for _, prompt := range []Prompt{TrainingAnalysisPrompt(), RecoveryCheckPrompt(), WeeklyPlanningPrompt(), WeeklyReviewPrompt(), ShareableTrainingReportPrompt(), PlanHealthReviewPrompt(), RaceWeekTaperPrompt(), CoachRosterTriagePrompt(), CoachAthleteOnboardingPrompt()} {
		text := renderPromptText(t, prompt, requiredArgsForPrompt(prompt.Name))
		if !strings.Contains(text, "icuvisor://") {
			t.Fatalf("prompt %s missing resource URI:\n%s", prompt.Name, text)
		}
		if strings.Contains(strings.ToLower(text), "workout dsl grammar") || strings.Count(text, "\n") > 25 {
			t.Fatalf("prompt %s appears too verbose or schema-like:\n%s", prompt.Name, text)
		}
	}
}

func renderPromptText(t *testing.T, prompt Prompt, arguments map[string]string) string {
	t.Helper()
	result, err := prompt.Handler(context.Background(), Request{Name: prompt.Name, Arguments: arguments})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("messages = %#v, want one", result.Messages)
	}
	if result.Messages[0].Role != RoleUser {
		t.Fatalf("message role = %q, want user", result.Messages[0].Role)
	}
	return result.Messages[0].Text
}

func requiredArgsForPrompt(name string) map[string]string {
	if name == CoachRosterTriageName || name == CoachAthleteOnboardingName {
		return map[string]string{"athlete_id": "i12345"}
	}
	if name == RaceWeekTaperName {
		return map[string]string{"race_date": "2026-06-07"}
	}
	return nil
}
