package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

const (
	getEventByIDName                    = "get_event_by_id"
	getEventByIDDescription             = "Fetch a single calendar event detail by event_id. If the detail endpoint 404s, performs one bounded list scan and returns a structured unavailable result instead of exposing the raw 404."
	invalidGetEventByIDArgumentsMessage = "invalid get_event_by_id arguments; provide event_id and optional YYYY-MM-DD date hints"
	fetchEventByIDMessage               = "could not fetch event; check intervals.icu credentials, athlete ID, event ID, and date hints"
	defaultEventByIDScanRadiusDays      = 30
	maxEventByIDScanRangeDays           = 61
	fallbackEventByIDLimit              = 500
)

// EventByIDClient retrieves event details and lists events for fallback scans.
type EventByIDClient interface {
	GetEvent(context.Context, string) (intervals.Event, error)
	ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error)
}

type getEventByIDRequest struct {
	EventID     string `json:"event_id"`
	Date        string `json:"date,omitempty"`
	Oldest      string `json:"oldest,omitempty"`
	Newest      string `json:"newest,omitempty"`
	Resolve     *bool  `json:"resolve,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getEventByIDResponse struct {
	Event       *getEventsRow         `json:"event,omitempty"`
	Unavailable *eventByIDUnavailable `json:"unavailable,omitempty"`
	Meta        getEventByIDMeta      `json:"_meta"`
}

type eventByIDUnavailable struct {
	Reason  string   `json:"reason"`
	Retried []string `json:"retried"`
}

type getEventByIDMeta struct {
	Source       string         `json:"source"`
	Recovered    bool           `json:"recovered"`
	Timezone     string         `json:"timezone"`
	IncludeFull  bool           `json:"include_full"`
	ScannedRange *dateRangeMeta `json:"scanned_range,omitempty"`
	Limit        int            `json:"limit,omitempty"`
	Count        int            `json:"count,omitempty"`
	Truncated    bool           `json:"truncated,omitempty"`
}

func newGetEventByIDTool(client EventByIDClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return newGetEventByIDToolWithClock(client, profileClient, version, timezoneFallback, debugMetadata, time.Now, shapeCfg)
}

func newGetEventByIDToolWithClock(client EventByIDClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getEventByIDName, Description: getEventByIDDescription, InputSchema: getEventByIDInputSchema(), OutputSchema: getEventByIDOutputSchema(), Handler: getEventByIDHandler(client, profileClient, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func getEventByIDHandler(client EventByIDClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetEventByIDRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetEventByIDArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchEventByIDMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchEventByIDMessage, errors.New("missing event client"))
		}

		event, err := client.GetEvent(ctx, args.EventID)
		if err == nil {
			payload, shapeErr := shapeGetEventByIDDetailResponse(event, args.IncludeFull, timezoneName)
			if shapeErr != nil {
				return Result{}, fmt.Errorf("shaping get_event_by_id detail response: %w", shapeErr)
			}
			return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, getEventByIDName, unitSystem, shapeCfg)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return Result{}, err
		}
		if !errors.Is(err, intervals.ErrNotFound) {
			return Result{}, NewUserError(fetchEventByIDMessage, err)
		}

		scanRange := eventByIDScanRange(args, timezoneName, now())
		resolve := true
		if args.Resolve != nil {
			resolve = *args.Resolve
		}
		events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: scanRange.Oldest, Newest: scanRange.Newest, Limit: fallbackEventByIDLimit, Resolve: &resolve})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchEventByIDMessage, err)
		}
		payload, err := shapeGetEventByIDScanResponse(events, args, timezoneName, scanRange)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_event_by_id fallback response: %w", err)
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, getEventByIDName, unitSystem, shapeCfg)
	}
}

func decodeGetEventByIDRequest(raw json.RawMessage) (getEventByIDRequest, error) {
	var args getEventByIDRequest
	if err := decodeStrict(raw, &args); err != nil {
		return args, err
	}
	args.EventID = strings.TrimSpace(args.EventID)
	args.Date = strings.TrimSpace(args.Date)
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	if args.EventID == "" {
		return args, errors.New("event_id is required")
	}
	if args.Date != "" && (args.Oldest != "" || args.Newest != "") {
		return args, errors.New("provide either date or oldest/newest, not both")
	}
	if args.Date != "" && !validDate(args.Date) {
		return args, errors.New("date must be YYYY-MM-DD")
	}
	if (args.Oldest == "") != (args.Newest == "") {
		return args, errors.New("oldest and newest must be provided together")
	}
	if args.Oldest != "" && (!validDate(args.Oldest) || !validDate(args.Newest)) {
		return args, errors.New("oldest and newest must be YYYY-MM-DD")
	}
	if args.Oldest != "" {
		if err := validateEventByIDRange(args.Oldest, args.Newest); err != nil {
			return args, err
		}
	}
	return args, nil
}

func eventByIDScanRange(args getEventByIDRequest, timezoneName string, now time.Time) dateRangeMeta {
	if args.Oldest != "" {
		return dateRangeMeta{Oldest: args.Oldest, Newest: args.Newest}
	}
	if args.Date != "" {
		center, _ := time.Parse(time.DateOnly, args.Date)
		return dateRangeMeta{Oldest: center.AddDate(0, 0, -defaultEventByIDScanRadiusDays).Format(time.DateOnly), Newest: center.AddDate(0, 0, defaultEventByIDScanRadiusDays).Format(time.DateOnly)}
	}
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		loc = time.UTC
	}
	localToday := now.In(loc)
	return dateRangeMeta{Oldest: localToday.AddDate(0, 0, -defaultEventByIDScanRadiusDays).Format(time.DateOnly), Newest: localToday.AddDate(0, 0, defaultEventByIDScanRadiusDays).Format(time.DateOnly)}
}

func validateEventByIDRange(oldest string, newest string) error {
	oldestDate, _ := time.Parse(time.DateOnly, oldest)
	newestDate, _ := time.Parse(time.DateOnly, newest)
	if newestDate.Before(oldestDate) {
		return errors.New("newest must be on or after oldest")
	}
	if int(newestDate.Sub(oldestDate).Hours()/24)+1 > maxEventByIDScanRangeDays {
		return fmt.Errorf("scan range must be %d days or fewer", maxEventByIDScanRangeDays)
	}
	return nil
}

func shapeGetEventByIDDetailResponse(event intervals.Event, includeFull bool, timezoneName string) (getEventByIDResponse, error) {
	row, err := eventRow(event, includeFull, timezoneName)
	if err != nil {
		return getEventByIDResponse{}, err
	}
	return getEventByIDResponse{Event: &row, Meta: getEventByIDMeta{Source: "detail", Recovered: false, Timezone: timezoneName, IncludeFull: includeFull}}, nil
}

func shapeGetEventByIDScanResponse(events []intervals.Event, args getEventByIDRequest, timezoneName string, scanRange dateRangeMeta) (getEventByIDResponse, error) {
	truncated := len(events) > fallbackEventByIDLimit
	if truncated {
		events = events[:fallbackEventByIDLimit]
	}
	meta := getEventByIDMeta{Source: "list_scan", Recovered: true, Timezone: timezoneName, IncludeFull: args.IncludeFull, ScannedRange: &scanRange, Limit: fallbackEventByIDLimit, Count: len(events), Truncated: truncated}
	for _, event := range events {
		if normalizedEventID(event) != args.EventID {
			continue
		}
		row, err := eventRow(event, args.IncludeFull, timezoneName)
		if err != nil {
			return getEventByIDResponse{}, err
		}
		return getEventByIDResponse{Event: &row, Meta: meta}, nil
	}
	return getEventByIDResponse{Unavailable: &eventByIDUnavailable{Reason: "upstream_inconsistency", Retried: []string{"detail", "list_scan"}}, Meta: meta}, nil
}

func normalizedEventID(event intervals.Event) string {
	if id := strings.TrimSpace(event.ID); id != "" {
		return id
	}
	return anyString(event.Raw["id"])
}

func getEventByIDInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"event_id"}, "properties": map[string]any{
		"event_id":     map[string]any{"type": "string", "description": "Required upstream event ID. Numeric upstream IDs should be passed as strings."},
		"date":         map[string]any{"type": "string", "description": "Optional athlete-local YYYY-MM-DD date hint. On detail 404, scans ±30 days around this date."},
		"oldest":       map[string]any{"type": "string", "description": "Optional athlete-local scan start YYYY-MM-DD. Must be paired with newest and span at most 61 days."},
		"newest":       map[string]any{"type": "string", "description": "Optional athlete-local scan end YYYY-MM-DD. Must be paired with oldest and span at most 61 days."},
		"resolve":      map[string]any{"type": "boolean", "description": "Optional upstream resolve flag for the fallback list scan; defaults to true to recover recurring or derived event instances."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include the raw upstream event payload under event.full; default response is terse."},
	}}
}

func getEventByIDOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "One calendar event as {event,_meta}, or a structured non-error {unavailable,_meta} result when detail 404 plus bounded list scan cannot find the ID."}
}
