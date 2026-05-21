package tools

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type stravaNativeProvider struct {
	key     string
	display string
}

var stravaNativeProviders = []stravaNativeProvider{
	{key: "garmin", display: "Garmin"},
	{key: "wahoo", display: "Wahoo"},
	{key: "coros", display: "Coros"},
	{key: "suunto", display: "Suunto"},
	{key: "polar", display: "Polar"},
}

func stravaBlockedWorkaround(raw map[string]any) string {
	provider := inferStravaNativeProvider(raw)
	if provider == "" {
		return stravaUnknownProviderWorkaround
	}
	return fmt.Sprintf("Open the intervals.icu Connections page, choose %s, and click Download old data so historical activities are re-imported directly from %s instead of through Strava's restricted API.", provider, provider)
}

func inferStravaNativeProvider(raw map[string]any) string {
	externalID := strings.ToLower(activityRawText(raw, "external_id"))
	deviceName := strings.ToLower(activityRawText(raw, "device_name"))
	for _, provider := range stravaNativeProviders {
		if hasProviderPrefix(externalID, provider.key) || containsProviderWord(deviceName, provider.key) {
			return provider.display
		}
	}
	return ""
}

func hasProviderPrefix(value string, provider string) bool {
	if value == provider {
		return true
	}
	if !strings.HasPrefix(value, provider) {
		return false
	}
	if len(value) == len(provider) {
		return true
	}
	next := rune(value[len(provider)])
	return next == '-' || next == '_' || next == '.' || next == ':'
}

func containsProviderWord(value string, provider string) bool {
	idx := strings.Index(value, provider)
	for idx >= 0 {
		beforeOK := idx == 0 || !unicode.IsLetter(rune(value[idx-1]))
		afterIdx := idx + len(provider)
		afterOK := afterIdx == len(value) || !unicode.IsLetter(rune(value[afterIdx]))
		if beforeOK && afterOK {
			return true
		}
		next := strings.Index(value[idx+len(provider):], provider)
		if next < 0 {
			break
		}
		idx += len(provider) + next
	}
	return false
}

func activityRawText(raw map[string]any, key string) string {
	if raw == nil {
		return ""
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func isStravaBlocked(activity intervals.Activity) bool {
	source := strings.ToLower(stringValue(activity.Source))
	note := strings.ToLower(stringValue(activity.Note))
	if strings.Contains(source, "strava") || strings.Contains(note, "strava") {
		return true
	}
	if source != "" {
		return false
	}
	meaningful := 0
	for _, key := range []string{"name", "type", "distance", "icu_distance", "moving_time", "elapsed_time"} {
		if value, ok := activity.Raw[key]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
			meaningful++
		}
	}
	if meaningful > 0 {
		return false
	}
	if len(activity.Raw) == 0 {
		return true
	}
	_, hasHiddenNote := activity.Raw["_note"]
	if hasHiddenNote {
		return true
	}
	stubKeys := map[string]bool{"id": true, "icu_athlete_id": true, "athlete_id": true, "start_date_local": true, "start_date": true, "external_id": true}
	nullableStubKeys := map[string]bool{"name": true, "type": true, "sub_type": true, "distance": true, "icu_distance": true, "moving_time": true, "elapsed_time": true, "source": true, "_note": true}
	for key, value := range activity.Raw {
		if stubKeys[key] {
			continue
		}
		if value == nil && nullableStubKeys[key] {
			continue
		}
		return false
	}
	return true
}
