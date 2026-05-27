# Plan Review: Step 2 — Investigate and implement activity tag handling if supported

**Verdict: Needs revision before implementation**

The Step 2 checklist in `STATUS.md` is still too high-level for the activity path. It correctly calls out investigation first, but it does not define the implementation decision, extraction semantics, or shared-row impact. Activity responses reuse `activityRow` in more places than just `get_activities` and `get_activity_details`, so the plan should be tightened before coding.

## Required clarifications

1. **Record the investigation result explicitly.**
   - Current `intervals.Activity` preserves `Raw` but has no typed `Tags` field, and the checked-in activity fixtures do not appear to contain `tags`.
   - The plan should state whether Step 2 will still support `Raw["tags"]` when upstream provides a valid array, or will intentionally document the current fixture/API gap without adding a row field.

2. **Use tolerant raw extraction if implementing.**
   - Do not add a plain `[]string json:"tags"` field to `intervals.Activity`; that risks decode failures or losing missing/null/empty distinctions.
   - Match the event semantics: emit only an upstream JSON string array, preserve order, preserve an explicit empty array, copy the slice, and omit missing/null/non-array/mixed values without guessing.
   - This likely means `Tags *[]string json:"tags,omitempty"` on `getActivitiesRow` populated from `activity.Raw["tags"]`.

3. **Acknowledge all shared `activityRow` callers.**
   - Updating `activityRow` affects `get_activities`, `get_activity_details`, `get_today` completed activities, and delete activity response shaping via `delete_common.go`.
   - The plan should either accept this inherited shared-row behavior and add/adjust coverage where appropriate, or describe how it will limit the field to only the named read tools.
   - Also decide whether valid tags should be retained on Strava-blocked/unavailable rows; if yes, initialize them before the early return.

4. **Include schema/description updates if the field is added.**
   - `getActivitiesOutputSchema()` and `activityReadOutputSchema()` should mention activity `tags` once it becomes a documented terse field. `get_today` may also need a description tweak if completed activities inherit tags.

5. **Define targeted tests now, not only in Step 3.**
   - Add activity tests for tags present/order, explicit empty array emitted, missing/null/non-array/mixed omitted, and `include_full` preserving the raw payload.
   - Cover both `get_activities` and `get_activity_details` at minimum; add a `get_today` completed-activity assertion if shared-row inheritance is accepted.
   - If deciding not to implement activity tags because upstream support is not evidenced, add a regression/discovery test proving valid current fixtures do not emit guessed tags and log the discovery in `STATUS.md`.

With these details added, Step 2 should be safe and consistent with the event implementation from Step 1.
