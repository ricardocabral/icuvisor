# TP-082 — Null stripping for write-tool responses

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Cross-tool response-shaping sweep with public behavior impact; tests are the main deliverable.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-082-write-response-null-stripping/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Verify and enforce that write-tool echo responses are shaped as tersely as reads: null keys from upstream payloads are stripped by default, while `include_full` preserves raw detail. This prevents sparse event/wellness/workout/custom-item echoes from wasting context.

## Dependencies

- **Task:** TP-007 (response shaper exists)
- **Task:** TP-020 (event write cluster exists)
- **Task:** TP-021 (`update_wellness` exists)
- **Task:** TP-022 (`update_sport_settings` exists)
- **Task:** TP-023 (workout-library writes exist)
- **Task:** TP-024 (custom-item writes exist)
- **Task:** TP-026 (`apply_training_plan` exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/response/* — shaper behavior and tests.
- internal/tools/add_or_update_event.go, update_wellness.go, update_sport_settings.go, create_workout.go, update_workout.go, create_custom_item.go, update_custom_item.go — write response paths.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/response/*`
- `internal/tools/add_or_update_event*.go`
- `internal/tools/update_wellness*.go`
- `internal/tools/update_sport_settings*.go`
- `internal/tools/create_workout*.go`
- `internal/tools/update_workout*.go`
- `internal/tools/create_custom_item*.go`
- `internal/tools/update_custom_item*.go`
- `internal/tools/apply_training_plan*.go`
- `internal/intervals/testdata/**/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Audit write response shaping

- [ ] List every write tool that returns an upstream echo payload.
- [ ] Identify whether each path applies the same shaper/null-strip policy as reads.
- [ ] Record any intentional exception in STATUS.md before changing code.

### Step 2: Add failing golden tests

- [ ] Add or update write-tool fixture tests that include sparse/null upstream fields.
- [ ] Assert terse default strips null keys and preserves meaningful zero/false/empty-string values.
- [ ] Assert `include_full` or debug path preserves raw detail where supported.

### Step 3: Apply shared shaping consistently

- [ ] Route write responses through shared response shaping rather than bespoke map cleanup.
- [ ] Preserve existing `_meta.server_version`, scale labels, and missing-field behavior.
- [ ] Avoid changing request payloads.

### Step 4: Verify full write cluster

- [ ] Run targeted write-tool tests.
- [ ] Run `make test`, `make build`, and `make lint`.
- [ ] Update CHANGELOG.md.

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

- Every listed write tool has fixture-backed null-stripping coverage.
- Terse defaults strip nulls; meaningful zero/false/empty values remain.
- Read response behavior is unchanged.
- Full suite/build/lint pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-082` for traceability. Examples:

- `feat(TP-082): complete step 1 — scope current behavior`
- `fix(TP-082): repair regression found during analyzer tests`
- `test(TP-082): add golden coverage for roadmap behavior`
- `hydrate: TP-082 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not alter upstream write request payloads while fixing response shaping.

---

## Amendments

_Add amendments below this line only._
