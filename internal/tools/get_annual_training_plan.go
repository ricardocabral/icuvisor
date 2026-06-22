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
	getAnnualTrainingPlanName        = "get_annual_training_plan"
	getAnnualTrainingPlanDescription = "Use when the prompt asks about annual training plan, season phase, weekly load/TSS targets, recovery weeks, taper context, or periodization summary; do not manually join raw get_events rows in chat. Summarizes existing PLAN, TARGET, and NOTE calendar events into phases, weekly targets, recovery/context notes, and projection-ready weekly_plan_targets without writing calendar data."
	invalidAnnualTrainingPlanMessage = "invalid get_annual_training_plan arguments; provide oldest/newest as YYYY-MM-DD with an optional capped limit"
	fetchAnnualTrainingPlanMessage   = "could not fetch annual training plan events; check intervals.icu credentials, athlete ID, and date range"

	annualTrainingPlanEventLimit = 500
	annualTrainingPlanMaxRange   = 366
	annualTrainingPlanEndpoint   = "/athlete/{id}/events"
	annualTrainingPlanSchema     = "annual_training_plan.v1"
)

type annualTrainingPlanRequest struct {
	Oldest      string `json:"oldest"`
	Newest      string `json:"newest"`
	CalendarID  string `json:"calendar_id,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type annualTrainingPlanResponse struct {
	Summary     annualTrainingPlanSummary      `json:"summary"`
	Phases      []annualTrainingPlanPhase      `json:"phases"`
	Weeks       []annualTrainingPlanWeek       `json:"weeks"`
	Notes       []annualTrainingPlanNote       `json:"notes"`
	Unavailable *annualTrainingPlanUnavailable `json:"unavailable,omitempty"`
	Meta        annualTrainingPlanMeta         `json:"_meta"`
}

type annualTrainingPlanSummary struct {
	PhaseCount                int           `json:"phase_count"`
	WeekCount                 int           `json:"week_count"`
	NoteCount                 int           `json:"note_count"`
	TargetEventCount          int           `json:"target_event_count"`
	WeeksWithLoadTargets      int           `json:"weeks_with_load_targets"`
	TotalLoadTarget           float64       `json:"total_load_target"`
	TotalTimeTargetSeconds    int           `json:"total_time_target_seconds"`
	TotalDistanceTargetMeters float64       `json:"total_distance_target_meters"`
	CurrentPhaseID            string        `json:"current_phase_id,omitempty"`
	DateRange                 dateRangeMeta `json:"date_range"`
}

type annualTrainingPlanPhase struct {
	PhaseID              string         `json:"phase_id"`
	SourceEventID        string         `json:"source_event_id,omitempty"`
	Name                 string         `json:"name,omitempty"`
	Type                 string         `json:"type,omitempty"`
	StartDate            string         `json:"start_date"`
	EndDate              string         `json:"end_date"`
	EndDateSource        string         `json:"end_date_source"`
	Description          string         `json:"description,omitempty"`
	Tags                 []string       `json:"tags,omitempty"`
	LoadTarget           *float64       `json:"load_target,omitempty"`
	TimeTargetSeconds    *int           `json:"time_target_seconds,omitempty"`
	DistanceTargetMeters *float64       `json:"distance_target_meters,omitempty"`
	Full                 map[string]any `json:"full,omitempty"`
}

type annualTrainingPlanWeek struct {
	WeekStartDate          string                          `json:"week_start_date"`
	WeekEndDate            string                          `json:"week_end_date"`
	RangeOverlapStart      string                          `json:"range_overlap_start"`
	RangeOverlapEnd        string                          `json:"range_overlap_end"`
	PartialWeek            bool                            `json:"partial_week"`
	PhaseIDs               []string                        `json:"phase_ids"`
	TargetEventCount       int                             `json:"target_event_count"`
	LoadTarget             *float64                        `json:"load_target,omitempty"`
	TimeTargetSeconds      *int                            `json:"time_target_seconds,omitempty"`
	DistanceTargetMeters   *float64                        `json:"distance_target_meters,omitempty"`
	MissingLoadTargetCount int                             `json:"missing_load_target_count"`
	NoteCount              int                             `json:"note_count"`
	RecoveryNoteCount      int                             `json:"recovery_note_count"`
	NoteIDs                []string                        `json:"note_ids"`
	TargetEvents           []annualTrainingPlanTargetEvent `json:"target_events,omitempty"`
}

type annualTrainingPlanTargetEvent struct {
	EventID              string         `json:"event_id,omitempty"`
	Name                 string         `json:"name,omitempty"`
	Type                 string         `json:"type,omitempty"`
	StartDate            string         `json:"start_date"`
	LoadTarget           *float64       `json:"load_target,omitempty"`
	TimeTargetSeconds    *int           `json:"time_target_seconds,omitempty"`
	DistanceTargetMeters *float64       `json:"distance_target_meters,omitempty"`
	Full                 map[string]any `json:"full,omitempty"`
}

type annualTrainingPlanNote struct {
	NoteID         string         `json:"note_id"`
	SourceEventID  string         `json:"source_event_id,omitempty"`
	Name           string         `json:"name,omitempty"`
	Type           string         `json:"type,omitempty"`
	StartDate      string         `json:"start_date"`
	EndDate        string         `json:"end_date"`
	Description    string         `json:"description,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	RecoveryHint   bool           `json:"recovery_hint"`
	PhaseIDs       []string       `json:"phase_ids"`
	WeekStartDates []string       `json:"week_start_dates"`
	Full           map[string]any `json:"full,omitempty"`
}

