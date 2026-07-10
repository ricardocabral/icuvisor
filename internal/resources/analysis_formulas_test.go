package resources

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestAnalysisFormulasMarkdownGolden(t *testing.T) {
	t.Parallel()

	const wantSHA256 = "1d7aa6e75501914a935f7c3d2f1c15e94d306e170c7ac23f4a3deda66be9295c"

	got := AnalysisFormulasMarkdown()
	want, err := os.ReadFile("testdata/analysis_formulas.md")
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Fatalf("AnalysisFormulasMarkdown() mismatch with testdata/analysis_formulas.md; canonical formula changes are breaking definition-drift events and require product review before updating the golden")
	}
	if gotHash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); gotHash != wantSHA256 {
		t.Fatalf("AnalysisFormulasMarkdown() sha256 = %s, want %s; formula refs/text drifted", gotHash, wantSHA256)
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
			ref:      AnalysisFormulaRefHRDrift,
			formula:  "100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half",
			boundary: "external load is stable",
			citation: "Joe Friel",
		},
		{
			ref:      AnalysisFormulaRefPwHRDecoupling,
			formula:  "100 * (ratio_first - ratio_second) / ratio_first",
			boundary: "power and HR in both halves",
			citation: "TrainingPeaks/WKO",
		},
		{
			ref:      AnalysisFormulaRefPolarization,
			formula:  "log10((low_share / moderate_share) * (high_share / moderate_share) * 100)",
			boundary: "high share is zero",
			citation: "Seiler",
		},
		{
			ref:      AnalysisFormulaRefEfficiencyFactor,
			formula:  "normalized_power / avg_hr",
			boundary: "missing normalized power",
			citation: "Coggan",
		},
		{
			ref:      AnalysisFormulaRefVariabilityIndex,
			formula:  "normalized_power / avg_power",
			boundary: "sport lacks power data",
			citation: "Variability Index",
		},
		{
			ref:      AnalysisFormulaRefZScore,
			formula:  "(current_value - baseline_mean) / sample_standard_deviation",
			boundary: "standard deviation is zero",
			citation: "NIST/SEMATECH",
		},
		{
			ref:      AnalysisFormulaRefPerformancePotential,
			formula:  "copy only explicit athlete-profile threshold fields",
			boundary: "never zero-fill missing thresholds",
			citation: "get_power_curves",
		},
		{
			ref:      AnalysisFormulaRefPowerZoneMechanicalWork,
			formula:  "work_i = power_i * delta_t_i",
			boundary: "final zone open-ended and an explicit below-zone bucket `[0, first_boundary)`",
			citation: "BIPM",
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
