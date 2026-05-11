package response

import (
	"strings"

	"github.com/ricardocabral/icuvisor/internal/units"
)

const (
	kilometersToMiles  = 0.621371192237334
	metersPerKilometer = 1000
	metersPerMile      = 1609.344
	metersPerYard      = 0.9144
	secondsPerHour     = 3600
)

// UnitSystem is the athlete's response-boundary distance unit preference.
type UnitSystem string

// PreferredUnitValue contains response-boundary unit conversion output without losing the upstream value.
type PreferredUnitValue struct {
	Value             float64
	Unit              units.Unit
	UnitLabel         string
	FieldSuffix       string
	OriginalValue     float64
	OriginalUnit      units.Unit
	OriginalUnitLabel string
	Converted         bool
	UnknownUnit       string
}

const (
	UnitSystemMetric   UnitSystem = "metric"
	UnitSystemImperial UnitSystem = "imperial"
)

// UnitSystemFromPreferredUnits normalizes known intervals.icu preferred_units values.
func UnitSystemFromPreferredUnits(preferredUnits string) (UnitSystem, bool) {
	return parseUnitSystem(preferredUnits)
}

// UnitSystemFromProfile derives the active unit system from profile preferences.
func UnitSystemFromProfile(preferredUnits string, measurementPreference string, weightPrefLB bool) (UnitSystem, bool) {
	if unitSystem, ok := parseUnitSystem(preferredUnits); ok {
		return unitSystem, true
	}
	if unitSystem, ok := parseUnitSystem(measurementPreference); ok {
		return unitSystem, true
	}
	if weightPrefLB {
		return UnitSystemImperial, true
	}
	return "", false
}

func parseUnitSystem(value string) (UnitSystem, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "unknown":
		return "", false
	case "metric", "km", "kms", "kilometer", "kilometers", "kilometres":
		return UnitSystemMetric, true
	case "imperial", "mile", "miles", "mi":
		return UnitSystemImperial, true
	}
	if strings.Contains(normalized, "imperial") || strings.Contains(normalized, "mile") {
		return UnitSystemImperial, true
	}
	if strings.Contains(normalized, "metric") || strings.Contains(normalized, "kilomet") {
		return UnitSystemMetric, true
	}
	return "", false
}

// DistanceFieldName returns the unit-disambiguated JSON field name for a distance base name.
func (u UnitSystem) DistanceFieldName(base string) string {
	base = strings.TrimSuffix(strings.TrimSpace(base), "_km")
	base = strings.TrimSuffix(base, "_mi")
	if base == "" {
		base = "distance"
	}
	if u == UnitSystemImperial {
		return base + "_mi"
	}
	return base + "_km"
}

// ConvertDistanceKM converts a kilometer value into the active unit system.
func (u UnitSystem) ConvertDistanceKM(kilometers float64) float64 {
	if u == UnitSystemImperial {
		return kilometers * kilometersToMiles
	}
	return kilometers
}

// ToPreferred converts value into the athlete-preferred unit family where a safe generic rule exists.
func ToPreferred(value float64, fromUnit units.Unit, sys UnitSystem) PreferredUnitValue {
	return ToPreferredWithRaw(value, fromUnit, "", sys)
}

// ToPreferredWithRaw converts value like ToPreferred while preserving an unknown raw upstream unit label.
func ToPreferredWithRaw(value float64, fromUnit units.Unit, rawUnit string, sys UnitSystem) PreferredUnitValue {
	originalLabel := preferredOriginalLabel(fromUnit, rawUnit)
	result := PreferredUnitValue{
		Value:             value,
		Unit:              fromUnit,
		UnitLabel:         originalLabel,
		FieldSuffix:       unitFieldSuffix(fromUnit),
		OriginalValue:     value,
		OriginalUnit:      fromUnit,
		OriginalUnitLabel: originalLabel,
	}
	if fromUnit == units.UnitUnknown {
		result.UnknownUnit = originalLabel
		return result
	}
	if sys == "" {
		sys = UnitSystemMetric
	}
	if converted, ok := convertDistance(value, fromUnit, sys); ok {
		return converted
	}
	if converted, ok := convertSpeed(value, fromUnit, sys); ok {
		return converted
	}
	if converted, ok := convertRunPace(value, fromUnit, sys); ok {
		return converted
	}
	return result
}

// Metadata returns the response _meta.units payload for the active unit system.
func (u UnitSystem) Metadata() map[string]string {
	if u == UnitSystemImperial {
		return map[string]string{"system": string(UnitSystemImperial), "distance": "mi", "pace": "min/mi", "speed": "mph"}
	}
	return map[string]string{"system": string(UnitSystemMetric), "distance": "km", "pace": "min/km", "speed": "km/h"}
}

