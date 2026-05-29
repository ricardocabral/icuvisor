package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/workoutdoc"
)

const (
	addOrUpdateEventName                    = "add_or_update_event"
	addOrUpdateEventDescription             = "Create or update a non-destructive calendar event such as a planned workout, race, or note. Omitting event_id creates a new event; providing event_id updates that event without deleting or replacing unrelated events; use delete_event to remove. `description` is a replacement for the upstream event description/DSL, not append-only notes; omit it on updates to leave the current description unchanged. For WORKOUT updates, supplying `description` without `workout_doc` can replace existing structured steps; include the desired `workout_doc` to preserve or merge structure. When both are supplied, icuvisor merges them into the upstream description DSL and the `<!-- icuvisor:steps -->` sentinel controls serialized-step placement. Prefer `workout_doc` when the structure is known, and call `validate_workout` first if uncertain about the DSL syntax (see icuvisor://workout-syntax)."
	descriptionOnlyWorkoutWarning           = "Description was written without workout_doc; if this item previously had structured steps, they may have been replaced. Include workout_doc when preserving or merging workout structure."
	invalidAddOrUpdateEventArgumentsMessage = "invalid add_or_update_event arguments; provide date as athlete-local YYYY-MM-DD, category, type for WORKOUT events, name for NOTE creates, and optional event_id for updates"
	writeEventMessage                       = "could not write event; check intervals.icu credentials, athlete ID, event ID, and writable event fields"
)

// EventWriterClient creates or updates intervals.icu calendar events for tools.
type EventWriterClient interface {
	AddOrUpdateEvent(context.Context, intervals.WriteEventParams) (intervals.Event, error)
}

type addOrUpdateEventRequest struct {
	Date               string                 `json:"date"`
	EventID            string                 `json:"event_id,omitempty"`
	Category           string                 `json:"category"`
	Type               string                 `json:"type,omitempty"`
	Name               string                 `json:"name,omitempty"`
	Description        *string                `json:"description,omitempty"`
	WorkoutDoc         *workoutdoc.WorkoutDoc `json:"workout_doc,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	Indoor             *bool                  `json:"indoor,omitempty"`
	TargetLoad         *float64               `json:"target_load,omitempty"`
	DistanceMeters     *float64               `json:"distance_meters,omitempty"`
	MovingTimeSeconds  *int                   `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds *int                   `json:"elapsed_time_seconds,omitempty"`
	IncludeFull        bool                   `json:"include_full,omitempty"`

	tagsProvided bool
}

type addOrUpdateEventResponse struct {
	Event getEventsRow         `json:"event"`
	Meta  addOrUpdateEventMeta `json:"_meta"`
}

type addOrUpdateEventMeta struct {
	Operation                     string `json:"operation"`
	Date                          string `json:"date"`
	Timezone                      string `json:"timezone"`
	WorkoutDocUploaded            string `json:"workout_doc_uploaded,omitempty"`
	WorkoutDocWarning             string `json:"workout_doc_warning,omitempty"`
	DescriptionOnlyWorkoutWarning string `json:"description_only_workout_warning,omitempty"`
	IncludeFull                   bool   `json:"include_full"`
}

func newAddOrUpdateEventTool(client EventWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: addOrUpdateEventName, Description: addOrUpdateEventDescription, InputSchema: addOrUpdateEventInputSchema(), OutputSchema: addOrUpdateEventOutputSchema(), Requirement: RequirementWrite, Handler: addOrUpdateEventHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func addOrUpdateEventHandler(client EventWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAddOrUpdateEventRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAddOrUpdateEventArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(writeEventMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(writeEventMessage, errors.New("missing event writer client"))
		}
		params, workoutDocUploaded, err := eventWriteParams(args)
		if err != nil {
			return Result{}, NewUserError(invalidAddOrUpdateEventArgumentsMessage, err)
		}
		event, err := client.AddOrUpdateEvent(ctx, params)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(writeEventMessage, err)
		}
		payload, err := shapeAddOrUpdateEventResponse(event, args, timezoneName, workoutDocUploaded)
		if err != nil {
			return Result{}, fmt.Errorf("shaping add_or_update_event response: %w", err)
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addOrUpdateEventName, unitSystem, shapeCfg)
	}
}

