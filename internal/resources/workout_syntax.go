package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	WorkoutSyntaxURI      = "icuvisor://workout-syntax"
	WorkoutSyntaxMIMEType = "text/markdown"
)

// WorkoutSyntaxResource returns the structured workout DSL resource definition.
func WorkoutSyntaxResource() Resource {
	return Resource{
		URI:         WorkoutSyntaxURI,
		Name:        "workout_syntax",
		Title:       "Workout syntax",
		Description: "Intervals.icu structured-workout DSL syntax supported by icuvisor.",
		MIMEType:    WorkoutSyntaxMIMEType,
		Handler: func(ctx context.Context, _ Request) (Result, error) {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			text, err := WorkoutSyntaxMarkdown()
			if err != nil {
				return Result{}, err
			}
			return Result{URI: WorkoutSyntaxURI, MIMEType: WorkoutSyntaxMIMEType, Text: text}, nil
		},
	}
}

// WorkoutSyntaxMarkdown renders the workout syntax resource from workoutdoc descriptors.
func WorkoutSyntaxMarkdown() (string, error) {
	spec := workoutdoc.WorkoutSyntaxSpec()
	var b strings.Builder
	b.WriteString("# Workout syntax\n\n")
	b.WriteString("This resource documents the Intervals.icu structured-workout DSL emitted by icuvisor. Examples are generated from `internal/workoutdoc` structured steps with `workoutdoc.Serialize`, so the resource follows the serializer rather than a separate hand-authored grammar.\n\n")
	b.WriteString("## General form\n\n")
	b.WriteString("- Simple steps begin with `- ` and include an optional description, a duration or distance, then at most one primary target plus optional cadence.\n")
	b.WriteString("- Repeat blocks use an `Nx` header and two-space-indented child steps.\n")
	b.WriteString("- Numeric ranges use `low-high`; zones use `Zlow-Zhigh`.\n\n")

	b.WriteString("## Supported features\n")
	for _, feature := range spec.Features {
		if feature.Key == "" || feature.Title == "" {
			return "", fmt.Errorf("workout syntax feature is missing key/title")
		}
		b.WriteString("\n### ")
		b.WriteString(feature.Title)
		b.WriteString("\n\n")
		b.WriteString(feature.Description)
		b.WriteString("\n\n")
		for _, example := range feature.Examples {
			if example.Key == "" {
				return "", fmt.Errorf("workout syntax example in %s is missing key", feature.Key)
			}
			dsl, err := workoutdoc.Serialize(workoutdoc.WorkoutDoc{Steps: []workoutdoc.Step{example.Step}})
			if err != nil {
				return "", fmt.Errorf("serializing workout syntax example %s: %w", example.Key, err)
			}
			b.WriteString("- `")
			b.WriteString(example.Key)
			b.WriteString("`: ")
			b.WriteString(example.Description)
			b.WriteString("\n\n")
			b.WriteString("```text\n")
			b.WriteString(dsl)
			b.WriteString("\n```\n\n")
		}
	}

	b.WriteString("## Limitations\n\n")
	for _, limitation := range spec.Limitations {
		if limitation.Key == "" || limitation.Description == "" {
			return "", fmt.Errorf("workout syntax limitation is missing key/description")
		}
		b.WriteString("- `")
		b.WriteString(limitation.Key)
		b.WriteString("`: ")
		b.WriteString(limitation.Description)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n", nil
}
