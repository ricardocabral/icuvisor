# TP-084 — Upstream-signal regression pack from 2026-05 behavior review

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Mostly test/fixture hardening, but touches user-facing edge-case semantics and may require small fixes.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 1, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-084-upstream-signal-regression-pack/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Lock the known upstream-behavior edge cases into regression fixtures: Strava numeric/no-`i` empty stubs from sync chains return structured unavailable markers, event detail 404-after-list remains `upstream_inconsistency`, and NOTE creates keep accepting date-only tool input while sending upstream-required local datetime payloads.

## Dependencies

- **Task:** TP-009 (Strava-blocked detection exists)
- **Task:** TP-012 (`get_event_by_id` inconsistency handling exists)
- **Task:** TP-075 (NOTE create date-only fix exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- docs/dogfood/v0.3-findings.md — NOTE create regression context.
- docs/upstream-gaps/event-note-payload.md — NOTE payload contract.
- internal/tools/get_activities_strava_test.go, get_event_by_id_test.go, add_or_update_event_test.go — existing edge-case tests.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/tools/get_activities_strava_test.go`
- `internal/tools/get_activity_details_test.go`
- `internal/tools/get_event_by_id_test.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/intervals/testdata/**/*`
- `docs/upstream-gaps/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Collect fixtures and expected markers

- [ ] Identify existing fixtures for each upstream signal and add sanitized fixtures where coverage is missing.
- [ ] Define exact expected structured markers for Strava stubs and event inconsistency.
- [ ] Confirm NOTE expected outbound payload uses local datetime for date-only input.

### Step 2: Add regression tests

- [ ] Add tests for numeric/no-`i` Strava empty stubs from Wahoo/MyWhoosh/TrainerRoad chains.
- [ ] Add/strengthen `get_event_by_id` 404-after-list fixture coverage.
- [ ] Add/strengthen NOTE date-only create serialization test.

### Step 3: Fix any regressions exposed by the pack

- [ ] Apply minimal code fixes if new tests fail.
- [ ] Keep changes additive and schema-stable.
- [ ] Run targeted affected tool tests.

### Step 4: Verify and document

- [ ] Run full suite/build/lint.
- [ ] Update CHANGELOG.md and upstream-gap notes only if new behavior changed.
- [ ] Record fixture provenance/redaction notes in STATUS.md.

### Step 5: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- CHANGELOG.md — record user-visible behavior under [Unreleased] if code or docs behavior changes.
- STATUS.md — keep execution state current.

**Check If Affected:**
- README.md — update if public setup/tool behavior changes.
- web/content/reference/tools.md — update if tool catalog descriptions or generated docs are affected.
- docs/prd/PRD-icuvisor.md — check only if behavior intentionally diverges from product scope.

## Completion Criteria

- All three upstream signals have regression tests.
- Structured unavailable/inconsistency markers are exact and stable.
- NOTE date-only input continues to serialize to upstream-required local datetime.
- No live secrets or athlete-identifying data are committed.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-084` for traceability. Examples:

- `feat(TP-084): complete step 1 — scope current behavior`
- `fix(TP-084): repair regression found during analyzer tests`
- `test(TP-084): add golden coverage for roadmap behavior`
- `hydrate: TP-084 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not commit unredacted live API payloads.
- Do not convert structured markers into generic errors.

---

## Amendments

_Add amendments below this line only._
