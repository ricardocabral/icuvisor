package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/streams"
)

const (
	getClimbSegmentsName        = "get_climb_segments"
	getClimbSegmentsDescription = "Use when the prompt asks which sustained climbs occurred in one activity or asks to compare climb segments; do not fetch raw streams and reduce them in chat. This full-toolset read-only analyzer detects bounded climbs from distance and altitude, bridges only deterministic shelves, and reports stream-quality limitations without coaching or physiology claims."
	invalidGetClimbSegmentsArgs = "invalid get_climb_segments arguments; provide activity_id and bounded climb parameters"
	getClimbSegmentsFetchMsg    = "could not analyze activity climbs; check intervals.icu credentials, athlete ID, and stream availability"
)

type getClimbSegmentsRequest struct {
	ActivityID               string   `json:"activity_id"`
	MinGradePercent          *float64 `json:"min_grade_percent,omitempty"`
	MinElevationGainM        *float64 `json:"min_elevation_gain_m,omitempty"`
	MaxGapDistanceM          *float64 `json:"max_gap_distance_m,omitempty"`
	MaxBridgedElevationLossM *float64 `json:"max_bridged_elevation_loss_m,omitempty"`
	IncludeFull              bool     `json:"include_full,omitempty"`
}

func (r *getClimbSegmentsRequest) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return errors.New("arguments must be a JSON object")
	}
	for key := range fields {
		switch key {
		case "activity_id", "min_grade_percent", "min_elevation_gain_m", "max_gap_distance_m", "max_bridged_elevation_loss_m", "include_full":
		default:
			return fmt.Errorf("unsupported property %q", key)
		}
	}
	activityID, ok := fields["activity_id"]
	if !ok || string(activityID) == "null" || json.Unmarshal(activityID, &r.ActivityID) != nil || strings.TrimSpace(r.ActivityID) == "" {
		return errors.New("activity_id is required and must be a non-empty string")
	}
	if value, ok := fields["min_grade_percent"]; ok {
		number, err := decodeClimbParameter(value, "min_grade_percent")
		if err != nil {
			return err
		}
		r.MinGradePercent = &number
	}
	if value, ok := fields["min_elevation_gain_m"]; ok {
		number, err := decodeClimbParameter(value, "min_elevation_gain_m")
		if err != nil {
			return err
		}
		r.MinElevationGainM = &number
	}
	if value, ok := fields["max_gap_distance_m"]; ok {
		number, err := decodeClimbParameter(value, "max_gap_distance_m")
		if err != nil {
			return err
		}
		r.MaxGapDistanceM = &number
	}
	if value, ok := fields["max_bridged_elevation_loss_m"]; ok {
		number, err := decodeClimbParameter(value, "max_bridged_elevation_loss_m")
		if err != nil {
			return err
		}
		r.MaxBridgedElevationLossM = &number
	}
	if value, ok := fields["include_full"]; ok {
		if string(value) == "null" || json.Unmarshal(value, &r.IncludeFull) != nil {
			return errors.New("include_full must be a boolean")
		}
	}
	return nil
}

func decodeClimbParameter(data json.RawMessage, name string) (float64, error) {
	if string(data) == "null" {
		return 0, fmt.Errorf("%s must be a finite number", name)
	}
	var value float64
	if err := json.Unmarshal(data, &value); err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("%s must be a finite number", name)
	}
	return value, nil
}

func newGetClimbSegmentsTool(client ActivityStreamsClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{
		Name:         getClimbSegmentsName,
		Description:  getClimbSegmentsDescription,
		InputSchema:  getClimbSegmentsInputSchema(),
		OutputSchema: genericOutputSchema("Deterministic climb segments, data quality diagnostics, and effective analyzer parameters."),
		Requirement:  RequirementRead,
		Handler:      getClimbSegmentsHandler(client, version, debugMetadata, shapeCfg),
	})
}

func getClimbSegmentsHandler(client ActivityStreamsClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		var args getClimbSegmentsRequest
		if err := decodeJSONArgs(req.Arguments, &args); err != nil {
			return Result{}, NewUserError(invalidGetClimbSegmentsArgs, err)
		}
		input := analysis.ClimbSegmentsInput{
			MinGradePercent:          analysis.DefaultClimbMinGradePercent,
			MinElevationGainM:        analysis.DefaultClimbMinElevationGainM,
			MaxGapDistanceM:          analysis.DefaultClimbMaxGapDistanceM,
			MaxBridgedElevationLossM: analysis.DefaultClimbMaxBridgedLossM,
		}
		if args.MinGradePercent != nil {
			input.MinGradePercent = *args.MinGradePercent
		}
		if args.MinElevationGainM != nil {
			input.MinElevationGainM = *args.MinElevationGainM
		}
		if args.MaxGapDistanceM != nil {
			input.MaxGapDistanceM = *args.MaxGapDistanceM
		}
		if args.MaxBridgedElevationLossM != nil {
			input.MaxBridgedElevationLossM = *args.MaxBridgedElevationLossM
		}
		if _, err := analysis.AnalyzeClimbSegments(input); err != nil {
			return Result{}, NewUserError(invalidGetClimbSegmentsArgs, err)
		}
		rows, err := client.GetActivityStreams(ctx, intervals.ActivityStreamsParams{ActivityID: strings.TrimSpace(args.ActivityID), Types: []string{"distance", "altitude", "time", "heartrate", "watts"}, IncludeDefaults: false})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(getClimbSegmentsFetchMsg, err)
		}
		input.Distance = climbStreamFromRows(rows, "distance")
		input.Altitude = climbStreamFromRows(rows, "altitude")
		input.Time = climbStreamFromRows(rows, "time")
		input.HeartRate = climbStreamFromRows(rows, "heart_rate")
		input.Watts = climbStreamFromRows(rows, "watts")
		computed, err := analysis.AnalyzeClimbSegments(input)
		if err != nil {
			return Result{}, NewUserError(invalidGetClimbSegmentsArgs, err)
		}
		status := computed.DataQuality.Status
		assumptions := map[string]any{
			"min_grade_percent":            computed.Parameters.MinGradePercent,
			"min_elevation_gain_m":         computed.Parameters.MinElevationGainM,
			"max_gap_distance_m":           computed.Parameters.MaxGapDistanceM,
			"max_bridged_elevation_loss_m": computed.Parameters.MaxBridgedElevationLossM,
			"resample_m":                   analysis.ClimbResampleM,
			"quality_status":               status,
		}
		insufficient := status == analysis.ClimbStatusMissing || status == analysis.ClimbStatusNull || status == analysis.ClimbStatusInvalidDistance || status == analysis.ClimbStatusInsufficient
		n := computed.DataQuality.UsableSamples
		if status == analysis.ClimbStatusMissing || status == analysis.ClimbStatusInvalidDistance {
			n = 0
		}
		return encodeAnalyzerResponse(analyzerResponseInput{
			Result: computed,
			Meta: analysis.AnalyzerMetaInput{
				Method:             "distance_normalized_climb_detection_with_bounded_bridging",
				SourceTools:        []string{getActivityStreamsName},
				N:                  n,
				MinSamples:         2,
				MissingAction:      analysis.MissingActionSkip,
				InsufficientSample: &insufficient,
				Assumptions:        assumptions,
			},
		}, args.IncludeFull, version, debugMetadata, getClimbSegmentsName, response.UnitSystemMetric, shapeCfg)
	}
}

