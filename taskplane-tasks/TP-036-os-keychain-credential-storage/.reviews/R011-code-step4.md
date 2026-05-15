# R011 Code Review — Step 4: Tests + manual sweep

## Verdict: request changes

The diff for this step only updates `STATUS.md`; no committed test or implementation changes were made. `go test -race ./...` is clean on this host, but Step 4 is marking several security/test obligations as complete while the underlying gaps from the prior reviews remain open.

## Findings

### 1. Legacy plaintext keys still do not warn when overridden by env/keychain

- **Severity:** High
- **Files:** `internal/config/config.go:165-186`, `internal/config/config.go:398-402`, `internal/config/config_test.go:523-542`, `taskplane-tasks/TP-036-os-keychain-credential-storage/STATUS.md:46`

Step 4 marks “Table-driven precedence tests” complete, but the unresolved Step 3 security gap is still present and still untested. `warnLegacyAPIKey` returns unless the selected credential source is `file`, so a plaintext `api_key`/`.env` secret left on disk produces no migration warning when it is overridden by process env or OS keychain.

That contradicts the recorded Step 3 legacy warning plan in `STATUS.md:78` (“any plaintext file-sourced API key ... emits one WARN at load”) and the Step 4 plan review request. The existing precedence cases cover env/keychain winning over plaintext files, but they do not capture logs or assert that a redacted legacy warning is emitted for those overridden plaintext credentials.

Please fix the loader to track plaintext-file presence independently from the selected `APIKeySource`, then add table coverage for at least:

- process env wins, plaintext JSON/`.env` exists, one redacted warning is emitted, selected source remains `env`;
- keychain wins, plaintext JSON/`.env` exists, one redacted warning is emitted, selected source remains `keychain`;
- the sentinel plaintext value never appears in logs/errors/`Config.String()`.

### 2. The live keychain test was temporary and is not reproducible

- **Severity:** Medium
- **Files:** `taskplane-tasks/TP-036-os-keychain-credential-storage/STATUS.md:75`, `taskplane-tasks/TP-036-os-keychain-credential-storage/STATUS.md:83`

The task prompt and the status plan say real keychain hits should be gated behind `//go:build keychain_live`, but there is still no committed `keychain_live` test (`grep keychain_live` finds only task/review text). `STATUS.md:83` says a temporary `TestTP036LiveOSKeychainDummy` was run, but because that test is not in the repository, maintainers cannot rerun the same validation on macOS/Windows/Linux or during future onboarding work.

Please add a small committed, build-tagged live integration test that writes a random dummy secret with `credstore.OSKeychain().Set`, reads it back, deletes it in cleanup, and skips/actionably fails when the host keychain is unavailable. It must not use or log a real intervals.icu key.

### 3. STATUS records rejected reviews as approved and closes checkboxes without the requested evidence

- **Severity:** Medium
- **Files:** `taskplane-tasks/TP-036-os-keychain-credential-storage/STATUS.md:46-49`, `taskplane-tasks/TP-036-os-keychain-credential-storage/STATUS.md:97-98`

`STATUS.md` still records R009 and R010 as `APPROVE`, but the actual review files are both “request changes.” R010 explicitly asked for this inconsistency to be fixed and for the Step 4 plan to include the missing automated checks/manual evidence. The current commit instead checks off Step 4 items and appends a manual-validation note while leaving those review outcomes incorrect.

Please correct the review history to match the review files and do not mark the Step 4 test items complete until the requested automated cases/live-test hook are committed (or explicitly document which accepted steering decision waives which item).

## Verification performed

- `git diff ff8788ac1eec87c28683f0de75a45d6866a3681c..HEAD --name-only` shows only `STATUS.md` changed.
- `go test -race ./...` passes on this host.
