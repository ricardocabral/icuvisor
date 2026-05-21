# Plan Review R003 — Step 3: Tests

**Verdict:** Approved with required test coverage guardrails.

The Step 3 direction is appropriate: keep tests in `internal/app/app_test.go`, pin top-level help with `internal/app/testdata/help.golden`, and cross-check env-var names against the exported constants. Before marking the step complete, make the test plan concrete enough to catch the CLI contract issues from the prompt and the Step 2 guardrails.

## Required coverage for Step 3

1. **Exercise `RunCLI` for process-level behavior.**
   Since `cmd/icuvisor/main.go` now delegates exit handling to `app.RunCLI`, tests should assert:
   - `--help`, `-h`, `help`, `version --help`, and `version` return exit code `0`.
   - unknown flags/commands and missing flag values return exit code `2`.
   - runtime/config/server errors still return exit code `1`.

2. **Assert stdout/stderr separation.**
   Help and version output must go to `opts.Stdout` with an empty stderr. Usage errors must go to `opts.Stderr` via `RunCLI` and should not write partial help to stdout.

3. **Pin the required usage diagnostic, not just the final hint.**
   The prompt asks for a one-line `Usage:` on unknown-flag errors and the acceptance criteria require `Run 'icuvisor --help' for usage.`. Include assertions for both, especially on `--bogus`/unknown command. Missing-value errors should also assert the same usage-error classification and help hint.

4. **Golden-test the exact top-level help output.**
   Compare `Run(... --help ...)` stdout to `internal/app/testdata/help.golden`. Add equivalent cases proving `-h` and `help` produce the same bytes, so aliases cannot drift.

5. **Cover subcommand help separately.**
   `version --help` should print the version help stanza and must not print the injected version string or load config/start the server. A small golden/string assertion is sufficient; a second `version_help.golden` is optional.

6. **Test help precedence at more than one position.**
   Step 2 required help recognition “at any position.” Add at least one case such as `--transport http --help` or `--config --help` that proves help short-circuits before config loading/server startup and before missing-value validation.

7. **Keep parser-regression coverage for all current flags.**
   Existing valid-flag coverage should include both separated and inline forms for `--config`, `--transport`, `--http-bind`, and `--env-file`, and should assert the resulting `config.Options` values.

8. **Cross-check env-var names from constants.**
   Build the list from `config.Env*`, `safety.Env*`, and `response.EnvDebugMetadata` constants and assert each appears in the golden help. This catches stale fixture text when env vars are renamed or added.

## Notes

- Avoid spawning the compiled binary unless you want a very small smoke/integration test; unit tests against `Run`/`RunCLI` are enough for this step.
- Keep tests table-driven with `t.Run`, matching the existing style.
- The current task status still has only high-level Step 3 bullets, so update it after implementation with the actual tests added and any intentional omissions.
