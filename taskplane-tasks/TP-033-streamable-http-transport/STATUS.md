# TP-033-streamable-http-transport: TP-033-streamable-http-transport — Status

**Current Step:** Step 3: Security posture
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 7
**Iteration:** 2
**Size:** M

---

### Step 1: Transport selection plumbing

**Status:** ✅ Complete

- [x] Config/flag selects `stdio` (default) or `http`, with HTTP bind address defaulting to `127.0.0.1:<port>`.
- [x] Non-loopback bind requires an explicit config value and logs a clear WARNING when active.
- [x] Invalid transport and bind values fail loudly at startup.

### Step 2: Streamable HTTP transport

**Status:** ✅ Complete

- [x] Wire the Go SDK Streamable HTTP transport onto the shared server core; the tool/resource/prompt registry is identical across transports.
- [x] Single shared server lifecycle (startup/shutdown, context cancellation honored).
- [x] Graceful shutdown closes the listener and in-flight requests cleanly.
- [x] Normalize or reject bind-address whitespace so validated values match the listener address used at runtime.

### Step 3: Security posture

**Status:** 🟨 In Progress

- [ ] Default bind is loopback only; confirm with a test that the default config never produces a non-loopback listener.
- [ ] No API keys or athlete IDs in HTTP logs; reuse the existing redaction conventions.
- [ ] Document the LAN-bind threat model briefly in README (anyone on the LAN can reach the server with no auth — opt in deliberately).

### Step 4: Parity tests

**Status:** ⏳ Not started

### Step 5: Docs

**Status:** ⏳ Not started

### Step 6: Verify

**Status:** ⏳ Not started

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 17:03 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 17:03 | Step 1 started | Transport selection plumbing |
| 2026-05-14 17:11 | Worker iter 1 | done in 479s, tools: 26 |
| 2026-05-14 17:11 | Soft progress | Iteration 1: 0 new checkboxes but uncommitted source changes detected — not counting as stall |

---

## Blockers

_None_

---

## Notes

- Go SDK Streamable HTTP docs: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#NewStreamableHTTPHandler and https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp#StreamableClientTransport
- Step 1 revised plan after R001: add config JSON fields `transport` and `http_bind`, `.env`/process env keys `ICUVISOR_TRANSPORT` and `ICUVISOR_HTTP_BIND`, and CLI overrides `--transport`/`--http-bind`; precedence is JSON < `.env` for absent values < process env < CLI options. `config.Config` carries strict `TransportStdio`/`TransportHTTP` and `HTTPBindAddress`; `config.Options` carries CLI overrides so `cmd/icuvisor/main.go` remains thin and `internal/app` owns parsing.
- Step 1 validation plan: default transport is `stdio`; default HTTP bind is `127.0.0.1:8765`; transport parsing is strict and invalid values fail startup; bind parsing requires explicit IP host plus numeric port 1-65535 (IPv4 and bracketed IPv6 accepted), rejects wildcard-by-omission like `:8765`, URL strings, missing port, non-numeric port, and out-of-range port. Non-loopback addresses are accepted only when explicitly configured because the default is loopback; the startup warning is emitted only when `transport=http` and the active bind is non-loopback.
- Step 1 logging/test plan: keep config validation side-effect free; log the LAN-bind warning from startup using structured `slog.Warn` with transport and bind address only (no API key or athlete ID). Cover defaults, JSON/env/CLI selection, `.env` recognition, invalid transport/bind errors, non-loopback warning, and backward-compatible `version`, `--config path`, and `--config=path` CLI parsing.
- Step 2 revised plan after R005: keep `internal/mcp.NewServer` as the single shared SDK server/registry constructor. For stdio, keep `Server.Run(ctx)` over `sdkmcp.StdioTransport`. For HTTP, add an `internal/mcp` runner that mounts the SDK handler on `/mcp` using the v1.4.1 construction shape `mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return sharedSDKServer }, options)`, so Streamable HTTP creates/manages per-session transports without duplicating tool/resource/prompt registration. `internal/app.defaultStartServer` dispatches on `Config.Transport` and passes the same constructed registries into either transport.
- Step 2 HTTP serving plan: listen on canonical `Config.HTTPBindAddress`, expose the endpoint at `http://<bind>/mcp`, and use concrete SDK `StreamableHTTPOptions`: `Stateless: false`, `JSONResponse: false`, `Logger: logger`, `SessionTimeout: 30 * time.Minute`, `DisableLocalhostProtection: false`, and `CrossOriginProtection: nil`.
- Step 2 lifecycle/test plan: implement HTTP serving with `net.Listen`, `http.Server`, request contexts rooted in the worker context, a serve goroutine that treats `http.ErrServerClosed` as expected, and cancellation-triggered `Shutdown` with a short bounded timeout followed by `Close` on timeout. Add a test seam that accepts an injected `net.Listener` so tests use `127.0.0.1:0`; cover app transport dispatch, HTTP initialize smoke, cancellation closing the listener, and graceful shutdown. Normalize bind addresses to the canonical `netip.AddrPort` string during config validation so warning logs and `net.Listen` use the exact validated address.
| 2026-05-14 17:33 | Review R005 | plan Step 2: REVISE |
| 2026-05-14 17:16 | Review R002 | plan Step 1: APPROVE |
| 2026-05-14 17:27 | Review R003 | code Step 1: APPROVE |
| 2026-05-14 17:31 | Review R004 | plan Step 2: REVISE |
| 2026-05-14 17:31 | Review R004 | plan Step 2: REVISE |
| 2026-05-14 17:34 | Review R005 | plan Step 2: REVISE |
| 2026-05-14 17:36 | Review R006 | plan Step 2: APPROVE |
| 2026-05-14 17:49 | Review R007 | code Step 2: APPROVE |
