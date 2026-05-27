# TP-107: Unit semantics audit — Status

**Current Step:** Step 2: Add work/energy and unknown-unit regression coverage
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 2
**Review Counter:** 8
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
**Status:** ⬜ Not Started

- [ ] Activity `calories_burned` and wellness `calories_intake` distinction covered
- [ ] `hydration` versus `hydrationVolume` semantics covered or clarified
- [ ] Explanatory metadata added if needed without bloating terse responses
- [ ] Targeted wellness/activity tests passing

---

### Step 4: Changelog and full verification
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated if behavior or metadata changes
- [ ] Unit-surface discoveries logged
- [ ] Targeted tests passing

---

### Step 5: Testing & Verification
**Status:** ⬜ Not Started

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

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Preflight scope: workout target units are centralized in `workoutTargetUnits`/`formatTarget`; interval units use `units.ParseUnit` with `unknown_unit`; extended metrics currently convert raw joules fields to `*_kj`; wellness calories use `calories_intake` while activity rows use `calories_burned`; hydration and `hydrationVolume` are emitted as separate wellness fields. | Drives regression coverage in Steps 1-3. | `internal/workoutdoc/serialize.go`; `internal/tools/get_activity_details.go`; `internal/tools/get_extended_metrics.go`; `internal/tools/get_wellness_data.go` |
| Structured workout serializer does not support `MINS_KM`/`MINS_MILE` pace target units; Step 1 locks this as an unsupported-unit regression instead of silently coercing absolute pace into numeric `PACE`. | Covered by `TestSerializeRejectsUnsupportedAbsolutePaceUnits`; no serializer fix applied. | `internal/workoutdoc/workoutdoc_test.go` |
| Step 2 audit: workout-library `joules` / `joules_above_ftp` exist only on the internal upstream DTO; current `get_workout_library` and `get_workouts_in_folder` terse rows do not expose them, and include-full preserves raw `workout_doc` rather than relabeling energy fields. Custom items are preserved verbatim, and activity histograms emit only power/HR/pace units. | Treat as audit-only for TP-107; no additive labels needed unless these fields become public tool fields later. | `internal/intervals/workout_library.go`; `internal/tools/get_workout_library.go`; `internal/tools/get_workouts_in_folder.go`; `internal/tools/get_custom_items.go`; `internal/tools/get_activity_histogram.go` |
| Unknown upstream unit preservation already exists in activity intervals (`unit: UNKNOWN` plus `unknown_unit`) and `units.ParseUnit`; Step 2 added preferred-unit energy pass-through coverage so KJ/KCAL are not converted by distance-unit preferences. | Existing interval coverage retained; response coverage extended. | `internal/tools/get_activity_details_test.go`; `internal/units/unit_test.go`; `internal/response/units_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-26 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 12:51 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 12:51 | Step 0 started | Preflight |

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
| 2026-05-27 12:57 | Review R001 | plan Step 1: REVISE |
| 2026-05-27 12:59 | Review R002 | plan Step 1: REVISE |
| 2026-05-27 13:01 | Review R003 | plan Step 1: APPROVE |
| 2026-05-27 13:07 | Review R004 | code Step 1: REVISE |
| 2026-05-27 13:10 | Review R005 | code Step 1: APPROVE |
| 2026-05-27 13:13 | Review R006 | plan Step 2: REVISE |
| 2026-05-27 13:15 | Review R007 | plan Step 2: APPROVE |
| 2026-05-27 13:20 | Review R008 | code Step 2: APPROVE |
