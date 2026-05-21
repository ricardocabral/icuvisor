# TP-090 — Auto-lap disambiguation on `get_activity_intervals`

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Adds inference heuristics to a stable read tool and affects analyzer behavior.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-090-auto-lap-disambiguation/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add additive `_meta.interval_source` and `_meta.auto_lap_suspected` signals to `get_activity_intervals` when intervals look like device auto-laps rather than structured workout segments, and ensure interval-consuming analyzers propagate/decline per-interval execution claims when the signal is true.

## Dependencies

- **Task:** TP-012 (event/workout read context exists)
- **Task:** TP-089 (analyzer meta skeleton exists before analyzer propagation wiring)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_activity_intervals* (or activity interval read implementation) — current interval response path.
- internal/intervals/activity_details.go and fixtures — interval payloads.
- docs/prd/PRD-icuvisor.md — auto-lap analyzer requirement.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/intervals/activity_details*.go`
- `internal/tools/get_activity_intervals*.go`
- `internal/tools/analyze_efforts_delta*.go`
- `internal/tools/compute_compliance_rate*.go`
- `internal/analysis/intervals*.go`
- `internal/tools/testdata/**/*`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Model interval-source heuristics

- [ ] Inspect current interval payload fields to find any explicit source markers.
- [ ] Define near-uniform distance/duration heuristic with unit-aware tolerances.
- [ ] Document false-positive tradeoffs in STATUS.md.

### Step 2: Add additive meta to interval reads

- [ ] Emit `_meta.interval_source` as `structured_workout`, `device_laps`, or `unknown`.
- [ ] Emit `_meta.auto_lap_suspected: true` for near-uniform auto-lap patterns.
- [ ] Do not remove or rename existing interval fields.

### Step 3: Propagate to analyzers

- [ ] Add shared helper so analyzer source_tools/meta can propagate auto-lap suspicion.
- [ ] Ensure interval-consuming analyzers decline per-interval execution quality claims when suspicion is true.
- [ ] Add placeholder propagation tests if analyzer tools are not implemented yet.

### Step 4: Tests and docs

- [ ] Add fixtures for structured intervals, 1 km/1 mi auto-laps, and unknown source.
- [ ] Update docs/reference and CHANGELOG.md.
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

- `get_activity_intervals` emits additive source/auto-lap metadata.
- Near-uniform auto-laps are detected in tests.
- Analyzers have a propagation/decline path for auto-lap suspicion.
- Stable schema fields are not broken.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-090` for traceability. Examples:

- `feat(TP-090): complete step 1 — scope current behavior`
- `fix(TP-090): repair regression found during analyzer tests`
- `test(TP-090): add golden coverage for roadmap behavior`
- `hydrate: TP-090 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not remove interval rows or change existing field names.
- Do not claim structured-workout compliance when auto-lap is suspected.

---

## Amendments

_Add amendments below this line only._
