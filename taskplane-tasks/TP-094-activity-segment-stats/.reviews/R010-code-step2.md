# Code review — Step 2: Implement stream slicing and stats

**Verdict: Request changes**

The R009 fixes are present, and the calculator/tool tests pass. I found two remaining contract issues before Step 2 should be closed: the public segment object is emitted with Go field names, and split-half derived stats can split on observed sample endpoints rather than the requested elapsed segment.

## Findings

### 1. Medium — `segment` response uses `Axis`/`Start`/`End` instead of JSON contract names

- **Where:** `internal/analysis/segment_stats.go:65-69`, returned through `internal/tools/compute_activity_segment_stats.go:40` and `:86`.
- **Problem:** `analysis.SegmentBounds` has no `json` tags. Because `activitySegmentStatsResult.segment` embeds that type, the MCP payload will contain `{"Axis":"time","Start":0,"End":40}` after response shaping, not snake_case/lowercase fields such as `axis`, `start`, and `end`.
- **Why this matters:** This tool is introducing a public return shape, and project conventions rely on JSON tags for stable snake_case API fields. The request schema and task notes describe `start_seconds`/`end_seconds` or distance bounds; exposing PascalCase Go identifiers in the result is inconsistent and easy for clients/LLMs to misuse.
- **Suggested fix:** Add JSON tags to `SegmentBounds` (for example `axis`, `start`, `end`, or more explicit response fields if preferred) and add a handler/encoding test that inspects `payload["result"].segment` to lock the public shape.

### 2. Medium — split-half derived stats can use sample endpoints instead of the requested segment endpoints

- **Where:** `internal/analysis/segment_stats.go:367-409`, called by drift/decoupling at `:208` and `:236`.
- **Problem:** `splitHalfPairs` and `splitHalfIndices` compute the midpoint from `times[indices[0]]` and `times[indices[len(indices)-1]]`. For a time-selected request whose bounds are inside stream coverage but whose first/last included samples do not lie near the requested bounds, this shifts the midpoint away from the requested segment midpoint. Example: for a request `start_seconds=30,end_seconds=90` with samples only at `40,50,55,60,65` inside the segment, the requested elapsed midpoint is `60`, but the code splits at `(40+65)/2 = 52.5`, moving samples between halves and changing drift/decoupling.
- **Why this matters:** Step 2 says the stats are over the specified time or distance segment and Step 1 defines drift/decoupling over elapsed-time halves. Using observed sample endpoints makes the result depend on sampling gaps at the segment edges rather than the user-selected elapsed segment, and can change both the numeric value and insufficient-sample status.
- **Suggested fix:** For time-axis selections, split using `bounds.Start`/`bounds.End` as the elapsed-time endpoints. For distance-axis selections, either document/test the discrete sample-endpoint behavior or derive/interpolate elapsed endpoints consistently. Add a pure analysis test with sparse edge samples to lock the intended split.

## Verification performed

- `git diff b93bbe1..HEAD --name-only`
- `git diff b93bbe1..HEAD`
- `go test ./internal/analysis ./internal/tools` — passed
- `go test ./...` — passed
- `golangci-lint run ./internal/analysis ./internal/tools` — still fails only on Step 3 registration-related unused functions: `newComputeActivitySegmentStatsTool` and `computeActivitySegmentStatsInputSchema`.