func convertDistance(value float64, fromUnit units.Unit, sys UnitSystem) (PreferredUnitValue, bool) {
	meters, ok := distanceMeters(value, fromUnit)
	if !ok {
		return PreferredUnitValue{}, false
	}
	if sys == UnitSystemImperial {
		return preferredConverted(value, fromUnit, meters/metersPerMile, units.UnitMI, "mi", "mi"), true
	}
	return preferredConverted(value, fromUnit, meters/metersPerKilometer, units.UnitKM, "km", "km"), true
}

func distanceMeters(value float64, fromUnit units.Unit) (float64, bool) {
	switch fromUnit {
	case units.UnitM:
		return value, true
	case units.UnitKM:
		return value * metersPerKilometer, true
	case units.UnitMI:
		return value * metersPerMile, true
	case units.UnitYD:
		return value * metersPerYard, true
	default:
		return 0, false
	}
}

func convertSpeed(value float64, fromUnit units.Unit, sys UnitSystem) (PreferredUnitValue, bool) {
	metersPerSecond, ok := speedMetersPerSecond(value, fromUnit)
	if !ok {
		return PreferredUnitValue{}, false
	}
	if sys == UnitSystemImperial {
		return preferredConverted(value, fromUnit, metersPerSecond*secondsPerHour/metersPerMile, units.UnitMPH, "mph", "mph"), true
	}
	return preferredConverted(value, fromUnit, metersPerSecond*secondsPerHour/metersPerKilometer, units.UnitKMH, "km/h", "kmh"), true
}

func speedMetersPerSecond(value float64, fromUnit units.Unit) (float64, bool) {
	switch fromUnit {
	case units.UnitKMH:
		return value * metersPerKilometer / secondsPerHour, true
	case units.UnitMPH:
		return value * metersPerMile / secondsPerHour, true
	case units.UnitMS:
		return value, true
	default:
		return 0, false
	}
}

func convertRunPace(value float64, fromUnit units.Unit, sys UnitSystem) (PreferredUnitValue, bool) {
	switch fromUnit {
	case units.UnitMinsKM:
		if sys == UnitSystemImperial {
			return preferredConverted(value, fromUnit, value*metersPerMile/metersPerKilometer, units.UnitMinsMile, "min/mi", "minutes_per_mile"), true
		}
		return preferredConverted(value, fromUnit, value, units.UnitMinsKM, "min/km", "minutes_per_km"), true
	case units.UnitMinsMile:
		if sys == UnitSystemImperial {
			return preferredConverted(value, fromUnit, value, units.UnitMinsMile, "min/mi", "minutes_per_mile"), true
		}
		return preferredConverted(value, fromUnit, value*metersPerKilometer/metersPerMile, units.UnitMinsKM, "min/km", "minutes_per_km"), true
	default:
		return PreferredUnitValue{}, false
	}
}

func preferredConverted(originalValue float64, originalUnit units.Unit, value float64, unit units.Unit, unitLabel string, fieldSuffix string) PreferredUnitValue {
	return PreferredUnitValue{
		Value:             value,
		Unit:              unit,
		UnitLabel:         unitLabel,
		FieldSuffix:       fieldSuffix,
		OriginalValue:     originalValue,
		OriginalUnit:      originalUnit,
		OriginalUnitLabel: string(originalUnit),
		Converted:         unit != originalUnit || value != originalValue,
	}
}

func preferredOriginalLabel(fromUnit units.Unit, rawUnit string) string {
	if raw := strings.TrimSpace(rawUnit); raw != "" || fromUnit == units.UnitUnknown {
		return raw
	}
	return string(fromUnit)
}

func unitFieldSuffix(unit units.Unit) string {
	switch unit {
	case units.UnitM:
		return "m"
	case units.UnitKM:
		return "km"
	case units.UnitMI:
		return "mi"
	case units.UnitYD:
		return "yd"
	case units.UnitKMH:
		return "kmh"
	case units.UnitMPH:
		return "mph"
	case units.UnitMS:
		return "mps"
	case units.UnitMinsKM:
		return "minutes_per_km"
	case units.UnitMinsMile:
		return "minutes_per_mile"
	case units.UnitSecs100M:
		return "seconds_per_100m"
	case units.UnitSecs500M:
		return "seconds_per_500m"
	default:
		return strings.ToLower(string(unit))
	}
}
