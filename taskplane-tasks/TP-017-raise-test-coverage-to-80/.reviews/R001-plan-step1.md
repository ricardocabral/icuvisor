# Review R001 — Plan review for Step 1

**Verdict:** Approved

## Summary

The selected target list is appropriate for the task. It prioritizes behavior-rich, currently uncovered production paths in `internal/intervals` and then moves to low-covered `get_extended_metrics` branches before considering deterministic `toolchecks` guards as fallback. This matches the prompt's intent to raise coverage through regression-value tests rather than trivial command/main coverage or production rewrites.

## Checks performed

- Read `PROMPT.md` and `STATUS.md`.
- Confirmed the checked-in `coverage.out` baseline still reports **76.9%**.
- Spot-checked the per-function coverage for the proposed targets:
  - `internal/intervals/activity_streams.go`: 22 uncovered statements.
  - `internal/intervals/wellness.go`: 77 uncovered statements.
  - `internal/tools/get_extended_metrics.go`: 51 uncovered statements remaining.
  - `internal/toolchecks` has additional deterministic fallback coverage available.

## Notes / recommendations

- The coverage math supports the plan, but Step 2 alone is unlikely to be enough: activity-streams plus wellness have a maximum gain of about **99 statements**, while the repo needs about **107 additional covered statements** to cross 80.0% from the current profile. Plan to include at least a small set of `get_extended_metrics` tests before expecting the final coverage gate to pass.
- Keep the fallback `internal/toolchecks` work genuinely fallback-only. The intervals and extended-metrics tests carry higher product regression value and should be exhausted first.
- For the intervals HTTP-client cases, prefer the existing `httptest`/client helper patterns and assert request paths/query strings plus wrapped error behavior; do not add network-dependent tests.
- For `get_extended_metrics`, the highest-value cases appear to be the Strava-blocked response, optional `ErrNotFound`/`ErrUnauthorized` handling, non-optional error propagation, and `include_full` raw payload inclusion.

No blocking issues found with the Step 1 plan.
