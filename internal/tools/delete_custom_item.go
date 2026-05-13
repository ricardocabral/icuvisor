package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	deleteCustomItemName                    = "delete_custom_item"
	deleteCustomItemDescription             = "Delete one custom item definition by item_id. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteCustomItemArgumentsMessage = "invalid delete_custom_item arguments; provide item_id only"
	deleteCustomItemMessage                 = "could not delete custom item; check intervals.icu credentials, athlete ID, item ID, and delete-mode configuration"
	deleteCustomItemEndpoint                = "/athlete/{id}/custom-item/{itemId}"
)

// CustomItemDeleterClient retrieves and deletes custom items for tools.
type CustomItemDeleterClient interface {
	GetCustomItem(context.Context, string) (intervals.CustomItem, error)
	DeleteCustomItem(context.Context, string) error
}

type deleteCustomItemRequest struct {
	ItemID string `json:"item_id"`
}

func newDeleteCustomItemTool(client CustomItemDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: deleteCustomItemName, Description: deleteCustomItemDescription, InputSchema: deleteCustomItemInputSchema(), OutputSchema: deleteCustomItemOutputSchema(), Requirement: RequirementDelete, Handler: deleteCustomItemHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func deleteCustomItemHandler(client CustomItemDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteCustomItemRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteCustomItemArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(deleteCustomItemMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(deleteCustomItemMessage, errors.New("missing custom item deleter client"))
		}
		item, err := client.GetCustomItem(ctx, args.ItemID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteCustomItemMessage, err)
		}
		before, err := customItemDeleteEcho(item, args.ItemID)
		if err != nil {
			return Result{}, fmt.Errorf("shaping delete_custom_item before echo: %w", err)
		}
		if err := client.DeleteCustomItem(ctx, args.ItemID); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteCustomItemMessage, err)
		}
		payload := newDeleteResourceResponse(args.ItemID, "custom_item", deleteCustomItemEndpoint, before)
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteCustomItemName, unitSystem)
	}
}

func decodeDeleteCustomItemRequest(raw json.RawMessage) (deleteCustomItemRequest, error) {
	var args deleteCustomItemRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.ItemID = strings.TrimSpace(args.ItemID)
	if args.ItemID == "" {
		return args, errors.New("item_id is required")
	}
	return args, nil
}

func deleteCustomItemInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"item_id"}, "properties": map[string]any{
		"item_id": map[string]any{"type": "string", "description": "Required opaque upstream custom item ID to delete. This destructive operation has no confirm argument; the tool is registered only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func deleteCustomItemOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete confirmation with deleted_id, status, and _meta.deleted containing a terse echo of the custom item before deletion."}
}
