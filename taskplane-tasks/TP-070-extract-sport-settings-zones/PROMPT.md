# TP-070 — Extract zones-merge logic from `internal/tools/update_sport_settings.go` (audit God-module)

## Mission

`internal/tools/update_sport_settings.go` (526 LOC) is a single tool — acceptable — but mixes:

- arg decode and per-field validation (FTP, HR, pace)
- **the zones merge** (gated by `ICUVISOR_DELETE_MODE`, the destructive-edit-sensitive path)
- unit conversion
- result shaping

The zones-merge logic is the complex, dangerous part — it's the only path here that can wipe an athlete's custom zones. It should live in its own file with focused tests.

Goal: extract `internal/tools/update_sport_settings_zones.go` (and `_test.go`) containing the zones-merge logic and its delete-mode gate. The main tool file shrinks and the dangerous code becomes greppable.

Audit ref: 2026-05-16 Go audit, God-module section.

PRD anchors: §7.2.C `update_sport_settings`; §7.2.D destructive-edit gating; safety policy in `internal/safety`.
CLAUDE.md hard rules: destructive ops registration-time gated; never accept model-controlled confirm flags.

Complexity: Blast radius 2 (tools package internal), Pattern novelty 1, Security 3 (zones overwrite is destructive), Reversibility 1 = 7 → Review Level 2. Size: S.

## Dependencies

- **TP-018** (delete-mode safety gate) — context only; behaviour must not regress.
- **TP-022** (`update_sport_settings` originally) — context only.

## Context to Read First

- `internal/tools/update_sport_settings.go` — full file.
- `internal/tools/update_sport_settings_test.go` — locked-in zones-merge cases.
- `internal/safety/` — destructive-op gate currently in use.
- PRD §7.2.D / §7.2.C `update_sport_settings` row.

## File Scope

- New: `internal/tools/update_sport_settings_zones.go` — the zones merge logic + delete-mode check.
- New: `internal/tools/update_sport_settings_zones_test.go` — table-driven coverage for: no zones change (no gate needed), zones change in safe mode (rejected), zones change in destructive mode (allowed).
- Trim: `internal/tools/update_sport_settings.go` — keep arg decode, FTP/HR/pace handling, result shaping; delegate the merge.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Changing the safety gate semantics or the env var name.
- Adding model-controlled `confirm` parameters (forbidden by CLAUDE.md).
- Changing zone structures or wire format.
- Other parts of the tool (FTP, HR, pace).

## Steps

### Step 1: Lock in safety-gate regression tests
- [ ] Verify the existing test file exercises: zones-edit blocked in safe mode; zones-edit allowed in destructive mode. If not, add the missing cases first.

### Step 2: Extract
- [ ] Move the zones-merge + delete-mode-check helpers to `update_sport_settings_zones.go`.
- [ ] Update the main handler to call the helper.
- [ ] Mirror the test split.

### Step 3: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Adversarial safety tests (TP-028's set) all pass.
- [ ] `scripts/snapshot_tool_schemas.go` diff empty.

## Acceptance Criteria

- Zones-merge logic lives in its own file with its own focused test.
- Main `update_sport_settings.go` shrinks.
- Safety-gate behaviour unchanged (regression tests prove it).
- No model-controlled confirm flag introduced.
- All `make` checks pass.

## Do NOT

- Do not change the destructive-op gate semantics or env var.
- Do not introduce model-controlled overrides.
- Do not change zone structures or wire format.
- Do not extract FTP/HR/pace into separate files in this task — out of scope.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).

## Git Commit Convention

Conventional Commits, prefixed `TP-070`.

---

## Amendments

_Add amendments below this line only._
