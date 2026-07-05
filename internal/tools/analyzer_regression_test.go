package tools

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/response"
)

func TestAggregateActivityDayDerivesPaceFromTotalDistanceAndTime(t *testing.T) {
	km1, km9 := 1000.0, 9000.0
	sec300, sec1800 := 300, 1800
	day, ok, assumptions := aggregateActivityDay("2026-05-01", []intervals.Activity{
		{ID: "a1", Distance: &km1, MovingTime: &sec300},
		{ID: "a2", Distance: &km9, MovingTime: &sec1800},
	}, analysis.Metric("pace_seconds_per_km"), response.UnitSystemMetric)
	if !ok || day.Value != 210 {
		t.Fatalf("daily pace = (%#v, %v), want 210 sec/km", day, ok)
	}
	if assumptions["aggregation"] != "total_moving_time_per_total_distance" {
		t.Fatalf("aggregation = %#v", assumptions)
	}
}

func TestLoadAnalyzerSeriesLoadsDerivedWeeklyMetrics(t *testing.T) {
	client := &fakeFitnessMetricsClient{summaries: decodeSummaries(t, `[
		{"date":"2026-05-01","training_load":10,"time":3600},
		{"date":"2026-05-02","training_load":20,"time":7200},
		{"date":"2026-05-15","training_load":30,"time":1800}
	]`)}
	window, err := analysis.ParseWindow(analysis.Window{StartDate: "2026-05-01", EndDate: "2026-05-21"}, 366)
	if err != nil {
		t.Fatalf("parse window: %v", err)
	}
	series, err := loadAnalyzerSeries(context.Background(), analyzerClients{fitness: client}, analysis.Metric("weekly_tss"), window, analysis.SampleGrainDaily, "", response.UnitSystemMetric, nil, true)
	if err != nil {
		t.Fatalf("load weekly series: %v", err)
	}
	if len(series.Samples) != 2 || series.Samples[0].Bucket != 0 || series.Samples[0].Value != 30 || series.Samples[1].Bucket != 2 || series.Samples[1].Value != 30 {
		t.Fatalf("weekly samples = %#v", series.Samples)
	}
	if series.MissingDays != 18 || series.Assumptions["sample_grain"] != string(analysis.SampleGrainWeekly) {
		t.Fatalf("weekly metadata missing=%d assumptions=%#v", series.MissingDays, series.Assumptions)
	}
}

func TestAnalyzeCorrelationSupportsVO2MaxLikeCustomFieldAgainstCTL(t *testing.T) {
	t.Parallel()

	fitness := &fakeFitnessMetricsClient{summaries: decodeSummaries(t, `[
		{"date":"2026-05-01","fitness":60},
		{"date":"2026-05-02","fitness":61},
		{"date":"2026-05-03","fitness":63}
	]`)}
	activities := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride 1","type":"Ride","start_date_local":"2026-05-01T07:00:00","moving_time":3600,"vo2max_est":51.2}`,
		`{"id":"a2","name":"Ride 2","type":"Ride","start_date_local":"2026-05-02T07:00:00","moving_time":4200,"vo2max_est":52.1}`,
		`{"id":"a3","name":"Ride 3","type":"Ride","start_date_local":"2026-05-03T07:00:00","moving_time":3900,"vo2max_est":53.0}`,
	}, "metric")
	activities.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newAnalyzeCorrelationTool(fitness, nil, activities, activities, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric_x":"custom:vo2max_est","metric_y":"ctl","window":{"start_date":"2026-05-01","end_date":"2026-05-03"},"custom_fields":["vo2max_est"],"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	body := payload["result"].(map[string]any)
	if body["metric_x"] != "custom:vo2max_est" || body["metric_y"] != "ctl" || body["n"] != float64(3) {
		t.Fatalf("result = %#v, want custom metric paired with ctl n=3", body)
	}
	if len(payload["series"].([]any)) != 3 {
		t.Fatalf("series = %#v, want full paired rows", payload["series"])
	}
	meta := payload["_meta"].(map[string]any)
	tools := meta["source_tools"].([]any)
	if !slices.Contains(tools, any("get_activities.custom_fields.vo2max_est")) {
		t.Fatalf("source_tools = %#v, want custom field provenance", tools)
	}
	seriesAssumptions := meta["assumptions"].(map[string]any)["series"].(map[string]any)["custom:vo2max_est"].(map[string]any)
	if seriesAssumptions["custom_field_code"] != "vo2max_est" || seriesAssumptions["aggregation"] != "moving_time_weighted_mean" {
		t.Fatalf("custom assumptions = %#v", seriesAssumptions)
	}
}

