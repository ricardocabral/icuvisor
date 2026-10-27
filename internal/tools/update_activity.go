package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/config"
	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	updateActivityName                    = "update_activity"
	updateActivityDescription             = "Update one completed non-Strava activity by activity_id: rename it, replace its description/tags, set commute, feel (1-5), RPE (1-10), subtype, manual-only distance in metres, or athlete-logged carbohydrate intake in whole grams. Sparse update: omit a field to leave it unchanged; pass an explicit empty string for description to clear it. carbs_ingested_g accepts 0-2147483647, where zero is a logged zero; clearing is not supported, and read-only carbs_used_g is a distinct upstream estimate. Intervals.icu does not update Strava activities through this endpoint. Gear assignment is not supported by the verified activity-write contract; use get_gear_list only to resolve existing activity gear IDs. Activity descriptions are prose metadata only, not planned-workout structure. For WeightTraining actuals, use one Markdown '- ' line per exercise and distinguish performed sets, reps, and loads from the plan. This non-destructive metadata edit does not alter recorded streams or intervals; upstream may recalculate derived metrics. New metadata fields are checked against the write response or one bounded readback; a mismatch or missing evidence is reported explicitly. Time corrections are unsupported. Use set_activity_intervals (delete-mode tool) to write a structured workout_doc as the activity's interval set."
	invalidUpdateActivityArgumentsMessage = "invalid update_activity arguments; provide activity_id plus a supported writable metadata field; gear assignment is not supported"
	updateActivityMessage                 = "could not update activity; check intervals.icu credentials, activity ID, and writable activity fields"
	maxActivityCarbsIngestedG             = 1<<31 - 1
)

// ActivityUpdaterClient sparsely updates supported activity metadata.
type ActivityUpdaterClient interface {
	UpdateActivity(context.Context, intervals.UpdateActivityParams) (intervals.Activity, error)
}

