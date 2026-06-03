# TP-150: Workout DSL metric suffix from sport priority — Status

**Current Step:** Step 1: Design the sport-aware suffix boundary
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
- [x] Confirm current tests lock bare power-zone serialization

---

### Step 1: Design the sport-aware suffix boundary
**Status:** 🟨 In Progress

- [ ] Decide implementation boundary for suffix selection
- [ ] Preserve no-sport-context serializer behavior
- [ ] Define expected suffix behavior for primary metric orders
- [ ] Document upstream ambiguity in STATUS.md

---

### Step 2: Implement and test metric suffix behavior
**Status:** ⬜ Not Started

- [ ] Add Run `POWER_HR_PACE` regression test
- [ ] Add HR-primary / pace-primary coverage where applicable
- [ ] Implement minimal behavior change
- [ ] Targeted tests passing

---

### Step 3: Refresh schemas and user guidance
**Status:** ⬜ Not Started

- [ ] Tool descriptions/schema wording updated if needed
- [ ] Schema snapshots regenerated if needed
- [ ] End-user workout docs updated if needed
- [ ] CHANGELOG updated if user-visible

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passes
- [ ] All failures fixed
- [ ] Build passes

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Clean-room behavior source summarized

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Step 0 evidence: `go test ./internal/workoutdoc -run TestSerializeTargetUnitSemantics -count=1` passed and existing case `POWER_ZONE` expects bare `Z2`.
