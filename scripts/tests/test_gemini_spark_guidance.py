#!/usr/bin/env python3
"""Keep Gemini Spark guidance aligned with the core/host compatibility boundary."""

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
GEMINI_GUIDE = ROOT / "web/content/connect/gemini-spark.md"
CLIENT_INDEX = ROOT / "web/content/connect/_index.md"
OTHER_CLIENTS = ROOT / "web/content/connect/other-clients.md"
HTTP_GUIDE = ROOT / "web/content/guides/http-transport.md"


def require(content: str, phrase: str, path: Path) -> list[str]:
    if phrase not in content:
        return [f"{path.relative_to(ROOT)} must contain: {phrase!r}"]
    return []


def forbid(content: str, phrase: str, path: Path, reason: str) -> list[str]:
    if phrase.casefold() in content.casefold():
        return [f"{path.relative_to(ROOT)} must not {reason}: {phrase!r}"]
    return []


def main() -> int:
    failures: list[str] = []
    gemini = GEMINI_GUIDE.read_text(encoding="utf-8")
    index = CLIENT_INDEX.read_text(encoding="utf-8")
    other_clients = OTHER_CLIENTS.read_text(encoding="utf-8")
    http = HTTP_GUIDE.read_text(encoding="utf-8")

    for phrase in (
        "**core icuvisor** connection boundary",
        "in-process Streamable HTTP smoke test",
        "`initialize`, `ping`, `tools/list`",
        "No end-to-end Gemini Spark account or mobile test has been performed",
        "http://127.0.0.1:8765/mcp",
        "Core local HTTP is loopback-bound and has no OAuth layer",
        "Store the intervals.icu API key with `icuvisor setup`",
        "A Gemini web or mobile surface cannot normally reach `127.0.0.1`",
        "Do not replace it with a generic public tunnel",
        "public HTTPS MCP URL, OAuth, or Dynamic Client Registration (DCR)",
        "belongs in `icuvisor-host`",
        "start a fresh Gemini conversation",
    ):
        failures.extend(require(gemini, phrase, GEMINI_GUIDE))

    for phrase, reason in (
        ("https://connect.icuvisor.app/mcp", "hard-code a hosted endpoint"),
        ("Gemini supports mobile", "claim an unverified mobile integration"),
        ("Gemini mobile supports icuvisor", "claim an unverified mobile integration"),
        ("core supports OAuth", "claim hosted authentication in core"),
        ("core supports DCR", "claim hosted registration in core"),
        ("OAuth is supported", "claim hosted authentication in core"),
    ):
        failures.extend(forbid(gemini, phrase, GEMINI_GUIDE, reason))

    for phrase in (
        'link="gemini-spark"',
        "Gemini Spark",
    ):
        failures.extend(require(index, phrase, CLIENT_INDEX))
    for phrase in (
        "| Gemini Spark",
        '[Gemini Spark setup]({{< relref "gemini-spark" >}})',
        "local Streamable HTTP only",
        "Gemini web/mobile surface",
        "hosted integration, which core does not provide",
    ):
        failures.extend(require(other_clients, phrase, OTHER_CLIENTS))

    for phrase in (
        "http://127.0.0.1:8765/mcp",
        "Keep `127.0.0.1:8765`",
        "unauthenticated MCP server",
        "a tunnel URL is not authentication",
    ):
        failures.extend(require(http, phrase, HTTP_GUIDE))

    if failures:
        print("Gemini Spark guidance contract failed:", file=sys.stderr)
        print("\n".join(f"- {failure}" for failure in failures), file=sys.stderr)
        return 1

    print("Gemini Spark guidance contract passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
