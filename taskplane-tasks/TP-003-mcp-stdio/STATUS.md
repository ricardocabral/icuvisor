# TP-003 — Status

**Issue:** v0.1 — MCP stdio
**Iteration:** 1
**Current Step:** Step 1: SDK spike and plan
**Last Updated:** 2026-05-10
**State:** Ready

## Step 1: SDK spike and plan

**Status:** 🟡 In Progress

- [x] Read official Go MCP SDK docs/examples
- [x] Record chosen SDK APIs and limitations
- [x] Confirm SDK license is permissive
- [ ] Define tiny internal registry interface
- [ ] R001: Record registry contract and SDK boundary decision
- [ ] R001: Record panic-to-error safe registrar strategy
- [ ] R001: Record dependency transitive license scan result

## Step 2: Add the MCP SDK and stdio server skeleton

**Status:** ⬜ Not started

- [ ] Add official Go MCP SDK dependency
- [ ] Create internal MCP server constructor
- [ ] Wire stdio as default v0.1 behavior
- [ ] Keep `main` thin
- [ ] Honor context cancellation and return errors instead of panicking

## Step 3: Add registry/test tool scaffolding

**Status:** ⬜ Not started

- [ ] Define tool registration contract
- [ ] Add fake/noop tool sufficient for protocol tests
- [ ] Ensure tool names are snake_case and stable
- [ ] Ensure user-facing errors are short

## Step 4: Test protocol behavior without Claude

**Status:** ⬜ Not started

- [ ] Test MCP initialize
- [ ] Test tool listing
- [ ] Test tool call dispatch
- [ ] Test malformed requests and handler errors

## Step 5: Verify and document

**Status:** ⬜ Not started

- [ ] Run `go fmt ./...`
- [ ] Run `go mod tidy`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

| 2026-05-10 23:08 | Task started | Runtime V2 lane-runner execution |
| 2026-05-10 23:08 | Step 1 started | SDK spike and plan |
| 2026-05-10 | Official Go MCP SDK APIs chosen: `mcp.NewServer`, `mcp.AddTool`, `(*mcp.Server).Run`, `mcp.StdioTransport`, `mcp.IOTransport`, `mcp.NewInMemoryTransports`, `mcp.NewClient`, `ClientSession.ListTools`, and `ClientSession.CallTool`. | Stdio server can be tested without Claude via SDK client/in-memory transports; production default uses newline-delimited JSON over stdin/stdout. |
| 2026-05-10 | SDK limitation: latest v1.6.0 requires Go 1.25, while icuvisor currently declares Go 1.23; v1.3.1 is the newest stable SDK release whose module declares Go 1.23 and has the needed stdio/tool APIs. `StdioTransport` hard-wires `os.Stdin`/`os.Stdout`; tests should use `IOTransport` or `NewInMemoryTransports`. | Pin `github.com/modelcontextprotocol/go-sdk` to v1.3.1 unless project Go version is intentionally raised. |
| 2026-05-10 | `github.com/modelcontextprotocol/go-sdk` v1.3.1 `LICENSE` is MIT; latest v1.6.0 is transitioning to Apache-2.0 with unrelicensed MIT contributions. | Both MIT and Apache-2.0 are permissive and compatible with project policy; v1.3.1 keeps the dependency wholly MIT. |
| 2026-05-10 23:14 | Review R001 | plan Step 1: REVISE |