type annualTrainingPlanUnavailable struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

type annualTrainingPlanMeta struct {
	SourceEndpoint          string                             `json:"source_endpoint"`
	DateRange               dateRangeMeta                      `json:"date_range"`
	Timezone                string                             `json:"timezone"`
	Limit                   int                                `json:"limit"`
	FetchedEventCount       int                                `json:"fetched_event_count"`
	PeriodizationEventCount int                                `json:"periodization_event_count"`
	PlanEventCount          int                                `json:"plan_event_count"`
	TargetEventCount        int                                `json:"target_event_count"`
	NoteEventCount          int                                `json:"note_event_count"`
	MalformedEventCount     int                                `json:"malformed_event_count"`
	Truncated               bool                               `json:"truncated"`
	IncludeFull             bool                               `json:"include_full"`
	SchemaVersion           string                             `json:"schema_version"`
	Caveats                 []string                           `json:"caveats"`
	ProjectionBridge        annualTrainingPlanProjectionBridge `json:"projection_bridge"`
}

type annualTrainingPlanProjectionBridge struct {
	TargetTool        string                                     `json:"target_tool"`
	TargetArgument    string                                     `json:"target_argument"`
	WeeklyPlanTargets []annualTrainingPlanProjectionWeeklyTarget `json:"weekly_plan_targets"`
	IncludedWeekCount int                                        `json:"included_week_count"`
	ExcludedWeekCount int                                        `json:"excluded_week_count"`
	Caveats           []string                                   `json:"caveats"`
}

type annualTrainingPlanProjectionWeeklyTarget struct {
	WeekStartDate string  `json:"week_start_date"`
	TrainingLoad  float64 `json:"training_load"`
}

type annualTrainingPlanEvents struct {
	plans     []annualTrainingPlanEvent
	targets   []annualTrainingPlanEvent
	notes     []annualTrainingPlanEvent
	malformed int
}

type annualTrainingPlanEvent struct {
	event     intervals.Event
	category  string
	startDate time.Time
	endDate   time.Time
	endValid  bool
	tags      []string
}

func newGetAnnualTrainingPlanTool(client EventsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	return newGetAnnualTrainingPlanToolWithClock(client, profileClient, version, timezoneFallback, debugMetadata, time.Now, shaping...)
}

