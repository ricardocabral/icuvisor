# Code Review — Step 3: Implement

**Decision: REVISE.**

The R010 fixes for non-finite floats, float32 formatting, and provenance `query_type` preservation look good, and the single-walker consolidation is still on the right track. I found one remaining JSON-round-trip semantic gap in the reflection converter that can change shaped output for supported stdlib JSON values.

## Blocking finding

1. **`json.Number` is converted to a JSON string instead of preserving `encoding/json` number semantics.**  
   `internal/response/shaper.go:153-158` handles every `reflect.String` value by returning `value.String()`, and `marshalSpecialValue` (`internal/response/shaper.go:299-311`) does not special-case `json.Number`. The old marshal/unmarshal path treats `json.Number("8.5")` as the JSON number `8.5` and rejects invalid numbers like `json.Number("bad")`. The new converter returns the string `"8.5"` and accepts `"bad"`, which violates the Step 3 requirement to preserve JSON-round-trip semantics and byte-identical output for the same inputs. This is especially relevant because `json.Number` already appears in internal generic-value handling/tests.

   Please route `json.Number` through the narrow `marshalJSONValue` fallback (or equivalent validation/number decoding) before the generic string case, and add regression coverage for both a valid number and an invalid number.

## Verification run

- `go test ./internal/response` ✅
- `go test ./...` ✅
