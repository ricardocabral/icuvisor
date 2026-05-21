# Review R003 — Plan Review for Step 1

**Verdict:** Approved

The Step 1 measurement method in `STATUS.md` now defines a reproducible audit scope and resolves the issues from R001/R002. It names the fixture roots, applies path/type exclusions before broad object-shape detection, lists the precomputed zone-time fields, defines positive-total validation, gives per-family opportunity semantics, and correctly treats `fallback_count` as missing precomputed coverage that would require stream math rather than behavior the current tools already perform.

## Notes for Step 2

- Implement the exclusions as hard precedence rules so fixtures under events, wellness, workout library, custom items, activity messages, activity intervals, analyzer goldens, and schema snapshots cannot be reintroduced by `id`/date shape matching.
- For family applicability signals, make the script value-aware where possible: a present-but-null field such as `distance: null` or `moving_time: null` should not create a pace fallback opportunity.
- Preserve the plan's distinction between `fallback_count` and `unknown_count`; otherwise the current sparse activity fixtures will be easy to overstate as analyzer failures.
- Keep the threshold policy as recorded: no agreed numeric threshold; non-zero measured missing-precomputed opportunities are risky evidence for Step 3, not an automatic pass/fail gate.

No blocking changes are required before implementing/running the audit.
