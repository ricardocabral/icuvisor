# TP-150: Workout DSL metric suffix from sport priority — Status

**Current Step:** Step 1: Design the sport-aware suffix boundary
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
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
**Status:** 🟨 In Progress

- [x] Decide implementation boundary for suffix selection
- [x] Preserve no-sport-context serializer behavior
- [x] Define expected suffix behavior for primary metric orders
- [x] Document upstream ambiguity in STATUS.md
- [ ] Decide and document `workout_order` decode/source
- [ ] Decide and document `update_workout` missing-sport fallback
- [ ] Decide and document `apply_training_plan` planned-write coverage

---

### Step 2: Implement and test metric suffix behavior
**Status:** ⬜ Not Started

- [ ] Add Run `POWER_HR_PACE` regression test
- [ ] Add HR-primary / pace-primary coverage where applicable
- [ ] Implement minimal behavior change
- [ ] Targeted tests passing

---

### Step 3: Refresh schemas and user guidance
**Status:** ⬜ Not Started

- [ ] Tool descriptions/schema wording updated if needed
- [ ] Schema snapshots regenerated if needed
- [ ] End-user workout docs updated if needed
- [ ] CHANGELOG updated if user-visible

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passes
- [ ] All failures fixed
- [ ] Build passes

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Clean-room behavior source summarized

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Upstream public behavior confirms explicit metric suffixes avoid zone-family ambiguity, but this task does not include authoritative upstream docs for when bare `Z2` is safe for each `workout_order`. | Implement sport-aware writes with explicit zone family suffixes whenever supported workout order context is known; preserve existing bare serializer for no-context callers. | Step 1 design |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Step 0 evidence: `go test ./internal/workoutdoc -run TestSerializeTargetUnitSemantics -count=1` passed and existing case `POWER_ZONE` expects bare `Z2`.
- Step 1 boundary: add an options-aware WorkoutDoc serialization path in `internal/workoutdoc` (for example `SerializeWithOptions`/`MergeDescriptionWithOptions`) and have planned-workout write call sites pass the target sport's `workout_order` from the athlete profile; keep suffix rendering centralized in the serializer rather than duplicating string rewrites in each tool.
- No-sport-context behavior: existing `workoutdoc.Serialize` and `workoutdoc.MergeDescription` remain unchanged, so resources, validators, activity-interval writes, and tests that lack sport settings continue to emit bare power zones (`Z2`).
- Sport-aware suffix expectation: when a known planned-workout sport setting exposes `workout_order` (`POWER_HR_PACE`, `HR_POWER_PACE`, or `PACE_HR_POWER`), zone targets serialize with explicit metric suffixes by target family: power `Z2 Power`/`Z2-Z3 Power`, heart rate `Z2 HR`/`Z2-Z3 HR`, pace `Z2 Pace`/`Z2-Z3 Pace`. The order identifies that the upstream sport has metric priority context; icuvisor avoids relying on whichever family upstream treats as bare default.
- R001 plan review requires explicit decisions for decoding `workout_order`, `update_workout` without sport, and `apply_training_plan` write coverage before Step 2.
| 2026-06-03 21:33 | Review R001 | plan Step 1: UNKNOWN |
