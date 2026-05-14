package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	updateCustomItemName                    = "update_custom_item"
	updateCustomItemDescription             = "Update one custom item definition by item_id with sparse fields only. Content patches are validated against the existing item; see icuvisor://custom-item-schemas."
	invalidUpdateCustomItemArgumentsMessage = "invalid update_custom_item arguments; provide item_id plus at least one sparse field matching the existing custom-item schema"
	updateCustomItemMessage                 = "could not update custom item; check intervals.icu credentials, athlete ID, item ID, writable custom-item fields, and schema"
)

// CustomItemUpdaterClient updates custom items for tools.
type CustomItemUpdaterClient interface {
	UpdateCustomItem(context.Context, intervals.WriteCustomItemParams) (intervals.CustomItem, error)
}

type updateCustomItemRequest struct {
	ItemID      string          `json:"item_id"`
	Name        string          `json:"name,omitempty"`
	Visibility  *string         `json:"visibility,omitempty"`
	Description *string         `json:"description,omitempty"`
	Image       *string         `json:"image,omitempty"`
	Index       *int            `json:"index,omitempty"`
	HideScript  *bool           `json:"hide_script,omitempty"`
	Content     json.RawMessage `json:"content,omitempty"`

	nameProvided        bool
	visibilityProvided  bool
	descriptionProvided bool
	imageProvided       bool
	indexProvided       bool
	hideScriptProvided  bool
	contentProvided     bool
}

func newUpdateCustomItemTool(client CustomItemUpdaterClient, readClient CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: updateCustomItemName, Description: updateCustomItemDescription, InputSchema: updateCustomItemInputSchema(), OutputSchema: updateCustomItemOutputSchema(), Requirement: RequirementWrite, Handler: updateCustomItemHandler(client, readClient, profileClient, version, timezoneFallback, debugMetadata)})
}

func updateCustomItemHandler(client CustomItemUpdaterClient, readClient CustomItemsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeUpdateCustomItemRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateCustomItemArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(updateCustomItemMessage, err)
		}
		if client == nil || readClient == nil {
			return Result{}, NewUserError(updateCustomItemMessage, errors.New("missing custom item updater or schema client"))
		}
		params, itemType, schemaSourceCount, err := updateCustomItemParams(ctx, readClient, args)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateCustomItemArgumentsMessage, err)
		}
		item, err := client.UpdateCustomItem(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(updateCustomItemMessage, err)
		}
		payload := shapeCustomItemWriteResponse(item, "update", customItemByIDEndpoint, args.ItemID, itemType, updateCustomItemFieldsUpdated(args), schemaSourceCount)
		return encodeShaped(payload, true, nil, version, debugMetadata, updateCustomItemName, unitSystem)
	}
}

func decodeUpdateCustomItemRequest(raw json.RawMessage) (updateCustomItemRequest, error) {
	fields, err := rawObjectFields(raw)
	if err != nil {
		return updateCustomItemRequest{}, err
	}
	var args updateCustomItemRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ItemID = strings.TrimSpace(args.ItemID)
	args.Name = strings.TrimSpace(args.Name)
	trimOptionalString(args.Visibility)
	trimOptionalString(args.Image)
	args.nameProvided = fields["name"]
	args.visibilityProvided = fields["visibility"]
	args.descriptionProvided = fields["description"]
	args.imageProvided = fields["image"]
	args.indexProvided = fields["index"]
	args.hideScriptProvided = fields["hide_script"]
	args.contentProvided = fields["content"]
	if args.ItemID == "" {
		return args, errors.New("item_id is required")
	}
	if args.nameProvided && args.Name == "" {
		return args, errors.New("name cannot be empty when supplied")
	}
	if args.visibilityProvided && (args.Visibility == nil || *args.Visibility == "") {
		return args, errors.New("visibility cannot be null or empty when supplied")
	}
	if args.imageProvided && (args.Image == nil || *args.Image == "") {
		return args, errors.New("image cannot be null or empty when supplied")
	}
	if args.contentProvided && strings.TrimSpace(string(args.Content)) == "null" {
		return args, errors.New("content cannot be null when supplied")
	}
	if len(updateCustomItemFieldsUpdated(args)) == 0 {
		return args, errors.New("at least one sparse field is required")
	}
	return args, nil
}

