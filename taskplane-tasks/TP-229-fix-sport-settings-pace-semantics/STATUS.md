# TP-229: Treat threshold pace as m/s and pace zones as percentages — Status

**Current Step:** Step 1: Define canonical pace conversions
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 3
**Review Counter:** 0
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

**Status:** 🟨 In Progress

- [ ] Read and write m/s formulas defined
- [ ] pace_units presentation role defined
- [ ] pace_zones percentage contract defined
- [ ] Compatibility migration decided

---

### Step 2: Correct read shaping and typed models

**Status:** ⬜ Not Started

- [ ] Typed upstream fields completed
- [ ] Threshold pace read shaping corrected
- [ ] Percentage zone response added
- [ ] Unknown-unit fallback preserved

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
