package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	getFitnessProjectionName        = "get_fitness_projection"
	getFitnessProjectionDescription = "Project CTL, ATL, and TSB forward from a current fitness row using deterministic load assumptions: weekly ramp %, optional recovery-week cadence, horizon, and optional explicit planned daily loads. Returns the curve only with include_full:true."
	invalidFitnessProjectionMessage = "invalid fitness projection arguments; provide start_date plus exactly one horizon_date or horizon_days, bounded ramp/recovery settings, and no free-form physiology model"
	fetchFitnessProjectionMessage   = "could not fetch current fitness data; check intervals.icu credentials, athlete ID, and start date"

	fitnessProjectionModel = "deterministic_ctl_atl_tsb"
)

const (
	defaultProjectionHorizonDays         = 42
	maxProjectionHorizonDays             = 180
	defaultProjectionWeeklyRampPct       = 5
	minProjectionWeeklyRampPct           = -50
	maxProjectionWeeklyRampPct           = 50
	defaultProjectionRecoveryWeekCadence = 4
	minProjectionRecoveryWeekCadence     = 2
	maxProjectionRecoveryWeekCadence     = 12
	defaultProjectionRecoveryLoadPct     = 60
	maxProjectionRecoveryLoadPct         = 100
	maxProjectionPlannedDailyLoad        = 1000
)

type fitnessProjectionRequest struct {
	StartDate               string                         `json:"start_date"`
	HorizonDate             string                         `json:"horizon_date,omitempty"`
	HorizonDays             int                            `json:"horizon_days,omitempty"`
	WeeklyRampPct           *float64                       `json:"weekly_ramp_pct,omitempty"`
	RecoveryWeekCadence     *int                           `json:"recovery_week_cadence,omitempty"`
	RecoveryWeekLoadPct     *float64                       `json:"recovery_week_load_pct,omitempty"`
	PlannedDailyLoads       []fitnessProjectionPlannedLoad `json:"planned_daily_loads,omitempty"`
	Model                   string                         `json:"model,omitempty"`
	IncludeFull             bool                           `json:"include_full,omitempty"`
	resolvedHorizonDays     int
	resolvedWeeklyRampPct   float64
	resolvedRecoveryCadence int
	resolvedRecoveryLoadPct float64
}

type fitnessProjectionPlannedLoad struct {
	Date         string  `json:"date"`
	TrainingLoad float64 `json:"training_load"`
}

