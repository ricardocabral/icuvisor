# R005 Code Review — Step 1: Design request/response contracts

**Verdict:** Changes requested

The Step 1 notes are much more complete than the original plan and resolve most of the earlier schema/source/meta concerns. I found one remaining public-contract gap that should be settled before Step 2, because it affects metric compatibility, response grain, `_meta.n`, and golden fixtures for PRD-listed metrics.

## Blocking finding

1. **Derived weekly metrics are still incompatible with the “always daily” analyzer contract.**

   `weekly_tss` and `weekly_hours` are canonical `analysis_metric` values, and PRD §7.2.C explicitly lists weekly TSS as an `analyze_trend` use case. The Step 1 notes currently say derived weekly metrics use `FitnessClient.ListAthleteSummary` weekly buckets (`STATUS.md:131`), but then say `analyze_trend` always computes over athlete-local daily samples and reports missing days as inclusive days minus sampled days (`STATUS.md:135`). Those two rules cannot both hold for weekly-derived metrics.

   Before implementation, please define the contract for `SourceDerivedWeekly` metrics across the affected analyzers:

   - whether `analyze_trend` accepts weekly-derived metrics as weekly samples, expands them to daily samples, or rejects them despite the PRD example;
   - how weekly bucket boundaries are aligned to an arbitrary inclusive `window`/`baseline_window`;
   - what `_meta.n`, `_meta.missing_days` (or `missing_weeks`/assumption fields), `sample_grain`, and `aggregation` mean for weekly-derived metrics;
   - how `rolling_window_days` applies to weekly samples, or whether a separate rolling/bucket rule is used;
   - whether `analyze_distribution` and `analyze_correlation` support derived weekly metrics, and if not, what short user-facing rejection hint is returned.

   Without this, Step 2 will either invent behavior in code or accidentally compute misleading missing-day counts such as “28 expected days minus 4 weekly buckets,” and tests will lock in an ambiguous public API.

## Non-blocking notes

- The older Step 1 prose still says `rolling_window_days` is clipped (`STATUS.md:122`) while the later R002 section says it is rejected (`STATUS.md:133`). The later rule is clear enough to implement, but cleaning up the stale sentence would avoid future reviewer/implementer confusion.
- The review-history rows at `STATUS.md:136-139` are currently appended under `## Notes` rather than the `## Reviews` or `## Execution Log` tables. This is process hygiene rather than a contract blocker.
