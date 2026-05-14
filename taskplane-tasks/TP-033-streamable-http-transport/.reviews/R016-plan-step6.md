# Plan Review: Step 6 — Verify

**Verdict:** REVISE

I read `PROMPT.md` and `STATUS.md`. The plan covers the required automated commands at a high level, but the manual verification part is not specific enough to satisfy the Step 6 checklist.

## Blocking gaps

1. **`tools/list` is not a tool call.**
   The checklist requires driving a tool call over HTTP from an MCP client. Keep `tools/list` as a useful smoke check, but add an actual `tools/call` invocation. Prefer a no-network tool such as `icuvisor_list_advanced_capabilities` so the manual test does not require real intervals.icu credentials or mutate data.

2. **Define how the loopback-only bind will be confirmed.**
   “Launch HTTP mode on the default loopback bind” does not prove it binds only to `127.0.0.1`. The plan should name the check: for example, inspect the startup log endpoint, use `lsof`/`netstat`/`ss` to verify the listener is `127.0.0.1:8765` and not `0.0.0.0`/`::`, and avoid passing `--http-bind`/`ICUVISOR_HTTP_BIND` so the default path is exercised.

3. **Make the manual client and config shape explicit.**
   Say which MCP client/script will connect to `http://127.0.0.1:8765/mcp`, and use dummy/sentinel local config values that do not leak secrets. If a tool call is chosen that makes no intervals.icu API calls, document that choice in the verification notes.

The automated portion can remain concise as: `make test`, `make build`, `make lint`, and `go test -race ./...`. After the manual plan includes both a listener-address check and a real HTTP `tools/call`, Step 6 is ready to execute.
