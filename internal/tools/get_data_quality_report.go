package tools

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/athleteprofile"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	getDataQualityReportName                    = "get_data_quality_report"
	getDataQualityReportDescription             = "Report whether icuvisor can see enough activities, streams, load, wellness, thresholds, and calendar data to answer common coaching questions, with actionable diagnostics for Strava restrictions, HR/TRIMP-only load, stale wellness, missing thresholds, and sparse history."
	invalidGetDataQualityReportArgumentsMessage = "invalid get_data_quality_report arguments; provide start_date/end_date as YYYY-MM-DD, optional sport, and optional include_full"
	fetchDataQualityReportMessage               = "could not fetch data quality inputs; check intervals.icu credentials, athlete ID, and date range"

	dataQualityActivityFetchLimit = maxActivityFetchLimit
	dataQualityEventFetchLimit    = maxEventsLimit
)

type dataQualityReportClient interface {
	ProfileClient
	ActivitiesClient
	FitnessClient
	WellnessClient
	EventsClient
}

type getDataQualityReportRequest struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Sport       string `json:"sport,omitempty"`
	IncludeFull bool   `json:"include_full,omitempty"`
}

type dataQualityReportResponse struct {
	Summary     dataQualitySummary      `json:"summary"`
	Sections    dataQualitySections     `json:"sections"`
	Diagnostics []dataQualityDiagnostic `json:"diagnostics,omitempty"`
	Full        *dataQualityFull        `json:"full,omitempty"`
	Meta        dataQualityMeta         `json:"_meta"`
}

type dataQualitySummary struct {
	Status                 string   `json:"status"`
	WorstSeverity          string   `json:"worst_severity"`
	WindowDays             int      `json:"window_days"`
	SectionsOK             int      `json:"sections_ok"`
	SectionsWarning        int      `json:"sections_warning"`
	SectionsCritical       int      `json:"sections_critical"`
	PrimaryRecommendations []string `json:"primary_recommendations,omitempty"`
}

type dataQualitySections struct {
	ActivityCoverage   dataQualitySection `json:"activity_coverage"`
	StreamAvailability dataQualitySection `json:"stream_availability"`
	SourceRestrictions dataQualitySection `json:"source_restrictions"`
	LoadBasis          dataQualitySection `json:"load_basis"`
	ThresholdsZones    dataQualitySection `json:"thresholds_zones"`
	WellnessFreshness  dataQualitySection `json:"wellness_freshness"`
	CalendarRaceData   dataQualitySection `json:"calendar_race_data"`
}

type dataQualitySection struct {
	Status      string                  `json:"status"`
	Severity    string                  `json:"severity"`
	Message     string                  `json:"message"`
	Evidence    map[string]any          `json:"evidence,omitempty"`
	Diagnostics []dataQualityDiagnostic `json:"diagnostics,omitempty"`
}

type dataQualityDiagnostic struct {
	Code           string                    `json:"code"`
	Severity       string                    `json:"severity"`
	Message        string                    `json:"message"`
	Evidence       map[string]any            `json:"evidence,omitempty"`
	Recommendation dataQualityRecommendation `json:"recommendation"`
}

type dataQualityRecommendation struct {
	Action     string   `json:"action"`
	Tool       string   `json:"tool,omitempty"`
	Fields     []string `json:"fields,omitempty"`
	Workaround string   `json:"workaround,omitempty"`
}

