package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

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
