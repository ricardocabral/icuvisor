# TP-011 — Status

**Issue:** v0.2 — read path
**Status:** ✅ Complete
**Review Level:** 2
**Iteration:** 3
**Current Step:** Step 6: Tests
**Last Updated:** 2026-05-11
**State:** Complete

_Task scaffolded from PROMPT.md; execution in progress._

## Step 1: Map the wellness payload

**Status:** ✅ Complete

- [x] Define the Step 1 mapping artifact schema in `STATUS.md`, including upstream key/type/nullability, category, planned response path, source/native scale, and evidence
- [x] Document clean-room evidence sources and GPL exclusion rules for the wellness mapping
- [x] Document black-box probe/fixture strategy, including target cases and sanitization rules
- [x] Document provider/provenance decision rules and custom-field handling rules
- [x] Catalog every field exposed by the intervals.icu wellness endpoint, distinguishing: athlete-entered scales, device-imported normalized fields, raw native sub-fields, and custom fields
- [x] Identify each bridged field's provider source (Polar / Garmin / Oura / Whoop / Apple Health / manual / unknown) and native scale
- [x] Record findings and any uncertainty in `STATUS.md`
- [x] R001 code review: add `feel` to the subjective-scale inventory or record it as an upstream OpenAPI gap
- [x] R001 code review: rewrite stale bridge fixture/provenance wording so `fetched_at` older than the wellness date/reference by >24h is stale and exactly 24h is not stale

## Step 2: Implement typed decoding

**Status:** ✅ Complete

- [x] Define the Step 2 typed-decoding plan, including client endpoint shape, raw/custom field preservation, and native sidecar boundaries
- [x] R001 plan review: specify response-row construction from raw/custom keys plus typed overlays before `response.Shape`
- [x] R001 plan review: specify native key hoisting, top-level native alias removal, and nested/top-level provider key support
- [x] Implement the intervals wellness client method and typed row decoder without dropping unknown/custom JSON keys
- [x] Add the `get_wellness_data` tool response shell so decoded rows can be shaped in later steps
- [x] Decode `sleepQuality` (1–4) and `sleepScore` (0–100) as **distinct fields** — no aliasing, no collapse
- [x] Decode `sleepSecs` as its own field; do not derive from `sleepScore`
- [x] Decode raw native sub-fields under a sidecar struct and surface under `_native.<source>.<field>`
- [x] Preserve custom-field rows (intervals.icu wellness custom fields) — they participate in null-stripping the same as standard fields

## Step 3: Provenance and staleness `_meta` assembly

**Status:** ✅ Complete

- [x] Define the Step 3 provenance/staleness plan, including source inference boundaries, timestamp selection, and 24h reference rule
- [x] R001 plan review: revise timestamp rules to exclude `updated` fallback and use non-null unknown `fetched_at`
- [x] R001 plan review: refine native-scale defaults, dual-use field provenance rules, stale-reason wording, and required shaper preservation of provenance `fetched_at`
- [x] For every bridged field, emit `_meta.provenance.<field> = { source, native_scale, fetched_at }`
- [x] Where provenance cannot be determined, emit `source: "unknown"` rather than omitting the marker
- [x] If `now - fetched_at > 24h` relative to the wellness `date`, emit `_meta.stale: true` with a one-line `_meta.stale_reason` (e.g. `"polar bridge refresh requires user to open intervals.icu"`)

## Step 4: In-response scale labels

**Status:** ✅ Complete

- [x] Define the Step 4 scale-label plan against `internal/response/shaper.go` and wellness row shaping
- [x] Register every subjective scale in the TP-007 scale-label registry: `feel` 1–5, `sleepQuality` 1–4, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, `injury`
- [x] Verify `_meta.scales` appears in the response for every registered field present in the row

## Step 5: Null-stripping integration

**Status:** ✅ Complete

- [x] Define the Step 5 null-stripping integration plan for wellness row collections, `_meta.missing_fields`, and `include_full`
- [x] Confirm wellness rows pass through the TP-007 null-strip pipeline
- [x] `_meta.missing_fields: [...]` per row when stripping removed at least one key
- [x] `include_full: true` opt-out is honoured

## Step 6: Tests

**Status:** ✅ Complete

- [x] Add sanitized wellness fixture files for Polar fresh/stale, Garmin body battery, Oura sleep score, manual-only, and custom/null-heavy rows
- [x] Table-driven tests over the `testdata/wellness/` fixtures
- [x] Assert: distinct sleep fields; provenance per bridged field; `_meta.stale` boundary at exactly 24h; `_native` round-trip for each provider in fixtures; null-strip + `_meta.missing_fields`; scale labels for every subjective field
- [x] `make test`, `make build`, `make lint` pass
- [x] Update `CHANGELOG.md` and README catalog for `get_wellness_data`

## Step 1 Plan and Findings — Wellness field catalog

### Mapping artifact schema

Each observed upstream wellness key is cataloged with the following columns:

| Column                    | Meaning                                                                                                                                                                     |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Upstream key/path         | Exact JSON key or nested path observed from intervals.icu wellness responses.                                                                                               |
| Observed type/nullability | JSON type(s) seen and whether null/missing is expected.                                                                                                                     |
| Category                  | One of: athlete-entered subjective scale, manual/body metric, normalized device-imported field, raw native sub-field, metadata/provenance marker, custom field, or unknown. |
| Planned response key/path | Response field planned for `get_wellness_data`, including `_native.<source>.<field>` for native provider payloads.                                                          |
| Provider source           | `manual`, `polar`, `garmin`, `oura`, `whoop`, `apple_health`, or `unknown`; never inferred without evidence.                                                                |
| Native scale              | Provider/native scale or unit when known; `unknown` when not supported by evidence.                                                                                         |
| Evidence                  | Public API docs, public forum reference, black-box fixture/probe, MIT endpoint-shape reference, or explicit uncertainty.                                                    |

