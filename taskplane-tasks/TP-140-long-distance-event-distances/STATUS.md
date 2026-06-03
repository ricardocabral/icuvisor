# TP-140: Long-distance event distance regression coverage — Status

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

### Step 1: Audit event distance handling
**Status:** ⬜ Not Started

- [ ] Inspect event write/read validation for distance limits, units, and load/target-load wording.
- [ ] Confirm whether any Icuvisor-local cap below 1200 km exists; record upstream-only constraints in STATUS.md.
- [ ] Check response wording does not imply Icuvisor auto-calculates load from distance/duration unless it actually does.
- [ ] Run targeted tests: `go test ./internal/tools`.

---

### Step 2: Add long-distance regression tests
**Status:** ⬜ Not Started

- [ ] Add tests for creating/updating and reading a 1200 km event or workout/race distance as meters.
- [ ] Remove or relax any arbitrary Icuvisor-local cap below randonneuring distances, preserving upstream error passthrough as actionable user errors.
- [ ] Assert no truncation and no false auto-load claim in response metadata/rows.
- [ ] Run targeted tests: `go test ./internal/tools`.

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
