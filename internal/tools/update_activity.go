package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	updateActivityName                    = "update_activity"
	updateActivityDescription             = "Update one completed non-Strava activity by activity_id: rename it, replace its free-text description, and/or set athlete-logged carbohydrate intake in whole grams. Sparse update: omit a field to leave it unchanged; pass an explicit empty string for description to clear it. carbs_ingested_g accepts 0-2147483647, where zero is a logged zero; clearing is not supported, and read-only carbs_used_g is a distinct upstream estimate. Intervals.icu does not update Strava activities through this endpoint. Activity descriptions are prose metadata only, not planned-workout structure. This non-destructive metadata edit does not alter recorded streams, intervals, or analyzed metrics. Use set_activity_intervals (delete-mode tool) to write a structured workout_doc as the activity's interval set."
	invalidUpdateActivityArgumentsMessage = "invalid update_activity arguments; provide activity_id plus at least one of name, description, or carbs_ingested_g"
	updateActivityMessage                 = "could not update activity; check intervals.icu credentials, activity ID, and writable activity fields"
	maxActivityCarbsIngestedG             = 1<<31 - 1
)

// ActivityUpdaterClient sparsely updates supported activity metadata.
type ActivityUpdaterClient interface {
	UpdateActivity(context.Context, intervals.UpdateActivityParams) (intervals.Activity, error)
}

type updateActivityRequest struct {
	ActivityID     string  `json:"activity_id"`
	Name           string  `json:"name,omitempty"`
	Description    *string `json:"description,omitempty"`
	CarbsIngestedG *int    `json:"carbs_ingested_g,omitempty"`
	IncludeFull    bool    `json:"include_full,omitempty"`

	nameProvided           bool
	descriptionProvided    bool
	carbsIngestedGProvided bool
}

type updateActivityResponse struct {
	ActivityID    string             `json:"activity_id"`
	Status        string             `json:"status"`
	FieldsUpdated []string           `json:"fields_updated"`
	Full          map[string]any     `json:"full,omitempty"`
	Meta          updateActivityMeta `json:"_meta"`
}

type updateActivityMeta struct {
	AthleteID      string `json:"athleteId,omitempty"`
	AppendOnly     bool   `json:"append_only"`
	Destructive    bool   `json:"destructive"`
	SourceEndpoint string `json:"source_endpoint"`
	IncludeFull    bool   `json:"include_full"`
}

func newUpdateActivityTool(client ActivityUpdaterClient, profileClient ProfileClient, version string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: updateActivityName, Description: updateActivityDescription, InputSchema: updateActivityInputSchema(), OutputSchema: updateActivityOutputSchema(), Requirement: RequirementWrite, Handler: updateActivityHandler(client, profileClient, version, debugMetadata, shapeCfg)})
}

func updateActivityHandler(client ActivityUpdaterClient, profileClient ProfileClient, version string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeUpdateActivityRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidUpdateActivityArgumentsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(updateActivityMessage, errors.New("missing activity updater client"))
		}
		params := intervals.UpdateActivityParams{ActivityID: args.ActivityID}
		if args.nameProvided {
			params.Name = args.Name
			params.NameSet = true
		}
		if args.descriptionProvided {
			if args.Description != nil {
				params.Description = *args.Description
			}
			params.DescriptionSet = true
		}
		if args.carbsIngestedGProvided {
			params.CarbsIngested = *args.CarbsIngestedG
			params.CarbsIngestedSet = true
		}
		activity, err := client.UpdateActivity(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(updateActivityMessage, err)
		}
		athleteID := ""
		if profileClient != nil {
			profile, profileErr := profileClient.GetAthleteProfile(ctx)
			if profileErr != nil {
				if errors.Is(profileErr, context.Canceled) || errors.Is(profileErr, context.DeadlineExceeded) {
					return Result{}, profileErr
				}
			} else {
				athleteID = config.NormalizeAthleteIDForDisplay(profile.ID)
			}
		}
		payload := updateActivityResponse{
			ActivityID:    args.ActivityID,
			Status:        "updated",
			FieldsUpdated: updateActivityFieldsUpdated(args),
			Meta: updateActivityMeta{
				AthleteID:      athleteID,
				AppendOnly:     false,
				Destructive:    false,
				SourceEndpoint: "/activity/{activityId}",
				IncludeFull:    args.IncludeFull,
			},
		}
		if args.IncludeFull {
			payload.Full = activity.Raw
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, updateActivityName, "", shapeCfg)
	}
}