### Clean-room evidence rules

Allowed evidence for this task:

- intervals.icu public API documentation and public intervals.icu forum posts.
- Black-box wellness endpoint responses collected from accounts with permission, with identifiers sanitized before committing fixtures.
- Synthetic fixtures derived from observed public field names and documented scales when live probing is unavailable.
- MIT-licensed `hhopke/intervals-icu-mcp` may be consulted only for endpoint shape, not copied into this project.

Excluded evidence:

- GPLv3 wellness handling from `mvilanova/intervals-mcp-server` must not be opened, read, paraphrased, or used.
- No copied code or payload fixtures from copyleft repositories.

### Black-box probe and fixture strategy

Endpoint to inspect: `GET /api/v1/athlete/{athlete_id}/wellness?oldest={YYYY-MM-DD}&newest={YYYY-MM-DD}` through the local typed client contract. If live credentials are unavailable in this environment, fixtures will be synthetic but constrained to public docs, public forum behavior, and documented acceptance criteria.

Target fixture/probe cases for later steps:

- Polar bridge fresh: normalized recovery/sleep fields plus `_native.polar.ans_charge` and `_native.polar.nightly_recharge_status`, with bridge `fetched_at` at the wellness row date/reference time or no more than 24h older.
- Polar bridge stale: same source markers with bridge `fetched_at` more than 24h older than the wellness row date/reference time used by Step 3 tests.
- Garmin body battery: normalized body-battery fields plus raw min/max values under `_native.garmin`.
- Oura raw sleep score: normalized `sleepScore` plus raw `_native.oura.sleep_score`.
- Manual-only row: athlete-entered subjective scales and body metrics without device provenance.
- Custom-fields row: user-defined keys not in the static catalog, including null-heavy keys to exercise null stripping and `_meta.missing_fields`.

Sanitization rules:

- Never commit API keys, bearer/basic credentials, real athlete IDs, names, emails, or unredacted account metadata.
- Fixture athlete IDs use stable fake values such as `i000001`; dates may be shifted while preserving relative staleness windows.
- Numeric wellness values may be rounded or synthesized when exact values could identify an athlete; scale ranges and null/missing shape must be preserved.

### Field catalog (Step 1)

