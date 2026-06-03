# TP-151: Event external_id idempotency — Status

**Current Step:** Step 4: Refresh schemas, routing, and docs
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 11
**Iteration:** 4
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Current event external_id handling identified

---

### Step 1: Design external_id contract
**Status:** ✅ Complete

- [x] Create/update/omit/clear semantics decided
- [x] apply_training_plan deterministic ID strategy decided
- [x] Event read-row exposure decided
- [x] Upstream uncertainty recorded
- [x] Retry/preflight behavior with external_id decided

---

### Step 2: Implement event write/read support
**Status:** ✅ Complete

- [x] WriteEventParams and payload support external_id
- [x] add_or_update_event schema/decoder/handler supports external_id
- [x] Create/update and preflight tests added
- [x] Event row exposure implemented/tested as decided
- [x] Targeted tests passing

---

### Step 3: Make apply_training_plan retry-safer
**Status:** ✅ Complete

- [x] Stable plan event external IDs generated
- [x] Deterministic external-ID helper contract pinned with canonical tuple serialization and digest-length tests
- [x] Plan ID/start date/workout ID/relative day/event date tuple threaded into event creation
- [x] Existing matching external_id conflicts are protected before replace_existing deletes same-day workouts
- [x] Dry-run proposed events expose hashed non-leaking external_id
- [x] Repeated apply payload stability tests added
- [x] Dry-run metadata reviewed for safety/usefulness
- [x] Targeted tests passing

---

### Step 4: Refresh schemas, routing, and docs
**Status:** 🟨 In Progress

- [x] Schema snapshots regenerated
- [x] Tool-routing expectations updated if affected
- [ ] Build-workouts documents optional manual external_id usage, stable non-provider namespaces, no secrets, blank ignored/no-clear behavior, and best-effort retry caveats
- [ ] Season-and-block-plan documents deterministic apply_training_plan external IDs, retry review visibility, and same-day/upstream caveats
- [ ] Schema snapshot expectation recorded: only add_or_update_event input schema changes; apply_training_plan snapshot is unchanged because its input schema did not change
- [ ] CHANGELOG updated under [Unreleased] with add_or_update_event.external_id and deterministic apply_training_plan external IDs
- [ ] User docs updated if affected

---

### Step 5: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passes
- [ ] All failures fixed
- [ ] Build passes

---

