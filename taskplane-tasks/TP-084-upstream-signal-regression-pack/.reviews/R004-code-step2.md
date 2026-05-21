# Code Review — TP-084 Step 2

**Verdict:** REQUEST CHANGES

I reviewed the diff from `20e8c79..HEAD` and ran the targeted tests.

## Findings

1. **Blocking — `get_event_by_id` regression does not exercise the listed-event mismatch fixture.**  
   In `TestGetEventByIDListedFixtureDetail404ReturnsInconsistencyWhenRescanMisses`, the fixture is only read to obtain `listedID`, but the fake client is configured with `events: nil` (`internal/tools/get_event_by_id_test.go:127-143`). That means the tool is testing the already-covered “detail 404 + empty list scan” path, not the task’s “event was listed, then detail returned 404” mismatch. The existing recovery test still asserts that a detail 404 is recovered when the fallback list contains the requested ID (`TestGetEventByIDFallbackScansDateWindowWithResolveAndCap`, same file), so the user-facing contract from the prompt/STATUS is still not locked. Please either feed `listedEvents` into the fake list response and update the expected behavior to `upstream_inconsistency`, or otherwise add an explicit test that fails if a listed/detail-404 mismatch is returned as a recovered event.

2. **Blocking — numeric sync-chain Strava fixture is not tested through `get_activities` structured output.**  
   The new fixture-backed activity coverage is helper-level classification (`internal/tools/get_activities_strava_test.go:69-84`) plus `get_activity_details` (`internal/tools/get_activity_details_test.go:105-126`). The mission and Step 1 notes explicitly require `get_activities` rows for these Wahoo/MyWhoosh/TrainerRoad numeric/no-`i` empty stubs to surface structured unavailable markers. A regression in list pagination/filtering or list response shaping for this exact fixture could still pass these tests. Add a `get_activities` handler test loading `strava_sync_chain_empty_stubs.json` and assert all three rows survive default unnamed filtering with numeric IDs, `strava_imported: true`, `unavailable.reason: "strava_tos"`, non-empty workaround, and no fabricated metric fields.

3. **Non-blocking — STATUS records the previous review incorrectly and outside the execution table.**  
   `STATUS.md:116-118` appends execution-log rows under `## Notes` instead of the `## Execution Log` table, and it records `Review R003 | plan Step 2: APPROVE` even though the committed `R003-plan-step2.md` says `REQUEST CHANGES`. Please fix this before closing the step so the task audit trail is reliable.

## Tests run

- `go test ./internal/tools`
- `go test ./internal/tools ./internal/intervals`
