# Task: TP-111 - Clarify description replacement wording

**Created:** 2026-05-27
**Size:** S

## Review Level: 1

**Assessment:** LLM-facing wording changes across write-tool descriptions and docs. Low code risk, but wording directly affects write behavior and safety.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-111-description-replacement-wording/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Clarify that `description` on workout/event write tools is a replacement for the upstream description field/DSL, not an append-only notes field. The wording should tell assistants that preserving existing structured workout steps requires supplying `workout_doc` (or using the merge sentinel with an explicit structured doc), especially on updates.

## Dependencies

- **Task:** TP-109 (warning terminology should align with the safety warning/guard)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and MCP schema conventions.
- `docs/prd/PRD-icuvisor.md` — event/workout write contract.
- `web/content/cookbook/build-workouts.md` — user-facing workout write recipe.
- `CHANGELOG.md` — record user-visible tool-description changes.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `internal/tools/add_or_update_event.go`
- `internal/tools/create_workout.go`
- `internal/tools/update_workout.go`
- `internal/tools/update_activity.go`
- `internal/tools/schema_snapshot/add_or_update_event.json`
- `internal/tools/schema_snapshot/create_workout.json`
- `internal/tools/schema_snapshot/update_workout.json`
- `internal/tools/schema_snapshot/update_activity.json`
- `internal/prompts/catalog.go`
- `internal/prompts/testdata/weekly_planning.md`
- `web/content/cookbook/build-workouts.md`
- `web/content/explain/calendar-notes.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing write-tool descriptions and docs reviewed

### Step 1: Update write-tool wording

> **Plan-review checkpoint** — Confirm the final wording pattern before applying it across tools.

- [ ] Update `add_or_update_event` schema/tool text to say `description` replaces the event description/DSL and may clear structured steps if used alone on workouts.
- [ ] Update `create_workout` and `update_workout` wording to distinguish creating a merged description from replacing an existing template description.
- [ ] Check `update_activity` wording for consistency without implying activity metadata descriptions carry planned workout structure.
- [ ] Regenerate or update schema snapshots for changed descriptions.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/*.go` (modified)
- `internal/tools/schema_snapshot/*.json` (modified)

### Step 2: Update prompt/docs wording

- [ ] Update `weekly_planning` prompt text if it currently implies `description` is a safe append-only notes channel.
- [ ] Update cookbook/explainer docs to say update writes replace descriptions and bulk edits should preserve structured steps explicitly.
- [ ] Update prompt golden tests if prompt text changes.
- [ ] Update `CHANGELOG.md` under `[Unreleased]`.
- [ ] Run targeted tests: `go test ./internal/prompts ./internal/tools`

**Artifacts:**
- `internal/prompts/catalog.go` / `internal/prompts/testdata/weekly_planning.md` (modified if affected)
- `web/content/**/*` (modified if affected)
- `CHANGELOG.md` (modified)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run lint if available locally: `make lint`
- [ ] Build passes: `make build`
- [ ] Fix all failures or document pre-existing unrelated failures in `STATUS.md`

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified.
- [ ] "Check If Affected" docs reviewed.
- [ ] Discoveries logged in `STATUS.md`.
- [ ] Commit at step boundary with the task ID in the message.

## Documentation Requirements

**Must Update:**
- Tool schema descriptions for affected write tools.
- `CHANGELOG.md` — note the clearer replacement-field wording.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `web/content/cookbook/build-workouts.md` — align user recipe with replacement semantics.
- `web/content/explain/calendar-notes.md` — keep NOTE guidance simple and accurate.
- `internal/prompts/catalog.go` and prompt golden files — align curated prompt guardrails.

## Completion Criteria

- LLM-facing wording clearly states `description` is a replacement field on writes/updates.
- Workout guidance explicitly says to include `workout_doc` when structured steps must be preserved.
- Schema snapshots and prompt golden files are updated where affected.
- Full tests/build pass or unrelated pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-111` for traceability. Examples:

- `docs(TP-111): complete step 2 — clarify description replacement wording`
- `fix(TP-111): clarify write-tool description semantics`
- `hydrate: TP-111 expand step checkboxes`

## Do NOT

- Do not change runtime behavior except wording/snapshot fallout; TP-109 owns warning behavior.
- Do not reintroduce `description`/`workout_doc` mutually-exclusive wording.
- Do not overstate that every description-only update is destructive; phrase the risk specifically around structured workout steps.
- Do not include private athlete examples.

---

## Amendments

_Add amendments below this line only._
