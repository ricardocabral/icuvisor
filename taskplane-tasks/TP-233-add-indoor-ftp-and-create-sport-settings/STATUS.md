# TP-233: Support indoor FTP and missing sport-setting creation — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 17
**Iteration:** 6
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

**Status:** ✅ Complete

- [x] PRD catalog updated
- [x] Website tool/schema data regenerated
- [x] Changelog updated
- [x] R010: Generator golden fixtures and website data reflect the 70-tool catalog and public create/update schemas
- [x] R010: PRD makes threshold-only create and indoor FTP write boundaries explicit
- [x] R010: Unreleased Added entry covers both capabilities and no-zone-replacement limit
- [x] R010: Docs generation and cmd/gendocs freshness tests pass with a reviewed diff
- [x] R011: PRD current-catalog counts are 70 total, 30 core, and 40 additional full tools

---

### Step 5: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Race suite passing
- [x] Lint passing
- [x] Build passes
- [x] Generated docs clean
- [x] R014: Docs regeneration leaves no tracked output changes and passes whitespace validation

---

### Step 6: Documentation & Delivery

**Status:** ✅ Complete

- [x] Must Update docs modified
- [x] Check If Affected docs reviewed
- [x] Discoveries logged

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
| Generated website data already exposes `create_sport_settings` and both create/update `indoor_ftp` arguments; README does not enumerate tool counts or sport-setting writes. | Kept generated data unchanged; no README change needed after review. | `web/data/tools.json`, `web/data/tool_schemas.json`, `README.md` |
| The FTP cookbook only covered outdoor FTP/zone updates, so a safe indoor-FTP variation makes the new sparse update discoverable without implying zone replacement. | Added an explicit-approval indoor-only example that preserves outdoor FTP and zones. | `web/content/cookbook/ftp-and-zones.md` |

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
| 2026-07-11 01:09 | Worker iter 4 | done in 5592s, tools: 146 |
| 2026-07-11 01:09 | Step 5 started | Testing & Verification |
| 2026-07-11 01:25 | Exit intercept reprompt | Supervisor provided instructions (261 chars) — reprompting worker |
| 2026-07-11 02:01 | Worker iter 5 | done in 3120s, tools: 20 |
| 2026-07-11 02:01 | Step 6 started | Documentation & Delivery |
| 2026-07-11 02:03 | Worker iter 6 | done in 158s, tools: 21 |
| 2026-07-11 02:03 | Task complete | .DONE created |

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
| 2026-07-10 | R010 Step 4 plan | Regenerate website plus `cmd/gendocs/testdata` goldens to 70 tools; inspect settings/full/write create and update indoor FTP schema. PRD/changelog must state threshold-only missing-sport semantics, one threshold requirement, no zones/recalc/history application, and no FTP ordering invention. Run docs generation and uncached gendocs/catalog/toolcheck tests. |
| 2026-07-10 | R011 Step 4 plan | Update every PRD statement of the current generated catalog to 70 total tools, 30 core tools, and 40 additional full tools, alongside the behavior contract and generated output refresh. |
| 2026-07-10 | R014 Step 5 plan | After docs regeneration, verify both `git diff --check` and no tracked changes relative to the pre-regeneration baseline; investigate and commit any intended output before rerunning verification. |
| 2026-07-10 22:43 | Review R002 | plan Step 1: APPROVE |
| 2026-07-10 22:49 | Review R003 | code Step 1: APPROVE |
| 2026-07-10 22:53 | Review R004 | plan Step 2: REVISE |
| 2026-07-10 22:55 | Review R005 | plan Step 2: APPROVE |
| 2026-07-10 23:37 | Review R006 | plan Step 3: REVISE |
| 2026-07-10 23:38 | Review R006 | plan Step 3: REVISE |
| 2026-07-10 23:40 | Review R007 | plan Step 3: APPROVE |
| 2026-07-10 23:47 | Review R010 | plan Step 4: REVISE |
| 2026-07-10 23:48 | Review R011 | plan Step 4: REVISE |
| 2026-07-10 23:51 | Review R014 | plan Step 5: REVISE |
| 2026-07-10 23:49 | Review R009 | code Step 3: APPROVE |
| 2026-07-10 23:51 | Review R010 | plan Step 4: REVISE |
| 2026-07-10 23:54 | Review R011 | plan Step 4: REVISE |
| 2026-07-10 23:55 | Review R012 | plan Step 4: APPROVE |
| 2026-07-11 00:00 | Review R013 | code Step 4: APPROVE |
| 2026-07-11 00:03 | Review R014 | plan Step 5: REVISE |
| 2026-07-11 00:19 | Review R016 | plan Step 5: APPROVE |
