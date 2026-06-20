package tools

import (
	"bytes"
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
	getPlanningContextName        = "get_planning_context"
	getPlanningContextDescription = "Fetch read-only weekly planning context without creating an ATP or calendar writes: athlete-local week events/workouts, active training-plan assignment summary, current fitness context, upcoming races, and caveats. This does not fill a calendar or call write/delete tools; include_full widens source read payloads only."
	fetchPlanningContextMessage   = "could not fetch planning context; check intervals.icu credentials, athlete ID, timezone, and date window"

	planningContextEventLimit      = 500
	planningContextRaceScanDays    = 84
	planningContextFitnessDays     = 7
	planningContextSeasonLookback  = 3
	planningContextSeasonLookahead = 1
	planningContextSeasonCategory  = "SEASON_START"
	planningContextScopeContext    = "context_only"
	planningContextReadOnlyCaveat  = "read_only_no_atp"
)

type planningContextClient interface {
	EventsClient
	TrainingPlanClient
	FitnessClient
}

type getPlanningContextRequest struct {
	WeekStart   string `json:"week_start,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type getPlanningContextResponse struct {
	Week          planningContextWeek       `json:"week"`
	WeekEvents    planningContextWeekEvents `json:"week_events"`
	TrainingPlan  getTrainingPlanResponse   `json:"training_plan"`
	Fitness       planningFitnessContext    `json:"fitness_context"`
	UpcomingRaces []getEventsRow            `json:"upcoming_races"`
	SeasonContext planningSeasonContext     `json:"season_context"`
	Caveats       []planningContextCaveat   `json:"caveats"`
	Meta          planningContextMeta       `json:"_meta"`
}

type planningContextWeek struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Timezone  string `json:"timezone"`
	Anchor    string `json:"anchor"`
}

type planningContextWeekEvents struct {
	PlannedWorkouts []getEventsRow `json:"planned_workouts"`
	Races           []getEventsRow `json:"races"`
	Notes           []getEventsRow `json:"notes"`
	OtherEvents     []getEventsRow `json:"other_events"`
}

type planningFitnessContext struct {
	Current *fitnessRow  `json:"current,omitempty"`
	Rows    []fitnessRow `json:"rows"`
}

type planningSeasonContext struct {
	Seasons []planningSeasonRow `json:"seasons"`
	Meta    planningSeasonMeta  `json:"_meta"`
}

type planningSeasonRow struct {
	EventID     string         `json:"event_id,omitempty"`
	StartDate   string         `json:"start_date,omitempty"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Full        map[string]any `json:"full,omitempty"`
}

type planningSeasonMeta struct {
	Category              string          `json:"category"`
	Window                dateRangeMeta   `json:"window"`
	Chunks                []dateRangeMeta `json:"chunks"`
	ChunkCount            int             `json:"chunk_count"`
	MaxChunkDays          int             `json:"max_chunk_days"`
	EventLimitPerChunk    int             `json:"event_limit_per_chunk"`
	Count                 int             `json:"count"`
	SeasonsMayBeTruncated bool            `json:"seasons_may_be_truncated"`
}

type planningContextCaveat struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type planningContextMeta struct {
	SourceTools     []string                   `json:"source_tools"`
	AsOf            string                     `json:"as_of"`
	AsOfDate        string                     `json:"as_of_date"`
	AsOfWeekday     string                     `json:"as_of_weekday"`
	Timezone        string                     `json:"timezone"`
	IncludeFull     bool                       `json:"include_full"`
	ReadOnly        bool                       `json:"read_only"`
	WritesPerformed bool                       `json:"writes_performed"`
	PlanningScope   string                     `json:"planning_scope"`
	WeekWindow      dateRangeMeta              `json:"week_window"`
	FitnessWindow   dateRangeMeta              `json:"fitness_window"`
	RaceWindow      dateRangeMeta              `json:"race_window"`
	EventLimits     planningContextEventLimits `json:"event_limits"`
	Truncation      planningContextTruncation  `json:"truncation"`
	SectionCounts   map[string]int             `json:"section_counts"`
	CaveatCodes     []string                   `json:"caveat_codes"`
}

type planningContextEventLimits struct {
	WeekEvents int `json:"week_events"`
	RaceScan   int `json:"race_scan"`
}

