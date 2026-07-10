# Task: TP-237 - Add coaching conversation handoff prompt

**Created:** 2026-07-10
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** This adapts the established MCP prompt and client prompt-pack patterns across one service and documentation surface. It is read-only and reversible, but privacy and source-grounding details need plan review.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-237-coaching-handoff-prompt/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a read-only `coaching_handoff` MCP prompt and portable client prompt pack that lets an athlete close a long conversation and start a fresh one without losing the durable coaching context. The handoff must separate user-stated decisions from live Icuvisor evidence, include athlete-local dates and evidence timestamps, identify unresolved questions and stale/missing data, and exclude credentials, raw athlete identifiers, local paths, and unsupported physiological conclusions.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — prompt catalog, privacy, timezone, and terse-response rules
- `internal/prompts/catalog.go` — registered MCP prompt patterns
- `internal/prompts/testdata/shareable_training_report.md` — privacy-safe report precedent
- `web/content/guides/claude-project-instructions.md` — current long-lived client guidance

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/registry.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/coaching_handoff_test.go`
- `internal/prompts/testdata/coaching_handoff.md`
- `docs/prompts/client-prompt-packs/coaching-handoff.md`
- `web/content/cookbook/conversation-handoff.md`
- `web/content/cookbook/_index.md`
- `web/content/guides/claude-project-instructions.md`
- `web/content/reference/resources-prompts.md`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `docs/prd/PRD-icuvisor.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing prompt registration, golden, portable-pack, and eval patterns reviewed

### Step 1: Define the handoff contract

**Plan-review checkpoint**

- [ ] Define a compact Markdown handoff with goals, constraints, accepted decisions, current plan state, evidence dates/sources, unresolved questions, and next actions
- [ ] Distinguish statements from the current conversation from facts retrieved through Icuvisor tools
- [ ] Define privacy exclusions and rules for stale, missing, partial, or paginated data
- [ ] Keep the workflow read-only and suitable for pasting into Claude, ChatGPT, Cursor, or another client
- [ ] Run targeted tests: `go test ./internal/prompts -run 'CoachingHandoff'`

**Artifacts:**

- `internal/prompts/catalog.go` (modified)
- `docs/prompts/client-prompt-packs/coaching-handoff.md` (new)

### Step 2: Register the prompt and add golden coverage

- [ ] Add `coaching_handoff` with bounded optional lookback/race context arguments and the minimum necessary read tools
- [ ] Register the prompt and update prompt catalog expectations without adding write/delete instructions
- [ ] Add a deterministic golden fixture and a new focused test file covering argument rendering, privacy exclusions, source separation, and athlete-local date anchoring
- [ ] Ensure missing advanced tools route through `icuvisor_list_advanced_capabilities` rather than chat-side invented calculations
- [ ] Run targeted tests: `go test ./internal/prompts -run 'Prompt|CoachingHandoff'`

**Artifacts:**

- `internal/prompts/catalog.go` (modified)
- `internal/prompts/registry.go` (modified)
- `internal/prompts/coaching_handoff_test.go` (new)
- `internal/prompts/testdata/coaching_handoff.md` (new)

### Step 3: Publish the portable workflow and eval

- [ ] Add a cookbook page and downloadable/copyable client prompt pack with a clear fresh-chat workflow
- [ ] Add an eval scenario that requires source-labelled decisions, explicit freshness, unresolved questions, and no credentials/IDs
- [ ] Update the prompt reference, cookbook index, Claude Project guidance, PRD catalog, and Unreleased changelog
- [ ] Run targeted checks: `go test ./internal/prompts && python3 scripts/eval/run_eval.py --validate`

**Artifacts:**

- `docs/prompts/client-prompt-packs/coaching-handoff.md` (new)
- `web/content/cookbook/conversation-handoff.md` (new)
- `web/content/cookbook/_index.md` (modified)
- `web/content/guides/claude-project-instructions.md` (modified)
- `web/content/reference/resources-prompts.md` (modified)
- `scripts/eval/scenarios/cookbook_scenarios.json` (modified)
- `docs/prd/PRD-icuvisor.md` (modified)
- `CHANGELOG.md` (modified)

### Step 4: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run prompt eval validation: `make eval-validate`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify clean Markdown and diff: `git diff --check`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `web/content/cookbook/conversation-handoff.md` — publish the workflow and expected output
- `web/content/reference/resources-prompts.md` — document the registered prompt
- `docs/prd/PRD-icuvisor.md` — add the prompt to the current catalog
- `CHANGELOG.md` — record the new prompt and prompt pack

**Check If Affected:**

- `web/content/guides/claude-project-instructions.md` — link the handoff workflow where long-chat guidance appears
- `web/content/cookbook/_index.md` — add the new recipe card and prompt count
- `README.md` — update only if it lists every registered prompt

## Completion Criteria

- [ ] `coaching_handoff` is registered and read-only
- [ ] Handoff output separates user decisions from sourced tool evidence
- [ ] Dates, freshness, missing data, and unresolved questions are explicit
- [ ] Credentials, raw athlete IDs, local paths, and unsupported conclusions are excluded
- [ ] New focused test, golden fixture, client prompt pack, cookbook page, and eval scenario exist
- [ ] Full tests, eval validation, lint, build, and diff checks pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-237:

- **Step completion:** `feat(TP-237): complete Step N — description`
- **Bug fixes:** `fix(TP-237): description`
- **Tests:** `test(TP-237): description`
- **Hydration:** `hydrate: TP-237 expand Step N checkboxes`

## Do NOT

- Add persistent server-side conversation storage
- Include API keys, athlete IDs, local paths, config files, or raw secrets in the handoff
- Present model memory or chat summaries as Icuvisor-sourced facts
- Call write or delete tools
- Dump raw streams or full histories by default
- Claim a client automatically imports or remembers the handoff
- Commit without TP-237 in the message

---

## Amendments (Added During Execution)

### R001 plan clarification

- `docs/prompts/README.md` is added to File Scope and Documentation Requirements so the portable pack remains discoverable from the existing pack index.
- Step 1 owns the contract-bearing `CoachingHandoffPrompt` function and portable pack, plus focused contract tests in `internal/prompts/coaching_handoff_test.go`. Step 2 registers the function, adds it to shared catalog/golden expectations, and creates the deterministic golden fixture.
- Step 1 verification is non-vacuous: `go test ./internal/prompts -run 'CoachingHandoff'` runs the focused contract tests created in Step 1.
