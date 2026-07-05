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

func TestRepeatTrailingCooldownStaysOutsideRepeat(t *testing.T) {
	t.Parallel()

	doc := readStructured(t, filepath.Join("testdata", "07-repeat-trailing-cooldown-structured.json"))
	if len(doc.Steps) != 3 {
		t.Fatalf("fixture has %d top-level steps, want warmup, repeat, cooldown", len(doc.Steps))
	}
	if got := doc.Steps[1]; got.Description != "Main Set" || got.Reps != 3 || len(got.Steps) != 2 {
		t.Fatalf("middle step = %#v, want 3x Main Set with two children", got)
	}
	if got := doc.Steps[2]; got.Description != "Cooldown" || got.Duration != 480 {
		t.Fatalf("trailing step = %#v, want top-level cooldown", got)
	}

	parsed, err := Parse(readGolden(t, filepath.Join("testdata", "07-repeat-trailing-cooldown-dsl.txt")))
	if err != nil {
		t.Fatalf("Parse(repeat trailing cooldown fixture) error = %v", err)
	}
	if !reflect.DeepEqual(parsed.Steps, doc.Steps) {
		gotJSON, _ := json.MarshalIndent(parsed, "", "  ")
		wantJSON, _ := json.MarshalIndent(doc, "", "  ")
		t.Fatalf("Parse(repeat trailing cooldown fixture) mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
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

func TestWorkoutDocYardDistanceSerializeParseValidate(t *testing.T) {
	t.Parallel()

	doc := WorkoutDoc{Steps: []Step{{Description: "Swim", Distance: &Length{Value: 100, Unit: "yards"}, Pace: targetValue(95, "PERCENT_THRESHOLD")}}}
	got, err := Serialize(doc)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	if got != "- Swim 100yd 95% Pace" {
		t.Fatalf("Serialize() = %q, want canonical yd", got)
	}
	parsed, err := Parse(got)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Steps[0].Distance == nil || parsed.Steps[0].Distance.Unit != "yd" || parsed.Steps[0].Distance.Value != 100 {
		t.Fatalf("parsed distance = %#v, want 100yd", parsed.Steps[0].Distance)
	}
	validated := ValidateDoc(doc)
	if len(validated.Errors) != 0 {
		t.Fatalf("ValidateDoc() errors = %+v, want none", validated.Errors)
	}
}

func TestWorkoutDocDistanceAliasesRemainCanonical(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		unit string
		want string
	}{
		{name: "m alias remains meters", unit: "m", want: "- Stride 400mtr 120%"},
		{name: "meters alias remains meters", unit: "meters", want: "- Stride 400mtr 120%"},
		{name: "kilometers alias remains km", unit: "kilometers", want: "- Stride 0.4km 120%"},
		{name: "miles alias remains mi", unit: "miles", want: "- Stride 0.25mi 120%"},
		{name: "yards alias emits yd", unit: "yards", want: "- Stride 25yd 120%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			value := 400.0
			if tc.unit == "kilometers" {
				value = 0.4
			}
			if tc.unit == "miles" {
				value = 0.25
			}
			if tc.unit == "yards" {
				value = 25
			}
			doc := WorkoutDoc{Steps: []Step{{Description: "Stride", Distance: &Length{Value: value, Unit: tc.unit}, Power: targetValue(120, "PERCENT_FTP")}}}
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

func TestSerializeRejectsFractionalPercentTargets(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		step Step
	}{
		{name: "power percent FTP scalar", step: Step{Description: "Bad FTP", Duration: 600, Power: targetValue(0.95, "PERCENT_FTP")}},
		{name: "power percent FTP alias", step: Step{Description: "Bad alias", Duration: 600, Power: targetValue(0.95, "%FTP")}},
		{name: "power blank percent default", step: Step{Description: "Bad blank", Duration: 600, Power: targetValue(0.95, "")}},
		{name: "power percent FTP range", step: Step{Description: "Bad range", Duration: 600, Power: targetRange(0.88, 0.94, "PERCENT_FTP")}},
		{name: "heart rate percent", step: Step{Description: "Bad HR", Duration: 600, HR: targetValue(0.8, "PERCENT_HR")}},
		{name: "pace percent", step: Step{Description: "Bad pace", Duration: 600, Pace: targetValue(0.95, "PERCENT_THRESHOLD")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Serialize(WorkoutDoc{Steps: []Step{tc.step}})
			if err == nil {
				t.Fatalf("Serialize() = %q, want fractional percent error", got)
			}
			if strings.Contains(got, "0.") || strings.Contains(err.Error(), "0.95%") {
				t.Fatalf("fractional percent was silently serialized: dsl=%q err=%q", got, err.Error())
			}
			if !strings.Contains(err.Error(), "percent points") {
				t.Fatalf("Serialize() error = %q, want percent-point guidance", err.Error())
			}
		})
	}
}

func TestSerializeRampDirectionsAreExplicit(t *testing.T) {
	t.Parallel()

	doc := WorkoutDoc{Steps: []Step{
		{Description: "Warmup", Duration: 600, Ramp: true, Power: targetRamp(55, 75, "PERCENT_FTP")},
		{Description: "Cooldown", Duration: 600, Ramp: true, Power: targetRamp(65, 45, "PERCENT_FTP")},
	}}
	got, err := Serialize(doc)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	want := "- Warmup 10m ramp 55-75%\n- Cooldown 10m ramp 65-45%"
	if got != want {
		t.Fatalf("Serialize() = %q, want %q", got, want)
	}

	parsed, err := Parse(got)
	if err != nil {
		t.Fatalf("Parse(serialized ramps) error = %v", err)
	}
	if !reflect.DeepEqual(parsed.Steps, doc.Steps) {
		gotJSON, _ := json.MarshalIndent(parsed.Steps, "", "  ")
		wantJSON, _ := json.MarshalIndent(doc.Steps, "", "  ")
		t.Fatalf("Parse(Serialize(ramps)) mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}

	var syntaxRamp Step
	for _, feature := range WorkoutSyntaxSpec().Features {
		if feature.Key != "ramps" || len(feature.Examples) == 0 {
			continue
		}
		syntaxRamp = feature.Examples[0].Step
		break
	}
	syntaxDSL, err := Serialize(WorkoutDoc{Steps: []Step{syntaxRamp}})
	if err != nil {
		t.Fatalf("Serialize(syntax ramp) error = %v", err)
	}
	if !strings.Contains(syntaxDSL, "ramp 70-95%") {
		t.Fatalf("syntax ramp = %q, want ascending generated ramp example", syntaxDSL)
	}
}

func TestSerializeRecoveryStepsOmitCadenceUnlessExplicit(t *testing.T) {
	t.Parallel()

	doc := WorkoutDoc{Steps: []Step{
		{Description: "Recovery", Duration: 240, Power: targetValue(50, "PERCENT_FTP")},
		{Description: "Cooldown", Duration: 600, Freeride: true},
		{Description: "Recovery spin", Duration: 180, Power: targetValue(55, "PERCENT_FTP"), Cadence: targetRange(85, 95, "RPM")},
	}}
	got, err := Serialize(doc)
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}
	want := "- Recovery 4m 50%\n- Cooldown 10m freeride\n- Recovery spin 3m 55% 85-95rpm"
	if got != want {
		t.Fatalf("Serialize() = %q, want %q", got, want)
	}
	if strings.Contains(strings.Split(got, "\n")[0], "rpm") || strings.Contains(strings.Split(got, "\n")[1], "rpm") {
		t.Fatalf("recovery/cooldown without cadence emitted rpm: %q", got)
	}

	parsed, err := Parse("- Recovery 4m 50%\n- Cooldown 10m freeride")
	if err != nil {
		t.Fatalf("Parse(recovery without cadence) error = %v", err)
	}
	for i, step := range parsed.Steps {
		if step.Cadence != nil {
			t.Fatalf("parsed step %d Cadence = %#v, want nil", i, step.Cadence)
		}
	}
}

func TestFullSurfaceCandidateGoldenLocksIssue25RiskAreas(t *testing.T) {
	t.Parallel()

	doc := readStructured(t, filepath.Join("testdata", "06-full-surface-upstream-candidate-structured.json"))
	want := readGolden(t, filepath.Join("testdata", "06-full-surface-upstream-candidate-dsl.txt"))

	got, err := Serialize(doc)
	if err != nil {
		t.Fatalf("Serialize(full surface candidate) error = %v", err)
	}
	if got != want {
		t.Fatalf("Serialize(full surface candidate) mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}

	parsed, err := Parse(want)
	if err != nil {
		t.Fatalf("Parse(full surface candidate) error = %v", err)
	}
	if !reflect.DeepEqual(parsed.Steps, doc.Steps) {
		gotJSON, _ := json.MarshalIndent(parsed, "", "  ")
		wantJSON, _ := json.MarshalIndent(doc, "", "  ")
		t.Fatalf("Parse(full surface candidate) mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}
}

func TestSerializeWithOptionsKnownWorkoutOrdersEmitZoneMetricSuffixes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		order string
		step  Step
		want  string
	}{
		{name: "hr primary order keeps HR zone explicit", order: "HR_POWER_PACE", step: Step{Description: "HR", Duration: 600, HR: targetValue(2, "HR_ZONE")}, want: "- HR 10m Z2 HR"},
		{name: "pace primary order keeps pace zone explicit", order: "PACE_HR_POWER", step: Step{Description: "Pace", Duration: 600, Pace: targetValue(2, "PACE_ZONE")}, want: "- Pace 10m Z2 Pace"},
		{name: "power zone range gets power suffix", order: "POWER_HR_PACE", step: Step{Description: "Power", Duration: 900, Power: targetRange(2, 3, "POWER_ZONE")}, want: "- Power 15m Z2-Z3 Power"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := SerializeWithOptions(WorkoutDoc{Steps: []Step{tc.step}}, SerializeOptions{WorkoutOrder: tc.order})
			if err != nil {
				t.Fatalf("SerializeWithOptions() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("SerializeWithOptions() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSerializeWithOptionsDeviceTargetPriorityOrdersEmitStructuredZoneTargets(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		order string
		step  Step
		want  string
	}{
		{name: "pace before HR before power", order: "PACE_HR_POWER", step: Step{Description: "Pace", Duration: 600, Pace: targetValue(2, "PACE_ZONE")}, want: "- Pace 10m Z2 Pace"},
		{name: "power before pace before HR", order: "POWER_PACE_HR", step: Step{Description: "Power", Duration: 600, Power: targetValue(2, "POWER_ZONE")}, want: "- Power 10m Z2 Power"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := SerializeWithOptions(WorkoutDoc{Steps: []Step{tc.step}}, SerializeOptions{WorkoutOrder: tc.order})
			if err != nil {
				t.Fatalf("SerializeWithOptions() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("SerializeWithOptions() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSerializeAbsolutePaceUnitsEmitStructuredPaceTargets(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		step Step
		want string
	}{
		{name: "minutes per km scalar", step: Step{Description: "Metric pace", Duration: 300, Pace: targetValue(300, "MINS_KM")}, want: "- Metric pace 5m 5:00/km Pace"},
		{name: "minutes per km range", step: Step{Description: "Metric range", Duration: 300, Pace: targetRange(285, 300, "MINS_KM")}, want: "- Metric range 5m 4:45-5:00/km Pace"},
		{name: "minutes per mile scalar", step: Step{Description: "Imperial pace", Duration: 480, Pace: targetValue(480, "MINS_MILE")}, want: "- Imperial pace 8m 8:00/mi Pace"},
		{name: "minutes per mile range", step: Step{Description: "Imperial range", Duration: 480, Pace: targetRange(465, 480, "MINS_MILE")}, want: "- Imperial range 8m 7:45-8:00/mi Pace"},
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
	return strings.TrimSuffix(string(data), "\n")
}

func floatPtr(value float64) *float64 {
	return &value
}
