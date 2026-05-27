package response

import (
	"fmt"
	"strings"
	"time"
)

// AsOfMetadata reports the athlete-local current-time anchor for time-relative reads.
type AsOfMetadata struct {
	AsOf        string `json:"as_of"`
	AsOfDate    string `json:"as_of_date"`
	AsOfWeekday string `json:"as_of_weekday"`
	Timezone    string `json:"timezone"`
}

// AsOfMetadataInTimezone renders t as athlete-local as-of metadata in the configured IANA timezone.
func AsOfMetadataInTimezone(t time.Time, timezone string) (AsOfMetadata, error) {
	loc, zone, err := athleteLocation(timezone)
	if err != nil {
		return AsOfMetadata{}, err
	}
	local := t.In(loc)
	return AsOfMetadata{
		AsOf:        local.Format(time.RFC3339),
		AsOfDate:    local.Format(time.DateOnly),
		AsOfWeekday: local.Weekday().String(),
		Timezone:    zone,
	}, nil
}

// RenderTimeInTimezone renders t in the athlete's configured IANA timezone.
func RenderTimeInTimezone(t time.Time, timezone string) (string, error) {
	loc, _, err := athleteLocation(timezone)
	if err != nil {
		return "", err
	}
	return t.In(loc).Format(time.RFC3339), nil
}

// RenderDateInTimezone renders the calendar date for t in the athlete's configured IANA timezone.
func RenderDateInTimezone(t time.Time, timezone string) (string, error) {
	loc, _, err := athleteLocation(timezone)
	if err != nil {
		return "", err
	}
	return t.In(loc).Format(time.DateOnly), nil
}

func athleteLocation(timezone string) (*time.Location, string, error) {
	zone := strings.TrimSpace(timezone)
	if zone == "" {
		zone = "UTC"
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return nil, "", fmt.Errorf("loading athlete timezone %q: %w", zone, err)
	}
	return loc, zone, nil
}