| Upstream key/path                                       | Observed type/nullability                                                 | Category                                               | Planned response key/path                 | Provider source                              | Native scale                                                                                          | Evidence                                                        |
| ------------------------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------ | ----------------------------------------- | -------------------------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| id                                                      | string; present                                                           | metadata/provenance marker                             | date                                      | manual                                       | ISO local date id / n/a                                                                               | OpenAPI Wellness.id; list endpoint docs                         |
| ctl                                                     | number(float); nullable/missing                                           | training load context                                  | ctl                                       | manual                                       | fitness load units                                                                                    | OpenAPI                                                         |
| atl                                                     | number(float); nullable/missing                                           | training load context                                  | atl                                       | manual                                       | fitness load units                                                                                    | OpenAPI                                                         |
| rampRate                                                | number(float); nullable/missing                                           | training load context                                  | rampRate                                  | manual                                       | CTL ramp rate                                                                                         | OpenAPI                                                         |
| ctlLoad                                                 | number(float); nullable/missing                                           | training load context                                  | ctlLoad                                   | manual                                       | load units                                                                                            | OpenAPI                                                         |
| atlLoad                                                 | number(float); nullable/missing                                           | training load context                                  | atlLoad                                   | manual                                       | load units                                                                                            | OpenAPI                                                         |
| sportInfo                                               | array<SportInfo>; nullable/missing                                        | training load context                                  | sportInfo                                 | manual                                       | sport-specific load summary                                                                           | OpenAPI                                                         |
| updated                                                 | string(date-time); nullable/missing                                       | metadata/provenance marker                             | \_meta.provenance.\*.fetched_at candidate | unknown                                      | RFC3339 timestamp; row update time, not provider-specific                                             | OpenAPI; bridge-specific semantics uncertain                    |
| weight                                                  | number(float); nullable/missing                                           | manual/body metric                                     | weight                                    | manual                                       | kg per intervals.icu default; unit normalization out of Step 1                                        | OpenAPI/PRD                                                     |
| restingHR                                               | integer; nullable/missing                                                 | manual/body metric                                     | restingHR                                 | manual or bridged unknown                    | bpm                                                                                                   | OpenAPI/PRD; provider uncertain when imported                   |
| hrv                                                     | number(float); nullable/missing                                           | manual/body metric                                     | hrv                                       | manual or bridged unknown                    | ms/rMSSD likely; upstream docs do not state                                                           | OpenAPI/PRD; uncertainty                                        |
| hrvSDNN                                                 | number(float); nullable/missing                                           | manual/body metric                                     | hrvSDNN                                   | manual or bridged unknown                    | ms SDNN                                                                                               | OpenAPI                                                         |
| menstrualPhase                                          | string enum; nullable/missing                                             | manual/body metric                                     | menstrualPhase                            | manual                                       | PERIOD/FOLLICULAR/OVULATING/LUTEAL/NONE                                                               | OpenAPI/PRD                                                     |
| menstrualPhasePredicted                                 | string enum; nullable/missing                                             | manual/body metric                                     | menstrualPhasePredicted                   | manual/unknown                               | same enum; predicted                                                                                  | OpenAPI                                                         |
| kcalConsumed                                            | integer; nullable/missing                                                 | manual/body metric                                     | kcalConsumed                              | manual                                       | kcal                                                                                                  | OpenAPI                                                         |
| sleepSecs                                               | integer; nullable/missing                                                 | normalized device-imported field or manual/body metric | sleepSecs                                 | manual or bridged unknown                    | seconds slept                                                                                         | OpenAPI/PRD; distinct from sleep scores                         |
| sleepScore                                              | number(float); nullable/missing                                           | normalized device-imported field                       | sleepScore                                | garmin/oura/whoop/apple_health/polar/unknown | 0-100 device nightly score; Polar raw 1-100 has same name                                             | OpenAPI/PRD/forum #58                                           |
| sleepQuality                                            | integer; nullable/missing                                                 | athlete-entered subjective scale                       | sleepQuality                              | manual                                       | 1-4, 1=poor 4=great                                                                                   | OpenAPI/PRD                                                     |
| feel                                                    | integer; nullable/missing; not present in current OpenAPI Wellness schema | athlete-entered subjective scale                       | feel                                      | manual                                       | 1-5                                                                                                   | PRD/Roadmap/response registry requirement; OpenAPI upstream gap |
| avgSleepingHR                                           | number(float); nullable/missing                                           | normalized device-imported field                       | avgSleepingHR                             | bridged unknown                              | bpm                                                                                                   | OpenAPI; provider uncertain                                     |
| soreness                                                | integer; nullable/missing                                                 | athlete-entered subjective scale                       | soreness                                  | manual                                       | 1-5 assumed; verify                                                                                   | OpenAPI/PRD Step 4                                              |
| fatigue                                                 | integer; nullable/missing                                                 | athlete-entered subjective scale                       | fatigue                                   | manual                                       | 1-5                                                                                                   | OpenAPI/PRD                                                     |
| stress                                                  | integer; nullable/missing                                                 | athlete-entered subjective scale                       | stress                                    | manual                                       | 1-5 assumed; verify                                                                                   | OpenAPI/PRD Step 4                                              |
| mood                                                    | integer; nullable/missing                                                 | athlete-entered subjective scale                       | mood                                      | manual                                       | 1-5                                                                                                   | OpenAPI/PRD                                                     |
| motivation                                              | integer; nullable/missing                                                 | athlete-entered subjective scale                       | motivation                                | manual                                       | 1-5 assumed; verify                                                                                   | OpenAPI/PRD Step 4                                              |
| injury                                                  | integer; nullable/missing                                                 | athlete-entered subjective scale                       | injury                                    | manual                                       | 1-5 assumed; verify                                                                                   | OpenAPI/PRD Step 4                                              |
| spO2                                                    | number(float); nullable/missing                                           | manual/body metric                                     | spO2                                      | manual or bridged unknown                    | percent oxygen saturation                                                                             | OpenAPI/PRD                                                     |
| systolic                                                | integer; nullable/missing                                                 | manual/body metric                                     | systolic                                  | manual                                       | mmHg                                                                                                  | OpenAPI/PRD                                                     |
| diastolic                                               | integer; nullable/missing                                                 | manual/body metric                                     | diastolic                                 | manual                                       | mmHg                                                                                                  | OpenAPI/PRD                                                     |
| hydration                                               | integer; nullable/missing                                                 | manual/body metric                                     | hydration                                 | manual                                       | upstream scale unknown                                                                                | OpenAPI; uncertainty                                            |
| hydrationVolume                                         | number(float); nullable/missing                                           | manual/body metric                                     | hydrationVolume                           | manual                                       | volume; unit unknown                                                                                  | OpenAPI; uncertainty                                            |
| readiness                                               | number(float); nullable/missing                                           | normalized device-imported field                       | readiness                                 | polar/garmin/oura/whoop/apple_health/unknown | provider-normalized readiness; Polar maps nightly_recharge_status 1-6 without normalization per forum | OpenAPI/forum #56/#58                                           |
| baevskySI                                               | number(float); nullable/missing                                           | normalized device-imported field                       | baevskySI                                 | bridged unknown                              | Baevsky stress index                                                                                  | OpenAPI; source uncertain                                       |
| bloodGlucose                                            | number(float); nullable/missing                                           | manual/body metric                                     | bloodGlucose                              | manual                                       | glucose unit upstream/user dependent                                                                  | OpenAPI/PRD                                                     |
| lactate                                                 | number(float); nullable/missing                                           | manual/body metric                                     | lactate                                   | manual                                       | mmol/L likely; verify                                                                                 | OpenAPI/PRD                                                     |
| bodyFat                                                 | number(float); nullable/missing                                           | manual/body metric                                     | bodyFat                                   | manual                                       | percent                                                                                               | OpenAPI/PRD                                                     |
| abdomen                                                 | number(float); nullable/missing                                           | manual/body metric                                     | abdomen                                   | manual                                       | circumference unit upstream/user dependent                                                            | OpenAPI/PRD                                                     |
| vo2max                                                  | number(float); nullable/missing                                           | manual/body metric                                     | vo2max                                    | manual or bridged unknown                    | ml/kg/min                                                                                             | OpenAPI/PRD                                                     |
| comments                                                | string; nullable/missing                                                  | manual/body metric                                     | comments                                  | manual                                       | free text                                                                                             | OpenAPI                                                         |
| steps                                                   | integer; nullable/missing                                                 | normalized device-imported field or manual/body metric | steps                                     | manual or bridged unknown                    | count                                                                                                 | OpenAPI                                                         |
| respiration                                             | number(float); nullable/missing                                           | manual/body metric                                     | respiration                               | manual or bridged unknown                    | breaths/min likely; verify                                                                            | OpenAPI/PRD                                                     |
| carbohydrates                                           | number(float); nullable/missing                                           | manual/body metric                                     | carbohydrates                             | manual                                       | grams likely; verify                                                                                  | OpenAPI                                                         |
| protein                                                 | number(float); nullable/missing                                           | manual/body metric                                     | protein                                   | manual                                       | grams likely; verify                                                                                  | OpenAPI                                                         |
| fatTotal                                                | number(float); nullable/missing                                           | manual/body metric                                     | fatTotal                                  | manual                                       | grams likely; verify                                                                                  | OpenAPI                                                         |
| locked                                                  | boolean; nullable/missing                                                 | metadata/provenance marker                             | locked                                    | manual                                       | prevents device sync overwrite                                                                        | OpenAPI/PRD                                                     |
| tempWeight                                              | boolean; nullable/missing                                                 | metadata/provenance marker                             | tempWeight                                | manual                                       | temporary/generated weight marker                                                                     | OpenAPI; semantics uncertain                                    |
| tempRestingHR                                           | boolean; nullable/missing                                                 | metadata/provenance marker                             | tempRestingHR                             | manual                                       | temporary/generated resting HR marker                                                                 | OpenAPI; semantics uncertain                                    |
| custom top-level keys                                   | any JSON type; nullable/missing                                           | custom field                                           | same custom key                           | manual/unknown                               | user-defined                                                                                          | OpenAPI fields query + PRD custom wellness fields               |
| polar.ans_charge / ans_charge                           | number; nullable/missing                                                  | raw native sub-field                                   | \_native.polar.ans_charge                 | polar                                        | -10 to +10                                                                                            | Forum #58 / acceptance fixture requirement                      |
| polar.nightly_recharge_status / nightly_recharge_status | integer; nullable/missing                                                 | raw native sub-field                                   | \_native.polar.nightly_recharge_status    | polar                                        | 1-6                                                                                                   | Forum #58 / acceptance fixture requirement                      |
| polar.sleep_score / sleep_score                         | number; nullable/missing                                                  | raw native sub-field                                   | \_native.polar.sleep_score                | polar                                        | 1-100                                                                                                 | Forum #58                                                       |
| garmin.body_battery_min                                 | integer; nullable/missing                                                 | raw native sub-field                                   | \_native.garmin.body_battery_min          | garmin                                       | 0-100 body battery                                                                                    | Acceptance fixture requirement; field name to verify by probe   |
| garmin.body_battery_max                                 | integer; nullable/missing                                                 | raw native sub-field                                   | \_native.garmin.body_battery_max          | garmin                                       | 0-100 body battery                                                                                    | Acceptance fixture requirement; field name to verify by probe   |
| oura.sleep_score                                        | integer; nullable/missing                                                 | raw native sub-field                                   | \_native.oura.sleep_score                 | oura                                         | 0-100 Oura sleep score                                                                                | Acceptance fixture requirement; field name to verify by probe   |

