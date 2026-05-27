# TP-104: As-of metadata for time-relative reads — Status

**Current Step:** Step 1: Design and implement shared athlete-local as-of helper
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 2
**Review Counter:** 2
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

- [x] Helper returns local RFC3339 datetime, date, weekday, and timezone
- [x] Shared helper API/location and single-localized-instant contract documented in plan
- [x] Clock contract for deterministic current-day checks documented in plan
- [x] Timezone error/fallback behavior documented in plan
- [x] Helper test coverage plan includes offset, weekday, empty/trimmed, and invalid-zone cases
- [x] Timezone edge cases covered with deterministic tests
- [x] Existing timezone error behavior preserved
- [x] Targeted helper tests passing

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
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | .reviews/R002-plan-step1.md |

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
- Step 1 plan: add a shared `internal/response.AsOfMetadata(now time.Time, timezone string)` helper returning one struct with `as_of`, `as_of_date`, `as_of_weekday`, and `timezone`, all derived from a single localized instant. The helper will reuse the existing timezone loading path used by `RenderTimeInTimezone`/`RenderDateInTimezone`; malformed zones return the existing wrapped load error and empty timezone continues to resolve to UTC. `get_today` keeps using its injectable `now func() time.Time`; Step 3 tools will receive injectable clock constructors before calling the helper/current-day range predicate, avoiding direct untestable `time.Now()` in handlers. Tests will cover positive/negative offset date shifts, weekday consistency, trimmed and empty timezone behavior, and invalid-zone errors.
| 2026-05-27 12:19 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-27 12:22 | Review R002 | plan Step 1: APPROVE |
