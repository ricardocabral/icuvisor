package response

import (
	"fmt"
	"strings"
	"time"
)

// RenderTimeInTimezone renders t in the athlete's configured IANA timezone.
func RenderTimeInTimezone(t time.Time, timezone string) (string, error) {
	loc, err := athleteLocation(timezone)
	if err != nil {
		return "", err
	}
	return t.In(loc).Format(time.RFC3339), nil
}

// RenderDateInTimezone renders the calendar date for t in the athlete's configured IANA timezone.
func RenderDateInTimezone(t time.Time, timezone string) (string, error) {
	loc, err := athleteLocation(timezone)
	if err != nil {
		return "", err
	}
	return t.In(loc).Format(time.DateOnly), nil
}

func athleteLocation(timezone string) (*time.Location, error) {
	zone := strings.TrimSpace(timezone)
	if zone == "" {
		zone = "UTC"
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("loading athlete timezone %q: %w", zone, err)
	}
	return loc, nil
}
