# TP-003 — MCP stdio transport and server registry

## Mission

Wire the v0.1 MCP server over stdio using the official Go SDK, with a small registry that later exposes `get_athlete_profile`. The transport should be testable without Claude Desktop and ready for manual macOS Claude Desktop smoke testing in TP-005.

Roadmap item: **MCP stdio transport wired up via `github.com/modelcontextprotocol/go-sdk`**.

Complexity: Blast radius 2, Pattern novelty 2, Security 1, Reversibility 1 = 6 → Review Level 2. Size: M.

## Dependencies

- **TP-001** — config/version foundation

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` — especially §7.2.B MCP transports, §7.2.D response shaping, §7.2.E toolset tiers, §7.3 Technology, §7.4 #7 and #14
- `ROADMAP.md` — v0.1 only
- `CONTRIBUTING.md`
- `internal/config/` from TP-001
- Official `github.com/modelcontextprotocol/go-sdk` documentation/examples before changing transport code
- `hhopke/intervals-icu-mcp` (MIT, Python) may be consulted as a non-dependency reference for MCP tool metadata/behavior patterns; record consulted behavior in `STATUS.md`

## File Scope

Expected files:

- `internal/mcp/` server/stdio implementation and tests
- `internal/tools/` only for a minimal interface/registry contract, not real tool logic
- `cmd/icuvisor/main.go` only to call the internal server startup path
- `go.mod` / `go.sum` for the official MCP SDK dependency
- `CHANGELOG.md`
- `taskplane-tasks/TP-003-mcp-stdio/STATUS.md`

Do not implement intervals.icu API calls or `get_athlete_profile` tool behavior here.

## Steps

### Step 1: SDK spike and plan

- [ ] Read official Go MCP SDK docs/examples for stdio server setup and tool registration
- [ ] Record the chosen SDK APIs and any limitations in `STATUS.md`
- [ ] Confirm the SDK license is permissive and record it in `STATUS.md`
- [ ] Define a tiny internal registry interface that `internal/tools` can implement

### Step 2: Add the MCP SDK and stdio server skeleton

- [ ] Add `github.com/modelcontextprotocol/go-sdk` to `go.mod`
- [ ] Create an internal MCP server constructor that takes config, version, logger, and tool registry dependencies
- [ ] Wire stdio transport as the default binary behavior for v0.1
- [ ] Keep `cmd/icuvisor/main.go` thin: parse high-level command, load config, run stdio server, handle errors
- [ ] Ensure server startup honors context cancellation and returns errors instead of panicking

### Step 3: Add registry/test tool scaffolding

- [ ] Define how tools register names, descriptions, input schema, handler functions, and response content
- [ ] Add a test-only/noop tool or registry fake sufficient to test MCP initialize/tools/list/tools/call behavior without intervals.icu
- [ ] Ensure tool names are snake_case and stable
- [ ] Ensure user-facing errors are short and internal details are not returned to the LLM

### Step 4: Test protocol behavior without Claude

- [ ] Add unit or integration tests using in-memory pipes/stdin-stdout buffers where practical
- [ ] Verify MCP initialize succeeds
- [ ] Verify tool listing returns registered tools
- [ ] Verify tool call dispatch reaches the registry handler
- [ ] Verify malformed requests and handler errors produce short actionable MCP errors

### Step 5: Verify and document

- [ ] Run `go fmt ./...`
- [ ] Run `go mod tidy`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md` and `STATUS.md`

## Reference Implementation Policy

- You may use `hhopke/intervals-icu-mcp` as a permissively licensed, Python-only reference implementation for intervals.icu API reverse-engineering and MCP tool semantics. It must not be added as a dependency. When reading it, extract behavior and API contracts, then implement idiomatic Go from first principles. Record the consulted files/behaviors in `STATUS.md`.
- You may not use `mvilanova/intervals-mcp-server` source code as an implementation reference because it is GPLv3 and this project is MIT clean-room. Do not read, copy, paraphrase, transliterate, or port its code. Use the PRD's summarized insights, intervals.icu public docs, black-box testing, and permissively licensed references instead.

## Acceptance Criteria

- Running `icuvisor` starts an MCP stdio server rather than exiting with `not yet implemented` once valid v0.1 config is supplied.
- MCP protocol behavior is test-covered without a real MCP client.
- Official Go MCP SDK dependency is added and its license is noted.
- No Streamable HTTP, SSE, or installer behavior is added.

## Do NOT

- Do not implement SSE; the PRD says SSE is deprecated and out of scope.
- Do not implement Streamable HTTP; that is v0.5.
- Do not expose LAN HTTP bindings.
- Do not implement all v1 tools or toolset tiers in this task.
- Do not read, copy, paraphrase, transliterate, or port GPL implementation code, including `mvilanova/intervals-mcp-server`.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

Check if affected:

- `README.md` only if user-facing command behavior changes

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-003`, for example: `TP-003 wire mcp stdio server`.

---

## Amendments

_Add amendments below this line only._
