# R006 Plan Review — Step 3: Apply promotion if evidence exists

**Verdict:** APPROVE

The revised Step 3 plan addresses the blocker from R005. `docs/kr5-benchmark.md` contains positive TP-098 promotion evidence for exactly `analyze_trend`, `compute_zone_time`, and `compute_baseline`, and the plan now explicitly selects that promotion path while keeping all other analyzer-family tools in `full`.

## What looks good

- The plan names the only three constructors that should change from `fullTool(...)` to `coreTool(...)`:
  - `internal/tools/analyze_trend.go`
  - `internal/tools/compute_zone_time.go`
  - `internal/tools/compute_baseline.go`
- It calls out the necessary tier-test updates so the Step 2 default-full assertions do not conflict with the post-promotion policy.
- It preserves the guardrail that non-candidate analyzer-family tools remain `full`.
- It includes the advanced-capabilities expectation: promoted core analyzers should disappear from the hidden/full-only capability list, while remaining full-only analyzers stay advertised.
- It includes regeneration/review of generated docs data via `make docs-tools`.

## Minor execution notes

- Since evidence is already present and positive, treat the absent/negative-evidence checkbox as not applicable rather than reopening the decision during implementation.
- Ensure the generated artifact `web/data/tools.json` is included if `make docs-tools` changes it; `web/content/reference/tools.md` likely only needs review/status documentation unless the generator changes it.
- Do not defer the user-visible behavior note indefinitely: `CHANGELOG.md` is required by the task, even though it is listed under Step 4 in `STATUS.md`.

No further plan changes are required before implementing Step 3.
