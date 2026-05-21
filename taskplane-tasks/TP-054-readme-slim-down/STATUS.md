# TP-054-readme-slim-down — Status

**Current Step:** Step 5: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 2
**Review Counter:** 11
**Iteration:** 1
**Size:** S

---

### Step 1: Pre-flight verification

**Status:** ✅ Complete

- [x] Confirm every section being deleted has a complete website destination. For each deletion, paste the destination URL into `STATUS.md` Step 1 with local Hugo `web/public` evidence approved by supervisor for pre-launch while live DNS/hosting is unavailable.
- [x] Confirm TP-051 has the tool-catalog generator + website page live. Evidence: `taskplane-tasks/TP-051-tool-catalog-generator/.DONE` exists; git log includes `4b862e3 merge: wave 2 lane 1 — TP-051, TP-076`, `0feb614 checkpoint: TP-051 task artifacts (.DONE, STATUS.md)`, and TP-051 implementation/review commits; `internal/toolcatalog/` exists; `web/data/tools.json` contains 40 generated tool entries; local Hugo output `web/public/reference/tools/index.html` includes `Tool reference` and generated rows such as `get_activities`.
- [x] Confirm TP-052 has migrated every end-user `docs/*.md` page. Evidence: `taskplane-tasks/TP-052-migrate-end-user-docs/.DONE` exists; git log includes `ada097a merge: wave 3 lane 1 — TP-052` and `c984b25 checkpoint: TP-052 task artifacts (.DONE, STATUS.md)`; source pages exist under `web/content/install/macos.md`, `web/content/connect/claude-desktop.md`, `web/content/connect/claude-code.md`, `web/content/guides/coach-mode.md`, `web/content/explain/coach-mode.md`, and `web/content/guides/after-upgrade.md`; Hugo output contains the corresponding `web/public/**/index.html` pages with expected headings/content.
- [x] Confirm TP-053 has shipped the ChatGPT tutorial. Evidence: `taskplane-tasks/TP-053-getting-started-tutorial-chatgpt/.DONE` exists; git log includes `6749ce5 merge: wave 4 lane 1 — TP-053`, `5661561 checkpoint: TP-053 task artifacts (.DONE, STATUS.md)`, `707b822 TP-053 draft chatgpt getting-started tutorial`, and `132d1e9 TP-053 add chatgpt tutorial screenshots`; `web/content/tutorials/getting-started-chatgpt.md` and `web/content/connect/chatgpt.md` exist; Hugo output includes `web/public/tutorials/getting-started-chatgpt/index.html`, `web/public/connect/chatgpt/index.html`, and six `web/public/img/tutorials/chatgpt/*.png` screenshots.
- [x] Confirm TP-055 has landed. Evidence: `taskplane-tasks/TP-055-reconcile-doc-conflicts/.DONE` exists; `taskplane-tasks/TP-055-reconcile-doc-conflicts/STATUS.md` top-level status is `✅ Complete`; git log includes `a1f0bcc merge: wave 1 lane 1 — TP-055`, `71ba348 checkpoint: TP-055 task artifacts (.DONE, STATUS.md)`, and `4a335f9 TP-055 verify documentation reconciliation`; TP-055 STATUS records resolution of analyzer, `get_planning_parameters`, and `update_wellness` documentation conflicts.

#### Website destination verification

