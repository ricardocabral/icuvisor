# Code Review — Step 3: Implement

**Decision: REVISE.**

The refactor is moving in the intended direction and the response test suite passes, but I found behavior changes in the new converter/walker that violate the Step 3 requirement to preserve the previous JSON-round-trip semantics and existing path predicates.

## Blocking findings

1. **JSON float semantics are no longer preserved for `NaN` / `Inf` (and `float32` byte output).**  
   In `internal/response/shaper.go:161-162`, the reflection converter returns float values directly. The old `json.Marshal` leg rejected non-finite floats (`unsupported value: NaN/+Inf`) before any shaped response escaped. The new path accepts them and returns a shaped value that cannot be encoded as JSON later, changing the shaper from an early, wrapped error into a latent output failure. The same direct `float32 -> float64` conversion can also change byte output (`encoding/json` formats float32 with 32-bit precision; converting to float64 first can expose extra digits). Please mirror `encoding/json` here: reject non-finite values and either preserve float32 formatting equivalence or route float32 through the narrow marshal fallback. Add tests for `math.NaN()` / `math.Inf(1)` and a representative `float32` value.

2. **`query_type` under provenance is now dropped even though the old walker preserved it.**  
   `dropDebugVisitor` (`internal/response/shaper.go:634-637`) drops any path matched by `isDebugPath`. The old `dropDebugMetadata` skipped deleting both `fetched_at` and `query_type` whenever the current map path matched `isProvenancePath`, so `_meta.provenance.<field>.query_type` was preserved. The new implementation only exempts provenance `fetched_at` via `isProvenanceFetchedAtPath` (`internal/response/shaper.go:617-623`) and drops provenance `query_type`. That is a path-predicate semantic change. Please make the drop visitor honor `isProvenancePath` equivalently to the old map-level delete logic, and add a regression test with `_meta.provenance.foo.query_type`.

## Non-blocking note

- `isMetaPath` (`internal/response/shaper.go:687-689`) uses `strings.Contains(path, "._meta")`, which also matches path segments like `._metadata` / `._meta2`. The old scale collector skipped only keys exactly named `_meta`. Consider making this segment-exact while you are touching the walker predicates.

## Verification run

- `go test -count=1 ./internal/response` ✅
- `go test ./...` ✅
