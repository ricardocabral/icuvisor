#!/usr/bin/env python3
"""Block release builds for tags that already have a published GitHub release."""

import argparse
import json
import re
import subprocess
import sys


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("repository")
    parser.add_argument("tag")
    parser.add_argument("--skip-package-managers", action="store_true")
    args = parser.parse_args()
    repository, tag = args.repository, args.tag
    if not re.fullmatch(
        r"v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?", tag
    ):
        parser.error("expected a release tag such as v1.2.3, not a branch or Git ref")
    try:
        result = subprocess.run(
            ["gh", "api", "--paginate", "--slurp", f"repos/{repository}/releases"],
            capture_output=True,
            text=True,
            check=True,
            timeout=60,
        )
        pages = json.loads(result.stdout)
        if not isinstance(pages, list) or not pages:
            raise ValueError("expected release pages")
        published = False
        drafts = []
        for page in pages:
            if not isinstance(page, list):
                raise ValueError("expected a release list")
            for release in page:
                if not isinstance(release, dict):
                    raise ValueError("expected a release")
                if not isinstance(release["tag_name"], str) or not isinstance(
                    release["draft"], bool
                ):
                    raise ValueError("invalid release state")
                if release["tag_name"] == tag:
                    if release["draft"]:
                        drafts.append(release)
                    else:
                        published = True
        missing_archives = set()
        if args.skip_package_managers and not published and len(drafts) == 1:
            version = tag.removeprefix("v")
            expected_archives = {
                f"icuvisor_{version}_{suffix}"
                for suffix in (
                    "linux_amd64.tar.gz",
                    "linux_arm64.tar.gz",
                    "windows_amd64.zip",
                    "windows_arm64.zip",
                    "macos_universal.tar.gz",
                )
            }
            assets = drafts[0]["assets"]
            if not isinstance(assets, list):
                raise ValueError("expected draft assets")
            names = set()
            for asset in assets:
                if not isinstance(asset, dict) or not isinstance(asset["name"], str):
                    raise ValueError("invalid draft asset")
                if (
                    asset.get("state") == "uploaded"
                    and type(asset.get("size")) is int
                    and asset["size"] > 0
                ):
                    names.add(asset["name"])
            missing_archives = expected_archives - names
    except (OSError, subprocess.SubprocessError, ValueError, KeyError) as error:
        print(
            f"error: cannot verify release state for {tag} ({type(error).__name__}); "
            "check GitHub access and retry",
            file=sys.stderr,
        )
        return 1

    if published:
        print(
            f"error: {tag} is already published; use the manual Winget recovery "
            "workflow for Winget failures, or ship a new patch for incorrect artifacts",
            file=sys.stderr,
        )
        return 1

    if args.skip_package_managers:
        if len(drafts) != 1:
            print(
                "error: skipping package managers requires exactly one existing draft; "
                "reconcile the release state before recovery",
                file=sys.stderr,
            )
            return 1
        if missing_archives:
            print(
                "error: draft is missing original archives: "
                + ", ".join(sorted(missing_archives))
                + "; reconcile Homebrew/Scoop metadata and retry without skipping package managers",
                file=sys.stderr,
            )
            return 1

    print(f"Release {tag} is unpublished; a new build or draft recovery may proceed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
