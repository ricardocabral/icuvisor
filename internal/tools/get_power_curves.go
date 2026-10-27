package tools

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getPowerCurvesName        = "get_power_curves"
	getPowerCurvesDescription = "Get the upstream-computed mean-maximal power curve for a date range. Default terse output returns only common duration buckets; include_full returns raw arrays."
	fetchPowerCurvesMessage   = "could not fetch power curves; check intervals.icu credentials, athlete ID, sport, and date range"
	defaultPowerCurveSport    = "Ride"
)

var powerDurabilityCurveSuffixes = []string{"-kj0", "-kj1"}

// PowerCurvesClient retrieves athlete curve sets.
type PowerCurvesClient interface {
	ListAthletePowerCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
}

type powerCurvesRequest struct {
	Oldest          string `json:"oldest"`
	Newest          string `json:"newest"`
	Sport           string `json:"sport,omitempty"`
	DurationSeconds []int  `json:"duration_seconds,omitempty"`
	IncludeFull     bool   `json:"include_full,omitempty"`
}

type powerCurvesResponse struct {
	Sport            string                 `json:"sport"`
	Points           []powerCurvePoint      `json:"points"`
	PowerModels      []powerModelSummary    `json:"power_models,omitempty"`
	DurabilityCurves []powerDurabilityCurve `json:"durability_curves,omitempty"`
	Full             map[string]any         `json:"full,omitempty"`
	Meta             powerCurvesMeta        `json:"_meta"`
}

type powerCurvePoint struct {
	DurationSeconds int      `json:"duration_seconds"`
	Watts           *float64 `json:"watts,omitempty"`
	ActivityID      string   `json:"activity_id,omitempty"`
	EffortDate      string   `json:"effort_date,omitempty"`
	WattsPerKG      *float64 `json:"watts_per_kg,omitempty"`
	WKGActivityID   string   `json:"wkg_activity_id,omitempty"`
	WKGEffortDate   string   `json:"wkg_effort_date,omitempty"`
}

type powerModelSummary struct {
	Type               string   `json:"type"`
	Source             string   `json:"source"`
	CriticalPowerWatts *float64 `json:"critical_power_watts,omitempty"`
	WPrimeJoules       *float64 `json:"w_prime_joules,omitempty"`
	PMaxWatts          *float64 `json:"p_max_watts,omitempty"`
	FTPWatts           *float64 `json:"ftp_watts,omitempty"`
}

type powerDurabilityCurve struct {
	AfterKJ        int               `json:"after_kj"`
	WorkThreshold  workThreshold     `json:"work_threshold"`
	Points         []powerCurvePoint `json:"points"`
	MissingBuckets []int             `json:"missing_buckets,omitempty"`
}

type workThreshold struct {
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}

type powerCurvesMeta struct {
	ServerVersion        string `json:"server_version"`
	Sport                string `json:"sport"`
	Oldest               string `json:"oldest"`
	Newest               string `json:"newest"`
	CurveSpec            string `json:"curve_spec"`
	DurationSeconds      []int  `json:"duration_seconds"`
	MissingBuckets       []int  `json:"missing_buckets,omitempty"`
	DurabilityCurveCount int    `json:"durability_curve_count,omitempty"`
	WorkThresholdUnit    string `json:"work_threshold_unit,omitempty"`
	IncludeFull          bool   `json:"include_full"`
}

func newGetPowerCurvesTool(client PowerCurvesClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getPowerCurvesName, Description: getPowerCurvesDescription, InputSchema: powerCurvesInputSchema(), OutputSchema: genericOutputSchema("Mean-maximal power curve bucket points."), Handler: getPowerCurvesHandler(client, version, debugMetadata, shapeCfg)})
}

func getPowerCurvesHandler(client PowerCurvesClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodePowerCurvesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCurveArgumentsMessage, err)
		}
		curveSpec := rangeCurveSpec(args.Oldest, args.Newest)
		requestCurveSpec := powerCurveRequestSpec(curveSpec)
		set, err := client.ListAthletePowerCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: requestCurveSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchPowerCurvesMessage, err)
		}
		curve := standardPowerCurve(set)
		points, missing := bucketPowerCurve(curve, args.DurationSeconds)
		addPowerCurveDates(points, set.Activities)
		durabilityCurves := bucketDurabilityPowerCurves(set, args.DurationSeconds)
		payload := powerCurvesResponse{Sport: args.Sport, Points: points, PowerModels: summarizePowerModels(curve.PowerModels), DurabilityCurves: durabilityCurves, Meta: powerCurvesMeta{ServerVersion: normalizeVersion(version), Sport: args.Sport, Oldest: args.Oldest, Newest: args.Newest, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds, MissingBuckets: missing, DurabilityCurveCount: len(durabilityCurves), WorkThresholdUnit: workThresholdUnit(durabilityCurves), IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = set.Raw
		}
		return encodeShaped(payload, args.IncludeFull, []string{"points"}, version, debugMetadata, getPowerCurvesName, response.UnitSystemMetric, shapeCfg)
	}
}

