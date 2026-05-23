package workoutdoc

import (
	"strings"
	"testing"
)

func TestMergeDescription(t *testing.T) {
	t.Parallel()

	simpleDoc := WorkoutDoc{Steps: []Step{{Description: "Warm up", Duration: 600, RPE: &Target{Value: ptrFloat(3), Units: "RPE"}}}}
	wantSimpleDSL := "- Warm up 10m RPE 3"

	tests := []struct {
		name  string
		prose string
		doc   WorkoutDoc
		want  string
	}{
		{name: "doc only", prose: "", doc: simpleDoc, want: wantSimpleDSL},
		{name: "prose only", prose: "Easy spin morning ride.", doc: WorkoutDoc{}, want: "Easy spin morning ride."},
		{name: "appends after prose with blank line", prose: "Race week opener.", doc: simpleDoc, want: "Race week opener.\n\n" + wantSimpleDSL},
		{name: "sentinel placement replaces line", prose: "Header line.\n" + StepsSentinel + "\nTrailing note.", doc: simpleDoc, want: "Header line.\n" + wantSimpleDSL + "\nTrailing note."},
		{name: "trailing newline in prose preserved", prose: "Notes.\n", doc: simpleDoc, want: "Notes.\n" + wantSimpleDSL},
		{name: "no steps returns prose untouched", prose: "Notes only.\n", doc: WorkoutDoc{}, want: "Notes only.\n"},
		{name: "crlf prose is normalized", prose: "Line1.\r\nLine2.", doc: WorkoutDoc{}, want: "Line1.\nLine2."},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := MergeDescription(tc.prose, tc.doc)
			if err != nil {
				t.Fatalf("MergeDescription() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("MergeDescription() mismatch\n--- got ---\n%q\n--- want ---\n%q", got, tc.want)
			}
		})
	}
}

func TestMergeDescriptionPropagatesSerializeError(t *testing.T) {
	t.Parallel()
	doc := WorkoutDoc{Steps: []Step{{Reps: 2, Steps: []Step{{Reps: 2, Steps: []Step{{Duration: 60, RPE: &Target{Value: ptrFloat(2), Units: "RPE"}}}}}}}}
	_, err := MergeDescription("", doc)
	if err == nil {
		t.Fatal("MergeDescription() with nested repeats expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nested repeats") {
		t.Fatalf("MergeDescription() error = %v, want nested-repeats failure", err)
	}
}

func ptrFloat(v float64) *float64 { return &v }
