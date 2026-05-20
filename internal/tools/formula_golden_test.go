package tools

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type toolFormulaGolden struct {
	Decoupling       toolSegmentFormulaGolden      `json:"decoupling"`
	Polarization     toolPolarizationFormulaGolden `json:"polarization"`
	VariabilityIndex toolVariabilityFormulaGolden  `json:"variability_index"`
	ZScore           toolZScoreFormulaGolden       `json:"z_score"`
}

type toolSegmentFormulaGolden struct {
	FormulaRef string  `json:"formula_ref"`
	Value      float64 `json:"value"`
}

type toolPolarizationFormulaGolden struct {
	FormulaRef     string    `json:"formula_ref"`
	InputZones     []float64 `json:"input_zones"`
	Index          float64   `json:"index"`
	State          string    `json:"state"`
	Classification string    `json:"classification"`
}

type toolVariabilityFormulaGolden struct {
	FormulaRef   string  `json:"formula_ref"`
	SourceField  string  `json:"source_field"`
	OutputField  string  `json:"output_field"`
	FixtureValue float64 `json:"fixture_value"`
}

type toolZScoreFormulaGolden struct {
	FormulaRef   string  `json:"formula_ref"`
	BaselineMean float64 `json:"baseline_mean"`
	SampleStdDev float64 `json:"sample_stddev"`
	ZScore       float64 `json:"z_score"`
}

func loadToolFormulaGolden(t *testing.T) toolFormulaGolden {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "analysis", "formula_golden.json"))
	if err != nil {
		t.Fatalf("read formula golden: %v", err)
	}
	var golden toolFormulaGolden
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("decode formula golden: %v", err)
	}
	return golden
}

func TestFormulaGoldenComputeActivitySegmentStatsTool(t *testing.T) {
	t.Parallel()
	golden := loadToolFormulaGolden(t)
	client := &segmentStatsStreamsClient{rows: []intervals.ActivityStream{
		{Type: "time", Data: []float64{0, 10, 20, 30, 40, 50}},
		{Type: "heartrate", Data: []float64{100, 100, 100, 100, 100, 100}},
		{Type: "watts", Data: []float64{200, 200, 200, 180, 180, 180}},
	}}
	handler := computeActivitySegmentStatsHandler(client, "test", false, responseShaping{})
	result, err := handler(context.Background(), Request{Arguments: json.RawMessage(`{"activity_id":"a1","stat":"decoupling","start_seconds":0,"end_seconds":50,"include_full":true}`)})
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}
	payload := result.StructuredContent.(map[string]any)
	body := payload["result"].(map[string]any)
	meta := payload["_meta"].(map[string]any)
	assertToolGoldenFloat(t, "decoupling result.value", body["value"].(float64), golden.Decoupling.Value, 1e-9)
	if meta["formula_ref"] != golden.Decoupling.FormulaRef {
		t.Fatalf("decoupling _meta.formula_ref = %v, want %s", meta["formula_ref"], golden.Decoupling.FormulaRef)
	}
}

