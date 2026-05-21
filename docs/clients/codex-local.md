# Codex CLI local MCP validation (macOS, v0.1)

This guide documents the v0.1 manual path for validating icuvisor with a local Codex CLI session. It is intended for maintainers and power users who want to prove the binary -> MCP stdio -> Codex -> intervals.icu path without editing persistent Codex MCP settings.

Use placeholders in examples and replace them only on your own machine. Do not commit API keys, real athlete IDs, local Codex config files, shell history containing secrets, or raw personal training data.

## Prerequisites

- macOS with Codex CLI installed and authenticated.
- Go 1.23 or newer.
- An intervals.icu account and API key for real end-to-end validation.
- A local clone of this repository.

The validation task used a Codex executable at:

```text
/Users/YOU/Library/pnpm/codex
```

If `codex` is on your `PATH`, you can use `codex` in place of the absolute executable path.

## Build the local binary

From the repository root:

```bash
make build
realpath bin/icuvisor
./bin/icuvisor version
```

Use the absolute path reported by `realpath` for MCP launch. icuvisor v0.1 starts its stdio MCP server when launched with no arguments. The `version` command is only for checking the build.

## Provide credentials safely

For real intervals.icu validation, prefer the normal setup flow so the API key is stored in the OS keychain and the generated config contains only non-secret metadata:

```bash
./bin/icuvisor setup
```

If this validation must run in a deliberately headless shell where keychain access is unavailable, make the required values available in the Codex process environment without printing them:

```bash
export INTERVALS_ICU_API_KEY="YOUR_INTERVALS_ICU_API_KEY"
export INTERVALS_ICU_ATHLETE_ID="i12345"
# Optional:
export ICUVISOR_TIMEZONE="America/Sao_Paulo"
```

A local untracked `.env` can also be used for maintainer smoke testing as the same fallback category, but do not display or commit it. Before committing, verify only its git status, not its contents:

```bash
git status --short .env
```

## Prefer ephemeral Codex MCP configuration

Codex supports MCP servers under `mcp_servers.<name>` in TOML config. For validation, prefer per-command overrides instead of `codex mcp add`, so no MCP server entry or secret is written to `~/.codex/config.toml`.

Use `env_vars` to pass non-secret variable names through from the Codex process environment. After `icuvisor setup`, pass only values such as `INTERVALS_ICU_ATHLETE_ID` and `ICUVISOR_TIMEZONE`; the API key is read from the OS keychain by the local icuvisor process. Include `INTERVALS_ICU_API_KEY` in `env_vars` only for the deliberate headless fallback described above, and never paste the key value into Codex config or prompts.

```bash
CODEX=/Users/YOU/Library/pnpm/codex
ICUVISOR=/absolute/path/to/bin/icuvisor
REPO=/absolute/path/to/icuvisor

"$CODEX" exec \
  --ignore-user-config \
  --ignore-rules \
  --ephemeral \
  -C "$REPO" \
  -c 'approval_policy="never"' \
  -c 'sandbox_mode="danger-full-access"' \
  -c "mcp_servers.icuvisor.command=\"$ICUVISOR\"" \
  -c "mcp_servers.icuvisor.cwd=\"$REPO\"" \
  -c 'mcp_servers.icuvisor.env_vars=["INTERVALS_ICU_ATHLETE_ID","ICUVISOR_TIMEZONE"]' \
  'List the icuvisor MCP tools by name only. Do not print environment variables.'
```

Notes:

- `--ignore-user-config` prevents loading your user `config.toml`, while `-c` supplies only the validation MCP server for this command.
- `--ephemeral` avoids persisting the Codex session transcript.
- `approval_policy="never"` is required for non-interactive `codex exec` MCP tool calls. Without it, Codex may list the tool but cancel a tool call before dispatch.
- `sandbox_mode="danger-full-access"` is used here only to prevent Codex's sandbox from blocking local MCP process execution during validation. Keep prompts narrow and do not ask Codex to edit files.
- Even with `--ignore-user-config`, Codex may update project trust metadata in `~/.codex/config.toml`. Check and restore the file if your validation policy requires zero persistent config drift.

## Confirm the tool catalog

Ask Codex to list the icuvisor tools:

