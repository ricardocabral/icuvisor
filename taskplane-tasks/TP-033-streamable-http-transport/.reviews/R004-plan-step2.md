# Plan Review: Step 2 — Streamable HTTP transport

## Verdict

APPROVE

## Notes

The Step 2 plan in `STATUS.md` is appropriately scoped and satisfies the prompt requirements for this phase:

- It keeps `internal/mcp.NewServer` as the single shared SDK server/registry constructor, avoiding forked tool/resource/prompt handler logic across transports.
- It uses the official Go SDK Streamable HTTP handler mounted at a stable `/mcp` endpoint.
- It separates production listener ownership (`RunStreamableHTTP`) from testable serving over an injected listener (`ServeStreamableHTTP`).
- The lifecycle plan covers context-rooted requests, expected `http.ErrServerClosed` handling, bounded graceful shutdown, fallback close, and listener-close verification.
- The proposed tests cover the Step 2 essentials: app dispatch into HTTP mode, HTTP initialize smoke, and cancellation closing the listener. Broader protocol parity remains correctly deferred to Step 4.

## Carry-forward checks for implementation/code review

1. Do not duplicate registry construction or transport-specific handler behavior; HTTP and stdio must share the same `*mcp.Server` setup path.
2. Keep HTTP logs limited to version/transport/address/path-style metadata; no config dumps, API keys, athlete IDs, or request/response payloads.
3. Ensure cancellation returns a predictable error shape (`context.Canceled` is acceptable) while treating `http.ErrServerClosed` as normal shutdown.
4. Avoid disabling SDK/localhost protections unless a later reviewed change explicitly justifies it for LAN opt-in behavior.

No blocking revisions are required before implementation.
