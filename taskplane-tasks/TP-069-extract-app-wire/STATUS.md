# TP-069-extract-app-wire — Status

**Current Step:** Step 4: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S

---

### Step 1: Inventory the wiring graph

**Status:** ✅ Complete

- [x] List every collaborator `defaultStartServer` builds, in order. Record in `STATUS.md`.
  - Inventory (current build order): `slog.Default()` logger; normalized version; `safety.Capability` fallback; parsed delete mode; parsed toolset; `intervals.NewClient`; `coach.NewSelectionStore`; `defaultRecentToolCallRecorder`; `prompts.NewRegistry`; `resources.NewRegistryWithOptions`; `tools.NewRegistryWithOptions`; `mcpserver.NewServer`; transport runner (`RunStreamableHTTP` or `Run`).
- [x] Decide whether a `Deps` struct is worth introducing (recommended: yes if ≥ 4 collaborators).
  - Decision: introduce package-private `deps` (lowercase) because wiring has more than four construction seams. Keep zero-value defaults for production and function fields for tests around client, registries, selection store, recorder, server construction, logger, and transport runner.

### Step 2: Move

**Status:** ✅ Complete

- [x] Create `wire.go` with package-private `deps` and construction-only `wireServer(ctx, info, deps) (*mcp.Server, func(), error)` returning a non-nil no-op cleanup for current resources.
- [x] Move the wiring graph into `wireServer` while preserving existing logging, fallback, options, registry, and recorder behavior exactly.
- [x] Move `defaultStartServer` to `wire.go` as the runtime wrapper that defers cleanup and dispatches `RunStreamableHTTP`/`Run`; keep `app.go` CLI/dispatch-only.

### Step 3: Tests

**Status:** ✅ Complete

- [x] Add `wireServer` coverage that swaps in fakes for at least two collaborators and asserts the constructed server options.
- [x] Existing `app_test.go` still passes unchanged. Verified with `go test ./internal/app`.

### Step 4: Verify

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` `[Unreleased]` with the internal wiring refactor.
- [x] Run `make build`, `make test`, `make test-race`, and `make lint`. All passed; rerun after final app trim also passed.
- [x] Confirm `wc -l internal/app/app.go` is ≤ ~200 LOC. Verified 170 LOC.
- [x] Commit `TP-069 extract app wireServer`. Committed verification and final trim in `chore(TP-069): verify app wire extraction`.

| 2026-05-17 11:22 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 11:22 | Step 1 started | Inventory the wiring graph |
| 2026-05-17 11:25 | Review R001 | plan Step 1: APPROVE |
| 2026-05-17 11:28 | Review R002 | plan Step 2: UNKNOWN |
| 2026-05-17 11:29 | Review R003 | plan Step 2: APPROVE |
| 2026-05-17 11:35 | Review R004 | plan Step 3: APPROVE |

| 2026-05-17 11:41 | Worker iter 1 | done in 1121s, tools: 92 |
| 2026-05-17 11:41 | Task complete | .DONE created |