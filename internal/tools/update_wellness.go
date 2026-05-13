package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	updateWellnessName                    = "update_wellness"
	updateWellnessDescription             = "Update one athlete-local wellness row with sparse manual fields: subjective scales, measurements, injury text, and locked; device-owned sleepScore and _native bridge fields are not writable."
	invalidUpdateWellnessArgumentsMessage = "invalid update_wellness arguments; provide date as YYYY-MM-DD and writable wellness fields with documented ranges"
	writeWellnessMessage                  = "could not update wellness; check intervals.icu credentials, athlete ID, date, lock state, and writable fields"
	poundsToKilograms                     = 0.45359237
)

var updateWellnessSubjectiveScaleFields = []string{
	"feel",
	"fatigue",
	"mood",
	"sleepQuality",
	"motivation",
	"soreness",
	"stress",
}

var updateWellnessMeasurementFields = []string{
	"weight",
	"bodyFat",
	"systolic",
	"diastolic",
	"bloodGlucose",
	"lactate",
	"restingHR",
	"hrv",
}

var updateWellnessFreeTextFields = []string{
	"injury",
}

var updateWellnessFlagFields = []string{
	"locked",
}

var updateWellnessReadOnlyFields = []string{
	"sleepScore",
	"_native",
}

// WellnessWriterClient updates athlete wellness rows for tools.
type WellnessWriterClient interface {
	UpdateWellness(context.Context, intervals.WriteWellnessParams) (intervals.Wellness, error)
}

type updateWellnessRequest struct {
	Date         string   `json:"date"`
	Feel         *int     `json:"feel,omitempty"`
	Fatigue      *int     `json:"fatigue,omitempty"`
	Mood         *int     `json:"mood,omitempty"`
	SleepQuality *int     `json:"sleepQuality,omitempty"`
	Motivation   *int     `json:"motivation,omitempty"`
	Soreness     *int     `json:"soreness,omitempty"`
	Stress       *int     `json:"stress,omitempty"`
	Weight       *float64 `json:"weight,omitempty"`
	BodyFat      *float64 `json:"bodyFat,omitempty"`
	Systolic     *int     `json:"systolic,omitempty"`
	Diastolic    *int     `json:"diastolic,omitempty"`
	BloodGlucose *float64 `json:"bloodGlucose,omitempty"`
	Lactate      *float64 `json:"lactate,omitempty"`
	RestingHR    *int     `json:"restingHR,omitempty"`
	HRV          *float64 `json:"hrv,omitempty"`
	Injury       *string  `json:"injury,omitempty"`
	Locked       *bool    `json:"locked,omitempty"`
	IncludeFull  bool     `json:"include_full,omitempty"`
}

type updateWellnessResponse struct {
	Wellness map[string]any       `json:"wellness"`
	Meta     updateWellnessMeta   `json:"_meta"`
}

type updateWellnessMeta struct {
	Date               string   `json:"date"`
	FieldsUpdated      []string `json:"fields_updated"`
	WeightInputUnit    string   `json:"weight_input_unit,omitempty"`
	WeightUpstreamUnit string   `json:"weight_upstream_unit,omitempty"`
	Locked             bool     `json:"locked,omitempty"`
	IncludeFull        bool     `json:"include_full"`
}

func newUpdateWellnessTool(client WellnessWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: updateWellnessName, Description: updateWellnessDescription, InputSchema: updateWellnessInputSchema(), OutputSchema: updateWellnessOutputSchema(), Requirement: RequirementWrite, Handler: updateWellnessHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func updateWellnessHandler(client WellnessWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeUpdateWellnessRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateWellnessArgumentsMessage, err)
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			return Result{}, NewUserError(writeWellnessMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(writeWellnessMessage, errors.New("missing wellness writer client"))
		}
		params, meta := wellnessWriteParams(args, profile)
		updated, err := client.UpdateWellness(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(writeWellnessMessage, err)
		}
		payload, err := shapeUpdateWellnessResponse(updated, meta, args.IncludeFull, version, debugMetadata, updateWellnessName, profileUnitSystem(profile))
		if err != nil {
			return Result{}, fmt.Errorf("shaping update_wellness response: %w", err)
		}
		text, err := json.Marshal(payload)
		if err != nil {
			return Result{}, fmt.Errorf("encoding update_wellness response: %w", err)
		}
		return Result{Content: []Content{{Type: ContentTypeText, Text: string(text)}}, StructuredContent: payload}, nil
	}
}

