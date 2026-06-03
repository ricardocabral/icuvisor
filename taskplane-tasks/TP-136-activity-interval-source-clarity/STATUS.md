# TP-136: Activity interval-source clarity in details and routing — Status
**Current Step:** Step 4: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 3
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

### Step 1: Audit current interval-source exposure
**Status:** ✅ Complete

- [x] Inspect `get_activity_details`, `get_activity_intervals`, interval-source tests, and tool descriptions.
- [x] Decide whether to expose interval-source metadata on `get_activity_details`, strengthen descriptions only, or both.
- [x] Record the decision and any upstream limitation in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/analysis`.

---

### Step 2: Implement clarity and regression coverage
**Status:** ✅ Complete

- [x] Add or update tests showing device-lap/auto-lap/structured-workout source metadata is surfaced or routed correctly.
- [x] Update tool descriptions/schema snapshots as needed so assistants know to call `get_activity_intervals` before analyzing laps/reps.
- [x] Ensure terse defaults stay compact and `include_full` remains the raw-payload opt-in.
- [x] Run targeted tests: `go test ./internal/tools ./internal/analysis`.

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
| 2026-06-03 | Step 1 | Decision: strengthen routing/descriptions and preserve `get_activity_intervals` as the source-metadata surface instead of adding interval-source metadata to `get_activity_details`. `get_activity_details` only fetches the activity detail payload, while reliable interval-source classification requires the intervals payload (`icu_intervals`/`icu_groups`). | Assistants must call `get_activity_intervals` before making lap/rep/interval-execution claims; details can warn about that without issuing an extra upstream intervals request or implying unavailable metadata. |
| 2026-06-03 | Step 4 | Reviewed `docs/prd/PRD-icuvisor.md` and `docs/dogfood/v0.2-prompts.md`; existing PRD contract already lists activity detail/interval tool split and dogfood prompt T-08 already targets `get_activity_intervals` for intervals/laps. | No material response-field contract or dogfood prompt update needed beyond tool-description/schema-description wording and CHANGELOG entry. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 16:41 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:41 | Step 0 started | Preflight |
| 2026-06-03 16:43 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 16:47 | Review R002 | plan Step 2: APPROVE |
| 2026-06-03 16:50 | Review R003 | plan Step 3: APPROVE |
