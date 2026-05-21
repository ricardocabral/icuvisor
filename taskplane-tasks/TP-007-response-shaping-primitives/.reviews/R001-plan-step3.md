# Plan Review — TP-007 Step 3

Verdict: **Not approved yet**. `STATUS.md` only repeats the Step 3 checklist; it does not describe a concrete implementation plan for `_meta.server_version` injection or the debug-metadata gate. Because this step is meant to establish a single response chokepoint used by every read tool, the missing design details can easily lead to duplicated per-tool metadata or environment reads scattered through handlers.

## Blocking findings

1. **No single chokepoint is defined.**
   - Step 3 requires `_meta.server_version` to be injected into every response from one place.
   - The current code still has `get_athlete_profile` constructing `GetAthleteProfileMeta.ServerVersion` directly, while `internal/response.Shape` has no server-version option or common metadata merge.
   - The plan must name the API that callers will use and state whether it returns the shaped value only, or also the JSON text used for MCP `Content` so `StructuredContent` and text cannot diverge.

2. **Root `_meta` semantics are not specified for all response shapes.**
   - `Shape` can currently return a single object, a bare array, or a wrapper with row collections.
   - A bare array has nowhere obvious to put `_meta.server_version`; row-level `_meta` from Step 2 is not the same as response-level metadata.
   - The plan must decide whether read tools must return wrapper objects, whether the helper wraps bare arrays, or whether arrays are rejected by the common metadata chokepoint.

3. **Metadata merge and collision rules are missing.**
   - Step 2 already merges row `_meta.fields_present` / `_meta.missing_fields` with existing `_meta` values.
   - Step 3 must define how common metadata merges with existing top-level `_meta`: preserve non-conflicting values, overwrite/reserve `server_version`, and avoid accidentally adding response-level fields to every row unless that is intentional.
   - The plan should also cover behavior when an input has `_meta` with a non-object value.

4. **Startup debug configuration path is undefined.**
   - The requirement says `ICUVISOR_DEBUG_METADATA` is read once at startup, not per call.
   - The plan does not say where the env var is read (`app.Run`/server startup/config), how the boolean is carried through `ServerInfo`, `mcp.Options`, `tools.NewRegistry`, and response options, or how tests inject the value without mutating process env during handler calls.
   - Avoid a package-level `init`/global env read; it is hard to test and does not match “startup” for this app.

5. **Accepted env values and invalid-value behavior are not defined.**
   - The plan should state the parser semantics. At minimum, `true` enables debug metadata and unset/empty/invalid values quietly disable it.
   - If using `strconv.ParseBool`, note that values such as `1`, `t`, and `TRUE` will also enable it; if that is not desired, use an exact case-insensitive `true` check.

6. **`fetched_at` and `query_type` value sources are unspecified.**
   - The plan must define what `query_type` contains (tool name? a caller-supplied query classification?) and where it is supplied.
   - `fetched_at` should be generated at response construction time only when debug metadata is enabled, preferably via an injectable clock in tests; otherwise tests will be nondeterministic.
   - It should also specify the timestamp format/timezone (UTC RFC3339/RFC3339Nano would align with project time conventions).

7. **Interim interaction with `get_athlete_profile` is unclear.**
   - Step 6 is scheduled to refactor `get_athlete_profile` onto the new helpers, but Step 3 already requires server-version injection into every response.
   - The plan must explain whether Step 3 will update the existing handler to use the new common helper now, or only add the helper and defer adoption. If adoption is deferred, the step does not satisfy its own checklist.

## Required additions before approval

Please update `STATUS.md` with a real Step 3 plan covering at least:

- Proposed function/types in `internal/response` for common metadata injection, including parameters for server version, debug-enabled flag, query type/tool name, and clock/fetched time.
- The exact response-level `_meta` placement for single objects, wrappers with row collections, and bare arrays.
- Merge/collision behavior for existing top-level `_meta` and row-level `_meta`.
- The startup env parsing location and propagation path through app/server/registry/tool construction.
- Accepted `ICUVISOR_DEBUG_METADATA` values and quiet invalid-value handling.
- The exact shape/format of debug metadata (`fetched_at`, `query_type`) and tests for enabled, disabled, and invalid env values.
- How `get_athlete_profile` will avoid a duplicate tool-specific `ServerVersion` source once common metadata exists.
- Whether `README.md` or client docs need a short note for the new user-visible debug env var.

Once those decisions are recorded, Step 3 should be straightforward to implement and review.
