package tools

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeComputeClient struct {
	fakeProfileClient
	summaries   []intervals.SummaryWithCats
	wellness    []intervals.Wellness
	events      []intervals.Event
	activities  []intervals.Activity
	details     map[string]intervals.Activity
	intervals   map[string]intervals.IntervalsDTO
	intervalErr error

	summaryCalls  int
	activityCalls int
	detailCalls   []string
	intervalCalls int
}

func newFakeComputeClient() *fakeComputeClient {
	return &fakeComputeClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC"}},
		details:           map[string]intervals.Activity{},
		intervals:         map[string]intervals.IntervalsDTO{},
	}
}

func (c *fakeComputeClient) ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	c.summaryCalls++
	return append([]intervals.SummaryWithCats(nil), c.summaries...), nil
}

func (c *fakeComputeClient) ListActivities(context.Context, intervals.ListActivitiesParams) ([]intervals.Activity, error) {
	c.activityCalls++
	return append([]intervals.Activity(nil), c.activities...), nil
}

func (c *fakeComputeClient) GetActivity(_ context.Context, activityID string) (intervals.Activity, error) {
	c.detailCalls = append(c.detailCalls, activityID)
	if activity, ok := c.details[activityID]; ok {
		return activity, nil
	}
	for _, activity := range c.activities {
		if activity.ID == activityID {
			return activity, nil
		}
	}
	return intervals.Activity{}, errors.New("activity not found")
}

func (c *fakeComputeClient) GetActivityIntervals(_ context.Context, activityID string) (intervals.IntervalsDTO, error) {
	c.intervalCalls++
	if c.intervalErr != nil {
		return intervals.IntervalsDTO{}, c.intervalErr
	}
	return c.intervals[activityID], nil
}

func (c *fakeComputeClient) GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error) {
	return intervals.PowerVsHR{}, nil
}

func (c *fakeComputeClient) ListWellness(context.Context, intervals.WellnessParams) ([]intervals.Wellness, error) {
	return append([]intervals.Wellness(nil), c.wellness...), nil
}

func (c *fakeComputeClient) ListEvents(context.Context, intervals.ListEventsParams) ([]intervals.Event, error) {
	return append([]intervals.Event(nil), c.events...), nil
}

func TestComputeZoneTimePrefersSummaryAndReportsMissingDays(t *testing.T) {
	client := newFakeComputeClient()
	client.summaries = []intervals.SummaryWithCats{{Date: "2026-05-01", TimeInZones: []float64{1800, 900, 300}, TimeInZonesTot: 3000, TrainingLoad: 75}}
	tool := newComputeZoneTimeTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-02","zone_metric":"power","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	meta := got["_meta"].(map[string]any)
	if status := resultMap["status"]; status != "partial" {
		t.Fatalf("status = %v, want partial for missing day", status)
	}
	if missing := meta["missing_days"]; missing != float64(1) {
		t.Fatalf("_meta.missing_days = %v, want 1", missing)
	}
	if client.activityCalls != 0 || len(client.detailCalls) != 0 || client.intervalCalls != 0 {
		t.Fatalf("unexpected fallback calls: activities=%d details=%v intervals=%d", client.activityCalls, client.detailCalls, client.intervalCalls)
	}
	sourceTools := stringSliceFromAny(meta["source_tools"])
	if !slices.Equal(sourceTools, []string{getTrainingSummaryName}) {
		t.Fatalf("source_tools = %v, want training summary only", sourceTools)
	}
}

