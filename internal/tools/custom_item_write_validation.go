package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const customItemSchemaDocumentation = "inline_v0.2_tool_description; moves_to_resource_v0.4"

type customItemWriteResponse struct {
	CustomItem map[string]any      `json:"custom_item"`
	Meta       customItemWriteMeta `json:"_meta"`
}

type customItemWriteMeta struct {
	Operation           string   `json:"operation"`
	SourceEndpoint      string   `json:"source_endpoint"`
	ItemID              string   `json:"item_id,omitempty"`
	ItemType            string   `json:"item_type,omitempty"`
	FieldsUpdated       []string `json:"fields_updated,omitempty"`
	ContentPreserved    bool     `json:"content_preserved"`
	SchemaDocumentation string   `json:"schema_documentation"`
	SchemaSourceCount   int      `json:"schema_source_count,omitempty"`
	DefaultPayloadScope string   `json:"default_payload_scope"`
}

type customItemContentSchema struct {
	Kind       string
	Props      map[string]*customItemContentSchema
	Elem       *customItemContentSchema
	SourceHits int
}

func customItemContentFromRaw(raw json.RawMessage) (map[string]any, error) {
	var content map[string]any
	if err := json.Unmarshal(raw, &content); err != nil {
		return nil, err
	}
	if content == nil {
		return nil, errors.New("content must be an object")
	}
	return content, nil
}

func validateCustomItemContentAgainstReadSchema(items []intervals.CustomItem, itemType string, content map[string]any, requireComplete bool) (int, error) {
	itemType = strings.TrimSpace(itemType)
	if itemType == "" {
		return 0, errors.New("item_type is required")
	}
	var schema *customItemContentSchema
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
			schema = &customItemContentSchema{Kind: "object", Props: map[string]*customItemContentSchema{}}
		}
		mergeCustomItemContentSchema(schema, contentMap)
		sourceCount++
	}
	if schema == nil || sourceCount == 0 {
		return 0, fmt.Errorf("no readable custom item schema found for item_type %q", itemType)
	}
	if err := validateCustomItemValue(schema, content, "content"); err != nil {
		return sourceCount, err
	}
	if requireComplete {
		missing := missingCustomItemContentKeys(schema, content)
		if len(missing) > 0 {
			return sourceCount, fmt.Errorf("content missing schema keys: %s", strings.Join(missing, ", "))
		}
	}
	return sourceCount, nil
}

func mergeCustomItemContentSchema(schema *customItemContentSchema, value any) {
	kind := customItemJSONKind(value)
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
			schema.Props = map[string]*customItemContentSchema{}
		}
		for key, child := range typed {
			childSchema := schema.Props[key]
			if childSchema == nil {
				childSchema = &customItemContentSchema{}
				schema.Props[key] = childSchema
			}
			mergeCustomItemContentSchema(childSchema, child)
		}
	case []any:
		if schema.Elem == nil {
			schema.Elem = &customItemContentSchema{}
		}
		for _, child := range typed {
			mergeCustomItemContentSchema(schema.Elem, child)
		}
	}
}

func validateCustomItemValue(schema *customItemContentSchema, value any, path string) error {
	kind := customItemJSONKind(value)
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
			if err := validateCustomItemValue(childSchema, child, path+"."+key); err != nil {
				return err
			}
		}
	case []any:
		if schema.Elem == nil || schema.Elem.Kind == "" {
			return nil
		}
		for i, child := range typed {
			if err := validateCustomItemValue(schema.Elem, child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func missingCustomItemContentKeys(schema *customItemContentSchema, content map[string]any) []string {
	missing := []string{}
	for key := range schema.Props {
		if _, ok := content[key]; !ok {
			missing = append(missing, "content."+key)
		}
	}
	sort.Strings(missing)
	return missing
}

func customItemJSONKind(value any) string {
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

func shapeCustomItemWriteResponse(item intervals.CustomItem, operation string, endpoint string, itemID string, itemType string, fieldsUpdated []string, schemaSourceCount int) customItemWriteResponse {
	readShape := shapeGetCustomItemByIDResponse(item, itemID)
	if itemType == "" {
		itemType = readShape.Meta.ItemType
	}
	return customItemWriteResponse{CustomItem: readShape.CustomItem, Meta: customItemWriteMeta{Operation: operation, SourceEndpoint: endpoint, ItemID: readShape.Meta.ItemID, ItemType: itemType, FieldsUpdated: fieldsUpdated, ContentPreserved: readShape.Meta.ContentPreserved, SchemaDocumentation: customItemSchemaDocumentation, SchemaSourceCount: schemaSourceCount, DefaultPayloadScope: "full upstream custom item with content preserved verbatim; same custom_item read shape as get_custom_item_by_id"}}
}
