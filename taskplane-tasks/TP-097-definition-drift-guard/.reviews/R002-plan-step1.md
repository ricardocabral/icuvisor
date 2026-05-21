# Review R002 — Plan Review for Step 1

**Verdict:** APPROVED

The revised Step 1 plan in `STATUS.md` addresses the blocking gaps from R001. It defines a task-local inventory artifact in the `STATUS.md` Notes section, expands the trace from math functions into tool adapters and `_meta.formula_ref` propagation, calls out EF/VI as resource-only or upstream-mapped candidates rather than silently skipping them, and requires deterministic expected outputs/boundary cases before Step 2 starts.

## What looks good

- The inventory schema is specific enough for Step 2: formula/ref ID, canonical text/hash target, implementation functions, tool adapters, existing coverage, planned golden cases, and gaps/status.
- The search targets cover the main formula-sensitive paths:
  - `AnalysisFormulaRef` and `formula_ref` propagation.
  - `ComputeActivitySegmentStats` for HR drift / Pw:HR decoupling.
  - `ComputeZoneBalance` for polarization.
  - `ComputeBaselineStats` and `ComputeTrend` for z-score.
  - `get_extended_metrics` / `analysis/metrics.go` status for EF and VI.
- Boundary cases named in the plan match the task risk: zero/insufficient denominators, polarization moderate/high zero states, sample standard deviation (`n-1`), and explicit EF/VI handling.
- Reusing the existing `internal/resources/testdata/analysis_formulas.md` golden is the right default and avoids unnecessary docs churn.

## Conditions before marking Step 1 complete

- Make the golden fixture layout decision concrete in the Step 1 artifact. It is fine to decide during inventory, but the completed inventory should state the final path(s), not leave the current conditional “repo-root `testdata/analysis/` if cross-package, otherwise package-local” wording unresolved.
- Include tool-level metadata propagation in the inventory, not only package-level computations. In particular, record where `_meta.formula_ref` is emitted for `compute_activity_segment_stats`, `compute_zone_time` / `compute_load_balance`, `compute_baseline`, and `analyze_trend`.
- For EF and VI, explicitly record whether the current behavior is only canonical resource text plus upstream field mapping (`get_extended_metrics` / `metricCatalog`) and whether any local computation is intentionally absent.

No further plan changes are required before proceeding with Step 1 execution.
