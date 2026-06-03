# TP-134: Calendar write idempotency and duplicate prevention — Status
**Current Step:** Step 4: Documentation & Delivery
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 8
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

### Step 1: Audit write retry and duplicate behavior
**Status:** ✅ Complete

- [x] Inspect `apply_training_plan` and `add_or_update_event` for retry, repeated-call, and concurrent-call behavior.
- [x] Identify whether duplicate detection can be done deterministically from existing event fields before writes.
- [x] Record the chosen idempotency contract and any upstream limitations in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools`.

---

### Step 2: Implement duplicate prevention or explicit duplicate warnings
**Status:** ✅ Complete

- [x] Add tests for repeated `apply_training_plan` calls against the same plan/date range and for duplicate same-day planned events.
- [x] Implement deduplication, stable skip behavior, idempotency keys/metadata, or explicit duplicate warnings using existing upstream fields only.
- [x] Ensure dry-run output makes potential duplicate/conflict outcomes clear before any write.
- [x] Run targeted tests: `go test ./internal/tools`.
- [x] Fix R004: exact duplicate matching must compare the full writable create shape, treating omitted create fields as absent/empty rather than wildcards.
- [x] Fix R005: exact duplicate matching must compare writable target fields only and treat matching actual/completed metrics without targets as conflicts/warnings.

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
| 2026-06-03 | Step 1 | Current client retry policy never retries POST creates, but PUT updates may retry; `apply_training_plan` preflights calendar conflicts once, then creates, so repeated calls skip only after upstream-created events are visible; same-plan duplicate days and concurrent calls can both pass the initial preflight. | Chosen contract: use deterministic same-day duplicate/conflict detection from upstream event fields (date, category, type, name, targets, tags, indoor, description/workout_doc summary) before each create, skip exact matches, warn/skip same-day conflicts for `skip_existing`, and keep unavoidable concurrent race limits explicit because upstream exposes no compare-and-set create or unique idempotency key. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|
| 2026-06-03 | Step 1 plan | APPROVE | Audit plan approved. |
| 2026-06-03 | Step 1 code | APPROVE | Audit findings and targeted tests approved. |
| 2026-06-03 | Step 2 code | REVISE | R004: exact duplicate matching used omitted create fields as wildcards. |
| 2026-06-03 | Step 2 code | REVISE | R005: exact duplicate matching fell back from writable targets to actual metrics. |
| 2026-06-03 | Step 2 code | APPROVE | Duplicate prevention revisions approved. |
| 2026-06-03 | Step 3 plan | APPROVE | Verification plan approved. |
| 2026-06-03 | Step 3 code | APPROVE | Quality gate results approved. |

| 2026-06-03 15:43 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 15:43 | Step 0 started | Preflight |
| 2026-06-03 15:45 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 15:48 | Review R002 | code Step 1: APPROVE |
| 2026-06-03 15:49 | Review R003 | plan Step 2: APPROVE |
| 2026-06-03 15:57 | Review R004 | code Step 2: REVISE |
| 2026-06-03 16:02 | Review R005 | code Step 2: REVISE |
| 2026-06-03 16:06 | Review R006 | code Step 2: APPROVE |
| 2026-06-03 16:07 | Review R007 | plan Step 3: APPROVE |
| 2026-06-03 16:09 | Review R008 | code Step 3: APPROVE |
