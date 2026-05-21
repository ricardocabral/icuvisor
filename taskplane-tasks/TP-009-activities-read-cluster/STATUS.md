# TP-009 — Activities read cluster — Status

**Issue:** v0.2 — read path
**Status:** ✅ Complete
**Iteration:** 2
**Review Level:** 2 (Plan + Code)
**Current Step:** Step 6: Tests
**Last Updated:** 2026-05-11
**State:** Complete

_Task scaffolded from PROMPT.md; execution in progress._

## Execution Log

| Time             | Event           | Notes                                                 |
| ---------------- | --------------- | ----------------------------------------------------- |
| 2026-05-11 15:39 | Task started    | Runtime V2 lane-runner execution                      |
| 2026-05-11 15:39 | Step 1 started  | Plan endpoints, types, and pagination shape           |
| 2026-05-11 15:40 | Hydrated STATUS | Added step checklists from PROMPT.md for resumability |
| 2026-05-11 15:42 | Review R001     | plan Step 1: APPROVE                                  |
| 2026-05-11 15:47 | Review R001     | code Step 1: REVISE                                   |
| 2026-05-11 16:11 | Review R001     | code Step 1: REVISE                                   |
| 2026-05-11 16:13 | Review R001     | code Step 1: REVISE                                   |
| 2026-05-11 16:17 | Review R001     | code Step 1: APPROVE                                  |
| 2026-05-11 16:20 | Review R001     | plan Step 2: REVISE                                   |
| 2026-05-11 16:24 | Review R001     | plan Step 2: APPROVE                                  |
| 2026-05-11 16:46 | Exit intercept timeout | Supervisor did not respond within 60s — closing session |
| 2026-05-11 16:46 | ⚠️ Steering | Continue TP-009; do not exit for the previous wrap-up condition. I checked the lane worktree and there is no current `.task-wrap-up` signal present. You have implementation changes in progress (`inter |
| 2026-05-11 16:46 | Worker iter 1 | done in 4019s, tools: 122 |
| 2026-05-11 16:46 | Soft progress | Iteration 1: 0 new checkboxes but uncommitted source changes detected — not counting as stall |
| 2026-05-11 16:46 | Step 1 started | Plan endpoints, types, and pagination shape |
| 2026-05-11 17:05 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 17:15 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 17:25 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 17:35 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 17:45 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 17:55 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 18:05 | Review R001     | code Step 2: REVISE                                  |
| 2026-05-11 18:15 | Review R001     | code Step 2: APPROVE                                 |
| 2026-05-11 18:16 | Step 3 started  | Implement `get_activity_details`, `_intervals`       |
| 2026-05-11 18:17 | Review R001     | plan Step 3: APPROVE                                 |
| 2026-05-11 18:30 | Review R001     | code Step 3: REVISE                                  |
| 2026-05-11 18:40 | Review R001     | code Step 3: APPROVE                                 |
| 2026-05-11 18:41 | Step 4 started  | Implement `get_activity_streams` and `get_activity_splits` |
| 2026-05-11 18:42 | Review R001     | plan Step 4: APPROVE                                 |
| 2026-05-11 18:55 | Review R001     | code Step 4: APPROVE                                 |
| 2026-05-11 18:56 | Step 5 started  | Implement `get_activity_messages`                    |
| 2026-05-11 18:57 | Review R001     | plan Step 5: APPROVE                                 |
| 2026-05-11 19:05 | Review R001     | code Step 5: REVISE                                  |
| 2026-05-11 19:15 | Review R001     | code Step 5: REVISE                                  |
| 2026-05-11 19:25 | Review R001     | code Step 5: APPROVE                                 |
| 2026-05-11 19:26 | Step 6 started  | Tests                                                |

## Step 1: Plan endpoints, types, and pagination shape

**Status:** ✅ Complete

- [x] Identify endpoints, query params, and response shapes from public docs; record uncertainty in `STATUS.md`
- [x] Decide pagination contract: opaque `next_page_token`, default page size that fits the §7.2.D ~30k-token soft ceiling
- [x] Decide how `include_unnamed` filters (issue #67) — server-side filter vs client-side post-filter; prefer server-side
- [x] R001 fix: correct `include_full` plan so raw/full responses preserve upstream nulls while terse mode null-strips
- [x] R001 fix: specify client-side unnamed filtering pagination loop, over-fetch cap, token payload, validation, and same-timestamp ordering
- [x] R001 fix: clarify stream canonicalization over `ActivityStream.type`/response map shape and explicit key mapping
- [x] R001 follow-up: move review log rows out of Blockers and keep reviewer-local state untracked

### Step 1 Notes

Endpoint plan from the public OpenAPI spec at `https://intervals.icu/api/v1/docs`:

| Tool                     | Upstream endpoint                                                          | Query/path params                                                                                                                                                                                 | Upstream response shape                                                                                                                          | Shaped response boundary                                                                                                                                                                                                                                                                                                                                                      |
| ------------------------ | -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `get_activities`         | `GET /api/v1/athlete/{id}/activities`                                      | path `id`; query `oldest` required local ISO date/date-time, `newest` optional, `route_id` optional, `limit` optional, `fields` optional comma-separated allowlist that also excludes null values | `[]Activity`; docs guarantee descending date order and state Strava activities are empty stubs                                                   | wrapper `{ activities: [...], _meta: { page_size, next_page_token, server_version, units } }`; terse rows keep id/name/type/start/duration/distance/load/core metrics only; `include_full` preserves raw typed fields and upstream nulls as far as decoding allows                                                                                                            |
| `get_activity_details`   | `GET /api/v1/activity/{id}`                                                | path `id`; query `intervals=false` for details-only                                                                                                                                               | `Activity`, `ActivityWithIntervals`, or `Hidden`; docs state Strava activities are empty stubs                                                   | single object with normalized id/name/type/start/duration/distance/load/zones/metrics plus `unavailable` for hidden/stub rows; `include_full` carries extra typed fields                                                                                                                                                                                                      |
| `get_activity_intervals` | `GET /api/v1/activity/{id}/intervals`                                      | path `id`                                                                                                                                                                                         | `IntervalsDTO` with `id`, `analyzed`, `icu_intervals: []Interval`, `icu_groups: []IntervalGroup`                                                 | wrapper `{ activity_id, intervals, groups, _meta }`; interval rows use typed unit enum/labels where unit-bearing target fields appear and preserve unknown units via `_meta.unknown_unit`                                                                                                                                                                                     |
| `get_activity_streams`   | `GET /api/v1/activity/{id}/streams{ext}`                                   | path `id`, `ext` empty for JSON; query `types`, `includeDefaults`                                                                                                                                 | `[]ActivityStream` (`type`, `name`, `data`, `data2`, `valueTypeIsArray`, `anomalies`, `custom`, `allNull`)                                       | wrapper `{ activity_id, streams, _meta }`; response reshapes the array into a map keyed by canonical snake_case stream name; canonicalize both requested `keys` and upstream `ActivityStream.type` with TP-008 `streams.CanonicalKey`, surface unknown requested/upstream keys in `_meta.unknown_stream_keys`; full stream payload requires `include_full` or explicit `keys` |
| `get_activity_splits`    | no distinct public split endpoint found; derive from intervals and streams | read intervals first; if no manual laps, read distance/time/moving streams via `streams{ext}` with explicit keys                                                                                  | manual laps are `Interval` rows when present; virtual splits derive from stream samples                                                          | wrapper `{ activity_id, split_unit, splits, source, _meta }`; per-km for metric and per-mile for imperial unless caller requests `split_unit`; algorithm documented in package doc                                                                                                                                                                                            |
| `get_activity_messages`  | `GET /api/v1/activity/{id}/messages`                                       | path `id`; query `sinceId`, `limit` default 100                                                                                                                                                   | `[]Message` (`id`, `athlete_id`, `name`, `created`, `type`, `content`, `activity_id`, interval indexes, answer/attachment/deleted/seen metadata) | wrapper `{ activity_id, messages, _meta }`; timestamps rendered in athlete/configured timezone and raw UTC kept only in `include_full`                                                                                                                                                                                                                                        |

Typed model outline:

- `internal/intervals` owns tolerant upstream structs: `Activity` with stable fields used by terse shaping plus `Raw map[string]any` for `include_full` fallback if needed; `HiddenActivity`/stub detection helper; `IntervalsDTO`, `Interval`, `IntervalGroup`; `ActivityStream`; `Message`.
- `internal/tools` owns public MCP response structs per tool and does unit, timezone, pagination, Strava-unavailable, and stream-key shaping at the response boundary.
- TP-007 `response.Shape` remains the last boundary for terse-mode null stripping, `_meta.server_version`, optional debug metadata, scale labels, and unit metadata. Full mode uses `response.Options{IncludeFull:true}` so raw/debug payloads preserve upstream nulls instead of applying null stripping. TP-008 `streams.CanonicalKey` canonicalizes requested stream keys and upstream `ActivityStream.type`; `streams.CanonicalizeRow` is reserved for map-shaped raw stream data. `units.ParseUnit` is used before the response boundary so unknown-unit metadata is present in shaped output.
- `_meta` conventions: every wrapper includes `server_version`; list wrappers include `page_size`, `next_page_token` when present, and `more_available`; stream wrappers include `unknown_stream_keys`; unit conversions include `units` from TP-007; upstream totals are not reported because the public activities endpoint does not return total counts.

Uncertainties to keep explicit in code/tests:

- The exact Strava-blocked marker should be detected defensively: docs mention both an empty stub object and a `Hidden` shape (`id`, `icu_athlete_id`, `start_date_local`, `source`, `_note`). Treat rows with only stub/hidden fields, `source` resembling Strava, or `_note` mentioning Strava as unavailable until black-box testing confirms the stable marker.
- Public docs do not expose an upstream offset/page token for `get_activities`; pagination must be implemented by bounded date-range fetch + local opaque token unless later black-box testing finds a hidden cursor.
- Pagination contract: `page_size` defaults to 50 terse rows, accepts 1-200, and each upstream request uses `limit=min(page_size*2+1, 201)` so client-side filtering can fill pages without unbounded payloads. The tool may issue up to 5 upstream requests per page while advancing a date/id cursor; if filtered rows exhaust those windows before filling `page_size`, it returns the rows found plus a `next_page_token` when a lookahead row proves more data may exist. This keeps default responses well under the PRD §7.2.D ~30k-token soft ceiling even with `_meta` and null-stripping metadata.
- Cursor/order details: normalize each fetched window into deterministic local order `start_date_local desc, id desc` (falling back to `start_date` then original upstream order only when IDs are absent). The token cursor is exclusive and contains `before_start_date_local`, `before_id`, and `skip_ids_at_boundary` for rows already emitted or filtered at the same timestamp. When a whole window is filtered (for example unnamed-only rows), the cursor still advances past the filtered boundary group so the next upstream request cannot loop or duplicate rows.
- Token payload before URL-safe base64 JSON encoding: `{ "v": 1, "oldest": "YYYY-MM-DD", "newest": "YYYY-MM-DD", "route_id": <optional>, "include_unnamed": false, "include_full": false, "page_size": 50, "fields": [<upstream field allowlist>], "before_start_date_local": "...", "before_id": "...", "skip_ids_at_boundary": ["..."] }`. Supplying a token with mismatched explicit filters, invalid JSON/base64, unsupported version, or cursor outside the requested date range returns a short invalid-arguments user error.
- `include_unnamed` decision: public docs expose no upstream query parameter for excluding unnamed activities, so implement a client-side post-filter after fetching bounded page windows. Default `include_unnamed=false` drops activities whose trimmed `name` is empty; `include_unnamed=true` preserves them. The token includes `include_unnamed` and cursor state because filtering can consume upstream rows that never appear in the public page.
- Stream key mapping: the public `keys` argument accepts canonical snake_case names or known upstream aliases. The tool maps each requested key to the best upstream `types` query token, calls `GET /streams` with `types`, then emits `streams.<canonical_key>` entries. Unknown requested keys are still passed through best-effort as the upstream `types` value and listed in `_meta.unknown_stream_keys`; unknown upstream `ActivityStream.type` values are converted with `ToSnakeCase` and listed in the same metadata. If no `keys` and no `include_full`, the tool returns only available stream names/metadata, not samples.
- The spec has no manual lap/splits endpoint distinct from intervals; `get_activity_splits` should use intervals as manual laps when present and compute virtual splits from distance/time streams when not.

## Step 2: Implement `get_activities` with pagination

**Status:** ✅ Complete

- [x] Add typed intervals activity list client/types and tests for query params/page-window behavior
- [x] Date-range arguments, `include_unnamed: bool`, `page_size`, `next_page_token`
- [x] Terse default rows; full payload behind `include_full: true`
- [x] Strava-blocked rows surface as `unavailable: { reason: "strava_tos", workaround: "connect device directly to intervals.icu (Garmin, Wahoo, Coros, Suunto, Polar)" }` — do not emit `N/A`-laden rows
- [x] Apply `preferred_units` via TP-007 unit-system plumbing for distance / pace fields
- [x] Register `get_activities` and update README/CHANGELOG catalog entries
- [x] R001 plan fix: specify unit-disambiguated terse row schema for metric and imperial athletes
- [x] R001 plan fix: specify `include_full` raw/null preservation and terse upstream `fields` allowlist strategy
- [x] R001 plan fix: specify registry/client interfaces for activity reads plus profile-unit lookup
- [x] R001 plan fix: specify Step 2 targeted tests for pagination, tokens, filtering, Strava, full nulls, units, and registration
- [x] R001 code fix: return a continuation token when filtered full windows hit the over-fetch cap before exhausting upstream data, with regression coverage
- [x] R001 code fix: preserve context cancellation/deadline errors from profile lookup instead of wrapping them as user-facing fetch errors
- [x] R001 code fix: keep Strava-blocked empty-name rows visible under default `include_unnamed=false`, with regression coverage
- [x] R001 code fix: emit pace fields only for run-like activities while preserving distance/speed fields for cycling and other sports, with regression coverage
- [x] R001 code fix: broaden Strava-blocked detection for documented empty/stub shapes and cover them in tests
- [x] R001 code fix: prevent same-timestamp filtered-window pagination loops when no semantic cursor progress is possible, with regression coverage
- [x] R001 code fix: prevent continuation tokens for ID/date-less Strava stubs unless pagination cursor advances, with regression coverage
- [x] R001 code fix: use the fetched athlete profile timezone as the activity row fallback when per-row timezone is missing, with regression coverage
- [x] R001 status fix: keep STATUS current-step metadata and review log placement valid for Step 2 review
- [x] R001 code fix: advance past fully filtered same-timestamp boundary windows so older eligible activities remain reachable, with static limit-filter fake coverage
- [x] R001 code fix: widen same-timestamp boundary lookahead before crossing filtered full windows so eligible same-timestamp rows are not skipped, with regression coverage
- [x] R001 status fix: remove stray review-log rows from Blockers
- [x] R001 code fix: stop before issuing inverted lower-bound activity ranges after filtered boundary rewinds, with regression coverage
- [x] R001 code fix: return an explicit bounded-pagination error instead of crossing full same-timestamp groups that still may contain eligible rows beyond max lookahead, with regression coverage
- [x] R001 code fix: classify null-only documented hidden stub shapes as Strava-unavailable, with regression coverage

### Step 2 Notes

Terse `get_activities` row schema:

- Common fields: `activity_id`, `name`, `sport`, `sub_type`, `start_date_local` (athlete-local upstream timestamp), `start_date_utc` when available, `timezone`, `moving_time_seconds`, `elapsed_time_seconds`, `training_load`, `average_heart_rate_bpm`, `max_heart_rate_bpm`, `average_cadence_rpm`, `calories_burned`, `device_name`, `has_streams`, `strava_imported` when detected, and `unavailable` instead of metric fields for Strava-blocked stubs.
- Metric athletes receive `distance_km`, `pace_seconds_per_km` for run-like activities, and `average_speed_kmh` for speed fields when emitted.
- Imperial athletes receive `distance_mi`, `pace_seconds_per_mile` for run-like activities, and `average_speed_mph` for speed fields when emitted.
- Elevation is stable regardless of measurement preference as `elevation_gain_m`/`elevation_loss_m` in terse mode unless a later task adds elevation conversion; raw upstream `total_elevation_gain` names appear only in `include_full`.
- Raw upstream field names (`distance`, `icu_distance`, `average_speed`, `calories`, etc.) are never used in terse rows when the field needs unit or semantic disambiguation.

`include_full` and upstream `fields` strategy:

- `internal/intervals.Activity` uses pointer fields for nullable stable fields and stores the original object as `Raw map[string]any` or `json.RawMessage` during unmarshal so `include_full` can preserve upstream nulls and unmodeled fields.
- `include_full=true` sends no upstream `fields` query because the public API says `fields` excludes null values; full mode must inspect the raw upstream shape.
- Terse mode may use a conservative `fields` allowlist for payload control, but it must always include pagination/sorting fields (`id`, `start_date_local`, `start_date`), display fields (`name`, `type`, `sub_type`, `timezone`), unit-conversion fields (`distance`, `icu_distance`, `moving_time`, `elapsed_time`, `average_speed`, `max_speed`, elevation fields), core metrics, and Strava/hidden detection fields (`source`, `_note`, `icu_athlete_id`, `external_id`, `stream_types`).
- The token records the effective allowlist so a later page cannot silently switch from terse to full or lose Strava marker fields.

Registry/client wiring plan:

- Add `ActivitiesClient` in `internal/tools` with `ListActivities(ctx, intervals.ListActivitiesParams) ([]intervals.Activity, error)`; keep `ProfileClient` for unit/timezone lookup.
- Update registry construction to accept one dependency that may satisfy both interfaces (`intervals.Client` in app wiring) or explicit `RegistryOptions` clients if tests need fakes. Avoid globals; handlers receive interfaces via constructors.
- `get_activities` fetches the athlete profile once per call through `ProfileClient` to determine `preferred_units`/timezone; if profile lookup fails, return a short user error rather than silently defaulting units. Future caching can be added in registry-level dependency wrappers, not hidden in package state.
- Update `internal/app` default wiring, registry tests, and fakes so `intervals.Client` backs both profile and activity-list reads.

Step 2 targeted tests before code review:

- `internal/intervals` httptest coverage for `ListActivities` path/query construction, no-network behavior, full raw/null preservation, and retry/error propagation.
- `internal/tools/get_activities_test.go` coverage for page-token round trip, invalid base64/version/mismatched-filter token user errors, `include_unnamed=false` filtering across same-timestamp boundaries and all-filtered windows, Strava-blocked row shaping, `include_full` preserving raw nulls, metric vs imperial field names and converted distance/speed/pace values, and registry/schema exposure.
- Targeted command after implementation: `go test ./internal/intervals ./internal/tools ./internal/app` before Step 2 code review; full `make test`, `make build`, and `make lint` remain in Step 6.

## Step 3: Implement `get_activity_details`, `_intervals`

**Status:** ✅ Complete

- [x] Terse default; `include_full` for the raw payload
- [x] Interval rows use the canonical unit enum from TP-008
- [x] Strava-blocked single-activity reads return the same `unavailable` shape, not a 4xx upstream propagation
- [x] R001 code fix: make `get_activity_intervals` return the structured Strava unavailable marker for hidden/stub success payloads and not-found/forbidden fallbacks confirmed by details lookup
- [x] R001 code fix: surface the top-level raw `IntervalsDTO` payload when `include_full:true`

## Step 4: Implement `get_activity_streams` and `get_activity_splits`

**Status:** ✅ Complete

- [x] Streams: canonicalize every key to snake_case via TP-008; unknown keys surface in `_meta.unknown_stream_keys`
- [x] Streams are heavy — require explicit `include_full: true` or an explicit `keys: [...]` arg listing requested channels; default should not return full streams
- [x] Splits: when no manual laps exist, compute per-km or per-mile virtual splits from streams (upstream public discussion); honour `preferred_units`
- [x] Document the splits algorithm and edge cases (paused segments, missing samples) in package doc

## Step 5: Implement `get_activity_messages`

**Status:** ✅ Complete

- [x] List messages on a given activity, terse-by-default
- [x] Render timestamps in the athlete's configured TZ
- [x] R001 code fix: add confirmed Strava-unavailable fallback for message fetch errors using activity details lookup
- [x] R001 code fix: preserve context cancellation/deadline errors during profile and message fetches
- [x] R001 code fix: validate and cap `limit`, expose effective limit metadata, and improve message argument schema descriptions
- [x] R001 code fix: add intervals httptest coverage for activity messages path/query and raw null preservation
- [x] R001 code fix: preserve explicit `seen:false`/`deleted:false` message booleans in terse responses
- [x] R001 code fix: propagate cancellation/deadline errors from the Strava fallback detail lookup
- [x] R001 code fix: reject negative `since_id`, document schema minimum, and expose effective `since_id` metadata
- [x] R001 status/test hygiene: remove stray Blockers review rows and assert messages tool registration

## Step 6: Tests

**Status:** ✅ Complete

- [x] Table-driven tests using `httptest.Server` + `testdata/` fixtures; never hit the network
- [x] Cover: pagination round-trip; `include_unnamed` filter; `include_full` opt-in; Strava-blocked rows; stream-key canonicalization; splits for both unit systems; TZ rendering on messages
- [x] `make test`, `make build`, `make lint` pass

## Discoveries

| Date | Area | Discovery |
| ---- | ---- | --------- |

## Blockers

_None._
| 2026-05-11 18:32 | Review R001 | code Step 5: APPROVE |
