# Task: TP-162 - Codex Streamable HTTP JSON-RPC smoke coverage

**Created:** 2026-06-09
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds protocol compatibility smoke coverage for the HTTP transport. It should be mostly tests, but transport handshake behavior is client-critical and cross-client visible.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-162-codex-streamable-http-jsonrpc-smoke/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add Codex-compatible Streamable HTTP smoke coverage that verifies initialization and ping responses are strict JSON-RPC 2.0 envelopes. A public Montis report showed Codex handshake failure until ping/initialization responses were wrapped correctly; icuvisor should prevent regressions in its HTTP transport.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — MCP transport and loopback-bind rules.
- `docs/clients/codex-local.md` — Codex client setup expectations.
- `docs/prd/PRD-icuvisor.md` — MCP transport requirements.

## Environment

- **Workspace:** Go module root
- **Services required:** None; use in-process HTTP test server only.

## File Scope

- `internal/mcp/transport.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `internal/mcp/protocol_test.go`
- `internal/mcp/protocol_registry_test_helpers_test.go`
- `docs/clients/codex-local.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Add Streamable HTTP JSON-RPC handshake smoke tests

- [ ] Add an in-process HTTP transport test for `initialize` that asserts the response is a JSON-RPC 2.0 object with `jsonrpc: "2.0"`, matching `id`, and a valid `result` object.
- [ ] Add a ping test that asserts the response is also a JSON-RPC 2.0 envelope, not a bare object/string/null payload.
- [ ] Include request/response headers used by Codex-like Streamable HTTP clients without binding beyond loopback.
- [ ] Run targeted tests: `go test ./internal/mcp -run 'Streamable|JSONRPC|Codex|Protocol|Ping|Initialize'`

**Artifacts:**
- `internal/mcp/server_test.go` (modified if best location)
- `internal/mcp/protocol_test.go` (modified if best location)
- `internal/mcp/protocol_registry_test_helpers_test.go` (modified if helper needed)

### Step 2: Fix transport/protocol behavior only if tests fail

- [ ] If the smoke tests fail, update `internal/mcp/transport.go` or server wiring so Streamable HTTP responses preserve strict JSON-RPC 2.0 envelopes.
- [ ] Preserve existing stdio behavior and default HTTP loopback binding.
- [ ] Ensure protocol errors remain short/actionable and do not expose internals.
- [ ] Run targeted tests: `go test ./internal/mcp`

**Artifacts:**
- `internal/mcp/transport.go` (modified if needed)
- `internal/mcp/server.go` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `docs/clients/codex-local.md` mentions the compatibility smoke coverage or updates troubleshooting if needed.
- [ ] `CHANGELOG.md` notes the Codex/Streamable HTTP compatibility guard.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note compatibility smoke coverage/fix.

**Check If Affected:**
- `docs/clients/codex-local.md` — update troubleshooting or compatibility notes if behavior changes.
- `README.md` — update only if transport support wording changes.

## Completion Criteria

- [ ] Initialize and ping Streamable HTTP tests assert strict JSON-RPC 2.0 response envelopes.
- [ ] HTTP remains loopback-bound by default.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-162): complete Step N — description`
- **Bug fixes:** `fix(TP-162): description`
- **Tests:** `test(TP-162): description`
- **Hydration:** `hydrate: TP-162 expand Step N checkboxes`

## Do NOT

- Bind HTTP transport beyond `127.0.0.1` by default.
- Require a live Codex process or external network in tests.
- Copy competitor implementation; use only the public handshake failure signal.
- Change stdio behavior unless tests prove it is part of the same bug.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
