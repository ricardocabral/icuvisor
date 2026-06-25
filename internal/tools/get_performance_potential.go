package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getPerformancePotentialName        = "get_performance_potential"
	getPerformancePotentialDescription = "Summarize per-sport performance potential from explicit athlete profile thresholds plus upstream power, pace, and heart-rate curve anchors. Returns deterministic source fields and caveats; it does not invent proprietary scores or unavailable threshold estimates."
	invalidPerformancePotentialMessage = "invalid performance potential arguments; provide start_date/end_date as YYYY-MM-DD and optional sports or bucket arrays"
	fetchPerformanceProfileMessage     = "could not fetch athlete profile for performance potential; check intervals.icu credentials and athlete ID"
)

var defaultPerformancePotentialSports = []string{"Ride", "Run", "Swim"}

// PerformancePotentialClient retrieves curve sources needed for performance-potential summaries.
type PerformancePotentialClient interface {
	PowerCurvesClient
	PaceCurvesClient
	HRCurvesClient
}

type performancePotentialRequest struct {
	StartDate            string   `json:"start_date"`
	EndDate              string   `json:"end_date"`
	Sports               []string `json:"sports,omitempty"`
	PowerDurationSeconds []int    `json:"power_duration_seconds,omitempty"`
	HRDurationSeconds    []int    `json:"hr_duration_seconds,omitempty"`
	PaceDistanceMeters   []int    `json:"pace_distance_meters,omitempty"`
	IncludeFull          bool     `json:"include_full,omitempty"`
}

type performancePotentialResponse struct {
	Sports []performancePotentialSport `json:"sports"`
	Meta   performancePotentialMeta    `json:"_meta"`
}

type performancePotentialSport struct {
	Sport            string                                               `json:"sport"`
	SportFamily      string                                               `json:"sport_family"`
	Thresholds       performancePotentialThresholds                       `json:"thresholds"`
	CurveAnchors     performancePotentialCurveAnchors                     `json:"curve_anchors"`
	ThresholdContext performancePotentialThresholdContext                 `json:"threshold_context"`
	Unavailable      []analysis.PerformancePotentialUnavailable           `json:"unavailable,omitempty"`
	Caveats          []string                                             `json:"caveats,omitempty"`
	Full             map[string]any                                       `json:"full,omitempty"`
	Sources          map[string]analysis.PerformancePotentialSourceStatus `json:"sources,omitempty"`
}

type performancePotentialThresholds struct {
	FTPWatts                    *int                                     `json:"ftp_watts,omitempty"`
	IndoorFTPWatts              *int                                     `json:"indoor_ftp_watts,omitempty"`
	WPrimeJoules                *int                                     `json:"w_prime_joules,omitempty"`
	PMaxWatts                   *int                                     `json:"p_max_watts,omitempty"`
	LTHRBPM                     *int                                     `json:"lthr_bpm,omitempty"`
	MaxHRBPM                    *int                                     `json:"max_hr_bpm,omitempty"`
	ThresholdPaceSecondsPerKM   *float64                                 `json:"threshold_pace_seconds_per_km,omitempty"`
	ThresholdPaceSecondsPerMile *float64                                 `json:"threshold_pace_seconds_per_mile,omitempty"`
	ThresholdPaceSecondsPer100M *float64                                 `json:"threshold_pace_seconds_per_100m,omitempty"`
	ThresholdPaceSecondsPer500M *float64                                 `json:"threshold_pace_seconds_per_500m,omitempty"`
	ThresholdPaceValue          *float64                                 `json:"threshold_pace_value,omitempty"`
	PaceDistanceUnit            string                                   `json:"pace_distance_unit,omitempty"`
	PaceUnitsSource             string                                   `json:"pace_units_source,omitempty"`
	CriticalPower               analysis.PerformancePotentialUnavailable `json:"critical_power"`
}

type performancePotentialCurveAnchors struct {
	Power     performancePotentialPowerCurve `json:"power"`
	Pace      performancePotentialPaceCurve  `json:"pace"`
	HeartRate performancePotentialHRCurve    `json:"heart_rate"`
}

type performancePotentialPowerCurve struct {
	Status         string            `json:"status"`
	Reason         string            `json:"reason,omitempty"`
	Unit           string            `json:"unit,omitempty"`
	Points         []powerCurvePoint `json:"points,omitempty"`
	MissingBuckets []int             `json:"missing_buckets,omitempty"`
}