### Provider/provenance decision rules

- Prefer explicit upstream provider/source markers when present. Normalize provider labels to `polar`, `garmin`, `oura`, `whoop`, `apple_health`, `manual`, or `unknown`.
- Treat athlete-entered subjective scales and body metrics as `source: manual` unless a bridge marker explicitly says otherwise.
- Treat normalized device-imported wellness fields as bridged fields only when provider evidence exists in the row, a native provider sub-field exists for the same concept, or a fixture/probe documents the bridge.
- Set `source: unknown` and `native_scale: unknown` when a non-manual value appears without enough evidence to assign a provider. Do not invent provenance from field names alone unless the field is explicitly provider-native (for example `polar.*`, `garmin.*`, or `oura.*`).
- Use upstream refresh/import timestamps as `fetched_at` when present. If absent, keep the provenance marker with an unknown/zero `fetched_at` representation decided in Step 3 tests rather than fabricating a timestamp.
- Staleness is evaluated as `reference_time - fetched_at`, where `reference_time` is the wellness row's date boundary chosen by the tool/test (or an explicit test `now` when provided). Step 3 marks stale only when that bridge age is greater than 24h; exactly 24h old is not stale, and a `fetched_at` after the row date/reference is treated as fresh/not stale.

### Custom field handling rules

- Unknown top-level fields not claimed by the static catalog or reserved metadata are preserved as custom fields.
- Custom fields keep their original JSON key spelling in the response unless a safety rule requires namespacing.
- Custom fields participate in null stripping and `_meta.missing_fields` exactly like standard fields.
- A custom field is never reclassified as a provider-native field without explicit source evidence.

### Bridged field source/native-scale matrix

