# R010 Plan Review — Step 2: Implement computations

**Verdict:** Revise

The Step 1 public contract is now strong enough, but the Step 2 implementation plan in `STATUS.md:40-46` is still only the original outcome checklist plus one broad loader bullet. For a review-level 3 task adding four analyzer computations, that is not enough to review before code: the plan does not pin the deterministic math variants, source-loader shape, or implementation sequencing that will determine public results and golden fixtures.

## Blocking findings

1. **Step 2 needs a concrete implementation plan, not just outcome bullets.**

   `STATUS.md:43-46` says to implement source loading, validation, math, missing-day counts, and terse behavior, but it does not say where the reusable computation code will live, which helper types/functions will be added under `internal/analysis`, how the four tool files will call those helpers, or what order will keep the diff reviewable. Please hydrate Step 2 with a short package/file-level plan, for example:

   - shared sample/window structs and validation helpers;
   - source-loader interfaces and metric-source selection;
   - trend/distribution/correlation/efforts computation helpers;
   - tool adapter responsibilities versus pure math responsibilities;
   - which parts are deferred to Step 3 registration versus implemented now.

2. **Deterministic numeric semantics are still not pinned for the computations Step 2 will implement.**

   The Step 1 notes define user-visible fields, but Step 2 has to choose exact formulas. Before implementation, record these choices so code and golden tests do not accidentally lock in ad hoc behavior:

   - rolling mean alignment and early-window behavior;
   - trend OLS x-axis (`0..n-1`, calendar-day offsets, or week-bucket indexes) and zero-variance handling;
   - percent-delta denominator behavior when baseline is zero;
   - z-score variance convention and zero-stddev behavior;
   - distribution standard deviation convention (sample vs population), quantile interpolation method, bucket behavior for out-of-range explicit buckets, and float rounding policy;
   - Pearson zero-variance handling;
   - Spearman tie handling/ranking method;
   - coefficient strength/direction thresholds.

   R009 already called out formula variants as an implementation caution; because this step is specifically “Implement computations,” these choices should be written into the Step 2 plan before coding.

3. **The shared loader plan is too high-level for the metric/source complexity.**

   `STATUS.md:144-152` defines source behavior conceptually, but the Step 2 plan should make the implementation path explicit enough to avoid inconsistent source selection across tools. Please specify how the loader will:

   - choose among multiple sources for the same metric deterministically (for example `ctl`/`atl`, activity vs training-summary metrics);
   - map source rows to metric values without reflection/stringly field drift;
   - enforce unsupported interval/extended-source rejections with the promised short hints;
   - page `get_activities` to completion and surface the boundary error without partial results;
   - compute expected/missing counts for daily, weekly, and activity-grain samples;
   - acquire athlete-local timezone and unit preferences needed for dates and pace efforts.

   Without this, two analyzers could satisfy the same Step 1 contract differently while still appearing to complete the broad checklist.

## Non-blocking notes

- Keep the Step 1 hygiene noted in R009 on the radar: add the R009 review row and move the stray execution-log rows currently under `## Notes` (`STATUS.md:155-156`). This does not block the Step 2 plan, but it should be cleaned up before delivery.
- The plan should explicitly say that Step 2 remains pure computation/source loading; catalog registration, public descriptions, generated docs, and CHANGELOG can stay in later steps as already scheduled.

## Tests

Not run; reviewed `PROMPT.md`, `STATUS.md`, prior reviews, and existing `internal/analysis` skeletons only.
