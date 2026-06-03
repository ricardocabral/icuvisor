# TP-151: Event external_id idempotency — Status

**Current Step:** Not Started
**Status:** 🔵 Ready for Execution
**Last Updated:** 2026-06-03
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ⬜ Not Started

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current event external_id handling identified

---

### Step 1: Design external_id contract
**Status:** ⬜ Not Started

- [ ] Create/update/omit/clear semantics decided
- [ ] apply_training_plan deterministic ID strategy decided
- [ ] Event read-row exposure decided
- [ ] Upstream uncertainty recorded

---

### Step 2: Implement event write/read support
**Status:** ⬜ Not Started

- [ ] WriteEventParams and payload support external_id
- [ ] add_or_update_event schema/decoder/handler supports external_id
- [ ] Create/update and preflight tests added
- [ ] Event row exposure implemented/tested as decided
- [ ] Targeted tests passing

---

### Step 3: Make apply_training_plan retry-safer
**Status:** ⬜ Not Started

- [ ] Stable plan event external IDs generated
- [ ] Repeated apply payload stability tests added
- [ ] Dry-run metadata reviewed for safety/usefulness
- [ ] Targeted tests passing

---

### Step 4: Refresh schemas, routing, and docs
**Status:** ⬜ Not Started

- [ ] Schema snapshots regenerated
- [ ] Tool-routing expectations updated if affected
- [ ] User docs updated if affected
- [ ] CHANGELOG updated

---

### Step 5: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passes
- [ ] All failures fixed
- [ ] Build passes

---

### Step 6: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Remaining caveats summarized

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
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
