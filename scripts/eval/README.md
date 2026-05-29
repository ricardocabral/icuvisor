# Cookbook eval harness

End-to-end tests for the [cookbook recipes](../../web/content/cookbook/) and
prompt library. Where the [KR5 benchmark](../benchmark/) measures the *token
cost* of icuvisor, this harness measures **answer quality**: does a recipe
prompt make an assistant activate the right tools, ground every claim in real
data, and refuse to guess?

## What it checks

Each scenario is a self-contained prompt (a filled-in recipe). A run captures
the MCP tool calls the assistant made and its final answer, then an LLM judge
scores six dimensions 1-5:

| Dimension | Question |
| --- | --- |
| `tool_activation` | Did it call the expected tools and avoid forbidden ones? |
| `grounding` | Does every number trace to a tool result? Is missing data named? |
| `scale_correctness` | Sleep 1-4, feel 1-5, RPE 1-10, Strava-import limits honored? |
| `coverage` | Were all `must_address` points answered? |
| `actionability` | Is the answer decision-shaped and specific? |
| `conciseness` | Is the depth appropriate, no raw-payload dumps? |

A scenario **passes** only if the weighted overall is >= 4.0, `tool_activation`
and `grounding` are each >= 3, and no dimension scores 1. See
[`judge/judge_prompt.md`](judge/judge_prompt.md) for the full rubric.

## Layout

```
scripts/eval/
  run_eval.py                      runner (validate + live modes)
  scenarios/cookbook_scenarios.json scenario specs
  judge/judge_prompt.md            LLM judge system prompt + rubric
  results/                         run output (git-ignored except .gitkeep)
```

## Validate mode (CI-safe, no API)

```bash
python3 scripts/eval/run_eval.py --validate
# or
make eval-validate
```

Validation confirms the scenario file is well-formed, every tool name in
`expected_tools` / `bonus_tools` / `forbidden_tools` exists in
`web/data/tools.json`, every `recipe` points at a real cookbook page, and the
judge prompt is present. Run it after editing scenarios or the tool catalog so
the eval never references a renamed tool.

## Live mode (real LLM + real account)

```bash
export ANTHROPIC_API_KEY=sk-...
python3 scripts/eval/run_eval.py --server-cmd ./bin/icuvisor
```

Live mode needs `pip install anthropic mcp`, an `ANTHROPIC_API_KEY`, and an
icuvisor binary configured against a **test** intervals.icu account (never a
real athlete's). It spawns icuvisor as an MCP server over stdio, runs each
scenario through an agent loop, judges the result, and writes
`results/latest.json`.

Useful flags: `--filter CB-WEEKLY` (subset by id prefix), `--model` /
`--judge-model`, `--max-steps`. Coach scenarios are skipped automatically when
the server is not in coach mode; full-toolset scenarios are skipped on a
core-only server.

### Handling run-to-run noise

The agent and the judge are both stochastic, so a single run scores ±1 point.
The runner manages this:

- `--repeats N` runs each scenario N times; the per-scenario verdict uses the
  **mean overall**, so one noisy run cannot flip it. Output reports mean / min /
  max per scenario.
- `--scenario-delay` (default 5s) and `--tool-delay` (default 0.5s) throttle
  requests to keep intervals.icu from rate-limiting a long run.
- `--retries` (default 3) retries transient tool errors (rate limit, gateway,
  timeout) with exponential backoff.

When comparing a recipe change before/after, use `--repeats 3` or higher on the
affected scenarios and compare the means.

## The iteration loop

The judge does not just score — it returns `improvement_suggestions` aimed at
the **recipe prompt**. Use them to harden the cookbook until activation is
reliable:

1. Run the eval (`run_eval.py`).
2. For any scenario below threshold, read `improvement_suggestions`.
3. Edit the recipe in `web/content/cookbook/` and the matching scenario prompt.
4. Re-run. Repeat until the pass rate and dimension means stop improving.

A recipe is "really good" when it passes across runs with `tool_activation` and
`grounding` consistently >= 4. Commit `results/` snapshots only when comparing
runs intentionally.

## Adding a scenario

Add an object to `scenarios/cookbook_scenarios.json` with: `id` (unique,
`CB-` prefix), `recipe` (matching a cookbook page slug), `persona`
(`athlete` or `coach`), a self-contained `prompt` (no real IDs or dates — use
relative windows and descriptive references), `expected_tools`, optional
`bonus_tools`, `forbidden_tools`, `must_address`, and `anti_patterns`. For
multi-step activity questions, put the lookup tool and downstream detail tools in
the intended order in `expected_tools`, then make the ordering and grounding
requirements explicit in `must_address` and `anti_patterns` because validate mode
checks schema/catalog integrity while the judge scores transcript order and
unsupported claims. Then run `--validate`.
