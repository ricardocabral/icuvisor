# TP — Status

**Issue:** v0.2 — read path
**Status:** ✅ Complete
**Iteration:** 3
**Current Step:** Step 4: Tests
**Last Updated:** 2026-05-11
**State:** Complete

_Task scaffolded from PROMPT.md; hydrated with executable step checklists._

## Step 1: Black-box probe extended-metric availability (§7.4 #4)

**Status:** ✅ Complete

- [x] Define the PRD candidate metric inventory and black-box evidence standard, including `available` vs `computed` vs `not_observed` rules
- [x] Map each candidate field in PRD §7.2.C (`get_extended_metrics`) across documented public endpoints/query variants to whether intervals.icu exposes it directly
- [x] Capture sanitized minimal fixtures for every available field and record checked endpoints for unavailable/not-observed fields
- [x] Record findings in `testdata/extended-metrics/availability.md`; cite the endpoint, JSON pointer, unit/scale, and fixture for each "yes"
- [x] Fields with no upstream exposure are **dropped from the tool**, not zero-filled
- [x] Fix R001 code-review fixture evidence mismatches so every cited JSON pointer is present or explicitly narrowed

### Step 1 Probe Plan

**Candidate inventory and proposed response keys**

| PRD candidate metric        | Proposed `get_extended_metrics` key   |
| --------------------------- | ------------------------------------- |
| Ground contact time         | `ground_contact_time_ms`              |
| Vertical oscillation        | `vertical_oscillation_cm`             |
| Stride length               | `stride_length_m`                     |
| Ground contact time balance | `ground_contact_time_balance_percent` |
| DFA α1                      | `dfa_alpha1`                          |
| W' balance                  | `w_prime_balance_kj`                  |
| Core temperature            | `core_temperature_c`                  |
| Cardiac decoupling / Pw:HR  | `cardiac_decoupling_percent`          |
| HR drift %                  | `hr_drift_percent`                    |
| Aerobic decoupling          | `aerobic_decoupling_percent`          |
| Power-zone distribution     | `power_zone_distribution`             |
| Pace-zone time              | `pace_zone_time`                      |
| Cadence-by-zone             | `cadence_by_zone`                     |
| Joules above FTP            | `joules_above_ftp_kj`                 |
| Intensity factor            | `intensity_factor`                    |
| Variability index           | `variability_index`                   |
| Polarization index          | `polarization_index`                  |
| TRIMP                       | `trimp`                               |
| Strain score                | `strain_score`                        |
| HR load                     | `hr_load`                             |
| Pace load                   | `pace_load`                           |
| Power load                  | `power_load`                          |
| Left/right balance          | `left_right_balance_percent`          |
| RPE                         | `rpe`                                 |
| Feel                        | `feel`                                |
| Session RPE                 | `session_rpe`                         |
| Compliance %                | `compliance_percent`                  |
| Device name                 | `device_name`                         |

**Endpoint/query matrix**

| Candidate group                                                                                                     | Endpoints and variants to probe                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Running dynamics, L/R balance, device name                                                                          | `GET /api/v1/athlete/{id}/activities?oldest={date}&newest={date}` list payload; `GET /api/v1/activity/{activityId}` detail payload; documented `fields=`/include variants when present; `GET /api/v1/activity/{activityId}/streams` stream metadata and stream keys |
| DFA α1, W' balance, core temperature                                                                                | Activity list/detail payloads; activity streams/metadata; any documented analyzed-interval/lap endpoint that exposes server-computed or device-supplied advanced sensor metrics                                                                                     |
| Decoupling, HR drift, aerobic decoupling, IF, VI, TRIMP, strain/load variants, joules above FTP, zone distributions | Activity list/detail payloads; fitness/summary endpoints; analyzed intervals/laps; power-curve/best-effort endpoints only if their schemas expose these metrics directly                                                                                            |
| RPE, feel, session-RPE                                                                                              | Activity list/detail subjective fields; wellness/daily fields for same-date subjective scales; activity messages/notes only as evidence of absence, not as structured metric sources                                                                             |
| Compliance %                                                                                                        | Calendar event/activity pairing fields from `GET /api/v1/athlete/{id}/events{format}` and activity detail/list fields related to planned/completed pairing                                                                                                          |

