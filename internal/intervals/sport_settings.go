package intervals

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// WriteSportSettingsParams contains sparse sport-setting fields for one sport.
type WriteSportSettingsParams struct {
	SportSettingID int
	RecalcHRZones  bool
	FTP            *int
	IndoorFTP      *int
	ThresholdHR    *int
	ThresholdPace  *SportSettingsPace
	ZonesProvided  bool
	Zones          []SportSettingsZoneDefinition
}

// CreateSportSettingsParams contains threshold-only fields for one new sport setting.
type CreateSportSettingsParams struct {
	Sport         string
	FTP           *int
	IndoorFTP     *int
	ThresholdHR   *int
	ThresholdPace *SportSettingsPace
}

// SportSettingsPace contains a threshold pace stored in m/s with independent display metadata.
type SportSettingsPace struct {
	Value        float64
	PaceUnits    string
	PaceLoadType string
}

// SportSettingsZoneDefinition contains one replacement zone set for a sport-setting metric.
type SportSettingsZoneDefinition struct {
	Kind       string
	Boundaries []float64
	Names      []string
}

// UpdateSportSettings updates sparse sport settings.
func (c *Client) UpdateSportSettings(ctx context.Context, params WriteSportSettingsParams) (SportSettings, error) {
	if params.SportSettingID <= 0 {
		return SportSettings{}, fmt.Errorf("updating sport settings: sport setting ID is required")
	}
	if err := validateSportSettingsThresholds("updating", params.FTP, params.IndoorFTP, params.ThresholdHR, params.ThresholdPace); err != nil {
		return SportSettings{}, err
	}
	body, err := writeSportSettingsBody(params)
	if err != nil {
		return SportSettings{}, err
	}
	var setting SportSettings
	id := strconv.Itoa(params.SportSettingID)
	query := url.Values{"recalcHrZones": []string{strconv.FormatBool(params.RecalcHRZones)}}
	if err := c.doJSONBodyQuery(ctx, http.MethodPut, body, &setting, query, "athlete", c.athleteID, "sport-settings", id); err != nil {
		return SportSettings{}, fmt.Errorf("updating sport settings %s: %w", id, err)
	}
	return setting, nil
}

// CreateSportSettings creates threshold-only settings for one sport.
func (c *Client) CreateSportSettings(ctx context.Context, params CreateSportSettingsParams) (SportSettings, error) {
	body, err := createSportSettingsBody(params)
	if err != nil {
		return SportSettings{}, err
	}
	var setting SportSettings
	if err := c.doJSONBody(ctx, http.MethodPost, body, &setting, "athlete", c.athleteID, "sport-settings"); err != nil {
		return SportSettings{}, fmt.Errorf("creating sport settings: %w", err)
	}
	return setting, nil
}

// ApplySportSettings asks upstream to recompute activities affected by a sport-setting change.
func (c *Client) ApplySportSettings(ctx context.Context, sportSettingID int) error {
	if sportSettingID <= 0 {
		return fmt.Errorf("applying sport settings: sport setting ID is required")
	}
	var response map[string]any
	id := strconv.Itoa(sportSettingID)
	if err := c.doJSONNoBody(ctx, http.MethodPut, &response, "athlete", c.athleteID, "sport-settings", id, "apply"); err != nil {
		return fmt.Errorf("applying sport settings %s: %w", id, err)
	}
	return nil
}

func writeSportSettingsBody(params WriteSportSettingsParams) (map[string]any, error) {
	body := map[string]any{}
	setSportSettingsThresholdFields(body, params.FTP, params.IndoorFTP, params.ThresholdHR, params.ThresholdPace)
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

func createSportSettingsBody(params CreateSportSettingsParams) (map[string]any, error) {
	sport := strings.TrimSpace(params.Sport)
	if sport == "" {
		return nil, fmt.Errorf("creating sport settings: sport is required")
	}
	if err := validateSportSettingsThresholds("creating", params.FTP, params.IndoorFTP, params.ThresholdHR, params.ThresholdPace); err != nil {
		return nil, err
	}
	body := map[string]any{"types": []string{sport}}
	setSportSettingsThresholdFields(body, params.FTP, params.IndoorFTP, params.ThresholdHR, params.ThresholdPace)
	return body, nil
}

func validateSportSettingsThresholds(operation string, ftp *int, indoorFTP *int, thresholdHR *int, thresholdPace *SportSettingsPace) error {
	if ftp != nil && *ftp <= 0 {
		return fmt.Errorf("%s sport settings: ftp must be > 0", operation)
	}
	if indoorFTP != nil && *indoorFTP <= 0 {
		return fmt.Errorf("%s sport settings: indoor FTP must be > 0", operation)
	}
	if thresholdHR != nil && *thresholdHR <= 0 {
		return fmt.Errorf("%s sport settings: threshold HR must be > 0", operation)
	}
	if thresholdPace != nil && (thresholdPace.Value <= 0 || math.IsNaN(thresholdPace.Value) || math.IsInf(thresholdPace.Value, 0)) {
		return fmt.Errorf("%s sport settings: threshold pace must be finite and > 0", operation)
	}
	return nil
}

func setSportSettingsThresholdFields(body map[string]any, ftp *int, indoorFTP *int, thresholdHR *int, thresholdPace *SportSettingsPace) {
	setSparse(body, "ftp", ftp)
	setSparse(body, "indoor_ftp", indoorFTP)
	setSparse(body, "lthr", thresholdHR)
	if thresholdPace == nil {
		return
	}
	body["threshold_pace"] = thresholdPace.Value
	if paceUnits := strings.TrimSpace(thresholdPace.PaceUnits); paceUnits != "" {
		body["pace_units"] = paceUnits
	}
	if paceLoadType := strings.TrimSpace(thresholdPace.PaceLoadType); paceLoadType != "" {
		body["pace_load_type"] = paceLoadType
	}
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
