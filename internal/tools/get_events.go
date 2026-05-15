package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getEventsName                    = "get_events"
	getEventsDescription             = "List calendar events across a bounded athlete-local YYYY-MM-DD date range. Returns terse rows by default, raw upstream event payloads only with include_full:true, and preserves upstream category enum values."
	invalidGetEventsArgumentsMessage = "invalid get_events arguments; provide oldest/newest as YYYY-MM-DD with an optional capped limit"
	fetchEventsMessage               = "could not fetch events; check intervals.icu credentials, athlete ID, and date range"
	defaultEventsLimit               = 100
	maxEventsLimit                   = 500
	maxEventsRangeDays               = 366
)

// EventsClient lists intervals.icu calendar events for tools.
type EventsClient interface {
	ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error)
}

type getEventsRequest struct {
	Oldest      string `json:"oldest"`
	Newest      string `json:"newest"`
	Category    string `json:"category,omitempty"`
	CalendarID  string `json:"calendar_id,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Resolve     *bool  `json:"resolve,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getEventsResponse struct {
	Events []getEventsRow `json:"events"`
	Meta   getEventsMeta  `json:"_meta"`
}

type getEventsRow struct {
	EventID                  string         `json:"event_id,omitempty"`
	Category                 string         `json:"category,omitempty"`
	Type                     string         `json:"type,omitempty"`
	Name                     string         `json:"name,omitempty"`
	StartDateLocal           string         `json:"start_date_local,omitempty"`
	EndDateLocal             string         `json:"end_date_local,omitempty"`
	Description              string         `json:"description,omitempty"`
	WorkoutDocSummary        map[string]any `json:"workout_doc_summary,omitempty"`
	TrainingLoad             *float64       `json:"icu_training_load,omitempty"`
	LoadTarget               *float64       `json:"load_target,omitempty"`
	DistanceMeters           *float64       `json:"distance_meters,omitempty"`
	DistanceTargetMeters     *float64       `json:"distance_target_meters,omitempty"`
	MovingTimeSeconds        int            `json:"moving_time_seconds,omitempty"`
	TimeTargetSeconds        int            `json:"time_target_seconds,omitempty"`
	ElapsedTimeSeconds       int            `json:"elapsed_time_seconds,omitempty"`
	ElapsedTimeTargetSeconds int            `json:"elapsed_time_target_seconds,omitempty"`
	TrainingPlanID           string         `json:"training_plan_id,omitempty"`
	CalendarID               string         `json:"calendar_id,omitempty"`
	PlanApplied              string         `json:"plan_applied,omitempty"`
	PlanAppliedLocal         string         `json:"plan_applied_local,omitempty"`
	Updated                  string         `json:"updated,omitempty"`
	UpdatedLocal             string         `json:"updated_local,omitempty"`
	Full                     map[string]any `json:"full,omitempty"`
}

type getEventsMeta struct {
	DateRange   dateRangeMeta `json:"date_range"`
	Timezone    string        `json:"timezone"`
	Limit       int           `json:"limit"`
	Count       int           `json:"count"`
	Truncated   bool          `json:"truncated"`
	IncludeFull bool          `json:"include_full"`
}

type dateRangeMeta struct {
	Oldest string `json:"oldest"`
	Newest string `json:"newest"`
}

func newGetEventsTool(client EventsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getEventsName, Description: getEventsDescription, InputSchema: getEventsInputSchema(), OutputSchema: getEventsOutputSchema(), Handler: getEventsHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getEventsHandler(client EventsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetEventsRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetEventsArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchEventsMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchEventsMessage, errors.New("missing events client"))
		}
		events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: args.Oldest, Newest: args.Newest, Category: args.Category, CalendarID: args.CalendarID, Limit: args.Limit, Resolve: args.Resolve})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchEventsMessage, err)
		}
		payload, err := shapeGetEventsResponse(events, args, timezoneName)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_events response: %w", err)
		}
		return encodeShaped(payload, args.IncludeFull, []string{"events"}, version, debugMetadata, getEventsName, unitSystem, shapeCfg)
	}
}

func decodeGetEventsRequest(raw json.RawMessage) (getEventsRequest, error) {
	var args getEventsRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	args.Category = strings.TrimSpace(args.Category)
	args.CalendarID = strings.TrimSpace(args.CalendarID)
	if !validDate(args.Oldest) || !validDate(args.Newest) {
		return args, errors.New("oldest and newest must be YYYY-MM-DD")
	}
	oldest, _ := time.Parse(time.DateOnly, args.Oldest)
	newest, _ := time.Parse(time.DateOnly, args.Newest)
	if newest.Before(oldest) {
		return args, errors.New("newest must be on or after oldest")
	}
	if int(newest.Sub(oldest).Hours()/24)+1 > maxEventsRangeDays {
		return args, fmt.Errorf("date range must be %d days or fewer", maxEventsRangeDays)
	}
	if args.Limit <= 0 {
		args.Limit = defaultEventsLimit
	}
	if args.Limit > maxEventsLimit {
		args.Limit = maxEventsLimit
	}
	return args, nil
}

