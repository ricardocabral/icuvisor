# TP-158: Sport-settings profile readiness warnings — Status

**Current Step:** Step 1: Design and add readiness warning shape
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Design and add readiness warning shape
**Status:** 🟨 In Progress

- [x] `_meta.warnings` warning codes added
- [x] Warnings are terse, sport-scoped, and non-sensitive
- [x] Warnings provide actionable planning preflight context
- [x] Targeted profile/sport tests pass
- [ ] Prefer sport setting `types` over legacy `type` when deriving warning scope
- [ ] Restrict heart-rate readiness warnings to applicable endurance sport types

---

### Step 2: Propagate to tool/resource schemas and tests
**Status:** ⬜ Not Started

- [ ] get_athlete_profile warnings covered by tests
- [ ] athlete-profile resource covered if shared shaping applies
- [ ] Schema snapshot refreshed if needed
- [ ] update_sport_settings guidance/tests reviewed
- [ ] Targeted tool/resource tests pass

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Integration tests (if applicable)
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `README.md` updated if affected
- [ ] `CHANGELOG.md` updated
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-10 11:40 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 11:40 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

Public signals: IcuSync forum #263 and LeCoach forum #406 highlight threshold/zone readiness problems.
| 2026-06-10 11:43 | Review R001 | plan Step 1: APPROVE |
| 2026-06-10 11:48 | Review R002 | code Step 1: UNKNOWN |
