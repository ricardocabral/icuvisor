package athleteprofile

import (
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
	"github.com/ricardocabral/icuvisor/internal/units"
)

const queryType = "get_athlete_profile"

// Response is the structured athlete-profile response shared by tools and resources.
type Response struct {
	AthleteID                   string  `json:"athlete_id"`
	Name                        string  `json:"name,omitempty"`
	FirstName                   string  `json:"first_name,omitempty"`
	LastName                    string  `json:"last_name,omitempty"`
	Timezone                    string  `json:"timezone,omitempty"`
	Locale                      string  `json:"locale,omitempty"`
	Units                       Units   `json:"units"`
	SportSettings               []Sport `json:"sport_settings,omitempty"`
	Meta                        Meta    `json:"_meta"`
	MeasurementPreferenceSource string  `json:"measurement_preference_source,omitempty"`
}

// Units describes athlete unit preferences.
type Units struct {
	MeasurementPreference string `json:"measurement_preference,omitempty"`
	Weight                string `json:"weight,omitempty"`
	Temperature           string `json:"temperature,omitempty"`
}

// Sport contains thresholds and zones for one sport setting.
type Sport struct {
	Types                       []string       `json:"types,omitempty"`
	FTPWatts                    int            `json:"ftp_watts,omitempty"`
	IndoorFTPWatts              int            `json:"indoor_ftp_watts,omitempty"`
	WPrimeJoules                int            `json:"w_prime_joules,omitempty"`
	PMaxWatts                   int            `json:"p_max_watts,omitempty"`
	LTHRBPM                     int            `json:"lthr_bpm,omitempty"`
	MaxHRBPM                    int            `json:"max_hr_bpm,omitempty"`
	PowerZonesWatts             []int          `json:"power_zones_watts,omitempty"`
	PowerZoneNames              []string       `json:"power_zone_names,omitempty"`
	HRZonesBPM                  []int          `json:"hr_zones_bpm,omitempty"`
	HRZoneNames                 []string       `json:"hr_zone_names,omitempty"`
	ThresholdPaceSecondsPerKM   *float64       `json:"threshold_pace_seconds_per_km,omitempty"`
	PaceZonesSecondsPerKM       []float64      `json:"pace_zones_seconds_per_km,omitempty"`
	ThresholdPaceSecondsPerMile *float64       `json:"threshold_pace_seconds_per_mile,omitempty"`
	PaceZonesSecondsPerMile     []float64      `json:"pace_zones_seconds_per_mile,omitempty"`
	ThresholdPaceSecondsPer100M *float64       `json:"threshold_pace_seconds_per_100m,omitempty"`
	PaceZonesSecondsPer100M     []float64      `json:"pace_zones_seconds_per_100m,omitempty"`
	ThresholdPaceSecondsPer100Y *float64       `json:"threshold_pace_seconds_per_100y,omitempty"`
	PaceZonesSecondsPer100Y     []float64      `json:"pace_zones_seconds_per_100y,omitempty"`
	ThresholdPaceSecondsPer500M *float64       `json:"threshold_pace_seconds_per_500m,omitempty"`
	PaceZonesSecondsPer500M     []float64      `json:"pace_zones_seconds_per_500m,omitempty"`
	ThresholdPaceValue          *float64       `json:"threshold_pace_value,omitempty"`
	PaceZonesValues             []float64      `json:"pace_zones_values,omitempty"`
	PaceUnitsSource             string         `json:"pace_units_source,omitempty"`
	PaceDistanceUnit            string         `json:"pace_distance_unit,omitempty"`
	PaceZoneNames               []string       `json:"pace_zone_names,omitempty"`
	SportSettingID              int            `json:"sport_setting_id,omitempty"`
	SportSettingAthleteID       string         `json:"sport_setting_athlete_id,omitempty"`
	Meta                        map[string]any `json:"_meta,omitempty"`
}

// ReadinessWarning flags missing sport settings that can affect analysis or planning.
type ReadinessWarning struct {
	Code       string   `json:"code"`
	SportTypes []string `json:"sport_types"`
	Field      string   `json:"field"`
	Message    string   `json:"message"`
	Action     string   `json:"action"`
}