**Availability states and decision rules**

- `yes`: the public API directly returns the metric as a documented or observed field; availability table must cite endpoint/query, JSON pointer, unit/scale, fixture, and probe date.
- `conditional`: the API directly returns the metric only for qualifying sports/devices/activities; table must state the condition and include a fixture.
- `computed_not_allowed`: the metric can be derived from streams/zones locally but is not directly returned by intervals.icu; do not expose it in `get_extended_metrics`.
- `not_observed`: docs or sample strategy indicate the metric may exist, but no representative sample is available; do not expose it until evidence is added.
- `no`: documented and available samples/endpoints were checked and no direct upstream field exists; do not expose it.

**Fixture capture/redaction**

- Store sanitized minimal JSON under `testdata/extended-metrics/` using names like `activity-detail-run-dynamics.json`, `activity-detail-ride-balance.json`, `activity-streams-advanced.json`, `wellness-subjective.json`, and `events-compliance.json`.
- Fixtures must contain only fields needed to prove availability/absence decisions; scrub API keys, raw athlete IDs, exact activity IDs when possible, notes, GPS/location traces, device serials, and personally identifying free text.
- Raw unredacted probe output must not be committed. If no local credentials/config are present, rely on public OpenAPI schema evidence and mark live probes skipped/limited in `availability.md`.

**Representative sample plan**

- Run with running dynamics-capable device.
- Ride with power meter and left/right balance.
- Activity or wellness day with subjective RPE/feel/session-RPE.
- Planned event paired to a completed activity for compliance.
- Samples capable of core temperature, DFA α1, and W' balance if available; otherwise record `not_observed` with limitations.

## Step 2: Implement the four straightforward reads

**Status:** ✅ Complete

- [x] `get_fitness`: date-range CTL / ATL / TSB; honour TZ; terse-by-default
- [x] `get_best_efforts`: PRs across sports; structure by sport + duration buckets
- [x] `get_power_curves`: mean-maximal curve; date-range arg; `include_full` for raw sample arrays
- [x] `get_training_summary`: aggregated volume / TSS / zones over a date range

### Step 2 Implementation Plan

| Tool                   | Client method and endpoint                                                                                                                                                                                                                                        | Request shape                                                                                                                                                                                                                                 | Terse response shape                                                                                                                                                                                              | Full-payload rule                                                                                                              |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `get_fitness`          | `ListFitness(ctx, FitnessParams)` against `GET /api/v1/athlete/{id}/athlete-summary.json?start&end` using `SummaryWithCats.fitness`, `fatigue`, and `form` as CTL/ATL/TSB                                                                                         | `start_date`, `end_date`, optional `include_full`                                                                                                                                                                                             | rows sorted by local date with `date`, `ctl`, `atl`, `tsb`; `_meta.server_version`, `_meta.units`, date range, count                                                                                              | Include raw `SummaryWithCats` rows only when `include_full:true`                                                               |
| `get_best_efforts`     | `ListBestEfforts(ctx, BestEffortsParams)` fans out per sport/type across upstream-computed athlete curve endpoints: `GET /api/v1/athlete/{id}/power-curves.json`, `hr-curves.json`, and `pace-curves.json`                                                        | optional `oldest`/`newest` (when omitted use all-time upstream curve spec), optional `sports` defaulting to `Ride`, `Run`, and `Swim`, optional `duration_seconds` for power/HR, optional `distance_meters` for pace, optional `include_full` | grouped by sport and bucket with best power watts, HR bpm, and pace values plus activity IDs where upstream provides them; `_meta.sports_requested` and unavailable sport notes make the default fan-out explicit | Include upstream curve arrays and activity map only when `include_full:true`                                                   |
| `get_power_curves`     | `ListPowerCurves(ctx, PowerCurvesParams)` against upstream-computed `GET /api/v1/athlete/{id}/power-curves.json`; translate `oldest`/`newest` to the documented range curve selector (`r.YYYY-MM-DD.YYYY-MM-DD`) rather than recomputing from per-activity curves | `oldest`, `newest`, optional `sport` defaulting to `Ride` because upstream `type` is required, optional `duration_seconds`, optional `include_full`                                                                                           | mean-maximal curve points `{duration_seconds, watts}` from `DataCurve.secs`/`values` plus `_meta.sport`; raw arrays omitted by default                                                                            | Include raw `DataCurveSetPowerCurve` list, activity map, `activity_id`, ranks, and submax arrays only when `include_full:true` |
| `get_training_summary` | `ListTrainingSummary(ctx, TrainingSummaryParams)` against `GET /api/v1/athlete/{id}/athlete-summary.json?start&end`                                                                                                                                               | `start_date`, `end_date`, optional `include_full`                                                                                                                                                                                             | totals for time, moving time, distance, calories, neutral `training_load`, sRPE, elevation, upstream `timeInZones` zone-order seconds, and per-sport category rollups                                             | Include raw summary rows only when `include_full:true`                                                                         |

