# Plan Review: Step 5 — Docs

## Verdict

APPROVE

## Findings

No blocking findings.

## Notes

- The plan covers the Step 5 checklist: README transport-selection documentation and a concise `[Unreleased]` CHANGELOG entry.
- README coverage should explicitly include that `stdio` remains the default, Streamable HTTP is opt-in via env/CLI/config, and the default endpoint is `http://127.0.0.1:8765/mcp`.
- Keep the LAN-bind warning direct: HTTP mode is unauthenticated in this task, so anyone who can reach the bind address can invoke registered tools using the configured intervals.icu credentials.
- Include both env/CLI and JSON config names (`ICUVISOR_TRANSPORT`, `ICUVISOR_HTTP_BIND`, `--transport`, `--http-bind`, `transport`, `http_bind`) so users can map the docs to the implemented config surface.
- The CHANGELOG entry should stay user-facing and avoid over-documenting internal test/lifecycle details that belong in `STATUS.md`.
