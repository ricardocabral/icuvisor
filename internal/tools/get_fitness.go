package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	getFitnessName                 = "get_fitness"
	getBestEffortsName             = "get_best_efforts"
	getPowerCurvesName             = "get_power_curves"
	getTrainingSummaryName         = "get_training_summary"
	getFitnessDescription          = "Get CTL, ATL, and TSB fitness trends for a local date range. Dates are athlete-local YYYY-MM-DD values."
	getBestEffortsDescription      = "Get upstream best efforts grouped by sport and default/requested power, heart-rate, and pace buckets. Defaults to Ride, Run, and Swim with terse bucket rows."
	getPowerCurvesDescription      = "Get the upstream-computed mean-maximal power curve for a date range. Default terse output returns only common duration buckets; include_full returns raw arrays."
	getTrainingSummaryDescription  = "Get aggregated training volume, neutral training load, sRPE, and upstream zone-order totals for a local date range."
	invalidFitnessArgumentsMessage = "invalid fitness arguments; provide start_date/end_date as YYYY-MM-DD and optional include_full"
	invalidCurveArgumentsMessage   = "invalid curve arguments; provide valid dates, sports, and positive bucket values"
	fetchFitnessMessage            = "could not fetch fitness data; check intervals.icu credentials, athlete ID, and date range"
	fetchBestEffortsMessage        = "could not fetch best efforts; check intervals.icu credentials, athlete ID, sports, and date range"
	fetchPowerCurvesMessage        = "could not fetch power curves; check intervals.icu credentials, athlete ID, sport, and date range"
	fetchTrainingSummaryMessage    = "could not fetch training summary; check intervals.icu credentials, athlete ID, and date range"
	defaultPowerCurveSport         = "Ride"
)

var (
	defaultBestEffortSports    = []string{"Ride", "Run", "Swim"}
	defaultDurationBuckets     = []int{5, 15, 30, 60, 300, 1200, 3600}
	defaultRunDistanceBuckets  = []int{400, 1000, 1609, 5000, 10000}
	defaultSwimDistanceBuckets = []int{50, 100, 200, 400, 1500}
)

// FitnessClient retrieves athlete summary rows.
type FitnessClient interface {
	ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error)
}

// PowerCurvesClient retrieves athlete curve sets.
type PowerCurvesClient interface {
	ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
}

// BestEffortsClient retrieves athlete curve sets for best efforts.
type BestEffortsClient interface {
	ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
	ListAthleteHRCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
	ListAthletePaceCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
}

type dateRangeRequest struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type powerCurvesRequest struct {
	Oldest          string `json:"oldest"`
	Newest          string `json:"newest"`
	Sport           string `json:"sport,omitempty"`
	DurationSeconds []int  `json:"duration_seconds,omitempty"`
	IncludeFull     bool   `json:"include_full,omitempty"`
}

type bestEffortsRequest struct {
	Oldest          string   `json:"oldest,omitempty"`
	Newest          string   `json:"newest,omitempty"`
	Sports          []string `json:"sports,omitempty"`
	DurationSeconds []int    `json:"duration_seconds,omitempty"`
	DistanceMeters  []int    `json:"distance_meters,omitempty"`
	IncludeFull     bool     `json:"include_full,omitempty"`
}

type fitnessResponse struct {
	Rows []fitnessRow `json:"fitness"`
	Meta fitnessMeta  `json:"_meta"`
}

type fitnessRow struct {
	Date string         `json:"date"`
	CTL  *float64       `json:"ctl,omitempty"`
	ATL  *float64       `json:"atl,omitempty"`
	TSB  *float64       `json:"tsb,omitempty"`
	Full map[string]any `json:"full,omitempty"`
}

