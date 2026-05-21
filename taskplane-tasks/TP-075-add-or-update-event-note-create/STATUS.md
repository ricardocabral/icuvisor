# TP-075-add-or-update-event-note-create — Status

**Current Step:** Step 6: Close the issue
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 14
**Iteration:** 1
**Size:** M
**Closes:** #6
**Requires live API access:** YES (.env-dev test athlete)

---

### Step 1: Live probe to isolate the upstream contract

**Status:** ✅ Complete

- [x] Source `.env-dev` to get the test athlete credentials.
- [x] Out-of-process probe: POST minimal NOTE event directly to intervals.icu, varying date format, `type`, and description one field at a time.
- [x] Record the accepted request/response combination in sanitized `internal/intervals/testdata/events/note_create_request.json` and `note_create_response.json` fixtures.
- [x] Clean up every probe event created on the test athlete and verify no probe events remain.

- [x] R002: Redact live-account calendar metadata from `note_create_response.json`.
- [x] R002: Complete and document the probe matrix, including category casing and name/description combinations.

**Discovery:** Direct bulk NOTE create accepts `start_date_local` as `YYYY-MM-DDT00:00:00` with uppercase `category: "NOTE"` and a non-empty `name`; `category: "Note"` and `"note"` are rejected with HTTP 400 JSON parse errors. Date-only `YYYY-MM-DD` is rejected with HTTP 422. `type` may be omitted, `""`, or `"Note"`. Name-only and name+description NOTE creates are accepted; description without name is rejected with HTTP 422 `Name is required`.

### Step 2: Add a failing test

**Status:** ✅ Complete

- [x] Add an intervals client serialization regression test for NOTE create using the captured fixtures as the expected outbound body/response.
- [x] Add a tool validation regression test showing NOTE creates require a non-empty `name`.
- [x] Confirm the NOTE regression tests fail before the tool fix.

### Step 3: Fix the tool

**Status:** ✅ Complete

- [x] Serialize date-only NOTE writes as `YYYY-MM-DDT00:00:00`, preserving existing WORKOUT behavior and leaving unprobed categories unchanged.
- [x] Require a non-empty `name` only for NOTE creates, and update schema text plus the public invalid-arguments message.
- [x] Run targeted NOTE regression plus existing WORKOUT create/update tests and confirm they pass.

### Step 4: Build + lint + race + live re-validation

**Status:** ✅ Complete

- [x] Run `make build`, `make test`, `make test-race`, and `make lint` successfully.
- [x] Live re-validate the built binary via stdio MCP using date-only NOTE create arguments with a unique name and no `type`; verify via `get_events`, delete the returned event ID with `.env-dev` delete mode, then re-run `get_events` and confirm the unique name is absent.

### Step 5: Document amendment

**Status:** ✅ Complete

- [x] Document the live-probed NOTE create payload contract in `docs/upstream-gaps/event-note-payload.md`.

### Step 6: Close the issue

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` under `[Unreleased] → Fixed` with the NOTE create fix and root cause.
- [x] Update `STATUS.md` with final delivery/issue-close notes.
- [x] Commit the final delivery changes with a `Closes #6` reference.

**Delivery note:** The final delivery commit includes `Closes #6` so the issue closes when this lane is merged; no probe or live-validation NOTE events remain on the test athlete.

| 2026-05-17 01:55 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 01:55 | Step 1 started | Live probe to isolate the upstream contract |
| 2026-05-17 01:58 | Review R001 | plan Step 1: APPROVE |
| 2026-05-17 02:05 | Review R002 | code Step 1: APPROVE |
| 2026-05-17 02:07 | Review R003 | plan Step 2: UNKNOWN |
| 2026-05-17 02:11 | Review R004 | code Step 1: APPROVE |
| 2026-05-17 02:13 | Review R005 | plan Step 2: APPROVE |
| 2026-05-17 02:17 | Review R006 | code Step 2: APPROVE |
| 2026-05-17 02:20 | Review R007 | plan Step 3: UNKNOWN |
| 2026-05-17 02:22 | Review R008 | plan Step 3: APPROVE |
| 2026-05-17 02:26 | Review R009 | code Step 3: APPROVE |
| 2026-05-17 02:28 | Review R010 | plan Step 4: UNKNOWN |
| 2026-05-17 02:29 | Review R011 | plan Step 4: APPROVE |
| 2026-05-17 02:34 | Review R012 | code Step 4: APPROVE |
| 2026-05-17 02:37 | Review R013 | plan Step 5: APPROVE |
| 2026-05-17 02:39 | Review R014 | code Step 5: APPROVE |

| 2026-05-17 02:43 | Worker iter 1 | done in 2891s, tools: 166 |
| 2026-05-17 02:43 | Task complete | .DONE created |