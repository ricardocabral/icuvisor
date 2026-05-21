# Plan Review R001 — Step 1: Draft canonical formulas

**Verdict:** Changes requested

The current Step 1 plan captures the right high-level outcomes, but it is too thin for a formula registry that later analyzer tools will treat as a stable contract. Before implementation starts, Step 1 should produce a concrete formula table/draft with exact refs, equations, source citations, and boundary decisions.

## Blocking concerns

1. **No concrete Step 1 deliverable is defined.**
   The checklist says to draft definitions, cite sources, and assign anchors, but it does not say where those decisions are recorded before Step 2. Add an explicit output for this step, e.g. a table in `STATUS.md` Notes/Discoveries or a draft section that Step 2 will translate into `internal/resources/analysis_formulas.go`.

2. **Formula refs need to be specified now.**
   Analyzer responses will later link via `_meta.formula_ref`, so Step 1 should choose the stable ref format, not leave it implicit. Prefer full URI fragments such as:
   - `icuvisor://analysis-formulas#hr_drift`
   - `icuvisor://analysis-formulas#pw_hr_decoupling`
   - `icuvisor://analysis-formulas#polarization_index`
   - `icuvisor://analysis-formulas#efficiency_factor`
   - `icuvisor://analysis-formulas#variability_index`
   - `icuvisor://analysis-formulas#z_score`

3. **Definitions must include exact math and edge-case decisions, not only prose.**
   The PRD says definitions are locked and definition drift is breaking. The Step 1 plan should require each entry to include the canonical equation/method string that future golden tests can pin. At minimum capture:
   - split method for HR drift and Pw:HR decoupling, including first-half vs second-half semantics;
   - whether EF is `normalized_power / avg_hr` and how non-cycling equivalents are represented or deferred;
   - whether VI is `normalized_power / avg_power` and what happens when normalized power is unavailable;
   - the exact polarization-index formula/classification basis from zone buckets;
   - whether z-score uses sample or population standard deviation, baseline window definition, and zero-variance handling.

4. **HR drift vs Pw:HR decoupling must be disambiguated.**
   These are often conflated. The plan should require separate canonical wording explaining when each ref applies and whether HR drift is HR-only drift under stable external load or the same ratio-based aerobic decoupling calculation with a different output stream.

5. **Citation/source policy needs to be explicit enough to preserve clean-room constraints.**
   The task requires accepted public sources and warns against copying. Add a source-selection rule: cite bibliographic/public docs/papers with URLs or publication details, summarize in original words, do not quote long text, and do not derive wording from GPL/copyleft implementations. For sources named by the PRD (Friel / Seiler / Coggan), record exactly which public article, paper, book, or documentation page supports each formula.

## Suggested Step 1 acceptance shape

Add a table with columns like:

| ref | label | canonical equation/method | boundary handling | citation(s) |
| --- | --- | --- | --- | --- |

This will make Step 2 straightforward and gives TP-097 a clear target for definition-drift guard tests.

## Non-blocking notes

- Keep each rendered resource entry to one paragraph as required, but the Step 1 draft can be slightly more structured to capture implementation decisions.
- If a formula is dependent on upstream Intervals.icu precomputed values, state that explicitly to stay aligned with the PRD's “no competing physiology models” rule.
- Record final source decisions in `STATUS.md`; waiting until Step 4 is acceptable for final notes, but the initial decisions should exist before code is written.
