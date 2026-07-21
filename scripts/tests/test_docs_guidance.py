#!/usr/bin/env python3
"""Prevent regressions in published athlete-ID and hosted HTTP guidance."""

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ATHLETE_ID_GUIDES = (
    ROOT / "web/content/reference/config-file.md",
    ROOT / "web/content/guides/coach-mode.md",
    ROOT / "CONTRIBUTING.md",
)
HTTP_GUIDE = ROOT / "web/content/guides/http-transport.md"
README = ROOT / "README.md"


def require(content: str, phrase: str, path: Path) -> list[str]:
    if phrase not in content:
        return [f"{path.relative_to(ROOT)} must contain: {phrase!r}"]
    return []


def main() -> int:
    failures: list[str] = []

    expected_guidance = {
        ATHLETE_ID_GUIDES[0]: (
            "Surrounding whitespace is trimmed",
            "optional leading `I` is lowercased to `i`",
            "remaining characters must be digits",
            "`12345` remains `12345`",
            "`i12345` remains `i12345`",
        ),
        ATHLETE_ID_GUIDES[1]: (
            "trims surrounding whitespace",
            "lowercases only an optional leading `I`",
            "validates the remaining digits",
            "`12345` remains `12345`",
            "`i12345` remains `i12345`",
        ),
        ATHLETE_ID_GUIDES[2]: (
            "trim whitespace",
            "lowercase only an optional leading `I`",
            "validate digits",
            "never add or remove the `i` prefix",
        ),
    }
    for path, phrases in expected_guidance.items():
        content = path.read_text(encoding="utf-8")
        for phrase in phrases:
            failures.extend(require(content, phrase, path))

    false_normalization = re.compile(
        r"normaliz\w*[^.\n]{0,100}?to\s+`?i12345`?", re.IGNORECASE
    )
    published_docs = list((ROOT / "web/content").rglob("*.md")) + [
        ROOT / "CONTRIBUTING.md"
    ]
    for path in published_docs:
        if false_normalization.search(path.read_text(encoding="utf-8")):
            failures.append(
                f"{path.relative_to(ROOT)} must not claim athlete IDs normalize to i12345"
            )

    http_content = HTTP_GUIDE.read_text(encoding="utf-8")
    for phrase in (
        "https://connect.icuvisor.app/mcp",
        "hosted ICU Visor connector",
        "Do not expose local loopback HTTP through a generic public tunnel",
        "a tunnel URL is not authentication",
    ):
        failures.extend(require(http_content, phrase, HTTP_GUIDE))

    readme_content = README.read_text(encoding="utf-8")
    for phrase in (
        "https://icuvisor.app/install/",
        "https://icuvisor.app/connect/",
        "https://icuvisor.app/cookbook/",
    ):
        failures.extend(require(readme_content, phrase, README))
    for heading in (
        "### Connect from Cursor",
        "### MCP discovery",
        "### Downloadable prompt packs",
        "### Fitness projection with ATP/periodization targets",
    ):
        if heading in readme_content:
            failures.append(
                f"README.md must route user documentation to icuvisor.app instead of retaining {heading!r}"
            )

    if failures:
        print("Documentation guidance contract failed:", file=sys.stderr)
        print("\n".join(f"- {failure}" for failure in failures), file=sys.stderr)
        return 1

    print("Documentation guidance contract passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
