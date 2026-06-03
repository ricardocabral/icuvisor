package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeExtendedMetricsClient struct {
	fakeProfileClient
	activity     intervals.Activity
	activityErr  error
	intervals    intervals.IntervalsDTO
	intervalsErr error
	powerVsHR    intervals.PowerVsHR
	powerErr     error
}

func (f *fakeExtendedMetricsClient) GetActivity(context.Context, string) (intervals.Activity, error) {
	return f.activity, f.activityErr
}

func (f *fakeExtendedMetricsClient) GetActivityIntervals(context.Context, string) (intervals.IntervalsDTO, error) {
	return f.intervals, f.intervalsErr
}

func (f *fakeExtendedMetricsClient) GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error) {
	return f.powerVsHR, f.powerErr
}

func TestExtendedMetricsUnavailableReasons(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		activity intervals.Activity
		err      error
		reason   string
	}{
		{name: "strava_blocked", activity: decodeExtendedMetricsActivity(t, `{"id":"stub1","source":"Strava","_note":"hidden"}`), reason: "strava_blocked"},
		{name: "not_found", err: intervals.ErrNotFound, reason: "not_found"},
		{name: "unauthorized", err: intervals.ErrUnauthorized, reason: "unauthorized"},
		{name: "rate_limited", err: intervals.ErrRateLimited, reason: "rate_limited"},
		{name: "upstream_unavailable_500", err: &intervals.Error{StatusCode: 500, Kind: intervals.ErrUpstream}, reason: "upstream_unavailable"},
		{name: "upstream_unavailable_400", err: &intervals.Error{StatusCode: 400, Kind: intervals.ErrUpstream}, reason: "upstream_unavailable"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeExtendedMetricsClient(t)
			client.activity = tc.activity
			client.activityErr = tc.err
			tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"stub1"}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v, want structured unavailable", err)
			}
			assertUnavailableReason(t, resultMap(t, result), tc.reason)
		})
	}
}

func TestExtendedMetricsStravaUnavailableIncludesFullWhenRequested(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"strava-1","source":"Strava","_note":"Imported from Strava","external_id":"wahoo-synthetic-98765"}`)
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"fallback","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	if got["activity_id"] != "strava-1" || got["strava_imported"] != true {
		t.Fatalf("strava response = %#v", got)
	}
	assertUnavailableReasonAndWorkaround(t, got, "strava_blocked", wantWahooStravaWorkaround)
	full := got["full"].(map[string]any)
	activity := full["activity"].(map[string]any)
	if activity["source"] != "Strava" {
		t.Fatalf("full activity = %#v", activity)
	}
	meta := got["_meta"].(map[string]any)
	if meta["include_full"] != true || meta["server_version"] != "test" {
		t.Fatalf("meta = %#v", meta)
	}
}

func TestExtendedMetricsOptionalSourcesReportPartialAvailability(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-1","name":"Run","average_stride":1.23,"icu_zone_times":[10,20],"pace_zone_times":[30,40]}`)
	client.intervalsErr = intervals.ErrNotFound
	client.powerErr = intervals.ErrUnauthorized
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	metrics := got["metrics"].(map[string]any)
	if metrics["stride_length_m"] != 1.23 {
		t.Fatalf("metrics = %#v", metrics)
	}
	if _, ok := got["intervals"]; ok {
		t.Fatalf("intervals present despite optional source miss: %#v", got["intervals"])
	}
	if _, ok := got["full"]; ok {
		t.Fatalf("full present in terse partial response: %#v", got["full"])
	}
	meta := got["_meta"].(map[string]any)
	if meta["partial"] != true {
		t.Fatalf("meta = %#v, want partial", meta)
	}
	unavailable := meta["unavailable_sources"].([]any)
	if strings.Join([]string{unavailable[0].(string), unavailable[1].(string)}, ",") != "intervals,power_vs_hr" {
		t.Fatalf("unavailable_sources = %#v", unavailable)
	}
}

