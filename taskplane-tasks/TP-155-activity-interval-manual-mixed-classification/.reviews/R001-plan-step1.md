# Plan Review: Step 1 — Add classifier states and fixture coverage

**Verdict:** Changes requested

The step scope is right (`internal/analysis/interval_source.go` and `_test.go` only), but the plan needs one important clarification before implementation.

## Findings

1. **Guard against overclassifying missing raw evidence as `manual_added`.**
   The prompt explicitly says not to treat missing raw fields as manual. Step 1 should state that the `group_id` heuristic only applies when interval rows expose raw upstream evidence. Existing synthetic tests such as `genericDistanceIntervals` have nil/empty `Raw` maps and must keep their current `device_laps`/`unknown` behavior. Add a regression case for nil or empty raw maps remaining non-manual.

2. **Make the intended precedence explicit.**
   The implementation should document/test the order, e.g. explicit structured markers and explicit device-lap markers first, then the group-id manual/mixed heuristic, then the existing uniform auto-lap fallback only when group-id evidence is inconclusive/unavailable. This is the safest way to preserve the current structured/device behavior while adding `manual_added` and `mixed`.

3. **Test the actual upstream marker shape.**
   The new tests should use realistic interval `Raw` maps with `group_id` present/non-empty versus absent, rather than relying on `IntervalSourceInput.Groups` (which currently means structured group evidence). Include all required source outcomes: `structured_workout`, `device_laps`, `manual_added`, `mixed`, and `unknown`.

Once those points are reflected in the plan/tests, the step looks appropriately scoped.