### Step 6: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Remaining caveats summarized

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE (tool returned APPROVE, review file says not approved) | `.reviews/R001-plan-step1.md` |
| R002 | Plan | Step 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | Step 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | Plan | Step 2 | APPROVE | `.reviews/R004-plan-step2.md` |
| R005 | Code | Step 2 | APPROVE | `.reviews/R005-code-step2.md` |
| R006 | Plan | Step 3 | REVISE | `.reviews/R006-plan-step3.md` |
| R007 | Plan | Step 3 | APPROVE | `.reviews/R007-plan-step3.md` |
| R008 | Code | Step 3 | UNAVAILABLE | _(reviewer produced no file)_ |
| R009 | Code | Step 3 | UNAVAILABLE | _(reviewer produced no file)_ |
| R011 | Plan | Step 4 | REVISE | `.reviews/R011-plan-step4.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Current event write path has no typed `external_id`: `WriteEventParams`/`writeEventPayload` omit it; `add_or_update_event` request/schema omit it; event reads preserve raw payloads but terse rows do not expose `external_id`; `apply_training_plan` creates events without idempotency keys and relies on same-day duplicate matching. | Drives Step 1 contract and Step 2/3 implementation. | `internal/intervals/events.go`, `internal/tools/add_or_update_event.go`, `internal/tools/get_events.go`, `internal/tools/apply_training_plan.go` |
| Upstream acceptance/clear semantics for event `external_id` are not live-probed in this task; treat it as a best-effort idempotency key, keep same-day duplicate preflight, avoid clear/null semantics, and document retry caveats. | Conservative implementation and docs caveat. | Step 1 design |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |
| 2026-06-03 | Step 0 completed | Required files, dependencies, and current external_id handling identified |
| 2026-06-03 | Step 1 started | Design external_id contract |
| 2026-06-03 | Step 1 reviewed | R002 plan and R003 code APPROVE |
| 2026-06-03 | Step 2 started | Implement event write/read support |
| 2026-06-03 21:38 | Worker iter 1 | done in 644s, tools: 61 |
| 2026-06-03 21:39 | Worker iter 2 | done in 66s, tools: 18 |
| 2026-06-03 | Step 2 reviewed | R005 code APPROVE |
| 2026-06-03 | Step 3 started | Make apply_training_plan retry-safer |
| 2026-06-03 | Step 3 plan reviewed | R007 plan APPROVE |
| 2026-06-03 | Step 3 code review attempted | R008/R009 unavailable; proceeding per review protocol |
| 2026-06-03 | Step 4 started | Refresh schemas, routing, and docs |
| 2026-06-03 22:52 | Worker iter 3 | done in 4335s, tools: 113 |
| 2026-06-03 | Review R011 | plan Step 4: REVISE; docs/changelog plan expanded |

---

## Blockers

*None*

---

## Notes

### Step 1 external_id contract

- `add_or_update_event.external_id` is an optional, trimmed non-empty string idempotency key. Creates and updates forward a provided value as upstream `external_id`. Omitting it leaves the payload unchanged; empty/whitespace input is treated as omit. Clearing an existing upstream value is not exposed because upstream clear/null semantics are unproven, and sending empty strings could accidentally create a durable bad key.
- `apply_training_plan` will generate icuvisor-owned deterministic IDs with prefix `icuvisor-plan-v1-` plus a SHA-256-derived hex digest of `(plan_id, start_date, workout_id, relative_day, event_date)`. The hash avoids leaking raw plan/workout IDs in the upstream key, the prefix avoids known provider-owned prefixes such as `strava-`/`hevy-`, and including the anchor start date prevents collisions when the same library plan is applied to different blocks.
- Event read rows will expose `external_id` in terse mode when upstream returns it, and full mode will continue to include the raw payload. This is useful audit metadata for idempotent writes, low-token, and already exposed for activities in terse rows.
- Create preflight remains conservative. For creates with `external_id`, a same-day event with the same `external_id` is treated as an idempotent duplicate even if other writable fields drift; the response should skip creating another event and identify the existing event. Differing or missing `external_id` does not disable the existing exact writable-field duplicate check. Cross-day external-id lookups are not added because the list API is date-windowed and `apply_training_plan` IDs include the event date. Dry-run/proposed plan rows will show the hashed `external_id` for review; raw plan/workout IDs are not embedded in the key.
- R002 non-blocking implementation notes: pin hash input serialization/digest length in code/tests; include duplicate warning/existing event ID when external-ID preflight skips a drifted body; keep dry-run external_id exposure explicit in tests.
- R004 Step 2 plan notes: preserve trim/omit/no-clear semantics; test POST bulk-array and PUT single-object body shapes; make external-ID preflight behavior explicit; cover terse row omission/exposure; update add_or_update_event description away from “no idempotency key” wording.
- R006 Step 3 plan revision: protect matching external_id conflicts before replace_existing deletes same-day workouts; pin canonical hash tuple serialization/digest length; expose hashed proposed external_id in dry-run without raw plan/workout IDs.
- R011 Step 4 plan revision: docs must explicitly cover manual `external_id` usage in build-workouts, deterministic `icuvisor-plan-v1-...` IDs in season/apply workflows, conservative caveats/no-clear semantics, and `[Unreleased]` changelog wording. Schema snapshot expectation is that `add_or_update_event.json` changes while `apply_training_plan.json` remains unchanged because only output/proposed metadata changed.

*Reserved for execution notes*
| 2026-06-03 21:30 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 21:33 | Review R002 | plan Step 1: APPROVE |
| 2026-06-03 21:35 | Review R003 | code Step 1: APPROVE |
| 2026-06-03 21:37 | Review R004 | plan Step 2: APPROVE |
| 2026-06-03 21:45 | Review R005 | code Step 2: APPROVE |
| 2026-06-03 21:47 | Review R006 | plan Step 3: REVISE |
| 2026-06-03 21:48 | Review R007 | plan Step 3: APPROVE |
| 2026-06-03 22:54 | Review R011 | plan Step 4: REVISE |
