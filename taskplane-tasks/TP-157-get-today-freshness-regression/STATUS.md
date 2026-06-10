# TP-157: get_today current-day freshness regression — Status

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

### Step 1: Add explicit freshness regression tests
**Status:** ⬜ Not Started

- [ ] Fixed timezone/as_of test added
- [ ] Previous-day activity/event/fitness rows excluded
- [ ] Partial or absent wellness does not backfill yesterday
- [ ] Targeted get_today tests pass

---

### Step 2: Fix filtering or shaping if tests expose stale composition
**Status:** ⬜ Not Started

- [ ] Date filtering fixed if needed
- [ ] Existing wellness stale metadata preserved
- [ ] Fetch/filter boundaries are defensive against stale rows
- [ ] Targeted get_today tests pass

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
- [ ] README/PRD reviewed if affected
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

Public signal: IntervalCoach forum #852-855 described a mixed old/new daily briefing bug.