type performancePotentialPaceCurve struct {
	Status                string                          `json:"status"`
	Reason                string                          `json:"reason,omitempty"`
	PreferredUnit         string                          `json:"preferred_unit,omitempty"`
	ElapsedUnit           string                          `json:"elapsed_unit,omitempty"`
	Points                []performancePotentialPacePoint `json:"points,omitempty"`
	MissingDistanceMeters []int                           `json:"missing_distance_meters,omitempty"`
}

type performancePotentialHRCurve struct {
	Status         string         `json:"status"`
	Reason         string         `json:"reason,omitempty"`
	Unit           string         `json:"unit,omitempty"`
	Points         []hrCurvePoint `json:"points,omitempty"`
	MissingBuckets []int          `json:"missing_buckets,omitempty"`
}

type performancePotentialPacePoint struct {
	DistanceMeters     int      `json:"distance_meters"`
	ElapsedSeconds     *float64 `json:"elapsed_seconds,omitempty"`
	PaceSecondsPerKM   *float64 `json:"pace_seconds_per_km,omitempty"`
	PaceSecondsPerMile *float64 `json:"pace_seconds_per_mile,omitempty"`
	PaceSecondsPer100M *float64 `json:"pace_seconds_per_100m,omitempty"`
	ActivityID         string   `json:"activity_id,omitempty"`
}

type performancePotentialThresholdContext struct {
	AnaerobicThreshold []performancePotentialThresholdSource    `json:"anaerobic_threshold,omitempty"`
	AerobicThreshold   analysis.PerformancePotentialUnavailable `json:"aerobic_threshold"`
}

type performancePotentialThresholdSource struct {
	Field  string  `json:"field"`
	Value  float64 `json:"value"`
	Unit   string  `json:"unit"`
	Source string  `json:"source"`
}

type performancePotentialMeta struct {
	ServerVersion        string         `json:"server_version"`
	StartDate            string         `json:"start_date"`
	EndDate              string         `json:"end_date"`
	Sports               []string       `json:"sports"`
	PowerDurationSeconds []int          `json:"power_duration_seconds"`
	HRDurationSeconds    []int          `json:"hr_duration_seconds"`
	PaceDistanceMeters   []int          `json:"pace_distance_meters,omitempty"`
	CurveSpec            string         `json:"curve_spec"`
	SourceTools          []string       `json:"source_tools"`
	SourceEndpoints      []string       `json:"source_endpoints"`
	FormulaRefs          []string       `json:"formula_refs"`
	SourceUnits          map[string]any `json:"source_units"`
	Caveats              []string       `json:"caveats"`
	IncludeFull          bool           `json:"include_full"`
}

func newGetPerformancePotentialTool(client PerformancePotentialClient, profileClient ProfileClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getPerformancePotentialName, Description: getPerformancePotentialDescription, InputSchema: performancePotentialInputSchema(), OutputSchema: genericOutputSchema("Per-sport performance-potential thresholds, curve anchors, and explicit caveats."), Handler: getPerformancePotentialHandler(client, profileClient, version, debugMetadata, shapeCfg)})
}

