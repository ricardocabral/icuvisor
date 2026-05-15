# TP — Status

**Issue:** v0.2 — read path
**Review Level:** 1
**Status:** ✅ Complete
**Iteration:** 2
**Current Step:** Complete
**Last Updated:** 2026-05-15
**State:** Complete under operator-approved invited-athlete deferral; no outreach or fabricated participant data

_Task scaffolded from PROMPT.md; execution hydrated from PROMPT.md Step 1._

## Step 1: Assemble the prompt set

**Status:** ✅ Complete

- [x] Build a canonical prompt set covering each PRD §6 customer job (1–5, except #3 which requires writes)
- [x] Include at least one prompt per registered read tool, and one prompt per cluster (activities, fitness, wellness, events, workout library, custom items, periodization if shipped)
- [x] Include three adversarial prompts: (a) ask the LLM to compute a metric that requires a unit it lacks (force unit-system reasoning); (b) ask about Strava-imported activities (force the `unavailable` shape to surface correctly); (c) ask "did I sleep well" against a row with both `sleepQuality` and `sleepScore` (force dual-scale reporting)
- [x] Record the prompt set in `docs/dogfood/v0.2-prompts.md` with redactions

## Step 2: Solo dogfood against your own account

**Status:** ✅ Complete

- [x] Run the prompt set through Claude Desktop or Codex (per TP-006 recipe) against the maintainer's own intervals.icu account
- [x] For each prompt, record: tool calls made, response sizes (bytes), whether the LLM's final answer was correct, any scale / unit / Strava-detection failure
- [x] Measure the largest single response against the §7.2.D 30k-token soft ceiling; flag any tool that exceeds it

## Step 3: Invited athletes (2–3) read-only

**Status:** ✅ Complete — operator-approved invited-athlete deferral documented

- [x] R001 plan revision: define participant protocol, inclusion targets, and consent/privacy boundaries without naming people
- [x] R001 plan revision: pin exact run instructions, allowed clients, prompt set, and fresh-session requirements
- [x] R001 plan revision: add a redacted invited-athlete findings template and explicit forbidden-data list
- [x] R001 plan revision: document local response-size measurement, temporary-artifact cleanup, coverage fallback rules, and Step 4 triage boundary
- [x] R002 plan hardening: require invites to name the exact release artifact or git revision participants should run
- [x] Deferred/N/A by operator-approved acceptance change: recruit 2–3 forum-friendly athletes (one ideally an miles/imperial user; one ideally using a non-Garmin bridge like Polar or Oura to exercise wellness provenance). No outreach or recruitment was attempted.
- [x] Deferred/N/A by operator-approved acceptance change: provide participants the manual-config recipe (v0.1 docs) and have them run the same prompt set. The recipe/protocol is preserved in `docs/dogfood/v0.2-findings.md`; participant execution remains maintainer follow-up.
- [x] Deferred/N/A by operator-approved acceptance change: collect findings via a redacted template. No participant findings were collected or fabricated.
- [x] Deferred/N/A by operator-approved acceptance change: aggregate invited-athlete findings into `docs/dogfood/v0.2-findings.md`. The doc records that external participant rows do not exist for this batch and are not fabricated.

## Step 4: Triage findings

**Status:** ✅ Complete — solo triage accepted; invited-athlete-specific triage deferred

- [x] For each scale / unit / provenance / Strava-detection failure: open a GitHub issue tagged `v0.2-followup` linking the specific tool task — solo failures covered by issues #11 and #12; invited-athlete-specific failures deferred until real results exist
- [x] For latency / token-budget regressions: confirm against KR4 / KR5 targets; if a tool exceeds the soft 30k-token ceiling, open a follow-up issue for pagination or shape tightening — solo maximum was `get_wellness_data` at ~10.4k tokens, below ceiling; verbosity follow-up tracked in issue #12
- [x] Decide which findings are launch-blocking for v0.5 vs follow-up; record the call in `STATUS.md` — no additional v0.5 blocker for revised acceptance beyond issues #11/#12; real participant-result triage remains maintainer follow-up

Revised triage completed under operator-approved deferral: issue #11 covers activity detail read fetch failures from TP-009; issue #12 covers `get_workouts_in_folder` default verbosity from TP-013. Invited-athlete-specific triage is deferred until real participant results exist.

## Step 5: Sign-off

**Status:** ✅ Complete

- [x] Update `ROADMAP.md` v0.2 to check off the dogfood item — revised wording records solo dogfood completion and invited-athlete follow-up deferral
- [x] If any code/doc changed, run `make test`, `make build`, `make lint`, update `CHANGELOG.md` — `make test`, `make build`, and `make lint` passed on 2026-05-15; no user-visible behavior change, so `CHANGELOG.md` not updated
- [x] Confirm no athlete API keys, raw personal data, or training-load values are committed — inspected diff/status; only redacted docs/status/roadmap/review files are included

## Notes

- Step 1 plan inputs read: PRD §6/§7.2.C/§7.2.D, ROADMAP.md v0.2, TP-006 recipe, and TP-009…TP-015 STATUS summaries.
- Registered v0.2 read tools in the current registry: `get_athlete_profile`, `get_fitness`, `get_training_summary`, `get_wellness_data`, `get_best_efforts`, `get_power_curves`, `get_activities`, `get_events`, `get_event_by_id`, `get_training_plan`, `get_workout_library`, `get_workouts_in_folder`, `get_custom_items`, `get_custom_item_by_id`, `get_activity_details`, `get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, `get_activity_messages`, `get_extended_metrics`.
- `get_planning_parameters` / periodization parameters are not shipped in v0.2; TP-014 documents deferral pending upstream API exposure.

## Blockers

| Date       | Blocker                                                                                                                                                                                             | Attempts                                                                                                                                                                                                                                                                                                                    | Current Impact                                                                                                                                                                                       |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 2026-05-12 | Step 3 invited-athlete recruitment requires external human outreach and participant consent; the worker cannot recruit forum athletes, access private contacts, or post on the maintainer's behalf. | Created the participant protocol, run instructions, consent/privacy note, redacted findings template, coverage/fallback rules, and local measurement/cleanup instructions in `docs/dogfood/v0.2-findings.md`; plan re-review returned APPROVE. Did not attempt unauthenticated forum posting or impersonate the maintainer. | Resolved for this batch by 2026-05-15 operator-approved acceptance change: external invited-athlete validation is deferred to maintainer follow-up, with no outreach or fabricated participant data. |

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

## Execution Log

| 2026-05-12 06:38 | Task started | Runtime V2 lane-runner execution |
| 2026-05-12 06:38 | Step 1 started | Assemble the prompt set |
| 2026-05-12 | Hydrated | Expanded STATUS.md with Step 1 checkboxes and prompt-set inputs |
| 2026-05-12 07:27 | Exit intercept reprompt | Supervisor provided instructions (445 chars) — reprompting worker |
| 2026-05-12 07:29 | Review R001 | plan Step 1: APPROVE |
| 2026-05-12 | Step 1 prompt set recorded | `docs/dogfood/v0.2-prompts.md` contains 27 redacted prompts covering all current registered read tools, clusters, PRD §6 read-only jobs, and three adversarial cases |
| 2026-05-12 | Step 2 started | Solo dogfood against maintainer account |
| 2026-05-12 | Review R001 | plan Step 2: UNAVAILABLE |
| 2026-05-12 | Step 2 solo dogfood run | Codex CLI ran the 27-prompt set against the maintainer account using credentials from an untracked parent `.env`; raw temp logs were removed after extracting redacted results |
| 2026-05-12 | Step 2 largest response | `get_wellness_data` measured 41,459 bytes (~10.4k tokens), below the PRD §7.2.D 30k-token soft ceiling |
| 2026-05-12 | Step 3 started | Invited-athlete read-only validation |
| 2026-05-12 | Review R001 | plan Step 3: REVISE — add external-participant privacy/run protocol and redacted findings template before recruitment |
| 2026-05-12 | Step 3 plan revisions applied | Added invited-athlete protocol, run instructions, redacted template, measurement/cleanup guidance, fallback rules, and triage boundary to `docs/dogfood/v0.2-findings.md` |
| 2026-05-12 | Review R002 | plan Step 3: APPROVE |
| 2026-05-12 | Step 3 blocker | External recruitment and participant-run collection require maintainer action and cannot be completed by the worker without contacts/consent |
| 2026-05-12 07:52 | Exit intercept reprompt | Supervisor provided instructions (404 chars) — reprompting worker |
| 2026-05-12 08:09 | Review R001 | plan Step 3: REVISE |
| 2026-05-12 08:15 | Review R001 | plan Step 3: APPROVE |
| 2026-05-12 08:17 | Exit intercept timeout | Supervisor did not respond within 60s — closing session |
| 2026-05-12 08:17 | Agent escalate | Blocked at TP-016 Step 3. I completed and committed the invited-athlete protocol, exact run instructions, consent/privacy note, redacted findings template, local measurement/cleanup guidance, and got |
| 2026-05-12 08:17 | Worker iter 1 | done in 5915s, tools: 90 |
| 2026-05-12 08:17 | Soft progress | Iteration 1: 0 new checkboxes but uncommitted source changes detected — not counting as stall |
| 2026-05-12 08:17 | Step 1 started | Assemble the prompt set |
| 2026-05-12 | Steering | Supervisor confirmed Step 3 is genuinely blocked on external invited-athlete recruitment; worker must not attempt outreach/posting/impersonation/external account actions and must preserve local deliverables only. |
| 2026-05-12 | Step 3 blocked deliverables preserved | `docs/dogfood/v0.2-findings.md` records explicit blocked status, participant protocol, consent text, run instructions, redacted template, coverage fallback rules, and forbidden-data list. |
| 2026-05-12 08:21 | Exit intercept close | Supervisor directed session close: "let it fail" |
| 2026-05-12 08:21 | Agent reply | Acknowledged. TP-016 remains blocked at Step 3; I preserved the invited-athlete protocol/local deliverables, recorded the explicit blocked status and current impact in docs/status, committed them as 1 |
| 2026-05-12 08:21 | Worker iter 2 | done in 267s, tools: 15 |
| 2026-05-12 08:21 | No progress | Iteration 2: 0 new checkboxes (1/3 stall limit) |
| 2026-05-12 08:21 | Step 1 started | Assemble the prompt set |
| 2026-05-12 | Resume check | Step 1 and Step 2 are already complete; resumed at Step 3 and confirmed the remaining recruitment/run/collection checkboxes are still blocked on maintainer-supplied invited-athlete results. |
| 2026-05-12 08:24 | Exit intercept close | Supervisor directed session close: "let it fail" |
| 2026-05-12 08:24 | Agent escalate | Blocked at TP-016 Step 3. Step 1 and Step 2 are complete, Step 3 plan/protocol deliverables are preserved and committed, and the remaining checkboxes require external maintainer action: recruit 2–3 in |
| 2026-05-12 08:24 | Worker iter 3 | done in 184s, tools: 13 |
| 2026-05-12 08:24 | No progress | Iteration 3: 0 new checkboxes (2/3 stall limit) |
| 2026-05-12 08:24 | Step 1 started | Assemble the prompt set |
| 2026-05-12 | Steering | Supervisor directed: "let it fail". Remaining Step 3 invited-athlete recruitment/run/collection checkboxes remain blocked on external maintainer action and should not be completed by the worker. |
| 2026-05-12 08:26 | Exit intercept close | Supervisor directed session close: "let it fail" |
| 2026-05-12 08:26 | ⚠️ Steering | let it fail |
| 2026-05-12 08:26 | Worker iter 4 | done in 97s, tools: 7 |
| 2026-05-12 08:26 | No progress | Iteration 4: 0 new checkboxes (3/3 stall limit) |
| 2026-05-12 08:26 | Task blocked | No progress after 3 iterations |
| 2026-05-14 18:55 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 18:55 | Step 1 started | Assemble the prompt set |
| 2026-05-14 | Resume check | Step 1 and Step 2 remain complete; Step 3 protocol/template deliverables remain present in `docs/dogfood/v0.2-findings.md`. Remaining invited-athlete recruitment, participant runs, collection, Step 4 triage, and Step 5 sign-off are still blocked on external maintainer action/results. |
| 2026-05-14 19:01 | Exit intercept timeout | Supervisor did not respond within 60s — closing session |
| 2026-05-14 19:01 | Agent escalate | Blocked at TP-016 Step 3. Step 1 and Step 2 are complete, and the Step 3 participant protocol, run instructions, consent/privacy note, redacted findings template, cleanup/measurement guidance, fallbac |
| 2026-05-14 19:01 | Agent reply | Acknowledged. TP-016 remains blocked at Step 3; I will not attempt external outreach or fabricate invited-athlete results. STATUS.md records the steering and blocked state so the run can fail as direc |
| 2026-05-14 19:01 | Worker iter 1 | done in 176s, tools: 13 |
| 2026-05-14 19:01 | No progress | Iteration 1: 0 new checkboxes (1/3 stall limit) |
| 2026-05-14 19:01 | Step 1 started | Assemble the prompt set |
| 2026-05-14 | Resume correction | Current step reset to Step 3 because Step 1 and Step 2 are already complete; remaining invited-athlete recruitment/run/collection work is blocked on maintainer-supplied participants and redacted results. |
| 2026-05-14 | Step 3 plan hardening | Added invite guidance requiring the maintainer to name the exact release artifact or git revision each participant should run. |
| 2026-05-14 19:05 | Review R001 | plan Step 3: APPROVE |
| 2026-05-14 | Validation commands | `make test`, `make build`, and `make lint` passed after docs/status/roadmap updates. |
| 2026-05-14 | Step 4 partial triage | Opened v0.2 follow-up issue #11 for solo activity detail read fetch failures and issue #12 for `get_workouts_in_folder` default verbosity; invited-athlete evidence is still required before completing Step 4. |
| 2026-05-14 | Roadmap correction | Left the v0.2 dogfood ROADMAP checkbox unchecked because invited-athlete validation is not complete and sign-off remains blocked. |
| 2026-05-14 | Blocker escalation | Escalated that remaining Step 3 recruitment/run/collection, Step 4 final triage, and Step 5 sign-off require maintainer-supplied invited-athlete results; no outreach or fabricated data attempted. |
| 2026-05-14 19:10 | Exit intercept timeout | Supervisor did not respond within 60s — closing session |
| 2026-05-14 19:10 | Agent escalate | TP-016 remains blocked at Step 3 on maintainer-supplied invited-athlete recruitment/results. I did not attempt external outreach or fabricate participant data. New progress this iteration: hardened th |
| 2026-05-14 19:10 | Worker iter 2 | done in 505s, tools: 55 |
| 2026-05-14 19:10 | No progress | Iteration 2: 0 new checkboxes (2/3 stall limit) |
| 2026-05-14 19:10 | Step 1 started | Assemble the prompt set |
| 2026-05-14 | Resume correction | Current step reset to Step 3 because Step 1 and Step 2 are complete; invited-athlete recruitment/run/collection remains blocked on external maintainer action and redacted participant results. |
| 2026-05-15 09:42 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 09:42 | Step 1 started | Assemble the prompt set |
| 2026-05-15 | Steering | Operator-approved acceptance change: external invited-athlete validation is deferred; solo dogfood plus documented invited-athlete protocol/template is accepted for this batch. No outreach, recruitment, external account actions, or fabricated athlete findings should be performed. |
| 2026-05-15 09:44 | Review R001 | plan Step 3: APPROVE |
| 2026-05-15 09:47 | Review R001 | plan Step 4: APPROVE |
| 2026-05-15 09:49 | Review R001 | plan Step 5: APPROVE |
| 2026-05-15 | Validation commands | `make test && make build && make lint` passed. |
| 2026-05-15 | Secret/raw-data review | Inspected diff/status for API keys, athlete IDs, raw personal data, raw payloads, exact private training-load values, and temp logs; none are committed. |
| 2026-05-15 09:53 | Review R001 | code Step 5: REVISE |
| 2026-05-15 09:55 | Review R001 | code Step 5: APPROVE |
| 2026-05-15 09:57 | Review R001 | code Step 5: APPROVE |
| 2026-05-15 10:00 | Exit intercept timeout | Supervisor did not respond within 60s — closing session |
| 2026-05-15 10:00 | Agent reply | Received operator-approved acceptance change for TP-016. I will complete by documenting the invited-athlete deferral, using solo findings/issues #11/#12 for triage, updating ROADMAP/STATUS/docs, runni |
| 2026-05-15 10:00 | Worker iter 1 | done in 1059s, tools: 65 |
| 2026-05-15 10:00 | No progress | Iteration 1: 0 new checkboxes (1/3 stall limit) |
| 2026-05-15 10:00 | Step 1 started | Assemble the prompt set |
| 2026-05-15 | Resume correction | STATUS.md and `.DONE` already mark TP-016 complete; current step reset to Complete so the orchestrator does not restart finished checkboxes. |
