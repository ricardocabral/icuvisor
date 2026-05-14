# TP-033-streamable-http-transport: TP-033-streamable-http-transport — Status

**Current Step:** Step 4: Parity tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 7
**Iteration:** 1
**Size:** M

---

### Step 1: Transport selection plumbing

**Status:** ✅ Complete

- [x] Config/flag selects `stdio` (default) or `http`; HTTP bind defaults to `127.0.0.1:<port>`.
- [x] Non-loopback bind requires an explicit config value and logs a clear WARNING when active.
- [x] Invalid transport and bind values fail loudly at startup.

### Step 2: Streamable HTTP transport

**Status:** ✅ Complete

- [x] Wire the Go SDK Streamable HTTP transport onto the shared server core; the tool/resource/prompt registry is identical across transports.
- [x] Single shared server lifecycle (startup/shutdown, context cancellation honored).
- [x] Graceful shutdown closes the listener and in-flight requests cleanly.

### Step 3: Security posture

**Status:** ✅ Complete

- [x] Default bind is loopback only; confirm with a test that the default config never produces a non-loopback listener.
- [x] No API keys or athlete IDs in HTTP logs; reuse the existing redaction conventions.
- [x] Document the LAN-bind threat model briefly in README (anyone on the LAN can reach the server with no auth — opt in deliberately).

### Step 4: Parity tests

**Status:** 🟨 In Progress

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
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | plan | 2 | APPROVE | `.reviews/R004-plan-step2.md` |
| R005 | code | 2 | APPROVE | `.reviews/R005-code-step2.md` |
| R006 | plan | 3 | APPROVE | `.reviews/R006-plan-step3.md` |
| R007 | code | 3 | APPROVE | `.reviews/R007-code-step3.md` |

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
| 2026-05-14 19:57 | Review R003 | code Step 1: APPROVE |
| 2026-05-14 19:59 | Step 1 complete | Transport selection plumbing audited and approved; Step 2 started. |
| 2026-05-14 19:59 | Review R004 | plan Step 2: APPROVE |
| 2026-05-14 20:00 | Step 2 checkpoint | Shared SDK server and `NewStreamableHTTPHandler` wiring audited in `internal/mcp`; `go test ./internal/mcp ./internal/app` passed. |
| 2026-05-14 20:01 | Step 2 checkpoint | Startup/shutdown paths use one `Server` wrapper for stdio and HTTP with context cancellation honored. |
| 2026-05-14 20:02 | Step 2 checkpoint | Graceful shutdown and listener-close tests audited in `TestServeStreamableHTTPCancelClosesListener`. |
| 2026-05-14 20:01 | Review R005 | code Step 2: APPROVE |
| 2026-05-14 20:03 | Step 2 complete | Streamable HTTP transport wiring and lifecycle audited and approved; Step 3 started. |
| 2026-05-14 20:03 | Review R006 | plan Step 3: APPROVE |
| 2026-05-14 20:04 | Step 3 checkpoint | Default loopback bind audited in config tests; `go test ./internal/config ./internal/mcp ./internal/app` passed. |
| 2026-05-14 20:05 | Step 3 checkpoint | HTTP log redaction tests audited for malformed requests, startup/listen/shutdown, API keys, and athlete IDs. |
| 2026-05-14 20:06 | Step 3 checkpoint | README LAN-bind threat model audited. |
| 2026-05-14 20:06 | Review R007 | code Step 3: APPROVE |
| 2026-05-14 20:07 | Step 3 complete | Security posture audited and approved; Step 4 started. |

---

## Blockers

_None_

---

## Notes

- Go SDK Streamable HTTP docs: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#NewStreamableHTTPHandler and https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#StreamableClientTransport
- Step 1 plan: expose config JSON fields `transport` and `http_bind`, `.env`/process env keys `ICUVISOR_TRANSPORT` and `ICUVISOR_HTTP_BIND`, and CLI overrides `--transport`/`--http-bind`; precedence is JSON < `.env` for absent values < process env < CLI options. `cmd/icuvisor/main.go` stays thin and delegates to `internal/app`, which owns flag parsing.
- Step 1 validation plan: default transport is `stdio`; HTTP mode is selected only by explicit `http`; default HTTP bind is `127.0.0.1:8765`. Accepted transports are exactly `stdio` and `http`. Bind parsing requires explicit IP host plus numeric port 1-65535 (IPv4 and bracketed IPv6 accepted), rejects wildcard-by-omission such as `:8765`, URL strings, missing port, non-numeric port, and out-of-range port. Non-loopback addresses are accepted only when explicitly configured because the default is loopback.
- Step 1 warning/test plan: log a structured WARN only when `transport=http` and the active bind is non-loopback; include transport and bind address only, never API keys or raw athlete IDs. Cover defaults, JSON/env/CLI selection, invalid transport/bind errors, non-loopback detection, and backward-compatible `version`, `--config path`, and `--config=path` CLI parsing.
- Step 2 plan: keep `internal/mcp.NewServer` as the single shared SDK server/registry constructor. For stdio, keep `Server.Run(ctx)` over `sdkmcp.StdioTransport`. For HTTP, serve the same SDK server through `sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server { return sharedSDKServer }, options)` mounted at `/mcp`; do not duplicate tool/resource/prompt registration or handler logic.
- Step 2 lifecycle plan: `RunStreamableHTTP` owns `net.Listen`, `ServeStreamableHTTP` accepts an injected listener for tests, uses `http.Server` with request contexts rooted in the worker context, treats `http.ErrServerClosed` as expected, and on cancellation calls bounded `Shutdown` followed by `Close` if needed. Tests cover app transport dispatch, HTTP initialize smoke, and cancellation closing the listener.
- Step 3 plan: verify the default HTTP bind remains `127.0.0.1:8765` and is loopback via `internal/config` tests; keep non-loopback binds explicit and WARN-only. Audit HTTP logs for startup/listen/shutdown/malformed request paths to ensure no API keys, tokens, raw athlete IDs, or request payloads are logged.
- Step 3 docs plan: README must state the LAN-bind threat model clearly: Streamable HTTP has no auth in this task, so anyone on the LAN who can reach the bind address can invoke registered tools with the configured intervals.icu credentials.
