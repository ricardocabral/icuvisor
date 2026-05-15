# R010 Plan Review — Step 4: Tests + manual sweep

## Verdict: request changes

The Step 4 checklist is pointed in the right direction, but it is too thin for this security-sensitive step and it does not yet close a known Step 3 review gap. Please revise the plan before executing Step 4.

## Findings

### 1. Step 4 must explicitly cover the unresolved legacy-plaintext warning gap

- **Severity:** High
- **Files:** `internal/config/config.go:186`, `internal/config/config.go:398-402`, `internal/config/config_test.go:509-626`

`R009-code-step3.md` requested changes because `warnLegacyAPIKey` only warns when the selected credential source is `file`. If a plaintext `api_key` exists in `config.json` or `.env` but is overridden by process env or keychain, the user receives no migration warning even though the risky plaintext secret is still on disk.

The current Step 4 plan says only “Table-driven precedence tests,” which would allow this security regression to remain untested. Add explicit test cases and implementation work for:

- keychain wins over `config.json`/`.env`, but a single redacted legacy plaintext warning is still emitted;
- process env wins over `config.json`/`.env`, but a single redacted legacy plaintext warning is still emitted;
- the warning includes only source/path metadata and never the sentinel secret value;
- the selected `APIKeySource` remains `env` or `keychain`, not `file`.

Also fix the STATUS inconsistency: it records R009 as “APPROVE”, but the review file says “request changes”.

### 2. The manual sweep plan is not verifiable enough for a three-OS keychain feature

- **Severity:** Medium
- **Requirement:** Step 4 manual sweep and acceptance criterion “A key stored in the OS keychain on macOS/Windows/Linux is read by `icuvisor` at startup…”

The plan says “Manual three-OS sweep with platform-native UI” but does not specify where/how those OS checks will run, the exact credential identifiers, cleanup, or what output proves success. Please expand it into a matrix with at least:

- macOS: exact `security`/Keychain Access write and delete recipe for `service=icuvisor`, `account=intervals-icu-api-key`;
- Windows: exact `cmdkey`/Credential Manager target compatible with `go-keyring` (`icuvisor:intervals-icu-api-key`), plus delete command;
- Linux: exact `secret-tool` attributes (`service icuvisor username intervals-icu-api-key`), plus cleanup;
- the command used to start `icuvisor` with `INTERVALS_ICU_API_KEY`, JSON `api_key`, and `.env` API key unset;
- the smoke-test/tool call used to prove the loaded key works;
- where results will be recorded in `STATUS.md`, including any skipped OS and why.

If you cannot access all three platforms in this lane, the plan should say that explicitly and require CI/VM/human handoff evidence rather than silently marking the sweep complete.

### 3. Add a gated live-keychain integration test plan

- **Severity:** Medium
- **Requirement:** prompt file scope: “Real keychain hits are gated behind `-tags keychain_live` so CI does not need a desktop session.”

There are currently no `keychain_live` tests. The Step 4 plan should add a small gated integration test that writes a random sentinel via `credstore.OSKeychain().Set`, reads it back, and deletes it in cleanup. It should skip/actionably fail when the host keychain is unavailable, and it must never use or log a real intervals.icu API key.

This does not replace the native-UI manual sweep, but it gives maintainers a repeatable end-to-end check without touching CI by default.

### 4. Make log-redaction verification automated, not just manual

- **Severity:** Medium
- **Requirement:** Step 4 “startup log does not contain the key under any precedence path” and acceptance criterion “secret value is never written to logs”.

The plan should include automated slog-capture cases for sentinel secrets loaded from each source: process env, keychain fake, JSON, and `.env`. These should assert the sentinel never appears in logs, errors, or `Config.String()`, while the source indicator remains visible.

Manual inspection of binary startup logs is useful, but it is not enough for a regression-prone security invariant.

## Verification performed during review

- `go test ./...` passes.
- `go test -race ./...` passes on the current host.
