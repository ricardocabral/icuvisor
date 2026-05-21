//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var roots = []string{
	"internal/intervals/testdata",
	"internal/tools/testdata",
}

var precomputedKeys = map[string][]string{
	"power":      {"icu_zone_times", "power_zone_distribution_seconds", "power_zone_times"},
	"heart_rate": {"hr_zone_times", "heartrate_zone_times", "heart_rate_zone_times", "hr_time_in_zones"},
	"pace":       {"gap_zone_times", "pace_zone_times", "pace_zone_time_seconds"},
}

type unit struct {
	path string
	kind string
	raw  map[string]any
}

type familyCounts struct {
	precomputed int
	fallback    int
	unknown     int
}

type pathCounts struct {
	units    int
	families map[string]familyCounts
}

func main() {
	units, skipped, err := collectUnits()
	if err != nil {
		fmt.Fprintf(os.Stderr, "audit zone-time coverage: %v\n", err)
		os.Exit(1)
	}

	byPath := map[string]*pathCounts{}
	totals := map[string]familyCounts{}
	for _, family := range familyOrder() {
		totals[family] = familyCounts{}
	}

	for _, u := range units {
		counts := byPath[u.path]
		if counts == nil {
			counts = &pathCounts{families: map[string]familyCounts{}}
			for _, family := range familyOrder() {
				counts.families[family] = familyCounts{}
			}
			byPath[u.path] = counts
		}
		counts.units++

		for _, family := range applicableFamilies(u) {
			status := classifyFamily(u.raw, family)
			pathFamily := counts.families[family]
			totalFamily := totals[family]
			switch status {
			case "precomputed":
				pathFamily.precomputed++
				totalFamily.precomputed++
			case "fallback":
				pathFamily.fallback++
				totalFamily.fallback++
			case "unknown":
				pathFamily.unknown++
				totalFamily.unknown++
			}
			counts.families[family] = pathFamily
			totals[family] = totalFamily
		}
	}

	printReport(units, skipped, byPath, totals)
}

func collectUnits() ([]unit, int, error) {
	var units []unit
	skipped := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".json" || excludedPath(path) {
				skipped++
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var value any
			if err := json.Unmarshal(data, &value); err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			for _, raw := range objectUnits(value) {
				kind, ok := eligibleKind(raw)
				if !ok {
					skipped++
					continue
				}
				units = append(units, unit{path: filepath.ToSlash(path), kind: kind, raw: raw})
			}
			return nil
		})
		if err != nil {
			return nil, 0, err
		}
	}
	return units, skipped, nil
}

func objectUnits(value any) []map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return []map[string]any{typed}
	case []any:
		units := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if raw, ok := item.(map[string]any); ok {
				units = append(units, raw)
			}
		}
		return units
	default:
		return nil
	}
}

func excludedPath(path string) bool {
	path = filepath.ToSlash(path)
	excludedParts := []string{
		"/wellness/",
		"/events/",
		"/workout_library/",
		"/custom_items/",
		"/activity_intervals/",
		"/analyzer/",
		"/schema_snapshot/",
	}
	for _, part := range excludedParts {
		if strings.Contains(path, part) {
			return true
		}
	}
	switch filepath.Base(path) {
	case "activity_messages.json", "athlete_profile.json", "gear_list.json", "gear_list_empty.json":
		return true
	default:
		return false
	}
}

func eligibleKind(raw map[string]any) (string, bool) {
	if hasEventMarker(raw) {
		return "", false
	}
	if hasString(raw, "date") && (hasKey(raw, "timeInZones") || hasKey(raw, "timeInZonesTot")) {
		return "training_summary", true
	}
	if hasString(raw, "id") && hasAnyKey(raw, allZoneKeys()...) {
		return "extended_metrics", true
	}
	if hasString(raw, "id") && (hasString(raw, "start_date") || hasString(raw, "start_date_local")) {
		return "activity", true
	}
	return "", false
}

func hasEventMarker(raw map[string]any) bool {
	return hasAnyKey(raw, "category", "workout_doc", "show_as_note", "calendar_id", "paired_event_id")
}

func applicableFamilies(u unit) []string {
	if u.kind == "training_summary" {
		return []string{"power"}
	}
	out := make([]string, 0, 3)
	for _, family := range familyOrder() {
		if validZoneArray(rawValue(u.raw, precomputedKeys[family]...)) || hasFamilySignal(u.raw, family) {
			out = append(out, family)
		}
	}
	if len(out) == 0 {
		return familyOrder()
	}
	return out
}