| Normalized response field                                                            | Provider evidence accepted                                   | Provider source emitted             | Native scale emitted                                              | Notes                                                                                        |
| ------------------------------------------------------------------------------------ | ------------------------------------------------------------ | ----------------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `sleepScore`                                                                         | Oura native `sleep_score`                                    | `oura`                              | `0-100 Oura sleep score`                                          | Keep distinct from `sleepQuality`; native retained at `_native.oura.sleep_score`.            |
| `sleepScore`                                                                         | Polar native `sleep_score` or Polar bridge marker            | `polar`                             | `1-100 Polar sleep_score`                                         | Forum #58 says same name and not directly normalized against Garmin; retain native marker.   |
| `sleepScore`                                                                         | Garmin/Whoop/Apple Health bridge marker without native field | `garmin` / `whoop` / `apple_health` | `0-100 device nightly score` when documented; otherwise `unknown` | Do not infer provider without marker.                                                        |
| `sleepScore`                                                                         | No provider/native marker                                    | `unknown`                           | `0-100 device nightly score; provider unknown`                    | Provenance marker still emitted.                                                             |
| `sleepSecs`                                                                          | Sleep duration imported with a provider marker               | matching provider or `unknown`      | `seconds slept`                                                   | Field is independent; never derived from `sleepScore`.                                       |
| `readiness`                                                                          | Polar `nightly_recharge_status`                              | `polar`                             | `1-6 Polar nightly_recharge_status`                               | Forum #58 says Intervals maps this to readiness and does not normalize; raw native retained. |
| `readiness`                                                                          | Other provider readiness/recovery marker                     | matching provider or `unknown`      | provider-specific or `unknown`                                    | Do not compare scales without `_meta.provenance`.                                            |
| `avgSleepingHR`                                                                      | Sleep HR imported with provider/native marker                | matching provider or `unknown`      | `bpm`                                                             | Upstream schema lacks provider.                                                              |
| `restingHR`, `hrv`, `hrvSDNN`, `spO2`, `respiration`, `steps`, `vo2max`, `baevskySI` | Device/import marker in row or fixture                       | matching provider or `unknown`      | documented unit or `unknown`                                      | Otherwise treated as manual/body metric and no invented provider.                            |
| `_native.polar.ans_charge`                                                           | Native key `ans_charge` under Polar fixture/probe            | `polar`                             | `-10 to +10`                                                      | Raw only; can support readiness explanation.                                                 |
| `_native.polar.nightly_recharge_status`                                              | Native key `nightly_recharge_status`                         | `polar`                             | `1-6`                                                             | Raw only plus provenance for `readiness`.                                                    |
| `_native.garmin.body_battery_min/max`                                                | Native Garmin body-battery min/max keys                      | `garmin`                            | `0-100 body battery`                                              | Raw keys retained under `_native.garmin`.                                                    |
| `_native.oura.sleep_score`                                                           | Native Oura `sleep_score` key                                | `oura`                              | `0-100 Oura score`                                                | Raw key retained under `_native.oura`.                                                       |

### Uncertainties / upstream gaps

- Official OpenAPI documents the static `Wellness` schema and list endpoint parameters, but not provider-specific native sub-field shapes; native raw field names beyond Polar forum evidence are fixture/probe assumptions until black-box validated.
- Live authenticated probing was not available in this worker environment; test fixtures for Polar/Garmin/Oura will be synthetic and explicitly constrained by PRD acceptance criteria plus public forum evidence.
- `updated` is documented as a row timestamp but not as a per-provider bridge refresh timestamp. Step 3 should use an explicit bridge timestamp from fixture/probe data when available; otherwise emit provenance with `source: unknown`/unknown timestamp rather than fabricate freshness.
- `feel` is required by PRD/Roadmap/TP-007 response shaping and already exists in the scale-label registry, but the current public OpenAPI `Wellness` schema does not list it; keep support in the response registry and test if observed in fixtures. Subjective scales for `soreness`, `stress`, `motivation`, and `injury` are required by the task as scale labels, but public OpenAPI only gives integer types. Treat them as 1-5 for v0.2 and keep this evidence limitation visible in tests/docs.
- Several units (`hydration`, `hydrationVolume`, `lactate`, nutrition macros, circumference) are not specified in OpenAPI; do not add unit claims beyond what the response-shaping task requires.

### Fixture plan for later steps

- Create `internal/intervals/testdata/wellness/` fixtures for the six planned cases: Polar fresh, Polar stale, Garmin body battery, Oura sleep score, manual-only, and custom-fields/null-heavy row.
- Include explicit fake provider markers and bridge `fetched_at` timestamps in fixtures so Step 3 can test provenance and the 24h staleness boundary: one row exactly 24h old (not stale) and one row older than 24h (stale), never by putting `fetched_at` after the wellness date as the stale case.
- Include null values in every fixture family to exercise per-row `_meta.missing_fields` and `include_full: true` behavior.

## Step 5 Plan — Null-stripping integration

- Keep `get_wellness_data` using `encodeShaped(..., RowCollections: []string{"wellness"}, ...)` so `response.Shape` applies per-row null stripping to every wellness row rather than only to the wrapper.
- Because wellness rows are built from raw JSON before typed overlays, standard fields and custom wellness fields with `null` values should be stripped together and recorded in each row's `_meta.missing_fields`.
- Verify that `include_full:false` strips null standard/custom fields and emits per-row `_meta.fields_present`/`_meta.missing_fields`, while `include_full:true` keeps null fields and avoids missing-fields metadata; the raw upstream row remains available under `full` for debugging.
- If existing shaping behavior is insufficient, adjust row construction or `response.Shape` without changing the TP-007 contract that only JSON nulls are stripped.

## Step 4 Plan — In-response scale labels

- Update `internal/response/shaper.go` `defaultScaleLabels` to cover the full subjective wellness set required by TP-011: `feel`, `sleepQuality`, `fatigue`, `soreness`, `stress`, `mood`, `motivation`, and `injury`; keep existing `sleepScore` because it is a device score with its own 0-100 scale.
- Use concise labels that include the range and subjective meaning, with `sleepQuality` remaining 1-4 and all other listed subjective scales 1-5.
- Rely on existing recursive `addScaleMeta` / `collectScaleLabels` behavior so labels appear on each shaped wellness row when the field is present, while absent/null-stripped fields do not add scale noise.
- Verify with targeted response/tool tests that a wellness row containing every subjective field produces `_meta.scales` entries for each field.

