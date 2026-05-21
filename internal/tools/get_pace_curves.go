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
	getPaceCurvesName              = "get_pace_curves"
	getPaceCurvesDescription       = "Get upstream-computed pace best-effort curves for a date range. Default terse output returns common distance buckets with elapsed seconds and athlete-preferred pace units; include_full returns raw arrays."
	fetchPaceCurvesMessage         = "could not fetch pace curves; check intervals.icu credentials, athlete ID, sport, and date range"
	fetchPaceProfileMessage        = "could not fetch athlete profile for pace units; check intervals.icu credentials and athlete ID"
	metersPerKilometerForPaceCurve = 1000.0
	metersPerMileForPaceCurve      = 1609.344
)

// PaceCurvesClient retrieves athlete pace curve sets.
type PaceCurvesClient interface {
	ListAthletePaceCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
}

type paceCurvesRequest struct {
	Oldest         string `json:"oldest"`
	Newest         string `json:"newest"`
	Sport          string `json:"sport,omitempty"`
	DistanceMeters []int  `json:"distance_meters,omitempty"`
	IncludeFull    bool   `json:"include_full,omitempty"`
}

type paceCurvesResponse struct {
	Sport  string           `json:"sport,omitempty"`
	Points []paceCurvePoint `json:"points"`
	Full   map[string]any   `json:"full,omitempty"`
	Meta   paceCurvesMeta   `json:"_meta"`
}

type paceCurvePoint struct {
	DistanceMeters     int      `json:"distance_meters"`
	ElapsedSeconds     *float64 `json:"elapsed_seconds,omitempty"`
	PaceSecondsPerKM   *float64 `json:"pace_seconds_per_km,omitempty"`
	PaceSecondsPerMile *float64 `json:"pace_seconds_per_mile,omitempty"`
	ActivityID         string   `json:"activity_id,omitempty"`
}

type paceCurvesMeta struct {
	ServerVersion  string `json:"server_version"`
	Sport          string `json:"sport,omitempty"`
	Oldest         string `json:"oldest"`
	Newest         string `json:"newest"`
	CurveSpec      string `json:"curve_spec"`
	DistanceMeters []int  `json:"distance_meters"`
	MissingBuckets []int  `json:"missing_buckets,omitempty"`
	IncludeFull    bool   `json:"include_full"`
}

func newGetPaceCurvesTool(client PaceCurvesClient, profileClient ProfileClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getPaceCurvesName, Description: getPaceCurvesDescription, InputSchema: paceCurvesInputSchema(), OutputSchema: genericOutputSchema("Pace curve distance bucket points with elapsed seconds and preferred pace fields."), Handler: getPaceCurvesHandler(client, profileClient, version, debugMetadata, shapeCfg)})
}

func getPaceCurvesHandler(client PaceCurvesClient, profileClient ProfileClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodePaceCurvesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCurveArgumentsMessage, err)
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchPaceProfileMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		curveSpec := rangeCurveSpec(args.Oldest, args.Newest)
		set, err := client.ListAthletePaceCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: curveSpec, DistanceMeters: args.DistanceMeters})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchPaceCurvesMessage, err)
		}
		points, missing := bucketPaceCurve(firstCurve(set), args.DistanceMeters, unitSystem)
		payload := paceCurvesResponse{Sport: args.Sport, Points: points, Meta: paceCurvesMeta{ServerVersion: normalizeVersion(version), Sport: args.Sport, Oldest: args.Oldest, Newest: args.Newest, CurveSpec: curveSpec, DistanceMeters: args.DistanceMeters, MissingBuckets: missing, IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = set.Raw
		}
		return encodeShaped(payload, args.IncludeFull, []string{"points"}, version, debugMetadata, getPaceCurvesName, unitSystem, shapeCfg)
	}
}

func decodePaceCurvesRequest(raw json.RawMessage) (paceCurvesRequest, error) {
	var args paceCurvesRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[paceCurvesRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	args.Sport = strings.TrimSpace(args.Sport)
	if !validDate(args.Oldest) || !validDate(args.Newest) {
		return args, errors.New("oldest and newest must be YYYY-MM-DD")
	}
	if args.Newest < args.Oldest {
		return args, errors.New("newest must be on or after oldest")
	}
	args.DistanceMeters = normalizePositiveInts(args.DistanceMeters, defaultDistanceBucketsForSport(args.Sport))
	return args, nil
}

func bucketPaceCurve(curve intervals.DataCurve, buckets []int, unitSystem response.UnitSystem) ([]paceCurvePoint, []int) {
	values, missing := distanceCurveBucketValues(curve, buckets)
	points := make([]paceCurvePoint, 0, len(values))
	for _, value := range values {
		points = append(points, paceCurvePointFromValue(value, unitSystem))
	}
	return points, missing
}

func paceCurvePointFromValue(value curveBucketValue, unitSystem response.UnitSystem) paceCurvePoint {
	point := paceCurvePoint{DistanceMeters: value.Bucket, ElapsedSeconds: value.Value, ActivityID: value.ActivityID}
	if value.Value == nil || value.Bucket <= 0 {
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

func paceCurvesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Local start date YYYY-MM-DD."}, "newest": map[string]any{"type": "string", "description": "Local end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "description": "Optional intervals.icu sport/type filter for the upstream type parameter. Omit to let upstream aggregate its default pace curve set."}, "distance_meters": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Pace curve distance buckets in meters. Defaults to 400,1000,1609,5000,10000m, or swim-style 50,100,200,400,1500m when sport contains Swim. Points include upstream elapsed_seconds plus either pace_seconds_per_km or pace_seconds_per_mile based on athlete units."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
