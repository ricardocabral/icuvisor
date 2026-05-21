package tools

import (
	"context"
	"encoding/json"
	"errors"
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
	values, missing := durationCurveBucketValues(curve, buckets)
	points := make([]powerCurvePoint, 0, len(values))
	for _, value := range values {
		points = append(points, powerCurvePoint{DurationSeconds: value.Bucket, Watts: value.Value, ActivityID: value.ActivityID})
	}
	return points, missing
}

func powerCurvesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Local start date YYYY-MM-DD."}, "newest": map[string]any{"type": "string", "description": "Local end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "default": defaultPowerCurveSport, "description": "Intervals.icu sport/type for the required upstream type parameter; defaults to Ride."}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Power curve buckets to return. Defaults to 5,15,30,60,300,1200,3600 seconds."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