## Step 3 Plan — Provenance and staleness `_meta`

- Add provenance assembly in the wellness tool row builder, not in the low-level decoder: decoded rows carry raw/native data; tool rows decide which present fields receive `_meta.provenance`.
- Emit provenance for always-device/normalized bridged fields present in a row: `sleepScore`, `sleepSecs`, `readiness`, and `avgSleepingHR`. Emit provenance for dual-use fields (`restingHR`, `hrv`, `hrvSDNN`, `spO2`, `respiration`, `steps`, `vo2max`, `baevskySI`) only when explicit provider/native evidence is present; manual-only rows should not imply bridge ambiguity for athlete-entered/body metric values.
- Infer source only from explicit provider/native evidence: `_native` contents, provider marker fields such as `source`, `provider`, `wellness_source`, `wellnessSource`, `integration`, or provider-specific top-level aliases already claimed by Step 2. If an always-device bridged field is present without provider evidence, emit `source: "unknown"` while using the field's known native scale/unit where available (`sleepScore`: `0-100 device nightly score`, `sleepSecs`: `seconds`, `avgSleepingHR`: `bpm`). Use `native_scale: "unknown"` only for truly unknown/provider-specific scales such as readiness without native evidence.
- Choose `fetched_at` only from explicit bridge/import/provider refresh timestamp fields in priority order: `bridge_fetched_at`, `bridgeFetchedAt`, `provider_fetched_at`, `providerFetchedAt`, `imported_at`, `importedAt`, `importedAtUtc`, `imported_at_utc`, `fetched_at`, and `fetchedAt`. Do not use `updated` as a bridge freshness fallback because Step 1 identified it as a row update timestamp that can change after manual edits.
- Always include the provenance key `fetched_at`. When no explicit bridge/import/provider timestamp is present, emit the string sentinel `"unknown"` so default null stripping cannot remove the key; unknown timestamps do not participate in staleness calculation.
- Compute stale status from the wellness row date at UTC midnight as `reference_time - fetched_at`; mark stale only when the age is strictly greater than 24h. Exactly 24h old is fresh; timestamps after the row date/reference are fresh.
- When any provenance timestamp is stale, emit `_meta.stale: true` and `_meta.stale_reason` as one deterministic line: for a known source, `"<source> bridge data is older than 24h for this wellness date"`; for unknown source, `"wellness bridge data is older than 24h for this wellness date"`. Do not emit stale fields for fresh rows or unknown timestamps.
- Update `response.dropDebugMetadata`/missing-field filtering so debug stripping is scoped to top-level/debug response fields and preserves `_meta.provenance.*.fetched_at` in default mode while still stripping debug-only top-level `fetched_at` and `query_type`.

## Step 2 Plan — Typed decoding

- Add `internal/intervals/wellness.go` with `WellnessParams` (`Oldest`, `Newest`, optional `Fields`) and `Client.ListWellness(ctx, params)` calling the public list endpoint path `athlete/{id}/wellness.json` with `oldest`, `newest`, and optional comma-joined `fields`.
- Decode `Wellness` rows with pointer fields for nullable upstream values so `sleepQuality`, `sleepScore`, and `sleepSecs` remain distinct and unset/null can be represented without zero-value ambiguity.
- Preserve the original upstream JSON object in `Wellness.Raw` via custom `UnmarshalJSON`, matching existing `Activity` and `SummaryWithCats` patterns; unknown top-level keys remain available for custom fields and null-stripping.
- Add sidecar extraction helpers that split recognized provider-native fields into a `Native` map by source (`polar`, `garmin`, `oura`) while keeping the raw object available for include-full/debug behavior.
- Recognize native fields in both nested provider objects (`polar.ans_charge`, `polar.nightly_recharge_status`, `polar.sleep_score`, `garmin.body_battery_min`, `garmin.body_battery_max`, `oura.sleep_score`) and top-level aliases from fixtures/probes (`ans_charge`, `nightly_recharge_status`, `sleep_score`, `body_battery_min`, `body_battery_max`) when a provider marker or fixture context supports the source.
- In terse shaped rows, claimed native aliases are removed from the top level and emitted only under `_native.<source>.<field>` to avoid duplicate scale-ambiguous fields; `include_full:true` may expose the untouched upstream raw payload under a dedicated `full` field. Unrecognized top-level keys remain custom fields and are not guessed as provider-native.
- Add `internal/tools/get_wellness_data.go` with request validation, registry wiring when the profile client implements `WellnessClient`, and a response shell returning decoded row maps through `response.Shape`; provenance/staleness and final scale/null-strip assertions are completed in Steps 3-5.
- Build each response row as a `map[string]any` by starting from `Wellness.Raw` so custom keys and upstream nulls are retained, overlaying canonical typed/static fields (`date` from `id`, distinct `sleepQuality`, `sleepScore`, `sleepSecs`, and other decoded fields), adding `_native` from the sidecar, and then passing the wrapper through `response.Shape` with `RowCollections: []string{"wellness"}` so standard and custom keys are null-stripped together per row.
- Keep custom fields at their original top-level spelling unless a recognized native extractor claims them; do not build the terse row by marshaling only the typed struct, because that would drop dynamic wellness custom fields.
- Keep all source/provenance inference conservative: the decoder records raw/native/custom data but does not fabricate provider provenance in Step 2.

## Discoveries

