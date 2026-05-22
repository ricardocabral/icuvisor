package config

import (
	"errors"
	"strings"
	"unicode"
)

// NormalizeAthleteID validates intervals.icu athlete IDs.
//
// intervals.icu issues two ID shapes: the `i<digits>` form for accounts created on
// intervals.icu, and a bare-numeric form for accounts linked from Strava (the Strava
// athlete number). Both are accepted. The leading `i` is part of the ID, not a display
// convenience, so a bare-numeric ID is never rewritten with one — doing so would point
// at a different (or nonexistent) athlete.
func NormalizeAthleteID(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("missing athlete ID; set INTERVALS_ICU_ATHLETE_ID or athlete_id (find it in the intervals.icu URL, e.g. i12345 or 12345)")
	}

	digits := trimmed
	prefixed := false
	if trimmed[0] == 'i' || trimmed[0] == 'I' {
		digits = trimmed[1:]
		prefixed = true
	}
	if digits == "" {
		return "", invalidAthleteIDError()
	}
	for _, r := range digits {
		if !unicode.IsDigit(r) {
			return "", invalidAthleteIDError()
		}
	}
	if prefixed {
		return "i" + digits, nil
	}
	return digits, nil
}

func invalidAthleteIDError() error {
	return errors.New("invalid athlete ID; intervals.icu IDs are digits, optionally with a leading 'i', e.g. i12345 or 12345 (find yours in the intervals.icu URL)")
}

// NormalizeAthleteIDForDisplay returns the canonical public athlete ID when possible.
func NormalizeAthleteIDForDisplay(value string) string {
	normalized, err := NormalizeAthleteID(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return normalized
}
