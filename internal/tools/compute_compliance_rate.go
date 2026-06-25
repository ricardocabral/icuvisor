package tools

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

const (
	computeComplianceRateName                = "compute_compliance_rate"
	computeComplianceRateDescription         = "Use when the prompt asks how well completed activities matched scheduled workouts, targets, sport, or event type; do not fetch get_* rows or streams and reduce them in chat. Pairs events to activities, returns compliance rates and mean target deltas, and surfaces auto-lap caution when interval evidence is unsafe."
	invalidComputeComplianceArgumentsMessage = "invalid compute_compliance_rate arguments; provide valid dates, target_metric time/distance/load, and tolerance_percent 0..100"
	fetchComputeComplianceMessage            = "could not compute compliance rate; check intervals.icu credentials, athlete ID, and date range"
)

type computeComplianceRequest struct {
	StartDate        string  `json:"start_date"`
	EndDate          string  `json:"end_date"`
	Sport            string  `json:"sport,omitempty"`
	EventType        string  `json:"event_type,omitempty"`
	Category         string  `json:"category,omitempty"`
	TolerancePercent float64 `json:"tolerance_percent,omitempty"`
	TargetMetric     string  `json:"target_metric,omitempty"`
	IncludeFull      bool    `json:"include_full,omitempty"`
}

type complianceEventRow struct {
	EventID              string   `json:"event_id"`
	Name                 string   `json:"name,omitempty"`
	Date                 string   `json:"date,omitempty"`
	Sport                string   `json:"sport,omitempty"`
	EventType            string   `json:"event_type,omitempty"`
	Target               float64  `json:"target"`
	Actual               *float64 `json:"actual,omitempty"`
	Delta                *float64 `json:"delta,omitempty"`
	DeltaPercent         *float64 `json:"delta_percent,omitempty"`
	Compliant            bool     `json:"compliant"`
	WorkoutStatus        string   `json:"workout_status,omitempty"`
	WorkoutStatusCaveats []string `json:"workout_status_caveats,omitempty"`
	PairedActivityID     string   `json:"paired_activity_id,omitempty"`
	PairingSource        string   `json:"pairing_source,omitempty"`
	CautionReason        string   `json:"caution_reason,omitempty"`
}
type complianceResult struct {
	Status                      string                `json:"status"`
	StartDate                   string                `json:"start_date"`
	EndDate                     string                `json:"end_date"`
	Sport                       string                `json:"sport,omitempty"`
	EventType                   string                `json:"event_type,omitempty"`
	TargetMetric                string                `json:"target_metric"`
	TolerancePercent            float64               `json:"tolerance_percent"`
	ScheduledCount              int                   `json:"scheduled_count"`
	CompletedCount              int                   `json:"completed_count"`
	CompliantCount              int                   `json:"compliant_count"`
	ComplianceRate              *float64              `json:"compliance_rate,omitempty"`
	MeanDeltaPercent            *float64              `json:"mean_delta_percent,omitempty"`
	MeanDeltaSeconds            *float64              `json:"mean_delta_seconds,omitempty"`
	MeanDeltaMeters             *float64              `json:"mean_delta_meters,omitempty"`
	MeanDeltaLoad               *float64              `json:"mean_delta_load,omitempty"`
	BySport                     []complianceBreakdown `json:"by_sport,omitempty"`
	ByEventType                 []complianceBreakdown `json:"by_event_type,omitempty"`
	ExcludedEvents              int                   `json:"excluded_events,omitempty"`
	UnpairedEvents              int                   `json:"unpaired_events,omitempty"`
	MissedOrSkippedCount        int                   `json:"missed_or_skipped_count,omitempty"`
	PlannedCount                int                   `json:"planned_count,omitempty"`
	FutureCount                 int                   `json:"future_count,omitempty"`
	CompletedLinkedCount        int                   `json:"completed_linked_count,omitempty"`
	CompletedUnlinkedCount      int                   `json:"completed_unlinked_count,omitempty"`
	WorkoutStatusCaveats        []string              `json:"workout_status_caveats,omitempty"`
	AutoLapCaution              bool                  `json:"auto_lap_caution,omitempty"`
	TruncatedActivityCandidates bool                  `json:"truncated_activity_candidates,omitempty"`
	TruncatedEventCandidates    bool                  `json:"truncated_event_candidates,omitempty"`
	DeltaSampleCount            int                   `json:"delta_sample_count,omitempty"`
	InsufficientReason          string                `json:"insufficient_reason,omitempty"`
}
type complianceBreakdown struct {
	Key              string   `json:"key"`
	ScheduledCount   int      `json:"scheduled_count"`
	CompletedCount   int      `json:"completed_count"`
	CompliantCount   int      `json:"compliant_count"`
	ComplianceRate   *float64 `json:"compliance_rate,omitempty"`
	MeanDeltaPercent *float64 `json:"mean_delta_percent,omitempty"`
	MeanDelta        *float64 `json:"mean_delta,omitempty"`
	DeltaSampleCount int      `json:"delta_sample_count,omitempty"`
}

