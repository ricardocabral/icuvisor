# Code Review — TP-003 Step 4

## Summary

REVISE — the new protocol tests compile and pass, but they leave two Step 4 coverage gaps: they bypass the icuvisor `Server.Run`/transport wrapper, and they mark malformed-request coverage complete without actually sending a malformed MCP/JSON-RPC request.

## Findings

### P2 — Protocol tests bypass the production `Server.Run` path

**File:** `internal/mcp/protocol_test.go:136-137`

`connectTestClient` builds an icuvisor server with `NewServer`, but then connects directly to the embedded SDK server via `server.server.Connect(ctx, serverTransport, nil)`. That means the positive initialize/list/call tests do not exercise `Server.Run` or the `Options.Transport` injection path that production uses. A regression in the wrapper, such as ignoring `s.transport`, returning early, or failing to call `sdk.Server.Run`, would not be caught by these Step 4 protocol tests.

Please wire the helper through the public wrapper instead: create the SDK in-memory server/client transports, pass the server transport as `Options.Transport`, start `server.Run(ctx)` in a goroutine, and connect the SDK client to the client transport. Cleanup can close the client/cancel the context and accept the expected close/cancellation error from the goroutine. This still avoids Claude Desktop while covering the actual icuvisor serving path.

### P2 — “Malformed requests” are marked covered but no malformed request is tested

**Files:**
- `internal/mcp/protocol_test.go:101-104`
- `taskplane-tasks/TP-003-mcp-stdio/STATUS.md:49-52`

`TestProtocolMalformedRequestsAndHandlerErrors` only sends a well-formed `tools/call` for an unknown tool plus a well-formed call whose handler returns an error. The unknown-tool case is useful, but it is not a malformed MCP/JSON-RPC request, so the Step 4 checklist item “Verify malformed requests...” is not actually satisfied even though `STATUS.md` marks it checked.

Please add an actual malformed-request test, for example using `sdkmcp.IOTransport`/pipes or a raw transport to send invalid newline-delimited JSON or invalid JSON-RPC/MCP params, then assert the client observes a short protocol error without internal details. Alternatively, if the SDK cannot practically expose this in-process, leave the STATUS item unchecked and document the limitation.

## Verification

- Ran `go test ./...` — passed.
