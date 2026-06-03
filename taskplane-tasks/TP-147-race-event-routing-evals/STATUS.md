# TP-147: Race-event routing evals for add_or_update_event — Status

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

### Step 1: Add race-event routing cases
**Status:** ⬜ Not Started

- [ ] Find the existing tool-routing fixture/test pattern
- [ ] Add prompts for creating A/B/C races and assert the expected first tool is `add_or_update_event`
- [ ] Include a negative assertion or fixture note that a separate `add_race_event` tool should not be required
- [ ] Run targeted tests: `go test ./internal/toolrouting ./internal/prompts -run 'Race|Routing|Fixture'`

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