Implementation notes: validate date arguments as `YYYY-MM-DD` before calling upstream; `get_best_efforts` may omit both dates for all-time curves but rejects a single unpaired `oldest` or `newest`; render/output dates in athlete timezone by using local date strings and profile timezone metadata; keep all four responses null-stripped by `response.Shape` with `RowCollections` where rows are returned and include `_meta.server_version` plus `_meta.units`; fetch athlete profile for timezone/unit preference; decode arguments with `DisallowUnknownFields()` and trailing-token rejection; register tools only when the injected client implements the new interfaces.

Curve-backed terse bucket contract: default power/HR duration buckets are `5,15,30,60,300,1200,3600` seconds; default run pace distance buckets are `400,1000,1609,5000,10000` meters; default swim pace distance buckets are `50,100,200,400,1500` meters. `get_power_curves` returns only requested/default duration bucket points unless `include_full:true`. `get_best_efforts` returns only requested/default duration and distance buckets per sport/family. Missing buckets are omitted from rows and listed in `_meta.missing_buckets`; requested/default buckets are echoed in `_meta.duration_seconds` and `_meta.distance_meters`.

## Step 3: Implement `get_extended_metrics`

**Status:** ✅ Complete

- [x] Expose only the fields confirmed in Step 1
- [x] Use the canonical unit enum (TP-008) and the shaping pipeline (TP-007)
- [x] Heavy payloads (raw stream-derived series) gated behind `include_full`
- [x] Return standard Strava-unavailable marker for blocked activity stubs instead of empty metrics
- [x] Preserve scale metadata for nested `rpe`, `feel`, and `session_rpe` extended metrics

### Step 3 Implementation Plan

- Client contract: define `ExtendedMetricsClient` covering `GetActivity`, `GetActivityIntervals`, and new `GetActivityPowerVsHR(ctx, activityID)` for `GET /api/v1/activity/{id}/power-vs-hr.json`; do not call streams for unavailable/computed fields. Register `get_extended_metrics` only when the injected client implements this full interface.
- Tool request: `activity_id` required, `include_full` optional. No date-range or athlete-derived metric computation. Schema descriptions must state terse default and raw upstream payloads only with `include_full:true`.
- Terse response schema: `{ activity_id, metrics, intervals?, full?, _meta }`. `metrics` contains only Step 1 available/conditional activity-level fields: `stride_length_m`, `cardiac_decoupling_percent`, `pw_hr`, `aerobic_decoupling_percent`, `power_zone_distribution_seconds`, `pace_zone_time_seconds`, `joules_above_ftp_kj`, `intensity_factor`, `variability_index`, `polarization_index`, `trimp`, `strain_score`, `hr_load`, `pace_load`, `power_load`, `training_load`, `left_right_balance_percent`, `rpe`, `feel`, `session_rpe`, `compliance_percent`, and `device_name`.
- Interval response schema: `intervals` rows are emitted only when at least one exposed interval metric is present. Rows include `interval_id`, `label`, `dfa_alpha1`, `w_prime_balance_start_kj`, `w_prime_balance_end_kj`, `joules_above_ftp_kj`, `aerobic_decoupling_percent`, `left_right_balance_percent`, `stride_length_m`, `strain_score`, and `training_load` as available; omit empty/mostly-null rows.
- Dropped fields: ground contact time, vertical oscillation, GCT balance, core temperature, HR drift %, and cadence-by-zone are never emitted or zero-filled; terse mode emits only `device_name` and keeps power-meter serial-like fields inside explicit `full` raw payloads.
- Unit/scales/profile: fetch athlete profile with the existing `toolProfile` pattern, pass `UnitSystem` to `response.Shape`, keep stable response field names, and set `_meta.extended_metric_units` using canonical enum values where available. Convert upstream joules and W' balance from J to kJ before emitting `_kj` fields; fixtures are treated as upstream J values and tests must assert conversion.
- Partial behavior: `GetActivity` failure fails the tool. `GetActivityIntervals` and `GetActivityPowerVsHR` `ErrNotFound`/`ErrUnauthorized` are tolerated as optional sources; response includes `_meta.partial:true` and `_meta.unavailable_sources` with short source names. Other errors fail with a short user-facing message.
- Full payloads: when `include_full:true`, include raw activity, interval DTO, and power-vs-HR payloads; otherwise omit raw arrays/maps.

