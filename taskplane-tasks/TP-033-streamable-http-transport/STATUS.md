# TP-033-streamable-http-transport: TP-033-streamable-http-transport — Status

**Current Step:** Complete
**Status:** ✅ Complete
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 18
**Iteration:** 1
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

**Status:** ✅ Complete

- [x] Default bind is loopback only; confirm with a test that the default config never produces a non-loopback listener.
- [x] No API keys or athlete IDs in HTTP logs; reuse the existing redaction conventions.
- [x] Document the LAN-bind threat model briefly in README (anyone on the LAN can reach the server with no auth — opt in deliberately).
- [x] Add a race-safe HTTP log redaction test that exercises startup/listen/shutdown and a malformed request without leaking API keys or athlete IDs.
- [x] Make the app-level non-loopback warning redaction test race-safe so `go test -race ./internal/mcp ./internal/app` passes.

### Step 4: Parity tests

**Status:** ✅ Complete

- [x] The same protocol tests that cover stdio (initialize, tools/list, tool calls, resources, prompts, malformed requests, sanitized errors) run against the HTTP transport.
- [x] Handler behaviour is byte-identical across transports — assert this where practical.
- [x] Add HTTP-specific malformed POST coverage while keeping the raw newline malformed test for stdio/IO framing.

### Step 5: Docs

**Status:** ✅ Complete

- [x] README: transport selection, default loopback bind, opt-in LAN bind + security note.
- [x] CHANGELOG `[Unreleased]` entry.

### Step 6: Verify

**Status:** ✅ Complete