func newGetAnnualTrainingPlanToolWithClock(client EventsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getAnnualTrainingPlanName, Description: getAnnualTrainingPlanDescription, InputSchema: getAnnualTrainingPlanInputSchema(), OutputSchema: getAnnualTrainingPlanOutputSchema(), Handler: getAnnualTrainingPlanHandler(client, profileClient, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func getAnnualTrainingPlanHandler(client EventsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeAnnualTrainingPlanRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidAnnualTrainingPlanMessage, err)
		}
		if client == nil {
			return Result{}, NewUserError(fetchAnnualTrainingPlanMessage, errors.New("missing events client"))
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchAnnualTrainingPlanMessage, err)
		}
		fetchLimit := args.Limit + 1
		events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: args.Oldest, Newest: args.Newest, CalendarID: args.CalendarID, Limit: fetchLimit})
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchAnnualTrainingPlanMessage, err)
		}
		truncated := len(events) > args.Limit
		if truncated {
			events = events[:args.Limit]
		}
		asOf, err := response.AsOfMetadataInTimezone(now(), timezoneName)
		if err != nil {
			return Result{}, NewUserError(fetchAnnualTrainingPlanMessage, err)
		}
		payload, err := shapeAnnualTrainingPlanResponse(events, args, timezoneName, asOf.AsOfDate, truncated)
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_annual_training_plan response: %w", err)
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, getAnnualTrainingPlanName, unitSystem, shapeCfg)
	}
}

