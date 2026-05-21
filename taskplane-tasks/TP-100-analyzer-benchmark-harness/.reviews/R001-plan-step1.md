# Review R001 — Plan review for Step 1

Verdict: **Needs revision before implementation**

The step objective is right, and the existing harness has been correctly bounded to `scripts/benchmark/kr5_benchmark.py`, `scripts/benchmark/prompts/kr5_shared_prompts.json`, fixture JSON under `scripts/benchmark/testdata/fixtures/`, and `docs/kr5-benchmark.md`. However, the current Step 1 plan in `STATUS.md` only restates the task checklist. It does not yet define enough design detail to safely implement the benchmark modes or produce the raw-stream metric that TP-100 and ROADMAP v0.6 require.

## Blocking gaps to address in the plan

1. **Define analyzer-enabled/disabled as fair catalog variants.**
   Do not equate “disabled” with `ICUVISOR_TOOLSET=core` and “enabled” with `full`: analyzers currently live in `full`, but core/full also changes many non-analyzer tools and catalog tokens. The plan should specify whether the comparison is:
   - `icuvisor-full-analyzers-enabled` vs `icuvisor-full-analyzers-disabled` with only the analyzer family hidden, or
   - fixed fixture/call-plan variants that hold all non-analyzer tools constant.

   Without this, the token delta will conflate analyzer cost/savings with unrelated full-tool catalog changes and will not be useful for TP-098 core-promotion evidence.

2. **Be explicit that this harness is call-plan based, not an autonomous LLM benchmark.**
   The existing KR5 harness executes predetermined MCP `tools/call` entries; it does not send prompts to a model and observe tool choice. The plan should say how identical prompts map to two different call plans: analyzer-enabled calls use `analyze_*`/`compute_*`/histogram tools, while analyzer-disabled calls use the existing fetch-and-reduce source tools. Otherwise later docs may overclaim that the benchmark measured model activation behavior.

3. **Define the stream-pull metric before coding.**
   “Raw-stream pull count” needs a deterministic rule. Recommended: count MCP tool calls whose tool name is `get_activity_streams` or known reference equivalents, grouped per server/mode/prompt, and emit `raw_stream_pull_count` in the result summary. If analyzer `_meta.source_tools` is also inspected, keep it as a separate metric such as `analyzer_source_stream_count`; do not mix internal analyzer source evidence with LLM-visible raw-stream pulls. This distinction matters because `compute_activity_segment_stats` is explicitly the analyzer-family raw-stream exception, while ROADMAP’s zero raw-stream-pull target applies to trend/distribution/correlation shapes.

4. **Avoid relying on diagnostics log parsing unless the plan explains isolation.**
   `internal/diagnostics/recent_tool_calls.go` only records redacted tool names in a user-cache JSONL file, with no prompt ID, server ID, mode, or arguments. It is not sufficient for deterministic fixture-mode counts and risks cross-run contamination. Prefer counting from the harness’s own `ToolCall` rows. If log parsing is still chosen, the plan must include temporary path isolation, cleanup, and prompt/mode attribution.

5. **Plan the result schema/version update.**
   The existing result schema is `kr5-benchmark-result-v1` and only reports catalog tokens and median response bytes. Step 1 should decide the new schema/version fields before code changes, e.g. `schema_version: kr5-benchmark-result-v2`, `benchmark_modes`, per-call `mode`, `source_tool_usage`, and per-server/mode `raw_stream_pull_count`. This is necessary so Step 3 can update `docs/kr5-benchmark.md` reproducibly instead of patching ad hoc fields into v1.

## Non-blocking recommendations

- Keep fixture mode authoritative and deterministic; do not require live intervals.icu credentials for the new analyzer comparison.
- Add mode validation in `validate_measurement`/fixture loading so missing prompt-intent coverage fails per mode, not just per server.
- Keep prompt text vendor-neutral and identical across modes; put analyzer vs fallback differences only in call-plan/config/fixture mappings.
- Include expected analyzer source-tool usage in fixture/config metadata, but do not commit raw athlete payloads or live transcripts.

## Expected revised Step 1 plan

Before implementation, update `STATUS.md` or an execution note with a concrete design along these lines:

- existing harness locations identified;
- comparison modes named and their catalog/tool filtering strategy defined;
- result schema changes listed;
- raw-stream count rule defined from harness calls, including reference tool-name aliases;
- confirmation that fixture-mode determinism/redaction remains unchanged.

Once those details are captured, the implementation should be low-risk and well scoped to the benchmark harness.
