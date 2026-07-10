package prompts

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/config"
)

const (
	TrainingAnalysisName        = "training_analysis"
	RideAnalysisName            = "ride_analysis"
	FuelingReviewName           = "fueling_review"
	RecoveryCheckName           = "recovery_check"
	WeeklyPlanningName          = "weekly_planning"
	WeeklyReviewName            = "weekly_review"
	CoachingHandoffName         = "coaching_handoff"
	ShareableTrainingReportName = "shareable_training_report"
	PlanHealthReviewName        = "plan_health_review"
	RaceWeekTaperName           = "race_week_taper"
	CoachRosterTriageName       = "coach_roster_triage"
	CoachAthleteOnboardingName  = "coach_athlete_onboarding"
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

// RideAnalysisPrompt guides one-activity ride analysis with deterministic analyzers.
func RideAnalysisPrompt() Prompt {
	return Prompt{
		Name:        RideAnalysisName,
		Title:       "Ride analysis",
		Description: "Guide a unit-safe analysis of one ride using activity lookup, details, intervals, streams, and analyzer tools instead of chat-side reductions.",
		Arguments: []Argument{
			{Name: "activity_id", Title: "Activity ID", Description: "Optional intervals.icu activity ID when known; otherwise resolve the ride from date/name context first."},
			{Name: "activity_date", Title: "Activity date", Description: "Optional athlete-local date string (YYYY-MM-DD) used to find the ride when activity_id is not supplied."},
			{Name: "focus", Title: "Focus", Description: "Optional analysis focus such as pacing, intervals, power, heart rate, fueling, or durability."},
		},
		Handler: staticPromptHandler(promptSpec{
			Title:        "Ride analysis",
			DefaultScope: "analyze the specified ride; if activity_id is absent, resolve the ride from athlete-local date/name context before fetching details",
			ArgOrder:     []string{"activity_id", "activity_date", "focus"},
			Resources:    []string{"icuvisor://athlete-profile", "icuvisor://analysis-formulas"},
			Tools:        []string{"get_athlete_profile", "get_activities", "get_activity_details", "get_activity_intervals", "get_activity_streams", "get_activity_histogram", "compute_activity_segment_stats", "compute_zone_time", "analyze_distribution", "analyze_efforts_delta", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile first for athlete-local timezone, sport settings, thresholds/zones, and preferred units before comparing or labeling metrics.",
				"If activity_id is missing, use get_activities with athlete-local date/name context to identify the ride; do not guess from client-local dates or partial titles.",
				"Fetch get_activity_details before deeper analysis so tags, gear, calories_burned, carbs_ingested_g, carbs_used_g, unit-labelled metrics, and unavailable Strava-import fields are explicit.",
				"Use get_activity_intervals for lap/rep structure and interval_source/interval_source_caveat before judging workout execution; when a single collapsed/imported lap is ambiguous, say so and use compute_activity_segment_stats only for explicit segment questions.",
				"Prefer analyzer tools such as get_activity_histogram, compute_activity_segment_stats, compute_zone_time, analyze_distribution, and analyze_efforts_delta for deterministic math; cite `_meta.method`, `_meta.source_tools`, assumptions, caveats, and units instead of reducing raw streams in chat.",
				"Use get_activity_streams only when a deterministic analyzer cannot answer the user's specific question or when the user explicitly requests samples; keep include_full false unless full samples are required.",
			},
			Guardrails: []string{
				"Do not request or accept intervals.icu API keys in chat.",
				"Do not invent missing power, heart-rate, pace, weather, location, fueling, or baseline data; report unavailable fields and Strava import restrictions plainly.",
				"Do not diagnose medical issues or prescribe treatment from ride data; keep recommendations framed as training observations and questions.",
			},
			Return: "ride analysis with resolved activity identity, key unit-safe metrics, interval/segment evidence, deterministic analyzer findings with `_meta.method` and caveats, and focused next-step questions",
		}),
	}
}

// FuelingReviewPrompt guides a source-labelled review of logged fueling evidence.
func FuelingReviewPrompt() Prompt {
	return Prompt{
		Name:        FuelingReviewName,
		Title:       "Fueling review",
		Description: "Guide a read-only review of logged activity and daily nutrition evidence with transparent grams-per-hour calculations and explicit gaps.",
		Arguments: []Argument{
			{Name: "activity_id", Title: "Activity ID", Description: "Optional intervals.icu activity ID; cannot be combined with start_date or end_date."},
			{Name: "start_date", Title: "Start date", Description: "Optional athlete-local YYYY-MM-DD date; requires end_date and cannot be combined with activity_id."},
			{Name: "end_date", Title: "End date", Description: "Optional athlete-local YYYY-MM-DD date; requires start_date and cannot be combined with activity_id."},
			{Name: "race_date", Title: "Race date", Description: "Optional athlete-local YYYY-MM-DD date for same-day calendar context."},
			{Name: "race_name", Title: "Race name", Description: "Optional race name to disambiguate an event on the supplied race_date."},
		},
		Handler: fuelingReviewHandler,
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
				"Do not infer separate indoor/outdoor FTP from planned_events[].indoor or zone boundaries. Use get_athlete_profile sport_settings[].indoor_ftp_watts only when present; otherwise ask or confirm how to adjust the workout.",
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
			Tools:        []string{"get_athlete_profile", "resolve_calendar_dates", "get_planning_context", "propose_annual_training_plan", "get_events", "get_training_plan", "get_fitness", "get_training_summary", "get_activities", "compute_compliance_rate", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile/timezone, then ask or confirm the planning anchor: race date, priority/category, goal, and constraints when missing; for relative dates, weekdays, countdowns, or stale conversations, call resolve_calendar_dates and use its athlete-local date/weekday instead of UTC, client-time, or model arithmetic.",
				"Use get_planning_context when available to gather week events, active training-plan context, upcoming races, fitness context, and SEASON_START season boundaries before suggesting changes.",
				"Read planned events and active training-plan context before suggesting changes.",
				"Use fitness, training summary, recent activities, and compute_compliance_rate workout_status/status counts/caveats to summarize current load, fatigue/freshness, and planned-versus-completed work without inferring completion from calendar/activity co-occurrence.",
				"If get_training_plan or compute_compliance_rate is unavailable, call icuvisor_list_advanced_capabilities and proceed from get_events, get_fitness, get_training_summary, and activities.",
				"Use event categories and workout syntax resources by URI if the user asks for edits or workout details.",
				"For season or ATP proposals, call propose_annual_training_plan to get deterministic read-only phases, weekly targets, assumptions, warnings, and projection-ready weekly_plan_targets instead of doing ATP math in chat.",
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

// CoachingHandoffPrompt guides a privacy-safe, read-only conversation handoff.
func CoachingHandoffPrompt() Prompt {
	return Prompt{
		Name:        CoachingHandoffName,
		Title:       "Coaching conversation handoff",
		Description: "Guide a compact, source-labelled Markdown handoff that preserves durable coaching context for a fresh client conversation.",
		Arguments: []Argument{
			{Name: "lookback_days", Title: "Lookback days", Description: "Optional positive integer string from 1 to 90 for recent training evidence; default 28."},
			{Name: "race_context_days", Title: "Race context days", Description: "Optional positive integer string from 1 to 365 for upcoming race/event context; default 90."},
		},
		Handler: coachingHandoffHandler,
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
			Tools:        []string{"get_athlete_profile", "resolve_calendar_dates", "get_events", "get_training_plan", "get_activities", "compute_compliance_rate", "get_fitness", "get_training_summary", "compute_load_balance", "get_fitness_projection", "get_wellness_data", "icuvisor_list_advanced_capabilities"},
			Do: []string{
				"Read profile first for timezone, units, sport settings, and today's athlete-local date; call resolve_calendar_dates for relative planned windows, weekdays, countdowns, or stale conversations, then compare only returned athlete-local dates instead of UTC, client-time, or model arithmetic.",
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

// CoachAthleteOnboardingPrompt guides read-only coach/team athlete onboarding.
func CoachAthleteOnboardingPrompt() Prompt {
	return Prompt{
		Name:        CoachAthleteOnboardingName,
		Title:       "Coach athlete onboarding",
		Description: "Guide a coach through authorized athlete/team onboarding using existing read-only coach-mode tools.",
		Arguments: []Argument{
			{Name: "athlete_id", Title: "Athlete ID", Description: "Optional intervals.icu athlete selector string; IDs are digits, optionally with a leading 'i' (e.g. i12345 or 12345). This is not an API key or proof of authorization."},
			{Name: "start_date", Title: "Start date", Description: "Optional athlete-local date string (YYYY-MM-DD) for recent activity and wellness coverage."},
			{Name: "end_date", Title: "End date", Description: "Optional athlete-local date string (YYYY-MM-DD) for recent activity and wellness coverage."},
		},
		Handler: coachAthleteOnboardingHandler,
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

func fuelingReviewHandler(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	args := cloneArgs(req.Arguments)
	for _, key := range []string{"activity_id", "start_date", "end_date", "race_date", "race_name"} {
		args[key] = strings.TrimSpace(args[key])
	}
	activityID := args["activity_id"]
	startDate := args["start_date"]
	endDate := args["end_date"]
	raceDate := args["race_date"]
	raceName := args["race_name"]

	if activityID != "" && (startDate != "" || endDate != "") {
		return Result{}, NewUserError("invalid fueling review arguments; activity_id cannot be combined with start_date or end_date", nil)
	}
	if (startDate == "") != (endDate == "") {
		return Result{}, NewUserError("invalid date range; provide both start_date and end_date as YYYY-MM-DD", nil)
	}
	var start, end time.Time
	if startDate != "" {
		var err error
		start, err = parsePromptDate(startDate)
		if err != nil {
			return Result{}, NewUserError("invalid start_date; provide YYYY-MM-DD", err)
		}
		end, err = parsePromptDate(endDate)
		if err != nil {
			return Result{}, NewUserError("invalid end_date; provide YYYY-MM-DD", err)
		}
		if end.Before(start) {
			return Result{}, NewUserError("invalid date range; end_date must be on or after start_date", nil)
		}
		if int(end.Sub(start).Hours()/24)+1 > 90 {
			return Result{}, NewUserError("invalid date range; provide 90 athlete-local days or fewer", nil)
		}
	}
	if raceDate != "" {
		if _, err := parsePromptDate(raceDate); err != nil {
			return Result{}, NewUserError("invalid race_date; provide YYYY-MM-DD", err)
		}
	}
	if raceName != "" && raceDate == "" {
		return Result{}, NewUserError("missing race_date; provide YYYY-MM-DD", nil)
	}

	return renderSpec(promptSpec{
		Title:        "Fueling review",
		DefaultScope: "resolve athlete-local offsets -14 and -1, then review those 14 completed days unless an activity or date range is supplied",
		ArgOrder:     []string{"activity_id", "start_date", "end_date", "race_date", "race_name"},
		Resources:    []string{"icuvisor://athlete-profile"},
		Tools:        []string{"get_athlete_profile", "resolve_calendar_dates", "get_activities", "get_activity_details", "get_wellness_data", "get_training_summary", "get_events"},
		Do: []string{
			"Read profile first for athlete-local timezone and units. For a default or relative window, call resolve_calendar_dates with offsets -14 and -1; use returned athlete-local dates, not UTC or client-time arithmetic.",
			"For activity_id, read terse get_activity_details once. For a date range, read terse paginated get_activities with include_unnamed:true for duration, load, carbs_ingested_g, and carbs_used_g; do not fetch details for every row.",
			"Fetch every activity page needed before calling a range complete; otherwise state covered count/window as partial, count unavailable or Strava-blocked rows separately, and never reveal next_page_token. Preserve activity timezones and label current-day _meta.as_of data partial.",
			"Read get_wellness_data only when daily nutrition evidence is useful, with fields kcalConsumed, carbohydrates, protein, and fatTotal plus only explicitly requested custom codes. Keep its returned calories_intake, carbs_g, protein_g, and fat_g as daily fields; do not broaden into health fields when nutrition provenance or freshness is unavailable.",
			"Keep carbs_ingested_g as athlete-logged during-activity intake, carbs_used_g as an upstream used/burned estimate, and calories_burned/load as context only. Preserve custom-field codes and unknown meanings; never substitute carbs_used_g, wellness totals, calories, or load for intake.",
			"Use moving_time_seconds only: logged carbs/hour = carbs_ingested_g / (moving_time_seconds / 3600), labelled g/h. Calculate only a non-negative numeric ingested value with positive duration; zero yields 0 g/h, while absent/negative intake, invalid duration, and unavailable rows are counted exclusions. Aggregate only eligible grams and those same durations, then state eligible/total-session coverage.",
			"When race_date is supplied, call get_events with oldest/newest equal to that athlete-local date and limit:100; race_name only disambiguates that result. Mark _meta.truncated race context partial, and call a complete no-match unconfirmed rather than inventing an event.",
			"Return separate Sourced activity evidence, Sourced daily-wellness evidence, Sourced race/calendar context, Labelled calculations, Coverage and data gaps, and General educational guidance sections. Surface _meta.stale, _meta.missing_fields, field semantics/provenance, and availability warnings.",
		},
		Guardrails: []string{
			"This workflow is read-only: do not call write/delete tools, include_full, streams, or raw payloads.",
			"Do not treat missing data as zero or inadequate fueling, extrapolate a rate, invent intake or targets, or create a food/product library.",
			"Do not diagnose health conditions or eating disorders, prescribe individualized nutrition, infer deficits, recommend carbohydrate/calorie/sodium/fluid/sweat-rate targets, or claim product/performance effects.",
			"Keep general material visibly separate as conditional educational guidance; refer individualized or medical nutrition requests to a qualified sports dietitian or clinician.",
		},
		Return: "source-labelled logged-fueling evidence, transparent g/h calculations where valid, covered-session counts and exclusions, calendar context when confirmed, and clearly separated educational guidance",
	}, args), nil
}

func parsePromptDate(value string) (time.Time, error) {
	date, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return time.Time{}, err
	}
	if date.Format(time.DateOnly) != value {
		return time.Time{}, fmt.Errorf("date must use YYYY-MM-DD")
	}
	return date, nil
}

func coachingHandoffHandler(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	lookbackDays, err := boundedPromptDays(req.Arguments["lookback_days"], 28, 90)
	if err != nil {
		return Result{}, NewUserError("invalid lookback_days; provide an integer from 1 to 90", err)
	}
	raceContextDays, err := boundedPromptDays(req.Arguments["race_context_days"], 90, 365)
	if err != nil {
		return Result{}, NewUserError("invalid race_context_days; provide an integer from 1 to 365", err)
	}
	args := cloneArgs(req.Arguments)
	args["lookback_days"] = strconv.Itoa(lookbackDays)
	args["race_context_days"] = strconv.Itoa(raceContextDays)

	return renderSpec(promptSpec{
		Title:        "Coaching conversation handoff",
		DefaultScope: "use the last 28 athlete-local days and the next 90 athlete-local days of race/event context",
		ArgOrder:     []string{"lookback_days", "race_context_days"},
		Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories"},
		Tools:        []string{"get_athlete_profile", "resolve_calendar_dates", "get_events", "get_training_plan", "get_fitness", "get_training_summary", "get_activities", "get_wellness_data", "icuvisor_list_advanced_capabilities"},
		Do: []string{
			"Call get_athlete_profile for the athlete timezone and resolve_calendar_dates for today's athlete-local anchor or any relative date; state the generated-on date, timezone, lookback window, and race-context window.",
			"Return compact Markdown sections in this exact order: Handoff scope; Conversation-stated context (Goals, Constraints, Accepted decisions); Icuvisor evidence; Current plan state; Data gaps and unresolved questions; Next actions.",
			"Put only user-stated goals and constraints in Conversation-stated context. Put a decision in Accepted decisions only when the user explicitly stated or accepted it; never promote an assistant suggestion, model summary, or calendar row into a user decision.",
			"Keep tool-sourced facts separate. Render Icuvisor evidence as Claim | Source tool | Athlete-local evidence date/window | Freshness/as-of, and keep Current plan state limited to sourced event and training-plan state.",
			"Distinguish the date/window when evidence applies from returned as_of or provider freshness. Preserve trustworthy freshness markers; write 'not provided' when none exists, and never invent fetched_at or another retrieval timestamp.",
			"Use terse get_events, get_training_plan, get_fitness, get_training_summary, get_activities, and get_wellness_data only as needed for durable context; do not dump full history.",
			"Surface _meta.stale, _meta.missing_fields, unavailable or Strava-blocked data, current-day partial data, and unresolved tool failures. Never treat a missing value as zero or fill a tool gap from chat memory.",
			"When next_page_token is present, fetch every page needed before claiming completeness or label the evidence partial with the covered window/count; never include the opaque token in the handoff.",
			"If an advanced analyzer material to the conversation is unavailable, call icuvisor_list_advanced_capabilities, name the missing capability, and preserve the unresolved question instead of calculating a substitute in chat.",
			"Ask the athlete to review the Markdown, then manually copy it into a fresh Claude, ChatGPT, Cursor, or other client conversation; do not claim any client automatically imports, persists, or remembers it.",
		},
		Guardrails: []string{
			"This workflow is read-only: do not call write or delete tools.",
			"Never use include_full, raw streams, raw tool payloads, or full histories for the handoff.",
			"Exclude credentials, API/OAuth tokens, secrets, raw athlete identifiers, local or config paths, pagination tokens, and transport or debug metadata.",
			"Omit health details, precise locations, and private free-text notes by default; include only the minimum the user explicitly approves.",
			"Do not make diagnoses, unsupported physiological conclusions, or claims sourced only from model memory.",
		},
		Return: "the six-section Markdown handoff with explicit source separation, athlete-local dates, freshness and coverage caveats, unresolved questions, next actions, and a manual-review reminder",
	}, args), nil
}

func boundedPromptDays(value string, defaultValue, maxValue int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue, nil
	}
	days, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if days < 1 || days > maxValue {
		return 0, fmt.Errorf("value %d outside 1-%d", days, maxValue)
	}
	return days, nil
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
		Tools:        []string{"get_athlete_profile", "resolve_calendar_dates", "get_events", "get_training_plan", "get_fitness", "get_training_summary", "get_activities", "compute_compliance_rate", "get_fitness_projection", "icuvisor_list_advanced_capabilities"},
		Do: []string{
			"Find the race event by date/name and confirm priority/category, sport, distance, expected duration, and goal when missing; if the user supplied a relative race date, countdown, weekday, or weekday/date pairing, first call resolve_calendar_dates and use its athlete-local result instead of UTC, client-time, or model arithmetic.",
			"Review active plan, planned events, fitness, training summary, recent activities, compute_compliance_rate workout_status/status counts/caveats, and get_fitness_projection race-day form assumptions without pulling raw streams or inferring completion from calendar/activity co-occurrence.",
			"If advanced helpers are unavailable, call icuvisor_list_advanced_capabilities and proceed from events, fitness, summary, activities, and explicit projection assumptions.",
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

func coachAthleteOnboardingHandler(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	args := cloneArgs(req.Arguments)
	if athleteID := strings.TrimSpace(args["athlete_id"]); athleteID != "" {
		normalized, err := config.NormalizeAthleteID(athleteID)
		if err != nil {
			return Result{}, NewUserError("invalid athlete_id; intervals.icu IDs are digits, optionally with a leading 'i', e.g. i12345 or 12345", err)
		}
		args["athlete_id"] = normalized
	}
	return renderSpec(promptSpec{
		Title:        "Coach athlete onboarding",
		DefaultScope: "list the coach roster, choose one athlete, and use the last 28 athlete-local days unless a window is supplied",
		ArgOrder:     []string{"athlete_id", "start_date", "end_date"},
		Resources:    []string{"icuvisor://athlete-profile", "icuvisor://event-categories"},
		Tools:        []string{"list_athletes", "select_athlete", "get_athlete_profile", "get_activities", "get_training_summary", "get_fitness", "get_wellness_data", "get_events", "get_training_plan", "icuvisor_list_advanced_capabilities"},
		Do: []string{
			"Start with list_athletes; if athlete_id is supplied, select_athlete for that normalized selector, otherwise ask the coach which roster athlete to onboard.",
			"Before summarizing data, confirm the selected athlete's canonical ID/label and state that the coach must already have authorization and athlete consent to view and analyze this data.",
			"Read profile first for identity, timezone, units, thresholds/zones, and `_meta.warnings`; then check recent activities, training summary, fitness, wellness/HRV, events/races, and training-plan context.",
			"Call icuvisor_list_advanced_capabilities when a checklist item depends on a missing or ACL-hidden tool; name unavailable data rather than guessing.",
			"Produce checklist rows for thresholds/zones, activity coverage, wellness/HRV baseline, races/events/goals, devices/sources/sync gaps, missing data warnings, and coach follow-up questions.",
			"Keep this onboarding read-only; propose any calendar/settings changes separately and wait for explicit reviewed approval before using write tools.",
		},
		Guardrails: []string{
			"athlete_id selects a configured athlete; it is not a credential, consent artifact, invite token, or proof of upstream authorization.",
			"Do not request or accept intervals.icu API keys, OAuth tokens, invite links, or private identifiers in chat.",
			"Do not expose raw wellness/location details beyond what the coach needs for onboarding; ask the coach to review/redact any summary before sharing.",
			"Do not run live account tests or claim upstream roster import, consent capture, device inventory, or bulk team analytics exists.",
		},
		Return: "authorized-athlete confirmation, onboarding checklist with pass/warn/missing status, baseline profile, goals/races questions, device/source caveats, and first coach actions",
	}, args), nil
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
