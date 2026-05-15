package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	deleteGearName                    = "delete_gear"
	deleteGearDescription             = "Delete one gear item by gear_id. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteGearArgumentsMessage = "invalid delete_gear arguments; provide gear_id only"
	deleteGearMessage                 = "could not delete gear; check intervals.icu credentials, athlete ID, gear ID, and delete-mode configuration"
	deleteGearEndpoint                = "/athlete/{id}/gear/{gearId}"
)

// GearDeleterClient retrieves and deletes gear for tools.
type GearDeleterClient interface {
	GetGear(context.Context, string) (intervals.Gear, error)
	DeleteGear(context.Context, string) error
}

type deleteGearRequest struct {
	GearID string `json:"gear_id"`
}

func newDeleteGearTool(client GearDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: deleteGearName, Description: deleteGearDescription, InputSchema: deleteGearInputSchema(), OutputSchema: deleteGearOutputSchema(), Requirement: RequirementDelete, Handler: deleteGearHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func deleteGearHandler(client GearDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteGearRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteGearArgumentsMessage, err)
		}
		unitSystem, _, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(deleteGearMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(deleteGearMessage, errors.New("missing gear deleter client"))
		}
		gear, err := client.GetGear(ctx, args.GearID)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteGearMessage, err)
		}
		before := gearDeleteEcho(gear, args.GearID)
		if err := client.DeleteGear(ctx, args.GearID); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteGearMessage, err)
		}
		payload := newDeleteResourceResponse(args.GearID, "gear", deleteGearEndpoint, before)
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteGearName, unitSystem, shapeCfg)
	}
}

func decodeDeleteGearRequest(raw json.RawMessage) (deleteGearRequest, error) {
	var args deleteGearRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.GearID = strings.TrimSpace(args.GearID)
	if args.GearID == "" {
		return args, errors.New("gear_id is required")
	}
	return args, nil
}

func deleteGearInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"gear_id"}, "properties": map[string]any{
		"gear_id": map[string]any{"type": "string", "description": "Required opaque upstream gear ID to delete. This destructive operation has no confirm argument; the tool is registered only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func deleteGearOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete confirmation with deleted_id, status, and _meta.deleted containing a terse echo of the gear item before deletion."}
}