type fitnessMeta struct {
	ServerVersion string `json:"server_version"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Timezone      string `json:"timezone"`
	Count         int    `json:"count"`
	IncludeFull   bool   `json:"include_full"`
}

type powerCurvesResponse struct {
	Sport  string            `json:"sport"`
	Points []powerCurvePoint `json:"points"`
	Full   map[string]any    `json:"full,omitempty"`
	Meta   powerCurvesMeta   `json:"_meta"`
}

type powerCurvePoint struct {
	DurationSeconds int      `json:"duration_seconds"`
	Watts           *float64 `json:"watts,omitempty"`
	ActivityID      string   `json:"activity_id,omitempty"`
}

type powerCurvesMeta struct {
	ServerVersion   string `json:"server_version"`
	Sport           string `json:"sport"`
	Oldest          string `json:"oldest"`
	Newest          string `json:"newest"`
	CurveSpec       string `json:"curve_spec"`
	DurationSeconds []int  `json:"duration_seconds"`
	MissingBuckets  []int  `json:"missing_buckets,omitempty"`
	IncludeFull     bool   `json:"include_full"`
}

type bestEffortsResponse struct {
	Sports []bestEffortsSport `json:"sports"`
	Meta   bestEffortsMeta    `json:"_meta"`
}

type bestEffortsSport struct {
	Sport   string          `json:"sport"`
	Efforts []bestEffortRow `json:"efforts,omitempty"`
	Full    map[string]any  `json:"full,omitempty"`
}

type bestEffortRow struct {
	Family          string   `json:"family"`
	DurationSeconds int      `json:"duration_seconds,omitempty"`
	DistanceMeters  int      `json:"distance_meters,omitempty"`
	PowerWatts      *float64 `json:"power_watts,omitempty"`
	HeartRateBPM    *float64 `json:"heart_rate_bpm,omitempty"`
	PaceValue       *float64 `json:"pace_value,omitempty"`
	ActivityID      string   `json:"activity_id,omitempty"`
}

type bestEffortsMeta struct {
	ServerVersion   string           `json:"server_version"`
	SportsRequested []string         `json:"sports_requested"`
	Oldest          string           `json:"oldest,omitempty"`
	Newest          string           `json:"newest,omitempty"`
	CurveSpec       string           `json:"curve_spec"`
	DurationSeconds []int            `json:"duration_seconds"`
	DistanceMeters  []int            `json:"distance_meters"`
	MissingBuckets  map[string][]int `json:"missing_buckets,omitempty"`
	IncludeFull     bool             `json:"include_full"`
}

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
	ServerVersion string `json:"server_version"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Timezone      string `json:"timezone"`
	ZoneFamily    string `json:"zone_family"`
	ZoneOrder     string `json:"zone_order"`
	IncludeFull   bool   `json:"include_full"`
}

func newGetFitnessTool(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getFitnessName, Description: getFitnessDescription, InputSchema: dateRangeInputSchema("local start date for fitness rows"), OutputSchema: genericOutputSchema("Fitness rows with CTL, ATL, and TSB."), Handler: getFitnessHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func newGetBestEffortsTool(client BestEffortsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getBestEffortsName, Description: getBestEffortsDescription, InputSchema: bestEffortsInputSchema(), OutputSchema: genericOutputSchema("Best efforts grouped by sport and bucket."), Handler: getBestEffortsHandler(client, version, debugMetadata, shapeCfg)})
}

func newGetPowerCurvesTool(client PowerCurvesClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getPowerCurvesName, Description: getPowerCurvesDescription, InputSchema: powerCurvesInputSchema(), OutputSchema: genericOutputSchema("Mean-maximal power curve bucket points."), Handler: getPowerCurvesHandler(client, version, debugMetadata, shapeCfg)})
}

func newGetTrainingSummaryTool(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getTrainingSummaryName, Description: getTrainingSummaryDescription, InputSchema: dateRangeInputSchema("local start date for summary rows"), OutputSchema: genericOutputSchema("Aggregated training summary."), Handler: getTrainingSummaryHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getFitnessHandler(client FitnessClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDateRangeRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidFitnessArgumentsMessage, err)
		}
		unitSystem, timezone, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchFitnessMessage, err)
		}
		rows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.StartDate, End: args.EndDate})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchFitnessMessage, err)
		}
		payload := fitnessResponse{Rows: shapeFitnessRows(rows, args.IncludeFull), Meta: fitnessMeta{ServerVersion: normalizeVersion(version), StartDate: args.StartDate, EndDate: args.EndDate, Timezone: timezone, Count: len(rows), IncludeFull: args.IncludeFull}}
		return encodeShaped(payload, args.IncludeFull, []string{"fitness"}, version, debugMetadata, getFitnessName, unitSystem, shapeCfg)
	}
}