// Meta contains response-shaping metadata.
type Meta struct {
	ServerVersion      string             `json:"server_version"`
	AthleteIDFormat    string             `json:"athlete_id_format"`
	TimezoneConvention string             `json:"timezone_convention"`
	PaceConvention     string             `json:"pace_convention"`
	IncludeFull        bool               `json:"include_full"`
	Warnings           []ReadinessWarning `json:"warnings,omitempty"`
}

// Shape returns the shaped athlete profile used by get_athlete_profile and icuvisor://athlete-profile.
func Shape(profile intervals.AthleteWithSportSettings, version string, timezoneFallback string, includeFull bool, debugMetadata bool, shaping ...response.Options) (any, error) {
	profileResponse := NewResponse(profile, version, timezoneFallback, includeFull)
	opts := response.Options{
		IncludeFull:   includeFull,
		ServerVersion: NormalizeVersion(version),
		DebugMetadata: debugMetadata,
		QueryType:     queryType,
		UnitSystem:    profileUnitSystem(profile),
	}
	if len(shaping) > 0 {
		opts.DeleteMode = shaping[0].DeleteMode
		opts.Toolset = shaping[0].Toolset
		opts.CatalogHash = shaping[0].CatalogHash
	}
	return response.Shape(profileResponse, opts)
}

// NewResponse builds the typed athlete profile response before response-boundary shaping.
func NewResponse(profile intervals.AthleteWithSportSettings, version string, timezoneFallback string, includeFull bool) Response {
	athleteID := config.NormalizeAthleteIDForDisplay(profile.ID)
	units := profileUnits(profile)
	response := Response{
		AthleteID:     athleteID,
		Name:          strings.TrimSpace(profile.Name),
		FirstName:     strings.TrimSpace(profile.FirstName),
		LastName:      strings.TrimSpace(profile.LastName),
		Timezone:      profileTimezone(profile.Timezone, timezoneFallback),
		Locale:        strings.TrimSpace(profile.Locale),
		Units:         units,
		SportSettings: make([]Sport, 0, len(profile.SportSettings)),
		Meta: Meta{
			ServerVersion:      NormalizeVersion(version),
			AthleteIDFormat:    "i-prefixed intervals.icu athlete ID",
			TimezoneConvention: "IANA timezone from athlete profile when available; config timezone fallback otherwise",
			PaceConvention:     "paces are seconds per athlete pace distance unit; run pace uses threshold_pace_seconds_per_km or threshold_pace_seconds_per_mile, swim pace uses threshold_pace_seconds_per_100m or threshold_pace_seconds_per_100y when upstream reports SECS_100M/SECS_100Y, row pace uses threshold_pace_seconds_per_500m, and pace_units_source preserves the upstream enum",
			IncludeFull:        includeFull,
		},
	}
	if includeFull && profile.MeasurementPreference != "" && profile.MeasurementPreference != units.MeasurementPreference {
		response.MeasurementPreferenceSource = profile.MeasurementPreference
	}
	unitSystem := profileUnitSystem(profile)
	for _, setting := range profile.SportSettings {
		response.SportSettings = append(response.SportSettings, profileSport(setting, includeFull, unitSystem))
		response.Meta.Warnings = append(response.Meta.Warnings, sportReadinessWarnings(setting)...)
	}
	return response
}

// NormalizeTimezoneFallback returns a non-empty IANA timezone fallback.
func NormalizeTimezoneFallback(values ...string) string {
	if fallback := firstNonEmpty(values...); fallback != "" {
		return fallback
	}
	return config.DefaultTimezone
}

// NormalizeVersion returns a non-empty server version.
func NormalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "dev"
	}
	return version
}

func profileUnits(profile intervals.AthleteWithSportSettings) Units {
	measurement := string(profileUnitSystem(profile))
	weight := "kg"
	if profile.WeightPrefLB {
		weight = "lb"
	}
	temperature := "celsius"
	if profile.Fahrenheit {
		temperature = "fahrenheit"
	}
	return Units{
		MeasurementPreference: measurement,
		Weight:                weight,
		Temperature:           temperature,
	}
}

