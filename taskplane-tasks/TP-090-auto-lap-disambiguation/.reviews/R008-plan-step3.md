# Plan Review: Step 3 — Propagate to analyzers

## Verdict

Approve.

## Review scope

- Read `PROMPT.md` and current `STATUS.md` for TP-090.
- Re-read prior plan review `R007-plan-step3.md` to verify the requested refinements were incorporated.
- Checked the current analyzer support points in `internal/analysis/meta.go`, `internal/tools/analyzer_common.go`, and existing analyzer skeleton tests.
- Checked downstream TP-091/TP-093 prompts to confirm analyzer implementation remains out of TP-090 scope.

## Assessment

The revised Step 3 plan now addresses the earlier blocking gaps:

- It explicitly limits TP-090 Step 3 to shared `internal/analysis` support and placeholder tests, avoiding accidental implementation/registration of analyzer tools that belong to TP-091/TP-093.
- It defines a concrete analyzer metadata contract around `_meta.source_tools`, `_meta.interval_source`, and `_meta.auto_lap_suspected`, including omitting interval fields for non-interval analyzers via optional/pointer semantics.
- It defines the execution-claim policy helper and stable reason string `auto_lap_suspected`, which gives downstream analyzer tasks an actionable contract instead of relying on prose alone.
- It calls out preserving current non-interval analyzer metadata/goldens, which is important because `AnalyzerMeta` currently emits mandatory fields only.
- It includes an appropriate test plan for the current state of the codebase, where public analyzer tools are still absent.
- It requires handoff notes for TP-091/TP-093, matching the dependency chain in those task prompts.

## Non-blocking implementation guidance

When implementing, make the “known/evaluated” distinction explicit in code and tests: an evaluated classifier result of `IntervalSourceUnknown` should still be eligible to propagate as `_meta.interval_source: "unknown"` if interval evidence was actually consumed, while analyzers with no interval evidence should omit both interval-specific fields.

Otherwise, the plan is specific enough to proceed with Step 3.
