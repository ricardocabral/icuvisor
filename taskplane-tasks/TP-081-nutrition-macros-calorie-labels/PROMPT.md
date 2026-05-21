# TP-081 — Nutrition macros and calories-label clarification

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Public response-shape additions across read clusters; existing shaping patterns apply.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-081-nutrition-macros-calorie-labels/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Surface upstream nutrition macros under disambiguated keys and make calorie semantics explicit across wellness and activity reads. The assistant should distinguish active calories burned from any total/caloric-intake fields and should not guess hidden field names.

## Dependencies

- **Task:** TP-009 (activity reads exist)
- **Task:** TP-011 (wellness reads exist)
- **Task:** TP-007 (response shaping primitives exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/intervals/activities.go and wellness.go — decoded upstream payloads.
- internal/tools/get_activities_row.go, get_activity_details.go, get_wellness_data.go — response shaping.
- internal/response/scales.go and meta.go — meta labels if needed.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/intervals/activities*.go`
- `internal/intervals/wellness*.go`
- `internal/tools/get_activities*.go`
- `internal/tools/get_activity_details*.go`
- `internal/tools/get_wellness_data*.go`
- `internal/response/*`
- `internal/tools/testdata/*`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Map nutrition fields from upstream fixtures

- [ ] Identify carbs/protein/fat/calorie fields exposed on wellness and activity payloads.
- [ ] Add typed fields with JSON names and fixtures; if an expected field is absent, document the gap.
- [ ] Choose disambiguated response keys such as `carbs_g`, `protein_g`, `fat_g`, `calories_burned`, `calories_intake`/`calories_total` only when upstream semantics support them.

### Step 2: Shape activity and wellness responses

- [ ] Expose macro fields only when present; keep null stripping intact.
- [ ] Ensure `calories_burned` remains active/exercise calories and does not collide with intake/total fields.
- [ ] Add `_meta` labels where semantics could be confused.

### Step 3: Regression tests

- [ ] Add fixture-backed tests for activities and wellness with nutrition present and absent.
- [ ] Assert key names and null stripping behavior.
- [ ] Run targeted read-tool tests.

### Step 4: Docs and full verification

- [ ] Update tool reference/examples for nutrition fields.
- [ ] Update CHANGELOG.md.
- [ ] Run full quality gate.

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

- Nutrition fields are surfaced with units in field names.
- Calories fields are semantically disambiguated and documented.
- Absent nutrition data is stripped/called out consistently with existing missing-field rules.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-081` for traceability. Examples:

- `feat(TP-081): complete step 1 — scope current behavior`
- `fix(TP-081): repair regression found during analyzer tests`
- `test(TP-081): add golden coverage for roadmap behavior`
- `hydrate: TP-081 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not invent macro totals if upstream does not provide them.
- Do not overload `calories` as an ambiguous key.

---

## Amendments

_Add amendments below this line only._
