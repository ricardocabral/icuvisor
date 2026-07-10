# TP-239: Add transparent masters plan review prompt — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 10
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

**Status:** ✅ Complete

<!-- R007 plan revision items -->
- [x] R007-1: Replace the static prompt handler with a validating handler: default to a 14-day athlete-local planned review (today through day 13), 28 completed-history days immediately before planned_start, and 56 personal-baseline days immediately before history; resolve the same non-overlapping sequence for supplied planned dates. Accept history 1-90 days and baseline 1-180 days; require strict paired YYYY-MM-DD planned dates in ascending order and no more than 90 days inclusive (same-day is valid); trim and strictly validate race_date and require it for race_name; return short UserErrors and render normalized supplied values
- [x] R007-2: Extend `masters_plan_review_test.go` with table-driven valid/default and invalid handler cases (non-integer, zero, out-of-range lookbacks; malformed, start-only, end-only, reversed, and overlong planned dates; valid same-day and 90-day boundary; malformed/name-only race; whitespace normalization), use `errors.As` to assert a `UserError` and its exact public message, and add the normalized valid scope to the golden fixture
- [x] R007-3: Update the portable pack to mirror the handler's normalized scope/defaults, lookback ranges, paired-date/max-window validation, and race-name/date dependency. Register `MastersPlanReviewPrompt()` in `NewRegistry()` and update the catalog and protocol prompt-list counts/order, rendered golden case, `TestPromptResourceCitationsStayTerse`, and portable-pack registry-link coverage. Test the exact six-argument allowlist (no credential, age-policy, write, or delete arguments), deterministic analyzer/advanced-capability fallback route, and the pack's bounded/default scope terms
- [x] Prompt implemented and registered
- [x] Existing analyzers and fallback routing used
- [x] Focused test and golden fixture added
- [x] Calendar recommendations remain proposals

---

### Step 3: Publish the portable workflow and evals

**Status:** ✅ Complete

<!-- R009 plan revision items -->
- [x] R009-1: Deliberately revise the existing canonical pack and add a front-matter-valid `masters-plan-review` cookbook page plus index card. Present both MCP-prompt and pasted-pack entry points, bounded/default athlete-local windows, the five ordered sections, source/window/coverage/freshness evidence, stated preferences, race/current-day caveats, hard-session/baseline/projection limits, absolute read-only behavior, and affected-dimension insufficient-evidence questions.
- [x] R009-2: Add self-contained `CB-MASTERS-*` well-instrumented, stale/missing-wellness, and universal-age-rule eval records for `recipe: "masters-plan-review"`. Encode the athlete-local/source route, expected registered reads, all registered write/delete and raw/heavy routes as forbidden, required output/limitations, and anti-patterns for chat-side calculations, weak hard-session classification, missing-data completeness, age policy, opaque scores, and writes.
- [x] R009-3: Update prompt reference and guardrail prose; every registry-backed public prompt count/list in cookbook index and PRD; roadmap v2.2; Unreleased changelog; `docs/prompts/README.md`; and the root README's pack list. Record the `season-and-block-plan.md` cross-link and other Check If Affected dispositions.
- [x] R009-4: Validate prompt tests, eval schema, and Hugo build with `go test ./internal/prompts`, `python3 scripts/eval/run_eval.py --validate`, and `make web-build`.
- [x] Cookbook page and prompt pack added
- [x] Positive and refusal eval scenarios added
- [x] References, PRD, roadmap, and changelog updated
- [x] Future rule engine remains separate

---

### Step 4: Testing & Verification

**Status:** ✅ Complete

- [x] FULL test suite passing
- [x] Prompt eval validation passing
- [x] Lint passing
- [x] All failures fixed
- [x] Build passes
- [x] Markdown and diff clean

---

### Step 5: Documentation & Delivery

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
| R005 | Plan | 2 | REVISE | `.reviews/R005-plan-step2.md` |
| R006 | Plan | 2 | REVISE | `.reviews/R006-plan-step2.md` |
| R007 | Plan | 2 | REVISE | `.reviews/R007-plan-step2.md` |
| R009 | Plan | 3 | REVISE | `.reviews/R009-plan-step3.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| `season-and-block-plan` is a related existing-plan audit entry point. | Added a focused link to the strictly read-only masters evidence review; prompt-pack indexes and README were also updated. | `web/content/cookbook/season-and-block-plan.md`, `docs/prompts/README.md`, `README.md` |
| `taskplane-tasks/CONTEXT.md` was not provisioned in this lane. | Used the specified prompt, PRD, roadmap, plan-health/recovery fixtures, and TP-235 design/status as the available task context; no replacement context file was created. | Step 0 preflight |

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
| 2026-07-10 22:07 | Review R007 | plan Step 2: REVISE |
| 2026-07-10 22:10 | Review R008 | plan Step 2: APPROVE |
| 2026-07-10 22:21 | Review R009 | plan Step 3: REVISE |
| 2026-07-10 22:23 | Review R010 | plan Step 3: APPROVE |
