# TP-068 — Split `internal/app/setup.go` and extract reusable `terminalPrompter` (audit God-module)

## Mission

`internal/app/setup.go` is 562 LOC and packs three concerns:

- CLI arg parsing + help (`parseSetupArgs`, `resolveSetupConfigPath`, `runSetupCommand`).
- The setup flow itself — `RunSetup` is ~113 LOC, borderline cyclomatic-heavy.
- A generic `terminalPrompter` (~lines 492-562) that would serve other interactive commands.

Goal:

1. `internal/app/setup_cmd.go` — arg parsing + dispatch.
2. `internal/app/setup_flow.go` — `RunSetup` and its step helpers.
3. `internal/cli/prompt/` (new package) — lift `terminalPrompter` so future interactive commands can reuse it. Narrow interface (e.g., `Prompter.AskString`, `Prompter.AskYesNo`).

This is the first step toward a `cli/` shared package; resist over-designing — only the methods `setup.go` already uses get exported.

Audit ref: 2026-05-16 Go audit, God-module section.

PRD anchors: §7.1 first-run onboarding.
CLAUDE.md hard rules: small interfaces; `internal/` default home for non-exported helpers.

Complexity: Blast radius 2 (app package internal; new `internal/cli/prompt` package), Pattern novelty 2 (new package), Security 1, Reversibility 2 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-038** (first-run onboarding) — context only; behaviour must not regress.
- **TP-049** Step 3 — sequence after (env-read move). Re-baseline against the post-TP-049 file.

## Context to Read First

- `internal/app/setup.go` — full file.
- `internal/app/setup_test.go` — locked-in flow tests.
- `internal/app/app.go` — to find any other consumer of the prompter (probably none yet).
- TP-038 PROMPT.md + STATUS.md — flow semantics to preserve.

## File Scope

- `internal/app/setup_cmd.go` — `parseSetupArgs`, `resolveSetupConfigPath`, `runSetupCommand`.
- `internal/app/setup_flow.go` — `RunSetup` plus step helpers (slice `RunSetup` into smaller helpers like `setupCollectAPIKey`, `setupVerifyConnection`, etc.).
- `internal/cli/prompt/prompt.go` — `Prompter` interface + `Terminal` impl.
- `internal/cli/prompt/prompt_test.go` — table-driven.
- Tests in `internal/app/` — split alongside.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Changing the setup flow's user-visible steps or prompts.
- Adding non-`setup` interactive commands.
- Generalizing the prompter beyond the methods setup uses today.
- Adding new dependencies (e.g., bubbletea / survey).

## Steps

### Step 1: Capture flow regression-gate
- [ ] Ensure `setup_test.go` covers happy path + at least one error path (invalid API key, keychain error). If not, add tests.

### Step 2: Slice `RunSetup`
- [ ] Break the 113-LOC function into 4-6 step helpers in `setup_flow.go`. No logic change.
- [ ] Run tests after the slice.

### Step 3: Extract `cli/prompt`
- [ ] Create `internal/cli/prompt/`. Move `terminalPrompter` + the minimal interface.
- [ ] Update `internal/app/setup_flow.go` to import the new package.
- [ ] Run tests.

### Step 4: Split arg parsing
- [ ] Move `parseSetupArgs`, `resolveSetupConfigPath`, `runSetupCommand` to `setup_cmd.go`.
- [ ] Run tests.

### Step 5: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Manual smoke: `bin/icuvisor setup --help` works; `bin/icuvisor setup` golden path works in a clean keychain.
- [ ] `wc -l internal/app/setup*.go` — each ≤ ~300 LOC; `internal/cli/prompt/*.go` ≤ ~200 LOC.

## Acceptance Criteria

- `setup_cmd.go`, `setup_flow.go`, and `internal/cli/prompt/` exist as described.
- `RunSetup` is sliced into step helpers; no helper > 60 LOC.
- `internal/cli/prompt.Prompter` interface is minimal (only methods setup uses).
- Setup flow output and prompts are byte-identical (smoke-test signoff in `STATUS.md`).
- All `make` checks pass.

## Do NOT

- Do not change prompt wording, order, or defaults.
- Do not add features (resume / non-interactive mode / etc.) in this task.
- Do not introduce a third-party prompter library.
- Do not export from `internal/cli/prompt` more than setup needs today.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).

## Git Commit Convention

Conventional Commits, prefixed `TP-068`.

---

## Amendments

_Add amendments below this line only._