func decodeAddOrUpdateEventRequest(raw json.RawMessage) (addOrUpdateEventRequest, error) {
	fields, err := rawObjectFields(raw)
	if err != nil {
		return addOrUpdateEventRequest{}, err
	}
	var args addOrUpdateEventRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[addOrUpdateEventRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Date = strings.TrimSpace(args.Date)
	args.EventID = strings.TrimSpace(args.EventID)
	args.Category = strings.TrimSpace(args.Category)
	args.Type = strings.TrimSpace(args.Type)
	args.Name = strings.TrimSpace(args.Name)
	args.tagsProvided = fields["tags"]
	if !validDate(args.Date) {
		return args, errors.New("date must be athlete-local YYYY-MM-DD")
	}
	if args.Category == "" {
		return args, errors.New("category is required")
	}
	if strings.EqualFold(args.Category, "WORKOUT") && args.Type == "" {
		return args, errors.New("type is required for WORKOUT events")
	}
	if args.EventID == "" && strings.EqualFold(args.Category, "NOTE") && args.Name == "" {
		return args, errors.New("name is required for NOTE creates")
	}
	if args.MovingTimeSeconds != nil && *args.MovingTimeSeconds < 0 {
		return args, errors.New("moving_time_seconds must be non-negative")
	}
	if args.ElapsedTimeSeconds != nil && *args.ElapsedTimeSeconds < 0 {
		return args, errors.New("elapsed_time_seconds must be non-negative")
	}
	if args.TargetLoad != nil && *args.TargetLoad < 0 {
		return args, errors.New("target_load must be non-negative")
	}
	if args.DistanceMeters != nil && *args.DistanceMeters < 0 {
		return args, errors.New("distance_meters must be non-negative")
	}
	return args, nil
}

func eventWriteParams(args addOrUpdateEventRequest) (intervals.WriteEventParams, string, error) {
	params := intervals.WriteEventParams{
		EventID:            args.EventID,
		Date:               args.Date,
		Category:           args.Category,
		Type:               args.Type,
		Name:               args.Name,
		Description:        args.Description,
		Tags:               append([]string(nil), args.Tags...),
		TagsSet:            args.tagsProvided,
		Indoor:             args.Indoor,
		TargetLoad:         args.TargetLoad,
		DistanceMeters:     args.DistanceMeters,
		MovingTimeSeconds:  args.MovingTimeSeconds,
		ElapsedTimeSeconds: args.ElapsedTimeSeconds,
	}
	if args.WorkoutDoc == nil {
		return params, "", nil
	}
	prose := ""
	if args.Description != nil {
		prose = *args.Description
	}
	dsl, err := workoutdoc.MergeDescription(prose, *args.WorkoutDoc)
	if err != nil {
		return intervals.WriteEventParams{}, "", fmt.Errorf("merging workout_doc with description: %w", err)
	}
	params.Description = &dsl
	return params, "description_dsl", nil
}

func shapeAddOrUpdateEventResponse(event intervals.Event, args addOrUpdateEventRequest, timezoneName string, workoutDocUploaded string) (addOrUpdateEventResponse, error) {
	row, err := eventRow(event, args.IncludeFull, timezoneName)
	if err != nil {
		return addOrUpdateEventResponse{}, err
	}
	operation := "create"
	if args.EventID != "" {
		operation = "update"
	}
	uploadedSteps := args.WorkoutDoc != nil && len(args.WorkoutDoc.Steps) > 0
	return addOrUpdateEventResponse{Event: row, Meta: addOrUpdateEventMeta{Operation: operation, Date: args.Date, Timezone: timezoneName, WorkoutDocUploaded: workoutDocUploaded, WorkoutDocWarning: workoutDocRenderWarning(uploadedSteps, event.WorkoutDoc), DescriptionOnlyWorkoutWarning: addOrUpdateEventDescriptionOnlyWorkoutWarning(args), IncludeFull: args.IncludeFull}}, nil
}

func addOrUpdateEventDescriptionOnlyWorkoutWarning(args addOrUpdateEventRequest) string {
	if args.EventID == "" || args.Description == nil || args.WorkoutDoc != nil || !strings.EqualFold(args.Category, "WORKOUT") {
		return ""
	}
	return descriptionOnlyWorkoutWarning
}

func addOrUpdateEventInputSchema() map[string]any {
	examples := addOrUpdateEventInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"date", "category"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"date":                 map[string]any{"type": "string", "description": "Required athlete-local event date as YYYY-MM-DD; interpreted in the configured athlete timezone."},
		"event_id":             map[string]any{"type": "string", "description": "Optional upstream event ID to update. Omit to create a new event; this tool never deletes events."},
		"category":             map[string]any{"type": "string", "description": intervals.EventCategoryReferenceDescription("Required upstream event category.")},
		"type":                 map[string]any{"type": "string", "description": "Required for WORKOUT events: upstream sport/activity type such as Ride, Run, Swim, or the athlete account's configured activity type. Surrounding whitespace is trimmed."},
		"name":                 map[string]any{"type": "string", "description": "Event title/name shown on the athlete calendar. Required when creating NOTE events; optional for other supported writes and updates unless upstream requires it."},
		"description":          map[string]any{"type": "string", "description": "Optional replacement for the upstream event description/DSL, not append-only notes. Omit on updates to leave unchanged. For WORKOUT updates, supplying description without workout_doc can replace existing structured steps; include the desired workout_doc to preserve or merge structure. Preserved verbatim, including whitespace and line breaks. May be supplied with workout_doc; use the " + workoutdoc.StepsSentinel + " sentinel on its own line to choose where serialized steps are inserted."},
		"workout_doc":          map[string]any{"type": "object", "description": "Optional structured WorkoutDoc. Serialized to the upstream workout DSL and merged with description when both are supplied. For WORKOUT updates, include the desired structured steps when changing prose so the replacement description/DSL preserves the workout structure. In each structured step, description is a label/comment only: do not include duration or distance tokens there; use duration seconds or distance instead. Syntax reference: icuvisor://workout-syntax."},
		"tags":                 map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional replacement event tags, in caller-provided order. Omit to leave unchanged on updates; provide [] to clear all tags."},
		"indoor":               map[string]any{"type": "boolean", "description": "Optional planned-event indoor/trainer flag. Set true for indoor trainer rides; commonly paired with type VirtualRide, but this boolean controls intervals.icu's Indoor toggle."},
		"target_load":          map[string]any{"type": "number", "minimum": 0, "description": "Optional planned training load / TSS equivalent when supported upstream."},
		"distance_meters":      map[string]any{"type": "number", "minimum": 0, "description": "Optional planned distance in meters when supported upstream."},
		"moving_time_seconds":  map[string]any{"type": "integer", "minimum": 0, "description": "Optional planned moving duration in seconds when supported upstream."},
		"elapsed_time_seconds": map[string]any{"type": "integer", "minimum": 0, "description": "Optional planned elapsed duration in seconds when supported upstream."},
		"include_full":         map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream event payload under event.full; default response matches get_event_by_id's terse event shape."},
	}}
}

