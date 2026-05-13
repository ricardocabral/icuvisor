package intervals

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// WriteSportSettingsParams contains sparse sport-setting fields for one sport.
type WriteSportSettingsParams struct {
	SportSettingID int
	EffectiveDate  string
	FTP            *int
	ThresholdHR    *int
	ThresholdPace  *SportSettingsPace
	ZonesProvided  bool
	Zones          []SportSettingsZoneDefinition
}

// SportSettingsPace contains a threshold pace value in an intervals.icu pace unit.
type SportSettingsPace struct {
	Value float64
	Unit  string
}

// SportSettingsZoneDefinition contains one replacement zone set for a sport-setting metric.
type SportSettingsZoneDefinition struct {
	Kind       string
	Boundaries []float64
	Names      []string
}

// UpdateSportSettings updates sparse sport settings and applies them from the effective date when supplied.
func (c *Client) UpdateSportSettings(ctx context.Context, params WriteSportSettingsParams) (SportSettings, error) {
	if params.SportSettingID <= 0 {
		return SportSettings{}, fmt.Errorf("updating sport settings: sport setting ID is required")
	}
	body, err := writeSportSettingsBody(params)
	if err != nil {
		return SportSettings{}, err
	}
	var setting SportSettings
	id := strconv.Itoa(params.SportSettingID)
	if err := c.doJSONBody(ctx, http.MethodPut, body, &setting, "athlete", c.athleteID, "sport-settings", id); err != nil {
		return SportSettings{}, fmt.Errorf("updating sport settings %s: %w", id, err)
	}
	if effectiveDate := strings.TrimSpace(params.EffectiveDate); effectiveDate != "" {
		if err := c.ApplySportSettings(ctx, params.SportSettingID, effectiveDate); err != nil {
			return SportSettings{}, err
		}
	}
	return setting, nil
}

// ApplySportSettings asks upstream to recompute activities affected by a sport-setting change.
func (c *Client) ApplySportSettings(ctx context.Context, sportSettingID int, oldest string) error {
	if sportSettingID <= 0 {
		return fmt.Errorf("applying sport settings: sport setting ID is required")
	}
	body := map[string]any{}
	if oldest = strings.TrimSpace(oldest); oldest != "" {
		body["oldest"] = oldest
	}
	var response map[string]any
	id := strconv.Itoa(sportSettingID)
	if err := c.doJSONBody(ctx, http.MethodPut, body, &response, "athlete", c.athleteID, "sport-settings", id, "apply"); err != nil {
		return fmt.Errorf("applying sport settings %s: %w", id, err)
	}
	return nil
}

func writeSportSettingsBody(params WriteSportSettingsParams) (map[string]any, error) {
	body := map[string]any{}
	setSparse(body, "ftp", params.FTP)
	setSparse(body, "lthr", params.ThresholdHR)
	if params.ThresholdPace != nil {
		body["threshold_pace"] = params.ThresholdPace.Value
		if unit := strings.TrimSpace(params.ThresholdPace.Unit); unit != "" {
			body["pace_units"] = unit
		}
	}
	if params.ZonesProvided {
		for _, zone := range params.Zones {
			applySportSettingsZone(body, zone)
		}
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("updating sport settings: at least one threshold or zone field is required")
	}
	return body, nil
}

func applySportSettingsZone(body map[string]any, zone SportSettingsZoneDefinition) {
	kind := strings.ToLower(strings.TrimSpace(zone.Kind))
	switch kind {
	case "power":
		body["power_zones"] = roundedZoneBoundaries(zone.Boundaries)
		if len(zone.Names) > 0 {
			body["power_zone_names"] = append([]string(nil), zone.Names...)
		}
	case "hr", "heart_rate":
		body["hr_zones"] = roundedZoneBoundaries(zone.Boundaries)
		if len(zone.Names) > 0 {
			body["hr_zone_names"] = append([]string(nil), zone.Names...)
		}
	case "pace":
		body["pace_zones"] = append([]float64(nil), zone.Boundaries...)
		if len(zone.Names) > 0 {
			body["pace_zone_names"] = append([]string(nil), zone.Names...)
		}
	}
}

func roundedZoneBoundaries(values []float64) []int {
	out := make([]int, 0, len(values))
	for _, value := range values {
		out = append(out, int(value))
	}
	return out
}
