package intervals

import "strings"

const eventCategoriesResourceURI = "icuvisor://event-categories"

// EventCategory describes one documented intervals.icu calendar event category.
type EventCategory struct {
	Value       string
	Description string
}

var eventCategories = []EventCategory{
	{Value: "WORKOUT", Description: "Planned workout or training session on the athlete calendar."},
	{Value: "RACE_A", Description: "Highest-priority race or A goal event."},
	{Value: "RACE_B", Description: "Secondary-priority race or B goal event."},
	{Value: "RACE_C", Description: "Lower-priority race or C goal event."},
	{Value: "NOTE", Description: "Calendar note or general annotation."},
	{Value: "PLAN", Description: "Training-plan marker or plan-level calendar entry."},
	{Value: "HOLIDAY", Description: "Holiday or planned time away from normal training."},
	{Value: "SICK", Description: "Illness marker affecting training availability."},
	{Value: "INJURED", Description: "Injury marker affecting training availability."},
	{Value: "SET_EFTP", Description: "Fitness-model event that sets estimated FTP."},
	{Value: "FITNESS_DAYS", Description: "Fitness-model event that adjusts modeled fitness days."},
	{Value: "SEASON_START", Description: "Season start marker used by planning and fitness views."},
	{Value: "TARGET", Description: "Target event or goal marker."},
	{Value: "SET_FITNESS", Description: "Fitness-model event that sets modeled fitness."},
}

// EventCategories returns documented calendar event categories in upstream OpenAPI order.
func EventCategories() []EventCategory {
	out := make([]EventCategory, len(eventCategories))
	copy(out, eventCategories)
	return out
}

// EventCategoryValues returns documented category values in upstream OpenAPI order.
func EventCategoryValues() []string {
	categories := EventCategories()
	values := make([]string, 0, len(categories))
	for _, category := range categories {
		values = append(values, category.Value)
	}
	return values
}

// EventCategoryReferenceDescription describes event-category arguments without restricting custom upstream values.
func EventCategoryReferenceDescription(prefix string) string {
	values := strings.Join(EventCategoryValues(), ", ")
	return strings.TrimSpace(prefix) + " Documented upstream categories include " + values + "; see " + eventCategoriesResourceURI + ". Custom athlete/account category values are passed through."
}
