# Review R007 — Plan review for Step 2

Verdict: **REVISE**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- current Step 1 harness in `scripts/benchmark/kr5_benchmark.py`
- current prompt set and fixtures under `scripts/benchmark/`

The Step 2 direction is sound: add analyzer-specific prompt cases, keep prompt text single-copy across analyzer modes, and use a synthetic `icuvisor-analyzer-family` fixture with paired enabled/disabled call plans. However, the plan needs a little more precision before implementation because the current harness does not yet support two important pieces the plan relies on.

## Blocking issues

1. **The proposed “server scope” is not defined or supported by the current harness.**

   The plan says to add analyzer prompts to `kr5_shared_prompts.json` “with prompt-level `benchmark_modes` and a server scope so legacy reference fixtures are not forced to cover analyzer-only prompts.” Today `validate_measurement()` iterates every prompt in the prompt set for every measurement and requires call coverage for each prompt/mode. There is no prompt/server filtering in `mode_names_for_prompt()`, `validate_measurement()`, or `summarize()`.

   Without an explicit schema and code change, adding analyzer-only prompts to the shared prompt file will either:

   - fail all existing reference fixtures for missing analyzer prompt coverage, or
   - require unrelated legacy fixtures to grow synthetic analyzer rows, which defeats the stated scope isolation.

   Please revise the plan to define the scoping field and semantics, for example: `server_scope` / `measurement_scope` as a list of server IDs or tags, how unscoped prompts continue to apply to all current fixtures, and how `prompt_count`/coverage in the v2 result should treat prompts excluded from a given server.

2. **Expected source-tool usage is not yet a validated contract.**

   Step 2 requires recording expected source-tool usage. The current harness only normalizes and echoes per-call `source_tool_usage`; it does not compare it to prompt-level expectations or require it for analyzer prompt rows. The plan mentions “fixture call rows and expected tool metadata,” but only `expected_tools_by_mode` is currently validated, and that checks the top-level called tool, not source-tool usage.

   Please revise the plan to state where expected source-tool usage will live and how it will be checked. At minimum, the analyzer fixture should make it impossible to accidentally omit or drift the enabled analyzer `_meta.source_tools` / fallback source-tool plan without validation failing. This matters because TP-100’s output is supposed to inform TP-098 core-promotion decisions, not just list illustrative call rows.

## Non-blocking recommendations

- List the concrete analyzer prompt IDs and expected enabled/disabled top-level tools in the plan before implementation. A small table covering trend, distribution, baseline, correlation/compliance, and histogram would make review much easier and would reduce the chance of missing the roadmap’s trend/distribution/baseline promotion evidence.
- For the roadmap target of zero raw-stream pulls on trend/distribution/correlation analyzer shapes, make sure the disabled fallback plans include the relevant `get_activity_streams` rows where the benchmark is meant to demonstrate avoided fetch-and-reduce behavior, and that enabled rows do not call raw stream tools directly.
- Keep the existing Step 1 fairness invariant: `mode_catalogs.analyzers_enabled` and `mode_catalogs.analyzers_disabled` must share byte-identical non-analyzer catalog payloads, with only analyzer-family tools added in enabled mode.

