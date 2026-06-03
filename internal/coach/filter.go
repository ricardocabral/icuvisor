package coach

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

var (
	// ErrInvalidAthleteID reports a malformed athlete_id routing argument.
	ErrInvalidAthleteID = errors.New("invalid athlete_id")
	// ErrAthleteNotAuthorized reports a target that is not in the configured coach roster.
	ErrAthleteNotAuthorized = errors.New("athlete_id not authorized")
	// ErrToolNotAllowed reports a per-athlete ACL denial for the selected target.
	ErrToolNotAllowed = errors.New("tool not allowed for athlete")
)

// AthleteIDNormalizer canonicalizes an athlete ID string.
type AthleteIDNormalizer func(string) (string, error)

// ToolFilter owns coach-mode tool visibility and target authorization decisions.
type ToolFilter struct {
	evaluator Evaluator
}

// NewToolFilter creates a coach tool visibility filter from evaluator.
func NewToolFilter(evaluator Evaluator) ToolFilter {
	return ToolFilter{evaluator: evaluator}
}

// VisibleForAthlete reports whether toolName is visible for athleteID under coach ACLs.
func (f ToolFilter) VisibleForAthlete(athleteID string, toolName string) bool {
	if toolName == toolcatalog.ListAthletes || toolName == toolcatalog.SelectAthlete || toolName == toolcatalog.ICUvisorListAdvancedCapabilities {
		return true
	}
	if !toolcatalog.IsAthleteScopedTool(toolName) {
		return true
	}
	allowed, _ := f.evaluator.Evaluate(athleteID, toolName)
	return allowed
}

// VisibleToolNamesForAthlete returns the sorted subset of toolNames visible for athleteID.
func (f ToolFilter) VisibleToolNamesForAthlete(athleteID string, toolNames []string) []string {
	out := make([]string, 0, len(toolNames))
	for _, toolName := range toolNames {
		if f.VisibleForAthlete(athleteID, toolName) {
			out = append(out, toolName)
		}
	}
	sort.Strings(out)
	return out
}

// AllowedForAny reports whether at least one configured athlete may use toolName.
func (f ToolFilter) AllowedForAny(toolName string) bool {
	return f.evaluator.AllowedForAny(toolName)
}

// ResolveTarget selects, normalizes, and authorizes the request target athlete for toolName.
func (f ToolFilter) ResolveTarget(suppliedAthleteID, defaultAthleteID, selectedAthleteID, toolName string, normalize AthleteIDNormalizer) (string, error) {
	if normalize == nil {
		return "", errors.New("missing athlete ID normalizer")
	}
	targetAthleteID := strings.TrimSpace(suppliedAthleteID)
	if targetAthleteID == "" {
		targetAthleteID = strings.TrimSpace(selectedAthleteID)
	}
	if targetAthleteID == "" {
		targetAthleteID = strings.TrimSpace(defaultAthleteID)
	}
	normalized, err := normalize(targetAthleteID)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidAthleteID, err)
	}
	if !f.evaluator.HasAthlete(normalized) {
		return "", ErrAthleteNotAuthorized
	}
	allowed, reason := f.evaluator.Evaluate(normalized, toolName)
	if !allowed {
		return "", fmt.Errorf("%w: %s", ErrToolNotAllowed, reason)
	}
	return normalized, nil
}
