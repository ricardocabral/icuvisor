#!/usr/bin/env python3
"""Print the changelog body for one release tag."""

from __future__ import annotations

import re
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: release_notes_from_changelog.py vX.Y.Z", file=sys.stderr)
        return 2

    version = sys.argv[1].strip()
    if version.startswith("v"):
        version = version[1:]
    if not version:
        print("error: empty release version", file=sys.stderr)
        return 2

    changelog_path = Path(__file__).resolve().parents[1] / "CHANGELOG.md"
    changelog = changelog_path.read_text(encoding="utf-8")
    pattern = re.compile(
        rf"^## \[{re.escape(version)}\][^\n]*\n(?P<body>.*?)(?=^## \[|\Z)",
        re.MULTILINE | re.DOTALL,
    )
    match = pattern.search(changelog)
    if match is None:
        print(f"error: CHANGELOG.md has no section for [{version}]", file=sys.stderr)
        return 1

    body = match.group("body").strip()
    if not body:
        print(f"error: CHANGELOG.md section [{version}] is empty", file=sys.stderr)
        return 1

    print(body)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
