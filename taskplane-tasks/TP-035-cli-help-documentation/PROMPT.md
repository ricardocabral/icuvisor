# TP-035 — CLI help documentation (`--help` for commands, flags, env vars)

## Mission

Add discoverable CLI documentation to the `icuvisor` binary following Go CLI best practices: a top-level `--help` / `-h` that lists subcommands, flags, environment variables, and exit codes; per-subcommand help (currently only `version`, but the surface will grow); and a one-line `Usage:` on unknown-flag errors. Today the binary's hand-rolled flag parser in `internal/app/app.go` accepts `--config`, `--transport`, `--http-bind`, and the `version` subcommand, but there is no way to discover them without reading the README. A user typing `icuvisor --help` gets `unknown command or flag "--help"`, which is a poor first-run experience.

PRD anchors: §7.2.D "stdio is default; Streamable HTTP transport opt-in", §7.4 #1 (one-binary install), §7.4 #6 (clear safety-mode signaling). PRD KR1 (install success) is partly about a user being able to verify the install works without reading docs — `--help` is a basic affordance for that.

ROADMAP positioning: This is preparatory polish for v0.5 (internal beta), where forum-recruited athletes will run the binary directly and need self-serve discovery. Land it standalone — it is independent of v0.5's keychain/installer work.

Complexity: Blast radius 1 (CLI surface only, no protocol changes), Pattern novelty 1 (standard Go CLI library usage), Security 1 (no new credential paths), Reversibility 1 = 4 → Review Level 1. Size: S.

## Dependencies

- **TP-030** — toolset tiers; `ICUVISOR_TOOLSET` must be documented in `--help` env-var section.
- **TP-033** — Streamable HTTP transport; `--transport` and `--http-bind` flags exist and need help text.
- **TP-018** — delete-mode safety gate; `ICUVISOR_DELETE_MODE` must be documented in `--help` env-var section.

No blocking deps beyond "the flags and env vars being documented must exist."

## Context to Read First

- `CLAUDE.md` — Go conventions, "default to `internal/`", no `panic` outside `main`, structured errors.
- `cmd/icuvisor/main.go` — current entrypoint (25 lines, calls `app.Run`).
- `internal/app/app.go` — the hand-rolled flag parser (`parseDefaultArgs`) and the `version` subcommand handling.
- `internal/app/app_test.go` — existing CLI tests; mirror their style.
- `README.md` — Quickstart, MCP transport, Delete/write safety mode, Toolset tiers sections — all env vars and flags that must appear in `--help` output.
- `docs/prd/PRD-icuvisor.md` §7.2.D for transport semantics.

## File Scope

Expected files:

- `internal/app/app.go` — replace `parseDefaultArgs` with a help-aware parser (or wire in a small dependency — see "Library choice" below).
- `internal/app/help.go` (new) — the help-text template / builder if hand-rolled.
- `internal/app/app_test.go` — tests for `--help`, `-h`, `help` subcommand, per-subcommand help, unknown-flag error includes usage hint, exit codes.
- `cmd/icuvisor/main.go` — only if exit-code handling needs adjustment for help (help should exit 0, errors non-zero).
- `README.md` — short pointer that `icuvisor --help` is the source of truth for flags/env vars; keep README focused on quickstart.
- `CHANGELOG.md` — under `[Unreleased]`.
- `taskplane-tasks/TP-035-cli-help-documentation/STATUS.md`.

Out of scope: shell completion (`bash`/`zsh`/`fish`), man pages, restructuring into a full subcommand tree. Leave those as follow-ups if a v0.5 reviewer asks.

## Library choice

Two acceptable approaches — pick one and justify in `STATUS.md`:

1. **Stdlib `flag` package** — zero new deps, fully aligned with CLAUDE.md "stdlib first". Customize `flag.Usage` to emit the sections below. Sufficient for the current ~3 flags + 1 subcommand surface.
2. **`spf13/cobra` (Apache-2.0) or `urfave/cli/v3` (MIT)** — both permissive-licensed (CLAUDE.md compatible). Worth it only if the team expects the subcommand tree to grow (e.g., `icuvisor config show`, `icuvisor doctor`). Adds a dependency for a binary the PRD wants kept lean.

**Default recommendation:** stdlib `flag`. The current surface does not justify a CLI framework, and the existing hand-rolled parser already supports `--flag value` and `--flag=value` forms that we must preserve. Only escalate to cobra/urfave if you find the help template becomes unreadable.

