package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
)

const (
	TrainingAnalysisName        = "training_analysis"
	RecoveryCheckName           = "recovery_check"
	WeeklyPlanningName          = "weekly_planning"
	WeeklyReviewName            = "weekly_review"
	ShareableTrainingReportName = "shareable_training_report"
	PlanHealthReviewName        = "plan_health_review"
	RaceWeekTaperName           = "race_week_taper"
	CoachRosterTriageName       = "coach_roster_triage"
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
				"If the user explicitly mentions hypoxic training, altitude tents/chambers, or reduced oxygen exposure, state that CTL/ATL/Form use logged training_load: power-based load may under-represent extra hypoxic strain, HR/RPE/feel/recovery can be supporting context, and you must not apply a hypoxia multiplier without evidence.",
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
			Tools:        []string{"get_athlete_profile", "get_wellness_data", "get_fitness", "get_today"},
			Do: []string{
				"Read wellness first; preserve sleepQuality 1-4 and sleepScore 0-100 as separate fields.",
				"Check HRV, resting HR, readiness, fatigue, soreness, mood, and any `_meta.stale`, `_meta.missing_fields`, or provenance warnings.",
				"When readiness is present, cite `_meta.provenance.readiness.source` and `native_scale`; treat Garmin Body Battery, Oura readiness, Polar nightly recharge/ANS charge, WHOOP recovery, and unknown upstream readiness as provider-native signals, not a universal recovery score.",
				"If readiness is missing or null, say that plainly before interpreting other signals; do not invent a readiness score.",
				"Use HRV, resting HR, sleepSecs, sleepQuality (1-4), sleepScore (0-100), fatigue, soreness, stress, feel, mood, motivation, and available `_native` provider fields only as cautious supporting evidence.",
				"Use fitness only to contextualize recent load; do not turn recovery into a full training analysis.",
				"For today-specific or indoor/outdoor questions, call get_today and use only its weather.status/provenance, planned_events[].indoor, tags, and completed-activity context; if weather.status is forecast_unavailable, say weather is unavailable from icuvisor and do not invent conditions.",
				"When suggesting an indoor alternative, present it as a chat recommendation or preview first; do not write calendar changes, and do not create a second active workout for the same planned session unless the user explicitly approves replacing or adding one.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
				"Do not call write or delete tools for indoor/outdoor adaptation unless the user has reviewed and approved the exact change.",
			},
			Return: "green/yellow/red recovery guidance, the main evidence with provider/source labels, stale or missing fields, readiness-score absence when applicable, weather availability when relevant, and a 24-48h training adjustment",
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
			Tools:        []string{"get_athlete_profile", "get_planning_context", "get_events", "get_training_plan", "get_fitness", "get_training_summary", "get_activities", "compute_compliance_rate", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile/timezone, then ask or confirm the planning anchor: race date, priority/category, goal, and constraints when missing.",
				"Use get_planning_context when available to gather week events, active training-plan context, upcoming races, fitness context, and SEASON_START season boundaries before suggesting changes.",
				"Read planned events and active training-plan context before suggesting changes.",
				"Use fitness, training summary, recent activities, and compute_compliance_rate workout_status/status counts/caveats to summarize current load, fatigue/freshness, and planned-versus-completed work without inferring completion from calendar/activity co-occurrence.",
				"If get_training_plan or compute_compliance_rate is unavailable, call icuvisor_list_advanced_capabilities and proceed from get_events, get_fitness, get_training_summary, and activities.",
				"Use event categories and workout syntax resources by URI if the user asks for edits or workout details.",
				"Draft a season/block/week proposal with assumptions, load constraints, and follow-up questions before any edits.",
				"When proposing endurance workouts, prefer the structured `workout_doc` form on write tools and include any coaching notes via `description` on the same event; both fields coexist, but `description` replaces the upstream description/DSL on writes, so for updates include the desired `workout_doc` whenever preserving structured steps matters. Call `validate_workout` before the write if uncertain about the DSL syntax, and read `icuvisor://workout-syntax` for the cheat sheet and common mistakes.",
				"When the user asks for gym or strength work, schedule a simple `NOTE` time block or free-text supported calendar event; do not invent structured exercises, sets, reps, loads, or rest periods unless documented upstream strength-training support is available.",
				"Before workout create/update/schedule writes, show a before/after preview with total duration, key steps, target intensities, load/distance/time changes, and what existing title/prose/tags/structured steps are preserved.",
				"Before bulk calendar/workout writes, validate or preview one representative structured payload, perform one representative write, read it back, and inspect validation warnings, existing write `_meta` warning fields such as `workout_doc_warning` when present, and `workout_doc_summary`/stored description before writing the rest. Avoid parallel bulk writes while schema wording, warning metadata, or description/`workout_doc` preservation semantics are ambiguous.",
				"For approved writes, use event categories and workout syntax resources, validate workout_doc when uncertain, and write only the exact user-approved changes.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
				"Do not automatically fill the calendar, create ATP notes, or call write/delete tools; first return a reviewed proposal and wait for approval of exact changes.",
			},
			Return: "season/block/week proposal, race context, current load, active-plan and event evidence, compliance notes, and questions before writes",
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
				"Define the athlete-local review window before reading data; do not include wellness, activities, or summary rows after that end date unless the user requested next-week or current-day context.",
				"Use fitness, training summary, and compute_zone_time to summarize load, volume, intensity mix, and fatigue/freshness changes.",
				"Use compute_load_balance and compute_compliance_rate when available; read compute_compliance_rate workout_status, status counts, and caveats before describing planned-vs-completed gaps; otherwise call icuvisor_list_advanced_capabilities, continue from available reads, and name the missing helper.",
				"Review activities, race/other events, and training plan for planned-versus-completed work; include race date/priority when relevant and the upcoming-week preview only when include_next_week is true or the user asks.",
				"Use wellness data for sleep/readiness/HRV context; check `_meta.stale`, `_meta.missing_fields`, provenance warnings, and treat current-day `_meta.as_of` as partial-day context only.",
				"When readiness is present, cite `_meta.provenance.readiness.source` and `native_scale`; treat Garmin Body Battery, Oura readiness, Polar nightly recharge/ANS charge, WHOOP recovery, and unknown upstream readiness as provider-native signals, not a universal recovery score.",
				"If readiness is missing, null, stale, or absent, say that explicitly and do not invent, infer, or backfill a readiness score; use HRV, resting HR, sleep duration/quality/score, subjective fatigue/soreness/stress/feel/mood/motivation, and available `_native` provider fields as cautious supporting context only.",
				"Use analyze_trend only for specific trend questions; keep raw activity rows terse unless evidence is missing.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
				"Do not call write or delete tools unless the user explicitly approves the exact change first.",
				"Do not auto-fill calendars or create ATP notes from the review; propose exact changes for user approval first.",
			},
			Return: "weekly review with wins, concerns, planned-vs-completed gaps, wellness caveats with provider/source labels, load/intensity evidence, next-week preview when requested, and explicit follow-up questions before any write",
		}),
	}
}

