# TP-035-cli-help-documentation: CLI help documentation — Status

**Current Step:** Step 1: Help-text design
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** S

---

### Step 1: Help-text design

**Status:** ⏳ Not started

- [ ] Draft the `--help` output as a fixture before writing code
- [ ] Confirm env-var list against `internal/config/`, `internal/safety/`, `internal/response/`
- [ ] Document exit codes

### Step 2: Parser changes

**Status:** ⏳ Not started

- [ ] Recognize `--help`, `-h`, `help`
- [ ] Per-subcommand help (`version --help`)
- [ ] Unknown-flag errors include usage hint and exit code 2
- [ ] Preserve `--flag value` and `--flag=value` parsing
- [ ] All I/O routed through `opts.Stdout` / `opts.Stderr`

### Step 3: Tests

**Status:** ⏳ Not started

- [ ] Table-driven coverage of help flags / subcommand / errors
- [ ] Golden-file fixture for `--help` output
- [ ] Env-var names cross-checked against resolved config

### Step 4: Documentation

**Status:** ⏳ Not started

- [ ] README pointer to `--help`
- [ ] CHANGELOG `[Unreleased]` entry

### Step 5: Verify

**Status:** ⏳ Not started

- [ ] `make build`, `make test`, `make test-race`, `make lint`
- [ ] Manual smoke of each help/error path

---

## Decisions

- **Library choice:** TBD in Step 1. Default recommendation in PROMPT.md is stdlib `flag` (CLAUDE.md "stdlib first"); escalate to cobra/urfave only if the help template becomes unreadable.

## Notes

_Add notes as work progresses._
