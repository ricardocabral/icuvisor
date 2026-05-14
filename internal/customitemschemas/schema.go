package customitemschemas

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

// ContentSchema describes the inferred JSON shape of custom-item content.
type ContentSchema struct {
	Kind       string
	Props      map[string]*ContentSchema
	Elem       *ContentSchema
	SourceHits int
}

// SchemaPath describes one inferred content path and JSON kind.
type SchemaPath struct {
	Path string
	Kind string
}

// ValidateContentAgainstReadSchema validates content against readable custom-item samples.
func ValidateContentAgainstReadSchema(items []intervals.CustomItem, itemType string, content map[string]any, requireComplete bool) (int, error) {
	itemType = strings.TrimSpace(itemType)
	if itemType == "" {
		return 0, errors.New("item_type is required")
	}
	var schema *ContentSchema
	var sourceCount int
	for _, item := range items {
		if customItemType(item) != itemType || item.Content == nil {
			continue
		}
		contentMap, ok := item.Content.(map[string]any)
		if !ok {
			continue
		}
		if schema == nil {
			schema = &ContentSchema{Kind: "object", Props: map[string]*ContentSchema{}}
		}
		MergeContentSchema(schema, contentMap)
		sourceCount++
	}
	if schema == nil || sourceCount == 0 {
		return 0, fmt.Errorf("no readable custom item schema found for item_type %q", itemType)
	}
	if err := ValidateValue(schema, content, "content"); err != nil {
		return sourceCount, err
	}
	if requireComplete {
		missing := MissingContentKeys(schema, content)
		if len(missing) > 0 {
			return sourceCount, fmt.Errorf("content missing schema keys: %s", strings.Join(missing, ", "))
		}
	}
	return sourceCount, nil
}

// InferContentSchema infers a schema from representative content samples.
func InferContentSchema(samples []map[string]any) (*ContentSchema, error) {
	if len(samples) == 0 {
		return nil, errors.New("at least one content sample is required")
	}
	schema := &ContentSchema{Kind: "object", Props: map[string]*ContentSchema{}}
	for _, sample := range samples {
		if sample == nil {
			return nil, errors.New("content sample must be an object")
		}
		MergeContentSchema(schema, sample)
	}
	return schema, nil
}

// MergeContentSchema merges one JSON value into an inferred content schema.
func MergeContentSchema(schema *ContentSchema, value any) {
	kind := JSONKind(value)
	if schema.Kind == "" {
		schema.Kind = kind
	}
	if schema.Kind != kind {
		schema.Kind = "mixed"
	}
	schema.SourceHits++
	switch typed := value.(type) {
	case map[string]any:
		if schema.Props == nil {
			schema.Props = map[string]*ContentSchema{}
		}
		for key, child := range typed {
			childSchema := schema.Props[key]
			if childSchema == nil {
				childSchema = &ContentSchema{}
				schema.Props[key] = childSchema
			}
			MergeContentSchema(childSchema, child)
		}
	case []any:
		if schema.Elem == nil {
			schema.Elem = &ContentSchema{}
		}
		for _, child := range typed {
			MergeContentSchema(schema.Elem, child)
		}
	}
}

// ValidateValue validates a JSON value against an inferred schema.
func ValidateValue(schema *ContentSchema, value any, path string) error {
	kind := JSONKind(value)
	if schema.Kind != "mixed" && schema.Kind != kind {
		return fmt.Errorf("%s must be %s", path, schema.Kind)
	}
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			childSchema := schema.Props[key]
			if childSchema == nil {
				return fmt.Errorf("%s.%s is not in the readable schema", path, key)
			}
			if err := ValidateValue(childSchema, child, path+"."+key); err != nil {
				return err
			}
		}
	case []any:
		if schema.Elem == nil || schema.Elem.Kind == "" {
			return nil
		}
		for i, child := range typed {
			if err := ValidateValue(schema.Elem, child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// MissingContentKeys returns required top-level keys absent from content.
func MissingContentKeys(schema *ContentSchema, content map[string]any) []string {
	missing := []string{}
	for key := range schema.Props {
		if _, ok := content[key]; !ok {
			missing = append(missing, "content."+key)
		}
	}
	sort.Strings(missing)
	return missing
}

// SchemaPaths returns deterministic inferred paths for a content schema.
func SchemaPaths(schema *ContentSchema) []SchemaPath {
	if schema == nil {
		return nil
	}
	var paths []SchemaPath
	collectSchemaPaths(&paths, "content", schema)
	return paths
}

func collectSchemaPaths(paths *[]SchemaPath, path string, schema *ContentSchema) {
	if schema == nil || schema.Kind == "" {
		return
	}
	*paths = append(*paths, SchemaPath{Path: path, Kind: schema.Kind})
	keys := make([]string, 0, len(schema.Props))
	for key := range schema.Props {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		collectSchemaPaths(paths, path+"."+key, schema.Props[key])
	}
	if schema.Elem != nil {
		collectSchemaPaths(paths, path+"[]", schema.Elem)
	}
}

// JSONKind returns a stable JSON-kind label for schema inference.
func JSONKind(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case float64, float32, int, int64, int32, json.Number:
		return "number"
	case bool:
		return "boolean"
	default:
		return "value"
	}
}

func customItemType(item intervals.CustomItem) string {
	if item.Type != nil && strings.TrimSpace(*item.Type) != "" {
		return strings.TrimSpace(*item.Type)
	}
	for _, key := range []string{"type", "item_type"} {
		if value, ok := item.Raw[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
