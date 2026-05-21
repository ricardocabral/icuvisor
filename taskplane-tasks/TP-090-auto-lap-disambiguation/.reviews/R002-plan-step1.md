# Plan Review: Step 1 — Model interval-source heuristics

## Verdict

Approved. The revised Step 1 plan is specific enough to drive implementation and later code/test review.

## What improved since R001

- The field inventory now matches the current DTO/shape path: typed interval/group fields plus preserved `Raw` maps are explicitly in scope, and the absence of known explicit interval-origin markers is recorded.
- The classifier contract is deterministic: source precedence, minimum usable row count, edge-row dropping, distance and duration targets, tolerances, invalid data handling, and contiguity requirements are documented.
- Structured-workout false positives are addressed by making `icu_groups`, workout-step raw markers, and non-generic names/types/labels take precedence over uniformity.
- Unit assumptions are explicit: interval distances are treated as meters at the response boundary, 1 mi is represented as `1609.344m`, and ambiguous/contradictory evidence falls back to `unknown`.
- Shared placement in `internal/analysis` with typed constants is planned, which should prevent string drift when analyzers propagate the signal.
- Metadata behavior is documented as additive, with `_meta.interval_source` and boolean `_meta.auto_lap_suspected` on successful interval responses and existing unavailable/Strava-blocked shapes preserved unless rows exist to evaluate.
- Acceptance examples cover structured, 1 km auto-lap, 1 mi auto-lap, insufficient/unknown, and structured-repeat negative cases.

## Non-blocking implementation notes

1. When implementing the edge-row drop rule, make the definition of "partial" objective in code/tests rather than relying on name matching alone. The current plan is acceptable, but Step 4 fixtures should lock this down.
2. Keep diagnostic reasons/confidence internal or test-only as planned; do not add extra public `_meta` fields unless the tool reference and schema snapshots are updated intentionally.
3. If an explicit upstream marker for manual/device laps is discovered during Step 2, preserve the plan's conservative precedence and decide explicitly whether it sets only `interval_source=device_laps` or also `auto_lap_suspected=true`; near-uniform rows should remain the main trigger for auto-lap suspicion.

## Summary

The Step 1 planning requirements are satisfied. Proceed to Step 2 using this STATUS.md contract as the source of truth for classifier implementation and fixture design.