func decodeUpdateWellnessRequest(raw json.RawMessage) (updateWellnessRequest, error) {
	if err := rejectReadOnlyWellnessFields(raw); err != nil {
		return updateWellnessRequest{}, err
	}
	var args updateWellnessRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.Date = strings.TrimSpace(args.Date)
	if !validDate(args.Date) {
		return args, errors.New("date must be athlete-local YYYY-MM-DD")
	}
	if err := validateUpdateWellnessRanges(args); err != nil {
		return args, err
	}
	if len(updateWellnessFieldsUpdated(args)) == 0 {
		return args, errors.New("at least one writable wellness field is required")
	}
	return args, nil
}

func rejectReadOnlyWellnessFields(raw json.RawMessage) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil
	}
	if _, ok := fields["sleepScore"]; ok {
		return errors.New("field_not_writable: sleepScore (device-managed)")
	}
	if _, ok := fields["_native"]; ok {
		return errors.New("field_not_writable: _native (bridge-managed)")
	}
	return nil
}

func validateUpdateWellnessRanges(args updateWellnessRequest) error {
	for field, value := range map[string]*int{
		"feel":       args.Feel,
		"fatigue":    args.Fatigue,
		"mood":       args.Mood,
		"motivation": args.Motivation,
		"soreness":   args.Soreness,
		"stress":     args.Stress,
	} {
		if err := validateIntRange(field, value, 1, 5); err != nil {
			return err
		}
	}
	if err := validateIntRange("sleepQuality", args.SleepQuality, 1, 4); err != nil {
		return err
	}
	for field, value := range map[string]*float64{
		"weight":       args.Weight,
		"bodyFat":      args.BodyFat,
		"bloodGlucose": args.BloodGlucose,
		"lactate":      args.Lactate,
		"hrv":          args.HRV,
	} {
		if err := validateFloatMin(field, value, 0); err != nil {
			return err
		}
	}
	for field, value := range map[string]*int{"systolic": args.Systolic, "diastolic": args.Diastolic, "restingHR": args.RestingHR} {
		if err := validateIntMin(field, value, 0); err != nil {
			return err
		}
	}
	return nil
}

func validateIntRange(field string, value *int, min int, max int) error {
	if value == nil {
		return nil
	}
	if *value < min || *value > max {
		return fmt.Errorf("%s must be %d-%d", field, min, max)
	}
	return nil
}

func validateIntMin(field string, value *int, min int) error {
	if value == nil {
		return nil
	}
	if *value < min {
		return fmt.Errorf("%s must be >= %d", field, min)
	}
	return nil
}

func validateFloatMin(field string, value *float64, min float64) error {
	if value == nil {
		return nil
	}
	if *value < min {
		return fmt.Errorf("%s must be >= %g", field, min)
	}
	return nil
}

func wellnessWriteParams(args updateWellnessRequest, profile intervals.AthleteWithSportSettings) (intervals.WriteWellnessParams, updateWellnessMeta) {
	params := intervals.WriteWellnessParams{
		Date:         args.Date,
		Feel:         args.Feel,
		Fatigue:      args.Fatigue,
		Mood:         args.Mood,
		SleepQuality: args.SleepQuality,
		Motivation:   args.Motivation,
		Soreness:     args.Soreness,
		Stress:       args.Stress,
		BodyFat:      args.BodyFat,
		Systolic:     args.Systolic,
		Diastolic:    args.Diastolic,
		BloodGlucose: args.BloodGlucose,
		Lactate:      args.Lactate,
		RestingHR:    args.RestingHR,
		HRV:          args.HRV,
		Injury:       args.Injury,
		Locked:       args.Locked,
	}
	meta := updateWellnessMeta{Date: args.Date, FieldsUpdated: updateWellnessFieldsUpdated(args), IncludeFull: args.IncludeFull}
	if args.Weight != nil {
		weight := *args.Weight
		meta.WeightInputUnit = "kg"
		meta.WeightUpstreamUnit = "kg"
		if profile.WeightPrefLB {
			weight *= poundsToKilograms
			meta.WeightInputUnit = "lb"
		}
		params.Weight = &weight
	}
	return params, meta
}

