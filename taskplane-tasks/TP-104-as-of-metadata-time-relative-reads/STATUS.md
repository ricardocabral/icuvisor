# TP-104: As-of metadata for time-relative reads — Status

**Current Step:** Step 1: Design and implement shared athlete-local as-of helper
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm current metadata conventions before adding new keys

---

### Step 1: Design and implement shared athlete-local as-of helper
**Status:** 🟨 In Progress

- [ ] Helper returns local RFC3339 datetime, date, weekday, and timezone
- [ ] Timezone edge cases covered with deterministic tests
- [ ] Existing timezone error behavior preserved
- [ ] Targeted helper tests passing

---

### Step 2: Add metadata to `get_today`
**Status:** ⬜ Not Started

- [ ] `get_today` meta includes `as_of`, `as_of_date`, `as_of_weekday`, and timezone
- [ ] Existing injectable clock used in tests
- [ ] Existing `date`, `activity_window`, and counts preserved
- [ ] Targeted `get_today` tests passing

---

### Step 3: Add metadata to current-day range reads
**Status:** ⬜ Not Started

- [ ] `get_activities` current-day range metadata added
- [ ] `get_events` current-day range metadata added
- [ ] `get_wellness_data` current-day range metadata added
- [ ] Pagination/null-stripping/terse-full behavior preserved
- [ ] Targeted tool tests passing

---

### Step 4: Regression tests and changelog
**Status:** ⬜ Not Started

- [ ] Positive/negative timezone boundary cases covered
- [ ] Date ranges including/excluding local today covered
- [ ] Past-only range behavior verified
- [ ] `CHANGELOG.md` updated

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
| 2026-05-27 12:12 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 12:12 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/31
