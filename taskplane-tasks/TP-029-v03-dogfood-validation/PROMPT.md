# TP-029 — v0.3 dogfood: dedicated test-athlete account, write-path validation

## Mission

Close the v0.3 gate by validating the write-path catalog against a dedicated test athlete account. Confirm that writes round-trip, deletes are gated as designed, the workout_doc serializer preserves fidelity end-to-end, and the adversarial safety claims hold under a real LLM client.

Roadmap item (ROADMAP.md v0.3): **Dogfooded against a dedicated test athlete account; no production athletes yet.**

PRD anchors: KR4 (reliability), KR5 (token efficiency), KR6 partial (Claude Desktop / Codex tested for writes).

Complexity: Blast radius 2 (test athlete only), Pattern novelty 1, Security 3 (mutates a real account), Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-018**…**TP-028** — every v0.3 task

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C catalog, §7.2.D safety
- `ROADMAP.md` v0.3 (every checkbox)
- `taskplane-tasks/TP-006-codex-local-mcp-validation/STATUS.md`
- `taskplane-tasks/TP-016-v02-dogfood-validation/` — recipe + redaction conventions
- `taskplane-tasks/TP-018`…`TP-028/STATUS.md`

## File Scope

Expected files:

- `docs/dogfood/v0.3-prompts.md` — canonical write-path prompt set (redacted)
- `docs/dogfood/v0.3-findings.md` — per-tool round-trip + safety findings
- `taskplane-tasks/TP-029-v03-dogfood-validation/STATUS.md`
- `CHANGELOG.md` only if user-visible behavior changes
- Code changes only when validation reveals a concrete bug; keep minimal and covered by tests

## Steps

### Step 1: Load the test-athlete credentials

- [ ] Source `.env-dev` (untracked, gitignored) at the repo root for the test-athlete credentials: `INTERVALS_API_KEY` and `INTERVALS_ATHLETE_ID` (and any sibling vars the dev profile uses). The maintainer maintains this file out-of-band; do not create or commit it.
- [ ] Confirm `.env-dev` is gitignored and never staged; verify with `git check-ignore .env-dev`.
- [ ] Export the vars into the binary's environment for the dogfood run (e.g. `set -a; source .env-dev; set +a`); do **not** echo the key, do **not** pass it as a tool argument, and do **not** copy it into `STATUS.md` or any committed doc — redact to `INTERVALS_API_KEY=<from .env-dev>` in notes.
- [ ] Record only the *redacted* athlete-ID prefix (e.g. `i12345…`) in `STATUS.md` along with confirmation that the credentials came from `.env-dev`, not from a production athlete.
- [ ] Confirm no production athlete is in scope by cross-checking the athlete ID against the maintainer's known production accounts before any write or delete call.

### Step 2: Assemble the write prompt set

- [ ] One prompt per write tool: `add_or_update_event` (create + update), `link_activity_to_event`, `add_activity_message`, `update_wellness` (subjective + measurement), `update_sport_settings` (threshold-only and zones), `create_workout` / `update_workout`, `create_custom_item` / `update_custom_item`, `apply_training_plan` (dry-run + apply)
- [ ] One prompt per gated delete tool, run twice: once in `safe` (expect tool-not-found surrender) and once in `full` (expect success)
- [ ] Three adversarial prompts pulled from TP-028 corpus
- [ ] Record set in `docs/dogfood/v0.3-prompts.md` with redactions

### Step 3: Run in `safe` mode

- [ ] Run the write subset; confirm round-trips by re-reading after each write
- [ ] Run the destructive subset; expect every one to fail tool-not-found
- [ ] Run the adversarial subset; expect every one to surrender

### Step 4: Run in `full` mode

- [ ] Re-run the destructive subset against the test athlete
- [ ] Confirm `_meta.deleted` echoes and `delete_events_by_date_range` enforces its range cap
- [ ] Restore the test athlete to a known state or note its disposable status

### Step 5: Aggregate findings + triage

- [ ] `docs/dogfood/v0.3-findings.md`: per-tool pass/fail with round-trip evidence, byte/token sizes, latency, observed `_meta` fields
- [ ] Open `v0.3-followup` project issues for any failure
- [ ] Decide which findings are launch-blocking for v0.4 entry vs follow-up

### Step 6: Sign-off

- [ ] Tick the v0.3 dogfood item in `ROADMAP.md`
- [ ] `make test`, `make build`, `make lint` if any code/doc changed
- [ ] Confirm no test-athlete API key or raw account data is committed

## Acceptance Criteria

- A documented, redacted prompt set covers every v0.3 write tool, every gated delete, and three adversarial prompts.
- `safe`-mode runs prove the destructive subset is unreachable.
- `full`-mode runs prove the destructive subset works and surfaces `_meta.deleted`.
- Every failure has either a fix landed or a tagged follow-up issue.
- ROADMAP v0.3 dogfood checkbox is ticked.

## Do NOT

- Do not run any v0.3 write or delete tool against a production athlete; the `.env-dev` test-athlete credentials are mandatory.
- Do not commit, log, or echo the contents of `.env-dev`. Never copy `INTERVALS_API_KEY` into `STATUS.md`, the findings doc, transcripts, or commit messages. The athlete ID is redacted to a prefix in any committed note.
- Do not stage `.env-dev`; if `git status` shows it, stop and fix `.gitignore` before continuing.
- Do not silently rewrite tools mid-dogfood; if a bug surfaces, fix it in a separate commit with a test.
- Do not extend dogfood scope to v0.4 (token-efficiency benchmarks have their own task).

## Documentation

Must update:

- `STATUS.md`
- `docs/dogfood/v0.3-prompts.md`
- `docs/dogfood/v0.3-findings.md`
- `ROADMAP.md` (tick the dogfood checkbox)

Check if affected:

- `CHANGELOG.md`
- `README.md` if any user-visible behavior changes from triaged fixes

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-029`, for example: `TP-029 record v0.3 safe-mode dogfood findings`.

---

## Amendments

_Add amendments below this line only._
