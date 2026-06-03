# TP-140: Long-distance event distance regression coverage — Status

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

### Step 1: Audit event distance handling
**Status:** ✅ Complete

- [x] Inspect event write/read validation for distance limits, units, and load/target-load wording.
- [x] Confirm whether any Icuvisor-local cap below 1200 km exists; record upstream-only constraints in STATUS.md.
- [x] Check response wording does not imply Icuvisor auto-calculates load from distance/duration unless it actually does.
- [x] Run targeted tests: `go test ./internal/tools`.

---

### Step 2: Add long-distance regression tests
**Status:** ✅ Complete

- [x] Add tests for creating/updating and reading a 1200 km event or workout/race distance as meters.
- [x] Remove or relax any arbitrary Icuvisor-local cap below randonneuring distances, preserving upstream error passthrough as actionable user errors.
- [x] Assert no truncation and no false auto-load claim in response metadata/rows.
- [x] Run targeted tests: `go test ./internal/tools`.

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
| 2026-06-03 | Step 1 | Event write validation only rejects negative `distance_meters`/`target_load`; no Icuvisor-local maximum distance cap found in `internal/tools` or `internal/intervals`. Writes send `distance_meters` as upstream `distance_target` and reads preserve upstream `distance`/`distance_target` as meter-valued float fields. | Long-distance acceptance is governed by upstream intervals.icu constraints; regression tests should prove Icuvisor preserves 1200 km values locally. |
| 2026-06-03 | Step 1 | Response/schema wording says `target_load` is optional planned load when supported upstream and `distance_meters` is optional planned distance; no response metadata claims Icuvisor calculates load from distance/duration. | Add assertions around response rows/metadata to prevent false auto-load wording from being introduced. |
| 2026-06-03 | Step 2 | No arbitrary local cap required removal; new tool-level regression tests cover 1,200,000 meter create/update/read paths so future caps below randonneuring distances fail fast. | Upstream errors remain passed through by existing write error handling; local code continues not to reject high non-negative distances. |
| 2026-06-03 | Step 4 | Reviewed `docs/prd/PRD-icuvisor.md`; no material product contract change was needed because event tools already describe calendar event read/write behavior without imposing a distance maximum. | Changelog is sufficient user-visible documentation for this regression coverage. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|
| 2026-06-03 | Step 1 plan | APPROVE | Plan review approved before audit implementation. |
| 2026-06-03 | Step 2 plan | APPROVE | Plan review approved before adding regression tests. |
| 2026-06-03 | Step 3 plan | APPROVE | Plan review approved before full verification. |

| 2026-06-03 16:07 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:07 | Step 0 started | Preflight |
| 2026-06-03 16:08 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 16:11 | Review R002 | plan Step 2: APPROVE |
| 2026-06-03 16:13 | Review R003 | plan Step 3: APPROVE |

| 2026-06-03 16:15 | Worker iter 1 | done in 490s, tools: 83 |
| 2026-06-03 16:15 | Task complete | .DONE created |