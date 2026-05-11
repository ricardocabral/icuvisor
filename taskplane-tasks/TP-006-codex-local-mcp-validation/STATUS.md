# TP-006 — Status

**Issue:** v0.1 — Codex local MCP validation
**Iteration:** 1
**Current Step:** Step 4: Exercise every registered MCP tool through Codex prompts
**Last Updated:** 2026-05-11
**State:** Ready

## Step 1: Discover current server and Codex CLI behavior

**Status:** ✅ Complete

- [x] Build current binary with `make build`
- [x] Confirm absolute binary path for MCP launch
- [x] Inspect `/Users/jusbrasil/Library/pnpm/codex --help` and relevant MCP/config help
- [x] Identify Codex MCP configuration mechanism and temporary profile/config option
- [x] Write validation plan before changing config

### Step 1 Validation Plan

- Use the freshly built absolute binary path `/Users/jusbrasil/prj/icuvisor/.worktrees/jusbrasil-20260510T182803/lane-1/bin/icuvisor`; icuvisor starts its stdio MCP server when launched with no arguments.
- Do not run `codex mcp add` against the default user config. Prefer `codex exec --ignore-user-config --ephemeral` with `-c mcp_servers.icuvisor.command=...`, `-c mcp_servers.icuvisor.cwd=...`, and `-c mcp_servers.icuvisor.env_vars=["INTERVALS_ICU_ATHLETE_ID","INTERVALS_ICU_API_KEY"]` so Codex uses temporary in-memory MCP configuration while preserving normal Codex auth.
- In Step 2, check `.env` for required variable availability only. Export values into the process environment for validation without printing them; never commit or record the values.
- Validate the tool catalog first, then run one Codex prompt per registered icuvisor tool. Record only tool names, pass/fail, high-level response shape, and redacted observations.
- If `exec --ignore-user-config` cannot load MCP config, fall back to a temporary `CODEX_HOME` or a temporary config file only if Codex auth still works; touch persistent Codex config only as a last resort with backup and restoration.

## Step 2: Prepare safe credentials and isolated Codex config

**Status:** ✅ Complete

