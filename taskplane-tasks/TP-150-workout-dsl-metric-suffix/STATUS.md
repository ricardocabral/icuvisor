# TP-150: Workout DSL metric suffix from sport priority — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 9
**Iteration:** 3
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm current tests lock bare power-zone serialization

---

### Step 1: Design the sport-aware suffix boundary
**Status:** ✅ Complete

- [x] Decide implementation boundary for suffix selection
- [x] Preserve no-sport-context serializer behavior
- [x] Define expected suffix behavior for primary metric orders
- [x] Document upstream ambiguity in STATUS.md
- [x] Decide and document `workout_order` decode/source
- [x] Decide and document `update_workout` missing-sport fallback
- [x] Decide and document `apply_training_plan` planned-write coverage

---

### Step 2: Implement and test metric suffix behavior
**Status:** ✅ Complete

- [x] Add Run `POWER_HR_PACE` regression test
- [x] Add HR-primary / pace-primary coverage where applicable
- [x] Add `workout_order` decode/helper coverage and update apply-training-plan path coverage
- [x] Implement minimal behavior change
- [x] Targeted tests passing

---

### Step 3: Refresh schemas and user guidance
**Status:** ✅ Complete

- [x] Tool descriptions/schema wording updated if needed
- [x] Schema snapshots regenerated if needed
- [x] End-user workout docs updated if needed
- [x] CHANGELOG updated if user-visible

---

### Step 4: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Lint passes
- [x] All failures fixed
- [x] Build passes

---

### Step 5: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Clean-room behavior source summarized

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | plan | 2 | APPROVE | `.reviews/R004-plan-step2.md` |
| R005 | code | 2 | APPROVE | `.reviews/R005-code-step2.md` |
| R006 | plan | 3 | APPROVE | `.reviews/R006-plan-step3.md` |
| R007 | code | 3 | APPROVE | `.reviews/R007-code-step3.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Upstream public behavior confirms explicit metric suffixes avoid zone-family ambiguity, but this task does not include authoritative upstream docs for when bare `Z2` is safe for each `workout_order`. | Implement sport-aware writes with explicit zone family suffixes whenever supported workout order context is known; preserve existing bare serializer for no-context callers. | Step 1 design |
| Documentation review found the product contract already covers canonical WorkoutDoc serialization; the user-facing cookbook needed the operational guidance, while PRD/resource reference did not need a behavior-contract change. | Updated cookbook and changelog; left PRD and resources-prompts unchanged after review. | Step 5 delivery |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |
| 2026-06-03 21:42 | Worker iter 1 | done in 883s, tools: 105 |
| 2026-06-03 21:59 | Worker iter 2 | done in 994s, tools: 60 |
| 2026-06-03 22:31 | Exit intercept reprompt | Supervisor provided instructions (549 chars) — reprompting worker |

---

## Blockers

*None*

---

## Notes

- Step 5 evidence: Must-update docs satisfied by CHANGELOG.md `[Unreleased]` fixed entry for planned workout writes emitting explicit zone metric suffixes from athlete sport settings.
- Step 5 evidence: Check-if-affected docs reviewed: `web/content/cookbook/build-workouts.md` updated with zone-suffix/update-workout sport guidance; `web/content/reference/resources-prompts.md` and `docs/prd/PRD-icuvisor.md` reviewed with no product contract/reference table change needed.
- Step 5 clean-room summary: implementation was derived from icuvisor's existing WorkoutDoc serializer/tests plus the task's public upstream behavior signal that planned Run workouts with ambiguous sport metric priority require explicit zone-family suffixes such as `Z2 Power`, `Z2 HR`, or `Z2 Pace`; no external/copyleft implementation code was used.
- Step 0 evidence: `go test ./internal/workoutdoc -run TestSerializeTargetUnitSemantics -count=1` passed and existing case `POWER_ZONE` expects bare `Z2`.
- Step 1 boundary: add an options-aware WorkoutDoc serialization path in `internal/workoutdoc` (for example `SerializeWithOptions`/`MergeDescriptionWithOptions`) and have planned-workout write call sites pass the target sport's `workout_order` from the athlete profile; keep suffix rendering centralized in the serializer rather than duplicating string rewrites in each tool.
- No-sport-context behavior: existing `workoutdoc.Serialize` and `workoutdoc.MergeDescription` remain unchanged, so resources, validators, activity-interval writes, and tests that lack sport settings continue to emit bare power zones (`Z2`).
- Sport-aware suffix expectation: when a known planned-workout sport setting exposes `workout_order` (`POWER_HR_PACE`, `HR_POWER_PACE`, or `PACE_HR_POWER`), zone targets serialize with explicit metric suffixes by target family: power `Z2 Power`/`Z2-Z3 Power`, heart rate `Z2 HR`/`Z2-Z3 HR`, pace `Z2 Pace`/`Z2-Z3 Pace`. The order identifies that the upstream sport has metric priority context; icuvisor avoids relying on whichever family upstream treats as bare default.
- R001 plan review requires explicit decisions for decoding `workout_order`, `update_workout` without sport, and `apply_training_plan` write coverage before Step 2.
- R001 resolution (`workout_order` source): add `WorkoutOrder string `json:"workout_order"`` to `intervals.SportSettings` and table-test JSON decoding. A tool-layer helper will select a sport setting by exact `type` or membership in `types`, normalize only known values (`POWER_HR_PACE`, `HR_POWER_PACE`, `PACE_HR_POWER`), and return zero options for missing/unknown orders.
- R001 resolution (`update_workout` no sport): `update_workout` will use sport-aware serialization only when the request supplies `sport`; otherwise it intentionally falls back to existing bare serialization because the sparse update path has no existing workout fetch in scope. The schema/docs will tell callers to include `sport` with `workout_doc` when they need sport-aware metric suffixes.
- R001 resolution (`apply_training_plan`): keep it in scope. `applyTrainingPlan` already has athlete profile and each template's `type`; pass profile/sport-derived serialization options through `eventParamsFromPlanWorkout`/`eventWriteParams` so applied calendar events get the same sport-aware zone suffixes as direct planned workout writes.
| 2026-06-03 21:33 | Review R001 | plan Step 1: UNKNOWN |
| 2026-06-03 21:35 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 21:36 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 21:38 | Review R004 | plan Step 2: APPROVE |
| 2026-06-03 21:46 | Review R005 | code Step 2: APPROVE |
| 2026-06-03 21:48 | Review R006 | plan Step 3: APPROVE |
| 2026-06-03 21:52 | Review R007 | code Step 3: APPROVE |
| 2026-06-03 21:53 | Review R008 | plan Step 4: APPROVE |
| 2026-06-03 22:55 | Review R009 | code Step 4: APPROVE |
