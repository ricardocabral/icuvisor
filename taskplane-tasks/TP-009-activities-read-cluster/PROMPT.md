# TP-009 — Activities read cluster (list, details, intervals, streams, splits, messages) + Strava detection

## Mission

Ship the activity-side read tools that are the daily-use core of any coaching prompt, with pagination, canonical streams, and Strava-blocked detection that prevents the LLM from hallucinating over upstream-blocked rows.

Roadmap items (ROADMAP.md v0.2):

- `get_activities` (date-range list; `include_unnamed` per issue #67; server-side pagination)
- `get_activity_details`, `get_activity_intervals`, `get_activity_streams` (canonical snake_case keys)
- `get_activity_splits` (virtual per-km / per-mile splits when no manual laps; honours `preferred_units`)
- `get_activity_messages`
- Strava-blocked activity detection returning structured `unavailable: { reason: "strava_tos", workaround: ... }`

PRD anchors: §7.2.C Activities, §7.2.D response shaping and unit normalization.

Complexity: Blast radius 2, Pattern novelty 2, Security 1, Reversibility 1 = 6 → Review Level 2. Size: L (largest v0.2 task — six tools).

## Dependencies

- **TP-007** — response shaping primitives
- **TP-008** — unit enum + stream-key canonicalization

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C Activities, §7.2.D response shaping
- `ROADMAP.md` v0.2
- `internal/intervals/`, `internal/response/`, `internal/streams/`, `internal/units/` from prior tasks
- Public intervals.icu API docs for activities, intervals, streams, splits, messages endpoints

## File Scope

Expected files:

- `internal/intervals/activities.go` (and friends) — typed client methods
- `internal/tools/get_activities.go`
- `internal/tools/get_activity_details.go`
- `internal/tools/get_activity_intervals.go`
- `internal/tools/get_activity_streams.go`
- `internal/tools/get_activity_splits.go`
- `internal/tools/get_activity_messages.go`
- `_test.go` for each, with `testdata/` fixtures
- `CHANGELOG.md`
- `taskplane-tasks/TP-009-activities-read-cluster/STATUS.md`

## Steps

### Step 1: Plan endpoints, types, and pagination shape

- [ ] Identify endpoints, query params, and response shapes from public docs; record uncertainty in `STATUS.md`
- [ ] Decide pagination contract: opaque `next_page_token`, default page size that fits the §7.2.D ~30k-token soft ceiling
- [ ] Decide how `include_unnamed` filters (issue #67) — server-side filter vs client-side post-filter; prefer server-side

### Step 2: Implement `get_activities` with pagination

- [ ] Date-range arguments, `include_unnamed: bool`, `page_size`, `next_page_token`
- [ ] Terse default rows; full payload behind `include_full: true`
- [ ] Strava-blocked rows surface as `unavailable: { reason: "strava_tos", workaround: "connect device directly to intervals.icu (Garmin, Wahoo, Coros, Suunto, Polar)" }` — do not emit `N/A`-laden rows
- [ ] Apply `preferred_units` via TP-007 unit-system plumbing for distance / pace fields

### Step 3: Implement `get_activity_details`, `_intervals`

- [ ] Terse default; `include_full` for the raw payload
- [ ] Interval rows use the canonical unit enum from TP-008
- [ ] Strava-blocked single-activity reads return the same `unavailable` shape, not a 4xx upstream propagation

### Step 4: Implement `get_activity_streams` and `get_activity_splits`

- [ ] Streams: canonicalize every key to snake_case via TP-008; unknown keys surface in `_meta.unknown_stream_keys`
- [ ] Streams are heavy — require explicit `include_full: true` or an explicit `keys: [...]` arg listing requested channels; default should not return full streams
- [ ] Splits: when no manual laps exist, compute per-km or per-mile virtual splits from streams (upstream public discussion); honour `preferred_units`
- [ ] Document the splits algorithm and edge cases (paused segments, missing samples) in package doc

### Step 5: Implement `get_activity_messages`

- [ ] List messages on a given activity, terse-by-default
- [ ] Render timestamps in the athlete's configured TZ

### Step 6: Tests

- [ ] Table-driven tests using `httptest.Server` + `testdata/` fixtures; never hit the network
- [ ] Cover: pagination round-trip; `include_unnamed` filter; `include_full` opt-in; Strava-blocked rows; stream-key canonicalization; splits for both unit systems; TZ rendering on messages
- [ ] `make test`, `make build`, `make lint` pass

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint semantics, especially around Strava-blocked markers and splits derivation. Do not depend on it.
- GPL/copyleft implementation code is off limits.

## Acceptance Criteria

- All six activity read tools registered with the MCP server.
- Pagination on `get_activities` works against fixtures.
- Streams return canonical snake_case keys; unknown keys surface in `_meta`.
- Strava-blocked detection returns the structured `unavailable` shape.
- `preferred_units` flows through distance and pace fields.
- Tests cover the cases in Step 6.

## Do NOT

- Do not implement any write tools.
- Do not bypass the TP-007 response shaping pipeline.
- Do not emit full stream payloads by default — heavy data requires explicit opt-in.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`
- README catalog (add the six tools)

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-009`, for example: `TP-009 add get_activities pagination`.

---

## Amendments

_Add amendments below this line only._