func profileUnitSystem(profile intervals.AthleteWithSportSettings) response.UnitSystem {
	if unitSystem, ok := response.UnitSystemFromProfile(profile.PreferredUnits, profile.MeasurementPreference, profile.WeightPrefLB); ok {
		return unitSystem
	}
	return response.UnitSystemMetric
}

func profileSport(setting intervals.SportSettings, includeFull bool, unitSystem response.UnitSystem) Sport {
	sport := Sport{
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
	applyProfilePace(&sport, setting, unitSystem)
	if includeFull {
		sport.SportSettingID = setting.ID
		sport.SportSettingAthleteID = config.NormalizeAthleteIDForDisplay(setting.AthleteID)
	}
	return sport
}

func profileTimezone(profileTimezone string, fallback string) string {
	if timezone := strings.TrimSpace(profileTimezone); timezone != "" {
		return timezone
	}
	return strings.TrimSpace(fallback)
}

func sportReadinessWarnings(setting intervals.SportSettings) []ReadinessWarning {
	sportTypes := readinessSportTypes(setting)
	var warnings []ReadinessWarning
	if isRideSport(sportTypes) {
		if setting.FTP <= 0 {
			warnings = append(warnings, readinessWarning("missing_power_threshold", sportTypes, "ftp_watts", "power threshold is missing for this sport", "Use update_sport_settings with ftp for this sport before power-based planning."))
		}
		if len(setting.PowerZones) == 0 {
			warnings = append(warnings, readinessWarning("missing_power_zones", sportTypes, "power_zones_watts", "power zones are missing for this sport", "Use update_sport_settings with zones kind=power for this sport before zone-based planning."))
		}
	}
	if usesHeartRateReadiness(sportTypes) {
		if setting.LTHR <= 0 && setting.FTHR <= 0 {
			warnings = append(warnings, readinessWarning("missing_hr_threshold", sportTypes, "lthr_bpm", "heart-rate threshold is missing for this sport", "Use update_sport_settings with threshold_hr for this sport before heart-rate-based planning."))
		}
		if len(setting.HRZones) == 0 {
			warnings = append(warnings, readinessWarning("missing_hr_zones", sportTypes, "hr_zones_bpm", "heart-rate zones are missing for this sport", "Use update_sport_settings with zones kind=hr for this sport before zone-based planning."))
		}
	}
	if usesPaceReadiness(sportTypes) {
		if setting.ThresholdPace <= 0 && setting.PaceThreshold <= 0 {
			warnings = append(warnings, readinessWarning("missing_pace_threshold", sportTypes, "threshold_pace", "pace threshold is missing for this sport", "Use update_sport_settings with threshold_pace for this sport before pace-based planning."))
		}
		if len(setting.PaceZones) == 0 {
			warnings = append(warnings, readinessWarning("missing_pace_zones", sportTypes, "pace_zones", "pace zones are missing for this sport", "Use update_sport_settings with zones kind=pace for this sport before zone-based planning."))
		}
	}
	return warnings
}

func readinessWarning(code string, sportTypes []string, field string, message string, action string) ReadinessWarning {
	return ReadinessWarning{Code: code, SportTypes: sportTypes, Field: field, Message: message, Action: action}
}

func readinessSportTypes(setting intervals.SportSettings) []string {
	values := setting.Types
	if len(values) == 0 {
		values = []string{setting.Type}
	}
	seen := map[string]bool{}
	var sportTypes []string
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || seen[strings.ToLower(trimmed)] {
			continue
		}
		seen[strings.ToLower(trimmed)] = true
		sportTypes = append(sportTypes, trimmed)
	}
	if len(sportTypes) == 0 {
		return []string{"unknown"}
	}
	return sportTypes
}

func isRideSport(sportTypes []string) bool {
	for _, sportType := range sportTypes {
		switch normalizeSportType(sportType) {
		case "ride", "virtualride", "bike", "bikeride", "cycle", "cycling", "indoorcycling", "mountainbike", "mtb", "gravelride", "ebikeride":
			return true
		}
	}
	return false
}

