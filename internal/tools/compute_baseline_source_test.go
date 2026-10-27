package tools

import (
	"context"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestComputeBaselineHRVSourceContinuity(t *testing.T) {
	for _, tc := range []struct {
		name           string
		metric         string
		first          string
		second         string
		current        string
		wantStatus     string
		wantContinuity string
	}{
		{name: "provider changes across windows", metric: "hrv", first: `"source":"oura"`, second: `"source":"oura"`, current: `"source":"whoop"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed"},
		{name: "device changes within same provider", metric: "hrv", first: `"source":"polar","device":"Polar H10"`, second: `"source":"polar","device":"Polar H10"`, current: `"source":"polar","device":"Polar Vantage"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed"},
		{name: "source changes inside baseline", metric: "hrv_sdnn", first: `"source":"apple_health"`, second: `"source":"oura"`, current: `"source":"oura"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed"},
		{name: "same explicit source", metric: "hrv", first: `"source":"oura"`, second: `"source":"oura"`, current: `"source":"oura"`, wantStatus: "ok", wantContinuity: "consistent"},
		{name: "missing provenance is unknown", metric: "hrv", first: `"unrelated":true`, second: `"unrelated":true`, current: `"unrelated":true`, wantStatus: "ok", wantContinuity: "unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := newFakeComputeClient()
			client.wellness = []intervals.Wellness{
				decodeWellnessRow(t, `{"id":"2026-05-01","hrv":40,"hrvSDNN":40,`+tc.first+`}`),
				decodeWellnessRow(t, `{"id":"2026-05-02","hrv":60,"hrvSDNN":60,`+tc.second+`}`),
				decodeWellnessRow(t, `{"id":"2026-05-03","hrv":90,"hrvSDNN":90,`+tc.current+`}`),
			}
			tool := newComputeBaselineTool(nil, client, nil, nil, client, "test", "UTC", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"` + tc.metric + `","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-03","current_end_date":"2026-05-03","min_samples":2}`)})
			if err != nil {
				t.Fatal(err)
			}
			body := resultMap(t, result)["result"].(map[string]any)
			if body["status"] != tc.wantStatus {
				t.Fatalf("status = %v, want %s", body["status"], tc.wantStatus)
			}
			continuity, ok := body["source_continuity"].(map[string]any)
			if !ok || continuity["status"] != tc.wantContinuity {
				t.Fatalf("source_continuity = %#v, want %s", body["source_continuity"], tc.wantContinuity)
			}
			if tc.wantContinuity == "changed" {
				for _, field := range []string{"z_score", "baseline_mean", "baseline_stddev", "current_value"} {
					if _, ok := body[field]; ok {
						t.Fatalf("mixed-source result exposes %s: %#v", field, body)
					}
				}
				if body["interpretation"] != "not_interpreted" || body["insufficient_reason"] != "hrv_source_changed" {
					t.Fatalf("mixed-source verdict = %#v", body)
				}
			}
		})
	}
}

func TestComputeBaselineNeverUsesSDNNAsRMSSDFallback(t *testing.T) {
	client := newFakeComputeClient()
	client.wellness = []intervals.Wellness{
		decodeWellnessRow(t, `{"id":"2026-05-01","hrv":40}`),
		decodeWellnessRow(t, `{"id":"2026-05-02","hrv":60}`),
		decodeWellnessRow(t, `{"id":"2026-05-03","hrvSDNN":90}`),
	}
	tool := newComputeBaselineTool(nil, client, nil, nil, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"hrv","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-03","current_end_date":"2026-05-03","min_samples":2}`)})
	if err != nil {
		t.Fatal(err)
	}
	body := resultMap(t, result)["result"].(map[string]any)
	if body["status"] != "insufficient_current_sample" || body["n_current"] != float64(0) {
		t.Fatalf("SDNN substituted for RMSSD: %#v", body)
	}
}