func updateCustomItemParams(ctx context.Context, readClient CustomItemsClient, args updateCustomItemRequest) (intervals.WriteCustomItemParams, string, int, error) {
	params := intervals.WriteCustomItemParams{ItemID: args.ItemID, Name: args.Name, NameSet: args.nameProvided, Visibility: args.Visibility, VisibilitySet: args.visibilityProvided, Description: args.Description, DescriptionSet: args.descriptionProvided, Image: args.Image, ImageSet: args.imageProvided, Index: args.Index, IndexSet: args.indexProvided, HideScript: args.HideScript, HideScriptSet: args.hideScriptProvided}
	if !args.contentProvided {
		return params, "", 0, nil
	}
	patch, err := customItemContentFromRaw(args.Content)
	if err != nil {
		return intervals.WriteCustomItemParams{}, "", 0, err
	}
	existing, err := readClient.GetCustomItem(ctx, args.ItemID)
	if err != nil {
		return intervals.WriteCustomItemParams{}, "", 0, err
	}
	itemType := customItemType(existing)
	schemaSourceCount, err := validateCustomItemContentAgainstReadSchema([]intervals.CustomItem{existing}, itemType, patch, false)
	if err != nil {
		return intervals.WriteCustomItemParams{}, itemType, schemaSourceCount, err
	}
	existingContent, ok := existing.Content.(map[string]any)
	if !ok {
		return intervals.WriteCustomItemParams{}, itemType, schemaSourceCount, errors.New("existing content schema is not an object")
	}
	params.Content = mergeCustomItemContentPatch(existingContent, patch)
	params.ContentSet = true
	return params, itemType, schemaSourceCount, nil
}

func mergeCustomItemContentPatch(base map[string]any, patch map[string]any) map[string]any {
	out := cloneJSONMap(base)
	for key, value := range patch {
		if patchMap, ok := value.(map[string]any); ok {
			if baseMap, ok := out[key].(map[string]any); ok {
				out[key] = mergeCustomItemContentPatch(baseMap, patchMap)
				continue
			}
		}
		out[key] = value
	}
	return out
}

func updateCustomItemFieldsUpdated(args updateCustomItemRequest) []string {
	fields := []string{}
	if args.nameProvided {
		fields = append(fields, "name")
	}
	if args.visibilityProvided {
		fields = append(fields, "visibility")
	}
	if args.descriptionProvided {
		fields = append(fields, "description")
	}
	if args.imageProvided {
		fields = append(fields, "image")
	}
	if args.indexProvided {
		fields = append(fields, "index")
	}
	if args.hideScriptProvided {
		fields = append(fields, "hide_script")
	}
	if args.contentProvided {
		fields = append(fields, "content")
	}
	sort.Strings(fields)
	return fields
}

func updateCustomItemInputSchema() map[string]any {
	examples := updateCustomItemInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"item_id"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"item_id":     map[string]any{"type": "string", "description": "Required intervals.icu custom item ID to update. Surrounding whitespace is trimmed."},
		"name":        map[string]any{"type": "string", "description": "Optional replacement custom item name. Omit to leave unchanged."},
		"visibility":  map[string]any{"type": "string", "description": "Optional replacement upstream visibility value. Omit to leave unchanged."},
		"description": map[string]any{"type": "string", "description": "Optional replacement custom item description. Omit to leave unchanged; empty strings are preserved when intentionally supplied."},
		"image":       map[string]any{"type": "string", "description": "Optional replacement upstream image identifier or URL when supported by the custom-item type. Omit to leave unchanged."},
		"index":       map[string]any{"type": "integer", "description": "Optional replacement display/order index. Omit to leave unchanged."},
		"hide_script": map[string]any{"type": "boolean", "description": "Optional replacement upstream hide_script flag. Omit to leave unchanged."},
		"content":     map[string]any{"type": "object", "description": "Optional sparse content patch. Validated against the existing item_type schema, then merged so omitted keys stay untouched; see icuvisor://custom-item-schemas."},
	}}
}

func updateCustomItemInputExamples() []map[string]any {
	return []map[string]any{
		{
			"item_id": "custom-item-example-3",
			"name":    "Training load trend - coach view",
		},
		{
			"item_id":     "custom-item-example-4",
			"visibility":  "PRIVATE",
			"description": "Updated description for the coach dashboard.",
			"content": map[string]any{
				"layout": map[string]any{"height": 300},
			},
		},
		{
			"item_id":     "custom-item-example-5",
			"index":       30,
			"hide_script": true,
			"content": map[string]any{
				"series": []any{map[string]any{"field": "atl", "color": "orange"}},
			},
		},
	}
}

func updateCustomItemOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Update confirmation containing the same full custom_item read shape as get_custom_item_by_id, with content preserved verbatim, fields_updated, and schema-validation metadata."}
}
