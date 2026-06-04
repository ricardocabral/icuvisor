package tools

import (
	"fmt"
	"sort"
	"strings"
)

const maxCatalogInputExamples = 2

// ToolSchemaDescriptor describes one tool's input arguments for generated documentation.
type ToolSchemaDescriptor struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Arguments   []ToolSchemaArgument `json:"arguments"`
	Examples    []any                `json:"examples,omitempty"`
}

// ToolSchemaArgument is a concise JSON Schema projection for generated documentation.
type ToolSchemaArgument struct {
	Name                 string               `json:"name,omitempty"`
	Required             bool                 `json:"required"`
	Type                 string               `json:"type,omitempty"`
	Description          string               `json:"description,omitempty"`
	Enum                 []any                `json:"enum,omitempty"`
	Default              any                  `json:"default,omitempty"`
	Format               string               `json:"format,omitempty"`
	Items                *ToolSchemaArgument  `json:"items,omitempty"`
	Children             []ToolSchemaArgument `json:"children,omitempty"`
	AdditionalProperties string               `json:"additional_properties,omitempty"`
}

// SchemaCatalog returns a generated-docs projection of registered tool input schemas.
func SchemaCatalog() map[string]ToolSchemaDescriptor {
	tools := catalogTools()
	catalog := make(map[string]ToolSchemaDescriptor, len(tools))
	for _, tool := range tools {
		schema, _ := tool.InputSchema.(map[string]any)
		catalog[tool.Name] = ToolSchemaDescriptor{
			Name:        tool.Name,
			Description: toolSummary(tool.Description),
			Arguments:   schemaCatalogArguments(schema),
			Examples:    schemaCatalogExamples(schema),
		}
	}
	return catalog
}

func schemaCatalogArguments(schema map[string]any) []ToolSchemaArgument {
	properties, _ := schema["properties"].(map[string]any)
	if len(properties) == 0 {
		return nil
	}
	required := schemaRequiredSet(schema["required"])
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)

	args := make([]ToolSchemaArgument, 0, len(names))
	for _, name := range names {
		prop, _ := properties[name].(map[string]any)
		args = append(args, schemaCatalogArgument(name, prop, required[name], 0))
	}
	return args
}

func schemaCatalogArgument(name string, schema map[string]any, required bool, depth int) ToolSchemaArgument {
	arg := ToolSchemaArgument{
		Name:                 name,
		Required:             required,
		Type:                 schemaTypeLabel(schema),
		Description:          stringSchemaValue(schema["description"]),
		Enum:                 schemaAnySlice(schema["enum"]),
		Default:              schema["default"],
		Format:               stringSchemaValue(schema["format"]),
		AdditionalProperties: additionalPropertiesSummary(schema),
	}
	if items, ok := schema["items"].(map[string]any); ok {
		item := schemaCatalogArgument("", items, false, depth+1)
		arg.Items = &item
	}
	if depth == 0 {
		arg.Children = schemaCatalogChildren(schema)
	}
	return arg
}

func schemaCatalogChildren(schema map[string]any) []ToolSchemaArgument {
	properties, _ := schema["properties"].(map[string]any)
	if len(properties) == 0 {
		return nil
	}
	required := schemaRequiredSet(schema["required"])
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	children := make([]ToolSchemaArgument, 0, len(names))
	for _, name := range names {
		prop, _ := properties[name].(map[string]any)
		children = append(children, schemaCatalogArgument(name, prop, required[name], 1))
	}
	return children
}

func schemaCatalogExamples(schema map[string]any) []any {
	if len(schema) == 0 {
		return nil
	}
	examples := schemaAnySlice(schema["input_examples"])
	if len(examples) == 0 {
		examples = schemaAnySlice(schema["examples"])
	}
	if len(examples) > maxCatalogInputExamples {
		examples = examples[:maxCatalogInputExamples]
	}
	return examples
}

func schemaRequiredSet(value any) map[string]bool {
	required := make(map[string]bool)
	for _, name := range schemaStringSlice(value) {
		required[name] = true
	}
	return required
}

func schemaTypeLabel(schema map[string]any) string {
	if len(schema) == 0 {
		return "unknown"
	}
	if label := typeValueLabel(schema["type"]); label != "" {
		return label
	}
	for _, key := range []string{"anyOf", "oneOf"} {
		branches := schemaAnySlice(schema[key])
		if len(branches) == 0 {
			continue
		}
		labels := make([]string, 0, len(branches))
		seen := make(map[string]bool, len(branches))
		for _, branch := range branches {
			branchSchema, _ := branch.(map[string]any)
			label := schemaTypeLabel(branchSchema)
			if label == "" || seen[label] {
				continue
			}
			labels = append(labels, label)
			seen[label] = true
		}
		if len(labels) > 0 {
			return strings.Join(labels, " | ")
		}
	}
	if _, ok := schema["properties"].(map[string]any); ok {
		return "object"
	}
	if _, ok := schema["items"]; ok {
		return "array"
	}
	return "unknown"
}

func typeValueLabel(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []string:
		return strings.Join(typed, " | ")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, part := range typed {
			if text, ok := part.(string); ok && text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " | ")
	default:
		return ""
	}
}

func additionalPropertiesSummary(schema map[string]any) string {
	value, ok := schema["additionalProperties"]
	if !ok {
		if schemaTypeLabel(schema) == "object" {
			return "allowed"
		}
		return ""
	}
	switch typed := value.(type) {
	case bool:
		if typed {
			return "allowed"
		}
		return ""
	case map[string]any:
		return "allowed values: " + schemaTypeLabel(typed)
	default:
		return fmt.Sprintf("allowed values: %v", typed)
	}
}

func schemaAnySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return append([]any(nil), typed...)
	case []map[string]any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	case []string:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	case []int:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	case []float64:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	default:
		return nil
	}
}

func schemaStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func stringSchemaValue(value any) string {
	text, _ := value.(string)
	return text
}
