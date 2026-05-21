# TP-052-migrate-end-user-docs — Status

**Current Step:** Step 8: Build + link check
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 20
**Iteration:** 1
**Size:** L

---

### Step 1: Inventory + audit

**Status:** ✅ Complete

- [x] Verify blocking/recommended dependency gates before migration: TP-050 scaffold evidence, TP-055 conflict reconciliation, and TP-051 tool-catalog status.
- [x] Confirm TP-055 has landed by verifying the analyzer family, `get_planning_parameters`, and `update_wellness` error-contract conflicts are resolved in PRD, ROADMAP, and README; record the result here before migration.
- [x] Read every source file listed in PROMPT.md and audit factual claims against internal code; record code-vs-prose discrepancies here.
- [x] R003: Correct the TP-050 scaffold dependency gate with accurate evidence by creating/verifying the missing Hugo section structure or recording a blocker before migration proceeds.

**Dependency preflight:**

| Dependency | Required evidence                                                                                                                   | Result                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| TP-050     | Scaffold landed: `web/` structure and Hugo config exist, or explicit explanation for stale task status.                             | Pass after R003 correction: task metadata still says open/no `.DONE`, but the worktree now has `web/hugo.toml`, `web/content/_index.md`, and verified section indexes for `web/content/{install,connect,guides,reference,explain}/_index.md` (`find web/content -maxdepth 2 -name _index.md`).                                                                                                                                                                                                                     |
| TP-055     | `.DONE`/complete status plus conflict-specific grep evidence for analyzer family, `get_planning_parameters`, and `update_wellness`. | Pass: `.DONE` exists and TP-055 `STATUS.md` is complete. Re-verified: PRD keeps analyzers as `Planned analyzers` / v0.6 roadmap scope, ROADMAP v0.6 lists analyzer family as future, README points to generated catalog; `get_planning_parameters` appears only as upstream-gap deferral in ROADMAP and future conditional PRD text, not README/catalog/internal tools; README, PRD, `internal/tools/update_wellness.go`, and `web/data/tools.json` all surface `field_not_writable: sleepScore (device-managed)`. |
| TP-051     | `.DONE`/generated tool catalog status; note whether `reference/tools.md` must be left as a stub.                                    | Note: no TP-051 `.DONE` found in this worktree, but `web/content/reference/tools.md` already contains the generated catalog shortcode and `web/data/tools.json` exists; do not hand-author the tool list.                                                                                                                                                                                                                                                                                                          |

**Audit evidence matrix:**

