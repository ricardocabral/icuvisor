# Review R004 — Plan Review for Step 2: Subcommand wiring

Verdict: **REQUEST CHANGES**

The revised plan is much more concrete and covers the key parser/dispatch concerns: `setup` runs before config loading/server startup, has injectable dependencies for tests, does not introduce command-line secret input, and treats `--force` as config-only rather than silently overwriting credentials. Those are the right directions.

Before coding, tighten the plan on a few Step 2 requirements that are still underspecified:

1. **Config target path and existing-config prompt need explicit Step 2 behavior.**
   Step 2's checklist includes overwrite prompts for both keychain and config-file, but the plan only fully specifies the existing-key prompt. Add concrete rules for:
   - target path precedence: `setup --config`, likely `ICUVISOR_CONFIG`, then the platform default (`os.UserConfigDir()/icuvisor/config.json` or a central helper);
   - existence check that does **not** call `config.Load` or require the config to be valid;
   - prompt text/default for an existing config file;
   - `--force` bypassing only the config-file clobber prompt;
   - tests for existing config default-no and `--force`.

2. **Define the prompt/password abstraction, not just streams.**
   Production must use `golang.org/x/term` `ReadPassword`, which does not work against an arbitrary `io.Reader`. The dependency bundle should include a narrow `SecretReader`/`Prompter` function or interface so tests can fake masked input without a TTY while production still uses the standard masked path. Keep the API key out of args, logs, and returned errors.

3. **Spell out dispatch/help ordering with the current parser shape.**
   The current `helpRequested` path will otherwise tend to route `icuvisor setup --help` to top-level help unless dispatch is changed deliberately. The plan should explicitly say `setup --help` uses setup-specific help, does not load config, does not touch keychain, and does not start the server. Add a parser test for that exact path.

4. **Define user-abort semantics.**
   For default-no on an existing key/config prompt, specify whether `RunSetup` returns nil with a “setup canceled; nothing changed” message or a usage/runtime error, and the expected exit code. Tests should assert that aborting leaves key/config/server untouched.

5. **Revisit `--config` placement if needed.**
   Supporting `icuvisor setup --config <path>` is probably sufficient given the documented `icuvisor <command> [flags]` form, but the task says “Honour `--config` (existing flag).” If `icuvisor --config <path> setup` is intentionally unsupported, make sure help/docs do not imply global pre-command flags for subcommands; otherwise support it in the parser.

Once these are added to the plan, it should be ready to implement Step 2.
