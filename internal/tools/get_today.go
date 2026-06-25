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
	getTodayName        = "get_today"
	getTodayDescription = "Answer how's today looking with one terse daily digest: today's CTL/ATL/TSB fitness, wellness, completed activities, planned events, and NOTE/race annotations. Reuses the same response shaping as get_fitness, get_wellness_data, get_activities, and get_events; include_full widens each section with raw upstream payloads."
	fetchTodayMessage   = "could not fetch today's digest; check intervals.icu credentials and athlete timezone"

	defaultTodayEventsLimit = 100
)

type todayClient interface {
	FitnessClient
	WellnessClient
	ActivitiesClient
	EventsClient
}

type getTodayRequest struct {
	IncludeFull bool `json:"include_full,omitempty"`
}

type getTodayResponse struct {
	Fitness             []fitnessRow       `json:"fitness"`
	Wellness            []map[string]any   `json:"wellness"`
	CompletedActivities []getActivitiesRow `json:"completed_activities"`
	PlannedEvents       []getEventsRow     `json:"planned_events"`
	Annotations         []getEventsRow     `json:"annotations"`
	Weather             getTodayWeather    `json:"weather"`
	Meta                getTodayMeta       `json:"_meta"`
}

type getTodayWeather struct {
	Status                        string `json:"status"`
	Summary                       string `json:"summary"`
	Provenance                    string `json:"provenance"`
	CompletedActivityWeatherCount int    `json:"completed_activity_weather_count,omitempty"`
}

type getTodayMeta struct {
	Date           string         `json:"date"`
	AsOf           string         `json:"as_of"`
	AsOfDate       string         `json:"as_of_date"`
	AsOfWeekday    string         `json:"as_of_weekday"`
	Timezone       string         `json:"timezone"`
	IncludeFull    bool           `json:"include_full"`
	SourceTools    []string       `json:"source_tools"`
	SectionCounts  map[string]int `json:"section_counts"`
	ActivityWindow string         `json:"activity_window"`
}

func newGetTodayTool(client todayClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return newGetTodayToolWithClock(client, profileClient, gearClient, gearCache, customFieldClient, customFieldCache, version, timezoneFallback, debugMetadata, time.Now, shapeCfg)
}

func newGetTodayToolWithClock(client todayClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getTodayName, Description: getTodayDescription, InputSchema: getTodayInputSchema(), OutputSchema: getTodayOutputSchema(), Handler: getTodayHandler(client, profileClient, gearClient, gearCache, customFieldClient, customFieldCache, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func getTodayHandler(client todayClient, profileClient ProfileClient, gearClient GearListClient, gearCache *gearListCache, customFieldClient ActivityCustomFieldClient, customFieldCache *customFieldCache, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		args, err := decodeGetTodayRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError("invalid get_today arguments; provide only optional include_full", err)
		}
		if client == nil || profileClient == nil {
			return Result{}, NewUserError(fetchTodayMessage, errors.New("missing today or profile client"))
		}
		profile, err := profileClient.GetAthleteProfile(ctx)
		if err != nil {
			if isContextError(err) || errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return Result{}, firstNonNilError(ctx.Err(), err)
			}
			return Result{}, NewUserError(fetchTodayMessage, err)
		}
		unitSystem := profileUnitSystem(profile)
		timezoneName := profileTimezone(profile.Timezone, timezoneFallback)
		asOf, err := response.AsOfMetadataInTimezone(now(), timezoneName)
		if err != nil {
			return Result{}, NewUserError(fetchTodayMessage, err)
		}
		today := asOf.AsOfDate

		fitnessRows, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: today, End: today})
		if err != nil {
			return todayFetchError(err)
		}
		fitnessRows = filterTodayFitnessRows(fitnessRows, today)
		wellness, err := client.ListWellness(ctx, intervals.WellnessParams{Oldest: today, Newest: today})
		if err != nil {
			return todayFetchError(err)
		}
		wellness = filterTodayWellnessRows(wellness, today)
		customFieldCodes := todayCustomFieldCodes(ctx, customFieldClient, customFieldCache)
		activities, _, err := fetchActivitiesPage(ctx, client, GetActivitiesRequest{Oldest: today, PageSize: defaultActivitiesPageSize, IncludeFull: args.IncludeFull}, nil, targetAthleteID(ctx), customFieldCodes)
		if err != nil {
			if errors.Is(err, errActivitiesPaginationBoundary) {
				return Result{}, NewUserError(activitiesPaginationBoundaryMessage, err)
			}
			return todayFetchError(err)
		}
		activities = filterTodayActivities(activities, today)
		if gearCache == nil {
			gearCache = newGearListCache()
		}
		gearResolutions, err := resolveActivityGear(ctx, gearClient, gearCache, activities)
		if err != nil {
			return Result{}, err
		}
		events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: today, Newest: today, Limit: defaultTodayEventsLimit})
		if err != nil {
			return todayFetchError(err)
		}
		events = filterTodayEvents(events, today)

		payload, err := shapeGetTodayResponse(todayDigestInputs{today: today, asOf: asOf, timezone: asOf.Timezone, includeFull: args.IncludeFull, unitSystem: unitSystem, profile: profile, fitnessRows: fitnessRows, wellnessRows: wellness, activities: activities, gearResolutions: gearResolutions, customFieldCodes: customFieldCodes, events: events})
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_today response: %w", err)
		}
		shaped, err := response.Shape(payload, shapeCfg.options(args.IncludeFull, []string{"fitness", "wellness", "completed_activities", "planned_events", "annotations"}, version, debugMetadata, getTodayName, unitSystem))
		if err != nil {
			return Result{}, fmt.Errorf("shaping get_today response: %w", err)
		}
		if _, err := json.Marshal(shaped); err != nil {
			return Result{}, fmt.Errorf("encoding get_today response: %w", err)
		}
		return TextResult(shaped), nil
	}
}

