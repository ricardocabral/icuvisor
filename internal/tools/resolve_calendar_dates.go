package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	resolveCalendarDatesName           = "resolve_calendar_dates"
	resolveCalendarDatesDescription    = "Resolve athlete-local calendar anchors before planning, scheduling, or answering date-sensitive prompts such as today, tomorrow, next week, or N days from today. Returns deterministic dates and weekdays from the athlete timezone so assistants do not infer weekday/date pairings with model arithmetic."
	invalidResolveCalendarDatesMessage = "invalid resolve_calendar_dates arguments; provide optional base_date as YYYY-MM-DD and unique offsets between -366 and 366"
	fetchResolveCalendarDatesMessage   = "could not resolve calendar dates; check athlete timezone"

	defaultCalendarOffset = 0
	maxCalendarOffsets    = 32
	minCalendarOffset     = -366
	maxCalendarOffset     = 366
)

type resolveCalendarDatesRequest struct {
	BaseDate string `json:"base_date,omitempty"`
	Offsets  []int  `json:"offsets,omitempty"`
}

type resolveCalendarDatesResponse struct {
	Dates []calendarDateAnchor     `json:"dates"`
	Meta  resolveCalendarDatesMeta `json:"_meta"`
}

type calendarDateAnchor struct {
	OffsetDays int    `json:"offset_days"`
	Date       string `json:"date"`
	Weekday    string `json:"weekday"`
}

type resolveCalendarDatesMeta struct {
	Timezone    string `json:"timezone"`
	BaseDate    string `json:"base_date"`
	BaseWeekday string `json:"base_weekday"`
	Count       int    `json:"count"`
}

func newResolveCalendarDatesTool(profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	return newResolveCalendarDatesToolWithClock(profileClient, version, timezoneFallback, debugMetadata, time.Now, shaping...)
}

func newResolveCalendarDatesToolWithClock(profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: resolveCalendarDatesName, Description: resolveCalendarDatesDescription, InputSchema: resolveCalendarDatesInputSchema(), OutputSchema: resolveCalendarDatesOutputSchema(), Handler: resolveCalendarDatesHandler(profileClient, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func resolveCalendarDatesHandler(profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		args, err := decodeResolveCalendarDatesRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidResolveCalendarDatesMessage, err)
		}
		if profileClient == nil {
			return Result{}, NewUserError(fetchResolveCalendarDatesMessage, errors.New("missing profile client"))
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchResolveCalendarDatesMessage, err)
		}
		payload, err := shapeResolveCalendarDates(args, now, timezoneName)
		if err != nil {
			return Result{}, NewUserError(invalidResolveCalendarDatesMessage, err)
		}
		return encodeShaped(payload, false, []string{"dates"}, version, debugMetadata, resolveCalendarDatesName, unitSystem, shapeCfg)
	}
}

func decodeResolveCalendarDatesRequest(raw json.RawMessage) (resolveCalendarDatesRequest, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		trimmed = []byte(`{}`)
	}
	if trimmed[0] != '{' {
		return resolveCalendarDatesRequest{}, errors.New("arguments must be a JSON object")
	}
	args, err := DecodeStrict[resolveCalendarDatesRequest](trimmed)
	if err != nil {
		return args, err
	}
	if args.BaseDate != "" && !validDate(args.BaseDate) {
		return args, errors.New("base_date must be YYYY-MM-DD")
	}
	if len(args.Offsets) == 0 {
		args.Offsets = []int{defaultCalendarOffset}
	}
	if len(args.Offsets) > maxCalendarOffsets {
		return args, fmt.Errorf("offsets must contain %d items or fewer", maxCalendarOffsets)
	}
	seen := make(map[int]struct{}, len(args.Offsets))
	for _, offset := range args.Offsets {
		if offset < minCalendarOffset || offset > maxCalendarOffset {
			return args, fmt.Errorf("offset %d is outside %d..%d", offset, minCalendarOffset, maxCalendarOffset)
		}
		if _, exists := seen[offset]; exists {
			return args, fmt.Errorf("duplicate offset %d", offset)
		}
		seen[offset] = struct{}{}
	}
	return args, nil
}

func shapeResolveCalendarDates(args resolveCalendarDatesRequest, now func() time.Time, timezoneName string) (resolveCalendarDatesResponse, error) {
	loc, err := time.LoadLocation(timezoneName)
	if err != nil {
		return resolveCalendarDatesResponse{}, fmt.Errorf("loading timezone %q: %w", timezoneName, err)
	}
	base, err := resolveCalendarBaseDate(args.BaseDate, now, loc)
	if err != nil {
		return resolveCalendarDatesResponse{}, err
	}
	dates := make([]calendarDateAnchor, 0, len(args.Offsets))
	for _, offset := range args.Offsets {
		date := base.AddDate(0, 0, offset)
		dates = append(dates, calendarDateAnchor{OffsetDays: offset, Date: date.Format(time.DateOnly), Weekday: date.Weekday().String()})
	}
	return resolveCalendarDatesResponse{Dates: dates, Meta: resolveCalendarDatesMeta{Timezone: timezoneName, BaseDate: base.Format(time.DateOnly), BaseWeekday: base.Weekday().String(), Count: len(dates)}}, nil
}

func resolveCalendarBaseDate(baseDate string, now func() time.Time, loc *time.Location) (time.Time, error) {
	if baseDate != "" {
		return time.ParseInLocation(time.DateOnly, baseDate, loc)
	}
	if now == nil {
		now = time.Now
	}
	localNow := now().In(loc)
	return time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc), nil
}

func resolveCalendarDatesInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"base_date": map[string]any{"type": "string", "description": "Optional athlete-local YYYY-MM-DD base date. Defaults to today's date from the athlete timezone, not UTC or the chat client's timezone."},
		"offsets":   map[string]any{"type": "array", "default": []int{defaultCalendarOffset}, "maxItems": maxCalendarOffsets, "uniqueItems": true, "items": map[string]any{"type": "integer", "minimum": minCalendarOffset, "maximum": maxCalendarOffset}, "description": "Integer day offsets from base_date. 0 is the base date, 1 is tomorrow, 7 is one week later, and negative values look backward. Each date is computed in the athlete timezone with calendar-day arithmetic."},
	}}
}

func resolveCalendarDatesOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Deterministic athlete-local calendar anchors. Each dates row has offset_days, date, and weekday; _meta has timezone, base_date, base_weekday, count, and server_version."}
}