func usesHeartRateReadiness(sportTypes []string) bool {
	for _, sportType := range sportTypes {
		switch normalizeSportType(sportType) {
		case "ride", "virtualride", "bike", "bikeride", "cycle", "cycling", "indoorcycling", "mountainbike", "mtb", "gravelride", "ebikeride",
			"run", "virtualrun", "trailrun", "treadmill", "walk", "hike",
			"swim", "openwaterswim", "poolswim",
			"row", "rowing", "kayak", "canoe":
			return true
		}
	}
	return false
}

func usesPaceReadiness(sportTypes []string) bool {
	for _, sportType := range sportTypes {
		switch normalizeSportType(sportType) {
		case "run", "virtualrun", "trailrun", "treadmill", "walk", "hike", "swim", "openwaterswim", "poolswim", "row", "rowing", "kayak", "canoe":
			return true
		}
	}
	return false
}

func normalizeSportType(sportType string) string {
	return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.TrimSpace(sportType)))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func applyProfilePace(sport *Sport, setting intervals.SportSettings, unitSystem response.UnitSystem) {
	pace := setting.ThresholdPace
	if strings.TrimSpace(setting.PaceUnits) == "" && pace <= 0 && len(setting.PaceZones) == 0 {
		return
	}
	parsedUnit, rawUnit := units.ParseUnit(setting.PaceUnits)
	if parsedUnit == units.UnitUnknown {
		sport.Meta = map[string]any{"unknown_unit": rawUnit}
	}
	if pace > 0 {
		converted := response.ToPreferredWithRaw(pace, parsedUnit, rawUnit, unitSystem)
		assignProfileThresholdPace(sport, converted)
	}
	if len(setting.PaceZones) > 0 {
		convertedZones := make([]float64, 0, len(setting.PaceZones))
		var converted response.PreferredUnitValue
		for _, zone := range setting.PaceZones {
			converted = response.ToPreferredWithRaw(zone, parsedUnit, rawUnit, unitSystem)
			convertedZones = append(convertedZones, converted.Value)
		}
		assignProfilePaceZones(sport, converted, convertedZones)
	}
	if setting.PaceUnits != "" || pace > 0 || len(setting.PaceZones) > 0 {
		converted := response.ToPreferredWithRaw(pace, parsedUnit, rawUnit, unitSystem)
		sport.PaceDistanceUnit = profilePaceDistanceUnit(converted)
	}
}

func assignProfileThresholdPace(sport *Sport, converted response.PreferredUnitValue) {
	value := converted.Value
	switch converted.Unit {
	case units.UnitMinsKM:
		sport.ThresholdPaceSecondsPerKM = &value
	case units.UnitMinsMile:
		sport.ThresholdPaceSecondsPerMile = &value
	case units.UnitSecs100M:
		sport.ThresholdPaceSecondsPer100M = &value
	case units.UnitSecs100Y:
		sport.ThresholdPaceSecondsPer100Y = &value
	case units.UnitSecs500M:
		sport.ThresholdPaceSecondsPer500M = &value
	default:
		sport.ThresholdPaceValue = &value
	}
}

func assignProfilePaceZones(sport *Sport, converted response.PreferredUnitValue, values []float64) {
	switch converted.Unit {
	case units.UnitMinsKM:
		sport.PaceZonesSecondsPerKM = values
	case units.UnitMinsMile:
		sport.PaceZonesSecondsPerMile = values
	case units.UnitSecs100M:
		sport.PaceZonesSecondsPer100M = values
	case units.UnitSecs100Y:
		sport.PaceZonesSecondsPer100Y = values
	case units.UnitSecs500M:
		sport.PaceZonesSecondsPer500M = values
	default:
		sport.PaceZonesValues = values
	}
}

func profilePaceDistanceUnit(converted response.PreferredUnitValue) string {
	switch converted.Unit {
	case units.UnitMinsKM:
		return "km"
	case units.UnitMinsMile:
		return "mile"
	case units.UnitSecs100M:
		return "100m"
	case units.UnitSecs100Y:
		return "100y"
	case units.UnitSecs500M:
		return "500m"
	default:
		return converted.UnitLabel
	}
}