func decodeFitnessProjectionRequest(raw json.RawMessage) (fitnessProjectionRequest, error) {
	var args fitnessProjectionRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[fitnessProjectionRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.HorizonDate = strings.TrimSpace(args.HorizonDate)
	args.Model = strings.TrimSpace(args.Model)
	if args.Model == "" {
		args.Model = fitnessProjectionModel
	}
	if args.Model != fitnessProjectionModel {
		return args, fmt.Errorf("model must be %q; free-form physiology models are not supported", fitnessProjectionModel)
	}
	if !validDate(args.StartDate) {
		return args, errors.New("start_date must be YYYY-MM-DD")
	}
	if args.HorizonDate != "" && args.HorizonDays != 0 {
		return args, errors.New("provide horizon_date or horizon_days, not both")
	}
	args.resolvedHorizonDays = defaultProjectionHorizonDays
	if args.HorizonDays != 0 {
		args.resolvedHorizonDays = args.HorizonDays
	}
	if args.HorizonDate != "" {
		days, err := projectionHorizonDays(args.StartDate, args.HorizonDate)
		if err != nil {
			return args, err
		}
		args.resolvedHorizonDays = days
	}
	if args.resolvedHorizonDays < 1 || args.resolvedHorizonDays > maxProjectionHorizonDays {
		return args, fmt.Errorf("horizon must be 1-%d days", maxProjectionHorizonDays)
	}
	args.resolvedWeeklyRampPct = defaultProjectionWeeklyRampPct
	if args.WeeklyRampPct != nil {
		args.resolvedWeeklyRampPct = *args.WeeklyRampPct
	}
	if args.resolvedWeeklyRampPct < minProjectionWeeklyRampPct || args.resolvedWeeklyRampPct > maxProjectionWeeklyRampPct {
		return args, fmt.Errorf("weekly_ramp_pct must be between %d and %d", minProjectionWeeklyRampPct, maxProjectionWeeklyRampPct)
	}
	args.resolvedRecoveryCadence = defaultProjectionRecoveryWeekCadence
	if args.RecoveryWeekCadence != nil {
		args.resolvedRecoveryCadence = *args.RecoveryWeekCadence
	}
	if args.resolvedRecoveryCadence != 0 && (args.resolvedRecoveryCadence < minProjectionRecoveryWeekCadence || args.resolvedRecoveryCadence > maxProjectionRecoveryWeekCadence) {
		return args, fmt.Errorf("recovery_week_cadence must be 0 or %d-%d", minProjectionRecoveryWeekCadence, maxProjectionRecoveryWeekCadence)
	}
	args.resolvedRecoveryLoadPct = defaultProjectionRecoveryLoadPct
	if args.RecoveryWeekLoadPct != nil {
		args.resolvedRecoveryLoadPct = *args.RecoveryWeekLoadPct
	}
	if args.resolvedRecoveryLoadPct < 0 || args.resolvedRecoveryLoadPct > maxProjectionRecoveryLoadPct {
		return args, fmt.Errorf("recovery_week_load_pct must be between 0 and %d", maxProjectionRecoveryLoadPct)
	}
	if err := validateProjectionPlannedLoads(args.StartDate, args.resolvedHorizonDays, args.PlannedDailyLoads); err != nil {
		return args, err
	}
	return args, nil
}

func projectionHorizonDays(startDate string, horizonDate string) (int, error) {
	if !validDate(horizonDate) {
		return 0, errors.New("horizon_date must be YYYY-MM-DD")
	}
	start, _ := time.Parse(time.DateOnly, startDate)
	horizon, _ := time.Parse(time.DateOnly, horizonDate)
	days := int(horizon.Sub(start).Hours() / 24)
	if days < 1 {
		return 0, errors.New("horizon_date must be after start_date")
	}
	return days, nil
}

func validateProjectionPlannedLoads(startDate string, horizonDays int, loads []fitnessProjectionPlannedLoad) error {
	start, _ := time.Parse(time.DateOnly, startDate)
	seen := map[string]struct{}{}
	for _, load := range loads {
		date := strings.TrimSpace(load.Date)
		if !validDate(date) {
			return errors.New("planned_daily_loads.date must be YYYY-MM-DD")
		}
		if _, ok := seen[date]; ok {
			return fmt.Errorf("planned_daily_loads contains duplicate date %s", date)
		}
		seen[date] = struct{}{}
		parsed, _ := time.Parse(time.DateOnly, date)
		offset := int(parsed.Sub(start).Hours() / 24)
		if offset < 1 || offset > horizonDays {
			return fmt.Errorf("planned_daily_loads date %s must be within the projection horizon after start_date", date)
		}
		if load.TrainingLoad < 0 || load.TrainingLoad > maxProjectionPlannedDailyLoad {
			return fmt.Errorf("planned_daily_loads training_load for %s must be between 0 and %d", date, maxProjectionPlannedDailyLoad)
		}
	}
	return nil
}

func fitnessProjectionInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"start_date"},
		"properties": map[string]any{
			"start_date":             map[string]any{"type": "string", "description": "Athlete-local YYYY-MM-DD date whose get_fitness CTL/ATL/TSB values seed the projection."},
			"horizon_date":           map[string]any{"type": "string", "description": "Optional athlete-local YYYY-MM-DD end date; provide either horizon_date or horizon_days, not both."},
			"horizon_days":           map[string]any{"type": "integer", "minimum": 1, "maximum": maxProjectionHorizonDays, "default": defaultProjectionHorizonDays, "description": "Optional number of days after start_date to simulate."},
			"weekly_ramp_pct":        map[string]any{"type": "number", "minimum": minProjectionWeeklyRampPct, "maximum": maxProjectionWeeklyRampPct, "default": defaultProjectionWeeklyRampPct, "description": "Week-over-week percent change applied to modeled daily load when planned_daily_loads does not specify a day."},
			"recovery_week_cadence":  map[string]any{"type": "integer", "minimum": 0, "maximum": maxProjectionRecoveryWeekCadence, "default": defaultProjectionRecoveryWeekCadence, "description": "Every Nth week uses recovery_week_load_pct of modeled load; use 0 to disable recovery weeks."},
			"recovery_week_load_pct": map[string]any{"type": "number", "minimum": 0, "maximum": maxProjectionRecoveryLoadPct, "default": defaultProjectionRecoveryLoadPct, "description": "Percent of modeled load used during recovery weeks."},
			"planned_daily_loads":    map[string]any{"type": "array", "description": "Optional explicit planned training load values by athlete-local date; these replace modeled ramp load for matching dates.", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"date", "training_load"}, "properties": map[string]any{"date": map[string]any{"type": "string", "description": "Athlete-local YYYY-MM-DD date after start_date and within the horizon."}, "training_load": map[string]any{"type": "number", "minimum": 0, "maximum": maxProjectionPlannedDailyLoad, "description": "Planned daily training load/TSS-equivalent stress for this date."}}}},
			"model":                  map[string]any{"type": "string", "enum": []string{fitnessProjectionModel}, "default": fitnessProjectionModel, "description": "Closed enum. Free-form physiology models are rejected; this tool only supports deterministic CTL/ATL/TSB impulse-response projection."},
			"include_full":           map[string]any{"type": "boolean", "default": false, "description": "When true, include the projected daily CTL/ATL/TSB curve series."},
		},
	}
}

func fitnessProjectionAssumptions(args fitnessProjectionRequest) map[string]any {
	return map[string]any{
		"model":                          fitnessProjectionModel,
		"horizon_days":                   args.resolvedHorizonDays,
		"weekly_ramp_pct":                round(args.resolvedWeeklyRampPct, 3),
		"recovery_week_cadence":          args.resolvedRecoveryCadence,
		"recovery_week_load_pct":         round(args.resolvedRecoveryLoadPct, 3),
		"planned_daily_load_count":       len(args.PlannedDailyLoads),
		"ctl_time_constant_days":         42,
		"atl_time_constant_days":         7,
		"modeled_load_without_plan":      "starts from current CTL as daily load proxy and changes by weekly_ramp_pct each week",
		"planned_load_override_behavior": "planned_daily_loads replace modeled load on matching dates only",
	}
}

func fitnessProjectionBoundaries() []string {
	return []string{
		"projection is deterministic scenario modeling, not a physiological prediction",
		"requires a start_date with non-null current CTL, ATL, and TSB from get_fitness",
		"does not read hidden upstream periodization or future calendar fields",
		fmt.Sprintf("horizon is capped at %d days", maxProjectionHorizonDays),
		"only the deterministic_ctl_atl_tsb model is supported",
	}
}
