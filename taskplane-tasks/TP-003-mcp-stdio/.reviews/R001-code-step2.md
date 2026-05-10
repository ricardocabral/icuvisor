# Code Review — TP-003 Step 2: Add the MCP SDK and stdio server skeleton

Verdict: **APPROVE**

I reviewed the full diff from `0468667..HEAD` and the changed files in context.

## Checks run

- `git diff 0468667..HEAD --name-only`
- `git diff 0468667..HEAD`
- `go test ./...`
- `go build ./...`
- `gofmt -l internal/app/app.go internal/mcp/server.go internal/mcp/server_test.go internal/tools/registry.go`
- `go mod tidy` produced no `go.mod` / `go.sum` diff

## Findings

No blocking issues for Step 2.

The implementation adds the official SDK dependency, keeps the CLI path thin through `internal/app`, defaults production construction to `mcp.StdioTransport`, preserves a transport seam for later in-memory protocol tests, checks cancellation before construction/run, and contains SDK construction/registration panics behind error-returning boundaries.

## Follow-up notes for Step 3/4

- `internal/mcp/server.go:109-110` currently maps handler errors directly to MCP tool error text via `err.Error()`. That is acceptable to leave unexercised in this Step 2 skeleton, but before real or test tools are considered complete, handler failures should be converted to short user-safe messages and the detailed error should be logged outside stdout. This is required by the MCP-server convention that LLM-visible errors must not expose internal details.
- `internal/mcp/server.go:74` / `internal/app/app.go:108` return SDK run errors without additional context. Consider wrapping non-context errors with `%w` when Step 4 adds protocol error tests, while preserving `errors.Is` behavior for cancellation/connection-close cases.
