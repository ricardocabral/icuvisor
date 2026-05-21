# TP-084: Upstream-signal regression pack from 2026-05 behavior review — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 10
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Collect fixtures and expected markers
**Status:** ✅ Complete

- [x] Identify existing fixtures for each upstream signal and add sanitized fixtures where coverage is missing.
- [x] Define exact expected structured markers for Strava stubs and event inconsistency.
- [x] Confirm NOTE expected outbound payload uses local datetime for date-only input.

---

### Step 2: Add regression tests
**Status:** ✅ Complete

- [x] Add tests for numeric/no-`i` Strava empty stubs from Wahoo/MyWhoosh/TrainerRoad chains.
- [x] Add/strengthen `get_event_by_id` 404-after-list fixture coverage.
- [x] Add/strengthen NOTE date-only create serialization test.

---

### Step 3: Fix any regressions exposed by the pack
**Status:** ✅ Complete

- [x] Apply minimal code fixes if new tests fail.
- [x] Keep changes additive and schema-stable.
- [x] Run targeted affected tool tests.

---

### Step 4: Verify and document
**Status:** ✅ Complete

- [x] Run full suite/build/lint.
- [x] Update CHANGELOG.md and upstream-gap notes only if new behavior changed.
- [x] Record fixture provenance/redaction notes in STATUS.md.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Strava sync-chain fixture uses synthetic Wahoo/MyWhoosh/TrainerRoad external IDs, numeric/no-`i` activity IDs, scrubbed athlete ID `i00000`, and no personal activity names/metrics. | Added sanitized regression fixture and README; no live payloads committed. | `internal/intervals/testdata/activities/strava_sync_chain_empty_stubs.json`, `internal/intervals/testdata/activities/README.md` |
| Event inconsistency and NOTE fixtures were already sanitized/synthetic or placeholder-based. | Reused existing fixtures; no new upstream personal data captured. | `internal/intervals/testdata/events/inconsistent/*`, `internal/intervals/testdata/events/note_create_*.json` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 13:29 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 13:29 | Step 0 started | Preflight |
| 2026-05-20 14:07 | Worker iter 1 | done in 2297s, tools: 158 |
| 2026-05-20 14:07 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- Step 1 expected markers: `get_activities` / `get_activity_details` rows for empty Strava stubs use `strava_imported: true` and `unavailable.reason: "strava_tos"`; activity subresource fallbacks (`get_activity_intervals`, streams, splits, extended metrics) use `unavailable.reason: "strava_blocked"` with the existing non-empty `workaround` string; `get_event_by_id` list/detail mismatch uses `unavailable.reason: "upstream_inconsistency"` and `retried: ["detail", "list_scan"]`.
- Step 1 NOTE payload confirmation: `internal/intervals/events.go` appends `T00:00:00` for date-only `WORKOUT` and `NOTE` writes; `internal/intervals/testdata/events/note_create_request.json` expects `"start_date_local": "2026-05-25T00:00:00"` for date-only tool/client input.
- Step 3 targeted regression run: `go test ./internal/tools ./internal/intervals` passed; no minimal production-code fixes were required.
- Step 4 docs decision: CHANGELOG.md and docs/upstream-gaps were reviewed; no user-facing behavior changed, so no changelog/upstream-gap content update was required.
- Step 5 verification: targeted `go test ./internal/tools ./internal/intervals`, `make test`, `make build`, and `make lint` all passed with no unrelated or pre-existing failures.
| 2026-05-20 13:32 | Review R001 | plan Step 1: APPROVE |
| 2026-05-20 13:37 | Review R002 | code Step 1: APPROVE |
| 2026-05-20 13:40 | Review R003 | plan Step 2: APPROVE |
| 2026-05-20 13:47 | Review R004 | code Step 2: APPROVE |
| 2026-05-20 13:50 | Review R005 | plan Step 3: APPROVE |
| 2026-05-20 13:54 | Review R006 | code Step 3: APPROVE |
| 2026-05-20 13:55 | Review R007 | plan Step 4: APPROVE |
| 2026-05-20 13:59 | Review R008 | code Step 4: APPROVE |
| 2026-05-20 14:02 | Review R009 | plan Step 5: APPROVE |
| 2026-05-20 14:05 | Review R010 | code Step 5: APPROVE |
