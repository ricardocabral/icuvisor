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