| Deleted README/docs content                                         | Website destination URL                                                                                  | Evidence                                                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| README `Features (planned for v1.0)`                                | https://icuvisor.app/                                                                                    | `cd web && hugo --minify --gc` passed; `web/public/index.html` title `icuvisor — Talk to your intervals.icu data`, h1 `Talk to your intervals.icu data.`, includes feature bullets for generated MCP catalog, coach mode, stdio/HTTP transports, signed binary, keychain/local privacy.                                                      |
| README `MCP tool catalog`                                           | https://icuvisor.app/reference/tools/                                                                    | `web/public/reference/tools/index.html` title `Tool reference · icuvisor`, h1 `Tool reference`, generated registry table includes tool rows such as `get_activities`, `get_activity_details`, `list_athletes`, and grouped domains Activities/Coach mode/Custom items/Events/Fitness/Meta/Settings/Wellness/Workout library.                 |
| README `MCP resources`                                              | https://icuvisor.app/reference/resources-prompts/#resources                                              | `web/public/reference/resources-prompts/index.html` title `MCP resources and prompts · icuvisor`, h1 `MCP resources and prompts`, includes resource table with `icuvisor://workout-syntax`, `icuvisor://event-categories`, `icuvisor://custom-item-schemas`, and `icuvisor://athlete-profile`.                                               |
| README `MCP prompts`                                                | https://icuvisor.app/reference/resources-prompts/#prompts                                                | Same local page includes prompts table with `training_analysis`, `recovery_check`, `weekly_planning`, `race_week_taper`, and `coach_roster_triage`, plus prompt guardrails.                                                                                                                                                                  |
| README macOS install/DMG prose and `docs/install/macos.md`          | https://icuvisor.app/install/macos/                                                                      | `web/public/install/macos/index.html` title/h1 `Install on macOS`, includes signed/notarized DMG install steps, `/Applications/icuvisor.app/Contents/MacOS/icuvisor`, Keychain setup, Gatekeeper/notarization checks, and MCP client configuration.                                                                                          |
| README quickstart                                                   | https://icuvisor.app/install/ and https://icuvisor.app/guides/api-key/                                   | `web/public/install/index.html` title/h1 `Install icuvisor` covers platform install paths; `web/public/guides/api-key/index.html` title/h1 `Get and store an API key` covers intervals.icu API key setup and storage.                                                                                                                        |
| README API-key/keychain prose                                       | https://icuvisor.app/guides/api-key/                                                                     | `web/public/guides/api-key/index.html` includes Intervals.icu settings, masked `icuvisor setup`, OS keychain storage, manual keychain commands, and client JSON warning not to paste API keys.                                                                                                                                               |
| README MCP transport prose                                          | https://icuvisor.app/guides/http-transport/ and https://icuvisor.app/reference/cli/                      | `web/public/guides/http-transport/index.html` h1 `Use Streamable HTTP transport` includes stdio default, `ICUVISOR_TRANSPORT=http`, loopback `127.0.0.1:8765/mcp`, `ICUVISOR_HTTP_BIND`, and LAN warning; `web/public/reference/cli/index.html` h1 `CLI reference` covers serve/stdio/http flags.                                            |
| README delete/write safety mode                                     | https://icuvisor.app/reference/safety-modes/ and https://icuvisor.app/explain/safety-modes/              | `web/public/reference/safety-modes/index.html` title/h1 `Safety modes and toolset tiers` includes `ICUVISOR_DELETE_MODE`, delete/write/read safety table, `_meta.delete_mode`, and tiers; `web/public/explain/safety-modes/index.html` h1 `Why safety modes exist` explains no model-controlled `confirm: true`, `none`, `safe`, and `full`. |
| README toolset tiers                                                | https://icuvisor.app/reference/safety-modes/#toolset-tier and https://icuvisor.app/explain/safety-modes/ | Same reference page includes toolset tiers and `_meta.toolset` echoes; explain page describes when to use `none`, `safe`, and `full` and points to the complete mode/tier table.                                                                                                                                                             |
| README post-upgrade/new-conversation line and `docs/post-update.md` | https://icuvisor.app/guides/after-upgrade/                                                               | `web/public/guides/after-upgrade/index.html` title/h1 `After upgrading icuvisor`, includes starting a new conversation, reconnecting clients, refreshing cached tool catalog, checking `icuvisor version`, and rerunning setup after config/keychain changes.                                                                                |
| `docs/clients/claude-desktop.md`                                    | https://icuvisor.app/connect/claude-desktop/                                                             | `web/public/connect/claude-desktop/index.html` title/h1 `Connect Claude Desktop`, includes `claude_desktop_config.json`, binary path, env fields, and verification prompt.                                                                                                                                                                   |
| `docs/clients/claude-code.md`                                       | https://icuvisor.app/connect/claude-code/                                                                | `web/public/connect/claude-code/index.html` title/h1 `Connect Claude Code`, includes `claude mcp add`, stdio configuration, env fields, and verification.                                                                                                                                                                                    |
| `docs/coach-mode.md`                                                | https://icuvisor.app/guides/coach-mode/ and https://icuvisor.app/explain/coach-mode/                     | `web/public/guides/coach-mode/index.html` title/h1 `Set up coach mode`, includes `ICUVISOR_COACH_MODE`, roster config, ACL patterns, `list_athletes`, `select_athlete`, and per-call `athlete_id`; `web/public/explain/coach-mode/index.html` title/h1 `Coach mode model` explains coach-scoped key and athlete routing.                     |
| TP-053 ChatGPT tutorial dependency                                  | https://icuvisor.app/tutorials/getting-started-chatgpt/ and https://icuvisor.app/connect/chatgpt/        | `web/public/tutorials/getting-started-chatgpt/index.html` title/h1 `Getting started with ChatGPT`, includes build/setup/API-key/MCP connector tutorial; `web/public/connect/chatgpt/index.html` title/h1 `Connect ChatGPT`, includes ChatGPT MCP configuration, stdio JSON shape, HTTP alternative, and verification.                        |

