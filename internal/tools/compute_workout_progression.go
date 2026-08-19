package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	computeWorkoutProgressionName             = "compute_workout_progression"
	computeWorkoutProgressionDescription      = "Use when the user supplies an explicit ordered sequence of repeated workout activity IDs and asks for factual progression evidence. Compares only the supplied sequence: prescribed/completed structure, target adherence, stability, duration/recovery, source-labelled feel/RPE, and optional wellness fields. It reports evidence gaps and never assigns a progression score, recommends changes, or writes the calendar."
	invalidWorkoutProgressionArgumentsMessage = "invalid compute_workout_progression arguments; provide 2-20 ordered activity objects with non-empty activity_id values"
	fetchWorkoutProgressionMessage            = "could not compute workout progression evidence; check activity IDs, intervals.icu credentials, and athlete-local dates"
	maxWorkoutProgressionActivities           = 20
	maxWorkoutProgressionAudit                = 200
	maxWorkoutProgressionIntervals            = 1000
)

var workoutProgressionDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(?:$|T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?$)`)

// WorkoutProgressionExtendedClient is the narrow optional extended source used by the analyzer.
type WorkoutProgressionExtendedClient interface {
	GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error)
}

// WorkoutProgressionEventClient retrieves one exact event and never lists events.
type WorkoutProgressionEventClient interface {
	GetEvent(context.Context, string) (intervals.Event, error)
}

type computeWorkoutProgressionRequest struct {
	Activities       []computeWorkoutProgressionActivity `json:"activities"`
	IncludeFull      bool                                `json:"include_full,omitempty"`
	IncludeReadiness bool                                `json:"include_readiness,omitempty"`
}

type computeWorkoutProgressionActivity struct {
	ActivityID string
	Label      string
	EventID    string
}

// newComputeWorkoutProgressionTool creates the read-only progression evidence tool.
//
//nolint:unparam // version is supplied by registry wiring when the full toolset is registered.
func newComputeWorkoutProgressionTool(details ActivityDetailsClient, intervalClient ActivityIntervalsClient, extended WorkoutProgressionExtendedClient, eventClient WorkoutProgressionEventClient, wellness WellnessClient, profile ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{
		Name:         computeWorkoutProgressionName,
		Description:  computeWorkoutProgressionDescription,
		InputSchema:  computeWorkoutProgressionInputSchema(),
		OutputSchema: genericOutputSchema("Ordered workout progression evidence with explicit structure, metric, and missing-data diagnostics; no score or recommendation."),
		Handler:      computeWorkoutProgressionHandler(details, intervalClient, extended, eventClient, wellness, profile, version, timezoneFallback, debugMetadata, shapeCfg),
	})
}

func computeWorkoutProgressionHandler(details ActivityDetailsClient, intervalClient ActivityIntervalsClient, extended WorkoutProgressionExtendedClient, eventClient WorkoutProgressionEventClient, wellness WellnessClient, profile ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeComputeWorkoutProgressionRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidWorkoutProgressionArgumentsMessage, err)
		}
		sourceTools := []string{}
		profileTimezone := ""
		if args.IncludeReadiness {
			if profile == nil {
				return Result{}, NewUserError(fetchWorkoutProgressionMessage, errors.New("readiness requires an athlete profile source"))
			}
			sourceTools = append(sourceTools, "get_athlete_profile")
			athlete, profileErr := profile.GetAthleteProfile(ctx)
			if profileErr != nil {
				if contextError(profileErr) {
					return Result{}, profileErr
				}
				return Result{}, NewUserError(fetchWorkoutProgressionMessage, profileErr)
			}
			profileTimezone = strings.TrimSpace(athlete.Timezone)
			if profileTimezone == "" {
				profileTimezone = strings.TrimSpace(timezoneFallback)
			}
			if _, tzErr := time.LoadLocation(profileTimezone); tzErr != nil {
				return Result{}, NewUserError(fetchWorkoutProgressionMessage, fmt.Errorf("invalid profile timezone %q: %w", profileTimezone, tzErr))
			}
		}
		inputs := make([]analysis.WorkoutProgressionActivity, len(args.Activities))
		validDates := []string{}
		missingDays := 0
		for i, requested := range args.Activities {
			input := analysis.WorkoutProgressionActivity{ID: requested.ActivityID, Label: requested.Label}
			if args.IncludeReadiness {
				input.Readiness = emptyProgressionReadiness("")
			}
			if details == nil {
				input.InitialReasons = append(input.InitialReasons, "missing_source")
				inputs[i] = input
				continue
			}
			sourceTools = append(sourceTools, "get_activity_details")
			activity, detailErr := details.GetActivity(ctx, requested.ActivityID)
			if detailErr != nil {
				if contextError(detailErr) {
					return Result{}, detailErr
				}
				input.InitialReasons = append(input.InitialReasons, "missing_source")
				inputs[i] = input
				continue
			}
			input.Sport = activityType(activity)
			if date, dateOK := progressionActivityDate(activity); dateOK {
				input.Date = date
				if args.IncludeReadiness {
					input.Readiness = emptyProgressionReadiness(date)
				}
			}
			if args.IncludeReadiness {
				applyReadinessDateDiagnostics(&input, activity, profileTimezone)
				if input.Date != "" && !progressionHasInvalidDate(input.InitialReasons) {
					validDates = append(validDates, input.Date)
				}
			}
			if isRestrictedProgressionActivity(activity) {
				input.InitialReasons = append(input.InitialReasons, "restricted_source")
				inputs[i] = input
				continue
			}
			linkedEventID, aliasConflict := progressionEventID(activity)
			if requested.EventID != "" {
				linkedEventID = requested.EventID
				aliasConflict = false
			}
			if aliasConflict {
				input.InitialReasons = append(input.InitialReasons, "missing_prescription")
			} else if linkedEventID != "" && eventClient != nil {
				sourceTools = append(sourceTools, "get_event_by_id")
				event, eventErr := eventClient.GetEvent(ctx, linkedEventID)
				if eventErr != nil {
					if contextError(eventErr) {
						return Result{}, eventErr
					}
					input.InitialReasons = append(input.InitialReasons, "missing_prescription")
				} else if prescription, prescriptionErr := progressionPrescriptionFromEvent(event); prescriptionErr == nil {
					input.Prescription = prescription
				} else {
					input.InitialReasons = append(input.InitialReasons, "missing_prescription")
				}
			} else if linkedEventID != "" {
				input.InitialReasons = append(input.InitialReasons, "missing_prescription")
			} else if rawDoc, present := activity.Raw["workout_doc"]; present {
				if prescription, prescriptionErr := progressionPrescriptionFromValue(rawDoc, "activity_workout_doc"); prescriptionErr == nil {
					input.Prescription = prescription
				} else {
					input.InitialReasons = append(input.InitialReasons, "missing_prescription")
				}
			}
			if intervalClient == nil {
				input.InitialReasons = append(input.InitialReasons, "missing_intervals")
			} else {
				sourceTools = append(sourceTools, "get_activity_intervals")
				dto, intervalsErr := intervalClient.GetActivityIntervals(ctx, requested.ActivityID)
				if intervalsErr != nil {
					if contextError(intervalsErr) {
						return Result{}, intervalsErr
					}
					input.InitialReasons = append(input.InitialReasons, "missing_intervals")
				} else {
					input.Completed = progressionCompletedFromDTO(dto)
					if extended != nil {
						sourceTools = append(sourceTools, "get_extended_metrics")
						metrics, metricsErr := extended.GetActivityPowerVsHR(ctx, requested.ActivityID)
						if metricsErr != nil {
							if contextError(metricsErr) {
								return Result{}, metricsErr
							}
						} else {
							input.PowerHR = finiteFloatPointer(metrics.PowerHR)
							input.DecouplingPercent = finiteFloatPointer(metrics.Decoupling)
						}
					}
				}
			}
			input.DurationSeconds, input.DurationSource = progressionDuration(activity)
			input.Feel = finiteIntAsFloat(activity.Feel)
			input.RPE = finiteIntAsFloat(activity.RPE)
			input.UpstreamCompliancePercent = progressionRawFloat(activity.Raw, "compliance")
			inputs[i] = input
		}
		if args.IncludeReadiness && len(validDates) > 0 {
			var readinessErr error
			missingDays, readinessErr = loadProgressionReadiness(ctx, inputs, validDates, wellness, profileTimezone, &sourceTools)
			if readinessErr != nil {
				if contextError(readinessErr) {
					return Result{}, readinessErr
				}
				for i := range inputs {
					if inputs[i].Date != "" && !progressionHasInvalidDate(inputs[i].InitialReasons) {
						inputs[i].InitialReasons = append(inputs[i].InitialReasons, "missing_readiness")
					}
				}
			}
		}
		computed := analysis.AnalyzeWorkoutProgression(inputs, args.IncludeFull)
		meta := analysis.AnalyzerMetaInput{Method: analysis.WorkoutProgressionMethod, SourceTools: sourceTools, N: len(inputs), MissingDays: missingDays, MissingAction: analysis.MissingActionSkip, MinSamples: 2, IncludeMinSamples: true, FormulaRef: analysis.WorkoutProgressionFormulaRef, InsufficientSample: progressionBoolPtr(computed.Status != "ok")}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: computed, Meta: meta}, args.IncludeFull, version, debugMetadata, computeWorkoutProgressionName, response.UnitSystemMetric, shapeCfg)
	}
}

func decodeComputeWorkoutProgressionRequest(raw json.RawMessage) (computeWorkoutProgressionRequest, error) {
	var args computeWorkoutProgressionRequest
	if len(raw) == 0 || string(raw) == "null" || strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return args, errors.New("arguments must be a JSON object")
	}
	allowed := map[string]bool{"activities": true, "include_full": true, "include_readiness": true}
	for key := range object {
		if !allowed[key] {
			return args, fmt.Errorf("unknown argument %q", key)
		}
	}
	activityRaw, ok := object["activities"]
	if !ok {
		return args, errors.New("activities is required")
	}
	var activityMessages []json.RawMessage
	if err := json.Unmarshal(activityRaw, &activityMessages); err != nil || activityMessages == nil {
		return args, errors.New("activities must be an array")
	}
	args.Activities = make([]computeWorkoutProgressionActivity, len(activityMessages))
	if len(args.Activities) < 2 || len(args.Activities) > maxWorkoutProgressionActivities {
		return args, fmt.Errorf("activities must contain 2-%d items", maxWorkoutProgressionActivities)
	}
	if value, present := object["include_full"]; present {
		if string(value) == "null" || json.Unmarshal(value, &args.IncludeFull) != nil || !isJSONBool(value) {
			return args, errors.New("include_full must be a boolean")
		}
	}
	if value, present := object["include_readiness"]; present {
		if string(value) == "null" || json.Unmarshal(value, &args.IncludeReadiness) != nil || !isJSONBool(value) {
			return args, errors.New("include_readiness must be a boolean")
		}
	}
	seen := map[string]bool{}
	for i, rawActivity := range activityMessages {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(rawActivity, &fields); err != nil || fields == nil {
			return args, fmt.Errorf("activities[%d] must be an object", i)
		}
		for key := range fields {
			if key != "activity_id" && key != "label" && key != "event_id" {
				return args, fmt.Errorf("activities[%d] has unknown key %q", i, key)
			}
		}
		idRaw, present := fields["activity_id"]
		if !present || !isJSONString(idRaw) || json.Unmarshal(idRaw, &args.Activities[i].ActivityID) != nil {
			return args, fmt.Errorf("activities[%d].activity_id must be a string", i)
		}
		args.Activities[i].ActivityID = strings.TrimSpace(args.Activities[i].ActivityID)
		if args.Activities[i].ActivityID == "" {
			return args, fmt.Errorf("activities[%d].activity_id must not be empty", i)
		}
		if seen[args.Activities[i].ActivityID] {
			return args, fmt.Errorf("duplicate activity_id %q", args.Activities[i].ActivityID)
		}
		seen[args.Activities[i].ActivityID] = true
		if labelRaw, present := fields["label"]; present {
			if !isJSONString(labelRaw) || json.Unmarshal(labelRaw, &args.Activities[i].Label) != nil {
				return args, fmt.Errorf("activities[%d].label must be a string", i)
			}
			args.Activities[i].Label = strings.TrimSpace(args.Activities[i].Label)
		}
		if eventRaw, present := fields["event_id"]; present {
			if !isJSONString(eventRaw) || json.Unmarshal(eventRaw, &args.Activities[i].EventID) != nil {
				return args, fmt.Errorf("activities[%d].event_id must be a string", i)
			}
			args.Activities[i].EventID = strings.TrimSpace(args.Activities[i].EventID)
			if args.Activities[i].EventID == "" {
				return args, fmt.Errorf("activities[%d].event_id must not be empty when provided", i)
			}
		}
	}
	return args, nil
}

func computeWorkoutProgressionInputSchema() map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"required": []string{"activities"},
		"properties": map[string]any{
			"activities":        map[string]any{"type": "array", "minItems": 2, "maxItems": maxWorkoutProgressionActivities, "description": "Ordered comparable completed activities; labels are display-only and activity IDs are never matched by name.", "items": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{"activity_id": map[string]any{"type": "string", "description": "Required activity ID."}, "label": map[string]any{"type": "string", "description": "Optional display label only."}, "event_id": map[string]any{"type": "string", "description": "Optional exact planned event ID; no event list search is performed."}}}},
			"include_full":      map[string]any{"type": "boolean", "description": "Include at most the first 200 normalized prescription/completed interval audit rows; never includes raw streams."},
			"include_readiness": map[string]any{"type": "boolean", "description": "Include source-labelled daily wellness fields matched to each activity-local date."},
		},
	}
}

func progressionActivityDate(activity intervals.Activity) (string, bool) {
	if activity.StartDateLocal == nil {
		return "", false
	}
	value := strings.TrimSpace(*activity.StartDateLocal)
	if !workoutProgressionDatePattern.MatchString(value) || len(value) < 10 {
		return "", false
	}
	if _, err := time.Parse("2006-01-02", value[:10]); err != nil {
		return "", false
	}
	return value[:10], true
}

func applyReadinessDateDiagnostics(input *analysis.WorkoutProgressionActivity, activity intervals.Activity, profileTimezone string) {
	if activity.StartDateLocal == nil || strings.TrimSpace(progressionStringValue(activity.StartDateLocal)) == "" {
		input.InitialReasons = append(input.InitialReasons, "invalid_activity_date", "missing_readiness")
		return
	}
	if date, ok := progressionActivityDate(activity); ok {
		input.Date = date
		if input.Readiness == nil {
			input.Readiness = emptyProgressionReadiness(date)
		} else {
			input.Readiness.Date = date
		}
	} else {
		input.InitialReasons = append(input.InitialReasons, "invalid_activity_date", "missing_readiness")
		return
	}
	timezone := strings.TrimSpace(progressionStringValue(activity.Timezone))
	if timezone == "" {
		timezone = profileTimezone
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		input.InitialReasons = append(input.InitialReasons, "invalid_activity_date", "missing_readiness")
	}
}

func loadProgressionReadiness(ctx context.Context, inputs []analysis.WorkoutProgressionActivity, validDates []string, client WellnessClient, _ string, sourceTools *[]string) (int, error) {
	validRowCount := 0
	for _, input := range inputs {
		if input.Date != "" && !progressionHasInvalidDate(input.InitialReasons) {
			validRowCount++
		}
	}
	if client == nil {
		return validRowCount, errors.New("missing wellness source")
	}
	sortedDates := append([]string{}, validDates...)
	sort.Strings(sortedDates)
	start, end := sortedDates[0], sortedDates[len(sortedDates)-1]
	startTime, _ := time.Parse(time.DateOnly, start)
	endTime, _ := time.Parse(time.DateOnly, end)
	if int(endTime.Sub(startTime)/(24*time.Hour))+1 > 366 {
		for i := range inputs {
			if inputs[i].Date != "" && !progressionHasInvalidDate(inputs[i].InitialReasons) {
				inputs[i].InitialReasons = append(inputs[i].InitialReasons, "readiness_window_too_large", "missing_readiness")
				if inputs[i].Readiness != nil {
					inputs[i].Readiness.Reasons = append(inputs[i].Readiness.Reasons, "readiness_window_too_large", "missing_readiness")
				}
			}
		}
		return validRowCount, nil
	}
	*sourceTools = append(*sourceTools, "get_wellness_data")
	rows, err := client.ListWellness(ctx, intervals.WellnessParams{Oldest: start, Newest: end, Fields: progressionWellnessFields()})
	if err != nil {
		return validRowCount, err
	}
	missingDays := 0
	byDate := map[string]intervals.Wellness{}
	for _, row := range rows {
		date := progressionWellnessDate(row)
		if date == "" {
			continue
		}
		if _, exists := byDate[date]; !exists {
			byDate[date] = row
		}
	}
	for i := range inputs {
		if inputs[i].Date == "" || progressionHasInvalidDate(inputs[i].InitialReasons) {
			continue
		}
		row, ok := byDate[inputs[i].Date]
		if !ok {
			inputs[i].InitialReasons = append(inputs[i].InitialReasons, "missing_readiness")
			missingDays++
			continue
		}
		fields := map[string]analysis.ReadinessField{}
		for _, field := range progressionWellnessMetricFields() {
			value, present := progressionWellnessValue(row, field)
			if !present {
				continue
			}
			entry, stale := wellnessProvenanceEntry(row, field)
			readiness := analysis.ReadinessField{Value: value, Source: stringValueAny(entry["source"]), NativeScale: stringValueAny(entry["native_scale"]), Stale: stale}
			if fetched := stringValueAny(entry["fetched_at"]); fetched != "" && fetched != "unknown" {
				readiness.FetchedAt = fetched
			}
			fields[field] = readiness
		}
		inputs[i].Readiness = &analysis.ReadinessEvidence{Date: inputs[i].Date, Fields: fields, ExpectedFields: progressionWellnessMetricFields()}
	}
	return missingDays, nil
}

func progressionWellnessFields() []string {
	return []string{"id", "date", "updated", "feel", "fatigue", "soreness", "stress", "mood", "motivation", "sleepQuality", "sleepScore", "hrv", "hrvSDNN", "readiness", "source", "provider", "device", "wellnessSource", "wellness_source", "integration", "bridge_fetched_at", "bridgeFetchedAt", "fetched_at", "fetchedAt", "polar", "garmin", "oura", "whoop"}
}

func progressionWellnessMetricFields() []string {
	return []string{"feel", "fatigue", "soreness", "stress", "mood", "motivation", "sleepQuality", "sleepScore", "hrv", "hrvSDNN", "readiness"}
}

func emptyProgressionReadiness(date string) *analysis.ReadinessEvidence {
	return &analysis.ReadinessEvidence{Date: date, Fields: map[string]analysis.ReadinessField{}, ExpectedFields: progressionWellnessMetricFields(), Reasons: []string{"missing_readiness"}}
}

func progressionWellnessDate(row intervals.Wellness) string {
	candidates := []string{}
	if row.ID != nil {
		candidates = append(candidates, *row.ID)
	}
	for _, key := range []string{"date", "id"} {
		if value, ok := row.Raw[key].(string); ok {
			candidates = append(candidates, value)
		}
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if _, err := time.Parse(time.DateOnly, candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func progressionWellnessValue(row intervals.Wellness, field string) (any, bool) {
	value, present := row.Raw[field]
	if !present || value == nil {
		return nil, false
	}
	switch typed := value.(type) {
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return nil, false
		}
		return typed, true
	default:
		return nil, false
	}
}

func progressionPrescriptionFromEvent(event intervals.Event) (*analysis.WorkoutPrescription, error) {
	if raw, present := event.Raw["workout_doc"]; present {
		if prescription, err := progressionPrescriptionFromValue(raw, "explicit_event"); err == nil {
			return prescription, nil
		}
	} else if event.WorkoutDoc != nil {
		if prescription, err := progressionPrescriptionFromValue(event.WorkoutDoc, "explicit_event"); err == nil {
			return prescription, nil
		}
	}
	if event.Description != nil {
		validated := workoutdoc.ValidateDescription(*event.Description)
		if len(validated.Errors) == 0 && len(validated.Warnings) == 0 && validated.StructuredStepLines > 0 {
			return progressionPrescriptionFromDoc(validated.Doc, "explicit_event")
		}
	}
	return nil, errors.New("event has no valid structured prescription")
}

func progressionPrescriptionFromValue(value any, source string) (*analysis.WorkoutPrescription, error) {
	if value == nil {
		return nil, errors.New("workout_doc is null")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var doc workoutdoc.WorkoutDoc
	if err := json.Unmarshal(encoded, &doc); err != nil {
		return nil, err
	}
	return progressionPrescriptionFromDoc(doc, source)
}

func progressionPrescriptionFromDoc(doc workoutdoc.WorkoutDoc, source string) (*analysis.WorkoutPrescription, error) {
	if len(doc.Steps) == 0 {
		return nil, errors.New("workout_doc has no steps")
	}
	if err := prevalidateProgressionDoc(doc); err != nil {
		return nil, err
	}
	validated := workoutdoc.ValidateDoc(doc)
	if len(validated.Errors) > 0 {
		return nil, errors.New("workout_doc validation failed")
	}
	warnings := make([]analysis.PrescriptionWarning, 0, len(validated.Warnings))
	for _, warning := range validated.Warnings {
		warnings = append(warnings, analysis.PrescriptionWarning{Code: warning.Code, Message: warning.Message, StepIndex: warning.StepIndex})
	}
	intervalsOut := []analysis.PrescriptionInterval{}
	saturated := false
	for _, step := range doc.Steps {
		if !appendProgressionStep(&intervalsOut, step, 0, &saturated) {
			break
		}
	}
	if saturated {
		return &analysis.WorkoutPrescription{Source: source, Intervals: intervalsOut, Saturated: true, TotalCount: maxWorkoutProgressionIntervals + 1, Warnings: warnings}, nil
	}
	return &analysis.WorkoutPrescription{Source: source, Intervals: intervalsOut, TotalCount: len(intervalsOut), Warnings: warnings}, nil
}

func prevalidateProgressionDoc(doc workoutdoc.WorkoutDoc) error {
	for _, step := range doc.Steps {
		if err := prevalidateProgressionStep(step, false); err != nil {
			return err
		}
	}
	return nil
}

func prevalidateProgressionStep(step workoutdoc.Step, inRepeat bool) error {
	if step.Duration < 0 {
		return errors.New("step duration must not be negative")
	}
	if step.Distance != nil && (!finiteValue(step.Distance.Value) || step.Distance.Value < 0 || progressionDistanceFactor(step.Distance.Unit) == 0) {
		return errors.New("step distance is invalid or has an unsupported unit")
	}
	for _, target := range []*workoutdoc.Target{step.Power, step.HR, step.Pace, step.RPE, step.Cadence} {
		if target == nil {
			continue
		}
		for _, value := range []*float64{target.Value, target.Min, target.Max, target.Start, target.End} {
			if value != nil && !finiteValue(*value) {
				return errors.New("target contains a non-finite value")
			}
		}
		if target.Min != nil && target.Max != nil && *target.Min > *target.Max {
			return errors.New("target range is inverted")
		}
	}
	if len(step.Steps) > 0 {
		if inRepeat || step.Reps < 1 || step.Reps > 100 {
			return errors.New("unsupported repeat cardinality or nested repeat")
		}
		for _, child := range step.Steps {
			if child.Reps != 0 || len(child.Steps) > 0 {
				return errors.New("nested repeat is unsupported")
			}
			if err := prevalidateProgressionStep(child, true); err != nil {
				return err
			}
		}
		return nil
	}
	if step.Reps != 0 {
		return errors.New("simple step must have zero reps")
	}
	return nil
}

func appendProgressionStep(out *[]analysis.PrescriptionInterval, step workoutdoc.Step, _ int, saturated *bool) bool {
	if len(step.Steps) == 0 {
		if len(*out) >= maxWorkoutProgressionIntervals {
			*saturated = true
			return false
		}
		*out = append(*out, progressionPrescriptionInterval(step))
		return true
	}
	for repetition := 0; repetition < step.Reps; repetition++ {
		for _, child := range step.Steps {
			if !appendProgressionStep(out, child, 0, saturated) {
				return false
			}
		}
	}
	return true
}

func progressionPrescriptionInterval(step workoutdoc.Step) analysis.PrescriptionInterval {
	interval := analysis.PrescriptionInterval{Recovery: progressionRecoveryDescription(step.Description), Ramp: step.Ramp, Freeride: step.Freeride, Target: progressionTarget(step)}
	if step.Duration > 0 {
		value := float64(step.Duration)
		interval.DurationSeconds = &value
	}
	if step.Distance != nil && finiteValue(step.Distance.Value) {
		factor := progressionDistanceFactor(step.Distance.Unit)
		if factor > 0 {
			value := step.Distance.Value * factor
			interval.DistanceMeters = &value
		}
	}
	return interval
}

func progressionTarget(step workoutdoc.Step) analysis.WorkoutTarget {
	candidates := []struct {
		kind   string
		target *workoutdoc.Target
	}{
		{kind: "power", target: step.Power},
		{kind: "heart_rate", target: step.HR},
		{kind: "pace", target: step.Pace},
		{kind: "rpe", target: step.RPE},
		{kind: "cadence", target: step.Cadence},
	}
	for _, candidate := range candidates {
		if candidate.target != nil {
			return analysis.WorkoutTarget{Kind: candidate.kind, Unit: strings.TrimSpace(candidate.target.Units), Value: finiteFloatPointer(candidate.target.Value), Min: finiteFloatPointer(candidate.target.Min), Max: finiteFloatPointer(candidate.target.Max), Start: finiteFloatPointer(candidate.target.Start), End: finiteFloatPointer(candidate.target.End), Text: strings.TrimSpace(candidate.target.Text)}
		}
	}
	return analysis.WorkoutTarget{}
}

func progressionCompletedFromDTO(dto intervals.IntervalsDTO) *analysis.WorkoutCompleted {
	intervalSource := analysis.InferIntervalSource(analysis.IntervalSourceInput{Raw: dto.Raw, Intervals: progressionSourceIntervals(dto.ICUIntervals), Groups: progressionSourceGroups(dto.ICUGroups)})
	completed := &analysis.WorkoutCompleted{IntervalSource: intervalSource.Source, AutoLapSuspected: intervalSource.AutoLapSuspected}
	if intervalSource.Source != analysis.IntervalSourceStructuredWorkout {
		completed.IntervalSourceCaveat = "interval rows are not verified structured workout execution"
	}
	for _, interval := range dto.ICUIntervals {
		completed.Intervals = append(completed.Intervals, analysis.CompletedInterval{DurationSeconds: progressionIntervalValue(interval.Duration, interval.Raw, "moving_time", "duration", "elapsed_time"), DistanceMeters: progressionIntervalValue(interval.Distance, interval.Raw, "distance"), Kind: strings.ToLower(strings.TrimSpace(progressionIntervalString(interval.Type, interval.Raw, "type", "kind"))), ObservedPowerWatts: progressionIntervalValue(interval.AveragePower, interval.Raw, "average_power", "average_watts", "icu_average_watts"), ObservedHeartRateBPM: progressionIntervalValue(interval.AverageHR, interval.Raw, "average_hr", "average_heartrate"), ObservedPace: progressionIntervalValue(interval.Pace, interval.Raw, "pace"), ObservedPaceUnit: progressionIntervalString(interval.Unit, interval.Raw, "pace_unit", "pace_units", "unit")})
	}
	return completed
}

func progressionSourceIntervals(rows []intervals.ActivityInterval) []analysis.IntervalSourceInterval {
	out := make([]analysis.IntervalSourceInterval, 0, len(rows))
	for _, row := range rows {
		label := progressionIntervalString(nil, row.Raw, "label")
		out = append(out, analysis.IntervalSourceInterval{Name: progressionStringValue(row.Name), Type: progressionStringValue(row.Type), Label: label, Raw: row.Raw, StartIndex: row.StartIndex, EndIndex: row.EndIndex, StartDistance: row.StartDistance, EndDistance: row.EndDistance, Distance: row.Distance, Duration: row.Duration})
	}
	return out
}

func progressionSourceGroups(rows []intervals.IntervalGroup) []analysis.IntervalSourceGroup {
	out := make([]analysis.IntervalSourceGroup, 0, len(rows))
	for _, row := range rows {
		out = append(out, analysis.IntervalSourceGroup{Name: progressionStringValue(row.Name), Type: progressionStringValue(row.Type), Raw: row.Raw, StartIndex: row.StartIndex, EndIndex: row.EndIndex})
	}
	return out
}

func progressionIntervalValue(primary *float64, raw map[string]any, aliases ...string) *float64 {
	if value := finiteFloatPointer(primary); value != nil {
		return value
	}
	for _, alias := range aliases {
		if value, ok := raw[alias].(float64); ok && finiteValue(value) {
			return &value
		}
	}
	return nil
}

func progressionIntervalString(primary *string, raw map[string]any, aliases ...string) string {
	if value := strings.TrimSpace(progressionStringValue(primary)); value != "" {
		return value
	}
	for _, alias := range aliases {
		if value, ok := raw[alias].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func progressionDuration(activity intervals.Activity) (*float64, string) {
	if activity.MovingTime != nil && *activity.MovingTime >= 0 {
		value := float64(*activity.MovingTime)
		return &value, "moving_time"
	}
	if activity.ElapsedTime != nil && *activity.ElapsedTime >= 0 {
		value := float64(*activity.ElapsedTime)
		return &value, "elapsed_time"
	}
	return nil, ""
}

func progressionEventID(activity intervals.Activity) (string, bool) {
	values := []string{}
	for _, key := range []string{"paired_event_id", "event_id", "calendar_event_id", "icu_event_id"} {
		if value := progressionRawID(activity.Raw, key); value != "" {
			values = append(values, value)
		}
	}
	if len(values) == 0 {
		return "", false
	}
	for _, value := range values[1:] {
		if value != values[0] {
			return "", true
		}
	}
	return values[0], false
}

func progressionRawID(raw map[string]any, key string) string {
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		if finiteValue(typed) && typed == math.Trunc(typed) {
			return strconv.FormatInt(int64(typed), 10)
		}
	case json.Number:
		return strings.TrimSpace(string(typed))
	case int:
		return strconv.Itoa(typed)
	}
	return ""
}

func progressionRawFloat(raw map[string]any, key string) *float64 {
	value, ok := raw[key].(float64)
	if !ok || !finiteValue(value) {
		return nil
	}
	return &value
}

func activityType(activity intervals.Activity) string {
	return strings.TrimSpace(progressionStringValue(activity.Type))
}

func isRestrictedProgressionActivity(activity intervals.Activity) bool {
	return isStravaBlocked(activity)
}

func progressionDistanceFactor(unit string) float64 {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "m", "mtr", "meter", "meters":
		return 1
	case "km":
		return 1000
	case "mi":
		return 1609.344
	case "yd", "yrd", "yard", "yards":
		return 0.9144
	default:
		return 0
	}
}

func progressionRecoveryDescription(description string) bool {
	normalized := strings.ToLower(strings.TrimSpace(description))
	for _, marker := range []string{"recovery", "recover", "rest", "cooldown", "cool down"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func finiteFloatPointer(value *float64) *float64 {
	if value == nil || !finiteValue(*value) {
		return nil
	}
	return value
}

func finiteIntAsFloat(value *int) *float64 {
	if value == nil {
		return nil
	}
	out := float64(*value)
	return &out
}

func finiteValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func progressionStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringValueAny(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func isJSONBool(value json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(value))
	return trimmed == "true" || trimmed == "false"
}

func isJSONString(value json.RawMessage) bool {
	return len(value) > 0 && strings.TrimSpace(string(value))[0] == '"'
}

func progressionHasInvalidDate(values []string) bool {
	for _, value := range values {
		if value == "invalid_activity_date" {
			return true
		}
	}
	return false
}

func progressionBoolPtr(value bool) *bool { return &value }
