# Plan Review: Step 5 — Docs

**Verdict:** APPROVE

I read `PROMPT.md`, `STATUS.md`, and the current README/CHANGELOG structure. The Step 5 plan is appropriately scoped for a documentation-only step: it covers README transport selection docs, explicit env/flag/config names, the loopback default endpoint, the opt-in LAN bind threat note, and a concise `[Unreleased]` changelog entry.

Implementation notes:

- In README, make sure the config-file field names are included alongside `ICUVISOR_TRANSPORT`/`--transport` and `ICUVISOR_HTTP_BIND`/`--http-bind` so JSON users can discover `transport` and `http_bind` too.
- Keep the default posture unambiguous: `stdio` remains default; `http` is opt-in; HTTP without a bind override listens on `127.0.0.1:8765` at `/mcp`.
- Preserve the required security wording: a LAN bind exposes an unauthenticated MCP server reachable by anyone who can connect to that address, and those callers can invoke registered tools using the configured intervals.icu credentials.
- Put the changelog bullet under `[Unreleased]` / `Added`; no extra PRD or roadmap edits are needed for this step.

No blocking gaps found.
