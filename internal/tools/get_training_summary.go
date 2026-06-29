package tools

import (
	"context"
	"errors"
	"sort"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	getTrainingSummaryName        = "get_training_summary"
	getTrainingSummaryDescription = "Get aggregated training volume, neutral training load, sRPE, and upstream zone-order totals for a local date range."
	fetchTrainingSummaryMessage   = "could not fetch training summary; check intervals.icu credentials, athlete ID, and date range"
)

type trainingSummaryResponse struct {
	Summary trainingSummaryTotals `json:"summary"`
	Sports  []trainingSportTotals `json:"sports,omitempty"`
	Full    []map[string]any      `json:"full,omitempty"`
	Meta    trainingSummaryMeta   `json:"_meta"`
}

type trainingSummaryTotals struct {
	Count                   int       `json:"count"`
	TimeSeconds             int       `json:"time_seconds,omitempty"`
	MovingTimeSeconds       int       `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds      int       `json:"elapsed_time_seconds,omitempty"`
	CaloriesBurned          int       `json:"calories_burned,omitempty"`
	ElevationGainM          float64   `json:"elevation_gain_m,omitempty"`
	DistanceKM              *float64  `json:"distance_km,omitempty"`
	DistanceMI              *float64  `json:"distance_mi,omitempty"`
	TrainingLoad            int       `json:"training_load,omitempty"`
	SessionRPE              int       `json:"session_rpe,omitempty"`
	TimeInZonesSeconds      []float64 `json:"time_in_zones_seconds,omitempty"`
	TimeInZonesTotalSeconds int       `json:"time_in_zones_total_seconds,omitempty"`
}

type trainingSportTotals struct {
	Sport              string   `json:"sport"`
	Count              int      `json:"count"`
	TimeSeconds        int      `json:"time_seconds,omitempty"`
	MovingTimeSeconds  int      `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds int      `json:"elapsed_time_seconds,omitempty"`
	CaloriesBurned     int      `json:"calories_burned,omitempty"`
	ElevationGainM     float64  `json:"elevation_gain_m,omitempty"`
	DistanceKM         *float64 `json:"distance_km,omitempty"`
	DistanceMI         *float64 `json:"distance_mi,omitempty"`
	TrainingLoad       int      `json:"training_load,omitempty"`
	SessionRPE         int      `json:"session_rpe,omitempty"`
}

type trainingSummaryMeta struct {
	ServerVersion   string                       `json:"server_version"`
	StartDate       string                       `json:"start_date"`
	EndDate         string                       `json:"end_date"`
	Timezone        string                       `json:"timezone"`
	ZoneFamily      string                       `json:"zone_family"`
	ZoneOrder       string                       `json:"zone_order"`
	IncludeFull     bool                         `json:"include_full"`
	LoadDiagnostics []dataAvailabilityDiagnostic `json:"load_diagnostics,omitempty"`
}

func newGetTrainingSummaryTool(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getTrainingSummaryName, Description: getTrainingSummaryDescription, InputSchema: dateRangeInputSchema("local start date for summary rows"), OutputSchema: genericOutputSchema("Aggregated training summary."), Handler: getTrainingSummaryHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getTrainingSummaryHandler(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDateRangeRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidFitnessArgumentsMessage, err)
		}
		unitSystem, timezone, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchTrainingSummaryMessage, err)
		}
		rows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.StartDate, End: args.EndDate})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchTrainingSummaryMessage, err)
		}
		payload := shapeTrainingSummary(rows, args, timezone, unitSystem, version)
		return encodeShaped(payload, args.IncludeFull, []string{"sports"}, version, debugMetadata, getTrainingSummaryName, unitSystem, shapeCfg)
	}
}

func shapeTrainingSummary(rows []intervals.SummaryWithCats, args dateRangeRequest, timezone string, unitSystem response.UnitSystem, version string) trainingSummaryResponse {
	payload := trainingSummaryResponse{Meta: trainingSummaryMeta{ServerVersion: normalizeVersion(version), StartDate: args.StartDate, EndDate: args.EndDate, Timezone: timezone, ZoneFamily: "upstream_timeInZones", ZoneOrder: "upstream", IncludeFull: args.IncludeFull, LoadDiagnostics: loadDiagnostics(rows)}}
	categoryTotals := map[string]*trainingSportTotals{}
	var distanceMeters float64
	for _, row := range rows {
		payload.Summary.Count += row.Count
		payload.Summary.TimeSeconds += row.Time
		payload.Summary.MovingTimeSeconds += row.MovingTime
		payload.Summary.ElapsedTimeSeconds += row.ElapsedTime
		payload.Summary.CaloriesBurned += row.Calories
		payload.Summary.ElevationGainM += row.TotalElevationGain
		payload.Summary.TrainingLoad += summaryTrainingLoad(row).Value
		payload.Summary.SessionRPE += row.SRPE
		payload.Summary.TimeInZonesSeconds = addFloatSlices(payload.Summary.TimeInZonesSeconds, row.TimeInZones)
		payload.Summary.TimeInZonesTotalSeconds += row.TimeInZonesTot
		distanceMeters += row.Distance
		if args.IncludeFull {
			payload.Full = append(payload.Full, row.Raw)
		}
		for _, category := range row.ByCategory {
			total := categoryTotals[category.Category]
			if total == nil {
				total = &trainingSportTotals{Sport: category.Category}
				categoryTotals[category.Category] = total
			}
			total.Count += category.Count
			total.TimeSeconds += category.Time
			total.MovingTimeSeconds += category.MovingTime
			total.ElapsedTimeSeconds += category.ElapsedTime
			total.CaloriesBurned += category.Calories
			total.ElevationGainM += category.TotalElevationGain
			total.TrainingLoad += categoryTrainingLoad(category).Value
			total.SessionRPE += category.SRPE
			addDistance(total, category.Distance, unitSystem)
		}
	}
	setDistance(&payload.Summary, distanceMeters, unitSystem)
	for _, total := range categoryTotals {
		payload.Sports = append(payload.Sports, *total)
	}
	sort.Slice(payload.Sports, func(i, j int) bool { return payload.Sports[i].Sport < payload.Sports[j].Sport })
	return payload
}

func addFloatSlices(left []float64, right []float64) []float64 {
	if len(right) > len(left) {
		grown := make([]float64, len(right))
		copy(grown, left)
		left = grown
	}
	for i, value := range right {
		left[i] += value
	}
	return left
}

func setDistance(total *trainingSummaryTotals, meters float64, unitSystem response.UnitSystem) {
	converted := response.ToPreferred(meters, units.UnitM, unitSystem)
	value := round(converted.Value, 3)
	if converted.Unit == units.UnitMI {
		total.DistanceMI = &value
	} else {
		total.DistanceKM = &value
	}
}

func addDistance(total *trainingSportTotals, meters float64, unitSystem response.UnitSystem) {
	converted := response.ToPreferred(meters, units.UnitM, unitSystem)
	value := round(converted.Value, 3)
	if converted.Unit == units.UnitMI {
		total.DistanceMI = addPtr(total.DistanceMI, value)
	} else {
		total.DistanceKM = addPtr(total.DistanceKM, value)
	}
}

func addPtr(existing *float64, value float64) *float64 {
	if existing == nil {
		return &value
	}
	*existing = round(*existing+value, 3)
	return existing
}
