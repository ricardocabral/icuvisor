# Plan Review: Step 2 — Implement OpenAPI diff tooling

Verdict: **APPROVE**

The revised Step 2 plan addresses the prior testability concern by using a normal `scripts/openapidiff/` package/command with fixture-based tests rather than burying logic in a build-ignored root script. It also satisfies the task constraints: normal tests remain offline, live fetching is opt-in or CI-scheduled/manual, output is a human triage artifact, and the tooling does not auto-generate or auto-implement endpoints.

Implementation watchpoints:

- Keep the pinned baseline in a path covered by task scope, and document exactly when/how maintainers intentionally update it after triage.
- Sort path keys in output/tests so added/removed/no-change results are deterministic.
- If parsing YAML is needed, verify any new dependency license first; prefer JSON/stdlib if the upstream spec URL supports it.
- Make the workflow produce a job summary and/or artifact without failing routine CI solely because upstream added endpoints, unless that behavior is intentionally documented.
- Ensure fixture tests cover added paths, removed paths, and no-change output without network access.

Proceed with Step 2 implementation.
