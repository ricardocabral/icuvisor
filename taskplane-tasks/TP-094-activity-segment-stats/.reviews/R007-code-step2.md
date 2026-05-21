# Code review — Step 2: Implement stream slicing and stats

**Verdict: Request changes**

The pure calculator covers the requested scalar and derived paths, and `go test ./...` passes. However, there is a blocker in the analyzer metadata contract: some valid-but-insufficient derived-stat results will report `result.insufficient_sample:true` while top-level analyzer `_meta.insufficient_sample:false`. That contradicts TP-094's mandatory analyzer meta behavior and should be fixed before proceeding.

## Findings

### 1. Blocker — `_meta.insufficient_sample` can disagree with the computed result

- **Where:** `internal/tools/compute_activity_segment_stats.go:82`; driven by `internal/analysis/segment_stats.go:203-211` and `:234-248`.
- **Problem:** `computeDrift` and `computeDecoupling` can mark `SegmentStatsResult.InsufficientSample = true` even when `N >= MinSamples`, e.g. when one elapsed-time half has fewer than two finite paired samples, or when positive denominator checks fail. But the tool passes only `N` and `MinSamples` into `analysis.NewAnalyzerMeta`, whose `InsufficientSample` is recomputed as `n < min_samples`. In those cases the returned payload will have contradictory fields:
  - `result.insufficient_sample: true`
  - `_meta.insufficient_sample: false`
- **Example:** a drift segment with five finite HR samples distributed four in the first half and one in the second half yields `N=5`, `MinSamples=4`, and `computed.InsufficientSample=true`; `_meta.insufficient_sample` will be false because `5 < 4` is false.
- **Why this matters:** Step 2 explicitly requires valid slices with too few usable samples to return analyzer payloads with `_meta.insufficient_sample:true`; the analyzer `_meta` is the mandatory contract clients consume.
- **Suggested fix:** Preserve the calculator's boolean in analyzer meta instead of recomputing solely from `N/MinSamples` for this tool. Options include extending `AnalyzerMetaInput` with an optional explicit insufficient flag, or mapping derived-stat per-half/denominator failures into metadata in a way that forces `_meta.insufficient_sample` to match `computed.InsufficientSample`. Add tests for uneven split-half samples and zero/negative denominator cases through the tool response, not just the pure calculator result.

### 2. Medium — `include_full:true` has no audit series for decoupling

- **Where:** `internal/analysis/segment_stats.go:250-253`.
- **Problem:** `computeDecoupling` returns aggregate `Details` but never sets `result.Audit`. Since the tool's full-response path exposes `computed.Audit` as `series`, `include_full:true` for decoupling does not include the sliced paired HR/power inputs or calculation points needed to audit the stat, unlike scalar, drift, and NP.
- **Why this matters:** The Step 1/2 contract says full mode may expose only the sliced canonical audit inputs/calculation points needed to audit the stat. Returning no series for one derived stat makes full mode inconsistent and less auditable.
- **Suggested fix:** Populate a decoupling audit series, for example first/second-half HR and watts slices (or paired first/second calculation points), while still excluding unrelated raw streams.

### 3. Medium — invalid request arguments can trigger an upstream stream fetch before local validation

- **Where:** `internal/tools/compute_activity_segment_stats.go:58-73`, with validation deferred to `analysis.ComputeActivitySegmentStats` at `:73`.
- **Problem:** `activitySegmentStatsInput` validates selector presence and required stream keys, but it does not validate numeric bounds or `ftp_watts` compatibility. Requests such as `stat=mean` with `ftp_watts`, negative/inverted ranges, or non-finite pure inputs will call `GetActivityStreams` before being rejected by the calculator.
- **Why this matters:** These are deterministic argument errors and should be rejected locally without I/O; otherwise users may see an upstream failure instead of the intended invalid-arguments error, and the raw-stream exception does unnecessary work for malformed requests.
- **Suggested fix:** validate bounds and `ftp_watts` compatibility before fetching streams, reusing/exporting the calculator validation where practical so tool and pure API behavior stay aligned.

## Verification performed

- `git diff b93bbe1..HEAD --name-only`
- `git diff b93bbe1..HEAD`
- `go test ./internal/analysis ./internal/tools` — passed
- `go test ./...` — passed
- `golangci-lint run ./internal/analysis ./internal/tools` — failed. Most failures are expected until Step 3 registers the new tool, but `internal/analysis/segment_stats.go:493` (`round`'s `precision` parameter always receives `6`) will remain an `unparam` issue unless adjusted.
