# TP-083 — Per-source sleep-score scale labels in wellness provenance

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Focused wellness response-shape change with user-facing semantics.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-083-per-source-sleep-score-scales/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Make `_meta.provenance` report source-specific native sleep/readiness scales for Garmin, Whoop, Oura, and Polar instead of assuming a single canonical 0–100 label. This keeps assistants from comparing bridged wellness values across incompatible native scales.

## Dependencies

- **Task:** TP-011 (wellness provenance and `_native` sidecars exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_wellness_data.go and tests — wellness response shaping.
- internal/intervals/wellness.go — upstream wellness fields.
- internal/response/scales.go — existing scale labels.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/tools/get_wellness_data*.go`
- `internal/intervals/wellness*.go`
- `internal/response/scales.go`
- `internal/response/meta.go`
- `internal/tools/testdata/wellness/*`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Define source scale mapping

- [ ] Review existing provenance fields and native sidecars.
- [ ] Define exact `native_scale` labels for at least Garmin, Whoop, Oura, and Polar based on public docs/observed fixtures.
- [ ] Represent unknown sources as `unknown`, not a guessed scale.

### Step 2: Apply provenance labels

- [ ] Update wellness shaping so each bridged field uses source-specific `native_scale`.
- [ ] Keep response-level canonical scale labels separate from native provenance labels.
- [ ] Ensure stale/provenance behavior remains intact.

### Step 3: Fixture coverage

- [ ] Add fixtures for at least two divergent sources and assert the exact `native_scale` strings.
- [ ] Add unknown-source fallback test.
- [ ] Run targeted wellness tests.

### Step 4: Docs and verification

- [ ] Update tool docs/reference wording.
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

- At least two divergent provider fixtures assert source-specific `native_scale` labels.
- Unknown source does not claim a scale.
- Existing wellness stale/provenance behavior remains green.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-083` for traceability. Examples:

- `feat(TP-083): complete step 1 — scope current behavior`
- `fix(TP-083): repair regression found during analyzer tests`
- `test(TP-083): add golden coverage for roadmap behavior`
- `hydrate: TP-083 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not collapse native and canonical scales into one label.
- Do not silently omit provenance when source cannot be determined.

---

## Amendments

_Add amendments below this line only._
