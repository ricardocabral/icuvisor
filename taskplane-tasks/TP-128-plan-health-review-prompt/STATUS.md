# TP-128: Plan health review prompt — Status

**Current Step:** Step 5: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 13
**Iteration:** 2
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Design plan-health prompt contract
**Status:** ✅ Complete

- [x] Inspect existing `weekly_review`, `weekly_planning`, `race_week_taper`, analyzer tools, and cookbook pages.
- [x] Decide whether to add a new `plan_health_review` prompt or extend `weekly_review` without duplicating TP-122 season-planning scope.
- [x] Define required tool sequence: events/training plan, fitness/projection, planned-vs-completed compliance, recent wellness, and caveats for deload/recovery weeks.
- [x] Record explicit prompt contract in Discoveries: name/approach, arguments, date windows, output sections, race-risk behavior, and test implications.
- [x] Include formula-transparency fallbacks and scope boundaries in Discoveries: formula resource, analyzer `_meta.method` assumptions, advanced capability fallback, no black-box score, no plan-filler/autonomous coaching/calendar writes.
- [x] Run targeted tests: `go test ./internal/prompts`

---

### Step 2: Implement prompt and golden tests
**Status:** ✅ Complete

- [x] Add or update prompt text with transparent scoring/caveats, explicit missing-data handling, and no hidden black-box score unless computed from surfaced values.
- [x] Add/update prompt registry golden tests.
- [x] Ensure prompt asks for a reviewed proposal before any calendar writes.
- [x] Run targeted tests: `go test ./internal/prompts`
- [x] Address R004/R005 review findings: verify registry/golden/contract invariant coverage, update MCP protocol prompt-list expectations to seven prompts, and run `go test ./internal/prompts ./internal/mcp`.
- [x] Address R007 review findings: restore terse/include_full guardrail in `plan_health_review`, cover it in tests/golden, and correct review history notes for R004/R005/R007.

---

### Step 3: Document cookbook workflow
**Status:** ✅ Complete

- [x] Confirm Step 2 carry-over is resolved with `go test ./internal/prompts ./internal/mcp` before docs work.
- [x] Add cookbook guidance showing when to use weekly review vs plan-health review vs season planning.
- [x] Include caveats for deload weeks, planned races, and incomplete wellness/readiness data.
- [x] Update MCP prompt/cookbook reference pages affected by adding `plan_health_review`, including `web/content/reference/resources-prompts.md` and `web/content/cookbook/_index.md`.
- [x] Run targeted tests: `go test ./internal/prompts ./internal/mcp` plus docs build if available (`make web-build` or equivalent).
- [x] Address R010 docs review findings: keep deload/compliance/wellness caveat out of the base weekly-review recipe unless fetched, and fix season-plan wording/link for the prompt-library copy.

---

### Step 4: Testing & Verification
**Status:** ✅ Complete

- [x] FULL test suite passing: `make test`
- [x] Lint passes or pre-existing linter limitations are documented: `make lint`
- [x] Build passes: `make build`
- [x] All failures fixed or clearly documented as pre-existing

---

