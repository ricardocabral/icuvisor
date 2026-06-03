# TP-138: Weekly report timezone and stale-data guardrails — Status
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

### Step 1: Audit weekly/date-window safeguards
**Status:** ⬜ Not Started

- [ ] Inspect prompt text and tests for weekly review, plan-health review, wellness reads, and `_meta.as_of`.
- [ ] Identify whether prompts explicitly forbid including wellness after the requested report window.
- [ ] Record any missing stale-date guardrails in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/prompts ./internal/tools`.

---

### Step 2: Add prompt and regression coverage
**Status:** ⬜ Not Started

- [ ] Update weekly/plan-health prompt guidance to anchor all report windows in athlete-local dates and treat current-day `_meta.as_of` as partial-day context only.
- [ ] Add or strengthen golden tests so stale/current-day caveats are preserved in prompt output.
- [ ] Add targeted tool tests only if an existing `_meta.as_of` edge case is uncovered.
- [ ] Run targeted tests: `go test ./internal/prompts ./internal/tools`.

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
