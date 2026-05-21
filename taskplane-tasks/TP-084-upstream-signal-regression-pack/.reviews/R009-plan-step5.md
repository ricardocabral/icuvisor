# Plan Review — TP-084 Step 5

**Verdict:** REQUEST CHANGES

I reviewed `PROMPT.md`, `STATUS.md`, the current review history, and the current regression-test state. The Step 5 checklist names the required high-level verification commands, but the plan is not approvable because the task has advanced into final testing while earlier blocking reviews remain unresolved and are recorded incorrectly.

## Blocking plan gaps

1. **Do not run Step 5 as a closure step over unresolved review blockers.**  
   `R008-code-step4.md` requested changes, but `STATUS.md:51-68` marks Step 4 complete and moves the task to Step 5. `STATUS.md:127` also records R008 as `APPROVE` even though the review verdict is `REQUEST CHANGES`. Step 5 needs an explicit precondition to resolve or formally disposition R008 (and the earlier R004/R006/R007 blockers) before marking any verification as final.

2. **The plan still does not address the missing regression coverage required by the mission.**  
   The current tests still leave the same gaps called out by prior reviews:
   - `internal/tools/get_event_by_id_test.go:124-166` loads the listed-event fixture only to extract the ID, then configures the fake list response with `events: nil`, so it does not lock the listed-event/detail-404 mismatch against recovery semantics.
   - `internal/tools/get_activities_strava_test.go:69-86` covers the sync-chain fixture only at `isStravaBlocked` helper level, while `internal/tools/get_activity_details_test.go:105-127` covers detail output; there is still no `get_activities` handler-output assertion for the numeric/no-`i` Wahoo/MyWhoosh/TrainerRoad rows.

   Step 5 verification over the current pack would not satisfy the prompt's completion criteria that all three upstream signals have stable regression tests.

3. **Fix the `STATUS.md` audit trail before final verification.**  
   The Reviews table is empty (`STATUS.md:82-85`), while review rows are appended under Notes (`STATUS.md:120-127`) and record multiple `REQUEST CHANGES` reviews as `APPROVE` (`R001`, `R003`, `R004`, `R005`, `R006`, `R007`, `R008`). Step 5's plan should include correcting this history and moving entries to the appropriate Reviews/Execution Log sections before claiming all failures are fixed or documented.

4. **Make targeted verification explicit and non-cached after fixes.**  
   The checklist includes `make test`, `make build`, and `make lint`, which is good, but it should also name the focused regression commands that must be rerun after the missing coverage/semantic fixes, for example:
   - `go test -count=1 ./internal/tools ./internal/intervals`
   - focused `go test -count=1 ./internal/tools -run 'TestGetActivities|TestGetActivityDetails|TestGetEventByID|TestAddOrUpdateEvent'` (or equivalent exact test names after edits)

   Any failures from those commands should be fixed, or clearly documented in `STATUS.md` as pre-existing and unrelated.

## Summary

The Step 5 command set is directionally correct, but this task is not ready for final testing/verification. Revise the plan to first resolve the outstanding regression-coverage blockers and repair the status/review audit trail, then rerun targeted non-cached tests plus `make test`, `make build`, and `make lint` with exact outcomes recorded in `STATUS.md`.
