# R013 Code Review — Step 4: Connect section

Verdict: APPROVE

No blocking findings.

What I checked:

- Reviewed the full diff from `718ea2157b1bedccb020094802e0fe58eb2527a3..HEAD`.
- Read the Step 4 connect pages and compared the migrated Claude Desktop / Claude Code content against `docs/clients/claude-desktop.md` and `docs/clients/claude-code.md`.
- Verified the MCP JSON snippets keep the intervals.icu API key out of client config and use only non-secret env values.
- Verified tool references use Hugo `relref` anchors to `reference/tools.md`.
- Ran `cd web && hugo --minify --gc`; the site builds successfully with no broken relrefs.

Non-blocking follow-up for later drift/link sweep:

- Several env vars appear inside JSON/bash examples, which cannot practically be linked inline. The connect index links to the CLI reference, but the per-client pages could optionally add a short “see CLI reference for environment variables” sentence during Step 7 if the task owner wants stricter compliance with the cross-reference guidance.
