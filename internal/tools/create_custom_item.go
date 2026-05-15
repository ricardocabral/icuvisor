package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	createCustomItemName                    = "create_custom_item"
	createCustomItemDescription             = "Create one custom item definition. Content is validated against readable samples; see icuvisor://custom-item-schemas for item_type content guidance."
	invalidCreateCustomItemArgumentsMessage = "invalid create_custom_item arguments; provide item_type, name, content matching a readable custom-item schema, and optional visibility/description/index/hide_script"
	createCustomItemMessage                 = "could not create custom item; check intervals.icu credentials, athlete ID, writable custom-item fields, and available schema samples"
)

// CustomItemCreatorClient creates custom items for tools.
type CustomItemCreatorClient interface {
	CreateCustomItem(context.Context, intervals.WriteCustomItemParams) (intervals.CustomItem, error)
}

type createCustomItemRequest struct {
	ItemType    string          `json:"item_type"`
	Name        string          `json:"name"`
	Visibility  *string         `json:"visibility,omitempty"`
	Description *string         `json:"description,omitempty"`
	Image       *string         `json:"image,omitempty"`
	Index       *int            `json:"index,omitempty"`
	HideScript  *bool           `json:"hide_script,omitempty"`
	Content     json.RawMessage `json:"content"`
}

func newCreateCustomItemTool(client CustomItemCreatorClient, readClient CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: createCustomItemName, Description: createCustomItemDescription, InputSchema: createCustomItemInputSchema(), OutputSchema: createCustomItemOutputSchema(), Requirement: RequirementWrite, Handler: createCustomItemHandler(client, readClient, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func createCustomItemHandler(client CustomItemCreatorClient, readClient CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeCreateCustomItemRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidCreateCustomItemArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(createCustomItemMessage, err)
		}
		if client == nil || readClient == nil {
			return Result{}, NewUserError(createCustomItemMessage, errors.New("missing custom item creator or schema client"))
		}
		params, schemaSourceCount, err := createCustomItemParams(ctx, readClient, args)
		if err != nil {
			return Result{}, NewUserError(invalidCreateCustomItemArgumentsMessage, err)
		}
		item, err := client.CreateCustomItem(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(createCustomItemMessage, err)
		}
		payload := shapeCustomItemWriteResponse(item, "create", customItemsEndpoint, item.ID, args.ItemType, nil, schemaSourceCount)
		return encodeShaped(payload, true, nil, version, debugMetadata, createCustomItemName, unitSystem, shapeCfg)
	}
}

func decodeCreateCustomItemRequest(raw json.RawMessage) (createCustomItemRequest, error) {
	var args createCustomItemRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ItemType = strings.TrimSpace(args.ItemType)
	args.Name = strings.TrimSpace(args.Name)
	trimOptionalString(args.Visibility)
	trimOptionalString(args.Image)
	if args.ItemType == "" {
		return args, errors.New("item_type is required")
	}
	if args.Name == "" {
		return args, errors.New("name is required")
	}
	if len(args.Content) == 0 || strings.TrimSpace(string(args.Content)) == "null" {
		return args, errors.New("content is required")
	}
	if args.Visibility != nil && *args.Visibility == "" {
		return args, errors.New("visibility cannot be empty when supplied")
	}
	if args.Image != nil && *args.Image == "" {
		return args, errors.New("image cannot be empty when supplied")
	}
	return args, nil
}

func createCustomItemParams(ctx context.Context, readClient CustomItemsClient, args createCustomItemRequest) (intervals.WriteCustomItemParams, int, error) {
	content, err := customItemContentFromRaw(args.Content)
	if err != nil {
		return intervals.WriteCustomItemParams{}, 0, err
	}
	items, err := customItemSchemaSamples(ctx, readClient, args.ItemType)
	if err != nil {
		return intervals.WriteCustomItemParams{}, 0, err
	}
	schemaSourceCount, err := validateCustomItemContentAgainstReadSchema(items, args.ItemType, content, true)
	if err != nil {
		return intervals.WriteCustomItemParams{}, schemaSourceCount, err
	}
	return intervals.WriteCustomItemParams{ItemType: args.ItemType, Name: args.Name, NameSet: true, Visibility: args.Visibility, VisibilitySet: args.Visibility != nil, Description: args.Description, DescriptionSet: args.Description != nil, Image: args.Image, ImageSet: args.Image != nil, Index: args.Index, IndexSet: args.Index != nil, HideScript: args.HideScript, HideScriptSet: args.HideScript != nil, Content: content, ContentSet: true}, schemaSourceCount, nil
}

func customItemSchemaSamples(ctx context.Context, readClient CustomItemsClient, itemType string) ([]intervals.CustomItem, error) {
	items, err := readClient.ListCustomItems(ctx)
	if err != nil {
		return nil, err
	}
	samples := make([]intervals.CustomItem, 0, len(items))
	for _, item := range items {
		if customItemType(item) != itemType {
			continue
		}
		if _, ok := item.Content.(map[string]any); ok {
			samples = append(samples, item)
			continue
		}
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		detail, err := readClient.GetCustomItem(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		samples = append(samples, detail)
	}
	return samples, nil
}

func trimOptionalString(value *string) {
	if value != nil {
		*value = strings.TrimSpace(*value)
	}
}

func createCustomItemInputSchema() map[string]any {
	examples := createCustomItemInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"item_type", "name", "content"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"item_type":   map[string]any{"type": "string", "description": "Required upstream custom-item type. The value must match an existing readable schema sample, for example FITNESS_CHART, INPUT_FIELD, ACTIVITY_FIELD, ACTIVITY_STREAM, ACTIVITY_PANEL, ACTIVITY_HISTOGRAM, ACTIVITY_MAP, ACTIVITY_HEATMAP, TRACE_CHART, FITNESS_TABLE, or ZONES."},
		"name":        map[string]any{"type": "string", "description": "Required custom item name. Surrounding whitespace is trimmed."},
		"visibility":  map[string]any{"type": "string", "description": "Optional upstream visibility value. Omit to use intervals.icu defaults."},
		"description": map[string]any{"type": "string", "description": "Optional custom item description. Preserved verbatim when supplied."},
		"image":       map[string]any{"type": "string", "description": "Optional upstream image identifier or URL when supported by the custom-item type."},
		"index":       map[string]any{"type": "integer", "description": "Optional display/order index for item types that support ordering."},
		"hide_script": map[string]any{"type": "boolean", "description": "Optional upstream hide_script flag for script/formula based custom items."},
		"content":     map[string]any{"type": "object", "description": "Required item_type-specific content object. Validated against existing readable custom items before upload; see icuvisor://custom-item-schemas for schema guidance."},
	}}
}

func createCustomItemInputExamples() []map[string]any {
	return []map[string]any{
		{
			"item_type": "FITNESS_CHART",
			"name":      "Training load trend",
			"content": map[string]any{
				"series": []any{map[string]any{"field": "ctl", "color": "blue"}},
				"layout": map[string]any{"height": 240},
			},
		},
		{
			"item_type":   "INPUT_FIELD",
			"name":        "Travel fatigue note",
			"visibility":  "PRIVATE",
			"description": "Short coach-facing note captured on travel weeks.",
			"index":       20,
			"hide_script": false,
			"content": map[string]any{
				"field": map[string]any{"key": "travel_fatigue", "type": "text", "label": "Travel fatigue"},
			},
		},
	}
}

func createCustomItemOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Create confirmation containing the same full custom_item read shape as get_custom_item_by_id, with content preserved verbatim and schema-validation metadata."}
}
