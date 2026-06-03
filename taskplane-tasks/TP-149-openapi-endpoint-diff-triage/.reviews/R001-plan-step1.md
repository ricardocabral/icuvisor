# Plan Review: Step 1 — Design endpoint-diff triage workflow

Verdict: **Changes requested before implementation**

The proposed workflow is directionally correct: offline fixture-based diffing, opt-in/live fetch only via explicit flag or scheduled/manual CI, Markdown triage output, and no auto-generation of tools all match the task requirements and clean-room constraints.

One design issue should be resolved before Step 2:

- **Testability of a root-level standalone script:** the plan names `scripts/diff_openapi_endpoints.go`. Existing root `scripts/*.go` files use `//go:build ignore` and multiple `main` packages, so logic placed only in a root build-ignored script will not be covered by `go test ./...` and is awkward to unit test with fixtures. Step 2 explicitly requires fixture-based tests. Prefer a testable layout such as `scripts/openapidiff/` (normal Go package/command with `_test.go`) or a small build-ignored wrapper plus testable diff logic in a package. This keeps normal tests offline while making added/removed/no-change detection part of the suite.

Implementation notes once that is addressed:

- Keep live fetching out of normal tests; use local fixture specs for unit tests.
- Ensure the scheduled/manual workflow writes a summary/artifact and does not imply endpoints are automatically accepted into product scope.
- Document where the pinned baseline lives and how maintainers intentionally update it after triage.
