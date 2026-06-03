# TP-153: Snapshot every registered MCP tool schema — Status

**Current Step:** Step 1: Decide snapshot coverage policy
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Current registered-tool and snapshot counts recorded

---

### Step 1: Decide snapshot coverage policy
**Status:** 🟨 In Progress

- [ ] Live catalog compared to current whitelist
- [ ] Mode coverage policy decided
- [ ] Coach-mode injected schema policy decided
- [ ] Intentional exclusions documented if any

---

### Step 2: Implement full coverage guard
**Status:** ⬜ Not Started

- [ ] Whitelist replaced/extended to prevent silent gaps
- [ ] Missing-snapshot tests added
- [ ] No-network deterministic generation preserved
- [ ] Targeted tests passing

---

### Step 3: Regenerate snapshots and review churn
**Status:** ⬜ Not Started

- [ ] Snapshots regenerated
- [ ] Added/changed snapshots reviewed for secrets/paths/ordering
- [ ] Noise policy documented if needed
- [ ] Targeted tests passing after refresh

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passes
- [ ] All failures fixed
- [ ] Build passes

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final snapshot policy summarized

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
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |
| 2026-06-03 21:35 | Step 0 completed | Required files exist; Go 1.26.4 and Make targets available; full-mode coach registry has 60 tools; current schema snapshots/whitelist cover 37 tools. |
| 2026-06-03 21:35 | Step 1 started | Coverage policy decision |

---

## Blockers

*None*

---

## Notes

- Preflight counts: full-mode coach-enabled registry currently registers 60 tools; `schemaCatalogToolNames` and committed `internal/tools/schema_snapshot/*.json` currently contain 37 matching snapshots.
