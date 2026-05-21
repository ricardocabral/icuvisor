# TP-074-activity-detail-fetch-fallback — Status

**Current Step:** Step 6: Close the issue
**Status:** ✅ Complete
**Last Updated:** 2026-05-16
**Review Level:** 2
**Review Counter:** 15
**Iteration:** 3
**Size:** M
**Closes:** #11

---

### Step 1: Add failing tests for each error class

**Status:** ✅ Complete

- [x] Add compile-safe, table-driven activity detail read tool tests covering strava_blocked, not_found, unauthorized, rate_limited, upstream_unavailable_500, and upstream_unavailable_400 for all four tools using route-based servers or existing fake-client call sites.
- [x] Update existing Strava expectation tests in the activity detail read tools from strava_tos to strava_blocked without changing unrelated tool contracts.
- [x] Assert the structured unavailable response shape: handler success, workaround only for strava_blocked, no fabricated detail data on unavailable responses, and success-path tests remain unchanged.
- [x] Target each tool's actual terminal error source in tests: intervals from GetActivityIntervals, streams from GetActivityStreams plus detail lookup for Strava detection, splits from streams fallback after no manual splits, and extended metrics from primary activity fetch for terminal categories.
- [x] Confirm the new tests fail before the implementation change (targeted `go test ./internal/tools -run 'Test(GetActivityIntervalsUnavailableReasons|GetActivityStreamsUnavailableReasons|GetActivitySplitsUnavailableReasons|ExtendedMetricsUnavailableReasons|GetActivityIntervalsUnavailableForHiddenSuccessPayload|GetActivityIntervalsFallbacksToDetailsForBlockedError|ExtendedMetricsStravaUnavailableIncludesFullWhenRequested)'` fails with current generic errors / strava_tos expectations).

### Step 2: Broaden the fallback predicate

**Status:** ✅ Complete

- [x] Broaden activity-read fallback eligibility to cover ErrUpstream and ErrRateLimited in addition to not_found and unauthorized, while preserving context cancellation behavior.
- [x] Return strava_blocked unavailable responses with workaround when a fallback/detail activity lookup proves the activity is Strava-blocked.
- [x] Return categorized unavailable responses for terminal errors based on the original sentinel: not_found, unauthorized, rate_limited, and upstream_unavailable.
- [x] Keep get_activity_messages on its previous not_found/unauthorized fallback predicate so Step 2 does not broaden out-of-scope message endpoint behavior.

### Step 3: Mirror the shape across the four tools

**Status:** ✅ Complete

- [x] Extract shared activity unavailable helpers: one classifier based on the original endpoint error and one optional Strava-block detector for endpoint failures that are not already GetActivity.
- [x] Add explicit detail-client wiring for streams/splits fallback detection where needed, updating registry and tests without relying on accidental fake-client methods.
- [x] Wire get_activity_intervals, get_activity_streams, get_activity_splits, and get_extended_metrics to return successful structured unavailable payloads while preserving success-path response shapes and omitting fabricated detail collections on unavailable responses.
- [x] Keep extended metrics from refetching GetActivity on primary activity errors; classify those terminal errors directly.
- [x] Use a dedicated interval unavailable response shape so get_activity_intervals does not emit fabricated analyzed:false on unavailable responses.

### Step 4: Fixtures

**Status:** ✅ Complete

- [x] Verify whether fixture files are needed after the fake-client test approach; prefer no unused JSON fixtures when only HTTP status/sentinel behavior is asserted (tests use sentinel fake clients and inline Strava payloads, so no fixture body is consumed).
- [x] Keep fixture coverage minimal by relying on inline/fake test setup for empty 400/403/404/429/500 bodies and existing inline Strava payloads.

### Step 5: Build + lint + race + manual smoke

**Status:** ✅ Complete

- [x] `make build`, `make test`, `make test-race`, `make lint` all pass.
- [x] Manual smoke using `.env-dev` when possible: call `get_activity_intervals` via stdio MCP against a real Strava-imported or v0.2 failing activity and confirm structured `unavailable` shape, or document why credentials/data are unavailable.
  - Manual smoke not run: no `.env-dev` or other local env credentials were present in the worktree (`find ... -name '.env*'` returned no files), so no real athlete/activity data was available for stdio MCP validation.

### Step 6: Close the issue

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` under `[Unreleased] → Fixed` with the structured unavailable activity-detail read behavior.
- [x] Record the issue-closing delivery note in `STATUS.md`, including `Closes #11` PR-body guidance and the post-merge `gh issue close` fallback.
  - Delivery note: include `Closes #11` in the PR body so the issue closes on merge; if it remains open, close issue #11 manually after merge.

| 2026-05-16 22:50 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 22:50 | Step 1 started | Add failing tests for each error class |
| 2026-05-16 22:55 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-16 22:57 | Review R002 | plan Step 1: REVISE |
| 2026-05-16 22:59 | Review R003 | plan Step 1: APPROVE |
| 2026-05-16 23:05 | Review R004 | code Step 1: APPROVE |
| 2026-05-16 23:08 | Review R005 | plan Step 2: APPROVE |
| 2026-05-16 23:13 | Review R006 | code Step 2: REVISE |
| 2026-05-16 23:17 | Review R007 | code Step 2: APPROVE |
| 2026-05-16 23:20 | Review R008 | plan Step 3: REVISE |
| 2026-05-16 23:22 | Review R009 | plan Step 3: APPROVE |
| 2026-05-16 23:30 | Review R010 | code Step 3: REVISE |
| 2026-05-16 23:35 | Review R011 | code Step 3: APPROVE |
| 2026-05-16 23:38 | Review R012 | plan Step 4: APPROVE |

| 2026-05-17 00:28 | Worker iter 1 | done in 5861s, tools: 138 |
| 2026-05-17 00:28 | Step 5 started | Build + lint + race + manual smoke |
| 2026-05-17 01:03 | Exit intercept reprompt | Supervisor provided instructions (716 chars) — reprompting worker |

| 2026-05-17 01:50 | Worker iter 2 | done in 4933s, tools: 16 |
| 2026-05-17 01:53 | Review R015 | code Step 5: APPROVE |

| 2026-05-17 01:55 | Worker iter 3 | done in 298s, tools: 29 |
| 2026-05-17 01:55 | Task complete | .DONE created |