func decodeAnnualTrainingPlanRequest(raw json.RawMessage) (annualTrainingPlanRequest, error) {
	var args annualTrainingPlanRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[annualTrainingPlanRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.Oldest = strings.TrimSpace(args.Oldest)
	args.Newest = strings.TrimSpace(args.Newest)
	args.CalendarID = strings.TrimSpace(args.CalendarID)
	if !validDate(args.Oldest) || !validDate(args.Newest) {
		return args, errors.New("oldest and newest must be YYYY-MM-DD")
	}
	oldest, _ := time.Parse(time.DateOnly, args.Oldest)
	newest, _ := time.Parse(time.DateOnly, args.Newest)
	if newest.Before(oldest) {
		return args, errors.New("newest must be on or after oldest")
	}
	if int(newest.Sub(oldest).Hours()/24)+1 > annualTrainingPlanMaxRange {
		return args, fmt.Errorf("date range must be %d days or fewer", annualTrainingPlanMaxRange)
	}
	if args.Limit <= 0 {
		args.Limit = annualTrainingPlanEventLimit
	}
	if args.Limit > annualTrainingPlanEventLimit {
		args.Limit = annualTrainingPlanEventLimit
	}
	return args, nil
}

func shapeAnnualTrainingPlanResponse(events []intervals.Event, args annualTrainingPlanRequest, timezoneName string, asOfDate string, truncated bool) (annualTrainingPlanResponse, error) {
	oldest, _ := time.Parse(time.DateOnly, args.Oldest)
	newest, _ := time.Parse(time.DateOnly, args.Newest)
	classified := annualTrainingPlanPeriodizationEvents(events)
	phases, phaseCaveats := annualTrainingPlanPhases(classified.plans, args.IncludeFull, oldest, newest)
	notes := annualTrainingPlanNotes(classified.notes, phases, oldest, newest, args.IncludeFull)
	weeks, weekCaveats := annualTrainingPlanWeeks(classified.targets, phases, notes, oldest, newest, args.IncludeFull)
	bridge := annualTrainingPlanBridge(weeks)
	caveats := []string{"upstream athlete-level periodization parameters such as ramp rate, recovery cadence, taper percent, and intensity distribution are not exposed; this summarizes calendar PLAN/TARGET/NOTE events only"}
	caveats = append(caveats, phaseCaveats...)
	caveats = append(caveats, weekCaveats...)
	caveats = append(caveats, bridge.Caveats...)
	if classified.malformed > 0 {
		caveats = append(caveats, "some PLAN/TARGET/NOTE events had missing or malformed athlete-local dates and were skipped")
	}
	if truncated {
		caveats = append(caveats, "event scan reached the requested limit; additional periodization events may exist")
	}
	periodizationCount := len(classified.plans) + len(classified.targets) + len(classified.notes)
	summary := annualTrainingPlanSummary{PhaseCount: len(phases), WeekCount: len(weeks), NoteCount: len(notes), TargetEventCount: len(classified.targets), DateRange: dateRangeMeta{Oldest: args.Oldest, Newest: args.Newest}}
	for _, week := range weeks {
		if week.LoadTarget != nil {
			summary.WeeksWithLoadTargets++
			summary.TotalLoadTarget += *week.LoadTarget
		}
		if week.TimeTargetSeconds != nil {
			summary.TotalTimeTargetSeconds += *week.TimeTargetSeconds
		}
		if week.DistanceTargetMeters != nil {
			summary.TotalDistanceTargetMeters += *week.DistanceTargetMeters
		}
	}
	if validDate(asOfDate) {
		current, _ := time.Parse(time.DateOnly, asOfDate)
		for _, phase := range phases {
			start, _ := time.Parse(time.DateOnly, phase.StartDate)
			end, _ := time.Parse(time.DateOnly, phase.EndDate)
			if !current.Before(start) && !current.After(end) {
				summary.CurrentPhaseID = phase.PhaseID
				break
			}
		}
	}
	payload := annualTrainingPlanResponse{Summary: summary, Phases: phases, Weeks: weeks, Notes: notes, Meta: annualTrainingPlanMeta{SourceEndpoint: annualTrainingPlanEndpoint, DateRange: dateRangeMeta{Oldest: args.Oldest, Newest: args.Newest}, Timezone: timezoneName, Limit: args.Limit, FetchedEventCount: len(events), PeriodizationEventCount: periodizationCount, PlanEventCount: len(classified.plans), TargetEventCount: len(classified.targets), NoteEventCount: len(classified.notes), MalformedEventCount: classified.malformed, Truncated: truncated, IncludeFull: args.IncludeFull, SchemaVersion: annualTrainingPlanSchema, Caveats: uniqueStrings(caveats), ProjectionBridge: bridge}}
	if periodizationCount == 0 {
		payload.Unavailable = &annualTrainingPlanUnavailable{Reason: "no_periodization_events", Detail: "no PLAN, TARGET, or NOTE calendar events were returned for the requested range"}
	}
	return payload, nil
}

func annualTrainingPlanPeriodizationEvents(events []intervals.Event) annualTrainingPlanEvents {
	out := annualTrainingPlanEvents{}
	for _, event := range events {
		category := strings.ToUpper(strings.TrimSpace(firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"]))))
		if category != "PLAN" && category != "TARGET" && category != "NOTE" {
			continue
		}
		start, ok := annualTrainingPlanEventDateOnly(event.StartDateLocal)
		if !ok {
			out.malformed++
			continue
		}
		end, endValid := annualTrainingPlanEventDateOnly(event.EndDateLocal)
		if !endValid {
			end = start
		}
		item := annualTrainingPlanEvent{event: event, category: category, startDate: start, endDate: end, endValid: endValid, tags: annualTrainingPlanTags(event)}
		switch category {
		case "PLAN":
			out.plans = append(out.plans, item)
		case "TARGET":
			out.targets = append(out.targets, item)
		case "NOTE":
			out.notes = append(out.notes, item)
		}
	}
	sortAnnualTrainingPlanEvents(out.plans)
	sortAnnualTrainingPlanEvents(out.targets)
	sortAnnualTrainingPlanEvents(out.notes)
	return out
}

func sortAnnualTrainingPlanEvents(events []annualTrainingPlanEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].startDate.Equal(events[j].startDate) {
			return events[i].startDate.Before(events[j].startDate)
		}
		updatedI := stringValue(events[i].event.Updated)
		updatedJ := stringValue(events[j].event.Updated)
		if updatedI != updatedJ {
			return updatedI < updatedJ
		}
		return events[i].event.ID < events[j].event.ID
	})
}

