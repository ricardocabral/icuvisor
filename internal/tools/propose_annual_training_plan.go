package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	proposeAnnualTrainingPlanName        = "propose_annual_training_plan"
	proposeAnnualTrainingPlanDescription = "Use when the prompt asks to propose a season plan, annual training plan, race build, taper, recovery weeks, weekly load targets, or projection-ready ATP bridge; do not roll ATP/projection math in chat and do not write calendar data. Returns deterministic read-only phases, weekly targets, assumptions, warnings, and get_fitness_projection weekly_plan_targets."
	invalidSeasonPlanProposalMessage     = "invalid propose_annual_training_plan arguments; provide goal_date, optional future-Monday start_date, bounded load/hour targets, and a horizon of at most 53 weeks"
	fetchSeasonPlanProposalMessage       = "could not resolve athlete profile/timezone for season-plan proposal"

	seasonPlanProposalMethod        = "deterministic_read_only_season_plan_proposal"
	seasonPlanProposalSchemaVersion = "season_plan_proposal.v1"
	seasonPlanMaxWeeks              = 53
	seasonPlanLoadPerHour           = 50.0
	seasonPlanDefaultWeeklyLoad     = 300.0
	seasonPlanDefaultTargetFactor   = 1.15
	seasonPlanRecoveryCadenceWeeks  = 4
	seasonPlanRecoveryLoadPct       = 60.0
	seasonPlanDefaultTaperWeeks     = 1
	seasonPlanMinTaperWeeks         = 1
	seasonPlanMaxTaperWeeks         = 3
	seasonPlanDefaultTaperLoadPct   = 60.0
	seasonPlanMinTaperLoadPct       = 1.0
	seasonPlanMaxTaperLoadPct       = 100.0
)

var allowedSeasonGoalTypes = map[string]struct{}{"race": {}, "event": {}, "season_goal": {}, "other": {}}

type seasonPlanProposalRequest struct {
	StartDate          string   `json:"start_date,omitempty"`
	GoalDate           string   `json:"goal_date"`
	GoalName           string   `json:"goal_name,omitempty"`
	GoalType           string   `json:"goal_type,omitempty"`
	TargetWeeklyLoad   *float64 `json:"target_weekly_load,omitempty"`
	CurrentWeeklyLoad  *float64 `json:"current_weekly_load,omitempty"`
	TargetHoursPerWeek *float64 `json:"target_hours_per_week,omitempty"`
	CurrentWeeklyHours *float64 `json:"current_weekly_hours,omitempty"`
	MaxHoursPerWeek    *float64 `json:"max_hours_per_week,omitempty"`
	TaperWeeks         *int     `json:"taper_weeks,omitempty"`
	TaperTargetLoadPct *float64 `json:"taper_target_load_pct,omitempty"`
	Sports             []string `json:"sports,omitempty"`
	StrengthContext    string   `json:"strength_context,omitempty"`
	IncludeFull        bool     `json:"include_full,omitempty"`
}

type seasonPlanProposalResponse struct {
	Summary       seasonPlanProposalSummary        `json:"summary"`
	Phases        []seasonPlanProposalPhase        `json:"phases"`
	WeeklyTargets []seasonPlanProposalWeeklyTarget `json:"weekly_targets"`
	RecoveryWeeks []seasonPlanProposalRecoveryWeek `json:"recovery_weeks"`
	RaceAnchors   []seasonPlanProposalRaceAnchor   `json:"race_anchors"`
	Assumptions   []seasonPlanProposalNotice       `json:"assumptions"`
	Warnings      []seasonPlanProposalNotice       `json:"warnings"`
	Meta          seasonPlanProposalMeta           `json:"_meta"`
}

type seasonPlanProposalSummary struct {
	StartDate          string   `json:"start_date"`
	EndDate            string   `json:"end_date"`
	GoalDate           string   `json:"goal_date"`
	TotalWeeks         int      `json:"total_weeks"`
	PhaseCount         int      `json:"phase_count"`
	RecoveryWeekCount  int      `json:"recovery_week_count"`
	RaceAnchorCount    int      `json:"race_anchor_count"`
	TargetWeeklyLoad   float64  `json:"target_weekly_load"`
	TargetHoursPerWeek float64  `json:"target_hours_per_week"`
	MaxHoursPerWeek    *float64 `json:"max_hours_per_week,omitempty"`
}

