#!/usr/bin/env python3
"""Prevent regressions in portable structured-workout cookbook guidance."""

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
COOKBOOK = ROOT / "web/content/cookbook/build-workouts.md"
DOGFOOD = ROOT / "docs/dogfood/resource-blind-workout-authoring.md"


def require(content: str, phrase: str, path: Path) -> list[str]:
    if phrase not in content:
        return [f"{path.relative_to(ROOT)} must contain: {phrase!r}"]
    return []


def forbid(content: str, pattern: str, path: Path, description: str) -> list[str]:
    if re.search(pattern, content, re.IGNORECASE | re.DOTALL):
        return [f"{path.relative_to(ROOT)} must not {description}"]
    return []


def require_matrix_row(content: str, case: str, path: Path) -> list[str]:
    match = re.search(rf"^\|\s*{re.escape(case)}\s*\|(?P<row>[^\n]*)$", content, re.MULTILINE)
    if match is None:
        return [f"{path.relative_to(ROOT)} must contain a matrix row for {case}"]
    if match.group("row").count("not run") < 8:
        return [f"{path.relative_to(ROOT)} matrix row for {case} must remain not run"]
    return []


def main() -> int:
    failures: list[str] = []
    cookbook = COOKBOOK.read_text(encoding="utf-8")
    dogfood = DOGFOOD.read_text(encoding="utf-8")

    for phrase in (
        "Resource registration alone is not enough.",
        "A listed Resource,\nits URI, or `resources/list` is not evidence that you can read it.",
        "portable structured-tool path",
        "structured `workout_doc`",
        "Call\n`validate_workout` with that `workout_doc`.",
        "Continue only when `valid: true`",
        "Ask for approval of the exact preview. Only after I approve",
        "`workout.workout_doc_summary`",
        "`event.workout_doc_summary`",
        "`_meta.workout_doc_warning`",
        "does not verify structured rendering",
        "did not render the uploaded structure or only partially preserved it",
    ):
        failures.extend(require(cookbook, phrase, COOKBOOK))

    failures.extend(
        forbid(
            cookbook,
            r"(?:must|always|first)\s+(?:read|check|consult)"
            r".{0,160}?icuvisor://workout-syntax",
            COOKBOOK,
            "make the workout-syntax Resource mandatory before authoring",
        )
    )
    failures.extend(
        forbid(
            cookbook,
            r"forcing\s+the\s+assistant\s+to\s+consult",
            COOKBOOK,
            "claim that forcing a Resource read is required",
        )
    )

    for phrase in (
        "manual test protocol, not a host-compatibility report",
        "dedicated test athlete",
        "Observable Resource-read state to model",
        "Exact validation/write inputs",
        "Returned summary evidence",
        "`_meta.workout_doc_warning`",
        "B1 bike",
        "R1 run",
        "SM1 metric swim",
        "SY1 yard swim",
        "RP1 repeat",
        "RM1 ramp",
        "Do not turn blank rows into a compatibility matrix",
    ):
        failures.extend(require(dogfood, phrase, DOGFOOD))

    for case in ("B1 bike", "R1 run", "SM1 metric swim", "SY1 yard swim", "RP1 repeat", "RM1 ramp"):
        failures.extend(require_matrix_row(dogfood, case, DOGFOOD))

    if failures:
        print("Build-workouts guidance contract failed:", file=sys.stderr)
        print("\n".join(f"- {failure}" for failure in failures), file=sys.stderr)
        return 1

    print("Build-workouts guidance contract passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