#### Blockers

- 2026-05-17 20:39 — Resolved by supervisor direction at 20:40. Live-site preflight failed because private/pre-launch `icuvisor.app` DNS/hosting is not live (`curl --http1.1 --retry 2 --retry-delay 1 -fsSL https://icuvisor.app/` and `/reference/tools/` failed with `curl: (35) Send failure: Broken pipe` / timeout to `162.255.119.238:443`; HTTP returned Namecheap URL Forward; HTTPS `www.icuvisor.app` failed TLS with `tlsv1 unrecognized name`). Supervisor approved local Hugo build/content evidence as the authoritative substitute for this batch, and `cd web && hugo --minify --gc` passed.

### Step 2: Inbound link sweep

**Status:** ✅ Complete

- [x] Run `git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**'` and record every product hit.
- [x] Add and follow a context-sensitive hit/action table; exclude historical `taskplane-tasks/**` artifacts from the product-link sweep because they are task records, not live inbound product links.
- [x] Replace each non-deferred deleted-doc inbound link with the appropriate icuvisor.app URL for end-user context, or remove redundant links. Updated `SECURITY.md`, `docs/dogfood/v0.2-findings.md`, `docs/internal-beta/onboarding-playbook.md`, `internal/app/setup.go`, and `internal/app/setup_test.go`; README hits are deferred to the Step 3 rewrite, and `docs/install/macos.md` internal hits are deferred to Step 4 deletion.
- [x] Re-run the scoped product grep and record remaining hits with their planned resolving step or confirm none remain outside README/deleted docs.
- [x] Address R006 missed relative links by running the broader product sweep `git grep -nE '(install/macos|claude-desktop|claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**'`, updating live product docs that point at deleted files, and recording any intentional false positives/deferred hits. Updated `docs/internal-beta/README.md` and `docs/internal-beta/onboarding-playbook.md`; remaining broader-grep hits are README hits deferred to Step 3, `docs/install/macos.md` hits deferred to Step 4 deletion, and an intentional kept link to `docs/threat-models/coach-mode.md` from `docs/coach-mode.md`.

#### Scoped inbound-link sweep

Command: `git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**'`

