# Analysis formulas

Canonical formula refs for analyzer `_meta.formula_ref` values. These definitions are intentionally stable; changing a ref or formula is a breaking definition-drift event.

## Formula refs

### HR drift

Ref: `icuvisor://analysis-formulas#hr_drift`. HR drift reports heart-rate-only change across an eligible steady segment: split elapsed moving time into equal first and second halves and calculate `100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half`. Interpret it only when external load is stable enough for the comparison; require positive average HR in both halves and return insufficient data instead of forcing the formula when the segment is not steady. Sources: Joe Friel, "Aerobic Endurance and Decoupling" (joefrieltraining.com), and TrainingPeaks public education on aerobic decoupling/cardiac drift.

### Pw:HR decoupling

Ref: `icuvisor://analysis-formulas#pw_hr_decoupling`. Pw:HR decoupling compares power per heartbeat across an eligible cycling segment: split elapsed moving time into equal first and second halves, compute `ratio_first = avg_power_first_half / avg_hr_first_half` and `ratio_second = avg_power_second_half / avg_hr_second_half`, then report `100 * (ratio_first - ratio_second) / ratio_first` so positive values mean less power per heartbeat later in the segment. Require power and HR in both halves with positive denominators; pace-based siblings should use their own future ref rather than overloading this one. Sources: Joe Friel, "Aerobic Endurance and Decoupling" (joefrieltraining.com), and TrainingPeaks/WKO public documentation on Pw:HR decoupling.

### Polarization index

Ref: `icuvisor://analysis-formulas#polarization_index`. Polarization index uses a three-bucket intensity distribution with low = time in Z1+Z2, moderate = time in Z3, and high = time in Z4+; for nonzero moderate and high shares, calculate `log10((low_share / moderate_share) * (high_share / moderate_share) * 100)` from fractional shares. Require total bucketed time above zero, return an explicit saturated/undefined state rather than dividing by zero when moderate share is zero, and treat PI as undefined for polarized classification when high share is zero so bucket shares drive a non-polarized label. Sources: Stephen Seiler public work on three-zone endurance intensity distribution, and Treff et al., "The Polarization-Index: a simple calculation to distinguish polarized from non-polarized training intensity distributions" (Frontiers in Physiology, 2017).

### Efficiency factor (EF)

Ref: `icuvisor://analysis-formulas#efficiency_factor`. Cycling efficiency factor is `normalized_power / avg_hr` for the selected activity or segment, summarizing normalized output per heartbeat and staying distinct from decoupling, which compares EF-like ratios across halves. Require normalized power and positive average HR; for non-cycling activities or missing normalized power, return unavailable unless a later sport-specific ref defines a normalized pace or speed equivalent. Sources: Allen and Coggan, Training and Racing with a Power Meter, and TrainingPeaks/WKO public documentation on Efficiency Factor and Normalized Power.

### Variability index (VI)

Ref: `icuvisor://analysis-formulas#variability_index`. Cycling variability index is `normalized_power / avg_power` for the selected activity or segment; values near 1.00 indicate steadier output and higher values indicate more variable output. Require normalized power and positive average power; if normalized power is unavailable or the sport lacks power data, return unavailable instead of substituting raw speed or pace. Sources: Allen and Coggan, Training and Racing with a Power Meter, and TrainingPeaks/WKO public documentation on Variability Index and Normalized Power.

### z-score

Ref: `icuvisor://analysis-formulas#z_score`. For baseline comparisons, z-score is `z = (current_value - baseline_mean) / sample_standard_deviation`, where the baseline mean and sample standard deviation are calculated over the analyzer-selected baseline window after skipping missing days. Require at least the analyzer minimum baseline sample (`n >= 7` unless a tool sets a stricter rule), use the sample standard deviation denominator (`n-1`), and report insufficient variance when standard deviation is zero. Sources: NIST/SEMATECH e-Handbook of Statistical Methods on standard scores and sample standard deviation, and PRD §7.2.C analyzer minimum-sample and missing-day rules.
