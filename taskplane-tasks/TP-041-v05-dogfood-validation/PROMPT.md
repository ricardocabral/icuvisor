# TP-041 — v0.5 dogfood prep (scripts, docs, checklists for the internal beta)

## Mission

Produce the *artifacts* needed to run the v0.5 internal beta later — recruitment script, onboarding playbook, measurement template, diagnostics subcommand, exit-interview script — without performing the recruitment or running the cohort. Cohort execution, recruitment, and synthesis are async / calendar-bound and explicitly out of scope for batch completion; the maintainer drives those manually after this task closes.

This task is done when the next person (or the maintainer next month) can sit down, follow the checklist, and run the beta end-to-end without needing to invent any document, script, or measurement template.

PRD anchors: §4 KR1; §5 Primary + Secondary segments; §7.4 #6, #8, #12 (assumptions the beta will eventually validate); §7.2.I "Copy diagnostics" (CLI form). ROADMAP positioning: closes the v0.5 milestone *prep*; the v0.5 ship gate is the maintainer-run beta itself, tracked outside taskplane.

Complexity: Blast radius 1 (docs + one small subcommand), Pattern novelty 1 (copy from TP-016 / TP-029 protocols), Security 2 (diagnostics output must not leak secrets), Reversibility 1 = 5 → Review Level 2. Size: M.

## Dependencies

- **TP-035** — CLI parser / `--help` golden file (the new `diagnostics` subcommand registers here).
- **TP-036…TP-040** — exist as referenced features in the docs, but the docs can be drafted in parallel; final cross-references resolved at the end.
- **TP-016**, **TP-029** — prior dogfood protocols to mine for reusable structure.

## File Scope

- `internal/app/diagnostics.go` + `_test.go` — new `icuvisor diagnostics` subcommand. Prints: server version, catalog hash (TP-040), config source (env/keychain/file, never the value), resolved transport, `ICUVISOR_DELETE_MODE`, `ICUVISOR_TOOLSET`, `ICUVISOR_COACH_MODE`, OS + Go runtime, last N tool-call names with timestamps (no arguments, no payloads). Test asserts zero secrets in output across all paths.
- `internal/app/app.go` — wire the subcommand; update TP-035 `--help` golden file.
- `docs/internal-beta/README.md` — index that links the four docs below and states the running order.
- `docs/internal-beta/protocol.md` — recruitment script, consent statement, eligibility filters (OS, AI client, coach yes/no, mobile-need signal), cohort cap (5–10), 14-day recruitment time-box.
- `docs/internal-beta/onboarding-playbook.md` — operator's exact terminal recipe: DMG download → install → `icuvisor setup` → Claude Desktop config snippet → first-call verification. Includes coach-mode variant (TP-039 roster + ACL walkthrough) and a troubleshooting section pointing to `icuvisor diagnostics`.
- `docs/internal-beta/measurement.md` — template `findings.md` table with one row per athlete, columns for install-to-first-call time (KR1), top-5 tool calls, mobile-need answer (PRD §7.4 #8), free-text surprises, blockers filed. Top of file: KR1 / §7.4 #12 measurement procedure.
- `docs/internal-beta/exit-interview.md` — 8–12 question script for end-of-beta interviews: coach-mode usability, schema-change-notification clarity, willingness to recommend, blockers for daily use.
- `docs/internal-beta/findings.md` — *empty template only* (header + the measurement table skeleton from `measurement.md`); the maintainer fills it during the run.
- `docs/internal-beta/checklist.md` — single-page operator checklist (Recruit → Onboard → Run → Synthesize), each step a one-line action with a doc link.
- `taskplane-tasks/TP-041-v05-dogfood-validation/STATUS.md`.

Out of scope: posting the forum recruitment notice, recruiting athletes, running the cohort, populating `findings.md`, updating PRD §7.4 with measured numbers, opening v1.0 blocker issues, opt-in telemetry, tray icon, public website, non-mac installers.

## Steps

### Step 1: Diagnostics subcommand

- [ ] `icuvisor diagnostics` prints the fields listed in File Scope; routed through `opts.Stdout`.
- [ ] Test: across every code path, output contains no value matching the API key, the athlete ID raw form, or any token-shaped string. Use a fixture-injected secret and grep the captured stdout.
- [ ] Update TP-035 `--help` golden file in the same change.

### Step 2: Docs

- [ ] Draft each `docs/internal-beta/*.md` file. Keep each under ~150 lines; prefer checklists and tables over prose.
- [ ] Consent statement in `protocol.md` is explicit about what the maintainer receives (tool-call descriptions, not data values) and how to revoke.
- [ ] `onboarding-playbook.md` includes the verbatim Claude Desktop / Code JSON snippets (cross-reference TP-037 client docs; do not duplicate — link).
- [ ] `measurement.md` references KR1, §7.4 #6/#8/#12 by name so the maintainer can later mark them validated.

### Step 3: Cross-check + verify

- [ ] `docs/internal-beta/README.md` links all five docs in execution order.
- [ ] Maintainer reads `checklist.md` cold; can identify what to do next at each phase without consulting other docs first.
- [ ] `make test`, `make lint`, `make build` clean.

## Acceptance Criteria

- `icuvisor diagnostics` subcommand ships; test proves no secret leakage.
- All seven `docs/internal-beta/` files exist, link to each other, and reference the relevant PRD anchors and TP IDs.
- `findings.md` is an empty template (no fabricated data).
- TP-035 `--help` golden file is updated to include `diagnostics`.
- Nothing in this task requires waiting on external people, a release tag, or calendar time. A reviewer can verify acceptance in a single pass.

## Do NOT

- Do not recruit athletes, post forum notices, or contact anyone as part of this task.
- Do not populate `findings.md` with example data — empty template only, so a future reader is not confused about what is real.
- Do not ship opt-in telemetry, tray icon, or auto-update.
- Do not have `diagnostics` print the API key value, the raw config-file contents, or recent tool-call arguments / payloads. Names and timestamps only.
- Do not duplicate the Claude Desktop / Code JSON snippets across TP-037 docs and the onboarding playbook — link.
- Do not retag a release as part of this task (the mid-beta `beta.2` exercise from the previous draft is a maintainer-run step, documented in `onboarding-playbook.md` as instructions, not executed here).

## Documentation

Must create / update:

- `internal/app/testdata/help-fixture` (TP-035 golden file)
- `docs/internal-beta/README.md`
- `docs/internal-beta/protocol.md`
- `docs/internal-beta/onboarding-playbook.md`
- `docs/internal-beta/measurement.md`
- `docs/internal-beta/exit-interview.md`
- `docs/internal-beta/findings.md` (empty template)
- `docs/internal-beta/checklist.md`
- `CHANGELOG.md` (`diagnostics` subcommand under Added)
- `STATUS.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-041`, for example: `TP-041 add diagnostics subcommand`.

---

## Amendments

_Add amendments below this line only._
