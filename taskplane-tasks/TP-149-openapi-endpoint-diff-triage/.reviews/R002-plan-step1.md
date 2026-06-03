# Plan Review: Step 1 — Design endpoint-diff triage workflow

Verdict: **Approved for implementation**

The revised plan addresses the prior review concern by moving the diff logic into a testable `scripts/openapidiff/` Go package/command instead of relying only on a build-ignored root script. It satisfies the Step 1 requirements:

- normal tests remain offline via local `-baseline`/`-latest` fixture inputs;
- live fetching is opt-in (`-latest-url`) and limited to scheduled/manual CI use;
- output is a Markdown human-triage summary focused on added/removed OpenAPI path keys;
- the workflow explicitly avoids auto-generating tools or accepting new endpoints into product scope.

Implementation guardrails for Step 2:

- Keep the pinned baseline in an in-scope OpenAPI path and document how maintainers intentionally refresh it after triage.
- Add fixture-based tests for added paths, removed paths, and no-change output.
- Ensure any scheduled/manual workflow produces a step summary and/or artifact without making normal CI/network tests flaky.
