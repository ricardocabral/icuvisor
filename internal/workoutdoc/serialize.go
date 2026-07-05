package workoutdoc

import (
	"fmt"
	"math"
	"strings"
)

// SerializeOptions configures context-aware WorkoutDoc serialization without
// changing the default no-context Serialize behavior.
type SerializeOptions struct {
	WorkoutOrder string
}

// Serialize emits a deterministic Intervals.icu workout-description DSL string.
func Serialize(doc WorkoutDoc) (string, error) {
	return SerializeWithOptions(doc, SerializeOptions{})
}

// SerializeWithOptions emits a deterministic Intervals.icu workout-description DSL string.
func SerializeWithOptions(doc WorkoutDoc, options SerializeOptions) (string, error) {
	lines := make([]string, 0, len(doc.Steps))
	for _, step := range doc.Steps {
		emitted, err := serializeStep(step, 0, false, options)
		if err != nil {
			return "", err
		}
		lines = append(lines, emitted...)
	}
	return strings.Join(lines, "\n"), nil
}

func serializeStep(step Step, depth int, inRepeat bool, options SerializeOptions) ([]string, error) {
	if step.Reps > 0 || len(step.Steps) > 0 {
		return serializeRepeat(step, depth, inRepeat, options)
	}
	line, err := serializeSimpleStep(step, depth, options)
	if err != nil {
		return nil, err
	}
	return []string{line}, nil
}

func serializeRepeat(step Step, depth int, inRepeat bool, options SerializeOptions) ([]string, error) {
	if err := descriptionStructuralTokenError(step, "repeat"); err != nil {
		return nil, err
	}
	if inRepeat {
		return nil, unsupported(step, "nested repeats are not supported by the upstream workout DSL")
	}
	if step.Reps <= 0 {
		return nil, unsupported(step, "repeat block requires reps greater than zero")
	}
	if len(step.Steps) == 0 {
		return nil, unsupported(step, "repeat block requires child steps")
	}
	if step.Duration != 0 || step.Distance != nil || step.Power != nil || step.HR != nil || step.Pace != nil || step.RPE != nil || step.Cadence != nil || step.Ramp || step.Freeride {
		return nil, unsupported(step, "repeat block cannot also carry simple step fields")
	}

	header := fmt.Sprintf("%dx", step.Reps)
	if step.Description != "" {
		header = step.Description + " " + header
	}
	lines := []string{indent(depth) + header}
	for _, child := range step.Steps {
		emitted, err := serializeStep(child, depth+1, true, options)
		if err != nil {
			return nil, err
		}
		lines = append(lines, emitted...)
	}
	return lines, nil
}

func serializeSimpleStep(step Step, depth int, options SerializeOptions) (string, error) {
	if err := descriptionStructuralTokenError(step, "step"); err != nil {
		return "", err
	}
	if step.Duration <= 0 && step.Distance == nil {
		return "", unsupported(step, "step requires duration or distance")
	}
	parts := make([]string, 0, 5)
	if step.Description != "" {
		parts = append(parts, step.Description)
	}
	if step.Duration > 0 {
		parts = append(parts, formatDuration(step.Duration))
	} else if step.Distance != nil {
		distance, err := formatDistance(*step.Distance)
		if err != nil {
			return "", unsupported(step, err.Error())
		}
		parts = append(parts, distance)
	}

	target, err := stepTarget(step, options)
	if err != nil {
		return "", err
	}
	if target != "" {
		parts = append(parts, target)
	}
	if step.Cadence != nil {
		cadence, err := formatCadence(*step.Cadence)
		if err != nil {
			return "", unsupported(step, err.Error())
		}
		parts = append(parts, cadence)
	}
	return indent(depth) + "- " + strings.Join(parts, " "), nil
}