func updateWellnessFieldsUpdated(args updateWellnessRequest) []string {
	fields := make([]string, 0, len(updateWellnessSubjectiveScaleFields)+len(updateWellnessMeasurementFields)+len(updateWellnessFreeTextFields)+len(updateWellnessFlagFields))
	add := func(name string, present bool) {
		if present {
			fields = append(fields, name)
		}
	}
	add("feel", args.Feel != nil)
	add("fatigue", args.Fatigue != nil)
	add("mood", args.Mood != nil)
	add("sleepQuality", args.SleepQuality != nil)
	add("motivation", args.Motivation != nil)
	add("soreness", args.Soreness != nil)
	add("stress", args.Stress != nil)
	add("weight", args.Weight != nil)
	add("bodyFat", args.BodyFat != nil)
	add("systolic", args.Systolic != nil)
	add("diastolic", args.Diastolic != nil)
	add("bloodGlucose", args.BloodGlucose != nil)
	add("lactate", args.Lactate != nil)
	add("restingHR", args.RestingHR != nil)
	add("hrv", args.HRV != nil)
	add("injury", args.Injury != nil)
	add("locked", args.Locked != nil)
	sort.Strings(fields)
	return fields
}

func shapeUpdateWellnessResponse(row intervals.Wellness, meta updateWellnessMeta, includeFull bool, version string, debugMetadata bool, queryType string, unitSystem response.UnitSystem) (updateWellnessResponse, error) {
	shapedRow, err := response.Shape(wellnessRow(row, includeFull), response.Options{IncludeFull: includeFull, ServerVersion: version, DebugMetadata: debugMetadata, QueryType: queryType, UnitSystem: unitSystem})
	if err != nil {
		return updateWellnessResponse{}, err
	}
	wellness, ok := shapedRow.(map[string]any)
	if !ok {
		return updateWellnessResponse{}, errors.New("wellness response row did not shape to object")
	}
	if row.Locked != nil && *row.Locked {
		meta.Locked = true
	}
	return updateWellnessResponse{Wellness: wellness, Meta: meta}, nil
}

func updateWellnessInputSchema() map[string]any {
	scales := response.RegisteredScaleLabels()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"date"}, "properties": map[string]any{
		"date":         map[string]any{"type": "string", "description": "Required athlete-local wellness date as YYYY-MM-DD."},
		"feel":         scaleSchema(scales, "feel", 1, 5),
		"fatigue":      scaleSchema(scales, "fatigue", 1, 5),
		"mood":         scaleSchema(scales, "mood", 1, 5),
		"sleepQuality": scaleSchema(scales, "sleepQuality", 1, 4),
		"motivation":   scaleSchema(scales, "motivation", 1, 5),
		"soreness":     scaleSchema(scales, "soreness", 1, 5),
		"stress":       scaleSchema(scales, "stress", 1, 5),
		"weight":       map[string]any{"type": "number", "minimum": 0, "description": "Manual body weight in the athlete's preferred weight unit from get_athlete_profile (_meta.units / units.weight); converted to upstream kg at the API boundary."},
		"bodyFat":      map[string]any{"type": "number", "minimum": 0, "maximum": 100, "description": "Manual body fat percentage, 0-100%."},
		"systolic":     map[string]any{"type": "integer", "minimum": 0, "description": "Manual systolic blood pressure in mmHg."},
		"diastolic":    map[string]any{"type": "integer", "minimum": 0, "description": "Manual diastolic blood pressure in mmHg."},
		"bloodGlucose": map[string]any{"type": "number", "minimum": 0, "description": "Manual blood glucose in the upstream intervals.icu wellness unit."},
		"lactate":      map[string]any{"type": "number", "minimum": 0, "description": "Manual blood lactate in mmol/L."},
		"restingHR":    map[string]any{"type": "integer", "minimum": 0, "description": "Manual resting heart rate in bpm."},
		"hrv":          map[string]any{"type": "number", "minimum": 0, "description": "Manual HRV in milliseconds rMSSD."},
		"injury":       map[string]any{"type": "string", "description": "Optional free-text injury or limitation note. Preserved verbatim."},
		"locked":       map[string]any{"type": "boolean", "description": "When true, ask upstream to lock the wellness row against device-sync overwrites."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream wellness row under wellness.full and keep null fields."},
	}}
}

func scaleSchema(scales map[string]string, field string, min int, max int) map[string]any {
	return map[string]any{"type": "integer", "minimum": min, "maximum": max, "description": fmt.Sprintf("%s; %s scale.", scales[field], field)}
}

func updateWellnessOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Updated wellness row using the same terse read shape as get_wellness_data, plus write metadata and delete-mode/unit metadata."}
}
