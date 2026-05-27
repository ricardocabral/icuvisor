# Task: TP-110 - Workout description schema regression tests

**Created:** 2026-05-27
**Size:** S

## Review Level: 1

**Assessment:** Adds test coverage around registered tool metadata and schema wording. Runtime behavior should not change, but this guards LLM-facing contracts that affect write safety.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-110-workout-description-schema-regression/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add regression tests that fail if workout write tools reintroduce contradictory `description`/`workout_doc` wording such as "mutually exclusive" while the server supports merged prose plus structured steps. This protects assistants from stale or conflicting parameter-level instructions that can cause failed writes or unsafe fallbacks.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository testing and MCP schema conventions.
- `CHANGELOG.md` — update only if the project tracks test-only safety hardening.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `internal/tools/*_test.go`
- `internal/tools/catalog_test.go`
- `internal/toolchecks/**/*`
- `internal/tools/schema_snapshot/add_or_update_event.json`
- `internal/tools/schema_snapshot/create_workout.json`
- `internal/tools/schema_snapshot/update_workout.json`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing schema/catalog tests identified

### Step 1: Add metadata invariant tests

> **Plan-review checkpoint** — Confirm the invariant location and exact tools covered before implementation.

- [ ] Add a regression test over registered tool/input-schema descriptions for `add_or_update_event`, `create_workout`, and `update_workout`.
- [ ] Assert those descriptions do not contain `mutually exclusive` or equivalent contradictory wording for `description` and `workout_doc`.
- [ ] Assert the schema still advertises coexistence/merge behavior or the sentinel guidance.
- [ ] Keep the test focused on LLM-facing contract text, not exact prose snapshots unless necessary.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolchecks`

**Artifacts:**
- `internal/tools/*_test.go` or `internal/toolchecks/**/*` (modified/new)

### Step 2: Refresh affected snapshots and docs if needed

- [ ] Regenerate schema snapshots if the chosen invariant exposes stale snapshot text.
- [ ] Update `CHANGELOG.md` only if this safety-hardening test is tracked there.
- [ ] Ensure no generated docs still contain the contradictory wording.

**Artifacts:**
- `internal/tools/schema_snapshot/*.json` (modified if needed)
- `CHANGELOG.md` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run lint if available locally: `make lint`
- [ ] Build passes: `make build`
- [ ] Fix all failures or document pre-existing unrelated failures in `STATUS.md`

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified if required.
- [ ] "Check If Affected" docs reviewed.
- [ ] Discoveries logged in `STATUS.md`.
- [ ] Commit at step boundary with the task ID in the message.

## Documentation Requirements

**Must Update:**
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `CHANGELOG.md` — update only if test-only safety hardening is tracked under `[Unreleased]`.
- Generated schema snapshots/docs — update if any still contain contradictory wording.

## Completion Criteria

- A regression test fails if the relevant write-tool descriptions say `description` and `workout_doc` are mutually exclusive.
- The test covers at least `add_or_update_event`, `create_workout`, and `update_workout`.
- Targeted and full tests pass or unrelated pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-110` for traceability. Examples:

- `test(TP-110): complete step 1 — guard workout description schema wording`
- `hydrate: TP-110 expand step checkboxes`

## Do NOT

- Do not make the invariant brittle against harmless wording changes.
- Do not change runtime write behavior in this task.
- Do not remove existing schema snapshot coverage.
- Do not broaden into full tool-description rewriting; TP-111 owns wording improvements.

---

## Amendments

_Add amendments below this line only._
