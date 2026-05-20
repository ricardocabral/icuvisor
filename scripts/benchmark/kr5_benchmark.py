#!/usr/bin/env python3
"""Measure KR5 call-plan catalog, response, and stream-pull metrics for MCP servers."""

from __future__ import annotations

import argparse
import json
import os
import select
import statistics
import subprocess
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any

TOKENIZER_NAME = "cl100k_base"
TOKENIZER_PACKAGE = "tiktoken==0.12.0"
DEFAULT_BENCHMARK_MODE = "default"
ANALYZER_ENABLED_MODE = "analyzers_enabled"
ANALYZER_DISABLED_MODE = "analyzers_disabled"
RAW_STREAM_TOOL_NAMES = frozenset(
    {
        "get_activity_streams",
        "icu_get_activity_streams",
        "reference:get_activity_streams",
    }
)
ANALYZER_TOOL_PREFIXES = ("analyze_", "compute_")
ANALYZER_DIRECT_TOOLS = frozenset({"get_activity_histogram", "get_fitness_projection"})


class BenchmarkError(Exception):
    pass


def canonical_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":"))


def canonical_bytes(value: Any) -> bytes:
    return canonical_json(value).encode("utf-8")


class TokenCounter:
    def __init__(self, allow_approx: bool = False) -> None:
        try:
            import tiktoken  # type: ignore
        except ImportError as exc:
            if not allow_approx:
                raise BenchmarkError(
                    f"{TOKENIZER_PACKAGE} is required for KR5 results; install it with "
                    f"python3 -m pip install {TOKENIZER_PACKAGE!r}, or pass "
                    "--allow-approx-tokenizer for smoke tests only"
                ) from exc
            self.name = "approx_byte4_smoke_only"
            self.package = "stdlib"
            self._encoding = None
            return
        self.name = TOKENIZER_NAME
        self.package = TOKENIZER_PACKAGE
        self._encoding = tiktoken.get_encoding(TOKENIZER_NAME)

    def count(self, text: str) -> int:
        if self._encoding is None:
            return max(1, (len(text.encode("utf-8")) + 3) // 4)
        return len(self._encoding.encode(text))


@dataclass
class ToolCall:
    prompt_id: str
    intent: str
    tool: str
    arguments: dict[str, Any]
    result: dict[str, Any] | None = None
    mode: str = DEFAULT_BENCHMARK_MODE
    source_tool_usage: list[dict[str, Any]] | None = None


@dataclass
class ServerMeasurement:
    server_id: str
    display_name: str
    version: str
    source: str
    tools: list[dict[str, Any]]
    calls: list[ToolCall]
    measurement_env: dict[str, str] | None = None
    mode_catalogs: dict[str, list[dict[str, Any]]] | None = None

    def catalogs_by_mode(self) -> dict[str, list[dict[str, Any]]]:
        if self.mode_catalogs:
            return self.mode_catalogs
        return {DEFAULT_BENCHMARK_MODE: self.tools}


class MCPClient:
    def __init__(
        self,
        command: list[str],
        env: dict[str, str] | None = None,
        timeout: float = 20.0,
    ) -> None:
        self.command = command
        merged_env = os.environ.copy()
        if env:
            merged_env.update(env)
        self.timeout = timeout
        self.proc = subprocess.Popen(
            command,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1,
            env=merged_env,
        )
        self._next_id = 1

    def close(self) -> None:
        if self.proc.poll() is None:
            self.proc.terminate()
            try:
                self.proc.wait(timeout=2)
            except subprocess.TimeoutExpired:
                self.proc.kill()
        if self.proc.stderr:
            try:
                stderr = self.proc.stderr.read()
            except Exception:
                stderr = ""
            if stderr.strip():
                print(stderr, file=sys.stderr)

    def request(
        self, method: str, params: dict[str, Any] | None = None
    ) -> dict[str, Any]:
        req_id = self._next_id
        self._next_id += 1
        message: dict[str, Any] = {"jsonrpc": "2.0", "id": req_id, "method": method}
        if params is not None:
            message["params"] = params
        self._write(message)
        deadline = time.time() + self.timeout
        while time.time() < deadline:
            response = self._read_line(deadline)
            if response is None:
                continue
            if response.get("id") != req_id:
                continue
            if "error" in response:
                raise BenchmarkError(f"MCP {method} failed: {response['error']}")
            result = response.get("result")
            if not isinstance(result, dict):
                raise BenchmarkError(
                    f"MCP {method} returned non-object result: {response}"
                )
            return result
        raise BenchmarkError(f"timed out waiting for MCP response to {method}")

    def notify(self, method: str, params: dict[str, Any] | None = None) -> None:
        message: dict[str, Any] = {"jsonrpc": "2.0", "method": method}
        if params is not None:
            message["params"] = params
        self._write(message)

    def _write(self, message: dict[str, Any]) -> None:
        if self.proc.stdin is None:
            raise BenchmarkError("MCP process stdin is closed")
        self.proc.stdin.write(canonical_json(message) + "\n")
        self.proc.stdin.flush()

    def _read_line(self, deadline: float) -> dict[str, Any] | None:
        if self.proc.stdout is None:
            raise BenchmarkError("MCP process stdout is closed")
        timeout = max(0.0, deadline - time.time())
        ready, _, _ = select.select([self.proc.stdout], [], [], timeout)
        if not ready:
            return None
        line = self.proc.stdout.readline()
        if line == "":
            stderr = ""
            if self.proc.stderr:
                stderr = self.proc.stderr.read()
            raise BenchmarkError(
                f"MCP process exited while reading stdout; stderr={stderr!r}"
            )
        return json.loads(line)


def load_prompt_set(path: Path) -> dict[str, Any]:
    data = json.loads(path.read_text(encoding="utf-8"))
    prompt_ids = [prompt["id"] for prompt in data.get("prompts", [])]
    if len(prompt_ids) != len(set(prompt_ids)):
        raise BenchmarkError("prompt IDs must be unique")
    if not prompt_ids:
        raise BenchmarkError("prompt set is empty")
    return data


def unavailable_result(server_id: str, intent: str) -> dict[str, Any]:
    return {
        "isError": True,
        "content": [
            {"type": "text", "text": f"intent unavailable on {server_id}: {intent}"}
        ],
    }


def normalize_source_tool_usage(value: Any) -> list[dict[str, Any]] | None:
    if value is None:
        return None
    if not isinstance(value, list):
        raise BenchmarkError("source_tool_usage must be a list when present")
    rows: list[dict[str, Any]] = []
    for item in value:
        if isinstance(item, str):
            rows.append({"tool": item, "count": 1})
            continue
        if not isinstance(item, dict) or not isinstance(item.get("tool"), str):
            raise BenchmarkError(
                "source_tool_usage entries must be tool strings or objects with a tool"
            )
        count = item.get("count", 1)
        if not isinstance(count, int) or count < 1:
            raise BenchmarkError("source_tool_usage count must be a positive integer")
        rows.append({"tool": item["tool"], "count": count})
    return rows


def load_fixture(path: Path) -> ServerMeasurement:
    data = json.loads(path.read_text(encoding="utf-8"))
    calls = [
        ToolCall(
            prompt_id=item["prompt_id"],
            intent=item["intent"],
            tool=item["tool"],
            arguments=item.get("arguments", {}),
            result=item["result"],
            mode=item.get("mode", DEFAULT_BENCHMARK_MODE),
            source_tool_usage=normalize_source_tool_usage(item.get("source_tool_usage")),
        )
        for item in data.get("calls", [])
    ]
    mode_catalogs = data.get("mode_catalogs")
    if mode_catalogs is not None and not isinstance(mode_catalogs, dict):
        raise BenchmarkError("mode_catalogs must be an object keyed by benchmark mode")
    return ServerMeasurement(
        server_id=data["server_id"],
        display_name=data.get("display_name", data["server_id"]),
        version=data.get("version", "unknown"),
        source=data.get("source", "fixture"),
        tools=data["tools"],
        calls=calls,
        measurement_env=data.get("measurement_env"),
        mode_catalogs=mode_catalogs,
    )


def load_fixtures(fixture_dir: Path) -> list[ServerMeasurement]:
    files = sorted(fixture_dir.glob("*.json"))
    if not files:
        raise BenchmarkError(f"no fixture JSON files found in {fixture_dir}")
    return [load_fixture(path) for path in files]


def load_live_measurements(config_path: Path) -> list[ServerMeasurement]:
    config = json.loads(config_path.read_text(encoding="utf-8"))
    measurements: list[ServerMeasurement] = []
    for server in config.get("servers", []):
        command = server.get("command")
        if not isinstance(command, list) or not command:
            raise BenchmarkError(f"server {server.get('id')} missing command array")
        client = MCPClient(
            command,
            env=server.get("env", {}),
            timeout=float(server.get("timeout_seconds", 20)),
        )
        try:
            client.request(
                "initialize",
                {
                    "protocolVersion": server.get("protocol_version", "2024-11-05"),
                    "capabilities": {},
                    "clientInfo": {"name": "icuvisor-kr5-benchmark", "version": "dev"},
                },
            )
            client.notify("notifications/initialized")
            tools_result = client.request("tools/list")
            calls: list[ToolCall] = []
            for item in server.get("calls", []):
                tool = item["tool"]
                if tool.startswith("unavailable:"):
                    result = unavailable_result(server["id"], item["intent"])
                else:
                    params = {"name": tool, "arguments": item.get("arguments", {})}
                    result = client.request("tools/call", params)
                calls.append(
                    ToolCall(
                        prompt_id=item["prompt_id"],
                        intent=item["intent"],
                        tool=tool,
                        arguments=item.get("arguments", {}),
                        result=result,
                        mode=item.get("mode", server.get("mode", DEFAULT_BENCHMARK_MODE)),
                        source_tool_usage=normalize_source_tool_usage(
                            item.get("source_tool_usage")
                        ),
                    )
                )
            measurements.append(
                ServerMeasurement(
                    server_id=server["id"],
                    display_name=server.get("display_name", server["id"]),
                    version=server.get("version", "unknown"),
                    source="live",
                    tools=tools_result.get("tools", []),
                    calls=calls,
                    measurement_env=server.get("env", {}),
                    mode_catalogs={
                        server.get("mode", DEFAULT_BENCHMARK_MODE): tools_result.get(
                            "tools", []
                        )
                    },
                )
            )
        finally:
            client.close()
    return measurements


def catalog_payload(tools: list[dict[str, Any]]) -> list[dict[str, Any]]:
    rows = []
    for tool in tools:
        rows.append(
            {
                "name": tool.get("name", ""),
                "description": tool.get("description", ""),
                "inputSchema": tool.get("inputSchema", tool.get("input_schema", {})),
            }
        )
    return sorted(rows, key=lambda row: row["name"])


def mode_names_for_prompt(prompt_set: dict[str, Any], prompt: dict[str, Any]) -> list[str]:
    modes = prompt.get("benchmark_modes", prompt_set.get("benchmark_modes"))
    if modes is None:
        return [DEFAULT_BENCHMARK_MODE]
    if not isinstance(modes, list) or not all(isinstance(mode, str) for mode in modes):
        raise BenchmarkError("benchmark_modes must be a list of mode strings")
    return modes


def is_analyzer_tool(tool_name: str) -> bool:
    return tool_name.startswith(ANALYZER_TOOL_PREFIXES) or tool_name in ANALYZER_DIRECT_TOOLS


def validate_analyzer_mode_catalogs(
    measurement: ServerMeasurement, tool_names_by_mode: dict[str, set[str]]
) -> None:
    if ANALYZER_ENABLED_MODE not in tool_names_by_mode or ANALYZER_DISABLED_MODE not in tool_names_by_mode:
        return
    enabled = tool_names_by_mode[ANALYZER_ENABLED_MODE]
    disabled = tool_names_by_mode[ANALYZER_DISABLED_MODE]
    disabled_analyzers = sorted(name for name in disabled if is_analyzer_tool(name))
    if disabled_analyzers:
        raise BenchmarkError(
            f"{measurement.server_id} disabled analyzer mode exposes analyzer tools: "
            + ", ".join(disabled_analyzers)
        )
    enabled_non_analyzers = {name for name in enabled if not is_analyzer_tool(name)}
    disabled_non_analyzers = {name for name in disabled if not is_analyzer_tool(name)}
    if enabled_non_analyzers != disabled_non_analyzers:
        raise BenchmarkError(
            f"{measurement.server_id} analyzer mode catalogs differ outside analyzer tools"
        )


def validate_measurement(
    prompt_set: dict[str, Any], measurement: ServerMeasurement
) -> None:
    prompt_by_id = {prompt["id"]: prompt for prompt in prompt_set["prompts"]}
    prompt_intents: dict[tuple[str, str, str], int] = {}
    for prompt in prompt_set["prompts"]:
        for mode in mode_names_for_prompt(prompt_set, prompt):
            for intent in prompt.get("required_intents", []):
                prompt_intents[(prompt["id"], intent, mode)] = 0
    catalogs = measurement.catalogs_by_mode()
    tool_names_by_mode = {
        mode: {tool.get("name", "") for tool in tools}
        for mode, tools in catalogs.items()
    }
    validate_analyzer_mode_catalogs(measurement, tool_names_by_mode)
    calls_by_prompt_mode: dict[tuple[str, str], set[str]] = {}
    for call in measurement.calls:
        if call.mode not in tool_names_by_mode:
            raise BenchmarkError(
                f"{measurement.server_id} call {call.prompt_id}:{call.intent} "
                f"uses unknown benchmark mode {call.mode!r}"
            )
        key = (call.prompt_id, call.intent, call.mode)
        if key in prompt_intents:
            prompt_intents[key] += 1
        calls_by_prompt_mode.setdefault((call.prompt_id, call.mode), set()).add(call.tool)
        if call.tool.startswith("unavailable:"):
            if not call.result or call.result.get("isError") is not True:
                raise BenchmarkError(
                    f"{measurement.server_id} unavailable call "
                    f"{call.prompt_id}:{call.intent} must set isError=true"
                )
            continue
        if call.tool not in tool_names_by_mode[call.mode]:
            raise BenchmarkError(
                f"{measurement.server_id} call {call.prompt_id}:{call.intent} "
                f"uses unlisted tool {call.tool!r} in mode {call.mode!r}"
            )
        if call.mode == ANALYZER_DISABLED_MODE and is_analyzer_tool(call.tool):
            raise BenchmarkError(
                f"{measurement.server_id} disabled analyzer mode calls {call.tool!r}"
            )
    for prompt in prompt_set["prompts"]:
        expected_by_mode = prompt.get("expected_tools_by_mode", {})
        if not isinstance(expected_by_mode, dict):
            raise BenchmarkError("expected_tools_by_mode must be an object when present")
        for mode, expected_tools in expected_by_mode.items():
            if not isinstance(expected_tools, list) or not all(
                isinstance(tool, str) for tool in expected_tools
            ):
                raise BenchmarkError("expected_tools_by_mode values must be string lists")
            called = calls_by_prompt_mode.get((prompt["id"], mode), set())
            missing_expected = sorted(set(expected_tools) - called)
            if missing_expected:
                raise BenchmarkError(
                    f"{measurement.server_id} {prompt['id']} mode {mode} missing "
                    f"expected tools: {', '.join(missing_expected)}"
                )
    missing = [
        f"{prompt_id}:{intent}:{mode}"
        for (prompt_id, intent, mode), count in prompt_intents.items()
        if count == 0 and prompt_id in prompt_by_id
    ]
    if missing:
        raise BenchmarkError(
            f"{measurement.server_id} missing call coverage for {', '.join(missing)}"
        )


def audited_response_bytes(result: dict[str, Any]) -> int | None:
    content = result.get("content")
    if not isinstance(content, list) or not content:
        return None
    first = content[0]
    if not isinstance(first, dict) or not isinstance(first.get("text"), str):
        return None
    try:
        payload = json.loads(first["text"])
    except json.JSONDecodeError:
        return None
    audit = payload.pop("redaction_audit", None)
    if not isinstance(audit, dict):
        return None
    raw_bytes = audit.get("raw_response_bytes")
    redacted_bytes = audit.get("redacted_response_bytes")
    if not isinstance(raw_bytes, int) or raw_bytes <= 0:
        raise BenchmarkError(
            "redaction_audit.raw_response_bytes must be a positive integer"
        )
    if not isinstance(redacted_bytes, int) or redacted_bytes <= 0:
        raise BenchmarkError(
            "redaction_audit.redacted_response_bytes must be a positive integer"
        )
    stripped_result = dict(result)
    stripped_content = list(content)
    stripped_first = dict(first)
    stripped_first["text"] = canonical_json(payload)
    stripped_content[0] = stripped_first
    stripped_result["content"] = stripped_content
    actual_redacted_bytes = len(canonical_bytes(stripped_result))
    redacted_tolerance = max(1, int(redacted_bytes * 0.01))
    if abs(actual_redacted_bytes - redacted_bytes) > redacted_tolerance:
        raise BenchmarkError(
            "redaction_audit.redacted_response_bytes does not match "
            "the committed redacted result after removing audit metadata"
        )
    raw_tolerance = max(1, int(raw_bytes * 0.01))
    if abs(raw_bytes - redacted_bytes) > raw_tolerance:
        raise BenchmarkError(
            "redacted response byte audit differs from raw by more than 1%"
        )
    return raw_bytes


def response_size(result: dict[str, Any]) -> int:
    audited = audited_response_bytes(result)
    if audited is not None:
        return audited
    return len(canonical_bytes(result))


def redact_env(env: dict[str, str] | None) -> dict[str, str]:
    if not env:
        return {}
    redacted: dict[str, str] = {}
    for key, value in sorted(env.items()):
        upper = key.upper()
        if (
            "ATHLETE" in upper
            or "KEY" in upper
            or "TOKEN" in upper
            or "SECRET" in upper
            or "PASSWORD" in upper
        ):
            redacted[key] = "<redacted>"
        else:
            redacted[key] = value
    return redacted


def raw_stream_pull_count(calls: list[ToolCall]) -> int:
    return sum(
        1
        for call in calls
        if not call.tool.startswith("unavailable:") and call.tool in RAW_STREAM_TOOL_NAMES
    )


def summarize(
    prompt_set: dict[str, Any],
    measurements: list[ServerMeasurement],
    token_counter: TokenCounter,
    generated_at: str | None = None,
) -> dict[str, Any]:
    servers = []
    benchmark_modes: set[str] = set()
    for measurement in measurements:
        validate_measurement(prompt_set, measurement)
        catalogs = measurement.catalogs_by_mode()
        benchmark_modes.update(catalogs)
        calls_by_mode: dict[str, list[ToolCall]] = {mode: [] for mode in catalogs}
        response_rows = []
        for index, call in enumerate(measurement.calls, start=1):
            if call.result is None:
                raise BenchmarkError(
                    f"{measurement.server_id} call {index} has no result"
                )
            if call.mode not in calls_by_mode:
                calls_by_mode[call.mode] = []
            calls_by_mode[call.mode].append(call)
            size = response_size(call.result)
            response_tokens = token_counter.count(canonical_json(call.result))
            row = {
                "prompt_id": call.prompt_id,
                "intent": call.intent,
                "mode": call.mode,
                "tool": call.tool,
                "response_bytes": size,
                "response_tokens": response_tokens,
            }
            if call.source_tool_usage is not None:
                row["source_tool_usage"] = call.source_tool_usage
            response_rows.append(row)
        mode_summaries: dict[str, dict[str, Any]] = {}
        for mode, tools in sorted(catalogs.items()):
            catalog = catalog_payload(tools)
            catalog_json = canonical_json(catalog)
            mode_calls = calls_by_mode.get(mode, [])
            mode_response_bytes = [response_size(call.result or {}) for call in mode_calls]
            mode_response_tokens = [
                token_counter.count(canonical_json(call.result or {})) for call in mode_calls
            ]
            mode_summaries[mode] = {
                "tool_count": len(catalog),
                "description_tokens": token_counter.count(catalog_json),
                "catalog_bytes": len(catalog_json.encode("utf-8")),
                "call_count": len(mode_calls),
                "response_tokens": sum(mode_response_tokens),
                "median_response_bytes": statistics.median(mode_response_bytes)
                if mode_response_bytes
                else 0,
                "median_response_tokens": statistics.median(mode_response_tokens)
                if mode_response_tokens
                else 0,
                "raw_stream_pull_count": raw_stream_pull_count(mode_calls),
            }
        default_summary = mode_summaries.get(DEFAULT_BENCHMARK_MODE) or next(
            iter(mode_summaries.values())
        )
        servers.append(
            {
                "server_id": measurement.server_id,
                "display_name": measurement.display_name,
                "version": measurement.version,
                "source": measurement.source,
                "measurement_env": redact_env(measurement.measurement_env),
                "tool_count": default_summary["tool_count"],
                "description_tokens": default_summary["description_tokens"],
                "catalog_bytes": default_summary["catalog_bytes"],
                "median_response_bytes": default_summary["median_response_bytes"],
                "call_count": len(response_rows),
                "mode_summaries": mode_summaries,
                "calls": response_rows,
            }
        )
    return {
        "schema_version": "kr5-benchmark-result-v2",
        "generated_at": generated_at
        or time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "prompt_set_version": prompt_set["version"],
        "prompt_count": len(prompt_set["prompts"]),
        "benchmark_modes": sorted(benchmark_modes) or [DEFAULT_BENCHMARK_MODE],
        "tokenizer": {"name": token_counter.name, "package": token_counter.package},
        "servers": sorted(servers, key=lambda item: item["server_id"]),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--mode", choices=["fixtures", "live"], default="fixtures")
    parser.add_argument("--prompt-set", type=Path, required=True)
    parser.add_argument(
        "--fixture-dir", type=Path, default=Path("scripts/benchmark/testdata/fixtures")
    )
    parser.add_argument("--config", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument(
        "--allow-approx-tokenizer",
        action="store_true",
        help="smoke tests only; not valid for KR5 results",
    )
    parser.add_argument(
        "--generated-at",
        help="fixed RFC3339 timestamp for reproducible committed results",
    )
    args = parser.parse_args()

    try:
        prompt_set = load_prompt_set(args.prompt_set)
        token_counter = TokenCounter(allow_approx=args.allow_approx_tokenizer)
        if args.mode == "fixtures":
            measurements = load_fixtures(args.fixture_dir)
        else:
            if args.config is None:
                raise BenchmarkError("--config is required for live mode")
            measurements = load_live_measurements(args.config)
        result = summarize(
            prompt_set, measurements, token_counter, generated_at=args.generated_at
        )
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(canonical_json(result) + "\n", encoding="utf-8")
        print(f"wrote {args.output}")
        return 0
    except BenchmarkError as exc:
        print(f"kr5_benchmark: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
