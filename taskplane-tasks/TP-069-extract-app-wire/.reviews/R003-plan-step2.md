# R003 Plan Review — Step 2: Move

Verdict: **APPROVE**

The updated Step 2 plan addresses the blocking concerns from R002 and is aligned with TP-069's acceptance criteria.

## What looks good

- `wireServer` is now explicitly **construction-only** and returns `(*mcp.Server, func(), error)` with a non-nil no-op cleanup for current resources.
- `defaultStartServer` is planned to move to `wire.go` as the runtime wrapper that defers cleanup and dispatches `RunStreamableHTTP` vs `Run`, keeping transport/run concerns out of `app.go`.
- The dependency holder is package-private (`deps`) and has zero-value production defaults, which preserves encapsulation and gives tests seams without exporting DI surface.
- The plan calls out preserving existing logging, safety fallback, parsed mode/toolset, registry options, selection store, and recent-call recorder warning behavior.

## Implementation guardrails

Please keep these details tight during the move:

1. `app.go` should no longer import wiring graph packages such as `coach`, `intervals`, `internal/mcp`, `prompts`, `resources`, `tools`, or `log/slog`. It can keep CLI/config/loading concerns needed by `Run`/`startServer`.
2. Use aliases consistently if needed, e.g. `mcpserver "github.com/ricardocabral/icuvisor/internal/mcp"`, so the `wireServer` return type is the local wrapper `*mcpserver.Server`, not the SDK server type.
3. Keep cleanup semantics simple: return a non-nil no-op cleanup after successful construction, and have `defaultStartServer` `defer cleanup()` only after `wireServer` succeeds.
4. Do not broaden registry/client interfaces in this step. Test seams should come from `deps` factory functions, not from changing public/internal constructor APIs.
5. Update `STATUS.md` as Step 2 completes; remember the task also requires a `CHANGELOG.md` `[Unreleased]` entry before final acceptance.

With those constraints, the Step 2 implementation should be low-risk and behavior-preserving.