| Source path                                                                   | Claims audited                                                                                                     | Source of truth checked                                                                                                                                                | Discrepancy found                                                                                                                                                                                                           | Follow-up note                                                                                                                                                                                                                      |
| ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `README.md`                                                                   | Install, API-key/keychain, transport, safety modes, toolset tiers, resources/prompts, schema-change advice         | `internal/config/config.go`, `internal/credstore/credstore.go`, `internal/safety/*`, `internal/resources/*`, `internal/prompts/*`, `internal/app/testdata/help.golden` | No blocking discrepancy for migrated user docs.                                                                                                                                                                             | Use code/help for final env-var/flag lists; README links still point to `docs/` because TP-054 owns slim-down.                                                                                                                      |
| `docs/install/macos.md`                                                       | DMG flow, executable path, setup flags, Keychain service/account, Gatekeeper checks, uninstall                     | CLI help, config/credstore constants, app path conventions                                                                                                             | No blocking discrepancy.                                                                                                                                                                                                    | Strip maintainer release-operator checklist from migrated page.                                                                                                                                                                     |
| `docs/clients/claude-desktop.md`                                              | Config file path, MCP JSON shape, env vars, executable path, smoke checks                                          | CLI help, config env constants, default transport                                                                                                                      | No blocking discrepancy.                                                                                                                                                                                                    | Keep API-key warning; convert smoke checklist to collapsible section.                                                                                                                                                               |
| `docs/clients/claude-code.md`                                                 | `.mcp.json` shape, env vars, executable path, restart/check advice                                                 | CLI help, config env constants, default transport                                                                                                                      | No blocking discrepancy.                                                                                                                                                                                                    | End-user page should avoid committing local athlete IDs.                                                                                                                                                                            |
| `docs/clients/codex-local.md`                                                 | Developer/power-user local validation workflow                                                                     | Current generated catalog / `internal/toolcatalog/catalog.go`                                                                                                          | Stale as end-user truth: it is explicitly v0.1 maintainer validation and expects only `get_athlete_profile`, while current catalog constants include the broader v0.5 surface.                                              | Confirmed developer content; leave in `docs/` and do not migrate.                                                                                                                                                                   |
| `docs/coach-mode.md`                                                          | `ICUVISOR_COACH_MODE`, coach config schema, ACL patterns, `list_athletes`/`select_athlete`, cache caveat           | `internal/coach/config.go`, `internal/toolcatalog/catalog.go`, `internal/mcp/protocol_test.go` grep evidence                                                           | No blocking discrepancy.                                                                                                                                                                                                    | Split setup/how-to into guides and conceptual model into explanation.                                                                                                                                                               |
| `docs/post-update.md`                                                         | `_meta.server_version`, `_meta.catalog_hash`, `_meta.schema_changed` and new-conversation advice                   | `internal/response/meta.go`, `internal/mcp/catalog_hash.go` grep evidence                                                                                              | No blocking discrepancy.                                                                                                                                                                                                    | Use plain-language explanation in migrated guide.                                                                                                                                                                                   |
| `docs/prd/PRD-icuvisor.md` §7.2.A/D/E/F/G                                     | Distribution/transport, response shaping, toolset tiers, destructive safety, resources/prompts, config assumptions | `internal/config/config.go`, `internal/safety/*`, `internal/mcp/registrar_tools.go`, `internal/resources/*`, `internal/prompts/*`                                      | Discrepancy: PRD §7.2.F still says invalid `ICUVISOR_DELETE_MODE` values fail loudly and its table suggests some safe-mode deletes; code parses empty/unknown to `safe` and safe mode registers writes but no delete tools. | Code wins for website safety-mode reference/explanation.                                                                                                                                                                            |
| `internal/config/`                                                            | Env-var names, defaults, config fields, validation, key precedence, coach/debug flags                              | Code source of truth                                                                                                                                                   | N/A                                                                                                                                                                                                                         | Actual file is `internal/config/config.go`; default config path is OS user config dir `icuvisor/config.json`; default HTTP bind `127.0.0.1:8765`; config fields include `api_key` legacy plus public non-secret fields and `coach`. |
| `internal/safety/`                                                            | Delete/write mode and toolset tier values/defaults                                                                 | Code source of truth                                                                                                                                                   | N/A                                                                                                                                                                                                                         | `ICUVISOR_DELETE_MODE`: `safe/full/none`, unknown or empty => `safe`; `ICUVISOR_TOOLSET`: `core/full`, unknown or empty => `core`.                                                                                                  |
| `internal/toolset/`                                                           | Prompt-listed source path                                                                                          | Repository audit                                                                                                                                                       | Path does not exist.                                                                                                                                                                                                        | Toolset implementation is `internal/safety/toolset.go`; use that as source of truth.                                                                                                                                                |
| `internal/prompts/`                                                           | Prompt names, titles, arguments, descriptions, default resources/tools                                             | Code source of truth                                                                                                                                                   | N/A                                                                                                                                                                                                                         | Five prompts registered: `training_analysis`, `recovery_check`, `weekly_planning`, `race_week_taper`, `coach_roster_triage`.                                                                                                        |
| CLI help (`internal/app/testdata/help_golden.txt` or `./bin/icuvisor --help`) | Flags, env vars, commands, exit codes                                                                              | `internal/app/testdata/help.golden`                                                                                                                                    | Prompt filename drift: actual golden fixture is `internal/app/testdata/help.golden`, not `help_golden.txt`.                                                                                                                 | `reference/cli.md` should render the actual fixture verbatim.                                                                                                                                                                       |

