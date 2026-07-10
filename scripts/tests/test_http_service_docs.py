#!/usr/bin/env python3
"""Prevent insecure regressions in persistent local HTTP service guidance."""

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SERVICE_GUIDE = ROOT / "web/content/guides/persistent-http-service.md"
HTTP_GUIDE = ROOT / "web/content/guides/http-transport.md"
GUIDES = (SERVICE_GUIDE, HTTP_GUIDE)
EXECUTABLE_LANGUAGES = frozenset(
    {
        "bash",
        "sh",
        "shell",
        "powershell",
        "ps1",
        "json",
        "xml",
        "plist",
        "ini",
        "systemd",
    }
)
FENCE = re.compile(r"```(?P<language>[\w+-]*)\n(?P<body>.*?)```", re.DOTALL)
CANONICAL_BIND = "127.0.0.1:8765"


def require(content: str, phrase: str, path: Path) -> list[str]:
    if phrase not in content:
        return [f"{path.relative_to(ROOT)} must contain: {phrase!r}"]
    return []


def executable_snippets(content: str) -> list[str]:
    return [
        match.group("body")
        for match in FENCE.finditer(content)
        if match.group("language").lower() in EXECUTABLE_LANGUAGES
    ]


def normalize_service_markup(snippet: str) -> str:
    return re.sub(r"<[^>]+>", " ", snippet)


def validate_executable_samples(path: Path, content: str) -> list[str]:
    failures: list[str] = []
    bind_count = 0

    for number, snippet in enumerate(executable_snippets(content), start=1):
        normalized = normalize_service_markup(snippet)
        location = f"{path.relative_to(ROOT)} fenced executable sample {number}"

        for pattern, description in (
            (r"\bINTERVALS_ICU_API_KEY\b", "API-key environment source"),
            (r"(?i)\.env\b", "dotenv source"),
            (r"--env-file\b", "environment-file source"),
            (r"(?i)\bapi_key\b", "plaintext config API-key field"),
            (r"(?i)\bEnvironmentVariables\b", "launchd environment directive"),
            (r"(?im)^\s*Environment(?:File)?\s*=", "systemd environment directive"),
        ):
            if re.search(pattern, normalized):
                failures.append(f"{location} must not contain a {description}")

        if re.search(r"\b0\.0\.0\.0(?::\d+)?\b", normalized):
            failures.append(f"{location} must not contain a wildcard IPv4 bind")
        if re.search(r"(?:\[::\]|::):\d+", normalized):
            failures.append(f"{location} must not contain a wildcard IPv6 bind")

        bind_patterns = (
            r"--http-bind\s+([^\s]+)",
            r'"http_bind"\s*:\s*"([^"]+)"',
            r"\bICUVISOR_HTTP_BIND\s*=\s*([^\s]+)",
        )
        for pattern in bind_patterns:
            for match in re.finditer(pattern, normalized):
                bind_count += 1
                address = match.group(1).strip("'\"`;,")
                if address != CANONICAL_BIND:
                    failures.append(
                        f"{location} must bind {CANONICAL_BIND}, not {address!r}"
                    )

    if bind_count == 0:
        failures.append(
            f"{path.relative_to(ROOT)} must include an executable loopback HTTP bind"
        )
    return failures


def main() -> int:
    failures: list[str] = []
    contents = {path: path.read_text(encoding="utf-8") for path in GUIDES}
    service_content = contents[SERVICE_GUIDE]
    http_content = contents[HTTP_GUIDE]

    for phrase in (
        "## macOS: LaunchAgent",
        "## Linux: systemd user service",
        "## Windows: Task Scheduler",
        "interactively as the same OS account",
        "credential store",
        "--transport http --http-bind 127.0.0.1:8765",
        "http://127.0.0.1:8765/mcp",
        "launchctl bootstrap",
        "launchctl print",
        "launchctl kickstart -k",
        "launchctl bootout",
        "tail -n 100",
        "systemctl --user enable --now",
        "systemctl --user status",
        "journalctl --user",
        "systemctl --user restart",
        "systemctl --user disable --now",
        "Register-ScheduledTask",
        "Get-ScheduledTaskInfo",
        "Get-Content -LiteralPath",
        "Stop-ScheduledTask",
        "Unregister-ScheduledTask",
        "https://connect.icuvisor.app/mcp",
        "hosted OAuth",
        "generic public tunnel",
        "use `icuvisor`",
    ):
        failures.extend(require(service_content, phrase, SERVICE_GUIDE))

    for phrase in (
        "--transport http --http-bind 127.0.0.1:8765",
        '"http_bind": "127.0.0.1:8765"',
        "http://127.0.0.1:8765/mcp",
        "Keep local HTTP running",
        "https://connect.icuvisor.app/mcp",
        "a tunnel URL is not authentication",
    ):
        failures.extend(require(http_content, phrase, HTTP_GUIDE))

    for path, content in contents.items():
        failures.extend(validate_executable_samples(path, content))

    if failures:
        print("Persistent HTTP service documentation contract failed:", file=sys.stderr)
        print("\n".join(f"- {failure}" for failure in failures), file=sys.stderr)
        return 1

    print("Persistent HTTP service documentation contract passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
