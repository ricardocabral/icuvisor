# Task: TP-239 - Add transparent masters plan review prompt

**Created:** 2026-07-10
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** This adds a read-only prompt and documentation by adapting the existing plan-health workflow. It does not alter physiology calculations, but wording must avoid unsupported age-based rules, medical claims, and opaque scores.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-239-masters-plan-review/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a read-only `masters_plan_review` MCP prompt and portable prompt pack for older endurance athletes who want a transparent audit of hard-session spacing, load ramp, recovery evidence, race proximity, and plan feasibility. Anchor conclusions in the athlete's own history and explicitly supplied preferences; do not infer a universal rule from age, invent an age-adjusted readiness score, or write calendar changes. The output should expose missing evidence and ask focused questions before recommending conservative plan edits.

## Dependencies

- **Task:** TP-235 (structured plan-filler constraint terminology and validation boundaries)

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — prompt catalog, analyzer evidence, and product boundaries
- `ROADMAP.md` — v2.2 science-backed plan safety scope
- `docs/design/plan-filler-constraints.md` — explicit constraint terminology from TP-235
- `internal/prompts/testdata/plan_health_review.md` — existing transparent plan-health workflow
- `internal/prompts/testdata/recovery_check.md` — personal-baseline and wellness freshness rules

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/registry.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/masters_plan_review_test.go`
- `internal/prompts/testdata/masters_plan_review.md`
- `docs/prompts/client-prompt-packs/masters-plan-review.md`
- `web/content/cookbook/masters-plan-review.md`
- `web/content/cookbook/_index.md`
- `web/content/reference/resources-prompts.md`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] TP-235 is complete
- [ ] Existing plan-health, recovery, projection, and explicit-constraint terminology reviewed

### Step 1: Define evidence and non-claim boundaries

**Plan-review checkpoint**

- [ ] Define a review sequence for personal baseline, recent hard-session spacing, load ramp, wellness freshness, race context, and explicit availability/duration preferences
- [ ] Separate observed evidence, user-supplied preferences, cautious interpretation, and proposed changes in the output
- [ ] Prohibit universal age cutoffs, automatic age-derived spacing/ramp values, medical conclusions, and black-box readiness or risk scores
- [ ] Require insufficient-evidence responses when history, zones, wellness, or plan detail cannot support the requested comparison
- [ ] Run targeted tests: `go test ./internal/prompts -run 'MastersPlanReview'`

**Artifacts:**

- `internal/prompts/catalog.go` (modified)
- `docs/prompts/client-prompt-packs/masters-plan-review.md` (new)

### Step 2: Register the prompt and add focused tests

- [ ] Add and register `masters_plan_review` with bounded planning/lookback/race context arguments and no credential, age-derived-policy, write, or delete arguments
- [ ] Reuse existing deterministic reads/analyzers and route unavailable advanced helpers through `icuvisor_list_advanced_capabilities`
- [ ] Add a deterministic golden fixture and a new focused test file covering personal-baseline language, stale wellness, explicit preferences, insufficient evidence, and read-only operation
- [ ] Ensure any calendar recommendation is presented as a reviewable proposal, never an automatic mutation
- [ ] Run targeted tests: `go test ./internal/prompts -run 'Prompt|MastersPlanReview'`

**Artifacts:**

- `internal/prompts/catalog.go` (modified)
- `internal/prompts/registry.go` (modified)
- `internal/prompts/masters_plan_review_test.go` (new)
- `internal/prompts/testdata/masters_plan_review.md` (new)

### Step 3: Publish the portable workflow and evals

- [ ] Add a cookbook page and portable client prompt pack that explain what the workflow can and cannot conclude
- [ ] Add eval scenarios for a well-instrumented athlete, stale/missing wellness, and a request for an unsupported universal age rule
- [ ] Update prompt reference, cookbook index, PRD prompt catalog, roadmap v2.2 wording, and Unreleased changelog
- [ ] Keep future science-backed rule-engine work explicitly separate from this prompt-only evidence review
- [ ] Run targeted checks: `go test ./internal/prompts && python3 scripts/eval/run_eval.py --validate`

**Artifacts:**

- `docs/prompts/client-prompt-packs/masters-plan-review.md` (new)
- `web/content/cookbook/masters-plan-review.md` (new)
- `web/content/cookbook/_index.md` (modified)
- `web/content/reference/resources-prompts.md` (modified)
- `scripts/eval/scenarios/cookbook_scenarios.json` (modified)
- `docs/prd/PRD-icuvisor.md` (modified)
- `ROADMAP.md` (modified)
- `CHANGELOG.md` (modified)

### Step 4: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run prompt eval validation: `make eval-validate`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify clean Markdown and diff: `git diff --check`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `web/content/cookbook/masters-plan-review.md` — document evidence, limits, and read-only workflow
- `web/content/reference/resources-prompts.md` — document the registered prompt
- `docs/prd/PRD-icuvisor.md` — add the prompt to the current catalog without claiming an age-aware physiology model
- `CHANGELOG.md` — record the new prompt and portable pack

**Check If Affected:**

- `ROADMAP.md` — keep future evidence-based rule engine separate from the shipped prompt
- `web/content/cookbook/season-and-block-plan.md` — link the review when appropriate
- `README.md` — update only if it lists every prompt

## Completion Criteria

- [ ] `masters_plan_review` is registered and read-only
- [ ] Conclusions use personal evidence and explicit preferences, not universal age rules
- [ ] Stale or missing data produces visible caveats or insufficient-evidence results
- [ ] No medical claims, opaque scores, or automatic calendar mutations are introduced
- [ ] New focused test, golden fixture, prompt pack, cookbook page, and eval scenarios exist
- [ ] Future science-backed validation remains a separate roadmap item
- [ ] Full tests, eval validation, lint, build, and diff checks pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-239:

- **Step completion:** `feat(TP-239): complete Step N — description`
- **Bug fixes:** `fix(TP-239): description`
- **Tests:** `test(TP-239): description`
- **Hydration:** `hydrate: TP-239 expand Step N checkboxes`

## Do NOT

- Start before TP-235 is complete
- Infer age, birth date, or a universal training rule from absent upstream fields
- Add automatic age-derived hard-day gaps, ramp limits, readiness scores, or injury-risk scores
- Make medical, diagnostic, or treatment claims
- Call write or delete tools
- Present a proposed calendar change as already applied
- Commit without TP-239 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
