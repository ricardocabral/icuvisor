# R001 Plan Review — Step 5: Null-stripping integration

**Decision:** APPROVE

The Step 5 plan covers the required integration points for TP-007 response shaping:

- It keeps `get_wellness_data` routed through `encodeShaped(..., RowCollections: []string{"wellness"}, ...)`, which is the correct mechanism for per-row null stripping and row-local `_meta.missing_fields`.
- It explicitly relies on rows being built from raw JSON before typed overlays, so both standard wellness keys and dynamic custom fields can reach `response.Shape` as JSON nulls and be reported together.
- It distinguishes terse mode from `include_full:true`: terse mode strips nulls and emits row `_meta.fields_present` / `_meta.missing_fields`; full mode preserves nulls that reach the shaper and suppresses missing-field metadata.
- It preserves the TP-007 contract that only JSON nulls are stripped, avoiding accidental removal of zero values, false booleans, and empty strings.

Non-blocking guardrails for implementation/code review:

1. When verifying `include_full:true`, account for the existing canonicalization/hoisting behavior: claimed native aliases may be removed from the canonical row, but the untouched upstream payload should remain under `full`. The null-strip opt-out should be judged on fields that reach the shaper, plus raw preservation under `full`.
2. Include a custom null field in the verification path, not just a standard field, because custom-field preservation is one of the main reasons wellness rows start from `Raw`.
3. Confirm row-level metadata is attached to each `wellness[]` entry, while wrapper-level `_meta` keeps response metadata such as date range, `include_full`, server version, and units.

No plan revisions required.
