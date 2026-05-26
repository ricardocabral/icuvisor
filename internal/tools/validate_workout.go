package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	validateWorkoutName                    = "validate_workout"
	validateWorkoutDescription             = "Validate an intervals.icu workout description, a structured workout_doc, or both, and return the canonical merged DSL that icuvisor would submit on a write. Read-only and athlete-independent; never hits the network and never rejects prose. Only malformed structured-step lines (lines starting with '- ' or a 'Nx' repeat header) surface as PARSE_ERROR. Free-text headers, comments, and notes pass through verbatim. Syntax reference: icuvisor://workout-syntax."
	invalidValidateWorkoutArgumentsMessage = "invalid validate_workout arguments; provide at least one of description (string) or workout_doc (object with steps[])"
)

type validateWorkoutRequest struct {
	Description *string                `json:"description,omitempty"`
	WorkoutDoc  *workoutdoc.WorkoutDoc `json:"workout_doc,omitempty"`
}

type validateWorkoutDiagnostic struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	StepIndex *int   `json:"step_index"`
	Line      *int   `json:"line"`
}

type validateWorkoutStats struct {
	StructuredSteps          int  `json:"structured_steps"`
	ProseLines               int  `json:"prose_lines"`
	HasRepeats               bool `json:"has_repeats"`
	EstimatedDurationSeconds *int `json:"estimated_duration_seconds"`
}

type validateWorkoutResponse struct {
	Valid        bool                        `json:"valid"`
	CanonicalDSL string                      `json:"canonical_dsl"`
	Errors       []validateWorkoutDiagnostic `json:"errors"`
	Warnings     []validateWorkoutDiagnostic `json:"warnings"`
	Stats        validateWorkoutStats        `json:"stats"`
	Meta         validateWorkoutMeta         `json:"_meta"`
}

type validateWorkoutMeta struct {
	Operation             string `json:"operation"`
	DescriptionStepsFound bool   `json:"description_steps_found"`
	WorkoutDocProvided    bool   `json:"workout_doc_provided"`
	StepsSentinelEmbedded bool   `json:"steps_sentinel_embedded"`
	MergeRule             string `json:"merge_rule"`
	StepSourcesOverridden bool   `json:"step_sources_overridden,omitempty"`
}

func newValidateWorkoutTool(version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{
		Name:         validateWorkoutName,
		Description:  validateWorkoutDescription,
		InputSchema:  validateWorkoutInputSchema(),
		OutputSchema: validateWorkoutOutputSchema(),
		Requirement:  RequirementRead,
		Handler:      validateWorkoutHandler(version, debugMetadata, shapeCfg),
	})
}

func validateWorkoutHandler(version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeValidateWorkoutRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidValidateWorkoutArgumentsMessage, err)
		}
		payload := validateWorkout(args)
		return encodeShaped(payload, false, nil, version, debugMetadata, validateWorkoutName, response.UnitSystemMetric, shapeCfg)
	}
}

func decodeValidateWorkoutRequest(raw json.RawMessage) (validateWorkoutRequest, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return validateWorkoutRequest{}, errors.New("arguments must be a JSON object")
	}
	args, err := DecodeStrict[validateWorkoutRequest](raw)
	if err != nil {
		return validateWorkoutRequest{}, err
	}
	if args.Description == nil && args.WorkoutDoc == nil {
		return validateWorkoutRequest{}, errors.New("provide description, workout_doc, or both")
	}
	return args, nil
}

func validateWorkout(args validateWorkoutRequest) validateWorkoutResponse {
	resp := validateWorkoutResponse{
		Errors:   []validateWorkoutDiagnostic{},
		Warnings: []validateWorkoutDiagnostic{},
		Meta: validateWorkoutMeta{
			Operation:          "validate",
			WorkoutDocProvided: args.WorkoutDoc != nil,
			MergeRule:          "prose lines retain order verbatim; structured steps replace a '" + workoutdoc.StepsSentinel + "' line when present, otherwise are appended after prose with a blank-line separator",
		},
	}

	description := ""
	if args.Description != nil {
		description = *args.Description
	}
	descResult := workoutdoc.ValidateDescription(description)
	resp.Errors = appendDiagnostics(resp.Errors, descResult.Errors)
	resp.Warnings = appendDiagnostics(resp.Warnings, descResult.Warnings)
	resp.Meta.DescriptionStepsFound = descResult.StructuredStepLines > 0

	var effectiveDoc workoutdoc.WorkoutDoc
	switch {
	case args.WorkoutDoc != nil:
		effectiveDoc = *args.WorkoutDoc
		docResult := workoutdoc.ValidateDoc(effectiveDoc)
		resp.Errors = appendDiagnostics(resp.Errors, docResult.Errors)
		resp.Warnings = appendDiagnostics(resp.Warnings, docResult.Warnings)
		if descResult.StructuredStepLines > 0 {
			resp.Meta.StepSourcesOverridden = true
			resp.Warnings = append(resp.Warnings, validateWorkoutDiagnostic{
				Code:    "STEP_SOURCES_OVERRIDDEN",
				Message: "description contains structured step lines and workout_doc was also provided; workout_doc wins for the canonical merged DSL",
			})
		}
	default:
		effectiveDoc = descResult.Doc
	}

	prose := descResult.Prose
	if args.WorkoutDoc != nil && descResult.StructuredStepLines == 0 && args.Description != nil {
		prose = strings.ReplaceAll(*args.Description, "\r\n", "\n")
	}
	resp.Meta.StepsSentinelEmbedded = strings.Contains(prose, workoutdoc.StepsSentinel)

	if len(resp.Errors) == 0 {
		merged, err := workoutdoc.MergeDescription(prose, effectiveDoc)
		if err != nil {
			resp.Errors = append(resp.Errors, validateWorkoutDiagnostic{
				Code:    serializeErrorCode(err),
				Message: err.Error(),
			})
		} else {
			resp.CanonicalDSL = merged
		}
	}

	resp.Stats = validateWorkoutStats{
		StructuredSteps: len(effectiveDoc.Steps),
		ProseLines:      countProseLines(prose),
		HasRepeats:      docHasRepeats(effectiveDoc),
	}
	if total := workoutdoc.EstimateDurationSeconds(effectiveDoc); total > 0 {
		resp.Stats.EstimatedDurationSeconds = &total
	}
	resp.Valid = len(resp.Errors) == 0
	return resp
}

