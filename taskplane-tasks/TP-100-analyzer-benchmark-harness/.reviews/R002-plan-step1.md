# Review R002 — Plan review for Step 1

Verdict: **REVISE before implementation**

The revised Step 1 checklist in `STATUS.md` addresses the themes from R001, especially fixture determinism, avoiding diagnostics log parsing, and explicitly treating this as a call-plan benchmark rather than autonomous model tool choice. That is progress, but the current plan is still mostly phrased as outcomes to define later. For this step, the risky parts are the definitions themselves, so they need to be captured before code is changed.

## Remaining blockers

1. **Analyzer modes are not actually defined yet.**
   The plan says to define fair enabled/disabled modes that hold non-analyzer tools constant, but it does not name the modes or state the mechanism. Before implementation, specify whether Step 1 will add fixture-only/call-plan modes, live catalog filtering, separate synthetic server IDs, or another approach. The plan should make clear that the catalog token comparison will not accidentally compare `core` vs `full` or otherwise change unrelated tools.

2. **Result schema v2 is too high-level.**
   `STATUS.md` says to upgrade to v2 with benchmark modes, per-call mode/source-tool usage, and stream-pull metrics, but does not list the concrete JSON fields or where they live. Please pin at least the intended top-level/per-server/per-call fields, e.g. how `benchmark_modes`, `mode`, `source_tool_usage`, and raw-stream counts will be represented, and whether existing v1 consumers/docs are intentionally broken or migrated.

3. **Raw-stream counting needs exact rules.**
   Counting from harness `ToolCall` rows and avoiding diagnostics logs is the right direction. The plan still needs the deterministic predicate: which tool names count as raw-stream pulls (`get_activity_streams` and any reference aliases), whether `unavailable:*` calls count, and how to handle analyzer-family tools whose `_meta.source_tools` mention streams. Keep LLM-visible raw stream pulls separate from internal analyzer source-tool evidence.

4. **Prompt/call-plan mapping is not concrete enough.**
   The plan notes that the harness is call-plan based, but it should say how the same prompt ID will have analyzer-enabled and analyzer-disabled call rows without duplicating or changing prompt text. This matters for validation: coverage should fail per prompt/mode/intent, not merely per prompt/intent as the current harness does.

5. **Fixture schema migration and validation are unspecified.**
   Step 1 will likely need fixture shape changes (`calls` with mode fields, mode-specific call-plan metadata, or split fixture files). The plan should identify the chosen migration path and the validation behavior for missing mode coverage before modifying fixtures.

## What is already acceptable

- The harness locations are correctly scoped to `scripts/benchmark/kr5_benchmark.py`, `scripts/benchmark/prompts/kr5_shared_prompts.json`, `scripts/benchmark/testdata/fixtures/`, and `docs/kr5-benchmark.md`.
- Fixture mode remains the authoritative deterministic path, with no live credentials required.
- Diagnostics/recent-tool-call log parsing is correctly avoided for fixture-mode stream counts.
- The plan no longer overclaims that this measures autonomous LLM tool-selection behavior.

## Expected revision

Add a short concrete design note to `STATUS.md` (or an adjacent execution note) before coding that includes:

- mode names and the exact analyzer-enabled/disabled comparison strategy;
- the v2 result schema fields to add;
- the raw-stream counting predicate and any reference-server aliases;
- the fixture/call-plan schema migration strategy;
- validation rules for prompt/intent coverage per mode.

After those details are pinned, Step 1 should be well-scoped and safe to implement.
