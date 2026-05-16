# TP-075 — `add_or_update_event` NOTE category create refused

## Mission

Resolves [GitHub issue #6](https://github.com/ricardocabral/icuvisor/issues/6) — *TP-029: `add_or_update_event` NOTE create refused during dogfood*.

During the v0.3 dogfood (TP-029, W-01 in [`docs/dogfood/v0.3-findings.md:19`](../../docs/dogfood/v0.3-findings.md)), creating a synthetic NOTE event named `TP-029 dogfood note` against the dedicated test athlete failed: the upstream POST was refused, and a same-day `get_events` re-read showed zero synthetic events. The structured WORKOUT create on W-03 in the same run passed, so the failure is specific to the NOTE category code path — not the tool wiring overall.

Suspected defects (from triage, ordered by likelihood — confirm via live probe):

1. **Date format gating** — [`internal/tools/add_or_update_event.go:189-193`](../../internal/tools/add_or_update_event.go) appends a `T00:00:00` suffix to `start_date_local` *only* for WORKOUT category. NOTEs may require the same datetime format, or a different one.
2. **Required `type` field** — for NOTE categories, intervals.icu may require a non-empty `type` field; the current payload omits it (`type omitempty`).
3. **`name` requirement** — NOTEs may have different `name` / `description` semantics than WORKOUTs.
4. **Less likely:** category enum casing (`NOTE` vs `Note` vs `note`) — verify against a working NOTE created via the intervals.icu web UI.

This task requires **live API probing** against the `.env-dev` test athlete to isolate which of the above is the actual upstream contract.

PRD anchors: §7.2.C event write field set; §7.4 schema-honesty rule.
CLAUDE.md: hard rule 5 (terse-by-default — schemas matter); MCP-server conventions ("Schemas matter"); "Idempotency: writes that can be safely retried should be idempotent."

Complexity: Blast radius 1 (one tool / one category branch), Pattern novelty 2 (new branch in date/required-field logic + live-probe loop), Security 1, Reversibility 2 (writes against real account but on dedicated test athlete) = 6 → Review Level 2. Size: M.

## Dependencies

- None. Independent of the other v0.2/v0.3 dogfood follow-ups (TP-072, TP-073, TP-074, TP-076, TP-077).
- Useful prior art: TP-020 (event write cluster) shipped the WORKOUT branch — read its STATUS/PROMPT for the original payload design.

## Context to Read First

- [`docs/prd/PRD-icuvisor.md:256-`](../../docs/prd/PRD-icuvisor.md) — "Events & workouts" section. Note the `description` verbatim preservation rule.
- [`docs/dogfood/v0.3-findings.md`](../../docs/dogfood/v0.3-findings.md) line 19 (W-01) and the per-tool triage row for `add_or_update_event`.
- [`internal/tools/add_or_update_event.go`](../../internal/tools/add_or_update_event.go):
  - Request struct: lines 26–40.
  - Handler + date logic: lines 55–193.
- [`internal/intervals/events.go`](../../internal/intervals/events.go):
  - `AddOrUpdateEvent`: lines 131–151.
  - `writeEventPayload`: lines 153–164.
- [`internal/tools/add_or_update_event_test.go`](../../internal/tools/add_or_update_event_test.go) — existing WORKOUT round-trip test (lines 28–71). Mirror its pattern for the NOTE case.
- Existing TP work: [`taskplane-tasks/TP-020-event-write-cluster/`](../TP-020-event-write-cluster/) for original design rationale.

## File Scope

- `internal/tools/add_or_update_event.go` — adjust the NOTE code path (date formatting + any missing required fields).
- `internal/intervals/events.go` — extend `writeEventPayload` if a new required field needs to be carried.
- `internal/tools/add_or_update_event_test.go` — add a NOTE round-trip test mirroring the WORKOUT one.
- `internal/intervals/testdata/events/` — new fixture for NOTE create request/response.
- `CHANGELOG.md` — `[Unreleased]` under "Fixed".
- `STATUS.md` (this dir).

Out of scope:
- Refactoring the WORKOUT path.
- Adding new event categories beyond NOTE (any further categories: separate task).
- Touching delete or update flows beyond what's needed for the round-trip test cleanup.

## Steps

### Step 1: Live probe to isolate the upstream contract

- [ ] Source `.env-dev` to get the test athlete credentials.
- [ ] **Out-of-process probe** (use `curl` or a tiny Go scratch program — do NOT commit it): POST a minimal NOTE event directly to the intervals.icu events endpoint. Vary one field at a time:
  - Minimal payload: `{ "start_date_local": "2026-05-25T00:00:00", "category": "NOTE", "name": "tp-075 probe a" }`.
  - Without `T00:00:00`: `start_date_local: "2026-05-25"`.
  - With `"type": ""` vs `"type": "Note"` vs omitting `type`.
  - With/without `description`.
- [ ] Record which combination is accepted. Save the request/response under `internal/intervals/testdata/events/note_create_request.json` + `note_create_response.json` (sanitize: redact athlete IDs, replace event IDs with `EVENT_ID_PLACEHOLDER`).
- [ ] **Clean up:** delete every probe event you created on the test athlete (`gh` is fine for the issue; use the intervals.icu UI or the existing `delete_event` tool to remove the probe data). Verify with `get_events` that no probe events remain.

### Step 2: Add a failing test

- [ ] Mirror `TestAddOrUpdateEventCreatePreservesFreeTextTagsAndReadShape` (line 28) for the NOTE case, using the captured fixture as the expected outbound body.
- [ ] Confirm the test fails on `main`.

### Step 3: Fix the tool

- [ ] Apply the minimum diff identified in Step 1. Likely candidates:
  - Generalize the date suffix logic so all categories use ISO datetime, not just WORKOUT.
  - Populate `type` with the upstream-required default for NOTEs if Step 1 confirmed that requirement.
- [ ] Keep the WORKOUT path's existing semantics unchanged — assert by running the existing WORKOUT test.

### Step 4: Build + lint + race + live re-validation

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] **Live re-validation** against `.env-dev`: call `add_or_update_event` via stdio MCP with a NOTE payload; confirm a new event appears in `get_events` for the chosen date; delete it; confirm it's gone.

