# R004 Plan Review — Step 2: Backends

## Verdict: changes requested before implementation

The Step 2 goals in `PROMPT.md` are clear, and the Step 1 decisions are a good foundation. However, `STATUS.md` does not yet contain enough of a concrete backend plan to safely implement this security-sensitive step. Please tighten the plan before coding, especially around Linux degradation, test seams, context behavior, and logging.

## Blocking plan gaps

1. **Define the backend wrapper and test seam explicitly.**
   `the selected OS keyring module` exposes package-level `Get/Set/Delete` functions and a global mock provider. The Step 2 plan should say how `internal/credstore.OSKeychain()` will wrap those calls without letting normal unit tests touch the real OS keychain. Prefer an internal function-field or small adapter interface that can be faked per test; avoid relying on `keyring.MockInit()` unless tests are serial and restore global state. Real keychain tests must remain behind `//go:build keychain_live`.

2. **Specify exact Linux “unavailable keyring” mapping.**
   The current note says unavailable D-Bus/session keyring maps to `credstore.ErrNotFound` on `Get`, but it does not define which errors qualify. Add a planned helper such as `isKeychainUnavailable(err)` with Linux-specific implementation/tests. It should map only headless/unavailable cases (for example no session bus, no Secret Service provider, missing/default collection where appropriate) to `ErrNotFound` on reads. Permission denials, unlock failures, malformed D-Bus responses, and other unexpected backend errors should stay wrapped and visible. `Set` and `Delete` must not silently degrade to `ErrNotFound` for unavailable Linux keyrings.

3. **Clarify context semantics with a non-context upstream library.**
   The project interface is context-aware, but keyring backend calls are not. The Step 2 plan should state the intended behavior: at minimum check `ctx.Err()` before invoking the backend and after it returns. Do not use a goroutine/select wrapper that claims cancellation while the OS keychain operation continues in the background, especially for `Set`, because that could write a secret after the caller has canceled.

4. **Make the logging contract testable.**
   Step 2 requires “Log only the action verb and account name” and a slog capture test proving the secret is absent. The plan should say which operations log, at what level, and which fields are allowed. Keep secret values out of both log messages and fields; if logging errors, be careful that upstream error strings could be included. A good test should exercise `Set` with a known secret and assert the captured logs do not contain it.

5. **Record build-tag / cross-compile expectations.**
   If the implementation is just one wrapper file over `go-keyring`, say that no project-specific OS files are needed because the dependency supplies OS selection. If platform-specific helpers are added, list the `_linux.go` and `_nonlinux.go` split. The plan should include a quick validation target for `CGO_ENABLED=0` on linux/darwin/windows to protect the GoReleaser matrix.

## Suggested plan additions

- Add `OSKeychain() Store` returning a concrete `keyringStore` using `credstore.ServiceName` internally.
- Add unit tests for `Get`, `Set`, and `Delete` covering success, upstream not-found mapping, wrapped unexpected errors, context cancellation, and no-secret logging.
- Add Linux classifier tests that do not require a live D-Bus session.
- Keep live OS keychain tests optional under `keychain_live` only.

Once these details are captured in `STATUS.md`, the implementation plan should be ready to proceed.
