# TP-096 — Activation-hint pass on analyzer descriptions

**Created:** 2026-05-20
**Size:** S

## Review Level: 1

**Assessment:** Catalog/documentation quality pass over public tool descriptions.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-096-analyzer-activation-hints/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Ensure every analyzer description leads with the user prompt shape that should trigger the tool and explicitly tells the LLM not to pull `get_*` rows/streams and reduce them itself. This is a catalog-quality pass after analyzer tools exist.

## Dependencies

- **Task:** TP-091 (`analyze_*` tools exist)
- **Task:** TP-092 (`get_activity_histogram` exists)
- **Task:** TP-093 (`compute_*` tools exist)
- **Task:** TP-094 (`compute_activity_segment_stats` exists)
- **Task:** TP-095 (`get_fitness_projection` exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/catalog.go and individual analyzer tool files — tool descriptions.
- internal/toolchecks/confusable_names.go — catalog quality checks.
- web/content/reference/tools.md / generated tool data — rendered docs.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/tools/analyze_*.go`
- `internal/tools/compute_*.go`
- `internal/tools/get_activity_histogram*.go`
- `internal/tools/get_fitness_projection*.go`
- `internal/toolchecks/*`
- `web/data/tools.json`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Audit analyzer descriptions

- [ ] List every analyzer-family tool and current first sentence.
- [ ] Identify missing activation hints and missing do-not-roll-your-own language.
- [ ] Check for confusable wording across similar tools.

### Step 2: Update descriptions and checks

- [ ] Rewrite descriptions so each starts with a concrete prompt shape.
- [ ] Add explicit raw-row/stream avoidance language where applicable.
- [ ] Add or update catalog tests/checks to keep future analyzer descriptions compliant.

### Step 3: Regenerate docs/catalog artifacts

- [ ] Run the catalog/docs generation command if descriptions feed generated docs.
- [ ] Review rendered docs for clarity and no stale examples.
- [ ] Update CHANGELOG.md if user-visible docs changed.

### Step 4: Verify

- [ ] Run targeted toolcheck/catalog tests.
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

- Every analyzer-family description has a strong activation hint.
- Every relevant description tells the LLM not to reduce raw `get_*` data itself.
- Catalog/toolcheck tests prevent drift.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-096` for traceability. Examples:

- `feat(TP-096): complete step 1 — scope current behavior`
- `fix(TP-096): repair regression found during analyzer tests`
- `test(TP-096): add golden coverage for roadmap behavior`
- `hydrate: TP-096 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not change tool names or remove arguments during this pass.

---

## Amendments

_Add amendments below this line only._
