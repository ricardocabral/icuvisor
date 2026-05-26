package workoutdoc

import (
	"fmt"
	"strings"
)

// StructuralTokenInDescriptionError reports a structured step whose free-text
// label contains a token the Intervals.icu DSL treats as duration or distance.
type StructuralTokenInDescriptionError struct {
	Step    Step
	Token   string
	Kind    string
	Context string
}

func (e *StructuralTokenInDescriptionError) Error() string {
	if e == nil {
		return "step description contains a structural token"
	}
	context := e.Context
	if context == "" {
		context = "step"
	}
	kind := e.Kind
	if kind == "" {
		kind = "duration or distance"
	}
	return fmt.Sprintf("%s description contains %s token %q; put duration/distance in structured fields, not in description", context, kind, e.Token)
}

func descriptionStructuralTokenError(step Step, context string) error {
	if token, kind, ok := structuralTokenInDescription(step.Description); ok {
		return &StructuralTokenInDescriptionError{Step: step, Token: token, Kind: kind, Context: context}
	}
	return nil
}

func structuralTokenInDescription(description string) (token string, kind string, ok bool) {
	for _, raw := range strings.Fields(description) {
		candidate := normalizeDescriptionToken(raw)
		if candidate == "" {
			continue
		}
		lower := strings.ToLower(candidate)
		if _, parsed := parseDurationToken(lower); parsed {
			return candidate, "duration", true
		}
		if _, parsed := parseDistanceToken(lower); parsed {
			return candidate, "distance", true
		}
	}
	return "", "", false
}

func normalizeDescriptionToken(token string) string {
	return strings.Trim(token, " \t\r\n.,;:!?()[]{}<>\"'`")
}
