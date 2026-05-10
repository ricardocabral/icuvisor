# Plan Review: TP-001 Step 3

Verdict: **Changes requested before implementation**

The high-level direction in `STATUS.md` is aligned with the prompt: add `internal/config`, keep credentials read-only, centralize athlete-ID normalization, and support JSON/env/`.env` inputs. However, the Step 3 plan is still too underspecified to implement safely and consistently. Please update `STATUS.md` with the concrete config contract before coding this step.

## Blocking plan gaps

1. **Name the public config contract.** Specify the JSON field names and env vars for each v0.1 input: API key, athlete ID, timezone, optional API base URL, optional HTTP timeout, and config file path. Also decide whether the app supports `--config`, only an env var, or a platform/default path for v0.1.

2. **Make precedence unambiguous.** The current note says “JSON first then environment/.env overrides,” but `.env` must not override explicit process/MCP-client env vars. A safer precedence should be written down, for example: defaults < JSON file < `.env` values loaded only when the key is absent < process env < CLI flags, if any. If you choose a different order, document why and test it.

3. **Define validation/default behavior.** The plan should state what is required vs defaulted: whether timezone is required or defaults to UTC/local, how HTTP timeout is represented in JSON/env and its default, what base URL default is used, how URL schemes are validated, and what short actionable errors are returned for missing API key / athlete ID / invalid timezone / invalid timeout.

4. **Clarify app integration for this step.** Step 2 intentionally avoided config loading. Step 3 should say whether default startup now calls `config.Load(...)` and passes the typed config into the starter/server info, or whether Step 3 only lands the package. If config loading is wired into default startup, ensure `icuvisor version` remains config-free.

5. **Plan secret redaction explicitly.** Add how API keys are kept out of `%v`/`String()`/errors/loggable structs. The loader may read keys from user-provided JSON/env in v0.1, but it must never create/write that file and must never include the raw key in validation errors.

6. **Scope `.env` support narrowly.** If implementing `.env` parsing instead of documenting an export command, specify that it is read-only developer convenience, limited to recognized `INTERVALS_ICU_*`/`ICUVISOR_*` keys, uses stdlib or a license-checked permissive dependency, and does not print loaded values.

## Implementation notes once the plan is updated

- Keep the package under `internal/config`; do not introduce `pkg/` or keychain storage in v0.1.
- Because file loading is I/O, prefer a context-aware loader shape such as `Load(ctx, Options)` and check cancellation before reading files.
- Centralize athlete ID normalization as a small exported function with the doc comment required by Go linting; accept `12345` and `i12345`, emit `i12345`, and reject empty/malformed values.
- Consider a redacted helper/type so future logs can safely include config summaries without accidental secret leakage.
- Tests are listed in Step 4, but the Step 3 design should already be table-testable; avoid adding behavior that will be hard to cover immediately afterward.

No source changes should be made for Step 3 until the above decisions are captured in `STATUS.md`.
