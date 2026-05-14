package tools

import (
	"strings"
	"testing"
)

func TestEventCategorySchemaDescriptionsReferenceResourceWithoutValidationEnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema map[string]any
	}{
		{name: "get_events", schema: getEventsInputSchema()},
		{name: "add_or_update_event", schema: addOrUpdateEventInputSchema()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			properties, ok := tc.schema["properties"].(map[string]any)
			if !ok {
				t.Fatalf("properties type = %T", tc.schema["properties"])
			}
			category, ok := properties["category"].(map[string]any)
			if !ok {
				t.Fatalf("category schema type = %T", properties["category"])
			}
			description, _ := category["description"].(string)
			for _, want := range []string{"icuvisor://event-categories", "WORKOUT", "SET_FITNESS", "Custom athlete/account category values are passed through"} {
				if !strings.Contains(description, want) {
					t.Fatalf("category description = %q, missing %q", description, want)
				}
			}
			if _, exists := category["enum"]; exists {
				t.Fatalf("category schema unexpectedly validates against enum: %#v", category)
			}
		})
	}
}
