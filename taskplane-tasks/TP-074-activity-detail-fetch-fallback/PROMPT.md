# TP-074 — Activity detail read tools: broaden upstream-error fallback + structured `unavailable` shape

## Mission

Resolves issue #11 — *TP-016: investigate activity detail read fetch failures*.

The v0.2 dogfood (TP-016) found that `get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, and `get_extended_metrics` returned generic fetch-failure errors against real activity IDs. The tools correctly avoided fabricating data, but the public error was uninformative — operators couldn't distinguish "Strava-blocked activity" from "bad ID" from "intervals.icu hiccup".

Root cause (from triage):

1. `isActivityReadFallbackCandidate` in [`internal/tools/get_activity_details.go:219-221`](../../internal/tools/get_activity_details.go) only catches `intervals.ErrNotFound` and `intervals.ErrUnauthorized`. Upstream 400 / 5xx are mapped to `intervals.ErrUpstream` ([`internal/intervals/errors.go:50-61`](../../internal/intervals/errors.go)) and bypass the Strava-block fallback entirely.
2. The fallback path that *does* detect a Strava-blocked activity already returns a structured `unavailable: { reason, workaround }`. We just need to route more error classes into it.
3. When the fallback also fails (e.g. true 5xx from upstream), the generic message obscures the real cause.

Fix: extend the fallback predicate, and make the terminal error surface a structured `unavailable: { reason: "..." }` shape with categories: `strava_blocked`, `not_found`, `unauthorized`, `upstream_unavailable`, `unknown`.

PRD anchors: §7.2.C activity read cluster; §7.4 reliability; Strava-blocked-activity detection (v0.2 roadmap, already implemented for `get_activities`).
CLAUDE.md: errors back to the LLM "short, actionable, and free of internal stack traces."

Complexity: Blast radius 2 (four tools share the predicate), Pattern novelty 2 (introduces a new categorized `unavailable` shape on a new code path), Security 1, Reversibility 2 = 7 → Review Level 2. Size: M.

## Dependencies

- None. Independent of the other v0.2/v0.3 dogfood follow-ups (TP-072, TP-073, TP-075, TP-076, TP-077).
- Related: the `get_activities` tool already implements `isStravaBlocked` ([`internal/tools/get_activities.go:1127-1130`](../../internal/tools/get_activities.go)) — reuse its detection logic; do not duplicate.

## Context to Read First

- [`docs/dogfood/v0.2-findings.md`](../../docs/dogfood/v0.2-findings.md) — search for T-08, T-09, T-10, T-12 for the exact observed error text.
- [`taskplane-tasks/TP-016-v02-dogfood-validation/STATUS.md`](../TP-016-v02-dogfood-validation/STATUS.md) — original follow-up tracking.
- [`internal/intervals/errors.go:50-61`](../../internal/intervals/errors.go) — `errorForStatus()` HTTP→sentinel mapping. Note 400 is NOT explicitly mapped (falls through to `ErrUpstream`).
- [`internal/intervals/activity_details.go:131-144`](../../internal/intervals/activity_details.go) — `GetActivityIntervals` + sibling client methods.
- [`internal/tools/get_activity_details.go:139-221`](../../internal/tools/get_activity_details.go) — handler + `isActivityReadFallbackCandidate`.
- [`internal/tools/get_activity_streams.go`](../../internal/tools/get_activity_streams.go) — sibling tool error path.
- [`internal/tools/get_activity_splits.go`](../../internal/tools/get_activity_splits.go) — sibling tool error path.
- [`internal/tools/get_extended_metrics.go:97-160`](../../internal/tools/get_extended_metrics.go) — sibling tool error path.
- [`internal/tools/get_activities.go:1127-1130`](../../internal/tools/get_activities.go) — existing `isStravaBlocked` helper.

## File Scope

- `internal/tools/get_activity_details.go` — broaden fallback predicate; add structured `unavailable` shape for terminal errors.
- `internal/tools/get_activity_streams.go`, `internal/tools/get_activity_splits.go`, `internal/tools/get_extended_metrics.go` — mirror the new error shape.
- `internal/intervals/errors.go` — *optional* — add explicit 400 → `ErrBadRequest` mapping if doing so simplifies the predicate; otherwise leave alone.
- `internal/intervals/testdata/` — new fixtures for 400 / 403 / 5xx responses on intervals/streams/splits endpoints.
- New / extended `_test.go` files for each tool covering: Strava block path, 404 path, 5xx path, success path.
- `CHANGELOG.md` — `[Unreleased]` under "Fixed".
- `STATUS.md` (this dir).

Out of scope:
- Refactoring the four tools into a shared base (each is small enough; not worth the abstraction).
- Adding retry logic for transient 5xx — that's `httpretry`'s job and is already configured at client level.
- Changing the success-path response shape.

## Steps

### Step 1: Add failing tests for each error class

- [ ] For each of the four tools, write a test that uses an `httptest.Server` returning, in sequence:
  - 403 on the detail endpoint + a Strava-marked activity payload available via `GetActivity` → expect `unavailable: { reason: "strava_blocked", workaround: "<msg>" }`.
  - 404 → expect `unavailable: { reason: "not_found" }`.
  - 500 → expect `unavailable: { reason: "upstream_unavailable" }`.
  - 400 → expect `unavailable: { reason: "upstream_unavailable" }` (or `"bad_request"` if you choose to map it explicitly).
- [ ] Confirm all the new tests fail on `main` (or partially fail — some 403 cases may already pass).

### Step 2: Broaden the fallback predicate

- [ ] Update `isActivityReadFallbackCandidate` to also accept `intervals.ErrUpstream` (so 400 / 5xx route into the Strava-block check).
- [ ] Inside the fallback: if the `GetActivity` fetch succeeds and `isStravaBlocked(activity)` is true → return the `strava_blocked` `unavailable` shape.
- [ ] If the fallback `GetActivity` also fails or the activity isn't Strava-blocked → return a *categorized* `unavailable` shape based on the original sentinel:
  - `ErrNotFound` → `not_found`
  - `ErrUnauthorized` → `unauthorized`
  - `ErrRateLimited` → `rate_limited`
  - `ErrUpstream` (or anything else) → `upstream_unavailable`

### Step 3: Mirror the shape across the four tools

- [ ] Extract a small helper (e.g. `internal/tools/activity_unavailable.go`) that takes `(ctx, client, athleteID, activityID, err) → (unavailableShape, classifiedErr)` so the four tools share one categorization path.
- [ ] Wire `get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, `get_extended_metrics` to use it.