func classifyFamily(raw map[string]any, family string) string {
	if family == "power" && validSummaryZones(raw) {
		return "precomputed"
	}
	if validZoneArray(rawValue(raw, precomputedKeys[family]...)) {
		return "precomputed"
	}
	if hasFamilySignal(raw, family) || hasAnyKey(raw, precomputedKeys[family]...) {
		return "fallback"
	}
	return "unknown"
}

func validSummaryZones(raw map[string]any) bool {
	return positiveNumber(raw["timeInZonesTot"]) && validZoneArray(raw["timeInZones"])
}

func validZoneArray(value any) bool {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return false
	}
	total := 0.0
	for _, item := range items {
		number, ok := item.(float64)
		if !ok {
			return false
		}
		total += number
	}
	return total > 0
}

func hasFamilySignal(raw map[string]any, family string) bool {
	switch family {
	case "power":
		return positiveAny(raw, "icu_training_load", "power_load", "average_watts", "weighted_average_watts", "max_watts") || streamTypeContains(raw, "watts", "power")
	case "heart_rate":
		return positiveAny(raw, "hr_load", "average_heartrate", "max_heartrate") || streamTypeContains(raw, "heartrate", "heart_rate", "hr")
	case "pace":
		return positiveAny(raw, "pace_load", "average_speed", "max_speed") || (positiveAny(raw, "distance", "icu_distance") && positiveAny(raw, "moving_time", "elapsed_time")) || streamTypeContains(raw, "velocity_smooth", "pace", "speed")
	default:
		return false
	}
}

func streamTypeContains(raw map[string]any, needles ...string) bool {
	value, ok := raw["stream_types"].([]any)
	if !ok {
		return false
	}
	for _, item := range value {
		streamType, ok := item.(string)
		if !ok {
			continue
		}
		streamType = strings.ToLower(streamType)
		for _, needle := range needles {
			if strings.Contains(streamType, needle) {
				return true
			}
		}
	}
	return false
}

func rawValue(raw map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := raw[key]; ok {
			return value
		}
	}
	return nil
}

func hasKey(raw map[string]any, key string) bool {
	_, ok := raw[key]
	return ok
}

func hasAnyKey(raw map[string]any, keys ...string) bool {
	for _, key := range keys {
		if hasKey(raw, key) {
			return true
		}
	}
	return false
}

func hasString(raw map[string]any, key string) bool {
	value, ok := raw[key].(string)
	return ok && strings.TrimSpace(value) != ""
}

func positiveAny(raw map[string]any, keys ...string) bool {
	for _, key := range keys {
		if positiveNumber(raw[key]) {
			return true
		}
	}
	return false
}

func positiveNumber(value any) bool {
	number, ok := value.(float64)
	return ok && number > 0
}

func allZoneKeys() []string {
	var keys []string
	for _, family := range familyOrder() {
		keys = append(keys, precomputedKeys[family]...)
	}
	return keys
}

func familyOrder() []string {
	return []string{"power", "heart_rate", "pace"}
}

func printReport(units []unit, skipped int, byPath map[string]*pathCounts, totals map[string]familyCounts) {
	fmt.Println("# Zone-time upstream fixture coverage audit")
	fmt.Println()
	fmt.Printf("fixture_count: %d eligible object(s)\n", len(units))
	fmt.Printf("skipped_objects_or_files: %d\n", skipped)
	fmt.Println()
	fmt.Println("## Totals by metric family")
	fmt.Println()
	fmt.Println("| metric_family | precomputed_count | fallback_count | unknown_count | coverage |")
	fmt.Println("|---|---:|---:|---:|---:|")
	for _, family := range familyOrder() {
		counts := totals[family]
		fmt.Printf("| %s | %d | %d | %d | %.1f%% |\n", family, counts.precomputed, counts.fallback, counts.unknown, coverage(counts))
	}
	fmt.Println()
	fmt.Println("## By source path")
	fmt.Println()
	fmt.Println("| path | fixture_units | metric_family | precomputed_count | fallback_count | unknown_count |")
	fmt.Println("|---|---:|---|---:|---:|---:|")
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		counts := byPath[path]
		for _, family := range familyOrder() {
			familyCounts := counts.families[family]
			fmt.Printf("| %s | %d | %s | %d | %d | %d |\n", path, counts.units, family, familyCounts.precomputed, familyCounts.fallback, familyCounts.unknown)
		}
	}
}

func coverage(counts familyCounts) float64 {
	known := counts.precomputed + counts.fallback
	if known == 0 {
		return 0
	}
	return float64(counts.precomputed) / float64(known) * 100
}
