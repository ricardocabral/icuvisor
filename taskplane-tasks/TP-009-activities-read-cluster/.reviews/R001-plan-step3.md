# Plan Review — TP-009 Step 3

Decision: **Changes requested**

## Summary

I cannot approve Step 3 yet because the Step 3 implementation plan has not been recorded beyond the three checklist bullets copied from the task prompt. `STATUS.md` does not define the concrete client methods, request/response contracts, unit handling, Strava-blocked fallback behavior, registry wiring, or targeted tests for `get_activity_details` and `get_activity_intervals`.

Step 1 captured the high-level endpoints, but Step 3 still needs an implementation-level plan before coding so these two tools do not drift from the already-approved `get_activities` response-shaping and Strava behavior.

## Blocking issues

1. **Details/intervals client and model contract is missing**
   - Specify the exact `internal/intervals` methods to add, e.g. detail fetch against `GET /api/v1/activity/{id}?intervals=false` and intervals fetch against `GET /api/v1/activity/{id}/intervals`.
   - Define the upstream structs to add/extend: detail activity fields, raw preservation for `include_full`, `IntervalsDTO`, interval rows, groups, and any hidden/stub shape needed for Strava detection.
   - Document input validation for `activity_id` and how upstream `ErrNotFound`, `ErrUnauthorized`, cancellation, and transient errors are mapped to short user-facing errors.

2. **Public response shapes are undefined**
   - Decide whether `get_activity_details` returns `{ activity: ..., _meta: ... }` or a top-level activity object, and list the terse fields, including unit-disambiguated names and timezone behavior.
   - Decide the `get_activity_intervals` wrapper shape (`activity_id`, `intervals`, `groups`, `_meta`) and which fields are included by default versus only under `include_full`.
   - Include `_meta` requirements: `server_version`, `include_full`, unit metadata, and any unknown-unit metadata. Keep null stripping and raw/null preservation aligned with TP-007 and the Step 2 `include_full` strategy.

3. **Canonical interval unit handling is not planned**
   - The plan must enumerate which upstream interval fields carry unit tokens or target units and how they map through TP-008 `units.ParseUnit`.
   - Define the shaped representation for known units and the behavior for unknown units, including preserving the raw upstream token and surfacing it in `_meta.unknown_units` (or a similarly explicit field).
   - Add fixture coverage for known and unknown units so the enum contract is locked down.

4. **Strava-blocked single-activity behavior needs an explicit fallback path**
   - Reuse or centralize the `get_activities` hidden/stub detection logic; do not duplicate a narrower detector.
   - For `get_activity_details`, specify how hidden/stub/Strava-sourced raw payloads become the same `unavailable: { reason: "strava_tos", workaround: ... }` shape.
   - For `get_activity_intervals`, specify how the tool avoids propagating an upstream 4xx for a Strava-blocked activity. For example, on a not-found/forbidden intervals response, perform a detail/stub check and return `unavailable` only when the detail payload confirms the Strava-blocked shape; do not mask genuine credential failures as Strava.

5. **Registry and test plan are absent**
   - Define the new tool/client interfaces and how `NewRegistry` registers `get_activity_details` and `get_activity_intervals` when the configured intervals client supports them.
   - Add targeted Step 3 tests rather than deferring all validation to Step 6: intervals client path/query tests, detail terse/full raw behavior, profile-unit conversion/timezone fallback, Strava unavailable shaping, interval unit enum/unknown-unit metadata, registry/schema exposure, and context-cancellation preservation.

## Required updates before approval

Please update `STATUS.md` with a `Step 3 Notes` section covering:

- Concrete endpoint/client/type changes for details and intervals.
- Exact request and response schemas for both tools, including terse vs `include_full` behavior.
- Unit-canonicalization rules for interval rows and unknown-unit metadata.
- Strava-blocked handling for both direct detail reads and intervals reads that encounter upstream 4xx/stub behavior.
- Registry/app wiring changes and a focused Step 3 test checklist.

## Non-blocking status hygiene

`STATUS.md` currently has review-log rows after `_None._` under `## Blockers`. Move or remove those rows so the Blockers section remains unambiguous.
