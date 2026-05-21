# TP-061 — Introduce `ErrInvalidInput` sentinel and route validation errors through it (audit Medium)

## Mission

Several tool handlers use bare `fmt.Errorf("…")` for input-validation failures. Examples flagged by the 2026-05-16 audit:

- `internal/tools/get_activities.go:244`
- `internal/tools/delete_events_by_date_range.go:120`
- `internal/tools/apply_training_plan.go:160`
- `internal/tools/update_wellness.go:222, 232`

These are correct messages but unpairable with `errors.Is` — downstream callers and tests resort to string matching, which CLAUDE.md explicitly forbids ("Use `errors.Is` / `errors.As` at call sites; never `err.Error() == "..."`").

Fix: introduce a single `tools.ErrInvalidInput` (or `tools.UserError` already exists — confirm and reuse) sentinel. Validation errors wrap it: `fmt.Errorf("%w: end_date must be after start_date", tools.ErrInvalidInput)`. Then:

- Registry/handler error mapping can `errors.Is(err, tools.ErrInvalidInput)` to format a short LLM-facing error.
- Tests can assert validation paths without string matching.

Audit ref: 2026-05-16 Go audit, "Medium" severity.

PRD anchors: §7.4 reliability; "errors back to the LLM must be short, actionable, free of internal stack traces."
CLAUDE.md hard rules: "Sentinel errors for stable contract points; use `errors.Is`/`errors.As`."

Complexity: Blast radius 2 (touches ~5 tool files plus registry error mapper), Pattern novelty 1, Security 1, Reversibility 1 = 5 → Review Level 1. Size: S.

## Dependencies

- **TP-042** — soft. TP-042 collapses registry error mapping; once it lands, plug the sentinel into the new mapper. If this lands first, leave a TODO at the mapper for TP-042 to pick up.
- **TP-048** — `tools.UserError` may already serve this purpose. **Read TP-048 PROMPT before deciding** whether to add a new sentinel or extend `UserError`. Prefer extending the existing type.

## Context to Read First

- `CLAUDE.md` — Errors section.
- `internal/tools/registry.go:269-397` — existing `UserError` type and where errors flow back to the protocol.
- The five cited validation sites.
- Tests for any of the above — to understand current assertion style.

## File Scope

- New (or extended): `internal/tools/errors.go` — declare `ErrInvalidInput` (or extend `UserError` with an `IsInvalidInput()` accessor) and `func Wrap(err error) UserError` if needed.
- The five validation sites — switch to wrap the sentinel.
- `internal/tools/registry.go` (or wherever errors are mapped to LLM-facing strings) — branch on `errors.Is(err, ErrInvalidInput)`.
- Affected tool tests — replace any string-match assertions with `errors.Is`.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Internal infra/transport errors (don't introduce sentinels for those here).
- Wholesale sweep of every `fmt.Errorf` in the tools package — only validation errors flagged by the audit.
- Renaming `UserError`.

## Steps

### Step 1: Decide on the sentinel home
- [ ] Read TP-048's plan for `UserError`. Decide: new `var ErrInvalidInput = errors.New("invalid input")` vs an `IsInvalidInput()` method on `UserError`. Record the decision in `STATUS.md`.

### Step 2: Wire the sentinel into the five sites
- [ ] At each site, change `fmt.Errorf("end_date must be after start_date")` to `fmt.Errorf("%w: end_date must be after start_date", tools.ErrInvalidInput)` (or the equivalent `UserError` constructor).
- [ ] Update or add tests at each site to assert with `errors.Is(err, tools.ErrInvalidInput)`.

### Step 3: Route through the LLM-facing mapper
- [ ] At the registry/handler boundary, return a short LLM-facing string for `ErrInvalidInput`; do not include the wrapped error chain in the LLM-visible message.
- [ ] Log the full wrapped error via `slog.Warn` for the operator (without API keys / IDs).

### Step 4: Verify
- [ ] `grep -n 'fmt.Errorf("[^%]' internal/tools/` for the five sites — each must now use `%w`.
- [ ] `grep -rn 'err.Error() ==' internal/tools/` — zero hits.
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-061 introduce ErrInvalidInput sentinel for tool validation`.

## Acceptance Criteria

- A single sentinel (`tools.ErrInvalidInput` or equivalent `UserError` capability) covers validation errors.
- The five cited sites are migrated.
- LLM-facing message stays short; full chain is logged only.
- `errors.Is` is the only assertion style in affected tests.
- All `make` checks pass.

## Do NOT

- Do not introduce a sentinel per error message — one is enough.
- Do not change wire/JSON structure of error responses.
- Do not expand scope to non-tools packages.
- Do not let the wrapped error chain leak into the LLM-facing string.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed".

## Git Commit Convention

Conventional Commits, prefixed `TP-061`.

---

## Amendments

_Add amendments below this line only._
