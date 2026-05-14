# Code Review: Step 4 — Parity tests

**Verdict:** APPROVE

I reviewed the Step 4 working-tree changes in `internal/mcp/protocol_test.go` plus the `STATUS.md` update. The new shared protocol suite runs the same initialize, tools, tool-call, resource, prompt, missing-object, and sanitized-error scenarios against the existing in-memory/stdio-equivalent path and Streamable HTTP. The HTTP harness uses an ephemeral loopback listener, `/mcp`, `sdkmcp.StreamableClientTransport`, short client timeouts, and cleanup via session close, context cancellation, and `waitForServerRun`.

The dedicated parity snapshot compares canonical JSON bytes for stable initialize/tool/resource/prompt results, and malformed coverage now keeps the raw IO framing test while adding an HTTP bad-POST test with leak/panic assertions.

Verification run:

- `go test ./internal/mcp`
- `go test -race ./internal/mcp`

No blocking issues found.

Note: `git diff ee5ce4c..HEAD` currently shows only `STATUS.md`; the Step 4 test implementation is present as an uncommitted working-tree change in `internal/mcp/protocol_test.go`.
