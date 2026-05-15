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
	deleteSportSettingsName                    = "delete_sport_settings"
	deleteSportSettingsDescription             = "Delete one sport-settings definition by sport_settings_id. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteSportSettingsArgumentsMessage = "invalid delete_sport_settings arguments; provide sport_settings_id only"
	deleteSportSettingsMessage                 = "could not delete sport settings; check intervals.icu credentials, athlete ID, sport-settings ID, and delete-mode configuration"
	deleteSportSettingsEndpoint                = "/athlete/{id}/sport-settings/{sportSettingsId}"
)

// SportSettingsDeleterClient deletes sport settings for tools.
type SportSettingsDeleterClient interface {
	DeleteSportSettings(context.Context, string) error
}

type deleteSportSettingsRequest struct {
	SportSettingsID string `json:"sport_settings_id"`
}

func newDeleteSportSettingsTool(client SportSettingsDeleterClient, profileClient ProfileClient, version string, _ string, debugMetadata bool) Tool {
	return fullTool(Tool{Name: deleteSportSettingsName, Description: deleteSportSettingsDescription, InputSchema: deleteSportSettingsInputSchema(), OutputSchema: deleteSportSettingsOutputSchema(), Requirement: RequirementDelete, Handler: deleteSportSettingsHandler(client, profileClient, version, debugMetadata)})
}

func deleteSportSettingsHandler(client SportSettingsDeleterClient, profileClient ProfileClient, version string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteSportSettingsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteSportSettingsArgumentsMessage, err)
		}
		if profileClient == nil || client == nil {
			return Result{}, NewUserError(deleteSportSettingsMessage, errors.New("missing sport settings deleter client"))
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteSportSettingsMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		setting, ok := sportSettingByID(profile.SportSettings, args.SportSettingsID)
		if !ok {
			return Result{}, NewUserError(deleteSportSettingsMessage, fmt.Errorf("sport-settings ID %q not found in athlete profile", args.SportSettingsID))
		}
		before := sportSettingsDeleteEcho(setting, args.SportSettingsID)
		if err := client.DeleteSportSettings(ctx, args.SportSettingsID); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteSportSettingsMessage, err)
		}
		payload := newDeleteResourceResponse(args.SportSettingsID, "sport_settings", deleteSportSettingsEndpoint, before)
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteSportSettingsName, unitSystem)
	}
}

func decodeDeleteSportSettingsRequest(raw json.RawMessage) (deleteSportSettingsRequest, error) {
	var args deleteSportSettingsRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[deleteSportSettingsRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.SportSettingsID = strings.TrimSpace(args.SportSettingsID)
	if args.SportSettingsID == "" {
		return args, errors.New("sport_settings_id is required")
	}
	return args, nil
}

func sportSettingByID(settings []intervals.SportSettings, id string) (intervals.SportSettings, bool) {
	trimmed := strings.TrimSpace(id)
	for _, setting := range settings {
		if fmt.Sprint(setting.ID) == trimmed {
			return setting, true
		}
	}
	return intervals.SportSettings{}, false
}

func deleteSportSettingsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"sport_settings_id"}, "properties": map[string]any{
		"sport_settings_id": map[string]any{"type": "string", "description": "Required opaque upstream sport-settings ID to delete. This destructive operation has no confirm argument; the tool is registered only when ICUVISOR_DELETE_MODE=full."},
	}}
}

func deleteSportSettingsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete confirmation with deleted_id, status, and _meta.deleted containing a terse echo of the sport settings before deletion."}
}
