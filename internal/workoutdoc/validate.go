package workoutdoc

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Diagnostic is a single validation finding emitted by ValidateDescription or
// ValidateDoc. Codes are stable, machine-readable strings; messages are short
// and meant for humans (or the LLM). StepIndex is the zero-based index of the
// top-level structured step that triggered the finding, or nil for
// whole-input findings. Line is the one-based line number within the input
// description, or nil when the finding is not tied to a specific line.
type Diagnostic struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	StepIndex *int   `json:"step_index"`
	Line      *int   `json:"line"`
}

// ValidationResult is the aggregate result returned by ValidateDescription and
// ValidateDoc.
type ValidationResult struct {
	// Doc is the parsed structured workout for the structured-step lines found
	// in the input. Doc.Steps is nil when the input contains no
	// recognizable structured-step lines.
	Doc WorkoutDoc
	// Prose holds the input lines that were not structured-step lines, in
	// their original order, joined by "\n". A StepsSentinel line is inserted
	// at the position where the structured-step block began so MergeDescription
	// can later reinsert serialized steps in the same place.
	Prose string
	// StructuredStepLines is the count of input lines recognized as structured
	// (step or repeat-header lines).
	StructuredStepLines int
	// Errors are unrecoverable findings: the input contains malformed
	// structured-step lines or a workout that cannot be serialized.
	Errors []Diagnostic
	// Warnings are recoverable findings worth surfacing to the LLM.
	Warnings []Diagnostic
}

// repeatHeaderRE matches a "Nx" repeat header, optionally prefixed by a
// description like "Main set 3x". Mirrors the parser's repeatLineRE.
var (
	repeatHeaderRE          = regexp.MustCompile(`^(?:.*?\S\s+)?[1-9][0-9]*x$`)
	malformedRepeatHeaderRE = regexp.MustCompile(`^-\s*[1-9][0-9]*\s*x$`)
)

// IsStructuredStepLine reports whether a single description line would be
// treated by the intervals.icu DSL parser as a structured-step line (either a
// "- ..." step or an "Nx" repeat header). Repeat-looking dashed headers are
// also treated as structured so validation can reject them instead of passing
// them through as prose.
func IsStructuredStepLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "- ") {
		return true
	}
	return repeatHeaderRE.MatchString(trimmed) || malformedRepeatHeaderRE.MatchString(trimmed)
}

// ValidateDescription scans a free-text intervals.icu description, validates
// the structured-step regions using Parse, and returns prose lines untouched.
//
// Unparseable lines that look like prose (no leading "- ", no "Nx" header) are
// never reported as errors. Only malformed structured-step lines surface as
// PARSE_ERROR diagnostics.
//
// Warnings cover MISSING_PRIMARY_TARGET, MIXED_PRIMARY_TARGETS, EMPTY_REPEAT_BLOCK,
// NESTED_REPEAT, and M_AMBIGUITY (bare "m" minutes token whose magnitude looks
// like meters).
func ValidateDescription(description string) ValidationResult {
	result := ValidationResult{}
	if description == "" {
		return result
	}
	lines := strings.Split(strings.ReplaceAll(description, "\r\n", "\n"), "\n")

	stepBlocks, proseLines := splitStepBlocks(lines)
	result.StructuredStepLines = sumStepBlockLines(stepBlocks)
	result.Prose = strings.Join(proseLines, "\n")

	if len(stepBlocks) == 0 {
		return result
	}
	if len(stepBlocks) > 1 {
		first := stepBlocks[0]
		line := first.startLine
		result.Errors = append(result.Errors, Diagnostic{
			Code:    "PARSE_ERROR",
			Message: "structured workout step lines must form a single contiguous block; multiple blocks separated by prose are not supported by the intervals.icu DSL",
			Line:    &line,
		})
		return result
	}

	block := stepBlocks[0]
	doc, err := Parse(block.text())
	if err != nil {
		startLine := block.startLine
		result.Errors = append(result.Errors, Diagnostic{
			Code:    "PARSE_ERROR",
			Message: err.Error(),
			Line:    &startLine,
		})
		return result
	}
	result.Doc = doc

	for index, step := range doc.Steps {
		idx := index
		collectStepDiagnostics(step, &idx, nil, &result)
	}
	collectMAmbiguity(block, &result)
	return result
}

// ValidateDoc validates a structured WorkoutDoc without any associated prose.
// It mirrors the warnings emitted by ValidateDescription for the structured
// portion. Errors from this function come from Serialize, including any
// *UnsupportedStepError.
func ValidateDoc(doc WorkoutDoc) ValidationResult {
	result := ValidationResult{Doc: doc}
	for index, step := range doc.Steps {
		idx := index
		collectStepDiagnostics(step, &idx, nil, &result)
	}
	dsl, err := Serialize(doc)
	if err != nil {
		code := serializeDiagnosticCode(err)
		if code != "STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION" || !hasDiagnosticCode(result.Errors, code) {
			result.Errors = append(result.Errors, Diagnostic{Code: code, Message: err.Error()})
		}
		return result
	}
	if _, err := Parse(dsl); err != nil {
		result.Errors = append(result.Errors, Diagnostic{Code: "SERIALIZED_DSL_PARSE_ERROR", Message: "serialized workout_doc did not parse: " + err.Error()})
		return result
	}
	result.StructuredStepLines = len(strings.Split(dsl, "\n"))
	return result
}

