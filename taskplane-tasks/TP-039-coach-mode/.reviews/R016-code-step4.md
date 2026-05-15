# R016 code review — Step 4: `list_athletes` + `select_athlete`

Verdict: **REVISE**

The basic session-selection path works in the happy-path protocol test, and `go test ./...` passes. However, the implementation still diverges from the Step 4 plan in the catalog-visibility contract: `select_athlete` computes its own ACL-only tool list instead of using the same effective catalog as `tools/list` / registration.

## Findings

1. **`select_athlete.allowed_tools` and `requires_new_conversation` ignore delete-mode/toolset gates.**
   - Location: `internal/tools/select_athlete.go:58-70`; related registration gates at `internal/mcp/server.go:342-350` and active visibility helper at `internal/mcp/server.go:416-421`.
   - `visibleToolsForAthlete` iterates `toolcatalog.AthleteScopedToolNames()` and applies only the coach ACL. It does not know whether a tool was actually registered after `ICUVISOR_DELETE_MODE` and `ICUVISOR_TOOLSET` filtering. As a result, a coach athlete with `allowed_tools: ["*"]` can get `select_athlete.allowed_tools` entries for delete tools while delete mode is safe, and full-only tools while the core toolset is active. Those names are not visible/callable in the effective MCP catalog.
   - The same bug affects `_meta.requires_new_conversation`: it can be `true` solely because hidden full/delete tools differ, even when the visible catalog returned by `tools/list` did not change, or it can otherwise drift from the real registered catalog.
   - This violates the approved Step 4 plan/R015 requirement to filter `tools/list`, advanced capabilities, and `select_athlete.allowed_tools` through one effective active-athlete visibility helper composing delete-mode AND toolset AND coach ACL.
   - Suggested fix: keep `internal/tools` SDK-agnostic, but inject an MCP-owned visible-catalog function or move the `select_athlete` handler assembly to the MCP layer. Compute `allowed_tools` and the conversation-change diff from the registered/effective catalog after capability + toolset + coach filtering, not from `toolcatalog` constants plus ACL only.

2. **Session selections are never cleaned up for Streamable HTTP sessions.**
   - Location: `internal/coach/selection.go:18-49`; selection store is created in `internal/mcp/server.go:94-100` / `internal/app/app.go:278-285` with no corresponding delete path.
   - Every non-empty SDK session ID can add an entry to `SelectionStore.selected`, but there is no removal on session close/timeout and no bounded-cache rationale beyond the process fallback note in `STATUS.md`. A long-running local Streamable HTTP server can accumulate stale per-session athlete selections indefinitely.
   - R014 explicitly asked for a cleanup strategy if feasible, or an explicit bounded/leak rationale in `STATUS.md`. Please either hook session close/timeout to delete the selection key, or document why this is acceptable and bounded for the supported transports.

## Tests run

- `go test ./internal/coach ./internal/tools ./internal/mcp`
- `go test ./...`