| Path                                           | Current context                                                          | Action                                                                                                                                                      | Replacement URL / deletion rationale                                                                                  |
| ---------------------------------------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `README.md:68`                                 | macOS install, Claude Desktop, and Claude Code user-doc links            | Defer to Step 3 rewrite                                                                                                                                     | README will be replaced with developer-only structure and user docs pointer to `https://icuvisor.app`.                |
| `README.md:91`                                 | coach-mode user-doc link                                                 | Defer to Step 3 rewrite                                                                                                                                     | README will remove user-facing coach-mode prose and point users to `https://icuvisor.app`.                            |
| `README.md:159`                                | post-upgrade/schema-change guidance                                      | Defer to Step 3 rewrite                                                                                                                                     | README will remove this line; destination is `https://icuvisor.app/guides/after-upgrade/`.                            |
| `README.md:161`                                | Claude Desktop and Codex local setup links                               | Defer to Step 3 rewrite                                                                                                                                     | README will remove Claude Desktop user setup link; keep developer-only Codex local docs if needed.                    |
| `SECURITY.md:53`                               | maintainer release-signing checklist points at deleted macOS install doc | Remove stale pointer, because the release secret/operator checklist is already duplicated inline in `SECURITY.md`; do not replace with public install page. | Deletion rationale: public macOS install page is end-user install content, not maintainer release-operator checklist. |
| `docs/dogfood/v0.2-findings.md:93`             | dogfood note points Claude Desktop manual config to deleted doc          | Update link to website connection guide.                                                                                                                    | `https://icuvisor.app/connect/claude-desktop/`                                                                        |
| `docs/install/macos.md:68`                     | internal link from file scheduled for deletion                           | Defer to Step 4 deletion.                                                                                                                                   | File will be removed.                                                                                                 |
| `docs/install/macos.md:69`                     | internal link from file scheduled for deletion                           | Defer to Step 4 deletion.                                                                                                                                   | File will be removed.                                                                                                 |
| `docs/internal-beta/onboarding-playbook.md:46` | stale tool-catalog/coach-mode troubleshooting pointer                    | Update to website coach-mode guide and explanation.                                                                                                         | `https://icuvisor.app/guides/coach-mode/` and `https://icuvisor.app/explain/coach-mode/`                              |
| `internal/app/setup.go:225`                    | user-visible setup output points at deleted Claude Desktop doc           | Update output to website connection guide.                                                                                                                  | `https://icuvisor.app/connect/claude-desktop/`                                                                        |
| `internal/app/setup_test.go:395`               | setup output assertion expects deleted path                              | Update expected string to website connection guide.                                                                                                         | `https://icuvisor.app/connect/claude-desktop/`                                                                        |

Historical `taskplane-tasks/**` hits are intentionally out of scope for Step 2 and Step 5 product-link sweeps; they are immutable task/review records and include this task's own prompt/status references.

#### Scoped post-edit grep

Command: `git grep -nE 'docs/(install/macos|clients/claude-desktop|clients/claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**' || true`

Remaining hits after Step 2 edits:

- `README.md:68`, `README.md:91`, `README.md:159`, `README.md:161` — planned resolution in Step 3 README rewrite.
- `docs/install/macos.md:68`, `docs/install/macos.md:69` — planned resolution in Step 4 deletion.

No remaining scoped hits exist in `SECURITY.md`, `docs/dogfood/v0.2-findings.md`, `docs/internal-beta/onboarding-playbook.md`, `internal/app/setup.go`, or `internal/app/setup_test.go`.

#### R006 broader post-edit grep

Command: `git grep -nE '(install/macos|claude-desktop|claude-code|coach-mode|post-update)\.md' -- . ':!taskplane-tasks/**' || true`

Remaining hits after R006 fixes:

- `README.md:68`, `README.md:91`, `README.md:159`, `README.md:161` — planned resolution in Step 3 README rewrite.
- `docs/install/macos.md:68`, `docs/install/macos.md:69` — planned resolution in Step 4 deletion.
- `docs/coach-mode.md:118` — intentional false positive for kept developer threat-model document `docs/threat-models/coach-mode.md`, not a link to deleted `docs/coach-mode.md`.

No remaining broader-grep hits exist in live product docs outside the README/deleted-doc deferrals and the kept threat-model false positive.

### Step 3: Rewrite README

**Status:** ✅ Complete

