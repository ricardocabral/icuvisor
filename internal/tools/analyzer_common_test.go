package tools

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/analysis"
	"github.com/ricardocabral/icuvisor/internal/resources"
)

func TestShapeAnalyzerResponseIncludesMandatoryMeta(t *testing.T) {
	encoded, err := encodeAnalyzerResponse(analyzerDemoInput(), false, "test", false, "demo_analyzer", "")
	if err != nil {
		t.Fatalf("encode analyzer response: %v", err)
	}
	got := encoded.StructuredContent
	assertAnalyzerGolden(t, "analyzer/demo_terse.golden.json", got)

	root := analyzerMap(t, got)
	meta := analyzerMap(t, root["_meta"])
	for _, key := range []string{"method", "source_tools", "n", "missing_days", "missing_action", "insufficient_sample"} {
		if _, ok := meta[key]; !ok {
			t.Fatalf("mandatory _meta.%s missing from %#v", key, meta)
		}
	}
}

func TestShapeAnalyzerResponseTerseAndFullSeries(t *testing.T) {
	terse, err := shapeAnalyzerResponse(analyzerDemoInput(), false, "test", false, "demo_analyzer", "")
	if err != nil {
		t.Fatalf("shape terse analyzer response: %v", err)
	}
	if _, ok := analyzerMap(t, terse)["series"]; ok {
		t.Fatal("terse analyzer response unexpectedly included series")
	}
	assertAnalyzerMissingDays(t, terse, 2)

	full, err := shapeAnalyzerResponse(analyzerDemoInput(), true, "test", false, "demo_analyzer", "")
	if err != nil {
		t.Fatalf("shape full analyzer response: %v", err)
	}
	assertAnalyzerGolden(t, "analyzer/demo_full.golden.json", full)
	series, ok := analyzerMap(t, full)["series"].([]any)
	if !ok || len(series) != 2 {
		t.Fatalf("full analyzer series = %#v, want two points", analyzerMap(t, full)["series"])
	}
	assertAnalyzerMissingDays(t, full, 2)
}

func assertAnalyzerMissingDays(t *testing.T, value any, want int) {
	t.Helper()
	meta := analyzerMap(t, analyzerMap(t, value)["_meta"])
	got, ok := meta["missing_days"].(float64)
	if !ok {
		t.Fatalf("_meta.missing_days = %T(%#v), want JSON number", meta["missing_days"], meta["missing_days"])
	}
	if got != float64(want) {
		t.Fatalf("_meta.missing_days = %v, want %d", got, want)
	}
}

func analyzerDemoInput() analyzerResponseInput {
	return analyzerResponseInput{
		Result: map[string]any{"summary": "demo", "score": 1.25},
		Series: []map[string]any{
			{"date": "2026-05-01", "value": 1.1},
			{"date": "2026-05-02", "value": 1.4},
		},
		Meta: analysis.AnalyzerMetaInput{
			Method:      "z_score",
			SourceTools: []string{"get_wellness_data"},
			N:           analysis.MinBaselineSamples,
			MissingDays: 2,
			MinSamples:  analysis.MinBaselineSamples,
			FormulaRef:  resources.AnalysisFormulaRefZScore,
		},
	}
}

func analyzerMap(t *testing.T, value any) map[string]any {
	t.Helper()
	out, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value = %T, want map[string]any", value)
	}
	return out
}

func assertAnalyzerGolden(t *testing.T, fixture string, got any) {
	t.Helper()
	gotJSON := canonicalAnalyzerJSON(t, got)
	path := filepath.Join("testdata", fixture)
	if os.Getenv("UPDATE_ANALYZER_GOLDENS") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create analyzer golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(gotJSON), 0o644); err != nil {
			t.Fatalf("update analyzer golden %s: %v", path, err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read analyzer golden %s: %v", path, err)
	}
	if gotJSON != string(bytes.TrimSpace(want))+"\n" {
		t.Fatalf("analyzer golden %s mismatch\n got: %s\nwant: %s", path, gotJSON, string(want))
	}
}

func canonicalAnalyzerJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal analyzer response: %v", err)
	}
	return string(data) + "\n"
}
