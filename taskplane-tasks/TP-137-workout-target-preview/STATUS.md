# TP-137: Resolved workout target previews for planned workouts — Status
**Current Step:** Step 1: Design compact resolved-target shape
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

### Step 1: Design compact resolved-target shape
**Status:** 🟨 In Progress

- [ ] Audit event/workout read rows and `workout_doc_summary` to find the least-bloated place for target previews.
- [ ] Use athlete profile thresholds/units only when already available or cheaply fetchable; avoid extra heavy calls or raw payload expansion.
- [ ] Record unsupported target cases and null/omission rules in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/workoutdoc`.

---

### Step 2: Implement target previews and tests
**Status:** ⬜ Not Started

- [ ] Add tests for `% FTP` planned workout targets resolving to watts from profile FTP.
- [ ] Add tests or explicit omissions for HR threshold, pace threshold, missing profile threshold, and non-numeric/text targets.
- [ ] Implement compact preview fields while preserving terse-by-default and `include_full` behavior.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/workoutdoc`.

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