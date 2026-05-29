# TP-128: Plan health review prompt — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Design plan-health prompt contract
**Status:** ⬜ Not Started

- [ ] Inspect existing `weekly_review`, `weekly_planning`, `race_week_taper`, analyzer tools, and cookbook pages.
- [ ] Decide whether to add a new `plan_health_review` prompt or extend `weekly_review` without duplicating TP-122 season-planning scope.
- [ ] Define required tool sequence: events/training plan, fitness/projection, planned-vs-completed compliance, recent wellness, and caveats for deload/recovery weeks.
- [ ] Run targeted tests: `go test ./internal/prompts`

---

### Step 2: Implement prompt and golden tests
**Status:** ⬜ Not Started

- [ ] Add or update prompt text with transparent scoring/caveats, explicit missing-data handling, and no hidden black-box score unless computed from surfaced values.
- [ ] Add/update prompt registry golden tests.
- [ ] Ensure prompt asks for a reviewed proposal before any calendar writes.
- [ ] Run targeted tests: `go test ./internal/prompts`

---

### Step 3: Document cookbook workflow
**Status:** ⬜ Not Started

- [ ] Add cookbook guidance showing when to use weekly review vs plan-health review vs season planning.
- [ ] Include caveats for deload weeks, planned races, and incomplete wellness/readiness data.
- [ ] Run targeted tests: `make test` or relevant docs validation if available.

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing: `make test`
- [ ] Lint passes or pre-existing linter limitations are documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or clearly documented as pre-existing

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
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
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 14:57 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:57 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