func stepTarget(step Step, options SerializeOptions) (string, error) {
	targets := 0
	for _, target := range []*Target{step.Power, step.HR, step.Pace, step.RPE} {
		if target != nil {
			targets++
		}
	}
	if step.Freeride {
		targets++
	}
	if targets > 1 {
		return "", unsupported(step, "step can only contain one primary target")
	}
	if step.Freeride {
		if step.Ramp {
			return "", unsupported(step, "freeride cannot be combined with ramp")
		}
		return "freeride", nil
	}
	var formatted string
	var err error
	switch {
	case step.Power != nil:
		formatted, err = formatTarget("power", *step.Power, step.Ramp, options)
	case step.HR != nil:
		formatted, err = formatTarget("hr", *step.HR, step.Ramp, options)
	case step.Pace != nil:
		formatted, err = formatTarget("pace", *step.Pace, step.Ramp, options)
	case step.RPE != nil:
		formatted, err = formatTarget("rpe", *step.RPE, step.Ramp, options)
	case step.Ramp:
		return "", unsupported(step, "ramp requires a power, heart-rate, pace, or RPE target")
	default:
		return "", nil
	}
	if err != nil {
		return "", unsupported(step, err.Error())
	}
	if step.Ramp {
		return "ramp " + formatted, nil
	}
	return formatted, nil
}

func formatTarget(family string, target Target, ramp bool, options SerializeOptions) (string, error) {
	if target.Text != "" {
		if ramp {
			return "", fmt.Errorf("text targets cannot be used for ramps")
		}
		return target.Text, nil
	}
	lo, hi, ranged, err := targetBounds(target, ramp)
	if err != nil {
		return "", err
	}
	unit := canonicalUnit(target.Units)
	if family == "pace" && isAbsolutePaceUnit(unit) {
		return formatAbsolutePaceTarget(lo, hi, ranged, unit)
	}
	for _, syntax := range workoutTargetUnits {
		if syntax.Family != family || !syntaxUnitMatches(syntax.Units, unit) {
			continue
		}
		if err := rejectFractionalPercentTarget(lo, hi, ranged, syntax); err != nil {
			return "", err
		}
		if syntax.Zone {
			suffix := syntax.Suffix
			if explicitZoneMetricSuffixes(options) {
				suffix = explicitZoneMetricSuffix(family, suffix)
			}
			return syntax.Prefix + formatZoneRange(lo, hi, ranged, suffix), nil
		}
		return syntax.Prefix + formatRange(lo, hi, ranged, syntax.Suffix), nil
	}
	return "", fmt.Errorf("unsupported %s target units %q", family, target.Units)
}

func rejectFractionalPercentTarget(lo float64, hi float64, ranged bool, syntax TargetUnitSyntax) error {
	if !strings.Contains(syntax.Suffix, "%") {
		return nil
	}
	if isFractionalPercentPoint(lo) || (ranged && isFractionalPercentPoint(hi)) {
		return fmt.Errorf("%s targets use percent points, not fractional ratios; use 95 for 95%%, not 0.95", syntax.Family)
	}
	return nil
}

func isFractionalPercentPoint(value float64) bool {
	return value > 0 && value < 1
}

func explicitZoneMetricSuffixes(options SerializeOptions) bool {
	switch canonicalUnit(options.WorkoutOrder) {
	case "POWER_HR_PACE", "POWER_PACE_HR", "HR_POWER_PACE", "HR_PACE_POWER", "PACE_POWER_HR", "PACE_HR_POWER":
		return true
	default:
		return false
	}
}

func explicitZoneMetricSuffix(family string, fallback string) string {
	switch family {
	case "power":
		return " Power"
	case "hr":
		return " HR"
	case "pace":
		return " Pace"
	default:
		return fallback
	}
}

func isAbsolutePaceUnit(unit string) bool {
	return unit == "MINS_KM" || unit == "MINS_MILE"
}

