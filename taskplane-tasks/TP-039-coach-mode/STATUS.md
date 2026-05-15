# TP-039-coach-mode: Coach mode + per-athlete tool ACLs — Status

**Current Step:** Step 1: Threat-model review + endpoint probe
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 4
**Review Counter:** 0
**Iteration:** 0
**Size:** L

---

### Step 1: Threat-model review + endpoint probe

**Status:** ⏳ Not started

- [ ] Threat model written (`athlete_id` cannot exfiltrate, escalate, or escape roster)
- [ ] Coach-roster endpoint probed; path/auth/shape documented OR gap documented
- [ ] Writeup in `docs/threat-models/coach-mode.md`

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

## Notes

_Add notes as work progresses._
