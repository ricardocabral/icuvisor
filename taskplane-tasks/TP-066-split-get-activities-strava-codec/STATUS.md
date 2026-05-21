# TP-066-split-get-activities-strava-codec — Status

**Current Step:** Step 5: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S

---

### Step 1: Capture golden tests

**Status:** ✅ Complete

- [x] Add or verify byte-identical `next_page_token` golden coverage for a fixed full-page sweep.
- [x] Add or verify table-driven `isStravaBlocked` golden coverage for Strava marker, N/A heuristics, and manual entries.

### Step 2: Extract Strava heuristic

**Status:** ✅ Complete

- [x] Move `isStravaBlocked` and its package-private helper data to `get_activities_strava.go` without changing behavior.
- [x] Move Strava heuristic golden tests to `get_activities_strava_test.go`.
- [x] Run targeted tool tests for the Strava split.

### Step 3: Extract cursor codec

**Status:** ✅ Complete

- [x] Move activities page token types, encode/decode, request token validation, and cursor progression helpers to `get_activities_cursor.go` without changing token bytes.
- [x] Add or move cursor codec/token tests to focused coverage while preserving byte-identical golden tokens.
- [x] Run targeted tool tests for cursor pagination and token invariants.

### Step 4: Extract row helpers

**Status:** ✅ Complete

- [x] Count row-shaping helper LOC and move row shaping to `get_activities_row.go` because it exceeds the 30 LOC threshold.
- [x] Move row-shaping tests or keep existing handler tests covering units and Strava full rows.
- [x] Run targeted tool tests for row shaping behavior.

### Step 5: Verify

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` `[Unreleased]` with the internal `get_activities` refactor.
- [x] Run `make build`, `make test`, `make test-race`, and `make lint`.
- [x] Verify `scripts/snapshot_tool_schemas.go` produces an empty schema diff.
- [x] Run `wc -l internal/tools/get_activities*.go` and confirm the main file is ≤ ~350 LOC.

| 2026-05-17 09:57 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 09:57 | Step 1 started | Capture golden tests |
| 2026-05-17 10:12 | Review R002 | plan Step 2: APPROVE |
| 2026-05-17 10:17 | Review R003 | plan Step 3: APPROVE |
| 2026-05-17 10:32 | Review R004 | plan Step 4: APPROVE |

| 2026-05-17 10:37 | Worker iter 1 | done in 2403s, tools: 112 |
| 2026-05-17 10:37 | Task complete | .DONE created |