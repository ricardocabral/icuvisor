# TP-033 — Streamable HTTP transport (localhost-bound by default)

## Mission

Add the Streamable HTTP transport alongside the existing stdio transport, bound to `127.0.0.1` by default. This serves clients that prefer HTTP and lays groundwork for the future hosted-relay story. A LAN bind is opt-in and must be documented.

Roadmap items (ROADMAP.md v0.4):

- Streamable HTTP transport (localhost-bound by default).

PRD anchors: §7.3 transports / §8 (Streamable HTTP bound to `127.0.0.1` by default, optional LAN binding), CLAUDE.md hard rule #7 ("Don't expose the HTTP transport beyond `127.0.0.1` by default"), §7.4 #1/#3 (Go SDK treated as production-ready for stdio + Streamable HTTP).

Complexity: Blast radius 2 (transport layer; tool handlers unchanged), Pattern novelty 2, Security 3 (network bind surface), Reversibility 2 = 9 → Review Level 2. Size: M.

## Dependencies

- **TP-003** — MCP stdio server skeleton; the HTTP transport reuses the same server core and tool/resource/prompt registry. Transport choice must not change handler behaviour.

## Context to Read First

- `CLAUDE.md` (esp. hard rule #7)
- `docs/prd/PRD-icuvisor.md` §7.3, §8, §7.4 #1/#3
- `ROADMAP.md` v0.4
- `internal/mcp/` — server core + existing stdio transport wiring
- `internal/config/` — config/env-var conventions for bind address + transport selection
- Go SDK docs for the Streamable HTTP transport (record the canonical link in `STATUS.md`)

## File Scope

Expected files:

- `internal/mcp/` — Streamable HTTP transport wiring next to the stdio transport; shared server core
- `internal/mcp/*_test.go`
- `internal/config/` — transport selection + bind-address config (default `127.0.0.1`, opt-in LAN bind)
- `cmd/icuvisor/` — flag/env to select transport and bind address (keep `main` thin)
- `README.md` — document the transport, the default bind, and the opt-in LAN bind with its security note
- `CHANGELOG.md`
- `taskplane-tasks/TP-033-streamable-http-transport/STATUS.md`

## Steps

### Step 1: Transport selection plumbing

- [ ] Config/flag to select `stdio` (default) or `http`; bind address config defaults to `127.0.0.1:<port>`
- [ ] A non-loopback bind is opt-in and requires an explicit config value — never the default; log a clear WARNING when a non-loopback bind is active
- [ ] Invalid transport/bind values fail loudly at startup

### Step 2: Streamable HTTP transport

- [ ] Wire the Go SDK Streamable HTTP transport onto the shared server core; the tool/resource/prompt registry is identical across transports
- [ ] Single shared server lifecycle (startup/shutdown, context cancellation honored)
- [ ] Graceful shutdown closes the listener and in-flight requests cleanly

### Step 3: Security posture

- [ ] Default bind is loopback only; confirm with a test that the default config never produces a non-loopback listener
- [ ] No API keys or athlete IDs in HTTP logs; reuse the existing redaction conventions
- [ ] Document the LAN-bind threat model briefly in README (anyone on the LAN can reach the server with no auth — opt in deliberately)

### Step 4: Parity tests

- [ ] The same protocol tests that cover stdio (initialize, tools/list, tool calls, resources, prompts, malformed requests, sanitized errors) run against the HTTP transport
- [ ] Handler behaviour is byte-identical across transports — assert this where practical

### Step 5: Docs

- [ ] README: transport selection, default loopback bind, opt-in LAN bind + security note
- [ ] CHANGELOG `[Unreleased]` entry

### Step 6: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: start in `http` mode; confirm it binds `127.0.0.1` only by default; drive a tool call over HTTP from one MCP client

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for transport ergonomics. Do not depend on it.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- Streamable HTTP transport works against the shared server core; stdio remains the default.
- Default bind is `127.0.0.1`; a non-loopback bind requires explicit opt-in config and logs a WARNING.
- Protocol-test parity across stdio and HTTP.
- Invalid transport/bind config fails loudly at startup.
- README documents the transport, default bind, and LAN-bind security note; CHANGELOG updated.

## Do NOT

- Do not bind beyond `127.0.0.1` by default, ever.
- Do not add authentication scope-creep here — that belongs to the future hosted-relay story; this task is loopback-first.
- Do not fork handler logic per transport — share the server core.
- Do not log API keys, tokens, or raw athlete identifiers in HTTP request/response logging.

## Documentation

Must update:

- `STATUS.md`
- `README.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-033`, for example: `TP-033 add streamable HTTP transport selection plumbing`.

---

## Amendments

_Add amendments below this line only._
