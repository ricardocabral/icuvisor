package workoutdoc

import (
	"strings"
	"testing"
)

func TestValidateDescriptionProsePassthrough(t *testing.T) {
	t.Parallel()
	prose := "Easy aerobic morning ride.\n# A heading the parser ignores.\nKeep cadence high."
	got := ValidateDescription(prose)
	if got.Prose != prose {
		t.Fatalf("Prose mismatch\n--- got ---\n%q\n--- want ---\n%q", got.Prose, prose)
	}
	if len(got.Errors) != 0 {
		t.Fatalf("expected no errors for prose-only input, got %+v", got.Errors)
	}
	if len(got.Doc.Steps) != 0 {
		t.Fatalf("expected no parsed steps, got %d", len(got.Doc.Steps))
	}
	if got.StructuredStepLines != 0 {
		t.Fatalf("StructuredStepLines = %d, want 0", got.StructuredStepLines)
	}
}

func TestValidateDescriptionParsesStepBlock(t *testing.T) {
	t.Parallel()
	input := "Warmup note.\n- 10m 60%\n- 20m 95-100%\nCooldown note."
	got := ValidateDescription(input)
	if len(got.Errors) != 0 {
		t.Fatalf("expected no errors, got %+v", got.Errors)
	}
	if len(got.Doc.Steps) != 2 {
		t.Fatalf("Doc.Steps len = %d, want 2", len(got.Doc.Steps))
	}
	if !strings.Contains(got.Prose, StepsSentinel) {
		t.Fatalf("Prose should contain steps sentinel; got %q", got.Prose)
	}
	if !strings.Contains(got.Prose, "Warmup note.") || !strings.Contains(got.Prose, "Cooldown note.") {
		t.Fatalf("Prose should preserve surrounding notes; got %q", got.Prose)
	}
}

func TestValidateDescriptionMAmbiguityWarning(t *testing.T) {
	t.Parallel()
	got := ValidateDescription("- 400m 90%")
	if len(got.Warnings) == 0 {
		t.Fatal("expected M_AMBIGUITY warning, got none")
	}
	if got.Warnings[0].Code != "M_AMBIGUITY" {
		t.Fatalf("Warning code = %q, want M_AMBIGUITY", got.Warnings[0].Code)
	}
}

func TestValidateDescriptionMalformedStepLine(t *testing.T) {
	t.Parallel()
	got := ValidateDescription("- not a valid step token")
	if len(got.Errors) == 0 {
		t.Fatal("expected PARSE_ERROR, got none")
	}
	if got.Errors[0].Code != "PARSE_ERROR" {
		t.Fatalf("Error code = %q, want PARSE_ERROR", got.Errors[0].Code)
	}
}

func TestValidateDescriptionMalformedRepeatHeaders(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"-3 x\n  - 1m RPE 2", "- 3x\n  - 1m RPE 2"} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			got := ValidateDescription(input)
			if got.StructuredStepLines == 0 {
				t.Fatalf("StructuredStepLines = 0, want malformed repeat-like line treated as structured")
			}
			if len(got.Errors) == 0 {
				t.Fatal("expected PARSE_ERROR, got none")
			}
			if got.Errors[0].Code != "PARSE_ERROR" {
				t.Fatalf("Error code = %q, want PARSE_ERROR", got.Errors[0].Code)
			}
		})
	}
}

func TestValidateDescriptionMultipleStepBlocksReported(t *testing.T) {
	t.Parallel()
	got := ValidateDescription("- 10m 60%\nProse interlude.\n- 20m 70%")
	if len(got.Errors) == 0 {
		t.Fatal("expected error for multi-block step regions, got none")
	}
	if got.Errors[0].Code != "PARSE_ERROR" {
		t.Fatalf("Error code = %q, want PARSE_ERROR", got.Errors[0].Code)
	}
}

func TestValidateDocMixedPrimaryTargets(t *testing.T) {
	t.Parallel()
	v := ptrFloat(200)
	hr := ptrFloat(150)
	doc := WorkoutDoc{Steps: []Step{{Duration: 600, Power: &Target{Value: v, Units: "WATTS"}, HR: &Target{Value: hr, Units: "BPM"}}}}
	got := ValidateDoc(doc)
	codes := warningCodes(got.Warnings)
	if !diagListContains(codes, "MIXED_PRIMARY_TARGETS") {
		t.Fatalf("expected MIXED_PRIMARY_TARGETS, got warnings %+v", codes)
	}
}

func TestValidateDocMissingPrimaryTarget(t *testing.T) {
	t.Parallel()
	doc := WorkoutDoc{Steps: []Step{{Duration: 600, Cadence: &Target{Value: ptrFloat(90), Units: "RPM"}}}}
	got := ValidateDoc(doc)
	codes := warningCodes(got.Warnings)
	if !diagListContains(codes, "MISSING_PRIMARY_TARGET") {
		t.Fatalf("expected MISSING_PRIMARY_TARGET, got warnings %+v", codes)
	}
}

func TestValidateDocNestedRepeatReportsError(t *testing.T) {
	t.Parallel()
	inner := Step{Reps: 2, Steps: []Step{{Duration: 60, RPE: &Target{Value: ptrFloat(2), Units: "RPE"}}}}
	doc := WorkoutDoc{Steps: []Step{{Reps: 2, Steps: []Step{inner}}}}
	got := ValidateDoc(doc)
	codes := errorCodes(got.Errors)
	if !diagListContains(codes, "NESTED_REPEAT") {
		t.Fatalf("expected NESTED_REPEAT error, got %+v", codes)
	}
}

func TestValidateDocUnsupportedStepError(t *testing.T) {
	t.Parallel()
	doc := WorkoutDoc{Steps: []Step{{Duration: 600, Freeride: true, Ramp: true}}}
	got := ValidateDoc(doc)
	codes := errorCodes(got.Errors)
	if !diagListContains(codes, "UNSUPPORTED_STEP") {
		t.Fatalf("expected UNSUPPORTED_STEP error, got %+v", codes)
	}
}

func TestEstimateDurationSecondsWithRepeats(t *testing.T) {
	t.Parallel()
	doc := WorkoutDoc{Steps: []Step{
		{Duration: 600, RPE: &Target{Value: ptrFloat(3), Units: "RPE"}},
		{Reps: 3, Steps: []Step{
			{Duration: 60, RPE: &Target{Value: ptrFloat(8), Units: "RPE"}},
			{Duration: 120, RPE: &Target{Value: ptrFloat(2), Units: "RPE"}},
		}},
	}}
	got := EstimateDurationSeconds(doc)
	want := 600 + 3*(60+120)
	if got != want {
		t.Fatalf("EstimateDurationSeconds() = %d, want %d", got, want)
	}
}

func warningCodes(diags []Diagnostic) []string {
	codes := make([]string, 0, len(diags))
	for _, d := range diags {
		codes = append(codes, d.Code)
	}
	return codes
}

func errorCodes(diags []Diagnostic) []string { return warningCodes(diags) }

func diagListContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