func formatAbsolutePaceTarget(lo, hi float64, ranged bool, unit string) (string, error) {
	if lo <= 0 || (ranged && hi <= 0) {
		return "", fmt.Errorf("absolute pace targets must be positive")
	}
	suffix := "/km"
	if unit == "MINS_MILE" {
		suffix = "/mi"
	}
	formatted := formatPaceDuration(lo)
	if ranged {
		formatted += "-" + formatPaceDuration(hi)
	}
	return formatted + suffix + " Pace", nil
}

func formatPaceDuration(seconds float64) string {
	total := int(math.Round(seconds))
	minutes := total / 60
	remaining := total % 60
	return fmt.Sprintf("%d:%02d", minutes, remaining)
}

func syntaxUnitMatches(units []string, unit string) bool {
	for _, candidate := range units {
		if canonicalUnit(candidate) == unit {
			return true
		}
	}
	return false
}

func targetBounds(target Target, ramp bool) (float64, float64, bool, error) {
	if ramp {
		if target.Start == nil || target.End == nil {
			return 0, 0, false, fmt.Errorf("ramp target requires start and end")
		}
		return *target.Start, *target.End, true, nil
	}
	if target.Value != nil {
		if target.Min != nil || target.Max != nil || target.Start != nil || target.End != nil {
			return 0, 0, false, fmt.Errorf("target value cannot be combined with range bounds")
		}
		return *target.Value, 0, false, nil
	}
	if target.Min != nil && target.Max != nil {
		return *target.Min, *target.Max, true, nil
	}
	return 0, 0, false, fmt.Errorf("target requires value or min/max range")
}

func formatCadence(target Target) (string, error) {
	unit := canonicalUnit(target.Units)
	if unit != "" && unit != "RPM" {
		return "", fmt.Errorf("unsupported cadence units %q", target.Units)
	}
	lo, hi, ranged, err := targetBounds(target, false)
	if err != nil {
		return "", err
	}
	return formatRange(lo, hi, ranged, "rpm"), nil
}

func formatDuration(seconds int) string {
	remaining := seconds
	hours := remaining / 3600
	remaining %= 3600
	minutes := remaining / 60
	remaining %= 60
	var b strings.Builder
	if hours > 0 {
		fmt.Fprintf(&b, "%dh", hours)
	}
	if minutes > 0 {
		fmt.Fprintf(&b, "%dm", minutes)
	}
	if remaining > 0 || b.Len() == 0 {
		fmt.Fprintf(&b, "%ds", remaining)
	}
	return b.String()
}

func formatDistance(distance Length) (string, error) {
	unit := strings.ToLower(strings.TrimSpace(distance.Unit))
	for _, syntax := range workoutDistanceUnits {
		for _, alias := range syntax.Aliases {
			if strings.ToLower(strings.TrimSpace(alias)) == unit {
				return formatNumber(distance.Value) + syntax.Canonical, nil
			}
		}
	}
	return "", fmt.Errorf("unsupported distance unit %q", distance.Unit)
}

func formatRange(lo float64, hi float64, ranged bool, suffix string) string {
	if !ranged {
		return formatNumber(lo) + suffix
	}
	return formatNumber(lo) + "-" + formatNumber(hi) + suffix
}

func formatZoneRange(lo float64, hi float64, ranged bool, suffix string) string {
	if !ranged {
		return "Z" + formatNumber(lo) + suffix
	}
	return "Z" + formatNumber(lo) + "-Z" + formatNumber(hi) + suffix
}

func formatNumber(value float64) string {
	if math.Trunc(value) == value {
		return fmt.Sprintf("%.0f", value)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", value), "0"), ".")
}

func canonicalUnit(unit string) string {
	unit = strings.TrimSpace(unit)
	unit = strings.ReplaceAll(unit, " ", "_")
	unit = strings.ReplaceAll(unit, "-", "_")
	return strings.ToUpper(unit)
}

func indent(depth int) string {
	return strings.Repeat("  ", depth)
}

func unsupported(step Step, reason string) *UnsupportedStepError {
	return &UnsupportedStepError{Step: step, Reason: reason}
}
