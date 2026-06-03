# TP-153: Snapshot every registered MCP tool schema — Status

**Current Step:** Step 4: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 2
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
**Status:** ✅ Complete

- [x] Live catalog compared to current whitelist
- [x] Mode coverage policy decided
- [x] Coach-mode injected schema policy decided
- [x] Intentional exclusions documented if any

---

### Step 2: Implement full coverage guard
**Status:** ✅ Complete

- [x] Whitelist replaced/extended to prevent silent gaps
- [x] Missing-snapshot tests added
- [x] No-network deterministic generation preserved
- [x] Targeted tests passing

---

### Step 3: Regenerate snapshots and review churn
**Status:** ✅ Complete

- [x] Snapshots regenerated
- [x] Added/changed snapshots reviewed for secrets/paths/ordering
- [x] Noise policy documented if needed
- [x] Targeted tests passing after refresh

---

### Step 4: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Lint passes
- [x] All failures fixed
- [x] Build passes

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
| 1 | plan | 1 | APPROVE | inline review_step |
| 2 | plan | 2 | APPROVE | inline review_step |
| 3 | plan | 3 | APPROVE | inline review_step |
| 4 | plan | 4 | APPROVE | inline review_step |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Current schema snapshot whitelist covers 37 of 60 full-mode coach-enabled registered tools; the 23 missing names are analyzer/planning additions, remaining custom/settings/delete helpers, gear helpers, validate_workout, and coach tools. | Step 1 policy input; Step 2 guard must derive coverage from the live full registry rather than a curated subset. | internal/toolchecks/schema_stability.go |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-03 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-03 21:28 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 21:28 | Step 0 started | Preflight |
| 2026-06-03 21:35 | Step 0 completed | Required files exist; Go 1.26.4 and Make targets available; full-mode coach registry has 60 tools; current schema snapshots/whitelist cover 37 tools. |
| 2026-06-03 21:35 | Step 1 started | Coverage policy decision |
| 2026-06-03 21:40 | Step 1 completed | Plan review APPROVE; policy is full-mode, coach-enabled, no public-tool exclusions. |
| 2026-06-03 21:40 | Step 2 started | Full coverage guard implementation |
| 2026-06-03 21:41 | Step 2 plan reviewed | APPROVE |
| 2026-06-03 21:45 | Step 3 started | Snapshot refresh and churn review |
| 2026-06-03 21:45 | Step 3 plan reviewed | APPROVE |
| 2026-06-03 21:50 | Step 4 started | Full verification |
| 2026-06-03 21:50 | Step 4 plan reviewed | APPROVE |
| 2026-06-03 21:38 | Worker iter 1 | done in 603s, tools: 65 |

---

## Blockers

*None*

---

## Notes

- Preflight counts: full-mode coach-enabled registry currently registers 60 tools; `schemaCatalogToolNames` and committed `internal/tools/schema_snapshot/*.json` currently contain 37 matching snapshots.
- Step 1 mode policy: generate and enforce snapshots from the full toolset with full delete/write capability so every public tool that can be registered is covered in a single canonical schema set; safe/core mode filtering is a registration policy and should not shrink schema drift coverage.
- Step 1 coach policy: enable coach mode during snapshot generation and include coach-only tools (`list_athletes`, `select_athlete`) plus the injected `athlete_id` argument in snapshots. This intentionally snapshots the broadest public schema; solo-mode schemas are subsets and remain protected because removing or changing a baseline property fails stability checks.
- Step 1 exclusions policy: no registered public MCP tools are intentionally excluded for TP-153. If a future generated schema must be excluded, Step 2 should require an explicit reason and test-enforce that the exclusion is not silent.
- Step 3 snapshot review: 60 JSON snapshots are parseable; grep found no API keys/secrets/local filesystem paths, only expected schema terms such as `next_page_token`, `credentials`, and `icuvisor://workout-syntax`; canonical generator ordering was verified by diffing two temp generations.
- Step 3 noise policy: no structural coverage was weakened; snapshot churn is accepted as the intentional broad full-mode coach baseline (coach `athlete_id` injection plus newly covered tools), while description/example text remains snapshotted to catch client-visible schema drift.
- Step 4 verification: `make test` and `make lint` completed with zero failures before build verification.
| 2026-06-03 21:32 | Review R001 | plan Step 1: APPROVE |
| 2026-06-03 21:35 | Review R002 | plan Step 2: APPROVE |
| 2026-06-03 21:45 | Review R003 | plan Step 3: APPROVE |
| 2026-06-03 21:50 | Review R004 | plan Step 4: APPROVE |
| 2026-06-03 21:40 | Review R003 | plan Step 3: APPROVE |
| 2026-06-03 21:42 | Review R004 | plan Step 4: APPROVE |
