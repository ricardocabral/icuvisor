# Plan Review — TP-003 Step 2: Add the MCP SDK and stdio server skeleton

Verdict: **approved to proceed to Step 2**.

I reviewed `PROMPT.md`, the current `STATUS.md`, the existing `internal/app` startup path, `cmd/icuvisor/main.go`, `go.mod`, the PRD MCP transport/response-shaping sections, and the roadmap v0.1 scope. The Step 2 plan is aligned with the task: add the official SDK, introduce an internal MCP constructor/run path, make stdio the default behavior, and keep real tool behavior for later steps.

## What I verified

- The Step 1 discoveries give Step 2 enough direction: pin `the MCP Go SDK/mcp` to **v1.3.1** to preserve the repository's current `go 1.23` baseline.
- The selected SDK APIs support the planned skeleton: `mcp.NewServer`, `(*mcp.Server).Run`, and `mcp.StdioTransport` are the right production path for stdio; `IOTransport`/in-memory transports should remain test seams.
- The current `cmd/icuvisor/main.go` is already thin and delegates through `internal/app`; Step 2 should preserve that shape rather than moving CLI/config logic into `main`.
- `internal/app.defaultStartServer` is the natural place to replace `ErrServerNotImplemented` with a call into `internal/mcp`.
- No Streamable HTTP, SSE, installer, or real intervals.icu tool behavior is planned for this step, which matches TP-003 scope.

## Guidance for implementation

- Keep stdout protocol-only while the stdio server is running. Any SDK logger or app diagnostics must write to stderr or a test/discard logger, never stdout.
- Make the registry dependency concrete enough in Step 2 for the server constructor to compile, even if Step 3 expands the contract and adds fake/noop tools. Do not leak MCP SDK types into `internal/tools`.
- Preserve a transport seam in `internal/mcp` so tests can use SDK in-memory/IO transports later. Production can use `&mcp.StdioTransport{}`.
- Be careful with `mcp.IOTransport`: it requires `io.ReadCloser` and `io.WriteCloser`, not plain `io.Reader`/`io.Writer`.
- Apply the recorded panic-to-error boundary anywhere Step 2 touches SDK registration/construction code that can panic. Startup should return wrapped errors, not panic.
- If adding SDK logging through `mcp.ServerOptions.Logger`, use a logger configured away from stdout and avoid logging raw config/API-key data.
- Do not raise the module Go version as part of this step unless the task is explicitly amended.

No blocking issues remain in the Step 2 plan.