| Date       | Step   | Finding                                                                                                                                                                                                           |
| ---------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-05-11 | Step 1 | Defined STATUS.md wellness mapping artifact schema for upstream key/type/nullability, field category, response path, source/native scale, and evidence.                                                           |
| 2026-05-11 | Step 1 | Cataloged official OpenAPI Wellness fields, dynamic custom keys, and required native provider sub-fields for Polar, Garmin, and Oura.                                                                             |
| 2026-05-11 | Step 1 | Added `feel` to the subjective-scale inventory as a PRD/Roadmap requirement and documented that it is absent from the current public OpenAPI Wellness schema.                                                     |
| 2026-05-11 | Step 2 | Planned typed decoding around `wellness.json`, nullable pointer fields, raw/custom preservation, and provider-native sidecar extraction.                                                                          |
| 2026-05-11 | Step 2 | Refined Step 2 plan so response rows start from raw/custom keys with typed overlays, and recognized native aliases are hoisted to `_native` instead of duplicated at top level.                                   |
| 2026-05-11 | Step 2 | Added `internal/intervals/wellness.go` with `ListWellness`, nullable typed fields, raw JSON preservation, and native sidecar extraction.                                                                          |
| 2026-05-11 | Step 2 | Added `get_wellness_data` tool shell, registry wiring, request validation, and row map shaping via `response.Shape`.                                                                                              |
| 2026-05-11 | Step 2 | Decoded `sleepQuality` and `sleepScore` into separate nullable typed fields and overlaid them independently into response rows.                                                                                   |
| 2026-05-11 | Step 2 | Decoded `sleepSecs` as a separate nullable typed field with no derivation from `sleepScore`.                                                                                                                      |
| 2026-05-11 | Step 2 | Extracted recognized Polar, Garmin, and Oura native fields into `Wellness.Native` and surfaced them under `_native.<source>.<field>` in tool rows.                                                                |
| 2026-05-11 | Step 2 | Built wellness response rows from raw JSON first, so dynamic custom fields and upstream nulls remain available to the null-strip pipeline.                                                                        |
| 2026-05-11 | Step 3 | Planned provenance assembly in the wellness tool row builder with conservative source inference, explicit timestamp priority, and strict `>24h` stale boundary.                                                   |
| 2026-05-11 | Step 3 | Revised plan to exclude row `updated` from bridge freshness, use `fetched_at: "unknown"`, preserve known scales with unknown sources, avoid provenance on manual-only dual-use fields, and scope debug stripping. |
| 2026-05-11 | Step 3 | Added wellness `_meta.provenance` assembly for bridged wellness fields with source, native scale, and non-null `fetched_at`.                                                                                      |
| 2026-05-11 | Step 3 | Unknown provider cases now keep provenance entries with `source: "unknown"` and known field scales where available.                                                                                               |
| 2026-05-11 | Step 3 | Added strict `>24h` bridge staleness detection, deterministic stale reasons, and response shaper preservation for `_meta.provenance.*.fetched_at`.                                                                |
| 2026-05-11 | Step 4 | Planned scale-label registry expansion for the complete wellness subjective scale set and per-row `_meta.scales` verification.                                                                                    |
| 2026-05-11 | Step 4 | Expanded `defaultScaleLabels` to include `soreness`, `stress`, `motivation`, and `injury` alongside existing wellness scale labels.                                                                               |
| 2026-05-11 | Step 4 | Updated response shaper tests to verify `_meta.scales` entries for every registered wellness subjective field present in a row.                                                                                   |
| 2026-05-11 | Step 5 | Planned verification that `get_wellness_data` row collections use TP-007 null stripping for standard and custom fields, with `include_full` preserving nulls.                                                     |
| 2026-05-11 | Step 5 | Added `get_wellness_data` handler test proving wellness rows run through `response.Shape` row-collection null stripping.                                                                                          |
| 2026-05-11 | Step 5 | Verified per-row `_meta.missing_fields` includes stripped standard and custom null keys (`hrv`, `custom_null`).                                                                                                   |
| 2026-05-11 | Step 5 | Verified `include_full: true` keeps null standard/custom keys and exposes the raw upstream row under `full` without missing-fields metadata.                                                                      |
| 2026-05-11 | Step 6 | Added sanitized fixture JSON files for Polar fresh/stale, Garmin body battery, Oura sleep score, manual-only, and custom/null-heavy wellness rows.                                                                |
| 2026-05-11 | Step 6 | Added table-driven `get_wellness_data` tests over all `internal/intervals/testdata/wellness/` fixture rows.                                                                                                       |
| 2026-05-11 | Step 6 | Fixture tests assert sleep field separation, provenance, exactly-24h stale boundary, provider `_native` fields, null stripping/missing fields, and subjective scales.                                             |
| 2026-05-11 | Step 6 | `make test`, `make build`, and `make lint` passed.                                                                                                                                                                |
| 2026-05-11 | Step 6 | Updated README tool catalog and CHANGELOG Unreleased entry for `get_wellness_data`.                                                                                                                               |
| 2026-05-11 | Step 6 | Re-ran `make test`, `make build`, and `make lint` after final fixture/test/docs updates; all passed.                                                                                                              |

## Blockers

| Date | Step | Blocker |
| ---- | ---- | ------- |

## Notes

| Date | Step | Note |
| ---- | ---- | ---- |

## Execution Log

