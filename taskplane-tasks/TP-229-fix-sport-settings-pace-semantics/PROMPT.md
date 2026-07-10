# Task: TP-229 - Treat threshold pace as m/s and pace zones as percentages

**Created:** 2026-07-10
**Size:** L

## Review Level: 3 (Full)

**Assessment:** This corrects unit semantics across upstream decoding, response shaping, writes, fixtures, and destructive zone replacement. Incorrect conversion can publish impossible paces or overwrite an athlete's pace zones with invalid values.
**Score:** 6/8 — Blast radius: 2, Pattern novelty: 2, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-229-fix-sport-settings-pace-semantics/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Correct icuvisor's sport-settings pace model to match intervals.icu's public behavior: `threshold_pace` is stored in SI meters per second for run and swim, while `pace_units` is a display preference; `pace_zones` are percentage-of-threshold boundaries, not durations. Convert read values into explicit athlete-facing seconds-per-distance fields, convert write inputs back to m/s, expose zone percentages truthfully, and replace synthetic fixtures that currently encode seconds as upstream threshold values. Preserve unknown-unit fallback behavior and avoid copying the competing PR's incorrect minute-based conversion.

## Dependencies

- **Task:** TP-228 (sport-settings update transport and public schema must be corrected first)

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — unit and sport-settings response requirements
- `ROADMAP.md` — stable-line constraints
- `https://forum.intervals.icu/t/not-a-bug-run-threshold-pace-units/123710` — public confirmation that threshold pace is m/s and pace_units is display-only
- `https://forum.intervals.icu/t/server-side-data-model-for-scripts/25781/16` — public server-model description for pace-zone percentages
- `https://intervals.icu/api/v1/docs` — current SportSettings schema

## Environment

- **Workspace:** repository root
- **Services required:** None; no real athlete writes

## File Scope

- `internal/intervals/types.go`
- `internal/intervals/testdata/athlete_profile.json`
- `internal/units/*`
- `internal/response/units.go`
- `internal/response/units_test.go`
- `internal/athleteprofile/profile.go`
- `internal/athleteprofile/profile_test.go`
- `internal/tools/update_sport_settings.go`
- `internal/tools/update_sport_settings_test.go`
- `internal/tools/update_sport_settings_zones_test.go`
- `internal/tools/sport_settings_pace_semantics_test.go`
- `internal/tools/schema_snapshot/update_sport_settings.json`
- `web/data/tools.json`
- `web/data/tool_schemas.json`
- `docs/prd/PRD-icuvisor.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] TP-228 is complete
- [ ] Confirm the public m/s and percentage semantics from the listed upstream evidence

### Step 1: Define canonical pace conversions

**Plan-review checkpoint**

- [ ] Define read conversion as seconds per requested distance equals distance metres divided by threshold m/s
- [ ] Define write conversion as threshold m/s equals distance metres divided by supplied seconds
- [ ] Define `pace_units` as presentation metadata only and `pace_zones` as percent-of-threshold boundaries
- [ ] Decide a compatibility-safe response migration for currently misnamed pace-zone duration fields; never continue returning false semantics
- [ ] Run targeted tests: `go test ./internal/units ./internal/response ./internal/athleteprofile`

**Artifacts:**

- `internal/response/units.go` (modified)
- `internal/athleteprofile/profile.go` (modified)

### Step 2: Correct read shaping and typed models

- [ ] Decode and preserve `threshold_pace`, `pace_units`, `pace_load_type`, and percentage zone boundaries with typed fields
- [ ] Emit correct threshold pace in seconds per km, mile, 100 m, 100 yd, or 500 m according to sport/display context
- [ ] Emit pace-zone boundaries as percentages with an unambiguous field name and metadata
- [ ] Preserve unknown upstream units without failing the whole profile
- [ ] Run targeted tests: `go test ./internal/athleteprofile ./internal/tools -run 'Profile|Pace'`

**Artifacts:**

- `internal/intervals/types.go` (modified)
- `internal/athleteprofile/profile.go` (modified)
- `internal/athleteprofile/profile_test.go` (modified)

### Step 3: Correct sport-settings writes

- [ ] Convert threshold pace inputs from explicit seconds/minutes-per-distance units to m/s before transport
- [ ] Send `pace_units` only as the athlete display convention and include `pace_load_type` when needed to disambiguate run/swim
- [ ] Treat pace-zone write boundaries as percentages, validate their range/order, and remove descriptions that request duration values
- [ ] Keep zone replacement behind the existing full delete-mode gate
- [ ] Run targeted tests: `go test ./internal/tools -run 'UpdateSportSettings|Pace|Zone'`

**Artifacts:**

- `internal/tools/update_sport_settings.go` (modified)
- `internal/tools/update_sport_settings_test.go` (modified)
- `internal/tools/update_sport_settings_zones_test.go` (modified)

### Step 4: Replace misleading fixtures and lock semantics

- [ ] Replace the synthetic profile fixture with realistic m/s threshold values and percent pace-zone boundaries
- [ ] Add table-driven round-trip tests for run metric/imperial, 100 m and 100 yd swim, and 500 m rowing pace
- [ ] Add regression tests proving 3.5714285 m/s becomes 280 seconds/km and 280 seconds/km writes back to approximately 3.5714285 m/s
- [ ] Assert pace-zone values such as 77.5 and 100 remain percentages and are never converted as durations
- [ ] Run targeted tests: `go test ./internal/response ./internal/athleteprofile ./internal/tools`

**Artifacts:**

- `internal/intervals/testdata/athlete_profile.json` (modified)
- `internal/tools/sport_settings_pace_semantics_test.go` (new)
- `internal/response/units_test.go` (modified)

### Step 5: Testing & Verification

**Code review checkpoint**

- [ ] Run FULL test suite: `make test`
- [ ] Run race suite: `make test-race`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Regenerate docs: `make docs-tools`
- [ ] Verify no generated or golden drift remains: `git diff --check`

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `docs/prd/PRD-icuvisor.md` — state that upstream threshold pace is m/s, pace_units is presentation-only, and pace_zones are percentages
- `CHANGELOG.md` — record corrected read/write pace semantics and migration implications

**Check If Affected:**

- Generated tool reference and schema data
- `web/content/cookbook/ftp-and-zones.md` — ensure user examples use the corrected percentage-zone contract
- `internal/resources/*` — update only if sport-setting semantics are duplicated there

## Completion Criteria

- [ ] Threshold pace reads and writes round-trip through m/s correctly
- [ ] Pace zones are represented and validated as percentages
- [ ] Fixture values match real upstream semantics
- [ ] Misleading seconds/minutes assumptions are removed from code, schemas, and docs
- [ ] New semantic regression test file exists
- [ ] Full tests, race, lint, build, and generated docs pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-229:

- **Step completion:** `fix(TP-229): complete Step N — description`
- **Bug fixes:** `fix(TP-229): description`
- **Tests:** `test(TP-229): description`
- **Hydration:** `hydrate: TP-229 expand Step N checkboxes`

## Do NOT

- Copy the competing implementation's minute-based threshold conversion
- Treat `MINS_KM` or `MINS_MILE` as the storage unit for threshold_pace
- Convert pace-zone percentages as pace durations
- Perform live writes or use real credentials
- Remove unknown-unit fallback behavior
- Bypass delete-mode gating for zone replacement
- Expand into indoor FTP/create support; TP-233 owns that feature
- Commit without TP-229 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
