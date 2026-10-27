package tools

import (
	"context"
	"slices"
	"strings"
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
		wantSources    []string
		wantUnknown    int
	}{
		{name: "provider changes across windows", metric: "hrv", first: `"source":"oura"`, second: `"source":"oura"`, current: `"source":"whoop"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantSources: []string{"oura", "whoop"}},
		{name: "device changes within same provider", metric: "hrv", first: `"source":"polar","device":"Polar H10"`, second: `"source":"polar","device":"Polar H10"`, current: `"source":"polar","device":"Polar Vantage"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantSources: []string{"polar"}},
		{name: "source changes inside baseline", metric: "hrv_sdnn", first: `"source":"apple_health"`, second: `"source":"oura"`, current: `"source":"oura"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantSources: []string{"apple_health", "oura"}},
		{name: "same explicit source", metric: "hrv", first: `"source":"oura"`, second: `"source":"oura"`, current: `"source":"oura"`, wantStatus: "ok", wantContinuity: "consistent", wantSources: []string{"oura"}},
		{name: "missing provenance is unknown", metric: "hrv", first: `"unrelated":true`, second: `"unrelated":true`, current: `"unrelated":true`, wantStatus: "ok", wantContinuity: "unknown", wantUnknown: 3},
		{name: "sleep sidecar cannot mask explicit source transition", metric: "hrv", first: `"source":"oura","polar_sleep_score":80`, second: `"source":"oura","polar_sleep_score":80`, current: `"source":"whoop","polar_sleep_score":80`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantSources: []string{"oura", "whoop"}},
		{name: "sleep sidecars do not identify HRV source", metric: "hrv", first: `"polar_sleep_score":80`, second: `"polar_sleep_score":82`, current: `"polar_sleep_score":83`, wantStatus: "ok", wantContinuity: "unknown", wantUnknown: 3},
		{name: "readiness sidecars do not identify SDNN source", metric: "hrv_sdnn", first: `"whoop_recovery_score":80`, second: `"whoop_recovery_score":82`, current: `"whoop_recovery_score":83`, wantStatus: "ok", wantContinuity: "unknown", wantUnknown: 3},
		{name: "arbitrary explicit providers change", metric: "hrv", first: `"provider":"kubios"`, second: `"provider":"kubios"`, current: `"provider":"elite_hrv"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantSources: []string{"elite_hrv", "kubios"}},
		{name: "provider identity only normalizes case and whitespace", metric: "hrv", first: `"provider":" KuBIOS "`, second: `"provider":"kubios"`, current: `"provider":"KUBIOS"`, wantStatus: "ok", wantContinuity: "consistent", wantSources: []string{"kubios"}},
		{name: "provider names are not collapsed to vendor substrings", metric: "hrv", first: `"source":"oura ring 3"`, second: `"source":"oura ring 3"`, current: `"source":"oura ring 4"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantSources: []string{"oura ring 3", "oura ring 4"}},
		{name: "explicit provider takes precedence over readiness sidecar", metric: "hrv", first: `"provider":"kubios","whoop_recovery_score":80`, second: `"provider":"kubios","whoop_recovery_score":82`, current: `"provider":"kubios","whoop_recovery_score":83`, wantStatus: "ok", wantContinuity: "consistent", wantSources: []string{"kubios"}},
		{name: "device-only provenance stays unknown without exposing identifiers", metric: "hrv", first: `"device":" sensor-1 "`, second: `"device":"SENSOR-1"`, current: `"device":"sensor-1"`, wantStatus: "ok", wantContinuity: "unknown", wantUnknown: 3},
		{name: "device-only transition does not expose device identifiers", metric: "hrv", first: `"device":"sensor-1"`, second: `"device":"sensor-1"`, current: `"device":"sensor-2"`, wantStatus: "insufficient_comparable_data", wantContinuity: "changed", wantUnknown: 3},
		{name: "partially missing provenance stays unknown", metric: "hrv", first: `"provider":"kubios"`, second: `"provider":"kubios"`, current: `"polar_sleep_score":83`, wantStatus: "ok", wantContinuity: "unknown", wantSources: []string{"kubios"}, wantUnknown: 1},
		{name: "unknown is not a provider identity", metric: "hrv", first: `"source":"unknown"`, second: `"source":"unknown"`, current: `"source":"unknown"`, wantStatus: "ok", wantContinuity: "unknown", wantUnknown: 3},
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
			if got := stringSliceFromAny(continuity["sources"]); !slices.Equal(got, tc.wantSources) {
				t.Fatalf("sources = %#v, want %#v", got, tc.wantSources)
			}
			if continuity["unknown_source_samples"] != float64(tc.wantUnknown) {
				t.Fatalf("unknown_source_samples = %v, want %d", continuity["unknown_source_samples"], tc.wantUnknown)
			}
			if strings.Contains(result.Content[0].Text, "sensor-") {
				t.Fatalf("device identifier exposed in response: %s", result.Content[0].Text)
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