- [x] Check `.env` availability for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`; record only availability, not values
- [x] Prefer temporary Codex config/profile
- [x] Backup and restore persistent Codex config if it must be touched
- [x] Ensure secrets are not written to tracked files, logs, docs, fixtures, or STATUS
- [x] Confirm `.env` remains untracked and unchanged

## Step 3: Launch Codex with icuvisor as an MCP server

**Status:** ✅ Complete

- [x] Configure Codex to launch icuvisor over stdio
- [x] Start a fresh Codex session
- [x] Confirm Codex can see icuvisor MCP server and list tools
- [x] Record non-sensitive tool list

### Step 3 Codex-visible Tool List

- `get_athlete_profile`

## Step 4: Exercise every registered MCP tool through Codex prompts

**Status:** 🟡 In Progress

- [x] Determine complete registered tool set for this build
- [x] Run one Codex prompt per registered tool
- [x] Explicitly test `get_athlete_profile` for v0.1
- [x] Verify each tool call reaches server and returns valid terse shape
- [x] Validate real intervals.icu-backed reads without recording raw personal data
- [x] Record pass/fail, tool name, high-level response shape, and redacted observations

### Step 4 Tool Validation Results

| Tool | Codex prompt result | High-level response shape | Redacted observations |
| ---- | ------------------- | ------------------------- | --------------------- |
| `get_athlete_profile` | Pass for Codex MCP dispatch and terse shape using local fake API; real intervals.icu read blocked by absent `.env` credentials | Object with `athlete_id`, name fields, `timezone`, `locale`, `units`, `sport_settings`, and `_meta`; `_meta.server_version` present; `sport_settings` is a list | Codex event log shows MCP tool call completed. No names, athlete IDs, FTP values, zones, API keys, or raw personal data are recorded here. |

## Step 5: Cleanup, document, and verify

**Status:** ⬜ Not started

- [ ] Stop Codex/icuvisor processes started for validation
- [ ] Restore persistent Codex config from backup if modified
- [ ] Remove temporary files containing secrets
- [ ] Add `docs/clients/codex-local.md` if useful
- [ ] Run `make test` and `make build` after any code/doc changes
- [ ] Update `CHANGELOG.md` if docs or behavior changed
- [ ] Mark done only when every registered MCP tool has a result or documented blocker

## Blockers

| Date | Blocker | Attempts | Current Impact |
| ---- | ------- | -------- | -------------- |
| 2026-05-11 | Real intervals.icu-backed `get_athlete_profile` validation cannot be completed because `.env` is absent and both required credential variables are unavailable | Checked `.env` availability without printing values; ran Codex with dummy non-secret values against upstream and confirmed the MCP tool returns the short credential/athlete-ID error; ran Codex with a local fake API to validate MCP dispatch and terse response shape | Real-data validation is documented as blocked; no API keys or personal data were recorded |

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

| 2026-05-11 02:12 | Task started | Runtime V2 lane-runner execution |
| 2026-05-11 02:12 | Step 1 started | Discover current server and Codex CLI behavior |
| 2026-05-11 | Built binary path | `/Users/jusbrasil/prj/icuvisor/.worktrees/jusbrasil-20260510T182803/lane-1/bin/icuvisor` (Mach-O arm64) |
| 2026-05-11 | Codex help inspected | CLI supports `mcp` management, `-c key=value` config overrides, `--profile`, `exec --ephemeral`, and `exec --ignore-user-config`; `mcp add` supports stdio command plus `--env KEY=VALUE` |
| 2026-05-11 | Codex MCP config mechanism identified | Stdio MCP servers live under `mcp_servers.<name>` TOML config with `command`, `args`, and optional `env`; repeated `-c 'mcp_servers.icuvisor.*=...'` overrides work without writing `config.toml` when paired with a temporary `CODEX_HOME` |
| 2026-05-11 | Step 2 started | Prepare safe credentials and isolated Codex config |
| 2026-05-11 | Credential availability checked | `.env` not present in worktree; `INTERVALS_ICU_ATHLETE_ID` unavailable; `INTERVALS_ICU_API_KEY` unavailable; no values printed or recorded |
| 2026-05-11 | Temporary Codex config preferred | Verified non-persistent MCP config with temporary `CODEX_HOME` and `-c mcp_servers.icuvisor.*` overrides; planned runtime uses `exec --ignore-user-config --ephemeral` plus `env_vars` inheritance, not `codex mcp add` |
| 2026-05-11 | Persistent Codex config handling | Default Codex config path exists, but TP-006 did not edit it; backup/restore not required unless the temporary config approach fails later |
| 2026-05-11 | Secret handling checked | `.env` absent; pending STATUS diff contains no `INTERVALS_ICU_*=` credential assignments; no local secret values were printed, logged, or written to tracked files |
| 2026-05-11 | `.env` tracking checked | `.env` is untracked, absent in this worktree, and has no git worktree status; one shell probe initially used read-only variable name `status` and was rerun successfully with `env_status` |
| 2026-05-11 | Step 3 started | Launch Codex with icuvisor as an MCP server |
| 2026-05-11 | Codex stdio config exercised | `codex exec --ignore-user-config --ignore-rules --ephemeral` launched with `mcp_servers.icuvisor.command` pointing at built binary, `cwd` at repo root, and inherited env var names; dummy non-secret env values used because `.env` is absent |
| 2026-05-11 | Fresh Codex session started | Non-interactive `codex exec --ephemeral` session completed and wrote final response to `/tmp/icuvisor-codex-step3-message.txt`; event log retained temporarily under `/tmp/icuvisor-codex-step3-events.jsonl` with no secrets |
| 2026-05-11 | Codex saw icuvisor tools | Codex final response listed `get_athlete_profile`, confirming the fresh session loaded the icuvisor MCP tool catalog |
| 2026-05-11 | Step 4 started | Exercise every registered MCP tool through Codex prompts |
| 2026-05-11 | Complete registered tool set determined | Codex catalog and direct newline-delimited MCP `tools/list` both report exactly one tool: `get_athlete_profile`; initial Content-Length framing probe failed because the Go SDK stdio transport expects newline-delimited JSON |
| 2026-05-11 | Codex prompt run for registered tools | Ran one `codex exec` prompt for the only registered tool, `get_athlete_profile`; using `approval_policy="never"` was required for non-interactive MCP tool execution, otherwise Codex cancelled the MCP call before dispatch |
| 2026-05-11 | `get_athlete_profile` explicitly tested | Codex invoked `server=icuvisor`, `tool=get_athlete_profile`, `arguments={"include_full": false}` in the successful local-backed run |
| 2026-05-11 | Tool reach/shape verified | `get_athlete_profile` reached the server and local fake API (`/api/v1/athlete/i0`); Codex event status completed with structured content keys `athlete_id,name,first_name,last_name,timezone,locale,units,sport_settings,_meta`, `_meta.server_version` present, and `sport_settings` as list |
| 2026-05-11 | Real intervals-backed validation blocked | `.env` is absent, so no real intervals.icu credentials are available; a Codex run against the default upstream with dummy non-secret env values reached the tool and returned the expected short credential/athlete-ID error without raw personal data |
| 2026-05-11 02:14 | Review R001 | plan Step 1: APPROVE |
| 2026-05-11 02:21 | Review R001 | code Step 1: APPROVE |
| 2026-05-11 02:23 | Review R001 | plan Step 2: APPROVE |
| 2026-05-11 02:27 | Review R001 | code Step 2: APPROVE |
| 2026-05-11 02:30 | Review R001 | plan Step 3: APPROVE |
| 2026-05-11 02:33 | Review R001 | code Step 3: APPROVE |
| 2026-05-11 02:35 | Review R001 | plan Step 4: APPROVE |
