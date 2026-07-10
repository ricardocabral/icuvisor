# TP-239: Add transparent masters plan review prompt — Status

**Current Step:** Step 2: Register the prompt and add focused tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 6
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

**Status:** ✅ Complete

<!-- R003 plan revision items -->
- [x] R003-1: In catalog and portable-pack instructions, order the athlete-local review as profile/timezone and resolved relative dates; non-overlapping baseline/history, completed, planned, and race windows; sourced event/plan/activity reads with pagination or truncation labelled partial; then permitted baseline, spacing, ramp, and wellness checks. Distinguish a confirmed calendar race from a user-supplied scenario anchor and retain current-day wellness as partial.
- [x] R003-2: Define mandatory, visibly separate sourced evidence, athlete-stated preference, cautious interpretation, insufficient-evidence/question, and conditional reviewable-proposal output sections; masters is an audience label only
- [x] R003-3: Permit hard-session spacing only for athlete-identified sessions or detailed, sourced activity/plan intensity evidence; use compute_baseline one eligible metric at a time with status, sample, missing-day, freshness, method, and formula metadata; use only copyable plan targets or athlete-supplied projection values, surface every projection assumption, and never treat projection defaults as policy
- [x] R003-4: Make the workflow absolutely read-only: it never calls write/delete tools, including after approval, and every change remains an unapplied conditional proposal
- [x] R003-5: Create `internal/prompts/masters_plan_review_test.go` in Step 1, table-driven and selected by `go test ./internal/prompts -run 'MastersPlanReview'`, to assert section/provenance separation, no age/medical/score claims, absolute no-write behavior, hard-session and projection fallbacks, and the insufficient-evidence matrix: ambiguous/unavailable hard-session or plan detail, absent/invalid zones, short/partial/truncated/missing history, missing/stale/partial wellness or missing/provider-native readiness, missing race context, and insufficient explicit projection targets; each gap names evidence, makes no affected conclusion, and asks one focused question while availability/requested duration remain athlete-stated context
- [x] Review evidence sequence defined

**Step 1 artifacts:** `internal/prompts/catalog.go`, `docs/prompts/client-prompt-packs/masters-plan-review.md`, and `internal/prompts/masters_plan_review_test.go`.
- [x] Evidence, preferences, interpretation, and proposals separated
- [x] Unsupported age and medical claims prohibited
- [x] Insufficient-evidence behavior defined

---

### Step 2: Register the prompt and add focused tests

**Status:** 🟨 In Progress

<!-- R006 plan revision items -->
- [ ] R006-1: Replace the static prompt handler with a validating handler: default to a 14-day athlete-local planned review (today through day 13), 28 completed-history days immediately before planned_start, and 56 personal-baseline days immediately before history; resolve the same non-overlapping sequence for supplied planned dates. Accept history 1-90 days and baseline 1-180 days; require strict paired YYYY-MM-DD planned dates in ascending order and no more than 90 days inclusive; trim and strictly validate race_date and require it for race_name; return short UserErrors and render normalized supplied values
- [ ] R006-2: Extend `masters_plan_review_test.go` with table-driven valid/default and invalid handler cases (non-integer, zero, out-of-range lookbacks; malformed, reversed, and overlong planned dates; malformed/name-only race; whitespace normalization), use `errors.As` to assert a `UserError` and its exact public message, and add the normalized valid scope to the golden fixture
- [ ] R006-3: Register `MastersPlanReviewPrompt()` in `NewRegistry()` and update the catalog and protocol prompt-list counts/order, rendered golden case, `TestPromptResourceCitationsStayTerse`, and portable-pack registry-link coverage. Test the exact six-argument allowlist (no credential, age-policy, write, or delete arguments), deterministic analyzer/advanced-capability fallback route, and the pack's bounded/default scope contract
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
| R003 | Plan | 1 | REVISE | `.reviews/R003-plan-step1.md` |
| R005 | Plan | 2 | REVISE | `.reviews/R005-plan-step2.md` |
| R006 | Plan | 2 | REVISE | `.reviews/R006-plan-step2.md` |

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
| 2026-07-10 21:51 | Review R003 | plan Step 1: REVISE |
| 2026-07-10 21:53 | Review R004 | plan Step 1: APPROVE |
| 2026-07-10 22:01 | Review R005 | plan Step 2: REVISE |
| 2026-07-10 22:04 | Review R006 | plan Step 2: REVISE |