func decodePowerCurvesRequest(raw json.RawMessage) (powerCurvesRequest, error) {
	var args powerCurvesRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[powerCurvesRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
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

func bucketPowerCurve(curve intervals.DataCurve, buckets []int) ([]powerCurvePoint, []int) {
	points := []powerCurvePoint{}
	missing := []int{}
	for _, bucket := range buckets {
		_, idx, found := valueAtBucket(curve.Secs, curve.Secs, bucket)
		if !found {
			missing = append(missing, bucket)
			continue
		}
		point := powerCurvePoint{DurationSeconds: bucket}
		if idx < len(curve.Values) {
			point.Watts = roundPtr(curve.Values[idx])
			point.ActivityID = activityIDAt(curve, idx)
		}
		if idx < len(curve.WattsPerKG) && curve.WattsPerKG[idx] != nil && *curve.WattsPerKG[idx] > 0 {
			point.WattsPerKG = roundPtr(*curve.WattsPerKG[idx])
			if idx < len(curve.WKGActivityID) {
				point.WKGActivityID = curve.WKGActivityID[idx]
			}
		}
		if point.Watts == nil && point.WattsPerKG == nil {
			missing = append(missing, bucket)
			continue
		}
		points = append(points, point)
	}
	return points, missing
}

func addPowerCurveDates(points []powerCurvePoint, activities map[string]any) {
	for i := range points {
		points[i].EffortDate = curveActivityDate(activities, points[i].ActivityID)
		points[i].WKGEffortDate = curveActivityDate(activities, points[i].WKGActivityID)
	}
}

func curveActivityDate(activities map[string]any, id string) string {
	activity, ok := activities[id].(map[string]any)
	if !ok {
		return ""
	}
	date, ok := activity["start_date_local"].(string)
	if !ok || len(date) < 10 || !validDate(date[:10]) {
		return ""
	}
	return date[:10]
}

func summarizePowerModels(models []intervals.PowerModel) []powerModelSummary {
	out := []powerModelSummary{}
	for _, model := range models {
		switch model.Type {
		case "MS_2P", "MORTON_3P", "FFT_CURVES", "ECP":
		default:
			continue
		}
		summary := powerModelSummary{Type: model.Type, Source: "upstream_fitted", CriticalPowerWatts: availableModelValue(model.CriticalPower), WPrimeJoules: availableModelValue(model.WPrime), PMaxWatts: availableModelValue(model.PMax), FTPWatts: availableModelValue(model.FTP)}
		if summary.CriticalPowerWatts == nil && summary.WPrimeJoules == nil && summary.PMaxWatts == nil && summary.FTPWatts == nil {
			continue
		}
		out = append(out, summary)
		if len(out) == 4 {
			break
		}
	}
	return out
}

func availableModelValue(value *float64) *float64 {
	if value == nil || *value <= 0 || math.IsNaN(*value) || math.IsInf(*value, 0) {
		return nil
	}
	return value
}

func powerCurveRequestSpec(base string) string {
	specs := make([]string, 0, len(powerDurabilityCurveSuffixes)+1)
	specs = append(specs, base)
	for _, suffix := range powerDurabilityCurveSuffixes {
		specs = append(specs, base+suffix)
	}
	return strings.Join(specs, ",")
}

func standardPowerCurve(set intervals.DataCurveSet) intervals.DataCurve {
	for _, curve := range set.List {
		if curve.AfterKJ == nil || *curve.AfterKJ <= 0 {
			return curve
		}
	}
	return intervals.DataCurve{}
}

func bucketDurabilityPowerCurves(set intervals.DataCurveSet, buckets []int) []powerDurabilityCurve {
	curves := make([]powerDurabilityCurve, 0)
	for _, curve := range set.List {
		if curve.AfterKJ == nil || *curve.AfterKJ <= 0 {
			continue
		}
		points, missing := bucketPowerCurve(curve, buckets)
		addPowerCurveDates(points, set.Activities)
		points, nonPositiveMissing := dropNonPositivePowerCurvePoints(points)
		missing = append(missing, nonPositiveMissing...)
		sort.Ints(missing)
		if len(points) == 0 {
			continue
		}
		curves = append(curves, powerDurabilityCurve{AfterKJ: *curve.AfterKJ, WorkThreshold: workThreshold{Value: *curve.AfterKJ, Unit: "kJ"}, Points: points, MissingBuckets: missing})
	}
	return curves
}

func dropNonPositivePowerCurvePoints(points []powerCurvePoint) ([]powerCurvePoint, []int) {
	kept := make([]powerCurvePoint, 0, len(points))
	missing := []int{}
	for _, point := range points {
		if point.Watts == nil || *point.Watts <= 0 {
			missing = append(missing, point.DurationSeconds)
			continue
		}
		kept = append(kept, point)
	}
	return kept, missing
}

func workThresholdUnit(curves []powerDurabilityCurve) string {
	if len(curves) == 0 {
		return ""
	}
	return "kJ"
}

func powerCurvesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Local start date YYYY-MM-DD."}, "newest": map[string]any{"type": "string", "description": "Local end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "default": defaultPowerCurveSport, "description": "Intervals.icu sport/type for the required upstream type parameter; defaults to Ride."}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Power curve buckets to return. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
