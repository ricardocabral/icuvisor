# Task: TP-103 - Tags in terse read responses

**Created:** 2026-05-26
**Size:** M

## Review Level: 1

**Assessment:** Additive public response-shape work across existing event/activity read paths using established shaping patterns.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-103-tags-in-terse-read-responses/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Expose upstream tag values in terse read responses so assistants can filter and reason about athlete-defined labels without requiring `include_full`. icuvisor can write event tags today; read tools should surface them as first-class fields when upstream returns them.

Tracking issue: https://github.com/ricardocabral/icuvisor/issues/30

## Dependencies

- **Task:** TP-009 (activity read cluster exists)
- **Task:** TP-012 (event read cluster exists)
- **Task:** TP-020 (event write cluster with tag writes exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and Go/MCP conventions.
- `docs/prd/PRD-icuvisor.md` — response-shaping and terse/full contracts.
- `internal/intervals/events.go` — event model and raw payload preservation.
- `internal/intervals/activities.go` — activity model and raw payload preservation.
- `internal/tools/get_events.go` — shared event row shaping.
- `internal/tools/get_activities_row.go` — activity row shaping.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None. Unit tests must not hit the network.

## File Scope

- `internal/intervals/events.go`
- `internal/intervals/events_test.go`
- `internal/intervals/activities.go`
- `internal/intervals/activity_*test.go`
- `internal/tools/get_events.go`
- `internal/tools/get_event_by_id.go`
- `internal/tools/get_today.go`
- `internal/tools/get_activities_row.go`
- `internal/tools/get_activities*.go`
- `internal/tools/get_activity_details.go`
- `internal/tools/*tags*_test.go`
- `internal/tools/*events*_test.go`
- `internal/tools/*activities*_test.go`
- `internal/intervals/testdata/**/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm issue #30 scope and avoid unrelated tag/filter features

### Step 1: Implement event tag read shaping

- [ ] Decode or extract upstream `tags` for events without guessing missing values.
- [ ] Add `tags` to shared terse event rows while preserving order.
- [ ] Ensure `get_events`, `get_event_by_id`, `add_or_update_event`, and `get_today` event sections share the same behavior.
- [ ] Run targeted event tests.

**Artifacts:**
- `internal/intervals/events.go` (modified if typed decode is used)
- `internal/tools/get_events.go` (modified)
- Event-related tests/fixtures (modified/new)

### Step 2: Investigate and implement activity tag handling if supported

- [ ] Inspect current activity fixtures/models for upstream activity tag availability.
- [ ] If activity tags are returned by the read endpoints, expose them on `get_activities` and `get_activity_details` terse rows.
- [ ] If activity tags are not returned by current endpoints/fixtures, add a regression test or discovery note documenting that no field is emitted.
- [ ] Run targeted activity tests.

**Artifacts:**
- `internal/intervals/activities.go` (modified only if needed)
- `internal/tools/get_activities_row.go` (modified only if needed)
- Activity-related tests/fixtures (modified/new)

### Step 3: Regression tests and docs

- [ ] Add tests for tags present, empty, null/missing, and `include_full` behavior.
- [ ] Verify null stripping does not emit guessed or empty fields except where an explicit empty upstream list should round-trip.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` for user-visible response additions.
- [ ] Run targeted tests for all touched read/write response paths.

**Artifacts:**
- `internal/tools/*_test.go` (modified/new)
- `internal/intervals/testdata/**/*` (modified/new)
- `CHANGELOG.md` (modified)

### Step 4: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — record the user-visible response addition under `[Unreleased]`.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `README.md` — update only if public examples mention event/activity response fields.
- `docs/prd/PRD-icuvisor.md` — update only if behavior intentionally changes product scope.
- Generated tool-reference data/docs — regenerate if catalog/response docs require it.

## Completion Criteria

- `get_events` terse rows include `tags` in upstream order when present.
- `get_event_by_id`, `add_or_update_event`, and `get_today` event sections expose the same tag shape through shared code.
- Activity tag availability is implemented or explicitly documented in tests/discoveries.
- Null/missing/non-array tag payloads do not panic or produce guessed values.
- `make test`, `make build`, and `make lint` pass or pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-103` for traceability. Examples:

- `feat(TP-103): complete step 1 — expose event tags`
- `test(TP-103): add tag response fixtures`
- `hydrate: TP-103 expand step checkboxes`

## Do NOT

- Do not add tag filtering/search unless already required to expose returned tag values.
- Do not guess tag names from descriptions, names, or categories.
- Do not emit API keys, athlete-identifying data, or live upstream payloads in tests.
- Do not broaden into unrelated response-shape refactors.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
