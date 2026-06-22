package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestGetFitnessToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	rows := got["fitness"].([]any)
	if rows[0].(map[string]any)["date"] != "2026-05-01" || rows[1].(map[string]any)["date"] != "2026-05-02" {
		t.Fatalf("fitness row order = %#v", rows)
	}
	meta := got["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" || meta["server_version"] != "test" {
		t.Fatalf("meta = %#v", meta)
	}
	if _, ok := got["per_sport_load_trends"]; ok {
		t.Fatalf("default response unexpectedly included per_sport_load_trends: %#v", got)
	}
	if len(client.summaryCalls) != 1 || client.summaryCalls[0].Start != "2026-05-01" || client.summaryCalls[0].End != "2026-05-02" {
		t.Fatalf("summary calls = %#v", client.summaryCalls)
	}
}

func TestGetFitnessPerSportLoadTrends(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.summaries = decodeSummaries(t, `[
		{"date":"2026-05-02","fitness":72,"fatigue":80,"form":-8,"training_load":125,"byCategory":[{"category":"Ride","training_load":90},{"category":"Open Water Swim","training_load":35}]},
		{"date":"2026-05-01","fitness":70,"fatigue":78,"form":-8,"training_load":55,"byCategory":[{"category":"Trail Run","training_load":40},{"category":"Strength","training_load":15}]}
	]`)
	tool := newGetFitnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-02","include_per_sport_load_trends":true,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	if len(client.summaryCalls) != 1 || client.summaryCalls[0].Start != "2026-02-06" || client.summaryCalls[0].End != "2026-05-02" {
		t.Fatalf("summary calls = %#v", client.summaryCalls)
	}
	buckets := perSportBuckets(t, got)
	assertTrendLoad(t, buckets, "running", "2026-05-01", 40)
	assertTrendLoad(t, buckets, "cycling", "2026-05-02", 90)
	assertTrendLoad(t, buckets, "swimming", "2026-05-02", 35)
	assertTrendLoad(t, buckets, "other", "2026-05-01", 15)
	assertTrendValue(t, buckets, "running", "2026-05-01", "ctl", 0.952)
	assertTrendValue(t, buckets, "running", "2026-05-01", "atl", 5.714)
	assertTrendValue(t, buckets, "running", "2026-05-01", "tsb", -4.762)
	if sumTrendLoadForDate(buckets, "2026-05-02") != 125 {
		t.Fatalf("per-sport load sum on 2026-05-02 = %v, want 125", sumTrendLoadForDate(buckets, "2026-05-02"))
	}
	meta := got["_meta"].(map[string]any)["per_sport_load_trends"].(map[string]any)
	if meta["method"] != perSportLoadTrendMethod || meta["warmup_summary_days_available"].(float64) != 0 {
		t.Fatalf("per-sport meta = %#v", meta)
	}
	categories := meta["source_categories_by_bucket"].(map[string]any)
	if !anySliceContains(categories["running"].([]any), "Trail Run") || !anySliceContains(categories["swimming"].([]any), "Open Water Swim") {
		t.Fatalf("source categories = %#v", categories)
	}
}

func TestGetFitnessPerSportLoadTrendCaveatsAndDateGaps(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.summaries = decodeSummaries(t, `[
		{"date":"2026-05-01","fitness":70,"fatigue":78,"form":-8,"training_load":50,"byCategory":[]},
		{"date":"2026-05-03","fitness":72,"fatigue":80,"form":-8,"training_load":60,"byCategory":[{"category":"Run"}]}
	]`)
	tool := newGetFitnessTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-03","include_per_sport_load_trends":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	buckets := perSportBuckets(t, got)
	if len(buckets["running"]) != 3 {
		t.Fatalf("running trend rows = %#v, want every requested date", buckets["running"])
	}
	assertTrendLoad(t, buckets, "running", "2026-05-02", 0)
	meta := got["_meta"].(map[string]any)["per_sport_load_trends"].(map[string]any)
	missing := meta["missing_requested_dates"].([]any)
	if len(missing) != 1 || missing[0] != "2026-05-02" {
		t.Fatalf("missing_requested_dates = %#v", missing)
	}
	caveats := joinedStrings(meta["caveats"].([]any))
	for _, want := range []string{"no byCategory sport breakdown", "omit training_load", "totals differ", "absent from upstream summary rows", "fewer than 84 warm-up"} {
		if !strings.Contains(caveats, want) {
			t.Fatalf("caveats %q missing %q", caveats, want)
		}
	}
}

func perSportBuckets(t *testing.T, got map[string]any) map[string][]map[string]any {
	t.Helper()
	out := map[string][]map[string]any{}
	for _, bucketRaw := range got["per_sport_load_trends"].([]any) {
		bucket := bucketRaw.(map[string]any)
		for _, rowRaw := range bucket["rows"].([]any) {
			out[bucket["sport"].(string)] = append(out[bucket["sport"].(string)], rowRaw.(map[string]any))
		}
	}
	return out
}

func trendRow(t *testing.T, buckets map[string][]map[string]any, sport string, date string) map[string]any {
	t.Helper()
	for _, row := range buckets[sport] {
		if row["date"] == date {
			return row
		}
	}
	t.Fatalf("missing trend row for %s %s in %#v", sport, date, buckets[sport])
	return nil
}

func assertTrendLoad(t *testing.T, buckets map[string][]map[string]any, sport string, date string, want float64) {
	t.Helper()
	got := trendRow(t, buckets, sport, date)["training_load"].(float64)
	if got != want {
		t.Fatalf("%s %s training_load = %v, want %v", sport, date, got, want)
	}
}

func assertTrendValue(t *testing.T, buckets map[string][]map[string]any, sport string, date string, field string, want float64) {
	t.Helper()
	got := trendRow(t, buckets, sport, date)[field].(float64)
	if got != want {
		t.Fatalf("%s %s %s = %v, want %v", sport, date, field, got, want)
	}
}

func sumTrendLoadForDate(buckets map[string][]map[string]any, date string) float64 {
	var sum float64
	for sport := range buckets {
		for _, row := range buckets[sport] {
			if row["date"] == date {
				sum += row["training_load"].(float64)
			}
		}
	}
	return sum
}

func anySliceContains(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func joinedStrings(values []any) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, value.(string))
	}
	return strings.Join(parts, "\n")
}