func getPerformancePotentialHandler(client PerformancePotentialClient, profileClient ProfileClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodePerformancePotentialRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidPerformancePotentialMessage, err)
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchPerformanceProfileMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		curveSpec := rangeCurveSpec(args.StartDate, args.EndDate)
		profileResponse := athleteprofile.NewResponse(profile, version, "", false)
		payload := performancePotentialResponse{Sports: make([]performancePotentialSport, 0, len(args.Sports)), Meta: performancePotentialMeta{ServerVersion: normalizeVersion(version), StartDate: args.StartDate, EndDate: args.EndDate, Sports: append([]string(nil), args.Sports...), PowerDurationSeconds: append([]int(nil), args.PowerDurationSeconds...), HRDurationSeconds: append([]int(nil), args.HRDurationSeconds...), PaceDistanceMeters: append([]int(nil), args.PaceDistanceMeters...), CurveSpec: curveSpec, SourceTools: analysis.NormalizeSourceTools([]string{getAthleteProfileName, getPowerCurvesName, getPaceCurvesName, getHRCurvesName}), SourceEndpoints: []string{"/api/v1/athlete/{id}", "/api/v1/athlete/{id}/power-curves.json", "/api/v1/athlete/{id}/pace-curves.json", "/api/v1/athlete/{id}/hr-curves.json"}, FormulaRefs: []string{analysis.PerformancePotentialFormulaRef}, SourceUnits: performancePotentialSourceUnits(unitSystem), Caveats: []string{"This is a deterministic summary of explicit profile thresholds and upstream curve anchors, not a proprietary score or medical/lactate diagnosis.", "Aerobic threshold, critical power, and model-derived threshold estimates are unavailable unless an explicit supported upstream source exists."}, IncludeFull: args.IncludeFull}}
		for _, sport := range args.Sports {
			profileSport := matchPerformancePotentialProfileSport(profileResponse.SportSettings, sport)
			row, err := buildPerformancePotentialSport(ctx, client, sport, profileSport, unitSystem, curveSpec, args)
			if err != nil {
				return Result{}, err
			}
			payload.Sports = append(payload.Sports, row)
		}
		return encodeShaped(payload, args.IncludeFull, []string{"sports"}, version, debugMetadata, getPerformancePotentialName, unitSystem, shapeCfg)
	}
}

func decodePerformancePotentialRequest(raw json.RawMessage) (performancePotentialRequest, error) {
	var args performancePotentialRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[performancePotentialRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date must be YYYY-MM-DD")
	}
	if args.EndDate < args.StartDate {
		return args, errors.New("end_date must be on or after start_date")
	}
	args.Sports = normalizePerformancePotentialSports(args.Sports)
	args.PowerDurationSeconds = normalizePositiveInts(args.PowerDurationSeconds, defaultDurationBuckets)
	args.HRDurationSeconds = normalizePositiveInts(args.HRDurationSeconds, defaultDurationBuckets)
	args.PaceDistanceMeters = normalizePositiveInts(args.PaceDistanceMeters, nil)
	return args, nil
}

func buildPerformancePotentialSport(ctx context.Context, client PerformancePotentialClient, sport string, profileSport *athleteprofile.Sport, unitSystem response.UnitSystem, curveSpec string, args performancePotentialRequest) (performancePotentialSport, error) {
	family := analysis.PerformancePotentialSportFamily(sport)
	row := performancePotentialSport{Sport: sport, SportFamily: family, Sources: map[string]analysis.PerformancePotentialSourceStatus{"athlete_profile": {Status: "available", Unit: string(unitSystem)}}}
	row.Thresholds, row.ThresholdContext, row.Unavailable, row.Caveats = performancePotentialThresholdsForSport(sport, profileSport)
	power, err := performancePotentialPowerAnchors(ctx, client, sport, curveSpec, args.PowerDurationSeconds, args.IncludeFull, &row)
	if err != nil {
		return performancePotentialSport{}, err
	}
	row.CurveAnchors.Power = power
	paceBuckets := args.PaceDistanceMeters
	if len(paceBuckets) == 0 {
		paceBuckets = defaultDistanceBucketsForSport(sport)
	}
	pace, err := performancePotentialPaceAnchors(ctx, client, sport, curveSpec, paceBuckets, unitSystem, args.IncludeFull, &row)
	if err != nil {
		return performancePotentialSport{}, err
	}
	row.CurveAnchors.Pace = pace
	heartRate, err := performancePotentialHRAnchors(ctx, client, sport, curveSpec, args.HRDurationSeconds, args.IncludeFull, &row)
	if err != nil {
		return performancePotentialSport{}, err
	}
	row.CurveAnchors.HeartRate = heartRate
	return row, nil
}

