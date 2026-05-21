# Plan Review — TP-084 Step 1

**Verdict:** Request changes / plan needs more specificity before implementation.

I reviewed `PROMPT.md`, `STATUS.md`, current edge-case tests, and existing fixtures under `internal/intervals/testdata`. Step 1 is currently represented only by the high-level checklist in `STATUS.md`; there is no concrete fixture/marker inventory or collection plan to approve. Because this task is specifically about locking upstream behavior into stable regression fixtures, Step 1 should produce an explicit inventory before Step 2 starts.

## Required plan additions

1. **List the exact existing fixtures and gaps.** The plan should name the current fixtures it will rely on and the new sanitized fixtures it expects to add. Current relevant files include:
   - `internal/intervals/testdata/events/inconsistent/synthetic_list.json`
   - `internal/intervals/testdata/events/inconsistent/synthetic_detail_404.txt`
   - `internal/intervals/testdata/events/note_create_request.json`
   - `internal/intervals/testdata/events/note_create_response.json`

   The apparent gap is the Strava numeric/no-`i` empty-stub set for Wahoo/MyWhoosh/TrainerRoad sync chains. The plan should state where those sanitized fixtures will live, e.g. a dedicated `internal/intervals/testdata/activities/strava_stubs/` directory or another existing testdata location, instead of embedding all payloads inline in tests.

2. **Define the exact structured markers as Step 1 output.** The plan should pin the expected payload semantics before writing tests:
   - `get_activities` Strava stub rows: `strava_imported: true`, `unavailable.reason: "strava_tos"`, non-empty `unavailable.workaround`, no fabricated metrics for empty payloads.
   - activity detail/interval/stream unavailable fallback where applicable: confirm whether the expected reason is `"strava_blocked"` or `"strava_tos"`; current list rows use `strava_tos`, while detail fallback tests use `strava_blocked` in some paths and `strava_tos` inside the activity row.
   - `get_event_by_id` 404-after-list miss: `unavailable.reason: "upstream_inconsistency"`, `unavailable.retried: ["detail", "list_scan"]`, no `event` object, and `_meta.scanned_range`, `_meta.count`, `_meta.truncated` expectations.

3. **Clarify the no-`i`/numeric stub fixture shape.** The mission calls out “Strava numeric/no-`i` empty stubs from sync chains.” The plan should specify the fields that make those rows representative and sanitized: numeric-looking `id`, omitted or non-`i` athlete identifier if present, source/sync-chain markers if upstream exposes them, and intentionally empty/null name/metrics. It should also document that all athlete IDs, calendar IDs, names, notes, and timestamps are synthetic or scrubbed.

4. **Confirm NOTE coverage at the correct layer.** Existing interval-client coverage already asserts `WriteEventParams{Date: "2026-05-25", Category: "NOTE"}` sends `start_date_local: "2026-05-25T00:00:00"` via `note_create_request.json`. Step 1 should explicitly decide whether Step 2 only strengthens this existing `internal/intervals/events_test.go` fixture test, or also adds a tool-level `add_or_update_event` test proving the public tool still accepts date-only NOTE input.

5. **Record fixture provenance/redaction in `STATUS.md` as part of Step 1.** The prompt requires provenance/redaction notes later, but Step 1 is the fixture collection step. The plan should include updating `STATUS.md` discoveries/notes with whether fixtures are synthetic, dogfood-derived and scrubbed, or copied from existing sanitized TP-012/TP-075 fixtures. Do not wait until delivery if new fixture content is added during Step 1.

## Minor recommendations

- Prefer fixture-backed tests for the upstream-signal examples and keep inline JSON only for tiny synthetic assertions.
- Avoid changing production code in Step 1; this step should only inventory/add sanitized fixtures and expected markers. If marker definitions reveal an implementation mismatch, defer code changes to Step 3.
- Include a targeted command list for verifying the fixture inventory compiles/parses, such as `go test ./internal/intervals ./internal/tools` after Step 2, but not necessarily during Step 1.

## Summary

The high-level checklist matches the task, but the plan is too underspecified for a regression-pack task. Please revise Step 1 to name the fixture files, exact marker contracts, NOTE serialization contract, and redaction/provenance notes before proceeding to test implementation.