type seasonPlanProposalPhase struct {
	PhaseID        string `json:"phase_id"`
	PhaseType      string `json:"phase_type"`
	Name           string `json:"name"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	WeekCount      int    `json:"week_count"`
	StartWeekIndex int    `json:"start_week_index"`
	EndWeekIndex   int    `json:"end_week_index"`
}

type seasonPlanProposalWeeklyTarget struct {
	WeekStartDate  string  `json:"week_start_date"`
	WeekEndDate    string  `json:"week_end_date"`
	WeekIndex      int     `json:"week_index"`
	PhaseID        string  `json:"phase_id"`
	PhaseType      string  `json:"phase_type"`
	TrainingLoad   float64 `json:"training_load"`
	TargetHours    float64 `json:"target_hours"`
	IsRecoveryWeek bool    `json:"is_recovery_week"`
	IsTaperWeek    bool    `json:"is_taper_week"`
	LoadSource     string  `json:"load_source"`
	HoursSource    string  `json:"hours_source"`
}

type seasonPlanProposalRecoveryWeek struct {
	WeekStartDate   string  `json:"week_start_date"`
	WeekIndex       int     `json:"week_index"`
	PhaseID         string  `json:"phase_id"`
	TrainingLoad    float64 `json:"training_load"`
	RecoveryLoadPct float64 `json:"recovery_load_pct"`
	Reason          string  `json:"reason"`
}

type seasonPlanProposalRaceAnchor struct {
	Date          string `json:"date"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type"`
	Source        string `json:"source"`
	WeekStartDate string `json:"week_start_date"`
}

type seasonPlanProposalNotice struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Value   any    `json:"value,omitempty"`
}

type seasonPlanProposalTaperScenario struct {
	RequestedTaperWeeks         *int     `json:"requested_taper_weeks"`
	ResolvedTaperWeeks          int      `json:"resolved_taper_weeks"`
	TaperWeeksSource            string   `json:"taper_weeks_source"`
	RequestedTaperTargetLoadPct *float64 `json:"requested_taper_target_load_pct"`
	ResolvedTaperTargetLoadPct  float64  `json:"resolved_taper_target_load_pct"`
	TaperTargetLoadPctSource    string   `json:"taper_target_load_pct_source"`
	Allocation                  string   `json:"allocation"`
	AdjustmentCode              string   `json:"adjustment_code"`
	AssumptionCodes             []string `json:"assumption_codes"`
}

type seasonPlanProposalMeta struct {
	Method                  string                             `json:"method"`
	SourceTools             []string                           `json:"source_tools"`
	ReadOnly                bool                               `json:"read_only"`
	WritesPerformed         bool                               `json:"writes_performed"`
	Timezone                string                             `json:"timezone"`
	SchemaVersion           string                             `json:"schema_version"`
	PlanningSourceEndpoints seasonPlanProposalSourceEndpoints  `json:"planning_source_endpoints"`
	ProjectionBridge        seasonPlanProposalProjectionBridge `json:"projection_bridge"`
	TaperScenario           seasonPlanProposalTaperScenario    `json:"taper_scenario"`
}

type seasonPlanProposalSourceEndpoints struct {
	ProfileTimezone string   `json:"profile_timezone"`
	NotCalled       []string `json:"not_called"`
}

type seasonPlanProposalProjectionBridge struct {
	TargetTool        string                                     `json:"target_tool"`
	TargetArgument    string                                     `json:"target_argument"`
	WeeklyPlanTargets []seasonPlanProposalProjectionWeeklyTarget `json:"weekly_plan_targets"`
	IncludedWeekCount int                                        `json:"included_week_count"`
}

type seasonPlanProposalProjectionWeeklyTarget struct {
	WeekStartDate string  `json:"week_start_date"`
	TrainingLoad  float64 `json:"training_load"`
}

type seasonPlanResolvedInputs struct {
	startDate            time.Time
	goalDate             time.Time
	goalWeekStart        time.Time
	currentWeeklyLoad    float64
	targetWeeklyLoad     float64
	currentWeeklyHours   float64
	targetWeeklyHours    float64
	maxHoursPerWeek      *float64
	loadSource           string
	hoursSource          string
	taperWeeks           int
	taperWeeksSource     string
	taperTargetLoadPct   float64
	taperTargetSource    string
	taperAdjustmentCode  string
	taperAssumptionCodes []string
	assumptions          []seasonPlanProposalNotice
	warnings             []seasonPlanProposalNotice
}

func newProposeAnnualTrainingPlanTool(profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	return newProposeAnnualTrainingPlanToolWithClock(profileClient, version, timezoneFallback, debugMetadata, time.Now, shaping...)
}

