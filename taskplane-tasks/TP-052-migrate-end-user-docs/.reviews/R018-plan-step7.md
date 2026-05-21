# R018 Plan Review — Step 7: Reconcile content with code (drift sweep)

Verdict: APPROVE

The Step 7 plan matches the prompt: it is a docs-only drift sweep, explicitly compares env vars, flags, tool names, and JSON shapes against `internal/`, and keeps the rule that code wins without changing Go code. The additional note to include links/anchors and generated catalog partials is important and addresses the stale generated-link issue found earlier.

Non-blocking implementation notes:

1. Build the comparison from authoritative inventories, not only ad hoc prose greps. Use `internal/app/testdata/help.golden` plus command-specific help/code for CLI flags, `internal/config/config.go` and `internal/safety/*` for env vars/defaults, `internal/toolcatalog/catalog.go` or `web/data/tools.json` for tool names, and `internal/resources/` / `internal/prompts/` for resources and prompts.
2. Include both `ICUVISOR_*` and `INTERVALS_ICU_*` env vars. A grep limited to `ICUVISOR_` would miss the required API-key and athlete-ID variables.
3. Classify flag hits before treating them as icuvisor flags. The website legitimately contains external command flags such as `codesign --verify --deep --strict`, `shasum --ignore-missing`, and `secret-tool --label`, while icuvisor itself has top-level flags plus `setup`-specific flags (`--config`, `--offline`, `--force`, `--help`) and intentionally no `--api-key` flag.
4. For tool-name mentions, verify both spelling and link target. Mentions should point to generated anchors under `reference/tools.md` where they are actual tools; ACL patterns such as `get_*` and metadata fields such as `allowed_tools` should be handled separately so false positives do not obscure real drift.
5. For JSON shapes, do more than check that code blocks parse. Compare examples against the relevant structs/contracts: MCP client `mcpServers` snippets, config file fields in `internal/config` and `internal/coach`, response `_meta` fields in `internal/response`, and safety/toolset values in `internal/safety`. Partial snippets are fine if the surrounding prose makes that clear.
6. Carry forward the Step 6 follow-up: soften `explain/what-is-mcp.md` from “the AI client starts the local icuvisor server” to “starts or connects to” so the wording also covers Streamable HTTP.
7. Record any drift discovered and fixed in `STATUS.md`, including the source-of-truth file used. Do not edit `README.md`, delete migrated `docs/` sources, or change Go code in this task.

With those guardrails, the plan is ready for the drift sweep.