## Step 4: Tests

**Status:** ✅ Complete

- [x] Table-driven tests using `httptest.Server` + fixtures; never hit the network
- [x] Cover: TZ-correct date math on fitness rows; sport-buckets on best-efforts; curve-shape correctness on power curves; field-drop behavior on extended metrics for fixtures that omit a tracked field
- [x] Update README catalog and CHANGELOG for the five new tools
- [x] `make test`, `make build`, `make lint` pass

## Discoveries

| Date       | Area             | Finding                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ---------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-05-11 | Extended metrics | Public OpenAPI documents direct fields for stride, interval DFA α1/W' balance, decoupling/Pw:HR, zone times, joules above FTP, IF/VI/polarization, TRIMP/strain/load, L/R balance, RPE/feel/session-RPE, compliance, and device name. Ground contact time, vertical oscillation, GCT balance, core temperature, HR drift %, and cadence-by-zone are not directly exposed and must be dropped rather than zero-filled or derived. |

## Blockers

None.

## Execution Log

| Time             | Event                         | Notes                                                                                                                                                                                      |
| ---------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 2026-05-11 21:04 | Task started                  | Runtime V2 lane-runner execution                                                                                                                                                           |
| 2026-05-11 21:04 | Step 1 started                | Black-box probe extended-metric availability (§7.4 #4)                                                                                                                                     |
| 2026-05-11 21:05 | Hydrated STATUS.md            | Added step checklists from PROMPT.md for crash-safe execution                                                                                                                              |
| 2026-05-11 21:06 | Plan review R001 requested    | Reviewer requested an explicit probe matrix, candidate inventory, evidence standard, fixture redaction process, and representative sample strategy                                         |
| 2026-05-11 21:07 | Step 1 plan revised           | Added candidate inventory, endpoint/query matrix, availability semantics, fixture redaction process, and representative sample strategy                                                    |
| 2026-05-11 21:14 | Code review R001 requested    | Reviewer found fixture evidence mismatches for cited JSON pointers in availability.md                                                                                                      |
| 2026-05-11 21:16 | Step 1 code review approved   | R001 returned APPROVE after fixture evidence fixes                                                                                                                                         |
| 2026-05-11 21:17 | Step 2 started                | Hydrated concrete endpoint/client/response plan for the four straightforward read tools                                                                                                    |
| 2026-05-11 21:18 | Plan review R001 requested    | Reviewer required athlete power-curves endpoint for get_power_curves, explicit best-effort bucket/default semantics, neutral training-load naming, and existing response decoding patterns |
| 2026-05-11 21:19 | Step 2 plan revised           | Set get_power_curves default sport to Ride and get_best_efforts default sport fan-out to Ride, Run, and Swim with explicit meta                                                            |
| 2026-05-11 21:20 | Step 2 plan revised           | Added terse default duration/distance bucket contract, missing-bucket metadata, and paired-date validation for curve-backed tools                                                          |
| 2026-05-11 21:24 | Step 2 implementation drafted | Added typed athlete-summary and curve clients plus four registered MCP tools; targeted package tests pass                                                                                  |
| 2026-05-11 21:25 | Step 2 code review approved   | R001 returned APPROVE                                                                                                                                                                      |
| 2026-05-11 21:26 | Step 3 started                | Hydrated get_extended_metrics plan from Step 1 availability evidence                                                                                                                       |
| 2026-05-11 21:27 | Plan review R001 requested    | Reviewer required exact response schema, J to kJ conversion, profile-derived unit metadata, interface-gated registration, and optional-source error behavior                               |
| 2026-05-11 21:31 | Step 3 implementation drafted | Added get_extended_metrics, power-vs-HR client method, registry wiring, J-to-kJ shaping, and optional-source partial metadata; targeted package tests pass                                 |
| 2026-05-11 21:32 | Code review R001 requested    | Reviewer required Strava-unavailable response handling and explicit scale metadata for nested subjective fields                                                                            |
| 2026-05-11 21:34 | Step 3 revisions applied      | Added Strava-unavailable response and recursive scale metadata detection; targeted packages pass                                                                                           |
| 2026-05-11 21:35 | Step 3 code review approved   | R001 returned APPROVE after revisions                                                                                                                                                      |
| 2026-05-11 21:36 | Step 4 started                | Added docs-update checkpoint alongside required test and full-suite verification items                                                                                                     |
| 2026-05-11 21:43 | Step 4 verification passed    | Added client/tool tests and docs; `make test`, `make build`, and `make lint` pass                                                                                                          |
| 2026-05-11 21:44 | Task complete                 | All steps complete and verification passed                                                                                                                                                 |
| 2026-05-11 21:07 | Review R001                   | plan Step 1: UNKNOWN                                                                                                                                                                       |
| 2026-05-11 21:09 | Review R001                   | plan Step 1: UNKNOWN                                                                                                                                                                       |
| 2026-05-11 21:12 | Review R001                   | plan Step 1: APPROVE                                                                                                                                                                       |
| 2026-05-11 21:18 | Review R001                   | code Step 1: UNKNOWN                                                                                                                                                                       |
| 2026-05-11 21:21 | Review R001                   | code Step 1: APPROVE                                                                                                                                                                       |
| 2026-05-11 21:25 | Review R001                   | plan Step 2: REVISE                                                                                                                                                                        |
| 2026-05-11 21:28 | Review R001                   | plan Step 2: REVISE                                                                                                                                                                        |
| 2026-05-11 21:31 | Review R001                   | plan Step 2: REVISE                                                                                                                                                                        |
| 2026-05-11 21:33 | Review R001                   | plan Step 2: APPROVE                                                                                                                                                                       |
| 2026-05-11 21:40 | Review R001                   | code Step 2: APPROVE                                                                                                                                                                       |
| 2026-05-11 21:43 | Review R001                   | plan Step 3: REVISE                                                                                                                                                                        |
| 2026-05-11 21:45 | Review R001                   | plan Step 3: APPROVE                                                                                                                                                                       |
| 2026-05-11 21:52 | Review R001                   | code Step 3: REVISE                                                                                                                                                                        |
| 2026-05-11 21:57 | Review R001                   | code Step 3: APPROVE                                                                                                                                                                       |
| 2026-05-11 22:04 | Exit intercept close          | Supervisor directed session close: "close"                                                                                                                                                 |
| 2026-05-11 22:04 | Worker iter 1                 | done in 3548s, tools: 195                                                                                                                                                                  |
| 2026-05-11 22:04 | No progress                   | Iteration 1: 0 new checkboxes (1/3 stall limit)                                                                                                                                            |
| 2026-05-11 22:04 | Step 1 started                | Black-box probe extended-metric availability (§7.4 #4)                                                                                                                                     |
| 2026-05-11 22:05 | Exit intercept close | Supervisor directed session close: "close" |
| 2026-05-11 22:05 | Worker iter 2 | done in 104s, tools: 10 |
| 2026-05-11 22:05 | No progress | Iteration 2: 0 new checkboxes (2/3 stall limit) |
| 2026-05-11 22:05 | Step 1 started | Black-box probe extended-metric availability (§7.4 #4) |
