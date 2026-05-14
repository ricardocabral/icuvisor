# Plan Review: Step 2 — Streamable HTTP transport

**Verdict:** APPROVE

I read `PROMPT.md`, `STATUS.md`, the relevant PRD/Roadmap anchors, the current Step 1 implementation shape, and the SDK v1.4.1 Streamable HTTP API. The updated Step 2 plan addresses the prior blocking gaps and is specific enough to implement safely.

## What is now covered

- **SDK wiring is explicit and correct for v1.4.1.** The plan uses `mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return sharedSDKServer }, options)`, which matches the current SDK signature and avoids treating `Server.Run` as the HTTP serving path.
- **Shared server/registry model is preserved.** The plan keeps `internal/mcp.NewServer` as the single constructor for the SDK server and registrations, with `internal/app.defaultStartServer` dispatching by `Config.Transport` rather than forking tool/resource/prompt registration.
- **HTTP surface and options are concrete.** `/mcp`, canonical bind address usage, stateful sessions, SSE-style responses (`JSONResponse: false`), logger, `30 * time.Minute` session timeout, and enabled localhost/cross-origin protections are all specified.
- **Lifecycle and graceful shutdown have an implementation shape.** The plan calls for `net.Listen`, `http.Server`, request contexts rooted in the worker context, a serve goroutine, expected `http.ErrServerClosed` handling, bounded `Shutdown`, and `Close` fallback.
- **Testability is planned without fixed-port flakes.** The injected-listener seam with `127.0.0.1:0` tests is the right approach for Step 2, while leaving full protocol parity for Step 4.
- **The bind validation/runtime mismatch is addressed.** Canonicalizing through `netip.AddrPort` during config validation should make logs, warning checks, and `net.Listen` operate on the same address string.

## Implementation notes

These are not blockers, but worth keeping in mind while coding:

- When wiring shutdown, avoid returning `http.ErrServerClosed` as an error on normal cancellation; preserve unexpected `Serve` errors.
- If request contexts are based directly on the parent worker context, cancellation will interrupt in-flight handlers immediately. That may be acceptable, but the shutdown test should reflect the intended behavior.
- Keep HTTP logs limited to transport/bind/lifecycle fields; do not add request/response logging that could expose API keys or athlete identifiers.

The plan is ready for implementation.
