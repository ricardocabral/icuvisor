# Plan Review — TP-078 Step 2

**Verdict:** APPROVE

## Summary

The Step 2 plan in `STATUS.md` is sufficient to proceed. It targets the only code gap surfaced by the Step 1 audit: `icuvisor setup` already uses `internal/credstore`, but `setupPersistConfigAndKey` currently writes the generated config before the keychain `Set`/`Get` verification. Reordering that sequence so keychain failures cannot leave a fresh onboarding config behind is the right implementation focus.

The plan also correctly keeps process env and legacy JSON/.env support as compatibility fallbacks rather than treating them as onboarding write paths.

## Guardrails for implementation

- Store and verify the API key with the existing `internal/credstore` service/account only: service `icuvisor`, account `intervals-icu-api-key`. Do not introduce a second credential namespace or config-selectable secret store.
- The generated config must not contain the API key value or an `api_key` field. Any credential metadata added to generated config should be non-secret and should reference only the existing keychain service/account.
- If config loading is updated to accept new credential metadata, keep it backward-compatible with existing configs and preserve current precedence: process env wins, keychain beats plaintext files when env is absent, and legacy JSON/.env remain fallback-only.
- Treat any new config metadata as descriptive/routing metadata, not as permission to weaken redaction or skip secret lookup rules.
- Keep setup failure behavior clear: keychain store/verification failure should happen before config write and should not print success text or leave a newly generated config.

## Non-blocking notes

- Because `readJSONConfig` currently uses `DisallowUnknownFields`, adding metadata to `config.Write` requires the read-side schema to accept the new field in the same step.
- Existing setup/config tests will likely need expectation updates after the write order and metadata changes; the fuller regression coverage can land in Step 3, but Step 2 should not leave the current test suite knowingly broken.
