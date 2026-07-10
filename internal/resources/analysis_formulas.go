package resources

import (
	"context"
	"strings"
)

const (
	AnalysisFormulasURI      = "icuvisor://analysis-formulas"
	AnalysisFormulasMIMEType = "text/markdown"

	AnalysisFormulaRefHRDrift                 = AnalysisFormulasURI + "#hr_drift"
	AnalysisFormulaRefPwHRDecoupling          = AnalysisFormulasURI + "#pw_hr_decoupling"
	AnalysisFormulaRefPolarization            = AnalysisFormulasURI + "#polarization_index"
	AnalysisFormulaRefEfficiencyFactor        = AnalysisFormulasURI + "#efficiency_factor"
	AnalysisFormulaRefVariabilityIndex        = AnalysisFormulasURI + "#variability_index"
	AnalysisFormulaRefZScore                  = AnalysisFormulasURI + "#z_score"
	AnalysisFormulaRefPerformancePotential    = AnalysisFormulasURI + "#performance_potential"
	AnalysisFormulaRefPowerZoneMechanicalWork = AnalysisFormulasURI + "#power_zone_mechanical_work"
)

type analysisFormulaEntry struct {
	ref       string
	label     string
	paragraph string
}

// analysisFormulaEntries is a public formula contract: changing refs or formula text is a breaking definition-drift event, not a routine refactor.
var analysisFormulaEntries = []analysisFormulaEntry{
	{
		ref:       AnalysisFormulaRefHRDrift,
		label:     "HR drift",
		paragraph: "HR drift reports heart-rate-only change across an eligible steady segment: split elapsed moving time into equal first and second halves and calculate `100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half`. Interpret it only when external load is stable enough for the comparison; require positive average HR in both halves and return insufficient data instead of forcing the formula when the segment is not steady. Sources: Joe Friel, \"Aerobic Endurance and Decoupling\" (joefrieltraining.com), and TrainingPeaks public education on aerobic decoupling/cardiac drift.",
	},
	{
		ref:       AnalysisFormulaRefPwHRDecoupling,
		label:     "Pw:HR decoupling",
		paragraph: "Pw:HR decoupling compares power per heartbeat across an eligible cycling segment: split elapsed moving time into equal first and second halves, compute `ratio_first = avg_power_first_half / avg_hr_first_half` and `ratio_second = avg_power_second_half / avg_hr_second_half`, then report `100 * (ratio_first - ratio_second) / ratio_first` so positive values mean less power per heartbeat later in the segment. Require power and HR in both halves with positive denominators; pace-based siblings should use their own future ref rather than overloading this one. Sources: Joe Friel, \"Aerobic Endurance and Decoupling\" (joefrieltraining.com), and TrainingPeaks/WKO public documentation on Pw:HR decoupling.",
	},
	{
		ref:       AnalysisFormulaRefPolarization,
		label:     "Polarization index",
		paragraph: "Polarization index uses a three-bucket intensity distribution with low = time in Z1+Z2, moderate = time in Z3, and high = time in Z4+; for nonzero moderate and high shares, calculate `log10((low_share / moderate_share) * (high_share / moderate_share) * 100)` from fractional shares. Require total bucketed time above zero, return an explicit saturated/undefined state rather than dividing by zero when moderate share is zero, and treat PI as undefined for polarized classification when high share is zero so bucket shares drive a non-polarized label. Sources: Stephen Seiler public work on three-zone endurance intensity distribution, and Treff et al., \"The Polarization-Index: a simple calculation to distinguish polarized from non-polarized training intensity distributions\" (Frontiers in Physiology, 2017).",
	},
	{
		ref:       AnalysisFormulaRefEfficiencyFactor,
		label:     "Efficiency factor (EF)",
		paragraph: "Cycling efficiency factor is `normalized_power / avg_hr` for the selected activity or segment, summarizing normalized output per heartbeat and staying distinct from decoupling, which compares EF-like ratios across halves. Require normalized power and positive average HR; for non-cycling activities or missing normalized power, return unavailable unless a later sport-specific ref defines a normalized pace or speed equivalent. Sources: Allen and Coggan, Training and Racing with a Power Meter, and TrainingPeaks/WKO public documentation on Efficiency Factor and Normalized Power.",
	},
	{
		ref:       AnalysisFormulaRefVariabilityIndex,
		label:     "Variability index (VI)",
		paragraph: "Cycling variability index is `normalized_power / avg_power` for the selected activity or segment; values near 1.00 indicate steadier output and higher values indicate more variable output. Require normalized power and positive average power; if normalized power is unavailable or the sport lacks power data, return unavailable instead of substituting raw speed or pace. Sources: Allen and Coggan, Training and Racing with a Power Meter, and TrainingPeaks/WKO public documentation on Variability Index and Normalized Power.",
	},
	{
		ref:       AnalysisFormulaRefZScore,
		label:     "z-score",
		paragraph: "For baseline comparisons, z-score is `z = (current_value - baseline_mean) / sample_standard_deviation`, where the baseline mean and sample standard deviation are calculated over the analyzer-selected baseline window after skipping missing days. Require at least the analyzer minimum baseline sample (`n >= 7` unless a tool sets a stricter rule), use the sample standard deviation denominator (`n-1`), and report insufficient variance when standard deviation is zero. Sources: NIST/SEMATECH e-Handbook of Statistical Methods on standard scores and sample standard deviation, and PRD §7.2.C analyzer minimum-sample and missing-day rules.",
	},
	{
		ref:       AnalysisFormulaRefPerformancePotential,
		label:     "Performance potential summary",
		paragraph: "Performance potential is a deterministic per-sport summary, not a numeric score: copy only explicit athlete-profile threshold fields for the requested sport family (FTP/indoor FTP/W′/Pmax for power sports, LTHR/max HR for HR-supported sports, threshold pace in the athlete/sport pace unit for pace sports) and pair them with selected upstream power, pace, and heart-rate curve anchors for the requested date range. Return unsupported or missing critical power, aerobic threshold, profile thresholds, and absent curves as explicit unavailable/caveat objects; never zero-fill missing thresholds, estimate hidden physiology from curve shape, or compare watts and pace across sports as equivalent units. Sources: intervals.icu athlete profile sport settings and upstream curve endpoints exposed through get_athlete_profile, get_power_curves, get_pace_curves, and get_hr_curves.",
	},
	{
		ref:       AnalysisFormulaRefPowerZoneMechanicalWork,
		label:     "Power-zone mechanical work",
		paragraph: "Power-zone mechanical work integrates canonical recorded power over elapsed sample timestamps using the left endpoint: for each eligible interval calculate `delta_t_i = t_(i+1) - t_i`, assign `delta_t_i` and `work_i = power_i * delta_t_i` to the configured lower-inclusive, upper-exclusive power zone containing `power_i`, then sum zone seconds and joules and convert with `zone_kJ = zone_joules / 1000`. Require finite timestamps, `0 < delta_t_i <= 60 seconds`, and finite nonnegative left-endpoint power; skip invalid or longer intervals, do not interpolate missing power, and give the final sample zero duration because it has no following timestamp. Reported kJ is external mechanical work only, not metabolic energy, calorie expenditure, or food calories. Source: BIPM, The International System of Units (SI Brochure), 9th edition, definitions of the joule and watt (`W = J/s`).",
	},
}

