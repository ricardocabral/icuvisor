# TP-135: Apply training plan conflict protection for non-workout calendar items — Status
**Current Step:** Step 4: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 8
**Iteration:** 2
**Size:** M
> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.
---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit conflict shape and replace policy
**Status:** ✅ Complete

- [x] Inspect `fetchApplyTrainingPlanConflicts` and `replace_existing` deletion behavior.
- [x] Confirm how event category/type/name/date are available from existing event rows and upstream raw fields.
- [x] Define a safe conflict taxonomy in STATUS.md: workout conflicts vs protected annotations/races/unavailable items.
- [x] Run targeted tests: `go test ./internal/tools`.
- [x] R002: Record the exact conflict-flow code path, duplicate short-circuit, and non-dry-run re-preflight behavior.
- [x] R002: Make the taxonomy explicit enough to implement, including default protection for non-WORKOUT categories and concrete race/unavailable-like categories.

---

### Step 2: Add protected-conflict behavior and tests
**Status:** ✅ Complete

- [x] Extend conflict output to include enough category/type/name information for LLMs to explain why a day was skipped.
- [x] Ensure `replace_existing` deletes only intended workout conflicts; protected NOTE, RACE, and UNAVAILABLE-like events are skipped/reported unless a clearly named server-side policy is added.
- [x] Add tests for mixed calendar days containing a workout plus NOTE/race/unavailable event.
- [x] R004: Implement the protected-day decision so `replace_existing` skips/reports any day with a protected conflict and only deletes pure replaceable WORKOUT conflicts.
- [x] R004: Apply the same classification to initial range preflight and non-dry-run per-day re-preflight before create.
- [x] R004: Ensure exact workout duplicates do not hide other same-day protected rows, preferably with apply-plan-specific conflict building/partitioning.
- [x] R004: Add concrete tests for mixed protected days, pure workout replacement, exact duplicate plus protected row, re-preflight-only protected conflict, conflict detail fields, raw-category fallback, and missing-category protection.
- [x] R004: Update output schema/description and CHANGELOG coverage for the user-visible conflict contract.
- [x] Run targeted tests: `go test ./internal/tools`.

---

### Step 3: Testing & Verification
**Status:** ✅ Complete

- [x] Run FULL test suite: `make test`
- [x] Run lint: `make lint`
- [x] Fix all failures or document pre-existing unrelated failures with exact command output
- [x] Build passes: `make build`

---

### Step 4: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged in STATUS.md

---

## Discoveries

| Date | Step | Finding | Impact |
|------|------|---------|--------|
| 2026-06-03 | Step 1 | Exact conflict flow is `fetchApplyTrainingPlanEvents` (one range `ListEvents`) -> `applyTrainingPlanConflictsForParams` -> `eventCreatePreflightFromEvents`; during non-dry-run, if the initial conflict list is empty, each day is re-read and re-preflighted immediately before create. | Step 2 must apply protected-conflict classification to both the initial range preflight and the per-day re-preflight path. |
| 2026-06-03 | Step 1 | `eventCreatePreflightFromEvents` currently returns immediately on `duplicate_existing_event`, replacing the accumulated conflict list and dropping any other same-day rows. | Step 2 must not let an exact workout duplicate hide protected NOTE/race/unavailable rows on mixed days. |
| 2026-06-03 | Step 1 | `replace_existing` currently deletes every conflict returned by preflight, because `shouldSkipApplyTrainingPlanConflicts` only protects `duplicate_existing_event` and `duplicate_plan_date`. | Step 2 must delete only classified replaceable workout conflicts and skip/report days with any protected conflict. |
| 2026-06-03 | Step 1 | Safe taxonomy for Step 2: replaceable = existing event with category `WORKOUT` and reason `existing_event_on_date`; protected = reason `duplicate_existing_event`, reason `duplicate_plan_date`, any non-`WORKOUT` category, missing/unknown category, documented races `RACE_A`/`RACE_B`/`RACE_C`, annotations `NOTE`/`PLAN`, unavailable-like blocks `HOLIDAY`/`SICK`/`INJURED`, and model/goal markers `SET_EFTP`/`FITNESS_DAYS`/`SEASON_START`/`TARGET`/`SET_FITNESS`. | Conservative default protects custom upstream categories and known non-workout calendar items; no server-side policy exists in this task to delete them. |
| 2026-06-03 | Step 1 | `intervals.Event` exposes typed `Category`, `Type`, `Name`, `StartDateLocal`, and preserved `Raw`; `eventRow` already falls back to raw `category` and `eventDateOnly` falls back to raw `start_date_local`/`start_date`. | Conflict output can include category/type/name/date fields without new interval client fields. |
| 2026-06-03 | Step 4 | `apply_training_plan` now has a user-visible protected-conflict contract, so PRD §7.2.C and the v0.3 dogfood apply prompts were affected in addition to the changelog. | Delivery docs now describe category/type/name/date conflict details and protected-note/race/unavailable `replace_existing` expectations. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 17:12 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 17:12 | Step 0 started | Preflight |
| 2026-06-03 17:14 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 17:16 | Review R002 | code Step 1: REVISE | Missing exact conflict-flow details and concrete protected taxonomy; revision checkboxes added. |
| 2026-06-03 17:18 | Review R003 | code Step 1: APPROVE | Conflict-flow and taxonomy clarifications approved. |
| 2026-06-03 17:19 | Review R004 | plan Step 2: REVISE | Expanded Step 2 plan for protected-day decisions, both preflight paths, duplicate short-circuit, concrete tests, and docs. |
| 2026-06-03 17:20 | Review R005 | plan Step 2: APPROVE | Expanded Step 2 plan approved. |
| 2026-06-03 17:25 | Review R006 | code Step 2: APPROVE | Protected conflict implementation approved. |
| 2026-06-03 17:26 | Review R007 | plan Step 3: APPROVE | Verification plan approved. |
| 2026-06-03 17:29 | Review R008 | code Step 3: APPROVE | Verification results approved. |
| 2026-06-03 17:18 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 17:20 | Review R004 | plan Step 2: UNKNOWN |
| 2026-06-03 17:21 | Review R005 | plan Step 2: APPROVE |
| 2026-06-03 17:28 | Review R006 | code Step 2: APPROVE |
| 2026-06-03 17:28 | Review R007 | plan Step 3: APPROVE |
| 2026-06-03 17:31 | Review R008 | code Step 3: APPROVE |

| 2026-06-03 17:31 | Worker iter 1 | done in 1184s, tools: 116 |
| 2026-06-03 17:31 | Step 4 started | Documentation & Delivery |