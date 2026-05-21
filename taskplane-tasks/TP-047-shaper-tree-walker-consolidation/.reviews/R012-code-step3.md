# Code Review — Step 3: Implement

**Decision: REVISE.**

The R010/R011 fixes are in place and `go test ./...` passes, but the new reflection converter still has behavior gaps versus the old `encoding/json` round-trip. Since Step 3's acceptance criteria require byte-identical output for the same inputs and wrapped failures for unsupported values, these should be fixed before moving on.

## Blocking findings

1. **Struct field selection/options still do not preserve `encoding/json` semantics.**  
   `structToJSONValue` / `jsonField` (`internal/response/shaper.go:232-278`) walks exported fields linearly and assigns `out[name] = item`. That differs from the old marshal/unmarshal path for standard JSON cases: duplicate/conflicting JSON field names are resolved by `encoding/json`'s dominance rules (for example two fields both tagged `json:"x"` are omitted, not last-writer-wins), and the `,string` tag option serializes numeric/bool fields as JSON strings. With the current converter those inputs produce different shaped output. Please either implement the relevant `encoding/json` field-resolution/tag semantics or route structs with conflicts / unsupported tag options through the narrow `marshalJSONValue` fallback. Add regression tests for at least duplicate tagged fields and a `json:",string"` numeric field.

2. **Recursive conversion no longer detects cycles; it can hang/crash instead of returning the old wrapped JSON error.**  
   `toJSONValue` recurses through maps, slices, pointers, and structs (`internal/response/shaper.go:119-230`) without a visited stack. A self-referential map/slice/pointer that previously failed through `json.Marshal` with an unsupported-cycle error will now recurse until stack exhaustion. That violates the Step 3 guardrail that unsupported values fail with a wrapped error comparable to the current marshal failure path. Please add cycle detection to the reflection path or fall back to `encoding/json` for container shapes where a cycle is detected, and cover this with a test that does not risk unbounded recursion.

## Non-blocking note

- `collectScaleLabels` now calls `walkJSON` (`internal/response/shaper.go:545-560`), and `walkJSON` always allocates cloned maps/slices even when the visitor only reads. That means every response, including large `include_full` payloads, gets an extra full-tree allocation pass just to collect scales. Consider a read-only walk mode/helper for scale collection so the refactor does not give back the allocation savings from removing the whole-response marshal round-trip.

## Verification run

- `go test ./internal/response` ✅
- `go test ./...` ✅
