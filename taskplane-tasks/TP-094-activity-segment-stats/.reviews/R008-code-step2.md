# Code review — Step 2: Implement stream slicing and stats

**Verdict: Request changes**

The R007 fixes are present: analyzer `_meta.insufficient_sample` now follows the calculator boolean, decoupling has an audit series, and deterministic argument validation happens before stream fetch. Targeted and full tests pass. I found one blocker in the required-stream selection that violates the Step 2 narrow-fetch contract for scalar distance segments, plus two medium contract issues around user-facing missing-stream errors and no-sample ranges.

## Findings

### 1. Blocker — scalar distance stats fetch and require `time` even though it is not used

- **Where:** `internal/analysis/segment_stats.go:90-93`, used by `internal/tools/compute_activity_segment_stats.go:58-62`.
- **Problem:** `RequiredSegmentStreamKeys` unconditionally appends `time` whenever the selector axis is `distance`. That is correct for derived stats that split or roll over elapsed time, but it is overbroad for scalar stats. For example, `mean`/`median`/`p90` of `watts` over `start_distance_m`/`end_distance_m` should require only `distance` + `watts`; the implementation requires and fetches `distance` + `time` + `watts`, and will fail if `time` is absent.
- **Why this matters:** Step 2 explicitly says scalar stats fetch “selector axis plus requested metric” and that the raw-stream exception must fetch canonical stream keys only for the requested calculation. This bug both broadens the upstream fetch and turns an otherwise computable scalar distance request into a missing-stream error.
- **Suggested fix:** Move the extra `time` requirement into the derived-stat branches (`drift`, `decoupling`, `np`, `if`) or otherwise add it only when the metric itself is `time`. Add a test like `RequiredSegmentStreamKeys(mean, watts, distance) == [distance, watts]`, plus a handler test asserting the upstream `Types` omit `time` for scalar distance stats.

### 2. Medium — missing required streams lose the actionable stream name in the public error

- **Where:** `internal/tools/compute_activity_segment_stats.go:69-72`; `canonicalActivitySegmentStreams` returns `ErrMissingSegmentStream` with the key at `:133-136`, but the handler wraps it with the generic `computeActivitySegmentStatsMessage`.
- **Problem:** If the upstream activity lacks a required stream, the public LLM-visible error is only `could not compute activity segment stats`. The actual missing key is only in the wrapped internal cause, and `UserError.Error()` intentionally hides it.
- **Why this matters:** The task completion criteria require missing streams to produce short actionable errors/meta. A generic compute failure does not tell the model/user whether they should choose another metric/stat or inspect available streams.
- **Suggested fix:** Detect `errors.Is(err, analysis.ErrMissingSegmentStream)` on this path and return a short public message that says a required activity stream is missing, ideally including the canonical key from the error or from a typed missing-stream error.

### 3. Medium — in-coverage ranges with no samples are reported as out-of-range errors instead of insufficient samples

- **Where:** `internal/analysis/segment_stats.go:323-346`.
- **Problem:** `segmentIndices` returns `ErrSegmentOutOfRange` whenever `len(indices) == 0`, even if the requested bounds are inside the axis coverage. For example, a time stream covering `0..20` with samples at `0,10,20` and a request for `5..6` is not outside stream coverage, but it returns `activity segment range outside stream coverage`.
- **Why this matters:** Step 2 distinguishes out-of-coverage ranges (errors) from valid slices with too few usable samples (an analyzer payload with `_meta.insufficient_sample:true`, `_meta.n` set, and no fabricated value). This implementation collapses an in-coverage/no-sample segment into the out-of-range error path and produces an inaccurate message.
- **Suggested fix:** Have `segmentIndices` reserve `ErrSegmentOutOfRange` for `bounds.Start < minAxis || bounds.End > maxAxis` (or no finite axis coverage). For in-coverage zero-index segments, return an empty index slice and let the stat calculators produce the insufficient-sample result, or add a dedicated insufficiency path that preserves analyzer meta.

## Verification performed

- `git diff b93bbe1..HEAD --name-only`
- `git diff b93bbe1..HEAD`
- `go test ./internal/analysis ./internal/tools` — passed
- `go test ./...` — passed
- `golangci-lint run ./internal/analysis ./internal/tools` — failed only on the expected Step 3 registration-related unused functions: `newComputeActivitySegmentStatsTool` and `computeActivitySegmentStatsInputSchema`.
