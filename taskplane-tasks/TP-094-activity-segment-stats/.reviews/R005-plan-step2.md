# Plan review — Step 2: Implement stream slicing and stats

**Verdict: Changes requested**

The Step 2 plan is on the right track: keeping the calculator pure in `internal/analysis`, doing stream client work only in `internal/tools`, deriving the required canonical streams from the selected stat, and preserving the Step 1 semantics for elapsed-time halves and NP/IF windows are all good directions.

Before implementation, tighten a few deterministic behavior points so the analyzer does not ship with test-fragile or ambiguous math/error handling.

## Blocking issues

1. **Define the percentile/stat algorithms and scalar minimums explicitly.**  
   The plan says to compute `mean`, `median`, and `p90`, but it does not specify the percentile convention for `p90` or the scalar minimum sample threshold. Pick and document one deterministic method before coding (for example: sort finite values; median averages the two middle values for even `n`; p90 uses nearest-rank `ceil(0.90*n)` or a named interpolation method). Also state that scalar stats require at least one finite sliced sample and set `_meta.n`/`insufficient_sample` from that count. This should be locked with tests because different p90 conventions produce different answers on small segments.

2. **Make the error-vs-insufficient-result contract deterministic.**  
   Step 2 currently says missing/out-of-range/insufficient cases are reported as “actionable errors or result statuses,” which leaves too much discretion to the implementation. Align it with Step 1 before coding: invalid arguments, missing required streams, mismatched stream lengths, and segment ranges outside axis coverage should return short user-facing errors; insufficient samples after a valid slice should return an analyzer response with `_meta.insufficient_sample: true`, the appropriate `_meta.n`, and no fabricated numeric value. This distinction matters for MCP callers and for the Step 3 fixture expectations.

3. **Specify canonical stream fetch/mapping details.**  
   The plan lists required canonical streams, but the implementation also needs a precise boundary rule for the intervals client call: dedupe the required canonical keys, request only those streams (plus the selector axis), canonicalize returned `type`/`name` values through `internal/streams`, and fail if a required canonical key is absent. If the upstream `types` query needs API spellings that differ from canonical keys (notably `heart_rate`/`heartrate` or `velocity_smooth`/`VelocitySmooth`), define that mapping now rather than relying on default streams or broad fetches. This is necessary to satisfy the task’s “canonical streams only” and “raw-stream exception” constraints.

## Non-blocking implementation notes

- Keep `IncludeDefaults` behavior intentional: do not let the analyzer accidentally retrieve unrelated default streams when only a small required key set is needed.
- Add focused pure-analysis tests in Step 2 if possible, even if tool registration fixtures land in Step 3; the calculator math is the highest-risk part of this task.
- Preserve the Step 1 full-mode rule: `include_full:true` should expose only sliced canonical inputs/calculation points needed to audit the selected stat, never unrelated activity streams or upstream raw objects.
