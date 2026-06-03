# Plan Review — Step 2

Result: **Request changes**

The Step 2 scope matches the task, but the executable plan is still too high-level for the failure modes discovered in Step 1. Before coding, expand/clarify Step 2 in `STATUS.md` so the implementation cannot miss these cases.

## Required plan clarifications

1. **State the protected-day decision explicitly.** For `replace_existing`, a day with any protected conflict (`duplicate_existing_event`, `duplicate_plan_date`, non-`WORKOUT`, missing/unknown category, races/notes/unavailable-like markers) should be skipped/reported and should not delete even replaceable workout conflicts on that mixed day. Only pure replaceable `WORKOUT`/`existing_event_on_date` conflicts should be deleted.

2. **Cover both conflict paths.** Add an outcome/check that the same classification and skip/delete partition is applied to both the initial range preflight and the non-dry-run per-day re-preflight immediately before create.

3. **Address the duplicate short-circuit directly.** The plan should say how exact workout duplicates stop hiding other same-day protected rows. Prefer an apply-plan-specific conflict builder/partition if changing shared `eventCreatePreflightFromEvents` would alter `add_or_update_event` duplicate semantics.

4. **Make the tests concrete.** Include tests for: mixed `WORKOUT` + `NOTE`/`RACE_A`/`HOLIDAY` (or sick/injured) under full `replace_existing` with no deletes/writes; pure workout conflict still deletes/recreates in full mode; exact duplicate plus protected row reports both; and a re-preflight-only protected conflict is skipped without mutation. Also assert new conflict fields (`category`, `type`, `name`, date) and raw-category fallback/missing-category protection.

5. **Include contract/doc update in the step.** Since conflict output changes, update the output schema/description and ensure `CHANGELOG.md` is covered (either in Step 2 or explicitly deferred to Step 4).

## Verification

Not run; this was a plan review.
