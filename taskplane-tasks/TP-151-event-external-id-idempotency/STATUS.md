# TP-151: Event external_id idempotency — Status

**Current Step:** Step 1: Design external_id contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
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
- [x] Current event external_id handling identified

---

### Step 1: Design external_id contract
**Status:** 🟨 In Progress

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
| Current event write path has no typed `external_id`: `WriteEventParams`/`writeEventPayload` omit it; `add_or_update_event` request/schema omit it; event reads preserve raw payloads but terse rows do not expose `external_id`; `apply_training_plan` creates events without idempotency keys and relies on same-day duplicate matching. | Drives Step 1 contract and Step 2/3 implementation. | `internal/intervals/events.go`, `internal/tools/add_or_update_event.go`, `internal/tools/get_events.go`, `internal/tools/apply_training_plan.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |
| 2026-06-03 | Step 0 completed | Required files, dependencies, and current external_id handling identified |
| 2026-06-03 | Step 1 started | Design external_id contract |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
