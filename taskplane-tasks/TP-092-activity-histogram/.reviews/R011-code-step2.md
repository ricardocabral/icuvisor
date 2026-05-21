# R011 Code Review — Step 2: Implement stream-backed histogram

**Verdict:** REVISE

The core histogram construction mostly follows the approved Step 2 plan, and `go test ./...` passes. I found one behavioral blocker in pace interval filtering plus a current lint/quality-gate blocker from the unregistered tool code.

## Blocking findings

1. **Pace histograms include intervals with negative distance deltas.**  
   `internal/tools/get_activity_histogram.go:215-222` computes pace for every adjacent distance/time pair and relies on `BuildHistogram` to drop invalid samples later. That drops non-positive/invalid time and `dd == 0` via `Inf`, but it does **not** drop `dd < 0`: a backwards/noisy distance sample produces a finite negative pace with positive seconds, so it contributes to buckets and can skew fixed-width edges and percentages. The Step 1 contract explicitly says pace intervals with non-positive time or distance deltas are skipped. Fix by checking `dt` and `dd` are finite and `> 0` before dividing/appending in `paceSamples`.

2. **Current branch fails lint because the new tool is entirely unreferenced.**  
   `golangci-lint run ./...` fails with 22 `unused` issues starting at `internal/tools/get_activity_histogram.go:17` (`getActivityHistogramName`, constructor, handler, helpers, etc.). This may be resolved naturally by Step 3 registration, but as the Step 2 code currently stands it cannot pass the repository quality gate at a step boundary. Either register/wire the constructor in the same coherent change or defer adding the unreferenced tool file until the registration step.

## Non-blocking notes

- The approved Step 2 plan said stream interval construction/duration weighting should stay in `internal/analysis`; the current `valueTimeSamples` and `paceSamples` live in `internal/tools`. This is still testable from tool-package tests, but moving that pure math into `analysis` would better preserve the planned package boundary.
- `stream_fetch_failed` unavailable responses currently omit `_meta.emitted_unit`; missing-stream and insufficient-sample paths set it. Consider setting `emittedUnit` before the stream fetch so pace unavailable responses are consistent.

## Verification

- `go test ./...` — passes.
- `golangci-lint run ./...` — fails with unused issues for the unregistered histogram tool code.
