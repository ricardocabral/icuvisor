# Plan Review — Step 1

Result: **Approved with required clarifications before/while executing Step 1**

The step is correctly scoped as an audit-only pass and the targeted baseline test command (`go test ./internal/tools`) currently passes. I do not see a reason to block execution, but the Step 1 notes should be more precise so Step 2 is built on the real conflict flow.

## Required clarifications

1. **Use actual function names/paths in the audit.** The prompt mentions `fetchApplyTrainingPlanConflicts`, but the current code path is `fetchApplyTrainingPlanEvents` → `applyTrainingPlanConflictsForParams` → shared `eventCreatePreflightFromEvents`, with deletion in `applyTrainingPlan`'s `replace_existing` branch. Record this in `STATUS.md` so the implementation doesn't chase a non-existent helper.

2. **Audit the duplicate short-circuit.** `eventCreatePreflightFromEvents` returns immediately on `duplicate_existing_event`, discarding any other same-day rows. That matters for “make conflicts explicit” when a day has an exact workout duplicate plus a protected NOTE/race/unavailable-like item.

3. **Define taxonomy with concrete category rules.** The taxonomy should state which conflicts are replaceable workout conflicts and which are protected. At minimum include `NOTE`, `RACE`, `RACE_*`, and unavailable-like categories from `internal/intervals/event_categories.go` such as `HOLIDAY`, `SICK`, and `INJURED`; consider whether all non-`WORKOUT` calendar markers should be protected by default.

4. **Include re-preflight behavior in the audit.** Non-dry-run conflict-free rows are re-listed for the same day before create. The replacement policy must be evaluated for both the initial date-range conflict list and this per-day preflight path.

