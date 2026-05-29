# TP-118: Activity tombstone delete endpoint — Status

**Current Step:** Step 1: Determine the correct activity deletion contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room guardrail confirmed

---

### Step 1: Determine the correct activity deletion contract
**Status:** 🟨 In Progress

- [ ] Existing delete implementation and tests inspected
- [ ] Public upstream evidence checked without competitor source
- [ ] Endpoint decision recorded in Discoveries
- [ ] `/api/v1` base-path handling captured for the selected endpoint
- [ ] Targeted tests run with regex covering intervals path and target-athlete safety tests

---

### Step 2: Implement and lock the endpoint behavior
**Status:** ⬜ Not Started

- [ ] Client path/method updated if needed
- [ ] httptest coverage asserts exact method/path and target-athlete safety
- [ ] Tool metadata/schema snapshots updated if affected
- [ ] Targeted tests run

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passing if source changed
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated if needed
- [ ] Roadmap/PRD checked if affected
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 15:21 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 15:21 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

Plan review R001 requires Step 1 tests to include `DeleteMethods|ActivityIDEndpointsRequireResolvedTargetOwnership` coverage and discoveries to explicitly note `/api/v1` base URL handling.
| 2026-05-29 15:25 | Review R001 | plan Step 1: REVISE |
