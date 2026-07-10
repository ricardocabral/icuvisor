# TP-229: Treat threshold pace as m/s and pace zones as percentages — Status

**Current Step:** Step 2: Correct read shaping and typed models
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 3
**Review Counter:** 10
**Iteration:** 1
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] TP-228 is complete
- [x] Public m/s and percentage semantics confirmed

---

### Step 1: Define canonical pace conversions

**Status:** ✅ Complete

- [x] Read and write m/s formulas defined
- [x] pace_units presentation role defined
- [x] pace_zones percentage contract defined
- [x] Compatibility migration decided
- [x] R002: Apply the declared m/s and percentage response migration before advertising it in `_meta`
- [x] R003: Propagate every recognized pace display unit and treat `NONE` as a known m/s fallback
- [x] R004: Reject finite pace inputs whose reciprocal conversion overflows
- [x] R005: Correct workout previews and configured histogram zones for m/s thresholds and percentage boundaries
- [x] R005: Omit ambiguous source-unit fallback values
- [x] R006: Honor every recognized `pace_units` display distance in workout target previews
- [x] R007: Route previews through finite canonical pace conversion
- [x] R008: Reject pace durations at the signed-int formatting boundary

---

### Step 2: Correct read shaping and typed models

**Status:** 🟨 In Progress

- [ ] Typed upstream fields completed
- [ ] Threshold pace read shaping corrected
- [ ] Percentage zone response added
- [ ] Unknown-unit fallback preserved
- [ ] R010: Document the typed-field, response/fallback, and table-driven coverage plan

---

### Step 3: Correct sport-settings writes

**Status:** ⬜ Not Started

- [ ] Explicit pace inputs convert to m/s
- [ ] pace_units and pace_load_type are correct
- [ ] Pace-zone percentage validation implemented
- [ ] Delete-mode zone gate preserved

---

### Step 4: Replace misleading fixtures and lock semantics

**Status:** ⬜ Not Started

- [ ] Realistic upstream fixture values installed
- [ ] Run/swim/row round-trip scenarios covered
- [ ] m/s regression assertions added
- [ ] Percentage zones remain unchanged

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] Build passes
- [ ] Generated docs clean

---

### Step 6: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | UNAVAILABLE | — |
| R002 | Code | 1 | REVISE | `.reviews/R002-code-step1.md` |
| R003 | Code | 1 | REVISE | `.reviews/R003-code-step1.md` |
| R004 | Code | 1 | REVISE | `.reviews/R004-code-step1.md` |
| R005 | Code | 1 | REVISE | `.reviews/R005-code-step1.md` |
| R006 | Code | 1 | REVISE | `.reviews/R006-code-step1.md` |
| R007 | Code | 1 | REVISE | `.reviews/R007-code-step1.md` |
| R008 | Code | 1 | REVISE | `.reviews/R008-code-step1.md` |
| R009 | Code | 1 | APPROVE | — |
| R010 | Plan | 2 | REVISE | `.reviews/R010-plan-step2.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| The planned `internal/athleteprofile/profile_test.go` does not exist; equivalent profile coverage currently lives under `internal/tools/get_athlete_profile_test.go`. | Create focused athleteprofile package tests during Step 2 if needed; all implementation paths and fixture locations exist. | `internal/athleteprofile/`, `internal/tools/get_athlete_profile_test.go` |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 19:11 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 19:11 | Step 0 started | Preflight |

## Blockers

- TP-228 must complete first.

## Notes

- Preflight evidence (2026-07-10): the upstream forum confirms `threshold_pace` is always stored in SI m/s and `MINS_KM` is GUI-only presentation metadata; the server-model reference defines `pace_zones` as percentage-of-threshold boundaries.
- Step 1 plan: centralize two inverse conversions: display seconds = selected pace-distance metres / stored m/s, and stored m/s = selected pace-distance metres / supplied seconds. `pace_units` selects only the display distance, while `pace_zones` are copied as percent values to a newly named percentage field; existing `pace_zones_seconds_per_*` fields will be omitted rather than returned with false duration semantics. The migration is intentionally additive for correct values (`pace_zones_percent_of_threshold`) and omits false legacy duration fields rather than retaining a deprecated lie. Preserve unknown display units by returning m/s plus the raw `pace_units` source. Live OpenAPI confirms `pace_load_type` values `RUN`/`SWIM` and presentation enums including `SECS_100M`, `SECS_100Y`, `MINS_KM`, `MINS_MILE`, and `SECS_500M`.
| 2026-07-10 19:30 | Review R002 | code Step 1: REVISE |
| 2026-07-10 19:39 | Review R003 | code Step 1: REVISE |
| 2026-07-10 19:45 | Review R004 | code Step 1: REVISE |
| 2026-07-10 19:50 | Review R005 | code Step 1: REVISE |
| 2026-07-10 20:00 | Review R006 | code Step 1: REVISE |
| 2026-07-10 20:05 | Review R007 | code Step 1: REVISE |
| 2026-07-10 20:11 | Review R008 | code Step 1: REVISE |
| 2026-07-10 20:15 | Review R009 | code Step 1: APPROVE |
| 2026-07-10 20:18 | Review R010 | plan Step 2: REVISE |
