# Review R006 — Code Review for Step 2: Subcommand wiring

Verdict: **CHANGES REQUESTED**

The setup dispatch, setup-specific help, dependency injection, config-path precedence, and safe default-no prompts are largely in place. However, there is one security-sensitive parser issue to fix before approving Step 2.

## Findings

1. **`--api-key=...` is echoed back in the usage error.**  
   In `parseSetupArgs`, unknown setup args are formatted with the full raw argument (`internal/app/setup.go:179-180`). For `icuvisor setup --api-key=supersecret`, the CLI prints:
   
   ```text
   icuvisor: unknown setup flag "--api-key=supersecret"
   Run 'icuvisor setup --help' for usage.
   ```
   
   This violates the spirit of the task's “do not accept the API key on the command line” / credential-handling requirements by reprinting a likely secret to stderr. Please special-case `--api-key`/`--api-key=...` (or redact values for unknown `--*=...` flags generally) so the error mentions only the flag name, e.g. `unknown setup flag "--api-key"`, and add a regression assertion that the supplied value is not present in the error/stderr.

## Non-blocking notes

- `golang.org/x/term` is directly imported by `internal/app/setup.go`, but `go.mod` lists it as `// indirect` (`go.mod:19`). Please run/update module tidy so direct dependencies are represented cleanly.

## What looks good

- `setup` is dispatched before runtime config loading and server startup.
- `setup --help` is setup-specific and bypasses config/keychain/prompt/server side effects.
- `setup --config <path>`, `setup --config=<path>`, `--offline`, and `--force` are parsed and passed through to the injected runner.
- `ICUVISOR_CONFIG` fallback and the new platform default path helper match the Step 2 plan.
- Existing key and existing config prompts are default-no and return success on user cancellation without reading the new secret.

## Validation

- Ran `go test ./...` — passing.