- [x] `make test`
- [x] `make build`
- [x] `make lint`
- [x] `go test -race ./...`
- [x] Manual: start in `http` mode; confirm it binds `127.0.0.1` only by default; drive a tool call over HTTP from one MCP client.

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
| 2026-05-14 18:28 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 18:38 | Step 4 implemented | Added shared in-memory/HTTP protocol suite, canonical parity snapshot, and HTTP malformed POST coverage; `go test ./internal/mcp` and `go test -race ./internal/mcp` pass. |

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
| 2026-05-14 17:53 | Review R008 | plan Step 3: REVISE |
- Step 3 revised plan after R008: extend `internal/config` coverage for `ICUVISOR_TRANSPORT=http` with no bind override and assert `DefaultHTTPBindAddress` is the canonical value, parses as loopback, and is not wildcard/non-loopback without starting a fixed port. Extend HTTP logging tests using the synchronized log-buffer pattern, with sentinel API key and raw athlete ID values, covering startup, non-loopback warning, listener log, shutdown, and one malformed HTTP request; assert neither sentinel appears. Do not add HTTP request/response/header/body access logging, and keep SDK Streamable HTTP logger use limited to lifecycle/spec-level messages. Update the README transport/configuration section to state that default Streamable HTTP is loopback-only and that `ICUVISOR_HTTP_BIND`/`--http-bind` to a LAN address exposes an unauthenticated MCP server whose reachable clients can invoke tools using the configured intervals.icu credentials.
- R010 code review fix plan: replace the remaining plain `bytes.Buffer` in the app-level non-loopback HTTP warning test with `safeAppLogBuffer`; keep the forbidden API-key and athlete-ID assertions in that app test because it exercises `defaultStartServer` with configured secrets and the LAN-bind warning path.
| 2026-05-14 17:56 | Review R009 | plan Step 3: APPROVE |
| 2026-05-14 18:01 | Review R010 | code Step 3: REVISE |
| 2026-05-14 18:05 | Review R011 | code Step 3: APPROVE |
| 2026-05-14 18:08 | Review R012 | plan Step 4: REVISE |
- Step 4 revised plan after R012: refactor `internal/mcp/protocol_test.go` around a `connectProtocolClient(t, transportKind, opts)` helper with two cases: the current in-memory/stdio-equivalent SDK transport and Streamable HTTP served on `127.0.0.1:0` at `StreamableHTTPPath` using `sdkmcp.StreamableClientTransport`, `MaxRetries: -1`, short HTTP timeouts, session close, context cancellation, and `waitForServerRun`. Run the shared protocol suite across both transports for initialize, `tools/list`, successful calls, missing/unknown tools, sanitized tool errors, `resources/list`, `resources/read`, missing resources, sanitized resource errors, and current `prompts/list` behaviour. Add a parity test that serializes stable SDK results to canonical JSON and compares bytes for initialize server info/capabilities, tools/list, call-tool content/structured content, resources/list/read, and prompts/list, excluding only unavoidable transport/session-specific metadata. Keep the existing raw newline malformed-request test for IO framing and add an HTTP-specific malformed POST to `/mcp` that asserts a client-visible error response without panic or leaked secrets.
| 2026-05-14 18:39 | Review R013 | code Step 4: APPROVE |
| 2026-05-14 18:40 | Step 4 complete | Code review R013 approved parity tests. |
| 2026-05-14 18:41 | Step 5 started | Plan: expand README MCP transport docs with env/flag/config names, loopback default endpoint, and LAN unauthenticated threat note; add a concise CHANGELOG `[Unreleased]` entry for Streamable HTTP. |
| 2026-05-14 18:42 | README docs updated | Added Streamable HTTP env/flag examples, JSON config fields, default loopback endpoint, invalid config behaviour, and LAN security warning. |
| 2026-05-14 18:43 | CHANGELOG updated | Added `[Unreleased]` Streamable HTTP transport entry. |
| 2026-05-14 18:44 | Review R015 | code Step 5: APPROVE |
| 2026-05-14 18:44 | Step 5 complete | Docs reviewed and approved. |
| 2026-05-14 18:45 | Step 6 started | Plan: run required make/go verification commands, then manually launch HTTP mode with dummy local config on the default loopback bind and drive an MCP `tools/list` call over Streamable HTTP. |
| 2026-05-14 18:46 | Review R016 | plan Step 6: REVISE |
| 2026-05-14 18:47 | Step 6 revised plan | Automated checks stay `make test`, `make build`, `make lint`, and `go test -race ./...`. Manual smoke will start `./bin/icuvisor --transport http` without `--http-bind`/`ICUVISOR_HTTP_BIND`, using sentinel config env values, verify the startup log and `lsof` listener are `127.0.0.1:8765` and not wildcard/non-loopback, then use a small Go MCP client script with `sdkmcp.StreamableClientTransport` pointed at `http://127.0.0.1:8765/mcp` to run both `tools/list` and a real `tools/call` for no-network `icuvisor_list_advanced_capabilities`. |
| 2026-05-14 18:48 | Verify | `make test` passed. |
| 2026-05-14 18:49 | Verify | `make build` passed. |
| 2026-05-14 18:51 | Verify | `make lint` initially caught copyloopvar/whitespace in the new protocol test; fixed and reran `go test ./internal/mcp` plus `make lint`, both passed. |
| 2026-05-14 18:52 | Verify | `go test -race ./...` passed. |
| 2026-05-14 18:53 | Verify | Manual HTTP smoke passed: started `./bin/icuvisor` with `ICUVISOR_TRANSPORT=http` and sentinel config, confirmed startup log and `lsof` listener were `127.0.0.1:8765` only, then used an SDK Streamable HTTP client to run `tools/list` and `tools/call` for no-network `icuvisor_list_advanced_capabilities`. |
| 2026-05-14 18:54 | Verify | Re-ran `make test` after the lint cleanup; passed. |
| 2026-05-14 18:42 | Review R014 | plan Step 5: APPROVE |
| 2026-05-14 18:44 | Review R015 | code Step 5: APPROVE |
| 2026-05-14 18:47 | Review R016 | plan Step 6: REVISE |
| 2026-05-14 18:48 | Review R017 | plan Step 6: APPROVE |
| 2026-05-14 18:54 | Review R018 | code Step 6: APPROVE |
| 2026-05-14 18:55 | Step 6 complete | All automated and manual verification passed; code review R018 approved. |