func getPowerCurvesHandler(client PowerCurvesClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodePowerCurvesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCurveArgumentsMessage, err)
		}
		curveSpec := rangeCurveSpec(args.Oldest, args.Newest)
		set, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchPowerCurvesMessage, err)
		}
		points, missing := bucketPowerCurve(firstCurve(set), args.DurationSeconds)
		payload := powerCurvesResponse{Sport: args.Sport, Points: points, Meta: powerCurvesMeta{ServerVersion: normalizeVersion(version), Sport: args.Sport, Oldest: args.Oldest, Newest: args.Newest, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds, MissingBuckets: missing, IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = set.Raw
		}
		return encodeShaped(payload, args.IncludeFull, []string{"points"}, version, debugMetadata, getPowerCurvesName, response.UnitSystemMetric, shapeCfg)
	}
}

func getBestEffortsHandler(client BestEffortsClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeBestEffortsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCurveArgumentsMessage, err)
		}
		curveSpec := bestEffortsCurveSpec(args)
		missing := map[string][]int{}
		payload := bestEffortsResponse{Sports: make([]bestEffortsSport, 0, len(args.Sports)), Meta: bestEffortsMeta{ServerVersion: normalizeVersion(version), SportsRequested: args.Sports, Oldest: args.Oldest, Newest: args.Newest, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds, DistanceMeters: args.DistanceMeters, MissingBuckets: missing, IncludeFull: args.IncludeFull}}
		for _, sport := range args.Sports {
			row, err := bestEffortsForSport(ctx, client, sport, curveSpec, args, missing)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(fetchBestEffortsMessage, err)
			}
			payload.Sports = append(payload.Sports, row)
		}
		if len(missing) == 0 {
			payload.Meta.MissingBuckets = nil
		}
		return encodeShaped(payload, args.IncludeFull, []string{"sports"}, version, debugMetadata, getBestEffortsName, response.UnitSystemMetric, shapeCfg)
	}
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

func decodeDateRangeRequest(raw json.RawMessage) (dateRangeRequest, error) {
	var args dateRangeRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date must be YYYY-MM-DD")
	}
	if args.EndDate < args.StartDate {
		return args, errors.New("end_date must be on or after start_date")
	}
	return args, nil
}

func decodePowerCurvesRequest(raw json.RawMessage) (powerCurvesRequest, error) {
	var args powerCurvesRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	args.Sport = firstNonEmpty(strings.TrimSpace(args.Sport), defaultPowerCurveSport)
	if !validDate(args.Oldest) || !validDate(args.Newest) {
		return args, errors.New("oldest and newest must be YYYY-MM-DD")
	}
	if args.Newest < args.Oldest {
		return args, errors.New("newest must be on or after oldest")
	}
	args.DurationSeconds = normalizePositiveInts(args.DurationSeconds, defaultDurationBuckets)
	return args, nil
}

func decodeBestEffortsRequest(raw json.RawMessage) (bestEffortsRequest, error) {
	var args bestEffortsRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	if (args.Oldest == "") != (args.Newest == "") {
		return args, errors.New("oldest and newest must be supplied together or both omitted")
	}
	if args.Oldest != "" && (!validDate(args.Oldest) || !validDate(args.Newest) || args.Newest < args.Oldest) {
		return args, errors.New("oldest/newest must be paired YYYY-MM-DD values in order")
	}
	args.Sports = normalizeSports(args.Sports)
	args.DurationSeconds = normalizePositiveInts(args.DurationSeconds, defaultDurationBuckets)
	args.DistanceMeters = normalizePositiveInts(args.DistanceMeters, nil)
	return args, nil
}

func decodeStrict(raw json.RawMessage, out any) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return errors.New("arguments must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing JSON")
	}
	return nil
}

func validDate(value string) bool { _, err := time.Parse(time.DateOnly, value); return err == nil }

func normalizeSports(values []string) []string {
	if len(values) == 0 {
		return append([]string(nil), defaultBestEffortSports...)
	}
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		sport := strings.TrimSpace(value)
		if sport != "" && !seen[sport] {
			seen[sport] = true
			out = append(out, sport)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), defaultBestEffortSports...)
	}
	return out
}

