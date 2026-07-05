package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

func TestValidateWorkoutMetadataAdvertisesPreviewPreflight(t *testing.T) {
	t.Parallel()

	tool := newValidateWorkoutTool("test", false)
	contractText := normalizeContractText(tool.Description + "\n" + validateWorkoutInputSchema()["description"].(string) + "\n" + validateWorkoutOutputSchema()["description"].(string))
	for _, want := range []string{"read-only preflight", "structured workout changes", "preview total duration", "key steps", "target intensities", "load/distance/time changes"} {
		if !strings.Contains(contractText, want) {
			t.Fatalf("validate_workout contract missing %q:\n%s", want, contractText)
		}
	}
}

func TestValidateWorkoutDescriptionOnlyValidDSL(t *testing.T) {
	t.Parallel()
	resp := runValidateWorkout(t, `{"description":"Easy spin\n- 30m 60-70% HR"}`)
	if !resp.Valid {
		t.Fatalf("Valid = false, want true; errors=%+v", resp.Errors)
	}
	if len(resp.Errors) != 0 {
		t.Fatalf("errors = %+v, want none", resp.Errors)
	}
	if !strings.Contains(resp.CanonicalDSL, "- 30m 60-70% HR") {
		t.Fatalf("CanonicalDSL missing serialized step: %q", resp.CanonicalDSL)
	}
	if !strings.Contains(resp.CanonicalDSL, "Easy spin") {
		t.Fatalf("CanonicalDSL missing prose: %q", resp.CanonicalDSL)
	}
	if resp.Stats.StructuredSteps != 1 {
		t.Fatalf("Stats.StructuredSteps = %d, want 1", resp.Stats.StructuredSteps)
	}
	if resp.Stats.EstimatedDurationSeconds == nil || *resp.Stats.EstimatedDurationSeconds != 1800 {
		t.Fatalf("Stats.EstimatedDurationSeconds = %v, want 1800", resp.Stats.EstimatedDurationSeconds)
	}
}

func TestValidateWorkoutDescriptionOnlyMAmbiguityWarning(t *testing.T) {
	t.Parallel()
	resp := runValidateWorkout(t, `{"description":"- 400m 90%"}`)
	if !resp.Valid {
		t.Fatalf("Valid = false, want true; errors=%+v", resp.Errors)
	}
	codes := diagCodes(resp.Warnings)
	if !diagContains(codes, "M_AMBIGUITY") {
		t.Fatalf("expected M_AMBIGUITY warning; got %+v", codes)
	}
}

func TestValidateWorkoutStepsOnlyHappyPath(t *testing.T) {
	t.Parallel()
	payload := `{"workout_doc":{"steps":[{"description":"Warm up","duration":600,"power":{"value":60,"units":"PERCENT_FTP"}},{"reps":2,"steps":[{"duration":300,"power":{"min":95,"max":100,"units":"PERCENT_FTP"}},{"description":"Recovery","duration":120,"power":{"value":50,"units":"PERCENT_FTP"}}]}]}}`
	resp := runValidateWorkout(t, payload)
	if !resp.Valid {
		t.Fatalf("Valid = false; errors=%+v", resp.Errors)
	}
	if !resp.Stats.HasRepeats {
		t.Fatalf("Stats.HasRepeats = false, want true")
	}
	if resp.Stats.StructuredSteps != 2 {
		t.Fatalf("Stats.StructuredSteps = %d, want 2", resp.Stats.StructuredSteps)
	}
	if resp.Stats.EstimatedDurationSeconds == nil || *resp.Stats.EstimatedDurationSeconds != 600+2*(300+120) {
		t.Fatalf("Stats.EstimatedDurationSeconds wrong: %v", resp.Stats.EstimatedDurationSeconds)
	}
	if !strings.Contains(resp.CanonicalDSL, "2x") {
		t.Fatalf("CanonicalDSL missing repeat header: %q", resp.CanonicalDSL)
	}
}

func TestValidateWorkoutRejectsFractionalPercentFTP(t *testing.T) {
	t.Parallel()
	payload := `{"workout_doc":{"steps":[{"description":"Threshold","duration":600,"power":{"value":0.95,"units":"PERCENT_FTP"}}]}}`
	resp := runValidateWorkout(t, payload)
	if resp.Valid {
		t.Fatalf("Valid = true, want false; canonical=%q warnings=%+v", resp.CanonicalDSL, resp.Warnings)
	}
	if resp.CanonicalDSL != "" || strings.Contains(resp.CanonicalDSL, "0.95%") {
		t.Fatalf("CanonicalDSL = %q, want no silently serialized fractional percent", resp.CanonicalDSL)
	}
	if !diagContains(diagCodes(resp.Errors), "UNSUPPORTED_STEP") {
		t.Fatalf("expected UNSUPPORTED_STEP error; got %+v", resp.Errors)
	}
	if !strings.Contains(resp.Errors[0].Message, "percent points") {
		t.Fatalf("error = %q, want percent-point guidance", resp.Errors[0].Message)
	}
}

