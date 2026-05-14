# Code Review: Step 5 — Docs

**Verdict:** APPROVE

I reviewed the Step 5 working-tree documentation changes in `README.md`, `CHANGELOG.md`, and `STATUS.md`. Note that `git diff e775787..HEAD` is empty because `HEAD` is still at the baseline commit; the Step 5 changes are currently uncommitted working-tree changes, so I reviewed `git diff e775787`.

The README now documents:

- `stdio` as the default transport.
- Streamable HTTP opt-in via `ICUVISOR_TRANSPORT=http` or `--transport http`.
- The default loopback-only endpoint `127.0.0.1:8765` at `/mcp`.
- CLI and config-file fields for transport/bind selection.
- Startup failure for invalid transport/bind values.
- The unauthenticated LAN-bind threat model and deliberate opt-in requirement.

The `[Unreleased]` changelog entry concisely captures the new Streamable HTTP transport, default loopback bind, env/CLI overrides, and LAN-bind warning behavior.

No blocking issues found.
