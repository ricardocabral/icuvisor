# TP-136: Activity interval-source clarity in details and routing — Status
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

### Step 1: Audit current interval-source exposure
**Status:** ⬜ Not Started

- [ ] Inspect `get_activity_details`, `get_activity_intervals`, interval-source tests, and tool descriptions.
- [ ] Decide whether to expose interval-source metadata on `get_activity_details`, strengthen descriptions only, or both.
- [ ] Record the decision and any upstream limitation in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/analysis`.

---

### Step 2: Implement clarity and regression coverage
**Status:** ⬜ Not Started

- [ ] Add or update tests showing device-lap/auto-lap/structured-workout source metadata is surfaced or routed correctly.
- [ ] Update tool descriptions/schema snapshots as needed so assistants know to call `get_activity_intervals` before analyzing laps/reps.
- [ ] Ensure terse defaults stay compact and `include_full` remains the raw-payload opt-in.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/analysis`.

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