type planningContextTruncation struct {
	WeekEventsMayBeTruncated bool `json:"week_events_may_be_truncated"`
	RaceScanMayBeTruncated   bool `json:"race_scan_may_be_truncated"`
}

type planningContextInputs struct {
	weekStart       time.Time
	weekEnd         time.Time
	weekAnchor      string
	fitnessStart    time.Time
	fitnessEnd      time.Time
	raceStart       time.Time
	raceEnd         time.Time
	seasonStart     time.Time
	seasonEnd       time.Time
	seasonChunks    []dateRangeMeta
	seasonTruncated bool
	asOf            response.AsOfMetadata
	timezone        string
	includeFull     bool
	unitSystem      response.UnitSystem
	weekEvents      []intervals.Event
	raceEvents      []intervals.Event
	seasonEvents    []intervals.Event
	trainingPlan    intervals.TrainingPlan
	fitnessRows     []intervals.SummaryWithCats
}

func newGetPlanningContextTool(client planningContextClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return newGetPlanningContextToolWithClock(client, profileClient, version, timezoneFallback, debugMetadata, time.Now, shapeCfg)
}

func newGetPlanningContextToolWithClock(client planningContextClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: getPlanningContextName, Description: getPlanningContextDescription, InputSchema: getPlanningContextInputSchema(), OutputSchema: getPlanningContextOutputSchema(), Handler: getPlanningContextHandler(client, profileClient, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func getPlanningContextHandler(client planningContextClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetPlanningContextRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError("invalid get_planning_context arguments; provide optional week_start as YYYY-MM-DD and include_full boolean", err)
		}
		if client == nil || profileClient == nil {
			return Result{}, NewUserError(fetchPlanningContextMessage, errors.New("missing planning context or profile client"))
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchPlanningContextMessage, err)
		}
		asOf, err := response.AsOfMetadataInTimezone(now(), timezoneName)
		if err != nil {
			return Result{}, NewUserError(fetchPlanningContextMessage, err)
		}
		asOfDate, err := time.Parse(time.DateOnly, asOf.AsOfDate)
		if err != nil {
			return Result{}, NewUserError(fetchPlanningContextMessage, err)
		}
		weekStart, weekAnchor, err := planningWeekStart(args.WeekStart, asOfDate)
		if err != nil {
			return Result{}, NewUserError("invalid get_planning_context arguments; week_start must be YYYY-MM-DD", err)
		}
		weekEnd := weekStart.AddDate(0, 0, 6)
		fitnessStart := asOfDate.AddDate(0, 0, -(planningContextFitnessDays - 1))
		fitnessEnd := asOfDate
		raceStart := asOfDate
		raceEnd := asOfDate.AddDate(0, 0, planningContextRaceScanDays)
		seasonStart := asOfDate.AddDate(-planningContextSeasonLookback, 0, 0)
		seasonEnd := asOfDate.AddDate(planningContextSeasonLookahead, 0, 0)

		weekEvents, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: formatDate(weekStart), Newest: formatDate(weekEnd), Limit: planningContextEventLimit})
		if err != nil {
			return planningContextFetchError(err)
		}
		raceEvents, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: formatDate(raceStart), Newest: formatDate(raceEnd), Limit: planningContextEventLimit})
		if err != nil {
			return planningContextFetchError(err)
		}
		seasonEvents, seasonChunks, seasonTruncated, err := listPlanningSeasonEvents(ctx, client, seasonStart, seasonEnd)
		if err != nil {
			return planningContextFetchError(err)
		}
		plan, err := client.GetTrainingPlan(ctx)
		if err != nil {
			return planningContextFetchError(err)
		}
		fitnessRows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: formatDate(fitnessStart), End: formatDate(fitnessEnd)})
		if err != nil {
			return planningContextFetchError(err)
		}

		payload, err := shapeGetPlanningContextResponse(planningContextInputs{weekStart: weekStart, weekEnd: weekEnd, weekAnchor: weekAnchor, fitnessStart: fitnessStart, fitnessEnd: fitnessEnd, raceStart: raceStart, raceEnd: raceEnd, seasonStart: seasonStart, seasonEnd: seasonEnd, seasonChunks: seasonChunks, seasonTruncated: seasonTruncated, asOf: asOf, timezone: timezoneName, includeFull: args.IncludeFull, unitSystem: unitSystem, weekEvents: weekEvents, raceEvents: raceEvents, seasonEvents: seasonEvents, trainingPlan: plan, fitnessRows: fitnessRows})
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_planning_context response: %w", err)
		}
		return encodeShaped(payload, args.IncludeFull, nil, version, debugMetadata, getPlanningContextName, unitSystem, shapeCfg)
	}
}

