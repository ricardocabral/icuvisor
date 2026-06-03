# TP-146: Unit regression pack for work, calories, and hydration — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 0
**Review Counter:** 0
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** 🟨 In Progress

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room constraint confirmed

---

### Step 1: Audit current unit coverage
**Status:** ⬜ Not Started

- [ ] Locate existing tests for extended metrics Joules/kJ, wellness kcal/hydration, activity calories semantics, and unit metadata
- [ ] Identify missing regression assertions without duplicating existing coverage

---

### Step 2: Add unit regression tests
**Status:** ⬜ Not Started

- [ ] Add or tighten tests for raw Joules emitted only as explicit kJ-derived fields where applicable
- [ ] Add or tighten tests for wellness `kcalConsumed` and `hydrationVolume` unit semantics
- [ ] Assert zero values are preserved and ambiguous raw field names are not emitted in terse responses
- [ ] Run targeted tests: `go test ./internal/tools ./internal/response ./internal/analysis -run 'Joule|Calories|Hydration|Unit'`

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

| 2026-06-03 16:38 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:38 | Step 0 started | Preflight |