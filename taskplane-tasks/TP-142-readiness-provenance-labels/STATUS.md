# TP-142: Readiness provenance labels and recovery wording guardrails — Status

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

### Step 1: Audit readiness/recovery wording
**Status:** ✅ Complete

- [x] Inspect wellness provenance shaping for Garmin, Oura, Polar, WHOOP, and unknown readiness sources.
- [x] Inspect recovery/weekly prompts for wording that could collapse provider-native readiness into generic recovery.
- [x] Record missing labels or ambiguous terms in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`.

---

### Step 2: Add provenance and prompt regressions
**Status:** ✅ Complete

- [x] Add or strengthen tests that `_meta.provenance.readiness.native_scale` is provider-specific and visible when readiness is present.
- [x] Update prompt wording/golden tests so assistants cite provider/source and do not invent a readiness score when missing or stale.
- [x] Ensure terse defaults remain compact and null stripping does not remove required provenance.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`.

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
| 2026-06-03 | Step 1 | Existing wellness provenance labels Garmin Body Battery, Polar nightly_recharge_status/ans_charge, WHOOP recovery, and unknown sources; Oura readiness label exists in code but lacks a dedicated readiness regression fixture/test. | Step 2 should add provider-specific readiness regression coverage, especially Oura and generic unknown readiness. |
| 2026-06-03 | Step 1 | Recovery and weekly prompts warn about missing readiness but do not require citing `_meta.provenance.readiness.source/native_scale` when readiness is present. | Step 2 should harden wording/goldens so assistants label provider-native readiness instead of calling it generic recovery. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 16:18 | Plan review | Step 3 APPROVE |
| 2026-06-03 16:17 | Plan review | Step 2 APPROVE |
| 2026-06-03 16:16 | Plan review | Step 1 APPROVE |
| 2026-06-03 16:15 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:15 | Step 0 started | Preflight |
| 2026-06-03 16:17 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 16:20 | Review R002 | plan Step 2: APPROVE |
| 2026-06-03 16:24 | Review R003 | plan Step 3: APPROVE |

| 2026-06-03 16:27 | Worker iter 1 | done in 692s, tools: 113 |
| 2026-06-03 16:27 | Task complete | .DONE created |