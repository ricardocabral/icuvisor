# TP-086 — NOTE-event discoverability pass

**Created:** 2026-05-20
**Size:** S

## Review Level: 1

**Assessment:** Primarily documentation/tool-example discoverability with minimal code risk.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-086-note-event-discoverability/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Improve docs and examples so assistants and users discover that `add_or_update_event` with `category: "NOTE"` is the supported way to create nutrition plans, travel logistics, daily reminders, and coach annotations. Do not add a confusable `add_note` tool unless future telemetry proves tool-selection failure.

## Dependencies

- **Task:** TP-075 (`add_or_update_event` NOTE create succeeds)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- docs/upstream-gaps/event-note-payload.md — NOTE payload contract.
- web/content/reference/tools.md and generated tool data — public tool docs.
- internal/tools/add_or_update_event.go — schema/examples if needed.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/tools/add_or_update_event*.go`
- `internal/tools/catalog*.go`
- `web/content/reference/tools.md`
- `web/content/guides/*`
- `web/content/tutorials/*`
- `README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Audit current NOTE discoverability

- [ ] Search docs and tool schema/examples for `category: "NOTE"` coverage.
- [ ] Identify whether `add_or_update_event` input examples are generated from code or static docs.
- [ ] Decide whether code-level `input_examples` need an additive example.

### Step 2: Add examples in the right surfaces

- [ ] Document NOTE examples for nutrition plans, travel logistics, daily reminders, and coach annotations.
- [ ] Add code/catalog examples if the generated reference pulls from tool metadata.
- [ ] Keep wording clear that NOTE is an event category, not a separate tool.

### Step 3: Validate docs generation

- [ ] Regenerate docs/tool reference if applicable.
- [ ] Run doc/site build if docs changed under `web/`.
- [ ] Run targeted tests if tool metadata changed.

### Step 4: Delivery

- [ ] Update CHANGELOG.md if tool metadata or public docs changed.
- [ ] Run `make test` if code changed; otherwise run docs/site verification.
- [ ] Record any telemetry question for future work in STATUS.md.

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

- Docs show at least four NOTE use cases with `add_or_update_event` and `category: "NOTE"`.
- No new `add_note` tool is added.
- Generated docs remain in sync if tool metadata changes.
- Site/docs build passes; code tests pass if code changed.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-086` for traceability. Examples:

- `feat(TP-086): complete step 1 — scope current behavior`
- `fix(TP-086): repair regression found during analyzer tests`
- `test(TP-086): add golden coverage for roadmap behavior`
- `hydrate: TP-086 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not add a separate `add_note` tool.
- Do not imply NOTE descriptions are normalized into workout DSL.

---

## Amendments

_Add amendments below this line only._
