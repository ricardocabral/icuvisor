package tools

import (
	"encoding/json"
	"errors"

	"github.com/ricardocabral/icuvisor/internal/customitemschemas"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const customItemSchemaDocumentation = "icuvisor://custom-item-schemas"

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
	return customitemschemas.ValidateContentAgainstReadSchema(items, itemType, content, requireComplete)
}

func shapeCustomItemWriteResponse(item intervals.CustomItem, operation string, endpoint string, itemID string, itemType string, fieldsUpdated []string, schemaSourceCount int) customItemWriteResponse {
	readShape := shapeGetCustomItemByIDResponse(item, itemID)
	if itemType == "" {
		itemType = readShape.Meta.ItemType
	}
	return customItemWriteResponse{CustomItem: readShape.CustomItem, Meta: customItemWriteMeta{Operation: operation, SourceEndpoint: endpoint, ItemID: readShape.Meta.ItemID, ItemType: itemType, FieldsUpdated: fieldsUpdated, ContentPreserved: readShape.Meta.ContentPreserved, SchemaDocumentation: customItemSchemaDocumentation, SchemaSourceCount: schemaSourceCount, DefaultPayloadScope: "full upstream custom item with content preserved verbatim; same custom_item read shape as get_custom_item_by_id"}}
}