func decodeUpdateActivityRequest(raw json.RawMessage) (updateActivityRequest, error) {
	if strings.TrimSpace(string(raw)) == "" {
		return updateActivityRequest{}, errors.New("arguments must be a JSON object")
	}
	fields, err := rawObjectFields(raw)
	if err != nil {
		return updateActivityRequest{}, err
	}
	decoded, err := DecodeStrict[updateActivityRequest](raw)
	if err != nil {
		return updateActivityRequest{}, err
	}
	decoded.ActivityID = strings.TrimSpace(decoded.ActivityID)
	if decoded.ActivityID == "" {
		return updateActivityRequest{}, errors.New("activity_id is required")
	}
	decoded.nameProvided = fields["name"]
	decoded.descriptionProvided = fields["description"]
	decoded.carbsIngestedGProvided = fields["carbs_ingested_g"]
	if decoded.nameProvided {
		decoded.Name = strings.TrimSpace(decoded.Name)
		if decoded.Name == "" {
			return updateActivityRequest{}, errors.New("name cannot be empty when supplied; omit to leave unchanged")
		}
	}
	if decoded.carbsIngestedGProvided {
		if decoded.CarbsIngestedG == nil {
			return updateActivityRequest{}, errors.New("carbs_ingested_g cannot be null; omit to leave unchanged")
		}
		if *decoded.CarbsIngestedG < 0 || *decoded.CarbsIngestedG > maxActivityCarbsIngestedG {
			return updateActivityRequest{}, errors.New("carbs_ingested_g must be an integer from 0 through 2147483647")
		}
	}
	if !decoded.nameProvided && !decoded.descriptionProvided && !decoded.carbsIngestedGProvided {
		return updateActivityRequest{}, errors.New("at least one of name, description, or carbs_ingested_g must be supplied")
	}
	return decoded, nil
}

func updateActivityFieldsUpdated(args updateActivityRequest) []string {
	fields := []string{}
	if args.nameProvided {
		fields = append(fields, "name")
	}
	if args.descriptionProvided {
		fields = append(fields, "description")
	}
	if args.carbsIngestedGProvided {
		fields = append(fields, "carbs_ingested_g")
	}
	sort.Strings(fields)
	return fields
}

func updateActivityInputSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"activity_id"},
		"properties": map[string]any{
			"activity_id":      map[string]any{"type": "string", "description": "Required intervals.icu activity ID. Surrounding whitespace is trimmed; the i-prefix is preserved verbatim."},
			"name":             map[string]any{"type": "string", "description": "Optional replacement activity title. Omit to leave unchanged. Empty strings are rejected to avoid accidentally blanking the title; intervals.icu's UI also rejects blank titles."},
			"description":      map[string]any{"type": "string", "description": "Optional replacement free-text activity description; this is not append-only. Omit to leave unchanged; pass an explicit empty string to clear the description. Prose metadata only — to write structured intervals, use set_activity_intervals with a workout_doc."},
			"carbs_ingested_g": map[string]any{"type": "integer", "minimum": 0, "maximum": maxActivityCarbsIngestedG, "description": "Optional athlete-logged carbohydrate consumed during this activity, in whole grams (0-2147483647). Zero is a logged zero; omit to leave unchanged. Clearing is not supported. Distinct from read-only carbs_used_g. Intervals.icu does not update Strava activities through this endpoint."},
			"include_full":     map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream updated-activity payload under full; default returns a terse update confirmation."},
		},
	}
}

func updateActivityOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Non-destructive activity metadata update confirmation with activity_id, status, fields_updated, normalized athlete_id, and source_endpoint metadata. Supports sparse name, free-text description, and athlete-logged carbs_ingested_g updates; carbs_used_g remains read-only. Does not alter recorded streams or interval analysis."}
}
