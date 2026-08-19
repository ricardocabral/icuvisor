package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetActivitySplitsVirtualRowsIncludeAlignedMetricsAndProvenance(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
		streams: decodeStreamFixtures(t,
			`{"type":"distance","data":[0,500,1500,2000]}`,
			`{"type":"time","data":[0,60,180,240]}`,
			`{"type":"heart_rate","data":[100,110,130,140]}`,
			`{"type":"watts","data":[200,220,260,280]}`,
			`{"type":"cadence","data":[80,82,86,88]}`,
			`{"type":"altitude","data":[10,15,12,20]}`,
		),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	rows := payload["splits"].([]any)
	if len(rows) != 2 {
		t.Fatalf("splits = %#v, want two rows", rows)
	}
	first := rows[0].(map[string]any)
	if first["provenance"] != "virtual_fixed_distance" || first["distance_basis"] != "fixed_distance_boundary" {
		t.Fatalf("first provenance = %#v, want fixed-distance virtual provenance", first)
	}
	for _, field := range []string{"average_heart_rate_bpm", "average_power_watts", "average_cadence_rpm", "elevation_gain_m"} {
		if _, ok := first[field]; !ok {
			t.Fatalf("first row = %#v, want %s", first, field)
		}
	}
	if first["average_heart_rate_bpm"] != float64(110) || first["average_power_watts"] != float64(220) || first["average_cadence_rpm"] != float64(82) || first["elevation_gain_m"] != float64(5) {
		t.Fatalf("first metrics = %#v, want HR 110 W 220 cadence 82 gain 5", first)
	}
	second := rows[1].(map[string]any)
	if second["average_heart_rate_bpm"] != float64(130) || second["average_power_watts"] != float64(260) || second["average_cadence_rpm"] != float64(86) || second["elevation_gain_m"] != float64(8) {
		t.Fatalf("second metrics = %#v, want HR 130 W 260 cadence 86 gain 8", second)
	}
	if _, ok := first["full"]; ok {
		t.Fatalf("split row = %#v, want no raw full stream payload", first)
	}
	meta := payload["_meta"].(map[string]any)
	if meta["interval_source"] != "unknown" || !strings.Contains(meta["algorithm"].(string), "fixed-distance") {
		t.Fatalf("meta = %#v, want explicit source and algorithm", meta)
	}
	if meta["units"].(map[string]any)["heart_rate"] != "bpm" || meta["units"].(map[string]any)["elevation"] != "m" {
		t.Fatalf("units = %#v, want enriched metric units", meta["units"])
	}
}

func TestGetActivitySplitsPoolSwimDefaultsToValidated100mAndPreservesUnits(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{
			PreferredUnits: "imperial",
			SportSettings:  []intervals.SportSettings{{Types: []string{"sWiM"}, PaceLoadType: "swim", PaceUnits: "secs_100m"}},
		}},
		activity: decodeActivityFixture(t, `{"id":"swim-1","type":"sWiM"}`),
		streams: decodeStreamFixtures(t,
			`{"type":"distance","unit":"m","data":[0,50,150,250]}`,
			`{"type":"time","data":[0,30,90,150]}`,
		),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"swim-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	if payload["split_unit"] != "100m" || len(payload["splits"].([]any)) != 2 {
		t.Fatalf("payload = %#v, want two 100m rows", payload)
	}
	row := payload["splits"].([]any)[0].(map[string]any)
	if row["distance_100m"] != float64(1) || row["pace_seconds_per_100m"] != float64(60) || row["pace_seconds"] != float64(60) {
		t.Fatalf("pool row = %#v, want explicit 100m fields", row)
	}
	units := payload["_meta"].(map[string]any)["units"].(map[string]any)
	if units["distance"] != "100m" || units["pace"] != "sec/100m" || units["elevation"] != "m" {
		t.Fatalf("pool units = %#v, want validated swim units after shaping", units)
	}
}

func TestGetActivitySplitsRejectsUnprovenExplicit100m(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"ride-1","type":"Ride"}`),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"ride-1","split_unit":"100m"}`)})
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "100m") {
		t.Fatalf("PublicErrorMessage(%v) = %q, %v; want explicit 100m validation error", err, message, ok)
	}
	if client.streamCalls != 0 {
		t.Fatalf("stream calls = %d, want no stream fetch for non-swim explicit 100m", client.streamCalls)
	}
}

func TestGetActivitySplitsKeepsDeviceLapAsUpstreamInterval(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"ride-1","type":"Ride"}`),
		intervals: decodeIntervalsFixture(t, `{"id":"ride-1","icu_intervals":[
			{"id":"lap-1","name":"Lap 1","distance":1000,"duration":100,"lap_type":"auto"},
			{"id":"lap-2","name":"Lap 2","distance":1000,"duration":101,"lap_type":"auto"}
		]}`),
		streamErr: errors.New("streams unavailable"),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"ride-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	rows := payload["splits"].([]any)
	if len(rows) != 2 {
		t.Fatalf("splits = %#v, want both valid upstream rows", rows)
	}
	row := rows[0].(map[string]any)
	if row["provenance"] != "device_lap" || row["distance_basis"] != "upstream_interval_distance" {
		t.Fatalf("device row = %#v, want non-fixed provenance", row)
	}
	diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
	if !hasSplitDiagnostic(diagnostics, "device_lap_not_fixed_distance") || !hasSplitDiagnostic(diagnostics, "base_stream_unavailable") {
		t.Fatalf("diagnostics = %#v, want device and base-stream caveats", diagnostics)
	}
}

func TestGetActivitySplitsPauseAndMissingMetricsRemainSourceHonest(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
		streams: decodeStreamFixtures(t,
			`{"type":"distance","data":[0,1000,1000,2000]}`,
			`{"type":"time","data":[0,100,160,260]}`,
			`{"type":"heart_rate","data":[120,130,131,140]}`,
		),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	row := payload["splits"].([]any)[0].(map[string]any)
	if row["duration_seconds"] != float64(160) || row["average_heart_rate_bpm"] != float64(127.1) {
		t.Fatalf("pause row = %#v, want elapsed pause and time-weighted HR", row)
	}
	for _, field := range []string{"average_power_watts", "average_cadence_rpm", "elevation_gain_m"} {
		if _, ok := row[field]; ok {
			t.Fatalf("row = %#v, want missing %s omitted", row, field)
		}
	}
	meta := payload["_meta"].(map[string]any)
	diagnostics := meta["data_availability"].([]any)
	for _, reason := range []string{"paused_samples_present", "missing_metric_stream"} {
		if !hasSplitDiagnostic(diagnostics, reason) {
			t.Fatalf("diagnostics = %#v, want %s", diagnostics, reason)
		}
	}
}

func hasSplitDiagnostic(diagnostics []any, reason string) bool {
	for _, value := range diagnostics {
		if value.(map[string]any)["reason"] == reason {
			return true
		}
	}
	return false
}
