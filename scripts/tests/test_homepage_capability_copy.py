#!/usr/bin/env python3
"""Lock factual homepage claims for workout fidelity and fitness scenarios."""

import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HOMEPAGE = ROOT / "web/layouts/index.html"
HUGO_CONFIG = ROOT / "web/hugo.toml"
WORKOUT_WARNING = (
    "structured workouts, with warnings when returned intervals.icu steps are missing or changed"
)
FITNESS_SCENARIO = (
    "Show a deterministic race-day fitness scenario from my planned load, not a forecast."
)
SEO_DESCRIPTION = (
    "Open-source MCP for intervals.icu data. Run locally by default or use the hosted HTTPS "
    "connector; returned workout-step warnings and deterministic fitness scenarios, not forecasts."
)
OVERSTATED_GUARANTEES = (
    "automatically successful",
    "valid for every device",
    "device-compatible",
    "appropriate for every athlete",
    "guarantees",
    "guaranteed",
)


def main() -> int:
    homepage = HOMEPAGE.read_text(encoding="utf-8")
    hugo_config = HUGO_CONFIG.read_text(encoding="utf-8")
    failures: list[str] = []

    for phrase, path in (
        (WORKOUT_WARNING, HOMEPAGE),
        (FITNESS_SCENARIO, HOMEPAGE),
        (SEO_DESCRIPTION, HUGO_CONFIG),
    ):
        if phrase not in path.read_text(encoding="utf-8"):
            failures.append(f"{path.relative_to(ROOT)} must contain: {phrase!r}")

    published_copy = f"{homepage}\n{hugo_config}".lower()
    for phrase in OVERSTATED_GUARANTEES:
        if phrase in published_copy:
            failures.append(f"homepage copy must not promise: {phrase!r}")

    if failures:
        print("Homepage capability copy contract failed:", file=sys.stderr)
        print("\n".join(f"- {failure}" for failure in failures), file=sys.stderr)
        return 1

    print("Homepage capability copy contract passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