func TestFormulaGoldenZoneTools(t *testing.T) {
	t.Parallel()
	golden := loadToolFormulaGolden(t)
	client := newFakeComputeClient()
	client.summaries = []intervals.SummaryWithCats{{Date: "2026-05-01", TimeInZones: golden.Polarization.InputZones, TimeInZonesTot: 1100, TrainingLoad: 75}}
	zoneTool := newComputeZoneTimeTool(client, client, client, client, "test", "UTC", false)

	zoneResult, err := zoneTool.Handler(context.Background(), Request{Name: zoneTool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","zone_metric":"power"}`)})
	if err != nil {
		t.Fatalf("zone handler error = %v", err)
	}
	zonePayload := resultMap(t, zoneResult)
	zoneBody := zonePayload["result"].(map[string]any)
	zoneMeta := zonePayload["_meta"].(map[string]any)
	assertToolGoldenFloat(t, "zone polarization_index", zoneBody["polarization_index"].(float64), golden.Polarization.Index, 1e-4)
	if zoneBody["polarization_state"] != golden.Polarization.State || zoneBody["classification"] != golden.Polarization.Classification {
		t.Fatalf("zone polarization = %v/%v, want %s/%s", zoneBody["polarization_state"], zoneBody["classification"], golden.Polarization.State, golden.Polarization.Classification)
	}
	if zoneMeta["formula_ref"] != golden.Polarization.FormulaRef {
		t.Fatalf("zone _meta.formula_ref = %v, want %s", zoneMeta["formula_ref"], golden.Polarization.FormulaRef)
	}

	loadTool := newComputeLoadBalanceTool(client, client, client, client, "test", "UTC", false)
	loadResult, err := loadTool.Handler(context.Background(), Request{Name: loadTool.Name, Arguments: []byte(`{"start_date":"2026-05-01","end_date":"2026-05-01","zone_metric":"power"}`)})
	if err != nil {
		t.Fatalf("load-balance handler error = %v", err)
	}
	loadPayload := resultMap(t, loadResult)
	loadBody := loadPayload["result"].(map[string]any)
	loadMeta := loadPayload["_meta"].(map[string]any)
	assertToolGoldenFloat(t, "load polarization_index", loadBody["polarization_index"].(float64), golden.Polarization.Index, 1e-4)
	if loadBody["polarization_state"] != golden.Polarization.State || loadBody["classification"] != golden.Polarization.Classification || loadMeta["formula_ref"] != golden.Polarization.FormulaRef {
		t.Fatalf("load-balance payload/meta = %#v/%#v, want polarization golden", loadBody, loadMeta)
	}
}

func TestFormulaGoldenBaselineTool(t *testing.T) {
	t.Parallel()
	golden := loadToolFormulaGolden(t)
	client := newFakeComputeClient()
	client.wellness = []intervals.Wellness{
		wellnessFixture("2026-05-01", map[string]any{"restingHR": 50.0}),
		wellnessFixture("2026-05-02", map[string]any{"restingHR": 60.0}),
		wellnessFixture("2026-05-04", map[string]any{"restingHR": 70.0}),
	}
	tool := newComputeBaselineTool(nil, client, nil, nil, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: []byte(`{"metric":"rhr","baseline_start_date":"2026-05-01","baseline_end_date":"2026-05-02","current_start_date":"2026-05-04","current_end_date":"2026-05-04","min_samples":2}`)})
	if err != nil {
		t.Fatalf("baseline handler error = %v", err)
	}
	payload := resultMap(t, result)
	body := payload["result"].(map[string]any)
	meta := payload["_meta"].(map[string]any)
	assertToolGoldenFloat(t, "baseline_mean", body["baseline_mean"].(float64), golden.ZScore.BaselineMean, 1e-9)
	assertToolGoldenFloat(t, "baseline_stddev", body["baseline_stddev"].(float64), golden.ZScore.SampleStdDev, 1e-4)
	assertToolGoldenFloat(t, "z_score", body["z_score"].(float64), golden.ZScore.ZScore, 1e-4)
	if meta["formula_ref"] != golden.ZScore.FormulaRef {
		t.Fatalf("baseline _meta.formula_ref = %v, want %s", meta["formula_ref"], golden.ZScore.FormulaRef)
	}
}

func TestFormulaGoldenExtendedMetricsVIIsUpstreamMapped(t *testing.T) {
	t.Parallel()
	golden := loadToolFormulaGolden(t)
	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeActivityFileFixture(t, "../../testdata/extended-metrics/activity-detail-extended.json")
	client.intervals = decodeIntervalsFileFixture(t, "../../testdata/extended-metrics/activity-intervals-extended.json")
	client.powerVsHR = decodePowerVsHRFixture(t, "../../testdata/extended-metrics/activity-power-vs-hr.json")
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-redacted"}`)})
	if err != nil {
		t.Fatalf("extended metrics handler error = %v", err)
	}
	payload := resultMap(t, result)
	metrics := payload["metrics"].(map[string]any)
	assertToolGoldenFloat(t, "variability_index", metrics[golden.VariabilityIndex.OutputField].(float64), golden.VariabilityIndex.FixtureValue, 1e-9)

	delete(client.activity.Raw, golden.VariabilityIndex.SourceField)
	result, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-redacted"}`)})
	if err != nil {
		t.Fatalf("extended metrics missing-VI handler error = %v", err)
	}
	payload = resultMap(t, result)
	metrics = payload["metrics"].(map[string]any)
	if _, ok := metrics[golden.VariabilityIndex.OutputField]; ok {
		t.Fatalf("%s emitted after upstream %s was removed; VI must not be locally recomputed or zero-filled without updating formula guards", golden.VariabilityIndex.OutputField, golden.VariabilityIndex.SourceField)
	}
}

func assertToolGoldenFloat(t *testing.T, name string, got float64, want float64, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %v, want %v (tolerance %g); analyzer formula golden drifted", name, got, want, tolerance)
	}
}