func TestComputeZoneTimeUsesActivityPrecomputedAndNeverStreams(t *testing.T) {
	client := newFakeComputeClient()
	client.activities = []intervals.Activity{
		computeActivityFixture("run-1", "Run", "2026-05-01T07:00:00", map[string]any{}),
		computeActivityFixture("ride-1", "Ride", "2026-05-01T09:00:00", map[string]any{"hr_zone_times": []any{60.0, 60.0, 60.0}}),
	}
	client.details["run-1"] = computeActivityFixture("run-1", "Run", "2026-05-01T07:00:00", map[string]any{"hr_zone_times": []any{300.0, 200.0, 100.0}})
	tool := newComputeZoneTimeTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","zone_metric":"heart_rate","sport":"Run","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	if status := resultMap["status"]; status != "ok" {
		t.Fatalf("status = %v, want ok", status)
	}
	if client.summaryCalls != 0 || client.intervalCalls != 0 {
		t.Fatalf("unexpected summary/stream calls: summaries=%d intervals=%d", client.summaryCalls, client.intervalCalls)
	}
	if !slices.Equal(client.detailCalls, []string{"run-1"}) {
		t.Fatalf("detail calls = %v, want filtered run detail only", client.detailCalls)
	}
	series := got["series"].([]any)
	row := series[0].(map[string]any)
	if row["source_key"] != "hr_zone_times" {
		t.Fatalf("source_key = %v, want hr_zone_times", row["source_key"])
	}
}

func TestComputeZoneTimeNoPrecomputedReturnsUnavailableBoundaryWithoutStreams(t *testing.T) {
	client := newFakeComputeClient()
	client.activities = []intervals.Activity{computeActivityFixture("run-1", "Run", "2026-05-01T07:00:00", map[string]any{})}
	client.details["run-1"] = computeActivityFixture("run-1", "Run", "2026-05-01T07:00:00", map[string]any{})
	tool := newComputeZoneTimeTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","zone_metric":"pace","sport":"Run","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	meta := got["_meta"].(map[string]any)
	if status := resultMap["status"]; status != "unavailable" {
		t.Fatalf("status = %v, want unavailable", status)
	}
	if reason := resultMap["insufficient_reason"]; reason != "missing_precomputed_zone_times" {
		t.Fatalf("insufficient_reason = %v", reason)
	}
	if client.intervalCalls != 0 {
		t.Fatalf("interval calls = %d, want no raw-stream fallback", client.intervalCalls)
	}
	if !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "raw streams are not reduced") {
		t.Fatalf("boundaries = %v, want raw-stream boundary", meta["boundaries"])
	}
}

func TestComputeLoadBalanceUsesLoadPriorityIndependentOfMetric(t *testing.T) {
	client := newFakeComputeClient()
	client.activities = []intervals.Activity{computeActivityFixture("run-1", "Run", "2026-05-01T07:00:00", map[string]any{
		"hr_zone_times": []any{240.0, 120.0, 60.0},
		"power_load":    91.0,
	})}
	tool := newComputeLoadBalanceTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","zone_metric":"heart_rate","sport":"Run"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	if load := resultMap["training_load_total"]; load != float64(91) {
		t.Fatalf("training_load_total = %v, want power_load fallback 91", load)
	}
	if client.intervalCalls != 0 {
		t.Fatalf("interval calls = %d, want no raw-stream fallback", client.intervalCalls)
	}
}