**Notes:**

- Code-vs-prose discrepancies: PRD §7.2.F safety-mode invalid-value/delete-table language differs from current code; `docs/clients/codex-local.md` is intentionally stale/developer-only; prompt referenced `internal/toolset/` and `internal/app/testdata/help_golden.txt`, but actual sources are `internal/safety/toolset.go` and `internal/app/testdata/help.golden`.
- TP-055 dependency status: verified complete and conflict reconciliations match code/catalog truth as of this audit.
- Plan review R001 requested explicit dependency preflight evidence before proceeding.
- R003 code review found the TP-050 scaffold gate was incorrectly recorded as pass: only `web/content/_index.md` and `web/content/reference/` existed. Step 1 was reopened and the missing `install`, `connect`, `guides`, and `explain` section indexes were created and verified.

### Step 2: Reference section

**Status:** ✅ Complete

- [x] Author `web/content/reference/cli.md` from `internal/app/testdata/help.golden` with framing prose and verbatim full help output.
- [x] Author `web/content/reference/safety-modes.md` for `ICUVISOR_DELETE_MODE` and `ICUVISOR_TOOLSET`, registration effects, and `_meta.delete_mode` / `_meta.toolset` echo behavior.
- [x] R004: Reconcile existing `web/content/reference/toolset-tiers.md` by folding useful content into `reference/safety-modes.md` and replacing/removing duplicate website content.
- [x] Author `web/content/reference/config-file.md` as a JSON config field reference grounded in `internal/config/` and coach config code.
- [x] Author `web/content/reference/resources-prompts.md` for registered MCP resources and prompts using `internal/resources/` and `internal/prompts/` as source of truth.
- [x] Verify `web/content/reference/tools.md` is generated/stubbed by TP-051 and do not hand-author the tool list.
- [x] R007: Update generated tool catalog tier-badge links so `full` badges point at the new canonical `reference/safety-modes.md` page instead of deleted `reference/toolset-tiers.md`.
- [x] R008: Correct the coach config example so per-athlete `allowed_tools` contains only athlete-scoped ACL patterns, not `icuvisor_list_advanced_capabilities`.

**Notes:**

- Step 1 found TP-051 `.DONE` absent but `web/content/reference/tools.md` plus `web/data/tools.json` already present; preserved generated shortcode page and verified `web/data/tools.json` has 40 entries.
- Plan review R004 requested a single canonical safety/toolset reference page instead of leaving duplicate `reference/toolset-tiers.md` content.
- Code review R007 found `web/layouts/partials/tool-catalog.html` still links full-tier badges to deleted `/reference/toolset-tiers/`; fixed to `/reference/safety-modes/#toolset-tier` and verified `hugo --minify --gc` plus grep.
- Code review R008 found `icuvisor_list_advanced_capabilities` is not an athlete-scoped ACL pattern and must not appear in `coach.athletes[].allowed_tools` examples; removed it from the JSON example and added linked prose explaining it is a meta/control tool.

### Step 3: Install section

**Status:** ✅ Complete

- [x] Author `web/content/install/_index.md` as an install overview chooser for macOS, Windows, Linux, and build-from-source.
- [x] Author `web/content/install/macos.md` from `docs/install/macos.md`, keeping Gatekeeper verification, DMG flow, uninstall, setup, and MCP executable path while dropping maintainer release-signing checklist.
- [x] Author `web/content/install/windows.md` as a v1.0 placeholder with build-from-source GitHub link.
- [x] Author `web/content/install/linux.md` as a v1.0 placeholder with build-from-source GitHub link.