type complianceActivity struct {
	ID          string
	Date        string
	Sport       string
	MovingTime  float64
	ElapsedTime float64
	Distance    float64
	Load        float64
	Raw         map[string]any
}

type complianceAccumulator struct {
	scheduled int
	completed int
	compliant int
	deltaPct  []float64
	delta     []float64
}

func newComputeComplianceRateTool(eventsClient EventsClient, activitiesClient ActivitiesClient, intervalsClient ActivityIntervalsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, shaping ...responseShaping) Tool {
	return newComputeComplianceRateToolWithClock(eventsClient, activitiesClient, intervalsClient, profileClient, version, timezoneFallback, debugMetadata, time.Now, shaping...)
}

func newComputeComplianceRateToolWithClock(eventsClient EventsClient, activitiesClient ActivitiesClient, intervalsClient ActivityIntervalsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shaping ...responseShaping) Tool {
	if now == nil {
		now = time.Now
	}
	shapeCfg := responseShapingOrDefault(shaping)
	return fullTool(Tool{Name: computeComplianceRateName, Description: computeComplianceRateDescription, InputSchema: computeComplianceInputSchema(), OutputSchema: genericOutputSchema("Scheduled-vs-completed compliance with analyzer metadata, explicit workout_status fields, and missed/skipped caveats."), Handler: computeComplianceRateHandler(eventsClient, activitiesClient, intervalsClient, profileClient, version, timezoneFallback, debugMetadata, now, shapeCfg)})
}

func computeComplianceRateHandler(eventsClient EventsClient, activitiesClient ActivitiesClient, intervalsClient ActivityIntervalsClient, profileClient ProfileClient, version string, timezoneFallback string, debugMetadata bool, now func() time.Time, shapeCfg responseShaping) Handler {
	return func(ctx context.Context, req Request) (Result, error) {
		args, err := decodeComputeComplianceRequest(req.Arguments)
		if err != nil {
			return Result{}, NewUserError(invalidComputeComplianceArgumentsMessage, err)
		}
		unitSystem, timezoneName, err := toolProfile(ctx, profileClient, timezoneFallback)
		if err != nil {
			return Result{}, NewUserError(fetchComputeComplianceMessage, err)
		}
		asOf, err := response.AsOfMetadataInTimezone(now(), timezoneName)
		if err != nil {
			return Result{}, NewUserError(fetchComputeComplianceMessage, err)
		}
		result, series, meta, err := computeCompliance(ctx, args, eventsClient, activitiesClient, intervalsClient, asOf.AsOfDate)
		if err != nil {
			if contextError(err) {
				return Result{}, err
			}
			return Result{}, NewUserError(fetchComputeComplianceMessage, err)
		}
		return encodeAnalyzerResponse(analyzerResponseInput{Result: result, Series: series, Meta: meta}, args.IncludeFull, version, debugMetadata, computeComplianceRateName, unitSystem, shapeCfg)
	}
}

