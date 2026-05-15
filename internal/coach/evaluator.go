package coach

import (
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

// Evaluator applies per-athlete coach ACLs to athlete-scoped tool names.
type Evaluator struct {
	enabled bool
	roster  map[string]Athlete
}

// NewEvaluator builds an immutable evaluator from normalized coach config.
func NewEvaluator(enabled bool, cfg Config) Evaluator {
	roster := make(map[string]Athlete, len(cfg.Athletes))
	for _, athlete := range cfg.Athletes {
		roster[athlete.ID] = athlete
	}
	return Evaluator{enabled: enabled, roster: roster}
}

// Enabled reports whether the evaluator is enforcing coach ACLs.
func (e Evaluator) Enabled() bool {
	return e.enabled
}

// HasAthlete reports whether athleteID is in the normalized coach roster.
func (e Evaluator) HasAthlete(athleteID string) bool {
	_, ok := e.roster[strings.TrimSpace(athleteID)]
	return ok
}

// Evaluate returns whether toolName is allowed for athleteID and a terse reason.
func (e Evaluator) Evaluate(athleteID, toolName string) (bool, string) {
	if !e.enabled || !toolcatalog.IsAthleteScopedTool(toolName) {
		return true, "coach_acl_not_applicable"
	}
	athlete, ok := e.roster[strings.TrimSpace(athleteID)]
	if !ok {
		return false, "athlete_not_in_roster"
	}
	if matchesAny(athlete.DeniedTools, toolName) {
		return false, "tool_denied_for_athlete"
	}
	if !matchesAny(athlete.AllowedTools, toolName) {
		return false, "tool_not_allowed_for_athlete"
	}
	return true, "allowed"
}

// AllowedForAny reports whether any roster athlete allows toolName.
func (e Evaluator) AllowedForAny(toolName string) bool {
	if !e.enabled || !toolcatalog.IsAthleteScopedTool(toolName) {
		return true
	}
	for athleteID := range e.roster {
		if allowed, _ := e.Evaluate(athleteID, toolName); allowed {
			return true
		}
	}
	return false
}

// MustEvaluate returns an error when Evaluate denies the tool.
func (e Evaluator) MustEvaluate(athleteID, toolName string) error {
	allowed, reason := e.Evaluate(athleteID, toolName)
	if allowed {
		return nil
	}
	return fmt.Errorf("coach ACL denied %s: %s", toolName, reason)
}

func matchesAny(patterns []string, toolName string) bool {
	for _, pattern := range patterns {
		if matchPattern(pattern, toolName) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, toolName string) bool {
	pattern = strings.TrimSpace(pattern)
	switch {
	case pattern == "*":
		return true
	case strings.HasSuffix(pattern, "*"):
		return strings.HasPrefix(toolName, strings.TrimSuffix(pattern, "*"))
	default:
		return pattern == toolName
	}
}
