package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	addUnavailableDateRangeName                    = "add_unavailable_date_range"
	addUnavailableDateRangeDescription             = "Create unavailable calendar markers across an inclusive athlete-local date range for Sick, Injured, or Holiday/time-off blocks. Writes one non-destructive per-day event with type Unavailable, skips identical retries, reports same-day conflicts such as workouts without deleting or overwriting them, and rejects unsupported categories or ranges over 31 days."
	invalidAddUnavailableDateRangeArgumentsMessage = "invalid add_unavailable_date_range arguments; provide start_date/end_date as YYYY-MM-DD, category Sick/Injured/Holiday, and a range of 31 days or fewer"
	writeUnavailableDateRangeMessage               = "could not write unavailable date range; check intervals.icu credentials, athlete ID, date range, and event fields"
	addUnavailableDateRangeExternalIDPrefix        = "icuvisor-unavailable-v1-"
	addUnavailableDateRangeDigestLength            = 24
	addUnavailableDateRangeType                    = "Unavailable"
	maxAddUnavailableDateRangeDays                 = 31
)

// UnavailableDateRangeWriterClient lists and writes calendar events for unavailable range creates.
type UnavailableDateRangeWriterClient interface {
	AddOrUpdateEvent(context.Context, intervals.WriteEventParams) (intervals.Event, error)
	ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error)
}

