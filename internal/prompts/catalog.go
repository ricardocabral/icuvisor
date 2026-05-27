package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
)

const (
	TrainingAnalysisName  = "training_analysis"
	RecoveryCheckName     = "recovery_check"
	WeeklyPlanningName    = "weekly_planning"
	WeeklyReviewName      = "weekly_review"
	RaceWeekTaperName     = "race_week_taper"
	CoachRosterTriageName = "coach_roster_triage"
)

// TrainingAnalysisPrompt guides training-load and trend analysis.
func TrainingAnalysisPrompt() Prompt {
	return Prompt{
		Name:        TrainingAnalysisName,
		Title:       "Training analysis",
		Description: "Guide a terse training-load, trend, and best-effort readout from existing icuvisor read tools.",
		Arguments: []Argument{
			{Name: "start_date", Title: "Start date", Description: "Optional athlete-local date string (YYYY-MM-DD) for the analysis window."},
			{Name: "end_date", Title: "End date", Description: "Optional athlete-local date string (YYYY-MM-DD) for the analysis window."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Training analysis",
			DefaultScope: "use the user's requested window; default to the last 42 days if absent",
			ArgOrder:     []string{"start_date", "end_date"},
			Resources:    []string{"icuvisor://athlete-profile"},
			Tools:        []string{"get_athlete_profile", "get_fitness", "get_training_summary", "get_best_efforts", "get_activities"},
			Do: []string{
				"Read profile first for timezone, sport settings, and units.",
				"Use fitness and summary rows for CTL/ATL/TSB, ramp, volume, load, and intensity mix.",
				"Use best efforts and recent activities only for context; keep raw rows terse unless the user asks for detail.",
			},
			Return: "load/trend readout with notable changes, likely drivers, missing-data caveats, and 2-3 next-step questions or actions",
		}),
	}
}

// RecoveryCheckPrompt guides wellness-led readiness analysis.
func RecoveryCheckPrompt() Prompt {
	return Prompt{
		Name:        RecoveryCheckName,
		Title:       "Recovery check",
		Description: "Guide a wellness-led recovery and readiness check with correct sleep scales and staleness handling.",
		Arguments: []Argument{
			{Name: "date", Title: "Date", Description: "Optional athlete-local date string (YYYY-MM-DD) to anchor the check; default today."},
			{Name: "lookback_days", Title: "Lookback days", Description: "Optional positive integer string for wellness history; default 14."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Recovery check",
			DefaultScope: "anchor on today with a 14-day lookback unless the user supplied date/lookback_days",
			ArgOrder:     []string{"date", "lookback_days"},
			Resources:    []string{"icuvisor://athlete-profile"},
			Tools:        []string{"get_athlete_profile", "get_wellness_data", "get_fitness"},
			Do: []string{
				"Read wellness first; preserve sleepQuality 1-4 and sleepScore 0-100 as separate fields.",
				"Check HRV, resting HR, readiness, fatigue, soreness, mood, and any `_meta.stale` or provenance warnings.",
				"Use fitness only to contextualize recent load; do not turn recovery into a full training analysis.",
			},
			Return: "green/yellow/red readiness, the main evidence, stale or missing fields, and a 24-48h training adjustment",
		}),
	}
}

// WeeklyPlanningPrompt guides planned-versus-completed week planning.
func WeeklyPlanningPrompt() Prompt {
	return Prompt{
		Name:        WeeklyPlanningName,
		Title:       "Weekly planning",
		Description: "Guide week planning from calendar events, training plans, and completed activity context.",
		Arguments: []Argument{
			{Name: "week_start", Title: "Week start", Description: "Optional athlete-local Monday date string (YYYY-MM-DD) for the planning week."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Weekly planning",
			DefaultScope: "use the upcoming athlete-local week unless week_start is supplied",
			ArgOrder:     []string{"week_start"},
			Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories", "icuvisor://workout-syntax"},
			Tools:        []string{"get_athlete_profile", "get_events", "get_training_plan", "get_activities", "get_training_summary", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile, planned events, and training-plan context before suggesting changes.",
				"If get_training_plan is unavailable in the active toolset, use icuvisor_list_advanced_capabilities and proceed from events/activities.",
				"Compare planned versus completed work where the week has already started.",
				"Use event categories and workout syntax resources by URI if the user asks for edits or workout details.",
				"When proposing workouts, prefer the structured `workout_doc` form on write tools and include any coaching notes via `description` on the same event; both fields coexist, but `description` replaces the upstream description/DSL on writes, so for updates include the desired `workout_doc` whenever preserving structured steps matters. Call `validate_workout` before the write if uncertain about the DSL syntax, and read `icuvisor://workout-syntax` for the cheat sheet and common mistakes.",
			},
			Return: "day-by-day plan, key load constraints, planned-vs-completed notes, and questions before any write tool is used",
		}),
	}
}

