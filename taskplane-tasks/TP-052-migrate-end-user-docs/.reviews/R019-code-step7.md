# R019 Code Review — Step 7: Reconcile content with code (drift sweep)

Verdict: UNKNOWN

I reviewed the Step 7 diff, read the changed docs/status files, spot-checked the drift-sweep claims against `internal/`, and ran `cd web && hugo --minify --gc` successfully.

## Findings

1. **HTTP transport drift remains in the MCP explanation** — `web/content/explain/what-is-mcp.md:6`

   The page still says: “the AI client starts the local icuvisor server”. That is true for stdio client configurations, but not for Streamable HTTP: the server can be started separately and the client connects to `http://127.0.0.1:8765/mcp` (`internal/mcp/transport.go` defines the HTTP endpoint, and the new HTTP guide documents that flow). This was also explicitly called out in R018 as a Step 7 follow-up: soften the sentence to “starts or connects to” so the explanation covers both supported transports.

   Please update the sentence, and add the fix to the Step 7 notes in `STATUS.md` so the drift-sweep record matches what was actually reconciled.

## Notes

- The newly added `relref` links in `connect/_index.md`, `install/_index.md`, and `reference/config-file.md` build cleanly.
- The env-var inventory in the website matches the current `ICUVISOR_*` / `INTERVALS_ICU_*` names from `internal/config`, `internal/safety`, and the help golden fixture.
