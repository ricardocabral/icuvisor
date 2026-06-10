# Plan Review: Step 2 — Propagate source evidence to tool/analyzer responses

**Verdict:** Approved

The Step 2 plan is appropriately scoped and follows the Step 1 classifier behavior. It covers analyzer metadata, `get_activity_intervals` response shaping/tests, and targeted `internal/tools` + `internal/analysis` verification.

## Notes for implementation

- Add explicit analyzer coverage for the new `manual_added` and `mixed` `interval_source` values. Keep a regression that `device_laps` still emits `auto_lap_suspected: true`, and verify manual/mixed evidence emits `auto_lap_suspected: false` rather than dropping the field when interval evidence is attached.
- Tool response tests should build interval rows with realistic raw `group_id` evidence and **no `icu_groups`** for manual/mixed cases, because `icu_groups` remains a structured-workout signal. Include an all-ungrouped/raw-evidence case for `manual_added` and a grouped-plus-ungrouped case for `mixed`.
- Preserve the missing-evidence guard in tool-level tests too: nil/empty raw interval maps should not become `manual_added` just because the response is terse.
- Update user-facing enum text in `getActivityIntervalsDescription` and `activityReadOutputSchema` from the old `structured_workout/device_laps/unknown` wording to include `manual_added` and `mixed`; extend the existing description test to assert those tokens.
- `internal/tools/schema_snapshot/get_activity_intervals.json` appears to snapshot the input schema only, so refresh it only if the actual input schema changes. If tool descriptions/catalog hashes are guarded elsewhere, run the relevant snapshot/catalog update command instead.

With those details observed, implementation can proceed.