func addOrUpdateEventInputExamples() []map[string]any {
	return []map[string]any{
		{
			"date":        "2026-06-12",
			"category":    "NOTE",
			"name":        "Race-week nutrition plan",
			"description": "Breakfast: oatmeal and banana. Lunch: rice bowl. Carry 90 g carbs/hour for the long ride.",
		},
		{
			"date":        "2026-06-15",
			"category":    "NOTE",
			"name":        "Travel logistics",
			"description": "Flight lands at 14:20. Pack pedals, charger, spare cleats, bottles, and race license.",
		},
		{
			"date":        "2026-06-17",
			"category":    "NOTE",
			"name":        "Daily reminder",
			"description": "Take resting HR after waking, do 10 minutes mobility, and log sleep quality before training.",
		},
		{
			"date":        "2026-06-18",
			"category":    "NOTE",
			"name":        "Coach annotation",
			"description": "Athlete reported tight calves; keep Thursday aerobic and reassess before intensity.",
		},
		{
			"date":     "2026-06-16",
			"category": "WORKOUT",
			"type":     "Ride",
			"name":     "Sweet spot repeats",
			"workout_doc": map[string]any{
				"steps": []any{
					map[string]any{"description": "Warm up", "duration": 600, "power": map[string]any{"value": 60, "units": "PERCENT_FTP"}},
					map[string]any{"description": "Main set", "reps": 3, "steps": []any{
						map[string]any{"duration": 480, "power": map[string]any{"value": 88, "units": "PERCENT_FTP"}},
						map[string]any{"duration": 240, "power": map[string]any{"value": 55, "units": "PERCENT_FTP"}},
					}},
				},
			},
			"tags":                []any{"sweet-spot", "indoor"},
			"indoor":              true,
			"target_load":         72,
			"moving_time_seconds": 4200,
		},
		{
			"date":                 "2026-09-13",
			"category":             "RACE_A",
			"type":                 "Run",
			"name":                 "Goal marathon",
			"description":          "A race. Confirm taper, fueling, pacing, and weather plan before adding workouts around it.",
			"tags":                 []any{"race", "goal-race"},
			"distance_meters":      42195,
			"target_load":          210,
			"elapsed_time_seconds": 10800,
		},
		{
			"event_id":             "evt-example-42",
			"date":                 "2026-06-21",
			"category":             "RACE_B",
			"type":                 "Run",
			"name":                 "10K tune-up race",
			"description":          "B race. Practice breakfast, warm-up, and even pacing.",
			"tags":                 []any{"race", "tune-up"},
			"distance_meters":      10000,
			"target_load":          85,
			"elapsed_time_seconds": 3300,
			"include_full":         true,
		},
		{
			"date":                "2026-05-30",
			"category":            "RACE_C",
			"type":                "Ride",
			"name":                "Local criterium practice",
			"description":         "C race. Treat as skills practice; keep surrounding plan priority higher.",
			"tags":                []any{"race", "practice"},
			"distance_meters":     40000,
			"target_load":         95,
			"moving_time_seconds": 5400,
		},
	}
}

func addOrUpdateEventOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Write confirmation containing the same terse event row shape used by get_event_by_id plus operation/date/timezone metadata. _meta.workout_doc_warning is set when intervals.icu stored the event but did not parse the uploaded workout_doc into a graphically rendered structured workout. _meta.description_only_workout_warning is set for WORKOUT event updates that supplied description without workout_doc."}
}
