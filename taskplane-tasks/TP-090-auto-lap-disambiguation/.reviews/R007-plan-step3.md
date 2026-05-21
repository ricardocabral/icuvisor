# Plan Review: Step 3 — Propagate to analyzers

## Verdict

Revise. The current Step 3 text in `STATUS.md` is still only the task checklist. That is not specific enough for this step because the interval-consuming analyzer tools are not implemented yet, and the step needs to establish a stable contract for TP-091/TP-093 rather than accidentally inventing analyzer behavior now.

## Review scope

- Read `PROMPT.md` and `STATUS.md` for TP-090.
- Checked the current analyzer skeleton in `internal/analysis/meta.go` and `internal/tools/analyzer_common.go`.
- Confirmed analyzer-family tools such as `analyze_efforts_delta` and `compute_compliance_rate` are not implemented/registered yet; `internal/tools/catalog_test.go` currently treats them as ghosts.
- Checked downstream task prompts: TP-091 and TP-093 both depend on TP-090 for auto-lap propagation/caution behavior.

## Required plan refinements

1. **State explicitly that Step 3 will not implement or register analyzer tools.**
   Since `analyze_efforts_delta`, `compute_compliance_rate`, and related tools do not exist yet, this step should add only shared analyzer support and placeholder tests. Do not create partial public tools, catalog entries, schemas, or docs for the analyzer family in TP-090.

2. **Define the analyzer metadata contract before changing structs.**
   The plan needs to name the exact optional fields future interval-consuming analyzers will emit, and when they are omitted. A safe shape would be along these lines:
   - `_meta.source_tools` includes `get_activity_intervals` when interval rows were used;
   - optional `_meta.interval_source` mirrors `structured_workout`, `device_laps`, or `unknown` when an analyzer has interval-source evidence;
   - optional `_meta.auto_lap_suspected` is emitted only by interval-consuming analyzers that evaluated interval source evidence, preferably via pointer/omitempty semantics so non-interval analyzers do not all gain a misleading `false`.

   If the implementation chooses different field names, the plan should justify them now so TP-091/TP-093 have a contract to follow.

3. **Define the “decline per-interval execution claims” signal.**
   Propagating `auto_lap_suspected=true` alone may not be enough to prevent downstream LLMs from claiming structured-workout execution quality. The plan should define a small helper/API for future analyzers, for example a policy/reason helper that returns “do not make per-interval execution-quality claims” when `AutoLapSuspected` is true. If this is exposed in `_meta`, name the field and values explicitly; if it stays internal, say that the future analyzer result text/summary must use the helper before making interval-level claims.

4. **Place helper code in `internal/analysis`, not in tool-specific response code.**
   Step 2 correctly put interval-source inference in `internal/analysis`. Step 3 should keep analyzer propagation there too, e.g. by extending `AnalyzerMetaInput`/`AnalyzerMeta` with optional interval-source evidence and/or adding an `ApplyIntervalSourceEvidence`/`IntervalExecutionClaimPolicy` helper. Avoid making future analyzers parse shaped `get_activity_intervals` JSON or depend on unexported `internal/tools` adapters.

5. **Preserve existing analyzer meta behavior for non-interval analyzers.**
   Existing analyzer skeleton tests/goldens assert the mandatory `_meta` fields. The plan should state that non-interval analyzer responses remain unchanged except for any intentional golden update. In particular, avoid adding non-omitempty zero-value fields such as `interval_source:""` or `auto_lap_suspected:false` to every analyzer response.

6. **Specify placeholder tests.**
   Because the real analyzer tools are absent, the plan should identify tests such as:
   - a meta/helper test proving interval evidence appends/deduplicates `get_activity_intervals` in `_meta.source_tools`;
   - a shaped analyzer demo test proving interval-source and `auto_lap_suspected=true` propagate when supplied;
   - a policy test proving per-interval execution quality is declined when `AutoLapSuspected` is true and not declined for structured/unknown false cases;
   - a non-interval demo/golden assertion proving existing analyzer meta output does not gain misleading interval fields.

7. **Record the downstream handoff in `STATUS.md`.**
   Since TP-091/TP-093 will implement the actual analyzers, Step 3 should leave a short note in `STATUS.md` with the helper names, field names, and the required behavior future analyzers must use.

## Summary

The direction is correct, but the plan needs one more layer of specificity before code starts. This step should produce a reusable analyzer-meta/policy contract and tests, while explicitly avoiding partial analyzer implementations or broad catalog/docs changes.
