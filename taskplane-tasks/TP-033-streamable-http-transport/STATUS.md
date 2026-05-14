# TP-033-streamable-http-transport: TP-033-streamable-http-transport — Status

**Current Step:** Step 2: Streamable HTTP transport
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 3
**Iteration:** 2
**Size:** M

---

### Step 1: Transport selection plumbing

**Status:** ✅ Complete

- [x] Config/flag selects `stdio` (default) or `http`, with HTTP bind address defaulting to `127.0.0.1:<port>`.
- [x] Non-loopback bind requires an explicit config value and logs a clear WARNING when active.
- [x] Invalid transport and bind values fail loudly at startup.

### Step 2: Streamable HTTP transport

**Status:** 🟨 In Progress

- [ ] Wire the Go SDK Streamable HTTP transport onto the shared server core; the tool/resource/prompt registry is identical across transports.
- [ ] Single shared server lifecycle (startup/shutdown, context cancellation honored).
- [ ] Graceful shutdown closes the listener and in-flight requests cleanly.

### Step 3: Security posture

**Status:** ⏳ Not started

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
| 2026-05-14 17:16 | Review R002 | plan Step 1: APPROVE |
| 2026-05-14 17:27 | Review R003 | code Step 1: APPROVE |
