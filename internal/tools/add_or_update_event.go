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
	addOrUpdateEventDescription             = "Create or update a non-destructive calendar event such as a planned workout, race, or note. Omitting event_id creates a new event; providing event_id updates that event without deleting or replacing unrelated events."
	invalidAddOrUpdateEventArgumentsMessage = "invalid add_or_update_event arguments; provide date as athlete-local YYYY-MM-DD, category, type for WORKOUT events, and optional event_id for updates"
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
	TargetLoad         *float64               `json:"target_load,omitempty"`
	DistanceMeters     *float64               `json:"distance_meters,omitempty"`
	MovingTimeSeconds  *int                   `json:"moving_time_seconds,omitempty"`
	ElapsedTimeSeconds *int                   `json:"elapsed_time_seconds,omitempty"`
	IncludeFull        bool                   `json:"include_full,omitempty"`
}

type addOrUpdateEventResponse struct {
	Event getEventsRow         `json:"event"`
	Meta  addOrUpdateEventMeta `json:"_meta"`
}

type addOrUpdateEventMeta struct {
	Operation          string `json:"operation"`
	Date               string `json:"date"`
	Timezone           string `json:"timezone"`
	WorkoutDocUploaded string `json:"workout_doc_uploaded,omitempty"`
	IncludeFull        bool   `json:"include_full"`
}

func newAddOrUpdateEventTool(client EventWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: addOrUpdateEventName, Description: addOrUpdateEventDescription, InputSchema: addOrUpdateEventInputSchema(), OutputSchema: addOrUpdateEventOutputSchema(), Requirement: RequirementWrite, Handler: addOrUpdateEventHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func addOrUpdateEventHandler(client EventWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
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
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, addOrUpdateEventName, unitSystem)
	}
}

func decodeAddOrUpdateEventRequest(raw json.RawMessage) (addOrUpdateEventRequest, error) {
	var args addOrUpdateEventRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.Date = strings.TrimSpace(args.Date)
	args.EventID = strings.TrimSpace(args.EventID)
	args.Category = strings.TrimSpace(args.Category)
	args.Type = strings.TrimSpace(args.Type)
	args.Name = strings.TrimSpace(args.Name)
	if !validDate(args.Date) {
		return args, errors.New("date must be athlete-local YYYY-MM-DD")
	}
	if args.Category == "" {
		return args, errors.New("category is required")
	}
	if strings.EqualFold(args.Category, "WORKOUT") && args.Type == "" {
		return args, errors.New("type is required for WORKOUT events")
	}
	if args.Description != nil && args.WorkoutDoc != nil {
		return args, errors.New("provide either free-text description or structured workout_doc, not both")
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
		TargetLoad:         args.TargetLoad,
		DistanceMeters:     args.DistanceMeters,
		MovingTimeSeconds:  args.MovingTimeSeconds,
		ElapsedTimeSeconds: args.ElapsedTimeSeconds,
	}
	if args.WorkoutDoc == nil {
		return params, "", nil
	}
	dsl, err := workoutdoc.Serialize(*args.WorkoutDoc)
	if err != nil {
		return intervals.WriteEventParams{}, "", fmt.Errorf("serializing workout_doc: %w", err)
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
	return addOrUpdateEventResponse{Event: row, Meta: addOrUpdateEventMeta{Operation: operation, Date: args.Date, Timezone: timezoneName, WorkoutDocUploaded: workoutDocUploaded, IncludeFull: args.IncludeFull}}, nil
}

func addOrUpdateEventInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"date", "category"}, "properties": map[string]any{
		"date":                 map[string]any{"type": "string", "description": "Required athlete-local event date as YYYY-MM-DD; interpreted in the configured athlete timezone."},
		"event_id":             map[string]any{"type": "string", "description": "Optional upstream event ID to update. Omit to create a new event; this tool never deletes events."},
		"category":             map[string]any{"type": "string", "description": "Required upstream event category enum such as WORKOUT, RACE, NOTE, or the athlete account's configured category value."},
		"type":                 map[string]any{"type": "string", "description": "Required for WORKOUT events: upstream sport/activity type such as Ride, Run, Swim, or the athlete account's configured activity type. Surrounding whitespace is trimmed."},
		"name":                 map[string]any{"type": "string", "description": "Optional event title/name shown on the athlete calendar."},
		"description":          map[string]any{"type": "string", "description": "Optional free-text athlete or coach notes. Preserved verbatim, including whitespace and line breaks; mutually exclusive with workout_doc."},
		"workout_doc":          map[string]any{"type": "object", "description": "Optional structured workout steps using icuvisor's WorkoutDoc shape. Mutually exclusive with description; the server serializes this to the intervals.icu workout DSL string in the upstream description field and never sends the structured object upstream."},
		"tags":                 map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Optional event tags to preserve on the upstream event, in caller-provided order."},
		"target_load":          map[string]any{"type": "number", "minimum": 0, "description": "Optional planned training load / TSS equivalent when supported upstream."},
		"distance_meters":      map[string]any{"type": "number", "minimum": 0, "description": "Optional planned distance in meters when supported upstream."},
		"moving_time_seconds":  map[string]any{"type": "integer", "minimum": 0, "description": "Optional planned moving duration in seconds when supported upstream."},
		"elapsed_time_seconds": map[string]any{"type": "integer", "minimum": 0, "description": "Optional planned elapsed duration in seconds when supported upstream."},
		"include_full":         map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream event payload under event.full; default response matches get_event_by_id's terse event shape."},
	}}
}

func addOrUpdateEventOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Write confirmation containing the same terse event row shape used by get_event_by_id plus operation/date/timezone metadata."}
}