func performancePotentialThresholdsForSport(sport string, profileSport *athleteprofile.Sport) (performancePotentialThresholds, performancePotentialThresholdContext, []analysis.PerformancePotentialUnavailable, []string) {
	thresholds := performancePotentialThresholds{CriticalPower: unavailablePerformanceField("critical_power", "unsupported", "critical power is not exposed as an explicit supported upstream field", "explicit critical_power field from athlete profile or upstream curve model")}
	context := performancePotentialThresholdContext{AerobicThreshold: unavailablePerformanceField("aerobic_threshold", "unsupported", "no public deterministic aerobic-threshold formula/source is implemented", "explicit upstream aerobic threshold or approved public formula")}
	unavailable := []analysis.PerformancePotentialUnavailable{thresholds.CriticalPower, context.AerobicThreshold}
	caveats := []string{}
	if profileSport == nil {
		missing := unavailablePerformanceField("sport_settings", "unavailable", "no matching athlete profile sport setting was found", "get_athlete_profile sport_settings")
		unavailable = append(unavailable, missing)
		caveats = append(caveats, "Profile thresholds are unavailable for this sport; curve anchors may still be shown when upstream curve endpoints return data.")
		return thresholds, context, unavailable, caveats
	}
	if profileSport.FTPWatts > 0 {
		thresholds.FTPWatts = intPointer(profileSport.FTPWatts)
		context.AnaerobicThreshold = append(context.AnaerobicThreshold, performancePotentialThresholdSource{Field: "ftp_watts", Value: float64(profileSport.FTPWatts), Unit: "W", Source: getAthleteProfileName})
	} else if analysis.PerformancePotentialSupportsPower(sport) {
		unavailable = append(unavailable, unavailablePerformanceField("ftp_watts", "unavailable", "profile FTP is missing for this power-based sport", "get_athlete_profile sport_settings.ftp_watts"))
	}
	if profileSport.IndoorFTPWatts > 0 {
		thresholds.IndoorFTPWatts = intPointer(profileSport.IndoorFTPWatts)
	}
	if profileSport.WPrimeJoules > 0 {
		thresholds.WPrimeJoules = intPointer(profileSport.WPrimeJoules)
	}
	if profileSport.PMaxWatts > 0 {
		thresholds.PMaxWatts = intPointer(profileSport.PMaxWatts)
	}
	if profileSport.LTHRBPM > 0 {
		thresholds.LTHRBPM = intPointer(profileSport.LTHRBPM)
		context.AnaerobicThreshold = append(context.AnaerobicThreshold, performancePotentialThresholdSource{Field: "lthr_bpm", Value: float64(profileSport.LTHRBPM), Unit: "bpm", Source: getAthleteProfileName})
	} else if analysis.PerformancePotentialSupportsHeartRate(sport) {
		unavailable = append(unavailable, unavailablePerformanceField("lthr_bpm", "unavailable", "profile LTHR is missing for this sport", "get_athlete_profile sport_settings.lthr_bpm"))
	}
	if profileSport.MaxHRBPM > 0 {
		thresholds.MaxHRBPM = intPointer(profileSport.MaxHRBPM)
	}
	assignPerformancePotentialPaceThresholds(&thresholds, &context, profileSport)
	if analysis.PerformancePotentialSupportsPace(sport) && !hasPerformancePotentialPaceThreshold(thresholds) {
		unavailable = append(unavailable, unavailablePerformanceField("threshold_pace", "unavailable", "profile threshold pace is missing for this pace-based sport", "get_athlete_profile sport_settings.threshold_pace"))
	}
	if len(context.AnaerobicThreshold) == 0 {
		caveats = append(caveats, "No explicit anaerobic-threshold proxy is available for this sport; no estimate was generated.")
	}
	return thresholds, context, unavailable, caveats
}

func assignPerformancePotentialPaceThresholds(thresholds *performancePotentialThresholds, context *performancePotentialThresholdContext, profileSport *athleteprofile.Sport) {
	thresholds.PaceDistanceUnit = profileSport.PaceDistanceUnit
	thresholds.PaceUnitsSource = profileSport.PaceUnitsSource
	add := func(field string, value *float64, unit string) {
		if value == nil || *value <= 0 {
			return
		}
		context.AnaerobicThreshold = append(context.AnaerobicThreshold, performancePotentialThresholdSource{Field: field, Value: *value, Unit: unit, Source: getAthleteProfileName})
	}
	thresholds.ThresholdPaceSecondsPerKM = profileSport.ThresholdPaceSecondsPerKM
	add("threshold_pace_seconds_per_km", profileSport.ThresholdPaceSecondsPerKM, "s/km")
	thresholds.ThresholdPaceSecondsPerMile = profileSport.ThresholdPaceSecondsPerMile
	add("threshold_pace_seconds_per_mile", profileSport.ThresholdPaceSecondsPerMile, "s/mile")
	thresholds.ThresholdPaceSecondsPer100M = profileSport.ThresholdPaceSecondsPer100M
	add("threshold_pace_seconds_per_100m", profileSport.ThresholdPaceSecondsPer100M, "s/100m")
	thresholds.ThresholdPaceSecondsPer500M = profileSport.ThresholdPaceSecondsPer500M
	add("threshold_pace_seconds_per_500m", profileSport.ThresholdPaceSecondsPer500M, "s/500m")
	thresholds.ThresholdPaceValue = profileSport.ThresholdPaceValue
	add("threshold_pace_value", profileSport.ThresholdPaceValue, "source_unit")
}

