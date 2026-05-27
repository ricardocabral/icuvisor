# TP-104: As-of metadata for time-relative reads — Status

**Current Step:** Step 3: Add metadata to current-day range reads
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 2
**Review Counter:** 9
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
**Status:** ✅ Complete

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
**Status:** ✅ Complete

- [x] `get_today` meta includes `as_of`, `as_of_date`, `as_of_weekday`, and timezone
- [x] Step 2 plan documents single `now()` anchor, existing meta preservation, helper timezone, and targeted tests
- [x] Existing injectable clock used in tests
- [x] Existing `date`, `activity_window`, and counts preserved
- [x] Targeted `get_today` tests passing

---

### Step 3: Add metadata to current-day range reads
**Status:** ✅ Complete

- [x] Shared current-day range predicate, as-of meta application helper, and injectable clock constructors added for range tools
- [x] `get_activities` current-day range metadata added
- [x] `get_events` current-day range metadata added
- [x] `get_wellness_data` current-day range metadata added
- [x] Pagination/null-stripping/terse-full behavior preserved
- [x] Targeted tool tests passing
- [x] Output schemas for `get_events` and `get_wellness_data` document current-day as-of metadata

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
| R003 | Code | 1 | APPROVE | .reviews/R003-code-step1.md |
| R004 | Plan | 2 | REVISE | .reviews/R004-plan-step2.md |
| R005 | Plan | 2 | APPROVE | .reviews/R005-plan-step2.md |
| R006 | Code | 2 | APPROVE | .reviews/R006-code-step2.md |
| R007 | Plan | 3 | APPROVE | .reviews/R007-plan-step3.md |
| R008 | Code | 3 | REVISE | .reviews/R008-code-step3.md |
| R009 | Code | 3 | APPROVE | .reviews/R009-code-step3.md |

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
- Step 2 plan: in `getTodayHandler`, call the injectable `now()` exactly once, pass that instant to `response.AsOfMetadataInTimezone`, use `asOf.AsOfDate` for the existing `today` fetch date, and pass the full helper result into `shapeGetTodayResponse` so `date` and `as_of_date` cannot diverge across midnight. Extend `getTodayMeta` with `as_of`, `as_of_date`, and `as_of_weekday` while preserving existing `date`, `timezone`, `include_full`, `source_tools`, `section_counts`, `activity_window`, and response-shaper-added metadata such as `units`. Populate `_meta.timezone` from the helper's trimmed/defaulted `Timezone`. Update `get_today` tests through `newGetTodayToolWithClock` to assert exact São Paulo boundary `as_of*` values and unchanged local-date fetches/counts, then run `go test ./internal/tools -run TestGetToday`.
- Step 3 plan: add a small tools-level helper that computes `response.AsOfMetadataInTimezone(now(), timezone)` once per request and attaches `as_of`, `as_of_date`, `as_of_weekday`, and helper-normalized `timezone` only when the normalized request date range includes `asOf.AsOfDate`; closed ranges require `oldest <= today <= newest`, and `get_activities` with blank `newest` treats the range as open-ended through upstream now. Add with-clock constructors for activities, events, and wellness so tests do not depend on wall-clock time. Preserve each tool's existing pagination token, count, null stripping, terse/full, and response-shaper metadata by only extending response meta structs immediately before shaping.
| 2026-05-27 12:19 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-27 12:22 | Review R002 | plan Step 1: APPROVE |
| 2026-05-27 12:29 | Review R003 | code Step 1: APPROVE |
| 2026-05-27 12:33 | Review R004 | plan Step 2: REVISE |
| 2026-05-27 12:36 | Review R005 | plan Step 2: APPROVE |
| 2026-05-27 12:44 | Review R006 | code Step 2: APPROVE |
| 2026-05-27 12:50 | Review R007 | plan Step 3: APPROVE |
| 2026-05-27 13:10 | Review R008 | code Step 3: REVISE |
| 2026-05-27 13:17 | Review R009 | code Step 3: APPROVE |
