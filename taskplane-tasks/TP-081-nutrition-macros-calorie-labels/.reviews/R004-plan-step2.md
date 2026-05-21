# R004 Plan Review — Step 2: Shape activity and wellness responses

**Verdict:** REVISE before coding Step 2.

I did not find a separate detailed Step 2 implementation plan beyond the Step 2 checklist in `PROMPT.md` / `STATUS.md` plus the Step 1 notes. The direction is correct, but the plan needs a few concrete shaping decisions to avoid leaking ambiguous upstream nutrition names or adding unclear `_meta` shapes.

## Findings

1. **The plan must explicitly remove/translate legacy wellness nutrition keys in terse rows.**  
   `wellnessRow` starts from `cloneJSONMap(row.Raw)` (`internal/tools/get_wellness_data.go:109`) and currently sets the upstream names back onto the top-level response (`kcalConsumed`, `carbohydrates`, `protein`, `fatTotal` at lines 130 and 159-161). If Step 2 only adds `calories_intake`, `carbs_g`, `protein_g`, and `fat_g`, terse responses will emit both the new disambiguated keys and the old ambiguous upstream keys. That undermines the task mission (“should not guess hidden field names” / “do not overload `calories` as an ambiguous key”). The plan should say whether Step 2 will:
   - delete `kcalConsumed`, `carbohydrates`, `protein`, and `fatTotal` from the top-level shaped wellness row before returning; and
   - keep raw upstream names only under `full` when `include_full: true`.

2. **Define the exact `_meta` contract before implementation.**  
   “Add `_meta` labels where semantics could be confused” is too open-ended. The plan should name the field and location, e.g. a stable `_meta.field_labels` / `_meta.field_semantics` map on the wrapper or per-row metadata. It should include at least:
   - `calories_burned`: active/exercise calories from activity `calories`;
   - `calories_intake`: consumed kcal from wellness `kcalConsumed`;
   - `carbs_g`, `protein_g`, `fat_g`: grams from wellness macro fields.

   Also account for the existing response shaper: top-level wrapper `_meta` will be merged with common metadata, while row-level wellness `_meta` may already contain provenance. The plan should state how the new labels coexist with existing provenance and scale metadata without overwriting them.

3. **Preserve null-stripping while distinguishing absent from present zero where practical.**  
   Wellness macros can use the existing pointer-aware `setWellnessField` pattern, which naturally omits absent values and preserves present zero values. Activity `calories_burned` is currently an `int` with `omitempty` (`internal/tools/get_activities.go:78`) populated via `intValue(activity.Calories)` (`internal/tools/get_activities_row.go:40`), so absent calories and present `0` both disappear. If this task is tightening calorie semantics, the plan should either convert `CaloriesBurned` to `*int` or explicitly document why existing zero elision is acceptable for activity calories.

4. **Activity shape is mostly already correct; avoid expanding scope.**  
   Step 1 established that no activity macro fields are supported by current fixtures/schemas. The Step 2 plan should say that activity changes are limited to preserving/clarifying `calories_burned` labels and not adding activity macro fields or `calories_total`.

## Required plan amendments

Before coding Step 2, add concise sub-bullets to `STATUS.md` (or a small implementation note) covering:

- exact public key mapping:
  - activity `calories` -> `calories_burned`;
  - wellness `kcalConsumed` -> `calories_intake`;
  - wellness `carbohydrates` -> `carbs_g`;
  - wellness `protein` -> `protein_g`;
  - wellness `fatTotal` -> `fat_g`;
  - no `calories_total` and no activity macros unless new fixture evidence appears;
- whether old wellness upstream names are removed from top-level terse rows and retained only in `full` for `include_full`;
- the exact `_meta` key/location for nutrition/calorie labels and how it merges with existing wellness provenance metadata;
- the intended handling for present-zero values, especially `calories_burned`.

With those amendments, the Step 2 plan should be safe to implement and can remain focused on response shaping rather than broader decoder or documentation work.
