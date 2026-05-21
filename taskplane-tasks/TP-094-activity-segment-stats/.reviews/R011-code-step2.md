# Code review — Step 2: Implement stream slicing and stats

**Verdict: Request changes**

The fixes from R010 are present: public segment fields are now JSON-tagged, and time-selected drift/decoupling split at the requested elapsed midpoint. The targeted and full test suites pass. I found one remaining NP/IF endpoint issue and one finite-time handling gap in the pure calculator that should be fixed before closing Step 2.

## Findings

### 1. Medium — NP/IF rolling-window eligibility is anchored to the first selected sample, not the requested segment start

- **Where:** `internal/analysis/segment_stats.go:281-283` and `:435-455`.
- **Problem:** `computeNP` does not pass the requested bounds into `rollingPowerWindows`, so the rolling-window code uses `times[indices[0]]` as the segment start guard (`endTime-times[indices[0]] < 30`). This adds an extra requirement that the first included sample be at least 30 seconds before the window end. Step 1/2 define the deterministic window by elapsed time (`window_end-30 < t <= window_end`) within the requested segment, not by the gap from the first sampled point. For a time-selected, in-coverage segment such as `start_seconds=5,end_seconds=60` with selected samples at `40s` and `60s`, the 60s window `(30,60]` is inside the requested segment and has finite watt samples, but the current code reports no valid window because `60-40 < 30`.
- **Why this matters:** This can turn a valid NP/IF segment into `_meta.insufficient_sample:true` and can undercount `_meta.n` depending on sampling gaps near the requested segment start. That violates the task's deterministic endpoint rule for NP/IF and makes results depend on sparse edge sampling rather than the caller's selected segment.
- **Suggested fix:** Pass `SegmentBounds` into `computeNP`/`computeIF`/`rollingPowerWindows` and use the requested elapsed start for time-selected segments when deciding whether a 30-second window is inside the segment. For distance-selected segments, either derive/interpolate elapsed start/end from the distance/time streams or explicitly test and document the discrete sample-endpoint behavior. Add pure tests for sparse edge samples around the 30-second window boundary.

### 2. Medium — distance-selected derived stats do not reject or skip non-finite time samples

- **Where:** `internal/analysis/segment_stats.go:367-414` and `:435-455`.
- **Problem:** When the selector axis is `distance`, `segmentIndices` validates finite distance values, but the derived calculations use the `time` stream without finite checks. In `splitHalfPairs`/`splitHalfIndices`, a selected sample with `times[idx] = NaN` falls into the second half because `NaN < mid` is false; if the first or last selected time is non-finite, `splitHalfTimeEndpoints` can make the midpoint non-finite and push every valid paired sample into the second half. In `rollingPowerWindows`, inner samples with non-finite timestamps are not skipped because both comparisons against `NaN` are false, so their watts can be included in every eligible window.
- **Why this matters:** Derived stats are defined over elapsed-time halves/windows. A non-finite elapsed timestamp is not a valid sample for those calculations and should not influence split-half means or NP/IF windows. The pure calculator currently can produce numeric values or inflated windows from invalid time samples instead of returning an insufficient result.
- **Suggested fix:** Require finite `times[idx]` wherever a selected sample participates in split-half or rolling-window calculations, and ensure distance-selected endpoints are finite before computing a midpoint. Add pure analysis tests with distance-selected segments containing `math.NaN()`/`math.Inf()` in the time stream.

## Verification performed

- `git diff b93bbe1..HEAD --name-only`
- `git diff b93bbe1..HEAD`
- `go test ./internal/analysis ./internal/tools` — passed
- `go test ./...` — passed
- `golangci-lint run ./internal/analysis ./internal/tools` — fails on the expected Step 3 registration-related unused functions: `newComputeActivitySegmentStatsTool` and `computeActivitySegmentStatsInputSchema`.
