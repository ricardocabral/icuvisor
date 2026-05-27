# TP-107: Unit semantics audit — Status

**Current Step:** Step 5: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 2
**Review Counter:** 16
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Current unit/metadata behavior scoped before changing code

---

### Step 1: Add workout target unit regression coverage
**Status:** ✅ Complete

- [x] Percent FTP / power target serialization tests added
- [x] Pace target range and unit tests added
- [x] Heart-rate percent variant tests added where supported
- [x] Serializer fixes applied only if required
- [x] Targeted workoutdoc tests passing
- [x] Direct table-driven serializer matrix covers scalar/range power, pace, HR, zone, watt, BPM, and text pace forms
- [x] Unsupported structured absolute pace target units are documented as discovery or fixed additively with tests
- [x] R004 coverage gap fixed for blank pace units and `%HR`/`HR` aliases

---

### Step 2: Add work/energy and unknown-unit regression coverage
**Status:** ✅ Complete

> ⚠️ Hydrate: Expand based on actual unit-bearing surfaces found during audit.

- [x] Joules/kilojoules surfaces audited and covered
- [x] Extended metrics raw-joule-to-kJ conversion tests cover activity, interval, and strain-score W' fields with `_meta.extended_metric_units`
- [x] Workout-library joules fields are audited as raw/full-only or covered if surfaced
- [x] Raw joules not mislabeled as kilojoules
- [x] Unknown units preserved rather than guessed
- [x] Response preferred-unit pass-through covers KJ/KCAL and unknown raw unit labels
- [x] Targeted unit/response tests passing

---

### Step 3: Add calories and hydration semantics coverage
**Status:** ✅ Complete

- [x] Activity `calories_burned` and wellness `calories_intake` distinction covered
- [x] `hydration` versus `hydrationVolume` semantics covered or clarified
- [x] Explanatory metadata added if needed without bloating terse responses
- [x] Targeted wellness/activity tests passing
- [x] Hydration row preserves `hydration` and `hydrationVolume` distinctly with terse metadata and include_full raw preservation
- [x] R011 null hydration fields do not leave stale field semantics in terse responses

---

### Step 4: Changelog and full verification
**Status:** ✅ Complete

- [x] `CHANGELOG.md` updated if behavior or metadata changes
- [x] Unit-surface discoveries logged
- [x] Targeted tests passing

---

### Step 5: Testing & Verification
**Status:** 🟨 In Progress

