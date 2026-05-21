# R002 Plan Review — Step 2: Move

Verdict: **REQUEST CHANGES**

The step direction is mostly aligned with TP-069, but the current Step 2 plan is too ambiguous around the boundary between **wiring** and **running**. Please tighten the plan before implementing so the refactor does not accidentally change behavior or leave transport concerns in `app.go`.

## Required clarifications / adjustments

1. **Keep `wireServer` construction-only.**
   - The required signature returns `(*mcp.Server, func(), error)`, so `wireServer` should build the logger/capability/client/registries/MCP server and return them.
   - It should **not** call `Run` or `RunStreamableHTTP`; that would conflict with the return shape and keep lifecycle/transport mixed into the wiring function.
   - Move the small runtime wrapper, likely `defaultStartServer`, to `wire.go`: call `wireServer`, `defer cleanup()`, then dispatch to `server.RunStreamableHTTP` or `server.Run` based on `info.Config.Transport`.

2. **Do not leave transport dispatch in `app.go`.**
   - Acceptance says `app.go` should be CLI/dispatch/error mapping only and should no longer import the wiring graph packages.
   - `startServer` can still default to `defaultStartServer`, but `defaultStartServer` itself should live in `wire.go` with the transport dispatch.

3. **Make the dependency type package-private.**
   - Step text says `Deps`, but Step 1 decided on lowercase `deps`; keep it unexported.
   - Prefer a `withDefaults`/`defaultDeps` helper so the zero value is production-safe without scattering nil checks through the wiring path.

4. **Avoid widening scope to make fake concrete clients possible.**
   - `intervals.NewClient` returns a concrete `*intervals.Client`, and the registry constructors currently accept that concrete type. Do not introduce new public interfaces or change registry constructor APIs in this step.
   - If tests need seams, fake the factory functions for registries/server construction rather than refactoring `intervals.Client`.

5. **Preserve current behavior exactly.**
   Ensure the move retains all existing details:
   - trim/normalize empty version to `dev`;
   - log `server starting` and the resolved delete mode/toolset exactly once;
   - warn for non-loopback HTTP binds without logging API key or athlete ID;
   - fallback capability from delete mode when missing;
   - parse delete mode and toolset from the same sources as today;
   - create `coach.SelectionStore` with `info.Config.Coach.DefaultAthleteID`;
   - treat recent-tool-call recorder creation errors as a warning only;
   - pass the same `ResourceOptions`, `RegistryOptions`, and `mcpserver.Options` fields.

6. **Define cleanup semantics explicitly.**
   - There are no obvious closeable resources today, so returning a non-nil no-op cleanup is fine.
   - The wrapper should call cleanup only after successful wiring and should `defer` it around the selected server run.

## Suggested shape

A good Step 2 target would be:

- `internal/app/wire.go`
  - package-private `type deps struct { ... }`
  - `func wireServer(ctx context.Context, info ServerInfo, d deps) (*mcpserver.Server, func(), error)`
  - `func defaultStartServer(ctx context.Context, info ServerInfo) error` as the run wrapper
- `internal/app/app.go`
  - no imports of `coach`, `intervals`, `internal/mcp`, `prompts`, `resources`, `tools`, or `log/slog`
  - existing `Options.StartServer` injection behavior unchanged

Once those boundaries are clarified, the move should be low-risk and behavior-preserving.