func decodeComputeComplianceRequest(raw json.RawMessage) (computeComplianceRequest, error) {
	var args computeComplianceRequest
	if strings.TrimSpace(string(raw)) == "" {
		return args, errors.New("arguments must be a JSON object")
	}
	decoded, err := DecodeStrict[computeComplianceRequest](raw)
	if err != nil {
		return args, err
	}
	args = decoded
	args.StartDate = strings.TrimSpace(args.StartDate)
	args.EndDate = strings.TrimSpace(args.EndDate)
	args.Sport = strings.TrimSpace(args.Sport)
	args.EventType = strings.TrimSpace(args.EventType)
	args.Category = strings.TrimSpace(args.Category)
	if args.Category == "" {
		args.Category = "WORKOUT"
	}
	args.TargetMetric = strings.ToLower(strings.TrimSpace(args.TargetMetric))
	if args.TargetMetric == "" {
		args.TargetMetric = "time"
	}
	if args.TolerancePercent == 0 {
		args.TolerancePercent = 20
	}
	if !validDate(args.StartDate) || !validDate(args.EndDate) || args.EndDate < args.StartDate {
		return args, errors.New("invalid date range")
	}
	if args.TargetMetric != "time" && args.TargetMetric != "distance" && args.TargetMetric != "load" {
		return args, errors.New("target_metric must be time, distance, or load")
	}
	if args.TolerancePercent < 0 || args.TolerancePercent > 100 {
		return args, errors.New("tolerance_percent must be 0..100")
	}
	return args, nil
}

