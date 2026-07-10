# TP-232: Correct athlete-ID normalization and hosted HTTP guidance — Status

**Current Step:** Step 3: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** S

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Hosted URL and athlete-ID behavior confirmed

---

### Step 1: Correct and lock the guidance

**Status:** ✅ Complete

**Expanded artifacts:** `web/content/guides/coach-mode.md`, `CONTRIBUTING.md`, `Makefile`, and `.github/workflows/ci.yml`, in addition to the prompt's original Step 1 artifacts. The content contract will reject the obsolete `normalize(s) ... to i12345` claim and require exact prefix-preserving guidance in `web/content/reference/config-file.md`, `web/content/guides/coach-mode.md`, and `CONTRIBUTING.md`; it will also require the hosted connector URL and generic-public-tunnel prohibition in `web/content/guides/http-transport.md`, running through CI.

- [x] Athlete-ID normalization wording corrected
- [x] Coach and contributor athlete-ID guidance corrected
- [x] Hosted HTTP troubleshooting wording corrected
- [x] Documentation content contract added
- [x] Documentation contract wired into CI
- [x] Targeted documentation test passing

---

### Step 2: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Documentation content test passing
- [x] Website build passes
- [x] Binary build passes
- [x] Markdown/diff clean

---

### Step 3: Documentation & Delivery

**Status:** ✅ Complete

- [x] Must Update docs modified
- [x] Check If Affected docs reviewed
- [x] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | REVISE | `.reviews/R003-plan-step1.md` |
| R004 | Plan | 1 | APPROVE | inline |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Coach-mode and contributor guidance also claimed bare IDs normalize to `i12345`. | Corrected and included in the contract test. | `web/content/guides/coach-mode.md`, `CONTRIBUTING.md` |
| The documentation contract was not part of the CI test path. | Added `make docs-guidance-test` and a CI invocation. | `Makefile`, `.github/workflows/ci.yml` |
| Task context file named in the prompt was not present. | Preflight proceeded from the authoritative source and hosted guidance files. | `taskplane-tasks/CONTEXT.md` |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 11:36 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 11:36 | Step 0 started | Preflight |
| 2026-07-10 11:40 | Step 1 plan review | R001 requested correction of coach-mode and contributor athlete-ID guidance |
| 2026-07-10 11:42 | Step 1 plan review | R002 required explicit expanded artifacts and CI execution for the content contract |
| 2026-07-10 11:43 | Step 1 plan review | R003 required path-specific athlete-ID contract coverage including CONTRIBUTING.md |
| 2026-07-10 11:44 | Step 1 plan review | R004 approved the revised plan |
| 2026-07-10 11:52 | Worker iter 1 | done in 949s, tools: 88 |
| 2026-07-10 11:52 | Task complete | .DONE created |

## Blockers

*None*

## Notes

*Reserved for execution notes*
| 2026-07-10 11:40 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 11:42 | Review R002 | plan Step 1: REVISE |
| 2026-07-10 11:44 | Review R003 | plan Step 1: REVISE |
| 2026-07-10 11:46 | Review R004 | plan Step 1: APPROVE |
