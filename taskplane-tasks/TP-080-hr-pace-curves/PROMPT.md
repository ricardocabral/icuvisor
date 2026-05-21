# TP-080 — HR and pace curve siblings to get_power_curves

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Two new tools adapt an existing pattern but touch the public catalog, so code review is required.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-080-hr-pace-curves/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add `get_hr_curves` and `get_pace_curves` as symmetric siblings to `get_power_curves`, using the same response conventions, pagination, units, and terse/full behavior. This lets assistants answer best-effort curve questions without pulling raw streams when upstream already exposes the data cheaply.

## Dependencies

- **Task:** TP-010 (`get_power_curves` pattern exists)
- **Task:** TP-008 (unit enum and pace-unit handling exist)
- **Task:** TP-030 (toolset tiers exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_power_curves.go and tests — canonical curve tool pattern.
- internal/intervals/fitness.go — existing curves client methods.
- internal/units or internal/response units tests — pace/HR units and preferred-unit conventions.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/intervals/fitness*.go`
- `internal/tools/get_power_curves*.go`
- `internal/tools/get_hr_curves*.go`
- `internal/tools/get_pace_curves*.go`
- `internal/tools/fitness_metrics_shared_test.go`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Extract reusable curve plumbing

- [ ] Review `get_power_curves` and identify shared request/response shaping that applies to HR and pace.
- [ ] Add/adjust typed intervals client support for HR and pace curve endpoints or document upstream gaps.
- [ ] Keep power-curve behavior unchanged with regression tests.

### Step 2: Implement HR and pace curve tools

- [ ] Create `get_hr_curves` and `get_pace_curves` tool files with schema descriptions that name units and scales.
- [ ] Apply athlete preferred-unit conversion for pace output.
- [ ] Register tools in the correct tier and catalog tests.

### Step 3: Test curve symmetry

- [ ] Add table-driven tests covering power/HR/pace shared shape, terse default, `include_full`, pagination if applicable, and unknown units.
- [ ] Run targeted tool/client tests.

### Step 4: Docs and full verification

- [ ] Update tool docs/generated reference.
- [ ] Update CHANGELOG.md.
- [ ] Run `make test`, `make build`, and `make lint`.

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

- Both tools are registered and discoverable.
- Response shape is symmetric with `get_power_curves` and unit-aware.
- No raw stream fetch is needed for curve use cases.
- Tests cover terse/full responses and unit handling.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-080` for traceability. Examples:

- `feat(TP-080): complete step 1 — scope current behavior`
- `fix(TP-080): repair regression found during analyzer tests`
- `test(TP-080): add golden coverage for roadmap behavior`
- `hydrate: TP-080 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not derive curves from streams if upstream exposes curve endpoints.
- Do not rename or change `get_power_curves`.

---

## Amendments

_Add amendments below this line only._
