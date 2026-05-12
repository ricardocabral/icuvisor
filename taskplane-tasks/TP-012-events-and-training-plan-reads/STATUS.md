# TP-012 — Status

**Issue:** v0.2 — read path
**Review Level:** 2
**Status:** ✅ Complete
**Iteration:** 5
**Current Step:** Step 3: Tests
**Last Updated:** 2026-05-12
**State:** Complete

_Task scaffolded from PROMPT.md; execution in progress._

## Execution Log

| Time             | Event                  | Notes                                                                                                                                                                                                    |
| ---------------- | ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-05-12 00:15 | Task started           | Runtime V2 lane-runner execution                                                                                                                                                                         |
| 2026-05-12 00:15 | Step 1 started         | Implement `get_events` and `get_training_plan`                                                                                                                                                           |
| 2026-05-12 00:16 | Step 1 hydrated        | Expanded STATUS.md with task steps and review level                                                                                                                                                      |
| 2026-05-12 00:17 | Plan review R001       | REVISE: Step 1 needed concrete endpoint/contract/date semantics plan                                                                                                                                     |
| 2026-05-12 00:18 | Step 1 plan hydrated   | Added endpoint, response contract, timezone, registration, docs, and deferred-test boundaries                                                                                                            |
| 2026-05-12 00:19 | Plan review R001 retry | REVISE: remove undocumented events fields query, tighten date/query/row/limit contracts                                                                                                                  |
| 2026-05-12 00:20 | Step 1 plan revised    | Aligned get_events with documented OpenAPI query params and explicit terse row fields                                                                                                                    |
| 2026-05-12 00:31 | Code review R001       | REVISE: get_events must enforce row cap after fetching and test truncation                                                                                                                               |
| 2026-05-12 01:39 | Exit intercept timeout | Supervisor did not respond within 60s — closing session                                                                                                                                                  |
| 2026-05-12 01:39 | ⚠️ Steering            | Ignore the stale fallback text about TP-011; you are currently on TP-012. Continue implementing TP-012 now. Address the remaining unchecked items in STATUS.md: enforce get_events limit/truncated/count |
| 2026-05-12 01:39 | Worker iter 1          | done in 5028s, tools: 71                                                                                                                                                                                 |
| 2026-05-12 01:39 | Soft progress          | Iteration 1: 0 new checkboxes but uncommitted source changes detected — not counting as stall                                                                                                            |
| 2026-05-12 01:39 | Step 1 started         | Implement `get_events` and `get_training_plan`                                                                                                                                                           |
| 2026-05-12 01:57 | Exit intercept timeout | Supervisor did not respond within 60s — closing session                                                                                                                                                  |
| 2026-05-12 01:57 | ⚠️ Steering            | You are making source progress on TP-012 but have not checked the remaining review items yet. Continue; do not exit. Commit or finish the in-progress changes in `internal/intervals/events.go`, `intern |
| 2026-05-12 01:57 | Worker iter 2          | done in 1101s, tools: 15                                                                                                                                                                                 |
| 2026-05-12 01:57 | Soft progress          | Iteration 2: 0 new checkboxes but uncommitted source changes detected — not counting as stall                                                                                                            |
| 2026-05-12 01:57 | Step 1 started         | Implement `get_events` and `get_training_plan`                                                                                                                                                           |
| 2026-05-12 02:02 | Code review R002       | APPROVE: Step 1 row cap revision accepted                                                                                                                                                                |
| 2026-05-12 02:03 | Step 2 started         | Implement `get_event_by_id` with fallback                                                                                                                                                                |
| 2026-05-12 02:04 | Plan review R001       | REVISE: hydrate concrete request, fallback, client, response, docs/tests, and fixture privacy contracts                                                                                                  |
| 2026-05-12 02:05 | Step 2 plan hydrated   | Added concrete get_event_by_id contract for implementation                                                                                                                                               |
| 2026-05-12 02:06 | Plan review R001 retry | REVISE: endpoint path, top-level response shapes, and deterministic clock hook needed                                                                                                                    |
| 2026-05-12 02:07 | Step 2 plan revised    | Pinned detail route, response envelopes, non-error miss result, and test clock hook                                                                                                                      |
| 2026-05-12 02:27 | Code review R001       | APPROVE: Step 2 implementation accepted                                                                                                                                                                  |
| 2026-05-12 02:28 | Step 3 started         | Tests                                                                                                                                                                                                    |
| 2026-05-12 02:39 | Verification           | `make test`, `make build`, and `make lint` passed                                                                                                                                                        |
| 2026-05-12 02:23 | Exit intercept close   | Supervisor directed session close: "close"                                                                                                                                                               |
| 2026-05-12 02:23 | Worker iter 3          | done in 1584s, tools: 118                                                                                                                                                                                |
| 2026-05-12 02:23 | No progress            | Iteration 3: 0 new checkboxes (1/3 stall limit)                                                                                                                                                          |
| 2026-05-12 02:23 | Step 1 started         | Implement `get_events` and `get_training_plan`                                                                                                                                                           |
| 2026-05-12 02:25 | Exit intercept close | Supervisor directed session close: "close" |
| 2026-05-12 02:25 | Worker iter 4 | done in 91s, tools: 10 |
| 2026-05-12 02:25 | No progress | Iteration 4: 0 new checkboxes (2/3 stall limit) |
| 2026-05-12 02:25 | Step 1 started | Implement `get_events` and `get_training_plan` |

## Step 1: Implement `get_events` and `get_training_plan`

**Status:** ✅ Complete

- [x] Implement `internal/intervals/events.go` with typed `Event`/`ListEventsParams` structs, `Raw map[string]any` preservation, and `ListEvents(ctx, params)` calling the documented public `GET /athlete/{id}/events` JSON route (empty `{format}` path unless fixtures/docs prove an explicit suffix is required) using only documented query params: required `oldest`/`newest`, optional `category`, `calendar_id`, `limit`, and `resolve`; do not send an undocumented projection/`fields` parameter.
- [x] Implement `internal/tools/get_events.go` with strict JSON argument decoding and athlete-local `YYYY-MM-DD` query dates only (`oldest` and `newest` required), optional safe filters (`category`, `calendar_id`, `resolve`), default `limit` 100 capped at 500, max date range 366 days, athlete-profile timezone fallback, and `_meta.date_range`, `_meta.timezone`, `_meta.limit`, `_meta.count`, `_meta.truncated`, and `_meta.include_full` fields.
- [x] Shape stable terse event rows at the response boundary with `event_id` as a stringified upstream ID, raw upstream `category`, `type`, `name`, `start_date_local`, `end_date_local`, `description`, `workout_doc_summary` (summary/counts only, never full structured workout steps by default), target/load fields when exposed (`icu_training_load`, distance/duration/target summaries), plan linkage fields when exposed (`training_plan_id`, `calendar_id`, `plan_applied`), and `full` only when `include_full:true`.
- [x] Render event time fields with explicit calendar semantics based on documented names: preserve date-like `start_date_local`/`end_date_local` values as supplied by upstream; parse timestamp fields such as `updated` and `plan_applied` when present and add timezone-rendered local companions via `internal/response` helpers without shifting date-only calendar fields.
- [x] Surface event categories as the raw upstream enum value in the terse row (`category`) without embedding long-form category documentation; defer verbose category docs to a later MCP Resource.
- [x] Implement `internal/intervals/training_plan.go` with typed `TrainingPlan`/assignment structs, `Raw map[string]any` preservation, and `GetTrainingPlan(ctx)` calling public `GET /athlete/{id}/training-plan` as the sole source of truth; treat 404/null/no `training_plan_id` as a non-retryable no-active-plan condition rather than a generic credential/date error.
- [x] Implement `internal/tools/get_training_plan.go` returning only upstream-exposed assignment plus lightweight plan summary fields (IDs, alias/name, start date, timezone, last-applied state, top-level nested plan metadata/counts), excluding nested `children[*].workout_doc` and full workout children by default, with `include_full:true` raw payload opt-in, `_meta.source_endpoint`, `_meta.include_full`, `_meta.timezone`, and an availability caveat instead of derived periodization assumptions; record the upstream gap in STATUS.md if only assignment metadata is exposed.
- [x] Wire `get_events` and `get_training_plan` through `internal/tools/registry.go` conditional interfaces, update README catalog plus CHANGELOG, and add focused compile/unit coverage for the new clients/tools and registry wiring now while leaving the broader fallback/inconsistency matrix to Step 3.
- [x] R001 code review: enforce `get_events` limit at the response boundary when upstream/fakes return more than requested, keep `_meta.count` within the cap, set `_meta.truncated` when rows are suppressed, and add a regression test.

## Step 2: Implement `get_event_by_id` with fallback

**Status:** ✅ Complete

- [x] Implement `internal/intervals.GetEvent(ctx, eventID string)` in `events.go` with trimmed non-empty ID validation, `Raw map[string]any` preservation, and the public athlete-scoped detail route `GET /athlete/{athlete_id}/events/{event_id}` with no `{format}` suffix; only `errors.Is(err, intervals.ErrNotFound)` is eligible for fallback while context cancellation/deadline, auth, rate, and transient upstream errors propagate.
- [x] Implement `internal/tools/get_event_by_id.go` with strict JSON request decoding: required non-empty string `event_id`, optional athlete-local `date` or bounded `oldest`/`newest` scan hints, optional `resolve`, optional `include_full`; derive a deterministic fallback scan window from provided bounds, else ±30 days around `date`, else the documented default of athlete-local `now`-30 through `now`+30, inject the `now func() time.Time` clock through the handler constructor for non-flaky timezone tests, and reject spans above 61 days.
- [x] On detail 404, scan via the combined `GetEvent` + `ListEvents` client interface (not by invoking the `get_events` handler) using the derived `oldest`/`newest`, `resolve=true` unless explicitly disabled, and a high-but-bounded fallback `limit`/cap of 500; match by normalized/stringified upstream ID and mark `_meta.truncated` when over-full results could hide a match.
- [x] Return top-level success as `{ "event": <terse row>, "_meta": ... }` for both detail and fallback using the same terse event row/date rendering/full-payload rules as `get_events`, with `_meta.source` (`detail` or `list_scan`), `_meta.recovered`, `_meta.include_full`, `_meta.timezone`, and `_meta.scanned_range` when a scan occurred.
- [x] If detail 404 plus list-scan miss persists, return a structured non-error tool result `{ "unavailable": { "reason": "upstream_inconsistency", "retried": ["detail", "list_scan"] }, "_meta": ... }` with `_meta.scanned_range`, `_meta.count`, `_meta.truncated`, `_meta.include_full`, and without inventing an event, returning a `NewUserError`, or exposing a raw 404.
- [x] Wire `get_event_by_id` through `internal/tools/registry.go`, update README catalog plus CHANGELOG, and add focused compile/unit coverage for the new client/tool/registry while leaving the full matrix to Step 3.
- [x] Capture every real reproducer encountered under `internal/intervals/testdata/events/inconsistent/` with the originating list response; scrub athlete identifiers and personal notes/descriptions, and add a README describing the fixture format. If no real reproducer is available, add synthetic fixtures documenting the mismatch shape.

## Step 3: Tests

**Status:** ✅ Complete

- [x] Table-driven tests using `httptest.Server` + fixtures
- [x] Cover: list / detail happy path; detail 404 → list-scan recovery; detail 404 + list-scan miss → structured `unavailable`; TZ rendering on event dates
- [x] `make test`, `make build`, `make lint` pass

## Discoveries

| Date       | Finding                                                                                                                                                                                                              | Impact                                                                                                                                                                              |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-05-12 | Public training-plan read is treated as an active assignment endpoint (`/athlete/{id}/training-plan`) with optional lightweight nested plan metadata; no generated periodization parameters are derived or invented. | `get_training_plan` returns assignment/summary plus an availability caveat and points callers to `get_events` for calendar workouts when upstream exposes only assignment metadata. |

## Blockers

| Date             | Blocker     | Attempts             |
| ---------------- | ----------- | -------------------- |
| 2026-05-12 00:19 | Review R001 | plan Step 1: REVISE  |
| 2026-05-12 00:22 | Review R001 | plan Step 1: REVISE  |
| 2026-05-12 00:25 | Review R001 | plan Step 1: APPROVE |
| 2026-05-12 00:50 | Review R001 | code Step 1: REVISE  |
| 2026-05-12 02:04 | Review R001 | plan Step 2: REVISE  |
| 2026-05-12 02:06 | Review R001 | plan Step 2: REVISE  |
| 2026-05-12 02:00 | Review R001 | code Step 1: APPROVE |
| 2026-05-12 02:03 | Review R001 | plan Step 2: REVISE  |
| 2026-05-12 02:06 | Review R001 | plan Step 2: REVISE  |
| 2026-05-12 02:08 | Review R001 | plan Step 2: APPROVE |
| 2026-05-12 02:19 | Review R001 | code Step 2: APPROVE |
