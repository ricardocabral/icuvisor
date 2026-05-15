# TP-039-coach-mode: Coach mode + per-athlete tool ACLs — Status

**Current Step:** Step 2: Config + feature flag
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 4
**Review Counter:** 2
**Iteration:** 2
**Size:** L

---

### Step 1: Threat-model review + endpoint probe

**Status:** ✅ Complete

- [x] Threat model written (`athlete_id` cannot exfiltrate, escalate, or escape roster)
- [x] Coach-roster endpoint probed; path/auth/shape documented OR gap documented
- [x] Writeup in `docs/threat-models/coach-mode.md`
- [x] R001 revision: mark authenticated coach-key roster probe as blocked/incomplete unless a real coach-scoped key is provided, and phrase config roster as a temporary fallback pending validation

### Step 2: Config + feature flag

**Status:** ⏳ Not started

- [ ] `ICUVISOR_COACH_MODE=on|off|auto`
- [ ] `coach.athletes[]` schema with `allowed_tools` / `denied_tools` / `default_athlete_id`
- [ ] Unknown tool names fail loudly

### Step 3: Tool registry plumbing

**Status:** ⏳ Not started

- [ ] `coach.Evaluator` third gate
- [ ] Compose order: delete-mode → toolset-tier → coach-ACL (any deny is final)
- [ ] Uniform optional `athlete_id` arg with consistent description
- [ ] Per-request normalization + roster check

### Step 4: `list_athletes` + `select_athlete`

**Status:** ⏳ Not started

- [ ] `list_athletes` (`_meta.source: "config" | "upstream"`)
- [ ] `select_athlete` session/process-scoped state
- [ ] `requires_new_conversation` `_meta` flag

### Step 5: Catalog-cache caveat + Tests

**Status:** ⏳ Not started

- [ ] §7.4 #7 caveat documented
- [ ] Composition truth-table coverage
- [ ] End-to-end with faked intervals client

### Step 6: Documentation

**Status:** ⏳ Not started

- [ ] `docs/coach-mode.md`
- [ ] README pointer
- [ ] CHANGELOG
- [ ] Follow-up issue for PRD §7.4 #5 status update

---

## Decisions

- **Per-athlete delegated keys:** out of scope for v0.5. v0.5 ships single-coach-key + many-athletes-it-can-already-see.

## Blockers

- Step 1 authenticated roster probe is blocked: this execution environment has no `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_CONFIG`, or `ICUVISOR_ENV_FILE`, no default config at `~/Library/Application Support/icuvisor/config.json` or `~/.config/icuvisor/config.json`, and no accessible `icuvisor`/`intervals-icu-api-key` OS keychain credential. No real coach-scoped intervals.icu key was provided. Public OpenAPI and unauthenticated probes identified the likely endpoint, but R001 correctly rejected that as insufficient to complete the authenticated coach-key probe requirement.

## Notes

- Step 1 writeup lives at `docs/threat-models/coach-mode.md`.
- Threat model conclusion: `athlete_id` is only a normalized target selector; it cannot exfiltrate credentials, bypass per-athlete ACLs, or escape the local roster if request-time roster checks remain authoritative and compose with delete-mode/toolset gates.
- Endpoint probe conclusion: public OpenAPI documents `GET /api/v1/athlete/{id}/athlete-summary{ext}` as “Summary information for followed athletes” with `SummaryWithCats[]` fields including `athlete_id` and `athlete_name`, but no real coach key was available in the task environment, so TP-039 should implement `list_athletes` from config first (`_meta.source: "config"`) and leave upstream roster support for a later authenticated probe.
- Supervisor steering on 2026-05-15 explicitly treats the authenticated black-box coach-roster probe as an external/operator-deferred validation gate; Step 1 is complete on the documented-gap/fallback basis, not because upstream roster discovery was locally proven.

| 2026-05-15 20:00 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 20:00 | Step 1 started | Threat-model review + endpoint probe |
| 2026-05-15 20:07 | Review R001 | code Step 1: UNKNOWN |

| 2026-05-15 20:09 | Agent escalate | Blocked on TP-039 Step 1 after code review R001. Reviewer correctly rejected marking the coach-roster endpoint probe complete because the task requires an authenticated black-box probe with a real coa |
| 2026-05-15 20:09 | Worker iter 1 | done in 541s, tools: 51 |
| 2026-05-15 20:09 | Steering | Authenticated coach-roster probe is operator-deferred; proceed with config-backed roster and mark Step 1 complete on documented-gap/fallback basis. |
| 2026-05-15 20:10 | Review R002 | code Step 1: reviewer repeated R001 objection; superseded by supervisor steering to treat gap as complete for TP-039 v0.5. |
| 2026-05-15 20:12 | Review R002 | code Step 1: UNKNOWN |