func shapeGetEventsResponse(events []intervals.Event, args getEventsRequest, timezoneName string) (getEventsResponse, error) {
	limit := args.Limit
	if limit <= 0 {
		limit = defaultEventsLimit
	}
	truncated := len(events) > limit
	if truncated {
		events = events[:limit]
	}

	rows := make([]getEventsRow, 0, len(events))
	for _, event := range events {
		row, err := eventRow(event, args.IncludeFull, timezoneName)
		if err != nil {
			return getEventsResponse{}, err
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].StartDateLocal != rows[j].StartDateLocal {
			return rows[i].StartDateLocal < rows[j].StartDateLocal
		}
		return rows[i].EventID < rows[j].EventID
	})
	return getEventsResponse{Events: rows, Meta: getEventsMeta{DateRange: dateRangeMeta{Oldest: args.Oldest, Newest: args.Newest}, Timezone: timezoneName, Limit: limit, Count: len(rows), Truncated: truncated, IncludeFull: args.IncludeFull}}, nil
}

func eventRow(event intervals.Event, includeFull bool, timezoneName string) (getEventsRow, error) {
	row := getEventsRow{EventID: event.ID, Category: firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"])), Type: stringValue(event.Type), Name: stringValue(event.Name), StartDateLocal: stringValue(event.StartDateLocal), EndDateLocal: stringValue(event.EndDateLocal), Description: stringValue(event.Description), TrainingLoad: event.TrainingLoad, LoadTarget: event.LoadTarget, DistanceMeters: event.Distance, DistanceTargetMeters: event.DistanceTarget, MovingTimeSeconds: intValue(event.MovingTime), TimeTargetSeconds: intValue(event.TimeTarget), ElapsedTimeSeconds: intValue(event.ElapsedTime), ElapsedTimeTargetSeconds: intValue(event.ElapsedTimeTarget), TrainingPlanID: anyString(firstRaw(event.Raw, "training_plan_id", "plan_id")), CalendarID: anyString(event.CalendarID), PlanApplied: stringValue(event.PlanApplied), Updated: stringValue(event.Updated)}
	if row.CalendarID == "" {
		row.CalendarID = anyString(event.Raw["calendar_id"])
	}
	if row.PlanApplied != "" {
		rendered, err := renderEventTimestamp(row.PlanApplied, timezoneName)
		if err != nil {
			return row, err
		}
		row.PlanAppliedLocal = rendered
	}
	if row.Updated != "" {
		rendered, err := renderEventTimestamp(row.Updated, timezoneName)
		if err != nil {
			return row, err
		}
		row.UpdatedLocal = rendered
	}
	if event.WorkoutDoc != nil {
		row.WorkoutDocSummary = workoutDocSummary(event.WorkoutDoc)
	}
	if includeFull {
		row.Full = cloneJSONMap(event.Raw)
	}
	return row, nil
}

func renderEventTimestamp(value string, timezoneName string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || validDate(trimmed) {
		return "", nil
	}
	parsed, parseErr := time.Parse(time.RFC3339Nano, trimmed)
	if parseErr == nil {
		return response.RenderTimeInTimezone(parsed, timezoneName)
	}
	return "", nil
}

func workoutDocSummary(value any) map[string]any {
	summary := map[string]any{"present": true}
	if typed, ok := value.(map[string]any); ok {
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		summary["top_level_keys"] = keys
		if steps, ok := typed["steps"].([]any); ok {
			summary["step_count"] = len(steps)
		}
		if name := anyString(typed["name"]); name != "" {
			summary["name"] = name
		}
		return summary
	}
	if steps, ok := value.([]any); ok {
		summary["step_count"] = len(steps)
	}
	return summary
}

func firstRaw(raw map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := raw[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func getEventsInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{
		"oldest":       map[string]any{"type": "string", "description": "Required athlete-local start date YYYY-MM-DD."},
		"newest":       map[string]any{"type": "string", "description": "Required athlete-local end date YYYY-MM-DD; date range is capped at 366 days."},
		"category":     map[string]any{"type": "string", "description": intervals.EventCategoryReferenceDescription("Optional upstream event category filter.")},
		"calendar_id":  map[string]any{"type": "string", "description": "Optional upstream calendar ID filter."},
		"limit":        map[string]any{"type": "integer", "default": defaultEventsLimit, "minimum": 1, "maximum": maxEventsLimit, "description": "Maximum event rows to request; defaults to 100 and values above 500 are capped."},
		"resolve":      map[string]any{"type": "boolean", "description": "Optional upstream resolve flag for recurring/derived events when supported by intervals.icu."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include each raw upstream event payload under full; default rows are terse and null-stripped."},
	}}
}

func getEventsOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Bounded calendar event rows with raw upstream category enum values, date-range metadata, truncation metadata, and optional full raw payloads."}
}
