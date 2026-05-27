# TP-116: Privacy and GDPR positioning — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 2
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied (TP-113 skipped; supervisor authorized constrained no-homepage/no-TP-113-overlap pass)
- [x] Existing privacy, security, coach-mode, and HTTP transport claims reviewed

---

### Step 1: Define privacy claims and boundaries
**Status:** ⬜ Not Started

- [ ] Substantiated privacy/security claims inventoried
- [ ] Explicit non-claims defined: no legal advice/certification, AI-client caveat, upstream data relationship caveat
- [ ] Dedicated page versus existing-page update decision made

**Plan-review checkpoint**

---

### Step 2: Add privacy-conscious user-facing copy
**Status:** ⬜ Not Started

- [ ] Local trust boundary, credential storage, HTTP bind defaults, and coach-mode credential handling documented
- [ ] EU/GDPR-conscious language framed as design posture/questions, not certification
- [ ] Homepage/README pointer added only if useful

---

### Step 3: Link from relevant docs and update changelog
**Status:** ⬜ Not Started

- [ ] Privacy positioning linked from local-first, coach-mode, HTTP transport, and/or safety docs where relevant
- [ ] `SECURITY.md` remains authoritative for security policy
- [ ] `CHANGELOG.md` updated under `[Unreleased]` if appropriate

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Docs/site build passing: `make web-build`
- [ ] Rendered links and page placement checked
- [ ] FULL test suite run if non-doc/generated app files are touched: `make test`
- [ ] Build passes if app strings or generated assets are touched: `make build`
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
| 1 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| TP-113 dependency is not actually complete in this worktree: `taskplane-tasks/TP-113-local-first-positioning-refresh/STATUS.md` is Ready/Not Started despite a merge commit subject mentioning TP-113. | Supervisor authorized a constrained TP-116 pass because TP-113 was intentionally skipped in wave 1; do not edit homepage/local-first positioning owned by TP-113 or assume TP-113 copy exists. | Step 0 dependency check; `.reviews/R001-plan-step1.md`; steering message iteration 2 |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 18:13 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 18:13 | Step 0 started | Preflight |
| 2026-05-27 18:17 | Agent escalate | TP-116 is blocked on its declared dependency TP-113. I initially marked the dependency satisfied based on a merge commit subject mentioning TP-113, but the reviewer pointed out and I confirmed `taskpl |
| 2026-05-27 18:17 | Worker iter 1 | done in 281s, tools: 32 |

---

## Blockers

- 2026-05-27: RESOLVED BY SUPERVISOR STEERING. TP-116 depends on TP-113 landing first, but TP-113 was intentionally skipped in wave 1. Proceed with a constrained privacy-only pass that avoids homepage/local-first positioning owned by TP-113 and does not assume TP-113 copy exists.

---

## Notes

Plan review R001 requested Step 1 plan artifacts before implementation and flagged TP-113 as unsatisfied. Supervisor authorized a constrained pass: use standalone privacy-specific docs, avoid homepage/local-first positioning owned by TP-113, clarify privacy boundaries in existing privacy/coach/http docs only where safe, and link to SECURITY.md.
| 2026-05-27 18:16 | Review R001 | plan Step 1: UNKNOWN |
