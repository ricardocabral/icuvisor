# TP-155: Activity interval manual and mixed classification — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-09
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

---

### Step 1: Add classifier states and fixture coverage
**Status:** ⬜ Not Started

- [ ] Add `manual_added` and `mixed` states in the classifier
- [ ] Add regression tests for grouped, ungrouped, mixed, structured, and device-lap cases
- [ ] Preserve existing precedence behavior
- [ ] Targeted analysis tests pass

---

### Step 2: Propagate source evidence to tool/analyzer responses
**Status:** ⬜ Not Started

- [ ] Analyzer metadata supports the new values
- [ ] get-activity-interval output/tests expose the new classifications
- [ ] Schema snapshot refreshed if needed
- [ ] Targeted tool and analysis tests pass

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Integration tests (if applicable)
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated
- [ ] `README.md` reviewed/updated if affected
- [ ] Discoveries logged

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
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |

---

## Blockers

*None*

---

## Notes

Public signal: IcuSync forum #265-266 reports auto-detected intervals have `group_id`; manually added intervals do not.