func TestExtendedMetricsIncludeFullContainsAvailableSources(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-2","name":"Ride","icu_power_hr":1.5,"use_gap_zone_times":true,"gap_zone_times":[50,60],"pace_zone_times":[1,2]}`)
	client.intervals = decodeExtendedMetricsIntervals(t, `{"id":"activity-2","analyzed":true,"icu_intervals":[{"id":"i1","label":"Work","average_dfa_a1":0.75},{"id":"i2","label":"Rest"}]}`)
	client.powerVsHR = decodeExtendedMetricsPowerVsHR(t, `{"powerHr":1.6,"decoupling":3.4}`)
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-2","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	metrics := got["metrics"].(map[string]any)
	paceZones := metrics["pace_zone_time_seconds"].([]any)
	if paceZones[0] != float64(50) || paceZones[1] != float64(60) {
		t.Fatalf("pace zones = %#v, want GAP zone times", paceZones)
	}
	rows := got["intervals"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["dfa_alpha1"] != 0.75 {
		t.Fatalf("intervals = %#v, want only interval with extended metrics", rows)
	}
	full := got["full"].(map[string]any)
	if full["activity"] == nil || full["intervals"] == nil || full["power_vs_hr"] == nil {
		t.Fatalf("full = %#v, want all available raw sources", full)
	}
}

func TestExtendedMetricsSourceErrorsReturnUserError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		activityErr error
		intervalErr error
		powerErr    error
	}{
		{name: "non optional intervals error", intervalErr: errors.New("intervals exploded")},
		{name: "non optional power-vs-hr error", powerErr: errors.New("power exploded")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeExtendedMetricsClient(t)
			client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-3","name":"Run"}`)
			client.activityErr = tc.activityErr
			client.intervalsErr = tc.intervalErr
			client.powerErr = tc.powerErr
			tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-3"}`)})
			if err == nil || !strings.Contains(err.Error(), fetchExtendedMetricsMessage) {
				t.Fatalf("Handler() error = %v, want user-facing fetch error", err)
			}
		})
	}
}

func TestExtendedMetricsDropsUnavailableFieldsAndConvertsUnits(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeActivityFileFixture(t, "../../testdata/extended-metrics/activity-detail-extended.json")
	delete(client.activity.Raw, "icu_variability_index")
	client.intervals = decodeIntervalsFileFixture(t, "../../testdata/extended-metrics/activity-intervals-extended.json")
	client.powerVsHR = decodePowerVsHRFixture(t, "../../testdata/extended-metrics/activity-power-vs-hr.json")

	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-redacted"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	metrics := got["metrics"].(map[string]any)
	if _, ok := metrics["ground_contact_time_ms"]; ok {
		t.Fatal("unavailable ground_contact_time_ms was emitted")
	}
	if _, ok := metrics["variability_index"]; ok {
		t.Fatal("omitted upstream variability_index was zero-filled/emitted")
	}
	if metrics["joules_above_ftp_kj"] != float64(0.042) {
		t.Fatalf("joules_above_ftp_kj = %v, want 0.042", metrics["joules_above_ftp_kj"])
	}
	intervalRows := got["intervals"].([]any)
	first := intervalRows[0].(map[string]any)
	if first["w_prime_balance_start_kj"] != float64(20.1) || first["joules_above_ftp_kj"] != float64(0.018) {
		t.Fatalf("interval converted units = %#v", first)
	}
	meta := got["_meta"].(map[string]any)
	scales := meta["scales"].(map[string]any)
	if scales["feel"] == nil || scales["rpe"] == nil || scales["session_rpe"] == nil {
		t.Fatalf("scales = %#v", scales)
	}
}