type addUnavailableDateRangeRequest struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Category    string `json:"category"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type addUnavailableDateRangeResponse struct {
	Events []getEventsRow              `json:"events"`
	Status string                      `json:"status"`
	Meta   addUnavailableDateRangeMeta `json:"_meta"`
}

type addUnavailableDateRangeMeta struct {
	Operation        string                        `json:"operation"`
	DateRange        dateRangeMeta                 `json:"date_range"`
	Timezone         string                        `json:"timezone"`
	Category         string                        `json:"category"`
	RequestedDays    int                           `json:"requested_days"`
	CreatedCount     int                           `json:"created_count"`
	SkippedCount     int                           `json:"skipped_count"`
	ConflictCount    int                           `json:"conflict_count"`
	RangeCapDays     int                           `json:"range_cap_days"`
	IncludeFull      bool                          `json:"include_full"`
	Skipped          []addUnavailableDateRangeSkip `json:"skipped,omitempty"`
	SameDayConflicts []applyTrainingPlanConflict   `json:"same_day_conflicts,omitempty"`
}

type addUnavailableDateRangeSkip struct {
	Date    string `json:"date"`
	EventID string `json:"event_id"`
	Reason  string `json:"reason"`
}

type addUnavailableDateRangeDay struct {
	Date      string
	Params    intervals.WriteEventParams
	Duplicate *intervals.Event
	Reason    string
	Conflicts []applyTrainingPlanConflict
}

func newAddUnavailableDateRangeTool(client UnavailableDateRangeWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: addUnavailableDateRangeName, Description: addUnavailableDateRangeDescription, InputSchema: addUnavailableDateRangeInputSchema(), OutputSchema: addUnavailableDateRangeOutputSchema(), Requirement: RequirementWrite, Handler: addUnavailableDateRangeHandler(client, profileClient, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func addUnavailableDateRangeHandler(client UnavailableDateRangeWriterClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAddUnavailableDateRangeRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAddUnavailableDateRangeArgumentsMessage, err)
		}
		profile, unitSystem, timezoneName, err := toolProfileDetails(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(writeUnavailableDateRangeMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(writeUnavailableDateRangeMessage, errors.New("missing unavailable date-range writer client"))
		}
		days, err := addUnavailableDateRangeDays(args)
		if err != nil {
			return Result{}, NewUserError(invalidAddUnavailableDateRangeArgumentsMessage, err)
		}
		existing, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: args.StartDate, Newest: args.EndDate, Limit: maxEventsLimit})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(writeUnavailableDateRangeMessage, err)
		}
		planned := addUnavailableDateRangePlan(args, days, existing)
		written := make([]intervals.Event, 0, len(planned))
		for _, day := range planned {
			if day.Duplicate != nil {
				continue
			}
			event, err := client.AddOrUpdateEvent(ctx, day.Params)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return Result{}, err
				}
				return Result{}, NewUserError(writeUnavailableDateRangeMessage, err)
			}
			written = append(written, event)
		}
		payload, err := shapeAddUnavailableDateRangeResponse(args, planned, written, timezoneName, profile, unitSystem)
		if err != nil {
			return Result{}, fmt.Errorf("shaping add_unavailable_date_range response: %w", err)
		}
		return encodeShaped(payload, args.IncludeFull, []string{"events"}, version, debugMetadata, addUnavailableDateRangeName, unitSystem, shapeCfg)
	}
}

func decodeAddUnavailableDateRangeRequest(raw json.RawMessage) (addUnavailableDateRangeRequest, error) {
	var args addUnavailableDateRangeRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[addUnavailableDateRangeRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	category, err := normalizeUnavailableDateRangeCategory(args.Category)
	if err != nil {
		return args, err
	}
	args.Category = category
	args.Name = strings.TrimSpace(args.Name)
	if args.Name == "" {
		args.Name = unavailableDateRangeDefaultName(args.Category)
	}
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date are required and must be YYYY-MM-DD")
	}
	start, _ := time.Parse(time.DateOnly, args.StartDate)
	end, _ := time.Parse(time.DateOnly, args.EndDate)
	if end.Before(start) {
		return args, errors.New("end_date must be on or after start_date")
	}
	if addUnavailableDateRangeDayCount(start, end) > maxAddUnavailableDateRangeDays {
		return args, fmt.Errorf("date range must be %d days or fewer", maxAddUnavailableDateRangeDays)
	}
	return args, nil
}

func normalizeUnavailableDateRangeCategory(category string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(category))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.Join(strings.Fields(normalized), "_")
	switch normalized {
	case "HOLIDAY", "HOLIDAYS", "VACATION", "PTO", "TIME_OFF":
		return "HOLIDAY", nil
	case "SICK", "ILL", "ILLNESS", "SICKNESS":
		return "SICK", nil
	case "INJURED", "INJURY":
		return "INJURED", nil
	default:
		return "", fmt.Errorf("unsupported unavailable category %q", strings.TrimSpace(category))
	}
}

func unavailableDateRangeDefaultName(category string) string {
	switch category {
	case "HOLIDAY":
		return "Holiday"
	case "SICK":
		return "Sick"
	case "INJURED":
		return "Injured"
	default:
		return category
	}
}

func addUnavailableDateRangeDays(args addUnavailableDateRangeRequest) ([]string, error) {
	start, err := time.Parse(time.DateOnly, args.StartDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse(time.DateOnly, args.EndDate)
	if err != nil {
		return nil, err
	}
	days := addUnavailableDateRangeDayCount(start, end)
	out := make([]string, 0, days)
	for day := 0; day < days; day++ {
		out = append(out, start.AddDate(0, 0, day).Format(time.DateOnly))
	}
	return out, nil
}

func addUnavailableDateRangeDayCount(start time.Time, end time.Time) int {
	return int(end.Sub(start).Hours()/24) + 1
}

func addUnavailableDateRangePlan(args addUnavailableDateRangeRequest, days []string, events []intervals.Event) []addUnavailableDateRangeDay {
	byDate := make(map[string][]intervals.Event, len(events))
	for _, event := range events {
		date := eventDateOnly(event)
		if date == "" {
			continue
		}
		byDate[date] = append(byDate[date], event)
	}
	planned := make([]addUnavailableDateRangeDay, 0, len(days))
	for _, date := range days {
		params := addUnavailableDateRangeParams(args, date)
		day := addUnavailableDateRangeDay{Date: date, Params: params}
		for _, event := range byDate[date] {
			if eventMatchesExternalID(event, params.ExternalID) && eventMatchesWriteParams(event, params) {
				if day.Duplicate == nil {
					duplicate := event
					day.Duplicate = &duplicate
					day.Reason = "matching_external_id"
				}
				continue
			}
			if eventMatchesWriteParams(event, params) {
				if day.Duplicate == nil {
					duplicate := event
					day.Duplicate = &duplicate
					day.Reason = "duplicate_existing_event"
				}
				continue
			}
			day.Conflicts = append(day.Conflicts, addUnavailableDateRangeConflict(event, date))
		}
		planned = append(planned, day)
	}
	return planned
}

func addUnavailableDateRangeParams(args addUnavailableDateRangeRequest, date string) intervals.WriteEventParams {
	var description *string
	if args.Description != "" {
		description = &args.Description
	}
	return intervals.WriteEventParams{ExternalID: addUnavailableDateRangeExternalID(args.Category, date, args.Name, args.Description), Date: date, Category: args.Category, Type: addUnavailableDateRangeType, Name: args.Name, Description: description}
}

func addUnavailableDateRangeExternalID(category string, date string, name string, description string) string {
	parts := []string{strings.TrimSpace(category), strings.TrimSpace(date), strings.TrimSpace(name), description}
	digest := sha256.Sum256([]byte(strings.Join(parts, "\n")))
	return addUnavailableDateRangeExternalIDPrefix + hex.EncodeToString(digest[:])[:addUnavailableDateRangeDigestLength]
}

func addUnavailableDateRangeConflict(event intervals.Event, date string) applyTrainingPlanConflict {
	return applyTrainingPlanConflict{EventID: normalizedEventID(event), Date: date, Category: firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"])), Type: firstNonEmpty(stringValue(event.Type), anyString(event.Raw["type"])), Name: firstNonEmpty(stringValue(event.Name), anyString(event.Raw["name"])), Reason: "existing_event_on_date", Protected: true}
}

func shapeAddUnavailableDateRangeResponse(args addUnavailableDateRangeRequest, planned []addUnavailableDateRangeDay, written []intervals.Event, timezoneName string, profile intervals.AthleteWithSportSettings, unitSystem response.UnitSystem) (addUnavailableDateRangeResponse, error) {
	writtenByDate := make(map[string]intervals.Event, len(written))
	for _, event := range written {
		writtenByDate[eventDateOnly(event)] = event
	}
	rows := make([]getEventsRow, 0, len(planned))
	skipped := make([]addUnavailableDateRangeSkip, 0)
	conflicts := make([]applyTrainingPlanConflict, 0)
	createdCount := 0
	for _, day := range planned {
		if day.Duplicate != nil {
			row, err := eventRow(*day.Duplicate, args.IncludeFull, timezoneName, workoutPreviewContextForEvent(*day.Duplicate, profile, unitSystem))
			if err != nil {
				return addUnavailableDateRangeResponse{}, err
			}
			rows = append(rows, row)
			skipped = append(skipped, addUnavailableDateRangeSkip{Date: day.Date, EventID: day.Duplicate.ID, Reason: day.Reason})
		} else if event, ok := writtenByDate[day.Date]; ok {
			row, err := eventRow(event, args.IncludeFull, timezoneName, workoutPreviewContextForEvent(event, profile, unitSystem))
			if err != nil {
				return addUnavailableDateRangeResponse{}, err
			}
			rows = append(rows, row)
			createdCount++
		}
		conflicts = append(conflicts, day.Conflicts...)
	}
	status := "created"
	if createdCount == 0 {
		status = "skipped"
	} else if len(skipped) > 0 || len(conflicts) > 0 {
		status = "partial"
	}
	meta := addUnavailableDateRangeMeta{Operation: "create_range", DateRange: dateRangeMeta{Oldest: args.StartDate, Newest: args.EndDate}, Timezone: timezoneName, Category: args.Category, RequestedDays: len(planned), CreatedCount: createdCount, SkippedCount: len(skipped), ConflictCount: len(conflicts), RangeCapDays: maxAddUnavailableDateRangeDays, IncludeFull: args.IncludeFull, Skipped: skipped, SameDayConflicts: conflicts}
	return addUnavailableDateRangeResponse{Events: rows, Status: status, Meta: meta}, nil
}

func addUnavailableDateRangeInputSchema() map[string]any {
	examples := addUnavailableDateRangeInputExamples()
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date", "category"}, "examples": examples, "input_examples": examples, "properties": map[string]any{
		"start_date":   map[string]any{"type": "string", "description": "Required athlete-local inclusive start date as YYYY-MM-DD."},
		"end_date":     map[string]any{"type": "string", "description": "Required athlete-local inclusive end date as YYYY-MM-DD; same-day ranges are allowed and ranges are capped at 31 days."},
		"category":     map[string]any{"type": "string", "enum": []string{"HOLIDAY", "SICK", "INJURED", "HOLIDAYS", "VACATION", "PTO", "TIME_OFF", "TIME OFF", "ILL", "ILLNESS", "SICKNESS", "INJURY"}, "description": "Closed unavailable category or alias. Accepted canonical values are HOLIDAY, SICK, and INJURED; time-off aliases normalize to HOLIDAY."},
		"name":         map[string]any{"type": "string", "description": "Optional event title/name. Defaults to Holiday, Sick, or Injured after category normalization; surrounding whitespace is trimmed."},
		"description":  map[string]any{"type": "string", "description": "Optional replacement event description written to each per-day unavailable marker."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream event payloads under each returned event.full; default response uses terse event rows."},
	}}
}

func addUnavailableDateRangeInputExamples() []map[string]any {
	return []map[string]any{
		{"start_date": "2026-07-01", "end_date": "2026-07-03", "category": "HOLIDAY", "description": "Family holiday; no training."},
		{"start_date": "2026-08-10", "end_date": "2026-08-11", "category": "SICK", "description": "Flu symptoms; rest only."},
		{"start_date": "2026-09-01", "end_date": "2026-09-05", "category": "injury", "name": "Calf injury block"},
	}
}

func addUnavailableDateRangeOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Unavailable range write result with terse event rows, status created/partial/skipped, and _meta date_range/category/count/skipped/conflict details. include_full:true includes raw upstream payloads for created or skipped event rows."}
}
