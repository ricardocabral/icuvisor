# TP-006 — Status

**Issue:** v0.1 — Codex local MCP validation
**State:** Ready

## Step 1: Discover current server and Codex CLI behavior

**Status:** ⬜ Not started

- [ ] Build current binary with `make build`
- [ ] Confirm absolute binary path for MCP launch
- [ ] Inspect `/Users/jusbrasil/Library/pnpm/codex --help` and relevant MCP/config help
- [ ] Identify Codex MCP configuration mechanism and temporary profile/config option
- [ ] Write validation plan before changing config

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
