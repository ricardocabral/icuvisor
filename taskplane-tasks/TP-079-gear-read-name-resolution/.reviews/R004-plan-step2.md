# Plan Review — TP-079 Step 2

**Verdict:** Revise

The Step 2 checklist covers the broad outcomes, but it is not yet precise enough for the risky part of this step: adding a shared per-athlete cache that will later be used by activity reads. Please tighten the plan before implementation so the cache cannot leak across coach-mode athletes and the tool registration lands on the right catalog surfaces.

## Blocking revisions

1. **Define cache ownership and keying.**
   - Use a single registry-scoped/shared gear cache that can be passed to both `get_gear_list` now and activity name-resolution in Step 3.
   - Do not use a package-global cache.
   - Key cache entries by the normalized resolved athlete target (`intervals.TargetAthleteIDFromContext(ctx)` in coach-mode calls, with a safe single-athlete fallback when no target is present). Add a test that two different target athletes do not share cached gear.

2. **Define refresh/staleness semantics.**
   - The plan says `refresh` bypasses stale entries, but it does not define what makes an entry stale.
   - Either make the cache manual-refresh-only, or specify a TTL and use an injectable clock in tests. Avoid `time.Sleep`-based tests.
   - Specify failure behavior: a failed refresh should not silently replace a previously good cache entry with an empty/partial result.

3. **Make the cache concurrency-safe.**
   - MCP handlers can be invoked concurrently, so any map-backed cache needs synchronization.
   - The plan should require cancellation-aware fetches and should not cache failed/canceled calls.

4. **Pin the tool API and response shape.**
   - The request schema should at minimum include `refresh: boolean` with a clear default/description. If raw upstream gear fields are exposed, gate them behind `include_full: true`.
   - Terse rows should use disambiguated fields such as `gear_id`, `name`, `type`, `brand`, `model`, and `retired`.
   - Missing gear names must be explicit rather than guessed or silently indistinguishable from absent data, e.g. via `name_missing: true`, an `_meta.unnamed_count`, or an equivalent terse signal.
   - `_meta` should include useful cache state such as `count`, `cached`/`refreshed`, and `include_full` if applicable.

5. **Call out all registration/catalog surfaces.**
   - Add `get_gear_list` to `internal/toolcatalog/catalog.go`; otherwise `Registry.Register` will reject it as an unknown tool.
   - Register it through `registryBaseTools` and classify it in `toolCatalogGroup` (likely `settings`).
   - Place it in the **full** toolset as a read-only tool (`RequirementRead`) per the roadmap/PRD note that gear tools belong behind `ICUVISOR_TOOLSET=full`; do not alter `delete_gear`'s `RequirementDelete` gating.
   - Add catalog/visibility coverage so `get_gear_list` is hidden from `core`, visible in `full`, and independent of delete-mode.

## Good parts of the current plan

- It correctly keeps `delete_gear` safety gating out of scope.
- It adds explicit empty-list and cached-vs-refreshed tests.
- It anticipates Step 3 by requiring the cache to be reusable by activity name resolution.

Once the cache semantics, coach-mode keying, response shape, and catalog/toolset placement are explicit, this step should be safe to implement.
