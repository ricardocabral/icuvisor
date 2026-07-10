# TP-233: Support indoor FTP and missing sport-setting creation — Status

**Current Step:** Step 4: Generated docs and public contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 9
**Iteration:** 4
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] TP-228 and TP-229 are complete
- [x] Public create contract confirmed

---

### Step 1: Extend the typed client

**Status:** ✅ Complete

- [x] Indoor FTP update support added
- [x] Typed CreateSportSettings operation added
- [x] Corrected threshold pace conversion reused
- [x] Threshold validation defined without invented constraints
- [x] Targeted client tests pass
- [x] R001 plan: exact typed boundary, sparse POST contract, validation, and client regression coverage recorded

---

### Step 2: Add and register MCP surfaces

**Status:** ✅ Complete

- [x] update_sport_settings indoor_ftp field added
- [x] create_sport_settings tool implemented
- [x] Full-toolset registration and annotations added
- [x] Schemas, examples, catalog, and snapshots updated
- [x] R004 plan: create validation ordering, exact catalog/snapshot integration, and public echo contract recorded

---

### Step 3: Regression and safety coverage

**Status:** ✅ Complete

- [x] Exact create wire tests added
- [x] Invalid input avoids network calls
- [x] Credential/confirm/zone arguments excluded
- [x] Tool counts and catalog guards updated
- [x] R006: Exact local-server POST wire matrix covers Ride indoor FTP and Run/Swim canonical pace sparse bodies
- [x] R006: Pre-I/O malformed create arguments and Type/Types duplicate lookup behavior have distinct no-write coverage
- [x] R006: Raw and registered create schemas are closed against credential, confirm, recalc, and zone arguments
- [x] R006: Safety matrix and core/full/coach catalog count guards deliberately include create_sport_settings

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
| R002 | Plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | Plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |

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
| 2026-07-10 23:14 | Worker iter 2 | done in 2050s, tools: 81 |
| 2026-07-10 23:14 | Step 2 started | Add and register MCP surfaces |
| 2026-07-10 23:36 | Worker iter 3 | done in 1293s, tools: 57 |
| 2026-07-10 23:36 | Step 3 started | Regression and safety coverage |

## Blockers

- TP-228 and TP-229 must complete first.

## Notes

| Date | Topic | Detail |
|---|---|---|
| 2026-07-10 | R001 Step 1 plan | Add `IndoorFTP *int` only to `WriteSportSettingsParams`; introduce separate `CreateSportSettingsParams{Sport, FTP, IndoorFTP, ThresholdHR, ThresholdPace}` and return `SportSettings`. `SportSettingsPace` is pre-normalized m/s plus selected `PaceUnits` and explicit/preserved `PaceLoadType`, serialized without client reinterpretation. |
| 2026-07-10 | R001 HTTP contract | Update remains `PUT /athlete/{athleteID}/sport-settings/{id}?recalcHrZones=<bool>` and writes `indoor_ftp` only when supplied. Create is no-retry `POST /athlete/{athleteID}/sport-settings`, no query string, with sparse `types:[sport]` and only `ftp`, `indoor_ftp`, `lthr`, `threshold_pace` (m/s), `pace_units`, and `pace_load_type`; it cannot carry ID, recalculation, or zones. |
| 2026-07-10 | R001 validation | Before transport reject blank sport, non-positive FTP/indoor FTP/HR, and non-finite/non-positive canonical pace; errors name create/update and make no request. No `indoor_ftp <= ftp` restriction: no confirmed upstream/product rule. Client leaves sport enum ownership to the MCP layer. |
| 2026-07-10 | R001 client tests | Local-server cases will assert update-only `indoor_ftp` and create `types:["Ride"]`/indoor FTP method/path/raw-query/sparse body/no zone-or-recalc fields plus returned echo, m/s pace keys, and table-driven invalid update/create calls with zero requests. |
| 2026-07-10 | Step 2 plan | Extend update request, conversion params, response echo, metadata, input schema, description, and examples with sparse `indoor_ftp` only: `indoor_ftp` → `WriteSportSettingsParams.IndoorFTP`, `indoor_ftp_watts` from returned value with supplied fallback, and sorted `_meta.fields_updated`. |
| 2026-07-10 | R004 validation order | `create_sport_settings` strictly decodes then canonicalizes/validates the documented sport and requires one non-nil threshold field before any profile or create call. It then reads the profile once, uses `findSportSetting` across `Type`/`Types` to reject duplicates with “use update_sport_settings”, and only then calls create. Request/schema exclude recalc, zones, confirmation, and credentials. |
| 2026-07-10 | R004 pace and public echo | Creation converts duration to canonical m/s using an empty existing setting, selecting input display units and Run/Swim load type. `sport_settings` returns sport, created `sport_setting_id`, only applicable FTP/indoor-FTP/HR pointers, and existing m/s-derived pace echo plus units/load type. `_meta` returns operation `create`, sorted `fields_created`, server/timezone/unit data, and pace metadata only when relevant; no apply/recalc/zone metadata. |
| 2026-07-10 | R004 catalog and snapshot | Register `newCreateSportSettingsTool` as full/write in `registryBaseTools`; add `CreateSportSettings` to the athlete-scoped canonical catalog, full-tier map, and `toolCatalogGroup` settings case. Update schema registry count 69→70 and required `create_sport_settings` snapshot, generated by `go run ./scripts/snapshot_tool_schemas.go`. Focused tests cover creator FTP/indoor/HR plus Run/Swim pace paths. |
| 2026-07-10 | R006 Step 3 plan | Client local-server wire tests must reject all extra keys across Ride/Run/Swim. Separate malformed pre-I/O/no-profile validation from duplicate `Type`/`Types` lookup/no-write behavior. Assert raw and registered create schemas exclude credential/confirm/recalc/zones (not update), add create to safety v03 catalog with safe/full/default counts 60/68/46, and target `internal/safety`. |
| 2026-07-10 22:43 | Review R002 | plan Step 1: APPROVE |
| 2026-07-10 22:49 | Review R003 | code Step 1: APPROVE |
| 2026-07-10 22:53 | Review R004 | plan Step 2: REVISE |
| 2026-07-10 22:55 | Review R005 | plan Step 2: APPROVE |
| 2026-07-10 23:37 | Review R006 | plan Step 3: REVISE |
| 2026-07-10 23:38 | Review R006 | plan Step 3: REVISE |
| 2026-07-10 23:40 | Review R007 | plan Step 3: APPROVE |
| 2026-07-10 23:49 | Review R009 | code Step 3: APPROVE |
