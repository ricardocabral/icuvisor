# R002 Plan Review — Step 1: Audit analyzer tier placement

**Verdict:** APPROVE

The revised Step 1 plan in `STATUS.md` addresses the gaps from R001. It now makes the audit reproducible by naming the analyzer-family sources of truth, requiring both constructor/effective-tier verification, and tying promotion eligibility to the existing TP-100/KR5 benchmark evidence.

## What is satisfactory

- The plan explicitly audits all analyzer-family tools rather than only the `analyzers` catalog group.
- It calls out the relevant sources of truth:
  - `analyzerFamilyCatalogNames()` in `internal/tools/catalog_test.go`
  - analyzer cases in `toolCatalogGroup()` in `internal/tools/catalog.go`
  - `get_fitness_projection` activation-test inclusion despite its `fitness` catalog group
- It requires checking both direct constructor wrapping and effective registered tiers.
- It includes the TP-100/KR5 evidence gate and limits promotion eligibility to `analyze_trend`, `compute_zone_time`, and `compute_baseline`.
- It keeps Step 1 scoped to audit/status handoff work; tier/test changes remain in later steps.

## Execution notes for Step 1

When performing the audit, make sure the table also captures analyzer-family tools whose catalog group is not `analyzers` but which are included by `analyzerFamilyCatalogNames()`; currently that includes `get_activity_histogram` and `get_fitness_projection`. This is not a blocker, but it will prevent the audit from accidentally treating catalog group as the only source of truth.

Recommended Step 1 output remains a concise `STATUS.md` table with: `tool`, `family source`, `catalog group`, `constructor tier`, `effective registered tier`, and `promotion eligibility/evidence`.
