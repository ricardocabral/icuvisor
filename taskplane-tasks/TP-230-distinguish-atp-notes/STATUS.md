# TP-230: Distinguish ATP-generated notes from personal calendar notes — Status

**Current Step:** Step 1: Design provenance-aware note shaping
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing provenance and keyword paths audited

---

### Step 1: Design provenance-aware note shaping

**Status:** 🟨 In Progress

- [ ] ATP and personal note statuses defined
- [ ] plan_applied provenance rule defined
- [ ] Counts cannot misclassify personal notes
- [ ] Ordering and terse/full behavior defined

---

### Step 2: Implement and cover classification

**Status:** ⬜ Not Started

- [ ] Provenance-aware note shape implemented
- [ ] English keyword dependence removed or constrained
- [ ] Personal and localized ATP fixtures covered
- [ ] Real TARGET week boundaries preserved
- [ ] Projection bridge remains explicit-target-only

---

### Step 3: Update schema and generated surfaces

**Status:** ⬜ Not Started

- [ ] Output schema snapshot updated
- [ ] Generated website data updated
- [ ] Tool description clarifies personal context

---

### Step 4: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] Build passes
- [ ] Generated docs clean

---

### Step 5: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `taskplane-tasks/CONTEXT.md` and the packet `.reviews/` directory were not staged; all implementation and documentation paths in task scope exist. | Continue from PROMPT and authoritative project docs; review tooling may create `.reviews/` when invoked. | Preflight |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 12:26 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 12:26 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

- Preflight audit: `intervals.Event.PlanApplied` already decodes `plan_applied`. The annual-plan path currently groups every `NOTE` as periodization context, increments week `note_count`, and derives `recovery_hint`/`recovery_note_count` solely via `annualTrainingPlanRecoveryHint`, which scans name/type/description/tags for English substrings `recovery`, `rest`, `taper`, and `deload`. No other annual-plan keyword recovery path was found.
