# Plan Review: Step 1 — Design deterministic weekly-target distribution

**Verdict:** Changes requested

The task direction is sound, but the Step 1 plan is still too underspecified to implement as failing tests. Before writing tests, capture the exact deterministic contract the tests will encode.

## Blocking gaps

1. **No concrete weekly-target shape is defined.** Step 1 should name the planned input shape that `get_fitness_projection` will eventually accept, e.g. weekly target entries with a week anchor/date and `training_load`/TSS value. Also decide whether this is manually supplied from `get_training_plan` output or extracted by a helper later; do not imply `get_fitness_projection` fetches plans implicitly.

2. **Week anchoring and partial weeks are ambiguous.** Define whether target weeks are athlete-local ISO/Monday weeks (or another Intervals convention), whether `start_date` itself is excluded from projection loads as current code does, and how to handle horizons that start/end mid-week.

3. **Override semantics need a precise formula.** “Explicit daily loads win” can mean either:
   - weekly target creates `target/7` candidate loads and explicit days replace those values, so weekly total may differ; or
   - explicit days consume part of the weekly target and the remainder is redistributed over missing days.

   Pick one. The simpler deterministic rule appears to be candidate `weekly_target/7` for each future day in that week, then exact-date `planned_daily_loads` replace candidates. Document the caveat in `_meta.assumptions`.

4. **Fallback behavior is not specified.** Tests should cover days with no weekly target and no explicit load continuing to use existing modeled ramp/recovery behavior, not zero load.

## Test plan adjustments

- Add an `internal/analysis/fitness_projection_test.go` table covering:
  - weekly target fills missing future days with deterministic daily loads;
  - explicit `planned_daily_loads` override exact dates;
  - uncovered dates retain `modeled_ramp`/`modeled_recovery_week` sources;
  - partial-week horizon/start-date exclusion behavior.
- Add tool-level tests that assert request decoding/schema-facing behavior and `_meta.assumptions`/`source_tools` include the training-plan-target bridge when weekly targets are supplied.
- Include validation cases for duplicate weekly target weeks, negative/out-of-range targets, invalid dates, and targets outside the projection horizon if those will be rejected or ignored.

## Metadata expectations

Plan for `_meta.assumptions` to include the distribution rule, target count, filled-day count, override count, week anchor convention, and a partial-week caveat. When weekly targets are provided, `_meta.source_tools` should include `get_training_plan` in addition to `get_fitness`.

Once these decisions are written into STATUS.md or the Step 1 test names/fixtures, the implementation plan will be reviewable and the failing tests will have a stable contract.
