# Plan Review — TP-079 Step 3

**Verdict:** Revise

The revised Step 3 checklist now covers most of the dependency and cache semantics requested in R008: it explicitly calls for registry-scoped cache wiring, `refresh=false`, no fetch when activity rows have no `gear_id`, shared-cache/athlete-isolation tests, and pagination field expectation updates. That is a meaningful improvement.

However, two API-contract details are still not pinned tightly enough for implementation. Activity read response shape is part of the MCP contract, and this step changes core tools, so the plan should be explicit before code starts.

## Blocking revisions

1. **Pin the exact row-level unresolved/unnamed/failure shape.**
   - The plan names `gear_id` and `gear_name`, but still only says unresolved/unnamed/failure cases must be "explicit".
   - Choose the actual typed fields/status values now, e.g. `gear_name_missing` plus `gear_resolution_status`/`gear_resolution_error`, or a single concise `gear_resolution` enum/reason. The important part is that matched-empty-name, unknown ID, and lookup-unavailable are distinguishable without guessing from absence of `gear_name`.
   - Apply the same shape to both `getActivitiesRow` and `get_activity_details` output.

2. **Separate context cancellation from non-context gear lookup failures.**
   - The checklist currently says gear-list lookup failures should not fail successful activity reads. That is correct for ordinary upstream/cache lookup errors, but the plan should explicitly preserve existing cancellation behavior: if the gear resolution fetch returns `context.Canceled`, `context.DeadlineExceeded`, or `ctx.Err()` is set, return the context error rather than a partial activity response.
   - For non-context gear-list errors after the activity read succeeded, return the activity with `gear_id` and the chosen lookup-unavailable marker.

3. **Add output schema update to Step 3.**
   - R008 asked for schema descriptions so clients/LLMs know unresolved IDs are not guessed. The current checklist does not explicitly include updating `getActivitiesOutputSchema` / `activityReadOutputSchema` descriptions for `gear_id`, `gear_name`, and the chosen unresolved/name-missing markers.
   - This is not merely Step 4 generated docs; it is part of the tool contract changed by Step 3.

## Non-blocking implementation notes

- The cache-wiring bullet should result in passing the existing `opts.gearCache` and `client` as `GearListClient` into `newGetActivitiesTool` and `newGetActivityDetailsTool`; avoid a package global or a second cache.
- Keep `get_activities` and `get_activity_details` as core tools, and keep `get_gear_list` as full-toolset/read-only.
- Add `gear_id` to `terseActivityFields` before updating cursor/token goldens, so terse default rows can resolve gear without `include_full=true`.

Once the response shape/schema and cancellation semantics are pinned in the plan, the rest of the Step 3 plan should be ready to implement.
