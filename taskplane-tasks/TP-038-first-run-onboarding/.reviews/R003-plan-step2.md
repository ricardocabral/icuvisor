# R003 plan review — Step 2: Subcommand wiring

Verdict: **REQUEST CHANGES**

The Step 2 section in `STATUS.md` currently repeats the task checklist, but it does not yet describe a concrete implementation plan. Because this step changes the top-level CLI dispatch and introduces the first credential-touching interactive command, please add a short Step 2 plan before coding.

## Blocking gaps to address in the plan

1. **Command dispatch shape is not specified.**
   - The plan should say where `setup` is detected in `internal/app.Run` and confirm it bypasses `LoadConfig`/`StartServer` entirely.
   - Tests should prove `icuvisor setup` does not start the MCP server or require an existing runtime config.
   - Prefer an injected setup runner/dependency in `app.Options` or similar so parser tests do not invoke real prompts/keychain access.

2. **Flag parsing contract needs to be explicit.**
   - Step 2 must cover `icuvisor setup --config <path>` and `icuvisor setup --config=<path>`.
   - Decide whether `icuvisor --config <path> setup` is supported; if not, document that the supported form follows current usage: `icuvisor <command> [flags]`.
   - Reserve/parse setup-only flags that later steps require (`--offline`, `--force`) now, or explicitly defer them with a plan to update help/tests in Step 3/4.
   - Confirm no `--api-key` or other command-line secret input is introduced.

3. **Help behavior is underspecified.**
   - Top-level `--help` golden fixture must list `setup`.
   - The plan should state what `icuvisor setup --help` does. Current `helpRequested` returns top-level help for every command except `version`; if that remains true, call it out intentionally. A setup-specific help path would be preferable if adding setup-only flags.

4. **Overwrite prompts need dependency and ordering details.**
   - For existing key detection, plan to call `credstore.Store.Get(ctx, credstore.IntervalsAPIKeyAccount)` and prompt `Overwrite? [y/N]` only on success; `credstore.ErrNotFound` should continue.
   - Default `no` should abort safely before reading a new API key or writing anything.
   - Define whether `--force` bypasses only config-file overwrite checks, or also the keychain prompt. The prompt text and acceptance criteria require warning before key overwrite; do not silently overwrite credentials unless the task is amended.

5. **Config path handling needs a source of truth.**
   - Step 2 says to honor `--config`, but setup must write to a path rather than call `config.Load` on a potentially nonexistent file. The plan should identify how the target path is resolved and passed into setup options.
   - Existing default config path behavior is not obvious in `internal/config/config.go`; if a helper is needed, include it in the plan so Step 4 does not duplicate path logic.

## Suggested minimum plan additions

Add a “Step 2 implementation plan” note to `STATUS.md` covering:

- Add `internal/app/setup.go` with `RunSetup(ctx, SetupOptions)` and a narrow interface for stdin/stdout/stderr, credential store, and config path.
- Extend `internal/app.Options` with an injectable setup runner/store (or equivalent) for tests.
- In `Run`, dispatch `setup` before default server startup and before runtime config loading.
- Add setup argument parser for `--config`, `--offline`, `--force`, and `--help`; unknown flags return `UsageError`/exit 2.
- Update `internal/app/help.go` and `internal/app/testdata/help.golden` in the same change.
- Add parser/dispatch tests for help, config path propagation, no server startup, unknown setup flags, and existing-key overwrite default-no.

Once these details are documented, the plan should be ready for implementation.