func normalizePositiveInts(values []int, defaults []int) []int {
	if len(values) == 0 {
		return append([]int(nil), defaults...)
	}
	seen := map[int]bool{}
	out := []int{}
	for _, value := range values {
		if value > 0 && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Ints(out)
	return out
}

func rangeCurveSpec(oldest, newest string) string { return "r." + oldest + "." + newest }
func bestEffortsCurveSpec(args bestEffortsRequest) string {
	if args.Oldest == "" {
		return "all"
	}
	return rangeCurveSpec(args.Oldest, args.Newest)
}

func shapeFitnessRows(rows []intervals.SummaryWithCats, includeFull bool) []fitnessRow {
	out := make([]fitnessRow, 0, len(rows))
	for _, row := range rows {
		ctl, atl, tsb := roundPtr(row.Fitness), roundPtr(row.Fatigue), roundPtr(row.Form)
		shaped := fitnessRow{Date: row.Date, CTL: ctl, ATL: atl, TSB: tsb}
		if includeFull {
			shaped.Full = row.Raw
		}
		out = append(out, shaped)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out
}

func firstCurve(set intervals.DataCurveSet) intervals.DataCurve {
	if len(set.List) == 0 {
		return intervals.DataCurve{}
	}
	return set.List[0]
}

func bucketPowerCurve(curve intervals.DataCurve, buckets []int) ([]powerCurvePoint, []int) {
	points := []powerCurvePoint{}
	missing := []int{}
	for _, bucket := range buckets {
		value, idx, ok := valueAtBucket(curve.Secs, curve.Values, bucket)
		if !ok {
			missing = append(missing, bucket)
			continue
		}
		point := powerCurvePoint{DurationSeconds: bucket, Watts: roundPtr(value)}
		if idx < len(curve.ActivityID) {
			point.ActivityID = curve.ActivityID[idx]
		}
		points = append(points, point)
	}
	return points, missing
}

func bestEffortsForSport(ctx context.Context, client BestEffortsClient, sport string, curveSpec string, args bestEffortsRequest, missing map[string][]int) (bestEffortsSport, error) {
	out := bestEffortsSport{Sport: sport}
	power, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds})
	if err != nil {
		return out, err
	}
	hr, err := client.ListAthleteHRCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds})
	if err != nil {
		return out, err
	}
	distances := args.DistanceMeters
	if len(distances) == 0 {
		distances = defaultDistanceBucketsForSport(sport)
	}
	pace, err := client.ListAthletePaceCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DistanceMeters: distances})
	if err != nil {
		return out, err
	}
	out.Efforts = append(out.Efforts, effortRowsFromDurationCurve("power", firstCurve(power), args.DurationSeconds, missing, sport)...)
	out.Efforts = append(out.Efforts, effortRowsFromDurationCurve("heart_rate", firstCurve(hr), args.DurationSeconds, missing, sport)...)
	out.Efforts = append(out.Efforts, effortRowsFromDistanceCurve(firstCurve(pace), distances, missing, sport)...)
	if args.IncludeFull {
		out.Full = map[string]any{"power": power.Raw, "heart_rate": hr.Raw, "pace": pace.Raw}
	}
	return out, nil
}

func effortRowsFromDurationCurve(family string, curve intervals.DataCurve, buckets []int, missing map[string][]int, sport string) []bestEffortRow {
	rows := []bestEffortRow{}
	for _, bucket := range buckets {
		value, idx, ok := valueAtBucket(curve.Secs, curve.Values, bucket)
		if !ok {
			missing[sport+":"+family] = append(missing[sport+":"+family], bucket)
			continue
		}
		row := bestEffortRow{Family: family, DurationSeconds: bucket, ActivityID: activityIDAt(curve, idx)}
		rounded := roundPtr(value)
		if family == "power" {
			row.PowerWatts = rounded
		} else {
			row.HeartRateBPM = rounded
		}
		rows = append(rows, row)
	}
	return rows
}

func effortRowsFromDistanceCurve(curve intervals.DataCurve, buckets []int, missing map[string][]int, sport string) []bestEffortRow {
	rows := []bestEffortRow{}
	for _, bucket := range buckets {
		value, idx, ok := valueAtBucket(curve.Distance, curve.Values, bucket)
		if !ok {
			missing[sport+":pace"] = append(missing[sport+":pace"], bucket)
			continue
		}
		rows = append(rows, bestEffortRow{Family: "pace", DistanceMeters: bucket, PaceValue: roundPtr(value), ActivityID: activityIDAt(curve, idx)})
	}
	return rows
}

