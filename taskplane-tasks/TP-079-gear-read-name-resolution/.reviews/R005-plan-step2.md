# Plan Review — TP-079 Step 2

**Verdict:** Approve

The revised Step 2 plan is now specific enough to implement safely. It addresses the prior blockers around cache scope, refresh semantics, concurrency, response shape, and catalog/toolset placement while keeping `delete_gear`'s destructive gating out of scope.

## What is now covered

- The gear cache is explicitly shared/registry-scoped rather than package-global, and is intended to be reused by Step 3 activity name resolution.
- Cache entries are per-athlete, keyed by the resolved target athlete, which is the important coach-mode isolation boundary.
- Refresh behavior is pinned as manual-refresh-only, with failed/canceled refreshes forbidden from replacing a previously good entry.
- Concurrent MCP calls are accounted for by requiring a concurrency-safe cache and cancellation-aware fetches.
- The `get_gear_list` API shape is constrained: `refresh`, `include_full`, terse gear fields, explicit unnamed-gear signaling, and `_meta` cache/count fields.
- Registration coverage now includes the shared `internal/toolcatalog` catalog, full-toolset/read-only visibility, hidden-from-core behavior, and independence from delete mode.

## Non-blocking implementation notes

- Add an explicit cache isolation test with two different target athletes in the same registry/cache instance. This is implied by the keying requirement, but it is important enough to pin directly in tests.
- For the single-athlete fallback key, keep it deterministic and distinct from coach-mode target IDs if there is no `TargetAthleteIDFromContext(ctx)` value.
- Treat the first successful fetch as `refreshed`/not cached in `_meta`; subsequent non-refresh hits should report cached state clearly.
- If `include_full` exposes raw upstream gear payloads, keep raw fields under a `full` object per row rather than mixing unknown upstream keys into the terse row.

With those notes, the plan is ready for implementation.
