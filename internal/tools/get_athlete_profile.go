package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getAthleteProfileName                    = "get_athlete_profile"
	getAthleteProfileDescription             = "Get the configured intervals.icu athlete profile, FTP/thresholds, zones, and sport settings. Use this for athlete identity, units, timezone, FTP, heart-rate thresholds, pace thresholds, and zone configuration; do not use it for activities, wellness, fitness trends, events, or workouts."
	invalidGetAthleteProfileArgumentsMessage = "invalid get_athlete_profile arguments; only include_full is supported"
	fetchAthleteProfileMessage               = "could not fetch athlete profile; check intervals.icu credentials and athlete ID"
)

// ProfileClient fetches athlete profile data for tools.
type ProfileClient interface {
	GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error)
}

// GetAthleteProfileRequest contains the get_athlete_profile tool arguments.
type GetAthleteProfileRequest struct {
	IncludeFull bool `json:"include_full,omitempty"`
}

// GetAthleteProfileResponse is the structured get_athlete_profile response.
type GetAthleteProfileResponse struct {
	AthleteID                   string                   `json:"athlete_id"`
	Name                        string                   `json:"name,omitempty"`
	FirstName                   string                   `json:"first_name,omitempty"`
	LastName                    string                   `json:"last_name,omitempty"`
	Timezone                    string                   `json:"timezone,omitempty"`
	Locale                      string                   `json:"locale,omitempty"`
	Units                       GetAthleteProfileUnits   `json:"units"`
	SportSettings               []GetAthleteProfileSport `json:"sport_settings,omitempty"`
	Meta                        GetAthleteProfileMeta    `json:"_meta"`
	MeasurementPreferenceSource string                   `json:"measurement_preference_source,omitempty"`
}

// GetAthleteProfileUnits describes athlete unit preferences.
type GetAthleteProfileUnits struct {
	MeasurementPreference string `json:"measurement_preference,omitempty"`
	Weight                string `json:"weight,omitempty"`
	Temperature           string `json:"temperature,omitempty"`
}

// GetAthleteProfileSport contains thresholds and zones for one sport setting.
type GetAthleteProfileSport struct {
	Types                       []string  `json:"types,omitempty"`
	FTPWatts                    int       `json:"ftp_watts,omitempty"`
	IndoorFTPWatts              int       `json:"indoor_ftp_watts,omitempty"`
	WPrimeJoules                int       `json:"w_prime_joules,omitempty"`
	PMaxWatts                   int       `json:"p_max_watts,omitempty"`
	LTHRBPM                     int       `json:"lthr_bpm,omitempty"`
	MaxHRBPM                    int       `json:"max_hr_bpm,omitempty"`
	PowerZonesWatts             []int     `json:"power_zones_watts,omitempty"`
	PowerZoneNames              []string  `json:"power_zone_names,omitempty"`
	HRZonesBPM                  []int     `json:"hr_zones_bpm,omitempty"`
	HRZoneNames                 []string  `json:"hr_zone_names,omitempty"`
	ThresholdPaceSecondsPerKM   *float64  `json:"threshold_pace_seconds_per_km,omitempty"`
	PaceZonesSecondsPerKM       []float64 `json:"pace_zones_seconds_per_km,omitempty"`
	ThresholdPaceSecondsPerMile *float64  `json:"threshold_pace_seconds_per_mile,omitempty"`
	PaceZonesSecondsPerMile     []float64 `json:"pace_zones_seconds_per_mile,omitempty"`
	PaceUnitsSource             string    `json:"pace_units_source,omitempty"`
	PaceDistanceUnit            string    `json:"pace_distance_unit,omitempty"`
	PaceZoneNames               []string  `json:"pace_zone_names,omitempty"`
	SportSettingID              int       `json:"sport_setting_id,omitempty"`
	SportSettingAthleteID       string    `json:"sport_setting_athlete_id,omitempty"`
}