func decodeGetPlanningContextRequest(raw json.RawMessage) (getPlanningContextRequest, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		trimmed = []byte(`{}`)
	}
	if trimmed[0] != '{' {
		return getPlanningContextRequest{}, errors.New("arguments must be a JSON object")
	}
	args, err := DecodeStrict[getPlanningContextRequest](trimmed)
	if err != nil {
		return getPlanningContextRequest{}, err
	}
	args.WeekStart = strings.TrimSpace(args.WeekStart)
	return args, nil
}

func planningWeekStart(value string, asOfDate time.Time) (time.Time, string, error) {
	if strings.TrimSpace(value) == "" {
		daysUntilMonday := (int(time.Monday) - int(asOfDate.Weekday()) + 7) % 7
		return asOfDate.AddDate(0, 0, daysUntilMonday), "upcoming_monday", nil
	}
	parsed, err := time.Parse(time.DateOnly, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, "", err
	}
	daysSinceMonday := (int(parsed.Weekday()) - int(time.Monday) + 7) % 7
	return parsed.AddDate(0, 0, -daysSinceMonday), "supplied_week_start_normalized_to_monday", nil
}

func shapeGetPlanningContextResponse(in planningContextInputs) (getPlanningContextResponse, error) {
	weekEvents, err := classifyPlanningWeekEvents(in.weekEvents, in.includeFull, in.timezone)
	if err != nil {
		return getPlanningContextResponse{}, err
	}
	upcomingRaces, err := planningRaceRows(in.raceEvents, in.includeFull, in.timezone)
	if err != nil {
		return getPlanningContextResponse{}, err
	}
	fitnessRows := shapeFitnessRows(in.fitnessRows, in.includeFull)
	currentFitness := latestFitnessRow(fitnessRows, in.asOf.AsOfDate)
	seasonContext := shapePlanningSeasonContext(in.seasonEvents, in.includeFull, in.seasonStart, in.seasonEnd, in.seasonChunks, in.seasonTruncated)
	trainingPlan := shapeGetTrainingPlanResponse(in.trainingPlan, in.includeFull, in.timezone)
	weekTruncated := len(in.weekEvents) >= planningContextEventLimit
	raceTruncated := len(in.raceEvents) >= planningContextEventLimit
	caveats := planningContextCaveats(weekEvents, trainingPlan, fitnessRows, upcomingRaces, weekTruncated, raceTruncated)
	caveatCodes := make([]string, 0, len(caveats))
	for _, caveat := range caveats {
		caveatCodes = append(caveatCodes, caveat.Code)
	}

	return getPlanningContextResponse{
		Week:          planningContextWeek{StartDate: formatDate(in.weekStart), EndDate: formatDate(in.weekEnd), Timezone: in.timezone, Anchor: in.weekAnchor},
		WeekEvents:    weekEvents,
		TrainingPlan:  trainingPlan,
		Fitness:       planningFitnessContext{Current: currentFitness, Rows: fitnessRows},
		UpcomingRaces: upcomingRaces,
		SeasonContext: seasonContext,
		Caveats:       caveats,
		Meta: planningContextMeta{
			SourceTools:     []string{getAthleteProfileName, getEventsName, getTrainingPlanName, getFitnessName},
			AsOf:            in.asOf.AsOf,
			AsOfDate:        in.asOf.AsOfDate,
			AsOfWeekday:     in.asOf.AsOfWeekday,
			Timezone:        in.asOf.Timezone,
			IncludeFull:     in.includeFull,
			ReadOnly:        true,
			WritesPerformed: false,
			PlanningScope:   planningContextScopeContext,
			WeekWindow:      dateRangeMeta{Oldest: formatDate(in.weekStart), Newest: formatDate(in.weekEnd)},
			FitnessWindow:   dateRangeMeta{Oldest: formatDate(in.fitnessStart), Newest: formatDate(in.fitnessEnd)},
			RaceWindow:      dateRangeMeta{Oldest: formatDate(in.raceStart), Newest: formatDate(in.raceEnd)},
			EventLimits:     planningContextEventLimits{WeekEvents: planningContextEventLimit, RaceScan: planningContextEventLimit},
			Truncation:      planningContextTruncation{WeekEventsMayBeTruncated: weekTruncated, RaceScanMayBeTruncated: raceTruncated},
			SectionCounts:   map[string]int{"planned_workouts": len(weekEvents.PlannedWorkouts), "races": len(weekEvents.Races), "notes": len(weekEvents.Notes), "other_events": len(weekEvents.OtherEvents), "fitness_rows": len(fitnessRows), "upcoming_races": len(upcomingRaces), "seasons": len(seasonContext.Seasons)},
			CaveatCodes:     caveatCodes,
		},
	}, nil
}

