package analysis

import (
	"fmt"
	"math"
	"time"
)

const (
	// FitnessProjectionMethod is the closed deterministic model emitted in analyzer metadata.
	FitnessProjectionMethod = "deterministic_ctl_atl_tsb"
	// FitnessProjectionCTLTimeConstantDays is the CTL exponential-response time constant.
	FitnessProjectionCTLTimeConstantDays = 42
	// FitnessProjectionATLTimeConstantDays is the ATL exponential-response time constant.
	FitnessProjectionATLTimeConstantDays = 7
)

// FitnessProjectionInput describes a deterministic CTL/ATL/TSB scenario.
type FitnessProjectionInput struct {
	StartDate           string
	StartCTL            float64
	StartATL            float64
	StartTSB            float64
	HorizonDays         int
	WeeklyRampPct       float64
	RecoveryWeekCadence int
	RecoveryWeekLoadPct float64
	PlannedDailyLoads   []FitnessProjectionPlannedLoad
}

// FitnessProjectionPlannedLoad overrides modeled load for one projected date.
type FitnessProjectionPlannedLoad struct {
	Date         string
	TrainingLoad float64
}

// FitnessProjectionPoint is one daily CTL/ATL/TSB point in the projection curve.
type FitnessProjectionPoint struct {
	Date               string
	Day                int
	TrainingLoad       float64
	TrainingLoadSource string
	CTL                float64
	ATL                float64
	TSB                float64
}

// FitnessProjectionResult contains the full projected curve and aggregate summary values.
type FitnessProjectionResult struct {
	StartDate        string
	EndDate          string
	StartCTL         float64
	StartATL         float64
	StartTSB         float64
	EndCTL           float64
	EndATL           float64
	EndTSB           float64
	CTLChange        float64
	ATLChange        float64
	TSBChange        float64
	AverageDailyLoad float64
	TotalLoad        float64
	MinTSB           float64
	MaxCTL           float64
	Points           []FitnessProjectionPoint
}

// ProjectFitness projects CTL/ATL/TSB using first-order impulse-response updates.
func ProjectFitness(input FitnessProjectionInput) (FitnessProjectionResult, error) {
	if input.HorizonDays < 1 {
		return FitnessProjectionResult{}, fmt.Errorf("horizon days must be positive")
	}
	start, err := time.Parse(time.DateOnly, input.StartDate)
	if err != nil {
		return FitnessProjectionResult{}, fmt.Errorf("parsing start date: %w", err)
	}
	plannedLoads := map[string]float64{}
	for _, load := range input.PlannedDailyLoads {
		plannedLoads[load.Date] = load.TrainingLoad
	}
	ctl := input.StartCTL
	atl := input.StartATL
	points := make([]FitnessProjectionPoint, 0, input.HorizonDays+1)
	points = append(points, FitnessProjectionPoint{Date: input.StartDate, Day: 0, TrainingLoadSource: "current_fitness", CTL: ctl, ATL: atl, TSB: input.StartTSB})
	result := FitnessProjectionResult{StartDate: input.StartDate, StartCTL: input.StartCTL, StartATL: input.StartATL, StartTSB: input.StartTSB, MinTSB: input.StartTSB, MaxCTL: input.StartCTL}
	baseLoad := math.Max(input.StartCTL, 0)
	for day := 1; day <= input.HorizonDays; day++ {
		date := start.AddDate(0, 0, day).Format(time.DateOnly)
		load, source := modeledProjectionLoad(baseLoad, input.WeeklyRampPct, input.RecoveryWeekCadence, input.RecoveryWeekLoadPct, day)
		if planned, ok := plannedLoads[date]; ok {
			load = planned
			source = "planned_daily_loads"
		}
		ctl = ctl + (load-ctl)/FitnessProjectionCTLTimeConstantDays
		atl = atl + (load-atl)/FitnessProjectionATLTimeConstantDays
		tsb := ctl - atl
		point := FitnessProjectionPoint{Date: date, Day: day, TrainingLoad: load, TrainingLoadSource: source, CTL: ctl, ATL: atl, TSB: tsb}
		points = append(points, point)
		result.TotalLoad += load
		if tsb < result.MinTSB {
			result.MinTSB = tsb
		}
		if ctl > result.MaxCTL {
			result.MaxCTL = ctl
		}
	}
	last := points[len(points)-1]
	result.EndDate = last.Date
	result.EndCTL = last.CTL
	result.EndATL = last.ATL
	result.EndTSB = last.TSB
	result.CTLChange = result.EndCTL - result.StartCTL
	result.ATLChange = result.EndATL - result.StartATL
	result.TSBChange = result.EndTSB - result.StartTSB
	result.AverageDailyLoad = result.TotalLoad / float64(input.HorizonDays)
	result.Points = points
	return result, nil
}

func modeledProjectionLoad(baseLoad float64, weeklyRampPct float64, recoveryWeekCadence int, recoveryWeekLoadPct float64, day int) (float64, string) {
	week := (day - 1) / 7
	load := baseLoad * math.Pow(1+weeklyRampPct/100, float64(week))
	if recoveryWeekCadence > 0 && (week+1)%recoveryWeekCadence == 0 {
		return load * recoveryWeekLoadPct / 100, "modeled_recovery_week"
	}
	return load, "modeled_ramp"
}
