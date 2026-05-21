# Plan Review — TP-068 Step 4

## Verdict

Approved with guardrails. Step 4 should be a mechanical file split: move setup command dispatch and argument parsing into `internal/app/setup_cmd.go` without changing CLI behavior or setup-flow semantics.

## Plan notes / guardrails

- Keep the split narrow. Move `runSetupCommand`, `setupArgs`, `parseSetupArgs`, `resolveSetupConfigPath`, and the private command-parsing helpers they directly need (for example `unknownSetupFlagName` / setup usage-error wrapping) as a block where it improves cohesion. Do not further refactor `RunSetup`, profile fetching, timezone detection, or prompt behavior in this step.
- Preserve the existing command side-effect order exactly:
  1. parse setup args,
  2. default `Stdout`/`Stderr`,
  3. handle `--help` before resolving config paths or constructing keychain/prompter defaults,
  4. resolve config path,
  5. choose injected/default runner, credential store, and prompter,
  6. call the runner with the same `SetupOptions` fields.
- Preserve config-path precedence: explicit `--config` / `--config=...` wins, then `ICUVISOR_CONFIG`, then `config.DefaultPath()`.
- Preserve setup usage errors byte-for-byte where tests assert behavior. In particular:
  - missing `--config` values should point to `icuvisor setup --help`, not top-level help;
  - unknown inline flags should redact values (`--api-key=secret` reports `--api-key`, never the secret);
  - top-level/pre-command `--config ... setup` behavior remains owned by the default parser and should not become a setup command special case.
- Be deliberate with `newSetupUsageError`: it is also used by setup-flow validation for an empty API key. If it is moved to `setup_cmd.go`, that is technically fine in the same package but semantically less clean; consider leaving shared setup error helpers in the remaining setup support file unless you move all setup-specific usage-error helpers together intentionally.
- Do not introduce exported symbols or a new package. This is an intra-`internal/app` file organization change only.
- After moving code, check imports carefully. `setup.go` should shed command-only imports, and `setup_cmd.go` should own the command-only imports. Avoid leaving dead imports or moving unrelated helpers just to satisfy import cleanup.
- Existing dispatch tests in `internal/app/app_test.go` are the right regression gate. Splitting tests into a `setup_cmd_test.go` is optional, but not required if it would add churn without coverage benefit.

## Checks expected after implementation

- `go test ./internal/app`
- A targeted run such as `go test ./internal/app -run 'TestRunSetupDispatch|TestRunSetupUsesConfigEnvironmentFallback|TestRunSetupFlagErrors|TestRunSetupRejectsAPIKeyFlag|TestRunCLI'`
- Optionally `go test ./...` because the setup command is wired through the top-level CLI.

## Blocking findings

None.
