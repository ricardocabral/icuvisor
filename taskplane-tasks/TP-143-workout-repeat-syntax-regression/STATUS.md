# TP-143: Workout repeat header syntax regression — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit repeat serialization and validation
**Status:** ⬜ Not Started

- [ ] Inspect WorkoutDoc serialize/parse/validate tests for repeat headers with and without descriptions.
- [ ] Confirm write-tool tests exercise repeat blocks through validation and event/workout serialization.
- [ ] Record any missing edge cases in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/workoutdoc ./internal/tools`.

---

### Step 2: Add repeat syntax regressions
**Status:** ⬜ Not Started

- [ ] Add tests asserting repeat headers serialize as `3x` or `<description> 3x` without a leading dash.
- [ ] Add parse/validate coverage rejecting or warning on malformed `-3 x` / `- 3x` style lines when appropriate.
- [ ] Add at least one write-tool regression showing a repeat workout_doc produces canonical DSL.
- [ ] Run targeted tests: `go test ./internal/workoutdoc ./internal/tools`.

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