**Notes:**

- Step 1 R003 created placeholder section indexes; replace the install index with user-facing content.

### Step 4: Connect section

**Status:** ✅ Complete

- [x] Author `web/content/connect/_index.md` chooser for Claude Desktop, Claude Code, ChatGPT, Cursor, Continue, Zed, and other MCP clients.
- [x] Author `web/content/connect/claude-desktop.md` from `docs/clients/claude-desktop.md`, with end-user voice and smoke checklist as a collapsible details block.
- [x] Author `web/content/connect/claude-code.md` from `docs/clients/claude-code.md`, with end-user voice.
- [x] Author `web/content/connect/chatgpt.md` as a minimal focused MCP connection how-to.
- [x] Author `web/content/connect/other-clients.md` for Cursor, Continue, Zed, Pi, and other MCP clients using the same JSON shape.

**Notes:**

- Do not migrate `docs/clients/codex-local.md`; Step 1 confirmed it is developer/power-user content.

### Step 5: Guides section

**Status:** ✅ Complete

- [x] Update `web/content/guides/_index.md` as a chooser for the guide pages.
- [x] Author `web/content/guides/api-key.md` for getting an intervals.icu API key, `icuvisor setup`, and manual keychain commands per OS.
- [x] Author `web/content/guides/http-transport.md` for enabling Streamable HTTP, loopback default, and explicit LAN warning.
- [x] Author `web/content/guides/coach-mode.md` with user-facing setup steps from `docs/coach-mode.md`.
- [x] Author `web/content/guides/after-upgrade.md` from `docs/post-update.md`, explaining `_meta.schema_changed` and new-conversation advice.
- [x] Author `web/content/guides/troubleshooting.md` with symptom-to-fix table for Gatekeeper, missing API key, stale schema, LAN refusal, and Linux libsecret.

**Notes:**

- Link guide references to existing install/connect/reference pages where possible; explanation pages are authored in Step 6.

### Step 6: Explanation section

**Status:** ✅ Complete

- [x] Update `web/content/explain/_index.md` as a chooser for explanation pages.
- [x] Author `web/content/explain/what-is-mcp.md`.
- [x] Author `web/content/explain/local-first.md`.
- [x] Author `web/content/explain/terse-by-default.md`.
- [x] Author `web/content/explain/safety-modes.md` explaining why `safe`/`full`/`none` exist.
- [x] Author `web/content/explain/coach-mode.md` covering server-held key, `athlete_id` selector semantics, and per-athlete ACLs.

**Notes:**

- Pair conceptual pages with existing reference/guide pages; avoid duplicating tables already in reference.

### Step 7: Reconcile with code (drift sweep)

**Status:** ✅ Complete

- [x] Grep website pages for every env-var, flag, and tool name and cross-check against `internal/` source truth.
- [x] Grep website pages for every JSON shape and confirm against `internal/` types and examples.
- [x] Fix any documentation discrepancy discovered during the drift sweep; code wins and Go code remains unchanged.
- [x] R019: Soften `what-is-mcp.md` language so MCP clients may start icuvisor over stdio or connect to an already running HTTP server.

**Notes:**

