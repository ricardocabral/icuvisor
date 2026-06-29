package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetTrainingSummaryToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetTrainingSummaryTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	summary := got["summary"].(map[string]any)
	if summary["training_load"] != float64(130) || summary["distance_km"] != float64(45) {
		t.Fatalf("summary = %#v", summary)
	}
	if _, ok := summary["tss"]; ok {
		t.Fatal("summary should not expose tss")
	}
}

func TestGetTrainingSummaryPreservesTRIMPLoadWithoutTSSLabel(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	client.summaries = decodeSummaries(t, `[
		{"date":"2026-05-01","trimp":35,"byCategory":[{"category":"Run","trimp":35}]},
		{"date":"2026-05-02","byCategory":[{"category":"Ride"}]}
	]`)
	tool := newGetTrainingSummaryTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	summary := got["summary"].(map[string]any)
	if summary["training_load"] != float64(35) {
		t.Fatalf("summary = %#v, want TRIMP fallback load preserved", summary)
	}
	if _, ok := summary["tss"]; ok {
		t.Fatal("summary should not expose tss for TRIMP fallback load")
	}
	meta := got["_meta"].(map[string]any)
	diagnostics := meta["load_diagnostics"].([]any)
	if !diagnosticReasonsContain(diagnostics, "trimp_or_hr_load_available") || !diagnosticReasonsContain(diagnostics, "missing_training_load") {
		t.Fatalf("load_diagnostics = %#v, want TRIMP and missing-load diagnostics", diagnostics)
	}
}
