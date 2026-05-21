# Plan Review: TP-082 Step 3 — Apply shared shaping consistently

## Verdict: APPROVE

The Step 3 plan is acceptable as a narrow implementation step, because the prior audit and Step 2 red-test evidence identify only one divergent response path: custom-item writes still call `encodeShaped(..., true, ...)` by default. The implementation should keep this scoped to routing `create_custom_item` and `update_custom_item` through the shared terse shaper, not perform a broad response-package refactor.

## What I checked

- Read `PROMPT.md` and `STATUS.md` for the task requirements and current Step 1/2 discoveries.
- Spot-checked the relevant code paths:
  - `internal/tools/create_custom_item.go`
  - `internal/tools/update_custom_item.go`
  - `internal/tools/custom_item_write_validation.go`
  - `internal/tools/get_custom_item_by_id.go`
  - `internal/response/shape.go`
- Re-ran the current targeted custom-item red tests:

```sh
go test ./internal/tools -run 'Test(CreateCustomItemDefaultStripsSparseNullsAndPreservesMapValues|UpdateCustomItemDefaultStripsSparseNullsAndPreservesMapValues)'
```

They fail for the expected reason: default custom-item write responses preserve `image: null` because the handlers use full-mode shaping.

## Implementation constraints for Step 3

- Keep the fix focused on the write handlers that currently hard-code full shaping:
  - `create_custom_item`
  - `update_custom_item`
- Prefer changing their response-boundary call to shared terse shaping (`includeFull=false`) rather than adding bespoke recursive null cleanup.
- Do not change request structs, input schemas, validation, or uploaded `intervals.WriteCustomItemParams` payloads.
- Do not change `get_custom_item_by_id`; its full custom-item read behavior is outside this write-response task and is intentionally full-preserving today.
- Do not add a new model-controlled `include_full` argument to custom-item writes in this step. Step 2 explicitly treated those tools as having no user-facing full path.
- Preserve wrapper metadata added by the shared shaper, especially `_meta.server_version`; allow the existing missing-field metadata behavior from `response.Shape` to report stripped null paths.
- If output-schema or `_meta.default_payload_scope` wording is updated to avoid saying the default write response is “full,” keep that change minimal and reflect any user-visible wording change in the Step 4 changelog update.

## Notes

- The simplest likely implementation is enough: route the two custom-item write responses through `encodeShaped(..., false, nil, ...)` and let `response.Shape` strip null map keys recursively while preserving zero, false, and empty-string values.
- After the code change, run the Step 2 targeted command for all write-tool tests. If failures appear outside the two planned custom-item tests, stop and record the unexpected behavior in `STATUS.md` before expanding the fix.
