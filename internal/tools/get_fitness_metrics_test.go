package tools

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeFitnessMetricsClient struct {
	fakeProfileClient
	summaries []intervals.SummaryWithCats
	activity  intervals.Activity
	intervals intervals.IntervalsDTO
	powerVsHR intervals.PowerVsHR
	curves    map[string]intervals.DataCurveSet
}

func (f *fakeFitnessMetricsClient) ListAthleteSummary(context.Context, intervals.AthleteSummaryParams) ([]intervals.SummaryWithCats, error) {
	return append([]intervals.SummaryWithCats(nil), f.summaries...), nil
}

func (f *fakeFitnessMetricsClient) ListAthletePowerCurves(_ context.Context, params intervals.CurveParams) (intervals.DataCurveSet, error) {
	return f.curves[params.Sport+":power"], nil
}

func (f *fakeFitnessMetricsClient) ListAthleteHRCurves(_ context.Context, params intervals.CurveParams) (intervals.DataCurveSet, error) {
	return f.curves[params.Sport+":hr"], nil
}

func (f *fakeFitnessMetricsClient) ListAthletePaceCurves(_ context.Context, params intervals.CurveParams) (intervals.DataCurveSet, error) {
	return f.curves[params.Sport+":pace"], nil
}

func (f *fakeFitnessMetricsClient) GetActivity(context.Context, string) (intervals.Activity, error) {
	return f.activity, nil
}

func (f *fakeFitnessMetricsClient) GetActivityIntervals(context.Context, string) (intervals.IntervalsDTO, error) {
	return f.intervals, nil
}

func (f *fakeFitnessMetricsClient) GetActivityPowerVsHR(context.Context, string) (intervals.PowerVsHR, error) {
	return f.powerVsHR, nil
}

func TestFitnessMetricsToolShapes(t *testing.T) {
	t.Parallel()

	profile := intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "America/Sao_Paulo"}
	client := &fakeFitnessMetricsClient{fakeProfileClient: fakeProfileClient{profile: profile}}
	client.summaries = decodeSummaries(t, `[
		{"date":"2026-05-02","fitness":72,"fatigue":80,"form":-8,"time":3600,"moving_time":3500,"distance":40000,"training_load":90,"timeInZones":[10,20],"timeInZonesTot":30,"byCategory":[{"category":"Ride","time":3600,"distance":40000,"training_load":90}]},
		{"date":"2026-05-01","fitness":70,"fatigue":78,"form":-8,"time":1800,"moving_time":1700,"distance":5000,"training_load":40,"timeInZones":[5,15],"timeInZonesTot":20,"byCategory":[{"category":"Run","time":1800,"distance":5000,"training_load":40}]}
	]`)
	client.curves = map[string]intervals.DataCurveSet{
		"Ride:power": curveSet(t, []float64{60, 300}, []float64{420, 310}),
		"Ride:hr":    curveSet(t, []float64{60, 300}, []float64{178, 165}),
		"Ride:pace":  distanceCurveSet(t, []float64{1000}, []float64{240}),
		"Run:power":  curveSet(t, []float64{60, 300}, []float64{390, 300}),
		"Run:hr":     curveSet(t, []float64{60, 300}, []float64{182, 170}),
		"Run:pace":   distanceCurveSet(t, []float64{1000}, []float64{230}),
	}

	tests := []struct {
		name   string
		tool   Tool
		args   string
		assert func(t *testing.T, got map[string]any)
	}{
		{
			name: "fitness rows sorted with timezone meta",
			tool: newGetFitnessTool(client, client, "test", "UTC", false),
			args: `{"start_date":"2026-05-01","end_date":"2026-05-02"}`,
			assert: func(t *testing.T, got map[string]any) {
				rows := got["fitness"].([]any)
				if rows[0].(map[string]any)["date"] != "2026-05-01" || rows[1].(map[string]any)["date"] != "2026-05-02" {
					t.Fatalf("fitness row order = %#v", rows)
				}
				meta := got["_meta"].(map[string]any)
				if meta["timezone"] != "America/Sao_Paulo" || meta["server_version"] != "test" {
					t.Fatalf("meta = %#v", meta)
				}
			},
		},
		{
			name: "best efforts grouped by sport and bucket",
			tool: newGetBestEffortsTool(client, "test", false),
			args: `{"sports":["Ride","Run"],"duration_seconds":[60],"distance_meters":[1000]}`,
			assert: func(t *testing.T, got map[string]any) {
				sports := got["sports"].([]any)
				if len(sports) != 2 {
					t.Fatalf("sports count = %d, want 2", len(sports))
				}
				efforts := sports[0].(map[string]any)["efforts"].([]any)
				if len(efforts) != 3 || efforts[0].(map[string]any)["duration_seconds"] != float64(60) {
					t.Fatalf("efforts = %#v", efforts)
				}
			},
		},
		{
			name: "power curve terse buckets omit full arrays",
			tool: newGetPowerCurvesTool(client, "test", false),
			args: `{"oldest":"2026-05-01","newest":"2026-05-07","duration_seconds":[60,300]}`,
			assert: func(t *testing.T, got map[string]any) {
				points := got["points"].([]any)
				if len(points) != 2 || points[1].(map[string]any)["watts"] != float64(310) {
					t.Fatalf("points = %#v", points)
				}
				if _, ok := got["full"]; ok {
					t.Fatal("full payload present in terse power curve response")
				}
			},
		},
		{
			name: "training summary aggregates without TSS relabel",
			tool: newGetTrainingSummaryTool(client, client, "test", "UTC", false),
			args: `{"start_date":"2026-05-01","end_date":"2026-05-02"}`,
			assert: func(t *testing.T, got map[string]any) {
				summary := got["summary"].(map[string]any)
				if summary["training_load"] != float64(130) || summary["distance_km"] != float64(45) {
					t.Fatalf("summary = %#v", summary)
				}
				if _, ok := summary["tss"]; ok {
					t.Fatal("summary should not expose tss")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := tc.tool.Handler(context.Background(), Request{Name: tc.tool.Name, Arguments: json.RawMessage(tc.args)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			tc.assert(t, resultMap(t, result))
		})
	}
}

func TestExtendedMetricsDropsUnavailableFieldsAndConvertsUnits(t *testing.T) {
	t.Parallel()

	profile := intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}
	client := &fakeFitnessMetricsClient{fakeProfileClient: fakeProfileClient{profile: profile}}
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

func decodeSummaries(t *testing.T, text string) []intervals.SummaryWithCats {
	t.Helper()
	var out []intervals.SummaryWithCats
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("decode summaries: %v", err)
	}
	return out
}

func curveSet(t *testing.T, secs []float64, values []float64) intervals.DataCurveSet {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"list": []map[string]any{{"id": "r", "secs": secs, "values": values, "activity_id": []string{"a1", "a2", "a3"}}}, "activities": map[string]any{}})
	var out intervals.DataCurveSet
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode curve set: %v", err)
	}
	return out
}

func distanceCurveSet(t *testing.T, distances []float64, values []float64) intervals.DataCurveSet {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"list": []map[string]any{{"id": "r", "distance": distances, "values": values, "activity_id": []string{"a1", "a2", "a3"}}}, "activities": map[string]any{}})
	var out intervals.DataCurveSet
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("decode distance curve set: %v", err)
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
