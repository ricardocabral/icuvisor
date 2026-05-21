# R005 Plan Review — Step 3: Apply promotion if evidence exists

**Verdict:** REVISE

The Step 3 plan is too under-specified for the current repository state. `docs/kr5-benchmark.md` already contains positive TP-098 core-promotion evidence for exactly `analyze_trend`, `compute_zone_time`, and `compute_baseline`, and `STATUS.md` records those three as eligible. Step 3 should therefore choose the promotion path explicitly rather than leave the absent/negative branch as an open decision.

## Required revisions

1. **State the selected promotion outcome.**
   - Evidence exists and is positive in `docs/kr5-benchmark.md`.
   - Promote only:
     - `analyze_trend`
     - `compute_zone_time`
     - `compute_baseline`
   - Keep every other analyzer-family tool in `full`.

2. **Plan the code edits explicitly.**
   - Change the three eligible tool constructors from `fullTool(...)` to `coreTool(...)`:
     - `internal/tools/analyze_trend.go`
     - `internal/tools/compute_zone_time.go`
     - `internal/tools/compute_baseline.go`
   - Do not change unrelated analyzer tools such as `analyze_correlation`, `compute_load_balance`, `compute_compliance_rate`, `get_activity_histogram`, or `get_fitness_projection`.

3. **Update the Step 2 tests that will otherwise conflict with promotion.**
   - `TestRegisteredToolTierMembership` currently expects the three candidates to be `full`; it must expect them as `core` after Step 3.
   - `TestAnalyzerFamilyDefaultsToFullToolset` currently requires every analyzer-family tool to be `full`; revise it to assert non-candidate analyzer-family tools remain `full`, or rename/split it so it reflects the post-promotion policy.
   - Strengthen or update the promotion-candidate test so the allowed candidate set is still exactly those three and, now that promotion is applied, those candidates are actually `core`.

4. **Account for generated catalog/docs updates.**
   - Regenerate `web/data/tools.json` with `make docs-tools` so the three promoted tools show `tier: "core"`.
   - Check `web/content/reference/tools.md`; it likely does not need manual tier edits because it renders from `web/data/tools.json`, but record that review in `STATUS.md`.
   - Update `CHANGELOG.md` under `[Unreleased]` because default `core` tool behavior changes.

5. **Clarify advanced-capabilities expectations.**
   - After promotion, the three core analyzers should no longer appear as hidden/full-only capabilities in `icuvisor_list_advanced_capabilities` for core mode.
   - The remaining full-only analyzer-family tools should still be advertised with clear activation hints.

With those revisions, the Step 3 plan will align with the task requirement and avoid immediately failing the tier tests added in Step 2.
