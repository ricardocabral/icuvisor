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

func TestNewAnalyzerMetaHonorsExplicitInsufficientSample(t *testing.T) {
	explicit := true
	got := NewAnalyzerMeta(AnalyzerMetaInput{N: MinBaselineSamples, MinSamples: MinBaselineSamples, InsufficientSample: &explicit})
	if !got.InsufficientSample {
		t.Fatalf("InsufficientSample = false, want explicit true")
	}
}

func TestApplyIntervalSourceEvidence(t *testing.T) {
	tests := []struct {
		name          string
		evidence      IntervalSourceResult
		wantSource    IntervalSource
		wantSuspected bool
	}{
		{name: "device laps keep auto lap suspicion", evidence: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true}, wantSource: IntervalSourceDeviceLaps, wantSuspected: true},
		{name: "manual added source propagates", evidence: IntervalSourceResult{Source: IntervalSourceManualAdded}, wantSource: IntervalSourceManualAdded},
		{name: "mixed source propagates", evidence: IntervalSourceResult{Source: IntervalSourceMixed}, wantSource: IntervalSourceMixed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := AnalyzerMetaInput{SourceTools: []string{"get_activity_intervals", "get_wellness_data"}}
			withEvidence := ApplyIntervalSourceEvidence(input, tc.evidence)
			meta := NewAnalyzerMeta(withEvidence)

			if !reflect.DeepEqual(meta.SourceTools, []string{"get_activity_intervals", "get_wellness_data"}) {
				t.Fatalf("SourceTools = %#v, want deduplicated interval source tools", meta.SourceTools)
			}
			if meta.IntervalSource != tc.wantSource {
				t.Fatalf("IntervalSource = %q, want %q", meta.IntervalSource, tc.wantSource)
			}
			if meta.AutoLapSuspected == nil || *meta.AutoLapSuspected != tc.wantSuspected {
				t.Fatalf("AutoLapSuspected = %#v, want %v pointer", meta.AutoLapSuspected, tc.wantSuspected)
			}
		})
	}
}

func TestIntervalExecutionClaimPolicy(t *testing.T) {
	tests := []struct {
		name string
		in   IntervalSourceResult
		want IntervalExecutionClaimDecision
	}{
		{name: "auto lap suspected declines", in: IntervalSourceResult{Source: IntervalSourceDeviceLaps, AutoLapSuspected: true}, want: IntervalExecutionClaimDecision{Decline: true, Reason: IntervalExecutionDeclineAutoLapSuspected}},
		{name: "structured workout does not decline", in: IntervalSourceResult{Source: IntervalSourceStructuredWorkout}, want: IntervalExecutionClaimDecision{}},
		{name: "unknown without suspicion does not decline", in: IntervalSourceResult{Source: IntervalSourceUnknown}, want: IntervalExecutionClaimDecision{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IntervalExecutionClaimPolicy(tc.in); got != tc.want {
				t.Fatalf("IntervalExecutionClaimPolicy() = %#v, want %#v", got, tc.want)
			}
		})
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
