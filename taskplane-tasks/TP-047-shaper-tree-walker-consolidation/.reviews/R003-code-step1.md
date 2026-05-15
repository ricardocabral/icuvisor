# Code Review — Step 1: Snapshot pre-refactor output

**Decision: Changes requested.**

The golden snapshot mechanism is close and the checked-in fixtures currently pass the direct golden test, but I found one test-isolation bug and one coverage gap that undercuts Step 1 as the safety net for the shaper refactor.

## Findings

### 1. Golden test leaks catalog runtime state into other tests

- **Severity:** Blocking
- **File:** `internal/response/shaper_test.go:484-512`

`TestShapeGoldenSnapshots` resets/sets the package-level catalog metadata inside each subtest, but never restores it after the final subtest. That leaves `catalogRuntime.current` as `golden-catalog-hash`, so later tests that rely on the default catalog hash become order-dependent.

Repro:

```sh
go test ./internal/response -shuffle=on -count=1
```

This failed locally with multiple mismatches where other tests received `catalog_hash:"golden-catalog-hash"` instead of `dev-catalog-hash`.

Please add cleanup around the golden test/subtests, for example `t.Cleanup(resetRuntimeCatalogMetadataForTest)` in `TestShapeGoldenSnapshots` (and/or per subtest with `defer resetRuntimeCatalogMetadataForTest()`), so the global catalog state is isolated regardless of test order, shuffle, or early returns in update mode.

### 2. Snapshot cases use only maps, so they do not lock the typed response contract

- **Severity:** Major
- **File:** `internal/response/shaper_test.go:538-590`

All five golden cases build inputs as `map[string]any` / `[]any`. Step 1 is meant to capture representative pre-refactor tool output before removing the marshal round-trip, and the checked status says the cases use synthetic DTO input. More importantly, the current shaper behavior for actual tool responses depends on JSON tags and `omitempty` during `marshalToJSONValue`; map-only fixtures bypass that entire contract.

This creates real blind spots for the upcoming refactor. For example, the `get_activities` fixture includes `"name":""` in a row (`shaper_test.go:541`), but the actual `getActivitiesRow.Name` field is tagged `json:"name,omitempty"`, so an equivalent typed response would omit that field before shaping. A refactor could break tag/omitempty behavior and these golden files would still pass.

Please make at least some representative golden inputs typed local DTO structs with JSON tags matching the tool response shapes, especially `get_activities` terse/full and `get_fitness`. Keep maps where they are intentionally exercising arbitrary nested metadata/provenance, but the snapshot set should cover the typed marshal/tag behavior that this refactor is about to change.

## Verification run

- `go test ./internal/response -run TestShapeGoldenSnapshots -count=1` — passes
- `go test ./internal/response -count=1` — passes
- `go test ./internal/response -shuffle=on -count=1` — fails due to leaked catalog runtime state
