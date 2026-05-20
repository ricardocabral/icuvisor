package analysis

import (
	"reflect"
	"testing"
)

func TestNewAnalyzerMetaNormalizesDefaults(t *testing.T) {
	got := NewAnalyzerMeta(AnalyzerMetaInput{
		Method:        " z_score ",
		SourceTools:   []string{" get_wellness_data ", "get_activities", "get_wellness_data", " "},
		N:             -3,
		MissingDays:   -2,
		MissingAction: " ",
		MinSamples:    MinBaselineSamples,
		FormulaRef:    " icuvisor://analysis-formulas#z_score ",
	})
	if got.Method != "z_score" {
		t.Fatalf("Method = %q, want z_score", got.Method)
	}
	if !reflect.DeepEqual(got.SourceTools, []string{"get_activities", "get_wellness_data"}) {
		t.Fatalf("SourceTools = %#v, want deterministic sorted unique tools", got.SourceTools)
	}
	if got.N != 0 || got.MissingDays != 0 {
		t.Fatalf("N/MissingDays = %d/%d, want negative counts clamped to zero", got.N, got.MissingDays)
	}
	if got.MissingAction != MissingActionSkip {
		t.Fatalf("MissingAction = %q, want %q", got.MissingAction, MissingActionSkip)
	}
	if !got.InsufficientSample {
		t.Fatal("InsufficientSample = false, want true for clamped n below minimum")
	}
	if got.FormulaRef != "icuvisor://analysis-formulas#z_score" {
		t.Fatalf("FormulaRef = %q, want trimmed caller-provided ref", got.FormulaRef)
	}
}

func TestInsufficientSample(t *testing.T) {
	tests := []struct {
		name string
		n    int
		min  int
		want bool
	}{
		{name: "no minimum", n: 0, min: 0, want: false},
		{name: "below minimum", n: MinBaselineSamples - 1, min: MinBaselineSamples, want: true},
		{name: "at minimum", n: MinBaselineSamples, min: MinBaselineSamples, want: false},
		{name: "above minimum", n: MinCorrelationSamples + 1, min: MinCorrelationSamples, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := InsufficientSample(tc.n, tc.min); got != tc.want {
				t.Fatalf("InsufficientSample(%d, %d) = %v, want %v", tc.n, tc.min, got, tc.want)
			}
		})
	}
}
