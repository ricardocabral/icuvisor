# Plan Review — TP-084 Step 2

**Verdict:** REQUEST CHANGES

I reviewed `PROMPT.md`, `STATUS.md`, the Step 1 review notes, the new activity fixture, and the current tests around activities, `get_event_by_id`, and NOTE event creation. The Step 2 plan is still only the high-level checklist from `STATUS.md`. Step 1 did define useful marker contracts, but Step 2 needs a concrete test map before implementation so this regression pack locks the intended upstream signals instead of duplicating or contradicting existing coverage.

## Blocking plan gaps

1. **Map each new regression to an exact test file/function and fixture.**
   The plan should state which tests will be added or strengthened, for example:
   - `internal/tools/get_activities_test.go` or `get_activities_strava_test.go` loading `internal/intervals/testdata/activities/strava_sync_chain_empty_stubs.json`.
   - `internal/tools/get_activity_details_test.go` using at least one numeric/no-`i` empty-stub fixture if Step 1's `get_activity_details` marker contract is in scope.
   - `internal/tools/get_event_by_id_test.go` using `internal/intervals/testdata/events/inconsistent/synthetic_list.json` and the matching 404 note, or an explicit reason if a tool-level fake is preferred.
   - `internal/tools/add_or_update_event_test.go` for public tool acceptance of date-only NOTE input, and/or `internal/intervals/events_test.go` for the wire payload assertion.

2. **Resolve the `get_event_by_id` semantic conflict before writing tests.**
   The task mission says a detail 404 after a listed event should remain `unavailable.reason: "upstream_inconsistency"`. Current `TestGetEventByIDFallbackScansDateWindowWithResolveAndCap` expects a detail 404 to recover from `list_scan` when the list contains the requested ID. A fixture-backed "listed ID + detail 404" test expecting `upstream_inconsistency` will conflict with that existing recovery test. The Step 2 plan must explicitly say whether to:
   - update/replace the recovery test because the desired contract is now structured inconsistency for listed-but-detail-404, or
   - narrow the new regression to the existing "detail 404 + list scan miss" behavior and explain why that still satisfies the mission wording.

   Without this decision, Step 2 risks adding mutually incompatible tests or preserving the wrong user-facing behavior.

3. **Define the Strava stub assertions beyond `reason`.**
   The plan should pin the expected response checks for the Wahoo/MyWhoosh/TrainerRoad fixture rows: numeric activity IDs are preserved, rows are not dropped by default unnamed filtering, `strava_imported: true`, `unavailable.reason: "strava_tos"`, `unavailable.workaround` is non-empty, and metric/full fields are not fabricated from empty/null upstream data. If classification-only tests are added in `get_activities_strava_test.go`, they should not be the only coverage; the task asks for structured tool response markers.

4. **Clarify the NOTE coverage layer.**
   There is already an intervals-client test asserting `WriteEventParams{Date: "2026-05-25", Category: "NOTE"}` sends `start_date_local: "2026-05-25T00:00:00"` via `internal/intervals/testdata/events/note_create_request.json`. Step 2 should state whether it will:
   - keep that as the wire-level serialization regression and add a tool-level `add_or_update_event` test proving date-only NOTE input is accepted without `type`, or
   - strengthen only the existing intervals-client test.

   A tool-level fake can only assert the `WriteEventParams` passed by the tool, not the final upstream JSON datetime, so the plan should avoid claiming tool-level tests prove wire serialization unless the intervals-client fixture remains part of the coverage.

5. **Include the expected targeted verification command.**
   Step 2 should list the affected test command(s), at minimum something like `go test ./internal/tools ./internal/intervals`. If the new `get_event_by_id` test is expected to fail until Step 3 exposes a real regression, call that out explicitly in `STATUS.md` rather than silently changing behavior during Step 2.

## Minor recommendations

- Prefer fixture-backed tests for `strava_sync_chain_empty_stubs.json` rather than re-embedding those payloads inline.
- Keep Step 2 to tests only. If a new test fails, document it and defer production code changes to Step 3 as the prompt requires.
- When updating `STATUS.md`, move the existing R001/R002 execution-log rows out of `## Notes` into the execution log table while adding the Step 2 test plan details.

## Summary

The high-level Step 2 checklist is aligned with TP-084, but it is not specific enough to approve. Please revise the plan to name the exact tests, fixtures, assertions, and especially the intended `get_event_by_id` detail/list mismatch semantics before implementing the regression tests.
