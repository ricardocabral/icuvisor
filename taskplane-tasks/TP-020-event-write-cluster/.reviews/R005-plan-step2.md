# Plan Review: TP-020 Step 2 (`link_activity_to_event`)

**Verdict: APPROVE**

I read `PROMPT.md`, `STATUS.md`, the prior R004 review, and the PRD/roadmap anchors for the event-write cluster. The revised Step 2 plan in `STATUS.md` now addresses the R004 blockers well enough to implement.

## What is now covered

- **Endpoint/write semantics:** The plan identifies a clean-room public-OpenAPI basis for `PUT /api/v1/activity/{id}` with a typed request containing numeric `paired_event_id`, plus a typed `LinkActivityToEventParams` client method and deliberate PUT retry behavior.
- **ID handling:** The plan resolves the earlier ambiguity by trimming and rejecting empty IDs, preserving upstream activity IDs as provided, parsing `event_id` only for the endpoint-specific numeric `paired_event_id`, and not applying athlete-ID normalization to activity/event IDs.
- **Mismatch warnings:** The plan calls for fetching activity/event dates via existing detail clients and surfacing stable `_meta.warnings[]` objects for date mismatches without turning warning-read failures into link failures.
- **Safety/registration:** The tool is planned as `RequirementWrite`, present in write-enabled modes, absent in `none`, and explicitly without a model-controlled `confirm` argument.
- **Schema/docs/tests:** The plan includes schema snapshot/catalog/README/CHANGELOG updates and covers success, mismatch warning, idempotent re-link, validation, and no-confirm schema behavior.

## Implementation notes to preserve during coding

These are not blockers for the plan, but should be made concrete in code/tests:

1. Capture the verified upstream request contract in intervals-client tests: relative path `activity/{id}` under the configured `/api/v1` base URL, `PUT`, JSON body with `paired_event_id` serialized as a number, expected response decoding, non-2xx wrapping, retry behavior for idempotent PUTs, and response-body closure.
2. Define the warning object shape before writing assertions; prefer stable fields such as `code`, `message`, `activity_date`, and `event_date` rather than string-only warnings.
3. Parse `event_id` with an explicit range/format check and reject non-numeric values with a short user-facing validation error. Preserve `activity_id` exactly after trimming, including IDs with prefixes such as `i...`.
4. Keep the public response terse by default: at minimum `activity_id`, `event_id`, a success/status field, and `_meta.warnings`; avoid exposing raw activity/event payloads unless an explicit `include_full` contract is added and tested.
5. Do not log raw IDs, activity notes, event descriptions, or any future message bodies while implementing this write path.

With those details carried into the implementation, Step 2 is ready to proceed.