func TestValidateWorkoutBothSetMergesWithSentinel(t *testing.T) {
	t.Parallel()
	prose := "Threshold day.\n" + workoutdoc.StepsSentinel + "\nCool down well."
	payload := mustMarshalArgs(t, map[string]any{
		"description": prose,
		"workout_doc": map[string]any{
			"steps": []any{
				map[string]any{"description": "Warm up", "duration": 600, "power": map[string]any{"value": 60, "units": "PERCENT_FTP"}},
			},
		},
	})
	resp := runValidateWorkout(t, payload)
	if !resp.Valid {
		t.Fatalf("Valid = false; errors=%+v", resp.Errors)
	}
	if !strings.HasPrefix(resp.CanonicalDSL, "Threshold day.\n- Warm up 10m 60%") {
		t.Fatalf("merge did not honour sentinel placement: %q", resp.CanonicalDSL)
	}
	if !strings.HasSuffix(resp.CanonicalDSL, "\nCool down well.") {
		t.Fatalf("trailing prose lost: %q", resp.CanonicalDSL)
	}
}

func TestValidateWorkoutNestedRepeatError(t *testing.T) {
	t.Parallel()
	payload := `{"workout_doc":{"steps":[{"reps":2,"steps":[{"reps":2,"steps":[{"duration":60,"rpe":{"value":2,"units":"RPE"}}]}]}]}}`
	resp := runValidateWorkout(t, payload)
	if resp.Valid {
		t.Fatalf("Valid = true, want false")
	}
	codes := diagCodes(resp.Errors)
	if !diagContains(codes, "NESTED_REPEAT") && !diagContains(codes, "UNSUPPORTED_STEP") {
		t.Fatalf("expected NESTED_REPEAT or UNSUPPORTED_STEP error; got %+v", codes)
	}
}

func TestValidateWorkoutUnsupportedStepError(t *testing.T) {
	t.Parallel()
	payload := `{"workout_doc":{"steps":[{"duration":600,"freeride":true,"ramp":true}]}}`
	resp := runValidateWorkout(t, payload)
	if resp.Valid {
		t.Fatalf("Valid = true, want false")
	}
	codes := diagCodes(resp.Errors)
	if !diagContains(codes, "UNSUPPORTED_STEP") {
		t.Fatalf("expected UNSUPPORTED_STEP; got %+v", codes)
	}
}

func TestValidateWorkoutRejectsStructuralTokenInStepDescription(t *testing.T) {
	t.Parallel()
	payload := `{"workout_doc":{"steps":[{"description":"Endurance 2h15m","duration":8100,"power":{"value":60,"units":"PERCENT_FTP"}}]}}`
	resp := runValidateWorkout(t, payload)
	if resp.Valid {
		t.Fatalf("Valid = true, want false")
	}
	if !diagContains(diagCodes(resp.Errors), "STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION") {
		t.Fatalf("expected STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION; got %+v", resp.Errors)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("errors = %+v, want one structural-token diagnostic", resp.Errors)
	}
	if resp.Errors[0].StepIndex == nil || *resp.Errors[0].StepIndex != 0 {
		t.Fatalf("step_index = %#v, want 0", resp.Errors[0].StepIndex)
	}
	if !strings.Contains(resp.Errors[0].Message, "duration/distance in structured fields") {
		t.Fatalf("message = %q, want structured-field guidance", resp.Errors[0].Message)
	}
}

func TestValidateWorkoutProsePassesThroughVerbatim(t *testing.T) {
	t.Parallel()
	prose := "# Heading the parser does not recognize\nFree-form coaching note: keep cadence above 85.\n## Another heading"
	payload := mustMarshalArgs(t, map[string]any{"description": prose})
	resp := runValidateWorkout(t, payload)
	if !resp.Valid {
		t.Fatalf("Valid = false; errors=%+v", resp.Errors)
	}
	if resp.CanonicalDSL != prose {
		t.Fatalf("prose round-trip mismatch\n--- got ---\n%q\n--- want ---\n%q", resp.CanonicalDSL, prose)
	}
	if len(resp.Errors) != 0 {
		t.Fatalf("prose passthrough should never error; got %+v", resp.Errors)
	}
}