### Step 5: Document amendment

- [ ] If Step 1 surfaced anything non-obvious about the upstream contract (e.g. `type` is required for NOTE but not WORKOUT, or vice versa), capture it in `docs/upstream-gaps/event-note-payload.md` (new file, 1–2 paragraphs). This is so future agents don't re-do the probe.

### Step 6: Close the GitHub issue

- [ ] Update `CHANGELOG.md` under `[Unreleased] → Fixed`: `add_or_update_event NOTE category create now succeeds against intervals.icu (was refused upstream due to <root cause>).`
- [ ] Update `STATUS.md`.
- [ ] Commit: `fix(events): repair add_or_update_event NOTE category create (TP-075, closes #6)`.
- [ ] Reference `Closes #6` in the PR body. After merge, verify auto-close; otherwise `gh issue close 6 --comment "Fixed in <commit-sha> / <PR>"`.

## Acceptance Criteria

- A NOTE create via `add_or_update_event` against the test athlete is accepted upstream, and the resulting event appears in `get_events` for the chosen date.
- The existing WORKOUT create path is unchanged (its test still passes).
- A new unit test exercises the NOTE create round-trip using a fixture captured from the live probe.
- `make build`, `make test`, `make test-race`, `make lint` pass.
- Any probe-created events on the test athlete have been deleted; the test athlete is in the same state as before.
- If a non-obvious upstream contract was discovered, `docs/upstream-gaps/event-note-payload.md` documents it.
- GitHub issue #6 closed.

## Do NOT

- Do not commit your probe scratch files.
- Do not leave probe-created events on the test athlete.
- Do not paste raw request/response bodies into git without redacting athlete IDs / event IDs / dates that could identify the test account.
- Do not introduce a `confirm: true` LLM-controlled override on this tool (CLAUDE.md hard rule 5).
- Do not refactor the WORKOUT code path "while you're here" — separate concern.

## Documentation

- `CHANGELOG.md` `[Unreleased]` under "Fixed".
- `STATUS.md` in this dir.
- *Optional but encouraged:* `docs/upstream-gaps/event-note-payload.md`.

## Git Commit Convention

Conventional Commits, prefixed with TP-075. Example:

```
fix(events): repair add_or_update_event NOTE category create

TP-075. Closes #6.
```

---

## Amendments

_Add amendments below this line only._
