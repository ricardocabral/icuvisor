# Task: TP-151 - Event external_id idempotency

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** This extends the calendar write contract and affects event/create/update paths plus plan application. It is reversible but touches upstream write payloads and public MCP schemas.
**Score:** 5/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-151-event-external-id-idempotency/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Make calendar writes retry-safer by adding explicit `external_id` support to icuvisor event writes and by using deterministic external IDs where icuvisor can safely own idempotency, especially `apply_training_plan`. Public WorkoutContext behavior showed duplicates when bulk upserts were attempted without stable IDs; icuvisor currently preflights exact same-day duplicates but cannot protect retries/near-concurrent writes as strongly because `WriteEventParams` has no `external_id`.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — clean-room, write/delete safety, and MCP conventions.
- `docs/prd/PRD-icuvisor.md` — event write/idempotency requirements.
- `docs/dogfood/v0.3-findings.md` — prior write dogfood failures and safety notes.
- `docs/upstream-gaps/event-note-payload.md` — upstream event create payload behavior.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/intervals/events.go`
- `internal/intervals/events_test.go`
- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/apply_training_plan.go`
- `internal/tools/apply_training_plan_test.go`
- `internal/tools/get_events.go`
- `internal/tools/get_event_by_id.go`
- `internal/tools/schema_snapshot/add_or_update_event.json`
- `internal/tools/schema_snapshot/apply_training_plan.json`
- `internal/toolrouting/testdata/cases.json`
- `web/content/cookbook/season-and-block-plan.md`
- `web/content/cookbook/build-workouts.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Identify current event read/write handling for raw `external_id`

### Step 1: Design external_id contract

**Plan-review checkpoint** — Review before changing public schemas.

- [ ] Decide exact `add_or_update_event.external_id` semantics for create, update, omit, and clear-if-supported
- [ ] Decide how `apply_training_plan` generates deterministic external IDs without colliding with user/provider IDs
- [ ] Decide whether event read rows should expose `external_id` in terse mode, full mode, or `_meta`
- [ ] Record any upstream uncertainty in STATUS.md and choose conservative behavior

**Artifacts:**
- `internal/intervals/events.go` (modified after plan acceptance)
- `internal/tools/add_or_update_event.go` (modified after plan acceptance)
- `STATUS.md` (discoveries)

### Step 2: Implement event write/read support

- [ ] Add `ExternalID` to `intervals.WriteEventParams` and event payload with tests proving JSON body shape
- [ ] Add `external_id` to `add_or_update_event` input schema/decoder and handler mapping
- [ ] Add tests for create/update preserving external ID and duplicate preflight behavior remaining intact
- [ ] Decide/read-test whether terse event rows include `external_id` for auditability
- [ ] Run targeted tests: `go test ./internal/intervals ./internal/tools -run 'Event|AddOrUpdateEvent'`

**Artifacts:**
- `internal/intervals/events.go` (modified)
- `internal/intervals/events_test.go` (modified)
- `internal/tools/add_or_update_event.go` (modified)
- `internal/tools/add_or_update_event_test.go` (modified)

### Step 3: Make apply_training_plan retry-safer

- [ ] Generate stable external IDs for plan-applied events from plan/workout/date or another documented deterministic tuple
- [ ] Add tests proving repeated apply payloads are stable and do not rely only on same-day matching
- [ ] Ensure dry-run exposes enough proposed metadata for user review without leaking sensitive IDs unnecessarily
- [ ] Run targeted tests: `go test ./internal/tools -run 'ApplyTrainingPlan'`

**Artifacts:**
- `internal/tools/apply_training_plan.go` (modified)
- `internal/tools/apply_training_plan_test.go` (modified)

### Step 4: Refresh schemas, routing, and docs

- [ ] Regenerate schema snapshots: `go run ./scripts/snapshot_tool_schemas.go`
- [ ] Update tool-routing expectations if event write prompts should mention idempotency/external IDs
- [ ] Update user docs for retry-safe calendar writes if public behavior changed
- [ ] Add a CHANGELOG `[Unreleased]` entry

**Artifacts:**
- `internal/tools/schema_snapshot/add_or_update_event.json` (modified)
- `internal/tools/schema_snapshot/apply_training_plan.json` (modified if affected)
- `internal/toolrouting/testdata/cases.json` (modified if affected)
- `web/content/cookbook/season-and-block-plan.md` (modified if affected)
- `CHANGELOG.md` (modified)

### Step 5: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md
- [ ] Summarize remaining idempotency caveats clearly

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — document new event idempotency/external ID behavior.

**Check If Affected:**
- `web/content/cookbook/season-and-block-plan.md` — update scheduling guidance if `apply_training_plan` behavior changes.
- `web/content/cookbook/build-workouts.md` — update event/template write guidance if relevant.
- `docs/prd/PRD-icuvisor.md` — update only if the product contract changes.

## Completion Criteria

- [ ] `add_or_update_event` accepts and writes `external_id`
- [ ] Event write payload tests prove upstream body shape
- [ ] `apply_training_plan` uses deterministic external IDs or documents why not
- [ ] Schema snapshots and docs are current
- [ ] All tests passing

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-151): complete Step N — description`
- **Bug fixes:** `fix(TP-151): description`
- **Tests:** `test(TP-151): description`
- **Hydration:** `hydrate: TP-151 expand Step N checkboxes`

## Do NOT

- Copy or transliterate competitor source. Use only the public behavior signal summarized in this prompt and upstream Intervals behavior.
- Add a bulk destructive tool or delete capability as part of this task.
- Generate external IDs that could collide with known provider-owned prefixes such as `hevy-` or `strava-`.
- Skip tests.
- Modify framework/standards docs without explicit user approval.
- Load docs not listed in "Context to Read First".
- Commit without the task ID prefix in the commit message.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
