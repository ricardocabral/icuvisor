package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetPowerCurvesToolContractUnchanged(t *testing.T) {
	t.Parallel()

	tool := newGetPowerCurvesTool(newFakeFitnessMetricsClient(t), "test", false)
	if tool.Name != getPowerCurvesName {
		t.Fatalf("tool name = %q, want %q", tool.Name, getPowerCurvesName)
	}
	if tool.Description != getPowerCurvesDescription {
		t.Fatalf("tool description changed: %q", tool.Description)
	}
	if tool.Requirement.effective() != RequirementRead {
		t.Fatalf("requirement = %q, want read", tool.Requirement.effective())
	}
	if got := tool.EffectiveToolset().String(); got != "full" {
		t.Fatalf("toolset = %q, want full", got)
	}
	schema := tool.InputSchema.(map[string]any)
	if required, ok := schema["required"].([]string); !ok || len(required) != 2 || required[0] != "oldest" || required[1] != "newest" {
		t.Fatalf("required schema fields = %#v", schema["required"])
	}
	properties := schema["properties"].(map[string]any)
	if _, ok := properties["duration_seconds"]; !ok {
		t.Fatalf("duration_seconds missing from schema properties: %#v", properties)
	}
}

func TestGetPowerCurvesToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetPowerCurvesTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","duration_seconds":[60,300]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	points := got["points"].([]any)
	if len(points) != 2 || points[0].(map[string]any)["duration_seconds"] != float64(60) || points[1].(map[string]any)["watts"] != float64(310) {
		t.Fatalf("points = %#v", points)
	}
	if _, ok := got["full"]; ok {
		t.Fatal("full payload present in terse power curve response")
	}
}

func TestGetPowerCurvesRequestsDurabilityCurveSpecs(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetPowerCurvesTool(client, "test", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","sport":"Ride","duration_seconds":[60,300]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.powerCalls) != 1 {
		t.Fatalf("power calls = %#v, want one call", client.powerCalls)
	}
	call := client.powerCalls[0]
	wantCurveSpec := "r.2026-05-01.2026-05-07,r.2026-05-01.2026-05-07-kj0,r.2026-05-01.2026-05-07-kj1"
	if call.CurveSpec != wantCurveSpec {
		t.Fatalf("CurveSpec = %q, want %q", call.CurveSpec, wantCurveSpec)
	}
	if call.Sport != "Ride" || len(call.DurationSeconds) != 2 || call.DurationSeconds[0] != 60 || call.DurationSeconds[1] != 300 {
		t.Fatalf("curve params = %#v", call)
	}
}

func TestGetPowerCurvesShapesDurabilityCurves(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.curves["Ride:power"] = decodeCurveSet(t, []byte(`{
		"list":[
			{"id":"base","secs":[60,300],"values":[420,310],"activity_id":["base60","base300"]},
			{"id":"fatigue0","after_kj":1000,"secs":[60,300],"values":[390,285],"activity_id":["fatigued60","fatigued300"]},
			{"id":"fatigue1","after_kj":1500,"secs":[60,300],"values":[370,270],"activity_id":["deep60","deep300"]}
		],
		"activities":{}
	}`))
	tool := newGetPowerCurvesTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","duration_seconds":[60,300]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	points := got["points"].([]any)
	if points[0].(map[string]any)["watts"] != float64(420) || points[1].(map[string]any)["activity_id"] != "base300" {
		t.Fatalf("base points = %#v", points)
	}
	durability := got["durability_curves"].([]any)
	if len(durability) != 2 {
		t.Fatalf("durability_curves = %#v, want two curves", durability)
	}
	first := durability[0].(map[string]any)
	threshold := first["work_threshold"].(map[string]any)
	if first["after_kj"] != float64(1000) || threshold["value"] != float64(1000) || threshold["unit"] != "kJ" {
		t.Fatalf("first durability threshold = %#v", first)
	}
	fatiguedPoints := first["points"].([]any)
	if fatiguedPoints[0].(map[string]any)["watts"] != float64(390) || fatiguedPoints[1].(map[string]any)["activity_id"] != "fatigued300" {
		t.Fatalf("fatigued points = %#v", fatiguedPoints)
	}
	meta := got["_meta"].(map[string]any)
	if meta["durability_curve_count"] != float64(2) || meta["work_threshold_unit"] != "kJ" {
		t.Fatalf("meta = %#v", meta)
	}
}

func TestGetPowerCurvesIncludesFullOnlyWhenRequested(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetPowerCurvesTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","include_full":true,"duration_seconds":[60]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	if _, ok := got["full"].(map[string]any); !ok {
		t.Fatalf("full payload missing or wrong type: %#v", got["full"])
	}
	meta := got["_meta"].(map[string]any)
	if meta["include_full"] != true {
		t.Fatalf("include_full meta = %#v", meta["include_full"])
	}
}

func TestGetPowerCurvesOmitsUnavailableDurabilityCurves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "no configured threshold returned",
			body: `{"list":[{"id":"base","secs":[60,300],"values":[420,310],"activity_id":["a1","a2"]}],"activities":{}}`,
		},
		{
			name: "durability threshold not fitted for requested buckets",
			body: `{"list":[{"id":"base","secs":[60,300],"values":[420,310],"activity_id":["a1","a2"]},{"id":"fatigue0","after_kj":1000,"secs":[1200],"values":[250],"activity_id":["a3"]}],"activities":{}}`,
		},
		{
			name: "durability threshold has zero-filled no power data",
			body: `{"list":[{"id":"base","secs":[60,300],"values":[420,310],"activity_id":["a1","a2"]},{"id":"fatigue0","after_kj":1000,"secs":[60,300],"values":[0,0],"activity_id":["a3","a4"]}],"activities":{}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newFakeFitnessMetricsClient(t)
			client.curves["Ride:power"] = decodeCurveSet(t, []byte(tc.body))
			tool := newGetPowerCurvesTool(client, "test", false)
			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","duration_seconds":[60,300]}`)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			got := resultMap(t, result)
			if _, ok := got["durability_curves"]; ok {
				t.Fatalf("durability_curves present for unavailable data: %#v", got["durability_curves"])
			}
			meta := got["_meta"].(map[string]any)
			if _, ok := meta["durability_curve_count"]; ok {
				t.Fatalf("durability_curve_count present in meta: %#v", meta)
			}
			if _, ok := meta["work_threshold_unit"]; ok {
				t.Fatalf("work_threshold_unit present in meta: %#v", meta)
			}
		})
	}
}

func TestGetPowerCurvesMissingBucketsAndDefaults(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetPowerCurvesTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	if got["sport"] != defaultPowerCurveSport {
		t.Fatalf("sport = %#v, want %q", got["sport"], defaultPowerCurveSport)
	}
	meta := got["_meta"].(map[string]any)
	if meta["curve_spec"] != "r.2026-05-01.2026-05-07" {
		t.Fatalf("curve_spec = %#v", meta["curve_spec"])
	}
	durations := meta["duration_seconds"].([]any)
	if len(durations) != len(defaultDurationBuckets) || durations[0] != float64(defaultDurationBuckets[0]) {
		t.Fatalf("duration_seconds = %#v", durations)
	}
	missing := meta["missing_buckets"].([]any)
	if len(missing) == 0 || missing[0] != float64(5) {
		t.Fatalf("missing_buckets = %#v", missing)
	}
}
