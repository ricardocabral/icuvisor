# Code Review — Step 3

Verdict: **REVISE**

## Findings

1. **Hydration `field_semantics` is added for null fields that terse shaping removes**  
   In `internal/tools/get_wellness_data.go:204-206`, the new hydration semantics check only tests key presence in `out`. Because `wellnessRow` starts from `cloneJSONMap(row.Raw)`, an upstream payload like `{ "hydration": null, "hydrationVolume": null }` leaves those keys in `out` with nil values. `response.Shape` then strips the null top-level fields in terse mode, but `_meta.field_semantics` remains and describes fields that are absent from the final row. This is inconsistent with the existing nutrition semantics behavior and can mislead clients/LLMs about whether hydration data was actually present.

   Suggested fix: only add semantics when the value is non-nil, or delete raw hydration keys before re-setting the typed non-nil fields, mirroring the nutrition-key pattern. Please add a regression test for null hydration/hydrationVolume in terse mode.

## Verification

- Ran: `go test ./internal/tools -run 'TestGetActivityDetails|TestGetWellnessData'` — passed.