func TestValidateWorkoutSportSpecificGeneratedDescriptionsAvoidCyclingWords(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		payload string
	}{
		{
			name:    "run workout",
			payload: `{"workout_doc":{"steps":[{"description":"Run warmup","duration":600,"pace":{"value":85,"units":"PERCENT_THRESHOLD"}},{"description":"Run intervals","duration":1200,"pace":{"min":95,"max":100,"units":"PERCENT_THRESHOLD"}},{"description":"Run cooldown","duration":600,"rpe":{"value":2,"units":"RPE"}}]}}`,
		},
		{
			name:    "swim workout",
			payload: `{"workout_doc":{"steps":[{"description":"Swim warmup","distance":{"value":200,"unit":"yards"},"pace":{"value":85,"units":"PERCENT_THRESHOLD"}},{"description":"Swim main","distance":{"value":800,"unit":"yards"},"pace":{"value":95,"units":"PERCENT_THRESHOLD"}},{"description":"Swim cooldown","distance":{"value":100,"unit":"yards"},"rpe":{"value":2,"units":"RPE"}}]}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resp := runValidateWorkout(t, tc.payload)
			if !resp.Valid {
				t.Fatalf("Valid = false; errors=%+v", resp.Errors)
			}
			lower := strings.ToLower(resp.CanonicalDSL)
			for _, forbidden := range []string{"ride", "spin"} {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("CanonicalDSL = %q contains cycling-only word %q", resp.CanonicalDSL, forbidden)
				}
			}
		})
	}
}

func TestValidateWorkoutKeepsUserProvidedCyclingProse(t *testing.T) {
	t.Parallel()

	resp := runValidateWorkout(t, `{"description":"Coach note: relaxed spin if legs are heavy.","workout_doc":{"steps":[{"description":"Run easy","duration":1800,"pace":{"value":80,"units":"PERCENT_THRESHOLD"}}]}}`)
	if !resp.Valid {
		t.Fatalf("Valid = false; errors=%+v", resp.Errors)
	}
	if !strings.Contains(resp.CanonicalDSL, "relaxed spin") {
		t.Fatalf("CanonicalDSL = %q, want user-provided prose preserved", resp.CanonicalDSL)
	}
}

func TestValidateWorkoutEmptyInputErrors(t *testing.T) {
	t.Parallel()
	tool := newValidateWorkoutTool("test", false)
	_, err := tool.Handler(context.Background(), Request{Name: validateWorkoutName, Arguments: json.RawMessage(`{}`)})
	if err == nil {
		t.Fatal("expected error for empty input, got nil")
	}
	if _, ok := PublicErrorMessage(err); !ok {
		t.Fatalf("expected UserError-style public message, got %v", err)
	}
}

func TestValidateWorkoutBothSetEmitsOverrideWarning(t *testing.T) {
	t.Parallel()
	payload := `{"description":"- 10m 60%","workout_doc":{"steps":[{"duration":600,"power":{"value":60,"units":"PERCENT_FTP"}}]}}`
	resp := runValidateWorkout(t, payload)
	if !resp.Valid {
		t.Fatalf("Valid = false; errors=%+v", resp.Errors)
	}
	if !diagContains(diagCodes(resp.Warnings), "STEP_SOURCES_OVERRIDDEN") {
		t.Fatalf("expected STEP_SOURCES_OVERRIDDEN warning; got %+v", resp.Warnings)
	}
	if !resp.Meta.StepSourcesOverridden {
		t.Fatalf("Meta.StepSourcesOverridden = false, want true")
	}
}

func runValidateWorkout(t *testing.T, args string) validateWorkoutResponse {
	t.Helper()
	tool := newValidateWorkoutTool("test", false)
	result, err := tool.Handler(context.Background(), Request{Name: validateWorkoutName, Arguments: json.RawMessage(args)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if result.StructuredContent == nil {
		t.Fatal("StructuredContent is nil")
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("Marshal(StructuredContent) error = %v", err)
	}
	var resp validateWorkoutResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("Unmarshal response error = %v; raw=%s", err, string(raw))
	}
	return resp
}

func mustMarshalArgs(t *testing.T, args map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("Marshal args error = %v", err)
	}
	return string(raw)
}

func diagCodes(diags []validateWorkoutDiagnostic) []string {
	codes := make([]string, 0, len(diags))
	for _, d := range diags {
		codes = append(codes, d.Code)
	}
	return codes
}

func diagContains(codes []string, needle string) bool {
	for _, code := range codes {
		if code == needle {
			return true
		}
	}
	return false
}
