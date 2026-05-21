# Review R005 — Plan Review for Step 2

**Verdict:** APPROVE

The expanded Step 2 plan is sufficient to proceed. It now ties the golden-file work back to the completed Step 1 inventory and covers the required surfaces: formula resource stability, shared analyzer fixtures, `internal/analysis` computation guards, `internal/tools` response/meta guards, and loud drift failures.

## Notes / guardrails for implementation

- Keep the shared `testdata/analysis/` fixture schema narrow and deterministic: inputs, expected rounded outputs, statuses, boundary states, and expected `formula_ref` values. Avoid storing large raw stream payloads beyond the minimum needed for the formula cases.
- Reuse the existing golden-test conventions where practical (`UPDATE_ANALYZER_GOLDENS` style update flow and canonical indented JSON), but make mismatch errors include the exact update command for the new fixture set.
- For EF and VI, preserve the Step 1 distinction:
  - EF is currently resource-only/no local analyzer output. Do not invent a computed EF in this step; guard that status explicitly and avoid conflating EF with existing intensity factor (`if` / `icu_intensity`).
  - VI is upstream-derived through `get_extended_metrics`; pin the upstream-mapped output and the missing-field omission behavior rather than locally recomputing `normalized_power / avg_power`.
- For tool-level checks, assert both the headline result fields and `_meta.formula_ref` where the tool emits one. This is especially important for `compute_activity_segment_stats`, `compute_zone_time`/`compute_load_balance`, `compute_baseline`, and the `analyze_trend` z-score path.
- Do not update `CHANGELOG.md` in this step unless the implementation changes user-visible behavior; pure test/guard additions can stay out of the changelog per the task instructions.