// WeeklyReviewPrompt guides structured weekly retrospective and next-week preview.
func WeeklyReviewPrompt() Prompt {
	return Prompt{
		Name:        WeeklyReviewName,
		Title:       "Weekly review",
		Description: "Guide a structured review of the previous training week and optional preview of the upcoming week using existing icuvisor tools.",
		Arguments: []Argument{
			{Name: "week_start", Title: "Week start", Description: "Optional athlete-local Monday date string (YYYY-MM-DD) for the week being reviewed."},
			{Name: "lookback_days", Title: "Lookback days", Description: "Optional positive integer string for context before week_start; default 7."},
			{Name: "include_next_week", Title: "Include next week", Description: "Optional boolean string; when true, include an upcoming-week preview after the review."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Weekly review",
			DefaultScope: "review the previous athlete-local week with a 7-day lookback and include next week only if requested",
			ArgOrder:     []string{"week_start", "lookback_days", "include_next_week"},
			Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories"},
			Tools:        []string{"get_athlete_profile", "get_wellness_data", "get_fitness", "get_training_summary", "get_activities", "get_events", "get_training_plan", "compute_zone_time", "compute_load_balance", "compute_compliance_rate", "analyze_trend", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile first to establish athlete-local timezone, current date, sport settings, and units; compare days only after converting to athlete-local dates.",
				"Use fitness, training summary, and compute_zone_time to summarize load, volume, intensity mix, and fatigue/freshness changes.",
				"Use compute_load_balance and compute_compliance_rate when available; otherwise call icuvisor_list_advanced_capabilities, continue from available reads, and name the missing helper.",
				"Review activities, events, and training plan for planned-versus-completed work; include the upcoming-week preview only when include_next_week is true or the user asks.",
				"Use wellness data for sleep/readiness/HRV context; check `_meta.stale`, missing fields, and provenance warnings, and do not infer readiness when wellness is stale or absent.",
				"Use analyze_trend only for specific trend questions; keep raw activity rows terse unless evidence is missing.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
				"Do not call write or delete tools unless the user explicitly approves the exact change first.",
			},
			Return: "weekly review with wins, concerns, planned-vs-completed gaps, wellness caveats, load/intensity evidence, next-week preview when requested, and explicit follow-up questions before any write",
		}),
	}
}

// RaceWeekTaperPrompt guides race-week taper framing.
func RaceWeekTaperPrompt() Prompt {
	return Prompt{
		Name:        RaceWeekTaperName,
		Title:       "Race-week taper",
		Description: "Guide race-week taper analysis using calendar race context and recent fitness/load reads.",
		Arguments: []Argument{
			{Name: "race_date", Title: "Race date", Description: "Required athlete-local race date string (YYYY-MM-DD).", Required: true},
			{Name: "race_name", Title: "Race name", Description: "Optional race name string to disambiguate events on the same date."},
		},
		Handler: raceWeekTaperHandler,
	}
}

// CoachRosterTriagePrompt guides coach-mode athlete triage.
func CoachRosterTriagePrompt() Prompt {
	return Prompt{
		Name:        CoachRosterTriageName,
		Title:       "Coach roster triage",
		Description: "Guide a coach-mode per-athlete scan; athlete_id is a selector, not a credential.",
		Arguments: []Argument{
			{Name: "athlete_id", Title: "Athlete ID", Description: "Required intervals.icu athlete selector string; IDs are digits, optionally with a leading 'i' (e.g. i12345 or 12345). This is not an API key or credential.", Required: true},
			{Name: "start_date", Title: "Start date", Description: "Optional athlete-local date string (YYYY-MM-DD) for the triage window."},
			{Name: "end_date", Title: "End date", Description: "Optional athlete-local date string (YYYY-MM-DD) for the triage window."},
		},
		Handler: coachRosterTriageHandler,
	}
}

type promptSpec struct {
	Title        string
	DefaultScope string
	ArgOrder     []string
	Resources    []string
	Tools        []string
	Do           []string
	Return       string
	Guardrails   []string
}

