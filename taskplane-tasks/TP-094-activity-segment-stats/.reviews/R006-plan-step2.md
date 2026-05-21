# Plan review — Step 2: Implement stream slicing and stats

**Verdict: Approved**

The revised Step 2 plan is ready to implement. It addresses the prior R005 blockers by locking deterministic scalar math, separating user-facing errors from valid-but-insufficient analyzer results, and spelling out the canonical stream fetch/mapping boundary for the raw-stream exception.

## What is now sufficiently specified

- The calculator stays pure in `internal/analysis`, with intervals.icu I/O contained in `internal/tools`.
- Required streams are derived per `stat`, deduped, and fetched narrowly rather than relying on defaults or broad stream expansion.
- Returned stream rows are canonicalized through `internal/streams`, and absent required canonical keys are treated as errors.
- Time and distance selectors use canonical axis streams with inclusive bounds and normalized units.
- Scalar stats have stable algorithms: finite-only samples, `min_samples=1`, arithmetic mean, even-`n` median averaging, and nearest-rank p90 with clamping.
- Derived stats inherit the approved Step 1 semantics: elapsed-time split halves, per-half finite paired sample minimums, positive denominator checks, 30-second elapsed-time NP windows, finite non-negative watts with zero values included, and `ftp_watts` validation for IF.
- The error contract is deterministic: invalid arguments, missing streams, mismatched lengths, and out-of-coverage ranges return short actionable errors; valid slices with too few usable samples return an analyzer payload with `_meta.insufficient_sample:true`, `_meta.n`, and no fabricated numeric value.

## Non-blocking implementation notes

- Keep `IncludeDefaults` explicitly false/omitted for this analyzer path unless there is a documented reason to include defaults; tests should assert the exact requested `Types` set where practical.
- Add pure `internal/analysis` table tests during this step for the percentile convention, boundary inclusion, distance slicing, insufficient scalar samples, split-half minima, and NP rolling-window endpoint rule.
- Be explicit in code about how non-finite selector-axis samples are handled if they appear in fixtures; even if upstream JSON should not contain NaN/Inf, the pure calculator tests may exercise those edge cases.
- Preserve the Step 1 full-response constraint when wiring the tool: `include_full:true` may expose only sliced canonical audit inputs/calculation points, not unrelated streams or upstream raw objects.

No further planning changes are required before coding Step 2.
