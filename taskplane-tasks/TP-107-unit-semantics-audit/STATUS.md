# TP-107: Unit semantics audit — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-26
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current unit/metadata behavior scoped before changing code

---

### Step 1: Add workout target unit regression coverage
**Status:** ⬜ Not Started

- [ ] Percent FTP / power target serialization tests added
- [ ] Pace target range and unit tests added
- [ ] Heart-rate percent variant tests added where supported
- [ ] Serializer fixes applied only if required
- [ ] Targeted workoutdoc tests passing

---

### Step 2: Add work/energy and unknown-unit regression coverage
**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand based on actual unit-bearing surfaces found during audit.

- [ ] Joules/kilojoules surfaces audited and covered
- [ ] Raw joules not mislabeled as kilojoules
- [ ] Unknown units preserved rather than guessed
- [ ] Targeted unit/response tests passing

---

### Step 3: Add calories and hydration semantics coverage
**Status:** ⬜ Not Started

- [ ] Activity `calories_burned` and wellness `calories_intake` distinction covered
- [ ] `hydration` versus `hydrationVolume` semantics covered or clarified
- [ ] Explanatory metadata added if needed without bloating terse responses
- [ ] Targeted wellness/activity tests passing

---

### Step 4: Changelog and full verification
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated if behavior or metadata changes
- [ ] Unit-surface discoveries logged
- [ ] Targeted tests passing

---

### Step 5: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final commit includes task ID

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
| 2026-05-26 | Task staged | PROMPT.md and STATUS.md created |

---

## Blockers

*None*

---

## Notes

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/34
