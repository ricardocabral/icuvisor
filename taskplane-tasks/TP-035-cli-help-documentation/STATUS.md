# TP-035-cli-help-documentation: CLI help documentation — Status

**Current Step:** Step 5: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S

---

### Step 1: Help-text design

**Status:** ✅ Complete

- [x] Draft the `--help` output as a fixture before writing code
- [x] Confirm env-var list against `internal/config/`, `internal/safety/`, `internal/response/`
- [x] Document exit codes

### Step 2: Parser changes

**Status:** ✅ Complete

- [x] Recognize `--help`, `-h`, `help`
- [x] Per-subcommand help (`version --help`)
- [x] Unknown-flag errors include usage hint and exit code 2
- [x] Preserve `--flag value` and `--flag=value` parsing
- [x] All I/O routed through `opts.Stdout` / `opts.Stderr`

### Step 3: Tests

**Status:** ✅ Complete

- [x] Table-driven coverage of help flags / subcommand / errors
- [x] Golden-file fixture for `--help` output
- [x] Env-var names cross-checked against resolved config

### Step 4: Documentation

**Status:** ✅ Complete

- [x] README pointer to `--help`
- [x] CHANGELOG `[Unreleased]` entry

### Step 5: Verify

**Status:** ✅ Complete

- [x] `make build`, `make test`, `make test-race`, `make lint`
- [x] Manual smoke of each help/error path

---

## Decisions

- **Library choice:** TBD in Step 1. Default recommendation in PROMPT.md is stdlib `flag` (CLAUDE.md "stdlib first"); escalate to cobra/urfave only if the help template becomes unreadable.

## Notes

- Step 1 env-var audit: `internal/config/config.go` defines `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_CONFIG`, `ICUVISOR_TIMEZONE`, `ICUVISOR_API_BASE_URL`, `ICUVISOR_HTTP_TIMEOUT`, `ICUVISOR_TRANSPORT`, `ICUVISOR_HTTP_BIND`, and `ICUVISOR_ENV_FILE`; `internal/safety` adds `ICUVISOR_DELETE_MODE` and `ICUVISOR_TOOLSET`; `internal/response` adds `ICUVISOR_DEBUG_METADATA`.
- Library choice: stdlib/hand-rolled parser retained; current command surface is small and no CLI framework dependency is justified.
- Exit-code design documented in `internal/app/testdata/help.golden`: 0 for success/help/version, 2 for CLI usage errors, 1 for runtime startup/config/server errors.
- Verification: `make build` passed; `make test` initially hit a transient `internal/mcp TestProtocolTransportParity` shutdown timeout then passed on rerun; `make test-race` hit the same transient once then passed on rerun; `make lint` passed with 0 issues.
- Manual smoke: `./bin/icuvisor --help`, `-h`, `help`, `version --help`, and `version` exited 0; `./bin/icuvisor --bogus` exited 2 with the usage hint; `./bin/icuvisor --help | head` showed normal help text and no `panic:` or wrapped error noise.

| 2026-05-15 13:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:18 | Step 1 started | Help-text design |
| 2026-05-15 13:21 | Review R001 | plan Step 1: APPROVE |
| 2026-05-15 13:25 | Review R002 | plan Step 2: APPROVE |
| 2026-05-15 13:30 | Review R003 | plan Step 3: APPROVE |
| 2026-05-15 13:33 | Review R004 | plan Step 4: APPROVE |