```bash
"$CODEX" exec \
  --ignore-user-config \
  --ignore-rules \
  --ephemeral \
  -C "$REPO" \
  -c 'approval_policy="never"' \
  -c 'sandbox_mode="danger-full-access"' \
  -c "mcp_servers.icuvisor.command=\"$ICUVISOR\"" \
  -c "mcp_servers.icuvisor.cwd=\"$REPO\"" \
  -c 'mcp_servers.icuvisor.env_vars=["INTERVALS_ICU_ATHLETE_ID","ICUVISOR_TIMEZONE"]' \
  'List the MCP tools available from the icuvisor MCP server by name only. Do not run shell commands.'
```

For v0.1, the expected tool list is:

```text
get_athlete_profile
```

Start a fresh Codex session after rebuilding icuvisor or changing tool schemas. MCP clients can cache tool catalogs for a conversation.

## Exercise `get_athlete_profile`

After running `icuvisor setup` and keeping the API key in the OS keychain, run:

```bash
"$CODEX" exec \
  --ignore-user-config \
  --ignore-rules \
  --ephemeral \
  -C "$REPO" \
  -c 'approval_policy="never"' \
  -c 'sandbox_mode="danger-full-access"' \
  -c "mcp_servers.icuvisor.command=\"$ICUVISOR\"" \
  -c "mcp_servers.icuvisor.cwd=\"$REPO\"" \
  -c 'mcp_servers.icuvisor.env_vars=["INTERVALS_ICU_ATHLETE_ID","ICUVISOR_TIMEZONE"]' \
  'Use icuvisor to fetch my intervals.icu athlete profile. Summarize only non-sensitive fields and the response shape; do not include raw athlete IDs, API keys, or detailed training values.'
```

Expected result:

- Codex invokes `server=icuvisor`, `tool=get_athlete_profile`.
- The tool returns a terse structured object with profile fields, `units`, `sport_settings`, and `_meta`.
- `_meta.server_version` is present.
- Codex summarizes without printing API keys or unnecessary personal data.

If credentials are absent or invalid, the tool should return a short actionable error such as:

```text
could not fetch athlete profile; check intervals.icu credentials and athlete ID
```

That error still confirms Codex reached the MCP server and the tool handler, but it is not a real intervals.icu data validation.

## Optional direct MCP catalog check

If you need to distinguish a Codex issue from an icuvisor MCP issue, you can query the stdio server directly. The Go MCP SDK stdio transport used here expects newline-delimited JSON-RPC messages, not `Content-Length` framing.

For maintainers, a small script can send `initialize`, `notifications/initialized`, and `tools/list` to `bin/icuvisor` with non-secret dummy env values. The expected v0.1 result is exactly one tool: `get_athlete_profile`.

## Cleanup checklist

After validation:

1. Stop any Codex or icuvisor processes started for validation.
2. Remove temporary logs or transcripts that may contain local response data.
3. Confirm no MCP server entry or secret was written to tracked files.
4. If `~/.codex/config.toml` changed, restore it or document the exact persistent change.
5. Confirm `.env` remains untracked and unchanged.

## Troubleshooting

### Codex lists the tool but cancels the MCP call

Use config override `-c 'approval_policy="never"'` for non-interactive `codex exec` validation. Without it, Codex can cancel MCP tool calls before dispatch.

### `missing intervals.icu API key`

Run `./bin/icuvisor setup` to store the key in the OS keychain. For deliberate headless fallback only, set `INTERVALS_ICU_API_KEY` in the Codex process environment and include `INTERVALS_ICU_API_KEY` in `mcp_servers.icuvisor.env_vars`. Do not paste the key into the prompt.

### `missing athlete ID` or `invalid athlete ID`

Set `INTERVALS_ICU_ATHLETE_ID` to `i12345` or `12345` and include `INTERVALS_ICU_ATHLETE_ID` in `mcp_servers.icuvisor.env_vars`.

### Tool schema looks stale

Rebuild icuvisor, then start a fresh Codex session. Do not reuse an old conversation for schema validation.

### Persistent Codex config changed unexpectedly

Check `~/.codex/config.toml` metadata before and after validation if you need strict no-drift guarantees. During local validation, Codex may add a project trust block even when MCP server configuration is supplied with `--ignore-user-config` and `-c` overrides. Remove only the validation-added block if you need to restore the previous file state.
