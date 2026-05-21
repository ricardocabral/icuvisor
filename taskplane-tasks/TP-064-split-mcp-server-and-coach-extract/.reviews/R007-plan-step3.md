# R007 Plan Review — Step 3: Lift coach concerns to `internal/coach`

## Decision: APPROVE

The revised Step 3 plan addresses the R006 blockers. It now covers both catalog/list visibility and request-time target authorization, preserves the acyclic package boundary (`coach` must not import `tools` or `config`), pins the current gate order, and adds concrete regression tests for the coach filter, target authorization, and filtered advanced-capabilities behavior.

## What looks good

- The proposed `coach.ToolFilter` seam keeps coach policy in `internal/coach` while leaving MCP SDK/session/raw-JSON adaptation in `internal/mcp`, which is the right boundary.
- The plan explicitly preserves the important gate composition: capability gate, coach any-athlete gate for the advanced-capabilities source catalog, toolset gate, then selected-athlete filtering for `tools/list` and advanced-capabilities rows.
- Request-time authorization is now included without introducing a `coach -> config` cycle by keeping normalization in `mcp` or injecting a normalizer callback.
- Moving advanced-capabilities rendering to `internal/tools` via a filtered helper avoids an `internal/coach -> internal/tools` import cycle and places response-shaping logic with the base tool implementation.
- The planned tests cover the new coach-owned policy surface and keep the existing protocol tests as the wire-behavior regression gate.

## Implementation cautions

- Keep the filtered advanced-capabilities output byte-compatible with the current coach-mode wire response. The plan mentions reusing the base helper, but the existing `mcp` duplicate and `tools` implementation differ in summary extraction behavior. If the implementation intentionally switches to the canonical `tools` formatting, make sure the compatibility tests and review clearly call out whether this is a behavior-preserving refactor or a deliberate wire-output change.
- Update both `NewServer` and `ComputeToolCatalogHash` registrar construction to initialize/use the same `ToolFilter` seam; otherwise catalog hash/schema snapshots can diverge from the live server path.
- Ensure `VisibleToolNamesForAthlete` receives the post-capability/post-toolset registered tool names for `select_athlete` metadata, not the pre-toolset advanced-capabilities source catalog.
- If `ToolFilter` or target authorization types/functions are exported from `internal/coach`, add Go doc notes and the short `docs/coach-mode.md` note requested by the task prompt.

With those cautions observed during implementation, the Step 3 plan is ready to proceed.
