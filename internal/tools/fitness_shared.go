package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	invalidFitnessArgumentsMessage = "invalid fitness arguments; provide start_date/end_date as YYYY-MM-DD and optional include_full"
	invalidCurveArgumentsMessage   = "invalid curve arguments; provide valid dates, sports, and positive bucket values"
)

var defaultDurationBuckets = []int{5, 15, 30, 60, 300, 1200, 3600}

// FitnessClient retrieves athlete summary rows.
type FitnessClient interface {
	ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error)
}

type dateRangeRequest struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

func decodeDateRangeRequest(raw json.RawMessage) (dateRangeRequest, error) {
	var args dateRangeRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[dateRangeRequest](raw)
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
	return args, nil
}

func validDate(value string) bool { _, err := time.Parse(time.DateOnly, value); return err == nil }

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

func firstCurve(set intervals.DataCurveSet) intervals.DataCurve {
	if len(set.List) == 0 {
		return intervals.DataCurve{}
	}
	return set.List[0]
}

type curveBucketValue struct {
	Bucket     int
	Value      *float64
	ActivityID string
}

func durationCurveBucketValues(curve intervals.DataCurve, buckets []int) ([]curveBucketValue, []int) {
	return curveBucketValues(curve.Secs, curve, buckets)
}

func distanceCurveBucketValues(curve intervals.DataCurve, buckets []int) ([]curveBucketValue, []int) {
	return curveBucketValues(curve.Distance, curve, buckets)
}

func curveBucketValues(xs []float64, curve intervals.DataCurve, buckets []int) ([]curveBucketValue, []int) {
	points := []curveBucketValue{}
	missing := []int{}
	for _, bucket := range buckets {
		value, idx, ok := valueAtBucket(xs, curve.Values, bucket)
		if !ok {
			missing = append(missing, bucket)
			continue
		}
		points = append(points, curveBucketValue{Bucket: bucket, Value: roundPtr(value), ActivityID: activityIDAt(curve, idx)})
	}
	return points, missing
}

func valueAtBucket(xs []float64, values []float64, bucket int) (float64, int, bool) {
	for i, x := range xs {
		if int(math.Round(x)) == bucket && i < len(values) {
			return values[i], i, true
		}
	}
	return 0, 0, false
}

func toolProfile(ctx context.Context, profileClient ProfileClient, timezoneFallback string) (response.UnitSystem, string, error) {
	profile, unitSystem, timezone, err := toolProfileDetails(ctx, profileClient, timezoneFallback)
	_ = profile
	return unitSystem, timezone, err
}

func toolProfileDetails(ctx context.Context, profileClient ProfileClient, timezoneFallback string) (intervals.AthleteWithSportSettings, response.UnitSystem, string, error) {
	profile, err := profileClient.GetAthleteProfile(ctx)
	if err != nil {
		return intervals.AthleteWithSportSettings{}, "", "", err
	}
	unitSystem := profileUnitSystem(profile)
	timezone := profileTimezone(profile.Timezone, timezoneFallback)
	return profile, unitSystem, timezone, nil
}

func encodeShaped(payload any, includeFull bool, rowCollections []string, version string, debugMetadata bool, queryType string, unitSystem response.UnitSystem, shaping ...responseShaping) (Result, error) {
	shapeCfg := responseShapingOrDefault(shaping)
	shaped, err := response.Shape(payload, shapeCfg.options(includeFull, rowCollections, version, debugMetadata, queryType, unitSystem))
	if err != nil {
		return Result{}, err
	}
	if _, err := json.Marshal(shaped); err != nil {
		return Result{}, fmt.Errorf("encoding %s response: %w", queryType, err)
	}
	return TextResult(shaped), nil
}

func roundPtr(value float64) *float64 { rounded := round(value, 3); return &rounded }

func dateRangeInputSchema(startDescription string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{"start_date": map[string]any{"type": "string", "description": startDescription + " as YYYY-MM-DD in the athlete timezone."}, "end_date": map[string]any{"type": "string", "description": "local end date as YYYY-MM-DD in the athlete timezone."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream summary rows."}}}
}

func genericOutputSchema(description string) map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": description}
}
