# TP-069 — Extract `internal/app/wire.go` from `defaultStartServer` (audit God-package)

## Mission

`internal/app/app.go:262-325` `defaultStartServer` is the entire wiring graph for icuvisor:

- builds `*intervals.Client`
- builds the coach `SelectionStore`
- builds the recent-call recorder
- builds `resources.Registry`, `prompts.Registry`, `tools.Registry`
- builds the MCP `Server`

That single function ties every internal package to `internal/app`, making `internal/app` a god package (CLI parsing + dispatch + exit codes + the wiring graph). Adding a new cross-cutting concern (e.g., audit logger, rate-limit middleware, feature-flag wiring) requires touching `defaultStartServer`.

Goal: extract the wiring into `internal/app/wire.go` (a sibling file, **same package** — exporting it would over-couple consumers). `internal/app/app.go` keeps only `Run`, `RunCLI`, `UsageError`, `ExitCode`. `wire.go` exposes a single `func wireServer(ctx, cfg, deps) (*mcp.Server, func(), error)` (cleanup func for resource shutdown).

Optionally: introduce a tiny `Deps` struct with the half-dozen explicit collaborators, so tests can swap them.

Audit ref: 2026-05-16 Go audit, God-package section.

PRD anchors: §7.1 architecture overview.
CLAUDE.md hard rules: narrow interfaces; small packages.

Complexity: Blast radius 2 (`internal/app` internal; tests follow), Pattern novelty 2 (introducing `Deps` is a small architectural step), Security 1, Reversibility 1 = 6 → Review Level 1. Size: S.

## Dependencies

- **TP-042** — soft. TP-042 cleans up registry interface assertions; after it lands, `wireServer` is simpler (no `any` plumbing).
- **TP-064** — soft. TP-064 splits `mcp/server.go` and lifts coach. After it lands, `wireServer` calls cleaner constructors.
- **TP-068** — sequence after. TP-068 splits `setup.go`; sequencing this after avoids merge conflicts on `internal/app/`.

## Context to Read First

- `internal/app/app.go` — full file.
- `internal/app/app_test.go` — locked-in dispatch tests.
- `internal/coach/`, `internal/intervals/`, `internal/resources/`, `internal/prompts/`, `internal/tools/`, `internal/mcp/` — to know what `defaultStartServer` currently builds.

## File Scope

- New: `internal/app/wire.go` — package-private `wireServer` + (optional) `Deps`.
- `internal/app/app.go` — trim to CLI dispatch + `Run` + `RunCLI` + `UsageError` + `ExitCode`. Calls into `wireServer`.
- Tests: add `wire_test.go` if the new function deserves dedicated coverage. Otherwise rely on existing `app_test.go`.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Public exports — `wireServer` and `Deps` stay package-private.
- Changing the wiring graph (same collaborators, same order).
- Changing CLI dispatch / exit codes.
- Adding new collaborators.

## Steps

### Step 1: Inventory the wiring graph
- [ ] List every collaborator `defaultStartServer` builds, in order. Record in `STATUS.md`.
- [ ] Decide whether a `Deps` struct is worth introducing (recommended: yes if ≥ 4 collaborators).

### Step 2: Move
- [ ] Create `wire.go` with `wireServer(ctx, cfg, deps Deps) (*mcp.Server, func(), error)`. Cleanup func returns any resource shutdown work.
- [ ] Move `defaultStartServer`'s body into `wireServer`.
- [ ] `app.go` calls `wireServer` and handles exit on error.

### Step 3: Tests
- [ ] If a `Deps` struct exists, write a test for `wireServer` that swaps in fakes for at least two collaborators (e.g., a fake `intervals.Client`) and asserts the returned `*mcp.Server` is configured as expected.
- [ ] Existing `app_test.go` still passes unchanged.

### Step 4: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] `wc -l internal/app/app.go` — ≤ ~200 LOC.
- [ ] Commit: `TP-069 extract app wireServer`.

## Acceptance Criteria

- `internal/app/wire.go` exists and holds the wiring graph.
- `internal/app/app.go` is ≤ ~200 LOC and contains only CLI/dispatch/error mapping.
- `wireServer` and `Deps` are package-private.
- Behaviour unchanged.
- All `make` checks pass.

## Do NOT

- Do not export `wireServer` or `Deps`.
- Do not change CLI flags, subcommands, exit codes, or dispatch behaviour.
- Do not add new collaborators / dependencies.
- Do not introduce a DI framework.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).

## Git Commit Convention

Conventional Commits, prefixed `TP-069`.

---

## Amendments

_Add amendments below this line only._