- [ ] Targeted tests passing
- [ ] FULL test suite passing: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | APPROVE | `.reviews/R003-plan-step1.md` |
| R004 | Code | 1 | REVISE | `.reviews/R004-code-step1.md` |
| R005 | Code | 1 | APPROVE | `.reviews/R005-code-step1.md` |
| R006 | Plan | 2 | REVISE | `.reviews/R006-plan-step2.md` |
| R007 | Plan | 2 | APPROVE | `.reviews/R007-plan-step2.md` |
| R008 | Code | 2 | APPROVE | `.reviews/R008-code-step2.md` |
| R009 | Plan | 3 | REVISE | `.reviews/R009-plan-step3.md` |
| R010 | Plan | 3 | APPROVE | `.reviews/R010-plan-step3.md` |
| R011 | Code | 3 | REVISE | `.reviews/R011-code-step3.md` |
| R012 | Code | 3 | APPROVE | `.reviews/R012-code-step3.md` |
| R013 | Plan | 4 | REVISE | `.reviews/R013-plan-step4.md` |
| R014 | Plan | 4 | APPROVE | `.reviews/R014-plan-step4.md` |
| R015 | Code | 4 | APPROVE | `.reviews/R015-code-step4.md` |
| R016 | Plan | 5 | REVISE | `.reviews/R016-plan-step5.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Preflight scope: workout target units are centralized in `workoutTargetUnits`/`formatTarget`; interval units use `units.ParseUnit` with `unknown_unit`; extended metrics currently convert raw joules fields to `*_kj`; wellness calories use `calories_intake` while activity rows use `calories_burned`; hydration and `hydrationVolume` are emitted as separate wellness fields. | Drives regression coverage in Steps 1-3. | `internal/workoutdoc/serialize.go`; `internal/tools/get_activity_details.go`; `internal/tools/get_extended_metrics.go`; `internal/tools/get_wellness_data.go` |
| Structured workout serializer does not support `MINS_KM`/`MINS_MILE` pace target units; Step 1 locks this as an unsupported-unit regression instead of silently coercing absolute pace into numeric `PACE`. | Covered by `TestSerializeRejectsUnsupportedAbsolutePaceUnits`; no serializer fix applied. | `internal/workoutdoc/workoutdoc_test.go` |
| Step 2 audit: workout-library `joules` / `joules_above_ftp` exist only on the internal upstream DTO; current `get_workout_library` and `get_workouts_in_folder` terse rows do not expose them, and include-full preserves raw `workout_doc` rather than relabeling energy fields. Custom items are preserved verbatim, and activity histograms emit only power/HR/pace units. | Treat as audit-only for TP-107; no additive labels needed unless these fields become public tool fields later. | `internal/intervals/workout_library.go`; `internal/tools/get_workout_library.go`; `internal/tools/get_workouts_in_folder.go`; `internal/tools/get_custom_items.go`; `internal/tools/get_activity_histogram.go` |
| Unknown upstream unit preservation already exists in activity intervals (`unit: UNKNOWN` plus `unknown_unit`) and `units.ParseUnit`; Step 2 added preferred-unit energy pass-through coverage so KJ/KCAL are not converted by distance-unit preferences. | Existing interval coverage retained; response coverage extended. | `internal/tools/get_activity_details_test.go`; `internal/units/unit_test.go`; `internal/response/units_test.go` |
| Calories semantics were already covered on activity details and wellness rows: activity emits `calories_burned` and rejects wellness intake keys; wellness emits `calories_intake` and keeps raw `kcalConsumed` only under `full` in include-full mode. | Relied on existing regression tests for Step 3; hydration metadata coverage was added. | `internal/tools/get_activity_details_test.go`; `internal/tools/get_wellness_data_test.go` |
| Unit-semantics audit did not require workout serializer behavior changes; Step 1 is regression coverage only. Hydration metadata is additive and avoids stale field semantics when upstream hydration fields are null. | Changelog updated for user-visible hydration metadata; no generated tool docs/catalog change required because schemas/descriptions were not changed. | `internal/workoutdoc/workoutdoc_test.go`; `internal/tools/get_wellness_data.go`; `CHANGELOG.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-26 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 12:51 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 12:51 | Step 0 started | Preflight |
| 2026-05-27 | Step 4 affected-package tests | `go test ./internal/workoutdoc ./internal/units ./internal/response ./internal/tools` passed |

---

## Blockers

*None*

---