func TestExtendedMetricsJouleFieldsConvertToKilojoules(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-j","name":"Ride","icu_joules_above_ftp":42000,"ss_w_prime":19500}`)
	client.intervals = decodeExtendedMetricsIntervals(t, `{"id":"activity-j","analyzed":true,"icu_intervals":[{"id":"i1","label":"Work","wbal_start":20100,"wbal_end":18300,"joules_above_ftp":18000}]}`)
	client.powerErr = intervals.ErrNotFound
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-j"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	metrics := payload["metrics"].(map[string]any)
	if metrics["joules_above_ftp_kj"] != float64(42) || metrics["strain_score_w_prime_kj"] != float64(19.5) {
		t.Fatalf("activity joule conversions = %#v, want raw joules divided to kJ", metrics)
	}
	for _, key := range []string{"icu_joules_above_ftp", "ss_w_prime", "joules_above_ftp", "w_prime_balance"} {
		if _, ok := metrics[key]; ok {
			t.Fatalf("activity metrics emitted raw/ambiguous Joule key %s: %#v", key, metrics)
		}
	}
	intervalRows := payload["intervals"].([]any)
	first := intervalRows[0].(map[string]any)
	if first["w_prime_balance_start_kj"] != float64(20.1) || first["w_prime_balance_end_kj"] != float64(18.3) || first["joules_above_ftp_kj"] != float64(18) {
		t.Fatalf("interval joule conversions = %#v, want raw joules divided to kJ", first)
	}
	for _, key := range []string{"wbal_start", "wbal_end", "joules_above_ftp", "joules"} {
		if _, ok := first[key]; ok {
			t.Fatalf("interval row emitted raw/ambiguous Joule key %s: %#v", key, first)
		}
	}
	units := payload["_meta"].(map[string]any)["extended_metric_units"].(map[string]any)
	for _, key := range []string{"joules_above_ftp_kj", "strain_score_w_prime_kj", "w_prime_balance_start_kj", "w_prime_balance_end_kj"} {
		if units[key] != "KJ" {
			t.Fatalf("extended_metric_units[%s] = %v, want KJ in %#v", key, units[key], units)
		}
	}
}

func TestExtendedMetricsJouleDerivedZeroValuesStayExplicitKJ(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-zero-j","name":"Ride","icu_joules_above_ftp":0,"ss_w_prime":0}`)
	client.intervals = decodeExtendedMetricsIntervals(t, `{"id":"activity-zero-j","analyzed":true,"icu_intervals":[{"id":"i-zero","label":"Recovery","wbal_start":0,"wbal_end":0,"joules_above_ftp":0}]}`)
	client.powerErr = intervals.ErrNotFound
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-zero-j"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	metrics := payload["metrics"].(map[string]any)
	for _, key := range []string{"joules_above_ftp_kj", "strain_score_w_prime_kj"} {
		if metrics[key] != float64(0) {
			t.Fatalf("metrics[%s] = %v, want explicit zero kJ in %#v", key, metrics[key], metrics)
		}
	}
	first := payload["intervals"].([]any)[0].(map[string]any)
	for _, key := range []string{"w_prime_balance_start_kj", "w_prime_balance_end_kj", "joules_above_ftp_kj"} {
		if first[key] != float64(0) {
			t.Fatalf("interval[%s] = %v, want explicit zero kJ in %#v", key, first[key], first)
		}
	}
}

func TestExtendedMetricsStrainScoreModelParameters(t *testing.T) {
	t.Parallel()

	t.Run("present and W' converted to kJ", func(t *testing.T) {
		t.Parallel()

		client := newFakeExtendedMetricsClient(t)
		client.activity = decodeActivityFileFixture(t, "../../testdata/extended-metrics/activity-detail-extended.json")
		client.intervalsErr = intervals.ErrNotFound
		client.powerErr = intervals.ErrNotFound
		tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-redacted"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		metrics := resultMap(t, result)["metrics"].(map[string]any)
		if metrics["strain_score_cp_watts"] != float64(265) {
			t.Fatalf("strain_score_cp_watts = %v, want 265", metrics["strain_score_cp_watts"])
		}
		if metrics["strain_score_w_prime_kj"] != float64(19.5) {
			t.Fatalf("strain_score_w_prime_kj = %v, want 19.5", metrics["strain_score_w_prime_kj"])
		}
		if metrics["strain_score_p_max_watts"] != float64(1100) {
			t.Fatalf("strain_score_p_max_watts = %v, want 1100", metrics["strain_score_p_max_watts"])
		}
		units := resultMap(t, result)["_meta"].(map[string]any)["extended_metric_units"].(map[string]any)
		if units["strain_score_cp_watts"] != "WATTS" || units["strain_score_w_prime_kj"] != "KJ" || units["strain_score_p_max_watts"] != "WATTS" {
			t.Fatalf("extended_metric_units = %#v", units)
		}
	})

	t.Run("omitted when upstream lacks the model fit", func(t *testing.T) {
		t.Parallel()

		client := newFakeExtendedMetricsClient(t)
		client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-9","name":"Ride","strain_score":40,"ss_cp":null,"ss_w_prime":null,"ss_p_max":null}`)
		client.intervalsErr = intervals.ErrNotFound
		client.powerErr = intervals.ErrNotFound
		tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

		result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-9"}`)})
		if err != nil {
			t.Fatalf("Handler() error = %v", err)
		}
		metrics := resultMap(t, result)["metrics"].(map[string]any)
		for _, key := range []string{"strain_score_cp_watts", "strain_score_w_prime_kj", "strain_score_p_max_watts"} {
			if _, ok := metrics[key]; ok {
				t.Fatalf("%s emitted despite null upstream value", key)
			}
		}
	})
}

