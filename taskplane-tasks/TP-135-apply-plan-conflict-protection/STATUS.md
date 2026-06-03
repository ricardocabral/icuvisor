# TP-135: Apply training plan conflict protection for non-workout calendar items — Status
**Current Step:** Step 1: Audit conflict shape and replace policy
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
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
**Status:** 🟨 In Progress

- [x] Inspect `fetchApplyTrainingPlanConflicts` and `replace_existing` deletion behavior.
- [x] Confirm how event category/type/name/date are available from existing event rows and upstream raw fields.
- [x] Define a safe conflict taxonomy in STATUS.md: workout conflicts vs protected annotations/races/unavailable items.
- [x] Run targeted tests: `go test ./internal/tools`.
- [ ] R002: Record the exact conflict-flow code path, duplicate short-circuit, and non-dry-run re-preflight behavior.
- [ ] R002: Make the taxonomy explicit enough to implement, including default protection for non-WORKOUT categories and concrete race/unavailable-like categories.

---

### Step 2: Add protected-conflict behavior and tests
**Status:** ⬜ Not Started

- [ ] Extend conflict output to include enough category/type/name information for LLMs to explain why a day was skipped.
- [ ] Ensure `replace_existing` deletes only intended workout conflicts; protected NOTE, RACE, and UNAVAILABLE-like events are skipped/reported unless a clearly named server-side policy is added.
- [ ] Add tests for mixed calendar days containing a workout plus NOTE/race/unavailable event.
- [ ] Run targeted tests: `go test ./internal/tools`.

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures with exact command output
- [ ] Build passes: `make build`

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

---

## Discoveries

| Date | Step | Finding | Impact |
|------|------|---------|--------|
| 2026-06-03 | Step 1 | Conflict preflight currently treats every same-day event as `existing_event_on_date`; exact matching workout writes become `duplicate_existing_event`, duplicate plan rows become `duplicate_plan_date`, and `replace_existing` deletes every current conflict before creating. | Safe taxonomy: replaceable conflicts are existing WORKOUT events that are not exact duplicates; protected conflicts are exact duplicates, duplicate plan dates, NOTE events, RACE events, and UNAVAILABLE-like calendar annotations/blocks. |
| 2026-06-03 | Step 1 | `intervals.Event` exposes typed `Category`, `Type`, `Name`, `StartDateLocal`, and preserved `Raw`; `eventRow` already falls back to raw `category` and `eventDateOnly` falls back to raw `start_date_local`/`start_date`. | Conflict output can include category/type/name/date fields without new interval client fields. |

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
