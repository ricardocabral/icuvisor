# R017 Code Review — Step 6: Explanation section

Verdict: APPROVE

No blocking findings.

What I checked:

- Ran the requested changed-file and full-diff commands from `532dfb40b1ed1205047cbaa5a942dc52ab0c4b59..HEAD`.
- Read the Step 6 explanation pages under `web/content/explain/`.
- Checked that the explanation index links to the five required conceptual pages.
- Verified the safety-mode explanation uses the code-audited behavior from Step 1: `safe` as the default, no model-controlled `confirm: true`, and delete-capable workflows only in `full` with coach ACLs still composing.
- Cross-checked coach-mode claims against `internal/coach/`, `internal/mcp/registrar_tools.go`, and the generated tool catalog anchors for `list_athletes`, `select_athlete`, and `icuvisor_list_advanced_capabilities`.
- Verified `include_full` / terse-by-default prose matches the response-shaping pattern in `internal/tools/` and links to the canonical safety/toolset reference.
- Ran `cd web && hugo --minify --gc`; the site builds successfully with no broken relrefs.

Non-blocking note for the later drift/link sweep:

- `what-is-mcp.md` says the AI client starts the local icuvisor server, which is accurate for stdio-style client configs but not for the HTTP transport path where the user starts icuvisor and points the client at `127.0.0.1:8765/mcp`. The HTTP/connect pages already explain that distinction, so this is not blocking Step 6, but Step 7 could soften the sentence to “starts or connects to” for precision.
