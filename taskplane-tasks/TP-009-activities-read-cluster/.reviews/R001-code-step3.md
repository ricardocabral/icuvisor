# Code Review — TP-009 Step 3 (`get_activity_details`, `get_activity_intervals`)

Decision: **Approve**

## Findings

No blocking findings for the Step 3 implementation.

The current code addresses the prior Step 3 concerns: `get_activity_intervals` now returns the structured Strava unavailable marker for both hidden/stub success payloads and confirmed not-found/forbidden fallbacks, and `include_full:true` now exposes the top-level raw `IntervalsDTO` payload in addition to row-level raw interval/group payloads.

## Non-blocking notes

- `STATUS.md` still has review-log rows under `## Blockers` after `_None._`. This does not block Step 3 code approval, but it should be cleaned up before task wrap-up so the blockers section remains unambiguous.

## Validation run

- `git diff 36db5a23b213efbc544add64d0b0579832a6d0d0..HEAD --name-only`
- `git diff 36db5a23b213efbc544add64d0b0579832a6d0d0..HEAD`
- `git diff --check 36db5a23b213efbc544add64d0b0579832a6d0d0..HEAD` — passed
- `go test ./internal/intervals ./internal/tools ./internal/app` — passed
- `go test ./...` — passed