- Include links/anchors and generated catalog partials in the drift sweep because previous review found a stale generated-link target.
- Env/flag/tool sweep evidence: website env vars exactly match `internal/config/config.go` + `internal/safety/{mode,toolset}.go`; icuvisor flags mentioned are valid (`--config`, `--env-file`, `--transport`, `--http-bind`, setup `--offline`/`--force`, help) and `--api-key` is correctly documented as intentionally absent; other `--deep`/`--strict`/`--verify`/`--ignore-missing`/`--label` occurrences belong to external macOS/libsecret commands. Tool names mentioned are present in `web/data/tools.json`/`internal/toolcatalog`.
- JSON sweep evidence: all `json` fenced blocks parse as valid JSON; config examples use `athlete_id`, `timezone`, `api_base_url`, `http_timeout`, `transport`, `http_bind`, and `coach` fields matching `internal/config/config.go` / `internal/coach/config.go`; MCP client examples match existing source docs' `mcpServers.icuvisor.command/env` shape; `_meta` examples match `internal/response/meta.go` keys.
- Drift fixes applied: replaced temporary prose now that Step 5/6 pages exist (`connect/_index.md` now links to HTTP transport guide; `reference/config-file.md` links to coach guide/explanation; `install/_index.md` links to connect section). Verified `cd web && hugo --minify --gc` after fixes.
- R019 found `explain/what-is-mcp.md` still described only stdio-style client startup; fixed to say clients start icuvisor over stdio or connect to an already running Streamable HTTP server. Verified Hugo build.

### Step 8: Build + link check

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` `[Unreleased]` under Changed to mention user-facing docs migrated to icuvisor.app.
- [x] Run `cd web && hugo --minify --gc` with zero broken relrefs.
- [x] Verify Pagefind/search indexing includes new pages by running a search for a tool name and an environment variable, or record the available equivalent if the scaffold lacks Pagefind tooling.
- [x] Manually click through every nav entry, or perform an equivalent generated-site link sweep when browser navigation is unavailable.

**Notes:**

- Review protocol skips the final step, but all verification evidence must be recorded here.
- Hugo build evidence: `rm -rf web/public && cd web && hugo --minify --gc` succeeded with 29 pages and zero broken relrefs.
- Search/Pagefind evidence: scaffold currently has no Pagefind files or search code (`find web -iname '*pagefind*'` and grep returned none). Equivalent generated-site token checks found `get_athlete_profile` and `ICUVISOR_DELETE_MODE` in `web/public` HTML.
- Link/navigation evidence: browser was unavailable, so a generated-site HTML link sweep parsed `web/public/**/*.html`, ignored external links, and verified 249 internal links resolve to an existing generated file or in-page anchor target class; no missing internal links found.

| 2026-05-17 17:08 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 17:08 | Step 1 started | Inventory + audit |
| 2026-05-17 17:11 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-17 17:13 | Review R002 | plan Step 1: APPROVE |
| 2026-05-17 17:19 | Review R003 | code Step 1: APPROVE |
| 2026-05-17 17:21 | Review R004 | plan Step 2: UNKNOWN |
| 2026-05-17 17:25 | Review R005 | code Step 1: APPROVE |
| 2026-05-17 17:28 | Review R006 | plan Step 2: APPROVE |
| 2026-05-17 17:38 | Review R007 | code Step 2: UNKNOWN |
| 2026-05-17 17:43 | Review R008 | code Step 2: UNKNOWN |
| 2026-05-17 17:49 | Review R009 | code Step 2: APPROVE |
| 2026-05-17 17:51 | Review R010 | plan Step 3: APPROVE |
| 2026-05-17 17:57 | Review R011 | code Step 3: APPROVE |
| 2026-05-17 18:00 | Review R012 | plan Step 4: APPROVE |
| 2026-05-17 18:06 | Review R013 | code Step 4: APPROVE |
| 2026-05-17 18:08 | Review R014 | plan Step 5: APPROVE |
| 2026-05-17 18:17 | Review R015 | code Step 5: APPROVE |
| 2026-05-17 18:20 | Review R016 | plan Step 6: APPROVE |
| 2026-05-17 18:26 | Review R017 | code Step 6: APPROVE |
| 2026-05-17 18:30 | Review R018 | plan Step 7: APPROVE |
| 2026-05-17 18:37 | Review R019 | code Step 7: UNKNOWN |
| 2026-05-17 18:40 | Review R020 | code Step 7: APPROVE |

| 2026-05-17 18:45 | Worker iter 1 | done in 5840s, tools: 301 |
| 2026-05-17 18:45 | Task complete | .DONE created |