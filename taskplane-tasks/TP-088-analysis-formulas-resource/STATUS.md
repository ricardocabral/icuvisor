# TP-088: MCP Resource `icuvisor://analysis-formulas` — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 15
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Draft canonical formulas
**Status:** ✅ Complete

- [x] Record a concrete Step 1 formula draft table in STATUS.md Notes with ref, label, canonical equation/method, boundary handling, and citation(s).
- [x] Write concise definitions for HR drift, Pw:HR decoupling, polarization index, EF, VI, and z-score.
- [x] Include citations to accepted public sources; avoid copying proprietary/copyrighted text.
- [x] Assign stable formula refs/anchors.
- [x] Capture edge-case decisions for split semantics, missing NP/power inputs, zone buckets, sample standard deviation, baseline windows, and zero variance.

---

### Step 2: Implement resource
**Status:** ✅ Complete

- [x] Add `internal/resources/analysis_formulas.go` with `AnalysisFormulasURI`, `AnalysisFormulasMIMEType`, `AnalysisFormulasResource()`, and `AnalysisFormulasMarkdown()`.
- [x] Render compact markdown with six one-paragraph entries sourced from the Step 1 draft, exposing exact refs/fragments: `hr_drift`, `pw_hr_decoupling`, `polarization_index`, `efficiency_factor`, `variability_index`, and `z_score`.
- [x] Add `icuvisor://analysis-formulas` to the resources registry.
- [x] Return a compact markdown shape consistent with existing resources (`text/markdown`, resource name `analysis_formulas`).
- [x] Add golden and invariant tests under `internal/resources`, including metadata, handler read, canceled context, exact-once required refs, formula/boundary words, and citation presence.
- [x] Update MCP protocol default-resource coverage so `resources/list` and `resources/read` include `icuvisor://analysis-formulas` with `text/markdown`.
- [x] Carry the high-share-zero polarization-index boundary behavior into the resource, golden file, and invariant test.

---

### Step 3: Wire docs and catalog
**Status:** ✅ Complete

- [x] Update `web/content/reference/resources-prompts.md` to list `icuvisor://analysis-formulas` with name `analysis_formulas`, MIME `text/markdown`, and a description of canonical analyzer formula refs.
- [x] Review README/public docs for affected resource listings; update or log “checked, no change” in `STATUS.md`.
- [x] Add a code-level stable-ref surface for analyzers (exported constants/helpers for `hr_drift`, `pw_hr_decoupling`, `polarization_index`, `efficiency_factor`, `variability_index`, and `z_score`) and test that it matches the markdown refs exactly once.
- [x] Confirm there is no separate generated resource catalog/hash to refresh; do not modify the tool catalog.
- [x] Run targeted resource/MCP tests: `go test ./internal/resources ./internal/mcp`.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Run full suite/build/lint.
- [x] Update CHANGELOG.md.
- [x] Record formula-source decisions in STATUS.md.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Run `git status --short` and confirm the only pending changes are expected task/status/review updates.
- [x] Targeted tests passing: `go test ./internal/resources ./internal/mcp`; record pass/fail in `STATUS.md`.
- [x] FULL test suite passing: `make test`; record pass/fail in `STATUS.md`.
- [x] Build passes: `make build`; record pass/fail in `STATUS.md`.
- [x] Lint passes: `make lint`; record pass/fail in `STATUS.md`.
- [x] All failures fixed or documented as pre-existing unrelated failures.
- [x] Confirm R012’s missing verification-log finding is resolved before moving to Step 6.

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No out-of-scope discoveries; resource-catalog search found only tool-catalog hashing, so no generated resource catalog refresh was needed. | Logged in Step 3 notes; no follow-up required. | STATUS.md Notes |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 14:22 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 14:22 | Step 0 started | Preflight |
| 2026-05-20 15:19 | Worker iter 1 | done in 3415s, tools: 181 |
| 2026-05-20 15:19 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

### Step 1 source-selection rule

Use bibliographic or public web sources with URLs/publication details; summarize in original words; do not quote long passages; do not derive wording from GPL/copyleft implementations. Prefer named public sources aligned with the PRD vocabulary (Friel/aerobic decoupling, Coggan/Allen Training and Racing with a Power Meter concepts, Seiler/polarized training distribution) plus general statistical references for z-score.

### Step 6 delivery notes

- Must-update docs: `CHANGELOG.md` and `STATUS.md` updated.
- Check-if-affected docs: `README.md` updated for resource layout; `web/content/reference/resources-prompts.md` updated as resource reference; PRD was checked via §7.2.C and not changed because implementation matches existing analyzer formula-registry scope.

### Step 5 verification log

- `git status --short` (2026-05-20): only expected task/status/review updates pending (`STATUS.md`, `R014-plan-step5.md`).
- `go test ./internal/resources ./internal/mcp` (2026-05-20): PASS (`internal/resources`, `internal/mcp`).
- `make test` (2026-05-20): PASS (`go test ./...`).
- `make build` (2026-05-20): PASS (`go build ... ./cmd/icuvisor`).
- `make lint` (2026-05-20): PASS (`golangci-lint run ./...`, 0 issues).
- R012 verification-log concern resolved by recording every Step 5 verification command and result above before checking the corresponding items.

### Step 4 formula-source decisions

- HR drift and Pw:HR decoupling cite Joe Friel’s public aerobic decoupling writing and TrainingPeaks/WKO public education; resource wording is original and pins separate HR-only vs power-to-HR semantics.
- EF and VI cite Allen/Coggan power-meter terminology and TrainingPeaks/WKO public documentation; resource boundaries avoid inventing normalized pace/speed substitutes when normalized power is unavailable.
- Polarization index cites Seiler’s three-zone distribution work and Treff et al. 2017 for the PI equation; resource pins zero-moderate and zero-high boundary handling.
- z-score cites NIST/SEMATECH statistical references plus PRD analyzer sample/missing-day rules; resource pins sample standard deviation, `n >= 7`, and zero-variance handling.

