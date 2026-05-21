# R006 Code Review — Step 2: Shape activity and wellness responses

**Verdict:** APPROVE.

## Findings

None.

The Step 2 implementation matches the approved plan:

- activity `calories` is surfaced as pointer-shaped `calories_burned`, preserving present zero values while omitting absent calories;
- activity calorie semantics are attached under wrapper `_meta.field_semantics` for both `get_activities` and `get_activity_details` when `calories_burned` is present;
- wellness nutrition fields are normalized from upstream names to `calories_intake`, `carbs_g`, `protein_g`, and `fat_g`;
- legacy top-level wellness nutrition keys are removed while the raw upstream names remain available under `full` when `include_full: true`;
- wellness row `_meta.field_semantics` is merged without overwriting existing provenance/scale/missing-field metadata.

## Verification

- `go test ./internal/tools ./internal/response` — pass
- `go test ./...` — pass

## Notes

- Regression tests for the new response keys and metadata are still scheduled for Step 3, which is appropriate for this task plan.
- STATUS bookkeeping still omits R005 from the Reviews table even though the execution log includes it; non-blocking for this code step.
