# TP-161: Bridge training-plan targets into fitness projection — Status

**Current Step:** Step 1: Design deterministic weekly-target distribution
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 4
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Design deterministic weekly-target distribution
**Status:** 🟨 In Progress

- [x] Weekly target input shape and no-implicit-fetch contract defined
- [x] Week anchoring, partial-week horizon, and start-date exclusion defined
- [x] Override formula and fallback-to-modeled behavior defined
- [x] Overlap-based horizon validation and mid-week current-week target case defined
- [x] Weekly target source labels and source_tools metadata behavior defined
- [x] Weekly target date/load validation bounds defined
- [x] Failing bridge tests added
- [x] Weekly target distribution assumptions defined
- [x] Explicit daily-load precedence covered
- [x] Validation/ignore behavior for duplicate/invalid/out-of-horizon weekly targets covered
- [x] Tool-level metadata/schema-facing behavior covered
- [x] Initial targeted projection/training-plan tests run
- [x] R004 recovery-week timing assertion corrected
- [x] R004 weekly target filled-day metadata count corrected

---

### Step 2: Implement bridge in projection input and schema
**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand after inspecting current `get_training_plan` output fields and upstream plan target shapes.

- [ ] Optional typed request field/helper added
- [ ] Deterministic conversion implemented
- [ ] Metadata/source assumptions updated
- [ ] Schema snapshot refreshed
- [ ] Targeted projection/training-plan tests pass

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Integration tests (if applicable)
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `README.md` updated
- [ ] `CHANGELOG.md` updated
- [ ] PRD reviewed/updated if affected
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-10 11:55 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 11:55 | Step 0 started | Preflight |
| 2026-06-10 | Step 1 tests added | `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'` fails as expected on missing `WeeklyPlanTargets`/`weekly_plan_targets` bridge support |
| 2026-06-10 | Step 1 R004 revisions | Corrected recovery-week timing and weekly-target filled-day expected count; targeted tests still fail only on missing bridge support |

---

## Blockers

*None*

---

## Notes

Public signal: IntervalCoach forum #858-859 noted goal-progress projection falling off after 7 days despite weekly TSS targets.
Plan review R001 requested concrete target shape, week anchoring/partial week semantics, exact override formula, fallback behavior, validation cases, and metadata/source_tools expectations before tests.
Plan review R002 requested overlap-based horizon validation for mid-week current-week targets, exact `weekly_plan_targets` source labels/source_tools behavior, and weekly target date/load validation bounds before tests.
Code review R004 requested correcting the Step 1 tests so recovery-week timing preserves the existing day 8/second-week behavior and weekly target filled-day metadata counts partial current-week fills correctly.
Step 1 contract: `get_fitness_projection` will accept explicit `weekly_plan_targets` entries shaped as `{week_start_date: YYYY-MM-DD, training_load: number}` supplied by the caller from `get_training_plan`/planning context; the projection tool will not fetch training plans implicitly. `week_start_date` is the athlete-local ISO/Monday week anchor. Projection day 0 (`start_date`) remains the current-fitness seed and never receives projected load. For projected dates day 1..horizon, weekly targets create candidate loads of `training_load/7` for dates in that anchored week, including partial weeks without reweighting. Explicit `planned_daily_loads` replace candidates for exact dates and do not consume or redistribute weekly target load. Dates without explicit daily loads or matching weekly targets keep existing modeled ramp/recovery sources. Weekly target overlap validation is based on intersection with projected days 1..horizon, so a current-week Monday target is valid when `start_date` is mid-week and future dates remain in that week; targets with no overlap are rejected. Daily source label for filled dates will be `weekly_plan_targets`; `_meta.source_tools` always includes `get_fitness` and adds `get_training_plan` only when weekly targets are supplied. Assumptions will include target count, filled-day count, override count, ISO Monday anchor, partial-week/no-redistribution caveat, and even `training_load/7` distribution. `week_start_date` must be a valid Monday date after trimming, duplicate normalized week anchors are rejected, and `training_load` must be finite in `[0, 7*maxProjectionPlannedDailyLoad]`.
| 2026-06-10 11:58 | Review R001 | plan Step 1: REVISE |
| 2026-06-10 12:01 | Review R002 | plan Step 1: REVISE |
| 2026-06-10 12:04 | Review R003 | plan Step 1: APPROVE |
| 2026-06-10 12:09 | Review R004 | code Step 1: REVISE |
