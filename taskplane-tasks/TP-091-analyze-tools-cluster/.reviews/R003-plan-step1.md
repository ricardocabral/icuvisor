# R003 Plan Review — Step 1: Design request/response contracts

**Verdict:** Revise

The Step 1 notes now resolve the R002 blockers around correlation result fields, sample-size rules, efforts bucket validation, and rolling-window rejection. The plan is close, but I would not start Step 2 yet: a few remaining public-contract gaps can change the request/response semantics and golden fixtures.

## Blocking findings

1. **Trend/distribution sample grain and `missing_days` semantics are still inconsistent for activity-row metrics.**
   The common rule says `missing_days = expected days - usable values`, but the source mapping also says activity-only metrics use `ActivitiesClient.ListActivities`, and the response sketch allows `series` to be “daily or activity values.” That breaks for activity metrics with zero, one, or multiple activities on a day: `usable values` is not comparable to expected days, and days without a run/ride are not necessarily missing data for a pace or cadence distribution. Before implementation, define per-tool/per-source semantics for:
   - whether `analyze_trend` daily-aggregates activity metrics before slope/rolling calculations, or computes over per-activity samples;
   - what `_meta.n` counts for each grain;
   - when `_meta.missing_days` is meaningful vs `0`/not applicable for activity-grain distributions;
   - the exact `sample_grain`/aggregation method echoed in `_meta.assumptions`.

2. **`lag_days` is not defined for `pairing_grain=activity`.**
   The plan defines lag as “x on day D pairs with y on day D+lag,” which is clear for daily pairs. For activity-grain correlations, it is unclear whether lag is rejected, ignored unless zero, pairs same-activity fields only, or creates date-shifted many-to-many activity pairs. This is a user-visible interpretation issue. Please define one deterministic rule; the safest contract is likely `pairing_grain=activity` requires `lag_days=0`, with lagged correlations using daily pairing.

3. **Activity-backed source reads must specify full-window pagination/completeness.**
   The plan says activity-backed metrics use `ActivitiesClient.ListActivities`, but analyzer results require complete windows, not a single upstream/default page. Existing `get_activities` has cursor logic because activity listing can be bounded. Step 1 should state whether the analyzer loader reuses/implements full-window paging, what fetch limits/boundary errors look like, and how partial reads are surfaced. Otherwise `n`, missing counts, trends, histograms, and correlations can be silently computed on incomplete activity data.

## Non-blocking suggestions

- Add explicit auto-bucket behavior for `analyze_distribution` (`bucket_count`) when all sampled values are equal or when min/max are absent after skipping nulls.
- Echo the defaulted baseline-window formula in the notes (e.g. same inclusive length ending the day before `window.start_date`) so tests don’t drift.
- Include `sample_grain` and aggregation labels in every response sketch; they will help LLMs avoid over-describing per-activity samples as daily trends.

Once those grain/lag/completeness rules are captured in `STATUS.md`, the plan should be ready for Step 2 implementation.
