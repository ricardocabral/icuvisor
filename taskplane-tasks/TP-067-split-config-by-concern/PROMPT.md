# TP-067 — Split `internal/config/config.go` by concern (audit God-module)

## Mission

`internal/config/config.go` is 690 LOC and bundles seven responsibilities:

- `Load` (file + env composition)
- `validate` (~88 LOC single function with per-field branching)
- atomic file write (`writeConfigFile`)
- athlete-ID parse / normalize
- HTTP bind validation/normalization (5 helpers, ~50 LOC)
- `.env` parsing
- default-path resolution
- redaction / `String()`
- raw-config merge

Goal: split by concern into focused files inside the same package. Same exported names, same signatures.

Suggested layout:

- `config.go` — types, defaults, public `Load` entry point.
- `load.go` — composition of file + env + flag inputs.
- `validate.go` — sliced per-field (FTP, HR, bind, etc.).
- `write.go` — atomic-write helpers.
- `athlete.go` — athlete-ID parsing/normalization.
- `httpbind.go` — HTTP bind validation and normalization.
- `dotenv.go` — `.env` parsing.
- `redaction.go` — `String()` (and `LogValue()` once TP-062 lands).

Audit ref: 2026-05-16 Go audit, God-module section.

PRD anchors: §7.1 setup / config; §7.3 transport (HTTP bind); §7.4 reliability.
CLAUDE.md hard rules: small files; keychain-only secrets; never log API keys.

Complexity: Blast radius 2 (config package internal; callers unchanged), Pattern novelty 1, Security 2 (validation correctness must not regress), Reversibility 1 = 6 → Review Level 1. Size: S.

## Dependencies

- **TP-049** Step 3 — `DebugMetadata` env read moves into `config.Load`. Sequence after TP-049 so the new env read lands at the right home.
- **TP-062** — soft. If TP-062 has added `LogValue`, place it in `redaction.go`. If not yet landed, leave a TODO marker for TP-062 to slot into the new file.

## Context to Read First

- `internal/config/config.go` — full file.
- `internal/config/config_test.go` — golden coverage.
- TP-049 PROMPT.md — to know whether the env-read move has happened.
- Any caller of `internal/config` (very small surface).

## File Scope

- Inside `internal/config/`: split per the layout above.
- Tests: split alongside the source split where it improves clarity (e.g., `validate_test.go`, `httpbind_test.go`).
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Changing validation rules.
- Changing config file format / paths.
- Changing exported names or signatures.
- Adding new config fields.

## Steps

### Step 1: Inventory
- [ ] List every top-level decl in `config.go` and assign each to a target file. Record in `STATUS.md`.
- [ ] Slice `validate` into per-field sub-validators (e.g., `validateFTP`, `validateHTTPBind`) — same total logic.

### Step 2: Mechanical split
- [ ] One commit per file extraction. Run tests after each.
- [ ] Slice the `validate` body into per-field functions, called in the same order from a single `validate` entry point. No logic change.

### Step 3: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Round-trip test: load a fixture config, write it, reload it — output must be byte-identical to pre-refactor.
- [ ] `wc -l internal/config/*.go` — each file ≤ ~250 LOC.

## Acceptance Criteria

- `internal/config/config.go` holds only types + defaults + public entry point.
- Each concern lives in its own file.
- `validate` is sliced into per-field functions.
- Exported names and signatures unchanged.
- All `make` checks pass.
- Round-trip test gates byte-identical config IO.

## Do NOT

- Do not change config file format, paths, or env-var names.
- Do not change validation rules. Same rules, smaller functions.
- Do not collapse into a giant single commit — reviewers need per-file diffs.
- Do not add new config fields.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).

## Git Commit Convention

Conventional Commits, prefixed `TP-067`. One commit per extracted file.

---

## Amendments

_Add amendments below this line only._
