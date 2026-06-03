package workoutdoc

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGoldenRoundTripParseSerialize(t *testing.T) {
	for _, tc := range goldenCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			dsl := readGolden(t, tc.dslPath)
			doc, err := Parse(dsl)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got, err := Serialize(doc)
			if err != nil {
				t.Fatalf("Serialize(parsed) error = %v", err)
			}
			if got != dsl {
				t.Fatalf("Serialize(Parse(dsl)) mismatch\n--- got ---\n%s\n--- want ---\n%s", got, dsl)
			}
		})
	}
}

func TestGoldenRoundTripStructuredSerializeParse(t *testing.T) {
	for _, tc := range goldenCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			want := readStructured(t, tc.structuredPath)
			dsl, err := Serialize(want)
			if err != nil {
				t.Fatalf("Serialize(structured) error = %v", err)
			}
			got, err := Parse(dsl)
			if err != nil {
				t.Fatalf("Parse(serialized) error = %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				gotJSON, _ := json.MarshalIndent(got, "", "  ")
				wantJSON, _ := json.MarshalIndent(want, "", "  ")
				t.Fatalf("Parse(Serialize(structured)) mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
			}
		})
	}
}

func TestSerializeRepeatHeadersAreCanonical(t *testing.T) {
	t.Parallel()

	child := []Step{{Duration: 300, Power: targetValue(95, "PERCENT_FTP")}}
	for _, tc := range []struct {
		name string
		step Step
		want string
	}{
		{name: "bare", step: Step{Reps: 3, Steps: child}, want: "3x\n  - 5m 95%"},
		{name: "described", step: Step{Description: "Main Set", Reps: 3, Steps: child}, want: "Main Set 3x\n  - 5m 95%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Serialize(WorkoutDoc{Steps: []Step{tc.step}})
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Serialize() = %q, want %q", got, tc.want)
			}
			firstLine, _, _ := strings.Cut(got, "\n")
			if strings.HasPrefix(firstLine, "-") {
				t.Fatalf("repeat header = %q, want no leading dash", firstLine)
			}
		})
	}
}

func TestParseRejectsDashedRepeatHeaders(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"-3 x\n  - 1m RPE 2", "- 3x\n  - 1m RPE 2"} {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			if _, err := Parse(input); err == nil {
				t.Fatal("Parse() error = nil, want malformed repeat header error")
			}
		})
	}
}

func TestSerializeTargetUnitSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		step Step
		want string
	}{
		{name: "power blank defaults to percent FTP", step: Step{Description: "Blank", Duration: 600, Power: targetValue(75, "")}, want: "- Blank 10m 75%"},
		{name: "power percent FTP", step: Step{Description: "FTP", Duration: 600, Power: targetRange(88, 94, "PERCENT_FTP")}, want: "- FTP 10m 88-94%"},
		{name: "power percent FTP alias", step: Step{Description: "FTP alias", Duration: 600, Power: targetValue(105, "%FTP")}, want: "- FTP alias 10m 105%"},
		{name: "power watts", step: Step{Description: "Watts", Duration: 300, Power: targetValue(250, "WATTS")}, want: "- Watts 5m 250w"},
		{name: "power watt alias", step: Step{Description: "Watt", Duration: 300, Power: targetValue(240, "WATT")}, want: "- Watt 5m 240w"},
		{name: "power w alias", step: Step{Description: "W", Duration: 300, Power: targetRange(220, 260, "W")}, want: "- W 5m 220-260w"},
		{name: "power zone scalar", step: Step{Description: "Power zone", Duration: 900, Power: targetValue(2, "POWER_ZONE")}, want: "- Power zone 15m Z2"},
		{name: "power zone range", step: Step{Description: "Power zones", Duration: 900, Power: targetRange(2, 3, "ZONE")}, want: "- Power zones 15m Z2-Z3"},
		{name: "pace blank defaults to percent threshold", step: Step{Description: "Pace blank", Duration: 600, Pace: targetValue(95, "")}, want: "- Pace blank 10m 95% Pace"},
		{name: "pace percent threshold", step: Step{Description: "Pace", Duration: 600, Pace: targetValue(95, "PERCENT_THRESHOLD")}, want: "- Pace 10m 95% Pace"},
		{name: "pace percent threshold pace alias", step: Step{Description: "Pace threshold", Duration: 600, Pace: targetRange(92, 96, "PERCENT_THRESHOLD_PACE")}, want: "- Pace threshold 10m 92-96% Pace"},
		{name: "pace percent pace alias", step: Step{Description: "Pace percent", Duration: 600, Pace: targetValue(90, "PERCENT_PACE")}, want: "- Pace percent 10m 90% Pace"},
		{name: "pace percent symbol alias", step: Step{Description: "Pace symbol", Duration: 600, Pace: targetRange(90, 95, "%PACE")}, want: "- Pace symbol 10m 90-95% Pace"},
		{name: "pace numeric scalar", step: Step{Description: "Numeric pace", Duration: 300, Pace: targetValue(5, "PACE")}, want: "- Numeric pace 5m 5 Pace"},
		{name: "pace numeric range", step: Step{Description: "Numeric pace range", Duration: 300, Pace: targetRange(4.5, 5, "PACE")}, want: "- Numeric pace range 5m 4.5-5 Pace"},
		{name: "pace zone scalar", step: Step{Description: "Pace zone", Duration: 600, Pace: targetValue(2, "PACE_ZONE")}, want: "- Pace zone 10m Z2 Pace"},
		{name: "pace zone range", step: Step{Description: "Pace zones", Duration: 600, Pace: targetRange(2, 3, "ZONE")}, want: "- Pace zones 10m Z2-Z3 Pace"},
		{name: "pace text form", step: Step{Description: "Text pace", Duration: 300, Pace: &Target{Text: "5:00/km Pace"}}, want: "- Text pace 5m 5:00/km Pace"},
		{name: "heart rate percent HR", step: Step{Description: "HR", Duration: 600, HR: targetValue(80, "PERCENT_HR")}, want: "- HR 10m 80% HR"},
		{name: "heart rate percent max HR alias", step: Step{Description: "Max HR", Duration: 600, HR: targetRange(80, 85, "PERCENT_MAX_HR")}, want: "- Max HR 10m 80-85% HR"},
		{name: "heart rate percent symbol alias", step: Step{Description: "HR symbol", Duration: 600, HR: targetValue(82, "%HR")}, want: "- HR symbol 10m 82% HR"},
		{name: "heart rate HR alias", step: Step{Description: "HR alias", Duration: 600, HR: targetValue(83, "HR")}, want: "- HR alias 10m 83% HR"},
		{name: "heart rate percent LTHR", step: Step{Description: "LTHR", Duration: 600, HR: targetRange(95, 99, "PERCENT_LTHR")}, want: "- LTHR 10m 95-99% LTHR"},
		{name: "heart rate percent LTHR symbol alias", step: Step{Description: "LTHR symbol", Duration: 600, HR: targetValue(96, "%LTHR")}, want: "- LTHR symbol 10m 96% LTHR"},
		{name: "heart rate LTHR alias", step: Step{Description: "LTHR alias", Duration: 600, HR: targetValue(97, "LTHR")}, want: "- LTHR alias 10m 97% LTHR"},
		{name: "heart rate bpm", step: Step{Description: "BPM", Duration: 300, HR: targetValue(150, "BPM")}, want: "- BPM 5m 150bpm"},
		{name: "heart rate zone scalar", step: Step{Description: "HR zone", Duration: 600, HR: targetValue(2, "HR_ZONE")}, want: "- HR zone 10m Z2 HR"},
		{name: "heart rate zone range", step: Step{Description: "HR zones", Duration: 600, HR: targetRange(2, 3, "ZONE")}, want: "- HR zones 10m Z2-Z3 HR"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Serialize(WorkoutDoc{Steps: []Step{tc.step}})
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Serialize() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSerializeRejectsUnsupportedAbsolutePaceUnits(t *testing.T) {
	t.Parallel()

	for _, unit := range []string{"MINS_KM", "MINS_MILE"} {
		t.Run(unit, func(t *testing.T) {
			t.Parallel()

			_, err := Serialize(WorkoutDoc{Steps: []Step{{Description: "Absolute pace", Duration: 300, Pace: targetValue(5, unit)}}})
			if err == nil {
				t.Fatal("Serialize() error = nil, want unsupported pace unit error")
			}
			var unsupported *UnsupportedStepError
			if !errors.As(err, &unsupported) {
				t.Fatalf("Serialize() error = %T, want *UnsupportedStepError", err)
			}
			if !strings.Contains(unsupported.Reason, "unsupported pace target units") {
				t.Fatalf("UnsupportedStepError.Reason = %q, want unsupported pace unit", unsupported.Reason)
			}
		})
	}
}

func TestSerializeUnsupportedStepErrorContainsStep(t *testing.T) {
	bad := WorkoutDoc{Steps: []Step{{Description: "No duration", Power: &Target{Value: floatPtr(60), Units: "PERCENT_FTP"}}}}
	_, err := Serialize(bad)
	if err == nil {
		t.Fatal("Serialize() error = nil, want unsupported step error")
	}
	var unsupported *UnsupportedStepError
	if !errors.As(err, &unsupported) {
		t.Fatalf("Serialize() error = %T, want *UnsupportedStepError", err)
	}
	if unsupported.Step.Description != "No duration" || unsupported.Step.Power == nil {
		t.Fatalf("UnsupportedStepError.Step = %#v, want offending step", unsupported.Step)
	}
}

func TestSerializeRejectsDurationOrDistanceTokensInStepDescription(t *testing.T) {
	for _, tc := range []struct {
		name        string
		description string
	}{
		{name: "duration", description: "Endurance 2h15m"},
		{name: "duration with punctuation", description: "Warm up (45m)"},
		{name: "distance", description: "5km pace"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := WorkoutDoc{Steps: []Step{{Description: tc.description, Duration: 8100, Power: &Target{Value: floatPtr(60), Units: "PERCENT_FTP"}}}}
			_, err := Serialize(doc)
			if err == nil {
				t.Fatal("Serialize() error = nil, want structural token error")
			}
			var structural *StructuralTokenInDescriptionError
			if !errors.As(err, &structural) {
				t.Fatalf("Serialize() error = %T, want *StructuralTokenInDescriptionError", err)
			}
			if !strings.Contains(err.Error(), "duration/distance in structured fields") {
				t.Fatalf("error = %q, want structured-field guidance", err.Error())
			}
		})
	}
}

type goldenCase struct {
	name           string
	structuredPath string
	dslPath        string
}

func goldenCases(t *testing.T) []goldenCase {
	t.Helper()
	structured, err := filepath.Glob(filepath.Join("testdata", "*-structured.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(structured) == 0 {
		t.Fatal("no structured golden files found")
	}
	cases := make([]goldenCase, 0, len(structured))
	for _, structuredPath := range structured {
		name := strings.TrimSuffix(filepath.Base(structuredPath), "-structured.json")
		dslPath := filepath.Join("testdata", name+"-dsl.txt")
		if _, err := os.Stat(dslPath); err != nil {
			t.Fatalf("missing DSL golden for %s: %v", name, err)
		}
		cases = append(cases, goldenCase{name: name, structuredPath: structuredPath, dslPath: dslPath})
	}
	return cases
}

func readStructured(t *testing.T, path string) WorkoutDoc {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc WorkoutDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return doc
}

func readGolden(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func floatPtr(value float64) *float64 {
	return &value
}
