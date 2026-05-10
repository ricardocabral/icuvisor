# TP-001 — v0.1 foundation: project layout, CLI shape, and config contract

## Mission

Turn the existing scaffolding into a usable v0.1 foundation for the walking skeleton. Define the minimal package boundaries, CLI modes, version propagation, and manual configuration contract that later tasks can build on without inventing their own shape.

Roadmap item: **Go module + project layout** and prerequisite for **manual JSON config**.

Complexity: Blast radius 2, Pattern novelty 1, Security 1, Reversibility 1 = 5 → Review Level 1. Size: M.

## Dependencies

- **None**

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` — especially §7.2.H Configuration and §7.3 Technology
- `ROADMAP.md` — v0.1 only
- `CONTRIBUTING.md`
- `SECURITY.md`
- `README.md`
- `cmd/icuvisor/main.go`
- `Makefile`
- `go.mod`
- Local `.env` file, if present, for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` values used only during local end-to-end validation; never print or commit secrets

## File Scope

Expected files:

- `cmd/icuvisor/main.go`
- `internal/config/` files and tests
- `internal/version/` or equivalent internal package if useful
- `README.md` only if the v0.1 command shape changes user-visible docs
- `CHANGELOG.md`
- `taskplane-tasks/TP-001-v01-foundation/STATUS.md`

Avoid touching intervals.icu client, MCP SDK wiring, and tool implementation beyond interfaces/stubs needed to compile.

## Steps

### Step 1: Capture the foundation plan in STATUS.md

- [ ] Inspect the current module, empty internal directories, Makefile, README, and CI expectations
- [ ] Decide the minimal CLI shape for v0.1 (`version`, default stdio server mode, config path/env handling)
- [ ] Decide the internal package boundaries that later tasks must use
- [ ] Write the plan in `STATUS.md` before changing source files

### Step 2: Implement the CLI and version foundation

- [ ] Keep `icuvisor version` working and covered by a test where practical
- [ ] Add a default server startup path that delegates to an internal package instead of leaving logic in `main`
- [ ] Ensure build-time version can be passed to lower layers for `User-Agent` and `_meta.server_version`
- [ ] Return errors from internal packages; only `main` may decide exit code and stderr output

### Step 3: Implement minimal manual config loading

- [ ] Define a typed config struct for v0.1 inputs: intervals.icu API key, athlete ID, timezone, optional API base URL, optional HTTP timeout, optional config file path
- [ ] Load config from a manual JSON file and/or env vars suitable for MCP client config; document precedence in code/tests
- [ ] For local developer convenience, support reading the untracked `.env` file or document the exact command to export `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` from it before running end-to-end validation
- [ ] Normalize athlete IDs in one place (`12345` and `i12345` accepted; emit `i12345`)
- [ ] Do not store API keys in plaintext; v0.1 may read a key from env/manual client config but must not write it to disk
- [ ] Never log or echo API keys

### Step 4: Add tests for foundation behavior

- [ ] Add table-driven tests for athlete-ID normalization
- [ ] Add table-driven tests for config loading, validation, defaults, and secret redaction behavior
- [ ] Add tests for invalid/missing config returning short actionable errors

### Step 5: Verify and document

- [ ] Run `gofmt`/`go fmt ./...`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available locally
- [ ] Update `CHANGELOG.md` under `[Unreleased]`
- [ ] Update `STATUS.md` with results and mark this task done only when all checks pass

## Acceptance Criteria

- The binary has a thin `main` and a testable internal startup/config layer.
- Later tasks can construct an intervals client and MCP server from one typed config object.
- Athlete IDs are normalized centrally.
- Secrets are redacted from errors/loggable strings.
- `make build` and `make test` pass.

## Do NOT

- Do not add keychain storage; that is v0.5.
- Do not print, log, commit, or copy local `.env` values into docs/tests; `.env` is only for local validation inputs.
- Do not implement onboarding UI, installer behavior, tray icon, auto-update, Streamable HTTP, or full tool catalog.
- Do not read GPL source or add GPL dependencies.
- Do not put reusable packages in `pkg/` for v0.1.
- Do not accept API keys as MCP tool parameters.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

Check if affected:

- `README.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-001`, for example: `TP-001 add v0.1 config loader`.

---

## Amendments

_Add amendments below this line only._