### Step 3 docs/catalog notes

- README public layout was affected and updated to mention analysis formulas under `internal/resources/`.
- Grep for resource-catalog/hash surfaces found only tool-catalog hashing (`internal/mcp/catalog_hash.go`); there is no generated resource catalog/hash to refresh, so the tool catalog was left untouched.

### Step 1 canonical formula draft

| ref | label | canonical equation/method | boundary handling | citation(s) |
| --- | --- | --- | --- | --- |
| `icuvisor://analysis-formulas#hr_drift` | HR drift | For an eligible steady segment, split elapsed moving time into equal first and second halves and report `100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half`. This is a heart-rate-only drift measure under stable external load, not a power-to-HR or pace-to-HR ratio. | Require positive average HR in both halves and a caller/analyzer-selected steady segment; if external load is not stable enough to interpret the result, report insufficient data rather than forcing the formula. | Joe Friel, “Aerobic Endurance and Decoupling,” joefrieltraining.com; TrainingPeaks public education on aerobic decoupling/cardiac drift. |
| `icuvisor://analysis-formulas#pw_hr_decoupling` | Pw:HR decoupling | For cycling power, split an eligible steady segment into equal first and second halves, compute `ratio_first = avg_power_first_half / avg_hr_first_half` and `ratio_second = avg_power_second_half / avg_hr_second_half`, then report `100 * (ratio_first - ratio_second) / ratio_first` so positive values mean less power per heartbeat later in the segment. | Require power and HR in both halves with positive denominators; pace-based siblings should use their own future ref instead of overloading `pw_hr_decoupling`. Do not compute when power is unavailable. | Joe Friel, “Aerobic Endurance and Decoupling,” joefrieltraining.com; TrainingPeaks/WKO public documentation on Pw:HR decoupling. |
| `icuvisor://analysis-formulas#polarization_index` | Polarization index | Use a 3-bucket intensity distribution: low = time in Z1+Z2, moderate = time in Z3, high = time in Z4+. For nonzero moderate/high time, compute `log10((low_share / moderate_share) * (high_share / moderate_share) * 100)` from fractional shares. Analyzer classifications may additionally expose bucket shares for `polarized`, `pyramidal`, or `threshold` labels. | Require total bucketed time > 0. If moderate share is zero, return an explicit saturated/undefined state with bucket shares instead of dividing by zero; if high share is zero, PI is undefined for polarized classification and bucket shares drive a non-polarized label. | Stephen Seiler public work on three-zone endurance intensity distribution; Treff et al., “The Polarization-Index: a simple calculation to distinguish polarized from non-polarized training intensity distributions,” Frontiers in Physiology, 2017. |
| `icuvisor://analysis-formulas#efficiency_factor` | Efficiency factor (EF) | Cycling EF is `normalized_power / avg_hr` for the selected activity or segment. It summarizes normalized output per heartbeat and is distinct from decoupling, which compares EF-like ratios across halves. | Require normalized power and positive average HR. For non-cycling or missing NP, return unavailable unless a later sport-specific ref defines a normalized pace/speed equivalent; do not invent NP. | Allen & Coggan, Training and Racing with a Power Meter; TrainingPeaks/WKO public documentation on Efficiency Factor and Normalized Power. |
| `icuvisor://analysis-formulas#variability_index` | Variability index (VI) | Cycling VI is `normalized_power / avg_power` for the selected activity or segment. Values near 1.00 indicate steadier output; higher values indicate more variable output. | Require normalized power and positive average power. If NP is unavailable or the sport lacks power data, return unavailable instead of substituting raw speed/pace. | Allen & Coggan, Training and Racing with a Power Meter; TrainingPeaks/WKO public documentation on Variability Index and Normalized Power. |
| `icuvisor://analysis-formulas#z_score` | z-score | For baseline comparisons, compute `z = (current_value - baseline_mean) / sample_standard_deviation`, where the baseline mean and standard deviation are calculated over the analyzer-selected baseline window after skipping missing days. | Require at least the analyzer minimum baseline sample (`n >= 7` unless a tool sets a stricter rule). Use sample standard deviation (`n-1`). If standard deviation is zero, report insufficient variance; only present `z=0` when the caller explicitly requests a degenerate equal-value display. | NIST/SEMATECH e-Handbook of Statistical Methods on standard scores and sample standard deviation; PRD §7.2.C analyzer minimum-sample and missing-day rules. |
| 2026-05-20 14:25 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 14:28 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 14:33 | Review R003 | code Step 1: APPROVE |
| 2026-05-20 14:37 | Review R004 | plan Step 2: REVISE |
| 2026-05-20 14:39 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 14:56 | Review R006 | code Step 2: REVISE |
| 2026-05-20 14:59 | Review R007 | code Step 2: APPROVE |
| 2026-05-20 15:01 | Review R008 | plan Step 3: REVISE |
| 2026-05-20 15:02 | Review R009 | plan Step 3: APPROVE |
| 2026-05-20 15:07 | Review R010 | code Step 3: APPROVE |
| 2026-05-20 15:08 | Review R011 | plan Step 4: APPROVE |
| 2026-05-20 15:11 | Review R012 | code Step 4: APPROVE |
| 2026-05-20 15:12 | Review R013 | plan Step 5: REVISE |
| 2026-05-20 15:13 | Review R014 | plan Step 5: APPROVE |
| 2026-05-20 15:17 | Review R015 | code Step 5: APPROVE |