| Time             | Event                | Details                                                                                               |
| ---------------- | -------------------- | ----------------------------------------------------------------------------------------------------- |
| 2026-05-11 22:18 | Task started         | Runtime V2 lane-runner execution                                                                      |
| 2026-05-11 22:18 | Step 1 started       | Map the wellness payload                                                                              |
| 2026-05-11 22:19 | Hydrated             | Expanded STATUS.md with all PROMPT.md step checkboxes and review level.                               |
| 2026-05-11 22:20 | Plan review REVISE   | Added Step 1 planning outcomes requested by R001.                                                     |
| 2026-05-11 22:21 | Review R001          | plan Step 1: REVISE                                                                                   |
| 2026-05-11 22:24 | Review R001          | plan Step 1: APPROVE                                                                                  |
| 2026-05-11 22:31 | Review R001          | code Step 1: REVISE; added revision checkboxes.                                                       |
| 2026-05-11 22:36 | Review R001          | code Step 1: APPROVE; marked Step 1 complete.                                                         |
| 2026-05-11 22:36 | Step 2 started       | Implement typed decoding                                                                              |
| 2026-05-11 22:37 | Hydrated             | Added Step 2 typed decoding implementation outcomes before plan review.                               |
| 2026-05-11 22:40 | Review R001          | plan Step 2: REVISE; added response-row and native-hoisting plan items.                               |
| 2026-05-11 23:02 | Review R001          | plan Step 2: APPROVE                                                                                  |
| 2026-05-11 23:08 | Test                 | `go test ./internal/intervals ./internal/tools` passed for Step 2 implementation.                     |
| 2026-05-11 23:09 | Review R001          | code Step 2: APPROVE; marked Step 2 complete.                                                         |
| 2026-05-11 23:09 | Step 3 started       | Provenance and staleness `_meta` assembly                                                             |
| 2026-05-11 23:10 | Hydrated             | Added Step 3 provenance/staleness planning outcome before implementation.                             |
| 2026-05-11 23:14 | Review R001          | plan Step 3: REVISE; added timestamp/native-scale/stale-reason/shaper revision items.                 |
| 2026-05-11 23:20 | Review R001          | plan Step 3: APPROVE                                                                                  |
| 2026-05-11 23:32 | Test                 | `go test ./internal/intervals ./internal/tools ./internal/response` passed for Step 3 implementation. |
| 2026-05-11 23:33 | Review R001          | code Step 3: UNAVAILABLE once, retry APPROVE; marked Step 3 complete.                                 |
| 2026-05-11 23:33 | Step 4 started       | In-response scale labels                                                                              |
| 2026-05-11 23:34 | Hydrated             | Added Step 4 scale-label planning outcome before implementation.                                      |
| 2026-05-11 23:35 | Review R001          | plan Step 4: APPROVE                                                                                  |
| 2026-05-11 23:40 | Test                 | `go test ./internal/response ./internal/tools` passed for Step 4 implementation.                      |
| 2026-05-11 23:41 | Review R001          | code Step 4: APPROVE; marked Step 4 complete.                                                         |
| 2026-05-11 23:41 | Step 5 started       | Null-stripping integration                                                                            |
| 2026-05-11 23:42 | Hydrated             | Added Step 5 null-stripping planning outcome before implementation.                                   |
| 2026-05-11 23:43 | Review R001          | plan Step 5: APPROVE                                                                                  |
| 2026-05-11 23:49 | Test                 | `go test ./internal/tools ./internal/response` passed for Step 5 implementation.                      |
| 2026-05-11 23:50 | Review R001          | code Step 5: APPROVE; marked Step 5 complete.                                                         |
| 2026-05-11 23:50 | Step 6 started       | Tests                                                                                                 |
| 2026-05-11 23:51 | Hydrated             | Added Step 6 fixture and documentation outcomes.                                                      |
| 2026-05-11 23:59 | Step 6 complete      | Verification, fixtures, docs, build, test, and lint completed.                                        |
| 2026-05-11 23:59 | Task complete        | All TP-011 steps complete.                                                                            |
| 2026-05-11 22:30 | Review R001          | code Step 1: REVISE                                                                                   |
| 2026-05-11 22:33 | Review R001          | code Step 1: APPROVE                                                                                  |
| 2026-05-11 22:37 | Review R001          | plan Step 2: REVISE                                                                                   |
| 2026-05-11 22:39 | Review R001          | plan Step 2: APPROVE                                                                                  |
| 2026-05-11 22:46 | Review R001          | code Step 2: APPROVE                                                                                  |
| 2026-05-11 22:49 | Review R001          | plan Step 3: REVISE                                                                                   |
| 2026-05-11 22:52 | Review R001          | plan Step 3: APPROVE                                                                                  |
| 2026-05-11 23:14 | Review R001          | code Step 3: APPROVE                                                                                  |
| 2026-05-11 23:17 | Review R001          | plan Step 4: APPROVE                                                                                  |
| 2026-05-11 23:20 | Review R001          | code Step 4: APPROVE                                                                                  |
| 2026-05-11 23:22 | Review R001          | plan Step 5: APPROVE                                                                                  |
| 2026-05-11 23:26 | Review R001          | code Step 5: APPROVE                                                                                  |
| 2026-05-11 23:31 | Exit intercept close | Supervisor directed session close: "close"                                                            |
| 2026-05-11 23:31 | Worker iter 1        | done in 4386s, tools: 218                                                                             |
| 2026-05-11 23:31 | No progress          | Iteration 1: 0 new checkboxes (1/3 stall limit)                                                       |
| 2026-05-11 23:31 | Step 1 started       | Map the wellness payload                                                                              |
| 2026-05-11 23:33 | Exit intercept close | Supervisor directed session close: "close" |
| 2026-05-11 23:33 | Worker iter 2 | done in 86s, tools: 11 |
| 2026-05-11 23:33 | No progress | Iteration 2: 0 new checkboxes (2/3 stall limit) |
| 2026-05-11 23:33 | Step 1 started | Map the wellness payload |