func serializeDiagnosticCode(err error) string {
	var structural *StructuralTokenInDescriptionError
	if errors.As(err, &structural) {
		return "STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION"
	}
	var unsupported *UnsupportedStepError
	if errors.As(err, &unsupported) {
		return "UNSUPPORTED_STEP"
	}
	return "SERIALIZE_ERROR"
}

func hasDiagnosticCode(diags []Diagnostic, code string) bool {
	for _, diag := range diags {
		if diag.Code == code {
			return true
		}
	}
	return false
}

// EstimateDurationSeconds returns the total scheduled duration of doc's
// timed steps in seconds, ignoring distance-only steps. Repeat blocks
// multiply their children by their reps count.
func EstimateDurationSeconds(doc WorkoutDoc) int {
	total := 0
	for _, step := range doc.Steps {
		total += estimateStepDuration(step)
	}
	return total
}

func estimateStepDuration(step Step) int {
	if step.Reps > 0 && len(step.Steps) > 0 {
		inner := 0
		for _, child := range step.Steps {
			inner += estimateStepDuration(child)
		}
		return inner * step.Reps
	}
	return step.Duration
}

func collectStepDiagnostics(step Step, topIndex *int, parentReps *int, result *ValidationResult) {
	if err := descriptionStructuralTokenError(step, "step"); err != nil {
		result.Errors = append(result.Errors, Diagnostic{
			Code:      "STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION",
			Message:   err.Error(),
			StepIndex: topIndex,
		})
	}
	if step.Reps > 0 || len(step.Steps) > 0 {
		if parentReps != nil {
			result.Errors = append(result.Errors, Diagnostic{
				Code:      "NESTED_REPEAT",
				Message:   "nested repeat blocks are not supported by the intervals.icu workout DSL",
				StepIndex: topIndex,
			})
			return
		}
		if len(step.Steps) == 0 {
			result.Warnings = append(result.Warnings, Diagnostic{
				Code:      "EMPTY_REPEAT_BLOCK",
				Message:   "repeat block has no child steps",
				StepIndex: topIndex,
			})
		}
		reps := step.Reps
		for _, child := range step.Steps {
			collectStepDiagnostics(child, topIndex, &reps, result)
		}
		return
	}

	primaries := 0
	if step.Power != nil {
		primaries++
	}
	if step.HR != nil {
		primaries++
	}
	if step.Pace != nil {
		primaries++
	}
	if primaries > 1 {
		result.Warnings = append(result.Warnings, Diagnostic{
			Code:      "MIXED_PRIMARY_TARGETS",
			Message:   "step has more than one of power, heart rate, or pace as a primary target; intervals.icu supports only one",
			StepIndex: topIndex,
		})
	}
	if primaries == 0 && step.RPE == nil && !step.Freeride && (step.Duration > 0 || step.Distance != nil) {
		result.Warnings = append(result.Warnings, Diagnostic{
			Code:      "MISSING_PRIMARY_TARGET",
			Message:   "step has duration or distance but no power, heart rate, pace, or RPE target",
			StepIndex: topIndex,
		})
	}
}

// stepBlock represents a contiguous run of structured-step lines extracted
// from the input description, with the original line numbers preserved so
// diagnostics can point back at the source.
type stepBlock struct {
	lines     []string
	startLine int // 1-based
}

func (b stepBlock) text() string { return strings.Join(b.lines, "\n") }

func splitStepBlocks(lines []string) ([]stepBlock, []string) {
	var blocks []stepBlock
	var current stepBlock
	prose := make([]string, 0, len(lines))
	sentinelEmitted := false
	flush := func() {
		if len(current.lines) > 0 {
			blocks = append(blocks, current)
			current = stepBlock{}
		}
	}
	for i, line := range lines {
		if IsStructuredStepLine(line) {
			if len(current.lines) == 0 {
				current.startLine = i + 1
			}
			current.lines = append(current.lines, line)
			if !sentinelEmitted {
				prose = append(prose, StepsSentinel)
				sentinelEmitted = true
			}
			continue
		}
		flush()
		prose = append(prose, line)
	}
	flush()
	return blocks, prose
}

func sumStepBlockLines(blocks []stepBlock) int {
	total := 0
	for _, block := range blocks {
		total += len(block.lines)
	}
	return total
}

func collectMAmbiguity(block stepBlock, result *ValidationResult) {
	bareMRE := regexp.MustCompile(`(?:^|\s)([1-9][0-9]{2,})m(?:\s|$)`)
	for offset, line := range block.lines {
		if matches := bareMRE.FindStringSubmatch(line); matches != nil {
			value, _ := strconv.Atoi(matches[1])
			lineNum := block.startLine + offset
			result.Warnings = append(result.Warnings, Diagnostic{
				Code:    "M_AMBIGUITY",
				Message: fmt.Sprintf("%dm parsed as %d minutes; intervals.icu uses 'm' for minutes and 'mtr' for meters", value, value),
				Line:    &lineNum,
			})
		}
	}
}
