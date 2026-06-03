# TP-134: Calendar write idempotency and duplicate prevention — Status
**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 0
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
**Status:** ⬜ Not Started

- [ ] Inspect `apply_training_plan` and `add_or_update_event` for retry, repeated-call, and concurrent-call behavior.
- [ ] Identify whether duplicate detection can be done deterministically from existing event fields before writes.
- [ ] Record the chosen idempotency contract and any upstream limitations in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`.

---

### Step 2: Implement duplicate prevention or explicit duplicate warnings
**Status:** ⬜ Not Started

- [ ] Add tests for repeated `apply_training_plan` calls against the same plan/date range and for duplicate same-day planned events.
- [ ] Implement deduplication, stable skip behavior, idempotency keys/metadata, or explicit duplicate warnings using existing upstream fields only.
- [ ] Ensure dry-run output makes potential duplicate/conflict outcomes clear before any write.
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

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 15:43 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 15:43 | Step 0 started | Preflight |