# R015 plan review — Step 5: `_meta` surfacing + docs

Verdict: **APPROVE**

I reviewed `PROMPT.md`, the revised Step 5 section in `STATUS.md`, the prior R014 feedback, and the existing `internal/response`, `internal/app`, README, changelog, and `icuvisor_list_advanced_capabilities` plumbing. The revised plan now pins the response-owned metadata mechanism, the direct-return tool exception, test coverage, and documentation scope sufficiently for implementation.

## What looks good

- The plan extends the existing TP-018 response chokepoint instead of creating per-tool metadata paths: `internal/response` gains process-level `SetToolset`/`Toolset` state, normalized with `safety.ParseToolset`, and `addCommonMeta` owns `_meta.toolset` alongside `_meta.server_version`, `_meta.delete_mode`, and units.
- Startup propagation is correctly scoped: `app.defaultStartServer` should use the already-resolved `ServerInfo.Toolset`/local `toolset` value and must not re-read `ICUVISOR_TOOLSET` in handlers or response code.
- The merge semantics are explicit and correct: caller `_meta` keys such as counts, pagination, scales, and delete-mode notes are preserved, while stale caller-supplied `_meta.toolset` is overwritten with the normalized process value.
- The plan accounts for `icuvisor_list_advanced_capabilities`, which currently bypasses `response.Shape`, and requires text JSON and `StructuredContent` to use the same response-owned `_meta.toolset` source.
- The proposed tests cover the important failure modes called out in R014: default/invalid/empty fallback to `core`, explicit `full`, stale `_meta.toolset` overwrite, startup setter propagation, and advanced-capabilities structured/text metadata in both core and full modes.
- README and CHANGELOG scope is precise and includes the important user-facing caveats: default `core`, `full` opt-in, restart required, invalid/empty fallback, discoverability tool, and orthogonality with `ICUVISOR_DELETE_MODE`.

## Implementation notes

- Because `internal/response` uses process-global state, tests that call `SetToolset` should restore the default with `t.Cleanup`, as existing delete-mode tests do.
- For `icuvisor_list_advanced_capabilities`, keep `current_toolset`/status behavior tied to the active registry toolset, but source `_meta.toolset` from `response.Toolset()` as planned so common response metadata stays centralized.
- When updating exact expected `_meta` maps, check all response-shaper tests that assert common metadata, not just the new toolset-specific cases.

No further plan changes are required before coding Step 5.
