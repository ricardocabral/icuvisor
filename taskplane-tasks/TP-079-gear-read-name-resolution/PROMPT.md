# TP-079 — Gear read/name-resolution pass

**Created:** 2026-05-20
**Size:** L

## Review Level: 2

**Assessment:** Adds a new read tool plus activity response shaping across multiple files; existing tool patterns apply.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 1, Security: 1, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-079-gear-read-name-resolution/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add gear read support and inline gear-name resolution so activity rows can carry both `gear_id` and human-readable `gear_name` when upstream exposes gear IDs. The LLM should not have to chain through a separate lookup tool to understand which bike/shoes/device an activity used.

## Dependencies

- **Task:** TP-009 (activities read cluster exists)
- **Task:** TP-025 (destructive gear delete gating context exists)
- **Task:** TP-030 (`full` toolset tier exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/intervals/activities.go and activity fixtures — current activity payload shape.
- internal/tools/get_activities*.go and get_activity_details.go — row/detail shaping.
- internal/tools/delete_gear.go — existing destructive gear behavior.
- internal/tools/catalog.go — tool registration and toolset placement.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/intervals/gear*.go`
- `internal/intervals/activities*.go`
- `internal/tools/get_gear_list*.go`
- `internal/tools/get_activities*.go`
- `internal/tools/get_activity_details*.go`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `internal/response/*`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Probe and model upstream gear payloads

- [ ] Identify the intervals.icu gear-list endpoint and activity gear fields from public docs/black-box fixtures.
- [ ] Add typed intervals client structs and fixtures for gear list responses.
- [ ] Document any upstream gap in `docs/upstream-gaps/` if gear IDs are not exposed consistently.

### Step 2: Implement `get_gear_list` and cache/refresh behavior

- [ ] Add the read-only tool with terse default response and an explicit refresh path if a cache is used.
- [ ] Register the tool in the appropriate toolset without changing `delete_gear` safety gating.
- [ ] Add tests for athlete-ID normalization and empty gear lists.

### Step 3: Inline resolve gear on activity reads

- [ ] Surface `gear_id` and resolved `gear_name` in `get_activities` rows when available.
- [ ] Surface the same fields in `get_activity_details`.
- [ ] Ensure unresolved IDs remain explicit instead of guessed.

### Step 4: Verify and document

- [ ] Run targeted intervals/tools tests, then full suite.
- [ ] Update generated/user docs for the new tool and activity fields.
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

- `get_gear_list` is registered and tested.
- `get_activities` and `get_activity_details` include `gear_id` and `gear_name` when upstream data permits.
- Gear delete remains destructive and gated exactly as before.
- Unresolved/missing gear data is explicit and terse.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-079` for traceability. Examples:

- `feat(TP-079): complete step 1 — scope current behavior`
- `fix(TP-079): repair regression found during analyzer tests`
- `test(TP-079): add golden coverage for roadmap behavior`
- `hydrate: TP-079 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not add a model-controlled delete override.
- Do not require callers to chain a lookup before activity reads.

---

## Amendments

_Add amendments below this line only._
