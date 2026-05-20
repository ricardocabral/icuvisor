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
