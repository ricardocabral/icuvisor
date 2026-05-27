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

func TestNewRegistryRegistersSixPrompts(t *testing.T) {
	t.Parallel()

	registrar := &captureRegistrar{}
	if err := NewRegistry().Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	wantNames := []string{TrainingAnalysisName, RecoveryCheckName, WeeklyPlanningName, WeeklyReviewName, RaceWeekTaperName, CoachRosterTriageName}
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
		{name: "race_week_taper", prompt: RaceWeekTaperPrompt(), arguments: map[string]string{"race_date": "2026-06-07", "race_name": "A Race"}, goldenFile: "race_week_taper.md"},
		{name: "coach_roster_triage", prompt: CoachRosterTriagePrompt(), arguments: map[string]string{"athlete_id": "i12345", "start_date": "2026-05-01", "end_date": "2026-05-14"}, goldenFile: "coach_roster_triage.md"},
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
		"icuvisor_list_advanced_capabilities",
		"_meta.stale",
		"provenance warnings",
		"Do not call write or delete tools unless the user explicitly approves the exact change first.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("weekly review prompt missing %q:\n%s", want, text)
		}
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

	for _, prompt := range []Prompt{TrainingAnalysisPrompt(), RecoveryCheckPrompt(), WeeklyPlanningPrompt(), WeeklyReviewPrompt(), RaceWeekTaperPrompt(), CoachRosterTriagePrompt()} {
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
	if name == CoachRosterTriageName {
		return map[string]string{"athlete_id": "i12345"}
	}
	if name == RaceWeekTaperName {
		return map[string]string{"race_date": "2026-06-07"}
	}
	return nil
}
