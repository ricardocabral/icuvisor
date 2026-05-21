# Code Review: TP-082 Step 3 — Apply shared shaping consistently

## Verdict: APPROVE

No blocking findings. The previous schema-description mismatch from R009 has been addressed, and the Step 3 implementation remains scoped to routing custom-item write responses through the shared terse response shaper.

## What I checked

- Ran the required diff commands:
  - `git diff 6ba3f7e9d2b886e0f00cd912d64293150c7c6bab..HEAD --name-only`
  - `git diff 6ba3f7e9d2b886e0f00cd912d64293150c7c6bab..HEAD`
- Read the changed implementation files:
  - `internal/tools/create_custom_item.go`
  - `internal/tools/update_custom_item.go`
  - `internal/tools/custom_item_write_validation.go`
- Spot-checked the related response-shaping context:
  - `internal/tools/get_custom_item_by_id.go`
  - `internal/response/shape.go`
  - `internal/response/meta.go`

## Notes

- `create_custom_item` and `update_custom_item` now call `encodeShaped(..., false, ...)`, so nullable keys in the custom-item write echo are handled by the shared terse shaper rather than preserving the full read shape by default.
- The custom-item write output schema descriptions and `_meta.default_payload_scope` text now describe the new null-stripped default and no longer claim the response is the same full/verbatim shape as `get_custom_item_by_id`.
- Request decoding and write payload construction were not changed.

## Tests run

```sh
go test ./internal/tools -run 'Test(CreateCustomItemDefaultStripsSparseNullsAndPreservesMapValues|UpdateCustomItemDefaultStripsSparseNullsAndPreservesMapValues)'
go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'
```

Both passed.
