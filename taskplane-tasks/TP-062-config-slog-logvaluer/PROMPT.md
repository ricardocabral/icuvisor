# TP-062 — `Config` implements `slog.LogValuer` (audit Low)

## Mission

`internal/config/config.go:340` defines `Config.String()` that produces a single redacted line. The 2026-05-16 audit notes this is hard to consume in structured logs — operators have to parse a string out of `slog` attributes.

Fix: implement `LogValue() slog.Value` returning a `slog.Group` with explicit attrs (`api_base_url`, `default_athlete_id`, `http_bind`, `coach_athletes_count`, `delete_mode`, `toolset`). Continue to redact `api_key`. Keep `String()` for human/error contexts (config diagnostics CLI), but make `slog` consumers automatically get structured attrs.

Audit ref: 2026-05-16 Go audit, "Low" severity.

PRD anchors: §7.4 reliability.
CLAUDE.md hard rules: structured logging with `log/slog`; never log API keys.

Complexity: Blast radius 1, Pattern novelty 1, Security 2 (redaction must hold), Reversibility 1 = 5 → Review Level 1. Size: XS.

## Dependencies

- **TP-066** — soft. TP-066 splits `config.go` into multiple files; if it lands first, place `LogValue` in the file owning `String()` (likely `redaction.go` or `config.go`).

## Context to Read First

- `internal/config/config.go:320-350` — current `String()` and redaction.
- Go stdlib `log/slog` docs: `LogValuer` interface, `slog.Group`, `slog.Value`.
- CLAUDE.md "Logging" section.

## File Scope

- `internal/config/config.go` (or wherever `String()` lives post-TP-066) — add `LogValue() slog.Value`.
- `internal/config/config_test.go` — table-driven test asserting (a) `LogValue` returns a Group, (b) `api_key` is never present, (c) all other public fields are present, (d) `slog.Default().Info("config", "cfg", cfg)` produces the expected attrs (use a test handler).
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Removing `String()`; keep it for non-`slog` contexts.
- Changing redaction policy.
- Restructuring `Config` exported fields.

## Steps

### Step 1: Implement `LogValue`
- [ ] Add `func (c Config) LogValue() slog.Value` returning a `slog.GroupValue(slog.Group("config", attrs...).Value)` — never include `api_key`.
- [ ] If any nested struct (e.g., `CoachAthletes`) is included, also implement `LogValue` for it or summarize as a count.

### Step 2: Test
- [ ] Use `slog.New(slog.NewJSONHandler(&buf, nil))` and assert the produced JSON has the expected keys and never `api_key`.
- [ ] Add a fuzzy / negative test: set `Config.APIKey = "secret-xxxxx"` and assert `strings.Contains(buf.String(), "secret")` is false.

### Step 3: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-062 config: implement slog.LogValuer with redaction`.

## Acceptance Criteria

- `Config` implements `slog.LogValuer`.
- `api_key` is never present in the structured output.
- Test asserts redaction holds.
- All `make` checks pass.

## Do NOT

- Do not include the API key (or any prefix/suffix of it) in attrs.
- Do not change the human-readable `String()` output.
- Do not log athlete IDs in a way that's hard to scrub (use a redacted form if needed).

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed".

## Git Commit Convention

Conventional Commits, prefixed `TP-062`.

---

## Amendments

_Add amendments below this line only._