func TestAnalyzeCorrelationCustomFieldInsufficientDataMetadata(t *testing.T) {
	t.Parallel()

	activities := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride 1","type":"Ride","start_date_local":"2026-05-01T07:00:00","icu_training_load":50,"vo2max_est":51.2}`,
		`{"id":"a2","name":"Ride 2","type":"Ride","start_date_local":"2026-05-02T07:00:00","icu_training_load":70,"vo2max_est":"not numeric"}`,
		`{"id":"a3","name":"Ride 3","type":"Ride","start_date_local":"2026-05-03T07:00:00","icu_training_load":65}`,
	}, "metric")
	activities.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newAnalyzeCorrelationTool(nil, nil, activities, activities, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric_x":"custom:vo2max_est","metric_y":"training_load","window":{"start_date":"2026-05-01","end_date":"2026-05-03"},"pairing_grain":"activity","custom_fields":["vo2max_est"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	if meta["insufficient_sample"] != true || meta["n"] != float64(1) {
		t.Fatalf("meta = %#v, want insufficient sample n=1", meta)
	}
	seriesAssumptions := meta["assumptions"].(map[string]any)["series"].(map[string]any)["custom:vo2max_est"].(map[string]any)
	if seriesAssumptions["skipped_non_numeric"] != float64(1) || seriesAssumptions["missing_field_samples"] != float64(1) {
		t.Fatalf("custom assumptions = %#v", seriesAssumptions)
	}
	if _, ok := resultMap(t, result)["series"]; ok {
		t.Fatal("terse response included paired series without include_full")
	}
}

func TestAnalyzeCorrelationFetchesExplicitCustomFieldsForActivityMetrics(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Ride 1","type":"Ride","start_date_local":"2026-05-01T07:00:00","icu_training_load":50,"moving_time":3600,"vo2max_est":51.2}`,
		`{"id":"a2","name":"Ride 2","type":"Ride","start_date_local":"2026-05-02T07:00:00","icu_training_load":70,"moving_time":4200,"vo2max_est":52.1}`,
		`{"id":"a3","name":"Ride 3","type":"Ride","start_date_local":"2026-05-03T07:00:00","icu_training_load":65,"moving_time":3900,"vo2max_est":51.8}`,
	}, "metric")
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newAnalyzeCorrelationTool(nil, nil, client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric_x":"training_load","metric_y":"moving_time_seconds","window":{"start_date":"2026-05-01","end_date":"2026-05-03"},"pairing_grain":"activity","custom_fields":["vo2max_est"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.listCalls) == 0 || !slices.Contains(client.listCalls[0].Fields, "vo2max_est") {
		t.Fatalf("ListActivities fields = %#v, want explicit custom field", client.listCalls)
	}
}

func TestAnalyzeCorrelationUnknownCustomFieldUsesPublicHint(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"vo2max_est"}}`)
	tool := newAnalyzeCorrelationTool(nil, nil, client, client, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric_x":"training_load","metric_y":"moving_time_seconds","window":{"start_date":"2026-05-01","end_date":"2026-05-03"},"custom_fields":["missing_vo2"]}`)})
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "missing_vo2") || !strings.Contains(message, "vo2max_est") {
		t.Fatalf("public error = %q/%v; err=%v", message, ok, err)
	}
}

func TestAnalyzeDistributionRejectsBucketAndQuantileContractViolations(t *testing.T) {
	tool := newAnalyzeDistributionTool(nil, nil, nil, nil, "test", "UTC", false)
	cases := []string{
		`{"metric":"ctl","window":{"start_date":"2026-05-01","end_date":"2026-05-07"},"bucket_count":5,"buckets":[1,2]}`,
		`{"metric":"ctl","window":{"start_date":"2026-05-01","end_date":"2026-05-07"},"quantiles":[-0.1]}`,
		`{"metric":"ctl","window":{"start_date":"2026-05-01","end_date":"2026-05-07"},"quantiles":[1.1]}`,
	}
	for _, args := range cases {
		t.Run(args, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
			if message, ok := PublicErrorMessage(err); !ok || message != invalidAnalyzeDistributionArgs {
				t.Fatalf("Handler() error = %v, public=%q ok=%v", err, message, ok)
			}
		})
	}
}

