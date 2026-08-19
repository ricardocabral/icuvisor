package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getActivitiesAroundName                    = "get_activities_around"
	getActivitiesAroundDescription             = "Find activities recorded near or around a known reference activity_id without asking the model to invent a date window. Use this after an activity has been identified and the prompt asks for nearby, before, after, or around-this-activity context; use get_activities instead for arbitrary athlete-local date windows, seasons, weeks, or prompts that still need an activity ID lookup."
	invalidGetActivitiesAroundArgumentsMessage = "invalid get_activities_around arguments; provide activity_id, optional limit 1..50, and optional include_full"
	fetchActivitiesAroundMessage               = "could not fetch activities around activity; check intervals.icu credentials, athlete ID, activity_id, and limit"
	defaultActivitiesAroundLimit               = 10
	maxActivitiesAroundLimit                   = 50
)

// ActivitiesAroundClient lists intervals.icu activities around a reference activity.
type ActivitiesAroundClient interface {
	ListActivitiesAround(context.Context, intervals.ActivitiesAroundParams) ([]intervals.Activity, error)
}

type getActivitiesAroundRequest struct {
	ActivityID  string `json:"activity_id"`
	Limit       int    `json:"-"`
	LimitArg    *int   `json:"limit,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getActivitiesAroundResponse struct {
	ReferenceActivityID string                  `json:"reference_activity_id"`
	Activities          []getActivitiesRow      `json:"activities"`
	Meta                getActivitiesAroundMeta `json:"_meta"`
}

type getActivitiesAroundMeta struct {
	Limit          int               `json:"limit"`
	ReturnedCount  int               `json:"returned_count"`
	IncludeFull    bool              `json:"include_full"`
	Timezone       string            `json:"timezone,omitempty"`
	EmptyReason    string            `json:"empty_reason,omitempty"`
	Guidance       string            `json:"guidance"`
	FieldSemantics map[string]string `json:"field_semantics,omitempty"`
}

func newGetActivitiesAroundTool(activityClient ActivitiesAroundClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{
		Name:         getActivitiesAroundName,
		Description:  getActivitiesAroundDescription,
		InputSchema:  getActivitiesAroundInputSchema(),
		OutputSchema: getActivitiesAroundOutputSchema(),
		Handler:      getActivitiesAroundHandler(activityClient, profileClient, gearClient, gearCache, version, timezoneFallback, debugMetadata, shapeCfg),
	})
}

func getActivitiesAroundHandler(activityClient ActivitiesAroundClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		args, err := decodeGetActivitiesAroundRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetActivitiesAroundArgumentsMessage, err)
		}
		if activityClient == nil || profileClient == nil {
			return Result{}, NewUserError(fetchActivitiesAroundMessage, errors.New("missing activities-around or profile client"))
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return Result{}, ctxErr
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchActivitiesAroundMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		activityTimezoneFallback := profileTimezone(profile.Timezone, timezoneFallback)
		activities, err := activityClient.ListActivitiesAround(ctx, intervals.ActivitiesAroundParams{ActivityID: args.ActivityID, Limit: args.Limit})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchActivitiesAroundMessage, err)
		}
		gearResolutions, err := resolveActivityGear(ctx, gearClient, gearCache, activities)
		if err != nil {
			return Result{}, err
		}
		shaped, err := shapeGetActivitiesAroundResponse(args, activities, gearResolutions, version, activityTimezoneFallback, debugMetadata, unitSystem, shapeCfg)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_activities_around response: %w", err)
		}
		if _, err := json.Marshal(shaped); err != nil {
			return Result{}, fmt.Errorf("encoding get_activities_around response: %w", err)
		}
		return TextResult(shaped), nil
	}
}

func decodeGetActivitiesAroundRequest(raw json.RawMessage) (getActivitiesAroundRequest, error) {
	args, err := DecodeStrict[getActivitiesAroundRequest](raw)
	if err != nil {
		return getActivitiesAroundRequest{}, err
	}
	args.ActivityID = strings.TrimSpace(args.ActivityID)
	if args.ActivityID == "" {
		return getActivitiesAroundRequest{}, errors.New("activity_id is required")
	}
	if args.LimitArg == nil {
		args.Limit = defaultActivitiesAroundLimit
	} else {
		args.Limit = *args.LimitArg
	}
	if args.Limit < 1 || args.Limit > maxActivitiesAroundLimit {
		return getActivitiesAroundRequest{}, fmt.Errorf("limit must be 1..%d", maxActivitiesAroundLimit)
	}
	return args, nil
}

func shapeGetActivitiesAroundResponse(args getActivitiesAroundRequest, activities []intervals.Activity, gearResolutions map[string]activityGearResolution, version string, timezoneFallback string, debugMetadata bool, unitSystem response.UnitSystem, shapeCfg responseShaping) (any, error) {
	rows := make([]getActivitiesRow, 0, len(activities))
	for _, activity := range activities {
		rows = append(rows, activityRow(activity, args.IncludeFull, timezoneFallback, unitSystem, gearResolutions[activity.ID], nil))
	}
	meta := getActivitiesAroundMeta{Limit: args.Limit, ReturnedCount: len(rows), IncludeFull: args.IncludeFull, Timezone: timezoneFallback, Guidance: "Use get_activities for arbitrary date-window, season, week, or activity-ID lookup prompts; use this tool when a reference activity_id is already known."}
	if len(rows) == 0 {
		meta.EmptyReason = "no_activities_returned_for_reference_activity"
	}
	meta.FieldSemantics = activityFieldSemantics(rows)
	payload := getActivitiesAroundResponse{ReferenceActivityID: args.ActivityID, Activities: rows, Meta: meta}
	return response.Shape(payload, shapeCfg.options(args.IncludeFull, []string{"activities"}, version, debugMetadata, getActivitiesAroundName, unitSystem))
}

func getActivitiesAroundInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"activity_id"}, "properties": map[string]any{
		"activity_id":  map[string]any{"type": "string", "description": "Required intervals.icu reference activity ID. Use get_activities first if the prompt only describes the activity by date, name, or recent context."},
		"limit":        map[string]any{"type": "integer", "default": defaultActivitiesAroundLimit, "minimum": 1, "maximum": maxActivitiesAroundLimit, "description": "Maximum number of nearby activities to return. Omitted defaults to 10; values outside 1..50 are rejected."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream activity fields and preserve upstream nulls; default terse rows are unit-disambiguated and null-stripped."},
	}}
}

func getActivitiesAroundOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Activities recorded near a known reference_activity_id. Rows reuse get_activities terse activity shaping, including upstream commute, feel as the athlete-reported 1-5 scale, and rpe from native icu_rpe on the separate 1-10 exertion scale; these fields are distinct, source-labelled, and never inferred or converted. Rows also include unit-disambiguated distance/speed/pace, calories_burned as active/exercise calories, carbs_ingested_g and carbs_used_g in grams, tags, Strava unavailable markers, gear_id/gear_name when upstream permits, and historical activity weather only when Intervals.icu returns completed-activity weather fields. Empty results return activities: [] with _meta.empty_reason and routing guidance. Use get_activities instead for arbitrary athlete-local date windows or prompts that still need activity ID lookup."}
}
