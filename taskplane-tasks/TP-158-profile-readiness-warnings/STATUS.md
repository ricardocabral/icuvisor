# TP-158: Sport-settings profile readiness warnings — Status

**Current Step:** Step 2: Propagate to tool/resource schemas and tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 5
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Design and add readiness warning shape
**Status:** ✅ Complete

- [x] `_meta.warnings` warning codes added
- [x] Warnings are terse, sport-scoped, and non-sensitive
- [x] Warnings provide actionable planning preflight context
- [x] Targeted profile/sport tests pass
- [x] Prefer sport setting `types` over legacy `type` when deriving warning scope
- [x] Restrict heart-rate readiness warnings to applicable endurance sport types

---

### Step 2: Propagate to tool/resource schemas and tests
**Status:** 🟨 In Progress

- [x] get_athlete_profile warnings covered by tests
- [x] athlete-profile resource covered if shared shaping applies
- [x] Schema snapshot refreshed if needed
- [x] update_sport_settings guidance/tests reviewed
- [x] Targeted tool/resource tests pass
- [x] Align readiness warning actions with `update_sport_settings` argument fields
- [x] Add handler-level get_athlete_profile warning serialization tests including alias-complete settings

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
| `go run ./scripts/snapshot_tool_schemas.go` produced no `get_athlete_profile.json` diff because committed snapshots cover input schemas; readiness warnings/output description do not change arguments. | No schema snapshot change needed. | `internal/tools/schema_snapshot/get_athlete_profile.json` |

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
| 2026-06-10 11:53 | Review R003 | code Step 1: APPROVE |
| 2026-06-10 11:56 | Review R004 | plan Step 2: APPROVE |
| 2026-06-10 12:05 | Review R005 | code Step 2: UNKNOWN |
