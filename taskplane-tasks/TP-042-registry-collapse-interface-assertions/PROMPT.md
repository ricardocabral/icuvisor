# TP-042 — Collapse `internal/tools/registry.go` interface-assertion sprawl (audit P0)

## Mission

`internal/tools/registry.go:62-263` `defaultRegistry.Register` is a ~200-line god method built around runtime type assertions: a single `r.profileClient any` field is asserted against ~30 per-tool `XxxClient` interfaces (one defined in each `internal/tools/<tool>.go`), all satisfied by exactly one concrete production type (`*intervals.Client`). This is a textbook one-impl interface explosion that doesn't buy testability (each tool already takes its narrow interface in its own constructor) and forces a parallel stub fan-out in `internal/toolchecks/schema_stability.go:304-408` (`schemaCatalogClient` exists only to satisfy the assertion chain).

Goal: keep the **narrow per-tool interfaces** (real unit-test seam) but stop routing them through a generic `any` + assertion chain in the registry. Accept a typed dependency (either `*intervals.Client` directly, or a small `Deps` struct that holds it) and call each tool's constructor unconditionally with that typed value. `*intervals.Client` already implements every `XxxClient`, so this is a mechanical change.

This was identified in the 2026-05-15 Go audit as the top structural issue in the codebase.

PRD anchors: §7.2.C (catalog stability — must not regress registered tool names), §7.4 #6 (clear safety-mode signaling — delete-mode gating in registry must not regress).

ROADMAP positioning: maintenance / debt paydown. Independent of any version milestone. Land before v0.5 dogfood (TP-041) so registry behaviour is settled when external users hit it.

Complexity: Blast radius 2 (touches every tool's wiring), Pattern novelty 1 (standard Go), Security 1 (no new credential paths), Reversibility 2 (large diff, easy to revert) = 6 → Review Level 2. Size: M.

## Dependencies

- None blocking. The change is internal refactor; behaviour must be preserved bit-for-bit.

## Context to Read First

- `CLAUDE.md` — "Default to `internal/`", error wrapping, no panic outside main, table-driven tests.
- `internal/tools/registry.go` — the file under refactor.
- `internal/tools/get_athlete_profile.go`, `get_activities.go`, `get_fitness.go` — representative tool files; each declares a narrow `XxxClient` interface and a `newXxxTool(client XxxClient) Tool` constructor.
- `internal/intervals/client.go` — the concrete `*Client` type that implements every tool interface.
- `internal/toolchecks/schema_stability.go:304-408` — `schemaCatalogClient`; most of it disappears once the registry stops asking for `any`.
- `internal/app/app.go` — registry construction call site.

## File Scope

- `internal/tools/registry.go` — rewrite `Register` to accept typed deps and call constructors directly. Remove the `profileClient any` field and the assertion chain.
- `internal/toolchecks/schema_stability.go` — delete `schemaCatalogClient` stub methods that are no longer needed; use `*intervals.Client` (or a fake that implements just the few interfaces actually exercised).
- `internal/toolchecks/schema_stability_test.go` — adjust fixtures.
- `internal/app/app.go` — adjust call site.
- `internal/tools/registry_test.go` (if present) — update tests.
- Per-tool files in `internal/tools/`: **do not** remove the narrow interfaces — they are still the test seam.

Out of scope:
- Renaming or restructuring tool constructors.
- Touching the `Tool` interface itself.
- Changing tool-name or schema surface (CI snapshot tests will catch any drift).

## Steps

### Step 1: Map the current assertion chain

- [ ] Enumerate every `XxxClient` interface in `internal/tools/`. Verify all are satisfied by `*intervals.Client`.
- [ ] Identify the actual unit-test fakes in `*_test.go` — those must keep working unchanged.
- [ ] Decide between (a) passing `*intervals.Client` directly to `Register`, or (b) a `Deps` struct (`type Deps struct { Client *intervals.Client; ... }`). Default: (a), unless other deps (logger, clock) make a struct cleaner.

### Step 2: Refactor `Register`

- [ ] Change `Register`'s signature to take the typed dep.
- [ ] Replace each `if client, ok := r.profileClient.(XxxClient); ok { … }` block with a direct constructor call.
- [ ] Preserve **all** existing registration-time gating: delete-mode (`internal/safety`), toolset tier (`ICUVISOR_TOOLSET`), capability requirement. The gating logic must not move or change semantics.
- [ ] Preserve the existing error path for tool wiring failures.
- [ ] Keep the per-tool error message specific to that tool (the audit also flagged that the current error string hardcodes `getAthleteProfileName` — fix in passing).

### Step 3: Collapse `schemaCatalogClient`

- [ ] Replace `schemaCatalogClient` with the minimal fake needed (or `*intervals.Client` if a no-network mode exists). Goal: the file shrinks substantially.
- [ ] Keep schema-stability snapshot output byte-identical; verify with the existing snapshot tests.

### Step 4: Tests

- [ ] Run `make test` and `make test-race`.
- [ ] Verify the registered tool catalog (names + count) is unchanged via the existing schema-stability snapshots.
- [ ] Add a small registry-level test that confirms `*intervals.Client` registers every advertised tool (regression guard against silently dropping a tool when adding a new one in the future).

### Step 5: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] `git diff --stat` should show `registry.go` and `schema_stability.go` shrinking; no tool file should grow beyond its existing narrow-interface declaration.
- [ ] Manual smoke: start the server (stdio), confirm `list_tools` matches pre-refactor output exactly.

## Acceptance Criteria

- `defaultRegistry.Register` no longer contains a chain of `client, ok := r.profileClient.(XxxClient)` assertions; it calls each tool constructor directly with a typed dependency.
- `internal/toolchecks/schema_stability.go` no longer needs a hand-maintained `schemaCatalogClient` stub for every tool interface (or that stub is materially smaller).
- All per-tool narrow `XxxClient` interfaces remain (test seam preserved).
- Tool catalog is byte-identical: schema-stability snapshot tests pass unchanged.
- All gating semantics (delete-mode, toolset tier, capability) preserved.
- Error message for tool wiring failure references the failing tool, not `getAthleteProfileName`.

## Do NOT

- Do not remove the narrow per-tool interfaces — they are the unit-test seam.
- Do not change tool names, schemas, or `_meta` output. This is a wiring refactor only.
- Do not introduce a new "service locator" / DI container abstraction. Plain struct or plain argument list.
- Do not move gating logic out of `Register` into individual tools — gating must stay registration-time.
- Do not panic anywhere; preserve error wrapping with `%w`.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` (under "Changed" — internal refactor, no user-visible behaviour change, but worth noting).

## Git Commit Convention

`TP-042 collapse registry interface-assertion sprawl`, etc.

---

## Amendments

_Add amendments below this line only._
