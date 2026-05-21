# R004 Code Review — Step 1: Define histogram contract

**Verdict:** REQUEST CHANGES

Step 1 is much closer to an implementable contract, but I found one blocking contract mismatch against the existing stream canonicalization. If implemented as written, power histograms can be treated as unavailable even when power data exists.

## Blocking finding

1. **Power required stream key uses the wrong canonical key.**
   - Location: `taskplane-tasks/TP-092-activity-histogram/STATUS.md:136`
   - The contract says power requires canonical stream keys `power` and `time`, but this repository canonicalizes upstream `Power`, `Watts`, and `power` to `watts` (`internal/streams/canonicalizer.go:27`, `:35`, `:61`, `:75`). `get_activity_streams` responses are keyed by that canonical value.
   - Please change the Step 1 contract to require `watts` + `time` for `power_watts`, while keeping `power` only as an accepted input alias if desired. Otherwise Step 2 can legitimately implement `streams["power"]` lookups per the contract and fail all real power histograms.

## Non-blocking notes

- `STATUS.md:124` says Step 1 will extend `internal/analysis` with `pace_seconds_per_km`, but that metric already exists in the catalog as an activity-row metric. To avoid a duplicate-catalog implementation in Step 2, consider wording this as: add missing stream-grain metrics (`power_watts`, `heart_rate_bpm`) and define a histogram-specific allowed subset that includes the existing `pace_seconds_per_km`.
- The execution-log rows for R001/R002 are currently appended under `## Notes` rather than in the `## Execution Log` table. This is just STATUS hygiene, but it is worth moving while touching the file.

Tests were not run; this step only changes task/status documentation.