func performancePotentialPowerAnchors(ctx context.Context, client PerformancePotentialClient, sport string, curveSpec string, buckets []int, includeFull bool, row *performancePotentialSport) (performancePotentialPowerCurve, error) {
	if !analysis.PerformancePotentialSupportsPower(sport) {
		row.Sources["power_curves"] = analysis.PerformancePotentialSourceStatus{Status: "unsupported", Reason: "sport family does not use power-duration anchors by default"}
		return performancePotentialPowerCurve{Status: "unsupported", Reason: "sport family does not use power-duration anchors by default"}, nil
	}
	set, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: powerCurveRequestSpec(curveSpec), DurationSeconds: buckets})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return performancePotentialPowerCurve{}, ctxErr
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return performancePotentialPowerCurve{}, err
		}
		row.Sources["power_curves"] = analysis.PerformancePotentialSourceStatus{Status: "unavailable", Reason: "source_fetch_failed", Unit: "W"}
		row.Unavailable = append(row.Unavailable, unavailablePerformanceField("curve_anchors.power", "unavailable", "power curve source could not be fetched", getPowerCurvesName))
		return performancePotentialPowerCurve{Status: "unavailable", Reason: "source_fetch_failed", Unit: "W"}, nil
	}
	points, missing := bucketPowerCurve(standardPowerCurve(set), buckets)
	points, nonPositiveMissing := dropNonPositivePowerCurvePoints(points)
	missing = append(missing, nonPositiveMissing...)
	sort.Ints(missing)
	status := "available"
	reason := ""
	if len(points) == 0 {
		status = "unavailable"
		reason = "no matching positive power buckets returned"
		row.Unavailable = append(row.Unavailable, unavailablePerformanceField("curve_anchors.power", "unavailable", reason, getPowerCurvesName))
	}
	row.Sources["power_curves"] = analysis.PerformancePotentialSourceStatus{Status: status, Reason: reason, Unit: "W"}
	if includeFull {
		ensurePerformancePotentialFull(row)["power_curves"] = set.Raw
	}
	return performancePotentialPowerCurve{Status: status, Reason: reason, Unit: "W", Points: points, MissingBuckets: missing}, nil
}

func performancePotentialPaceAnchors(ctx context.Context, client PerformancePotentialClient, sport string, curveSpec string, buckets []int, unitSystem response.UnitSystem, includeFull bool, row *performancePotentialSport) (performancePotentialPaceCurve, error) {
	preferredUnit := performancePotentialPaceUnit(sport, unitSystem)
	if !analysis.PerformancePotentialSupportsPace(sport) {
		row.Sources["pace_curves"] = analysis.PerformancePotentialSourceStatus{Status: "unsupported", Reason: "sport family does not use pace-distance anchors by default", Unit: preferredUnit}
		return performancePotentialPaceCurve{Status: "unsupported", Reason: "sport family does not use pace-distance anchors by default", PreferredUnit: preferredUnit, ElapsedUnit: "seconds"}, nil
	}
	set, err := client.ListAthletePaceCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DistanceMeters: buckets})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return performancePotentialPaceCurve{}, ctxErr
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return performancePotentialPaceCurve{}, err
		}
		row.Sources["pace_curves"] = analysis.PerformancePotentialSourceStatus{Status: "unavailable", Reason: "source_fetch_failed", Unit: preferredUnit}
		row.Unavailable = append(row.Unavailable, unavailablePerformanceField("curve_anchors.pace", "unavailable", "pace curve source could not be fetched", getPaceCurvesName))
		return performancePotentialPaceCurve{Status: "unavailable", Reason: "source_fetch_failed", PreferredUnit: preferredUnit, ElapsedUnit: "seconds"}, nil
	}
	points, missing := bucketPerformancePotentialPaceCurve(firstCurve(set), buckets, unitSystem, sport)
	status := "available"
	reason := ""
	if len(points) == 0 {
		status = "unavailable"
		reason = "no matching positive pace buckets returned"
		row.Unavailable = append(row.Unavailable, unavailablePerformanceField("curve_anchors.pace", "unavailable", reason, getPaceCurvesName))
	}
	row.Sources["pace_curves"] = analysis.PerformancePotentialSourceStatus{Status: status, Reason: reason, Unit: preferredUnit}
	if includeFull {
		ensurePerformancePotentialFull(row)["pace_curves"] = set.Raw
	}
	return performancePotentialPaceCurve{Status: status, Reason: reason, PreferredUnit: preferredUnit, ElapsedUnit: "seconds", Points: points, MissingDistanceMeters: missing}, nil
}

