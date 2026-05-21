# R008 Code Review — Step 3: Regression tests

**Verdict:** REVISE

## Findings

### P2 — Wellness `field_semantics` presence assertions do not catch missing keys

In `internal/tools/get_wellness_data_test.go:109-113`, the new nutrition metadata check uses:

```go
if semantics[key] == "" { ... }
```

For a missing `map[string]any` entry, `semantics[key]` is `nil`, and `nil == ""` is false. That means the test passes even if `calories_intake`, `carbs_g`, `protein_g`, or `fat_g` is completely absent from `_meta.field_semantics`, which is one of the central Step 3 regression requirements.

Please change this to assert both existence and non-empty string value, e.g.:

```go
got, ok := semantics[key].(string)
if !ok || got == "" { ... }
```

### P3 — Absent-nutrition metadata is not regression-tested

`internal/tools/get_wellness_data_test.go:174-179` verifies that absent nutrition fields are not emitted at the top level, but it does not assert that `_meta.field_semantics` is absent or does not contain nutrition entries for the absence fixture. The Step 3 plan specifically called out that absent nutrition fields should not create a `field_semantics` entry. As written, a regression that emits stale nutrition semantics for `custom_fields.json` would still pass.

Please add an assertion on `absentNutritionRow["_meta"]` that either `field_semantics` is absent or contains none of `calories_intake`, `carbs_g`, `protein_g`, `fat_g`.

### P3 — `get_activity_details` still lacks the requested absent/zero coverage

The new details test in `internal/tools/get_activity_details_test.go:56-79` covers a non-zero activity calorie value and the ambiguous `calories` key, but the approved Step 3 regression matrix requested detail coverage for the same edge cases as the list tool: absent calories should be stripped and present zero calories should be preserved as `calories_burned: 0`. Those cases are only covered for `get_activities`, so a detail-only regression in `activityRow`/detail shaping could slip through.

Please add `get_activity_details` cases for:

- `"calories": 0` => top-level `calories_burned: 0` plus semantics;
- no upstream `calories` => no top-level `calories_burned` and no `field_semantics.calories_burned`.

### P3 — STATUS records R007 as approved, but the review file says revise

`taskplane-tasks/TP-081-nutrition-macros-calorie-labels/.reviews/R007-plan-step3.md:3` has `**Verdict:** REVISE.`, while the execution log in `STATUS.md:137` records `Review R007 | plan Step 3: APPROVE`. Please correct the task bookkeeping before closing the step.

## Verification

Ran:

```sh
go test -count=1 ./internal/tools
```

Result: pass.
