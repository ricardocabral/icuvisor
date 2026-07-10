package tools

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeDataQualityReportClient struct {
	profile       intervals.AthleteWithSportSettings
	activities    []intervals.Activity
	summaries     []intervals.SummaryWithCats
	wellness      []intervals.Wellness
	events        []intervals.Event
	activityCalls []intervals.ListActivitiesParams
	eventCalls    []intervals.ListEventsParams
}

func (f *fakeDataQualityReportClient) GetAthleteProfile(context.Context) (intervals.AthleteWithSportSettings, error) {
	return f.profile, nil
}

func (f *fakeDataQualityReportClient) ListActivities(_ context.Context, params intervals.ListActivitiesParams) ([]intervals.Activity, error) {
	f.activityCalls = append(f.activityCalls, params)
	return append([]intervals.Activity(nil), f.activities...), nil
}

func (f *fakeDataQualityReportClient) ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	return append([]intervals.SummaryWithCats(nil), f.summaries...), nil
}

func (f *fakeDataQualityReportClient) ListWellness(context.Context, intervals.WellnessParams) ([]intervals.Wellness, error) {
	return append([]intervals.Wellness(nil), f.wellness...), nil
}

func (f *fakeDataQualityReportClient) ListEvents(_ context.Context, params intervals.ListEventsParams) ([]intervals.Event, error) {
	f.eventCalls = append(f.eventCalls, params)
	return append([]intervals.Event(nil), f.events...), nil
}

