# R007 Code Review — Step 1: Design request/response contracts

**Verdict:** Changes requested

The R006 updates resolve the lag-window and weighted aggregation blockers well enough to proceed for trend/distribution/correlation. I found one remaining public-contract gap in `analyze_efforts_delta` that should be settled before Step 2, because it determines the response shape, units, delta semantics, `_meta.missing_days`, and golden fixtures.

## Blocking finding

1. **`analyze_efforts_delta` still does not define unit-explicit effort values or deterministic missing-day meta.**

   The request contract separates power/HR duration buckets from pace distance buckets (`STATUS.md:133`), but the response sketch only says per-bucket `current`, `baseline`, `absolute_delta`, and `percent_delta` (`STATUS.md:137`). For pace buckets, this leaves the public API ambiguous: is `current` the upstream curve `value`, elapsed seconds for the distance, `pace_seconds_per_km`, `pace_seconds_per_mile`, or a preferred-unit display value? It also does not say whether `absolute_delta` is lower-is-better elapsed/pace delta or a raw upstream-value delta. PRD §7.2.C calls this tool “unit-aware” for cases like “5k pace”, so the contract should name the unit-specific fields before implementation.

   Please define the per-family bucket row shape explicitly, for example:
   - power: `current_power_watts`, `baseline_power_watts`, `absolute_delta_watts`, `percent_delta`;
   - heart rate: `current_heart_rate_bpm`, `baseline_heart_rate_bpm`, `absolute_delta_bpm`, `percent_delta`;
   - pace: either elapsed-time fields (`current_elapsed_seconds`, `baseline_elapsed_seconds`, `absolute_delta_seconds`) plus athlete-preferred `pace_seconds_per_km`/`pace_seconds_per_mile`, or another explicit scheme, with `better_direction="lower"` if lower elapsed/pace is an improvement.

   The same section says `_meta.missing_days` for efforts “stays date-window missing days when known” (`STATUS.md:141`), but the chosen curve endpoints do not provide day-level completeness. Make this deterministic now, e.g. `_meta.missing_days=0` with `_meta.assumptions.missing_days_applicable=false`, while bucket availability is represented by per-row `current_missing`/`baseline_missing` and `_meta.assumptions.missing_buckets`.

## Non-blocking notes

- The R006 weighted average rule for HR/cadence (`STATUS.md:147`) is clear when all activities have usable moving time and when none do. The mixed case (some activities have HR/cadence but no usable moving time) should be covered in tests or tightened in prose so implementation does not accidentally diverge.
- The final review-history row is appended under `## Notes` as a bare table row (`STATUS.md:148`) rather than in `## Execution Log`; this is process hygiene, not a contract blocker.

## Tests

Not run; reviewed the Step 1 contract/status diff only.
