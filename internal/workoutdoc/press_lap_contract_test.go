package workoutdoc

import (
	"strings"
	"testing"
)

func TestPressLapControlCannotBecomeStructuredStep(t *testing.T) {
	t.Parallel()

	result := ValidateDescription("- Press lap when ready")
	if len(result.Errors) != 1 || result.Errors[0].Code != "PARSE_ERROR" {
		t.Fatalf("Errors = %+v, want one PARSE_ERROR", result.Errors)
	}
	if result.Doc.Steps != nil {
		t.Fatalf("Doc.Steps = %+v, want no structured step", result.Doc.Steps)
	}
	if result.StructuredStepLines != 1 {
		t.Fatalf("StructuredStepLines = %d, want 1", result.StructuredStepLines)
	}
	if strings.Contains(result.Prose, "Press lap") {
		t.Fatalf("Prose = %q, want rejected DSL line not passed through as prose", result.Prose)
	}
}

func TestPressLapProseRemainsProse(t *testing.T) {
	t.Parallel()

	const prose = "Press lap when ready"
	result := ValidateDescription(prose)
	if len(result.Errors) != 0 {
		t.Fatalf("Errors = %+v, want prose to pass through", result.Errors)
	}
	if result.Prose != prose {
		t.Fatalf("Prose = %q, want %q", result.Prose, prose)
	}
	if result.Doc.Steps != nil {
		t.Fatalf("Doc.Steps = %+v, want no structured step", result.Doc.Steps)
	}
	if result.StructuredStepLines != 0 {
		t.Fatalf("StructuredStepLines = %d, want 0", result.StructuredStepLines)
	}
}