func computeCompliance(ctx context.Context, args computeComplianceRequest, eventsClient EventsClient, activitiesClient ActivitiesClient, intervalsClient ActivityIntervalsClient, asOfDate string) (complianceResult, []complianceEventRow, analysis.AnalyzerMetaInput, error) {
	if eventsClient == nil || activitiesClient == nil {
		return complianceResult{}, nil, analysis.AnalyzerMetaInput{}, errors.New("missing events or activities client")
	}
	events, err := eventsClient.ListEvents(ctx, intervals.ListEventsParams{Oldest: args.StartDate, Newest: args.EndDate, Category: args.Category, Limit: maxEventsLimit, Resolve: boolPtr(true)})
	if err != nil {
		return complianceResult{}, nil, analysis.AnalyzerMetaInput{}, err
	}
	eventTruncated := len(events) >= maxEventsLimit
	activitiesRaw, err := activitiesClient.ListActivities(ctx, intervals.ListActivitiesParams{Oldest: args.StartDate, Newest: args.EndDate, Limit: maxComputeActivityCandidates})
	if err != nil {
		return complianceResult{}, nil, analysis.AnalyzerMetaInput{}, err
	}
	activityTruncated := len(activitiesRaw) >= maxComputeActivityCandidates
	activities := make([]complianceActivity, 0, len(activitiesRaw))
	for _, activity := range activitiesRaw {
		ca := complianceActivityFromActivity(activity)
		if args.Sport != "" && !sameFold(args.Sport, ca.Sport) {
			continue
		}
		activities = append(activities, ca)
	}
	sort.SliceStable(activities, func(i, j int) bool {
		if activities[i].Date != activities[j].Date {
			return activities[i].Date < activities[j].Date
		}
		return activities[i].ID < activities[j].ID
	})
	scheduled := filterComplianceEvents(events, args)
	used := map[string]bool{}
	linked := map[string]string{}
	for _, activity := range activities {
		for _, key := range []string{"paired_event_id", "event_id", "calendar_event_id", "icu_event_id"} {
			if value := anyString(activity.Raw[key]); value != "" {
				linked[value] = activity.ID
			}
		}
	}
	series := make([]complianceEventRow, 0, len(scheduled))
	bySport := map[string]*complianceAccumulator{}
	byType := map[string]*complianceAccumulator{}
	acc := &complianceAccumulator{}
	excluded := 0
	unpaired := 0
	autoLap := false
	intervalUsed := false
	intervalUnavailable := false
	intervalEvidence := analysis.IntervalSourceResult{}
	statusCounts := map[string]int{}
	statusCaveats := []string{}
	reservedByEvent, reservedActivities, linkConflicts := buildComplianceReservations(scheduled, activities, linked, args.TargetMetric)
	for _, event := range scheduled {
		target, targetKind, ok := complianceTarget(event, args.TargetMetric)
		if !ok {
			excluded++
			continue
		}
		eventID := event.ID
		eventSport := firstNonEmpty(stringValue(event.Type), args.Sport)
		eventType := stringValue(event.Type)
		row := complianceEventRow{EventID: eventID, Name: stringValue(event.Name), Date: localDatePrefix(stringValue(event.StartDateLocal)), Sport: eventSport, EventType: eventType, Target: round(target, 3)}
		acc.scheduled++
		sportAcc := accumulatorFor(bySport, emptyKey(eventSport))
		typeAcc := accumulatorFor(byType, emptyKey(eventType))
		sportAcc.scheduled++
		typeAcc.scheduled++
		var activity *complianceActivity
		source := ""
		if linkConflicts[eventID] {
			unpaired++
			row.PairingSource = "linked_conflict"
			applyComplianceUnpairedStatus(&row, asOfDate, statusCounts, &statusCaveats)
			series = append(series, row)
			continue
		}
		if linkedID := reservedByEvent[eventID]; linkedID != "" {
			activity = findComplianceActivity(activities, linkedID)
			source = "linked"
		}
		if activity == nil {
			activity, source = autoPairActivity(event, target, targetKind, args.Sport, activities, used, reservedActivities)
		}
		if activity == nil {
			unpaired++
			row.PairingSource = "unpaired"
			applyComplianceUnpairedStatus(&row, asOfDate, statusCounts, &statusCaveats)
			series = append(series, row)
			continue
		}
		used[activity.ID] = true
		actual := complianceActual(*activity, targetKind)
		delta := actual - target
		pct := 0.0
		if target != 0 {
			pct = delta / target * 100
		}
		compliant := math.Abs(pct) <= args.TolerancePercent
		if event.WorkoutDoc != nil {
			if intervalsClient == nil {
				intervalUnavailable = true
				row.CautionReason = "interval_evidence_unavailable"
			} else if evidence, unavailable, ierr := complianceIntervalEvidence(ctx, intervalsClient, activity.ID); ierr != nil {
				return complianceResult{}, nil, analysis.AnalyzerMetaInput{}, ierr
			} else if unavailable {
				intervalUsed = true
				intervalUnavailable = true
				row.CautionReason = "interval_evidence_unavailable"
			} else {
				intervalUsed = true
				intervalEvidence = evidence
				if analysis.IntervalExecutionClaimPolicy(evidence).Decline {
					autoLap = true
					row.CautionReason = analysis.IntervalExecutionDeclineAutoLapSuspected
				}
			}
		}
		row.Actual = roundPtr(actual)
		row.Delta = roundPtr(delta)
		row.DeltaPercent = roundPtr(pct)
		row.Compliant = compliant
		row.PairedActivityID = activity.ID
		row.PairingSource = source
		applyComplianceCompletedStatus(&row, source, statusCounts, &statusCaveats)
		series = append(series, row)
		addCompliance(acc, compliant, delta, pct)
		addCompliance(sportAcc, compliant, delta, pct)
		addCompliance(typeAcc, compliant, delta, pct)
	}
	result := complianceResult{Status: "ok", StartDate: args.StartDate, EndDate: args.EndDate, Sport: args.Sport, EventType: args.EventType, TargetMetric: args.TargetMetric, TolerancePercent: args.TolerancePercent, ScheduledCount: acc.scheduled, CompletedCount: acc.completed, CompliantCount: acc.compliant, ComplianceRate: ratePtr(acc.compliant, acc.scheduled), MeanDeltaPercent: meanPtr(acc.deltaPct), BySport: breakdowns(bySport), ByEventType: breakdowns(byType), ExcludedEvents: excluded, UnpairedEvents: unpaired, MissedOrSkippedCount: statusCounts[workoutStatusMissedOrSkipped], PlannedCount: statusCounts[workoutStatusPlanned], FutureCount: statusCounts[workoutStatusFuture], CompletedLinkedCount: statusCounts[workoutStatusCompletedLinked], CompletedUnlinkedCount: statusCounts[workoutStatusCompletedUnlinked], WorkoutStatusCaveats: statusCaveats, AutoLapCaution: autoLap, TruncatedActivityCandidates: activityTruncated, TruncatedEventCandidates: eventTruncated, DeltaSampleCount: acc.completed}
	switch args.TargetMetric {
	case "time":
		result.MeanDeltaSeconds = meanPtr(acc.delta)
	case "distance":
		result.MeanDeltaMeters = meanPtr(acc.delta)
	case "load":
		result.MeanDeltaLoad = meanPtr(acc.delta)
	}
	if activityTruncated || eventTruncated {
		result.Status = "partial"
	}
	if acc.scheduled == 0 {
		if result.Status != "partial" {
			result.Status = "insufficient_sample"
		}
		result.InsufficientReason = "no_scheduled_target_events"
	}
	meta := analysis.AnalyzerMetaInput{Method: "scheduled_completed_event_compliance", SourceTools: []string{getEventsName, getActivitiesName}, N: acc.scheduled, MissingDays: 0, MissingAction: analysis.MissingActionSkip, Assumptions: map[string]any{"target_metric": args.TargetMetric, "tolerance_percent": args.TolerancePercent, "activity_candidates_truncated": activityTruncated, "event_candidates_truncated": eventTruncated, "workout_status_as_of_date": asOfDate}, Boundaries: complianceBoundaries(activityTruncated, eventTruncated, intervalUnavailable, result.PlannedCount+result.FutureCount > 0), InsufficientSample: boolPtr(acc.scheduled == 0)}
	if intervalUsed {
		meta = analysis.ApplyIntervalSourceEvidence(meta, intervalEvidence)
	}
	return result, series, meta, nil
}