type updateActivityRequest struct {
	ActivityID     string    `json:"activity_id"`
	Name           string    `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	CarbsIngestedG *int      `json:"carbs_ingested_g,omitempty"`
	Tags           *[]string `json:"tags,omitempty"`
	Commute        *bool     `json:"commute,omitempty"`
	Feel           *int      `json:"feel,omitempty"`
	RPE            *int      `json:"rpe,omitempty"`
	SubType        *string   `json:"sub_type,omitempty"`
	DistanceM      *float64  `json:"distance_m,omitempty"`
	IncludeFull    bool      `json:"include_full,omitempty"`

	nameProvided           bool
	descriptionProvided    bool
	carbsIngestedGProvided bool
}

type updateActivityResponse struct {
	ActivityID    string                         `json:"activity_id"`
	Status        string                         `json:"status"`
	FieldsUpdated []string                       `json:"fields_updated"`
	Full          map[string]any                 `json:"full,omitempty"`
	Requested     *updateActivityRequestedValues `json:"requested,omitempty"`
	Meta          updateActivityMeta             `json:"_meta"`
}

type updateActivityRequestedValues struct {
	Tags      *[]string `json:"tags,omitempty"`
	Commute   *bool     `json:"commute,omitempty"`
	Feel      *int      `json:"feel,omitempty"`
	RPE       *int      `json:"rpe,omitempty"`
	SubType   *string   `json:"sub_type,omitempty"`
	DistanceM *float64  `json:"distance_m,omitempty"`
}

type updateActivityMeta struct {
	AthleteID      string                                  `json:"athleteId,omitempty"`
	AppendOnly     bool                                    `json:"append_only"`
	Destructive    bool                                    `json:"destructive"`
	SourceEndpoint string                                  `json:"source_endpoint"`
	IncludeFull    bool                                    `json:"include_full"`
	Verification   *intervals.ActivityMetadataVerification `json:"verification,omitempty"`
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
		params := args.params()
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
		activityIdentityMatches := activity.ID != "" && activity.ID == args.ActivityID
		verification := activity.MetadataVerification
		if verification == nil || !activityIdentityMatches {
			verification = intervals.VerifyActivityMetadataUpdate(activity, params)
		}
		if verification != nil {
			payload.Meta.Verification = verification
			if verification.Status != "verified" {
				payload.Status = verification.Status
			}
			confirmed := make([]string, 0, len(payload.FieldsUpdated))
			for _, field := range payload.FieldsUpdated {
				switch field {
				case "tags", "commute", "feel", "rpe", "sub_type", "distance_m":
					for _, verified := range verification.Confirmed {
						if field == verified {
							confirmed = append(confirmed, field)
							break
						}
					}
				default:
					confirmed = append(confirmed, field)
				}
			}
			if !activityIdentityMatches {
				confirmed = []string{}
			}
			payload.FieldsUpdated = confirmed
		}
		if args.Tags != nil || args.Commute != nil || args.Feel != nil || args.RPE != nil || args.SubType != nil || args.DistanceM != nil {
			payload.Requested = &updateActivityRequestedValues{Tags: args.Tags, Commute: args.Commute, Feel: args.Feel, RPE: args.RPE, SubType: args.SubType, DistanceM: args.DistanceM}
		}
		if args.IncludeFull && (verification == nil || activityIdentityMatches) {
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
	for key, isNil := range map[string]bool{"tags": decoded.Tags == nil, "commute": decoded.Commute == nil, "feel": decoded.Feel == nil, "rpe": decoded.RPE == nil, "sub_type": decoded.SubType == nil, "distance_m": decoded.DistanceM == nil} {
		if fields[key] && isNil {
			return updateActivityRequest{}, fmt.Errorf("%s cannot be null; omit to leave unchanged", key)
		}
	}
	if err := decoded.params().Validate(); err != nil {
		return updateActivityRequest{}, err
	}
	return decoded, nil
}

func (args updateActivityRequest) params() intervals.UpdateActivityParams {
	p := intervals.UpdateActivityParams{ActivityID: args.ActivityID, Name: args.Name, NameSet: args.nameProvided, DescriptionSet: args.descriptionProvided, CarbsIngestedSet: args.carbsIngestedGProvided, Tags: args.Tags, Commute: args.Commute, Feel: args.Feel, RPE: args.RPE, SubType: args.SubType, DistanceM: args.DistanceM}
	if args.Description != nil {
		p.Description = *args.Description
	}
	if args.CarbsIngestedG != nil {
		p.CarbsIngested = *args.CarbsIngestedG
	}
	return p
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
	if args.Tags != nil {
		fields = append(fields, "tags")
	}
	if args.Commute != nil {
		fields = append(fields, "commute")
	}
	if args.Feel != nil {
		fields = append(fields, "feel")
	}
	if args.RPE != nil {
		fields = append(fields, "rpe")
	}
	if args.SubType != nil {
		fields = append(fields, "sub_type")
	}
	if args.DistanceM != nil {
		fields = append(fields, "distance_m")
	}
	sort.Strings(fields)
	return fields
}

func updateActivityInputSchema() map[string]any {
	examples := []map[string]any{{
		"activity_id": "gym-example-1",
		"description": "## Completed strength\n\n- Back squat: 3 × 5 @ 80 kg total; last set RPE 8.\n- Split squat: 3 × 8/side @ 16 kg per hand; left leg felt harder.",
	}}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"activity_id"},
		"examples":             examples,
		"input_examples":       examples,
		"properties": map[string]any{
			"tags":             map[string]any{"type": "array", "maxItems": 100, "items": map[string]any{"type": "string", "minLength": 1, "maxLength": 128}, "description": "Replace all activity tags; [] clears tags. Omit to preserve existing tags; null is rejected. At most 100 non-empty tags of at most 128 bytes each."},
			"commute":          map[string]any{"type": "boolean", "description": "Explicit commute flag; false clears the flag. Omit to leave unchanged; null is rejected. No inference from title or tags."},
			"feel":             map[string]any{"type": "integer", "minimum": 1, "maximum": 5, "description": "Athlete-reported feel, scale 1-5. Omit to leave unchanged; null and zero are rejected and clearing is unsupported."},
			"rpe":              map[string]any{"type": "integer", "minimum": 1, "maximum": 10, "description": "Athlete-reported rating of perceived exertion, scale 1-10, written to upstream icu_rpe. Omit to leave unchanged; null and zero are rejected and clearing is unsupported. Distinct from device perceived_exertion."},
			"sub_type":         map[string]any{"type": "string", "enum": []string{"NONE", "COMMUTE", "WARMUP", "COOLDOWN", "RACE"}, "description": "Activity subtype; NONE clears the subtype. Omit to preserve; null is rejected. Independent of the explicitly supplied commute flag."},
			"distance_m":       map[string]any{"type": "number", "minimum": 0, "maximum": 3.4028234663852886e38, "description": "Manual-activity distance in metres; zero is explicit. Only accepted for upstream source MANUAL with no recorded file. File-backed and unknown-source corrections are withheld because recomputation semantics are unverified. Omit to preserve; null is rejected. Time correction is unsupported."},
			"activity_id":      map[string]any{"type": "string", "description": "Required intervals.icu activity ID. Surrounding whitespace is trimmed; the i-prefix is preserved verbatim."},
			"name":             map[string]any{"type": "string", "description": "Optional replacement activity title. Omit to leave unchanged. Empty strings are rejected to avoid accidentally blanking the title; intervals.icu's UI also rejects blank titles."},
			"description":      map[string]any{"type": "string", "description": "Optional replacement free-text activity description; this is not append-only. Omit to leave unchanged; pass an explicit empty string to clear the description. For completed WeightTraining, read and preserve existing notes when editing, and record performed exercises as one Markdown '- ' line each after a blank line (not '•'); give actual sets/reps or time/distance and explicit load units/basis when known, and do not invent missing values. Prose metadata only — to write structured endurance intervals, use set_activity_intervals with a workout_doc."},
			"carbs_ingested_g": map[string]any{"type": "integer", "minimum": 0, "maximum": maxActivityCarbsIngestedG, "description": "Optional athlete-logged carbohydrate consumed during this activity, in whole grams (0-2147483647). Zero is a logged zero; omit to leave unchanged. Clearing is not supported. Distinct from read-only carbs_used_g. Intervals.icu does not update Strava activities through this endpoint."},
			"include_full":     map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream updated-activity payload under full; default returns a terse update confirmation."},
		},
	}
}

func updateActivityOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Non-destructive activity metadata update confirmation with activity_id, status, fields_updated, normalized athlete_id, and source_endpoint metadata. Supports sparse name, description, tags, commute, feel, rpe, subtype, manual-only distance_m, and carbs_ingested_g updates; carbs_used_g and gear assignment remain unsupported/read-only. Does not alter recorded streams or intervals. Upstream may recalculate derived metrics. New fields appear in fields_updated only after native values match; _meta.verification reports missing/mismatched evidence and requested labels submitted values."}
}
