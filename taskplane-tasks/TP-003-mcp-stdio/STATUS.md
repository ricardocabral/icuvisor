# TP-003 — Status

**Issue:** v0.1 — MCP stdio
**State:** Ready

## Step 1: SDK spike and plan

**Status:** ⬜ Not started

- [ ] Read official Go MCP SDK docs/examples
- [ ] Record chosen SDK APIs and limitations
- [ ] Confirm SDK license is permissive
- [ ] Define tiny internal registry interface

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
