# R009 Code Review — Step 3: Precedence chain in `internal/config`

## Verdict: request changes

`go test ./...` passes, and the main precedence path (`env > keychain > legacy file`) is mostly implemented correctly. I found one security/migration behavior gap against the approved Step 3 plan.

## Findings

### 1. Legacy plaintext API keys are only warned when they are the selected source

- **File:** `internal/config/config.go:165-186`, `internal/config/config.go:398-402`
- **Severity:** Medium

The Step 3 plan says any plaintext file-sourced API key (`config.json api_key` or `.env INTERVALS_ICU_API_KEY`) should emit one WARN at load, without leaking the value. The current implementation calls `warnLegacyAPIKey(cfg, raw)` after validation, and that helper returns unless `cfg.APIKeySource == APIKeySourceFile`.

That means no migration warning is emitted when a plaintext key is present on disk but is overridden by a higher-priority source, for example:

- `config.json` contains `api_key`
- the OS keychain also contains a key
- `Load` correctly uses the keychain key, but `cfg.APIKeySource == keychain`, so the plaintext config warning is skipped

The same applies when a real process env var wins over a plaintext JSON/`.env` key. The plaintext credential still exists on disk and remains the security risk this task is trying to surface to users.

Suggested fix: track whether JSON or `.env` contained a non-empty API key before higher-priority sources override it, then emit a single redacted WARN if any legacy plaintext key was present. Keep the selected `APIKeySource` as-is for diagnostics. Add table coverage for “keychain wins but JSON/.env plaintext key still warns” and “process env wins but JSON/.env plaintext key still warns”.

## What looks good

- Process env API key skips keychain lookup.
- Only `credstore.ErrNotFound` falls through; unexpected keychain errors fail load.
- Keychain-sourced keys override plaintext file keys.
- `Config.String()` still redacts the API key and includes the source indicator.
- Production startup wires `credstore.OSKeychain()` while keeping `config.Options.CredentialStore == nil` deterministic for unit tests.

## Verification

- Ran `go test ./...` successfully.