// ShareableTrainingReportPrompt guides a user-reviewed shareable report draft.
func ShareableTrainingReportPrompt() Prompt {
	return Prompt{
		Name:        ShareableTrainingReportName,
		Title:       "Shareable training report",
		Description: "Guide a privacy-safe Markdown report draft the athlete can review and share manually.",
		Arguments: []Argument{
			{Name: "report_type", Title: "Report type", Description: "Optional report style such as weekly, monthly, race_prep, or training_journey."},
			{Name: "start_date", Title: "Start date", Description: "Optional athlete-local date string (YYYY-MM-DD) for the report window."},
			{Name: "end_date", Title: "End date", Description: "Optional athlete-local date string (YYYY-MM-DD) for the report window."},
			{Name: "race_date", Title: "Race date", Description: "Optional athlete-local race date string (YYYY-MM-DD) for race-prep context."},
			{Name: "audience", Title: "Audience", Description: "Optional intended audience, for example coach, teammates, family, newsletter, or self."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Shareable training report",
			DefaultScope: "draft a weekly, monthly, race-prep, or training-journey report from the user's requested window; default to the last completed athlete-local week if absent",
			ArgOrder:     []string{"report_type", "start_date", "end_date", "race_date", "audience"},
			Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories"},
			Tools:        []string{"get_athlete_profile", "get_fitness", "get_training_summary", "get_activities", "get_events", "get_training_plan", "get_wellness_data", "compute_zone_time", "compute_load_balance", "analyze_trend", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile first for athlete-local timezone, units, sport settings, and dates; define the report window before fetching data.",
				"Gather only the summary evidence needed for a public-facing story: fitness/form, volume/load, notable sessions, planned/race context, intensity mix, and wellness caveats when useful.",
				"Use analyzers such as compute_zone_time, compute_load_balance, or analyze_trend only when they support the requested story; if unavailable, call icuvisor_list_advanced_capabilities and continue from ordinary reads.",
				"Draft Markdown first with a short title, timeframe, highlights, one honest challenge, key numbers with tool citations, and a concise next-focus section.",
				"If the user asks for HTML, convert the reviewed Markdown to simple static HTML in chat; icuvisor does not generate, publish, upload, or host HTML.",
				"Ask the athlete to review and redact private health, location, notes, identifiers, and race logistics before copying, exporting, or posting anywhere.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Prefer terse default tool responses; do not use include_full, raw streams, or heavy payloads unless the user explicitly asks or evidence is missing.",
				"Do not publish, host, upload, auto-share, or connect to social platforms; the athlete manually shares only after review.",
				"Do not invent missing metrics, race details, locations, health claims, or emotional framing not supported by data or the user's words.",
			},
			Return: "Markdown report draft plus private-data review checklist, cited evidence, missing/stale-data caveats, and optional HTML-conversion offer only after user review",
		}),
	}
}

