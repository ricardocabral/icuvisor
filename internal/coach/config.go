package coach

import (
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/toolcatalog"
)

// Mode identifies the requested coach-mode feature flag state.
type Mode string

const (
	ModeOff  Mode = "off"
	ModeOn   Mode = "on"
	ModeAuto Mode = "auto"
)

// Config contains normalized coach-mode roster and ACL settings.
type Config struct {
	Athletes         []Athlete `json:"athletes"`
	DefaultAthleteID string    `json:"default_athlete_id"`
}

// Athlete contains one normalized roster entry and its per-athlete ACL patterns.
type Athlete struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	AllowedTools []string `json:"allowed_tools"`
	DeniedTools  []string `json:"denied_tools"`
}

// NormalizeAthleteIDFunc normalizes an athlete ID without making coach import config.
type NormalizeAthleteIDFunc func(string) (string, error)

// ParseMode parses ICUVISOR_COACH_MODE. Empty values default to off.
func ParseMode(value string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "":
		return ModeOff, nil
	case string(ModeOff):
		return ModeOff, nil
	case string(ModeOn):
		return ModeOn, nil
	case string(ModeAuto):
		return ModeAuto, nil
	default:
		return "", fmt.Errorf("invalid coach mode %q; use off, on, or auto", strings.TrimSpace(value))
	}
}

// EffectiveMode resolves auto based on whether the normalized roster is non-empty.
func EffectiveMode(mode Mode, cfg Config) Mode {
	if mode == ModeAuto {
		if len(cfg.Athletes) > 0 {
			return ModeOn
		}
		return ModeOff
	}
	if mode == "" {
		return ModeOff
	}
	return mode
}

// ValidateConfig normalizes and validates a coach config stanza for typo defense.
func ValidateConfig(raw Config, mode Mode, normalize NormalizeAthleteIDFunc) (Config, error) {
	if normalize == nil {
		return Config{}, fmt.Errorf("validating coach config: missing athlete ID normalizer")
	}

	out := Config{Athletes: make([]Athlete, 0, len(raw.Athletes))}
	seen := make(map[string]struct{}, len(raw.Athletes))
	for i, athlete := range raw.Athletes {
		normalizedID, err := normalize(athlete.ID)
		if err != nil {
			return Config{}, fmt.Errorf("invalid coach.athletes[%d].id; intervals.icu IDs are digits, optionally with a leading 'i', e.g. i12345 or 12345", i)
		}
		if _, ok := seen[normalizedID]; ok {
			return Config{}, fmt.Errorf("duplicate coach athlete id %q", normalizedID)
		}
		seen[normalizedID] = struct{}{}

		allowed, err := normalizeToolPatterns(fmt.Sprintf("coach.athletes[%d].allowed_tools", i), athlete.AllowedTools)
		if err != nil {
			return Config{}, err
		}
		denied, err := normalizeToolPatterns(fmt.Sprintf("coach.athletes[%d].denied_tools", i), athlete.DeniedTools)
		if err != nil {
			return Config{}, err
		}

		out.Athletes = append(out.Athletes, Athlete{
			ID:           normalizedID,
			Label:        strings.TrimSpace(athlete.Label),
			AllowedTools: allowed,
			DeniedTools:  denied,
		})
	}

	defaultID := strings.TrimSpace(raw.DefaultAthleteID)
	if defaultID == "" && len(out.Athletes) == 1 {
		defaultID = out.Athletes[0].ID
	}
	if defaultID != "" {
		normalizedDefault, err := normalize(defaultID)
		if err != nil {
			return Config{}, fmt.Errorf("invalid coach.default_athlete_id; intervals.icu IDs are digits, optionally with a leading 'i', e.g. i12345 or 12345")
		}
		if _, ok := seen[normalizedDefault]; !ok {
			return Config{}, fmt.Errorf("coach.default_athlete_id must be present in coach.athletes")
		}
		out.DefaultAthleteID = normalizedDefault
	}

	effective := EffectiveMode(mode, out)
	if mode == ModeOn && len(out.Athletes) == 0 {
		return Config{}, fmt.Errorf("coach mode is on but coach.athletes is empty")
	}
	if effective == ModeOn && len(out.Athletes) > 1 && out.DefaultAthleteID == "" {
		return Config{}, fmt.Errorf("coach.default_athlete_id is required when coach mode has multiple athletes")
	}
	return out, nil
}

func normalizeToolPatterns(field string, values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		pattern, err := toolcatalog.ValidateACLPattern(value)
		if err != nil {
			return nil, fmt.Errorf("invalid %s[%d]: %w", field, i, err)
		}
		if _, ok := seen[pattern]; ok {
			continue
		}
		seen[pattern] = struct{}{}
		out = append(out, pattern)
	}
	return out, nil
}
