# R001 plan review — Step 1

Decision: **APPROVE**

The revised Step 1 plan resolves the earlier blocking concerns. It now anchors `get_events` to the documented events list endpoint, removes the undocumented `fields` projection, requires athlete-local `YYYY-MM-DD` range arguments, defines a stable terse row shape, bounds payload size, handles event date/timestamp semantics explicitly, and keeps full raw payloads behind `include_full:true`. The `get_training_plan` plan is also appropriately constrained to upstream-exposed assignment/summary data with a no-active-plan path and no derived periodization assumptions.

No blocking findings.

## Follow-up notes for implementation

- Decode upstream event IDs defensively. The tool contract says `event_id` is a stringified upstream ID, so the intervals DTO should tolerate numeric and string JSON IDs rather than failing if the schema returns an integer.
- Make `get_events` truncation semantics deterministic. Prefer fetching one extra row when possible, or clearly document if `_meta.truncated` means `count == limit` and may be conservative.
- Validate `oldest <= newest` and the 366-day max after parsing as `YYYY-MM-DD`; return a short `UserError` for invalid ranges.
- Ensure both new tools get strict input schemas with descriptions for units/date semantics, `additionalProperties:false`, output schemas, and registry tests matching the existing tool patterns.
- For `get_training_plan`, represent the no-active-plan case as structured content (for example `available:false` plus `_meta.source_endpoint`) rather than an error, unless credentials/auth actually failed.

Step 1 can proceed with these implementation checks, leaving the `get_event_by_id` inconsistency fallback matrix to later steps as planned.
