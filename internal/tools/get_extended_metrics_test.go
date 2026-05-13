package tools

import (
	"context"
	"encoding/json"
	"errors"
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

func TestExtendedMetricsStravaUnavailableIncludesFullWhenRequested(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"strava-1","source":"Strava","_note":"Imported from Strava"}`)
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"fallback","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	if got["activity_id"] != "strava-1" || got["strava_imported"] != true {
		t.Fatalf("strava response = %#v", got)
	}
	unavailable := got["unavailable"].(map[string]any)
	if unavailable["reason"] != "strava_tos" || !strings.Contains(unavailable["workaround"].(string), "connect device directly") {
		t.Fatalf("unavailable = %#v", unavailable)
	}
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
		{name: "activity error", activityErr: intervals.ErrUnauthorized},
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
