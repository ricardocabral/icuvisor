# Plan Review R016 — Step 5: `icuvisor://athlete-profile`

**Verdict: APPROVE**

I read `PROMPT.md`, the current `STATUS.md`, the existing MCP resource plumbing, and the current `get_athlete_profile` implementation. The Step 5 plan is specific enough to proceed.

## What is satisfactory

- The plan correctly makes response shaping a shared implementation used by both `get_athlete_profile` and the resource, which is the right way to satisfy the unit/timezone/`_meta` parity requirement.
- The resource contract is pinned: `icuvisor://athlete-profile`, `athlete_profile`, human title, `application/json`, and a single text resource content item containing the default terse profile (`include_full=false`).
- The refresh/staleness policy is documented in `STATUS.md`: per-resource-instance cache, 15-minute TTL, no background polling, no retry loop, failed refreshes surface a short safe error, and cached reads avoid upstream calls.
- The plan routes the real intervals client and response-shaping options from `internal/app` through a resource registry option path, while preserving a static-resource path for tests that intentionally do not configure a profile client.
- It explicitly covers context cancellation before/while refreshing and includes tests for shape parity, cache hit/expiry behavior, cancellation, and list/read protocol coverage for all four resources.

## Implementation notes to preserve the approval

1. **Make the cache concurrency guarantee testable.** The plan says concurrent reads share the same cache; implement that as a real single refresh on a cold/expired cache, not just a mutex around the cached value after separate callers have already started upstream requests. Add a focused test where concurrent reads on an expired cache produce one `GetAthleteProfile` call and all callers receive the same shaped JSON.

2. **Keep registry semantics clear.** Normal server startup must advertise all four resources. If `resources.NewRegistry()` remains static-only for existing tests, consider naming/documenting that path clearly (or adding a `NewStaticRegistry`) so future callers do not assume it is the production default. The app path should use the configured-client registry and protocol tests should assert `resources/list` includes `icuvisor://athlete-profile`.

3. **Account for existing resource error handling.** `internal/mcp` currently converts ordinary resource handler failures to the generic `resource read failed; try again or check icuvisor logs` message. If the implementation wants the resource-specific profile-fetch message described in the plan, add an explicit safe/public resource-error mechanism or otherwise update the plan/tests to expect the generic resource error. Do not leak upstream error details.

4. **Prefer an injected clock/TTL for tests.** Avoid slow or flaky TTL tests by making the cache clock and/or TTL configurable in resource options while keeping the production default at 15 minutes.

5. **Avoid package cycles and semantic drift in the shaper extraction.** A small domain package such as `internal/athleteprofile` (or equivalent) is preferable to having `internal/resources` import `internal/tools` just to reuse response structs. Tool output schemas/tests should continue proving the same JSON fields, and resource tests should compare against the same shaper rather than duplicating expected transformation logic.

With those cautions, the plan satisfies Step 5 requirements and can move to implementation.