type dataQualityMeta struct {
	ServerVersion string `json:"server_version"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Timezone      string `json:"timezone"`
	Sport         string `json:"sport,omitempty"`
	IncludeFull   bool   `json:"include_full"`
	ReadOnly      bool   `json:"read_only"`
	StreamPolicy  string `json:"stream_policy"`
	ActivityLimit int    `json:"activity_limit"`
	EventLimit    int    `json:"event_limit"`
}

type dataQualityFull struct {
	ActivitySamples []dataQualityActivityEvidence `json:"activity_samples,omitempty"`
	SummaryDates    []string                      `json:"summary_dates,omitempty"`
	WellnessDates   []string                      `json:"wellness_dates,omitempty"`
	CalendarEvents  []dataQualityEventEvidence    `json:"calendar_events,omitempty"`
}

type dataQualityActivityEvidence struct {
	ActivityID     string   `json:"activity_id"`
	Date           string   `json:"date,omitempty"`
	Sport          string   `json:"sport,omitempty"`
	Source         string   `json:"source,omitempty"`
	StreamTypes    []string `json:"stream_types,omitempty"`
	StravaImported bool     `json:"strava_imported,omitempty"`
}

type dataQualityEventEvidence struct {
	EventID  string `json:"event_id,omitempty"`
	Date     string `json:"date,omitempty"`
	Category string `json:"category,omitempty"`
	Type     string `json:"type,omitempty"`
	Name     string `json:"name,omitempty"`
	RaceLike bool   `json:"race_like,omitempty"`
}

func newGetDataQualityReportTool(client dataQualityReportClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	shapeCfg := responseShapingOrDefault(shaping)
	return coreTool(Tool{Name: getDataQualityReportName, Description: getDataQualityReportDescription, InputSchema: dataQualityReportInputSchema(), OutputSchema: dataQualityReportOutputSchema(), Handler: getDataQualityReportHandler(client, version, timezoneFallback, debugMetadata, shapeCfg)})
}

func getDataQualityReportHandler(client dataQualityReportClient, version string, timezoneFallback string, debugMetadata bool, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeGetDataQualityReportRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidGetDataQualityReportArgumentsMessage, err)
		}
		profile, err := client.GetAthleteProfile(ctx)
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchDataQualityReportMessage, err)
		}
		activities, err := client.ListActivities(ctx, intervals.ListActivitiesParams{Oldest: args.StartDate, Newest: args.EndDate, Limit: dataQualityActivityFetchLimit, Fields: dataQualityActivityFields()})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchDataQualityReportMessage, err)
		}
		summaries, err := client.ListAthleteSummary(ctx, intervals.AthleteSummaryParams{Start: args.StartDate, End: args.EndDate})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchDataQualityReportMessage, err)
		}
		wellness, err := client.ListWellness(ctx, intervals.WellnessParams{Oldest: args.StartDate, Newest: args.EndDate})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchDataQualityReportMessage, err)
		}
		resolve := true
		events, err := client.ListEvents(ctx, intervals.ListEventsParams{Oldest: args.StartDate, Newest: args.EndDate, Limit: dataQualityEventFetchLimit, Resolve: &resolve})
		if err != nil {
			if isContextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchDataQualityReportMessage, err)
		}
		activities = filterDataQualityActivities(activities, args.Sport)
		summaries = filterDataQualitySummaries(summaries, args.Sport)
		payload := shapeDataQualityReport(dataQualityReportInputs{args: args, profile: profile, activities: activities, summaries: summaries, wellness: wellness, events: events, version: version, timezoneFallback: timezoneFallback})
		shaped, err := response.Shape(payload, shapeCfg.options(args.IncludeFull, nil, version, debugMetadata, getDataQualityReportName, profileUnitSystem(profile)))
		if err != nil {
			return Result{}, err
		}
		return TextResult(shaped), nil
	}
}

type dataQualityReportInputs struct {
	args             getDataQualityReportRequest
	profile          intervals.AthleteWithSportSettings
	activities       []intervals.Activity
	summaries        []intervals.SummaryWithCats
	wellness         []intervals.Wellness
	events           []intervals.Event
	version          string
	timezoneFallback string
}

func decodeGetDataQualityReportRequest(raw json.RawMessage) (getDataQualityReportRequest, error) {
	var args getDataQualityReportRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[getDataQualityReportRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.Sport = strings.TrimSpace(args.Sport)
	if !validDate(args.StartDate) || !validDate(args.EndDate) {
		return args, errors.New("start_date and end_date must be YYYY-MM-DD")
	}
	if args.EndDate < args.StartDate {
		return args, errors.New("end_date must be on or after start_date")
	}
	return args, nil
}

func shapeDataQualityReport(in dataQualityReportInputs) dataQualityReportResponse {
	windowDays := dateCount(in.args.StartDate, in.args.EndDate)
	profileResponse := newGetAthleteProfileResponse(in.profile, in.version, in.timezoneFallback)
	sections := dataQualitySections{
		ActivityCoverage:   dataQualityActivityCoverageSection(in.activities, windowDays),
		StreamAvailability: dataQualityStreamAvailabilitySection(in.activities),
		SourceRestrictions: dataQualitySourceRestrictionsSection(in.activities),
		LoadBasis:          dataQualityLoadBasisSection(in.summaries, in.args.Sport),
		ThresholdsZones:    dataQualityThresholdsSection(profileResponse.Meta.Warnings),
		WellnessFreshness:  dataQualityWellnessSection(in.wellness, in.args.EndDate),
		CalendarRaceData:   dataQualityCalendarSection(in.events),
	}
	payload := dataQualityReportResponse{Sections: sections, Meta: dataQualityMeta{ServerVersion: normalizeVersion(in.version), StartDate: in.args.StartDate, EndDate: in.args.EndDate, Timezone: profileTimezone(in.profile.Timezone, in.timezoneFallback), Sport: in.args.Sport, IncludeFull: in.args.IncludeFull, ReadOnly: true, StreamPolicy: "uses activity stream_types and summary fields only; does not fetch raw stream samples by default", ActivityLimit: dataQualityActivityFetchLimit, EventLimit: dataQualityEventFetchLimit}}
	payload.Diagnostics = collectDataQualityDiagnostics(sections)
	payload.Summary = summarizeDataQuality(sections, windowDays)
	if in.args.IncludeFull {
		payload.Full = &dataQualityFull{ActivitySamples: dataQualityActivitySamples(in.activities), SummaryDates: dataQualitySummaryDates(in.summaries), WellnessDates: dataQualityWellnessDates(in.wellness), CalendarEvents: dataQualityEventSamples(in.events)}
	}
	return payload
}

func dataQualityActivityCoverageSection(activities []intervals.Activity, windowDays int) dataQualitySection {
	evidence := map[string]any{"activity_count": len(activities), "active_days": len(dataQualityActivityDays(activities)), "window_days": windowDays}
	if len(activities) > 0 {
		evidence["oldest_activity_date"] = dataQualityActivityDate(activities[len(activities)-1])
		evidence["newest_activity_date"] = dataQualityActivityDate(activities[0])
	}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "Activity history is visible for this window.", Evidence: evidence}
	if len(activities) == 0 {
		section = dataQualitySection{Status: "critical", Severity: "critical", Message: "No completed activities are visible in this date window.", Evidence: evidence, Diagnostics: []dataQualityDiagnostic{{Code: "no_visible_activities", Severity: "critical", Message: "No completed activities were returned for the requested window.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Confirm the date range, athlete selection, and upstream sync status, then retry get_activities for the same window.", Tool: getActivitiesName}}}}
	} else if sparseActivityHistory(len(dataQualityActivityDays(activities)), windowDays) {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "Activity history is visible but sparse for reliable coaching context."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "sparse_activity_history", Severity: "warning", Message: "Few activity days are visible in the selected window, so trend or readiness answers may lack context.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Use a longer date range or wait for more synced activities before asking for trend-heavy coaching analysis.", Tool: getActivitiesName}})
	}
	if len(activities) >= dataQualityActivityFetchLimit {
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "activity_probe_truncated", Severity: "info", Message: "The activity probe reached its safety limit; counts are lower bounds.", Evidence: map[string]any{"limit": dataQualityActivityFetchLimit}, Recommendation: dataQualityRecommendation{Action: "Use get_activities pagination for an exact activity inventory if needed.", Tool: getActivitiesName}})
	}
	return section
}

func dataQualityStreamAvailabilitySection(activities []intervals.Activity) dataQualitySection {
	withStreams := 0
	missingStreams := 0
	streamSet := map[string]bool{}
	for _, activity := range activities {
		if len(activity.StreamTypes) == 0 {
			missingStreams++
			continue
		}
		withStreams++
		for _, streamType := range activity.StreamTypes {
			if trimmed := strings.TrimSpace(streamType); trimmed != "" {
				streamSet[trimmed] = true
			}
		}
	}
	evidence := map[string]any{"activities_checked": len(activities), "activities_with_stream_types": withStreams, "activities_missing_stream_types": missingStreams, "available_stream_types": sortedBoolKeys(streamSet)}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "At least some activities expose stream type metadata.", Evidence: evidence}
	if len(activities) == 0 {
		section.Status = "unknown"
		section.Message = "Stream availability cannot be assessed because no activities are visible."
		return section
	}
	if withStreams == 0 {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "No activities in this window expose stream type metadata."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "missing_stream_metadata", Severity: "warning", Message: "Icuvisor can see activity summaries, but no stream channels are advertised for this window.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Use activity summary fields when possible; for stream-dependent questions, inspect a specific activity with get_activity_streams or re-import restricted source activities directly from the native provider.", Tool: getActivityStreamsName}})
	} else if missingStreams > 0 {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "Some activities lack stream type metadata."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "partial_stream_metadata", Severity: "warning", Message: "Some visible activities do not advertise stream channels, so stream-dependent answers may be incomplete.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Resolve important workouts with get_activity_details, then inspect streams only for the specific activity IDs that need sample-level analysis.", Tool: getActivityStreamsName}})
	}
	return section
}

func dataQualitySourceRestrictionsSection(activities []intervals.Activity) dataQualitySection {
	restricted := []dataQualityActivityEvidence{}
	for _, activity := range activities {
		if isStravaBlocked(activity) {
			restricted = append(restricted, dataQualityActivityEvidence{ActivityID: activity.ID, Date: dataQualityActivityDate(activity), Sport: dataQualityActivitySport(activity), Source: stringValue(activity.Source), StravaImported: true})
		}
	}
	evidence := map[string]any{"activities_checked": len(activities), "restricted_source_count": len(restricted)}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "No Strava-restricted activity stubs were found in this window.", Evidence: evidence}
	if len(restricted) > 0 {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "Some activity summaries are visible only as Strava-restricted stubs."
		for _, item := range restricted {
			workaround := stravaBlockedWorkaround(map[string]any{"external_id": ""})
			if activity, ok := activityByID(activities, item.ActivityID); ok {
				workaround = stravaBlockedWorkaround(activity.Raw)
			}
			section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "restricted_source", Severity: "warning", Message: "A Strava-imported activity summary is visible, but detailed streams, intervals, and max-heart-rate samples may be unavailable through the API.", Evidence: map[string]any{"activity_id": item.ActivityID, "date": item.Date, "source": item.Source}, Recommendation: dataQualityRecommendation{Action: "Re-import historical data directly from the native provider where possible.", Tool: getActivitiesName, Workaround: workaround}})
		}
	}
	return section
}

func dataQualityLoadBasisSection(rows []intervals.SummaryWithCats, sport string) dataQualitySection {
	evidence := map[string]any{"summary_days": len(rows), "dates": dataQualitySummaryDates(rows)}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "Training-load and fitness summary fields are visible.", Evidence: evidence}
	if len(rows) == 0 {
		section.Status = "critical"
		section.Severity = "critical"
		section.Message = "No athlete-summary rows are visible for load and fitness diagnostics."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "missing_load_history", Severity: "critical", Message: "No fitness/load rows were returned for the requested window.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Check upstream sync and use get_fitness or get_training_summary for the same date range.", Tool: getFitnessName}})
		return section
	}
	diagnostics := loadDiagnostics(rows)
	if strings.TrimSpace(sport) != "" {
		diagnostics = dataQualityCategoryLoadDiagnostics(rows)
		evidence["sport"] = sport
	}
	for _, diagnostic := range diagnostics {
		severity := "warning"
		if diagnostic.Reason == "trimp_or_hr_load_available" {
			severity = "warning"
		}
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: diagnostic.Reason, Severity: severity, Message: diagnostic.Message, Evidence: dataAvailabilityEvidence(diagnostic), Recommendation: dataQualityRecommendation{Action: "Treat training_load as neutral load; do not relabel HR/TRIMP-derived values as TSS.", Tool: getTrainingSummaryName, Fields: diagnostic.SourceFields}})
	}
	if len(section.Diagnostics) > 0 {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "Load data is visible, but some rows use HR/TRIMP fallback fields or omit fitness/load fields."
	}
	return section
}

func dataQualityCategoryLoadDiagnostics(rows []intervals.SummaryWithCats) []dataAvailabilityDiagnostic {
	missingLoadDates := map[string]bool{}
	alternateFields := map[string]bool{}
	for _, row := range rows {
		if len(row.ByCategory) == 0 {
			missingLoadDates[row.Date] = true
			continue
		}
		for _, category := range row.ByCategory {
			load := categoryTrainingLoad(category)
			if !load.Present {
				missingLoadDates[row.Date] = true
				continue
			}
			if load.Source != "" && load.Source != "training_load" {
				alternateFields[load.Source] = true
			}
		}
	}
	diagnostics := make([]dataAvailabilityDiagnostic, 0, 2)
	if len(alternateFields) > 0 {
		diagnostics = append(diagnostics, dataAvailabilityDiagnostic{Reason: "trimp_or_hr_load_available", SourceFields: sortedBoolKeys(alternateFields), Message: "Sport-filtered training load was preserved from HR/TRIMP-like category fields and is exposed as neutral training_load, not TSS."})
	}
	if len(missingLoadDates) > 0 {
		diagnostics = append(diagnostics, dataAvailabilityDiagnostic{Reason: "missing_training_load", MissingFields: []string{"training_load"}, Dates: sortedBoolKeys(missingLoadDates), Message: "Some sport-filtered summary category rows omit training_load and have no recognized HR/TRIMP fallback field; load-dependent sport totals treat those rows as zero rather than inventing TSS."})
	}
	return diagnostics
}

func dataQualityThresholdsSection(warnings []athleteprofile.ReadinessWarning) dataQualitySection {
	evidence := map[string]any{"warning_count": len(warnings)}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "Sport thresholds and zones look ready for common coaching analysis.", Evidence: evidence}
	if len(warnings) == 0 {
		return section
	}
	section.Status = "warning"
	section.Severity = "warning"
	section.Message = "Some sport thresholds or zones are missing from the athlete profile."
	for _, warning := range warnings {
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: warning.Code, Severity: "warning", Message: warning.Message, Evidence: map[string]any{"sport_types": warning.SportTypes, "field": warning.Field}, Recommendation: dataQualityRecommendation{Action: warning.Action, Tool: updateSportSettingsName, Fields: []string{warning.Field}}})
	}
	return section
}

func dataQualityWellnessSection(rows []intervals.Wellness, endDate string) dataQualitySection {
	dates := dataQualityWellnessDates(rows)
	staleBridge := dataQualityStaleWellnessRows(rows)
	evidence := map[string]any{"wellness_days": len(rows), "dates": dates}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "Wellness data is visible and current for the requested window.", Evidence: evidence}
	if len(rows) == 0 {
		section.Status = "critical"
		section.Severity = "critical"
		section.Message = "No wellness rows are visible for this window."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "missing_wellness_history", Severity: "critical", Message: "No wellness rows were returned for the requested window.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Confirm wellness sync/permissions and inspect get_wellness_data for the same window.", Tool: getWellnessDataName}})
		return section
	}
	if len(dates) == 0 {
		section.Status = "unknown"
		section.Severity = "warning"
		section.Message = "Wellness rows are visible, but none include valid date IDs for freshness checks."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "wellness_dates_unknown", Severity: "warning", Message: "Wellness rows were returned without valid YYYY-MM-DD ids, so freshness cannot be determined safely.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Inspect get_wellness_data include_full output for upstream wellness IDs and sync metadata.", Tool: getWellnessDataName}})
		return section
	}
	latest := dates[len(dates)-1]
	evidence["latest_wellness_date"] = latest
	if len(staleBridge) > 0 {
		evidence["stale_bridge_rows"] = staleBridge
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "Wellness data is visible, but provider bridge metadata indicates stale sync."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "stale_wellness_bridge", Severity: "warning", Message: "At least one wellness row has provider/bridge fetched-at metadata older than 24h for that wellness date.", Evidence: map[string]any{"stale_bridge_rows": staleBridge}, Recommendation: dataQualityRecommendation{Action: "Refresh the upstream wellness bridge and re-check get_wellness_data provenance metadata.", Tool: getWellnessDataName}})
	}
	if dayGap(latest, endDate) > 1 {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "Wellness data is visible but stale relative to the report end date."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "stale_wellness", Severity: "warning", Message: "The newest wellness row is older than the report end date.", Evidence: map[string]any{"latest_wellness_date": latest, "end_date": endDate}, Recommendation: dataQualityRecommendation{Action: "Refresh the upstream wellness bridge or choose a window ending on the latest synced wellness date.", Tool: getWellnessDataName}})
	}
	return section
}

func dataQualityCalendarSection(events []intervals.Event) dataQualitySection {
	raceCount := 0
	for _, event := range events {
		if dataQualityRaceLikeEvent(event) {
			raceCount++
		}
	}
	evidence := map[string]any{"calendar_event_count": len(events), "race_like_event_count": raceCount}
	section := dataQualitySection{Status: "ok", Severity: "info", Message: "Calendar events are visible for this window.", Evidence: evidence}
	if len(events) == 0 {
		section.Status = "warning"
		section.Severity = "warning"
		section.Message = "No calendar events are visible for this window."
		section.Diagnostics = append(section.Diagnostics, dataQualityDiagnostic{Code: "no_calendar_events", Severity: "warning", Message: "No planned workouts, notes, races, or calendar annotations were returned.", Evidence: evidence, Recommendation: dataQualityRecommendation{Action: "Use get_events over a planning/race window if calendar context should exist.", Tool: getEventsName}})
	}
	return section
}

func summarizeDataQuality(sections dataQualitySections, windowDays int) dataQualitySummary {
	all := dataQualitySectionList(sections)
	summary := dataQualitySummary{Status: "ok", WorstSeverity: "info", WindowDays: windowDays}
	for _, section := range all {
		switch section.Status {
		case "critical":
			summary.SectionsCritical++
			summary.Status = "critical"
			summary.WorstSeverity = "critical"
		case "warning":
			summary.SectionsWarning++
			if summary.Status != "critical" {
				summary.Status = "warning"
				summary.WorstSeverity = "warning"
			}
		case "unknown":
			if section.Severity == "warning" || section.Severity == "critical" {
				summary.SectionsWarning++
				if summary.Status != "critical" {
					summary.Status = "warning"
					summary.WorstSeverity = "warning"
				}
			}
		case "ok":
			summary.SectionsOK++
		}
		for _, diagnostic := range section.Diagnostics {
			if diagnostic.Severity == "critical" || len(summary.PrimaryRecommendations) < 3 {
				summary.PrimaryRecommendations = appendUniqueString(summary.PrimaryRecommendations, diagnostic.Recommendation.Action)
			}
		}
	}
	return summary
}

func collectDataQualityDiagnostics(sections dataQualitySections) []dataQualityDiagnostic {
	out := []dataQualityDiagnostic{}
	for _, section := range dataQualitySectionList(sections) {
		out = append(out, section.Diagnostics...)
	}
	return out
}

func dataQualitySectionList(sections dataQualitySections) []dataQualitySection {
	return []dataQualitySection{sections.ActivityCoverage, sections.StreamAvailability, sections.SourceRestrictions, sections.LoadBasis, sections.ThresholdsZones, sections.WellnessFreshness, sections.CalendarRaceData}
}

func filterDataQualityActivities(activities []intervals.Activity, sport string) []intervals.Activity {
	return filterAnalyzerSport(activities, sport)
}

func filterDataQualitySummaries(rows []intervals.SummaryWithCats, sport string) []intervals.SummaryWithCats {
	trimmed := strings.ToLower(strings.TrimSpace(sport))
	if trimmed == "" {
		return rows
	}
	out := []intervals.SummaryWithCats{}
	for _, row := range rows {
		filtered := row
		filtered.ByCategory = nil
		for _, category := range row.ByCategory {
			if strings.ToLower(strings.TrimSpace(category.Category)) == trimmed {
				filtered.ByCategory = append(filtered.ByCategory, category)
			}
		}
		if len(filtered.ByCategory) > 0 {
			out = append(out, filtered)
		}
	}
	return out
}

func sparseActivityHistory(activeDays int, windowDays int) bool {
	if windowDays >= 28 {
		return activeDays < 6
	}
	if windowDays >= 7 {
		return activeDays < 2
	}
	return activeDays == 0
}

func dataQualityActivityDays(activities []intervals.Activity) map[string]bool {
	days := map[string]bool{}
	for _, activity := range activities {
		if date := dataQualityActivityDate(activity); date != "" {
			days[date] = true
		}
	}
	return days
}

func dataQualityActivityDate(activity intervals.Activity) string {
	for _, value := range []string{stringValue(activity.StartDateLocal), stringValue(activity.StartDate)} {
		trimmed := strings.TrimSpace(value)
		if len(trimmed) >= len(time.DateOnly) && validDate(trimmed[:len(time.DateOnly)]) {
			return trimmed[:len(time.DateOnly)]
		}
	}
	return ""
}

func dataQualityActivitySport(activity intervals.Activity) string {
	return firstNonEmpty(stringValue(activity.Type), stringValue(activity.SubType))
}

func dataQualityActivitySamples(activities []intervals.Activity) []dataQualityActivityEvidence {
	limit := min(len(activities), 20)
	out := make([]dataQualityActivityEvidence, 0, limit)
	for _, activity := range activities[:limit] {
		out = append(out, dataQualityActivityEvidence{ActivityID: activity.ID, Date: dataQualityActivityDate(activity), Sport: dataQualityActivitySport(activity), Source: stringValue(activity.Source), StreamTypes: append([]string(nil), activity.StreamTypes...), StravaImported: isStravaBlocked(activity)})
	}
	return out
}

func dataQualitySummaryDates(rows []intervals.SummaryWithCats) []string {
	dates := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Date != "" {
			dates = append(dates, row.Date)
		}
	}
	sort.Strings(dates)
	return dates
}

func dataQualityWellnessDates(rows []intervals.Wellness) []string {
	dates := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.ID != nil && validDate(strings.TrimSpace(*row.ID)) {
			dates = append(dates, strings.TrimSpace(*row.ID))
		}
	}
	sort.Strings(dates)
	return dates
}

func dataQualityStaleWellnessRows(rows []intervals.Wellness) []map[string]any {
	out := []map[string]any{}
	for _, row := range rows {
		date := ""
		if row.ID != nil {
			date = strings.TrimSpace(*row.ID)
		}
		if fetchedAt, stale := dataQualityUpdatedStale(row); stale {
			out = append(out, map[string]any{"date": date, "field": "updated", "source": "wellness_updated", "fetched_at": fetchedAt})
		}
		for _, field := range []struct {
			name    string
			present bool
		}{
			{name: "sleepScore", present: row.SleepScore != nil},
			{name: "sleepSecs", present: row.SleepSecs != nil},
			{name: "readiness", present: row.Readiness != nil},
			{name: "restingHR", present: row.RestingHR != nil},
			{name: "hrv", present: row.HRV != nil},
			{name: "hrvSDNN", present: row.HRVSDNN != nil},
		} {
			if !field.present {
				continue
			}
			entry, stale := wellnessProvenanceEntry(row, field.name)
			if !stale {
				continue
			}
			out = append(out, map[string]any{"date": date, "field": field.name, "source": entry["source"], "fetched_at": entry["fetched_at"]})
		}
	}
	return out
}

func dataQualityUpdatedStale(row intervals.Wellness) (string, bool) {
	if row.Updated == nil || row.ID == nil {
		return "", false
	}
	fetchedAt, parsed, ok := parseWellnessTimestamp(*row.Updated)
	return fetchedAt, ok && wellnessFetchedAtIsStale(row, parsed)
}

func dataQualityEventSamples(events []intervals.Event) []dataQualityEventEvidence {
	limit := min(len(events), 20)
	out := make([]dataQualityEventEvidence, 0, limit)
	for _, event := range events[:limit] {
		out = append(out, dataQualityEventEvidence{EventID: event.ID, Date: dataQualityEventDate(event), Category: stringValue(event.Category), Type: stringValue(event.Type), Name: stringValue(event.Name), RaceLike: dataQualityRaceLikeEvent(event)})
	}
	return out
}

func dataQualityEventDate(event intervals.Event) string {
	for _, value := range []string{stringValue(event.StartDateLocal), stringValue(event.EndDateLocal)} {
		trimmed := strings.TrimSpace(value)
		if len(trimmed) >= len(time.DateOnly) && validDate(trimmed[:len(time.DateOnly)]) {
			return trimmed[:len(time.DateOnly)]
		}
	}
	return ""
}

func dataQualityRaceLikeEvent(event intervals.Event) bool {
	joined := strings.ToLower(strings.Join([]string{stringValue(event.Category), stringValue(event.Type), stringValue(event.Name)}, " "))
	return strings.Contains(joined, "race") || strings.Contains(joined, "event") || strings.Contains(joined, "a-race") || strings.Contains(joined, "b-race") || strings.Contains(joined, "c-race")
}

func activityByID(activities []intervals.Activity, id string) (intervals.Activity, bool) {
	for _, activity := range activities {
		if activity.ID == id {
			return activity, true
		}
	}
	return intervals.Activity{}, false
}

func dataAvailabilityEvidence(diagnostic dataAvailabilityDiagnostic) map[string]any {
	evidence := map[string]any{}
	if len(diagnostic.SourceFields) > 0 {
		evidence["source_fields"] = diagnostic.SourceFields
	}
	if len(diagnostic.MissingFields) > 0 {
		evidence["missing_fields"] = diagnostic.MissingFields
	}
	if len(diagnostic.Dates) > 0 {
		evidence["dates"] = diagnostic.Dates
	}
	return evidence
}

func dayGap(startDate string, endDate string) int {
	start, startErr := time.Parse(time.DateOnly, startDate)
	end, endErr := time.Parse(time.DateOnly, endDate)
	if startErr != nil || endErr != nil || end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours() / 24)
}

func appendUniqueString(values []string, value string) []string {
	if strings.TrimSpace(value) == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func dataQualityActivityFields() []string {
	return append([]string(nil), terseActivityFields...)
}

func dataQualityReportInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{"start_date": map[string]any{"type": "string", "description": "Inclusive athlete-local report start date YYYY-MM-DD."}, "end_date": map[string]any{"type": "string", "description": "Inclusive athlete-local report end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "description": "Optional sport/category filter such as Ride, Run, Swim, or VirtualRide. Leave empty to inspect all visible data."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include bounded evidence lists (activity samples, summary dates, wellness dates, calendar events). Raw stream samples are never fetched by this report."}}}
}

func dataQualityReportOutputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": true, "description": "Read-only data quality report with summary status, section diagnostics for activity coverage, stream availability, Strava/source restrictions, load basis, thresholds/zones, wellness freshness, and calendar/race data. Diagnostics include machine-readable code/severity plus plain-language recommendations."}
}