func TestComputeBaselineWellnessZScoreIncludesFormulaAndInterpretation(t *testing.T) {
	client := newFakeComputeClient()
	client.wellness = []intervals.Wellness{
		wellnessFixture("2026-05-01", map[string]any{"restingHR": 50.0}),
		wellnessFixture("2026-05-02", map[string]any{"restingHR": 60.0}),
		wellnessFixture("2026-05-04", map[string]any{"restingHR": 70.0}),
	}
	tool := newComputeBaselineTool(nil, client, nil, nil, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"rhr","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","min_samples":2,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	meta := got["_meta"].(map[string]any)
	if resultMap["status"] != "ok" || resultMap["interpretation"] != "elevated" {
		t.Fatalf("result = %v, want ok elevated", resultMap)
	}
	assertFloatEqual(t, resultMap["baseline_mean"], 55)
	assertFloatEqual(t, resultMap["baseline_stddev"], 7.0711)
	assertFloatEqual(t, resultMap["z_score"], 2.1213)
	if meta["formula_ref"] != "icuvisor://analysis-formulas#z_score" || meta["n"] != float64(2) {
		t.Fatalf("meta = %v, want z-score formula and n=2", meta)
	}
	assumptions := meta["assumptions"].(map[string]any)
	if assumptions["interpretation_direction"] != "adverse_high" {
		t.Fatalf("assumptions = %v, want adverse_high", assumptions)
	}
	if len(got["series"].([]any)) != 3 {
		t.Fatalf("series len = %d, want 3", len(got["series"].([]any)))
	}
}

func TestComputeBaselineStatusGoldensAndValidation(t *testing.T) {
	tests := []struct {
		name       string
		wellness   []intervals.Wellness
		args       string
		wantStatus string
		wantReason string
		wantErr    bool
	}{
		{
			name:       "insufficient baseline samples",
			wellness:   []intervals.Wellness{wellnessFixture("2026-05-01", map[string]any{"restingHR": 50.0}), wellnessFixture("2026-05-04", map[string]any{"restingHR": 55.0})},
			args:       `{"metric":"rhr","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","min_samples":2}`,
			wantStatus: "insufficient_sample",
			wantReason: "not_enough_baseline_samples",
		},
		{
			name:       "missing current window",
			wellness:   []intervals.Wellness{wellnessFixture("2026-05-01", map[string]any{"restingHR": 50.0}), wellnessFixture("2026-05-02", map[string]any{"restingHR": 60.0})},
			args:       `{"metric":"rhr","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","min_samples":2}`,
			wantStatus: "insufficient_current_sample",
			wantReason: "no_current_samples",
		},
		{
			name:       "zero variance",
			wellness:   []intervals.Wellness{wellnessFixture("2026-05-01", map[string]any{"restingHR": 50.0}), wellnessFixture("2026-05-02", map[string]any{"restingHR": 50.0}), wellnessFixture("2026-05-04", map[string]any{"restingHR": 60.0})},
			args:       `{"metric":"rhr","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","min_samples":2}`,
			wantStatus: "insufficient_variance",
			wantReason: "zero_baseline_variance",
		},
		{
			name:    "cross window ordering validation",
			args:    `{"metric":"rhr","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-04","current_start_date":"2026-05-04","current_end_date":"2026-05-05","min_samples":2}`,
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeComputeClient()
			client.wellness = tc.wellness
			tool := newComputeBaselineTool(nil, client, nil, nil, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(tc.args)})
			if tc.wantErr {
				if err == nil {
					t.Fatal("Handler() error = nil, want validation error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			resultMap := resultMap(t, result)["result"].(map[string]any)
			if resultMap["status"] != tc.wantStatus || resultMap["insufficient_reason"] != tc.wantReason {
				t.Fatalf("result = %v, want status=%s reason=%s", resultMap, tc.wantStatus, tc.wantReason)
			}
		})
	}
}

func TestComputeBaselineActivityTruncationPrecedesInsufficientSamples(t *testing.T) {
	tests := []struct {
		name       string
		activities []intervals.Activity
		wantReason string
	}{
		{
			name:       "insufficient baseline samples remains partial under cap",
			activities: cappedActivities("Run", nil),
			wantReason: "not_enough_baseline_samples",
		},
		{
			name: "insufficient current samples remains partial under cap",
			activities: append([]intervals.Activity{
				activityWithHR("run-b1", "Run", "2026-05-01T07:00:00", 140),
				activityWithHR("run-b2", "Run", "2026-05-02T07:00:00", 150),
			}, cappedActivities("Run", nil)[:maxComputeActivityCandidates-2]...),
			wantReason: "no_current_samples",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeComputeClient()
			client.activities = tc.activities
			tool := newComputeBaselineTool(nil, nil, client, nil, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"average_heart_rate_bpm","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","sport":"Run","min_samples":2}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			got := resultMap(t, result)
			resultMap := got["result"].(map[string]any)
			meta := got["_meta"].(map[string]any)
			if resultMap["status"] != "partial" || resultMap["insufficient_reason"] != tc.wantReason {
				t.Fatalf("result = %v, want partial reason %s", resultMap, tc.wantReason)
			}
			if resultMap["truncated_activity_candidates"] != true {
				t.Fatalf("result = %v, want truncated flag", resultMap)
			}
			if !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "activity candidates truncated") {
				t.Fatalf("boundaries = %v, want truncation boundary", meta["boundaries"])
			}
		})
	}
}

