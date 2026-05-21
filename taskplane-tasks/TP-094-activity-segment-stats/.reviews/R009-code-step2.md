# Code review — Step 2: Implement stream slicing and stats

**Verdict: Request changes**

The R008 fixes are in place: scalar distance stats now fetch narrowly, missing stream errors include the canonical key, and in-coverage/no-sample slices become insufficient analyzer results. `go test ./...` passes. I found two remaining validation/error-contract gaps that should be addressed before closing Step 2.

## Findings

### 1. Medium — explicit `ftp_watts: 0` is accepted for non-IF stats

- **Where:** `internal/tools/compute_activity_segment_stats.go:94-100`; `internal/analysis/segment_stats.go:150-154`.
- **Problem:** The request type preserves whether `ftp_watts` was provided with `*float64`, but `activitySegmentStatsInput` collapses nil and explicit zero to the same `0.0` value before validation. As a result, requests like `{"stat":"mean","metric":"watts",...,"ftp_watts":0}` are treated as valid and will fetch streams, even though the Step 1/2 contract says `ftp_watts` is accepted only for `stat=if` and rejected for every other stat.
- **Why this matters:** This is a deterministic argument error that should be rejected locally without I/O. It also makes the server-side validation weaker than the schema/STATUS contract and can silently ignore a caller-provided field.
- **Suggested fix:** Validate the pointer-level request semantics in the tool layer before converting to `analysis.SegmentStatsInput`: require `args.FTPWatts != nil` and positive/finite for `stat=if`, and reject `args.FTPWatts != nil` for all other stats. Keep the pure analysis API's `FTPWatts == 0` behavior as the absent/default case if desired. Add a handler test that `stat=mean` with `ftp_watts:0` returns the invalid-arguments user error and does not call `GetActivityStreams`.

### 2. Medium — out-of-coverage ranges lose the actionable public error

- **Where:** `internal/tools/compute_activity_segment_stats.go:76-82`; `internal/analysis/segment_stats.go:361-363`.
- **Problem:** `segmentIndices` correctly returns `ErrSegmentOutOfRange` with coverage details, but the handler maps that to the generic public message `invalid compute_activity_segment_stats arguments; provide activity_id, one stat, exactly one time or distance range...`. `UserError.Error()` hides the wrapped cause, so the LLM/user never sees that the requested segment was outside stream coverage.
- **Why this matters:** The task requires invalid ranges to produce short actionable errors/meta. A generic shape hint is misleading when the request already had the right fields but the selected bounds were outside the activity's streams; it does not tell the caller to choose a segment within the available time/distance coverage.
- **Suggested fix:** Add a specific public message for `errors.Is(err, analysis.ErrSegmentOutOfRange)`, e.g. `activity segment range is outside available stream coverage`, optionally including the coverage range if you keep it concise and safe. Add a handler-level test so this remains visible at the MCP boundary.

## Verification performed

- `git diff b93bbe1..HEAD --name-only`
- `git diff b93bbe1..HEAD`
- `go test ./internal/analysis ./internal/tools` — passed
- `go test ./...` — passed
- `golangci-lint run ./internal/analysis ./internal/tools` — fails only on the expected Step 3 registration-related unused functions: `newComputeActivitySegmentStatsTool` and `computeActivitySegmentStatsInputSchema`.
