# Plan Review: TP-001 Step 2

Verdict: **Request changes before implementation**

The Step 2 direction is mostly aligned with the task: keep `main` thin, preserve `icuvisor version`, move CLI/default startup handling into an internal package, and inject the build-time version instead of letting lower layers discover it themselves.

## Required adjustments

1. **Do not pull Step 3 config loading into Step 2.**
   - `STATUS.md` currently says the default startup should "validate/load config" before returning a placeholder error. That belongs to Step 3.
   - For Step 2, the default path can delegate to an internal app/server-start function and return a short actionable placeholder such as "stdio server not implemented yet". It should not define env/JSON precedence, `.env` behavior, or validation yet.

2. **Add the Step 2 test target to the plan explicitly.**
   - The prompt requires `icuvisor version` to remain working and be covered by a test where practical.
   - Prefer testing `internal/app` with injected args/stdout/stderr rather than trying to test `main` directly. At minimum cover:
     - `version` writes the injected version to stdout and returns nil.
     - default invocation delegates to the startup path and returns its error.
     - unknown commands have a short actionable error, if unknown-command handling is introduced.

3. **Define the app entrypoint shape before coding.**
   - To keep `main` thin and testable, Step 2 should plan an API like `app.Run(ctx, app.Options{Version: version, Args: os.Args[1:], Stdout: os.Stdout, Stderr: os.Stderr}) error` or equivalent.
   - Internal packages should return errors only; `main` should be the only place that writes final stderr output and chooses `os.Exit(1)`.

4. **Make version propagation concrete.**
   - Keeping `main.version` is fine, but internal packages cannot depend on `main`; the version must be injected into `internal/app` options and stored in whatever future runtime/server config object lower layers will consume.
   - Avoid adding a separate `internal/version` package unless the Makefile/ldflags are updated consistently in this step.

## Non-blocking notes

- The default startup path should accept `context.Context` now, because it will become blocking I/O when stdio MCP wiring lands.
- Avoid importing the MCP SDK or intervals client in this step unless necessary to compile; the prompt explicitly asks to avoid implementing those areas beyond interfaces/stubs.
- `CHANGELOG.md` can wait until Step 5 per the task sequence, unless Step 2 is committed independently with user-visible CLI behavior changes.

After narrowing Step 2 away from config implementation and adding the test/entrypoint details above, the plan should be safe to execute.
