# Plan Review — TP-079 Step 3

**Verdict:** Revise

The current Step 3 plan captures the desired user-visible outcome, but it is still too high-level for the risky part of this change: activity reads are core, paginated, coach-scoped tools, and name resolution will introduce a second upstream read path through the shared gear cache. Please tighten the plan before implementation.

## Blocking revisions

1. **Define how activity tools get gear resolution dependencies.**
   - Reuse the registry-scoped `gearListCache` added in Step 2; do not add a package-global cache or a second cache.
   - Pass the cache and a `GearListClient` into `newGetActivitiesTool` and `newGetActivityDetailsTool` via `registryBaseTools`/`registryToolOptions`.
   - Keep `get_activities` and `get_activity_details` in the core toolset. This internal read-side lookup should not change tool registration or delete gating.

2. **Pin fetch/cache semantics for activity reads.**
   - Activity reads should use the existing cache with `refresh=false`; do not add a new model-controlled refresh argument to activity reads.
   - Avoid the gear-list fetch entirely when the returned activity page/detail has no `gear_id` values.
   - On a cache miss with one or more `gear_id` values, fetch the gear list once for the resolved athlete and populate the cache through the shared cache API.
   - Use the same cache keying as `get_gear_list` (`TargetAthleteIDFromContext(ctx)` with the existing fallback) so coach-mode athletes cannot share gear names.

3. **Define failure and unresolved behavior.**
   - Gear lookup failures must not make `get_activities` or `get_activity_details` fail if the activity read itself succeeded. Return the activity with `gear_id` and an explicit unresolved signal instead.
   - Distinguish the important cases tersely:
     - matched gear with non-empty name: `gear_id` and `gear_name`;
     - matched gear with empty/missing name: `gear_id` plus an explicit unnamed/missing-name signal;
     - no matching gear ID in the list: `gear_id` plus an explicit unresolved signal;
     - gear-list fetch failed/canceled after the activity read: preserve cancellation behavior for context cancellation, otherwise return the activity and mark resolution unavailable.
   - Do not guess names from `device_name`, sport type, brand/model, or raw payload fragments.

4. **Ensure `gear_id` is actually requested for terse list reads.**
   - Add `gear_id` to `terseActivityFields`; otherwise `get_activities` may not receive the upstream field when `include_full=false`.
   - Update the pagination token golden expectations affected by the changed `fields` list.
   - Keep `include_full=true` behavior preserving raw payloads under `full` without requiring full mode to see `gear_id`/`gear_name`.

5. **Pin response shape and metadata.**
   - Add typed fields to `getActivitiesRow` for the terse output, e.g. `gear_id`, `gear_name`, and a concise explicit missing/unresolved marker.
   - Decide whether `_meta` needs a small gear-resolution summary (`resolved_count`, `unresolved_count`, or lookup status). If not, the row-level signal must be sufficient and tested.
   - Update output schema descriptions so clients/LLMs know that unresolved IDs are not guesses.
   - Preserve Strava-unavailable behavior: if upstream exposes `gear_id` on such a row/detail, still surface the ID and explicit name status, but do not fabricate other metrics.

6. **Add targeted coverage for the new contract.**
   - `get_activities` terse default includes `gear_id` and resolved `gear_name` without `include_full`.
   - `get_activity_details` includes the same fields.
   - Rows with no `gear_id` do not trigger a gear-list fetch.
   - Unknown gear IDs and unnamed gear entries are explicit and not guessed.
   - Gear-list fetch errors do not fail an otherwise successful activity read.
   - Cache reuse is visible between `get_gear_list` and activity reads within one registry/cache instance, and cache isolation by target athlete is preserved.
   - Pagination/token tests are updated for the added `gear_id` terse field.

## Good parts of the current plan

- It correctly requires both list and detail activity reads to expose the same gear fields.
- It keeps the central product rule in view: callers should not have to chain through `get_gear_list` to understand activity gear.
- It explicitly calls out that unresolved IDs must remain explicit rather than guessed.

Once the dependency wiring, fetch/failure semantics, response shape, and targeted tests are spelled out, this step should be safe to implement.
