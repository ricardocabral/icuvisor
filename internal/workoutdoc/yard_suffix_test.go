package workoutdoc

import (
	"strings"
	"testing"
)

// TestYardSuffixCanonicalSerialization verifies that all yard input aliases
// (WorkoutDoc JSON units and DSL tokens) serialize to the canonical yrd suffix.
func TestYardSuffixCanonicalSerialization(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		unit string
		want string
	}{
		{name: "canonical yrd alias", unit: "yrd", want: "- Swim 100yrd 95% Pace"},
		{name: "legacy yd alias", unit: "yd", want: "- Swim 100yrd 95% Pace"},
		{name: "yard alias", unit: "yard", want: "- Swim 100yrd 95% Pace"},
		{name: "yards alias", unit: "yards", want: "- Swim 100yrd 95% Pace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := WorkoutDoc{Steps: []Step{{
				Description: "Swim",
				Distance:    &Length{Value: 100, Unit: tc.unit},
				Pace:        targetValue(95, "PERCENT_THRESHOLD"),
			}}}
			got, err := Serialize(doc)
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Serialize() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestYardSuffixDSLRoundTrip verifies that the legacy 100yd DSL token is still
// accepted by the parser and round-trips to canonical 100yrd on re-serialization.
func TestYardSuffixDSLRoundTrip(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		input       string
		wantUnit    string
		wantReserial string
	}{
		{
			name:        "canonical yrd token parses and is idempotent",
			input:       "- Swim 100yrd 95% Pace",
			wantUnit:    "yrd",
			wantReserial: "- Swim 100yrd 95% Pace",
		},
		{
			name:        "legacy yd token parses and round-trips to yrd",
			input:       "- Swim 100yd 95% Pace",
			wantUnit:    "yd",
			wantReserial: "- Swim 100yrd 95% Pace",
		},
		{
			name:        "yard token parses and round-trips to yrd",
			input:       "- Swim 100yard 95% Pace",
			wantUnit:    "yard",
			wantReserial: "- Swim 100yrd 95% Pace",
		},
		{
			name:        "yards token parses and round-trips to yrd",
			input:       "- Swim 100yards 95% Pace",
			wantUnit:    "yards",
			wantReserial: "- Swim 100yrd 95% Pace",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tc.input, err)
			}
			if len(parsed.Steps) != 1 || parsed.Steps[0].Distance == nil {
				t.Fatalf("Parse(%q) steps = %#v, want one distance step", tc.input, parsed.Steps)
			}
			if parsed.Steps[0].Distance.Unit != tc.wantUnit {
				t.Fatalf("parsed unit = %q, want %q", parsed.Steps[0].Distance.Unit, tc.wantUnit)
			}
			if parsed.Steps[0].Distance.Value != 100 {
				t.Fatalf("parsed value = %v, want 100", parsed.Steps[0].Distance.Value)
			}
			reserial, err := Serialize(parsed)
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}
			if reserial != tc.wantReserial {
				t.Fatalf("re-serialized = %q, want %q", reserial, tc.wantReserial)
			}
		})
	}
}

// TestYardSuffixDescriptionTokenError verifies that a yrd token in a step
// description field is correctly rejected as a structural token error.
func TestYardSuffixDescriptionTokenError(t *testing.T) {
	t.Parallel()

	doc := WorkoutDoc{Steps: []Step{{
		Description: "Swim 100yrd",
		Duration:    300,
		Freeride:    true,
	}}}
	result := ValidateDoc(doc)
	found := false
	for _, e := range result.Errors {
		if e.Code == "STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ValidateDoc() errors = %+v, want STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION", result.Errors)
	}
}

// TestYardSuffixOtherDistanceUnitsUnchanged guards that mtr, km, and mi
// canonical serialization is not affected by the yard suffix change.
func TestYardSuffixOtherDistanceUnitsUnchanged(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		unit string
		val  float64
		want string
	}{
		{name: "mtr unchanged", unit: "meters", val: 400, want: "- Stride 400mtr 120%"},
		{name: "km unchanged", unit: "kilometers", val: 5, want: "- Stride 5km 120%"},
		{name: "mi unchanged", unit: "miles", val: 1, want: "- Stride 1mi 120%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc := WorkoutDoc{Steps: []Step{{
				Description: "Stride",
				Distance:    &Length{Value: tc.val, Unit: tc.unit},
				Power:       targetValue(120, "PERCENT_FTP"),
			}}}
			got, err := Serialize(doc)
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Serialize() = %q, want %q", got, tc.want)
			}
			// Verify round-trip through parser produces the same output.
			parsed, err := Parse(got)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", got, err)
			}
			reserial, err := Serialize(parsed)
			if err != nil {
				t.Fatalf("re-Serialize() error = %v", err)
			}
			if reserial != tc.want {
				t.Fatalf("re-serialized = %q, want %q", reserial, tc.want)
			}
		})
	}
}

// TestYardSuffixValidateDescriptionLegacyYdInput verifies that a description
// containing legacy 100yd DSL is accepted by ValidateDescription.
func TestYardSuffixValidateDescriptionLegacyYdInput(t *testing.T) {
	t.Parallel()

	got := ValidateDescription("- Swim 100yd 95% Pace")
	if len(got.Errors) != 0 {
		t.Fatalf("ValidateDescription() errors = %+v, want none for legacy yd input", got.Errors)
	}
	if len(got.Doc.Steps) != 1 || got.Doc.Steps[0].Distance == nil {
		t.Fatalf("Doc = %#v, want one distance step", got.Doc)
	}

	// Re-serialize to confirm it canonicalizes to yrd.
	reserial, err := Serialize(got.Doc)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if !strings.Contains(reserial, "yrd") {
		t.Fatalf("re-serialized = %q, want yrd canonical suffix", reserial)
	}
}
