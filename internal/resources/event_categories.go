package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	EventCategoriesURI      = "icuvisor://event-categories"
	EventCategoriesMIMEType = "text/markdown"
)

// EventCategoriesResource returns the event-category enum resource definition.
func EventCategoriesResource() Resource {
	return Resource{
		URI:         EventCategoriesURI,
		Name:        "event_categories",
		Title:       "Event categories",
		Description: "Intervals.icu calendar event category enum values and meanings.",
		MIMEType:    EventCategoriesMIMEType,
		Handler: func(ctx context.Context, _ Request) (Result, error) {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			text, err := EventCategoriesMarkdown()
			if err != nil {
				return Result{}, err
			}
			return Result{URI: EventCategoriesURI, MIMEType: EventCategoriesMIMEType, Text: text}, nil
		},
	}
}

// EventCategoriesMarkdown renders documented event categories from the intervals descriptor.
func EventCategoriesMarkdown() (string, error) {
	categories := intervals.EventCategories()
	if len(categories) == 0 {
		return "", fmt.Errorf("event category descriptor is empty")
	}
	var b strings.Builder
	b.WriteString("# Event categories\n\n")
	b.WriteString("Documented Intervals.icu calendar event categories from the public OpenAPI `Event.category` / `EventEx.category` enum. icuvisor preserves and passes through custom upstream category values; this resource documents the known enum, not a validation allow-list.\n\n")
	b.WriteString("| Category | Description |\n")
	b.WriteString("| --- | --- |\n")
	seen := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		if strings.TrimSpace(category.Value) == "" || strings.TrimSpace(category.Description) == "" {
			return "", fmt.Errorf("event category descriptor has empty value or description")
		}
		if _, exists := seen[category.Value]; exists {
			return "", fmt.Errorf("duplicate event category %q", category.Value)
		}
		seen[category.Value] = struct{}{}
		b.WriteString("| `")
		b.WriteString(category.Value)
		b.WriteString("` | ")
		b.WriteString(category.Description)
		b.WriteString(" |\n")
	}
	return b.String(), nil
}
