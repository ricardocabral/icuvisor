package prompts

import "context"

type staticRegistry struct {
	entries []Prompt
}

// NewRegistry returns the default MCP prompt registry.
func NewRegistry() Registry {
	return staticRegistry{entries: []Prompt{
		TrainingAnalysisPrompt(),
		RideAnalysisPrompt(),
		FuelingReviewPrompt(),
		RecoveryCheckPrompt(),
		WeeklyPlanningPrompt(),
		WeeklyReviewPrompt(),
		CoachingHandoffPrompt(),
		ShareableTrainingReportPrompt(),
		PlanHealthReviewPrompt(),
		RaceWeekTaperPrompt(),
		CoachRosterTriagePrompt(),
		CoachAthleteOnboardingPrompt(),
	}}
}

func (r staticRegistry) Register(ctx context.Context, registrar Registrar) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, prompt := range r.entries {
		if err := registrar.AddPrompt(prompt); err != nil {
			return err
		}
	}
	return nil
}
