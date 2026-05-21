# Plan Review — Step 2: Switch to `math/rand/v2`

Result: **approved with one required guardrail**.

The step is correctly scoped to replacing only the entropy source in `addJitter`; the backoff base, span calculation, `RetryConfig`, and retry decision flow should stay unchanged.

## Required guardrail

Use the concurrency-safe top-level function:

```go
import rand "math/rand/v2"

offset := rand.Int64N(2*span+1) - span
```

Do **not** introduce an unsynchronized package-local `*rand.Rand`. The `math/rand/v2` docs say `Rand` and `Source` should be used by a single goroutine at a time; only the top-level functions are safe for concurrent use. A `sync.Once` or package `var` initializer only makes construction one-time, not calls to `pkgRand.Int64N` concurrency-safe. If a package-local `*rand.Rand` is used despite the simpler option, it must be protected by synchronization, but that is unnecessary for this task.

## Implementation notes

- Keep the existing guards before calling `Int64N`: `delay <= 0`, `ratio <= 0`, and `span <= 0` must still return `delay`, because `Int64N` panics for non-positive bounds.
- Consider computing the bound separately and guarding overflow defensively before `Int64N`, e.g. return `delay` if `2*span+1` would be non-positive. This avoids introducing a panic path for extreme caller-provided jitter ratios/delays.
- Do not add `math/rand` v1, third-party RNGs, or test hooks/seeding unless a later test explicitly requires deterministic control.
- Keep the new statistical test unchanged except for any necessary range assertion cleanup; it should go green after this change.
- Update `STATUS.md` after the code change, and leave `CHANGELOG.md` for the task documentation requirement if not already handled in the final verification step.

No blocker to proceeding with Step 2 under the top-level `math/rand/v2` approach.
