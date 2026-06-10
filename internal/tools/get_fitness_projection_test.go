package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetFitnessProjectionStandardRampGolden(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":14,"weekly_ramp_pct":7,"recovery_week_cadence":0}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	assertAnalyzerGolden(t, "analyzer/fitness_projection_standard_ramp.golden.json", result.StructuredContent)
	root := analyzerMap(t, result.StructuredContent)
	if _, ok := root["series"]; ok {
		t.Fatal("terse projection unexpectedly included series")
	}
	meta := analyzerMap(t, root["_meta"])
	for _, key := range []string{"method", "source_tools", "n", "missing_days", "missing_action", "insufficient_sample", "assumptions", "boundaries"} {
		if _, ok := meta[key]; !ok {
			t.Fatalf("mandatory projection _meta.%s missing from %#v", key, meta)
		}
	}
	assumptions := analyzerMap(t, meta["assumptions"])
	if assumptions["model"] != fitnessProjectionModel || assumptions["weekly_ramp_pct"] != float64(7) {
		t.Fatalf("assumptions = %#v", assumptions)
	}
}

func TestGetFitnessProjectionRecoveryWeekFullGolden(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":28,"weekly_ramp_pct":5,"recovery_week_cadence":4,"recovery_week_load_pct":0,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	assertAnalyzerGolden(t, "analyzer/fitness_projection_recovery_week_full.golden.json", result.StructuredContent)
	series := analyzerMap(t, result.StructuredContent)["series"].([]any)
	if len(series) != 29 {
		t.Fatalf("series length = %d, want start point plus 28 projected days", len(series))
	}
	foundRecoveryZero := false
	for _, rawPoint := range series {
		point := rawPoint.(map[string]any)
		if point["training_load_source"] == "modeled_recovery_week" {
			load, exists := point["training_load"]
			if !exists {
				t.Fatalf("recovery point omitted explicit zero training_load: %#v", point)
			}
			if load == float64(0) {
				foundRecoveryZero = true
				break
			}
		}
	}
	if !foundRecoveryZero {
		t.Fatal("full projection series did not include a modeled_recovery_week point with explicit zero training_load")
	}
}

func TestGetFitnessProjectionDefaultsOmittedHorizonAndAllowsCadenceOne(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","recovery_week_cadence":1}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := analyzerMap(t, analyzerMap(t, result.StructuredContent)["_meta"])
	assumptions := analyzerMap(t, meta["assumptions"])
	if assumptions["horizon_days"] != float64(defaultProjectionHorizonDays) || assumptions["recovery_week_cadence"] != float64(1) {
		t.Fatalf("assumptions = %#v, want default horizon and cadence 1 accepted", assumptions)
	}
}

func TestGetFitnessProjectionRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	cases := []struct {
		name string
		args string
	}{
		{name: "free form model", args: `{"start_date":"2026-05-01","horizon_days":7,"model":"coach_magic"}`},
		{name: "both horizon fields", args: `{"start_date":"2026-05-01","horizon_days":7,"horizon_date":"2026-05-08"}`},
		{name: "ramp out of bounds", args: `{"start_date":"2026-05-01","horizon_days":7,"weekly_ramp_pct":75}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatal("Handler() error = nil, want validation error")
			}
			message, ok := PublicErrorMessage(err)
			if !ok || message != invalidFitnessProjectionMessage {
				t.Fatalf("public error = %q (ok=%v), want %q", message, ok, invalidFitnessProjectionMessage)
			}
		})
	}
}

func TestGetFitnessProjectionInsufficientCurrentFitnessData(t *testing.T) {
	t.Parallel()

	client := &fakeFitnessMetricsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		summaries:         decodeSummaries(t, `[{"date":"2026-05-01","fitness":null,"fatigue":80,"form":-8}]`),
	}
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":7}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want insufficient-data error")
	}
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "insufficient current fitness data") {
		t.Fatalf("public error = %q (ok=%v), want insufficient current fitness data", message, ok)
	}
}

func TestGetFitnessProjectionWeeklyPlanTargetsMetadataAndSources(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":10,"weekly_ramp_pct":0,"recovery_week_cadence":0,"weekly_plan_targets":[{"week_start_date":"2026-04-27","training_load":700},{"week_start_date":"2026-05-04","training_load":840}],"planned_daily_loads":[{"date":"2026-05-05","training_load":42}],"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}

	root := analyzerMap(t, result.StructuredContent)
	meta := analyzerMap(t, root["_meta"])
	if !anyStringSliceContains(t, meta["source_tools"], getTrainingPlanName) || !anyStringSliceContains(t, meta["source_tools"], getFitnessName) {
		t.Fatalf("source_tools = %#v, want get_fitness and get_training_plan", meta["source_tools"])
	}
	assumptions := analyzerMap(t, meta["assumptions"])
	if assumptions["weekly_plan_target_count"] != float64(2) || assumptions["weekly_plan_target_filled_day_count"] != float64(6) || assumptions["weekly_plan_target_override_count"] != float64(1) {
		t.Fatalf("weekly target assumptions = %#v", assumptions)
	}
	if assumptions["weekly_plan_target_week_anchor"] != "athlete-local ISO Monday week_start_date" || !strings.Contains(assumptions["weekly_plan_target_distribution"].(string), "training_load/7") {
		t.Fatalf("weekly target distribution assumptions = %#v", assumptions)
	}

	series := root["series"].([]any)
	assertProjectionSeriesPoint(t, series, "2026-05-02", "weekly_plan_targets", 100)
	assertProjectionSeriesPoint(t, series, "2026-05-05", "planned_daily_loads", 42)
	assertProjectionSeriesPoint(t, series, "2026-05-06", "weekly_plan_targets", 120)
}

func TestDecodeFitnessProjectionRequestValidatesWeeklyPlanTargets(t *testing.T) {
	t.Parallel()

	valid, err := decodeFitnessProjectionRequest(json.RawMessage(`{"start_date":"2026-06-03","horizon_days":3,"weekly_plan_targets":[{"week_start_date":"2026-06-01","training_load":700}]}`))
	if err != nil {
		t.Fatalf("decode valid overlapping current-week target error = %v", err)
	}
	targets := reflect.ValueOf(valid).FieldByName("WeeklyPlanTargets")
	if !targets.IsValid() || targets.Len() != 1 {
		t.Fatalf("WeeklyPlanTargets = %#v, want one normalized target", valid)
	}

	cases := []struct {
		name    string
		args    string
		wantErr string
	}{
		{name: "duplicate normalized week", args: `{"start_date":"2026-06-03","horizon_days":7,"weekly_plan_targets":[{"week_start_date":"2026-06-01","training_load":700},{"week_start_date":" 2026-06-01 ","training_load":500}]}`, wantErr: "weekly_plan_targets contains duplicate week_start_date 2026-06-01"},
		{name: "non monday anchor", args: `{"start_date":"2026-06-03","horizon_days":7,"weekly_plan_targets":[{"week_start_date":"2026-06-02","training_load":700}]}`, wantErr: "weekly_plan_targets.week_start_date must be a Monday YYYY-MM-DD"},
		{name: "negative load", args: `{"start_date":"2026-06-03","horizon_days":7,"weekly_plan_targets":[{"week_start_date":"2026-06-01","training_load":-1}]}`, wantErr: "weekly_plan_targets training_load for 2026-06-01 must be between 0 and 7000"},
		{name: "too large load", args: `{"start_date":"2026-06-03","horizon_days":7,"weekly_plan_targets":[{"week_start_date":"2026-06-01","training_load":7001}]}`, wantErr: "weekly_plan_targets training_load for 2026-06-01 must be between 0 and 7000"},
		{name: "no overlap before horizon", args: `{"start_date":"2026-06-03","horizon_days":7,"weekly_plan_targets":[{"week_start_date":"2026-05-25","training_load":700}]}`, wantErr: "weekly_plan_targets week_start_date 2026-05-25 must overlap the projection horizon"},
		{name: "no overlap after horizon", args: `{"start_date":"2026-06-03","horizon_days":7,"weekly_plan_targets":[{"week_start_date":"2026-06-15","training_load":700}]}`, wantErr: "weekly_plan_targets week_start_date 2026-06-15 must overlap the projection horizon"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodeFitnessProjectionRequest(json.RawMessage(tc.args))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("decode error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestFitnessProjectionSchemaIncludesWeeklyPlanTargets(t *testing.T) {
	t.Parallel()

	properties := fitnessProjectionInputSchema()["properties"].(map[string]any)
	raw, ok := properties["weekly_plan_targets"].(map[string]any)
	if !ok {
		t.Fatalf("schema properties = %#v, want weekly_plan_targets", properties)
	}
	description := raw["description"].(string)
	if !strings.Contains(description, getTrainingPlanName) || !strings.Contains(description, "training_load/7") {
		t.Fatalf("weekly_plan_targets description = %q, want get_training_plan and training_load/7 guidance", description)
	}
}

func assertProjectionSeriesPoint(t *testing.T, series []any, date string, source string, load float64) {
	t.Helper()

	for _, rawPoint := range series {
		point := rawPoint.(map[string]any)
		if point["date"] != date {
			continue
		}
		if point["training_load_source"] != source || point["training_load"] != load {
			t.Fatalf("series point %s = source %q load %v, want source %q load %v", date, point["training_load_source"], point["training_load"], source, load)
		}
		return
	}
	t.Fatalf("series point %s not found in %#v", date, series)
}

func anyStringSliceContains(t *testing.T, raw any, want string) bool {
	t.Helper()

	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("%#v is %T, want []any", raw, raw)
	}
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
