# TP-016 — v0.2 dogfood: solo + 2–3 invited athletes, read-only validation

## Mission

Close the v0.2 gate by validating the read-path catalog against real prompts on real athlete accounts. Confirm that an LLM, given only icuvisor reads, produces training analysis without scale confusion, unit drift, hallucinated Strava data, or context-window blowouts.

Roadmap item (ROADMAP.md v0.2): **Dogfooded solo + 2–3 invited athletes, read-only.**

PRD anchors: KR1 (install success out of scope at v0.2 — manual config still), KR4 (reliability), KR5 (token efficiency — measured here against the soft 30k-token per-response ceiling), KR6 partial (Claude Desktop / Codex tested).

Complexity: Blast radius 1 (no code by default), Pattern novelty 1, Security 2 (real athlete data), Reversibility 1 = 5 → Review Level 1. Size: M.

## Dependencies

- **TP-009**, **TP-010**, **TP-011**, **TP-012**, **TP-013**, **TP-014**, **TP-015** — full v0.2 read catalog present and guarded

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C, §7.2.D, KR1–KR6
- `ROADMAP.md` v0.2 (every checkbox)
- `taskplane-tasks/TP-006-codex-local-mcp-validation/STATUS.md` — recipe for running Codex against the local binary
- `taskplane-tasks/TP-009`…`TP-015` STATUS.md files — known-good and known-fragile shapes per tool

## File Scope

Expected files:

- `docs/dogfood/v0.2-prompts.md` — the canonical prompt set used (read-only, redacted)
- `docs/dogfood/v0.2-findings.md` — per-tool pass/fail, scale-confusion observations, unit-drift observations, token-budget observations, latency observations
- `taskplane-tasks/TP-016-v02-dogfood-validation/STATUS.md`
- `CHANGELOG.md` only if user-visible behavior changes
- Code changes only if validation reveals a concrete bug — keep minimal and covered by tests

## Steps

### Step 1: Assemble the prompt set

- [ ] Build a canonical prompt set covering each PRD §6 customer job (1–5, except #3 which requires writes)
- [ ] Include at least one prompt per registered read tool, and one prompt per cluster (activities, fitness, wellness, events, workout library, custom items, periodization if shipped)
- [ ] Include three adversarial prompts: (a) ask the LLM to compute a metric that requires a unit it lacks (force unit-system reasoning); (b) ask about Strava-imported activities (force the `unavailable` shape to surface correctly); (c) ask "did I sleep well" against a row with both `sleepQuality` and `sleepScore` (force dual-scale reporting)
- [ ] Record the prompt set in `docs/dogfood/v0.2-prompts.md` with redactions

### Step 2: Solo dogfood against your own account

- [ ] Run the prompt set through Claude Desktop or Codex (per TP-006 recipe) against the maintainer's own intervals.icu account
- [ ] For each prompt, record: tool calls made, response sizes (bytes), whether the LLM's final answer was correct, any scale / unit / Strava-detection failure
- [ ] Measure the largest single response against the §7.2.D 30k-token soft ceiling; flag any tool that exceeds it

### Step 3: Invited athletes (2–3) read-only

- [ ] Recruit 2–3 invited athletes (one ideally an miles/imperial user; one ideally using a non-Garmin bridge like Polar or Oura to exercise wellness provenance)
- [ ] Provide them the manual-config recipe (v0.1 docs); have them run the same prompt set
- [ ] Collect findings via a redacted template; never receive raw athlete data
- [ ] Aggregate into `docs/dogfood/v0.2-findings.md`

### Step 4: Triage findings

- [ ] For each scale / unit / provenance / Strava-detection failure: open a issue tagged `v0.2-followup` linking the specific tool task
- [ ] For latency / token-budget regressions: confirm against KR4 / KR5 targets; if a tool exceeds the soft 30k-token ceiling, open a follow-up issue for pagination or shape tightening
- [ ] Decide which findings are launch-blocking for v0.5 vs follow-up; record the call in `STATUS.md`

### Step 5: Sign-off

- [ ] Update `ROADMAP.md` v0.2 to check off the dogfood item
- [ ] If any code/doc changed, run `make test`, `make build`, `make lint`, update `CHANGELOG.md`
- [ ] Confirm no athlete API keys, raw personal data, or training-load values are committed

## Reference Implementation Policy

- N/A — this is a validation task.

## Acceptance Criteria

- A documented, redacted prompt set covering every v0.2 read tool exists.
- Solo dogfood findings recorded with per-tool pass/fail.
- 2–3 invited-athlete dogfood findings recorded with per-tool pass/fail.
- Every failure has either a fix landed or a tagged follow-up issue.
- ROADMAP v0.2 dogfood checkbox is ticked.

## Do NOT

- Do not write or delete via icuvisor during v0.2 dogfood — read-only only.
- Do not commit raw athlete data, API keys, or athlete IDs to the findings doc; redact ruthlessly.
- Do not extend the dogfood scope to v0.3 write tools — those have their own validation cycle.
- Do not silently rewrite tools mid-dogfood; if a bug surfaces, fix it in a separate commit with a test.

## Documentation

Must update:

- `STATUS.md`
- `docs/dogfood/v0.2-prompts.md`
- `docs/dogfood/v0.2-findings.md`
- `ROADMAP.md` (tick the dogfood checkbox)

Check if affected:

- `CHANGELOG.md`
- `README.md` if any user-visible behavior changes from triaged fixes

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-016`, for example: `TP-016 record v0.2 solo dogfood findings`.

---

## Amendments

_Add amendments below this line only._