// AnalysisFormulasResource returns the analyzer formula registry resource definition.
func AnalysisFormulasResource() Resource {
	return Resource{
		URI:         AnalysisFormulasURI,
		Name:        "analysis_formulas",
		Title:       "Analysis formulas",
		Description: "Canonical formula references used by icuvisor analyzer _meta.formula_ref values.",
		MIMEType:    AnalysisFormulasMIMEType,
		Handler: func(ctx context.Context, _ Request) (Result, error) {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			return Result{URI: AnalysisFormulasURI, MIMEType: AnalysisFormulasMIMEType, Text: AnalysisFormulasMarkdown()}, nil
		},
	}
}

// AnalysisFormulasMarkdown renders the canonical analyzer formula registry.
func AnalysisFormulasMarkdown() string {
	var b strings.Builder
	b.WriteString("# Analysis formulas\n\n")
	b.WriteString("Canonical formula refs for analyzer `_meta.formula_ref` values. These definitions are intentionally stable; changing a ref or formula is a breaking definition-drift event.\n\n")
	b.WriteString("## Formula refs\n")
	for _, entry := range analysisFormulaEntries {
		b.WriteString("\n### ")
		b.WriteString(entry.label)
		b.WriteString("\n\nRef: `")
		b.WriteString(entry.ref)
		b.WriteString("`. ")
		b.WriteString(entry.paragraph)
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}
