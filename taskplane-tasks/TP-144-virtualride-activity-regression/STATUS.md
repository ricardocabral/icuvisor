# TP-144: VirtualRide activity regression coverage — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-03
**Review Level:** 0
**Review Counter:** 0
**Iteration:** 0
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Clean-room constraint confirmed

---

### Step 1: Add VirtualRide activity regression coverage
**Status:** ⬜ Not Started

- [ ] Add or update a get_activities fixture/test containing an upstream activity with `type: "VirtualRide"`
- [ ] Assert terse shaping preserves `sport`/type as VirtualRide and does not collapse it to Ride
- [ ] Assert the row remains present under the current default filters
- [ ] Run targeted tests: `go test ./internal/tools -run 'TestGetActivities|VirtualRide'`

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
