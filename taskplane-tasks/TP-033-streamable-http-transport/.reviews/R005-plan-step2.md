# Plan Review: Step 2 — Streamable HTTP transport

**Verdict:** REVISE

I read `PROMPT.md`, `STATUS.md`, the prior Step 2 review, and the current Step 1 implementation in `internal/app`, `internal/config`, and `internal/mcp`. The revised Step 2 plan is much closer: it now covers the shared registry, `/mcp` endpoint, listener lifecycle, cancellation/shutdown, test seams, and the bind-address whitespace edge case.

However, two plan details still need tightening before implementation.

## Blocking gaps

1. **Correct the SDK handler construction shape.**
   The plan says to mount `mcp.NewStreamableHTTPHandler(sharedSDKServer, options)`, but the SDK version in this repo (`github.com/modelcontextprotocol/go-sdk v1.4.1`) exposes:

   ```go
   mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server, *mcp.StreamableHTTPOptions)
   ```

   The implementation plan should explicitly say that the HTTP runner passes a `getServer` callback returning the already-constructed shared `*sdkmcp.Server` (or `nil` on rejected requests, if any are added later). This matters because `internal/mcp.Server.Run` is for one transport session, while Streamable HTTP uses the handler to create/manage per-session transports. The plan should avoid implying the wrapped server can be passed directly.

2. **Make the Streamable HTTP options concrete.**
   R004 asked for explicit option choices. The revised plan still says “stateful sessions unless the SDK default is safer” and “a bounded session timeout if the SDK option is available.” In the current SDK the option is available. Pick the values in the plan so the worker does not have to decide while coding, for example:

   - `Stateless: false` (stateful sessions, matching SDK default) unless there is a specific reason to choose stateless.
   - `JSONResponse: false` unless a client compatibility reason requires JSON responses.
   - `Logger: logger`.
   - `SessionTimeout: <chosen duration>` such as `30 * time.Minute`.
   - `DisableLocalhostProtection: false` and `CrossOriginProtection: nil` to keep SDK protections enabled.

## Non-blocking clarification

- The bind-address whitespace item can be implemented either by normalizing before storing/listening or by rejecting internal whitespace, but the plan should preferably choose one. Normalizing to the canonical `netip.AddrPort` string during config validation would make `Config.HTTPBindAddress`, warning logs, and `net.Listen` use the same address.

After these updates are reflected in `STATUS.md`, the Step 2 plan will be ready to implement.
