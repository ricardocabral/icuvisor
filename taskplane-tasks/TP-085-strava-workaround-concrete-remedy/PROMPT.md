# TP-085 — Concrete Strava-import unavailable workaround text

**Created:** 2026-05-20
**Size:** S

## Review Level: 1

**Assessment:** Small user-facing response text change with fixture coverage; plan review is enough unless code changes expand.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-085-strava-workaround-concrete-remedy/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Make Strava-sourced empty/blocked activity markers tell users the concrete recovery action: trigger Download old data for the native device provider on the intervals.icu Connections page, so historical activities are re-imported directly from Garmin/Wahoo/etc. instead of through Strava's restricted API.

## Dependencies

- **Task:** TP-009 (Strava-blocked activity detection exists)
- **Task:** TP-084 (upstream-signal fixtures provide related coverage)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/activity_unavailable.go — structured unavailable helper.
- internal/tools/get_activities_strava.go and tests — Strava marker detection.
- internal/tools/get_activity_details.go — detail unavailable path.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/tools/activity_unavailable.go`
- `internal/tools/get_activities_strava*.go`
- `internal/tools/get_activity_details*.go`
- `internal/tools/get_activity_streams*.go`
- `internal/tools/testdata/**/*`
- `web/content/guides/troubleshooting.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Audit current unavailable wording

- [ ] Find all Strava/import-blocked unavailable marker construction paths.
- [ ] Identify whether native provider can be inferred from current payloads.
- [ ] Define provider-aware and provider-unknown workaround strings.

### Step 2: Update marker text

- [ ] Return the concrete intervals.icu Connections-page remedy when Strava-blocked data is detected.
- [ ] Mention provider name when known; use provider-neutral wording when unknown.
- [ ] Keep `unavailable.reason` stable and structured.

### Step 3: Fixture assertions

- [ ] Add/modify fixtures to assert the exact workaround string, not only `reason`.
- [ ] Cover at least one known native provider and one unknown-provider case.
- [ ] Run targeted Strava/unavailable tests.

### Step 4: Docs and verification

- [ ] Update troubleshooting docs with the same remedy wording.
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

- Strava-blocked markers include actionable Connections-page instructions.
- Provider-aware wording appears where source can be known.
- Tests assert workaround strings exactly.
- Reason codes remain stable.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-085` for traceability. Examples:

- `feat(TP-085): complete step 1 — scope current behavior`
- `fix(TP-085): repair regression found during analyzer tests`
- `test(TP-085): add golden coverage for roadmap behavior`
- `hydrate: TP-085 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not tell users to bypass Strava restrictions; the remedy is direct native-provider import into intervals.icu.

---

## Amendments

_Add amendments below this line only._
