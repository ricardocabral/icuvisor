package toolcatalog

import (
	"fmt"
	"sort"
	"strings"
)

const (
	AddActivityMessage               = "add_activity_message"
	AddOrUpdateEvent                 = "add_or_update_event"
	ApplyTrainingPlan                = "apply_training_plan"
	CreateCustomItem                 = "create_custom_item"
	CreateWorkout                    = "create_workout"
	DeleteActivity                   = "delete_activity"
	DeleteCustomItem                 = "delete_custom_item"
	DeleteEvent                      = "delete_event"
	DeleteEventsByDateRange          = "delete_events_by_date_range"
	DeleteGear                       = "delete_gear"
	DeleteSportSettings              = "delete_sport_settings"
	DeleteWorkout                    = "delete_workout"
	GetActivities                    = "get_activities"
	GetActivityDetails               = "get_activity_details"
	GetActivityIntervals             = "get_activity_intervals"
	GetActivityMessages              = "get_activity_messages"
	GetActivitySplits                = "get_activity_splits"
	GetActivityStreams               = "get_activity_streams"
	GetAthleteProfile                = "get_athlete_profile"
	GetBestEfforts                   = "get_best_efforts"
	GetCustomItemByID                = "get_custom_item_by_id"
	GetCustomItems                   = "get_custom_items"
	GetEventByID                     = "get_event_by_id"
	GetEvents                        = "get_events"
	GetExtendedMetrics               = "get_extended_metrics"
	GetFitness                       = "get_fitness"
	GetPowerCurves                   = "get_power_curves"
	GetTrainingPlan                  = "get_training_plan"
	GetTrainingSummary               = "get_training_summary"
	GetWellnessData                  = "get_wellness_data"
	GetWorkoutLibrary                = "get_workout_library"
	GetWorkoutsInFolder              = "get_workouts_in_folder"
	ICUvisorListAdvancedCapabilities = "icuvisor_list_advanced_capabilities"
	LinkActivityToEvent              = "link_activity_to_event"
	ListAthletes                     = "list_athletes"
	SelectAthlete                    = "select_athlete"
	UpdateCustomItem                 = "update_custom_item"
	UpdateSportSettings              = "update_sport_settings"
	UpdateWellness                   = "update_wellness"
	UpdateWorkout                    = "update_workout"
)

var athleteScopedToolNames = []string{
	AddActivityMessage,
	AddOrUpdateEvent,
	ApplyTrainingPlan,
	CreateCustomItem,
	CreateWorkout,
	DeleteActivity,
	DeleteCustomItem,
	DeleteEvent,
	DeleteEventsByDateRange,
	DeleteGear,
	DeleteSportSettings,
	DeleteWorkout,
	GetActivities,
	GetActivityDetails,
	GetActivityIntervals,
	GetActivityMessages,
	GetActivitySplits,
	GetActivityStreams,
	GetAthleteProfile,
	GetBestEfforts,
	GetCustomItemByID,
	GetCustomItems,
	GetEventByID,
	GetEvents,
	GetExtendedMetrics,
	GetFitness,
	GetPowerCurves,
	GetTrainingPlan,
	GetTrainingSummary,
	GetWellnessData,
	GetWorkoutLibrary,
	GetWorkoutsInFolder,
	LinkActivityToEvent,
	UpdateCustomItem,
	UpdateSportSettings,
	UpdateWellness,
	UpdateWorkout,
}

var allToolNames = append(append([]string{}, athleteScopedToolNames...), ICUvisorListAdvancedCapabilities, ListAthletes, SelectAthlete)

// AthleteScopedToolNames returns the canonical sorted names accepted by per-athlete ACLs.
func AthleteScopedToolNames() []string {
	return sortedCopy(athleteScopedToolNames)
}

// AllToolNames returns every canonical MCP tool name known to icuvisor.
func AllToolNames() []string {
	return sortedCopy(allToolNames)
}

// IsKnownTool reports whether name is in the canonical MCP catalog.
func IsKnownTool(name string) bool {
	return contains(allToolNames, strings.TrimSpace(name))
}

// IsAthleteScopedTool reports whether name is eligible for per-athlete ACLs.
func IsAthleteScopedTool(name string) bool {
	return contains(athleteScopedToolNames, strings.TrimSpace(name))
}

// ValidateACLPattern validates an exact tool name, "*", or a suffix wildcard like "get_*".
func ValidateACLPattern(pattern string) (string, error) {
	trimmed := strings.TrimSpace(pattern)
	if trimmed == "" {
		return "", fmt.Errorf("empty tool ACL pattern")
	}
	if trimmed == "*" {
		return trimmed, nil
	}
	if strings.Count(trimmed, "*") > 1 || (strings.Contains(trimmed, "*") && !strings.HasSuffix(trimmed, "*")) {
		return "", fmt.Errorf("invalid tool ACL pattern %q; use a known tool name, '*', or a prefix wildcard like 'get_*'", trimmed)
	}
	if strings.HasSuffix(trimmed, "*") {
		prefix := strings.TrimSuffix(trimmed, "*")
		if prefix == "" || !matchesPrefix(athleteScopedToolNames, prefix) {
			return "", fmt.Errorf("unknown tool ACL pattern %q; it matches no athlete-scoped tools", trimmed)
		}
		return trimmed, nil
	}
	if !IsAthleteScopedTool(trimmed) {
		return "", fmt.Errorf("unknown athlete-scoped tool %q", trimmed)
	}
	return trimmed, nil
}

func sortedCopy(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	return out
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func matchesPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
