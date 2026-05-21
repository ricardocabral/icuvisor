# Plan Review: Step 2 — Prepare safe credentials and isolated Codex config

## Verdict

Approved with required guardrails before executing Codex.

The plan in `STATUS.md` is aligned with the task: read `.env` only for availability, avoid persistent Codex configuration, use inherited environment-variable names rather than writing credential values, and keep secrets out of tracked files and logs. The Step 1 discovery also identifies a safe primary path: `codex exec --ignore-user-config --ephemeral` with `-c mcp_servers.icuvisor.*` overrides and `env_vars=["INTERVALS_ICU_ATHLETE_ID","INTERVALS_ICU_API_KEY"]`.

## Required guardrails for Step 2 execution

1. **Do not print or persist secret values.**
   - Do not run `cat .env`, `printenv`, `env`, `set`, shell tracing (`set -x`), or any command that echoes `INTERVALS_ICU_API_KEY` or the athlete ID.
   - Do not place values in `codex -c ...`, `codex mcp add --env KEY=VALUE`, temp config files, logs, docs, or `STATUS.md`.
   - If variables must be loaded for the Codex child process, load them silently with tracing disabled and pass only variable names via Codex `env_vars`.

2. **Use inherited env-var names, not inline values, for Codex MCP config.**
   - The safe shape is a transient override that names variables to inherit, e.g. `mcp_servers.icuvisor.env_vars=["INTERVALS_ICU_ATHLETE_ID","INTERVALS_ICU_API_KEY"]`.
   - Avoid `mcp_servers.icuvisor.env.INTERVALS_ICU_API_KEY=<value>` or command-line `KEY=value` wrappers for Codex, because those can leak through transcripts or process listings.

3. **Keep temporary files in ignored locations only.**
   - Prefer absolute `/tmp/...` or the repo’s ignored `tmp/` directory.
   - Do not use repo `.tmp/` unless it is first confirmed ignored; current `.gitignore` ignores `/tmp/` but not `.tmp/`.
   - Treat any temp Codex home/config/log as sensitive until inspected or deleted.

4. **Be precise when recording availability.**
   - `STATUS.md` should record only booleans or redacted status, for example: `INTERVALS_ICU_API_KEY available: yes/no` and `INTERVALS_ICU_ATHLETE_ID available: yes/no`.
   - Do not record the athlete ID, even partially, unless a later step has an explicit redaction policy.

5. **Confirm `.env` is untracked without exposing it.**
   - Use Git metadata checks such as `git ls-files -- .env .env.*` and ignore checks, not file dumps.
   - Normal `git status` alone is insufficient to prove ignored files are untracked.

6. **Persistent Codex config remains last resort.**
   - Continue with transient `codex exec --ignore-user-config --ephemeral -c ...` unless it fails.
   - If a persistent Codex config must be touched, document the exact path before editing, create a timestamped backup, restore it before finishing the task, and record only the path and restore status. Do not store API-key values there.

## Suggested concrete Step 2 flow

- Check whether `.env` exists without reading it aloud.
- Check only for the two key names using quiet matching or a no-output parser; record availability only.
- In the shell/session that will run Codex later, silently export the two values only if both are present and tracing is disabled.
- Prepare the transient Codex config command using the already-built absolute `bin/icuvisor` path, repo `cwd`, and `env_vars` names only.
- Verify `.env` is not tracked and was not modified.
- Update `STATUS.md` with redacted availability, selected isolated config strategy, and any blocker.

## Notes

No application-code or documentation changes are expected in Step 2. If credential availability is missing, the safe outcome is to record a redacted blocker and proceed only with non-live/tool-list validation in later steps as appropriate.