### Step 5: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | APPROVE | inline |
| R003 | Code | 1 | APPROVE | inline |
| R004 | Plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | Code | 2 | REVISE | `.reviews/R005-code-step2.md` |
| R006 | Plan | 3 | REVISE | `.reviews/R006-plan-step3.md` |
| R007 | Code | 2 | REVISE | `.reviews/R007-code-step2.md` |
| R008 | Code | 2 | APPROVE | `.reviews/R008-code-step2.md` |
| R009 | Plan | 3 | APPROVE | `.reviews/R009-plan-step3.md` |
| R010 | Code | 3 | REVISE | `.reviews/R010-code-step3.md` |
| R011 | Code | 3 | APPROVE | `.reviews/R011-code-step3.md` |
| R012 | Plan | 4 | APPROVE | `.reviews/R012-plan-step4.md` |
| R013 | Code | 4 | APPROVE | `.reviews/R013-code-step4.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Add a new `plan_health_review` MCP prompt rather than extending `weekly_review`: weekly review remains retrospective/next-week preview, while plan-health review is a transparent current-plan risk audit over planned windows, compliance, load/form projection, wellness caveats, and race-date risk. This changes the prompt catalog from the PRD's current six prompts to seven, so Step 2 tests/docs must update the registry count/golden fixtures and Step 5 must review whether PRD §7.2.G needs a prompt-list update. | Design decision for Step 2 | `internal/prompts/catalog.go`; `internal/prompts/registry.go`; `docs/prd/PRD-icuvisor.md` |
| Plan-health tool order: read `get_athlete_profile`; read `get_events` and `get_training_plan` for planned/race context; read `get_activities` only as needed for completed context; run `compute_compliance_rate` for scheduled-vs-completed adherence; read `get_fitness`/`get_training_summary`; run `compute_load_balance` for intensity/block classification; run `get_fitness_projection` for the stated horizon/ramp/recovery assumptions; read `get_wellness_data` for recent sleep/readiness/HRV caveats; call `icuvisor_list_advanced_capabilities` and name missing full-tool analyzers when advanced tools are unavailable. | Prompt contract input for Step 2 | `internal/prompts/catalog.go` |
| `plan_health_review` contract: optional `planned_start`/`planned_end` athlete-local dates (default next 14 days), optional `completed_lookback_days` positive integer (default 14), optional `race_date` and `race_name` for risk anchoring. Output sections should be: data coverage/missing-data caveats; planned-vs-completed adherence; load/form trajectory; plan-health risk table with evidence and no opaque aggregate score; deload/recovery-week interpretation; race-date risk if a race date/event is available; reviewed proposal/questions before any writes. If no race event is found, say no confirmed race event was found and report any user-supplied race date as a scenario anchor rather than an observed race. | Prompt contract input for Step 2 | `internal/prompts/catalog.go`; `internal/prompts/testdata/plan_health_review.md` |
| Formula/scope guardrails: prompt should cite `icuvisor://analysis-formulas`, require analyzer `_meta.method`, `_meta.assumptions`, `_meta.formula_ref`, missing-days/sample-size caveats where present, and forbid a hidden black-box plan-health score. Risk labels may be low/medium/high only when backed by surfaced values. Deload/recovery weeks must be treated as intentional load reductions unless compliance/wellness/form evidence says otherwise. This is a review workflow only: no plan filler, no autonomous physiology model, and no calendar writes until the exact proposal has been shown and approved. | Prompt contract input for Step 2 and cookbook docs | `internal/prompts/catalog.go`; `web/content/cookbook/weekly-review.md`; `web/content/cookbook/season-and-block-plan.md` |
| Check-if-affected docs: PRD §7.2.G materially changed because the prompt catalog now has seven prompts, so the prompt list was updated. ROADMAP.md was reviewed and does not need a change because plan filler and science-guardrail future phase assumptions remain unchanged. | Completed in Step 5 | `docs/prd/PRD-icuvisor.md`; `ROADMAP.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 14:57 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:57 | Step 0 started | Preflight |
| 2026-05-29 15:14 | Worker iter 1 | done in 1018s, tools: 11 |
| 2026-05-29 15:14 | Step 1 started | Design plan-health prompt contract |
| 2026-05-29 15:55 | Worker iter 2 | done in 2487s, tools: 173 |
| 2026-05-29 15:55 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

R006 Step 3 plan review identified Step 2 carry-over issues; recovered by reverting premature Step 2 completion and adding the remaining Step 2 review item.
| 2026-05-29 15:17 | Review R001 | plan Step 1: REVISE |
| 2026-05-29 15:20 | Review R002 | plan Step 1: APPROVE |
| 2026-05-29 15:22 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 15:24 | Review R004 | plan Step 2: REVISE |
| 2026-05-29 15:29 | Review R005 | code Step 2: REVISE |
| 2026-05-29 15:31 | Review R006 | plan Step 3: REVISE |
| 2026-05-29 15:35 | Review R007 | code Step 2: REVISE |
| 2026-05-29 15:39 | Review R008 | code Step 2: APPROVE |
| 2026-05-29 15:45 | Review R009 | plan Step 3: APPROVE |
| 2026-05-29 15:48 | Review R010 | code Step 3: REVISE |
| 2026-05-29 15:52 | Review R011 | code Step 3: APPROVE |
| 2026-05-29 15:55 | Review R012 | plan Step 4: APPROVE |
| 2026-05-29 15:58 | Review R013 | code Step 4: APPROVE |
| 2026-05-29 15:38 | Review R008 | code Step 2: APPROVE |
| 2026-05-29 15:40 | Review R009 | plan Step 3: APPROVE |
| 2026-05-29 15:46 | Review R010 | code Step 3: REVISE |
| 2026-05-29 15:49 | Review R011 | code Step 3: APPROVE |
| 2026-05-29 15:50 | Review R012 | plan Step 4: APPROVE |
| 2026-05-29 15:53 | Review R013 | code Step 4: APPROVE |
