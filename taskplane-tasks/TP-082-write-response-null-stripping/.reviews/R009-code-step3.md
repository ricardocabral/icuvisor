# Code Review: TP-082 Step 3 — Apply shared shaping consistently

## Verdict: REQUEST CHANGES

The runtime fix is pointed in the right direction: `create_custom_item` and `update_custom_item` now call `encodeShaped(..., false, ...)`, and the targeted custom-item red tests pass. However, the public output schema descriptions for those same tools still promise the old full/verbatim response shape, which now contradicts the actual default null-stripped response.

## Findings

### 1. Output schemas still describe a full/verbatim custom-item response

- `internal/tools/create_custom_item.go:194`
- `internal/tools/update_custom_item.go:232`

Both output schemas still say the write confirmation contains “the same full custom_item read shape as get_custom_item_by_id” and that content is “preserved verbatim.” After this step, the handlers route through terse shared shaping (`includeFull=false`), so nullable keys from `custom_item` and nested `content` are stripped by default and the response is no longer the same full/verbatim read shape.

This is user-visible MCP schema text, and the project guidelines call out schemas/return shapes as part of the API. Please update these descriptions to match the new behavior, e.g. “custom-item detail shape with null keys stripped by default; non-null content fields preserved,” while keeping `get_custom_item_by_id`’s full/verbatim wording unchanged.

## What I checked

- Ran the required diff commands:
  - `git diff 6ba3f7e9d2b886e0f00cd912d64293150c7c6bab..HEAD --name-only`
  - `git diff 6ba3f7e9d2b886e0f00cd912d64293150c7c6bab..HEAD`
- Read the changed implementation files:
  - `internal/tools/create_custom_item.go`
  - `internal/tools/update_custom_item.go`
  - `internal/tools/custom_item_write_validation.go`
- Spot-checked shared shaping/metadata behavior in `internal/response/shape.go` and `internal/response/meta.go`.

## Tests run

```sh
go test ./internal/tools -run 'Test(CreateCustomItemDefaultStripsSparseNullsAndPreservesMapValues|UpdateCustomItemDefaultStripsSparseNullsAndPreservesMapValues)'
go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'
```

Both passed.
