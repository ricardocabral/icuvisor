# Task: TP-158 - Sport-settings profile readiness warnings

**Created:** 2026-06-09
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds user-visible readiness warnings to athlete profile shaping and may affect profile resource/tool schemas. It is read-only and reversible, but touches planning-critical output contracts.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-158-profile-readiness-warnings/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add focused sport-settings preflight diagnostics so planners can detect missing thresholds/zones before producing bad analysis or training plans. Public IcuSync/LeCoach feedback highlights missing thresholds/zones and threshold-pace update misses; icuvisor should surface readiness warnings through `get_athlete_profile` (and the athlete-profile resource if shared) before workouts are planned.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repo rules and clean-room constraints.
- `docs/prd/PRD-icuvisor.md` — tool catalog/profile output expectations.
- `ROADMAP.md` — phasing for prompts/resources/tools.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/athleteprofile/profile.go`
- `internal/resources/athlete_profile.go`
- `internal/resources/athlete_profile_test.go`
- `internal/tools/get_athlete_profile.go`
- `internal/tools/get_athlete_profile_test.go`
- `internal/tools/schema_snapshot/get_athlete_profile.json`
- `internal/tools/update_sport_settings_test.go`
- `README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Design and add readiness warning shape

- [ ] Add a compact `_meta.warnings` shape in `internal/athleteprofile/profile.go` with stable codes such as `missing_power_threshold`, `missing_hr_threshold`, `missing_pace_threshold`, `missing_power_zones`, `missing_hr_zones`, and `missing_pace_zones`.
- [ ] Scope warnings by sport setting/types and keep them terse by default; do not include raw athlete identifiers or raw upstream payloads.
- [ ] Include enough context for an LLM to decide whether to ask the user to update sport settings before analysis/planning.
- [ ] Run targeted tests: `go test ./internal/athleteprofile ./internal/tools -run 'Test.*AthleteProfile|Test.*Sport'`

**Artifacts:**
- `internal/athleteprofile/profile.go` (modified)
- `internal/tools/get_athlete_profile_test.go` (modified)

### Step 2: Propagate to tool/resource schemas and tests

- [ ] Add tests proving `get_athlete_profile` emits warnings for missing ride/run/swim thresholds/zones and omits warnings when settings are complete.
- [ ] Add or update `internal/resources/athlete_profile_test.go` so the `icuvisor://athlete-profile` resource shares the same warnings if it uses shared profile shaping.
- [ ] Refresh `internal/tools/schema_snapshot/get_athlete_profile.json` if output description/schema text changes.
- [ ] Review `update_sport_settings` tests/descriptions so warnings point users to the correct write tool without implying model-controlled credentials.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/resources`

**Artifacts:**
- `internal/resources/athlete_profile_test.go` (modified)
- `internal/tools/schema_snapshot/get_athlete_profile.json` (modified if needed)
- `internal/tools/update_sport_settings_test.go` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `README.md` documents profile readiness warnings if profile output is described there.
- [ ] `CHANGELOG.md` notes sport-settings preflight diagnostics.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note added readiness warnings.

**Check If Affected:**
- `README.md` — update get_athlete_profile output examples/descriptions if present.
- `docs/prd/PRD-icuvisor.md` — update only if this formalizes a new public contract.

## Completion Criteria

- [ ] Missing sport thresholds/zones produce deterministic, terse warnings.
- [ ] Complete sport settings do not produce false warnings.
- [ ] Shared tool/resource profile shaping remains consistent.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-158): complete Step N — description`
- **Bug fixes:** `fix(TP-158): description`
- **Tests:** `test(TP-158): description`
- **Hydration:** `hydrate: TP-158 expand Step N checkboxes`

## Do NOT

- Add a model-controlled credential or API-key argument.
- Add destructive writes or mutate sport settings in a read-profile task.
- Copy competitor code; use only public behavior signals.
- Emit raw athlete IDs in warnings beyond existing normalized display behavior.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