func applyComplianceUnpairedStatus(row *complianceEventRow, asOfDate string, counts map[string]int, resultCaveats *[]string) {
	status, caveats := dateDerivedWorkoutStatus(row.Date, asOfDate)
	row.WorkoutStatus = status
	row.WorkoutStatusCaveats = caveats
	if status != "" {
		counts[status]++
	}
	*resultCaveats = appendCaveats(*resultCaveats, caveats...)
}

func applyComplianceCompletedStatus(row *complianceEventRow, source string, counts map[string]int, resultCaveats *[]string) {
	if source == "linked" {
		row.WorkoutStatus = workoutStatusCompletedLinked
		counts[workoutStatusCompletedLinked]++
		return
	}
	row.WorkoutStatus = workoutStatusCompletedUnlinked
	row.WorkoutStatusCaveats = appendCaveats(row.WorkoutStatusCaveats, workoutCaveatDateMetricNotExplicit)
	counts[workoutStatusCompletedUnlinked]++
	*resultCaveats = appendCaveats(*resultCaveats, row.WorkoutStatusCaveats...)
}

func complianceBoundaries(activityTruncated bool, eventTruncated bool, intervalUnavailable bool, hasPendingOrFuture bool) []string {
	boundaries := []string{"one-to-one event/activity matching", "raw streams are not used for compliance"}
	if activityTruncated {
		boundaries = append(boundaries, "activity candidates truncated at deterministic cap")
	}
	if eventTruncated {
		boundaries = append(boundaries, "event candidates truncated at deterministic cap; scheduled denominators may exclude rows past the cap")
	}
	if intervalUnavailable {
		boundaries = append(boundaries, "interval execution evidence could not be verified for one or more paired workout-doc events")
	}
	if hasPendingOrFuture {
		boundaries = append(boundaries, "scheduled_count includes planned/future target events; use workout_status counts before interpreting compliance_rate as due-workout adherence")
	}
	return boundaries
}

