#!/usr/bin/env python3
"""Measure KR5 catalog-token and response-byte metrics for MCP servers."""

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


@dataclass
class ServerMeasurement:
    server_id: str
    display_name: str
    version: str
    source: str
    tools: list[dict[str, Any]]
    calls: list[ToolCall]
    measurement_env: dict[str, str] | None = None


class MCPClient:
    def __init__(self, command: list[str], env: dict[str, str] | None = None, timeout: float = 20.0) -> None:
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

    def request(self, method: str, params: dict[str, Any] | None = None) -> dict[str, Any]:
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
                raise BenchmarkError(f"MCP {method} returned non-object result: {response}")
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
            raise BenchmarkError(f"MCP process exited while reading stdout; stderr={stderr!r}")
        return json.loads(line)


def load_prompt_set(path: Path) -> dict[str, Any]:
    data = json.loads(path.read_text(encoding="utf-8"))
    prompt_ids = [prompt["id"] for prompt in data.get("prompts", [])]
    if len(prompt_ids) != len(set(prompt_ids)):
        raise BenchmarkError("prompt IDs must be unique")
    if not prompt_ids:
        raise BenchmarkError("prompt set is empty")
    return data


def load_fixture(path: Path) -> ServerMeasurement:
    data = json.loads(path.read_text(encoding="utf-8"))
    calls = [
        ToolCall(
            prompt_id=item["prompt_id"],
            intent=item["intent"],
            tool=item["tool"],
            arguments=item.get("arguments", {}),
            result=item["result"],
        )
        for item in data.get("calls", [])
    ]
    return ServerMeasurement(
        server_id=data["server_id"],
        display_name=data.get("display_name", data["server_id"]),
        version=data.get("version", "unknown"),
        source=data.get("source", "fixture"),
        tools=data["tools"],
        calls=calls,
        measurement_env=data.get("measurement_env"),
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
        client = MCPClient(command, env=server.get("env", {}), timeout=float(server.get("timeout_seconds", 20)))
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
                params = {"name": item["tool"], "arguments": item.get("arguments", {})}
                result = client.request("tools/call", params)
                calls.append(
                    ToolCall(
                        prompt_id=item["prompt_id"],
                        intent=item["intent"],
                        tool=item["tool"],
                        arguments=item.get("arguments", {}),
                        result=result,
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


def validate_measurement(prompt_set: dict[str, Any], measurement: ServerMeasurement) -> None:
    prompt_intents: dict[tuple[str, str], int] = {}
    for prompt in prompt_set["prompts"]:
        for intent in prompt.get("required_intents", []):
            prompt_intents[(prompt["id"], intent)] = 0
    for call in measurement.calls:
        key = (call.prompt_id, call.intent)
        if key in prompt_intents:
            prompt_intents[key] += 1
    missing = [f"{prompt_id}:{intent}" for (prompt_id, intent), count in prompt_intents.items() if count == 0]
    if missing:
        raise BenchmarkError(f"{measurement.server_id} missing call coverage for {', '.join(missing)}")


def redact_env(env: dict[str, str] | None) -> dict[str, str]:
    if not env:
        return {}
    redacted: dict[str, str] = {}
    for key, value in sorted(env.items()):
        upper = key.upper()
        if "KEY" in upper or "TOKEN" in upper or "SECRET" in upper or "PASSWORD" in upper:
            redacted[key] = "<redacted>"
        else:
            redacted[key] = value
    return redacted


def summarize(
    prompt_set: dict[str, Any],
    measurements: list[ServerMeasurement],
    token_counter: TokenCounter,
    generated_at: str | None = None,
) -> dict[str, Any]:
    servers = []
    for measurement in measurements:
        validate_measurement(prompt_set, measurement)
        catalog = catalog_payload(measurement.tools)
        catalog_json = canonical_json(catalog)
        response_rows = []
        response_bytes = []
        for index, call in enumerate(measurement.calls, start=1):
            if call.result is None:
                raise BenchmarkError(f"{measurement.server_id} call {index} has no result")
            size = len(canonical_bytes(call.result))
            response_bytes.append(size)
            response_rows.append(
                {
                    "prompt_id": call.prompt_id,
                    "intent": call.intent,
                    "tool": call.tool,
                    "response_bytes": size,
                }
            )
        median_bytes = int(statistics.median(response_bytes)) if response_bytes else 0
        servers.append(
            {
                "server_id": measurement.server_id,
                "display_name": measurement.display_name,
                "version": measurement.version,
                "source": measurement.source,
                "measurement_env": redact_env(measurement.measurement_env),
                "tool_count": len(catalog),
                "description_tokens": token_counter.count(catalog_json),
                "catalog_bytes": len(catalog_json.encode("utf-8")),
                "median_response_bytes": median_bytes,
                "call_count": len(response_rows),
                "calls": response_rows,
            }
        )
    return {
        "schema_version": "kr5-benchmark-result-v1",
        "generated_at": generated_at or time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "prompt_set_version": prompt_set["version"],
        "prompt_count": len(prompt_set["prompts"]),
        "tokenizer": {"name": token_counter.name, "package": token_counter.package},
        "servers": sorted(servers, key=lambda item: item["server_id"]),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--mode", choices=["fixtures", "live"], default="fixtures")
    parser.add_argument("--prompt-set", type=Path, required=True)
    parser.add_argument("--fixture-dir", type=Path, default=Path("scripts/benchmark/testdata/fixtures"))
    parser.add_argument("--config", type=Path)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--allow-approx-tokenizer", action="store_true", help="smoke tests only; not valid for KR5 results")
    parser.add_argument("--generated-at", help="fixed RFC3339 timestamp for reproducible committed results")
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
        result = summarize(prompt_set, measurements, token_counter, generated_at=args.generated_at)
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(canonical_json(result) + "\n", encoding="utf-8")
        print(f"wrote {args.output}")
        return 0
    except BenchmarkError as exc:
        print(f"kr5_benchmark: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
