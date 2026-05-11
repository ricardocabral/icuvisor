# Plan Review — TP-009 Step 5

Decision: **Changes requested**

## Summary

I cannot approve Step 5 yet because the Step 5 implementation plan has not been recorded beyond the two checklist bullets copied from the task prompt. `STATUS.md` has a useful Step 1 endpoint note for `GET /api/v1/activity/{id}/messages`, but it does not define the concrete client/types, MCP request and response contract, timezone conversion rules, Strava-blocked fallback behavior, registry wiring, or targeted tests for `get_activity_messages`.

This is a smaller step than streams/splits, but it still touches response-shaping, profile timezone lookup, raw/null preservation, and activity-read unavailable handling. Please lock down the public contract before coding.

## Blocking issues

1. **Messages client and upstream model contract is missing**
   - Specify the exact `internal/intervals` method and params to add, e.g. `GetActivityMessages(ctx, intervals.ActivityMessagesParams)` against `GET /api/v1/activity/{id}/messages`.
   - Define query handling for documented upstream parameters (`sinceId`, `limit`) and their defaults/validation. If the tool will not expose both parameters, explicitly say why and what fixed values the client sends.
   - Define the tolerant upstream `Message` struct and raw preservation strategy for `include_full`: `id`, `athlete_id`, `name`, `created`, `type`, `content`, `activity_id`, interval indexes, `answer`, `attachment`, `deleted`, `seen`, and any unmodeled fields.
   - Document input validation for blank `activity_id`, invalid `limit`/`since_id`, and how upstream `ErrNotFound`, `ErrUnauthorized`, cancellation, and transient errors map to user-facing errors.

2. **Public `get_activity_messages` request/response shape is undefined**
   - Define the input schema, including `activity_id`, `include_full`, and whether callers can pass `limit` and/or `since_id`/`sinceId`. Each argument needs an LLM-readable JSON Schema description.
   - Define the wrapper shape, e.g. `{ activity_id, messages, _meta }`, and the default terse message fields. Use disambiguated names such as `message_id`, `author_name`, `message_type`, `created_at`, `content`, `deleted`, `seen`, and interval index fields if included.
   - Decide whether default mode returns full message `content` or a bounded preview plus metadata. If full content is returned by default, justify that it is still terse enough for comments/notes; if it is truncated, define the truncation marker and how `include_full` exposes the raw content.
   - Define `_meta` fields: `server_version`, `include_full`, `timezone`, effective `limit`, `since_id` if used, and any pagination/incremental-fetch metadata. If no `next_page_token` is planned for messages, state that the upstream only supports `sinceId` and how callers should request subsequent/newer messages.
   - Keep `include_full:true` aligned with TP-007/Step 2 behavior: preserve raw upstream nulls/unmodeled fields, while terse mode goes through null stripping.

3. **Timezone rendering is not planned concretely**
   - Specify that the handler fetches the athlete profile via `ProfileClient` and renders `created` in the athlete/configured timezone, falling back through the existing `profileTimezone(..., timezoneFallback)` path.
   - Define the output timestamp fields and format. For example, `created_at` as RFC3339 with the athlete-local offset, and optionally `created_at_utc` only under `include_full` or as part of the raw payload.
   - Define behavior for missing or unparsable upstream timestamps: whether to omit the rendered field, preserve the raw value in `full`, and/or surface a short warning in `_meta`.
   - Preserve context cancellation/deadline errors from profile lookup and message fetches instead of wrapping them as ordinary user errors, consistent with previous activity tools.

4. **Strava-blocked activity behavior is not specified**
   - The TP-009 mission and PRD require structured unavailable handling across the activity read cluster. The plan should say how `get_activity_messages` returns `unavailable: { reason: "strava_tos", workaround: ... }` for blocked activities rather than sparse rows or a generic 4xx.
   - For a messages endpoint 404/403, specify the same detail-read confirmation path used by `get_activity_intervals`: only return Strava unavailable when `GetActivity` confirms the hidden/stub/Strava marker; do not mask genuine credential or missing-activity failures.
   - Define the unavailable response shape for this tool, including whether it is top-level next to `activity_id` and `_meta`, and how `include_full` carries any raw confirmation payload.

5. **Registry/app wiring and schema exposure are not planned**
   - Define the `ActivityMessagesClient` interface in `internal/tools` and how `NewRegistry` registers `get_activity_messages` when the configured intervals client supports it.
   - Specify any app/client wiring changes needed so the production `intervals.Client` satisfies both `ActivityMessagesClient` and the detail/profile interfaces needed for timezone and Strava fallback.
   - Add README catalog and `CHANGELOG.md` updates for the new tool as part of the step plan, since Step 5 completes the sixth activity read tool.

6. **Targeted Step 5 tests are missing**
   - Add focused tests before code review rather than deferring everything to Step 6:
     - `internal/intervals` `httptest.Server` coverage for path/query construction, `sinceId`/`limit`, raw/null preservation, and error propagation.
     - Tool tests for terse default shape, `include_full` raw preservation, timezone rendering using a non-UTC athlete timezone, invalid arguments, limit validation if exposed, and context cancellation preservation.
     - Strava-blocked fallback tests for messages endpoint not-found/forbidden plus confirmed hidden/stub detail payload, and a negative test that genuine unauthorized/not-found is not mislabeled as Strava.
     - Registry/schema exposure tests so `get_activity_messages` is actually registered with useful descriptions.

## Required updates before approval

Please update `STATUS.md` with a `Step 5 Notes` section covering:

- Concrete endpoint/client/type changes for activity messages.
- Exact request and response schemas, including terse vs `include_full`, limit/`sinceId`, `_meta`, and unavailable response shape.
- Timestamp parsing/rendering rules using the athlete timezone and fallback behavior.
- Strava-blocked fallback detail-read confirmation behavior.
- Registry/app wiring and a focused Step 5 test checklist.

## Non-blocking status hygiene

`STATUS.md` still has review-log rows after `_None._` under `## Blockers`. Move or remove those rows so the Blockers section remains unambiguous before task wrap-up.
