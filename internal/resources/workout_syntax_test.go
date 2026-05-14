package resources

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

type captureRegistrar struct {
	resources []Resource
}

func (r *captureRegistrar) AddResource(resource Resource) error {
	r.resources = append(r.resources, resource)
	return nil
}

func TestWorkoutSyntaxMarkdownGolden(t *testing.T) {
	t.Parallel()

	got, err := WorkoutSyntaxMarkdown()
	if err != nil {
		t.Fatalf("WorkoutSyntaxMarkdown() error = %v", err)
	}
	want, err := os.ReadFile("testdata/workout_syntax.md")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("WorkoutSyntaxMarkdown() mismatch with testdata/workout_syntax.md")
	}
}

func TestWorkoutSyntaxUnitMatricesAreRendered(t *testing.T) {
	t.Parallel()

	markdown, err := WorkoutSyntaxMarkdown()
	if err != nil {
		t.Fatalf("WorkoutSyntaxMarkdown() error = %v", err)
	}
	for _, unit := range workoutdoc.WorkoutDistanceUnitSyntax() {
		for _, want := range []string{"`" + unit.Key + "`", "`" + unit.Canonical + "`"} {
			if !strings.Contains(markdown, want) {
				t.Fatalf("markdown missing distance unit marker %q", want)
			}
		}
		for _, alias := range unit.Aliases {
			if !strings.Contains(markdown, "`"+alias+"`") {
				t.Fatalf("markdown missing distance unit alias %q for %s", alias, unit.Key)
			}
		}
	}
	for _, unit := range workoutdoc.WorkoutTargetUnitSyntax() {
		for _, want := range []string{"`" + unit.Key + "`", "`" + unit.Family + "`"} {
			if !strings.Contains(markdown, want) {
				t.Fatalf("markdown missing target unit marker %q", want)
			}
		}
		for _, alias := range unit.Units {
			if !strings.Contains(markdown, "`"+alias+"`") {
				t.Fatalf("markdown missing target unit alias %q for %s", alias, unit.Key)
			}
		}
	}
}

func TestWorkoutSyntaxSpecExamplesAreRenderedFromSerializer(t *testing.T) {
	t.Parallel()

	markdown, err := WorkoutSyntaxMarkdown()
	if err != nil {
		t.Fatalf("WorkoutSyntaxMarkdown() error = %v", err)
	}
	spec := workoutdoc.WorkoutSyntaxSpec()
	requiredFeatures := map[string]bool{
		"duration_steps":     false,
		"distance_steps":     false,
		"repeats":            false,
		"freeride":           false,
		"ramps":              false,
		"cadence_targets":    false,
		"power_targets":      false,
		"heart_rate_targets": false,
		"pace_targets":       false,
		"rpe_targets":        false,
	}
	for _, feature := range spec.Features {
		if _, ok := requiredFeatures[feature.Key]; ok {
			requiredFeatures[feature.Key] = true
		}
		if feature.Title == "" || feature.Description == "" || len(feature.Examples) == 0 {
			t.Fatalf("feature %q is incomplete: %#v", feature.Key, feature)
		}
		if !strings.Contains(markdown, "### "+feature.Title) {
			t.Fatalf("markdown missing feature title %q", feature.Title)
		}
		for _, example := range feature.Examples {
			dsl, err := workoutdoc.Serialize(workoutdoc.WorkoutDoc{Steps: []workoutdoc.Step{example.Step}})
			if err != nil {
				t.Fatalf("Serialize(%s/%s) error = %v", feature.Key, example.Key, err)
			}
			for _, want := range []string{"`" + example.Key + "`", dsl} {
				if !strings.Contains(markdown, want) {
					t.Fatalf("markdown missing %q for example %s/%s", want, feature.Key, example.Key)
				}
			}
		}
	}
	for key, seen := range requiredFeatures {
		if !seen {
			t.Fatalf("syntax spec missing required serializer feature %q", key)
		}
	}
}

func TestWorkoutSyntaxMarkdownDocumentsLimitations(t *testing.T) {
	t.Parallel()

	markdown, err := WorkoutSyntaxMarkdown()
	if err != nil {
		t.Fatalf("WorkoutSyntaxMarkdown() error = %v", err)
	}
	for _, limitation := range workoutdoc.WorkoutSyntaxSpec().Limitations {
		if limitation.Key == "" || limitation.Description == "" {
			t.Fatalf("incomplete limitation: %#v", limitation)
		}
		if !strings.Contains(markdown, "`"+limitation.Key+"`") || !strings.Contains(markdown, limitation.Description) {
			t.Fatalf("markdown missing limitation %#v", limitation)
		}
	}
}

func TestNewRegistryRegistersWorkoutSyntaxResource(t *testing.T) {
	t.Parallel()

	registrar := &captureRegistrar{}
	if err := NewRegistry().Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if len(registrar.resources) != 1 {
		t.Fatalf("registered resources = %d, want 1", len(registrar.resources))
	}
	resource := registrar.resources[0]
	if resource.URI != WorkoutSyntaxURI || resource.Name != "workout_syntax" || resource.Title != "Workout syntax" || resource.MIMEType != WorkoutSyntaxMIMEType {
		t.Fatalf("resource metadata = %#v, want workout syntax metadata", resource)
	}

	result, err := resource.Handler(context.Background(), Request{URI: WorkoutSyntaxURI})
	if err != nil {
		t.Fatalf("resource handler error = %v", err)
	}
	if result.URI != WorkoutSyntaxURI || result.MIMEType != WorkoutSyntaxMIMEType || !strings.Contains(result.Text, "# Workout syntax") {
		t.Fatalf("resource handler result = %#v, want URI/MIME/markdown", result)
	}
}

func TestWorkoutSyntaxResourceHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := WorkoutSyntaxResource().Handler(ctx, Request{URI: WorkoutSyntaxURI})
	if err == nil {
		t.Fatal("handler error = nil, want context cancellation")
	}
}
