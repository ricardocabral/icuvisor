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
	units := meta["units"].(map[string]any)
	if units["system"] != "metric" || units["distance"] != "km" || units["pace"] != "min/km" {
		t.Fatalf("normal units = %#v, want unchanged metric distance/pace metadata", units)
	}
}

func TestGetActivitySplitsManualBoundaryFallbackAndPrecedence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interval   string
		streams    []string
		wantDur    float64
		wantMetric float64
	}{
		{
			name:     "distance bounds take precedence",
			interval: `{"id":"distance-1","start_distance":0,"end_distance":1000,"start_time":50,"end_time":60,"distance":1000,"duration":100}`,
			streams:  []string{`{"type":"distance","data":[0,500,1000]}`, `{"type":"time","data":[0,50,100]}`, `{"type":"cadence","data":[80,90,100]}`},
			wantDur:  100, wantMetric: 90,
		},
		{
			name:     "time bounds when distance is unavailable",
			interval: `{"id":"time-1","start_distance":0,"end_distance":1000,"start_time":10,"end_time":20,"distance":1000,"duration":10}`,
			streams:  []string{`{"type":"time","data":[0,10,20,30]}`, `{"type":"cadence","data":[80,90,100,110]}`},
			wantDur:  10, wantMetric: 95,
		},
		{
			name:     "inclusive index bounds when distance is unavailable",
			interval: `{"id":"index-1","start_index":1,"end_index":2,"distance":1000,"duration":10}`,
			streams:  []string{`{"type":"time","data":[0,10,20,30]}`, `{"type":"cadence","data":[80,90,100,110]}`},
			wantDur:  10, wantMetric: 95,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			intervalsRaw := `{"id":"run-1","icu_intervals":[` + tc.interval + `]}`
			client := &fakeActivityReadClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
				activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
				intervals:         decodeIntervalsFixture(t, intervalsRaw),
				streams:           decodeStreamFixtures(t, tc.streams...),
			}
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			row := resultMap(t, result)["splits"].([]any)[0].(map[string]any)
			if row["duration_seconds"] != tc.wantDur || row["average_cadence_rpm"] != tc.wantMetric {
				t.Fatalf("row = %#v, want duration %v and cadence %v", row, tc.wantDur, tc.wantMetric)
			}
		})
	}
}

