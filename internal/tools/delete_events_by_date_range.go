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
)

const (
	deleteEventsByDateRangeName                    = "delete_events_by_date_range"
	deleteEventsByDateRangeDescription             = "Delete calendar events in a required athlete-local YYYY-MM-DD start_date/end_date range, optionally filtered by category. Same-day ranges are allowed; ranges must be 31 days or fewer. This destructive tool has no confirm argument and is registered only when ICUVISOR_DELETE_MODE=full."
	invalidDeleteEventsByDateRangeArgumentsMessage = "invalid delete_events_by_date_range arguments; provide start_date and end_date as YYYY-MM-DD, optional category, and a range of 31 days or fewer"
	deleteEventsByDateRangeMessage                 = "could not delete events by date range; check intervals.icu credentials, athlete ID, date range, category, and delete-mode configuration"
	deleteEventsByDateRangeEndpoint                = "/athlete/{id}/events"
	maxDeleteEventsByDateRangeDays                 = 31
	deleteEventsByDateRangeListLimit               = 500
)

// EventsByDateRangeDeleterClient lists and deletes calendar events for range deletes.
type EventsByDateRangeDeleterClient interface {
	ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error)
	DeleteEvent(context.Context, string) error
}

type deleteEventsByDateRangeRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Category  string `json:"category,omitempty"`
}

type deleteEventsByDateRangeResponse struct {
	DeletedIDs []string                            `json:"deleted_ids"`
	Status     string                              `json:"status"`
	Meta       deleteEventsByDateRangeResponseMeta `json:"_meta"`
}

type deleteEventsByDateRangeResponseMeta struct {
	Operation      string           `json:"operation"`
	ResourceType   string           `json:"resource_type"`
	SourceEndpoint string           `json:"source_endpoint"`
	DateRange      dateRangeMeta    `json:"date_range"`
	Timezone       string           `json:"timezone"`
	Category       string           `json:"category,omitempty"`
	RangeCapDays   int              `json:"range_cap_days"`
	DeletedCount   int              `json:"deleted_count"`
	Deleted        []map[string]any `json:"deleted"`
}

func newDeleteEventsByDateRangeTool(client EventsByDateRangeDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Tool {
	return Tool{Name: deleteEventsByDateRangeName, Description: deleteEventsByDateRangeDescription, InputSchema: deleteEventsByDateRangeInputSchema(), OutputSchema: deleteEventsByDateRangeOutputSchema(), Requirement: RequirementDelete, Handler: deleteEventsByDateRangeHandler(client, profileClient, version, timezoneFallback, debugMetadata)}
}

func deleteEventsByDateRangeHandler(client EventsByDateRangeDeleterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeDeleteEventsByDateRangeRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidDeleteEventsByDateRangeArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(deleteEventsByDateRangeMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(deleteEventsByDateRangeMessage, errors.New("missing events date-range deleter client"))
		}
		resolve := true
		events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: args.StartDate, Newest: args.EndDate, Category: args.Category, Limit: deleteEventsByDateRangeListLimit, Resolve: &resolve})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(deleteEventsByDateRangeMessage, err)
		}
		before, ids, err := eventDeleteRangeEchoes(events, timezoneName)
		if err != nil {
			return Result{}, fmt.Errorf("shaping delete_events_by_date_range before echoes: %w", err)
		}
		for _, id := range ids {
			if err := client.DeleteEvent(ctx, id); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(deleteEventsByDateRangeMessage, err)
			}
		}
		payload := deleteEventsByDateRangeResponse{DeletedIDs: ids, Status: "deleted", Meta: deleteEventsByDateRangeResponseMeta{Operation: "delete", ResourceType: "event", SourceEndpoint: deleteEventsByDateRangeEndpoint, DateRange: dateRangeMeta{Oldest: args.StartDate, Newest: args.EndDate}, Timezone: timezoneName, Category: args.Category, RangeCapDays: maxDeleteEventsByDateRangeDays, DeletedCount: len(ids), Deleted: before}}
		return encodeShaped(payload, false, nil, version, debugMetadata, deleteEventsByDateRangeName, unitSystem)
	}
}

func decodeDeleteEventsByDateRangeRequest(raw json.RawMessage) (deleteEventsByDateRangeRequest, error) {
	var args deleteEventsByDateRangeRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.Category = strings.TrimSpace(args.Category)
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date are required and must be YYYY-MM-DD")
	}
	start, _ := time.Parse(time.DateOnly, args.StartDate)
	end, _ := time.Parse(time.DateOnly, args.EndDate)
	if end.Before(start) {
		return args, errors.New("end_date must be on or after start_date")
	}
	if int(end.Sub(start).Hours()/24)+1 > maxDeleteEventsByDateRangeDays {
		return args, fmt.Errorf("date range must be %d days or fewer", maxDeleteEventsByDateRangeDays)
	}
	return args, nil
}

func eventDeleteRangeEchoes(events []intervals.Event, timezoneName string) ([]map[string]any, []string, error) {
	type pair struct {
		id   string
		echo map[string]any
	}
	pairs := make([]pair, 0, len(events))
	for _, event := range events {
		id := normalizedEventID(event)
		if id == "" {
			return nil, nil, errors.New("listed event is missing an ID")
		}
		echo, err := eventDeleteEcho(event, id, timezoneName)
		if err != nil {
			return nil, nil, err
		}
		pairs = append(pairs, pair{id: id, echo: echo})
	}
	sort.SliceStable(pairs, func(i, j int) bool {
		left := anyString(pairs[i].echo["start_date_local"])
		right := anyString(pairs[j].echo["start_date_local"])
		if left != right {
			return left < right
		}
		return pairs[i].id < pairs[j].id
	})
	before := make([]map[string]any, 0, len(pairs))
	ids := make([]string, 0, len(pairs))
	for _, item := range pairs {
		before = append(before, item.echo)
		ids = append(ids, item.id)
	}
	return before, ids, nil
}

func deleteEventsByDateRangeInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{
		"start_date": map[string]any{"type": "string", "description": "Required athlete-local start date YYYY-MM-DD. Open-ended ranges are rejected."},
		"end_date":   map[string]any{"type": "string", "description": "Required athlete-local end date YYYY-MM-DD. Same-day ranges are allowed; range size is capped at 31 inclusive days."},
		"category":   map[string]any{"type": "string", "description": "Optional upstream event category filter. Omit to delete all event categories in the bounded date range."},
	}}
}

func deleteEventsByDateRangeOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Delete confirmation with deleted_ids, _meta.deleted_count, and _meta.deleted terse echoes for events deleted from the bounded athlete-local date range."}
}
