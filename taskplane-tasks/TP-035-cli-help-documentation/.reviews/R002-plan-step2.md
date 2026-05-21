# Plan Review R002 — Step 2: Parser changes

**Verdict:** Approved with required implementation guardrails.

The Step 2 checklist matches the prompt and the Step 1 artifacts now cover the key design choices: stdlib/hand-rolled parser, `internal/app/testdata/help.golden`, full env-var inventory, and exit-code meanings. The parser work can proceed, but the implementation should explicitly handle the items below to avoid failing the acceptance criteria.

## Required guardrails for Step 2

1. **Add a real usage-error classification, not just a different error string.**
   `cmd/icuvisor/main.go` currently exits `1` for every `app.Run` error. To make unknown flags and missing flag values exit `2`, introduce a stable mechanism such as an `app.UsageError`/`app.ExitCode(err)` helper and update `main` to use it. Do not rely on string matching in `main`.

2. **Decide one owner for stderr output and avoid duplicate messages.**
   Help must be written to `opts.Stdout`. Usage diagnostics must be test-capturable via `opts.Stderr` or returned for `main` to print, but not both. If `app.Run` writes usage errors to `opts.Stderr`, then `main` should detect the already-classified usage error and only `os.Exit(2)` without printing a second `icuvisor: ...` line.

3. **Preserve no-config/no-server short-circuit behavior.**
   `--help`, `-h`, `help`, and `version --help` must return before config loading or server startup. `icuvisor version` must continue to print the injected version and exit without loading config.

4. **Make help precedence unambiguous.**
   Because the requirement says help is recognized “at any position,” scan exact tokens `--help` and `-h` before validating flag values. That prevents cases like `icuvisor --config --help` from becoming a missing-value error. Keep `version --help` as the subcommand-help exception; other placements of `--help`/`-h` should print top-level help.

5. **Keep the full current flag surface, including `--env-file`.**
   Step 2’s checklist mentions preserving `--config`, `--transport`, and `--http-bind`, but the existing parser also supports `--env-file` and `--env-file=...`. That flag is in the Step 1 help fixture and must not regress.

6. **Pin the usage diagnostic text now.**
   The prompt asks for a one-line `Usage:` on unknown-flag errors and the acceptance criteria require `Run 'icuvisor --help' for usage.`. A safe shape is:

   ```text
   icuvisor: unknown command or flag "--bogus"
   Usage: icuvisor [flags]
   Run 'icuvisor --help' for usage.
   ```

   Missing-value errors should use the same usage-error classification and include the same help hint, even if the first line names the specific missing flag.

7. **Do not broaden the command tree in this step.**
   `icuvisor help` is required. `icuvisor help version` or a general subcommand framework is not required; avoid adding semantics that are not in the prompt unless they fall out trivially and are tested.

## Notes for the next test step

- Add assertions for exit code `2` via the new helper/type, not by spawning the binary unless you specifically want a small `main` integration test.
- Include cases for `--config=value`, `--config value`, `--env-file=value`, and `--env-file value` to prove the parser refactor did not drop existing forms.
- Include a `version --help` test that proves it prints subcommand help rather than the version string.

With these guardrails, the Step 2 plan is appropriately scoped and should satisfy the parser-related acceptance criteria.
