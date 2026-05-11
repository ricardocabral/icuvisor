# Plan Review: Step 3 — Launch Codex with icuvisor as an MCP server

## Verdict

Approved with required guardrails before execution.

The planned direction is appropriate: launch the freshly built `bin/icuvisor` as a stdio MCP server from a fresh Codex session, use transient `-c mcp_servers.icuvisor.*` overrides instead of `codex mcp add`, and record only the non-sensitive tool catalog in `STATUS.md`. This keeps the task within the prompt's validation scope and avoids persistent user configuration changes.

## Required guardrails for Step 3 execution

1. **Make the exact launch method explicit in `STATUS.md` before running it.**
   - Record a redacted command shape using the absolute binary path and repo `cwd`.
   - Include `--ignore-user-config` and `--ephemeral` for the Codex session unless a documented blocker requires a fallback.
   - Be careful with TOML quoting for `-c` overrides, for example command/cwd as TOML strings and env inheritance as variable names only.

2. **Do not modify persistent Codex configuration for the primary path.**
   - Do not use `codex mcp add` against the default user config.
   - If transient config fails and a persistent config fallback is considered, stop first, document the exact file, back it up, and restore it before finishing as required by the task prompt.

3. **Do not leak credentials or compensate for missing credentials unsafely.**
   - Step 2 found `.env` absent and both `INTERVALS_ICU_*` values unavailable in this worktree. Step 3 can still validate server launch and tool listing, but should not attempt a live `get_athlete_profile` call yet.
   - Continue passing only environment variable names, not values, if env inheritance is configured.
   - Do not run `env`, `printenv`, `set`, `cat .env`, shell tracing, or commands that could echo credential values.

4. **Use a fresh Codex session and avoid schema-cache ambiguity.**
   - Start a new `codex exec --ephemeral` session after the latest `make build` output.
   - If a tool list looks stale or missing, rebuild/relaunch and record the cache concern rather than reusing an old session.

5. **Validate actual tool visibility, not only server configuration.**
   - `codex mcp list` is useful to verify the configured server entry, but it may only show configuration. The Step 3 acceptance item requires Codex to see the server's available tools.
   - Use a fresh Codex prompt that asks for the available icuvisor MCP tools, or a direct MCP `tools/list` probe as a cross-check if Codex cannot expose the list cleanly. Record which method produced the tool list.

6. **Keep output capture redacted and ignored.**
   - If using `--json` or logs, write them only under `/tmp` or another confirmed ignored location and treat them as sensitive until inspected.
   - Copy only tool names and non-sensitive observations into `STATUS.md`.

7. **Record expected and observed catalog clearly.**
   - For this build, `internal/tools/registry.go` appears to register `get_athlete_profile` only. Step 3 should still rely on the Codex/direct MCP result as the observed catalog, then record the non-sensitive list in `STATUS.md`.

## Suggested Step 3 flow

- Confirm the built binary path still exists.
- Run a fresh transient Codex session with `-c mcp_servers.icuvisor.command=...`, `-c mcp_servers.icuvisor.cwd=...`, and env-var-name inheritance only.
- Ask Codex to report the icuvisor MCP server/tool catalog without invoking credentialed tools.
- If Codex cannot provide the tool list, run a minimal direct MCP `tools/list` probe against the same binary as a diagnostic and document the Codex-specific blocker separately.
- Update `STATUS.md` with pass/fail, the observed non-sensitive tool list, and whether persistent config remained untouched.

## Notes

No application-code changes are expected in Step 3. Because credentials are unavailable, live intervals.icu validation should be deferred to Step 4 and recorded as blocked unless a safe credential source is provided without exposing values.
