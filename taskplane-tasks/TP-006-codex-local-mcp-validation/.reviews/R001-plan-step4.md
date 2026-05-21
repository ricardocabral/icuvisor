# Plan Review: Step 4 — Exercise every registered MCP tool through Codex prompts

## Verdict

Approved with required guardrails before execution.

The Step 4 direction is aligned with the task: use the fresh Codex session to invoke each registered icuvisor MCP tool, record only pass/fail and high-level response shape, and keep raw intervals.icu data out of `STATUS.md`. For this build, both Step 3 and `internal/tools/registry.go` indicate the registered catalog is only `get_athlete_profile`, so the step should be small and focused.

The main risk is that Step 2 found no local `.env` and no available `INTERVALS_ICU_ATHLETE_ID` / `INTERVALS_ICU_API_KEY`. Step 4 must not turn a missing-credential result into a successful live intervals.icu validation. It may still prove that Codex invoked the tool and that the server returned a safe, actionable credential error, but the real-data read should be recorded as blocked unless credentials are supplied through the already-approved safe path.

## Required guardrails for Step 4 execution

1. **Confirm the registered tool set from a fresh source.**
   - Use the Step 3 Codex-visible catalog as the primary observed list.
   - Cross-check against the current source registry or a direct MCP `tools/list` probe if there is any discrepancy or schema-cache concern.
   - Record the final non-sensitive catalog in `STATUS.md`; for the current build it should be `get_athlete_profile` only.

2. **Use Codex prompts that force a tool call but suppress raw data in the final answer.**
   - For `get_athlete_profile`, use a prompt that explicitly asks Codex to call icuvisor and report only success/failure plus top-level response shape.
   - Do not ask Codex to print the athlete name, athlete ID, first/last name, locale, exact thresholds, zones, or raw JSON.
   - A safe prompt shape is: `Use icuvisor to fetch my intervals.icu athlete profile with default arguments. Do not print actual profile values or raw JSON. Report only whether the call succeeded and which top-level fields were present, such as athlete_id, units, sport_settings, and _meta.`

3. **Handle missing credentials honestly.**
   - Since Step 2 recorded credentials unavailable, the expected live intervals.icu data validation is currently blocked.
   - If running with dummy or absent env vars, treat the outcome as an invocation/error-path validation only. Record it as blocked or failed for live data, not as a pass for real-data validation.
   - Do not fetch credentials from persistent Codex config, shell history, parent directories, password managers, or user files beyond the task-approved `.env`/environment path without explicit user approval.

4. **Keep Codex output capture sensitive and temporary.**
   - Codex JSON/event logs can contain full MCP tool arguments and tool results. Write them only under `/tmp` or another confirmed ignored location.
   - If real credentials become available and a successful profile call occurs, assume the logs contain personal/training data; redact observations before copying anything into `STATUS.md`, then remove the raw logs during cleanup.
   - Do not commit, paste, or summarize raw personal data in tracked files.

5. **Verify the server was actually reached.**
   - Use Codex event output, final response wording, or redacted logs to confirm a `get_athlete_profile` tool invocation occurred.
   - A generic Codex answer that merely describes the tool without invoking it is not sufficient.
   - If Codex refuses or fails to call the tool, record that as a Codex invocation blocker and optionally use direct MCP only as a diagnostic, not as a replacement for the required Codex prompt result.

6. **Validate the default terse path first.**
   - Invoke `get_athlete_profile` with default arguments (`include_full` omitted or false) for the required v0.1 check.
   - Do not use `include_full: true` unless there is a specific diagnostic need, because Step 4 is validating terse-by-default behavior and avoiding unnecessary identifiers.

7. **Record results in `STATUS.md` with redaction.**
   - For each tool, record: tool name, Codex invocation status, live-data status, high-level response/error shape, and redacted observation.
   - Acceptable examples: `tool invoked; returned safe credential error because credentials unavailable` or `tool invoked; structured response contained athlete_id, units, sport_settings, _meta; raw values not recorded`.
   - Do not mark Step 4 complete unless every registered tool has a Codex validation result or a clearly documented blocker.

## Suggested Step 4 flow

1. Reconfirm the catalog for the current binary/fresh Codex session: `get_athlete_profile`.
2. Run one fresh transient `codex exec --ignore-user-config --ignore-rules --ephemeral` prompt using the same non-persistent `mcp_servers.icuvisor.*` overrides from Step 3.
3. Prompt Codex to invoke `get_athlete_profile` with default args and to report only top-level shape or a safe error summary.
4. Inspect only redacted output/logs to confirm the MCP tool call occurred.
5. Update `STATUS.md` with the result. If credentials remain unavailable, explicitly record live intervals.icu profile validation as blocked due to missing credentials, while separately noting whether Codex successfully invoked the tool and received the safe credential error.

## Notes

No application-code changes are expected in Step 4. If validation reveals a real server bug, stop and document it before making code changes so the task remains focused on Codex/local MCP validation.
