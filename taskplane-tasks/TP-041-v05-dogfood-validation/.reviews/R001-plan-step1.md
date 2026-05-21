# Plan Review — Step 1: Diagnostics subcommand

Decision: **changes requested**

I do not see a concrete Step 1 implementation plan beyond the checklist in `STATUS.md`, so this is not yet approvable as a plan. Before implementing, please spell out the design for the diagnostics command and tests. In particular, the plan needs to resolve these issues:

1. **Source of “last N tool-call names with timestamps” is undefined.** `icuvisor diagnostics` runs as a separate CLI subcommand, while the current MCP server has no persistent tool-call recorder. An in-memory ring buffer in the server would be empty/unavailable to a later diagnostics process. The plan must define a safe source of this data, how it is written by MCP tool calls, where it lives, and how diagnostics reads it without storing arguments, payloads, athlete IDs, or credentials.

2. **Catalog hash must not duplicate server filtering logic.** The active catalog hash depends on the same registration-time gates as the server (`ICUVISOR_DELETE_MODE`, `ICUVISOR_TOOLSET`, coach mode/ACL visibility). The current hash implementation is internal to `internal/mcp`. The plan should use the server’s existing catalog-hash path or expose a small reusable helper, rather than reimplementing filtering in `internal/app` and risking drift.

3. **Diagnostics should not start the MCP server or hit the network.** The command should load config, compute/report metadata, write only to `opts.Stdout`, and exit. The plan should state how tests prove `StartServer` is not called and no intervals.icu request is made.

4. **Secret-leakage testing needs a precise strategy.** The task asks to reject API keys, raw athlete IDs, and token-shaped strings, but the required catalog hash is itself a long hex string. The plan must define the matcher/allowlist so the test catches credential-like tokens without failing on the expected catalog hash. Also test both raw and normalized athlete IDs are absent; diagnostics should not print athlete IDs at all.

5. **Config-source output must be source-only.** Use `config.Config.APIKeySource` (`env`, `keychain`, `file`) or an explicit unset/error marker; do not print config paths, raw config contents, API keys, or `Config` dumps that could accidentally expand later. If diagnostics should still produce partial output when config loading fails, define that behavior and sanitize the error path.

6. **Help fixture path mismatch.** The prompt mentions `internal/app/testdata/help-fixture`, but the repo currently uses `internal/app/testdata/help.golden` and tests read that file. The plan should update `help.golden` (or explicitly rename/update tests), not add an unused fixture.

7. **CLI dispatch/help needs to be specified.** Add `diagnostics` to top-level help and decide whether `icuvisor diagnostics --help` has command-specific help. The command should bypass default server startup parsing and should support the existing config flags it needs (`--config`, `--env-file`, `--transport`, `--http-bind`) consistently with server startup.

8. **CHANGELOG is part of the task’s required documentation.** Since `diagnostics` is a user-visible command and the prompt lists `CHANGELOG.md`, include the changelog update in Step 1 or explicitly defer it to a named later step.

Suggested plan shape:

- Add `internal/app/diagnostics.go` with a small typed diagnostics model and `runDiagnosticsCommand(ctx, opts, args)`.
- Dispatch `diagnostics` in `Run` before default server startup; keep all output on `opts.Stdout`.
- Reuse the normal config loader and credential-store injection, but never print secret-bearing values.
- Compute the catalog hash through the same MCP/server registration path or a shared helper.
- Add/define a redacted recent-tool-call recorder that records only `{timestamp_utc, name}` and is readable by the diagnostics subcommand, or document an explicit `unavailable`/empty behavior if persistence is intentionally out of scope.
- Add table-driven tests for success, help, config flags, server bypass, catalog hash mode differences, and secret redaction across success and error paths.
- Update `internal/app/testdata/help.golden` and `CHANGELOG.md`.
