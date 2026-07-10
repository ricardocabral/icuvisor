# TP-229: Treat threshold pace as m/s and pace zones as percentages — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 3
**Review Counter:** 23
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

**Status:** ✅ Complete

- [x] Typed upstream fields completed
- [x] Threshold pace read shaping corrected
- [x] Percentage zone response added
- [x] Unknown-unit fallback preserved
- [x] R010: Document the typed-field, response/fallback, and table-driven coverage plan

---

### Step 3: Correct sport-settings writes

**Status:** ✅ Complete

- [x] Explicit pace inputs convert to m/s
- [x] pace_units and pace_load_type are correct
- [x] Pace-zone percentage validation implemented
- [x] Delete-mode zone gate preserved
- [x] R013: Specify m/s transport, truthful write echo, percentage validation, and generated-schema coverage

---

### Step 4: Replace misleading fixtures and lock semantics

**Status:** ✅ Complete

- [x] Realistic upstream fixture values installed
- [x] Run/swim/row round-trip scenarios covered
- [x] m/s regression assertions added
- [x] Percentage zones remain unchanged
- [x] R016: Lock fixture consumers, replace remaining duration-shaped fixtures, and test returned m/s echoes
- [x] R017: Complete the repository-wide duration-shaped fixture audit
- [x] R019: Lock every fixture zone/name and distinguish returned m/s echoes from requests

---

### Step 5: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Race suite passing
- [x] Lint passing
- [x] Build passes
- [x] Generated docs clean
- [x] R021: Verify generator output has no committed-data drift

---

### Step 6: Documentation & Delivery

