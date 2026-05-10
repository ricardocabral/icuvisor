# TP-003 — Status

**Issue:** v0.1 — MCP stdio
**Iteration:** 1
**Current Step:** Step 3: Add registry/test tool scaffolding
**Last Updated:** 2026-05-10
**State:** Ready

## Step 1: SDK spike and plan

**Status:** ✅ Complete

- [x] Read official Go MCP SDK docs/examples
- [x] Record chosen SDK APIs and limitations
- [x] Confirm SDK license is permissive
- [x] Define tiny internal registry interface
- [x] R001: Record registry contract and SDK boundary decision
- [x] R001: Record panic-to-error safe registrar strategy
- [x] R001: Record dependency transitive license scan result
- [x] R001-code: Include all modules from a full SDK dependency graph license scan

## Step 2: Add the MCP SDK and stdio server skeleton

**Status:** ✅ Complete

- [x] Add official Go MCP SDK dependency
- [x] Create internal MCP server constructor
- [x] Wire stdio as default v0.1 behavior
- [x] Keep `main` thin
- [x] Honor context cancellation and return errors instead of panicking

## Step 3: Add registry/test tool scaffolding

**Status:** 🟡 In Progress

- [ ] Define tool registration contract
- [ ] R001: Define sanitized tool error contract
- [ ] R001: Document fake tool as test-only scaffolding
- [ ] R001: Record registrar invariants including duplicate-name rejection
- [ ] R001: Record v0.1 response content shape
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
| 2026-05-10 | Tiny registry interface decision: `internal/tools` will own SDK-free contracts (`Registry.Register(context.Context, Registrar) error`, `Registrar.AddTool(Tool) error`, `Tool{Name, Description, InputSchema, Handler}`); `internal/mcp` will adapt those definitions to SDK `mcp.Tool`/handlers. | Keeps tool packages testable and avoids leaking official SDK types outside the MCP adapter boundary. |
| 2026-05-10 | Registry/MCP boundary plan: `internal/mcp` constructor will take config, version, logger, `tools.Registry`, and an injectable SDK transport; production uses `mcp.StdioTransport`, tests use in-memory/IO transports. | Preserves thin `main`, avoids Claude Desktop in tests, and keeps future real tools under `internal/tools`. |
| 2026-05-10 | Safe registrar strategy: all tool registration goes through an `internal/mcp` adapter that validates snake_case names and required schemas before calling SDK registration, wraps SDK `AddTool`/`Server.AddTool` in a `defer recover` boundary, and returns startup errors instead of allowing SDK panics to escape. | Satisfies no-panic-outside-main rule while still using SDK schema validation and typed handler behavior. |
| 2026-05-10 | Full dependency graph license scan for a temp module importing `github.com/modelcontextprotocol/go-sdk/mcp@v1.3.1` (`go list -m all`): SDK MIT; `cloud.google.com/go/compute/metadata` Apache-2.0; `golang-jwt/jwt/v5` MIT; `google/go-cmp` BSD-3-Clause; `google/jsonschema-go` MIT; `segmentio/encoding` MIT; `yosida95/uritemplate/v3` BSD-3-Clause; `golang.org/x/oauth2`, `x/tools`, `x/sys` BSD-3-Clause; `segmentio/asm` MIT. | No GPL/copyleft modules found among the full newly introduced SDK dependency graph. |
| 2026-05-10 23:14 | Review R001 | plan Step 1: REVISE |
| 2026-05-10 23:19 | Review R001 | plan Step 1: APPROVE |
| 2026-05-10 23:22 | Review R001 | code Step 1: REVISE |
| 2026-05-10 23:24 | Review R001 | code Step 1: APPROVE |
| 2026-05-10 23:27 | Review R001 | plan Step 2: APPROVE |
| 2026-05-10 23:35 | Review R001 | code Step 2: APPROVE |
| 2026-05-10 23:37 | Review R001 | plan Step 3: REVISE |
