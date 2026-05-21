# TP-085: Concrete Strava-import unavailable workaround text — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 1
**Review Counter:** 7
**Iteration:** 2
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Audit current unavailable wording
**Status:** ✅ Complete

- [x] Find all Strava/import-blocked unavailable marker construction paths.
- [x] Audit `stravaWorkaround`, `strava_tos`, `strava_blocked`, and `StravaImported` search results, including constructor paths outside the original file scope or log exclusions.
- [x] Identify whether native provider can be inferred from current payloads.
- [x] Define safe provider inference rules that avoid implying native providers from ambiguous Strava/sync-chain evidence.
- [x] Define provider-aware and provider-unknown workaround strings.
- [x] Record Step 1 acceptance notes in STATUS.md discoveries: constructor list/reason codes, provider-inference availability, exact strings, and any Step 2/3 file-scope expansion.

---

### Step 2: Update marker text
**Status:** ✅ Complete

- [x] Return the concrete intervals.icu Connections-page remedy when Strava-blocked data is detected.
- [x] Add a shared workaround builder that infers only allowlisted native providers from explicit raw payload evidence.
- [x] Apply the shared workaround builder to list rows, messages, intervals, streams/splits fallback, and extended metrics constructors.
- [x] Mention provider name when known; use provider-neutral wording when unknown.
- [x] Keep `unavailable.reason` stable and structured.

---

### Step 3: Fixture assertions
**Status:** ✅ Complete

- [x] Add/modify fixtures to assert the exact workaround string, not only `reason`.
- [x] Add helper-level assertions for Wahoo external-id provider inference and unknown-provider fallback.
- [x] Add response-level assertions covering at least list rows, shared fallback streams/splits, and one direct constructor outside the shared fallback.
- [x] Add exact response-level workaround and reason assertions for every identified constructor: list rows, messages, intervals, streams/splits, and extended metrics.
- [x] Cover at least one known native provider and one unknown-provider case.
- [x] Run targeted Strava/unavailable tests.

---

### Step 4: Docs and verification
**Status:** ✅ Complete

- [x] Update troubleshooting docs with the same remedy wording.
- [x] Update CHANGELOG.md.
- [x] Run full quality gate.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | inline reviewer verdict |
| R003 | Plan | 2 | APPROVE | .reviews/R003-plan-step2.md |
| R004 | Plan | 3 | REVISE | .reviews/R004-plan-step3.md |
| R005 | Plan | 3 | APPROVE | inline reviewer verdict |
| R006 | Plan | 4 | APPROVE | inline reviewer verdict |
| R007 | Plan | 5 | APPROVE | inline reviewer verdict |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Strava-unavailable constructors found: `activityRow` and `stravaUnavailableMessagesResponse` use reason `strava_tos`; `detectActivityUnavailable` fallback helper, `stravaUnavailableIntervalsResponse`, and `stravaUnavailableExtendedMetricsResponse` use reason `strava_blocked`. Streams and splits consume the shared fallback helper. | Step 2 must update all constructors via shared helper/constant while keeping reason codes stable. | `internal/tools/get_activities_row.go`, `get_activity_messages.go`, `activity_unavailable.go`, `get_activity_details.go`, `get_activity_streams.go`, `get_extended_metrics.go` |
| Provider inference can use raw blocked activity payloads on every path: list/detail rows receive `intervals.Activity`, messages and extended metrics receive fallback/detail activity, intervals direct stub gets DTO raw, and streams/splits receive raw through `detectActivityUnavailable`. | Implement inference from explicit allowlisted native-provider evidence only. | `internal/intervals/activities.go`, `internal/tools/*activity*` |
| Safe provider inference rule: infer only allowlisted native providers Garmin, Wahoo, Coros, Suunto, Polar from explicit `external_id` prefixes or `device_name` text; do not infer from `source: Strava`, `_note`, or unallowlisted sync-chain prefixes such as MyWhoosh or TrainerRoad. | Step 2/3 should cover a known Wahoo case and an unknown-provider case. | `internal/intervals/testdata/activities/strava_sync_chain_empty_stubs.json` |
| Target provider-aware workaround: `Open the intervals.icu Connections page, choose Wahoo, and click Download old data so historical activities are re-imported directly from Wahoo instead of through Strava's restricted API.` Target unknown-provider workaround: `Open the intervals.icu Connections page for the activity's original device provider and click Download old data so historical activities are re-imported directly from that provider instead of through Strava's restricted API.` | Reuse exact wording in code tests and troubleshooting docs, substituting the inferred provider name when known. | Step 2/3/4 |
| File scope expands beyond prompt list to include `internal/tools/get_activity_messages.go` and `internal/tools/get_extended_metrics.go` for consistent Strava-unavailable wording. | Include these files in code and fixture assertions. | Step 2/3 |
| Check-if-affected review found README and generated tool reference did not describe Strava-unavailable workaround details; PRD had stale generic workaround wording and was updated to match the concrete Connections-page remedy. | Documentation scope completed without generated tool catalog changes. | `README.md`, `web/content/reference/tools.md`, `docs/prd/PRD-icuvisor.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 15:23 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 15:23 | Step 0 started | Preflight |
| 2026-05-20 15:42 | Worker iter 1 | done in 1119s, tools: 83 |
| 2026-05-20 16:14 | Exit intercept reprompt | Supervisor provided instructions (741 chars) — reprompting worker |
| 2026-05-20 16:25 | Worker iter 2 | done in 2566s, tools: 112 |
| 2026-05-20 16:25 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

Plan review R001 requested a tighter Step 1 audit plan covering all Strava-unavailable constructors, safe provider inference rules, exact target strings, and explicit Step 1 acceptance notes.
Plan review R004 requested exact response-level workaround and reason assertions for every identified Strava-unavailable constructor path, plus replacement of stale old-wording substring expectations.
| 2026-05-20 15:26 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 15:27 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 15:30 | Review R003 | plan Step 2: APPROVE |
| 2026-05-20 15:44 | Review R004 | plan Step 3: REVISE |
| 2026-05-20 15:45 | Review R005 | plan Step 3: APPROVE |
| 2026-05-20 16:18 | Review R006 | plan Step 4: APPROVE |
| 2026-05-20 16:21 | Review R007 | plan Step 5: APPROVE |