func annualTrainingPlanPhases(events []annualTrainingPlanEvent, includeFull bool, oldest time.Time, newest time.Time) ([]annualTrainingPlanPhase, []string) {
	phases := make([]annualTrainingPlanPhase, 0, len(events))
	caveats := []string{}
	for i, event := range events {
		end := event.endDate
		endSource := "event_end_date"
		if !event.endValid || end.Before(event.startDate) {
			if i+1 < len(events) {
				nextStart := events[i+1].startDate
				candidate := nextStart.AddDate(0, 0, -1)
				if candidate.Before(event.startDate) {
					candidate = event.startDate
				}
				end = candidate
				endSource = "next_phase_start_minus_one"
			} else {
				end = newest
				endSource = "range_end"
			}
		}
		if i+1 < len(events) && event.endValid && end.Equal(events[i+1].startDate) {
			endSource = "shared_boundary"
			caveats = append(caveats, "at least one PLAN phase ends on the same date the next phase starts; both explicit boundary dates were preserved")
		}
		if end.After(newest) {
			end = newest
		}
		if event.startDate.Before(oldest) && !event.endValid {
			endSource = "invalid_or_missing"
		}
		phase := annualTrainingPlanPhase{PhaseID: annualTrainingPlanRowID("phase", event.event.ID, len(phases)+1), SourceEventID: event.event.ID, Name: stringValue(event.event.Name), Type: stringValue(event.event.Type), StartDate: formatDate(event.startDate), EndDate: formatDate(end), EndDateSource: endSource, Description: stringValue(event.event.Description), Tags: event.tags, LoadTarget: event.event.LoadTarget, TimeTargetSeconds: annualTrainingPlanTimeTarget(event.event), DistanceTargetMeters: event.event.DistanceTarget}
		if includeFull {
			phase.Full = cloneJSONMap(event.event.Raw)
		}
		phases = append(phases, phase)
	}
	return phases, caveats
}

func annualTrainingPlanNotes(events []annualTrainingPlanEvent, phases []annualTrainingPlanPhase, oldest time.Time, newest time.Time, includeFull bool) []annualTrainingPlanNote {
	notes := make([]annualTrainingPlanNote, 0, len(events))
	for _, event := range events {
		start := event.startDate
		end := event.endDate
		if end.Before(start) {
			end = start
		}
		if start.Before(oldest) {
			start = oldest
		}
		if end.After(newest) {
			end = newest
		}
		note := annualTrainingPlanNote{NoteID: annualTrainingPlanRowID("note", event.event.ID, len(notes)+1), SourceEventID: event.event.ID, Name: stringValue(event.event.Name), Type: stringValue(event.event.Type), StartDate: formatDate(start), EndDate: formatDate(end), Description: stringValue(event.event.Description), Tags: event.tags, RecoveryHint: annualTrainingPlanRecoveryHint(event), PhaseIDs: annualTrainingPlanOverlappingPhaseIDs(phases, start, end), WeekStartDates: annualTrainingPlanWeekStarts(start, end)}
		if includeFull {
			note.Full = cloneJSONMap(event.event.Raw)
		}
		notes = append(notes, note)
	}
	return notes
}

