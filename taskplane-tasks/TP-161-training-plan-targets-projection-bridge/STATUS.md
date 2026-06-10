# TP-161: Bridge training-plan targets into fitness projection — Status

**Current Step:** Step 4: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 11
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Design deterministic weekly-target distribution
**Status:** ✅ Complete

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
**Status:** ✅ Complete

> ⚠️ Hydrate: Expand after inspecting current `get_training_plan` output fields and upstream plan target shapes.

- [x] Optional typed request field plus decoder validation for weekly targets added
- [x] Analyzer input/result types carry weekly targets and fill/override counts
- [x] Deterministic conversion implemented
- [x] Metadata/source assumptions updated
- [x] Schema snapshot refreshed
- [x] Targeted projection/training-plan tests pass
- [x] R007 generated tool-schema docs refreshed
- [x] R008 missing weekly target training_load rejected

---

### Step 3: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Integration tests (if applicable)
- [x] All failures fixed
- [x] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ✅ Complete

- [x] `README.md` updated
- [x] `CHANGELOG.md` updated
- [x] PRD reviewed/updated if affected
- [x] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | APPROVE | `.reviews/R003-plan-step1.md` |
| R004 | Code | 1 | REVISE | `.reviews/R004-code-step1.md` |
| R005 | Code | 1 | APPROVE | `.reviews/R005-code-step1.md` |
| R006 | Plan | 2 | APPROVE | `.reviews/R006-plan-step2.md` |
| R007 | Code | 2 | REVISE | `.reviews/R007-code-step2.md` |
| R008 | Code | 2 | REVISE | `.reviews/R008-code-step2.md` |
| R009 | Code | 2 | APPROVE | `.reviews/R009-code-step2.md` |
| R010 | Plan | 3 | APPROVE | `.reviews/R010-plan-step3.md` |
| R011 | Code | 3 | APPROVE | `.reviews/R011-code-step3.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `get_training_plan` currently returns assignment fields plus lightweight plan summary by default and preserves raw nested plan payload only with `include_full:true`; this bridge should therefore accept caller-supplied weekly targets rather than extracting them implicitly. | Reflected in `weekly_plan_targets` request contract and schema wording. | `internal/tools/get_training_plan.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-10 11:55 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 11:55 | Step 0 started | Preflight |
| 2026-06-10 | Step 1 tests added | `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'` fails as expected on missing `WeeklyPlanTargets`/`weekly_plan_targets` bridge support |
| 2026-06-10 | Step 1 R004 revisions | Corrected recovery-week timing and weekly-target filled-day expected count; targeted tests still fail only on missing bridge support |
| 2026-06-10 | Step 2 implementation | Added `weekly_plan_targets` request/schema bridge, deterministic analyzer fill with explicit-load precedence, metadata counts/source_tools, refreshed schema snapshot, and targeted tests pass |
| 2026-06-10 | Step 2 R007 revision | Regenerated website and gendocs tool-schema data/goldens; focused gendocs/projection/schema tests pass |
| 2026-06-10 | Step 2 R008 revision | Added required weekly target `training_load` decoding validation and regression test; focused projection/gendocs/schema tests pass |
| 2026-06-10 | Step 3 full test suite | `make test` passed |
| 2026-06-10 | Step 3 integration scope | No external-service integration tests apply; bridge is deterministic request/analyzer/schema logic covered by unit/golden tests |
| 2026-06-10 | Step 3 failure status | Prior targeted and generated-doc failures fixed; full suite is green |
| 2026-06-10 | Step 3 build | `make build` passed |
| 2026-06-10 | Step 4 documentation | README, CHANGELOG, and PRD updated; discovery table records no-implicit-fetch training-plan shape decision |
| 2026-06-10 12:39 | Worker iter 1 | done in 2604s, tools: 144 |
| 2026-06-10 12:39 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

Public signal: IntervalCoach forum #858-859 noted goal-progress projection falling off after 7 days despite weekly TSS targets.
Plan review R001 requested concrete target shape, week anchoring/partial week semantics, exact override formula, fallback behavior, validation cases, and metadata/source_tools expectations before tests.
Plan review R002 requested overlap-based horizon validation for mid-week current-week targets, exact `weekly_plan_targets` source labels/source_tools behavior, and weekly target date/load validation bounds before tests.
Code review R004 requested correcting the Step 1 tests so recovery-week timing preserves the existing day 8/second-week behavior and weekly target filled-day metadata counts partial current-week fills correctly.
Code review R007 requested refreshing generated gendocs schema outputs after adding `weekly_plan_targets` to `get_fitness_projection`.
Code review R008 requested presence-checking `weekly_plan_targets[].training_load` so omitted loads are rejected instead of decoded as zero.
Step 1 contract: `get_fitness_projection` will accept explicit `weekly_plan_targets` entries shaped as `{week_start_date: YYYY-MM-DD, training_load: number}` supplied by the caller from `get_training_plan`/planning context; the projection tool will not fetch training plans implicitly. `week_start_date` is the athlete-local ISO/Monday week anchor. Projection day 0 (`start_date`) remains the current-fitness seed and never receives projected load. For projected dates day 1..horizon, weekly targets create candidate loads of `training_load/7` for dates in that anchored week, including partial weeks without reweighting. Explicit `planned_daily_loads` replace candidates for exact dates and do not consume or redistribute weekly target load. Dates without explicit daily loads or matching weekly targets keep existing modeled ramp/recovery sources. Weekly target overlap validation is based on intersection with projected days 1..horizon, so a current-week Monday target is valid when `start_date` is mid-week and future dates remain in that week; targets with no overlap are rejected. Daily source label for filled dates will be `weekly_plan_targets`; `_meta.source_tools` always includes `get_fitness` and adds `get_training_plan` only when weekly targets are supplied. Assumptions will include target count, filled-day count, override count, ISO Monday anchor, partial-week/no-redistribution caveat, and even `training_load/7` distribution. `week_start_date` must be a valid Monday date after trimming, duplicate normalized week anchors are rejected, and `training_load` must be finite in `[0, 7*maxProjectionPlannedDailyLoad]`.
| 2026-06-10 11:58 | Review R001 | plan Step 1: REVISE |
| 2026-06-10 12:01 | Review R002 | plan Step 1: REVISE |
| 2026-06-10 12:04 | Review R003 | plan Step 1: APPROVE |
| 2026-06-10 12:09 | Review R004 | code Step 1: REVISE |
| 2026-06-10 12:13 | Review R005 | code Step 1: APPROVE |
| 2026-06-10 12:16 | Review R006 | plan Step 2: APPROVE |
| 2026-06-10 12:24 | Review R007 | code Step 2: REVISE |
| 2026-06-10 12:28 | Review R008 | code Step 2: REVISE |
| 2026-06-10 12:32 | Review R009 | code Step 2: APPROVE |
| 2026-06-10 12:33 | Review R010 | plan Step 3: APPROVE |
| 2026-06-10 12:36 | Review R011 | code Step 3: APPROVE |
