# TP-001 — Status

**Issue:** v0.1 — foundation
**Iteration:** 1
**Current Step:** Step 1: Capture the foundation plan in STATUS.md
**Last Updated:** 2026-05-10
**State:** Ready

## Step 1: Capture the foundation plan in STATUS.md

**Status:** ✅ Complete

- [x] Inspect current module/layout/Makefile/README/CI
- [x] Decide minimal CLI shape for v0.1
- [x] Decide internal package boundaries
- [x] Write plan before changing source files

**Plan:**
- Keep `cmd/icuvisor/main.go` as a thin entrypoint with build-time `main.version`; it should delegate all command handling to `internal/app` and only decide stderr/exit behavior.
- Implement `internal/app` with `version` command support and a default startup path representing stdio server mode. For TP-001, default startup should validate/load config and return a clear placeholder error until later MCP wiring lands.
- Propagate build version through `internal/app` options so future intervals User-Agent and MCP `_meta.server_version` can use one source of truth.
- Implement `internal/config` as the central v0.1 config contract: API key, normalized athlete ID, timezone, optional API base URL, optional HTTP timeout, optional config path; load JSON first then environment/.env overrides with tests documenting precedence.
- Redact secrets in any string/error surfaces and never write API keys to disk; `.env` support is read-only developer convenience and must not print values.

## Step 2: Implement the CLI and version foundation

**Status:** ⬜ Not started

- [ ] Keep `icuvisor version` working
- [ ] Delegate default startup from thin `main` to internal package
- [ ] Pass build version to lower layers
- [ ] Return errors from internal packages; handle exit in `main`

## Step 3: Implement minimal manual config loading

**Status:** ⬜ Not started

- [ ] Define typed v0.1 config inputs
- [ ] Load config from manual JSON and/or env with tested precedence
- [ ] Support/document safe local `.env` loading for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` without printing secrets
- [ ] Normalize athlete IDs centrally
- [ ] Do not write API keys to disk
- [ ] Never log or echo API keys

## Step 4: Add tests for foundation behavior

**Status:** ⬜ Not started

- [ ] Table-driven tests for athlete-ID normalization
- [ ] Table-driven tests for config loading/validation/defaults/redaction
- [ ] Tests for short actionable invalid/missing config errors

## Step 5: Verify and document

**Status:** ⬜ Not started

- [ ] Run `go fmt ./...`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

| 2026-05-10 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-05-10 21:28 | Step 1 started | Capture the foundation plan in STATUS.md |
| 2026-05-10 | Current foundation inspected | Repo has only `cmd/icuvisor/main.go`, no `internal/` packages/tests yet; Makefile/CI expect `go build ./...`, `go test -race ./...`, optional golangci-lint. |
| 2026-05-10 | v0.1 CLI shape decided | Keep `icuvisor version`; default invocation starts stdio MCP server path via internal app package; config path may come from flag/env, with env/manual JSON for credentials. |
| 2026-05-10 | Internal boundaries decided | Use thin `cmd/icuvisor`; `internal/app` for CLI/default startup and version propagation; `internal/config` for typed config, env/JSON/.env loading, ID normalization, validation, redaction; future `internal/intervals`, `internal/mcp`, and `internal/tools` consume config rather than parse env. |