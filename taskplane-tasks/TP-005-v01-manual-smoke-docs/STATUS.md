# TP-005 — Status

**Issue:** v0.1 — manual smoke docs
**Iteration:** 1
**Current Step:** Step 3: Add a repeatable local smoke checklist
**Last Updated:** 2026-05-11
**State:** Ready

## Step 1: Plan the manual config and smoke test

**Status:** ✅ Complete

- [x] Identify exact v0.1 config inputs
- [x] Check local `.env` availability for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`; record only availability, not values
- [x] Identify macOS Claude Desktop config file path and JSON shape
- [x] Use placeholders only for secrets/IDs
- [x] Write smoke-test plan in STATUS.md

### Smoke-test plan

1. Build or verify the binary with `make build`, then run `./bin/icuvisor version`.
2. Configure Claude Desktop on macOS by editing `~/Library/Application Support/Claude/claude_desktop_config.json` with an `mcpServers.icuvisor` entry that points to the local binary and supplies v0.1 config via placeholders-backed env vars or `--config` JSON.
3. Restart Claude Desktop and start a new chat so MCP schema caching does not retain an old tool catalog.
4. Confirm Claude lists `get_athlete_profile`, ask it to call the tool, and verify the response shape contains anonymized profile fields plus `_meta.server_version` without exposing credentials.
5. If local credentials are unavailable, document the remaining human maintainer verification instead of fabricating a network result.

## Step 2: Write manual setup documentation

**Status:** ✅ Complete

- [x] Document local build/install for v0.1
- [x] Document intervals.icu API key acquisition
- [x] Document config/env inputs
- [x] Document safe untracked `.env` flow without committing/displaying secrets
- [x] Provide Claude Desktop macOS JSON config example with placeholders
- [x] Explain MCP schema caching/new chat requirement
- [x] Include troubleshooting for common startup/auth/config errors

## Step 3: Add a repeatable local smoke checklist

**Status:** ✅ Complete

- [x] Checklist for `icuvisor version`
- [x] Checklist for `make build`
- [x] Checklist for Claude Desktop tool listing/callability
- [x] Expected anonymized `get_athlete_profile` response shape
- [x] Note manual smoke requires a real intervals.icu account/API key

## Step 4: Align code UX with docs if necessary

**Status:** ⬜ Not started

- [ ] Tighten confusing user-facing errors without leaking secrets
- [ ] Ensure invalid config failures are short/actionable
- [ ] Point README quickstart to detailed client guide
- [ ] Update `CHANGELOG.md`

## Step 5: Verify v0.1 gate

**Status:** ⬜ Not started

- [ ] Run `make build`
- [ ] Run `make test`
- [ ] Run `make lint` if available
- [ ] Perform manual Claude Desktop smoke test if credentials are available, or record remaining human verification
- [ ] Confirm every v0.1 roadmap checkbox is represented in TP-001 through TP-005

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |

| 2026-05-11 01:40 | Task started | Runtime V2 lane-runner execution |
| 2026-05-11 01:40 | Step 1 started | Plan the manual config and smoke test |
| 2026-05-11 | v0.1 config inputs identified | Required: `INTERVALS_ICU_API_KEY`/`api_key`, `INTERVALS_ICU_ATHLETE_ID`/`athlete_id`; optional: `ICUVISOR_TIMEZONE`/`timezone`, `ICUVISOR_API_BASE_URL`/`api_base_url`, `ICUVISOR_HTTP_TIMEOUT`/`http_timeout`, and `ICUVISOR_CONFIG` or `--config /path/to/icuvisor.json`. `.env` is read for recognized vars during local development; process env overrides `.env`. |
| 2026-05-11 | Local `.env` availability checked | `.env` is absent in this worktree, so no local `INTERVALS_ICU_ATHLETE_ID` or `INTERVALS_ICU_API_KEY` is available for automated maintainer smoke; no secret values were printed or recorded. |
| 2026-05-11 | Claude Desktop macOS config shape identified | File: `~/Library/Application Support/Claude/claude_desktop_config.json`. Shape: top-level `mcpServers` object; `icuvisor` entry with `command` set to the absolute binary path, optional `args` such as `--config /absolute/path/icuvisor.json`, and/or `env` entries for v0.1 config. |
| 2026-05-11 | Placeholder policy set | Documentation examples will use placeholders such as `/absolute/path/to/icuvisor`, `/Users/YOU/.config/icuvisor/icuvisor.json`, `YOUR_INTERVALS_ICU_API_KEY`, `i12345`, and `America/Sao_Paulo`; no real keys, athlete IDs, or machine-specific worker paths. |
| 2026-05-11 02:03 | Review R001 | plan Step 3: APPROVE |