func filterComplianceEvents(events []intervals.Event, args computeComplianceRequest) []intervals.Event {
	out := []intervals.Event{}
	for _, event := range events {
		if args.EventType != "" && !sameFold(args.EventType, stringValue(event.Type)) {
			continue
		}
		if args.Sport != "" && args.EventType == "" && !sameFold(args.Sport, stringValue(event.Type)) {
			continue
		}
		out = append(out, event)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if stringValue(out[i].StartDateLocal) != stringValue(out[j].StartDateLocal) {
			return stringValue(out[i].StartDateLocal) < stringValue(out[j].StartDateLocal)
		}
		return out[i].ID < out[j].ID
	})
	return out
}
func complianceTarget(event intervals.Event, metric string) (float64, string, bool) {
	switch metric {
	case "time":
		if event.TimeTarget != nil && *event.TimeTarget > 0 {
			return float64(*event.TimeTarget), "moving_time", true
		}
		if event.ElapsedTimeTarget != nil && *event.ElapsedTimeTarget > 0 {
			return float64(*event.ElapsedTimeTarget), "elapsed_time", true
		}
	case "distance":
		if event.DistanceTarget != nil && *event.DistanceTarget > 0 {
			return *event.DistanceTarget, "distance", true
		}
	case "load":
		if event.LoadTarget != nil && *event.LoadTarget > 0 {
			return *event.LoadTarget, "load", true
		}
	}
	return 0, "", false
}
func complianceActual(activity complianceActivity, kind string) float64 {
	switch kind {
	case "moving_time":
		return activity.MovingTime
	case "elapsed_time":
		return activity.ElapsedTime
	case "distance":
		return activity.Distance
	case "load":
		return activity.Load
	}
	return 0
}
func complianceActivityFromActivity(activity intervals.Activity) complianceActivity {
	ca := complianceActivity{ID: activity.ID, Date: localDatePrefix(stringValue(activity.StartDateLocal)), Sport: stringValue(activity.Type), Raw: activity.Raw}
	if activity.MovingTime != nil {
		ca.MovingTime = float64(*activity.MovingTime)
	}
	if activity.ElapsedTime != nil {
		ca.ElapsedTime = float64(*activity.ElapsedTime)
	}
	if activity.Distance != nil {
		ca.Distance = *activity.Distance
	} else if activity.ICUDistance != nil {
		ca.Distance = *activity.ICUDistance
	}
	if activity.TrainingLoad != nil {
		ca.Load = float64(*activity.TrainingLoad)
	}
	return ca
}
func buildComplianceReservations(events []intervals.Event, activities []complianceActivity, linked map[string]string, targetMetric string) (map[string]string, map[string]bool, map[string]bool) {
	reservedByEvent := map[string]string{}
	reservedActivities := map[string]bool{}
	conflicts := map[string]bool{}
	for _, event := range events {
		if _, _, ok := complianceTarget(event, targetMetric); !ok {
			continue
		}
		ids := linkedActivityIDsForEvent(event, linked)
		for _, id := range ids {
			if findComplianceActivity(activities, id) == nil {
				continue
			}
			if reservedActivities[id] {
				conflicts[event.ID] = true
				break
			}
			reservedByEvent[event.ID] = id
			reservedActivities[id] = true
			break
		}
	}
	return reservedByEvent, reservedActivities, conflicts
}

func linkedActivityIDsForEvent(event intervals.Event, linked map[string]string) []string {
	ids := []string{}
	for _, key := range []string{"activity_id", "icu_activity_id", "paired_activity_id", "completed_activity_id"} {
		if value := anyString(event.Raw[key]); value != "" {
			ids = append(ids, value)
		}
	}
	if linked[event.ID] != "" {
		ids = append(ids, linked[event.ID])
	}
	return ids
}