func classifyPlanningWeekEvents(events []intervals.Event, includeFull bool, timezoneName string) (planningContextWeekEvents, error) {
	out := planningContextWeekEvents{PlannedWorkouts: []getEventsRow{}, Races: []getEventsRow{}, Notes: []getEventsRow{}, OtherEvents: []getEventsRow{}}
	for _, event := range events {
		row, err := eventRow(event, includeFull, timezoneName)
		if err != nil {
			return out, err
		}
		switch planningEventClass(row.Category) {
		case "workout":
			out.PlannedWorkouts = append(out.PlannedWorkouts, row)
		case "race":
			out.Races = append(out.Races, row)
		case "note":
			out.Notes = append(out.Notes, row)
		default:
			out.OtherEvents = append(out.OtherEvents, row)
		}
	}
	sortEventRows(out.PlannedWorkouts)
	sortEventRows(out.Races)
	sortEventRows(out.Notes)
	sortEventRows(out.OtherEvents)
	return out, nil
}

func planningRaceRows(events []intervals.Event, includeFull bool, timezoneName string) ([]getEventsRow, error) {
	out := []getEventsRow{}
	for _, event := range events {
		row, err := eventRow(event, includeFull, timezoneName)
		if err != nil {
			return nil, err
		}
		if planningEventClass(row.Category) == "race" {
			out = append(out, row)
		}
	}
	sortEventRows(out)
	return out, nil
}

func listPlanningSeasonEvents(ctx context.Context, client EventsClient, start time.Time, end time.Time) ([]intervals.Event, []dateRangeMeta, bool, error) {
	chunks := planningSeasonChunks(start, end)
	events := []intervals.Event{}
	truncated := false
	for _, chunk := range chunks {
		chunkEvents, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: chunk.Oldest, Newest: chunk.Newest, Category: planningContextSeasonCategory, Limit: planningContextEventLimit})
		if err != nil {
			return nil, chunks, false, err
		}
		if len(chunkEvents) >= planningContextEventLimit {
			truncated = true
		}
		events = append(events, chunkEvents...)
	}
	return events, chunks, truncated, nil
}

func planningSeasonChunks(start time.Time, end time.Time) []dateRangeMeta {
	if end.Before(start) {
		return nil
	}
	chunks := []dateRangeMeta{}
	for chunkStart := start; !chunkStart.After(end); {
		chunkEnd := chunkStart.AddDate(0, 0, maxEventsRangeDays-1)
		if chunkEnd.After(end) {
			chunkEnd = end
		}
		chunks = append(chunks, dateRangeMeta{Oldest: formatDate(chunkStart), Newest: formatDate(chunkEnd)})
		chunkStart = chunkEnd.AddDate(0, 0, 1)
	}
	return chunks
}

func shapePlanningSeasonContext(events []intervals.Event, includeFull bool, start time.Time, end time.Time, chunks []dateRangeMeta, truncated bool) planningSeasonContext {
	rows := []planningSeasonRow{}
	for _, event := range events {
		if !isSeasonStartEvent(event) {
			continue
		}
		row := planningSeasonRow{EventID: event.ID, StartDate: stringValue(event.StartDateLocal), Name: stringValue(event.Name), Description: stringValue(event.Description)}
		if includeFull {
			row.Full = cloneJSONMap(event.Raw)
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].StartDate != rows[j].StartDate {
			return rows[i].StartDate < rows[j].StartDate
		}
		return rows[i].EventID < rows[j].EventID
	})
	return planningSeasonContext{Seasons: rows, Meta: planningSeasonMeta{Category: planningContextSeasonCategory, Window: dateRangeMeta{Oldest: formatDate(start), Newest: formatDate(end)}, Chunks: chunks, ChunkCount: len(chunks), MaxChunkDays: maxEventsRangeDays, EventLimitPerChunk: planningContextEventLimit, Count: len(rows), SeasonsMayBeTruncated: truncated}}
}

