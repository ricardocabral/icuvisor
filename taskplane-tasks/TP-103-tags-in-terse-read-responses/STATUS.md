# TP-103: Tags in terse read responses — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-05-26
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm issue #30 scope and avoid unrelated tag/filter features

---

### Step 1: Implement event tag read shaping
**Status:** ⬜ Not Started

- [ ] Decode or extract upstream `tags` for events without guessing missing values
- [ ] Add `tags` to shared terse event rows while preserving order
- [ ] Ensure `get_events`, `get_event_by_id`, `add_or_update_event`, and `get_today` share behavior
- [ ] Targeted event tests passing

---

### Step 2: Investigate and implement activity tag handling if supported
**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand based on actual activity payload/model support discovered in source and fixtures.

- [ ] Activity tag availability determined and either implemented or documented with a regression test/discovery
- [ ] Targeted activity tests passing

---

### Step 3: Regression tests and docs
**Status:** ⬜ Not Started

- [ ] Tags present/empty/missing/null cases covered
- [ ] `include_full` and null-stripping behavior covered
- [ ] `CHANGELOG.md` updated if behavior changes
- [ ] Targeted tests passing

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 5: Documentation & Delivery
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

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/30
