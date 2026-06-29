package tools

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type dataAvailabilityDiagnostic struct {
	Reason        string   `json:"reason"`
	Message       string   `json:"message"`
	Workaround    string   `json:"workaround,omitempty"`
	ActivityID    string   `json:"activity_id,omitempty"`
	Requested     []string `json:"requested,omitempty"`
	Available     []string `json:"available,omitempty"`
	SourceFields  []string `json:"source_fields,omitempty"`
	MissingFields []string `json:"missing_fields,omitempty"`
	Dates         []string `json:"dates,omitempty"`
}

func restrictedSourceDiagnostic(activityID string, unavailable *unavailableReason) *dataAvailabilityDiagnostic {
	if unavailable == nil || unavailable.Reason != "strava_blocked" {
		return nil
	}
	return &dataAvailabilityDiagnostic{
		Reason:     "restricted_source",
		ActivityID: activityID,
		Message:    "The activity summary is visible, but the source is Strava-restricted so detailed streams, intervals, and max-heart-rate samples may be unavailable through the API.",
		Workaround: unavailable.Workaround,
	}
}

func activityAvailabilityDiagnostics(rows []getActivitiesRow) []dataAvailabilityDiagnostic {
	diagnostics := make([]dataAvailabilityDiagnostic, 0)
	for _, row := range rows {
		if diagnostic := restrictedSourceDiagnostic(row.ActivityID, row.Unavailable); diagnostic != nil {
			diagnostics = append(diagnostics, *diagnostic)
		}
	}
	return diagnostics
}

func activityStreamMissingDiagnostics(activityID string, requested []string, streams map[string]activityStreamRow) []dataAvailabilityDiagnostic {
	if len(requested) == 0 {
		return nil
	}
	availableSet := map[string]bool{}
	for key := range streams {
		availableSet[key] = true
	}
	missingSet := map[string]bool{}
	for _, key := range requested {
		if key == "" || availableSet[key] {
			continue
		}
		missingSet[key] = true
	}
	missing := sortedBoolKeys(missingSet)
	if len(missing) == 0 {
		return nil
	}
	return []dataAvailabilityDiagnostic{{
		Reason:     "missing_stream",
		ActivityID: activityID,
		Requested:  missing,
		Available:  sortedBoolKeys(availableSet),
		Message:    "Requested stream channels are absent from the API response; answers that require those samples, such as max-heart-rate-from-stream, may be incomplete.",
		Workaround: "Use activity summary fields when present. If this is a Strava-sourced activity, re-import it directly from the native provider (Garmin, Zwift, Wahoo, etc.) from intervals.icu Connections so streams are available through the API.",
	}}
}

func loadDiagnostics(rows []intervals.SummaryWithCats) []dataAvailabilityDiagnostic {
	missingLoadDates := map[string]bool{}
	alternateFields := map[string]bool{}
	missingFitnessFields := map[string]bool{}
	missingFitnessDates := map[string]bool{}
	for _, row := range rows {
		load := summaryTrainingLoad(row)
		if !load.Present {
			missingLoadDates[row.Date] = true
		}
		if load.Source != "" && load.Source != "training_load" {
			alternateFields[load.Source] = true
		}
		for _, field := range []string{"fitness", "fatigue", "form"} {
			if !rawFieldPresent(row.Raw, field) {
				missingFitnessFields[field] = true
				missingFitnessDates[row.Date] = true
			}
		}
	}
	diagnostics := make([]dataAvailabilityDiagnostic, 0, 3)
	if len(alternateFields) > 0 {
		diagnostics = append(diagnostics, dataAvailabilityDiagnostic{Reason: "trimp_or_hr_load_available", SourceFields: sortedBoolKeys(alternateFields), Message: "Training load was preserved from HR/TRIMP-like upstream fields and is exposed as neutral training_load, not TSS."})
	}
	if len(missingLoadDates) > 0 {
		diagnostics = append(diagnostics, dataAvailabilityDiagnostic{Reason: "missing_training_load", MissingFields: []string{"training_load"}, Dates: sortedBoolKeys(missingLoadDates), Message: "Some summary rows omit training_load and have no recognized HR/TRIMP fallback field; load-dependent totals and trends treat those rows as zero rather than inventing TSS."})
	}
	if len(missingFitnessFields) > 0 {
		diagnostics = append(diagnostics, dataAvailabilityDiagnostic{Reason: "fitness_fields_missing", MissingFields: sortedBoolKeys(missingFitnessFields), Dates: sortedBoolKeys(missingFitnessDates), Message: "Some summary rows omit CTL/ATL/TSB source fields; missing fields are omitted from fitness rows instead of being reported as zero."})
	}
	return diagnostics
}

type loadValue struct {
	Value   int
	Source  string
	Present bool
}

func summaryTrainingLoad(row intervals.SummaryWithCats) loadValue {
	return rawTrainingLoad(row.Raw, row.TrainingLoad)
}

func categoryTrainingLoad(category intervals.CategorySummary) loadValue {
	return rawTrainingLoad(category.Raw, category.TrainingLoad)
}

func rawTrainingLoad(raw map[string]any, fallback int) loadValue {
	for _, key := range []string{"training_load", "trimp", "hr_load"} {
		if value, ok := rawNumericValue(raw, key); ok {
			return loadValue{Value: int(math.Round(value)), Source: key, Present: true}
		}
	}
	if raw == nil && fallback != 0 {
		return loadValue{Value: fallback, Source: "training_load", Present: true}
	}
	return loadValue{}
}

func rawFloatPtr(raw map[string]any, key string, fallback float64) *float64 {
	if value, ok := rawNumericValue(raw, key); ok {
		rounded := round(value, 3)
		return &rounded
	}
	if raw == nil {
		return roundPtr(fallback)
	}
	return nil
}

func rawNumericValue(raw map[string]any, key string) (float64, bool) {
	if raw == nil {
		return 0, false
	}
	value, ok := raw[key]
	if !ok || value == nil {
		return 0, false
	}
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case json.Number:
		parsed, err := typed.Float64()
		return parsed, err == nil
	default:
		parsed, err := parseNumericText(fmt.Sprint(typed))
		return parsed, err == nil
	}
}

func rawFieldPresent(raw map[string]any, key string) bool {
	if raw == nil {
		return true
	}
	_, ok := raw[key]
	return ok
}

func parseNumericText(value string) (float64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("empty numeric value")
	}
	return strconv.ParseFloat(trimmed, 64)
}

func sortedBoolKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
