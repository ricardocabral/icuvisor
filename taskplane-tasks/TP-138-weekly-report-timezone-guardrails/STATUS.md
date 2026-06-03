# TP-138: Weekly report timezone and stale-data guardrails — Status
**Current Step:** Step 4: Documentation & Delivery
**Status:** 🟡 In Progress
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

### Step 1: Audit weekly/date-window safeguards
**Status:** ✅ Complete

- [x] Inspect prompt text and tests for weekly review, plan-health review, wellness reads, and `_meta.as_of`.
- [x] Identify whether prompts explicitly forbid including wellness after the requested report window.
- [x] Record any missing stale-date guardrails in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/prompts ./internal/tools`.

---

### Step 2: Add prompt and regression coverage
**Status:** ✅ Complete

- [x] Update weekly/plan-health prompt guidance to anchor all report windows in athlete-local dates and treat current-day `_meta.as_of` as partial-day context only.
- [x] Add or strengthen golden tests so stale/current-day caveats are preserved in prompt output.
- [x] Add targeted tool tests only if an existing `_meta.as_of` edge case is uncovered.
- [x] Run targeted tests: `go test ./internal/prompts ./internal/tools`.

---

### Step 3: Testing & Verification
**Status:** ✅ Complete

- [x] Run FULL test suite: `make test`
- [x] Run lint: `make lint`
- [x] Fix all failures or document pre-existing unrelated failures with exact command output
- [x] Build passes: `make build`

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
| 2026-06-03 | Step 1 | Weekly and plan-health prompts already require athlete-local timezone/date conversion and `_meta.stale` caveats, and tool tests cover current-day `_meta.as_of` inclusion/exclusion. They do not explicitly forbid mixing wellness rows after the requested report window or treating current-day `_meta.as_of` as partial-day context only. | Step 2 should harden prompt text and golden/regression assertions; no tool `_meta.as_of` edge case found so far. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 16:23 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:23 | Step 0 started | Preflight |
| 2026-06-03 16:24 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 16:26 | Review R002 | plan Step 2: APPROVE |
| 2026-06-03 16:29 | Review R003 | plan Step 3: APPROVE |