func TestAnalyzeTrendHRVSDNNCurrentWindowFreshnessCaveat(t *testing.T) {
	client := newFakeComputeClient()
	client.wellness = []intervals.Wellness{
		decodeWellnessRow(t, `{"id":"2026-04-24","hrvSDNN":51,"source":"oura","bridge_fetched_at":"2026-04-24T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-04-25","hrvSDNN":52,"source":"oura","bridge_fetched_at":"2026-04-25T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-04-26","hrvSDNN":53,"source":"oura","bridge_fetched_at":"2026-04-26T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-04-27","hrvSDNN":54,"source":"oura","bridge_fetched_at":"2026-04-27T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-04-28","hrvSDNN":55,"source":"oura","bridge_fetched_at":"2026-04-28T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-04-29","hrvSDNN":56,"source":"oura","bridge_fetched_at":"2026-04-29T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-04-30","hrvSDNN":57,"source":"oura","bridge_fetched_at":"2026-04-30T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-05-01","hrvSDNN":47,"source":"oura","bridge_fetched_at":"2026-05-01T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-05-02","hrvSDNN":46,"source":"oura","bridge_fetched_at":"2026-05-02T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-05-03","hrvSDNN":45,"source":"oura","bridge_fetched_at":"2026-05-03T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-05-04","hrvSDNN":44,"source":"oura","bridge_fetched_at":"2026-05-04T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-05-05","hrvSDNN":43,"source":"oura","bridge_fetched_at":"2026-05-05T08:00:00Z"}`),
		decodeWellnessRow(t, `{"id":"2026-05-06","hrvSDNN":42,"source":"oura","bridge_fetched_at":"2026-05-06T08:00:00Z"}`),
	}
	tool := newAnalyzeTrendTool(nil, client, nil, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric":"hrv_sdnn","window":{"start_date":"2026-05-01","end_date":"2026-05-07"},"baseline_window":{"start_date":"2026-04-24","end_date":"2026-04-30"},"rolling_window_days":3,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	body := payload["result"].(map[string]any)
	meta := payload["_meta"].(map[string]any)
	if body["freshness_status"] != "stale_current_window" || body["latest_sample_date"] != "2026-05-06" || body["window_end_date"] != "2026-05-07" {
		t.Fatalf("trend freshness fields = %#v, want stale_current_window latest/end dates", body)
	}
	caveats := joinedStrings(body["caveats"].([]any))
	if !strings.Contains(caveats, "latest HRV-SDNN sample 2026-05-06 predates trend window end 2026-05-07") || !strings.Contains(caveats, "1 missing day") {
		t.Fatalf("caveats = %q, want latest-sample and missing-day wording", caveats)
	}
	if !slices.Contains(stringSliceFromAny(meta["source_tools"]), getWellnessDataName) {
		t.Fatalf("source_tools = %#v, want get_wellness_data", meta["source_tools"])
	}
	freshness := meta["assumptions"].(map[string]any)["wellness_freshness"].(map[string]any)
	if freshness["status"] != "stale_current_window" || freshness["latest_sample_date"] != "2026-05-06" || freshness["window_end_date"] != "2026-05-07" {
		t.Fatalf("wellness_freshness = %#v", freshness)
	}
}

func TestAnalyzeTrendHRVStaleProvenanceVisible(t *testing.T) {
	client := newFakeComputeClient()
	client.wellness = []intervals.Wellness{
		decodeWellnessRow(t, `{"id":"2026-04-24","hrv":81,"source":"garmin","bridge_fetched_at":"2026-04-24T08:00:00Z","garmin":{"hrv":81}}`),
		decodeWellnessRow(t, `{"id":"2026-04-25","hrv":82,"source":"garmin","bridge_fetched_at":"2026-04-25T08:00:00Z","garmin":{"hrv":82}}`),
		decodeWellnessRow(t, `{"id":"2026-04-26","hrv":83,"source":"garmin","bridge_fetched_at":"2026-04-26T08:00:00Z","garmin":{"hrv":83}}`),
		decodeWellnessRow(t, `{"id":"2026-04-27","hrv":84,"source":"garmin","bridge_fetched_at":"2026-04-27T08:00:00Z","garmin":{"hrv":84}}`),
		decodeWellnessRow(t, `{"id":"2026-04-28","hrv":85,"source":"garmin","bridge_fetched_at":"2026-04-28T08:00:00Z","garmin":{"hrv":85}}`),
		decodeWellnessRow(t, `{"id":"2026-04-29","hrv":86,"source":"garmin","bridge_fetched_at":"2026-04-29T08:00:00Z","garmin":{"hrv":86}}`),
		decodeWellnessRow(t, `{"id":"2026-04-30","hrv":87,"source":"garmin","bridge_fetched_at":"2026-04-30T08:00:00Z","garmin":{"hrv":87}}`),
		decodeWellnessRow(t, `{"id":"2026-05-01","hrv":76,"source":"garmin","bridge_fetched_at":"2026-05-01T08:00:00Z","garmin":{"hrv":76}}`),
		decodeWellnessRow(t, `{"id":"2026-05-02","hrv":77,"source":"garmin","bridge_fetched_at":"2026-05-02T08:00:00Z","garmin":{"hrv":77}}`),
		decodeWellnessRow(t, `{"id":"2026-05-03","hrv":78,"source":"garmin","bridge_fetched_at":"2026-05-01T00:00:00Z","garmin":{"hrv":78}}`),
		decodeWellnessRow(t, `{"id":"2026-05-04","hrv":79,"source":"garmin","bridge_fetched_at":"2026-05-04T08:00:00Z","garmin":{"hrv":79}}`),
		decodeWellnessRow(t, `{"id":"2026-05-05","hrv":80,"source":"garmin","bridge_fetched_at":"2026-05-05T08:00:00Z","garmin":{"hrv":80}}`),
		decodeWellnessRow(t, `{"id":"2026-05-06","hrv":81,"source":"garmin","bridge_fetched_at":"2026-05-06T08:00:00Z","garmin":{"hrv":81}}`),
		decodeWellnessRow(t, `{"id":"2026-05-07","hrv":82,"source":"garmin","bridge_fetched_at":"2026-05-07T08:00:00Z","garmin":{"hrv":82}}`),
	}
	tool := newAnalyzeTrendTool(nil, client, nil, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric":"hrv","window":{"start_date":"2026-05-01","end_date":"2026-05-07"},"baseline_window":{"start_date":"2026-04-24","end_date":"2026-04-30"},"rolling_window_days":3}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	body := payload["result"].(map[string]any)
	if body["freshness_status"] != "stale_provenance" {
		t.Fatalf("freshness_status = %#v, want stale_provenance in result %#v", body["freshness_status"], body)
	}
	caveats := joinedStrings(body["caveats"].([]any))
	if !strings.Contains(caveats, "garmin bridge data for HRV sample 2026-05-03 is older than 24h") {
		t.Fatalf("caveats = %q, want stale provenance wording", caveats)
	}
	freshness := payload["_meta"].(map[string]any)["assumptions"].(map[string]any)["wellness_freshness"].(map[string]any)
	staleRows := freshness["stale_sample_dates"].([]any)
	if !slices.Contains(staleRows, any("2026-05-03")) || freshness["stale_source"] != "garmin" {
		t.Fatalf("wellness_freshness = %#v, want stale 2026-05-03 garmin", freshness)
	}
}

func TestAnalyzeTrendPropagatesBaselineCancellation(t *testing.T) {
	client := &cancelSecondSummaryClient{rows: decodeSummaries(t, `[
		{"date":"2026-05-01","fitness":70},{"date":"2026-05-02","fitness":71},{"date":"2026-05-03","fitness":72},{"date":"2026-05-04","fitness":73},{"date":"2026-05-05","fitness":74},{"date":"2026-05-06","fitness":75},{"date":"2026-05-07","fitness":76}
	]`)}
	profile := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	tool := newAnalyzeTrendTool(client, nil, nil, profile, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric":"ctl","window":{"start_date":"2026-05-01","end_date":"2026-05-07"},"baseline_window":{"start_date":"2026-04-24","end_date":"2026-04-30"}}`)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Handler() error = %v, want context.Canceled", err)
	}
}

func TestAnalyzeTrendReportsEffectiveDefaultRollingWindow(t *testing.T) {
	client := &fakeFitnessMetricsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}, summaries: decodeSummaries(t, `[
		{"date":"2026-05-01","fitness":70},{"date":"2026-05-02","fitness":71},{"date":"2026-05-03","fitness":72},{"date":"2026-05-04","fitness":73},{"date":"2026-05-05","fitness":74},{"date":"2026-05-06","fitness":75},{"date":"2026-05-07","fitness":76}
	]`)}
	tool := newAnalyzeTrendTool(client, nil, nil, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"metric":"ctl","window":{"start_date":"2026-05-01","end_date":"2026-05-07"}}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	assumptions := meta["assumptions"].(map[string]any)
	if assumptions["rolling_window_days"] != float64(7) {
		t.Fatalf("rolling_window_days = %#v, want effective default 7; assumptions=%#v", assumptions["rolling_window_days"], assumptions)
	}
}

type cancelSecondSummaryClient struct {
	rows  []intervals.SummaryWithCats
	calls int
}

func (c *cancelSecondSummaryClient) ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	c.calls++
	if c.calls == 2 {
		return nil, context.Canceled
	}
	return append([]intervals.SummaryWithCats(nil), c.rows...), nil
}
