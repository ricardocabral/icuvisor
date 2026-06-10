# Review R001 — Plan Review for Step 1

Verdict: **Changes requested before implementation**

The step scope is correct, but the current plan/checklist is not concrete enough to guarantee coverage of the Codex regression. The key risk is writing another SDK-level client test: existing tests already prove `sdkmcp.StreamableClientTransport` can connect, but they do not inspect the wire payload that Codex rejected.

## Required plan adjustments

1. **Use raw in-process HTTP requests, not the MCP SDK client, for the new smoke assertions.**
   - Post JSON-RPC directly to `http://127.0.0.1:<port>/mcp` and decode the HTTP response body.
   - Assert the actual wire message has top-level `jsonrpc: "2.0"`, matching `id`, and `result` object.

2. **Account for the current Streamable HTTP response mode.**
   - `ServeStreamableHTTP` currently configures `JSONResponse: false`, so successful responses may be `text/event-stream` with `data: {json-rpc-envelope}` rather than an `application/json` body.
   - The test plan should explicitly parse either the configured SSE response and inspect the `data:` JSON-RPC message, or defer changing to JSON response mode to Step 2 if the new tests prove that is necessary. Do not accidentally assert only decoded SDK results.

3. **Model the full handshake/session lifecycle for ping.**
   - Send `initialize`, capture `Mcp-Session-Id`, send `notifications/initialized` with that session ID and expect `202 Accepted`, then send `ping` with the same session ID.
   - Include Codex-like headers: `Content-Type: application/json`, `Accept: application/json, text/event-stream`, and `Mcp-Protocol-Version` on post-initialize requests.

4. **Make the negative shape explicit.**
   - For both initialize and ping, reject a bare result payload (`{}`, string, or `null`) by checking for the envelope fields before looking inside `result`.
   - Assert no top-level `error` is present on the success path.

## Suggested placement

`internal/mcp/protocol_test.go` is the best home, reusing the existing loopback `ServeStreamableHTTP` setup helpers or adding a small raw-HTTP helper nearby. Test names should include `StreamableHTTP` and `JSONRPC` so they are picked up by the requested targeted `go test` regex.

Once these amendments are made, the plan is aligned with TP-162 Step 1 and should remain test-only unless the smoke tests fail.
