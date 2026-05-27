# TP-116: Privacy and GDPR positioning — Status

**Current Step:** Step 2: Add privacy-conscious user-facing copy
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 3
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
**Status:** ✅ Complete

- [x] Substantiated privacy/security claims inventoried
- [x] Explicit non-claims defined: no legal advice/certification, AI-client caveat, upstream data relationship caveat
- [x] Dedicated page versus existing-page update decision made

**Plan-review checkpoint**

---

### Step 2: Add privacy-conscious user-facing copy
**Status:** ✅ Complete

- [x] Local trust boundary, credential storage, HTTP bind defaults, and coach-mode credential handling documented
- [x] EU/GDPR-conscious language framed as design posture/questions, not certification
- [x] Homepage/README pointer added only if useful

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
| 2 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| 3 | plan | 2 | APPROVE | `.reviews/R003-plan-step2.md` |

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
| 2026-05-27 18:16 | Review R001 | plan Step 1: REVISE |
| 2026-05-27 18:22 | Review R002 | plan Step 1: APPROVE |
| 2026-05-27 18:24 | Review R003 | plan Step 2: APPROVE |

---

## Blockers

- 2026-05-27: RESOLVED BY SUPERVISOR STEERING. TP-116 depends on TP-113 landing first, but TP-113 was intentionally skipped in wave 1. Proceed with a constrained privacy-only pass that avoids homepage/local-first positioning owned by TP-113 and does not assume TP-113 copy exists.

---

## Notes

Plan review R001 requested Step 1 plan artifacts before implementation and flagged TP-113 as unsatisfied. Supervisor authorized a constrained pass: use standalone privacy-specific docs, avoid homepage/local-first positioning owned by TP-113, clarify privacy boundaries in existing privacy/coach/http docs only where safe, and link to SECURITY.md.

Step 1 plan notes:
- Substantiated claims: `SECURITY.md` says API keys are OS-keychain by default, plaintext file credentials are discouraged/warned, diagnostics redact secrets, HTTP binds to `127.0.0.1` by default, and icuvisor only contacts intervals.icu plus optional releases host. `web/content/explain/local-first.md` already says local binary, no icuvisor SaaS account/data host in normal local setup, keychain storage, and AI-client caveat. `web/content/explain/coach-mode.md` says coach API key stays in the server credential chain and `athlete_id` is only a selector gated by roster/ACL. `web/content/guides/http-transport.md` says HTTP loopback default and LAN bind exposes unauthenticated MCP. Code/tests substantiate `DefaultHTTPBindAddress = "127.0.0.1:8765"`, OS-keychain package, and coach ACL enforcement.
- Explicit non-claims: do not call icuvisor GDPR-compliant/certified, do not provide legal advice, and frame EU/privacy language as due-diligence questions. State that the chosen AI client/model provider may process conversation/tool-response content under its own terms, and intervals.icu remains the upstream service that stores/processes the athlete's training data under its own terms.
- Page decision: add a standalone `web/content/explain/privacy.md` privacy explanation and a card in `web/content/explain/_index.md`. This avoids editing homepage or local-first positioning owned by TP-113 while giving privacy-conscious users a stable target. Make only narrow cross-links/clarifications in coach-mode and HTTP transport docs; leave `SECURITY.md` authoritative and link to it instead of duplicating policy.
- Step 2 pointer decision: no homepage or README pointer added in the constrained pass because TP-113 owns the main local-first/homepage positioning and supervisor explicitly requested no homepage/no-TP-113-overlap work. Discoverability will come from the explain-section index and contextual doc links.
