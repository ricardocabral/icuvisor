// Package streams canonicalizes intervals.icu activity stream keys at response boundaries.
package streams

import (
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// KnownStreamKeys maps observed upstream stream key spellings to canonical snake_case keys.
var KnownStreamKeys = map[string]string{
	"Altitude":               "altitude",
	"Cadence":                "cadence",
	"CoreTemperature":        "core_temperature",
	"Distance":               "distance",
	"FormPower":              "form_power",
	"GradeSmooth":            "grade_smooth",
	"GroundContactBalance":   "ground_contact_balance",
	"GroundContactTime":      "ground_contact_time",
	"HeartRate":              "heart_rate",
	"HR":                     "heart_rate",
	"LatLng":                 "latlng",
	"LeftRightBalance":       "left_right_balance",
	"LegSpringStiffness":     "leg_spring_stiffness",
	"Moving":                 "moving",
	"Power":                  "watts",
	"StanceTime":             "stance_time",
	"StrideLength":           "stride_length",
	"Temperature":            "temperature",
	"Time":                   "time",
	"VelocitySmooth":         "velocity_smooth",
	"VerticalOscillation":    "vertical_oscillation",
	"WPrimeBalance":          "w_prime_balance",
	"Watts":                  "watts",
	"altitude":               "altitude",
	"cadence":                "cadence",
	"coreTemperature":        "core_temperature",
	"core_temperature":       "core_temperature",
	"distance":               "distance",
	"formPower":              "form_power",
	"form_power":             "form_power",
	"gradeSmooth":            "grade_smooth",
	"grade_smooth":           "grade_smooth",
	"groundContactBalance":   "ground_contact_balance",
	"groundContactTime":      "ground_contact_time",
	"ground_contact_balance": "ground_contact_balance",
	"ground_contact_time":    "ground_contact_time",
	"heartRate":              "heart_rate",
	"heart_rate":             "heart_rate",
	"heartrate":              "heart_rate",
	"hr":                     "heart_rate",
	"latLng":                 "latlng",
	"lat_lng":                "latlng",
	"latlng":                 "latlng",
	"leftRightBalance":       "left_right_balance",
	"left_right_balance":     "left_right_balance",
	"legSpringStiffness":     "leg_spring_stiffness",
	"leg_spring_stiffness":   "leg_spring_stiffness",
	"moving":                 "moving",
	"power":                  "watts",
	"stanceTime":             "stance_time",
	"stance_time":            "stance_time",
	"strideLength":           "stride_length",
	"stride_length":          "stride_length",
	"temp":                   "temperature",
	"temperature":            "temperature",
	"time":                   "time",
	"velocitySmooth":         "velocity_smooth",
	"velocity_smooth":        "velocity_smooth",
	"verticalOscillation":    "vertical_oscillation",
	"vertical_oscillation":   "vertical_oscillation",
	"wPrimeBalance":          "w_prime_balance",
	"w_prime_balance":        "w_prime_balance",
	"watts":                  "watts",
	"wbal":                   "w_prime_balance",
}

// CanonicalKey returns the canonical snake_case key and whether it was in the known map.
func CanonicalKey(key string) (string, bool) {
	trimmed := strings.TrimSpace(key)
	if canonical, ok := KnownStreamKeys[trimmed]; ok {
		return canonical, true
	}
	return ToSnakeCase(trimmed), false
}

// CanonicalizeRow canonicalizes stream keys in a single response row without mutating input.
func CanonicalizeRow(row map[string]any) map[string]any {
	out := make(map[string]any, len(row))
	meta := map[string]any{}
	if existing, ok := row["_meta"].(map[string]any); ok {
		for key, value := range existing {
			meta[key] = value
		}
	}

	keys := make([]string, 0, len(row))
	for key := range row {
		if key != "_meta" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	sourceKeys := map[string][]string{}
	var unknownKeys []string
	collisions := map[string][]string{}
	for _, key := range keys {
		canonical, known := CanonicalKey(key)
		if canonical == "" {
			canonical = key
		}
		if !known {
			unknownKeys = append(unknownKeys, key)
		}
		value := row[key]
		if existing, ok := out[canonical]; ok {
			sourceKeys[canonical] = append(sourceKeys[canonical], key)
			if reflect.DeepEqual(existing, value) {
				continue
			}
			_, alreadyCollided := collisions[canonical]
			out[canonical] = appendCollisionValue(existing, value, alreadyCollided)
			collisions[canonical] = append([]string(nil), sourceKeys[canonical]...)
			continue
		}
		out[canonical] = value
		sourceKeys[canonical] = []string{key}
	}
	if len(unknownKeys) > 0 {
		meta["unknown_stream_keys"] = sortedUnique(unknownKeys)
	}
	if len(collisions) > 0 {
		meta["stream_key_collisions"] = collisions
	}
	if len(meta) > 0 {
		out["_meta"] = meta
	}
	return out
}

// CanonicalizeRows canonicalizes every row in a response-boundary stream collection.
func CanonicalizeRows(rows []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, CanonicalizeRow(row))
	}
	return out
}

// ToSnakeCase converts unknown upstream keys to best-effort snake_case.
func ToSnakeCase(key string) string {
	key = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), " ", "_"))
	if key == "" {
		return ""
	}
	var builder strings.Builder
	var previous rune
	runes := []rune(key)
	for i, current := range runes {
		next := rune(0)
		if i+1 < len(runes) {
			next = runes[i+1]
		}
		if current == '_' {
			writeUnderscore(&builder)
			previous = current
			continue
		}
		if unicode.IsUpper(current) && i > 0 && previous != '_' && (unicode.IsLower(previous) || unicode.IsDigit(previous) || unicode.IsLower(next)) {
			writeUnderscore(&builder)
		}
		builder.WriteRune(unicode.ToLower(current))
		previous = current
	}
	return strings.Trim(builder.String(), "_")
}

func appendCollisionValue(existing any, value any, alreadyCollided bool) []any {
	if alreadyCollided {
		if existingValues, ok := existing.([]any); ok {
			out := append([]any(nil), existingValues...)
			return append(out, value)
		}
	}
	return []any{existing, value}
}

func writeUnderscore(builder *strings.Builder) {
	current := builder.String()
	if current == "" || strings.HasSuffix(current, "_") {
		return
	}
	builder.WriteRune('_')
}

func sortedUnique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
