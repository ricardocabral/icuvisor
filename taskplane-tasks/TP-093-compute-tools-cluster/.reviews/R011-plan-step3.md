# R011 plan review — Step 3: Register and document activation hints

Verdict: REVISE

The Step 3 checklist has been expanded in the right direction since R008: it now calls out registry/catalog/test surfaces, PRD-style activation hints, generated tool docs, CHANGELOG, and keeping `analyze_trend` unregistered. However, the plan still must not proceed because the task status is treating unresolved prior `REVISE` reviews as approvals and Step 3 would publicly register tools with known contract/implementation blockers.

## Blocking findings

### 1. Step 3 still advances past unresolved Step 1/Step 2 `REVISE` reviews

- Location: `STATUS.md:24-60`, `STATUS.md:108-119`, `.reviews/R009-code-step1.md:1-4`, `.reviews/R010-code-step2.md:1-5`
- Severity: High

`STATUS.md` marks Step 1 and Step 2 complete and records R009/R010 as `APPROVE`, but the review files themselves both say `Verdict: REVISE`. R009 still blocks on deterministic contract gaps for summary-window baseline sampling and two-window activity truncation semantics. R010 still blocks on public-behavior implementation bugs, including weekly baseline z-scores, activity-derived baseline fields, truncation status/boundaries, and compliance breakdown denominators.

Registration/documentation is the public exposure step. The plan needs an explicit prerequisite to resolve or supersede R009 and R010 with approving reviews before any registry/catalog/docs changes publish the four compute tools.

### 2. Status metadata still hides the blocker state

- Location: `STATUS.md:46-60`, `STATUS.md:118-120`, `STATUS.md:140-160`
- Severity: High

The task metadata says Step 2 is complete, `Blockers` is `None`, and the execution log lists several revised reviews as approved. That is unsafe for a step-boundary plan because a worker following the status file could register the tools without seeing the actual outstanding blockers.

Before Step 3 proceeds, update `STATUS.md` so it reflects the real review state: R009/R010 are `REVISE` unless replaced by new approving reviews, Step 1/Step 2 should not be considered complete for registration purposes, and the blockers/notes should name the outstanding contract and implementation issues.

### 3. Schema-stability/snapshot updates should be explicit, not implied

- Location: `STATUS.md:70-72`, `internal/toolchecks/schema_stability.go`, `internal/tools/schema_snapshot/`
- Severity: Medium

The expanded checklist mentions “canonical tool-name surfaces,” which may cover `internal/toolcatalog/catalog.go`, but it still does not explicitly mention the schema-stability allowlist and committed schema snapshots. If the compute tools become registered public MCP tools, Step 3 should decide whether they enter `schemaCatalogToolNames`; if they do, the plan should include regenerating/committing the corresponding `internal/tools/schema_snapshot/*.json` files. If they intentionally stay out of schema stability for this step, the plan should state why.

## Required plan changes

1. Add a Step 3 precondition: fix the blockers in R009 and R010 and obtain superseding approving reviews before registration.
2. Correct `STATUS.md` so the review table, execution log, blockers, and completed-step markers match the actual review files.
3. Make the schema-stability decision explicit: update `internal/toolchecks/schema_stability.go` and generated schema snapshots for the four tools, or document why this public registration does not enter that surface yet.

I did not run the test suite because this is a plan review.
