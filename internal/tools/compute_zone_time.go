package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/resources"
)

const (
	computeZoneTimeName                = "compute_zone_time"
	computeLoadBalanceName             = "compute_load_balance"
	computeZoneTimeDescription         = "Compute deterministic zone time from upstream precomputed zone fields; do not fetch rows or streams and reduce manually. Returns polarization metadata and explicit missing/unavailable signals."
	computeLoadBalanceDescription      = "Compute deterministic low/moderate/high load balance from upstream precomputed zone fields; do not fetch rows or streams and reduce manually. Returns polarization classification and explicit missing/unavailable signals."
	invalidComputeZoneArgumentsMessage = "invalid compute zone arguments; provide valid dates, optional sport, and zone_metric power/heart_rate/pace"
	fetchComputeZoneMessage            = "could not compute zone aggregate; check intervals.icu credentials, athlete ID, and date range"
	maxComputeActivityCandidates       = 500
)

type computeZoneRequest struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	ZoneMetric  string `json:"zone_metric,omitempty"`
	Sport       string `json:"sport,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type zoneAggregate struct {
	Zones          []float64
	Rows           []zoneSeriesRow
	SourceTools    []string
	MissingSources []string
	MissingDays    int
	N              int
	Truncated      bool
	UsedSummary    bool
}

type zoneSeriesRow struct {
	Date            string    `json:"date,omitempty"`
	ActivityID      string    `json:"activity_id,omitempty"`
	Sport           string    `json:"sport,omitempty"`
	SourceKey       string    `json:"source_key,omitempty"`
	ZoneSeconds     []float64 `json:"zone_seconds,omitempty"`
	LowSeconds      float64   `json:"low_seconds,omitempty"`
	ModerateSeconds float64   `json:"moderate_seconds,omitempty"`
	HighSeconds     float64   `json:"high_seconds,omitempty"`
	MissingReason   string    `json:"missing_reason,omitempty"`
}

type computeZoneResult struct {
	Status                      string           `json:"status"`
	ZoneMetric                  string           `json:"zone_metric"`
	Sport                       string           `json:"sport,omitempty"`
	StartDate                   string           `json:"start_date"`
	EndDate                     string           `json:"end_date"`
	Zones                       []computeZoneRow `json:"zones,omitempty"`
	TotalSeconds                float64          `json:"total_seconds,omitempty"`
	PolarizationIndex           *float64         `json:"polarization_index,omitempty"`
	PolarizationState           string           `json:"polarization_state,omitempty"`
	Classification              string           `json:"classification,omitempty"`
	MissingSources              []string         `json:"missing_sources,omitempty"`
	TruncatedActivityCandidates bool             `json:"truncated_activity_candidates,omitempty"`
	InsufficientReason          string           `json:"insufficient_reason,omitempty"`
}

type computeZoneRow struct {
	Zone    int     `json:"zone"`
	Seconds float64 `json:"seconds"`
	Share   float64 `json:"share"`
}

type computeLoadBalanceResult struct {
	Status                      string          `json:"status"`
	ZoneMetric                  string          `json:"zone_metric"`
	Sport                       string          `json:"sport,omitempty"`
	Buckets                     []loadBucketRow `json:"buckets,omitempty"`
	PolarizationIndex           *float64        `json:"polarization_index,omitempty"`
	PolarizationState           string          `json:"polarization_state,omitempty"`
	Classification              string          `json:"classification,omitempty"`
	TrainingLoadTotal           *float64        `json:"training_load_total,omitempty"`
	MissingSources              []string        `json:"missing_sources,omitempty"`
	TruncatedActivityCandidates bool            `json:"truncated_activity_candidates,omitempty"`
	InsufficientReason          string          `json:"insufficient_reason,omitempty"`
}

type loadBucketRow struct {
	Bucket  string  `json:"bucket"`
	Seconds float64 `json:"seconds"`
	Share   float64 `json:"share"`
}

func newComputeZoneTimeTool(fitnessClient FitnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: computeZoneTimeName, Description: computeZoneTimeDescription, InputSchema: computeZoneInputSchema(true), OutputSchema: genericOutputSchema("Deterministic precomputed zone-time aggregate with analyzer metadata."), Handler: computeZoneTimeHandler(fitnessClient, activitiesClient, extendedClient, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func newComputeLoadBalanceTool(fitnessClient FitnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: computeLoadBalanceName, Description: computeLoadBalanceDescription, InputSchema: computeZoneInputSchema(false), OutputSchema: genericOutputSchema("Deterministic precomputed zone load-balance aggregate with analyzer metadata."), Handler: computeLoadBalanceHandler(fitnessClient, activitiesClient, extendedClient, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func computeZoneTimeHandler(fitnessClient FitnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeComputeZoneRequest(req.Arguments, true)
		if err != nil {
			return Result{}, NewUserError(invalidComputeZoneArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchComputeZoneMessage, err)
		}
		agg, err := collectZoneAggregate(ctx, args, fitnessClient, activitiesClient, extendedClient)
		if err != nil {
			return computeZoneUserError(err)
		}
		balance := analysis.ComputeZoneBalance(agg.Zones)
		result := zoneTimeResult(args, agg, balance)
		meta := zoneAnalyzerMeta("precomputed_zone_time_sum", agg, balance, map[string]any{"zone_metric": args.ZoneMetric, "precomputed_only": true, "activity_candidates_truncated": agg.Truncated})
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: agg.Rows, Meta: meta}, args.IncludeFull, version, debugMetadata, computeZoneTimeName, unitSystem, shapeCfg)
	}
}

func computeLoadBalanceHandler(fitnessClient FitnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeComputeZoneRequest(req.Arguments, false)
		if err != nil {
			return Result{}, NewUserError(invalidComputeZoneArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchComputeZoneMessage, err)
		}
		agg, err := collectZoneAggregate(ctx, args, fitnessClient, activitiesClient, extendedClient)
		if err != nil {
			return computeZoneUserError(err)
		}
		balance := analysis.ComputeZoneBalance(agg.Zones)
		result := loadBalanceResult(args, agg, balance)
		meta := zoneAnalyzerMeta("precomputed_zone_load_balance", agg, balance, map[string]any{"zone_metric": args.ZoneMetric, "classification_rule": "polarized low>=0.70 and high>=moderate; pyramidal low>moderate>high; threshold moderate>=0.30 and >= low or high", "precomputed_only": true, "activity_candidates_truncated": agg.Truncated})
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: agg.Rows, Meta: meta}, args.IncludeFull, version, debugMetadata, computeLoadBalanceName, unitSystem, shapeCfg)
	}
}

func decodeComputeZoneRequest(raw json.RawMessage, requireMetric bool) (computeZoneRequest, error) {
	var args computeZoneRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[computeZoneRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.ZoneMetric = strings.ToLower(strings.TrimSpace(args.ZoneMetric))
	args.Sport = strings.TrimSpace(args.Sport)
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date must be YYYY-MM-DD")
	}
	if args.EndDate < args.StartDate {
		return args, errors.New("end_date must be on or after start_date")
	}
	if args.ZoneMetric == "" && !requireMetric {
		args.ZoneMetric = "power"
	}
	if args.ZoneMetric != "power" && args.ZoneMetric != "heart_rate" && args.ZoneMetric != "pace" {
		return args, errors.New("zone_metric must be power, heart_rate, or pace")
	}
	return args, nil
}

func collectZoneAggregate(ctx context.Context, args computeZoneRequest, fitnessClient FitnessClient, activitiesClient ActivitiesClient, extendedClient ExtendedMetricsClient) (zoneAggregate, error) {
	var agg zoneAggregate
	needsActivities := args.Sport != "" || args.ZoneMetric != "power"
	if !needsActivities && fitnessClient != nil {
		rows, err := fitnessClient.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.StartDate, End: args.EndDate})
		if err != nil {
			if contextError(err) {
				return agg, err
			}
			agg.MissingSources = append(agg.MissingSources, "get_training_summary:error")
		} else {
			agg.SourceTools = append(agg.SourceTools, getTrainingSummaryName)
			seenDates := map[string]bool{}
			for _, row := range rows {
				if len(row.TimeInZones) == 0 || row.TimeInZonesTot <= 0 {
					agg.Rows = append(agg.Rows, zoneSeriesRow{Date: row.Date, MissingReason: "missing_summary_time_in_zones"})
					continue
				}
				agg.Zones = addZoneSlices(agg.Zones, row.TimeInZones)
				agg.N++
				seenDates[row.Date] = true
				agg.UsedSummary = true
				balance := analysis.ComputeZoneBalance(row.TimeInZones)
				agg.Rows = append(agg.Rows, zoneSeriesRow{Date: row.Date, SourceKey: "timeInZones", ZoneSeconds: cloneFloatSlice(row.TimeInZones), LowSeconds: balance.LowSeconds, ModerateSeconds: balance.ModerateSeconds, HighSeconds: balance.HighSeconds})
			}
			agg.MissingDays = dateCount(args.StartDate, args.EndDate) - len(seenDates)
			if agg.N > 0 {
				return agg, nil
			}
		}
	}
	if activitiesClient == nil || extendedClient == nil {
		agg.MissingSources = append(agg.MissingSources, "get_activities_or_get_extended_metrics:missing_client")
		agg.MissingDays = dateCount(args.StartDate, args.EndDate)
		return agg, nil
	}
	activities, err := activitiesClient.ListActivities(ctx, intervals.ListActivitiesParams{Oldest: args.StartDate, Newest: args.EndDate, Limit: maxComputeActivityCandidates})
	if err != nil {
		return agg, err
	}
	agg.SourceTools = append(agg.SourceTools, getActivitiesName, getExtendedMetricsName)
	agg.Truncated = len(activities) >= maxComputeActivityCandidates
	sort.SliceStable(activities, func(i, j int) bool {
		if stringValue(activities[i].StartDateLocal) != stringValue(activities[j].StartDateLocal) {
			return stringValue(activities[i].StartDateLocal) < stringValue(activities[j].StartDateLocal)
		}
		return activities[i].ID < activities[j].ID
	})
	seenDates := map[string]bool{}
	for _, activity := range activities {
		if args.Sport != "" && !sameFold(args.Sport, stringValue(activity.Type)) {
			continue
		}
		date := localDatePrefix(stringValue(activity.StartDateLocal))
		if date == "" {
			date = localDatePrefix(stringValue(activity.StartDate))
		}
		zones, key := zoneSliceForMetric(activity.Raw, args.ZoneMetric)
		if len(zones) == 0 && extendedClient != nil && activity.ID != "" {
			detail, detailErr := extendedClient.GetActivity(ctx, activity.ID)
			if detailErr == nil {
				zones, key = zoneSliceForMetric(detail.Raw, args.ZoneMetric)
			}
		}
		if len(zones) == 0 {
			agg.Rows = append(agg.Rows, zoneSeriesRow{Date: date, ActivityID: activity.ID, Sport: stringValue(activity.Type), MissingReason: "missing_precomputed_zone_times"})
			continue
		}
		agg.Zones = addZoneSlices(agg.Zones, zones)
		agg.N++
		if date != "" {
			seenDates[date] = true
		}
		balance := analysis.ComputeZoneBalance(zones)
		agg.Rows = append(agg.Rows, zoneSeriesRow{Date: date, ActivityID: activity.ID, Sport: stringValue(activity.Type), SourceKey: key, ZoneSeconds: cloneFloatSlice(zones), LowSeconds: balance.LowSeconds, ModerateSeconds: balance.ModerateSeconds, HighSeconds: balance.HighSeconds})
	}
	agg.MissingDays = dateCount(args.StartDate, args.EndDate) - len(seenDates)
	if agg.N == 0 {
		agg.MissingSources = append(agg.MissingSources, "precomputed_zone_times")
	}
	return agg, nil
}

func zoneSliceForMetric(raw map[string]any, metric string) ([]float64, string) {
	switch metric {
	case "power":
		for _, key := range []string{"icu_zone_times", "power_zone_distribution_seconds", "power_zone_times"} {
			if values := rawNumberSlice(raw, key); len(values) > 0 {
				return values, key
			}
		}
	case "pace":
		for _, key := range []string{"gap_zone_times", "pace_zone_times", "pace_zone_time_seconds"} {
			if values := rawNumberSlice(raw, key); len(values) > 0 {
				return values, key
			}
		}
	case "heart_rate":
		for _, key := range []string{"hr_zone_times", "heartrate_zone_times", "heart_rate_zone_times", "hr_time_in_zones"} {
			if values := rawNumberSlice(raw, key); len(values) > 0 {
				return values, key
			}
		}
	}
	return nil, ""
}

func zoneTimeResult(args computeZoneRequest, agg zoneAggregate, balance analysis.ZoneBalance) computeZoneResult {
	status, reason := aggregateStatus(agg, balance)
	return computeZoneResult{Status: status, ZoneMetric: args.ZoneMetric, Sport: args.Sport, StartDate: args.StartDate, EndDate: args.EndDate, Zones: zoneRows(agg.Zones), TotalSeconds: round(balance.TotalSeconds, 3), PolarizationIndex: roundOptional(balance.Index), PolarizationState: balance.State, Classification: balance.Classification, MissingSources: agg.MissingSources, TruncatedActivityCandidates: agg.Truncated, InsufficientReason: reason}
}

func loadBalanceResult(args computeZoneRequest, agg zoneAggregate, balance analysis.ZoneBalance) computeLoadBalanceResult {
	status, reason := aggregateStatus(agg, balance)
	return computeLoadBalanceResult{Status: status, ZoneMetric: args.ZoneMetric, Sport: args.Sport, Buckets: []loadBucketRow{{Bucket: "low", Seconds: round(balance.LowSeconds, 3), Share: round(balance.LowShare, 4)}, {Bucket: "moderate", Seconds: round(balance.ModerateSeconds, 3), Share: round(balance.ModerateShare, 4)}, {Bucket: "high", Seconds: round(balance.HighSeconds, 3), Share: round(balance.HighShare, 4)}}, PolarizationIndex: roundOptional(balance.Index), PolarizationState: balance.State, Classification: balance.Classification, MissingSources: agg.MissingSources, TruncatedActivityCandidates: agg.Truncated, InsufficientReason: reason}
}

func aggregateStatus(agg zoneAggregate, balance analysis.ZoneBalance) (string, string) {
	if agg.N == 0 || balance.TotalSeconds == 0 {
		return "unavailable", "missing_precomputed_zone_times"
	}
	if agg.MissingDays > 0 || len(agg.MissingSources) > 0 || agg.Truncated {
		return "partial", "some_days_or_sources_missing"
	}
	return "ok", ""
}

func zoneAnalyzerMeta(method string, agg zoneAggregate, balance analysis.ZoneBalance, assumptions map[string]any) analysis.AnalyzerMetaInput {
	if balance.TotalSeconds > 0 {
		assumptions["polarization_state"] = balance.State
	}
	boundaries := []string{"precomputed zones only; raw streams are not reduced"}
	if agg.Truncated {
		boundaries = append(boundaries, "activity candidates truncated at deterministic cap")
	}
	return analysis.AnalyzerMetaInput{Method: method, SourceTools: agg.SourceTools, N: agg.N, MissingDays: agg.MissingDays, MissingAction: analysis.MissingActionSkip, FormulaRef: resources.AnalysisFormulaRefPolarization, Assumptions: assumptions, Boundaries: boundaries}
}

func zoneRows(zones []float64) []computeZoneRow {
	total := 0.0
	for _, value := range zones {
		total += value
	}
	out := make([]computeZoneRow, 0, len(zones))
	for idx, value := range zones {
		share := 0.0
		if total > 0 {
			share = value / total
		}
		out = append(out, computeZoneRow{Zone: idx + 1, Seconds: round(value, 3), Share: round(share, 4)})
	}
	return out
}

func addZoneSlices(base []float64, next []float64) []float64 {
	if len(base) < len(next) {
		grown := make([]float64, len(next))
		copy(grown, base)
		base = grown
	}
	for i, value := range next {
		if value > 0 {
			base[i] += value
		}
	}
	return base
}

func cloneFloatSlice(values []float64) []float64 {
	out := make([]float64, len(values))
	copy(out, values)
	return out
}
func roundOptional(value *float64) *float64 {
	if value == nil {
		return nil
	}
	rounded := round(*value, 4)
	return &rounded
}
func sameFold(a, b string) bool { return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) }
func contextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
func computeZoneUserError(err error) (Result, error) {
	if contextError(err) {
		return Result{}, err
	}
	return Result{}, NewUserError(fetchComputeZoneMessage, err)
}

func dateCount(start, end string) int {
	startTime, _ := time.Parse(time.DateOnly, start)
	endTime, _ := time.Parse(time.DateOnly, end)
	return int(endTime.Sub(startTime).Hours()/24) + 1
}

func computeZoneInputSchema(requireMetric bool) map[string]any {
	required := []string{"start_date", "end_date"}
	if requireMetric {
		required = append(required, "zone_metric")
	}
	return map[string]any{"type": "object", "additionalProperties": false, "required": required, "properties": map[string]any{
		"start_date":   map[string]any{"type": "string", "description": "Athlete-local inclusive start date YYYY-MM-DD."},
		"end_date":     map[string]any{"type": "string", "description": "Athlete-local inclusive end date YYYY-MM-DD."},
		"zone_metric":  map[string]any{"type": "string", "enum": []string{"power", "heart_rate", "pace"}, "default": "power", "description": "Precomputed zone family to aggregate; raw streams are never fetched as fallback."},
		"sport":        map[string]any{"type": "string", "description": "Optional exact case-insensitive activity sport/type filter."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include per-source audit rows. Default returns aggregate zones/buckets only."},
	}}
}

func _compileComputeZoneFmtUse() { _ = fmt.Sprintf }
