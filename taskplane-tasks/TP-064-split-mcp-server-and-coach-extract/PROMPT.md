# TP-064 — Split `internal/mcp/server.go` and lift coach middleware to `internal/coach` (audit God-module + God-package)

## Mission

`internal/mcp/server.go` is 818 LOC and mixes:

- server lifecycle (`NewServer`, `Run`, `RunStreamableHTTP`),
- a 270-LOC `safeRegistrar` (lines ~315-605) that does tool registration AND coach-mode visibility filtering AND athlete-ID schema injection,
- resource and prompt registrars,
- JSON schema mutation (`schemaWithAthleteID`, `stripAthleteID`),
- validation helpers + SDK ↔ tools type conversion,
- and a tool handler living in transport: `coachFilteredAdvancedCapabilitiesHandler` (lines ~566-606).

This couples transport ↔ authz ↔ tool shaping; any cross-cutting concern (audit, rate-limit, per-tool flag) needs edits in all three layers.

Goal:

1. **Split `mcp/server.go` by concern:**
   - `mcp/transport.go` — `Run`, `RunStreamableHTTP`, `normalizeHTTPServerError`, `transportName`.
   - `mcp/registrar_tools.go` — the `safeRegistrar` core minus coach concerns.
   - `mcp/registrar_resources.go`, `mcp/registrar_prompts.go`.
   - `mcp/schema.go` — `schemaWithAthleteID`, `stripAthleteID`, `validate*`, SDK type conversion.
   - `mcp/server.go` keeps `NewServer` + the public constructor only.

2. **Lift coach-mode concerns into `internal/coach`:** the existing `internal/coach` package owns selection state. Move coach-visibility filtering and `coachFilteredAdvancedCapabilitiesHandler` to live there (or to `internal/tools/list_advanced_capabilities.go`), expose a `coach.ToolFilter(ctx, registry, athleteID) []Tool` (signature TBD), and have `mcp/registrar_tools.go` call it. Transport stops importing coach internals.

Audit ref: 2026-05-16 Go audit, God-module + God-package sections.

PRD anchors: §7.2.G coach mode; §7.2.E catalog tiers.
CLAUDE.md hard rules: small packages, narrow interfaces.

Complexity: Blast radius 4 (touches transport, registrar, coach, tests), Pattern novelty 2 (new coach filter interface), Security 3 (coach ACL must not regress), Reversibility 2 (package surface change) = 11 → Review Level 3. Size: L.

## Dependencies

- **TP-042** — sequence after. TP-042 collapses registry interface-assertion sprawl; the cleaner registry surface makes this split safer.
- **TP-060** — soft. TP-060 hardens `withPanicRecovery`; if it lands first, port the helper to its new file location (likely `mcp/recover.go`).
- **TP-039** (coach mode) — context only; behaviour must be preserved exactly.

## Context to Read First

- `internal/mcp/server.go` — full file. Categorize every decl by target file.
- `internal/coach/*.go` — current coach package structure and selection store.
- `internal/tools/list_advanced_capabilities.go` — destination candidate for the filtered handler.
- `docs/coach-mode.md`, `docs/threat-models/coach-mode.md` — the ACL contract that must be preserved.
- `internal/mcp/server_test.go` — existing coverage. Identify the coach-ACL tests; they become the regression gate.

## File Scope

- Inside `internal/mcp/`: keep package boundary. Split into the new files listed above.
- Inside `internal/coach/`: add `filter.go` (or similar) housing the tool-visibility filter and the advanced-capabilities handler.
- `internal/tools/list_advanced_capabilities.go` — adjust if the handler lands here instead.
- `internal/mcp/server_test.go` — split alongside the source split where it improves clarity.
- Tests in `internal/coach/` — add coverage for the moved filter.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Changing the MCP wire protocol.
- Changing coach ACL semantics. Same matrix, just lifted.
- Adding new cross-cutting concerns (audit, rate-limit) — file follow-ups in `CONTEXT.md` Technical Debt.
- Touching tool schemas / catalog metadata.

## Steps

### Step 1: Inventory + capture regression-gate tests
- [ ] List every top-level decl in `mcp/server.go` and assign each to a target file. Record in `STATUS.md`.
- [ ] Identify all coach-ACL tests across packages (probably in `mcp/server_test.go` and `internal/tools/list_advanced_capabilities_test.go`). They become the immovable gate — every one must pass after the refactor.

### Step 2: Mechanical file split
- [ ] Move decls into the new files **without changing logic**. Cross-file private helpers stay package-private.
- [ ] Each commit moves one concern (transport, schema, etc.) — keep diffs reviewable.
- [ ] Run all checks after each move.

### Step 3: Lift coach concerns to `internal/coach`
- [ ] Design the seam: probably `coach.ToolFilter` taking the registry catalog + selection ctx and returning the filtered catalog. Sketch in `STATUS.md` first.
- [ ] Move `coachFilteredAdvancedCapabilitiesHandler` to its new home.
- [ ] `mcp/registrar_tools.go` calls into `coach` instead of inlining the filter.
- [ ] Re-run all tests; coach-ACL tests must pass unchanged.

### Step 4: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] `scripts/snapshot_tool_schemas.go` — diff empty.
- [ ] Coach-mode integration tests (in TP-039's test set) — all pass.
- [ ] No new public-API surface in `internal/mcp` beyond what existed before.

## Acceptance Criteria

- `internal/mcp/server.go` is < 200 LOC and holds only the constructor + package doc comment.
- The transport/registrar/schema responsibilities live in their own files.
- `internal/coach` owns tool-visibility filtering; `internal/mcp` no longer references coach internals beyond the seam.
- Coach ACL behaviour is byte-identical (regression-gate tests pass unchanged).
- All `make` checks pass.

## Do NOT

- Do not change MCP wire behaviour.
- Do not change coach ACL semantics or visible tool sets.
- Do not add new dependencies.
- Do not collapse the file split into a single big commit — reviewers need per-concern diffs.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).
- If the coach package gains an exported `ToolFilter` API, add a short note in `docs/coach-mode.md`.

## Git Commit Convention

Conventional Commits, prefixed `TP-064`. Recommended sequence:
1. `TP-064 split mcp transport into transport.go`
2. `TP-064 extract mcp schema helpers`
3. `TP-064 split mcp resource/prompt registrars`
4. `TP-064 lift coach filter into internal/coach`
5. `TP-064 finalize mcp/server.go to constructor only`

---

## Amendments

_Add amendments below this line only._
