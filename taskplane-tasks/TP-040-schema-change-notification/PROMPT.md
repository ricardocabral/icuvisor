# TP-040 — Post-update notification: tell the user to start a new conversation when tool schemas changed

## Mission

When the icuvisor binary is upgraded and the new build's tool catalog or argument schemas differ from the version a running MCP client is talking to, surface that fact to the *user* (via the MCP client's response stream) so they know to start a new conversation — because MCP clients cache the tool catalog per conversation (PRD §7.4 #7) and an in-flight chat will keep using the stale schema and report "the fix didn't work."

Today every response already carries `_meta.server_version` (v0.2 acceptance, ROADMAP). This task layers a *catalog hash* on top of the version string, persists the hash the client first saw on session start, and emits a `_meta.schema_changed: true` callout plus a one-line human-readable warning string when the hash diverges. The callout includes the previously-seen and current versions so the LLM can articulate "icuvisor was upgraded from vX to vY since this conversation started; tool schemas may have changed — please open a new chat to use the latest tools."

This task does *not* implement the auto-update / signed-release machinery (that's v1.0). It only implements the **notification mechanism** that v0.5's manually-upgraded internal-beta athletes need to avoid the "the fix didn't work" failure mode (PRD §6 Pains avoided, last bullet).

PRD anchors: §7.4 #7 ("MCP tool-schema caching is per-conversation on all target clients" — implication: tool argument changes must be additive-only on stable tools; every tool response embeds `_meta.server_version`; users must be told to start a new conversation after an update); §6 Pains avoided ("Confusing 'fix didn't land' experiences after server upgrades, caused by MCP clients caching the tool schema per conversation"); §7.1 Flow C (post-update notification copy).

ROADMAP positioning: v0.5 — Internal beta, fifth item.

Complexity: Blast radius 3 (touches the response-shaping middleware that already adds `_meta.server_version`), Pattern novelty 2 (catalog hashing + session-first-seen state), Security 1 (no new credential paths), Reversibility 1 = 7 → Review Level 2. Size: S.

## Dependencies

- **TP-015** — schema-stability rules in CI; this task surfaces the *runtime* signal that complements the CI-time additive-only guard. The CI guard prevents accidental breaks; this task tells users a benign change happened.
- **TP-007** — response-shaping primitives include `_meta.server_version`; we extend `_meta` here.

## Context to Read First

- `CLAUDE.md` "MCP-server conventions" — `_meta` is part of the API.
- `docs/prd/PRD-icuvisor.md` §7.4 #7, §6, §7.1 Flow C.
- ROADMAP.md v0.2 (`_meta.server_version` is already shipped) and v0.5 (this task).
- `internal/response/` — the package that owns `_meta` injection.
- `internal/mcp/` — server wiring for catalog construction.

## File Scope

Expected files:

- `internal/mcp/catalog_hash.go` (new) — deterministic SHA-256 over the registered tools' `(name, JSON schema)` pairs, stable across reorderings. Computed once at server start; exposed via `Server.CatalogHash() string`.
- `internal/mcp/catalog_hash_test.go` — fixture: reorder tool registrations, confirm hash stable; rename an argument, confirm hash changes.
- `internal/response/meta.go` (or equivalent) — extend the `_meta` injector. On each response, include `server_version` (already there) and `catalog_hash` (new). On session start (first response to a given client, identified by the MCP transport's session handle if available, otherwise by the first response after process start), record the hash the client *first saw* in process-local memory keyed by session.
- The injector compares the current hash with the first-seen hash; if they differ, adds:
  ```json
  "_meta": {
    "schema_changed": true,
    "schema_change_message": "icuvisor was upgraded from v0.4.1 to v0.5.0 since this conversation started; tool schemas may have changed. Open a new conversation to use the latest tools.",
    "previous_version": "v0.4.1",
    "current_version": "v0.5.0",
    "catalog_hash": "ab12cd…",
    "previous_catalog_hash": "9f3e22…"
  }
  ```
  In practice the binary doesn't restart mid-session inside a process, so `schema_changed` will *not* fire normally — it fires only when the MCP transport supports session resumption across binary restarts, or when an integration test simulates it. Document this clearly: the value of the field is in the integration-test guarantee and in the documented protocol, not in day-to-day firing.
- `internal/response/meta_test.go` — table-driven: session-start records, subsequent responses carry hash, simulated hash change injects `schema_changed`.
- `internal/mcp/` — wire the session-handle plumbing (or fall back to per-process state with a docstring caveat if the go-sdk has no session abstraction; same caveat documented in `STATUS.md` as TP-039 does).
- `docs/post-update.md` (new) — explains the field to users and AI clients; pairs with `CHANGELOG.md` so each release lists what schemas changed.
- `CHANGELOG.md`.
- `taskplane-tasks/TP-040-schema-change-notification/STATUS.md`.

## Steps

### Step 1: Catalog hash

- [ ] Define the input: serialized `(tool_name, marshalled JSON schema)` pairs sorted by name; SHA-256 of the concatenation. Lock with a golden fixture.
- [ ] Confirm the hash is stable across map-iteration order, registration order, and unrelated build flags. Verify by repeated runs in the test.
- [ ] Confirm the hash changes when (a) a tool is added, (b) a tool is removed, (c) an argument is renamed, (d) a description string changes — yes, description changes too, because the LLM uses the description (PRD §7.2.D "Self-explanatory shapes"). A description-only change still warrants a "schemas changed" notice.

### Step 2: `_meta` injector

- [ ] Add `catalog_hash` to every response's `_meta`.
- [ ] Track first-seen hash per session (best-effort; per-process if no session handle).
- [ ] When current ≠ first-seen, populate `schema_changed`, `schema_change_message`, `previous_version`, `current_version`, `previous_catalog_hash` fields.
- [ ] The human message string is templated; the template lives in code and is testable.

### Step 3: Tests

- [ ] Catalog-hash determinism + sensitivity tests (Step 1 list).
- [ ] Injector behaviour: session-start, steady-state, simulated change.
- [ ] Confirm `_meta.schema_changed` appears in the actual response JSON (not just in the Go struct) and that downstream golden-file tests for individual tools are not destabilised by the always-on `catalog_hash` field — pin a stable fixture-hash via test-only injection or move `catalog_hash` to a deterministic value in tests.

### Step 4: Documentation

- [ ] `docs/post-update.md` — what `_meta.schema_changed` means, the recommended user action ("open a new conversation in your AI client"), and the limits (cannot fire if your client doesn't surface `_meta` back to the LLM — some clients don't; the same caveat as in TP-007 in-response scale labels).
- [ ] CHANGELOG `[Unreleased]` entry.

## Acceptance Criteria

- Every tool response carries `_meta.server_version` (unchanged behaviour) and `_meta.catalog_hash` (new).
- A simulated change in catalog hash (test-injected) produces `_meta.schema_changed: true` with all six fields populated.
- Catalog hash is deterministic across registration order and stable across unrelated rebuilds.
- The human-readable `schema_change_message` is templated and asserted by a test.
- `docs/post-update.md` exists and is linked from the README post-install / upgrade section.
- Existing tool golden files do not destabilise from the new field (test-only fixed-hash injection or equivalent).

## Do NOT

- Do not log the catalog hash at INFO on every response — once at startup is enough.
- Do not surface the hash to the LLM in tool *descriptions*; it's a `_meta` runtime signal only.
- Do not exclude description-only changes from the hash. Descriptions are part of the contract the LLM reasons over.
- Do not implement auto-update or release-channel polling in this task — that's v1.0.
- Do not break golden-file tests on every build because `catalog_hash` shifted. Use a test-only fixed-hash injector for tool fixtures; the injector's own tests cover the real hashing.

## Documentation

Must update:

- `STATUS.md`
- `docs/post-update.md` (new)
- `CHANGELOG.md`
- `README.md` (one-line pointer under "Updating")

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-040`, for example: `TP-040 add catalog hash and schema-change meta`.

---

## Amendments

_Add amendments below this line only._
