# Review R008 — Plan review for Step 2

Verdict: **REVISE**

Reviewed:

- `taskplane-tasks/TP-100-analyzer-benchmark-harness/PROMPT.md`
- `taskplane-tasks/TP-100-analyzer-benchmark-harness/STATUS.md`
- prior review `.reviews/R007-plan-step2.md`
- current harness in `scripts/benchmark/kr5_benchmark.py`
- current prompt set in `scripts/benchmark/prompts/kr5_shared_prompts.json`
- TP-098 prompt and roadmap analyzer/core-promotion requirements

The revised Step 2 plan addresses the two blockers from R007: it defines a `server_scope` concept so analyzer-only prompts do not force coverage in legacy fixtures, and it adds a prompt-level `expected_source_tools_by_mode` validation contract. The overall direction is now much closer to implementable.

One important gap remains before coding.

## Blocking issue

### The analyzer prompt cases do not yet cover the TP-098 `compute_zone_time` promotion candidate

TP-100 completion criteria require the benchmark results to explicitly inform the TP-098 core-promotion decision. TP-098's mission names exactly three benchmark-gated promotion candidates:

- `analyze_trend`
- `compute_zone_time`
- `compute_baseline`

The current Step 2 plan covers `analyze_trend` in `KR5-A01` and `compute_baseline` in `KR5-A03`, but the planned distribution case is `KR5-A02` using `analyze_distribution` rather than `compute_zone_time`. That leaves no benchmark evidence for one of the three tools TP-098 is supposed to decide on.

Please revise the plan so at least one analyzer prompt/call plan exercises `compute_zone_time` directly. Good options:

- make `KR5-A02` a zone/time-in-zone distribution prompt, with enabled mode calling `compute_zone_time` and disabled mode using the fetch-and-reduce source tools; or
- keep an `analyze_distribution` case, but add an additional scoped analyzer case for `compute_zone_time` so TP-098 receives evidence for all three promotion candidates.

If `compute_load_balance` is paired with `compute_zone_time` for the prompt shape, that is fine, but `compute_zone_time` itself should be an expected enabled top-level tool and should have explicit expected source-tool usage.

## Non-blocking recommendations

- Spell out the exact `expected_source_tools_by_mode` schema before implementation. The STATUS note says `{mode: {called_tool: [source tools...]}}` but also says validation checks names/counts. Prefer a count-capable shape matching normalized call rows, for example `{ "tool": "get_activity_intervals", "count": 1 }`, or explicitly state that duplicate strings encode counts.
- In the concrete prompt table, list enabled and disabled top-level tools for each prompt ID. This will make it obvious that disabled plans avoid analyzer-family tools and that enabled plans avoid LLM-visible raw-stream pulls for the roadmap target shapes.
- Keep `prompt_count` as the shared prompt-set size if desired, but consider emitting an applicable/scoped prompt count per server or per mode later so Step 3 documentation does not confuse skipped scoped prompts with missing coverage.

Once the plan includes `compute_zone_time` coverage, the rest of the Step 2 design should be safe to implement within the stated benchmark fixture scope.
