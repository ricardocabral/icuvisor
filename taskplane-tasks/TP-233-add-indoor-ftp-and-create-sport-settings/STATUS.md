# TP-233: Support indoor FTP and missing sport-setting creation — Status

**Current Step:** Step 1: Extend the typed client
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 2
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] TP-228 and TP-229 are complete
- [x] Public create contract confirmed

---

### Step 1: Extend the typed client

**Status:** 🟨 In Progress

- [ ] Indoor FTP update support added
- [ ] Typed CreateSportSettings operation added
- [ ] Corrected threshold pace conversion reused
- [ ] Threshold validation defined without invented constraints
- [ ] Targeted client tests pass
- [x] R001 plan: exact typed boundary, sparse POST contract, validation, and client regression coverage recorded

---

### Step 2: Add and register MCP surfaces

**Status:** ⬜ Not Started

- [ ] update_sport_settings indoor_ftp field added
- [ ] create_sport_settings tool implemented
- [ ] Full-toolset registration and annotations added
- [ ] Schemas, examples, catalog, and snapshots updated

---

### Step 3: Regression and safety coverage

**Status:** ⬜ Not Started

- [ ] Exact create wire tests added
- [ ] Invalid input avoids network calls
- [ ] Credential/confirm/zone arguments excluded
- [ ] Tool counts and catalog guards updated

---

### Step 4: Generated docs and public contract

**Status:** ⬜ Not Started

- [ ] PRD catalog updated
- [ ] Website tool/schema data regenerated
- [ ] Changelog updated

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
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 22:35 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 22:35 | Step 0 started | Preflight |
| 2026-07-10 22:40 | Worker iter 1 | done in 264s, tools: 25 |

## Blockers

- TP-228 and TP-229 must complete first.

## Notes

| Date | Topic | Detail |
|---|---|---|
| 2026-07-10 | R001 Step 1 plan | Add `IndoorFTP *int` only to `WriteSportSettingsParams`; introduce separate `CreateSportSettingsParams{Sport, FTP, IndoorFTP, ThresholdHR, ThresholdPace}` and return `SportSettings`. `SportSettingsPace` is pre-normalized m/s plus selected `PaceUnits` and explicit/preserved `PaceLoadType`, serialized without client reinterpretation. |
| 2026-07-10 | R001 HTTP contract | Update remains `PUT /athlete/{athleteID}/sport-settings/{id}?recalcHrZones=<bool>` and writes `indoor_ftp` only when supplied. Create is no-retry `POST /athlete/{athleteID}/sport-settings`, no query string, with sparse `types:[sport]` and only `ftp`, `indoor_ftp`, `lthr`, `threshold_pace` (m/s), `pace_units`, and `pace_load_type`; it cannot carry ID, recalculation, or zones. |
| 2026-07-10 | R001 validation | Before transport reject blank sport, non-positive FTP/indoor FTP/HR, and non-finite/non-positive canonical pace; errors name create/update and make no request. No `indoor_ftp <= ftp` restriction: no confirmed upstream/product rule. Client leaves sport enum ownership to the MCP layer. |
| 2026-07-10 | R001 client tests | Local-server cases will assert update-only `indoor_ftp` and create `types:["Ride"]`/indoor FTP method/path/raw-query/sparse body/no zone-or-recalc fields plus returned echo, m/s pace keys, and table-driven invalid update/create calls with zero requests. |
