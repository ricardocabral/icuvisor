# Review R002 — Plan Review for Step 1

Verdict: **Approved for implementation**

The revised Step 1 plan now addresses the blocker from R001. It explicitly requires raw in-process HTTP wire assertions rather than SDK-decoded client results, accounts for the current Streamable HTTP SSE response mode, and models the initialize → initialized notification → ping session lifecycle with Codex-like headers and loopback-only binding.

## What looks sufficient

- Raw HTTP assertions will inspect the actual response envelope (`jsonrpc`, matching `id`, `result`) that Codex compatibility depends on.
- The plan covers both supported response shapes for the current server configuration: JSON bodies and SSE `data:` JSON-RPC envelopes.
- Ping is planned after a real initialized session, including `Mcp-Session-Id`, so it should exercise the same path a Streamable HTTP client uses.
- The negative shape check is explicit enough to fail on bare `{}`, string, or `null` payloads before looking inside `result`.
- Scope remains test-only for Step 1, with transport/server changes deferred to Step 2 only if the smoke tests expose a failure.

## Non-blocking implementation reminders

- Put the tests in `internal/mcp/protocol_test.go` unless reuse strongly favors another existing test file.
- Name tests with `StreamableHTTP` and `JSONRPC` so the requested targeted regex includes them.
- Send `Content-Type: application/json`, `Accept: application/json, text/event-stream`, and `Mcp-Protocol-Version` on post-initialize requests.
- Assert the initialized notification returns the expected HTTP success status (typically `202 Accepted`) without requiring a JSON-RPC response.

Proceed with Step 1 implementation.
