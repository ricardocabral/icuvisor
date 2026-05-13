// Package workoutdoc parses and serializes Intervals.icu workout description DSL.
package workoutdoc

import "fmt"

// WorkoutDoc is the structured Intervals.icu workout document shape used by reads.
type WorkoutDoc struct {
	Name  string `json:"name,omitempty"`
	Steps []Step `json:"steps"`
}

// Step is one workout step or a repeat block.
type Step struct {
	Description string  `json:"description,omitempty"`
	Duration    int     `json:"duration,omitempty"`
	Distance    *Length `json:"distance,omitempty"`

	Power   *Target `json:"power,omitempty"`
	HR      *Target `json:"hr,omitempty"`
	Pace    *Target `json:"pace,omitempty"`
	RPE     *Target `json:"rpe,omitempty"`
	Cadence *Target `json:"cadence,omitempty"`

	Ramp     bool   `json:"ramp,omitempty"`
	Freeride bool   `json:"freeride,omitempty"`
	Reps     int    `json:"reps,omitempty"`
	Steps    []Step `json:"steps,omitempty"`
}

// Length is a distance target for distance-based workout steps.
type Length struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// Target is a scalar or range target for power, heart rate, pace, RPE, or cadence.
type Target struct {
	Value *float64 `json:"value,omitempty"`
	Min   *float64 `json:"min,omitempty"`
	Max   *float64 `json:"max,omitempty"`
	Start *float64 `json:"start,omitempty"`
	End   *float64 `json:"end,omitempty"`
	Units string   `json:"units,omitempty"`
	Text  string   `json:"text,omitempty"`
}

// UnsupportedStepError reports a structured workout step that has no safe DSL representation.
type UnsupportedStepError struct {
	Step   Step
	Reason string
}

func (e *UnsupportedStepError) Error() string {
	if e == nil {
		return "unsupported workout step"
	}
	if e.Reason == "" {
		return fmt.Sprintf("unsupported workout step: %+v", e.Step)
	}
	return fmt.Sprintf("unsupported workout step: %s: %+v", e.Reason, e.Step)
}
