package tools

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type splitFixtureFile struct {
	Activity  json.RawMessage                    `json:"activity"`
	Profile   intervals.AthleteWithSportSettings `json:"profile"`
	Intervals json.RawMessage                    `json:"intervals"`
	Streams   []json.RawMessage                  `json:"streams"`
}

func loadSplitFixture(t *testing.T, name string) *fakeActivityReadClient {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "activity_splits", name+".json"))
	if err != nil {
		t.Fatalf("read split fixture %q: %v", name, err)
	}
	var fixture splitFixtureFile
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode split fixture %q: %v", name, err)
	}
	var activity intervals.Activity
	if err := json.Unmarshal(fixture.Activity, &activity); err != nil {
		t.Fatalf("decode fixture activity %q: %v", name, err)
	}
	var intervalRows intervals.IntervalsDTO
	if err := json.Unmarshal(fixture.Intervals, &intervalRows); err != nil {
		t.Fatalf("decode fixture intervals %q: %v", name, err)
	}
	streams := make([]intervals.ActivityStream, 0, len(fixture.Streams))
	for _, raw := range fixture.Streams {
		streams = append(streams, streamFixtureValue(t, raw))
	}
	return &fakeActivityReadClient{
		fakeProfileClient: fakeProfileClient{profile: fixture.Profile},
		activity:          activity,
		intervals:         intervalRows,
		streams:           streams,
	}
}

func streamFixtureValue(t *testing.T, raw []byte) intervals.ActivityStream {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("decode fixture stream: %v", err)
	}
	var stream intervals.ActivityStream
	if err := json.Unmarshal(raw, &stream); err != nil {
		withoutData := make(map[string]json.RawMessage, len(fields))
		for key, value := range fields {
			if key != "data" {
				withoutData[key] = value
			}
		}
		metadata, marshalErr := json.Marshal(withoutData)
		if marshalErr != nil || json.Unmarshal(metadata, &stream) != nil {
			t.Fatalf("decode fixture stream metadata: %v", err)
		}
	}
	var encodedData []json.RawMessage
	if value, ok := fields["data"]; ok {
		if err := json.Unmarshal(value, &encodedData); err != nil {
			t.Fatalf("decode fixture stream data: %v", err)
		}
	}
	rawData := make([]any, 0, len(encodedData))
	stream.Data = make([]float64, 0, len(encodedData))
	for _, encoded := range encodedData {
		trimmed := strings.TrimSpace(string(encoded))
		switch trimmed {
		case "null":
			rawData = append(rawData, nil)
			stream.Data = append(stream.Data, 0)
		case `"NaN"`:
			rawData = append(rawData, "NaN")
			stream.Data = append(stream.Data, math.NaN())
		case `"Inf"`, `"Infinity"`:
			rawData = append(rawData, "Inf")
			stream.Data = append(stream.Data, math.Inf(1))
		case `"-Inf"`, `"-Infinity"`:
			rawData = append(rawData, "-Inf")
			stream.Data = append(stream.Data, math.Inf(-1))
		default:
			var number float64
			if err := json.Unmarshal(encoded, &number); err != nil {
				t.Fatalf("decode fixture stream sample %s: %v", encoded, err)
			}
			rawData = append(rawData, number)
			stream.Data = append(stream.Data, number)
		}
	}
	stream.Raw = make(map[string]any, len(fields))
	for key, value := range fields {
		var decoded any
		if err := json.Unmarshal(value, &decoded); err != nil {
			t.Fatalf("decode fixture stream field %q: %v", key, err)
		}
		stream.Raw[key] = decoded
	}
	if _, ok := fields["data"]; ok {
		stream.Raw["data"] = rawData
	}
	return stream
}

func TestGetActivitySplitsFixtureRunRowsAndBothResponseModes(t *testing.T) {
	t.Parallel()

	for _, includeFull := range []bool{false, true} {
		t.Run(map[bool]string{false: "terse", true: "full"}[includeFull], func(t *testing.T) {
			client := loadSplitFixture(t, "run_all_metrics_aliases")
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
			args := `{"activity_id":"fixture-run-all"}`
			if includeFull {
				args = `{"activity_id":"fixture-run-all","include_full":true}`
			}
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			if payload["source"] != "virtual_streams" {
				t.Fatalf("source = %#v, want virtual_streams", payload["source"])
			}
			rows := payload["splits"].([]any)
			if len(rows) != 2 {
				t.Fatalf("rows = %#v, want two km rows", rows)
			}
			first := rows[0].(map[string]any)
			for _, field := range []string{"distance_km", "duration_seconds", "pace_seconds", "provenance", "distance_basis", "average_heart_rate_bpm", "average_power_watts", "average_cadence_rpm", "elevation_gain_m"} {
				if _, ok := first[field]; !ok {
					t.Fatalf("first row = %#v, want legacy/enriched field %s", first, field)
				}
			}
			if first["elevation_gain_m"] != float64(5) || rows[1].(map[string]any)["elevation_gain_m"] != float64(8) {
				t.Fatalf("rows = %#v, want positive ascent with descent excluded", rows)
			}
			units := payload["_meta"].(map[string]any)["units"].(map[string]any)
			if units["system"] != "metric" || units["distance"] != "km" || units["pace"] != "min/km" {
				t.Fatalf("units = %#v, want metric km metadata", units)
			}
			assertSplitPayloadHasNoRaw(t, payload)
		})
	}
}

