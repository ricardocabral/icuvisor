# TP-103: Tags in terse read responses — Status

**Current Step:** Step 2: Investigate and implement activity tag handling if supported
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 2
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm issue #30 scope and avoid unrelated tag/filter features

---

### Step 1: Implement event tag read shaping
**Status:** ✅ Complete

- [x] Extract event tags from raw upstream payload only when `tags` is a JSON string array, preserving explicit empty arrays and omitting missing/null/non-array/mixed values
- [x] Add copied `tags` pointer to shared terse event rows while preserving order and inherited shared-row behavior for all event response helpers
- [x] Ensure `get_events`, `get_event_by_id`, `add_or_update_event`, and `get_today` share behavior
- [x] Targeted event tests cover present/order, explicit empty, missing/null/malformed omission, include_full raw preservation, and affected event paths
- [x] Targeted event tests passing

---

### Step 2: Investigate and implement activity tag handling if supported
**Status:** 🟨 In Progress

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
| 2026-05-27 12:12 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 12:12 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/30
| 2026-05-27 12:19 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-27 12:22 | Review R002 | plan Step 1: APPROVE |
