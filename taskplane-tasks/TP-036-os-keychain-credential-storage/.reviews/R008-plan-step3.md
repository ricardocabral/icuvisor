# R008 Plan Review — Step 3: Precedence chain in `internal/config`

## Verdict: approve

The Step 3 plan in `STATUS.md` now addresses the blocking gaps from R007 and is specific enough to implement safely.

## What looks good

- **Dependency injection is explicit.** Adding `CredentialStore credstore.Store` to `config.Options`, with nil meaning no keychain lookup, keeps unit tests deterministic and avoids accidental live keychain access.
- **Production wiring is covered.** The plan says the runtime path will install `credstore.OSKeychain()` before calling `config.Load`; in this codebase that likely belongs in `internal/app` around `parseDefaultArgs` / `startServer`, keeping `cmd/icuvisor/main.go` thin.
- **Precedence is unambiguous.** The planned chain is process env `INTERVALS_ICU_API_KEY` > keychain > legacy file sources > missing-key error, with process env skipping keychain lookup entirely.
- **Keychain errors are handled securely.** Only `credstore.ErrNotFound` falls through; unexpected keychain errors fail load instead of silently using plaintext fallback credentials.
- **Legacy `.env` behavior is clarified.** Treating `.env` API keys as file-sourced legacy credentials below keychain, while preserving the existing JSON/`.env` relative behavior for non-secret settings, matches the task's security posture without breaking existing installs.
- **Diagnostics are planned without leaking secrets.** `APIKeySource` with `json:"-"`, `Config.String()` redaction, and an expanded missing-key error provide useful support information without exposing the key.
- **Warning behavior is now broad enough.** Warning on any plaintext file-sourced API key, including `.env`, is the right migration signal.

## Implementation notes to keep in mind

- Coalesce legacy plaintext warnings so a load with both JSON `api_key` and `.env` `INTERVALS_ICU_API_KEY` emits one clear WARN and never includes the credential value.
- Keep the process-env API-key check separate from `.env`; only the real process/`opts.Env` value should suppress the keychain lookup.
- Add table cases for env-skips-keychain, keychain-overrides-JSON-and-dotenv, `ErrNotFound` fallthrough, unexpected keychain error failure, legacy warning redaction, `APIKeySource` rendering, and the updated missing-key message.

No plan changes are required before starting the Step 3 implementation.
