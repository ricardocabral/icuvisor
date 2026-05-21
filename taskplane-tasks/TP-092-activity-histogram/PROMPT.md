# TP-092 — `get_activity_histogram` single-activity histogram tool

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** New stream-backed analyzer-adjacent tool with public schema and math behavior.
**Score:** 5/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-092-activity-histogram/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add `get_activity_histogram` for per-activity power/HR/pace distribution so assistants can answer 'how was this workout distributed?' without pulling streams and binning per-second samples themselves. Buckets should prefer athlete zones, fall back to fixed-width buckets, and report `_meta.bucket_method`.

## Dependencies

- **Task:** TP-087 (`analysis_metric` enum exists)
- **Task:** TP-089 (analyzer meta skeleton exists)
- **Task:** TP-008 (stream-key/unit canonicalization exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_activity_streams.go — raw stream access pattern.
- internal/tools/get_activity_details.go and sport settings/zones helpers — zone context.
- internal/streams/* — canonical stream keys.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/histogram*.go`
- `internal/tools/get_activity_histogram*.go`
- `internal/tools/get_activity_streams*.go`
- `internal/tools/get_activity_details*.go`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `internal/streams/*`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Define histogram contract

- [ ] Limit metrics to power/HR/pace through the closed metric enum.
- [ ] Define bucket output fields: label/range, seconds, percentage, unit.
- [ ] Define `_meta.bucket_method` values for configured zones vs fixed-width fallback.

### Step 2: Implement stream-backed histogram

- [ ] Fetch only required streams for one activity.
- [ ] Use athlete configured zones where available; fall back to fixed-width buckets with documented edges.
- [ ] Return terse per-bucket summary only; no raw samples.

### Step 3: Tool registration and activation hint

- [ ] Register in `full` with description leading on single-activity distribution prompts.
- [ ] Explicitly say not to pull `get_activity_streams` and bin manually.
- [ ] Add source_tools/method/n meta.

### Step 4: Tests and verification

- [ ] Add fixtures for zone-based and fixed-width buckets.
- [ ] Test unit conversion and missing stream handling.
- [ ] Run full quality gate and update docs/CHANGELOG.

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

- Tool returns per-bucket time and percentage without raw samples.
- `_meta.bucket_method` states zones vs fixed-width fallback.
- Missing streams/insufficient samples return structured terse errors/meta.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-092` for traceability. Examples:

- `feat(TP-092): complete step 1 — scope current behavior`
- `fix(TP-092): repair regression found during analyzer tests`
- `test(TP-092): add golden coverage for roadmap behavior`
- `hydrate: TP-092 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not include raw stream samples in terse default responses.
- Do not conflate this single-activity tool with multi-activity `analyze_distribution`.

---

## Amendments

_Add amendments below this line only._
