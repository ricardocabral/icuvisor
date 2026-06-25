package tools

import (
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	workoutStatusPlanned           = "planned"
	workoutStatusCompletedLinked   = "completed_linked"
	workoutStatusCompletedUnlinked = "completed_unlinked"
	workoutStatusMissedOrSkipped   = "missed_or_skipped"
	workoutStatusFuture            = "future"

	workoutCaveatSkippedMissedUnavailable = "upstream_does_not_distinguish_skipped_from_missed"
	workoutCaveatDeletedAbsentUnavailable = "deleted_or_cancelled_unavailable_when_event_absent"
	workoutCaveatUnknownLocalDate         = "unknown_local_date"
	workoutCaveatDateMetricNotExplicit    = "date_metric_match_not_explicit_link"
	workoutCaveatUnlinkedActivity         = "unlinked_activity_does_not_prove_planned_event_completion"
)

func applyEventWorkoutStatus(row *getEventsRow, event intervals.Event, asOfDate string) {
	if row == nil || !isWorkoutTargetEvent(event) {
		return
	}
	if pairedID := eventPairedActivityID(event); pairedID != "" {
		row.WorkoutStatus = workoutStatusCompletedLinked
		row.PairedActivityID = pairedID
		return
	}
	status, caveats := dateDerivedWorkoutStatus(localDatePrefix(stringValue(event.StartDateLocal)), asOfDate)
	row.WorkoutStatus = status
	row.WorkoutStatusCaveats = caveats
}

func applyCompletedActivityWorkoutStatus(row *getActivitiesRow, activity intervals.Activity) {
	if row == nil {
		return
	}
	if pairedID := activityPairedEventID(activity); pairedID != "" {
		row.WorkoutStatus = workoutStatusCompletedLinked
		row.PairedEventID = pairedID
		return
	}
	row.WorkoutStatus = workoutStatusCompletedUnlinked
	row.WorkoutStatusCaveats = []string{workoutCaveatUnlinkedActivity}
}

func dateDerivedWorkoutStatus(eventDate string, asOfDate string) (string, []string) {
	if eventDate == "" || asOfDate == "" {
		return "", []string{workoutCaveatUnknownLocalDate}
	}
	switch {
	case eventDate < asOfDate:
		return workoutStatusMissedOrSkipped, []string{workoutCaveatSkippedMissedUnavailable, workoutCaveatDeletedAbsentUnavailable}
	case eventDate == asOfDate:
		return workoutStatusPlanned, nil
	default:
		return workoutStatusFuture, nil
	}
}

func isWorkoutTargetEvent(event intervals.Event) bool {
	category := strings.ToUpper(strings.TrimSpace(firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"]))))
	if category == "WORKOUT" {
		return true
	}
	if category != "" {
		return false
	}
	return event.WorkoutDoc != nil || event.LoadTarget != nil || event.DistanceTarget != nil || event.TimeTarget != nil || event.ElapsedTimeTarget != nil || anyString(event.Raw["workout_doc"]) != ""
}

func eventPairedActivityID(event intervals.Event) string {
	for _, key := range []string{"activity_id", "icu_activity_id", "paired_activity_id", "completed_activity_id"} {
		if value := anyString(event.Raw[key]); value != "" {
			return value
		}
	}
	return ""
}

func activityPairedEventID(activity intervals.Activity) string {
	for _, key := range []string{"paired_event_id", "event_id", "calendar_event_id", "icu_event_id"} {
		if value := anyString(activity.Raw[key]); value != "" {
			return value
		}
	}
	return ""
}

func appendCaveats(values []string, caveats ...string) []string {
	seen := make(map[string]bool, len(values)+len(caveats))
	out := make([]string, 0, len(values)+len(caveats))
	for _, value := range append(values, caveats...) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