func TestExtendedMetricsRawHelpersHandleEdgeTypes(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"int":       7,
		"json":      json.Number("8.5"),
		"bad_json":  json.Number("bad"),
		"slice":     []any{float64(1), 2, json.Number("3.5"), "skip"},
		"bad_slice": "not a slice",
		"blank":     "  ",
		"name":      " Device ",
	}
	if got, ok := rawNumber(raw, "int"); !ok || got != 7 {
		t.Fatalf("rawNumber(int) = (%v, %v), want (7, true)", got, ok)
	}
	if got, ok := rawNumber(raw, "json"); !ok || got != 8.5 {
		t.Fatalf("rawNumber(json) = (%v, %v), want (8.5, true)", got, ok)
	}
	if _, ok := rawNumber(raw, "bad_json"); ok {
		t.Fatal("rawNumber(bad_json) ok = true, want false")
	}
	if got := rawNumberSlice(raw, "slice"); len(got) != 3 || got[2] != 3.5 {
		t.Fatalf("rawNumberSlice() = %#v, want numeric values only", got)
	}
	if got := rawNumberSlice(raw, "bad_slice"); got != nil {
		t.Fatalf("rawNumberSlice(bad_slice) = %#v, want nil", got)
	}
	if got := rawString(raw, "blank"); got != "" {
		t.Fatalf("rawString(blank) = %q, want empty", got)
	}
	if got := rawString(raw, "name"); got != "Device" {
		t.Fatalf("rawString(name) = %q, want trimmed", got)
	}
}

func newFakeExtendedMetricsClient(t *testing.T) *fakeExtendedMetricsClient {
	t.Helper()
	return &fakeExtendedMetricsClient{fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}}
}

func decodeExtendedMetricsActivity(t *testing.T, raw string) intervals.Activity {
	t.Helper()
	var out intervals.Activity
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("decode activity: %v", err)
	}
	return out
}

func decodeExtendedMetricsIntervals(t *testing.T, raw string) intervals.IntervalsDTO {
	t.Helper()
	var out intervals.IntervalsDTO
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("decode intervals: %v", err)
	}
	return out
}

func decodeExtendedMetricsPowerVsHR(t *testing.T, raw string) intervals.PowerVsHR {
	t.Helper()
	var out intervals.PowerVsHR
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("decode power-vs-hr: %v", err)
	}
	return out
}

func decodeActivityFileFixture(t *testing.T, path string) intervals.Activity {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read activity fixture: %v", err)
	}
	var out intervals.Activity
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode activity fixture: %v", err)
	}
	return out
}

func decodeIntervalsFileFixture(t *testing.T, path string) intervals.IntervalsDTO {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read intervals fixture: %v", err)
	}
	var out intervals.IntervalsDTO
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode intervals fixture: %v", err)
	}
	return out
}

func decodePowerVsHRFixture(t *testing.T, path string) intervals.PowerVsHR {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read power-vs-hr fixture: %v", err)
	}
	var out intervals.PowerVsHR
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode power-vs-hr fixture: %v", err)
	}
	return out
}
