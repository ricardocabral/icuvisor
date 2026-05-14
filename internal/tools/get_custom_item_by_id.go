package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getCustomItemByIDName                    = "get_custom_item_by_id"
	getCustomItemByIDDescription             = "Fetch one custom item by item_id and preserve its full content payload. For per-item_type content schema guidance, see icuvisor://custom-item-schemas."
	invalidGetCustomItemByIDArgumentsMessage = "invalid get_custom_item_by_id arguments; provide item_id"
	fetchCustomItemByIDMessage               = "could not fetch custom item; check intervals.icu credentials, athlete ID, and item ID"
	customItemByIDEndpoint                   = "/athlete/{id}/custom-item/{itemId}"
)

type getCustomItemByIDRequest struct {
	ItemID string `json:"item_id"`
}

type getCustomItemByIDResponse struct {
	CustomItem map[string]any        `json:"custom_item"`
	Meta       getCustomItemByIDMeta `json:"_meta"`
}

type getCustomItemByIDMeta struct {
	SourceEndpoint      string `json:"source_endpoint"`
	ItemID              string `json:"item_id"`
	ItemType            string `json:"item_type,omitempty"`
	ContentPreserved    bool   `json:"content_preserved"`
	SchemaDocumentation string `json:"schema_documentation"`
	DefaultPayloadScope string `json:"default_payload_scope"`
}

func newGetCustomItemByIDTool(client CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: getCustomItemByIDName, Description: getCustomItemByIDDescription, InputSchema: getCustomItemByIDInputSchema(), OutputSchema: getCustomItemByIDOutputSchema(), Handler: getCustomItemByIDHandler(client, profileClient, version, timezoneFallback, debugMetadata)})
}

func getCustomItemByIDHandler(client CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetCustomItemByIDRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetCustomItemByIDArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchCustomItemByIDMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchCustomItemByIDMessage, errors.New("missing custom items client"))
		}
		item, err := client.GetCustomItem(ctx, args.ItemID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchCustomItemByIDMessage, err)
		}
		payload := shapeGetCustomItemByIDResponse(item, args.ItemID)
		return encodeShaped(payload, true, nil, version, debugMetadata, getCustomItemByIDName, unitSystem)
	}
}

func decodeGetCustomItemByIDRequest(raw json.RawMessage) (getCustomItemByIDRequest, error) {
	var args getCustomItemByIDRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ItemID = strings.TrimSpace(args.ItemID)
	if args.ItemID == "" {
		return args, errors.New("item_id is required")
	}
	return args, nil
}

func shapeGetCustomItemByIDResponse(item intervals.CustomItem, requestedID string) getCustomItemByIDResponse {
	itemType := customItemType(item)
	detail := cloneJSONMap(item.Raw)
	if item.ID != "" {
		detail["id"] = item.ID
	} else {
		detail["id"] = requestedID
	}
	if itemType != "" {
		detail["item_type"] = itemType
	}
	if item.Content != nil {
		detail["content"] = item.Content
	}
	return getCustomItemByIDResponse{CustomItem: detail, Meta: getCustomItemByIDMeta{SourceEndpoint: customItemByIDEndpoint, ItemID: requestedID, ItemType: itemType, ContentPreserved: item.Content != nil || detail["content"] != nil, SchemaDocumentation: "icuvisor://custom-item-schemas", DefaultPayloadScope: "full upstream custom item with content preserved verbatim"}}
}

func getCustomItemByIDInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"item_id"}, "properties": map[string]any{
		"item_id": map[string]any{"type": "string", "description": "Required intervals.icu custom item ID."},
	}}
}

func getCustomItemByIDOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Full custom item payload with content preserved verbatim and item_type schema documentation noted in _meta."}
}