### Step 4: Fixtures

- [ ] Add `internal/intervals/testdata/activity_intervals_strava_blocked_403.json` (or wire the test server inline if no fixture body is needed — 403 with empty body is enough).
- [ ] Add `internal/intervals/testdata/activity_streams_not_found_404.json`.
- [ ] Add `internal/intervals/testdata/activity_splits_500.json` (empty body, just status).
- [ ] Keep fixtures minimal — favour inline `httptest.Server` over JSON files when there's no shape to assert.

### Step 5: Build + lint + race + manual smoke

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Manual smoke (optional but encouraged) using `.env-dev`: pick a real Strava-imported activity on the test athlete (or any failing activity from the v0.2 findings) and call `get_activity_intervals` via stdio MCP. Confirm the new structured `unavailable` shape.

### Step 6: Close the issue

- [ ] Update `CHANGELOG.md` under `[Unreleased] → Fixed`: `activity detail read tools (get_activity_intervals/streams/splits, get_extended_metrics) now return a structured unavailable:{reason} shape covering strava_blocked, not_found, unauthorized, rate_limited, and upstream_unavailable instead of a generic fetch error.`
- [ ] Update `STATUS.md`.
- [ ] Commit: `fix(activities): categorize unavailable responses for activity detail reads (TP-074, closes #11)`.
- [ ] Reference `Closes #11` in the PR body. After merge, verify auto-close; otherwise close issue #11 manually after merge.

## Acceptance Criteria

- All four tools (`get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, `get_extended_metrics`) emit `unavailable: { reason: "<category>" }` rather than a generic error string when upstream fails.
- `reason` ∈ { `strava_blocked`, `not_found`, `unauthorized`, `rate_limited`, `upstream_unavailable` }. `strava_blocked` also includes a `workaround` field.
- The fallback predicate covers `ErrUpstream` so 400 / 5xx routes through Strava detection.
- Each tool has tests covering all four error classes plus a success-path test.
- `make build`, `make test`, `make test-race`, `make lint` pass.
- Issue #11 closed.

## Do NOT

- Do not change the success-path response shape — schema stability per v0.2 contract.
- Do not retry the upstream call yourself — `httpretry` handles transient 5xx at the client layer.
- Do not log activity IDs or athlete IDs in a way that's hard to scrub later (CLAUDE.md: "Never log API keys, tokens, or raw athlete identifiers in a way that's hard to scrub later").
- Do not introduce a new sentinel error class unless `errors.Is`/`errors.As` callers actually need it. Reuse `ErrUpstream` and friends.

## Documentation

- `CHANGELOG.md` `[Unreleased]` under "Fixed".
- `STATUS.md` in this dir.

## Git Commit Convention

Conventional Commits, prefixed with TP-074. Example:

```
fix(activities): categorize unavailable responses for activity detail reads

TP-074. Closes #11.
```

---

## Amendments

_Add amendments below this line only._