func newProposeAnnualTrainingPlanToolWithClock(profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: proposeAnnualTrainingPlanName, Description: proposeAnnualTrainingPlanDescription, InputSchema: seasonPlanProposalInputSchema(), OutputSchema: seasonPlanProposalOutputSchema(), Handler: seasonPlanProposalHandler(profileClient, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func seasonPlanProposalHandler(profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeSeasonPlanProposalRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidSeasonPlanProposalMessage, err)
		}
		if profileClient == nil {
			return Result{}, NewUserError(fetchSeasonPlanProposalMessage, errors.New("missing profile client"))
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchSeasonPlanProposalMessage, err)
		}
		asOf, err := response.AsOfMetadataInTimezone(now(), timezoneName)
		if err != nil {
			return Result{}, NewUserError(fetchSeasonPlanProposalMessage, err)
		}
		asOfDate, err := time.Parse(time.DateOnly, asOf.AsOfDate)
		if err != nil {
			return Result{}, NewUserError(fetchSeasonPlanProposalMessage, err)
		}
		resolved, err := resolveSeasonPlanProposalInputs(args, asOfDate)
		if err != nil {
			return Result{}, NewUserError(invalidSeasonPlanProposalMessage, err)
		}
		payload := buildSeasonPlanProposal(args, resolved, timezoneName)
		return encodeShaped(payload, args.IncludeFull, []string{"phases", "weekly_targets", "recovery_weeks", "race_anchors", "assumptions", "warnings"}, version, debugMetadata, proposeAnnualTrainingPlanName, unitSystem, shapeCfg)
	}
}

