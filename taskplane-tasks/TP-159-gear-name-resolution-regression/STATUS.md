# TP-159: Gear name resolution regression — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-09
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** S

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

---

### Step 1: Add direct numeric-gear-id regression coverage
**Status:** ⬜ Not Started

- [ ] Activity fixture exposes gear ID without embedded name
- [ ] Gear list fixture resolves the name
- [ ] Unknown gear fallback covered
- [ ] Targeted gear/activity tests pass

---

### Step 2: Fix resolver behavior only if the regression fails
**Status:** ⬜ Not Started

- [ ] Resolver fixes applied if needed
- [ ] Cache/error behavior preserved
- [ ] README wording checked
- [ ] Targeted gear/activity tests pass

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

Public signal: IcuSync forum #269-270 noted activities may expose only numeric `gear_id`; resolving names requires fetching full gear list.