## Notes

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/34
- Step 1 implementation plan: add `TestSerializeTargetUnitSemantics` in `internal/workoutdoc/workoutdoc_test.go`. Each table case will build a small `WorkoutDoc`, call `Serialize`, and assert the exact DSL string.
- Step 1 serializer matrix: power blank/default percent FTP, `PERCENT_FTP`, `%FTP`, watts aliases (`WATTS`, `WATT`, `W`), and power zone scalar/range; pace percent-threshold aliases (`PERCENT_THRESHOLD`, `PERCENT_THRESHOLD_PACE`, `PERCENT_PACE`, `%PACE`), `PACE` numeric scalar/range, pace zone scalar/range, and text pace form (`5:00/km Pace`); HR `% HR` via `PERCENT_HR`/`PERCENT_MAX_HR`, `% LTHR` via `PERCENT_LTHR`/`%LTHR`/`LTHR`, BPM, and HR zone scalar/range.
- Step 1 `MINS_KM`/`MINS_MILE` decision: current structured workout target units do not list these tokens; add an unsupported-unit regression and discovery unless tests reveal documented syntax metadata requiring an additive serializer/syntax fix. Do not coerce them to `PACE` silently.
- Step 1 verification command: `go test ./internal/workoutdoc`; update discoveries with the proven behavior.
- Step 2 implementation plan: add extended-metrics regression coverage in `internal/tools/get_extended_metrics_test.go` for every raw joule field normalized to kJ: activity `icu_joules_above_ftp`, activity `ss_w_prime`, interval `wbal_start`, interval `wbal_end`, and interval `joules_above_ftp`; assert divided values and `_meta.extended_metric_units`.
- Step 2 unknown-unit/energy plan: extend `internal/response/units_test.go` with `ToPreferredWithRaw` assertions that `KJ`, `KCAL`, and an unknown raw token pass through without conversion or guessed labels; record existing interval `unknown_unit` coverage if no code change is needed there.
- Step 2 audit-only surfaces: `get_workout_library` / `get_workouts_in_folder` do not expose raw `joules` / `joules_above_ftp` as terse fields today; include-full workout docs are preserved rather than relabeled. Custom-item content is preserved verbatim and should not parse/relabel embedded units. `get_activity_histogram` emits only power/HR/pace histogram units unless audit finds a joule-bearing path.
- Step 2 audit command: grep/review `joules`, `KJ`, `KCAL`, and `unknown_unit` across `internal/units`, `internal/response`, and relevant tools; log audit-only results in discoveries.
- Step 2 verification commands: targeted `go test ./internal/units ./internal/response ./internal/tools` after narrower iterations as needed.
- Step 3 calories plan: rely on and, if needed, extend existing `internal/tools/get_activity_details_test.go` and `internal/tools/get_wellness_data_test.go` coverage asserting activity `calories_burned` is distinct from wellness `calories_intake`, ambiguous `calories`/`kcalConsumed` do not leak at top level, and `_meta.field_semantics` explains nutrition semantics.
- Step 3 hydration plan: add an inline or fixture-backed wellness row containing both upstream `hydration` and `hydrationVolume`; assert both top-level fields are preserved distinctly in terse mode, neither is renamed/collapsed, row-level `_meta.field_semantics` describes the distinction without adding bulky top-level fields, and `include_full:true` preserves raw upstream names only under `full`.
- Step 3 metadata decision: prefer row-level `_meta.field_semantics` for hydration semantics; do not rename public fields. Keep default rows terse apart from `_meta` entries.
- Step 3 verification command: `go test ./internal/tools -run 'TestGetActivityDetails|TestGetWellnessData'`; log any existing calories coverage used as a discovery.
- Step 4 changelog plan: update `CHANGELOG.md` `[Unreleased]` with a concise user-visible entry that wellness rows now include `_meta.field_semantics` for `hydration` and `hydrationVolume`, without renaming fields or inventing units.
- Step 4 discovery plan: ensure `STATUS.md` captures no serializer behavior change beyond tests, energy/joule tested and audit-only surfaces, hydration metadata addition, and null hydration fields avoiding stale semantics.
- Step 4 verification plan: run affected-package tests `go test ./internal/workoutdoc ./internal/units ./internal/response ./internal/tools` and record the command/result in the execution log. Leave `make test`, `make build`, and `make lint` for Step 5.
- Step 4 docs/catalog boundary: no tool schema or description text is planned; generated tool docs/catalog regeneration is not needed unless that changes.
- Step 5 verification plan: rerun targeted affected packages with `go test ./internal/workoutdoc ./internal/units ./internal/response ./internal/tools`, then run `make test`, `make build`, and `make lint` in that order. Record each command/result in the execution log.
- Step 5 failure handling: fix task-related failures before checking boxes; if a failure appears pre-existing/unrelated, document the command, failure excerpt, and rationale in `STATUS.md` and leave the checkbox truthful.
- Step 5 environment: no network-dependent tests or external services are expected.
| 2026-05-27 12:57 | Review R001 | plan Step 1: REVISE |
| 2026-05-27 12:59 | Review R002 | plan Step 1: REVISE |
| 2026-05-27 13:01 | Review R003 | plan Step 1: APPROVE |
| 2026-05-27 13:07 | Review R004 | code Step 1: REVISE |
| 2026-05-27 13:10 | Review R005 | code Step 1: APPROVE |
| 2026-05-27 13:13 | Review R006 | plan Step 2: REVISE |
| 2026-05-27 13:15 | Review R007 | plan Step 2: APPROVE |
| 2026-05-27 13:20 | Review R008 | code Step 2: APPROVE |
| 2026-05-27 13:23 | Review R009 | plan Step 3: REVISE |
| 2026-05-27 13:25 | Review R010 | plan Step 3: APPROVE |
| 2026-05-27 13:31 | Review R011 | code Step 3: REVISE |
| 2026-05-27 13:34 | Review R012 | code Step 3: APPROVE |
| 2026-05-27 13:37 | Review R013 | plan Step 4: REVISE |
| 2026-05-27 13:38 | Review R014 | plan Step 4: APPROVE |
| 2026-05-27 13:41 | Review R015 | code Step 4: APPROVE |
| 2026-05-27 13:43 | Review R016 | plan Step 5: REVISE |