func appendDiagnostics(into []validateWorkoutDiagnostic, from []workoutdoc.Diagnostic) []validateWorkoutDiagnostic {
	for _, diag := range from {
		into = append(into, validateWorkoutDiagnostic{Code: diag.Code, Message: diag.Message, StepIndex: diag.StepIndex, Line: diag.Line})
	}
	return into
}

func serializeErrorCode(err error) string {
	var structural *workoutdoc.StructuralTokenInDescriptionError
	if errors.As(err, &structural) {
		return "STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION"
	}
	var unsupported *workoutdoc.UnsupportedStepError
	if errors.As(err, &unsupported) {
		return "UNSUPPORTED_STEP"
	}
	return "SERIALIZE_ERROR"
}

func countProseLines(prose string) int {
	if prose == "" {
		return 0
	}
	count := 0
	for _, line := range strings.Split(prose, "\n") {
		if line == workoutdoc.StepsSentinel {
			continue
		}
		count++
	}
	return count
}

func docHasRepeats(doc workoutdoc.WorkoutDoc) bool {
	for _, step := range doc.Steps {
		if step.Reps > 0 || len(step.Steps) > 0 {
			return true
		}
	}
	return false
}

func validateWorkoutInputSchema() map[string]any {
	examples := validateWorkoutInputExamples()
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"description":          "At least one of description or workout_doc is required. Both may be set in the same call: structured steps from workout_doc are merged into the prose from description following the merge rule reported in _meta.merge_rule.",
		"examples":             examples,
		"input_examples":       examples,
		"properties": map[string]any{
			"description": map[string]any{
				"type":        "string",
				"description": "Optional free-text intervals.icu description. Prose, headers, comments, and any line the DSL parser does not recognize as a structured step pass through verbatim. Embed the sentinel '" + workoutdoc.StepsSentinel + "' on its own line to control where structured steps from workout_doc are inserted in the merged canonical DSL. Structured-step lines start with '- ' or are 'Nx' repeat headers — those are validated; malformed ones surface as PARSE_ERROR. Syntax reference: icuvisor://workout-syntax.",
			},
			"workout_doc": map[string]any{
				"type":        "object",
				"description": "Optional structured WorkoutDoc with a steps[] array. Validated via the same Serialize path that write tools use. In each structured step, description is a label/comment only: do not include duration or distance tokens there; use duration seconds or distance instead. Mutually compatible with description; when both contain structured steps, workout_doc wins and the response surfaces a STEP_SOURCES_OVERRIDDEN warning.",
			},
		},
	}
}

func validateWorkoutInputExamples() []map[string]any {
	return []map[string]any{
		{
			"description": "Easy Zone 2 spin.\n- 60m 60-70% HR",
		},
		{
			"description": "Threshold day. Keep cadence above 85.\n" + workoutdoc.StepsSentinel + "\nCool down well.",
			"workout_doc": map[string]any{
				"steps": []any{
					map[string]any{"description": "Warm up", "duration": 600, "power": map[string]any{"value": 60, "units": "PERCENT_FTP"}},
					map[string]any{"reps": 3, "steps": []any{
						map[string]any{"duration": 480, "power": map[string]any{"min": 95, "max": 100, "units": "PERCENT_FTP"}},
						map[string]any{"description": "Recovery", "duration": 240, "power": map[string]any{"value": 50, "units": "PERCENT_FTP"}},
					}},
					map[string]any{"description": "Cool down", "duration": 600, "power": map[string]any{"value": 50, "units": "PERCENT_FTP"}},
				},
			},
		},
	}
}

func validateWorkoutOutputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"description":          "Validation result with diagnostics split into errors/warnings, the canonical merged DSL that would be submitted on a write, and aggregate stats. valid=true iff errors is empty. Validation is never a precondition for writes; prose always passes through.",
	}
}