func TestGetDataQualityReportHealthyTerseReadOnlyContract(t *testing.T) {
	t.Parallel()

	client := &fakeDataQualityReportClient{
		profile: healthyDataQualityProfile(),
		activities: []intervals.Activity{
			decodeActivityFixture(t, `{"id":"ride-1","type":"Ride","start_date_local":"2026-01-07T07:00:00","source":"Garmin","stream_types":["time","heartrate","watts","cadence"]}`),
			decodeActivityFixture(t, `{"id":"run-1","type":"Run","start_date_local":"2026-01-05T07:00:00","source":"Garmin","stream_types":["time","heartrate","distance"]}`),
		},
		summaries: decodeSummaries(t, `[
			{"date":"2026-01-05","training_load":42,"fitness":50,"fatigue":45,"form":5,"byCategory":[{"category":"Run","training_load":42}]},
			{"date":"2026-01-07","training_load":55,"fitness":51,"fatigue":46,"form":5,"byCategory":[{"category":"Ride","training_load":55}]}
		]`),
		wellness: []intervals.Wellness{
			decodeWellnessRow(t, `{"id":"2026-01-07","updated":"2026-01-07T08:00:00Z","restingHR":48,"hrv":70,"sleepSecs":28800,"readiness":82}`),
		},
		events: decodeToolEvents(t, `{"id":"evt-1","category":"WORKOUT","type":"Ride","name":"Endurance","start_date_local":"2026-01-06"}`),
	}
	tool := newGetDataQualityReportTool(client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-07"}`)})
	if err != nil {
		t.Fatalf("get_data_quality_report error = %v", err)
	}
	payload := resultMap(t, result)
	if payload["full"] != nil {
		t.Fatalf("default payload should omit full evidence: %#v", payload["full"])
	}
	summary := payload["summary"].(map[string]any)
	if summary["status"] != "ok" {
		t.Fatalf("summary = %#v, want ok", summary)
	}
	meta := payload["_meta"].(map[string]any)
	streamPolicy, _ := meta["stream_policy"].(string)
	if meta["read_only"] != true || !strings.Contains(streamPolicy, "does not fetch raw stream samples") {
		t.Fatalf("meta = %#v, want read-only raw-stream-free policy", meta)
	}
	if len(client.activityCalls) != 1 || !slices.Equal(client.activityCalls[0].Fields, dataQualityActivityFields()) {
		t.Fatalf("activity fields = %#v, want explicit data-quality fields", client.activityCalls)
	}
	if len(client.eventCalls) != 1 || client.eventCalls[0].Newest != "2026-04-07" {
		t.Fatalf("event calls = %#v, want bounded future horizon", client.eventCalls)
	}
	if descriptor, ok := catalogDescriptor(getDataQualityReportName); !ok || descriptor.Safety != string(RequirementRead) {
		t.Fatalf("catalog descriptor = %#v present %v, want read-only tool", descriptor, ok)
	}
}

func TestGetDataQualityReportRequiredDegradedDiagnostics(t *testing.T) {
	t.Parallel()

	client := &fakeDataQualityReportClient{
		profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC", PreferredUnits: "metric", SportSettings: []intervals.SportSettings{{Types: []string{"Ride"}}}},
		activities: []intervals.Activity{
			decodeActivityFixture(t, `{"id":"strava-1","type":"Ride","start_date_local":"2026-01-10T07:00:00","source":"Strava","_note":"hidden"}`),
		},
		summaries: decodeSummaries(t, `[
			{"date":"2026-01-10","trimp":35,"byCategory":[{"category":"Ride","trimp":35}]},
			{"date":"2026-01-11","byCategory":[{"category":"Ride"}]}
		]`),
		wellness: []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-01-01","updated":"2025-12-30T00:00:00Z","readiness":70}`)},
	}
	tool := newGetDataQualityReportTool(client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-31"}`)})
	if err != nil {
		t.Fatalf("get_data_quality_report error = %v", err)
	}
	payload := resultMap(t, result)
	for _, code := range []string{"restricted_source", "missing_stream_metadata", "trimp_or_hr_load_available", "missing_training_load", "stale_wellness_bridge", "stale_wellness", "missing_power_threshold", "missing_power_zones", "sparse_activity_history"} {
		if !dataQualityDiagnosticsContain(payload, code) {
			t.Fatalf("diagnostics missing %s: %#v", code, payload["diagnostics"])
		}
	}
}

func TestGetDataQualityReportWellnessProvenanceStaleness(t *testing.T) {
	t.Parallel()

	client := &fakeDataQualityReportClient{
		profile:    healthyDataQualityProfile(),
		activities: []intervals.Activity{decodeActivityFixture(t, `{"id":"ride-1","type":"Ride","start_date_local":"2026-01-10T07:00:00","source":"Garmin","stream_types":["time","heartrate","watts"]}`)},
		summaries:  decodeSummaries(t, `[{"date":"2026-01-10","training_load":35,"fitness":40,"fatigue":41,"form":-1}]`),
		wellness:   []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-01-10","sleepScore":88,"source":"garmin","bridge_fetched_at":"2026-01-08T00:00:00Z"}`)},
		events:     decodeToolEvents(t, `{"id":"evt-1","category":"WORKOUT","start_date_local":"2026-01-10"}`),
	}
	tool := newGetDataQualityReportTool(client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-10"}`)})
	if err != nil {
		t.Fatalf("get_data_quality_report error = %v", err)
	}
	payload := resultMap(t, result)
	if !dataQualityDiagnosticsContain(payload, "stale_wellness_bridge") {
		t.Fatalf("diagnostics missing stale_wellness_bridge: %#v", payload["diagnostics"])
	}
}

func TestGetDataQualityReportStep2RegressionDiagnostics(t *testing.T) {
	t.Parallel()

	activities := make([]intervals.Activity, 0, dataQualityActivityFetchLimit)
	for i := range dataQualityActivityFetchLimit {
		activities = append(activities, decodeActivityFixture(t, `{"id":"ride-`+string(rune('a'+(i%26)))+`","type":"Ride","start_date_local":"2026-01-02T07:00:00","source":"Strava","_note":"hidden"}`))
	}
	client := &fakeDataQualityReportClient{
		profile: intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC", PreferredUnits: "metric", SportSettings: []intervals.SportSettings{
			{Types: []string{"Ride"}},
			{Types: []string{"Run"}, LTHR: 170, HRZones: []int{120, 140, 160}, ThresholdPace: 3.5714285, PaceUnits: "MINS_KM", PaceLoadType: "RUN", PaceZones: []float64{77.5, 90, 100}},
		}},
		activities: activities,
		summaries:  decodeSummaries(t, `[{"date":"2026-01-02","training_load":99,"fitness":10,"fatigue":11,"form":-1,"byCategory":[{"category":"Run","trimp":22}]}]`),
		wellness:   []intervals.Wellness{decodeWellnessRow(t, `{"id":"not-a-date","updated":"2025-12-30T00:00:00Z","readiness":70}`)},
		events:     cappedDataQualityEvents(),
	}
	tool := newGetDataQualityReportTool(client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-12-31","sport":"Run","include_full":true}`)})
	if err != nil {
		t.Fatalf("get_data_quality_report error = %v", err)
	}
	payload := resultMap(t, result)
	for _, code := range []string{"wellness_dates_unknown", "activity_probe_capped_before_sport_filter", "activity_probe_capped_before_stream_check", "activity_probe_capped_before_source_check", "calendar_future_probe_incomplete", "calendar_probe_capped", "trimp_or_hr_load_available"} {
		if !dataQualityDiagnosticsContain(payload, code) {
			t.Fatalf("diagnostics missing %s: %#v", code, payload["diagnostics"])
		}
	}
	if dataQualityDiagnosticsContain(payload, "missing_power_threshold") {
		t.Fatalf("sport=Run should not include Ride threshold warnings: %#v", payload["diagnostics"])
	}
	full := payload["full"].(map[string]any)
	if got := len(full["summary_dates"].([]any)); got > 60 {
		t.Fatalf("summary_dates len = %d, want bounded", got)
	}
}

func TestGetDataQualityReportAggregatesRestrictedSources(t *testing.T) {
	t.Parallel()

	activities := make([]intervals.Activity, 0, 7)
	for _, id := range []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"} {
		activities = append(activities, decodeActivityFixture(t, `{"id":"`+id+`","type":"Ride","start_date_local":"2026-01-02T07:00:00","source":"Strava","_note":"hidden"}`))
	}
	client := &fakeDataQualityReportClient{profile: healthyDataQualityProfile(), activities: activities, summaries: decodeSummaries(t, `[{"date":"2026-01-02","training_load":10,"fitness":10,"fatigue":10,"form":0}]`), wellness: []intervals.Wellness{decodeWellnessRow(t, `{"id":"2026-01-02","updated":"2026-01-02T00:00:00Z","readiness":70}`)}, events: decodeToolEvents(t, `{"id":"evt-1","category":"WORKOUT","start_date_local":"2026-01-02"}`)}
	tool := newGetDataQualityReportTool(client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-01-01","end_date":"2026-01-03"}`)})
	if err != nil {
		t.Fatalf("get_data_quality_report error = %v", err)
	}
	payload := resultMap(t, result)
	diagnostics := payload["diagnostics"].([]any)
	count := 0
	for _, raw := range diagnostics {
		diagnostic := raw.(map[string]any)
		if diagnostic["code"] != "restricted_source" {
			continue
		}
		count++
		evidence := diagnostic["evidence"].(map[string]any)
		if len(evidence["sample"].([]any)) != 5 || evidence["restricted_source_count"] != float64(7) {
			t.Fatalf("restricted evidence = %#v, want aggregate count and bounded sample", evidence)
		}
	}
	if count != 1 {
		t.Fatalf("restricted_source diagnostics = %d, want one aggregate", count)
	}
}

func healthyDataQualityProfile() intervals.AthleteWithSportSettings {
	return intervals.AthleteWithSportSettings{ID: "i12345", Timezone: "UTC", PreferredUnits: "metric", SportSettings: []intervals.SportSettings{
		{Types: []string{"Ride"}, FTP: 250, PowerZones: []int{100, 150, 200}, LTHR: 170, HRZones: []int{120, 140, 160}},
		{Types: []string{"Run"}, LTHR: 170, HRZones: []int{120, 140, 160}, ThresholdPace: 3.5714285, PaceUnits: "MINS_KM", PaceLoadType: "RUN", PaceZones: []float64{77.5, 90, 100}},
	}}
}

func cappedDataQualityEvents() []intervals.Event {
	events := make([]intervals.Event, 0, dataQualityEventFetchLimit)
	for i := range dataQualityEventFetchLimit {
		id := "evt"
		date := "2026-01-02"
		if i > 400 {
			date = "2026-12-30"
		}
		event := intervals.Event{ID: id, Category: ptrString("WORKOUT"), Type: ptrString("Ride"), StartDateLocal: ptrString(date), Raw: map[string]any{}}
		events = append(events, event)
	}
	return events
}

func dataQualityDiagnosticsContain(payload map[string]any, code string) bool {
	diagnostics, ok := payload["diagnostics"].([]any)
	if !ok {
		return false
	}
	for _, raw := range diagnostics {
		diagnostic, ok := raw.(map[string]any)
		if ok && diagnostic["code"] == code {
			return true
		}
	}
	return false
}

func catalogDescriptor(name string) (ToolDescriptor, bool) {
	for _, descriptor := range Catalog() {
		if descriptor.Name == name {
			return descriptor, true
		}
	}
	return ToolDescriptor{}, false
}