**Status:** 🟨 In Progress

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
| R011 | Plan | 2 | APPROVE | — |
| R012 | Code | 2 | APPROVE | — |
| R013 | Plan | 3 | REVISE | `.reviews/R013-plan-step3.md` |
| R014 | Plan | 3 | APPROVE | — |
| R015 | Code | 3 | APPROVE | — |
| R016 | Plan | 4 | REVISE | `.reviews/R016-plan-step4.md` |
| R017 | Plan | 4 | REVISE | `.reviews/R017-plan-step4.md` |
| R018 | Plan | 4 | APPROVE | — |
| R019 | Code | 4 | REVISE | `.reviews/R019-code-step4.md` |
| R020 | Code | 4 | APPROVE | — |
| R021 | Plan | 5 | REVISE | `.reviews/R021-plan-step5.md` |
| R022 | Plan | 5 | APPROVE | — |
| R023 | Code | 5 | APPROVE | — |

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
- Step 2 plan: add `PaceLoadType string \`json:"pace_load_type"\`` to `intervals.SportSettings` and preserve its raw `RUN`/`SWIM` value into a `pace_load_type` response field. Shape `ThresholdPace` only as m/s through `PaceSecondsFromMetersPerSecond` for its `pace_units` distance, never profile-wide units or `PaceThreshold`; copy `PaceZones` unchanged to `pace_zones_percent_of_threshold` with names and metadata legend. Known `NONE`, unknown tokens, and finite conversion overflows retain only `threshold_pace_meters_per_second`; only an unknown non-empty token gets `_meta.unknown_unit`. Add table-driven decoding/shaping coverage for run km/mile, swim 100m/100y, row 500m, pace-load preservation, percentages/names/no legacy duration zones, and all fallback cases; leave fixture replacement to Step 4.
- Step 3 plan: redesign `intervals.SportSettingsPace` so `Value` is m/s while separate `PaceUnits` and `PaceLoadType` fields are emitted by `writeSportSettingsBody`; local HTTP tests will assert the exact m/s body value, preserved valid display/load values, inferred input display, derived RUN/SWIM only when upstream load type is absent, and no preservation of `NONE`/unknown display tokens. Convert each explicit duration input to seconds for its named distance, then call `response.PaceMetersPerSecondFromSeconds`. Rewrite the update echo via `units.ParseUnit` and `response.PaceSecondsFromMetersPerSecond`, returning explicit seconds fields only for known display units and `threshold_pace_meters_per_second` for `NONE`, unknown, or overflow; preserve selected source/load metadata and cover returned plus params-fallback `3.5714285` m/s → 280 s/km. Treat pace-zone boundaries as finite, strictly increasing percent values in `(0, 200]`; do not transform their values, keep the full delete-mode gate before clients, replace duration schema descriptions/examples with `[77.5,100]`, and test zero/non-finite/>200/duplicate/descending validation plus safe-mode no-client calls. Regenerate schema snapshot, website data, and gendocs goldens; run the targeted tool/intervals/gendocs tests.
- Step 4 plan: change the checked-in athlete profile fixture to a RUN setting with `threshold_pace: 3.5714285`, display-only `MINS_KM`, `pace_load_type: RUN`, and percentage zones `[77.5,90,100]`; strengthen its real client decode test to assert type, m/s, presentation/load metadata, percentages, and names. Replace all remaining duration-shaped test records in get-athlete-profile readiness, data-quality, and histogram fallback tests with `ThresholdPace` m/s where a threshold is used, explicit applicable display units, and ascending percentage zones; retain the histogram fallback's no-threshold/empty-display condition while changing its zones to percentages. Remove the irrelevant duration-valued pace fields from the cycling performance-potential fixture. Audit the repository so no duration-shaped `PaceThreshold` fixtures or implausibly large `ThresholdPace` values remain. Add a table-driven semantic regression file that exercises tool-level duration input → m/s transport → returned upstream m/s selected-display echo for metric/imperial run, 100m/100y swim, and 500m row; explicitly prove 3.5714285 m/s renders as approximately 280 s/km and a 280 s/km write returns the same m/s; assert [77.5,100] is copied untouched into a pace-zone definition. Extend response converter tests only where necessary and run response, athleteprofile, intervals, get-athlete-profile, data-quality, histogram, performance-potential, and semantic tool targets.
- Step 5 plan: run `make test`, `make test-race`, `make lint`, and `make build`; execute `make docs-tools && git diff --exit-code -- web/data/tools.json web/data/tool_schemas.json && git diff --check` from the committed baseline, so generated-data drift and whitespace errors both fail verification. Record each gate in the execution log and rerun any gate affected by a correction.
| 2026-07-10 19:30 | Review R002 | code Step 1: REVISE |
| 2026-07-10 19:39 | Review R003 | code Step 1: REVISE |
| 2026-07-10 19:45 | Review R004 | code Step 1: REVISE |
| 2026-07-10 19:50 | Review R005 | code Step 1: REVISE |
| 2026-07-10 20:00 | Review R006 | code Step 1: REVISE |
| 2026-07-10 20:05 | Review R007 | code Step 1: REVISE |
| 2026-07-10 20:11 | Review R008 | code Step 1: REVISE |
| 2026-07-10 20:15 | Review R009 | code Step 1: APPROVE |
| 2026-07-10 20:18 | Review R010 | plan Step 2: REVISE |
| 2026-07-10 20:21 | Review R011 | plan Step 2: APPROVE |
| 2026-07-10 20:26 | Review R012 | code Step 2: APPROVE |
| 2026-07-10 21:00 | Verification | `make test` passed |
| 2026-07-10 21:01 | Verification | `make test-race` passed |
| 2026-07-10 21:02 | Verification | `make lint` passed |
| 2026-07-10 21:02 | Verification | `make build` passed |
| 2026-07-10 21:03 | Verification | `make docs-tools`, generated-file diff, and `git diff --check` passed |
| 2026-07-10 20:31 | Review R013 | plan Step 3: REVISE |
| 2026-07-10 20:33 | Review R014 | plan Step 3: APPROVE |
| 2026-07-10 20:44 | Review R015 | code Step 3: APPROVE |
| 2026-07-10 20:48 | Review R016 | plan Step 4: REVISE |
| 2026-07-10 20:50 | Review R017 | plan Step 4: REVISE |
| 2026-07-10 20:53 | Review R018 | plan Step 4: APPROVE |
| 2026-07-10 20:59 | Review R019 | code Step 4: REVISE |
| 2026-07-10 21:02 | Review R020 | code Step 4: APPROVE |
| 2026-07-10 21:03 | Review R021 | plan Step 5: REVISE |
| 2026-07-10 21:05 | Review R022 | plan Step 5: APPROVE |
| 2026-07-10 21:08 | Review R023 | code Step 5: APPROVE |