- [x] Replace `README.md` with the target developer-focused structure and remove migrated end-user sections.
- [x] Keep all existing badges and preserve the project layout block verbatim. Verified badges for Go Reference, Go Report Card, CI, Release, License, Go Version, codecov, OpenSSF Scorecard, and Conventional Commits; project layout block matches the previous README exactly.
- [x] Verify README style requirements from `CLAUDE.md` and confirm it stays concise (about 60-100 lines) with no deleted-doc links. Evidence: `wc -l README.md` reports 71 lines; grep for deleted-doc paths and removed end-user section headings returned zero hits in README; non-ASCII/emoji check found none.

### Step 4: Delete migrated docs/\*.md files

**Status:** ✅ Complete

- [x] Remove migrated files: `docs/install/macos.md`, `docs/clients/claude-desktop.md`, `docs/clients/claude-code.md`, `docs/coach-mode.md`, and `docs/post-update.md`.
- [x] Remove `docs/install/` only if empty; keep `docs/clients/` because `docs/clients/codex-local.md` must remain. Evidence: `docs/install/` was empty and removed; `docs/clients/codex-local.md` remains, so `docs/clients/` was kept.

### Step 5: Verify

**Status:** ✅ Complete

- [x] `git grep -n 'docs/install/macos\|docs/clients/claude-desktop\|docs/clients/claude-code\|docs/coach-mode\|docs/post-update'` shows zero product hits, with historical `taskplane-tasks/**` artifacts excluded and recorded. Evidence: `git grep -n 'docs/install/macos\|docs/clients/claude-desktop\|docs/clients/claude-code\|docs/coach-mode\|docs/post-update' -- . ':!taskplane-tasks/**'` returned zero hits; the same pattern under `taskplane-tasks/` has 127 historical task/review artifact hits intentionally excluded from product-link verification.
- [x] Update `CHANGELOG.md` `[Unreleased]` with the docs reorganization under `Changed`.
- [x] `make build`, `make test`, `make lint` pass. Evidence: `make build && make test && make lint` completed successfully; Go tests passed across `./...` and golangci-lint reported `0 issues.`
- [x] `cd web && hugo --minify --gc` passes. Evidence: Hugo v0.161.1 built 31 pages successfully in 34 ms.
- [x] Open/preview the rewritten README enough to verify badges render as Markdown image links and all links point to existing kept files or `https://icuvisor.app`. Evidence: Markdown link audit found 9 badge image links, local paths `ROADMAP.md`, `docs/kr5-benchmark.md`, `CONTRIBUTING.md`, `SECURITY.md`, `docs/prd/PRD-icuvisor.md`, and `LICENSE` all exist, `https://icuvisor.app` appears in the user pointers, and no deleted-doc references remain.

| 2026-05-17 20:28 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 20:28 | Step 1 started | Pre-flight verification |
| 2026-05-17 20:31 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-17 20:33 | Review R002 | plan Step 1: APPROVE |

| 2026-05-17 20:40 | Exit intercept reprompt | Supervisor provided instructions (964 chars) — reprompting worker |
| 2026-05-17 20:46 | Review R003 | code Step 1: APPROVE |
| 2026-05-17 20:49 | Review R004 | plan Step 2: REVISE |
| 2026-05-17 20:51 | Review R005 | plan Step 2: APPROVE |
| 2026-05-17 20:55 | Review R006 | code Step 2: REVISE |
| 2026-05-17 20:58 | Review R007 | code Step 2: APPROVE |
| 2026-05-17 21:00 | Review R008 | plan Step 3: APPROVE |
| 2026-05-17 21:03 | Review R009 | code Step 3: APPROVE |
| 2026-05-17 21:04 | Review R010 | plan Step 4: APPROVE |
| 2026-05-17 21:07 | Review R011 | code Step 4: APPROVE |

| 2026-05-17 21:10 | Agent escalate | TP-054 is blocked in Step 1 live-site preflight. The approved plan requires `curl -fsSL` evidence from https://icuvisor.app before deleting repo docs. Attempts failed: HTTPS to `icuvisor.app` timed ou |
| 2026-05-17 21:10 | Worker iter 1 | done in 2550s, tools: 160 |
| 2026-05-17 21:10 | Task complete | .DONE created |