// PlanHealthReviewPrompt guides transparent review of plan health and risk.
func PlanHealthReviewPrompt() Prompt {
	return Prompt{
		Name:        PlanHealthReviewName,
		Title:       "Plan health review",
		Description: "Guide a transparent plan-health audit using adherence, load/form projection, wellness caveats, and race-date risk from deterministic tools.",
		Arguments: []Argument{
			{Name: "planned_start", Title: "Planned start", Description: "Optional athlete-local date string (YYYY-MM-DD) for the planned-window start; default today."},
			{Name: "planned_end", Title: "Planned end", Description: "Optional athlete-local date string (YYYY-MM-DD) for the planned-window end; default 14 days after planned_start."},
			{Name: "completed_lookback_days", Title: "Completed lookback days", Description: "Optional positive integer string for completed-work adherence context; default 14."},
			{Name: "race_date", Title: "Race date", Description: "Optional athlete-local race date string (YYYY-MM-DD) for race-risk anchoring."},
			{Name: "race_name", Title: "Race name", Description: "Optional race name string to disambiguate matching calendar events."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Plan health review",
			DefaultScope: "review the next 14 athlete-local days, with a 14-day completed-work lookback, unless dates or lookback are supplied",
			ArgOrder:     []string{"planned_start", "planned_end", "completed_lookback_days", "race_date", "race_name"},
			Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories", "icuvisor://analysis-formulas"},
			Tools:        []string{"get_athlete_profile", "get_events", "get_training_plan", "get_activities", "compute_compliance_rate", "get_fitness", "get_training_summary", "compute_load_balance", "get_fitness_projection", "get_wellness_data", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile first for timezone, units, sport settings, and today's athlete-local date; convert all windows before comparing days.",
				"Separate completed-lookback, planned-window, and race-scenario dates; do not mix current-day or post-window wellness into completed adherence evidence.",
				"Read events and training plan for planned workouts and races; if no race event is found, say so and treat any supplied race_date as a scenario anchor only.",
				"Use compute_compliance_rate for scheduled-vs-completed adherence; interpret workout_status, missed/planned/future/completed status counts, and caveats before calling anything skipped, missed, or completed, then get_fitness, get_training_summary, compute_load_balance, and get_fitness_projection for load/form trajectory and future assumptions.",
				"Quote analyzer `_meta.method`, `_meta.assumptions`, `_meta.formula_ref`, missing-days, and sample-size caveats where present; call icuvisor_list_advanced_capabilities and name missing helpers when full-tool analyzers are unavailable.",
				"Read recent wellness for sleep/readiness/HRV caveats; treat current-day `_meta.as_of` as partial-day context only and do not infer readiness when data is stale, absent, or missing key fields.",
				"Treat planned deload or recovery weeks as intentional load reductions unless compliance, wellness, or form evidence shows a problem.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
				"Do not invent a black-box plan-health score; use only surfaced values and label risk low/medium/high with evidence.",
				"Do not create a season plan, fill a calendar, or act as an autonomous physiology model.",
				"Do not call write or delete tools unless the user has reviewed and approved the exact proposal first.",
			},
			Return: "data coverage, adherence, load/form trajectory, transparent risk table, deload/recovery caveats, race-date risk when anchored, and reviewed proposal/questions before any write",
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
		Tools:        []string{"get_athlete_profile", "get_events", "get_training_plan", "get_fitness", "get_training_summary", "get_activities", "compute_compliance_rate", "icuvisor_list_advanced_capabilities"},
		Do: []string{
			"Find the race event by date/name and confirm priority/category, sport, distance, expected duration, and goal when missing.",
			"Review active plan, planned events, fitness, training summary, recent activities, and compute_compliance_rate workout_status/status counts/caveats without pulling raw streams or inferring completion from calendar/activity co-occurrence.",
			"If advanced helpers are unavailable, call icuvisor_list_advanced_capabilities and proceed from events, fitness, summary, and activities.",
			"Frame taper guidance as risk management: freshness, sharpness, logistics, and no last-minute fitness chasing.",
		},
		Guardrails: []string{
			"Do not request or accept intervals.icu API keys in chat.",
			"Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.",
			"Do not automatically fill the calendar, create ATP notes, or call write/delete tools; first return a reviewed taper proposal and wait for approval of exact changes.",
		},
		Return: "race-week schedule proposal, taper risks, intensity guardrails, recovery priorities, missing race context, and questions before writes",
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
