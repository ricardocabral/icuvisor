package tools

import (
	"context"
	"encoding/json"
	"errors"
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
	series, err := loadAnalyzerSeries(context.Background(), analyzerClients{fitness: client}, analysis.Metric("weekly_tss"), window, analysis.SampleGrainDaily, "", response.UnitSystemMetric, true)
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
