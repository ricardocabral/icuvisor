package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/customitemschemas"
)

const (
	CustomItemSchemasURI      = "icuvisor://custom-item-schemas"
	CustomItemSchemasMIMEType = "text/markdown"
)

// CustomItemSchemasResource returns the custom-item content schema resource definition.
func CustomItemSchemasResource() Resource {
	return Resource{
		URI:         CustomItemSchemasURI,
		Name:        "custom_item_schemas",
		Title:       "Custom item schemas",
		Description: "Representative content schemas for Intervals.icu custom item types.",
		MIMEType:    CustomItemSchemasMIMEType,
		Handler: func(ctx context.Context, _ Request) (Result, error) {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			text, err := CustomItemSchemasMarkdown()
			if err != nil {
				return Result{}, err
			}
			return Result{URI: CustomItemSchemasURI, MIMEType: CustomItemSchemasMIMEType, Text: text}, nil
		},
	}
}

// CustomItemSchemasMarkdown renders static custom-item schema guidance from shared descriptors.
func CustomItemSchemasMarkdown() (string, error) {
	families := customitemschemas.Families()
	if len(families) == 0 {
		return "", fmt.Errorf("custom-item schema descriptor is empty")
	}
	var b strings.Builder
	b.WriteString("# Custom item content schemas\n\n")
	b.WriteString("This resource provides representative `content` samples for known Intervals.icu custom item families. icuvisor write tools still validate create/update payloads against readable custom items for the target athlete/item; these samples are guidance, not a validation allow-list, and unknown upstream item types can still pass through when the upstream API supports them.\n")
	for _, family := range families {
		if family.Key == "" || family.Title == "" || family.Description == "" || len(family.ItemTypes) == 0 || family.Sample == nil {
			return "", fmt.Errorf("custom-item schema family %q is incomplete", family.Key)
		}
		b.WriteString("\n## ")
		b.WriteString(family.Title)
		b.WriteString("\n\n")
		b.WriteString("Descriptor key: `")
		b.WriteString(family.Key)
		b.WriteString("`\n\n")
		b.WriteString(family.Description)
		b.WriteString("\n\n")
		b.WriteString("Item types: `")
		b.WriteString(strings.Join(family.ItemTypes, "`, `"))
		b.WriteString("`\n\n")
		b.WriteString("Representative `content` sample:\n\n")
		b.WriteString("```json\n")
		sample, err := json.MarshalIndent(family.Sample, "", "  ")
		if err != nil {
			return "", fmt.Errorf("rendering custom-item sample %s: %w", family.Key, err)
		}
		b.Write(sample)
		b.WriteString("\n```\n\n")
		schema, err := customitemschemas.InferContentSchema([]map[string]any{family.Sample})
		if err != nil {
			return "", fmt.Errorf("inferring custom-item schema %s: %w", family.Key, err)
		}
		b.WriteString("Inferred paths:\n\n")
		for _, path := range customitemschemas.SchemaPaths(schema) {
			b.WriteString("- `")
			b.WriteString(path.Path)
			b.WriteString("`: ")
			b.WriteString(path.Kind)
			b.WriteString("\n")
		}
	}
	return b.String(), nil
}