func TestGetActivitySplitsFixtureMatrixOmitsInvalidMetricsWithStableDiagnostics(t *testing.T) {
	t.Parallel()

	client := loadSplitFixture(t, "metric_channel_matrix")
	tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-matrix"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	for _, rowValue := range payload["splits"].([]any) {
		row := rowValue.(map[string]any)
		for _, field := range []string{"average_heart_rate_bpm", "average_power_watts", "average_cadence_rpm"} {
			if _, ok := row[field]; ok {
				t.Fatalf("row = %#v, want unavailable %s omitted", row, field)
			}
		}
		if _, ok := row["elevation_gain_m"]; !ok {
			t.Fatalf("row = %#v, want valid altitude enrichment", row)
		}
	}
	diagnostics := payload["_meta"].(map[string]any)["data_availability"].([]any)
	want := map[string]int{"metric_channel_invalid": 2, "metric_channel_length_mismatch": 1}
	for reason, count := range want {
		if got := splitDiagnosticCount(diagnostics, reason); got != count {
			t.Fatalf("%s count = %d, want %d; diagnostics = %#v", reason, got, count, diagnostics)
		}
	}
}

func TestGetActivitySplitsFixtureBaseFailuresPausesAndIntervalRows(t *testing.T) {
	t.Parallel()

	paused := loadSplitFixture(t, "paused_and_invalid_base")
	pausedTool := newGetActivitySplitsTool(paused, paused, paused, paused, "test", false)
	pausedResult, err := pausedTool.Handler(context.Background(), Request{Name: pausedTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-paused"}`)})
	if err != nil {
		t.Fatalf("paused Handler() error = %v", err)
	}
	pausedPayload := resultMap(t, pausedResult)
	if pausedPayload["splits"].([]any)[0].(map[string]any)["duration_seconds"] != float64(160) || !hasSplitDiagnostic(pausedPayload["_meta"].(map[string]any)["data_availability"].([]any), "paused_samples_present") {
		t.Fatalf("paused payload = %#v, want elapsed pause and diagnostic", pausedPayload)
	}

	for _, tc := range []struct {
		fixture string
		reason  string
	}{
		{fixture: "decreasing_distance", reason: "non_monotonic_distance"},
		{fixture: "decreasing_time", reason: "non_monotonic_time"},
		{fixture: "non_zero_origin", reason: "insufficient_split_coverage"},
	} {
		client := loadSplitFixture(t, tc.fixture)
		tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"` + client.activity.ID + `"}`)})
		if err != nil {
			t.Fatalf("%s Handler() error = %v", tc.fixture, err)
		}
		payload := resultMap(t, result)
		if len(payload["splits"].([]any)) != 0 || !hasSplitDiagnostic(payload["_meta"].(map[string]any)["data_availability"].([]any), tc.reason) {
			t.Fatalf("%s payload = %#v, want no rows and %s", tc.fixture, payload, tc.reason)
		}
	}

	partial := loadSplitFixture(t, "metric_partial_manual")
	partialTool := newGetActivitySplitsTool(partial, partial, partial, partial, "test", false)
	partialResult, err := partialTool.Handler(context.Background(), Request{Name: partialTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-partial-manual"}`)})
	if err != nil {
		t.Fatalf("partial Handler() error = %v", err)
	}
	partialPayload := resultMap(t, partialResult)
	if !hasSplitDiagnostic(partialPayload["_meta"].(map[string]any)["data_availability"].([]any), "metric_insufficient_coverage") {
		t.Fatalf("partial payload = %#v, want metric coverage diagnostic", partialPayload)
	}

	intervalsClient := loadSplitFixture(t, "interval_sources")
	intervalsTool := newGetActivitySplitsTool(intervalsClient, intervalsClient, intervalsClient, intervalsClient, "test", false)
	intervalsResult, err := intervalsTool.Handler(context.Background(), Request{Name: intervalsTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-intervals"}`)})
	if err != nil {
		t.Fatalf("interval source Handler() error = %v", err)
	}
	intervalsPayload := resultMap(t, intervalsResult)
	intervalRows := intervalsPayload["splits"].([]any)
	if intervalsPayload["source"] != "manual_intervals" || len(intervalRows) != 3 || intervalRows[0].(map[string]any)["distance_basis"] != "upstream_interval_distance" || intervalRows[0].(map[string]any)["pace_seconds"] != float64(100) {
		t.Fatalf("interval fixture payload = %#v, want legacy upstream rows", intervalsPayload)
	}

	baseFailure := loadSplitFixture(t, "ride_distance_time_only")
	baseFailure.streamErrors = map[string]error{"distance,time": intervals.ErrNotFound}
	baseTool := newGetActivitySplitsTool(baseFailure, baseFailure, baseFailure, baseFailure, "test", false)
	baseResult, err := baseTool.Handler(context.Background(), Request{Name: baseTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-ride-only"}`)})
	if err != nil {
		t.Fatalf("base failure Handler() error = %v", err)
	}
	basePayload := resultMap(t, baseResult)
	if basePayload["unavailable"] == nil || !hasSplitDiagnostic(basePayload["_meta"].(map[string]any)["data_availability"].([]any), "base_stream_unavailable") {
		t.Fatalf("base failure payload = %#v, want structured unavailable and diagnostic", basePayload)
	}

	optionalFailure := loadSplitFixture(t, "run_all_metrics_aliases")
	optionalFailure.streamErrors = map[string]error{"heart_rate,watts,cadence,altitude": errors.New("optional streams unavailable")}
	optionalTool := newGetActivitySplitsTool(optionalFailure, optionalFailure, optionalFailure, optionalFailure, "test", false)
	optionalResult, err := optionalTool.Handler(context.Background(), Request{Name: optionalTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-run-all"}`)})
	if err != nil {
		t.Fatalf("optional failure Handler() error = %v", err)
	}
	optionalPayload := resultMap(t, optionalResult)
	if !hasSplitDiagnostic(optionalPayload["_meta"].(map[string]any)["data_availability"].([]any), "metric_stream_unavailable") {
		t.Fatalf("optional failure payload = %#v, want metric_stream_unavailable", optionalPayload)
	}
	assertSplitPayloadHasNoRaw(t, optionalPayload)
}

func TestGetActivitySplitsFixtureUnitsAndPoolPrecedence(t *testing.T) {
	t.Parallel()

	ride := loadSplitFixture(t, "ride_distance_time_only")
	rideTool := newGetActivitySplitsTool(ride, ride, ride, ride, "test", false)
	rideResult, err := rideTool.Handler(context.Background(), Request{Name: rideTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-ride-only"}`)})
	if err != nil {
		t.Fatalf("ride Handler() error = %v", err)
	}
	ridePayload := resultMap(t, rideResult)
	if ridePayload["split_unit"] != "mi" || ridePayload["splits"].([]any)[0].(map[string]any)["distance_mi"] != float64(1) {
		t.Fatalf("ride payload = %#v, want imperial mile row", ridePayload)
	}
	rideUnits := ridePayload["_meta"].(map[string]any)["units"].(map[string]any)
	if rideUnits["system"] != "imperial" || rideUnits["distance"] != "mi" || rideUnits["pace"] != "min/mi" {
		t.Fatalf("ride units = %#v, want imperial mile metadata", rideUnits)
	}

	pool := loadSplitFixture(t, "pool_virtual_100m")
	poolTool := newGetActivitySplitsTool(pool, pool, pool, pool, "test", false)
	poolResult, err := poolTool.Handler(context.Background(), Request{Name: poolTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-pool-virtual"}`)})
	if err != nil {
		t.Fatalf("pool Handler() error = %v", err)
	}
	poolPayload := resultMap(t, poolResult)
	poolRow := poolPayload["splits"].([]any)[0].(map[string]any)
	if poolPayload["source"] != "virtual_streams" || poolRow["distance_100m"] != float64(1) || poolRow["pace_seconds_per_100m"] != float64(60) || poolRow["provenance"] != "virtual_fixed_distance" || poolRow["distance_basis"] != "fixed_distance_boundary" {
		t.Fatalf("pool payload = %#v, want validated virtual 100m row", poolPayload)
	}
	units := poolPayload["_meta"].(map[string]any)["units"].(map[string]any)
	if units["system"] != "metric" || units["distance"] != "100m" || units["pace"] != "sec/100m" {
		t.Fatalf("pool units = %#v, want metric 100m units", units)
	}

	laps := loadSplitFixture(t, "pool_manual_laps_100m")
	lapsTool := newGetActivitySplitsTool(laps, laps, laps, laps, "test", false)
	lapsResult, err := lapsTool.Handler(context.Background(), Request{Name: lapsTool.Name, Arguments: json.RawMessage(`{"activity_id":"fixture-pool-laps"}`)})
	if err != nil {
		t.Fatalf("laps Handler() error = %v", err)
	}
	lapsPayload := resultMap(t, lapsResult)
	lap := lapsPayload["splits"].([]any)[0].(map[string]any)
	if lapsPayload["source"] != "manual_intervals" || lap["pace_seconds"] != float64(60) || lap["pace_seconds_per_100m"] != float64(60) || lap["distance_basis"] != "upstream_interval_distance" || lap["provenance"] != "device_lap" {
		t.Fatalf("laps payload = %#v, want manual-over-virtual upstream row", lapsPayload)
	}
}

func TestGetActivitySplitsFixturePoolValidationModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fixture     string
		args        string
		wantDiag    string
		wantUserErr bool
	}{
		{name: "yards omitted fallback", fixture: "pool_yards_100m", args: `{"activity_id":"fixture-pool-yards"}`, wantDiag: "swim_semantics_unavailable"},
		{name: "unknown pace omitted fallback", fixture: "pool_unknown_pace", args: `{"activity_id":"fixture-pool-unknown"}`, wantDiag: "unsupported_swim_pace_units"},
		{name: "missing unit omitted fallback", fixture: "pool_missing_unit", args: `{"activity_id":"fixture-pool-missing"}`, wantDiag: "swim_semantics_unavailable"},
		{name: "ambiguous unit omitted fallback", fixture: "pool_ambiguous_unit", args: `{"activity_id":"fixture-pool-ambiguous"}`, wantDiag: "swim_semantics_unavailable"},
		{name: "yards explicit error", fixture: "pool_yards_100m", args: `{"activity_id":"fixture-pool-yards","split_unit":"100m"}`, wantUserErr: true},
		{name: "unknown pace explicit error", fixture: "pool_unknown_pace", args: `{"activity_id":"fixture-pool-unknown","split_unit":"100m"}`, wantUserErr: true},
		{name: "missing unit explicit error", fixture: "pool_missing_unit", args: `{"activity_id":"fixture-pool-missing","split_unit":"100m"}`, wantUserErr: true},
		{name: "ambiguous unit explicit error", fixture: "pool_ambiguous_unit", args: `{"activity_id":"fixture-pool-ambiguous","split_unit":"100m"}`, wantUserErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := loadSplitFixture(t, tc.fixture)
			tool := newGetActivitySplitsTool(client, client, client, client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if tc.wantUserErr {
				if _, ok := PublicErrorMessage(err); !ok {
					t.Fatalf("error = %v, want public user error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			payload := resultMap(t, result)
			if !hasSplitDiagnostic(payload["_meta"].(map[string]any)["data_availability"].([]any), tc.wantDiag) {
				t.Fatalf("payload = %#v, want %s", payload, tc.wantDiag)
			}
			assertSplitPayloadHasNoRaw(t, payload)
		})
	}
}

func assertSplitPayloadHasNoRaw(t *testing.T, value any) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "samples" || key == "full" || key == "data" || key == "data2" {
				t.Fatalf("split payload contains raw key %q: %#v", key, typed)
			}
			assertSplitPayloadHasNoRaw(t, child)
		}
	case []any:
		for _, child := range typed {
			assertSplitPayloadHasNoRaw(t, child)
		}
	}
}

func TestGetActivitySplitsDocumentationMentionsSourceAndPoolCaveats(t *testing.T) {
	t.Parallel()

	changelog, err := os.ReadFile(filepath.Join("..", "..", "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("read changelog: %v", err)
	}
	cookbook, err := os.ReadFile(filepath.Join("..", "..", "web", "content", "cookbook", "activity-retrospective.md"))
	if err != nil {
		t.Fatalf("read cookbook: %v", err)
	}
	for name, content := range map[string]string{"CHANGELOG.md": string(changelog), "activity-retrospective.md": string(cookbook)} {
		for _, want := range []string{"get_activity_splits", "100m", "provenance", "distance_basis"} {
			if !strings.Contains(content, want) {
				t.Fatalf("%s missing split guidance %q", name, want)
			}
		}
	}
}

func splitDiagnosticCount(diagnostics []any, reason string) int {
	count := 0
	for _, value := range diagnostics {
		if value.(map[string]any)["reason"] == reason {
			count++
		}
	}
	return count
}