// GetAthleteProfileMeta contains response-shaping metadata.
type GetAthleteProfileMeta struct {
	ServerVersion      string `json:"server_version"`
	AthleteIDFormat    string `json:"athlete_id_format"`
	TimezoneConvention string `json:"timezone_convention"`
	PaceConvention     string `json:"pace_convention"`
	IncludeFull        bool   `json:"include_full"`
}

func newGetAthleteProfileTool(client ProfileClient, version string, timezoneFallback string) Tool {
	return Tool{
		Name:         getAthleteProfileName,
		Description:  getAthleteProfileDescription,
		InputSchema:  getAthleteProfileInputSchema(),
		OutputSchema: getAthleteProfileOutputSchema(),
		Handler:      getAthleteProfileHandler(client, version, timezoneFallback),
	}
}

func getAthleteProfileHandler(client ProfileClient, version string, timezoneFallback string) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		args, err := decodeGetAthleteProfileRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetAthleteProfileArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchAthleteProfileMessage, errors.New("missing profile client"))
		}
		profile, err := client.GetAthleteProfile(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return Result{}, ctxErr
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAthleteProfileMessage, err)
		}
		response := newGetAthleteProfileResponse(profile, version, timezoneFallback, args.IncludeFull)
		text, err := json.Marshal(response)
		if err != nil {
			return Result{}, fmt.Errorf("encoding get_athlete_profile response: %w", err)
		}
		return Result{
			Content:           []Content{{Type: ContentTypeText, Text: string(text)}},
			StructuredContent: response,
		}, nil
	}
}