func annualTrainingPlanWeeks(events []annualTrainingPlanEvent, phases []annualTrainingPlanPhase, notes []annualTrainingPlanNote, oldest time.Time, newest time.Time, includeFull bool) ([]annualTrainingPlanWeek, []string) {
	weeksByStart := map[string]*annualTrainingPlanWeek{}
	for weekStart := isoWeekStart(oldest); !weekStart.After(newest); weekStart = weekStart.AddDate(0, 0, 7) {
		weekEnd := weekStart.AddDate(0, 0, 6)
		overlapStart := maxDate(weekStart, oldest)
		overlapEnd := minDate(weekEnd, newest)
		key := formatDate(weekStart)
		weeksByStart[key] = &annualTrainingPlanWeek{WeekStartDate: key, WeekEndDate: formatDate(weekEnd), RangeOverlapStart: formatDate(overlapStart), RangeOverlapEnd: formatDate(overlapEnd), PartialWeek: weekStart.Before(oldest) || weekEnd.After(newest), PhaseIDs: annualTrainingPlanOverlappingPhaseIDs(phases, overlapStart, overlapEnd), NoteIDs: []string{}}
	}
	for _, event := range events {
		week := weeksByStart[formatDate(isoWeekStart(event.startDate))]
		if week == nil {
			continue
		}
		week.TargetEventCount++
		if event.event.LoadTarget != nil {
			week.LoadTarget = addFloatPtr(week.LoadTarget, *event.event.LoadTarget)
		} else {
			week.MissingLoadTargetCount++
		}
		if value := annualTrainingPlanTimeTarget(event.event); value != nil {
			week.TimeTargetSeconds = addIntPtr(week.TimeTargetSeconds, *value)
		}
		if event.event.DistanceTarget != nil {
			week.DistanceTargetMeters = addFloatPtr(week.DistanceTargetMeters, *event.event.DistanceTarget)
		}
		if includeFull {
			week.TargetEvents = append(week.TargetEvents, annualTrainingPlanTargetEvent{EventID: event.event.ID, Name: stringValue(event.event.Name), Type: stringValue(event.event.Type), StartDate: formatDate(event.startDate), LoadTarget: event.event.LoadTarget, TimeTargetSeconds: annualTrainingPlanTimeTarget(event.event), DistanceTargetMeters: event.event.DistanceTarget, Full: cloneJSONMap(event.event.Raw)})
		}
	}
	for _, note := range notes {
		for _, weekStart := range note.WeekStartDates {
			week := weeksByStart[weekStart]
			if week == nil {
				continue
			}
			week.NoteCount++
			if note.RecoveryHint {
				week.RecoveryNoteCount++
			}
			week.NoteIDs = append(week.NoteIDs, note.NoteID)
		}
	}
	weeks := make([]annualTrainingPlanWeek, 0, len(weeksByStart))
	for _, week := range weeksByStart {
		sort.Strings(week.PhaseIDs)
		sort.Strings(week.NoteIDs)
		weeks = append(weeks, *week)
	}
	sort.SliceStable(weeks, func(i, j int) bool { return weeks[i].WeekStartDate < weeks[j].WeekStartDate })
	caveats := []string{}
	for _, week := range weeks {
		if week.PartialWeek && week.TargetEventCount > 0 {
			caveats = append(caveats, "partial edge weeks with TARGET events are summarized but excluded from projection_bridge weekly_plan_targets")
			break
		}
	}
	return weeks, caveats
}

func annualTrainingPlanBridge(weeks []annualTrainingPlanWeek) annualTrainingPlanProjectionBridge {
	bridge := annualTrainingPlanProjectionBridge{TargetTool: getFitnessProjectionName, TargetArgument: "weekly_plan_targets", WeeklyPlanTargets: []annualTrainingPlanProjectionWeeklyTarget{}}
	for _, week := range weeks {
		if week.TargetEventCount == 0 {
			continue
		}
		if week.PartialWeek || week.MissingLoadTargetCount > 0 || week.LoadTarget == nil {
			bridge.ExcludedWeekCount++
			continue
		}
		bridge.WeeklyPlanTargets = append(bridge.WeeklyPlanTargets, annualTrainingPlanProjectionWeeklyTarget{WeekStartDate: week.WeekStartDate, TrainingLoad: round(*week.LoadTarget, 3)})
	}
	bridge.IncludedWeekCount = len(bridge.WeeklyPlanTargets)
	if bridge.ExcludedWeekCount > 0 {
		bridge.Caveats = append(bridge.Caveats, "some TARGET weeks were excluded from projection_bridge weekly_plan_targets because they were partial range-edge weeks or had missing load_target values")
	}
	return bridge
}

