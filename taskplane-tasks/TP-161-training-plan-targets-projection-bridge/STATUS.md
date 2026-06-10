# TP-161: Bridge training-plan targets into fitness projection — Status

**Current Step:** Step 1: Design deterministic weekly-target distribution
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Design deterministic weekly-target distribution
**Status:** 🟨 In Progress

- [ ] Weekly target input shape and no-implicit-fetch contract defined
- [ ] Week anchoring, partial-week horizon, and start-date exclusion defined
- [ ] Override formula and fallback-to-modeled behavior defined
- [ ] Failing bridge tests added
- [ ] Weekly target distribution assumptions defined
- [ ] Explicit daily-load precedence covered
- [ ] Validation/ignore behavior for duplicate/invalid/out-of-horizon weekly targets covered
- [ ] Tool-level metadata/schema-facing behavior covered
- [ ] Initial targeted projection/training-plan tests run

---

### Step 2: Implement bridge in projection input and schema
**Status:** ⬜ Not Started

> ⚠️ Hydrate: Expand after inspecting current `get_training_plan` output fields and upstream plan target shapes.

- [ ] Optional typed request field/helper added
- [ ] Deterministic conversion implemented
- [ ] Metadata/source assumptions updated
- [ ] Schema snapshot refreshed
- [ ] Targeted projection/training-plan tests pass

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

- [ ] `README.md` updated
- [ ] `CHANGELOG.md` updated
- [ ] PRD reviewed/updated if affected
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
| 2026-06-10 11:55 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 11:55 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

Public signal: IntervalCoach forum #858-859 noted goal-progress projection falling off after 7 days despite weekly TSS targets.
Plan review R001 requested concrete target shape, week anchoring/partial week semantics, exact override formula, fallback behavior, validation cases, and metadata/source_tools expectations before tests.
| 2026-06-10 11:58 | Review R001 | plan Step 1: REVISE |
