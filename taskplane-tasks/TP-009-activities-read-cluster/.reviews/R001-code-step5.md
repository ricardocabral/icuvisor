# Code Review — TP-009 Step 5

Decision: **Changes requested**

## Summary

The previous Step 5 issues around `seen:false`/`deleted:false`, fallback cancellation, negative `since_id`, and registration have been addressed, and the focused tests pass. I found one remaining PRD-level contract issue before this can be approved: message author athlete IDs are emitted exactly as upstream returns them instead of using the project’s canonical `i12345` display format.

## Blocking issues

1. **Message row `athlete_id` is not normalized before being emitted**  
   `internal/tools/get_activity_messages.go:42-45`, `internal/tools/get_activity_messages.go:106-110`

   PRD §7.2.D and the repository MCP conventions require athlete IDs to be emitted consistently in canonical `i12345` form. `shapeActivityMessages` currently copies `message.AthleteID` directly into the public row:

   ```go
   row := activityMessageRow{ID: message.ID, AthleteID: message.AthleteID, ...}
   ```

   If intervals.icu returns `"athlete_id":"12345"`, `get_activity_messages` will expose `"athlete_id":"12345"`, making this tool inconsistent with `get_athlete_profile` and the rest of the public API contract. Please normalize this through `internal/config.NormalizeAthleteIDForDisplay` (or an existing centralized helper) before assigning the terse row field. Add a regression test with a message fixture containing `"athlete_id":"12345"` and assert that the response row contains `"athlete_id":"i12345"`.

## Non-blocking issues

- `internal/tools/get_activity_messages.go:99` returns the generic `fetchActivityDetailsMessage` for message-fetch failures, so the LLM sees “could not fetch activity details” even though this tool was fetching messages. A dedicated `fetchActivityMessagesMessage` would be clearer and more actionable.
- `internal/tools/get_activity_messages.go:157-158` declares `limit.minimum = 1`, but the handler treats an explicit `"limit":0` as “use default”. Either reject explicit zero or make the schema/description match the implementation. This is minor because normal omitted `limit` calls work.
- `_meta.since_id` is tagged `omitempty` in `activityReadMeta`, so an effective no-cursor value of `0` is not exposed despite the status note saying effective `since_id` is included. This is likely acceptable, but align the status/contract if the field is intentionally omitted when zero.

## Verification

- Ran: `go test ./internal/intervals ./internal/tools ./internal/app` — pass.
