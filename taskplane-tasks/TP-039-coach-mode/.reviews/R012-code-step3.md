# R012 code review — Step 3: Tool registry plumbing

Verdict: **REVISE**

I ran:

- `git diff a527c36..HEAD --name-only`
- `git diff a527c36..HEAD`
- `go test ./...`

The test suite passes, and the R011 fixes are largely in place: activity ownership is now checked for the direct activity endpoints, advanced-capabilities coach filtering is centralized in `internal/mcp.safeRegistrar`, and coach-mode-off schema compatibility is preserved. I found one remaining blocking target-routing gap.

## Finding

### `link_activity_to_event` checks the activity target but not the event target before writing

`LinkActivityToEvent` now verifies that `activity_id` belongs to the resolved target athlete, but it still accepts an arbitrary `event_id` and immediately writes it into the activity:

- `internal/intervals/activities.go:121-125` checks `ensureActivityIDTarget(ctx, activityID)` and then sends `paired_event_id: eventID` to `PUT /activity/{activityID}`.
- The available target-routed event lookup already exists at `internal/intervals/events.go:117-125` (`GET /athlete/{target}/events/{eventID}`), but it is not used before the write.
- `internal/tools/link_activity_to_event.go:58-75` only calls `linkActivityToEventWarnings` after the link is applied, and that helper silently ignores `GetEvent` failures, so it cannot enforce target ownership.

In coach mode this leaves a cross-object escape hatch: a model can provide a valid roster `athlete_id` and an activity owned by that athlete, but pair it to an event ID that belongs to another athlete or is outside the configured roster if upstream accepts the `paired_event_id`. Step 3’s invariant is that one call targets exactly one configured athlete and all athlete-scoped object operations are constrained by that resolved target; this operation only constrains half of the object IDs.

Please add a pre-write event target check for `LinkActivityToEvent` (for example, call `GetEvent(ctx, eventID)` under the same target context and fail before the PUT if it is not found/authorized), and add a regression test proving that a mismatched or missing target event prevents the `PUT /activity/{activityID}` from being sent. The existing matching-target link test should also assert the preflight uses `/athlete/{target}/events/{eventID}`.

## Notes

- `go test ./...` passes.
- I did not find a regression in the R011 advanced-capabilities or coach-mode-off catalog compatibility fixes.
