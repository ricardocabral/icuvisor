# Plan Review: Step 2 — Streamable HTTP transport

**Verdict:** REVISE

I read `PROMPT.md`, `STATUS.md`, and the current Step 1 implementation shape in `internal/app`, `internal/config`, and `internal/mcp`. The current Step 2 "plan" in `STATUS.md` only repeats the task checkboxes. For this step, that is not enough: Streamable HTTP has a different serving model than the existing stdio/in-memory transport path, and the lifecycle/security details are where the risk is.

## Blocking gaps to address in the revised plan

1. **Specify the SDK wiring model.**
   The existing `mcp.Server.Run` connects one `sdkmcp.Transport` session. The SDK Streamable HTTP server path should be based on `mcp.NewStreamableHTTPHandler`, which creates/connects `StreamableServerTransport` instances per HTTP session. The revised plan should state how the existing shared `*sdkmcp.Server` / registry setup will be reused without duplicating tool/resource/prompt registration, and how `defaultStartServer` dispatches between stdio and HTTP when `Config.Transport == http`.

2. **Define the HTTP serving surface.**
   Choose and document the handler mount path, e.g. `/mcp` or root, and the resulting local URL. State the intended `StreamableHTTPOptions` explicitly: whether it is stateful or stateless, whether JSON responses are enabled, logger use, session timeout choice, and that SDK localhost/cross-origin protections are not disabled.

3. **Make lifecycle and graceful shutdown concrete.**
   The plan needs an implementation shape for `net.Listen`, `http.Server`, cancellation, and shutdown: listener creation from `Config.HTTPBindAddress`, `BaseContext` or request context propagation, goroutine/error handling for `Serve`, `Shutdown` on context cancellation with a bounded timeout, fallback `Close` if shutdown times out, and normalization of `http.ErrServerClosed` so expected shutdown is not returned as a failure.

4. **Plan for testability without fixed-port flakes.**
   Add a test seam such as serving on an injected `net.Listener` or an internal helper that can use `127.0.0.1:0` in tests, while keeping user config validation strict. Step 2 should have focused tests for HTTP startup/smoke initialize, app transport dispatch, context cancellation closing the listener, and graceful shutdown. Full parity tests can remain Step 4, but Step 2 should not land untested listener/lifecycle code.

5. **Carry forward the Step 1 bind-address edge case.**
   R003 noted that values with internal whitespace can pass validation but may fail at listen time. The Step 2 plan should either normalize `Config.HTTPBindAddress` before storing/using it or reject that form during validation, so bind validation and actual listener behavior do not diverge.

Once these details are added to `STATUS.md` (or an equivalent plan note), the step will be ready to implement.
