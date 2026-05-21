#!/usr/bin/env python3
"""Cookbook activation + answer-quality eval for icuvisor.

Two modes:

  --validate   Stdlib-only. Checks the scenario file is well-formed, every tool
               name resolves against web/data/tools.json, every recipe page
               exists, and the judge prompt is present. Safe for CI; no API
               calls, no network.

  (live)       Default. Spawns the icuvisor MCP server over stdio, runs each
               scenario through an Anthropic agent loop, captures the tool
               calls and final answer, scores the run with the LLM judge, and
               writes results JSON plus a scoreboard.

Live mode needs `anthropic` and `mcp` installed and ANTHROPIC_API_KEY set, and
an icuvisor binary configured against a (test) intervals.icu account.

Examples:
  python3 scripts/eval/run_eval.py --validate
  python3 scripts/eval/run_eval.py --server-cmd ./bin/icuvisor
  python3 scripts/eval/run_eval.py --filter CB-WEEKLY --server-cmd ./bin/icuvisor
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
DEFAULT_SCENARIOS = REPO_ROOT / "scripts/eval/scenarios/cookbook_scenarios.json"
DEFAULT_JUDGE = REPO_ROOT / "scripts/eval/judge/judge_prompt.md"
DEFAULT_CATALOG = REPO_ROOT / "web/data/tools.json"
RECIPE_DIR = REPO_ROOT / "web/content/cookbook"
DEFAULT_RESULTS = REPO_ROOT / "scripts/eval/results/latest.json"

REQUIRED_FIELDS = ("id", "recipe", "persona", "prompt", "expected_tools",
                   "forbidden_tools", "must_address")
ASSISTANT_SYSTEM = (
    "You are a helpful assistant with access to the icuvisor MCP tools, which "
    "return real intervals.icu training data. Answer the user's request using "
    "those tools. Do not answer training-data questions from memory."
)


# --------------------------------------------------------------------------
# validate mode
# --------------------------------------------------------------------------
def load_catalog(path: Path) -> set[str]:
    data = json.loads(path.read_text())
    return {t["name"] for t in data}


def validate(scenarios_path: Path, catalog_path: Path, judge_path: Path) -> int:
    errors: list[str] = []
    warnings: list[str] = []

    if not catalog_path.exists():
        print(f"FAIL: tool catalog not found at {catalog_path}")
        print("      run `make docs-tools` first")
        return 1
    catalog = load_catalog(catalog_path)

    if not judge_path.exists():
        errors.append(f"judge prompt not found at {judge_path}")

    try:
        spec = json.loads(scenarios_path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        print(f"FAIL: cannot parse {scenarios_path}: {exc}")
        return 1

    scenarios = spec.get("scenarios", [])
    if not scenarios:
        errors.append("scenario file has no scenarios")

    seen_ids: set[str] = set()
    for i, sc in enumerate(scenarios):
        tag = sc.get("id", f"#{i}")
        for field in REQUIRED_FIELDS:
            if field not in sc:
                errors.append(f"{tag}: missing required field '{field}'")
        sid = sc.get("id")
        if sid in seen_ids:
            errors.append(f"{tag}: duplicate scenario id")
        seen_ids.add(sid)

        if not sc.get("prompt", "").strip():
            errors.append(f"{tag}: empty prompt")
        if sc.get("persona") not in ("athlete", "coach"):
            errors.append(f"{tag}: persona must be 'athlete' or 'coach'")

        expected = set(sc.get("expected_tools", []))
        bonus = set(sc.get("bonus_tools", []))
        forbidden = set(sc.get("forbidden_tools", []))
        for name in expected | bonus | forbidden:
            if name not in catalog:
                errors.append(f"{tag}: unknown tool '{name}' (not in catalog)")
        overlap = expected & forbidden
        if overlap:
            errors.append(f"{tag}: tools both expected and forbidden: {sorted(overlap)}")
        if not expected:
            warnings.append(f"{tag}: no expected_tools — activation cannot be scored")

        recipe = sc.get("recipe", "")
        if recipe and not (RECIPE_DIR / f"{recipe}.md").exists():
            errors.append(f"{tag}: recipe page '{recipe}.md' not found in {RECIPE_DIR}")

    print(f"scenarios:   {len(scenarios)}")
    print(f"tool catalog: {len(catalog)} tools")
    for w in warnings:
        print(f"  warn: {w}")
    if errors:
        print(f"\nFAIL: {len(errors)} error(s)")
        for e in errors:
            print(f"  - {e}")
        return 1
    print("\nOK: scenario file is valid and in sync with the tool catalog.")
    return 0


# --------------------------------------------------------------------------
# live mode
# --------------------------------------------------------------------------
def _tool_result_text(result) -> str:
    parts = []
    for block in getattr(result, "content", []) or []:
        text = getattr(block, "text", None)
        if text:
            parts.append(text)
    return "\n".join(parts) if parts else "(no textual content)"


def _final_text(content) -> str:
    return "\n".join(b.text for b in content if getattr(b, "type", "") == "text")


def _parse_judge_json(raw: str) -> dict:
    start, end = raw.find("{"), raw.rfind("}")
    if start == -1 or end == -1:
        raise ValueError(f"no JSON object in judge reply: {raw[:200]}")
    return json.loads(raw[start:end + 1])


async def run_live(args, spec: dict, judge_text: str) -> int:
    import anthropic
    from mcp import ClientSession, StdioServerParameters
    from mcp.client.stdio import stdio_client

    client = anthropic.Anthropic()
    server = StdioServerParameters(command=args.server_cmd, args=args.server_args)
    scenarios = [s for s in spec["scenarios"]
                 if not args.filter or s["id"].startswith(args.filter)]

    results = []
    async with stdio_client(server) as (read, write):
        async with ClientSession(read, write) as session:
            await session.initialize()
            listed = await session.list_tools()
            available = {t.name for t in listed.tools}
            anth_tools = [{"name": t.name,
                           "description": t.description or "",
                           "input_schema": t.inputSchema}
                          for t in listed.tools]
            print(f"server exposes {len(available)} tools\n")

            for sc in scenarios:
                missing = set(sc["expected_tools"]) - available
                if missing:
                    print(f"SKIP {sc['id']}: server lacks {sorted(missing)} "
                          f"(wrong toolset/mode?)")
                    results.append({"scenario_id": sc["id"], "verdict": "skipped",
                                    "reason": f"missing tools {sorted(missing)}"})
                    continue

                print(f"RUN  {sc['id']} ({sc['recipe']}) ...", flush=True)
                transcript = await _agent_loop(client, session, anth_tools, sc,
                                               args.model, args.max_steps)
                verdict = _judge(client, judge_text, sc, transcript,
                                 args.judge_model)
                results.append(verdict)
                mark = {"pass": "PASS", "fail": "FAIL"}.get(
                    verdict.get("verdict"), "????")
                print(f"     {mark}  overall={verdict.get('overall')}")

    _write_and_report(spec, results, args.output)
    failed = [r for r in results if r.get("verdict") == "fail"]
    return 1 if failed else 0


async def _agent_loop(client, session, tools, scenario, model, max_steps):
    messages = [{"role": "user", "content": scenario["prompt"]}]
    tool_calls = []
    for _ in range(max_steps):
        resp = client.messages.create(
            model=model, max_tokens=4096, system=ASSISTANT_SYSTEM,
            tools=tools, messages=messages)
        messages.append({"role": "assistant", "content": resp.content})
        if resp.stop_reason != "tool_use":
            return {"tool_calls": tool_calls, "final_answer": _final_text(resp.content)}
        tool_results = []
        for block in resp.content:
            if getattr(block, "type", "") != "tool_use":
                continue
            args = block.input or {}
            tool_calls.append({"name": block.name, "arguments": args})
            try:
                result = await session.call_tool(block.name, args)
                text = _tool_result_text(result)[:2000]
            except Exception as exc:  # noqa: BLE001 - report tool errors to model
                text = f"ERROR calling {block.name}: {exc}"
            tool_results.append({"type": "tool_result",
                                 "tool_use_id": block.id, "content": text})
        messages.append({"role": "user", "content": tool_results})
    return {"tool_calls": tool_calls,
            "final_answer": "(stopped: max steps reached)"}


def _judge(client, judge_text, scenario, transcript, judge_model) -> dict:
    payload = {
        "scenario": {k: scenario.get(k) for k in
                     ("id", "prompt", "expected_tools", "bonus_tools",
                      "forbidden_tools", "must_address", "anti_patterns")},
        "transcript": {
            "tool_calls": [t["name"] for t in transcript["tool_calls"]],
            "tool_calls_detail": transcript["tool_calls"],
            "final_answer": transcript["final_answer"],
        },
    }
    resp = client.messages.create(
        model=judge_model, max_tokens=2048, system=judge_text,
        messages=[{"role": "user",
                   "content": "Evaluate this run.\n\n" + json.dumps(payload, indent=2)}])
    try:
        verdict = _parse_judge_json(_final_text(resp.content))
    except (ValueError, json.JSONDecodeError) as exc:
        return {"scenario_id": scenario["id"], "verdict": "fail",
                "overall": 0.0, "error": f"unparseable judge reply: {exc}"}
    verdict.setdefault("scenario_id", scenario["id"])
    return verdict


def _write_and_report(spec, results, output: Path) -> None:
    scored = [r for r in results if r.get("verdict") in ("pass", "fail")]
    dims = spec.get("scoring", {}).get("dimensions", [])
    means = {}
    for d in dims:
        vals = [r["scores"][d] for r in scored
                if isinstance(r.get("scores"), dict) and d in r["scores"]]
        means[d] = round(sum(vals) / len(vals), 2) if vals else None
    passed = sum(1 for r in scored if r["verdict"] == "pass")
    summary = {
        "scenario_set": spec.get("version"),
        "scored": len(scored),
        "passed": passed,
        "pass_rate": round(passed / len(scored), 3) if scored else None,
        "dimension_means": means,
    }
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(json.dumps({"summary": summary, "results": results}, indent=2))

    print("\n" + "=" * 60)
    print(f"pass {passed}/{len(scored)} scored"
          + (f"  ({len(results) - len(scored)} skipped)"
             if len(results) != len(scored) else ""))
    for d, m in means.items():
        print(f"  {d:18s} {m}")
    print(f"results written to {output}")


# --------------------------------------------------------------------------
def main() -> int:
    p = argparse.ArgumentParser(description=__doc__,
                                formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--validate", action="store_true",
                   help="check scenario file only; no API calls (CI-safe)")
    p.add_argument("--scenarios", type=Path, default=DEFAULT_SCENARIOS)
    p.add_argument("--judge", type=Path, default=DEFAULT_JUDGE)
    p.add_argument("--catalog", type=Path, default=DEFAULT_CATALOG)
    p.add_argument("--output", type=Path, default=DEFAULT_RESULTS)
    p.add_argument("--filter", default="", help="run only scenarios whose id starts with this")
    p.add_argument("--server-cmd", default="icuvisor",
                   help="icuvisor binary to spawn as the MCP server")
    p.add_argument("--server-args", nargs="*", default=[],
                   help="extra args passed to the icuvisor server")
    p.add_argument("--model", default="claude-sonnet-4-6",
                   help="model under test (the assistant)")
    p.add_argument("--judge-model", default="claude-opus-4-7",
                   help="model used as the LLM judge")
    p.add_argument("--max-steps", type=int, default=12,
                   help="max agent tool-use turns per scenario")
    args = p.parse_args()

    if args.validate:
        return validate(args.scenarios, args.catalog, args.judge)

    spec = json.loads(args.scenarios.read_text())
    judge_text = args.judge.read_text()
    try:
        import asyncio
        return asyncio.run(run_live(args, spec, judge_text))
    except ImportError as exc:
        print(f"live mode needs extra packages: {exc}")
        print("install with: pip install anthropic mcp")
        return 2


if __name__ == "__main__":
    sys.exit(main())