func findComplianceActivity(activities []complianceActivity, id string) *complianceActivity {
	for i := range activities {
		if activities[i].ID == id {
			return &activities[i]
		}
	}
	return nil
}

func autoPairActivity(event intervals.Event, target float64, kind string, requestSport string, activities []complianceActivity, used map[string]bool, reserved map[string]bool) (*complianceActivity, string) {
	date := localDatePrefix(stringValue(event.StartDateLocal))
	sport := firstNonEmpty(requestSport, stringValue(event.Type))
	best := -1
	bestDiff := math.MaxFloat64
	for i := range activities {
		if used[activities[i].ID] || reserved[activities[i].ID] || activities[i].Date != date {
			continue
		}
		if sport != "" && !sameFold(sport, activities[i].Sport) {
			continue
		}
		diff := math.Abs(complianceActual(activities[i], kind) - target)
		if diff < bestDiff || (diff == bestDiff && (best < 0 || activities[i].ID < activities[best].ID)) {
			best = i
			bestDiff = diff
		}
	}
	if best < 0 {
		return nil, ""
	}
	return &activities[best], "date_metric_match"
}
func complianceIntervalEvidence(ctx context.Context, client ActivityIntervalsClient, activityID string) (analysis.IntervalSourceResult, bool, error) {
	dto, err := client.GetActivityIntervals(ctx, activityID)
	if err != nil {
		if contextError(err) {
			return analysis.IntervalSourceResult{}, false, err
		}
		return analysis.IntervalSourceResult{}, true, nil
	}
	return classifyActivityIntervalsDTO(dto), false, nil
}
func addCompliance(acc *complianceAccumulator, compliant bool, delta, pct float64) {
	acc.completed++
	if compliant {
		acc.compliant++
	}
	acc.delta = append(acc.delta, delta)
	acc.deltaPct = append(acc.deltaPct, pct)
}
func accumulatorFor(values map[string]*complianceAccumulator, key string) *complianceAccumulator {
	if values[key] == nil {
		values[key] = &complianceAccumulator{}
	}
	return values[key]
}
func ratePtr(num, den int) *float64 {
	if den == 0 {
		return nil
	}
	v := round(float64(num)/float64(den), 4)
	return &v
}
func meanPtr(values []float64) *float64 {
	if len(values) == 0 {
		return nil
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	v := round(total/float64(len(values)), 4)
	return &v
}
func breakdowns(values map[string]*complianceAccumulator) []complianceBreakdown {
	out := make([]complianceBreakdown, 0, len(values))
	for key, acc := range values {
		out = append(out, complianceBreakdown{Key: key, ScheduledCount: acc.scheduled, CompletedCount: acc.completed, CompliantCount: acc.compliant, ComplianceRate: ratePtr(acc.compliant, acc.scheduled), MeanDeltaPercent: meanPtr(acc.deltaPct), MeanDelta: meanPtr(acc.delta), DeltaSampleCount: acc.completed})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
func emptyKey(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}

func computeComplianceInputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false, "required": []string{"start_date", "end_date"}, "properties": map[string]any{"start_date": map[string]any{"type": "string", "description": "Athlete-local inclusive start date YYYY-MM-DD."}, "end_date": map[string]any{"type": "string", "description": "Athlete-local inclusive end date YYYY-MM-DD."}, "sport": map[string]any{"type": "string", "description": "Optional exact case-insensitive sport/type filter."}, "event_type": map[string]any{"type": "string", "description": "Optional exact case-insensitive event type filter."}, "category": map[string]any{"type": "string", "default": "WORKOUT", "description": "Optional upstream event category filter."}, "target_metric": map[string]any{"type": "string", "enum": []string{"time", "distance", "load"}, "default": "time", "description": "Target/actual metric used for compliance."}, "tolerance_percent": map[string]any{"type": "number", "minimum": 0, "maximum": 100, "default": 20, "description": "Absolute percent delta allowed for compliance."}, "include_full": map[string]any{"type": "boolean", "default": false, "description": "When true, include event-level pairing audit rows."}}}
}
