# R007 Plan Review — Step 3: Precedence chain in `internal/config`

## Verdict: changes requested before implementation

The Step 3 target behavior is clear, and the `STATUS.md` decisions cover the intended high-level precedence (`process env > keychain > legacy file`). However, the current Step 3 plan is still too terse for a security-sensitive config-loader change. Please tighten the plan in `STATUS.md` before coding so the implementation does not accidentally hit a developer's live keychain in tests, silently ignore keychain failures, or keep ambiguous legacy-file behavior.

## Blocking plan gaps

1. **Specify the config-loader injection and production wiring.**
   The prompt requires a `credstore.Store` injection point in `internal/config`, but the Step 3 plan does not say what field will be added to `config.Options`, what the nil/default behavior is, or where production startup gets `credstore.OSKeychain()`. This needs to be explicit. Prefer a deterministic test path where unit tests pass a fake/`NoopStore`, and production CLI startup wires `OSKeychain()` before calling `config.Load`; avoid making ordinary `config.Load` tests depend on whatever happens to be in the developer or CI keychain.

2. **Define exactly when the keychain is queried.**
   Because `INTERVALS_ICU_API_KEY` is the highest-priority explicit process-env source, the loader should check that value first and skip the keychain lookup when it is present. This avoids unnecessary OS-keychain prompts/latency/failures while preserving deterministic headless/CI behavior. The plan should state this directly.

3. **State keychain error handling in the Step 3 flow.**
   `credstore.ErrNotFound` should be the only fall-through case. Any other `Store.Get` error must fail config loading with an actionable, wrapped error and must not fall through to plaintext file credentials. This is already part of the backend contract, but Step 3 should restate it where the precedence chain is implemented.

4. **Clarify legacy file source handling, including `.env`.**
   `STATUS.md` says plaintext `.env` and JSON `api_key` are both legacy file sources below keychain, while the checklist only says “WARN on legacy file `api_key`.” The plan should specify whether a `.env` API key also emits the migration warning. Given the task's security goal, any plaintext file-sourced API key (`config.json` or `.env`) should warn once at load without logging the secret value. Also preserve the existing relative ordering between JSON and `.env` for non-secret settings unless intentionally changing it.

5. **Spell out source plumbing and diagnostics.**
   The plan should name the source field/values and how it reaches `Config.String()` without being serialized or exposing the key. For example: carry an internal `apiKeySource`/`APIKeySource` with `json:"-"`, values `env|keychain|file`, and render `api_key_source=<source>` alongside `api_key=<redacted>`. Also update the missing-key error to mention process env, OS keychain service/account, JSON `api_key`, and `.env` if that remains supported.

## Suggested plan additions

- Add `CredentialStore credstore.Store` (or similarly named) to `config.Options`; document nil behavior.
- Production path: wire `credstore.OSKeychain()` in the CLI/app startup path before invoking `config.Load`; tests use an in-memory fake or `credstore.NoopStore`.
- Precedence algorithm: read JSON and `.env` legacy file values; if trimmed process env API key exists, set API key/source `env` and do not query keychain; otherwise query `IntervalsAPIKeyAccount`; on success set source `keychain`; on `ErrNotFound` keep legacy file source; on any other error return a wrapped config error.
- Add/plan table cases for env skipping keychain, keychain overriding JSON/`.env`, `ErrNotFound` falling through, unexpected keychain error failing, legacy file warning redaction, source rendering, and updated missing-key text.

Once these details are captured in `STATUS.md`, the Step 3 implementation plan should be ready to proceed.
