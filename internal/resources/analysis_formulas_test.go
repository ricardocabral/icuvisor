package resources

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestAnalysisFormulasMarkdownGolden(t *testing.T) {
	t.Parallel()

	got := AnalysisFormulasMarkdown()
	want, err := os.ReadFile("testdata/analysis_formulas.md")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("AnalysisFormulasMarkdown() mismatch with testdata/analysis_formulas.md")
	}
}

func TestAnalysisFormulasMarkdownPinsRequiredFormulaRefs(t *testing.T) {
	t.Parallel()

	markdown := AnalysisFormulasMarkdown()
	checks := []struct {
		ref      string
		formula  string
		boundary string
		citation string
	}{
		{
			ref:      AnalysisFormulasURI + "#hr_drift",
			formula:  "100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half",
			boundary: "external load is stable",
			citation: "Joe Friel",
		},
		{
			ref:      AnalysisFormulasURI + "#pw_hr_decoupling",
			formula:  "100 * (ratio_first - ratio_second) / ratio_first",
			boundary: "power and HR in both halves",
			citation: "TrainingPeaks/WKO",
		},
		{
			ref:      AnalysisFormulasURI + "#polarization_index",
			formula:  "log10((low_share / moderate_share) * (high_share / moderate_share) * 100)",
			boundary: "high share is zero",
			citation: "Seiler",
		},
		{
			ref:      AnalysisFormulasURI + "#efficiency_factor",
			formula:  "normalized_power / avg_hr",
			boundary: "missing normalized power",
			citation: "Coggan",
		},
		{
			ref:      AnalysisFormulasURI + "#variability_index",
			formula:  "normalized_power / avg_power",
			boundary: "sport lacks power data",
			citation: "Variability Index",
		},
		{
			ref:      AnalysisFormulasURI + "#z_score",
			formula:  "(current_value - baseline_mean) / sample_standard_deviation",
			boundary: "standard deviation is zero",
			citation: "NIST/SEMATECH",
		},
	}
	for _, check := range checks {
		t.Run(check.ref, func(t *testing.T) {
			for _, want := range []string{check.formula, check.boundary, check.citation} {
				if !strings.Contains(markdown, want) {
					t.Fatalf("markdown missing %q", want)
				}
			}
			if count := strings.Count(markdown, check.ref); count != 1 {
				t.Fatalf("ref %s count = %d, want exactly 1", check.ref, count)
			}
		})
	}
}

func TestNewRegistryRegistersAnalysisFormulasResource(t *testing.T) {
	t.Parallel()

	registrar := &captureRegistrar{}
	if err := NewRegistry().Register(context.Background(), registrar); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	var resource Resource
	for _, candidate := range registrar.resources {
		if candidate.URI == AnalysisFormulasURI {
			resource = candidate
			break
		}
	}
	if resource.URI == "" {
		t.Fatalf("registered resources = %#v, missing %s", registrar.resources, AnalysisFormulasURI)
	}
	if resource.Name != "analysis_formulas" || resource.Title != "Analysis formulas" || resource.MIMEType != AnalysisFormulasMIMEType {
		t.Fatalf("resource metadata = %#v, want analysis formulas metadata", resource)
	}

	result, err := resource.Handler(context.Background(), Request{URI: AnalysisFormulasURI})
	if err != nil {
		t.Fatalf("resource handler error = %v", err)
	}
	if result.URI != AnalysisFormulasURI || result.MIMEType != AnalysisFormulasMIMEType || !strings.Contains(result.Text, "# Analysis formulas") {
		t.Fatalf("resource handler result = %#v, want URI/MIME/markdown", result)
	}
}

func TestAnalysisFormulasResourceHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := AnalysisFormulasResource().Handler(ctx, Request{URI: AnalysisFormulasURI})
	if err == nil {
		t.Fatal("handler error = nil, want context cancellation")
	}
}