func isSeasonStartEvent(event intervals.Event) bool {
	category := firstNonEmpty(stringValue(event.Category), anyString(event.Raw["category"]))
	return strings.EqualFold(strings.TrimSpace(category), planningContextSeasonCategory)
}

func planningEventClass(category string) string {
	switch normalized := strings.ToUpper(strings.TrimSpace(category)); {
	case normalized == "WORKOUT":
		return "workout"
	case normalized == "RACE" || strings.HasPrefix(normalized, "RACE_"):
		return "race"
	case normalized == "NOTE":
		return "note"
	default:
		return "other"
	}
}

func sortEventRows(rows []getEventsRow) {
	sort.SliceStable(rows, func(i, j int) bool { return eventRowsBefore(rows[i], rows[j]) })
}

func latestFitnessRow(rows []fitnessRow, asOfDate string) *fitnessRow {
	var current *fitnessRow
	for i := range rows {
		if rows[i].Date > asOfDate {
			continue
		}
		current = &rows[i]
	}
	return current
}

func planningContextCaveats(events planningContextWeekEvents, plan getTrainingPlanResponse, fitnessRows []fitnessRow, upcomingRaces []getEventsRow, weekTruncated bool, raceTruncated bool) []planningContextCaveat {
	caveats := []planningContextCaveat{{Code: planningContextReadOnlyCaveat, Message: "read-only context only; this tool does not create ATP notes, fill calendars, or perform writes"}}
	weekEventCount := len(events.PlannedWorkouts) + len(events.Races) + len(events.Notes) + len(events.OtherEvents)
	if weekEventCount == 0 {
		caveats = append(caveats, planningContextCaveat{Code: "no_week_events", Message: "no calendar events were returned for the planning week"})
	}
	if len(events.PlannedWorkouts) == 0 {
		caveats = append(caveats, planningContextCaveat{Code: "no_week_workouts", Message: "no WORKOUT events were returned for the planning week"})
	}
	if plan.Unavailable != nil && plan.Unavailable.Reason == "no_active_training_plan" {
		caveats = append(caveats, planningContextCaveat{Code: "no_active_training_plan", Message: "no active training-plan assignment was returned"})
	}
	if plan.TrainingPlan != nil && plan.TrainingPlan.PlanSummary == nil {
		caveats = append(caveats, planningContextCaveat{Code: "partial_training_plan_summary", Message: "active training-plan assignment lacks nested plan summary fields"})
	}
	if len(fitnessRows) == 0 {
		caveats = append(caveats, planningContextCaveat{Code: "no_fitness_rows", Message: "no fitness rows were returned for the current fitness window"})
	}
	if len(upcomingRaces) == 0 {
		caveats = append(caveats, planningContextCaveat{Code: "no_upcoming_races", Message: "no RACE or RACE_* events were returned in the upcoming race window"})
	}
	if weekTruncated {
		caveats = append(caveats, planningContextCaveat{Code: "week_events_may_be_truncated", Message: "week event scan reached the upstream limit; additional week events may exist"})
	}
	if raceTruncated {
		caveats = append(caveats, planningContextCaveat{Code: "upcoming_races_may_be_truncated", Message: "race scan reached the upstream limit; additional race events may exist"})
	}
	return caveats
}

func formatDate(value time.Time) string {
	return value.Format(time.DateOnly)
}

func planningContextFetchError(err error) (Result, error) {
	if isContextError(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return Result{}, err
	}
	return Result{}, NewUserError(fetchPlanningContextMessage, err)
}

func getPlanningContextInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"week_start":   map[string]any{"type": "string", "description": "Optional athlete-local YYYY-MM-DD date anchoring the planning week. The supplied date is normalized backward to Monday; omit to use the upcoming athlete-local Monday, with today used when today is Monday."},
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw source read payloads only: event full fields, fitness row full fields, and raw active training-plan assignment/nested payloads. This does not enable ATP creation or calendar writes."},
	}}
}

func getPlanningContextOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Read-only planning context with week_events, training_plan, current fitness_context, upcoming_races, season_context, caveats, and _meta planning_scope=context_only/read_only=true/writes_performed=false. It summarizes existing source reads, includes SEASON_START season-boundary context, and never creates ATP notes or calendar items."}
}