func staticPromptHandler(spec promptSpec) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		return renderSpec(spec, req.Arguments), nil
	}
}

func raceWeekTaperHandler(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if strings.TrimSpace(req.Arguments["race_date"]) == "" {
		return Result{}, NewUserError("missing race_date; provide YYYY-MM-DD", nil)
	}
	return renderSpec(promptSpec{
		Title:        "Race-week taper",
		DefaultScope: "race_date is required; use race_name only to disambiguate matching calendar events",
		ArgOrder:     []string{"race_date", "race_name"},
		Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories", "icuvisor://workout-syntax"},
		Tools:        []string{"get_athlete_profile", "get_events", "get_fitness", "get_training_summary", "get_activities"},
		Do: []string{
			"Find the race event or use the supplied race_date as the anchor.",
			"Review recent CTL/ATL/TSB, volume, intensity, and race-specific workouts without pulling raw streams.",
			"Frame taper guidance as risk management: freshness, sharpness, logistics, and no last-minute fitness chasing.",
		},
		Return: "race-week schedule review, taper risks, intensity guardrails, recovery priorities, and open assumptions",
	}, req.Arguments), nil
}

func coachRosterTriageHandler(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	athleteID := strings.TrimSpace(req.Arguments["athlete_id"])
	if athleteID == "" {
		return Result{}, NewUserError("missing athlete_id; intervals.icu IDs are digits, optionally with a leading 'i', e.g. i12345 or 12345", nil)
	}
	normalized, err := config.NormalizeAthleteID(athleteID)
	if err != nil {
		return Result{}, NewUserError("invalid athlete_id; intervals.icu IDs are digits, optionally with a leading 'i', e.g. i12345 or 12345", err)
	}
	args := cloneArgs(req.Arguments)
	args["athlete_id"] = normalized
	return renderSpec(promptSpec{
		Title:        "Coach roster triage",
		DefaultScope: "use the selected athlete and the user's requested window; default to the last 14 days if absent",
		ArgOrder:     []string{"athlete_id", "start_date", "end_date"},
		Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories"},
		Tools:        []string{"get_athlete_profile", "get_wellness_data", "get_fitness", "get_training_summary", "get_events", "get_activities"},
		Do: []string{
			"Treat athlete_id as a coach-mode selector for server-side calls, never as a credential; do not ask for API keys.",
			"Scan wellness, fitness/load, upcoming events, missed/completed activities, and stale data warnings.",
			"Prioritize interventions: urgent health/recovery flags, compliance drift, race/event risk, then routine follow-up.",
		},
		Return: "triage status, top risks, evidence by tool, recommended coach action, and what to check next",
	}, args), nil
}

func renderSpec(spec promptSpec, args map[string]string) Result {
	guardrails := spec.Guardrails
	if len(guardrails) == 0 {
		guardrails = []string{"Do not request or accept intervals.icu API keys in chat.", "Prefer terse default tool responses; use include_full only when the user asks or evidence is missing."}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Prompt: %s\n", spec.Title)
	fmt.Fprintf(&b, "Scope: %s.\n", scopeText(spec.DefaultScope, spec.ArgOrder, args))
	fmt.Fprintf(&b, "Resources: %s.\n", strings.Join(spec.Resources, ", "))
	fmt.Fprintf(&b, "Tools: %s.\n", strings.Join(spec.Tools, ", "))
	b.WriteString("Do:\n")
	for _, item := range spec.Do {
		fmt.Fprintf(&b, "- %s\n", item)
	}
	b.WriteString("Guardrails:\n")
	for _, item := range guardrails {
		fmt.Fprintf(&b, "- %s\n", item)
	}
	fmt.Fprintf(&b, "Return: %s.\n", spec.Return)
	return Result{
		Description: spec.Title,
		Messages: []Message{{
			Role: RoleUser,
			Text: strings.TrimRight(b.String(), "\n"),
		}},
	}
}

func scopeText(defaultScope string, order []string, args map[string]string) string {
	parts := make([]string, 0, len(order))
	for _, name := range order {
		if value := strings.TrimSpace(args[name]); value != "" {
			parts = append(parts, name+"="+value)
		}
	}
	if len(parts) == 0 {
		return defaultScope
	}
	return strings.Join(parts, ", ")
}

func cloneArgs(args map[string]string) map[string]string {
	cloned := make(map[string]string, len(args))
	for key, value := range args {
		cloned[key] = value
	}
	return cloned
}
