#!/usr/bin/env python3
"""Exercise the release preflight against GitHub CLI response fixtures."""

import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


SCRIPT = Path(__file__).resolve().parents[1] / "check_release_unpublished.py"


class ReleasePreflightTests(unittest.TestCase):
    def test_non_tag_inputs_are_rejected(self):
        for tag in ("refs/tags/v1.9.0", "main", "069f893e", "v1.9.0\n", ""):
            with self.subTest(tag=tag):
                result = self.run_guard("[[]]", tag=tag)
                self.assertEqual(result.returncode, 2, result.stderr)
                self.assertIn("release tag", result.stderr)

    def test_prerelease_and_build_metadata_tags(self):
        for tag in ("v1.9.0-rc.1", "v1.9.0+build.1", "v1.9.0-rc.1+build.1"):
            with self.subTest(tag=tag):
                result = self.run_guard("[[]]", tag=tag)
                self.assertEqual(result.returncode, 0, result.stderr)

    def test_package_manager_recovery(self):
        archives = [
            {"name": name, "state": "uploaded", "size": 1234}
            for name in (
                "icuvisor_1.9.0_linux_amd64.tar.gz",
                "icuvisor_1.9.0_linux_arm64.tar.gz",
                "icuvisor_1.9.0_windows_amd64.zip",
                "icuvisor_1.9.0_windows_arm64.zip",
                "icuvisor_1.9.0_macos_universal.tar.gz",
            )
        ]
        draft = {"tag_name": "v1.9.0", "draft": True, "assets": archives}
        cases = (
            ("complete draft", [[draft]], 0, "unpublished"),
            ("no draft", [[]], 1, "exactly one existing draft"),
            ("duplicate drafts", [[draft, draft]], 1, "exactly one existing draft"),
            (
                "incomplete original archives",
                [[dict(draft, assets=archives[:-1])]],
                1,
                "missing original archives",
            ),
            (
                "wrong version archives",
                [[dict(draft, assets=[{"name": "icuvisor_1.8.2_linux_amd64.tar.gz"}])]],
                1,
                "missing original archives",
            ),
            (
                "unfinished upload",
                [[dict(draft, assets=[dict(archives[0], state="starter")] + archives[1:])]],
                1,
                "missing original archives",
            ),
            (
                "empty archive",
                [[dict(draft, assets=[dict(archives[0], size=0)] + archives[1:])]],
                1,
                "missing original archives",
            ),
            (
                "invalid archive size",
                [[dict(draft, assets=[dict(archives[0], size="1234")] + archives[1:])]],
                1,
                "missing original archives",
            ),
            (
                "missing upload metadata",
                [[dict(draft, assets=[{"name": archives[0]["name"]}] + archives[1:])]],
                1,
                "missing original archives",
            ),
            (
                "published release",
                [[dict(draft, draft=False)]],
                1,
                "already published",
            ),
        )
        for name, pages, expected, message in cases:
            with self.subTest(name=name):
                result = self.run_guard(json.dumps(pages), skip_package_managers=True)
                self.assertEqual(result.returncode, expected, result.stderr)
                self.assertIn(message, result.stdout + result.stderr)

    def test_release_state(self):
        cases = (
            ("no releases", [[]], 0),
            ("draft recovery", [[{"tag_name": "v1.9.0", "draft": True}]], 0),
            ("other version", [[{"tag_name": "v1.8.2", "draft": False}]], 0),
            ("published", [[{"tag_name": "v1.9.0", "draft": False}]], 1),
            (
                "published on later page despite draft",
                [
                    [{"tag_name": "v1.9.0", "draft": True}],
                    [{"tag_name": "v1.9.0", "draft": False}],
                ],
                1,
            ),
        )
        for name, pages, expected in cases:
            with self.subTest(name=name):
                result = self.run_guard(json.dumps(pages))
                self.assertEqual(result.returncode, expected, result.stderr)
                if expected:
                    self.assertIn("already published", result.stderr)
                else:
                    self.assertIn("unpublished", result.stdout)

    def test_lookup_errors_block_release(self):
        cases = (
            ("authentication failure", "", 1),
            ("partial pagination failure", "[[]]", 1),
            ("invalid JSON", "not JSON", 0),
            ("missing API pages", "[]", 0),
            ("unexpected response", '{"message":"API error"}', 0),
            ("missing release state", '[[{"tag_name":"v1.9.0"}]]', 0),
            ("invalid release state", '[[{"tag_name":"v1.9.0","draft":null}]]', 0),
        )
        for name, response, exit_code in cases:
            with self.subTest(name=name):
                result = self.run_guard(response, exit_code)
                self.assertNotEqual(result.returncode, 0)
                self.assertIn("cannot verify", result.stderr)
                self.assertNotIn("unpublished", result.stdout)

    def test_missing_arguments(self):
        result = subprocess.run(
            [sys.executable, str(SCRIPT)], capture_output=True, text=True, check=False
        )
        self.assertEqual(result.returncode, 2)
        self.assertIn("usage:", result.stderr)

    def run_guard(self, response, exit_code=0, skip_package_managers=False, tag="v1.9.0"):
        with tempfile.TemporaryDirectory() as directory:
            gh = Path(directory) / "gh"
            gh.write_text(
                f"#!{sys.executable}\n"
                "import os, sys\n"
                "expected = ['api', '--paginate', '--slurp', "
                "'repos/ricardocabral/icuvisor/releases']\n"
                "if sys.argv[1:] != expected: sys.exit(99)\n"
                "print(os.environ['TEST_GH_RESPONSE'])\n"
                "sys.exit(int(os.environ['TEST_GH_EXIT_CODE']))\n",
                encoding="utf-8",
            )
            gh.chmod(0o755)
            env = dict(
                os.environ,
                TEST_GH_RESPONSE=response,
                TEST_GH_EXIT_CODE=str(exit_code),
            )
            env["PATH"] = directory + os.pathsep + env.get("PATH", "")
            command = [sys.executable, str(SCRIPT), "ricardocabral/icuvisor", tag]
            if skip_package_managers:
                command.append("--skip-package-managers")
            return subprocess.run(
                command,
                env=env,
                capture_output=True,
                text=True,
                check=False,
            )


if __name__ == "__main__":
    unittest.main()
