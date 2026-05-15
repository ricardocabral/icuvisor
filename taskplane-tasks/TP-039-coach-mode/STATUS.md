# TP-039-coach-mode: Coach mode + per-athlete tool ACLs — Status

**Current Step:** Step 1: Threat-model review + endpoint probe
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 4
**Review Counter:** 1
**Iteration:** 1
**Size:** L

---

### Step 1: Threat-model review + endpoint probe

**Status:** 🟨 In Progress

- [x] Threat model written (`athlete_id` cannot exfiltrate, escalate, or escape roster)
- [ ] Coach-roster endpoint probed; path/auth/shape documented OR gap documented
- [x] Writeup in `docs/threat-models/coach-mode.md`
- [ ] R001 revision: mark authenticated coach-key roster probe as blocked/incomplete unless a real coach-scoped key is provided, and phrase config roster as a temporary fallback pending validation

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

- Step 1 authenticated roster probe is blocked: this execution environment has no `INTERVALS_ICU_API_KEY`, `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_CONFIG`, or `ICUVISOR_ENV_FILE`, and no real coach-scoped intervals.icu key was provided. Public OpenAPI and unauthenticated probes identified the likely endpoint, but R001 correctly rejected that as insufficient to complete the authenticated coach-key probe requirement.

## Notes

- Step 1 writeup lives at `docs/threat-models/coach-mode.md`.
- Threat model conclusion: `athlete_id` is only a normalized target selector; it cannot exfiltrate credentials, bypass per-athlete ACLs, or escape the local roster if request-time roster checks remain authoritative and compose with delete-mode/toolset gates.
- Endpoint probe conclusion: public OpenAPI documents `GET /api/v1/athlete/{id}/athlete-summary{ext}` as “Summary information for followed athletes” with `SummaryWithCats[]` fields including `athlete_id` and `athlete_name`, but no real coach key was available in the task environment, so TP-039 should implement `list_athletes` from config first (`_meta.source: "config"`) and leave upstream roster support for a later authenticated probe.

| 2026-05-15 20:00 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 20:00 | Step 1 started | Threat-model review + endpoint probe |
| 2026-05-15 20:07 | Review R001 | code Step 1: UNKNOWN |