func valueAtBucket(xs []float64, values []float64, bucket int) (float64, int, bool) {
	for i, x := range xs {
		if int(math.Round(x)) == bucket && i < len(values) {
			return values[i], i, true
		}
	}
	return 0, 0, false
}

func activityIDAt(curve intervals.DataCurve, idx int) string {
	if idx >= 0 && idx < len(curve.ActivityID) {
		return curve.ActivityID[idx]
	}
	return ""
}

func defaultDistanceBucketsForSport(sport string) []int {
	if strings.Contains(strings.ToLower(sport), "swim") {
		return append([]int(nil), defaultSwimDistanceBuckets...)
	}
	return append([]int(nil), defaultRunDistanceBuckets...)
}

func shapeTrainingSummary(rows []intervals.SummaryWithCats, args dateRangeRequest, timezone string, unitSystem response.UnitSystem, version string) trainingSummaryResponse {
	payload := trainingSummaryResponse{Meta: trainingSummaryMeta{ServerVersion: normalizeVersion(version), StartDate: args.StartDate, EndDate: args.EndDate, Timezone: timezone, ZoneFamily: "upstream_timeInZones", ZoneOrder: "upstream", IncludeFull: args.IncludeFull}}
	categoryTotals := map[string]*trainingSportTotals{}
	var distanceMeters float64
	for _, row := range rows {
		payload.Summary.Count += row.Count
		payload.Summary.TimeSeconds += row.Time
		payload.Summary.MovingTimeSeconds += row.MovingTime
		payload.Summary.ElapsedTimeSeconds += row.ElapsedTime
		payload.Summary.CaloriesBurned += row.Calories
		payload.Summary.ElevationGainM += row.TotalElevationGain
		payload.Summary.TrainingLoad += row.TrainingLoad
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
			total.TrainingLoad += category.TrainingLoad
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

func toolProfile(ctx context.Context, profileClient ProfileClient, timezoneFallback string) (response.UnitSystem, string, error) {
	profile, err := profileClient.GetAthleteProfile(ctx)
	if err != nil {
		return "", "", err
	}
	unitSystem := profileUnitSystem(profile)
	timezone := profileTimezone(profile.Timezone, timezoneFallback)
	return unitSystem, timezone, nil
}

func encodeShaped(payload any, includeFull bool, rowCollections []string, version string, debugMetadata bool, queryType string, unitSystem response.UnitSystem, shaping ...responseShaping) (Result, error) {
	shapeCfg := responseShapingOrDefault(shaping)
	shaped, err := response.Shape(payload, shapeCfg.options(includeFull, rowCollections, version, debugMetadata, queryType, unitSystem))
	if err != nil {
		return Result{}, err
	}
	text, err := json.Marshal(shaped)
	if err != nil {
		return Result{}, fmt.Errorf("encoding %s response: %w", queryType, err)
	}
	return Result{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: shaped}, nil
}

func roundPtr(value float64) *float64 { rounded := round(value, 3); return &rounded }

func dateRangeInputSchema(startDescription string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{"start_date": map[string]any{"type": "string", "description": startDescription + " as YYYY-MM-DD in the athlete timezone."}, "end_date": map[string]any{"type": "string", "description": "local end date as YYYY-MM-DD in the athlete timezone."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream summary rows."}}}
}
func powerCurvesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Local start date YYYY-MM-DD."}, "newest": map[string]any{"type": "string", "description": "Local end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "default": defaultPowerCurveSport, "description": "Intervals.icu sport/type for the required upstream type parameter; defaults to Ride."}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Power curve buckets to return. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
func bestEffortsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Optional local start date YYYY-MM-DD. Supply with newest or omit both for all-time."}, "newest": map[string]any{"type": "string", "description": "Optional local end date YYYY-MM-DD. Supply with oldest or omit both for all-time."}, "sports": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Intervals.icu sports/types to fan out. Defaults to Ride, Run, Swim."}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Power/HR duration buckets. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "distance_meters": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Pace distance buckets. Defaults depend on sport: run-style 400,1000,1609,5000,10000m; swim 50,100,200,400,1500m."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
func genericOutputSchema(description string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": description}
}
