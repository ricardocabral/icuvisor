# Review R005 — Plan Review for Step 2: Subcommand wiring

Verdict: **APPROVE**

The updated Step 2 plan addresses the blocking gaps from R004 and is ready to implement.

What looks good:

- `setup` is explicitly dispatched before runtime config loading and server startup, so first-run setup does not require an existing config or accidentally launch the MCP server.
- The setup runner/dependency injection plan should keep parser tests away from real prompts, keychain access, and network/client behavior.
- Flag handling is now concrete for `setup --config <path>`, `setup --config=<path>`, `--offline`, `--force`, and `--help`, while correctly excluding any command-line API-key input.
- The plan intentionally keeps `icuvisor --config <path> setup` unsupported and calls for tests/docs to reflect the current `icuvisor <command> [flags]` parser shape.
- `setup --help` is specified as setup-specific help with exit 0 and no config/keychain/prompt/server side effects.
- Existing key handling uses `credstore.Store.Get(ctx, credstore.IntervalsAPIKeyAccount)`, treats `ErrNotFound` as continue, prompts before overwriting, and does not let `--force` silently overwrite credentials.
- Config target resolution and existing-file behavior are now explicit: `setup --config`, then `ICUVISOR_CONFIG`, then platform default; `os.Stat` only; default-no cancellation; `--force` bypasses only the config-file clobber prompt.
- The planned tests cover the important parser/dispatch and safe-cancel paths.

Non-blocking implementation note: order the existing-config prompt before collecting/validating a new secret where practical, so a default-no config cancellation truly leaves the flow with no secret input and no writes. At minimum, tests should assert the promised “nothing changed” behavior for existing-key and existing-config cancellations.