func climbStreamFromRows(rows []intervals.ActivityStream, canonical string) analysis.ClimbStream {
	for _, row := range rows {
		key, _ := streams.CanonicalKey(firstNonEmpty(row.Type, row.Name))
		if key == canonical {
			return climbStreamFromRow(row)
		}
	}
	return analysis.ClimbStream{}
}

func climbStreamFromRow(row intervals.ActivityStream) analysis.ClimbStream {
	stream := analysis.ClimbStream{Present: true, RawLength: len(row.Data), AllNull: row.AllNull}
	if row.Raw != nil {
		data, exists := row.Raw["data"]
		if exists {
			switch values := data.(type) {
			case nil:
				stream.DataState = "data_null"
				stream.RawLength = 0
			case []any:
				stream.Values = make([]float64, len(values))
				stream.Valid = make([]bool, len(values))
				stream.RawLength = len(values)
				for i, value := range values {
					if value == nil {
						stream.NullCount++
						continue
					}
					number, ok := rawFiniteNumber(value)
					if !ok {
						stream.InvalidCount++
						continue
					}
					stream.Values[i], stream.Valid[i] = number, true
				}
				if row.AllNull {
					stream.DataState = "all_null"
				} else if len(values) > 0 && stream.NullCount == len(values) {
					stream.DataState = "null"
				}
			default:
				stream.DataState = "null"
				stream.RawLength = 0
			}
			if stream.DataState == "" && row.AllNull {
				stream.DataState = "all_null"
			}
			return stream
		}
	}
	if row.AllNull {
		stream.DataState = "all_null"
		return stream
	}
	if len(row.Data) == 0 {
		stream.DataState = "empty"
		return stream
	}
	stream.Values = append([]float64(nil), row.Data...)
	stream.Valid = make([]bool, len(stream.Values))
	for i, value := range stream.Values {
		stream.Valid[i] = finiteClimbNumber(value)
		if !stream.Valid[i] {
			stream.InvalidCount++
		}
	}
	return stream
}

func rawFiniteNumber(value any) (float64, bool) {
	var number float64
	switch typed := value.(type) {
	case float64:
		number = typed
	case float32:
		number = float64(typed)
	case int:
		number = float64(typed)
	case int64:
		number = float64(typed)
	case json.Number:
		parsed, err := typed.Float64()
		if err != nil {
			return 0, false
		}
		number = parsed
	default:
		return 0, false
	}
	return number, finiteClimbNumber(number)
}

func finiteClimbNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func getClimbSegmentsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":                  map[string]any{"type": "string", "description": "Required intervals.icu activity ID whose distance, altitude, time, heart-rate, and power streams are analyzed."},
		"min_grade_percent":            map[string]any{"type": "number", "default": analysis.DefaultClimbMinGradePercent, "minimum": 0.1, "maximum": 100, "description": "Minimum inclusive average grade for a candidate climb, in percent; bounded 0.1..100."},
		"min_elevation_gain_m":         map[string]any{"type": "number", "default": analysis.DefaultClimbMinElevationGainM, "minimum": 0, "maximum": 100000, "description": "Minimum net elevation gain for a final segment, in meters; bounded 0..100000."},
		"max_gap_distance_m":           map[string]any{"type": "number", "default": analysis.DefaultClimbMaxGapDistanceM, "minimum": 0, "maximum": 10000, "description": "Maximum distance gap allowed when bridging adjacent candidate runs, in meters; bounded 0..10000."},
		"max_bridged_elevation_loss_m": map[string]any{"type": "number", "default": analysis.DefaultClimbMaxBridgedLossM, "minimum": 0, "maximum": 1000, "description": "Maximum downhill loss allowed across a bridged shelf, in meters; bounded 0..1000."},
		"include_full":                 map[string]any{"type": "boolean", "default": false, "description": "Accepted for analyzer compatibility; climb analysis never returns source arrays or a series field."},
	}}
}
