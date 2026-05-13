package workoutdoc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	repeatLineRE    = regexp.MustCompile(`^(.*?\S\s+)?([1-9][0-9]*)x$`)
	durationTokenRE = regexp.MustCompile(`^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$`)
	distanceTokenRE = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(mtr|km|mi)$`)
	wattsTokenRE    = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(?:-([0-9]+(?:\.[0-9]+)?))?w$`)
	percentTokenRE  = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(?:-([0-9]+(?:\.[0-9]+)?))?%$`)
	bpmTokenRE      = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(?:-([0-9]+(?:\.[0-9]+)?))?bpm$`)
	rpmTokenRE      = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(?:-([0-9]+(?:\.[0-9]+)?))?rpm$`)
	zoneTokenRE     = regexp.MustCompile(`^Z([0-9]+)(?:-Z([0-9]+))?$`)
)

// Parse reads the canonical Intervals.icu workout-description DSL emitted by Serialize.
func Parse(dsl string) (WorkoutDoc, error) {
	var doc WorkoutDoc
	lines := splitDSL(dsl)
	steps, next, err := parseLines(lines, 0, 0)
	if err != nil {
		return WorkoutDoc{}, err
	}
	if next != len(lines) {
		return WorkoutDoc{}, fmt.Errorf("unexpected indented workout line %q", strings.TrimSpace(lines[next]))
	}
	doc.Steps = steps
	return doc, nil
}

func splitDSL(dsl string) []string {
	raw := strings.Split(strings.ReplaceAll(dsl, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, strings.TrimRight(line, "\t "))
	}
	return lines
}

func parseLines(lines []string, index int, depth int) ([]Step, int, error) {
	var steps []Step
	for index < len(lines) {
		line := lines[index]
		lineDepth := leadingDepth(line)
		if lineDepth < depth {
			break
		}
		if lineDepth > depth {
			return nil, index, fmt.Errorf("unexpected indentation on workout line %q", strings.TrimSpace(line))
		}
		trimmed := strings.TrimSpace(line)
		if match := repeatLineRE.FindStringSubmatch(trimmed); match != nil {
			reps, _ := strconv.Atoi(match[2])
			description := strings.TrimSpace(match[1])
			children, next, err := parseLines(lines, index+1, depth+1)
			if err != nil {
				return nil, index, err
			}
			if len(children) == 0 {
				return nil, index, fmt.Errorf("repeat block %q has no child steps", trimmed)
			}
			steps = append(steps, Step{Description: description, Reps: reps, Steps: children})
			index = next
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			return nil, index, fmt.Errorf("workout step line must start with '- ': %q", trimmed)
		}
		step, err := parseSimpleLine(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
		if err != nil {
			return nil, index, err
		}
		steps = append(steps, step)
		index++
	}
	return steps, index, nil
}

func parseSimpleLine(body string) (Step, error) {
	tokens := strings.Fields(body)
	if len(tokens) == 0 {
		return Step{}, fmt.Errorf("empty workout step")
	}
	idx := -1
	var duration int
	var distance *Length
	for i, token := range tokens {
		if seconds, ok := parseDurationToken(token); ok {
			idx = i
			duration = seconds
			break
		}
		if parsed, ok := parseDistanceToken(token); ok {
			idx = i
			distance = &parsed
			break
		}
	}
	if idx < 0 {
		return Step{}, fmt.Errorf("workout step missing duration or distance: %q", body)
	}
	step := Step{Description: strings.Join(tokens[:idx], " "), Duration: duration, Distance: distance}
	remaining := tokens[idx+1:]
	if len(remaining) == 0 {
		return step, nil
	}
	if cadence, ok := parseCadenceAtEnd(remaining); ok {
		step.Cadence = cadence
		remaining = remaining[:len(remaining)-1]
	}
	if len(remaining) == 0 {
		return step, nil
	}
	if strings.EqualFold(remaining[0], "freeride") {
		if len(remaining) != 1 {
			return Step{}, fmt.Errorf("freeride step has extra tokens: %q", strings.Join(remaining, " "))
		}
		step.Freeride = true
		return step, nil
	}
	if strings.EqualFold(remaining[0], "ramp") {
		step.Ramp = true
		remaining = remaining[1:]
	}
	if len(remaining) == 0 {
		return Step{}, fmt.Errorf("target missing after ramp")
	}
	if err := parsePrimaryTarget(&step, remaining); err != nil {
		return Step{}, err
	}
	return step, nil
}

func parsePrimaryTarget(step *Step, tokens []string) error {
	if strings.EqualFold(tokens[0], "RPE") {
		if len(tokens) != 2 {
			return fmt.Errorf("invalid RPE target %q", strings.Join(tokens, " "))
		}
		target, err := parseNumberTarget(tokens[1], "RPE")
		if err != nil {
			return err
		}
		step.RPE = target
		return nil
	}
	if match := wattsTokenRE.FindStringSubmatch(tokens[0]); match != nil {
		if len(tokens) != 1 {
			return fmt.Errorf("power watts target has extra tokens: %q", strings.Join(tokens, " "))
		}
		step.Power = targetForStep(step, targetFromRegex(match, "WATTS"))
		return nil
	}
	if match := bpmTokenRE.FindStringSubmatch(tokens[0]); match != nil {
		if len(tokens) != 1 {
			return fmt.Errorf("heart-rate bpm target has extra tokens: %q", strings.Join(tokens, " "))
		}
		step.HR = targetForStep(step, targetFromRegex(match, "BPM"))
		return nil
	}
	if match := percentTokenRE.FindStringSubmatch(tokens[0]); match != nil {
		target := targetFromRegex(match, "PERCENT_FTP")
		if len(tokens) == 1 {
			step.Power = targetForStep(step, target)
			return nil
		}
		if len(tokens) == 2 && strings.EqualFold(tokens[1], "HR") {
			target.Units = "PERCENT_HR"
			step.HR = targetForStep(step, target)
			return nil
		}
		if len(tokens) == 2 && strings.EqualFold(tokens[1], "LTHR") {
			target.Units = "PERCENT_LTHR"
			step.HR = targetForStep(step, target)
			return nil
		}
		if len(tokens) == 2 && strings.EqualFold(tokens[1], "Pace") {
			target.Units = "PERCENT_THRESHOLD"
			step.Pace = targetForStep(step, target)
			return nil
		}
	}
	if match := zoneTokenRE.FindStringSubmatch(tokens[0]); match != nil {
		target := targetFromRegex(match, "ZONE")
		if len(tokens) == 1 {
			step.Power = targetForStep(step, target)
			return nil
		}
		if len(tokens) == 2 && strings.EqualFold(tokens[1], "HR") {
			target.Units = "HR_ZONE"
			step.HR = targetForStep(step, target)
			return nil
		}
		if len(tokens) == 2 && strings.EqualFold(tokens[1], "Pace") {
			target.Units = "PACE_ZONE"
			step.Pace = targetForStep(step, target)
			return nil
		}
	}
	if len(tokens) >= 2 && strings.EqualFold(tokens[len(tokens)-1], "Pace") {
		step.Pace = &Target{Text: strings.Join(tokens, " ")}
		return nil
	}
	return fmt.Errorf("unsupported workout target %q", strings.Join(tokens, " "))
}

func parseDurationToken(token string) (int, bool) {
	match := durationTokenRE.FindStringSubmatch(token)
	if match == nil || token == "" {
		return 0, false
	}
	seconds := 0
	seen := false
	multipliers := []int{3600, 60, 1}
	for i, multiplier := range multipliers {
		if match[i+1] == "" {
			continue
		}
		value, _ := strconv.Atoi(match[i+1])
		seconds += value * multiplier
		seen = true
	}
	return seconds, seen && seconds > 0
}

func parseDistanceToken(token string) (Length, bool) {
	match := distanceTokenRE.FindStringSubmatch(token)
	if match == nil {
		return Length{}, false
	}
	value, _ := strconv.ParseFloat(match[1], 64)
	return Length{Value: value, Unit: match[2]}, true
}

func parseCadenceAtEnd(tokens []string) (*Target, bool) {
	if len(tokens) == 0 {
		return nil, false
	}
	match := rpmTokenRE.FindStringSubmatch(tokens[len(tokens)-1])
	if match == nil {
		return nil, false
	}
	return targetFromRegex(match, "RPM"), true
}

func parseNumberTarget(token string, units string) (*Target, error) {
	value, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid numeric target %q", token)
	}
	return &Target{Value: &value, Units: units}, nil
}

func targetForStep(step *Step, target *Target) *Target {
	if !step.Ramp || target == nil || target.Min == nil || target.Max == nil {
		return target
	}
	return &Target{Start: target.Min, End: target.Max, Units: target.Units, Text: target.Text}
}

func targetFromRegex(match []string, units string) *Target {
	lo, _ := strconv.ParseFloat(match[1], 64)
	if len(match) > 2 && match[2] != "" {
		hi, _ := strconv.ParseFloat(match[2], 64)
		return &Target{Min: &lo, Max: &hi, Units: units}
	}
	return &Target{Value: &lo, Units: units}
}

func leadingDepth(line string) int {
	spaces := 0
	for _, r := range line {
		if r != ' ' {
			break
		}
		spaces++
	}
	return spaces / 2
}
