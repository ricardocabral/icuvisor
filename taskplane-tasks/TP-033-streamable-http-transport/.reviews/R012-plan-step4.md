# Plan Review: Step 4 — Parity tests

**Verdict:** REVISE

I read `PROMPT.md`, `STATUS.md`, and the current `internal/mcp` protocol/HTTP test coverage. The Step 4 section in `STATUS.md` currently only repeats the task checklist; it does not yet describe how parity will be implemented. Because this step is the acceptance gate that proves HTTP shares the same server core as the existing transport path, the plan needs more detail before implementation.

## Blocking gaps

1. **Define the shared protocol-test harness.**
   The plan should say how the existing protocol tests in `internal/mcp/protocol_test.go` will be run for both transports without copy/paste. A good shape would be a small transport/session factory abstraction with cases for the existing non-HTTP/stdio-equivalent test path and Streamable HTTP. The HTTP case should use `net.Listen("tcp", "127.0.0.1:0")`, `ServeStreamableHTTP`, and `sdkmcp.StreamableClientTransport` pointed at `http://<addr>/mcp`; no fixed ports.

2. **List the protocol scenarios that will run through both transports.**
   The plan needs to explicitly cover the Step 4 checklist: initialize, `tools/list`, successful tool call, unknown/missing tool, sanitized handler errors, `resources/list`, `resources/read`, missing resource, sanitized resource errors, and prompts behaviour. If prompts are not yet registered by icuvisor, still assert transport parity for the current SDK-visible prompt behaviour (for example, `prompts/list` returns the same empty result/error shape across transports).

3. **Clarify malformed-request coverage per transport framing.**
   Existing `TestProtocolMalformedRawRequest` writes newline-delimited JSON-RPC over a pipe, which is not the same framing as Streamable HTTP. The plan should keep that raw stdio/IO malformed test and add an HTTP-specific malformed POST to `/mcp` that asserts a protocol/HTTP error is returned without panic/secret leakage. For shared parity, also include malformed protocol-level cases that can be exercised through the SDK client on both transports, such as unknown tool and sanitized handler failure.

4. **Specify the byte-identical comparison strategy.**
   “Byte-identical across transports” needs an implementation plan. Add a parity test that drives the same deterministic registry/resource registry through both transports, serializes stable SDK results to canonical JSON, and compares the bytes/strings for practical results: initialize server info/capabilities, `tools/list`, tool-call result content/structured content, `resources/list`, `resources/read`, and `prompts/list`. Normalize or exclude transport/session-specific fields if any exist; do not weaken this to only checking counts/names.

5. **Account for lifecycle and race-safety in the HTTP test harness.**
   Each HTTP parity case should create its own server instance/listener and clean up with `session.Close()`, context cancellation, and `waitForServerRun`. Use short client HTTP timeouts and `MaxRetries: -1` as in the existing HTTP smoke test so failures do not hang. Avoid sharing mutable registries or log buffers across parallel subtests unless protected.

## Suggested STATUS.md addition

Add a Step 4 plan along these lines:

- Refactor protocol tests to use a `connectProtocolClient(t, transportKind, opts)` helper with two transport cases: current in-memory/stdio-equivalent transport and Streamable HTTP on `127.0.0.1:0` at `StreamableHTTPPath`.
- Run a table-driven protocol suite for both transports covering initialize, tools/list, tool calls, missing tools, sanitized tool errors, resources/list/read/missing/sanitized errors, and current prompt-list behaviour.
- Add a parity assertion test that captures canonical JSON for stable responses from both transports and compares bytes, excluding only unavoidable transport/session-specific metadata.
- Keep the existing raw newline malformed-request test for the stdio/IO framing and add an HTTP malformed-request test for a bad POST to `/mcp`, asserting no panic/internal detail/secret leakage.
- Keep all HTTP tests on ephemeral loopback listeners and cleanly close sessions/listeners to satisfy `go test -race ./internal/mcp`.

Once those details are captured in `STATUS.md`, the Step 4 plan should be ready to implement.
