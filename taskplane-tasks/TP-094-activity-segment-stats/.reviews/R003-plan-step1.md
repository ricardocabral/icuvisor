# Plan review — Step 1: Define segment selection and metrics

**Verdict: Approved**

The revised Step 1 plan is now concrete enough to implement. It addresses the prior blockers: the schema is reduced to one `stat`, scalar metrics are restricted to canonical stream keys, `ftp_watts` is explicitly required only for IF, incompatible metric/FTP combinations are rejected, NP/IF avoid non-existent formula refs rather than reusing the wrong refs, elapsed-time windows are used for NP/IF, zero-watt samples are valid, and `_meta.n` now has stat-specific semantics.

## Non-blocking implementation notes

- Define the minimum sample thresholds in code/tests, especially for scalar `mean`/`median`/`p90` when all sliced samples are non-finite. A simple `min_samples=1` is probably enough for those stats, while the plan already defines stricter eligibility for split-half and NP/IF calculations.
- Make the analyzer meta explicit in the implementation: `source_tools` should include `get_activity_streams`, `missing_days` should be `0` for this within-activity analyzer, and `missing_action` should match the skip policy for non-finite/negative samples.
- When coding NP/IF, keep the rolling-window endpoint rule deterministic and test it. The plan implies windows ending at stream sample timestamps with `window_end-30 < t <= window_end`; tests should lock that down so future changes do not drift into interpolation or sample-count-based behavior.
- For `include_full:true`, return only the sliced canonical inputs/calculation points described in the plan, not unrelated streams or full upstream raw objects.

No further planning changes are required for Step 1.
