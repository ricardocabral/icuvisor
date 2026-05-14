# TP-033-streamable-http-transport: TP-033-streamable-http-transport — Status

**Current Step:** Step 1: Transport selection plumbing
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
**Size:** M

---

### Step 1: Transport selection plumbing

**Status:** 🟨 In Progress

- [x] Config/flag selects `stdio` (default) or `http`; HTTP bind defaults to `127.0.0.1:<port>`.
- [x] Non-loopback bind requires an explicit config value and logs a clear WARNING when active.
- [x] Invalid transport and bind values fail loudly at startup.

### Step 2: Streamable HTTP transport

**Status:** ⏳ Not started

- [ ] Wire the Go SDK Streamable HTTP transport onto the shared server core; the tool/resource/prompt registry is identical across transports.
- [ ] Single shared server lifecycle (startup/shutdown, context cancellation honored).
- [ ] Graceful shutdown closes the listener and in-flight requests cleanly.

### Step 3: Security posture

**Status:** ⏳ Not started

- [ ] Default bind is loopback only; confirm with a test that the default config never produces a non-loopback listener.
- [ ] No API keys or athlete IDs in HTTP logs; reuse the existing redaction conventions.
- [ ] Document the LAN-bind threat model briefly in README (anyone on the LAN can reach the server with no auth — opt in deliberately).

### Step 4: Parity tests

**Status:** ⏳ Not started

- [ ] The same protocol tests that cover stdio (initialize, tools/list, tool calls, resources, prompts, malformed requests, sanitized errors) run against the HTTP transport.
- [ ] Handler behaviour is byte-identical across transports — assert this where practical.

### Step 5: Docs

**Status:** ⏳ Not started

- [ ] README: transport selection, default loopback bind, opt-in LAN bind + security note.
- [ ] CHANGELOG `[Unreleased]` entry.

### Step 6: Verify

**Status:** ⏳ Not started

- [ ] `make test`
- [ ] `make build`
- [ ] `make lint`
- [ ] `go test -race ./...`
- [ ] Manual: start in `http` mode; confirm it binds `127.0.0.1` only by default; drive a tool call over HTTP from one MCP client.

---

## Reviews

| #    | Type | Step | Verdict | File |
| ---- | ---- | ---- | ------- | ---- |
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |
| Current worktree already contains the Streamable HTTP implementation from the previous merged TP-033 history; this iteration is auditing, re-verifying, and restoring current task artifacts. | Use existing implementation where it satisfies the prompt; do not duplicate handler logic. | `internal/config`, `internal/app`, `internal/mcp` |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 19:48 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 19:48 | Step 1 started | Transport selection plumbing |
| 2026-05-14 19:50 | Review R001 | plan Step 1: REVISE; added concrete plan details below before marking work complete |
| 2026-05-14 19:54 | Review R002 | plan Step 1: APPROVE |
| 2026-05-14 19:55 | Step 1 checkpoint | Config and CLI transport selection audited in `internal/config` and `internal/app`; `go test ./internal/config ./internal/app` passed. |
| 2026-05-14 19:56 | Step 1 checkpoint | Non-loopback explicit bind and WARN path audited in `internal/config`/`internal/app`; logs avoid API keys and athlete IDs. |
| 2026-05-14 19:57 | Step 1 checkpoint | Invalid transport and bind values fail in config validation. |
| 2026-05-14 19:58 | Recovery | Reverted premature Step 1 completion before required code review. |

---

## Blockers

_None_

---

## Notes

- Go SDK Streamable HTTP docs: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#NewStreamableHTTPHandler and https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#StreamableClientTransport
- Step 1 plan: expose config JSON fields `transport` and `http_bind`, `.env`/process env keys `ICUVISOR_TRANSPORT` and `ICUVISOR_HTTP_BIND`, and CLI overrides `--transport`/`--http-bind`; precedence is JSON < `.env` for absent values < process env < CLI options. `cmd/icuvisor/main.go` stays thin and delegates to `internal/app`, which owns flag parsing.
- Step 1 validation plan: default transport is `stdio`; HTTP mode is selected only by explicit `http`; default HTTP bind is `127.0.0.1:8765`. Accepted transports are exactly `stdio` and `http`. Bind parsing requires explicit IP host plus numeric port 1-65535 (IPv4 and bracketed IPv6 accepted), rejects wildcard-by-omission such as `:8765`, URL strings, missing port, non-numeric port, and out-of-range port. Non-loopback addresses are accepted only when explicitly configured because the default is loopback.
- Step 1 warning/test plan: log a structured WARN only when `transport=http` and the active bind is non-loopback; include transport and bind address only, never API keys or raw athlete IDs. Cover defaults, JSON/env/CLI selection, invalid transport/bind errors, non-loopback detection, and backward-compatible `version`, `--config path`, and `--config=path` CLI parsing.
