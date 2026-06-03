# TP-135: Apply training plan conflict protection for non-workout calendar items — Status
**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M
> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.
---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit conflict shape and replace policy
**Status:** ⬜ Not Started

- [ ] Inspect `fetchApplyTrainingPlanConflicts` and `replace_existing` deletion behavior.
- [ ] Confirm how event category/type/name/date are available from existing event rows and upstream raw fields.
- [ ] Define a safe conflict taxonomy in STATUS.md: workout conflicts vs protected annotations/races/unavailable items.
- [ ] Run targeted tests: `go test ./internal/tools`.

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

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|
