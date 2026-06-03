# TP-137: Resolved workout target previews for planned workouts — Status
**Current Step:** Step 3: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 7
**Iteration:** 2
**Size:** M
> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.
---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Design compact resolved-target shape
**Status:** ✅ Complete

- [x] Audit event/workout read rows and `workout_doc_summary` to find the least-bloated place for target previews.
- [x] Document exact `workout_doc_summary.target_previews` shape, examples, call-site scope, and null/omission rules.
- [x] Use athlete profile thresholds/units only when already available or cheaply fetchable; avoid extra heavy calls or raw payload expansion.
- [x] Record unsupported target cases and null/omission rules in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/workoutdoc`.

---

### Step 2: Implement target previews and tests
**Status:** ✅ Complete

- [x] Add tests for `% FTP` planned workout targets resolving to watts from profile FTP.
- [x] Add tests or explicit omissions for HR threshold, pace threshold, missing profile threshold, and non-numeric/text targets.
- [x] Implement compact preview fields while preserving terse-by-default and `include_full` behavior.
- [x] Run targeted tests: `go test ./internal/tools ./internal/workoutdoc`.

---

### Step 3: Testing & Verification
**Status:** 🟨 In Progress

- [x] Run formatting/import-order check: `make fmt-check`
- [x] Run FULL test suite: `make test`
- [x] Run lint: `make lint`
- [x] Fix all failures or document pre-existing unrelated failures with exact command output
- [x] Build passes: `make build`

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

---

## Discoveries

| Date | Step | Finding | Impact |
|------|------|---------|--------|
| 2026-06-03 | Step 1 | Design shape: add optional `target_previews` inside existing `workout_doc_summary`; each item is compact `{step,path,description,family,target,preview,basis,repeat_reps?}` where `target` is the original target string (for example `88-94% FTP`) and `preview` is the resolved human label (for example `220-235 W`). Omit `target_previews` entirely when no supported target resolves; never emit null/empty placeholder arrays. | Keeps terse rows compact and avoids exposing raw `workout_doc`/step payloads by default. |
| 2026-06-03 | Step 1 | Call-site scope: `workout_doc_summary` rows are produced through `eventRow` (`get_events`, `get_event_by_id`, `get_today` annotations, add/update/apply/delete event response rows) and through workout rows (`get_workout_library`, `get_workouts_in_folder`, `create_workout`, `update_workout`). Implement profile-aware summaries for shared row helpers so read paths and write responses stay consistent when the handler has already fetched the profile. | Prevents divergent row shapes and identifies tests needed beyond the two originally scoped files. |
| 2026-06-03 | Step 1 | Profile rules: replace `toolProfile` usage in affected handlers with a helper that returns the already fetched `AthleteWithSportSettings`, unit system, and timezone; do not add a second profile API call. Match sport settings by event/workout sport/type against `Type` and `Types` case-insensitively; fall back only when exactly one setting exists. For indoor workouts use `indoor_ftp` when `indoor:true` and positive, otherwise `ftp`. | Reuses cheap profile data and avoids guessing thresholds from unrelated sports. |
| 2026-06-03 | Step 1 | Conversion rules: support numeric scalar/range/ramp bounds for `% FTP` to rounded watts (`ftp * percent / 100`), `% LTHR` to bpm (`lthr * percent / 100`), `% HR`/`% max HR` to bpm (`max_hr * percent / 100`), and threshold pace percent to preferred pace using speed-percent semantics (`target_seconds = threshold_seconds * 100 / percent`, so >100% is faster). Preserve path/repeat context for nested/repeated steps without expanding full raw steps. | Gives explicit formulas and avoids misrepresenting pace percentages. |
| 2026-06-03 | Step 1 | Omission rules: omit previews for missing/zero thresholds, unmatched sport settings, non-numeric/text targets, RPE/cadence/freeride, absolute watts/bpm/pace values, zones (power/HR/pace zone boundaries are future work to avoid open-ended zone ambiguity), invalid pace units, zero/negative pace percentages, and unrecognized workout_doc structures. | Unsupported targets fail closed with no misleading nulls. |
| 2026-06-03 | Step 1 | Unsupported target cases verified against workoutdoc syntax: `PERCENT_FTP`, `PERCENT_LTHR`, `PERCENT_HR`/`PERCENT_MAX_HR`, and `PERCENT_THRESHOLD`/pace aliases are the only planned preview inputs; absolute `WATTS`/`BPM`/`PACE`, zones, RPE, cadence, freeride, and text pace labels remain summarized only. `target_previews` is omitted instead of set to null or `[]` whenever all targets omit. | Gives Step 2 a closed list of positive and negative test expectations. |

## Blockers

| Date | Step | Blocker | Resolution |
|------|------|---------|------------|

## Review Notes

| Date | Review Type | Result | Notes |
|------|-------------|--------|-------|

| 2026-06-03 15:43 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 15:43 | Step 0 started | Preflight |
| 2026-06-03 15:46 | Review R001 | plan Step 1: REVISE | Added concrete shape, call-site, profile, conversion, and test-scope plan in Discoveries. |
| 2026-06-03 15:49 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 | Review | code Step 1: APPROVE |
| 2026-06-03 | Review | plan Step 2: APPROVE |
| 2026-06-03 | Review | code Step 2: APPROVE |
| 2026-06-03 | Review | plan Step 3: REVISE | Added explicit `make fmt-check` gate before full suite/lint/build. |

| 2026-06-03 15:49 | Worker iter 1 | done in 394s, tools: 39 |
| 2026-06-03 15:51 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 15:53 | Review R004 | plan Step 2: APPROVE |
| 2026-06-03 16:06 | Review R005 | code Step 2: APPROVE |
| 2026-06-03 16:07 | Review R006 | plan Step 3: REVISE |
| 2026-06-03 16:07 | Review R007 | plan Step 3: APPROVE |
