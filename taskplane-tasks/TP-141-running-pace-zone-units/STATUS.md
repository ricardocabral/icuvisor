# TP-141: Running pace-zone unit and label audit — Status

**Current Step:** Step 1: Audit run pace read/write coverage
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit run pace read/write coverage
**Status:** 🟨 In Progress

- [ ] Inspect athlete-profile and sport-settings tests for threshold pace, pace units, and pace-zone names.
- [ ] Confirm tests cover both `seconds_per_km` and `seconds_per_mile` inputs and upstream pace unit output.
- [ ] Record any ambiguous LLM-facing wording or missing scale/unit metadata in STATUS.md.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/units`.

---

### Step 2: Add pace-zone regressions and wording fixes
**Status:** ⬜ Not Started

- [ ] Add missing tests for Run threshold pace conversion and pace zone boundary/name round trips.
- [ ] Update schema descriptions or response labels if they could be misread as speed rather than pace seconds per distance.
- [ ] Ensure zone overwrite behavior remains gated by `ICUVISOR_DELETE_MODE=full` where applicable.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/units`.

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

| 2026-06-03 16:52 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:52 | Step 0 started | Preflight |
| 2026-06-03 16:52 | Step 0 complete | Preflight passed: required files exist; go list succeeded for target packages; clean-room constraint acknowledged. |
| 2026-06-03 16:52 | Step 1 started | Audit run pace read/write coverage |