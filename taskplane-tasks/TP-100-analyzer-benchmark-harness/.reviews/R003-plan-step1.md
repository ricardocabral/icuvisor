# Review R003 — Plan review for Step 1

Verdict: **REVISE before implementation**

The latest `STATUS.md` notes incorporate most of R002's requested design: fixture/call-plan mode names, v2 schema direction, deterministic stream counting from harness rows, and explicit avoidance of diagnostics log parsing. That is a substantial improvement. I also confirmed the existing harness is still `scripts/benchmark/kr5_benchmark.py` with v1 fixture/result shapes under `scripts/benchmark/testdata/fixtures/` and `scripts/benchmark/results/kr5-results.json`.

There are still a couple of Step 1 design details that need to be pinned before coding, because they determine whether the benchmark will produce useful TP-098 evidence.

## Blocking gaps

1. **Mode catalog/tool-availability strategy is still ambiguous.**
   The note says the modes should live "under the same synthetic server catalog" so non-analyzer tools stay constant. That wording conflicts with the task requirement that with/without runs differ by analyzer tool availability. Please define this as one of:
   - a synthetic server *family* with two mode-specific catalogs: `analyzers_enabled` = same base catalog plus analyzer family, `analyzers_disabled` = same base catalog minus analyzer family; or
   - a single fixture `tools` list plus explicit per-mode `available_tools`/`catalog_payload` fields used for validation and token counting.

   In either case, validation should prove that all non-analyzer tools are identical across modes and that the disabled mode does not expose analyzer tools. Avoid a design where both modes share the exact same catalog and only the call plan changes, because that would not measure analyzer availability or catalog-token cost.

2. **Per-mode token metrics are not pinned in the v2 schema.**
   The plan mentions `benchmark_modes`, per-call `mode`, `source_tool_usage`, and per-mode raw-stream counts, but TP-100 also requires token deltas. The existing v1 harness only has server-level `description_tokens` and median response bytes; with mixed modes in one server, those fields become insufficient. Before implementation, specify concrete v2 summary fields such as a per-server/per-mode `mode_summaries` block containing at least:
   - catalog/tool-description token count for that mode, if catalogs differ;
   - response/result token metric or an explicit decision to continue using response bytes as the payload-size proxy;
   - median response bytes/tokens per mode;
   - `raw_stream_pull_count` per mode.

   If the v0.6 "≥40% fewer tokens" target is meant to use canonical tool-result tokens rather than bytes, Step 1 should add that metric now instead of leaving Step 3 to infer it from v1 response bytes.

3. **Mode-specific call-plan validation needs tool-family assertions, not only coverage.**
   Requiring rows for both modes per prompt/intent is necessary but not sufficient. Add validation rules that analyzer-disabled rows for analyzer benchmark prompts do not call `analyze_*`, `compute_*`, `get_activity_histogram`, or `get_fitness_projection`, and that analyzer-enabled rows for analyzer prompt shapes call the expected analyzer-family tool(s) unless the fixture explicitly marks an unavailable/error case. This keeps a fixture typo from silently turning both modes into the same fetch-and-reduce baseline.

## Non-blocking clarifications

- The raw-stream predicate is now mostly concrete. If the benchmark keeps Python reference fixtures in the same result, include their real aliases in the configured list, e.g. `icu_get_activity_streams` in addition to `get_activity_streams`; otherwise document that stream-pull counts are only reported for the icuvisor analyzer comparison.
- Define the shape of `source_tool_usage` enough to validate it (`[]string` vs objects with counts, required for analyzer-family calls, optional for fallback calls). Keep it separate from LLM-visible raw-stream pulls as planned.
- Move the concrete design out of the historical R002 note into a current Step 1 design note or expanded checklist entry so the implementation plan is not only embedded as prior-review noteary.

Once those points are captured, the implementation should be well scoped and can proceed against the existing harness without broad refactors.
