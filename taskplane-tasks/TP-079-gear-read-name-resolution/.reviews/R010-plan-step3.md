# Plan Review — TP-079 Step 3

**Verdict:** APPROVE

The Step 3 plan is now specific enough to implement safely. It addresses the prior blockers from R008/R009 by pinning the dependency path through the registry-scoped `gearListCache`, keeping activity tools in the core toolset, using `refresh=false` lookups, avoiding gear fetches when no activity has a `gear_id`, and covering cache reuse/isolation and pagination field expectations.

Most importantly, the response contract is now explicit: activity list rows and details should use `gear_id`, `gear_name`, and a shared `gear_resolution` enum with `resolved`, `name_missing`, `unresolved`, and `lookup_unavailable`. The plan also correctly separates context cancellation/deadline errors from ordinary gear-list lookup failures, and includes output schema updates as part of this implementation step rather than deferring the tool contract to generated docs.

## Implementation notes

- Add `gear_id` to `terseActivityFields` before updating pagination token goldens so default `get_activities` rows can resolve gear without `include_full=true`.
- Route both `newGetActivitiesTool` and `newGetActivityDetailsTool` through `registryBaseTools`/`registryToolOptions` with the existing cache and `GearListClient`; avoid package globals or a second cache.
- Treat non-context gear lookup failures, including unavailable lookup dependencies at runtime, as `gear_resolution=lookup_unavailable` on otherwise successful activity reads.
- Keep `gear_resolution` absent or otherwise unambiguous for rows/details with no upstream `gear_id`; do not force a gear-list fetch just to prove absence.
- Ensure tests cover Strava-unavailable rows/details with upstream `gear_id` so gear fields are still surfaced without fabricating other metrics.

No further plan revisions are required before implementation.