func performancePotentialHRAnchors(ctx context.Context, client PerformancePotentialClient, sport string, curveSpec string, buckets []int, includeFull bool, row *performancePotentialSport) (performancePotentialHRCurve, error) {
	if !analysis.PerformancePotentialSupportsHeartRate(sport) {
		row.Sources["hr_curves"] = analysis.PerformancePotentialSourceStatus{Status: "unsupported", Reason: "sport family does not use heart-rate anchors by default"}
		return performancePotentialHRCurve{Status: "unsupported", Reason: "sport family does not use heart-rate anchors by default"}, nil
	}
	set, err := client.ListAthleteHRCurves(ctx, intervals.CurveParams{Sport: sport, CurveSpec: curveSpec, DurationSeconds: buckets})
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return performancePotentialHRCurve{}, ctxErr
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return performancePotentialHRCurve{}, err
		}
		row.Sources["hr_curves"] = analysis.PerformancePotentialSourceStatus{Status: "unavailable", Reason: "source_fetch_failed", Unit: "bpm"}
		row.Unavailable = append(row.Unavailable, unavailablePerformanceField("curve_anchors.heart_rate", "unavailable", "heart-rate curve source could not be fetched", getHRCurvesName))
		return performancePotentialHRCurve{Status: "unavailable", Reason: "source_fetch_failed", Unit: "bpm"}, nil
	}
	points, missing := bucketHRCurve(firstCurve(set), buckets)
	points, nonPositiveMissing := dropNonPositiveHRPoints(points)
	missing = append(missing, nonPositiveMissing...)
	sort.Ints(missing)
	status := "available"
	reason := ""
	if len(points) == 0 {
		status = "unavailable"
		reason = "no matching positive heart-rate buckets returned"
		row.Unavailable = append(row.Unavailable, unavailablePerformanceField("curve_anchors.heart_rate", "unavailable", reason, getHRCurvesName))
	}
	row.Sources["hr_curves"] = analysis.PerformancePotentialSourceStatus{Status: status, Reason: reason, Unit: "bpm"}
	if includeFull {
		ensurePerformancePotentialFull(row)["hr_curves"] = set.Raw
	}
	return performancePotentialHRCurve{Status: status, Reason: reason, Unit: "bpm", Points: points, MissingBuckets: missing}, nil
}

func bucketPerformancePotentialPaceCurve(curve intervals.DataCurve, buckets []int, unitSystem response.UnitSystem, sport string) ([]performancePotentialPacePoint, []int) {
	values, missing := distanceCurveBucketValues(curve, buckets)
	points := make([]performancePotentialPacePoint, 0, len(values))
	for _, value := range values {
		if value.Value == nil || *value.Value <= 0 || value.Bucket <= 0 {
			missing = append(missing, value.Bucket)
			continue
		}
		points = append(points, performancePotentialPacePointFromValue(value, unitSystem, sport))
	}
	sort.Ints(missing)
	return points, missing
}