func TestComputeBaselineWeeklyAndActivityGrains(t *testing.T) {
	t.Run("weekly training load buckets summary rows", func(t *testing.T) {
		client := newFakeComputeClient()
		client.summaries = []intervals.SummaryWithCats{
			{Date: "2026-05-04", TrainingLoad: 10}, {Date: "2026-05-05", TrainingLoad: 20},
			{Date: "2026-05-11", TrainingLoad: 30}, {Date: "2026-05-12", TrainingLoad: 30},
			{Date: "2026-05-18", TrainingLoad: 90},
		}
		tool := newComputeBaselineTool(client, nil, nil, nil, client, "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"weekly_tss","baseline_start_date":"2026-05-04","baseline_end_date":"2026-05-17","current_start_date":"2026-05-18","current_end_date":"2026-05-24","min_samples":2}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		resultMap := resultMap(t, result)["result"].(map[string]any)
		metricSource := resultMap["metric_source"].(map[string]any)
		if metricSource["grain"] != "derived_weekly" || resultMap["n_baseline"] != float64(2) {
			t.Fatalf("result = %v, want two derived weekly samples", resultMap)
		}
		assertFloatEqual(t, resultMap["current_value"], 90)
	})
	t.Run("activity row grain honors sport filter", func(t *testing.T) {
		client := newFakeComputeClient()
		client.activities = []intervals.Activity{
			activityWithHR("run-1", "Run", "2026-05-01T07:00:00", 140),
			activityWithHR("run-2", "Run", "2026-05-02T07:00:00", 150),
			activityWithHR("ride-1", "Ride", "2026-05-02T09:00:00", 190),
			activityWithHR("run-3", "Run", "2026-05-04T07:00:00", 160),
		}
		tool := newComputeBaselineTool(nil, nil, client, nil, client, "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"average_heart_rate_bpm","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","sport":"Run","min_samples":2}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		resultMap := resultMap(t, result)["result"].(map[string]any)
		metricSource := resultMap["metric_source"].(map[string]any)
		if metricSource["grain"] != "activity" || resultMap["n_baseline"] != float64(2) {
			t.Fatalf("result = %v, want activity grain with ride excluded", resultMap)
		}
		assertFloatEqual(t, resultMap["current_value"], 160)
	})
}

func TestComputeCompliancePairingDeltasAndBreakdowns(t *testing.T) {
	client := newFakeComputeClient()
	client.events = []intervals.Event{
		eventFixture("evt-targetless", "Workout", "2026-05-01", nil, map[string]any{"activity_id": "act-targetless"}, false),
		eventFixture("evt-auto", "Workout", "2026-05-01", intPtr(3600), nil, false),
		eventFixture("evt-nonworkout", "Race", "2026-05-01", intPtr(3600), nil, false),
		eventFixture("evt-linked", "Workout", "2026-05-02", intPtr(1800), map[string]any{"activity_id": "act-linked"}, false),
		eventFixture("evt-conflict", "Workout", "2026-05-02", intPtr(1800), map[string]any{"activity_id": "act-linked"}, false),
	}
	client.activities = []intervals.Activity{
		activityWithTime("act-targetless", "Run", "2026-05-01T07:00:00", 3500, map[string]any{"paired_event_id": "evt-targetless"}),
		activityWithTime("act-ride-closer", "Ride", "2026-05-01T07:05:00", 3600, nil),
		activityWithTime("act-linked", "Run", "2026-05-02T07:00:00", 1980, nil),
	}
	tool := newComputeComplianceRateTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-03","sport":"Run","event_type":"Workout","target_metric":"time","tolerance_percent":10,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	if resultMap["scheduled_count"] != float64(3) || resultMap["excluded_events"] != float64(1) || resultMap["completed_count"] != float64(2) || resultMap["compliant_count"] != float64(2) || resultMap["unpaired_events"] != float64(1) {
		t.Fatalf("result counts = %v", resultMap)
	}
	assertFloatEqual(t, resultMap["compliance_rate"], 0.6667)
	assertFloatEqual(t, resultMap["mean_delta_percent"], 3.6111)
	assertFloatEqual(t, resultMap["mean_delta_seconds"], 40)
	if resultMap["delta_sample_count"] != float64(2) {
		t.Fatalf("delta_sample_count = %v, want completed denominator 2", resultMap["delta_sample_count"])
	}
	rows := rowsByEventID(got["series"].([]any))
	if _, ok := rows["evt-nonworkout"]; ok {
		t.Fatalf("evt-nonworkout appeared in series despite event_type filter: %v", rows["evt-nonworkout"])
	}
	if rows["evt-auto"]["paired_activity_id"] != "act-targetless" || rows["evt-auto"]["pairing_source"] != "date_metric_match" {
		t.Fatalf("evt-auto row = %v, want targetless-linked Run activity available for auto-pair instead of closer Ride", rows["evt-auto"])
	}
	if rows["evt-conflict"]["pairing_source"] != "linked_conflict" && rows["evt-linked"]["pairing_source"] != "linked_conflict" {
		t.Fatalf("linked rows = %v / %v, want one linked conflict without activity reuse", rows["evt-linked"], rows["evt-conflict"])
	}
	breakdowns := resultMap["by_event_type"].([]any)
	if len(breakdowns) != 1 {
		t.Fatalf("by_event_type = %v, want one Workout row", breakdowns)
	}
	workout := breakdowns[0].(map[string]any)
	if workout["key"] != "Workout" || workout["delta_sample_count"] != float64(2) {
		t.Fatalf("by_event_type row = %v, want completed delta denominator", workout)
	}
}

func TestComputeComplianceWorkoutStatusCountsAndCaveats(t *testing.T) {
	client := newFakeComputeClient()
	client.profile.Timezone = "America/Sao_Paulo"
	client.events = []intervals.Event{
		eventFixture("evt-missed", "Run", "2026-05-23", intPtr(3600), nil, false),
		eventFixture("evt-planned", "Run", "2026-05-24", intPtr(3600), nil, false),
		eventFixture("evt-future", "Run", "2026-05-25", intPtr(3600), nil, false),
		eventFixture("evt-linked", "Run", "2026-05-22", intPtr(1800), map[string]any{"activity_id": "act-linked"}, false),
		eventFixture("evt-auto", "Run", "2026-05-21", intPtr(2400), nil, false),
	}
	client.activities = []intervals.Activity{
		activityWithTime("act-linked", "Run", "2026-05-22T07:00:00", 1800, nil),
		activityWithTime("act-auto", "Run", "2026-05-21T07:00:00", 2400, nil),
	}
	tool := newComputeComplianceRateToolWithClock(client, client, client, client, "test", "UTC", false, fixedTodayClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-21","end_date":"2026-05-25","sport":"Run","target_metric":"time","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	if resultMap["scheduled_count"] != float64(5) || resultMap["completed_count"] != float64(2) {
		t.Fatalf("result counts = %#v, want scheduled 5 completed 2 preserved", resultMap)
	}
	wantCounts := map[string]float64{
		"missed_or_skipped_count":  1,
		"planned_count":            1,
		"future_count":             1,
		"completed_linked_count":   1,
		"completed_unlinked_count": 1,
	}
	for key, want := range wantCounts {
		if resultMap[key] != want {
			t.Fatalf("%s = %#v, want %v in result %#v", key, resultMap[key], want, resultMap)
		}
	}
	resultCaveats := stringSliceFromAny(resultMap["workout_status_caveats"])
	if !stringSliceContains(resultCaveats, workoutCaveatSkippedMissedUnavailable) || !stringSliceContains(resultCaveats, workoutCaveatDateMetricNotExplicit) {
		t.Fatalf("result caveats = %#v, want missed/skipped and auto-pair caveats", resultCaveats)
	}
	rows := rowsByEventID(got["series"].([]any))
	if rows["evt-missed"]["workout_status"] != workoutStatusMissedOrSkipped || rows["evt-planned"]["workout_status"] != workoutStatusPlanned || rows["evt-future"]["workout_status"] != workoutStatusFuture {
		t.Fatalf("unpaired status rows = missed:%#v planned:%#v future:%#v", rows["evt-missed"], rows["evt-planned"], rows["evt-future"])
	}
	if rows["evt-linked"]["workout_status"] != workoutStatusCompletedLinked || rows["evt-auto"]["workout_status"] != workoutStatusCompletedUnlinked {
		t.Fatalf("completed status rows = linked:%#v auto:%#v", rows["evt-linked"], rows["evt-auto"])
	}
	if !stringSliceContains(stringSliceFromAny(rows["evt-auto"]["workout_status_caveats"]), workoutCaveatDateMetricNotExplicit) {
		t.Fatalf("auto-pair row caveats = %#v, want explicit-link caveat", rows["evt-auto"])
	}
	for id, row := range rows {
		if row["workout_status"] == "deleted_or_cancelled" {
			t.Fatalf("%s row spuriously emitted deleted_or_cancelled: %#v", id, row)
		}
	}
	meta := got["_meta"].(map[string]any)
	if !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "scheduled_count includes planned/future") {
		t.Fatalf("boundaries = %#v, want planned/future denominator caveat", meta["boundaries"])
	}
}

func TestComputeComplianceTruncationMetadata(t *testing.T) {
	t.Run("event cap marks partial with boundary", func(t *testing.T) {
		client := newFakeComputeClient()
		client.events = cappedEvents("Run")
		client.activities = []intervals.Activity{activityWithTime("act-000", "Run", "2026-05-01T07:00:00", 3600, nil)}
		tool := newComputeComplianceRateTool(client, client, client, client, "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","sport":"Run","target_metric":"time"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		got := resultMap(t, result)
		resultMap := got["result"].(map[string]any)
		meta := got["_meta"].(map[string]any)
		if resultMap["status"] != "partial" || resultMap["truncated_event_candidates"] != true {
			t.Fatalf("result = %v, want event truncation partial", resultMap)
		}
		assumptions := meta["assumptions"].(map[string]any)
		if assumptions["event_candidates_truncated"] != true || !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "event candidates truncated") {
			t.Fatalf("meta = %v, want event truncation assumption/boundary", meta)
		}
	})
	t.Run("activity cap marks partial with boundary", func(t *testing.T) {
		client := newFakeComputeClient()
		client.events = []intervals.Event{eventFixture("evt-1", "Run", "2026-05-01", intPtr(3600), nil, false)}
		client.activities = cappedActivitiesWithTime("Run", 3600)
		tool := newComputeComplianceRateTool(client, client, client, client, "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","sport":"Run","target_metric":"time"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		got := resultMap(t, result)
		resultMap := got["result"].(map[string]any)
		meta := got["_meta"].(map[string]any)
		if resultMap["status"] != "partial" || resultMap["truncated_activity_candidates"] != true {
			t.Fatalf("result = %v, want activity truncation partial", resultMap)
		}
		assumptions := meta["assumptions"].(map[string]any)
		if assumptions["activity_candidates_truncated"] != true || !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "activity candidates truncated") {
			t.Fatalf("meta = %v, want activity truncation assumption/boundary", meta)
		}
	})
}

func TestComputeComplianceIntervalCautions(t *testing.T) {
	t.Run("auto lap evidence sets caution and meta", func(t *testing.T) {
		client := newFakeComputeClient()
		client.events = []intervals.Event{eventFixture("evt-auto-lap", "Run", "2026-05-01", intPtr(3600), nil, true)}
		client.activities = []intervals.Activity{activityWithTime("act-auto-lap", "Run", "2026-05-01T07:00:00", 3600, nil)}
		client.intervals["act-auto-lap"] = autoLapIntervalsDTO()
		tool := newComputeComplianceRateTool(client, client, client, client, "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","sport":"Run","target_metric":"time","include_full":true}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		got := resultMap(t, result)
		resultMap := got["result"].(map[string]any)
		meta := got["_meta"].(map[string]any)
		row := got["series"].([]any)[0].(map[string]any)
		if resultMap["auto_lap_caution"] != true || row["caution_reason"] != "auto_lap_suspected" || meta["auto_lap_suspected"] != true {
			t.Fatalf("result=%v row=%v meta=%v, want auto-lap caution", resultMap, row, meta)
		}
		if !slices.Contains(stringSliceFromAny(meta["source_tools"]), "get_activity_intervals") {
			t.Fatalf("source_tools = %v, want interval evidence source", meta["source_tools"])
		}
	})
	t.Run("interval unavailable adds row caution and boundary", func(t *testing.T) {
		client := newFakeComputeClient()
		client.events = []intervals.Event{eventFixture("evt-unavailable", "Run", "2026-05-01", intPtr(3600), nil, true)}
		client.activities = []intervals.Activity{activityWithTime("act-unavailable", "Run", "2026-05-01T07:00:00", 3600, nil)}
		client.intervalErr = errors.New("upstream unavailable")
		tool := newComputeComplianceRateTool(client, client, client, client, "test", "UTC", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","sport":"Run","target_metric":"time","include_full":true}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		got := resultMap(t, result)
		meta := got["_meta"].(map[string]any)
		row := got["series"].([]any)[0].(map[string]any)
		if row["caution_reason"] != "interval_evidence_unavailable" {
			t.Fatalf("row = %v, want interval unavailable caution", row)
		}
		if !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "interval execution evidence could not be verified") {
			t.Fatalf("boundaries = %v, want interval unavailable boundary", meta["boundaries"])
		}
	})
}

func TestComputeZoneTimeTruncationWinsEvenWithoutUsableZones(t *testing.T) {
	client := newFakeComputeClient()
	client.activities = cappedActivities("Run", nil)
	tool := newComputeZoneTimeTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-31","zone_metric":"heart_rate","sport":"Run"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	assertTruncatedAnalyzerResult(t, result, "activity_candidates_truncated")
}

func TestComputeLoadBalanceActivityTruncationIncludesBoundary(t *testing.T) {
	client := newFakeComputeClient()
	client.activities = cappedActivities("Run", map[string]any{"hr_zone_times": []any{120.0, 60.0, 30.0}})
	tool := newComputeLoadBalanceTool(client, client, client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-31","zone_metric":"heart_rate","sport":"Run"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	assertTruncatedAnalyzerResult(t, result, "activity_candidates_truncated")
}

func computeActivityFixture(id string, sport string, start string, raw map[string]any) intervals.Activity {
	activity := intervals.Activity{ID: id, Type: ptrString(sport), StartDateLocal: ptrString(start), Raw: raw}
	if activity.Raw == nil {
		activity.Raw = map[string]any{}
	}
	return activity
}

func activityWithHR(id string, sport string, start string, hr int) intervals.Activity {
	activity := computeActivityFixture(id, sport, start, map[string]any{})
	activity.AverageHeartRate = intPtr(hr)
	return activity
}

func activityWithTime(id string, sport string, start string, movingTime int, raw map[string]any) intervals.Activity {
	activity := computeActivityFixture(id, sport, start, raw)
	activity.MovingTime = intPtr(movingTime)
	activity.ElapsedTime = intPtr(movingTime)
	return activity
}

func eventFixture(id string, eventType string, date string, timeTarget *int, raw map[string]any, workoutDoc bool) intervals.Event {
	if raw == nil {
		raw = map[string]any{}
	}
	raw["id"] = id
	event := intervals.Event{ID: id, Type: ptrString(eventType), Category: ptrString("WORKOUT"), StartDateLocal: ptrString(date + "T00:00:00"), TimeTarget: timeTarget, Raw: raw}
	if workoutDoc {
		event.WorkoutDoc = map[string]any{"steps": []any{map[string]any{"duration": 600.0}}}
	}
	return event
}

func autoLapIntervalsDTO() intervals.IntervalsDTO {
	intervalsOut := make([]intervals.ActivityInterval, 0, 5)
	for idx := range 5 {
		duration := 300.0
		name := "Lap"
		intervalsOut = append(intervalsOut, intervals.ActivityInterval{Name: &name, Duration: &duration, Raw: map[string]any{"auto_lap": true}, StartIndex: intPtr(idx * 100), EndIndex: intPtr(idx*100 + 100)})
	}
	return intervals.IntervalsDTO{ID: "act-auto-lap", Analyzed: true, ICUIntervals: intervalsOut, Raw: map[string]any{}}
}

func rowsByEventID(series []any) map[string]map[string]any {
	out := map[string]map[string]any{}
	for _, item := range series {
		row := item.(map[string]any)
		out[row["event_id"].(string)] = row
	}
	return out
}

func wellnessFixture(date string, raw map[string]any) intervals.Wellness {
	if raw == nil {
		raw = map[string]any{}
	}
	raw["id"] = date
	return intervals.Wellness{Raw: raw}
}

func ptrString(value string) *string { return &value }

func assertFloatEqual(t *testing.T, got any, want float64) {
	t.Helper()
	value, ok := got.(float64)
	if !ok {
		t.Fatalf("value = %T(%v), want float64 %v", got, got, want)
	}
	if math.Abs(value-want) > 0.0001 {
		t.Fatalf("value = %v, want %v", value, want)
	}
}

func stringSliceFromAny(value any) []string {
	items, _ := value.([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func cappedActivities(sport string, raw map[string]any) []intervals.Activity {
	activities := make([]intervals.Activity, 0, maxComputeActivityCandidates)
	for idx := range maxComputeActivityCandidates {
		activityRaw := map[string]any{}
		for key, value := range raw {
			activityRaw[key] = value
		}
		activities = append(activities, computeActivityFixture(fmt.Sprintf("a-%03d", idx), sport, "2026-05-01T07:00:00", activityRaw))
	}
	return activities
}

func cappedActivitiesWithTime(sport string, movingTime int) []intervals.Activity {
	activities := make([]intervals.Activity, 0, maxComputeActivityCandidates)
	for idx := range maxComputeActivityCandidates {
		activities = append(activities, activityWithTime(fmt.Sprintf("act-%03d", idx), sport, "2026-05-01T07:00:00", movingTime, nil))
	}
	return activities
}

func cappedEvents(eventType string) []intervals.Event {
	events := make([]intervals.Event, 0, maxEventsLimit)
	for idx := range maxEventsLimit {
		events = append(events, eventFixture(fmt.Sprintf("evt-%03d", idx), eventType, "2026-05-01", intPtr(3600), nil, false))
	}
	return events
}

func assertTruncatedAnalyzerResult(t *testing.T, result Result, wantReason string) {
	t.Helper()
	got := resultMap(t, result)
	resultMap := got["result"].(map[string]any)
	meta := got["_meta"].(map[string]any)
	if status := resultMap["status"]; status != "partial" {
		t.Fatalf("status = %v, want partial", status)
	}
	if reason := resultMap["insufficient_reason"]; reason != wantReason {
		t.Fatalf("insufficient_reason = %v, want %s", reason, wantReason)
	}
	if truncated := resultMap["truncated_activity_candidates"]; truncated != true {
		t.Fatalf("truncated_activity_candidates = %v, want true", truncated)
	}
	assumptions := meta["assumptions"].(map[string]any)
	if assumptions["activity_candidates_truncated"] != true {
		t.Fatalf("assumptions = %v, want activity_candidates_truncated", assumptions)
	}
	if !stringSliceContains(stringSliceFromAny(meta["boundaries"]), "activity candidates truncated") {
		t.Fatalf("boundaries = %v, want truncation boundary", meta["boundaries"])
	}
}

func stringSliceContains(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
