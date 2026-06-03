# TP-142: Readiness provenance labels and recovery wording guardrails — Status

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

### Step 1: Audit readiness/recovery wording
**Status:** ⬜ Not Started

- [ ] Inspect wellness provenance shaping for Garmin, Oura, Polar, WHOOP, and unknown readiness sources.
- [ ] Inspect recovery/weekly prompts for wording that could collapse provider-native readiness into generic recovery.
- [ ] Record missing labels or ambiguous terms in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`.

---

### Step 2: Add provenance and prompt regressions
**Status:** ⬜ Not Started

- [ ] Add or strengthen tests that `_meta.provenance.readiness.native_scale` is provider-specific and visible when readiness is present.
- [ ] Update prompt wording/golden tests so assistants cite provider/source and do not invent a readiness score when missing or stale.
- [ ] Ensure terse defaults remain compact and null stripping does not remove required provenance.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`.

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
