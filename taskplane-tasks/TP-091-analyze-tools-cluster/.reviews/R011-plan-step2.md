# R011 Plan Review — Step 2: Implement computations

**Verdict:** APPROVE

The hydrated Step 2 plan now gives enough implementation detail to proceed. It names the analysis/tool helper split, keeps catalog registration and generated docs in Step 3, pins source-selection behavior, commits to typed field extraction instead of reflection, and records the deterministic numeric choices needed for golden tests.

## Non-blocking notes

1. **Align helper filenames with the task file scope or record the scope expansion.**

   `PROMPT.md:50-63` scopes analysis changes to `trend*.go`, `distribution*.go`, `correlation*.go`, and `efforts*.go`, while the plan proposes `internal/analysis/window.go` and `internal/analysis/samples.go` (`STATUS.md:161`). Those helpers are sensible, but before coding either rename them to fit the approved glob (for example `trend_window.go` / `trend_samples.go` if shared) or explicitly log the scope expansion in `STATUS.md` so the later code review is not surprised by out-of-scope files.

2. **Make rolling-window missing-data behavior explicit in code/tests.**

   The plan says rolling means are trailing windows aligned to the sample date/bucket and omitted until enough preceding samples exist (`STATUS.md:165`). That appears to mean “last N usable samples/buckets,” not “calendar-day span with skipped missing dates.” That is acceptable if intentional, but the implementation and fixtures should make this obvious because the public argument is named `rolling_window_days` and missing dates are skipped rather than imputed.

3. **Define the activity full-window loop guard in implementation.**

   The plan correctly says activity loaders repeatedly call `fetchActivitiesPage` with `include_unnamed=true` and page size 200 until no token, aborting on boundary/leftover tokens (`STATUS.md:163`). When implementing, keep a clear upper bound or token-progress guard around that outer loop so a bad cursor/token cannot spin indefinitely, and surface the promised `activity window too large; narrow date range` error rather than partial results.

## Tests

Not run; reviewed `PROMPT.md`, `STATUS.md`, prior R010 feedback, and relevant existing `internal/analysis`/`internal/tools` source surfaces.
