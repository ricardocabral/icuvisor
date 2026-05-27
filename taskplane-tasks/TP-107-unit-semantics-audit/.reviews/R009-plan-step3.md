# Plan Review — Step 3

Verdict: **Changes requested before implementation**

The Step 3 checklist captures the right outcomes, but the current `STATUS.md` does not yet contain an implementation plan for this step. For a unit/semantics audit, it needs the same level of specificity as Steps 1 and 2 before coding starts: exact tests, exact fields, whether existing coverage is being relied on, and what metadata changes are allowed.

Required plan additions:

1. Name the concrete calories regression coverage. The plan should say whether Step 3 will extend `internal/tools/get_activity_details_test.go`, `internal/tools/get_wellness_data_test.go`, or record existing tests as sufficient. At minimum, it should assert that activity output emits `calories_burned`, does not emit wellness intake keys, and wellness output emits `calories_intake`, does not emit ambiguous/top-level `calories` / `kcalConsumed`, with `_meta.field_semantics` explaining the distinction.
2. Add explicit hydration coverage for a row containing both upstream `hydration` and `hydrationVolume`. Current wellness fixtures appear not to exercise these fields, so the plan should specify either a new wellness fixture or an inline `intervals.Wellness` row, plus assertions that both fields are preserved as distinct outputs and are not collapsed/renamed.
3. Decide the additive metadata shape for hydration before implementation. If current output is ambiguous, prefer row-level `_meta.field_semantics` and/or `_meta.units`; keep terse rows unchanged except for `_meta`. Do not rename `hydration` or `hydrationVolume` unless the task is explicitly amended.
4. Include terse/full assertions. The plan should cover default responses remaining terse, and `include_full: true` preserving raw upstream names only under `full` while keeping the disambiguated top-level keys.
5. Specify targeted verification commands, e.g. `go test ./internal/tools -run 'TestGetActivityDetails|TestGetWellnessData'`, or narrower exact test names after they are added.
6. Add a `STATUS.md` discovery/logging item for any existing calories coverage that Step 3 relies on, so the audit trail explains why only hydration or metadata tests were added.

Once those details are added, the step should be straightforward and low risk.
