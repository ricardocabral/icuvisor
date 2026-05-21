# Plan Review — TP-078 Step 3

**Verdict:** APPROVE

## Summary

The Step 3 plan in `STATUS.md` is sufficient to proceed. It targets the right regression surface for the Step 2 changes: generated setup config should contain only non-secret keychain metadata, keychain store/verify failures must not leave a fresh config behind, diagnostics/loggable output must stay sanitized, and config loading should have explicit coverage for supported and unsupported `credential_ref` metadata while preserving existing fallback precedence.

## Guardrails for implementation

- Use a distinctive fake API key in tests and assert it is absent from generated config bytes, setup stdout/stderr, diagnostics output, and any captured loggable config output.
- For keychain `Set`, verify `Get`, and mismatch failures, assert both the actionable service/account guidance and the absence of success text; also assert the target config path was not created.
- Cover the load side with table-driven cases for:
  - generated/supported `credential_ref` plus keychain secret;
  - unsupported type/service/account producing a short actionable error;
  - process env still winning and skipping keychain lookup;
  - keychain beating legacy JSON/.env when env is absent;
  - legacy JSON/.env fallback still working when keychain returns `credstore.ErrNotFound`.
- Treat `credential_ref` as non-secret metadata only. Tests should not make it configurable as an alternate credential namespace beyond the existing service `icuvisor` and account `intervals-icu-api-key`.
- Keep tests hermetic: use fake stores/prompters/loaders and no live OS keychain or network access.

## Verification expected

Run and record the targeted commands requested by the task:

```sh
go test ./internal/app ./internal/config ./internal/credstore
```

If broader packages are touched while adding diagnostics/logging coverage, include their targeted tests as well.
