# TP-001 — Status

**Issue:** v0.1 — foundation
**Iteration:** 1
**Current Step:** Step 3: Implement minimal manual config loading
**Last Updated:** 2026-05-10
**State:** Ready

## Step 1: Capture the foundation plan in STATUS.md

**Status:** ✅ Complete

- [x] Inspect current module/layout/Makefile/README/CI
- [x] Decide minimal CLI shape for v0.1
- [x] Decide internal package boundaries
- [x] Write plan before changing source files

**Plan:**
- Keep `cmd/icuvisor/main.go` as a thin entrypoint with build-time `main.version`; it should call `app.Run(ctx, app.Options{Version: version, Args: os.Args[1:], Stdout: os.Stdout, Stderr: os.Stderr})` (or equivalent) and only decide final stderr/exit behavior.
- Implement `internal/app` with `version` command support and a default startup path representing stdio server mode. For Step 2, default startup should delegate to an injectable internal starter and return a short placeholder such as `stdio server not implemented yet`; config loading/validation belongs to Step 3.
- Test Step 2 at the `internal/app` layer with injected args/stdout/starter: `version` writes injected version and returns nil, default invocation delegates and returns starter error, and unknown commands return a short actionable error.
- Propagate build version through `internal/app.Options.Version` into an app/server startup config (for example `ServerInfo{Version: ...}`) so future intervals User-Agent and MCP `_meta.server_version` use one injected source of truth; do not add `internal/version` unless ldflags are changed.
- Implement `internal/config` as the central v0.1 config contract: API key, normalized athlete ID, timezone, optional API base URL, optional HTTP timeout, optional config path; load JSON first then environment/.env overrides with tests documenting precedence.
- Redact secrets in any string/error surfaces and never write API keys to disk; `.env` support is read-only developer convenience and must not print values.

## Step 2: Implement the CLI and version foundation

**Status:** ✅ Complete

- [x] R001 plan review: narrow Step 2 plan so default startup does not load/validate config yet
- [x] R001 plan review: explicitly test internal app `version`, default delegation error, and short unknown-command errors
- [x] R001 plan review: define `app.Run(ctx, Options{...})` entrypoint shape before coding
- [x] R001 plan review: make version injection to lower runtime/server config concrete without importing from `main`
- [x] Keep `icuvisor version` working
- [x] Delegate default startup from thin `main` to internal package
- [x] Pass build version to lower layers
- [x] Return errors from internal packages; handle exit in `main`

## Step 3: Implement minimal manual config loading

**Status:** 🟡 In Progress

- [x] R001 plan review: name JSON fields, env vars, and config path support
- [x] R001 plan review: document precedence defaults < JSON < `.env` absent-only < process env < CLI flags
- [x] R001 plan review: document validation/default behavior and short errors
- [x] R001 plan review: clarify default startup loads config while `version` remains config-free
- [x] R001 plan review: specify secret redaction for strings/errors/loggable structs
- [x] R001 plan review: scope read-only `.env` parsing to recognized keys
- [ ] Define typed v0.1 config inputs

**Step 3 config plan:**
- Public contract: `internal/config.Config` with JSON fields `api_key`, `athlete_id`, `timezone`, `api_base_url`, and `http_timeout`; env vars `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, `ICUVISOR_API_BASE_URL`, `ICUVISOR_HTTP_TIMEOUT`, and `ICUVISOR_CONFIG` for the file path.
- CLI support in v0.1 is limited to `--config <path>` / `--config=<path>` for default startup. There is no automatic platform config path yet; if no path/env is provided, loading uses defaults plus env/`.env`.
- Precedence: built-in defaults < JSON file selected by `ICUVISOR_CONFIG` or `--config` < local `.env` values applied only for keys still absent < process environment < CLI flags. `.env` never overrides explicit MCP-client/process env.
- Defaults/validation: `api_base_url` defaults to `https://intervals.icu/api/v1` and must be absolute `http`/`https`; `http_timeout` defaults to `30s`, is parsed as a Go duration string, and must be positive; `timezone` defaults to `UTC` and must load with `time.LoadLocation`; `api_key` and `athlete_id` are required. Errors stay short/actionable: set API key, set athlete ID, invalid timezone, invalid timeout, invalid API base URL.
- App integration: after Step 3, default startup parses `--config`, calls `config.Load(ctx, Options{Path: ...})`, and passes the typed config on `ServerInfo`; `icuvisor version` must return without touching config.
- Redaction: `Config.String()`/loggable summaries show `api_key=<redacted>` or empty status only; raw API keys are never included in errors, and the loader only reads user-provided JSON/env/`.env` without creating or writing credential files.
- `.env` scope: parse a local `.env` read-only with stdlib code, accept only recognized `INTERVALS_ICU_*` and `ICUVISOR_*` keys, ignore comments/unknowns, and never print loaded values.
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

## Notes

- R001 plan review requested narrowing Step 2 away from config loading, explicit app-level tests, a concrete `app.Run(ctx, Options{...})` shape, and injected version propagation.
- R001 plan review for Step 3 requested a concrete config contract, precedence, defaults/validation, app integration, redaction, and narrow `.env` scope before coding.

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

| 2026-05-10 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-05-10 21:28 | Step 1 started | Capture the foundation plan in STATUS.md |
| 2026-05-10 | Current foundation inspected | Repo has only `cmd/icuvisor/main.go`, no `internal/` packages/tests yet; Makefile/CI expect `go build ./...`, `go test -race ./...`, optional golangci-lint. |
| 2026-05-10 | v0.1 CLI shape decided | Keep `icuvisor version`; default invocation starts stdio MCP server path via internal app package; config path may come from flag/env, with env/manual JSON for credentials. |
| 2026-05-10 | Internal boundaries decided | Use thin `cmd/icuvisor`; `internal/app` for CLI/default startup and version propagation; `internal/config` for typed config, env/JSON/.env loading, ID normalization, validation, redaction; future `internal/intervals`, `internal/mcp`, and `internal/tools` consume config rather than parse env. |
| 2026-05-10 | Step 2 targeted tests passed | `go test ./cmd/icuvisor ./internal/app` passed after adding app Run tests and thin main delegation. |
| 2026-05-10 21:33 | Review R001 | plan Step 2: UNKNOWN |
| 2026-05-10 21:36 | Review R001 | plan Step 2: APPROVE |
| 2026-05-10 21:40 | Review R001 | plan Step 3: REVISE |
