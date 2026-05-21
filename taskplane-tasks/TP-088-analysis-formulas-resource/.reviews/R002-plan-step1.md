# Plan Review R002 — Step 1: Draft canonical formulas

**Verdict:** Approved

The revised Step 1 plan is now concrete enough to guide the drafting work before implementation. It adds the key pre-code deliverable requested in R001: a formula draft table in `STATUS.md` with `ref`, label, canonical equation/method, boundary handling, and citations. It also explicitly calls out the major definition-drift risk areas: split semantics, missing NP/power inputs, zone buckets, sample standard deviation, baseline windows, and zero variance.

## What is now covered

- A tangible Step 1 output location is defined (`STATUS.md` Notes), so Step 2 has a stable source to translate into the resource.
- The formula entries must capture exact equations/methods, not just prose summaries.
- Stable refs/anchors are part of the planned output.
- HR drift / Pw:HR-related boundary decisions are called out through split semantics and missing-input handling.
- The source-selection rule is sufficiently explicit for clean-room work: public/bibliographic sources, original summaries, no long copied text, and no GPL/copyleft implementation-derived wording.

## Non-blocking follow-up

When executing Step 1, prefer full resource-fragment refs so later `_meta.formula_ref` values can be copied verbatim, for example:

- `icuvisor://analysis-formulas#hr_drift`
- `icuvisor://analysis-formulas#pw_hr_decoupling`
- `icuvisor://analysis-formulas#polarization_index`
- `icuvisor://analysis-formulas#efficiency_factor`
- `icuvisor://analysis-formulas#variability_index`
- `icuvisor://analysis-formulas#z_score`

Also make sure the table distinguishes HR drift from Pw:HR decoupling in the wording, not only in the equations, because those labels are commonly conflated and analyzer responses will depend on the distinction.
