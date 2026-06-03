# TP-145: Read-only planning context tool — Status

**Current Step:** Step 1: Design read-only planning context contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room constraint confirmed

---

### Step 1: Design read-only planning context contract
**Status:** ⬜ Not Started

- [ ] Inventory existing tool/client patterns for get_today, get_training_plan, get_events, get_fitness, and prompt planning guidance
- [ ] Define a terse default response with `_meta.source_tools`, timezone/as-of, week window, and no write behavior
- [ ] Plan-review checkpoint completed before implementation

---

### Step 2: Implement get_planning_context
**Status:** ⬜ Not Started

- [ ] Add the tool using existing intervals client methods and response shaping patterns
- [ ] Return week events/workouts, active training-plan summary, current/recent fitness context, upcoming race context, and caveats without creating/updating/deleting calendar items
- [ ] Register the tool in the catalog/toolcatalog with appropriate core/full tier placement
- [ ] Add input and output schema descriptions that clearly distinguish planning context from ATP creation

---

### Step 3: Add tests and docs
**Status:** ⬜ Not Started

- [ ] Add table-driven handler tests for terse default, include_full behavior, source_tools metadata, timezone/week window handling, and empty-data caveats
- [ ] Add catalog/registration tests if needed
- [ ] Update CHANGELOG and README/catalog docs if user-visible
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

---

### Step 99: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing
- [ ] Build passes if code changed
- [ ] All failures fixed

---

### Step 100: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] Must-update docs modified
- [ ] Check-if-affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

<!-- Workers log durable discoveries here. -->

| 2026-06-03 16:11 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:11 | Step 0 started | Preflight |