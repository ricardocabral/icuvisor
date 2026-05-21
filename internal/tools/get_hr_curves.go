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
	getHRCurvesName        = "get_hr_curves"
	getHRCurvesDescription = "Get upstream-computed heart-rate best-effort curves for a date range. Default terse output returns common duration buckets in beats per minute (BPM); include_full returns raw arrays."
	fetchHRCurvesMessage   = "could not fetch heart-rate curves; check intervals.icu credentials, athlete ID, sport, and date range"
)

// HRCurvesClient retrieves athlete heart-rate curve sets.
type HRCurvesClient interface {
	ListAthleteHRCurves(context.Context, intervals.CurveParams) (intervals.DataCurveSet, error)
}

type hrCurvesRequest struct {
	Oldest          string `json:"oldest"`
	Newest          string `json:"newest"`
	Sport           string `json:"sport,omitempty"`
	DurationSeconds []int  `json:"duration_seconds,omitempty"`
	IncludeFull     bool   `json:"include_full,omitempty"`
}

type hrCurvesResponse struct {
	Sport  string         `json:"sport,omitempty"`
	Points []hrCurvePoint `json:"points"`
	Full   map[string]any `json:"full,omitempty"`
	Meta   hrCurvesMeta   `json:"_meta"`
}

type hrCurvePoint struct {
	DurationSeconds int      `json:"duration_seconds"`
	HeartRateBPM    *float64 `json:"heart_rate_bpm,omitempty"`
	ActivityID      string   `json:"activity_id,omitempty"`
}

type hrCurvesMeta struct {
	ServerVersion   string `json:"server_version"`
	Sport           string `json:"sport,omitempty"`
	Oldest          string `json:"oldest"`
	Newest          string `json:"newest"`
	CurveSpec       string `json:"curve_spec"`
	DurationSeconds []int  `json:"duration_seconds"`
	MissingBuckets  []int  `json:"missing_buckets,omitempty"`
	IncludeFull     bool   `json:"include_full"`
}

func newGetHRCurvesTool(client HRCurvesClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getHRCurvesName, Description: getHRCurvesDescription, InputSchema: hrCurvesInputSchema(), OutputSchema: genericOutputSchema("Heart-rate curve bucket points in beats per minute."), Handler: getHRCurvesHandler(client, version, debugMetadata, shapeCfg)})
}

func getHRCurvesHandler(client HRCurvesClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeHRCurvesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCurveArgumentsMessage, err)
		}
		curveSpec := rangeCurveSpec(args.Oldest, args.Newest)
		set, err := client.ListAthleteHRCurves(ctx, intervals.CurveParams{Sport: args.Sport, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchHRCurvesMessage, err)
		}
		points, missing := bucketHRCurve(firstCurve(set), args.DurationSeconds)
		payload := hrCurvesResponse{Sport: args.Sport, Points: points, Meta: hrCurvesMeta{ServerVersion: normalizeVersion(version), Sport: args.Sport, Oldest: args.Oldest, Newest: args.Newest, CurveSpec: curveSpec, DurationSeconds: args.DurationSeconds, MissingBuckets: missing, IncludeFull: args.IncludeFull}}
		if args.IncludeFull {
			payload.Full = set.Raw
		}
		return encodeShaped(payload, args.IncludeFull, []string{"points"}, version, debugMetadata, getHRCurvesName, response.UnitSystemMetric, shapeCfg)
	}
}

func decodeHRCurvesRequest(raw json.RawMessage) (hrCurvesRequest, error) {
	var args hrCurvesRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[hrCurvesRequest](raw)
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
	args.DurationSeconds = normalizePositiveInts(args.DurationSeconds, defaultDurationBuckets)
	return args, nil
}

func bucketHRCurve(curve intervals.DataCurve, buckets []int) ([]hrCurvePoint, []int) {
	values, missing := durationCurveBucketValues(curve, buckets)
	points := make([]hrCurvePoint, 0, len(values))
	for _, value := range values {
		points = append(points, hrCurvePoint{DurationSeconds: value.Bucket, HeartRateBPM: value.Value, ActivityID: value.ActivityID})
	}
	return points, missing
}

func hrCurvesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{"oldest": map[string]any{"type": "string", "description": "Local start date YYYY-MM-DD."}, "newest": map[string]any{"type": "string", "description": "Local end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "description": "Optional intervals.icu sport/type filter for the upstream type parameter. Omit to let upstream aggregate its default heart-rate curve set."}, "duration_seconds": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "minimum": 1}, "description": "Heart-rate curve duration buckets in seconds. Defaults to 5,15,30,60,300,1200,3600 seconds; point values are heart_rate_bpm."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream curve arrays and activity maps."}}}
}