func performancePotentialPacePointFromValue(value curveBucketValue, unitSystem response.UnitSystem, sport string) performancePotentialPacePoint {
	point := performancePotentialPacePoint{DistanceMeters: value.Bucket, ElapsedSeconds: value.Value, ActivityID: value.ActivityID}
	if value.Value == nil || value.Bucket <= 0 {
		return point
	}
	if analysis.PerformancePotentialSportFamily(sport) == "swimming" {
		pace := round(*value.Value/(float64(value.Bucket)/100.0), 1)
		point.PaceSecondsPer100M = &pace
		return point
	}
	if unitSystem == response.UnitSystemImperial {
		pace := round(*value.Value/(float64(value.Bucket)/metersPerMileForPaceCurve), 1)
		point.PaceSecondsPerMile = &pace
		return point
	}
	pace := round(*value.Value/(float64(value.Bucket)/metersPerKilometerForPaceCurve), 1)
	point.PaceSecondsPerKM = &pace
	return point
}

func dropNonPositiveHRPoints(points []hrCurvePoint) ([]hrCurvePoint, []int) {
	kept := make([]hrCurvePoint, 0, len(points))
	missing := []int{}
	for _, point := range points {
		if point.HeartRateBPM == nil || *point.HeartRateBPM <= 0 {
			missing = append(missing, point.DurationSeconds)
			continue
		}
		kept = append(kept, point)
	}
	return kept, missing
}

func normalizePerformancePotentialSports(values []string) []string {
	if len(values) == 0 {
		return append([]string(nil), defaultPerformancePotentialSports...)
	}
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		key := strings.ToLower(trimmed)
		if trimmed == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, trimmed)
	}
	if len(out) == 0 {
		return append([]string(nil), defaultPerformancePotentialSports...)
	}
	return out
}

func matchPerformancePotentialProfileSport(sports []athleteprofile.Sport, requested string) *athleteprofile.Sport {
	requestedKey := normalizePerformancePotentialSportName(requested)
	for i := range sports {
		for _, sportType := range sports[i].Types {
			if normalizePerformancePotentialSportName(sportType) == requestedKey {
				return &sports[i]
			}
		}
	}
	return nil
}

func normalizePerformancePotentialSportName(value string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(value)))
}

func unavailablePerformanceField(field string, status string, reason string, requiredSource string) analysis.PerformancePotentialUnavailable {
	return analysis.PerformancePotentialUnavailable{Field: field, Status: status, Reason: reason, RequiredSource: requiredSource}
}

func ensurePerformancePotentialFull(row *performancePotentialSport) map[string]any {
	if row.Full == nil {
		row.Full = map[string]any{}
	}
	return row.Full
}

func hasPerformancePotentialPaceThreshold(thresholds performancePotentialThresholds) bool {
	return thresholds.ThresholdPaceSecondsPerKM != nil || thresholds.ThresholdPaceSecondsPerMile != nil || thresholds.ThresholdPaceSecondsPer100M != nil || thresholds.ThresholdPaceSecondsPer500M != nil || thresholds.ThresholdPaceValue != nil
}

func performancePotentialPaceUnit(sport string, unitSystem response.UnitSystem) string {
	if analysis.PerformancePotentialSportFamily(sport) == "swimming" {
		return "seconds_per_100m"
	}
	if unitSystem == response.UnitSystemImperial {
		return "seconds_per_mile"
	}
	return "seconds_per_km"
}

func performancePotentialSourceUnits(unitSystem response.UnitSystem) map[string]any {
	pace := "seconds_per_km"
	if unitSystem == response.UnitSystemImperial {
		pace = "seconds_per_mile"
	}
	return map[string]any{"power": "watts", "heart_rate": "bpm", "pace_default": pace, "swim_pace": "seconds_per_100m", "source_threshold_units": "profile pace_units_source is preserved per sport when available"}
}

func intPointer(value int) *int { return &value }

func performancePotentialInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{"start_date": map[string]any{"type": "string", "description": "Local start date YYYY-MM-DD for upstream curve anchors."}, "end_date": map[string]any{"type": "string", "description": "Local end date YYYY-MM-DD for upstream curve anchors."}, "sports": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Intervals.icu sport/type labels to summarize. Defaults to Ride, Run, Swim. Each sport remains in its own unit/source namespace."}, "power_duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Power-duration anchor buckets for power-based sports. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "hr_duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Heart-rate anchor buckets. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "pace_distance_meters": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Pace-distance anchor buckets for pace sports. Defaults depend on sport: run-style 400,1000,1609,5000,10000m; swim 50,100,200,400,1500m."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve payloads under each sport.full; terse default returns only selected anchors, statuses, caveats, and metadata."}}}
}
