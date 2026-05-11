# TP-006 — Status

**Issue:** v0.1 — Codex local MCP validation
**Iteration:** 1
**Current Step:** Step 1: Discover current server and Codex CLI behavior
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

**Status:** ⬜ Not started

- [ ] Check `.env` availability for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`; record only availability, not values
- [ ] Prefer temporary Codex config/profile
- [ ] Backup and restore persistent Codex config if it must be touched
- [ ] Ensure secrets are not written to tracked files, logs, docs, fixtures, or STATUS
- [ ] Confirm `.env` remains untracked and unchanged

## Step 3: Launch Codex with icuvisor as an MCP server

**Status:** ⬜ Not started

- [ ] Configure Codex to launch icuvisor over stdio
- [ ] Start a fresh Codex session
- [ ] Confirm Codex can see icuvisor MCP server and list tools
- [ ] Record non-sensitive tool list

## Step 4: Exercise every registered MCP tool through Codex prompts

**Status:** ⬜ Not started

- [ ] Determine complete registered tool set for this build
- [ ] Run one Codex prompt per registered tool
- [ ] Explicitly test `get_athlete_profile` for v0.1
- [ ] Verify each tool call reaches server and returns valid terse shape
- [ ] Validate real intervals.icu-backed reads without recording raw personal data
- [ ] Record pass/fail, tool name, high-level response shape, and redacted observations

## Step 5: Cleanup, document, and verify

**Status:** ⬜ Not started

- [ ] Stop Codex/icuvisor processes started for validation
- [ ] Restore persistent Codex config from backup if modified
- [ ] Remove temporary files containing secrets
- [ ] Add `docs/clients/codex-local.md` if useful
- [ ] Run `make test` and `make build` after any code/doc changes
- [ ] Update `CHANGELOG.md` if docs or behavior changed
- [ ] Mark done only when every registered MCP tool has a result or documented blocker

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

| 2026-05-11 02:12 | Task started | Runtime V2 lane-runner execution |
| 2026-05-11 02:12 | Step 1 started | Discover current server and Codex CLI behavior |
| 2026-05-11 | Built binary path | `/Users/jusbrasil/prj/icuvisor/.worktrees/jusbrasil-20260510T182803/lane-1/bin/icuvisor` (Mach-O arm64) |
| 2026-05-11 | Codex help inspected | CLI supports `mcp` management, `-c key=value` config overrides, `--profile`, `exec --ephemeral`, and `exec --ignore-user-config`; `mcp add` supports stdio command plus `--env KEY=VALUE` |
| 2026-05-11 | Codex MCP config mechanism identified | Stdio MCP servers live under `mcp_servers.<name>` TOML config with `command`, `args`, and optional `env`; repeated `-c 'mcp_servers.icuvisor.*=...'` overrides work without writing `config.toml` when paired with a temporary `CODEX_HOME` |
| 2026-05-11 02:14 | Review R001 | plan Step 1: APPROVE |
| 2026-05-11 02:21 | Review R001 | code Step 1: APPROVE |
