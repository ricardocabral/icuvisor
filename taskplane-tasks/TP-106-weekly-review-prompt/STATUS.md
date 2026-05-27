# TP-106: Weekly review MCP prompt — Status

**Current Step:** Step 1: Add `weekly_review` prompt registration
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Existing prompt style and golden-test format reviewed

---

### Step 1: Add `weekly_review` prompt registration
**Status:** ✅ Complete

- [x] Prompt registered with metadata, arguments, tool list, instructions, and return format
- [x] Athlete-local date/timezone guidance included
- [x] No-write-without-approval guidance included
- [x] Targeted prompt tests passing
- [x] Plan review revisions addressed: registry wiring, registration tests, deliberate analyzer/compute tool set, and explicit write/delete guardrail
- [x] Step 1 scope covers `catalog.go`, `registry.go`, and non-golden `catalog_test.go` updates; targeted command is `go test ./internal/prompts`
- [x] Weekly review tool sequence planned: profile/timezone, wellness with `_meta.stale`/provenance cautions, fitness + training summary, activities, events/training plan for planned-vs-completed/next-week preview, `compute_zone_time`, `compute_compliance_rate`, load-balance-equivalent analyzer guidance if available, optional `analyze_trend`, and `icuvisor_list_advanced_capabilities` fallback guidance

---

### Step 2: Add golden tests
**Status:** ⬜ Not Started

- [ ] Default golden prompt file added
- [ ] Explicit-arguments test added or documented as unnecessary under current test style
- [ ] Advanced-capability fallback guidance included where appropriate
- [ ] Targeted prompt tests passing

---

### Step 3: Changelog and full verification
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated
- [ ] Prompt docs/reference checked if applicable
- [ ] Targeted prompt tests passing

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | Step 1 | REVISE | .reviews/R002-plan-step1.md |
| R003 | Plan | Step 1 | REVISE | .reviews/R003-plan-step1.md |
| R004 | Plan | Step 1 | APPROVE | inline review_step verdict |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-26 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 14:15 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 14:15 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Tracking issue: https://github.com/ricardocabral/icuvisor/issues/33
| 2026-05-27 14:18 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-27 14:20 | Review R002 | plan Step 1: UNKNOWN |
| 2026-05-27 14:22 | Review R003 | plan Step 1: REVISE |
| 2026-05-27 14:24 | Review R004 | plan Step 1: APPROVE |
