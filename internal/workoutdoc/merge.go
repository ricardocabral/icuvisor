package workoutdoc

import (
	"strings"
)

// StepsSentinel is the literal placeholder a caller may embed in a prose
// description to control where serialized structured steps are inserted by
// MergeDescription. The line containing the sentinel is replaced verbatim by
// the serialized step block; surrounding prose lines keep their original
// position and content.
const StepsSentinel = "<!-- icuvisor:steps -->"

// MergeDescription combines free-text prose with the canonical DSL serialization
// of a structured WorkoutDoc into the single description string that
// intervals.icu accepts on write.
//
// Placement rule:
//
//  1. If prose contains a line whose trimmed content equals StepsSentinel,
//     the sentinel line is replaced by the serialized step block in place.
//     Prose lines before and after the sentinel retain their order verbatim.
//  2. Otherwise the serialized step block is appended after all prose, with a
//     single blank line between the prose and the steps when prose is
//     non-empty and does not already end in a newline.
//  3. When doc has no steps, the original prose is returned untouched.
//  4. When prose is empty, only the serialized step block is returned.
//
// Prose is never reformatted, trimmed, or otherwise rewritten — line endings
// are preserved as-is apart from \r\n being normalized to \n.
func MergeDescription(prose string, doc WorkoutDoc) (string, error) {
	prose = strings.ReplaceAll(prose, "\r\n", "\n")
	var dsl string
	if len(doc.Steps) > 0 {
		serialized, err := Serialize(doc)
		if err != nil {
			return "", err
		}
		dsl = serialized
	}
	if dsl == "" {
		return prose, nil
	}
	if prose == "" {
		return dsl, nil
	}

	lines := strings.Split(prose, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == StepsSentinel {
			out := make([]string, 0, len(lines))
			out = append(out, lines[:i]...)
			out = append(out, dsl)
			out = append(out, lines[i+1:]...)
			return strings.Join(out, "\n"), nil
		}
	}

	if strings.HasSuffix(prose, "\n") {
		return prose + dsl, nil
	}
	return prose + "\n\n" + dsl, nil
}