## Steps

### Step 1: Help-text design

- [ ] Draft the `--help` output as a fixture before writing code. Sections, in order: short description; `Usage:` line; `Commands:` (just `version` for now, plus implicit "no command = run server"); `Flags:` with type, default, and one-line description; `Environment variables:` (`INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TRANSPORT`, `ICUVISOR_HTTP_BIND`, `ICUVISOR_DELETE_MODE`, `ICUVISOR_TOOLSET`, `ICUVISOR_DEBUG_METADATA` if it exists); `Examples:` (stdio default, HTTP transport, config file); `Exit codes:`; one-line pointer to README + PRD for deeper docs.
- [ ] Confirm env-var list against `internal/config/`, `internal/safety/`, and `internal/response/` — do not invent variables, and do not omit ones that exist.
- [ ] Document exit codes: 0 success, 2 usage error (unknown flag / missing value), non-zero for runtime errors. Match Go convention.

### Step 2: Parser changes

- [ ] Recognize `--help`, `-h`, and `help` (subcommand form) at any position; print help to stdout and exit 0.
- [ ] Per-subcommand help: `icuvisor version --help` prints the version-subcommand stanza.
- [ ] Unknown-flag errors must include `Run 'icuvisor --help' for usage.` and exit with code 2 (usage error), not the generic non-zero code.
- [ ] Preserve existing `--flag value` and `--flag=value` parsing; do not regress.
- [ ] Keep all I/O routed through `opts.Stdout` / `opts.Stderr` so tests can capture it; do not write to `os.Stdout` directly from `internal/app`.

### Step 3: Tests

- [ ] Table-driven test in `internal/app/app_test.go`: `--help`, `-h`, `help`, `version --help`, unknown flag, missing flag value, valid flags still parse.
- [ ] Golden-file test for `--help` output (commit the fixture under `internal/app/testdata/`). Renaming a flag is a breaking change for users — the golden file makes that visible at review time.
- [ ] Verify env-var names in the help fixture match the real env vars by reading the resolved values from `internal/config` / `internal/safety` rather than hardcoded strings in the test, where practical.

### Step 4: Documentation

- [ ] `README.md`: add a one-line note under Quickstart that `./bin/icuvisor --help` lists all flags and env vars.
- [ ] `CHANGELOG.md`: `[Unreleased]` entry under "Added".
- [ ] Do **not** duplicate the full flag list across README and `--help`. The README points to `--help`; `--help` is the source of truth.

### Step 5: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Manual: run each of `./bin/icuvisor --help`, `-h`, `help`, `version --help`, `--bogus`, and `version` and eyeball the output.
- [ ] Confirm `./bin/icuvisor --help | head` does not show `panic:` or wrapped error noise.

## Acceptance Criteria

- `icuvisor --help`, `icuvisor -h`, and `icuvisor help` all print the same help text to stdout and exit 0.
- Help output covers: usage line, subcommands, flags (with defaults), environment variables (every one read by `internal/config` and `internal/safety`), examples, and exit codes.
- `icuvisor version --help` prints version-subcommand help and exits 0.
- Unknown flags print a short error to stderr that includes a `Run 'icuvisor --help' for usage.` hint and exit with code 2.
- Existing `--config`, `--transport`, `--http-bind` and the `version` subcommand still work unchanged.
- Golden-file test pins the `--help` output; any change to flags/env vars/help text shows up in review as a diff.
- README points to `--help`; CHANGELOG updated.

## Do NOT

- Do not add shell completions, man pages, or a `doctor` / `config` subcommand tree in this task — out of scope; file follow-ups if needed.
- Do not accept API keys as a CLI flag, even hidden. CLAUDE.md hard rule: credentials come from env / config file / keychain only.
- Do not document an env var that does not exist in the code, and do not omit one that does.
- Do not duplicate the flag list verbatim in README — `--help` is the source of truth and README should point to it.
- Do not change the public API of `app.Options` or `app.Run` in ways that ripple into the test seam used by `internal/app/app_test.go` without updating those tests.
- Do not regress the `--flag=value` inline form; existing users (and the README examples) depend on both forms.

## Documentation

Must update:

- `STATUS.md`
- `README.md` (one-line pointer)
- `CHANGELOG.md`
- (Optionally) a follow-up issue if a subcommand framework looks justified for v0.5+

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-035`, for example: `TP-035 add --help output and parser support`.

---

## Amendments

_Add amendments below this line only._
