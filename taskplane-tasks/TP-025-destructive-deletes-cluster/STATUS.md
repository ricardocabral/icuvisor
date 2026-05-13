# TP-025-destructive-deletes-cluster: TP-025-destructive-deletes-cluster — Status

**Current Step:** Step 4: Verify
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-13
**Review Level:** 0
**Review Counter:** 0
**Iteration:** 3
**Size:** M

---

### Step 1: Per-ID deletes

**Status:** ✅ Complete

- [x] `delete_event`, `delete_activity`, `delete_custom_item`, `delete_sport_settings`, `delete_gear`: each takes an opaque ID, returns the deleted ID and a short before-shape echo in `_meta.deleted` so the LLM can confirm
- [x] Registered only in `full` mode (TP-018 `CanDelete`)
- [x] No `confirm` argument

---

### Step 2: `delete_events_by_date_range`

**Status:** ✅ Complete

- [x] Inputs: `start_date`, `end_date` (athlete-TZ; both required; same-day allowed), optional `category` filter
- [x] Hard validation: range size capped (document the cap in the schema description and in `STATUS.md`); reject open-ended ranges
- [x] Response includes `_meta.deleted_count` and the ID list
- [x] Registered only in `full`

---

### Step 3: Tests

**Status:** ✅ Complete

- [x] Per tool: success in `full`, absent from catalog in `safe` and `none`
- [x] `delete_events_by_date_range`: range-cap rejection, athlete-TZ correctness on boundary dates
- [x] Idempotency where upstream supports it (re-delete returns 404 mapped to a typed error, not a 500)

---

### Step 4: Verify

**Status:** 🟨 In Progress

- [x] Update `README.md` catalog and `CHANGELOG.md` for the destructive delete tools
- [x] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete in `full` mode; never in production

#### Ready-to-run manual smoke plan (pending explicit operator approval/scope)

Do not execute this plan until the operator explicitly approves TP-025 destructive-delete smoke. Do not source credentials during preparation. When approved, use only `/Users/jusbrasil/prj/icuvisor/.env-dev`; do not use shell history, copied secrets, production account config, or any other credentials source. The smoke must target only a designated test athlete and must never touch pre-existing production data.

1. From the worktree root, verify `/Users/jusbrasil/prj/icuvisor/.env-dev` exists without printing its contents, then run an approved ephemeral shell that exports only recognized keys from that file and explicitly sets `ICUVISOR_DELETE_MODE=full` for the icuvisor MCP process. Keep command output free of API key and athlete ID values.
2. Build the current binary with `make build`, start a fresh MCP client/session so schemas are not cached, and confirm the full-mode catalog contains `delete_event`, `delete_events_by_date_range`, `delete_custom_item`, `delete_activity`, `delete_gear`, and `delete_sport_settings` with no `confirm` argument.
3. Event single-delete round trip: create a uniquely named future NOTE event via `add_or_update_event` on a clearly disposable future date, capture only the returned event ID in local scratch notes, call `delete_event` on that ID, then verify `get_event_by_id` or bounded `get_events` no longer returns it. This deletes only an artifact created during the same smoke run.
4. Event range-delete round trip: create two uniquely named future NOTE events via `add_or_update_event` on one isolated disposable future date, call `delete_events_by_date_range` with `start_date` equal to `end_date` and a `category` filter that matches those smoke events, then verify `_meta.deleted_count` and `deleted_ids` contain only the smoke-created event IDs and the date no longer lists those events. Do not widen the range beyond the smoke date.
5. Custom-item delete round trip, if the test athlete already has a readable schema sample for a disposable `item_type`: create a uniquely named disposable custom item via `create_custom_item` using content derived from the test account's readable schema requirements, capture the new item ID, call `delete_custom_item`, then verify it is absent from `get_custom_items`. If no safe schema sample is available, mark this sub-check blocked rather than deleting a pre-existing custom item.
6. `delete_activity` requires a maintainer-provided disposable activity ID that was uploaded/imported only for this smoke and is safe to destroy. Do not select an existing training activity, Strava/Garmin sync artifact, race, or any production workout from `get_activities`.
7. `delete_gear` requires a maintainer-provided disposable gear ID that was created specifically for this smoke and is safe to destroy. Do not delete real bikes, shoes, components, sensors, or historical production gear.
8. `delete_sport_settings` requires a maintainer-provided disposable sport-settings ID for an isolated test-only sport/settings row. Do not delete the athlete's real Ride/Run/Swim settings, threshold history, zones, or any row used by production analysis.
9. After approved execution, record only pass/fail/blocked status per tool, deleted smoke artifact IDs or maintainer disposable fixture IDs as redacted placeholders if necessary, and confirmation that `/Users/jusbrasil/prj/icuvisor/.env-dev` was the only credentials source. Do not commit secrets, raw payloads, or local transcripts.

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |

---

## Execution Log

| Timestamp        | Action                     | Outcome                                                                                                                                                                                                  |
| ---------------- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-05-13       | Task staged                | STATUS.md auto-generated by task-runner                                                                                                                                                                  |
| 2026-05-13 20:38 | Task started               | Runtime V2 lane-runner execution                                                                                                                                                                         |
| 2026-05-13 20:38 | Step 1 started             | Per-ID deletes                                                                                                                                                                                           |
| 2026-05-13 21:07 | Agent escalate             | Manual smoke for TP-025 Step 4 is blocked: INTERVALS_ICU_API_KEY and INTERVALS_ICU_ATHLETE_ID are unset, and no .env or icuvisor.json exists in the worktree. I did not run any destructive tools agains |
| 2026-05-13 21:07 | Worker iter 1              | done in 1714s, tools: 114                                                                                                                                                                                |
| 2026-05-13       | Steering received          | Operator forbade TP-025 destructive delete manual smoke for now; no credentials sourced and no live destructive calls performed.                                                                         |
| 2026-05-13       | Manual smoke plan prepared | Ready-to-run plan recorded above; execution remains blocked pending explicit operator approval/scope.                                                                                                    |
| 2026-05-13 21:11 | Agent reply | Acknowledged steering for TP-025: I did not source credentials or run live destructive calls. I documented and committed a ready-to-run destructive-delete manual smoke plan in STATUS.md using only /Us |
| 2026-05-13 21:11 | ⚠️ Steering | Do not run TP-025 destructive delete manual smoke yet. Prior full-mode approval was scoped to TP-022/TP-023/TP-024 smoke-created/restored artifacts, not the full destructive deletes cluster. For now,  |
| 2026-05-13 21:11 | Worker iter 2 | done in 289s, tools: 48 |
| 2026-05-13 21:11 | No progress | Iteration 2: 0 new checkboxes (1/3 stall limit) |
| 2026-05-13       | Blocked                    | Iteration 3 confirmed the only remaining Step 4 item is live destructive-delete manual smoke, and operator steering still forbids running TP-025 smoke without explicit approval/scope. |

---

## Blockers

- Manual smoke is blocked by operator steering pending explicit TP-025 destructive-delete approval/scope. No credentials were sourced, no `/Users/jusbrasil/prj/icuvisor/.env-dev` values were read, and no live destructive calls were run. When approval is granted, use only `/Users/jusbrasil/prj/icuvisor/.env-dev`, delete same-run smoke artifacts where possible, and require maintainer-provided disposable fixture IDs for activity, gear, and sport-settings deletes.
- Iteration 3 remains blocked for the same reason: TP-025 Step 4's only unchecked checkbox is manual smoke against a test athlete in full mode, but supervisor steering explicitly says not to run TP-025 destructive delete manual smoke yet.

---

## Notes

- `delete_events_by_date_range` uses a hard inclusive range cap of 31 athlete-local days.
