#!/usr/bin/env python3
"""Validate README distribution CTAs that are easy to break by hand."""

from __future__ import annotations

import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

README = Path("README.md")
CURSOR_LINK_RE = re.compile(r"https://cursor\.com/install-mcp\?[^)\s>]+")
SECRET_NAME_RE = re.compile(r"(api|token|secret|password|key|authorization)", re.IGNORECASE)
ALLOWED_EMPTY_ENV = {"ICUVISOR_CONFIG"}
LINKS_TO_CHECK = (
    "https://cursor.com/install-mcp?name=icuvisor",
    "https://icuvisor.app",
)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check-links", action="store_true", help="also make HEAD/GET requests to stable public URLs")
    args = parser.parse_args()

    errors = validate_readme(README)
    if args.check_links:
        errors.extend(check_links(LINKS_TO_CHECK))

    if errors:
        for error in errors:
            print(f"README distribution validation: {error}", file=sys.stderr)
        return 1
    print("README distribution validation: ok")
    return 0


def validate_readme(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    errors: list[str] = []

    match = CURSOR_LINK_RE.search(text)
    if not match:
        return ["Cursor install link is missing"]

    link = match.group(0)
    parsed = urllib.parse.urlparse(link)
    query = urllib.parse.parse_qs(parsed.query)
    if query.get("name") != ["icuvisor"]:
        errors.append("Cursor install link must use name=icuvisor")

    configs = query.get("config")
    if len(configs) != 1:
        errors.append("Cursor install link must contain exactly one config parameter")
        return errors

    try:
        config = json.loads(configs[0])
    except json.JSONDecodeError as exc:
        errors.append(f"Cursor config is not JSON: {exc}")
        return errors

    if not isinstance(config, dict):
        errors.append("Cursor config must decode to an object")
        return errors

    if config.get("command") != "icuvisor":
        errors.append("Cursor config command must be icuvisor")
    if config.get("args") != []:
        errors.append("Cursor config args must be an empty array")
    if config.get("type") not in (None, "stdio"):
        errors.append("Cursor config transport type, when present, must be stdio")

    env = config.get("env", {})
    if not isinstance(env, dict):
        errors.append("Cursor config env must be an object")
        return errors
    for name, value in env.items():
        if name not in ALLOWED_EMPTY_ENV:
            errors.append(f"Cursor config env contains unexpected variable {name}")
        if value not in ("", "<optional path to icuvisor config>"):
            errors.append(f"Cursor config env {name} must be empty or an obvious placeholder")
        if SECRET_NAME_RE.search(name) and value:
            errors.append(f"Cursor config env {name} looks secret-related and must not have a value")

    cursor_link_text = urllib.parse.unquote(match.group(0))
    forbidden = ("ICUVISOR_API_KEY", "Authorization", "Bearer ", "Basic ")
    for token in forbidden:
        if token in cursor_link_text:
            errors.append(f"Cursor install link must not embed {token}")

    if "Glama" in text and "No Glama badge" not in text:
        glama_links = [link for link in re.findall(r"https://glama\.ai/[^)\s>]+", text)]
        if not glama_links:
            errors.append("Glama is mentioned without either a stable Glama link or the no-badge rationale")

    return errors


def check_links(urls: tuple[str, ...]) -> list[str]:
    errors: list[str] = []
    for url in urls:
        request = urllib.request.Request(url, method="HEAD", headers={"User-Agent": "icuvisor-distribution-check"})
        try:
            with urllib.request.urlopen(request, timeout=10) as response:
                if response.status >= 400:
                    errors.append(f"{url} returned HTTP {response.status}")
        except urllib.error.HTTPError as exc:
            errors.append(f"{url} returned HTTP {exc.code}")
        except urllib.error.URLError:
            get_request = urllib.request.Request(url, method="GET", headers={"User-Agent": "icuvisor-distribution-check"})
            try:
                with urllib.request.urlopen(get_request, timeout=10) as response:
                    if response.status >= 400:
                        errors.append(f"{url} returned HTTP {response.status}")
            except Exception as exc:  # noqa: BLE001 - report validation failures tersely.
                errors.append(f"{url} could not be reached: {exc}")
    return errors


if __name__ == "__main__":
    raise SystemExit(main())