func decodeGetAthleteProfileRequest(raw json.RawMessage) (GetAthleteProfileRequest, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return GetAthleteProfileRequest{}, nil
	}
	if trimmed[0] != '{' {
		return GetAthleteProfileRequest{}, errors.New("arguments must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var args GetAthleteProfileRequest
	if err := decoder.Decode(&args); err != nil {
		return GetAthleteProfileRequest{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return GetAthleteProfileRequest{}, errors.New("unexpected trailing JSON")
	}
	return args, nil
}

func newGetAthleteProfileResponse(profile intervals.AthleteWithSportSettings, version string, timezoneFallback string, includeFull bool) GetAthleteProfileResponse {
	athleteID := normalizeProfileAthleteID(profile.ID)
	units := profileUnits(profile)
	response := GetAthleteProfileResponse{
		AthleteID:     athleteID,
		Name:          strings.TrimSpace(profile.Name),
		FirstName:     strings.TrimSpace(profile.FirstName),
		LastName:      strings.TrimSpace(profile.LastName),
		Timezone:      profileTimezone(profile.Timezone, timezoneFallback),
		Locale:        strings.TrimSpace(profile.Locale),
		Units:         units,
		SportSettings: make([]GetAthleteProfileSport, 0, len(profile.SportSettings)),
		Meta: GetAthleteProfileMeta{
			ServerVersion:      normalizeVersion(version),
			AthleteIDFormat:    "i-prefixed intervals.icu athlete ID",
			TimezoneConvention: "IANA timezone from athlete profile when available; config timezone fallback otherwise",
			PaceConvention:     "paces are seconds per athlete pace distance unit; metric athletes receive threshold_pace_seconds_per_km/pace_zones_seconds_per_km, imperial athletes receive threshold_pace_seconds_per_mile/pace_zones_seconds_per_mile, and pace_units_source preserves the upstream enum such as MINS_KM or MINS_MILE",
			IncludeFull:        includeFull,
		},
	}
	if includeFull && profile.MeasurementPreference != "" && profile.MeasurementPreference != units.MeasurementPreference {
		response.MeasurementPreferenceSource = profile.MeasurementPreference
	}
	for _, setting := range profile.SportSettings {
		response.SportSettings = append(response.SportSettings, profileSport(setting, includeFull))
	}
	return response
}

func profileUnits(profile intervals.AthleteWithSportSettings) GetAthleteProfileUnits {
	measurement := normalizedMeasurementPreference(profile.MeasurementPreference, profile.WeightPrefLB)
	weight := "kg"
	if profile.WeightPrefLB {
		weight = "lb"
	}
	temperature := "celsius"
	if profile.Fahrenheit {
		temperature = "fahrenheit"
	}
	return GetAthleteProfileUnits{
		MeasurementPreference: measurement,
		Weight:                weight,
		Temperature:           temperature,
	}
}

func profileSport(setting intervals.SportSettings, includeFull bool) GetAthleteProfileSport {
	sport := GetAthleteProfileSport{
		Types:           setting.Types,
		FTPWatts:        setting.FTP,
		IndoorFTPWatts:  setting.IndoorFTP,
		WPrimeJoules:    setting.WPrime,
		PMaxWatts:       setting.PMax,
		LTHRBPM:         setting.LTHR,
		MaxHRBPM:        setting.MaxHR,
		PowerZonesWatts: setting.PowerZones,
		PowerZoneNames:  setting.PowerZoneNames,
		HRZonesBPM:      setting.HRZones,
		HRZoneNames:     setting.HRZoneNames,
		PaceUnitsSource: strings.TrimSpace(setting.PaceUnits),
		PaceZoneNames:   setting.PaceZoneNames,
	}
	pace := setting.ThresholdPace
	if pace > 0 {
		if isMilePace(setting.PaceUnits) {
			sport.ThresholdPaceSecondsPerMile = &pace
		} else {
			sport.ThresholdPaceSecondsPerKM = &pace
		}
	}
	if len(setting.PaceZones) > 0 {
		if isMilePace(setting.PaceUnits) {
			sport.PaceZonesSecondsPerMile = setting.PaceZones
		} else {
			sport.PaceZonesSecondsPerKM = setting.PaceZones
		}
	}
	if setting.PaceUnits != "" || pace > 0 || len(setting.PaceZones) > 0 {
		if isMilePace(setting.PaceUnits) {
			sport.PaceDistanceUnit = "mile"
		} else {
			sport.PaceDistanceUnit = "km"
		}
	}
	if includeFull {
		sport.SportSettingID = setting.ID
		sport.SportSettingAthleteID = normalizeProfileAthleteID(setting.AthleteID)
	}
	return sport
}

func getAthleteProfileInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"include_full": map[string]any{
				"type":        "boolean",
				"default":     false,
				"description": "When true, include additional typed, non-secret profile and sport-setting identifiers. Defaults to false; raw upstream payloads and credentials are never returned.",
			},
		},
	}
}

func getAthleteProfileOutputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": true,
		"description":          "Terse athlete profile with normalized athlete_id, units, timezone, sport thresholds/zones, and _meta.server_version.",
	}
}

func profileTimezone(profileTimezone string, fallback string) string {
	if timezone := strings.TrimSpace(profileTimezone); timezone != "" {
		return timezone
	}
	return strings.TrimSpace(fallback)
}

func normalizedMeasurementPreference(value string, weightPrefLB bool) string {
	trimmed := strings.TrimSpace(value)
	upper := strings.ToUpper(trimmed)
	if strings.Contains(upper, "IMPERIAL") {
		return "imperial"
	}
	if strings.Contains(upper, "METRIC") {
		return "metric"
	}
	if trimmed == "" && weightPrefLB {
		return "imperial"
	}
	return "metric"
}

func normalizeTimezoneFallback(values ...string) string {
	if fallback := firstNonEmpty(values...); fallback != "" {
		return fallback
	}
	return config.DefaultTimezone
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeProfileAthleteID(value string) string {
	normalized, err := config.NormalizeAthleteID(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return normalized
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

func isMilePace(paceUnits string) bool {
	return strings.Contains(strings.ToUpper(paceUnits), "MILE")
}
