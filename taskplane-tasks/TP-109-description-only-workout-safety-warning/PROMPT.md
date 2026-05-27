# Task: TP-109 - Description-only workout safety warning

**Created:** 2026-05-27
**Size:** S

## Review Level: 1

**Assessment:** Small write-tool safety improvement touching one tool family, tests, and public metadata. The behavior is additive and easy to revert, but it changes LLM-facing write guidance so it needs plan review.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-109-description-only-workout-safety-warning/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a safety warning/guard for description-only updates on workout-shaped objects so assistants do not silently clear or replace structured workout steps when they only meant to add prose. The immediate incident was `add_or_update_event` updating `WORKOUT` calendar events with `description` but no `workout_doc`; the worker should also check `update_workout` for the same risk and apply the same additive warning pattern if affected.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and MCP write-tool conventions.
- `docs/prd/PRD-icuvisor.md` — event/workout write contract and safety expectations.
- `CHANGELOG.md` — record user-visible warning behavior.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/update_workout.go`
- `internal/tools/update_workout_test.go`
- `internal/tools/schema_snapshot/add_or_update_event.json`
- `internal/tools/schema_snapshot/update_workout.json`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current add/update event and workout update behavior reviewed

### Step 1: Design the warning/guard contract

> **Plan-review checkpoint** — Confirm the warning shape and whether `update_workout` needs the same treatment before implementation.

- [ ] Decide the minimal additive response metadata field(s) for description-only workout updates.
- [ ] Ensure the warning is terse, actionable, and does not block legitimate strength-session prose updates.
- [ ] Document the exact trigger condition, at minimum: `description` supplied, no `workout_doc`, and the target is a workout-shaped write (`category: WORKOUT` for events; equivalent for workout templates if affected).

**Artifacts:**
- `internal/tools/add_or_update_event.go` (modified plan target)
- `internal/tools/update_workout.go` (modified only if affected)

### Step 2: Implement warning behavior and tests

- [ ] Add additive `_meta` warning behavior for description-only `WORKOUT` event updates without `workout_doc`.
- [ ] Add equivalent `update_workout` warning behavior if a description-only sparse template update can replace structured content.
- [ ] Add table-driven tests covering warning-present and warning-absent cases.
- [ ] Regenerate or update schema snapshots if output/input metadata changes require it.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/add_or_update_event.go` (modified)
- `internal/tools/add_or_update_event_test.go` (modified)
- `internal/tools/update_workout.go` / `_test.go` (modified if affected)
- `internal/tools/schema_snapshot/*.json` (modified if generated snapshots change)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run lint if available locally: `make lint`
- [ ] Build passes: `make build`
- [ ] Fix all failures or document pre-existing unrelated failures in `STATUS.md`

### Step 4: Documentation & Delivery

- [ ] Update `CHANGELOG.md` under `[Unreleased]` for the new safety warning.
- [ ] "Check If Affected" docs reviewed.
- [ ] Discoveries logged in `STATUS.md`.
- [ ] Commit at step boundary with the task ID in the message.

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — mention additive warning for description-only workout writes.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `web/content/cookbook/build-workouts.md` — update only if user-facing guidance needs a short note.
- `internal/prompts/catalog.go` / weekly-planning prompt text — update only if prompt guardrails need to reference the warning.

## Completion Criteria

- Description-only `WORKOUT` event updates without `workout_doc` produce an actionable warning in `_meta`.
- `update_workout` is either covered by the same warning or explicitly documented in `STATUS.md` as not affected.
- Tests prove warning-present and warning-absent behavior.
- Full tests/build pass or unrelated pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-109` for traceability. Examples:

- `fix(TP-109): complete step 2 — warn on description-only workout updates`
- `test(TP-109): cover description-only workout warning`
- `hydrate: TP-109 expand step checkboxes`

## Do NOT

- Do not add a model-controlled `confirm` override.
- Do not make description-only strength-session updates impossible unless the PRD explicitly requires it.
- Do not log API keys, athlete IDs, or raw private workout prose.
- Do not broaden this into a workout DSL refactor.

---

## Amendments

_Add amendments below this line only._