func decodeGetTodayRequest(raw json.RawMessage) (getTodayRequest, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		trimmed = []byte(`{}`)
	}
	if trimmed[0] != '{' {
		return getTodayRequest{}, errors.New("arguments must be a JSON object")
	}
	return DecodeStrict[getTodayRequest](trimmed)
}

type todayDigestInputs struct {
	today            string
	asOf             response.AsOfMetadata
	timezone         string
	includeFull      bool
	unitSystem       response.UnitSystem
	profile          intervals.AthleteWithSportSettings
	fitnessRows      []intervals.SummaryWithCats
	wellnessRows     []intervals.Wellness
	activities       []intervals.Activity
	gearResolutions  map[string]activityGearResolution
	customFieldCodes []string
	events           []intervals.Event
}

func shapeGetTodayResponse(in todayDigestInputs) (getTodayResponse, error) {
	completed := make([]getActivitiesRow, 0, len(in.activities))
	for _, activity := range in.activities {
		row := activityRow(activity, in.includeFull, in.timezone, in.unitSystem, in.gearResolutions[activity.ID], in.customFieldCodes)
		applyCompletedActivityWorkoutStatus(&row, activity)
		completed = append(completed, row)
	}
	planned, annotations, err := splitTodayEvents(in.events, in.includeFull, in.timezone, in.today, in.profile, in.unitSystem)
	if err != nil {
		return getTodayResponse{}, err
	}
	return getTodayResponse{
		Fitness:             shapeFitnessRows(in.fitnessRows, in.includeFull),
		Wellness:            wellnessRows(in.wellnessRows, in.includeFull),
		CompletedActivities: completed,
		PlannedEvents:       planned,
		Annotations:         annotations,
		Weather:             todayWeatherContext(completed),
		Meta: getTodayMeta{
			Date:           in.today,
			AsOf:           in.asOf.AsOf,
			AsOfDate:       in.asOf.AsOfDate,
			AsOfWeekday:    in.asOf.AsOfWeekday,
			Timezone:       in.asOf.Timezone,
			IncludeFull:    in.includeFull,
			SourceTools:    []string{getFitnessName, getWellnessDataName, getActivitiesName, getEventsName},
			SectionCounts:  map[string]int{"fitness": len(in.fitnessRows), "wellness": len(in.wellnessRows), "completed_activities": len(completed), "planned_events": len(planned), "annotations": len(annotations)},
			ActivityWindow: "from athlete-local midnight through upstream now",
		},
	}, nil
}