func TestGetActivitySplitsFetchesBaseAndOptionalChannelsSeparately(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
		streams: decodeStreamFixtures(t,
			`{"type":"Distance","data":[0,1000]}`,
			`{"type":"Time","data":[0,100]}`,
			`{"type":"HR","data":[120,130]}`,
		),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	if _, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)}); err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.streamParamHistory) != 2 {
		t.Fatalf("stream requests = %#v, want base and optional requests", client.streamParamHistory)
	}
	if got := strings.Join(client.streamParamHistory[0].Types, ","); got != "distance,time" || !client.streamParamHistory[0].IncludeDefaults {
		t.Fatalf("base request = %+v, want distance,time with defaults", client.streamParamHistory[0])
	}
	if got := strings.Join(client.streamParamHistory[1].Types, ","); got != "heart_rate,watts,cadence,altitude" || !client.streamParamHistory[1].IncludeDefaults {
		t.Fatalf("optional request = %+v, want metric channels with defaults", client.streamParamHistory[1])
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
	if units["system"] != "metric" || units["distance"] != "100m" || units["pace"] != "sec/100m" || units["elevation"] != "m" || units["speed"] != nil {
		t.Fatalf("pool units = %#v, want validated metric swim units after shaping", units)
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

func TestGetActivitySplitsPoolManualRowsInclude100mPace(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{
			PreferredUnits: "imperial",
			SportSettings:  []intervals.SportSettings{{Types: []string{"Swim"}, PaceLoadType: "SWIM", PaceUnits: "SECS_100M"}},
		}},
		activity:  decodeActivityFixture(t, `{"id":"swim-1","type":"Swim"}`),
		intervals: decodeIntervalsFixture(t, `{"id":"swim-1","icu_intervals":[{"id":"lap-1","distance":200,"duration":120}]}`),
		streams:   decodeStreamFixtures(t, `{"type":"distance","unit":"m","data":[0,200]}`, `{"type":"time","data":[0,120]}`),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"swim-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["splits"].([]any)[0].(map[string]any)
	if row["distance_100m"] != float64(2) || row["pace_seconds_per_100m"] != float64(60) || row["pace_seconds"] != float64(120) {
		t.Fatalf("pool manual row = %#v, want 100m pace plus legacy duration", row)
	}
	if row["distance_basis"] != "upstream_interval_distance" {
		t.Fatalf("pool manual row = %#v, want upstream distance basis", row)
	}
}

func TestGetActivitySplitsRejects100mForYardOrUnlabelledDistance(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{
			PreferredUnits: "metric",
			SportSettings:  []intervals.SportSettings{{Types: []string{"Swim"}, PaceLoadType: "SWIM", PaceUnits: "SECS_100M"}},
		}},
		activity: decodeActivityFixture(t, `{"id":"swim-1","type":"Swim"}`),
		streams:  decodeStreamFixtures(t, `{"type":"distance","unit":"yd","data":[0,100]}`, `{"type":"time","data":[0,60]}`),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"swim-1","split_unit":"100m"}`)})
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "100m") {
		t.Fatalf("PublicErrorMessage(%v) = %q, %v; want explicit 100m rejection", err, message, ok)
	}
	if client.streamCalls != 1 {
		t.Fatalf("stream calls = %d, want one base fetch before distance-unit proof", client.streamCalls)
	}
}

func TestGetActivitySplitsSwimSemanticsFallbackIsExplicit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		profile  intervals.AthleteWithSportSettings
		distance string
		wantDiag string
	}{
		{
			name: "yards pace setting",
			profile: intervals.AthleteWithSportSettings{
				PreferredUnits: "metric",
				SportSettings:  []intervals.SportSettings{{Types: []string{"Swim"}, PaceLoadType: "SWIM", PaceUnits: "SECS_100Y"}},
			},
			distance: `{"type":"distance","unit":"m","data":[0,1000]}`,
			wantDiag: "unsupported_swim_pace_units",
		},
		{
			name:     "ambiguous sport settings",
			profile:  intervals.AthleteWithSportSettings{PreferredUnits: "metric"},
			distance: `{"type":"distance","units":"m","data":[0,1000]}`,
			wantDiag: "swim_semantics_unavailable",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{
				fakeProfileClient: fakeProfileClient{profile: tc.profile},
				activity:          decodeActivityFixture(t, `{"id":"swim-1","type":"Swim"}`),
				streams: decodeStreamFixtures(t,
					tc.distance,
					`{"type":"time","data":[0,600]}`,
				),
			}
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"swim-1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			if payload["split_unit"] != "km" {
				t.Fatalf("split_unit = %#v, want safe km fallback", payload["split_unit"])
			}
			if !hasSplitDiagnostic(payload["_meta"].(map[string]any)["data_availability"].([]any), tc.wantDiag) {
				t.Fatalf("meta = %#v, want %s", payload["_meta"], tc.wantDiag)
			}
		})
	}
}

func TestGetActivitySplitsPreservesEveryIntervalSourceProvenance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		interval   string
		wantSource string
		wantRow    string
		wantDiag   string
	}{
		{name: "structured workout", interval: `{"id":"structured-1","name":"Tempo","distance":1000,"duration":100}`, wantSource: "structured_workout", wantRow: "structured_workout_interval"},
		{name: "manual added", interval: `{"id":"manual-1","name":"Block","start_index":1,"end_index":2,"distance":1000,"duration":100}`, wantSource: "manual_added", wantRow: "manual_interval"},
		{name: "unknown", interval: `{"id":"unknown-1","name":"Segment","distance":1000,"duration":100}`, wantSource: "unknown", wantRow: "unknown_interval"},
		{name: "device lap", interval: `{"id":"device-1","name":"Lap 1","distance":1000,"duration":100,"lap_type":"auto"}`, wantSource: "device_laps", wantRow: "device_lap", wantDiag: "device_lap_not_fixed_distance"},
		{name: "mixed", interval: `{"id":"mixed-1","name":"Mixed","group_id":"group-1","distance":1000,"duration":100}`, wantSource: "mixed", wantRow: "mixed_interval", wantDiag: "mixed_interval_source"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			intervalJSON := tc.interval
			if tc.wantSource == "mixed" {
				intervalJSON = `{"id":"mixed-1","name":"Mixed","group_id":"group-1","distance":1000,"duration":100},` + `{"id":"mixed-2","name":"Added","start_index":1,"end_index":2,"distance":1000,"duration":100}`
			}
			client := &fakeActivityReadClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
				activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
				intervals:         decodeIntervalsFixture(t, `{"id":"run-1","icu_intervals":[`+intervalJSON+`]}`),
				streams:           decodeStreamFixtures(t, `{"type":"distance","data":[0,1000]}`, `{"type":"time","data":[0,100]}`),
			}
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			meta := payload["_meta"].(map[string]any)
			if meta["interval_source"] != tc.wantSource {
				t.Fatalf("interval_source = %#v, want %q", meta["interval_source"], tc.wantSource)
			}
			rows := payload["splits"].([]any)
			if len(rows) == 0 || rows[0].(map[string]any)["provenance"] != tc.wantRow || rows[0].(map[string]any)["distance_basis"] != "upstream_interval_distance" {
				t.Fatalf("rows = %#v, want source-honest upstream row %q", rows, tc.wantRow)
			}
			if tc.wantDiag != "" && !hasSplitDiagnostic(meta["data_availability"].([]any), tc.wantDiag) {
				t.Fatalf("diagnostics = %#v, want %s", meta["data_availability"], tc.wantDiag)
			}
		})
	}
}

func TestGetActivitySplitsFailureMappingKeepsSourceHonest(t *testing.T) {
	t.Parallel()

	t.Run("interval failure uses virtual rows", func(t *testing.T) {
		client := &fakeActivityReadClient{
			fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
			activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
			intervalErr:       errors.New("interval endpoint unavailable"),
			streams:           decodeStreamFixtures(t, `{"type":"distance","data":[0,1000]}`, `{"type":"time","data":[0,100]}`),
		}
		tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		payload := resultMap(t, result)
		if payload["source"] != "virtual_streams" || len(payload["splits"].([]any)) != 1 {
			t.Fatalf("payload = %#v, want one virtual row", payload)
		}
		if !hasSplitDiagnostic(payload["_meta"].(map[string]any)["data_availability"].([]any), "interval_source_unavailable") {
			t.Fatalf("meta = %#v, want interval_source_unavailable", payload["_meta"])
		}
	})

	t.Run("empty intervals preserve structured unavailable fallback", func(t *testing.T) {
		client := &fakeActivityReadClient{
			fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
			activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
			intervals:         decodeIntervalsFixture(t, `{"id":"run-1","icu_intervals":[{"id":"empty","distance":0,"duration":0}]}`),
			streamErr:         intervals.ErrNotFound,
		}
		tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		payload := resultMap(t, result)
		if payload["unavailable"] == nil {
			t.Fatalf("payload = %#v, want structured unavailable response", payload)
		}
		meta := payload["_meta"].(map[string]any)
		if !hasSplitDiagnostic(meta["data_availability"].([]any), "base_stream_unavailable") {
			t.Fatalf("meta = %#v, want base_stream_unavailable", meta)
		}
		if !strings.Contains(meta["algorithm"].(string), "elapsed-time interpolation") || strings.Contains(meta["algorithm"].(string), "ignoring paused") {
			t.Fatalf("algorithm = %#v, want current deterministic split contract", meta["algorithm"])
		}
	})

	t.Run("omitted swim base failure reports semantics", func(t *testing.T) {
		client := &fakeActivityReadClient{
			fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{
				PreferredUnits: "metric",
				SportSettings:  []intervals.SportSettings{{Types: []string{"Swim"}, PaceLoadType: "SWIM", PaceUnits: "SECS_100M"}},
			}},
			activity:  decodeActivityFixture(t, `{"id":"swim-1","type":"Swim"}`),
			intervals: decodeIntervalsFixture(t, `{"id":"swim-1","icu_intervals":[{"id":"manual-1","distance":1000,"duration":100}]}`),
			streamErr: errors.New("base stream unavailable"),
		}
		tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"swim-1"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		meta := resultMap(t, result)["_meta"].(map[string]any)
		if !hasSplitDiagnostic(meta["data_availability"].([]any), "swim_semantics_unavailable") {
			t.Fatalf("meta = %#v, want swim_semantics_unavailable", meta)
		}
	})

	t.Run("base failure retains scalar upstream metrics", func(t *testing.T) {
		client := &fakeActivityReadClient{
			fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
			activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
			intervals:         decodeIntervalsFixture(t, `{"id":"run-1","icu_intervals":[{"id":"manual-1","distance":1000,"duration":100,"average_hr":141.25,"average_power":250.55}]}`),
			streamErr:         errors.New("base stream unavailable"),
		}
		tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		row := resultMap(t, result)["splits"].([]any)[0].(map[string]any)
		if row["average_heart_rate_bpm"] != float64(141.3) || row["average_power_watts"] != float64(250.6) {
			t.Fatalf("row = %#v, want rounded interval scalar metrics", row)
		}
		if _, ok := row["average_cadence_rpm"]; ok {
			t.Fatalf("row = %#v, want no stream-derived cadence", row)
		}
	})
}

func TestGetActivitySplitsReportsMetricCoveragePerManualRow(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
		intervals:         decodeIntervalsFixture(t, `{"id":"run-1","icu_intervals":[{"id":"outside","start_time":100,"end_time":200,"distance":1000,"duration":100}]}`),
		streams: decodeStreamFixtures(t,
			`{"type":"distance","data":[0,1000]}`,
			`{"type":"time","data":[0,100]}`,
			`{"type":"cadence","data":[80,90]}`,
		),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	diagnostics := resultMap(t, result)["_meta"].(map[string]any)["data_availability"].([]any)
	for _, value := range diagnostics {
		if diagnostic := value.(map[string]any); diagnostic["reason"] == "metric_insufficient_coverage" {
			missing := diagnostic["missing_fields"].([]any)
			if len(missing) != 1 || missing[0] != "splits[0].average_cadence_rpm" {
				t.Fatalf("diagnostic = %#v, want row-specific missing field", diagnostic)
			}
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want metric_insufficient_coverage", diagnostics)
}

func TestGetActivitySplitsRejectsInvalidAndPartialBaseCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		distance  string
		time      string
		want      string
		wantCount int
	}{
		{name: "decreasing distance", distance: `{"type":"distance","data":[0,1200,1000]}`, time: `{"type":"time","data":[0,60,120]}`, want: "non_monotonic_distance"},
		{name: "decreasing time", distance: `{"type":"distance","data":[0,1000,2000]}`, time: `{"type":"time","data":[0,60,50]}`, want: "non_monotonic_time"},
		{name: "non-zero origin", distance: `{"type":"distance","data":[100,1100]}`, time: `{"type":"time","data":[0,100]}`, want: "insufficient_split_coverage"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeActivityReadClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
				activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
				streams:           decodeStreamFixtures(t, tc.distance, tc.time),
			}
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			if got := len(payload["splits"].([]any)); got != tc.wantCount {
				t.Fatalf("split count = %d, want %d; payload = %#v", got, tc.wantCount, payload)
			}
			if !hasSplitDiagnostic(payload["_meta"].(map[string]any)["data_availability"].([]any), tc.want) {
				t.Fatalf("meta = %#v, want %s", payload["_meta"], tc.want)
			}
		})
	}
}

func TestGetActivitySplitsReportsManualBoundaryID(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
		intervals:         decodeIntervalsFixture(t, `{"id":"run-1","icu_intervals":[{"id":"no-boundary","distance":1000,"duration":100}]}`),
		streams:           decodeStreamFixtures(t, `{"type":"distance","data":[0,1000]}`, `{"type":"time","data":[0,100]}`),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	diagnostics := resultMap(t, result)["_meta"].(map[string]any)["data_availability"].([]any)
	for _, value := range diagnostics {
		if diagnostic := value.(map[string]any); diagnostic["reason"] == "manual_boundary_unavailable" {
			if diagnostic["interval_id"] != "no-boundary" {
				t.Fatalf("diagnostic = %#v, want interval_id no-boundary", diagnostic)
			}
			return
		}
	}
	t.Fatalf("diagnostics = %#v, want manual_boundary_unavailable", diagnostics)
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

func TestGetActivitySplitsMetricAggregationIsRightContinuousAtDuplicateTimes(t *testing.T) {
	t.Parallel()

	client := &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{PreferredUnits: "metric"}},
		activity:          decodeActivityFixture(t, `{"id":"run-1","type":"Run"}`),
		streams: decodeStreamFixtures(t,
			`{"type":"distance","data":[0,500,750,1000,1500,2000]}`,
			`{"type":"time","data":[0,10,10,20,30,40]}`,
			`{"type":"heart_rate","data":[100,110,130,140,150,160]}`,
		),
	}
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"run-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["splits"].([]any)[0].(map[string]any)
	if row["average_heart_rate_bpm"] != float64(125) {
		t.Fatalf("row = %#v, want right-continuous duplicate-time average 125", row)
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

func TestValidateMetricStreamsUsesStableValidityDiagnostics(t *testing.T) {
	t.Parallel()

	metrics := splitMetricStreams{Rows: map[string]intervals.ActivityStream{
		"heart_rate": {Data: []float64{120, 0}, Raw: map[string]any{"data": []any{120.0, nil}}},
		"watts":      {Data: []float64{200}},
		"altitude":   {Data: []float64{10, 11}},
	}}
	diagnostics := validateMetricStreams("run-1", metrics, 2)
	if len(diagnostics) != 3 {
		t.Fatalf("diagnostics = %#v, want one per affected optional channel", diagnostics)
	}
	wantReasons := []string{"metric_channel_invalid", "metric_channel_length_mismatch", "missing_metric_stream"}
	for index, want := range wantReasons {
		if diagnostics[index].Reason != want {
			t.Fatalf("diagnostic[%d] = %#v, want reason %q", index, diagnostics[index], want)
		}
	}
	if diagnostics[0].Available == nil || strings.Join(diagnostics[0].Available, ",") != "heart_rate,watts,altitude" {
		t.Fatalf("available = %#v, want stable metric order", diagnostics[0].Available)
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
