# Task: TP-112 - Bulk calendar write preflight guidance

**Created:** 2026-05-27
**Size:** S

## Review Level: 1

**Assessment:** Updates prompts/docs to guide safer operational behavior for bulk workout writes. Low runtime risk, but it changes assistant guardrails for write workflows.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-112-bulk-calendar-write-preflight-guidance/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add operational guidance so assistants do not fire many calendar/workout writes in parallel when the schema or preservation semantics are uncertain. For bulk workout edits, the assistant should validate the intended payload, perform one representative write plus readback, inspect warnings/structured-step preservation, and only then proceed with the remaining batch.

## Dependencies

- **Task:** TP-109 (warning behavior should exist before docs reference warning semantics)
- **Task:** TP-111 (wording should first clarify replacement semantics)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and docs conventions.
- `docs/prd/PRD-icuvisor.md` — write-tool and validation expectations.
- `web/content/cookbook/build-workouts.md` — primary user-facing workout-write recipe.
- `web/content/cookbook/season-and-block-plan.md` — bulk scheduling guidance.
- `internal/prompts/catalog.go` — curated weekly-planning prompt guardrails.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None.

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/testdata/weekly_planning.md`
- `web/content/cookbook/build-workouts.md`
- `web/content/cookbook/season-and-block-plan.md`
- `web/content/reference/resources-prompts.md`
- `docs/dogfood/v0.3-prompts.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing prompt and cookbook guidance reviewed

### Step 1: Add curated prompt guardrails

> **Plan-review checkpoint** — Confirm the one-write/readback rule wording before updating goldens.

- [ ] Update `weekly_planning` or other relevant curated prompt text with a bulk-write preflight rule.
- [ ] Rule says: validate or preview representative structured payload, write one event/template, read it back, inspect `_meta` warnings and `workout_doc_summary`, then proceed with the rest.
- [ ] Rule says: avoid parallel bulk writes when schema wording, warning metadata, or preservation semantics are ambiguous.
- [ ] Update prompt golden files.
- [ ] Run targeted tests: `go test ./internal/prompts`

**Artifacts:**
- `internal/prompts/catalog.go` (modified)
- `internal/prompts/testdata/*.md` (modified)

### Step 2: Add user-facing cookbook guidance

- [ ] Update workout-building and/or season/block planning cookbook pages with the representative write/readback pattern.
- [ ] Explain why this matters: write tools replace descriptions and structured steps can be lost if omitted.
- [ ] Keep guidance concise and client-neutral.
- [ ] Update `CHANGELOG.md` under `[Unreleased]` for docs/prompt guardrail change.

**Artifacts:**
- `web/content/cookbook/*.md` (modified)
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
- Relevant curated prompt text and golden files.
- At least one user-facing cookbook page that covers workout/calendar writes.
- `CHANGELOG.md` — mention safer bulk write preflight guidance.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `web/content/reference/resources-prompts.md` — update only if prompt summary needs to mention the guardrail.
- `docs/dogfood/v0.3-prompts.md` — update only if dogfood prompts should include one-write/readback preflight.

## Completion Criteria

- Curated prompts tell assistants to validate/preview and perform one representative write/readback before bulk workout/calendar writes.
- Cookbook guidance includes the same operational pattern in user-facing language.
- Prompt golden tests pass.
- Full tests/build pass or unrelated pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-112` for traceability. Examples:

- `docs(TP-112): complete step 2 — add bulk write preflight guidance`
- `test(TP-112): update weekly planning prompt golden`
- `hydrate: TP-112 expand step checkboxes`

## Do NOT

- Do not change runtime write behavior in this task.
- Do not instruct assistants to paste API keys or private payloads into docs/examples.
- Do not ban parallel writes universally; scope the guidance to workout/calendar writes where schema or preservation semantics are uncertain.
- Do not make docs verbose enough to bury the recipe.

---

## Amendments

_Add amendments below this line only._
