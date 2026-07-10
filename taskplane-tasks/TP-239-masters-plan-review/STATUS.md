# TP-239: Add transparent masters plan review prompt — Status

**Current Step:** Step 1: Define evidence and non-claim boundaries
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 2
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] TP-235 complete
- [x] Existing evidence and constraint terminology reviewed

---

### Step 1: Define evidence and non-claim boundaries

**Status:** 🟨 In Progress

<!-- R002 plan revision items -->
- [ ] R002-1: Define mandatory, visibly separate sourced evidence, athlete-stated preference, cautious interpretation, insufficient-evidence/question, and conditional reviewable-proposal output sections; masters is an audience label only
- [ ] R002-2: Permit hard-session spacing only for athlete-identified sessions or detailed, sourced activity/plan intensity evidence; use compute_baseline one eligible metric at a time with status, sample, missing-day, freshness, method, and formula metadata; use only copyable plan targets or athlete-supplied projection values, surface every projection assumption, and never treat projection defaults as policy
- [ ] R002-3: Make the workflow absolutely read-only: it never calls write/delete tools, including after approval, and every change remains an unapplied conditional proposal
- [ ] R002-4: Add a Step 1 boundary-test artifact and matrix for ambiguous/unavailable hard-session or plan detail, absent/invalid zones, short/partial/truncated/missing history, missing/stale/partial wellness or missing/provider-native readiness, missing race context, and insufficient explicit projection targets; every gap names evidence, makes no affected conclusion, and asks one focused question while availability/requested duration remain athlete-stated context
- [ ] Review evidence sequence defined
- [ ] Evidence, preferences, interpretation, and proposals separated
- [ ] Unsupported age and medical claims prohibited
- [ ] Insufficient-evidence behavior defined

---

### Step 2: Register the prompt and add focused tests

**Status:** ⬜ Not Started

- [ ] Prompt implemented and registered
- [ ] Existing analyzers and fallback routing used
- [ ] Focused test and golden fixture added
- [ ] Calendar recommendations remain proposals

---

### Step 3: Publish the portable workflow and evals

**Status:** ⬜ Not Started

- [ ] Cookbook page and prompt pack added
- [ ] Positive and refusal eval scenarios added
- [ ] References, PRD, roadmap, and changelog updated
- [ ] Future rule engine remains separate

---

### Step 4: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Prompt eval validation passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes
- [ ] Markdown and diff clean

---

### Step 5: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 21:41 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 21:41 | Step 0 started | Preflight |
| 2026-07-10 21:42 | Exit intercept reprompt | Supervisor provided instructions (217 chars) — reprompting worker |

## Blockers

*None*

## Notes

*Reserved for execution notes*
| 2026-07-10 21:46 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 21:48 | Review R002 | plan Step 1: REVISE |
