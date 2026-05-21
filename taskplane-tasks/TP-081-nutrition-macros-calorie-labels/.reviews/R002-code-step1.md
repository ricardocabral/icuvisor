# R002 Code Review — Step 1: Map nutrition fields from upstream fixtures

**Verdict:** Request changes

## Findings

1. **Macro fixture test does not prove the key-to-field mapping.**  
   `internal/intervals/wellness_test.go:60` only checks that `Carbohydrates`, `Protein`, and `FatTotal` are non-nil positive values. Because the fixture values are all positive, swapped JSON tags or swapped field assignments would still pass. Step 1 is specifically about mapping upstream nutrition keys, so this should assert exact fixture values for each field, e.g. `Carbohydrates == 320.5`, `Protein == 132.25`, and `FatTotal == 78.75`.

2. **Activity calories assertion is paired with a field filter that excludes calories.**  
   `internal/intervals/activity_gear_test.go:33` calls `ListActivities` with `Fields: []string{"id", "gear_id"}` but the new assertion expects `Calories` to be present. A real upstream response honoring the `fields` query would not include `calories` for that request, so the test now documents an impossible/misleading contract. Include `"calories"` in the requested fields or remove the field filter for the list-side decoder assertion.

## Non-blocking notes

- The Step 1 mapping/gap is recorded in `STATUS.md`, which satisfies the earlier plan-review direction. The review/status bookkeeping is slightly inconsistent: the Reviews table is still empty, and the R001 review entry appears under Notes rather than the Reviews/Execution Log table.

## Tests run

- `go test ./internal/intervals`
- `go test ./internal/tools`