func annualTrainingPlanOverlappingPhaseIDs(phases []annualTrainingPlanPhase, start time.Time, end time.Time) []string {
	ids := []string{}
	for _, phase := range phases {
		phaseStart, startErr := time.Parse(time.DateOnly, phase.StartDate)
		phaseEnd, endErr := time.Parse(time.DateOnly, phase.EndDate)
		if startErr != nil || endErr != nil {
			continue
		}
		if !end.Before(phaseStart) && !start.After(phaseEnd) {
			ids = append(ids, phase.PhaseID)
		}
	}
	return ids
}

func annualTrainingPlanWeekStarts(start time.Time, end time.Time) []string {
	starts := []string{}
	for weekStart := isoWeekStart(start); !weekStart.After(end); weekStart = weekStart.AddDate(0, 0, 7) {
		starts = append(starts, formatDate(weekStart))
	}
	return starts
}

func annualTrainingPlanTimeTarget(event intervals.Event) *int {
	if event.TimeTarget != nil {
		return event.TimeTarget
	}
	return event.ElapsedTimeTarget
}

func annualTrainingPlanRecoveryHint(event annualTrainingPlanEvent) bool {
	parts := []string{stringValue(event.event.Name), stringValue(event.event.Type), stringValue(event.event.Description)}
	parts = append(parts, event.tags...)
	text := strings.ToLower(strings.Join(parts, " "))
	for _, needle := range []string{"recovery", "rest", "taper", "deload"} {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func annualTrainingPlanTags(event intervals.Event) []string {
	if tags := eventTags(event.Raw); tags != nil {
		return append([]string{}, (*tags)...)
	}
	return nil
}

func annualTrainingPlanEventDateOnly(value *string) (time.Time, bool) {
	text := strings.TrimSpace(stringValue(value))
	if len(text) >= len(time.DateOnly) {
		text = text[:len(time.DateOnly)]
	}
	if !validDate(text) {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.DateOnly, text)
	return parsed, err == nil
}

func isoWeekStart(value time.Time) time.Time {
	daysSinceMonday := (int(value.Weekday()) - int(time.Monday) + 7) % 7
	return value.AddDate(0, 0, -daysSinceMonday)
}

func minDate(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxDate(a time.Time, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func addFloatPtr(value *float64, add float64) *float64 {
	out := add
	if value != nil {
		out += *value
	}
	out = round(out, 3)
	return &out
}

func addIntPtr(value *int, add int) *int {
	out := add
	if value != nil {
		out += *value
	}
	return &out
}

func annualTrainingPlanRowID(prefix string, eventID string, index int) string {
	if strings.TrimSpace(eventID) != "" {
		return prefix + "_" + strings.TrimSpace(eventID)
	}
	return fmt.Sprintf("%s_%d", prefix, index)
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func getAnnualTrainingPlanInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"oldest", "newest"}, "properties": map[string]any{
		"oldest":       map[string]any{"type": "string", "description": "Required athlete-local start date YYYY-MM-DD for the ATP/periodization event scan; range is capped at 366 days."},
		"newest":       map[string]any{"type": "string", "description": "Required athlete-local end date YYYY-MM-DD for the ATP/periodization event scan; must be on or after oldest."},
		"calendar_id":  map[string]any{"type": "string", "description": "Optional upstream calendar ID filter when the athlete uses separate planning calendars."},
		"limit":        map[string]any{"type": "integer", "default": annualTrainingPlanEventLimit, "minimum": 1, "maximum": annualTrainingPlanEventLimit, "description": "Maximum raw calendar events to scan before filtering PLAN/TARGET/NOTE periodization rows; defaults to 500 and values above 500 are capped."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream PLAN/TARGET/NOTE event payloads only on the corresponding phase, target_event, and note rows. Default output is a compact summary."},
	}}
}

func getAnnualTrainingPlanOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Read-only annual training plan/periodization summary derived from existing PLAN, TARGET, and NOTE calendar events. Returns summary counts/totals, phases, ISO-week target totals, recovery/context notes, source/truncation caveats, and projection_bridge.weekly_plan_targets rows that can be copied to get_fitness_projection.weekly_plan_targets."}
}
