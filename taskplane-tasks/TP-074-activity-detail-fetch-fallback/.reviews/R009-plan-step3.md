# Plan Review — Step 3: Mirror the shape across the four tools

Decision: **approve plan**.

The revised Step 3 plan addresses the blockers from R008. It now explicitly splits the work into shared classification vs. optional Strava-block detection, adds intentional detail-client wiring for streams/splits instead of relying on the current test fake's extra methods, preserves success-path shapes while avoiding fabricated collections on unavailable responses, and keeps `get_extended_metrics` from refetching `GetActivity` after a primary activity fetch failure.

## Implementation guardrails

- Keep the shared helper(s) small: classify from the original endpoint error, and make Strava detection an optional layer used only for endpoint-specific failures (`intervals`, `streams`, and the terminal streams fallback in `splits`).
- Preserve cancellation semantics: if the original call or fallback detail lookup is `context.Canceled` / `context.DeadlineExceeded`, return that error rather than encoding `unavailable`.
- For `get_activity_splits`, continue treating the initial manual-intervals lookup as optional. Only the terminal streams fallback failure should produce the structured unavailable payload.
- For `get_activity_streams` and `get_activity_splits`, prefer separate unavailable response structs or otherwise targeted encoding over broad `omitempty` changes that could alter successful empty-collection responses.
- Update constructor signatures, `registry.go`, and tests together when adding the explicit `ActivityDetailsClient` dependency for streams/splits.

With those constraints, the plan is sufficiently specific for Step 3 and should align the four tools with the TP-074 acceptance criteria.