func todayWeatherContext(completed []getActivitiesRow) getTodayWeather {
	count := 0
	for _, activity := range completed {
		if activity.Weather != nil {
			count++
		}
	}
	if count > 0 {
		return getTodayWeather{
			Status:                        "completed_activity_weather_available",
			Summary:                       "Weather fields are available only for completed activities that Intervals.icu returned with historical activity weather; no forecast conditions are inferred for planned events.",
			Provenance:                    "intervals.icu activity historical weather fields via get_activities rows",
			CompletedActivityWeatherCount: count,
		}
	}
	return getTodayWeather{
		Status:     "forecast_unavailable",
		Summary:    "No weather conditions are available from the current Intervals.icu daily digest sources; do not infer or invent today's forecast.",
		Provenance: "upstream gap: supported get_today sources include fitness, wellness, activities, and events; current Intervals.icu forecast conditions are not exposed through the implemented read model",
	}
}

func filterTodayFitnessRows(rows []intervals.SummaryWithCats, today string) []intervals.SummaryWithCats {
	filtered := rows[:0]
	for _, row := range rows {
		if row.Date == today {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func filterTodayWellnessRows(rows []intervals.Wellness, today string) []intervals.Wellness {
	filtered := rows[:0]
	for _, row := range rows {
		if row.ID != nil && *row.ID == today {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func filterTodayActivities(activities []intervals.Activity, today string) []intervals.Activity {
	filtered := activities[:0]
	for _, activity := range activities {
		if localDateMatches(activity.StartDateLocal, today) {
			filtered = append(filtered, activity)
		}
	}
	return filtered
}

func filterTodayEvents(events []intervals.Event, today string) []intervals.Event {
	filtered := events[:0]
	for _, event := range events {
		if localDateMatches(event.StartDateLocal, today) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func localDateMatches(value *string, today string) bool {
	if value == nil {
		return false
	}
	date := strings.TrimSpace(*value)
	if len(date) >= len(today) {
		date = date[:len(today)]
	}
	return date == today
}

func splitTodayEvents(events []intervals.Event, includeFull bool, timezoneName string, asOfDate string, profile intervals.AthleteWithSportSettings, unitSystem response.UnitSystem) ([]getEventsRow, []getEventsRow, error) {
	planned := make([]getEventsRow, 0, len(events))
	annotations := make([]getEventsRow, 0, len(events))
	for _, event := range events {
		row, err := eventRow(event, includeFull, timezoneName, workoutPreviewContextForEvent(event, profile, unitSystem))
		if err != nil {
			return nil, nil, err
		}
		applyEventWorkoutStatus(&row, event, asOfDate)
		if isTodayAnnotation(row.Category) {
			annotations = append(annotations, row)
			continue
		}
		planned = append(planned, row)
	}
	sort.SliceStable(planned, func(i, j int) bool { return eventRowsBefore(planned[i], planned[j]) })
	sort.SliceStable(annotations, func(i, j int) bool { return eventRowsBefore(annotations[i], annotations[j]) })
	return planned, annotations, nil
}

func isTodayAnnotation(category string) bool {
	category = strings.ToUpper(strings.TrimSpace(category))
	return category == "NOTE" || category == "RACE" || strings.HasPrefix(category, "RACE_")
}

func eventRowsBefore(left, right getEventsRow) bool {
	if left.StartDateLocal != right.StartDateLocal {
		return left.StartDateLocal < right.StartDateLocal
	}
	return left.EventID < right.EventID
}

func todayCustomFieldCodes(_ context.Context, _ ActivityCustomFieldClient, _ *customFieldCache) []string {
	return nil
}

func targetAthleteID(ctx context.Context) string {
	athleteID, _ := intervals.TargetAthleteIDFromContext(ctx)
	return athleteID
}

func todayFetchError(err error) (Result, error) {
	if isContextError(err) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return Result{}, err
	}
	return Result{}, NewUserError(fetchTodayMessage, err)
}

func getTodayInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include raw upstream payloads under each digest section's rows. Defaults to false for a terse today digest."},
	}}
}

func getTodayOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "One-call athlete-local today digest with fitness, wellness, completed_activities, planned_events, annotations, explicit workout_status/workout_status_caveats on workout and completed-activity rows, explicit weather availability/provenance, athlete-local as-of metadata, source_tools, section counts, units, and scale labels. Terse by default; include_full adds raw upstream payloads per section."}
}