func decodeSeasonPlanProposalRequest(raw json.RawMessage) (seasonPlanProposalRequest, error) {
	var args seasonPlanProposalRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[seasonPlanProposalRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.GoalDate = strings.TrimSpace(args.GoalDate)
	args.GoalName = strings.TrimSpace(args.GoalName)
	args.GoalType = strings.TrimSpace(args.GoalType)
	args.StrengthContext = strings.TrimSpace(args.StrengthContext)
	if args.GoalDate == "" || !validDate(args.GoalDate) {
		return args, errors.New("goal_date must be YYYY-MM-DD")
	}
	if args.StartDate != "" && !validDate(args.StartDate) {
		return args, errors.New("start_date must be YYYY-MM-DD")
	}
	if len(args.GoalName) > 120 {
		return args, errors.New("goal_name must be at most 120 characters")
	}
	if args.GoalType == "" {
		args.GoalType = "race"
	}
	if _, ok := allowedSeasonGoalTypes[args.GoalType]; !ok {
		return args, errors.New("goal_type must be race, event, season_goal, or other")
	}
	if len(args.StrengthContext) > 500 {
		return args, errors.New("strength_context must be at most 500 characters")
	}
	if len(args.Sports) > 12 {
		return args, errors.New("sports must include at most 12 entries")
	}
	for i, sport := range args.Sports {
		args.Sports[i] = strings.TrimSpace(sport)
		if args.Sports[i] == "" || len(args.Sports[i]) > 40 {
			return args, errors.New("sports entries must be non-empty and at most 40 characters")
		}
	}
	for field, value := range map[string]*float64{
		"target_weekly_load":    args.TargetWeeklyLoad,
		"current_weekly_load":   args.CurrentWeeklyLoad,
		"target_hours_per_week": args.TargetHoursPerWeek,
		"current_weekly_hours":  args.CurrentWeeklyHours,
		"max_hours_per_week":    args.MaxHoursPerWeek,
	} {
		if value == nil {
			continue
		}
		maxValue := 80.0
		if strings.Contains(field, "load") {
			maxValue = 7000
		}
		if *value < 0 || *value > maxValue || math.IsNaN(*value) || math.IsInf(*value, 0) {
			return args, fmt.Errorf("%s must be between 0 and %.0f", field, maxValue)
		}
	}
	if args.TaperWeeks != nil && (*args.TaperWeeks < seasonPlanMinTaperWeeks || *args.TaperWeeks > seasonPlanMaxTaperWeeks) {
		return args, fmt.Errorf("taper_weeks must be between %d and %d", seasonPlanMinTaperWeeks, seasonPlanMaxTaperWeeks)
	}
	if args.TaperTargetLoadPct != nil && (*args.TaperTargetLoadPct < seasonPlanMinTaperLoadPct || *args.TaperTargetLoadPct > seasonPlanMaxTaperLoadPct || math.IsNaN(*args.TaperTargetLoadPct) || math.IsInf(*args.TaperTargetLoadPct, 0)) {
		return args, fmt.Errorf("taper_target_load_pct must be between %.0f and %.0f", seasonPlanMinTaperLoadPct, seasonPlanMaxTaperLoadPct)
	}
	return args, nil
}

func resolveSeasonPlanProposalInputs(args seasonPlanProposalRequest, asOfDate time.Time) (seasonPlanResolvedInputs, error) {
	goalDate, _ := time.Parse(time.DateOnly, args.GoalDate)
	startDate := nextStrictMonday(asOfDate)
	if args.StartDate != "" {
		parsedStart, _ := time.Parse(time.DateOnly, args.StartDate)
		if parsedStart.Weekday() != time.Monday {
			return seasonPlanResolvedInputs{}, errors.New("start_date must be a Monday")
		}
		if !parsedStart.After(asOfDate) {
			return seasonPlanResolvedInputs{}, errors.New("start_date must be strictly after athlete-local today")
		}
		startDate = parsedStart
	}
	if goalDate.Before(startDate) {
		return seasonPlanResolvedInputs{}, errors.New("goal_date must be on or after resolved start_date")
	}
	goalWeekStart := mondayOf(goalDate)
	totalWeeks := int(goalWeekStart.Sub(startDate).Hours()/24)/7 + 1
	if totalWeeks < 1 || totalWeeks > seasonPlanMaxWeeks {
		return seasonPlanResolvedInputs{}, fmt.Errorf("proposal horizon must be 1-%d weeks", seasonPlanMaxWeeks)
	}
	resolved := seasonPlanResolvedInputs{startDate: startDate, goalDate: goalDate, goalWeekStart: goalWeekStart, maxHoursPerWeek: args.MaxHoursPerWeek}
	resolved.currentWeeklyLoad, resolved.loadSource = resolveCurrentWeeklyLoad(args, &resolved)
	resolved.targetWeeklyLoad = resolveTargetWeeklyLoad(args, resolved.currentWeeklyLoad, &resolved)
	resolved.currentWeeklyHours, resolved.targetWeeklyHours, resolved.hoursSource = resolveWeeklyHours(args, resolved.currentWeeklyLoad, resolved.targetWeeklyLoad, &resolved)
	resolved.taperWeeks = seasonPlanDefaultTaperWeeks
	resolved.taperWeeksSource = "default"
	resolved.taperAssumptionCodes = []string{"taper_weeks_default"}
	if args.TaperWeeks != nil {
		resolved.taperWeeks = *args.TaperWeeks
		resolved.taperWeeksSource = "input"
		resolved.taperAssumptionCodes[0] = "taper_weeks_input"
		resolved.addAssumption("taper_weeks_input", "taper duration was provided by the caller; this is a load-only scenario input", "taper_weeks", *args.TaperWeeks)
	}
	resolved.taperTargetLoadPct = seasonPlanDefaultTaperLoadPct
	resolved.taperTargetSource = "default"
	resolved.taperAssumptionCodes = append(resolved.taperAssumptionCodes, "taper_target_load_pct_default")
	if args.TaperTargetLoadPct != nil {
		resolved.taperTargetLoadPct = *args.TaperTargetLoadPct
		resolved.taperTargetSource = "input"
		resolved.taperAssumptionCodes[1] = "taper_target_load_pct_input"
		resolved.addAssumption("taper_target_load_pct_input", "taper target load percentage was provided by the caller; no intensity or suitability is inferred", "taper_target_load_pct", *args.TaperTargetLoadPct)
	}
	if resolved.taperWeeks > totalWeeks {
		resolved.taperWeeks = totalWeeks
		resolved.taperAdjustmentCode = "taper_weeks_clamped_to_horizon"
		resolved.taperAssumptionCodes = append(resolved.taperAssumptionCodes, resolved.taperAdjustmentCode)
		resolved.addAssumption(resolved.taperAdjustmentCode, "requested taper duration exceeded the proposal horizon and was clamped to the available weeks", "taper_weeks", totalWeeks)
	} else {
		resolved.taperAdjustmentCode = "none"
	}
	return resolved, nil
}

func resolveCurrentWeeklyLoad(args seasonPlanProposalRequest, resolved *seasonPlanResolvedInputs) (float64, string) {
	if args.CurrentWeeklyLoad != nil {
		return *args.CurrentWeeklyLoad, "input_current_weekly_load"
	}
	if args.TargetWeeklyLoad != nil {
		resolved.addAssumption("current_load_from_target", "current weekly load omitted; using target_weekly_load as the current-load baseline", "current_weekly_load", *args.TargetWeeklyLoad)
		resolved.addWarning("missing_current_load", "current CTL/current weekly load was not provided or fetched", "current_weekly_load", nil)
		return *args.TargetWeeklyLoad, "fallback_target_weekly_load"
	}
	if args.CurrentWeeklyHours != nil {
		load := *args.CurrentWeeklyHours * seasonPlanLoadPerHour
		resolved.addAssumption("current_load_from_hours", "current weekly load derived from current_weekly_hours at 50 load points per hour", "current_weekly_hours", *args.CurrentWeeklyHours)
		resolved.addWarning("missing_current_load", "current CTL/current weekly load was not provided or fetched", "current_weekly_load", nil)
		return load, "fallback_current_hours_x_50"
	}
	resolved.addAssumption("default_current_weekly_load", "current weekly load omitted; using 300 load points as a conservative default", "current_weekly_load", seasonPlanDefaultWeeklyLoad)
	resolved.addWarning("missing_current_load", "current CTL/current weekly load was not provided or fetched", "current_weekly_load", nil)
	return seasonPlanDefaultWeeklyLoad, "fallback_default_300"
}

func resolveTargetWeeklyLoad(args seasonPlanProposalRequest, currentWeeklyLoad float64, resolved *seasonPlanResolvedInputs) float64 {
	if args.TargetWeeklyLoad != nil {
		return *args.TargetWeeklyLoad
	}
	if args.TargetHoursPerWeek != nil {
		load := *args.TargetHoursPerWeek * seasonPlanLoadPerHour
		resolved.addAssumption("target_load_from_hours", "target weekly load derived from target_hours_per_week at 50 load points per hour", "target_hours_per_week", *args.TargetHoursPerWeek)
		return load
	}
	load := currentWeeklyLoad * seasonPlanDefaultTargetFactor
	resolved.addAssumption("default_target_weekly_load", "target weekly load omitted; using current weekly load multiplied by 1.15", "target_weekly_load", round(load, 1))
	return load
}

func resolveWeeklyHours(args seasonPlanProposalRequest, currentWeeklyLoad float64, targetWeeklyLoad float64, resolved *seasonPlanResolvedInputs) (float64, float64, string) {
	currentHours := currentWeeklyLoad / seasonPlanLoadPerHour
	targetHours := targetWeeklyLoad / seasonPlanLoadPerHour
	source := "derived_load_div_50"
	if args.CurrentWeeklyHours != nil {
		currentHours = *args.CurrentWeeklyHours
		source = "input_current_weekly_hours"
	} else {
		resolved.addAssumption("current_hours_from_load", "current weekly hours derived from weekly load at 50 load points per hour", "current_weekly_hours", round(currentHours, 2))
	}
	if args.TargetHoursPerWeek != nil {
		targetHours = *args.TargetHoursPerWeek
		source = "input_target_hours_per_week"
	} else {
		resolved.addAssumption("target_hours_from_load", "target weekly hours derived from target weekly load at 50 load points per hour", "target_hours_per_week", round(targetHours, 2))
	}
	if args.MaxHoursPerWeek != nil {
		if currentHours > *args.MaxHoursPerWeek {
			currentHours = *args.MaxHoursPerWeek
			resolved.addWarning("infeasible_hours_cap", "current load-derived hours exceed max_hours_per_week; hours are capped while load targets are preserved", "max_hours_per_week", *args.MaxHoursPerWeek)
		}
		if targetHours > *args.MaxHoursPerWeek {
			targetHours = *args.MaxHoursPerWeek
			resolved.addWarning("infeasible_hours_cap", "target load-derived hours exceed max_hours_per_week; hours are capped while load targets are preserved", "max_hours_per_week", *args.MaxHoursPerWeek)
		}
	}
	return currentHours, targetHours, source
}

func buildSeasonPlanProposal(args seasonPlanProposalRequest, resolved seasonPlanResolvedInputs, timezoneName string) seasonPlanProposalResponse {
	totalWeeks := int(resolved.goalWeekStart.Sub(resolved.startDate).Hours()/24)/7 + 1
	phases := allocateSeasonPlanPhases(totalWeeks, resolved.startDate, resolved.taperWeeks)
	weeklyTargets, recoveryWeeks := buildSeasonPlanWeeklyTargets(phases, resolved, totalWeeks)
	projectionTargets := make([]seasonPlanProposalProjectionWeeklyTarget, 0, len(weeklyTargets))
	for _, target := range weeklyTargets {
		projectionTargets = append(projectionTargets, seasonPlanProposalProjectionWeeklyTarget{WeekStartDate: target.WeekStartDate, TrainingLoad: target.TrainingLoad})
	}
	endDate := resolved.startDate.AddDate(0, 0, totalWeeks*7-1)
	raceAnchor := seasonPlanProposalRaceAnchor{Date: args.GoalDate, Name: args.GoalName, Type: args.GoalType, Source: "input", WeekStartDate: formatDate(resolved.goalWeekStart)}
	assumptions := append([]seasonPlanProposalNotice{}, resolved.assumptions...)
	if len(args.Sports) > 0 {
		assumptions = append(assumptions, seasonPlanProposalNotice{Code: "sports_context_input_only", Message: "sports were provided by the caller and were not fetched or allocated from upstream data", Field: "sports", Value: args.Sports})
	}
	if args.StrengthContext != "" {
		assumptions = append(assumptions, seasonPlanProposalNotice{Code: "strength_context_input_only", Message: "strength context was provided by the caller and is not first-class upstream strength-set data", Field: "strength_context", Value: args.StrengthContext})
	}
	warnings := append([]seasonPlanProposalNotice{}, resolved.warnings...)
	warnings = append(warnings, seasonPlanProposalNotice{Code: "missing_power_curve_profile", Message: "power curve, CTL history, and athlete planning parameters were not fetched; load/hour conversions use explicit assumptions"})
	if len(args.Sports) > 1 {
		warnings = append(warnings, seasonPlanProposalNotice{Code: "multi_sport_not_allocated", Message: "multi-sport context is echoed as an assumption; weekly targets are not split by sport", Field: "sports", Value: args.Sports})
	}
	if args.StrengthContext != "" {
		warnings = append(warnings, seasonPlanProposalNotice{Code: "strength_not_first_class", Message: "strength context is an assumption only; no upstream strength endpoint was read", Field: "strength_context"})
	}
	maxHours := cloneFloatPtr(resolved.maxHoursPerWeek)
	return seasonPlanProposalResponse{
		Summary:       seasonPlanProposalSummary{StartDate: formatDate(resolved.startDate), EndDate: formatDate(endDate), GoalDate: args.GoalDate, TotalWeeks: totalWeeks, PhaseCount: len(phases), RecoveryWeekCount: len(recoveryWeeks), RaceAnchorCount: 1, TargetWeeklyLoad: round(resolved.targetWeeklyLoad, 1), TargetHoursPerWeek: round(resolved.targetWeeklyHours, 2), MaxHoursPerWeek: roundFloatPtr(maxHours, 2)},
		Phases:        phases,
		WeeklyTargets: weeklyTargets,
		RecoveryWeeks: recoveryWeeks,
		RaceAnchors:   []seasonPlanProposalRaceAnchor{raceAnchor},
		Assumptions:   assumptions,
		Warnings:      warnings,
		Meta:          seasonPlanProposalMeta{Method: seasonPlanProposalMethod, SourceTools: []string{"athlete_profile", "deterministic_season_plan_proposal"}, ReadOnly: true, WritesPerformed: false, Timezone: timezoneName, SchemaVersion: seasonPlanProposalSchemaVersion, PlanningSourceEndpoints: seasonPlanProposalSourceEndpoints{ProfileTimezone: "toolProfile", NotCalled: []string{"calendar_events", "training_plan", "fitness", "power_curves", "strength"}}, ProjectionBridge: seasonPlanProposalProjectionBridge{TargetTool: getFitnessProjectionName, TargetArgument: "weekly_plan_targets", WeeklyPlanTargets: projectionTargets, IncludedWeekCount: len(projectionTargets)}, TaperScenario: seasonPlanProposalTaperScenario{RequestedTaperWeeks: cloneIntPtr(args.TaperWeeks), ResolvedTaperWeeks: resolved.taperWeeks, TaperWeeksSource: resolved.taperWeeksSource, RequestedTaperTargetLoadPct: cloneFloatPtr(args.TaperTargetLoadPct), ResolvedTaperTargetLoadPct: round(resolved.taperTargetLoadPct, 2), TaperTargetLoadPctSource: resolved.taperTargetSource, Allocation: "final_contiguous_weeks", AdjustmentCode: resolved.taperAdjustmentCode, AssumptionCodes: append([]string(nil), resolved.taperAssumptionCodes...)}},
	}
}

func allocateSeasonPlanPhases(totalWeeks int, startDate time.Time, taperWeeks ...int) []seasonPlanProposalPhase {
	resolvedTaperWeeks := seasonPlanDefaultTaperWeeks
	if len(taperWeeks) > 0 {
		resolvedTaperWeeks = taperWeeks[0]
	}
	counts := phaseWeekCounts(totalWeeks, resolvedTaperWeeks)
	phases := make([]seasonPlanProposalPhase, 0, len(counts))
	weekStart := 1
	phaseSeq := 1
	for _, spec := range counts {
		if spec.weeks == 0 {
			continue
		}
		phaseStart := startDate.AddDate(0, 0, (weekStart-1)*7)
		phaseEnd := phaseStart.AddDate(0, 0, spec.weeks*7-1)
		phaseID := fmt.Sprintf("phase_%02d_%s", phaseSeq, spec.phaseType)
		phases = append(phases, seasonPlanProposalPhase{PhaseID: phaseID, PhaseType: spec.phaseType, Name: phaseName(spec.phaseType), StartDate: formatDate(phaseStart), EndDate: formatDate(phaseEnd), WeekCount: spec.weeks, StartWeekIndex: weekStart, EndWeekIndex: weekStart + spec.weeks - 1})
		weekStart += spec.weeks
		phaseSeq++
	}
	return phases
}

type seasonPlanPhaseCount struct {
	phaseType string
	weeks     int
}

func phaseWeekCounts(totalWeeks int, taperWeeks ...int) []seasonPlanPhaseCount {
	resolvedTaperWeeks := seasonPlanDefaultTaperWeeks
	if len(taperWeeks) > 0 {
		resolvedTaperWeeks = taperWeeks[0]
	}
	if resolvedTaperWeeks < seasonPlanMinTaperWeeks {
		resolvedTaperWeeks = seasonPlanMinTaperWeeks
	}
	if resolvedTaperWeeks > totalWeeks {
		resolvedTaperWeeks = totalWeeks
	}
	if totalWeeks == 1 || resolvedTaperWeeks == totalWeeks {
		return []seasonPlanPhaseCount{{phaseType: "race_taper", weeks: totalWeeks}}
	}
	if resolvedTaperWeeks == 1 {
		switch totalWeeks {
		case 2:
			return []seasonPlanPhaseCount{{phaseType: "build", weeks: 1}, {phaseType: "race_taper", weeks: 1}}
		case 3:
			return []seasonPlanPhaseCount{{phaseType: "base", weeks: 1}, {phaseType: "build", weeks: 1}, {phaseType: "race_taper", weeks: 1}}
		default:
			remaining := totalWeeks - 1
			base := int(math.Floor(float64(remaining) * 0.5))
			build := int(math.Floor(float64(remaining) * 0.3))
			peak := remaining - base - build
			return []seasonPlanPhaseCount{{phaseType: "base", weeks: base}, {phaseType: "build", weeks: build}, {phaseType: "peak", weeks: peak}, {phaseType: "race_taper", weeks: 1}}
		}
	}
	remaining := totalWeeks - resolvedTaperWeeks
	if remaining == 1 {
		return []seasonPlanPhaseCount{{phaseType: "base", weeks: 1}, {phaseType: "race_taper", weeks: resolvedTaperWeeks}}
	}
	if remaining == 2 {
		return []seasonPlanPhaseCount{{phaseType: "base", weeks: 1}, {phaseType: "build", weeks: 1}, {phaseType: "race_taper", weeks: resolvedTaperWeeks}}
	}
	base := int(math.Floor(float64(remaining) * 0.5))
	build := int(math.Floor(float64(remaining) * 0.3))
	peak := remaining - base - build
	return []seasonPlanPhaseCount{{phaseType: "base", weeks: base}, {phaseType: "build", weeks: build}, {phaseType: "peak", weeks: peak}, {phaseType: "race_taper", weeks: resolvedTaperWeeks}}
}

func buildSeasonPlanWeeklyTargets(phases []seasonPlanProposalPhase, resolved seasonPlanResolvedInputs, totalWeeks int) ([]seasonPlanProposalWeeklyTarget, []seasonPlanProposalRecoveryWeek) {
	weeklyTargets := make([]seasonPlanProposalWeeklyTarget, 0, totalWeeks)
	recoveryWeeks := []seasonPlanProposalRecoveryWeek{}
	for _, phase := range phases {
		for weekIndex := phase.StartWeekIndex; weekIndex <= phase.EndWeekIndex; weekIndex++ {
			weekStart := resolved.startDate.AddDate(0, 0, (weekIndex-1)*7)
			load := interpolate(resolved.currentWeeklyLoad, resolved.targetWeeklyLoad, weekIndex, totalWeeks)
			hours := interpolate(resolved.currentWeeklyHours, resolved.targetWeeklyHours, weekIndex, totalWeeks)
			isTaper := phase.PhaseType == "race_taper"
			isRecovery := weekIndex%seasonPlanRecoveryCadenceWeeks == 0 && !isTaper
			loadSource := resolved.loadSource
			if isTaper {
				load = resolved.targetWeeklyLoad * resolved.taperTargetLoadPct / 100
				if resolved.taperTargetLoadPct == seasonPlanRecoveryLoadPct {
					loadSource = "race_taper_60_pct_target"
				} else {
					loadSource = "race_taper_target_pct"
				}
			} else if isRecovery {
				load = load * seasonPlanRecoveryLoadPct / 100
				loadSource = "recovery_60_pct_interpolated"
			}
			if resolved.maxHoursPerWeek != nil && hours > *resolved.maxHoursPerWeek {
				hours = *resolved.maxHoursPerWeek
			}
			target := seasonPlanProposalWeeklyTarget{WeekStartDate: formatDate(weekStart), WeekEndDate: formatDate(weekStart.AddDate(0, 0, 6)), WeekIndex: weekIndex, PhaseID: phase.PhaseID, PhaseType: phase.PhaseType, TrainingLoad: round(load, 1), TargetHours: round(hours, 2), IsRecoveryWeek: isRecovery, IsTaperWeek: isTaper, LoadSource: loadSource, HoursSource: resolved.hoursSource}
			weeklyTargets = append(weeklyTargets, target)
			if isRecovery {
				recoveryWeeks = append(recoveryWeeks, seasonPlanProposalRecoveryWeek{WeekStartDate: target.WeekStartDate, WeekIndex: weekIndex, PhaseID: phase.PhaseID, TrainingLoad: target.TrainingLoad, RecoveryLoadPct: seasonPlanRecoveryLoadPct, Reason: "every_4th_week"})
			}
		}
	}
	return weeklyTargets, recoveryWeeks
}

func interpolate(start, end float64, oneBasedIndex, totalWeeks int) float64 {
	if totalWeeks <= 1 {
		return end
	}
	position := float64(oneBasedIndex-1) / float64(totalWeeks-1)
	return start + (end-start)*position
}

func nextStrictMonday(date time.Time) time.Time {
	days := (int(time.Monday) - int(date.Weekday()) + 7) % 7
	if days == 0 {
		days = 7
	}
	return date.AddDate(0, 0, days)
}

func mondayOf(date time.Time) time.Time {
	daysSinceMonday := (int(date.Weekday()) - int(time.Monday) + 7) % 7
	return date.AddDate(0, 0, -daysSinceMonday)
}

func phaseName(phaseType string) string {
	switch phaseType {
	case "base":
		return "Base"
	case "build":
		return "Build"
	case "peak":
		return "Peak"
	case "race_taper":
		return "Race taper"
	default:
		return phaseType
	}
}

func (r *seasonPlanResolvedInputs) addAssumption(code, message, field string, value any) {
	r.assumptions = append(r.assumptions, seasonPlanProposalNotice{Code: code, Message: message, Field: field, Value: value})
}

func (r *seasonPlanResolvedInputs) addWarning(code, message, field string, value any) {
	r.warnings = append(r.warnings, seasonPlanProposalNotice{Code: code, Message: message, Field: field, Value: value})
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func cloneFloatPtr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func seasonPlanProposalInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"goal_date"},
		"properties": map[string]any{
			"start_date":            map[string]any{"type": "string", "description": "Optional athlete-local Monday YYYY-MM-DD start. Must be strictly after athlete-local today; omitted defaults to the next Monday strictly after today."},
			"goal_date":             map[string]any{"type": "string", "description": "Required athlete-local YYYY-MM-DD goal/race date. Proposal horizon is capped at 53 inclusive weeks."},
			"goal_name":             map[string]any{"type": "string", "maxLength": 120, "description": "Optional race or season-goal name."},
			"goal_type":             map[string]any{"type": "string", "enum": []string{"race", "event", "season_goal", "other"}, "default": "race", "description": "Caller-provided anchor type; v1 does not fetch existing race events."},
			"target_weekly_load":    map[string]any{"type": "number", "minimum": 0, "maximum": 7000, "description": "Target weekly TSS/load points for the goal build."},
			"current_weekly_load":   map[string]any{"type": "number", "minimum": 0, "maximum": 7000, "description": "Current weekly TSS/load points. Omit to use explicit documented fallbacks and warnings."},
			"target_hours_per_week": map[string]any{"type": "number", "minimum": 0, "maximum": 80, "description": "Target weekly training hours; may cap time targets independently from load."},
			"current_weekly_hours":  map[string]any{"type": "number", "minimum": 0, "maximum": 80, "description": "Current weekly training hours used as a fallback when current_weekly_load is absent."},
			"max_hours_per_week":    map[string]any{"type": "number", "minimum": 0, "maximum": 80, "description": "Hard weekly time cap; infeasible load-derived hours warn instead of exceeding this cap."},
			"taper_weeks":           map[string]any{"type": "integer", "minimum": seasonPlanMinTaperWeeks, "maximum": seasonPlanMaxTaperWeeks, "default": seasonPlanDefaultTaperWeeks, "description": "Optional number of final contiguous taper weeks to model; bounded 1-3 and clamped to a shorter proposal horizon. This changes load scenario only, not intensity or calendar data."},
			"taper_target_load_pct": map[string]any{"type": "number", "minimum": seasonPlanMinTaperLoadPct, "maximum": seasonPlanMaxTaperLoadPct, "default": seasonPlanDefaultTaperLoadPct, "description": "Optional taper weekly load as a percentage of target_weekly_load, bounded 1-100. This is a caller-selected deterministic scenario, not a suitability or physiology recommendation."},
			"sports":                map[string]any{"type": "array", "maxItems": 12, "items": map[string]any{"type": "string", "maxLength": 40}, "description": "Caller-provided sport context echoed as assumptions only; no sport allocation is fetched or invented."},
			"strength_context":      map[string]any{"type": "string", "maxLength": 500, "description": "Free-form caller-provided strength context echoed as an assumption; not first-class upstream strength data."},
			"include_full":          map[string]any{"type": "boolean", "default": false, "description": "When true, include all proposal rows after normal response shaping."},
		},
	}
}

func seasonPlanProposalOutputSchema() map[string]any {
	return genericOutputSchema("Read-only deterministic season-plan proposal with phases, weekly_targets, recovery_weeks, race_anchors, assumptions, warnings, and _meta.projection_bridge.weekly_plan_targets for get_fitness_projection.")
}
