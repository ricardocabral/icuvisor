# R005 Plan Review — Step 2: Backends

## Verdict: approved

The revised Step 2 plan in `STATUS.md` addresses the blocking gaps from R004 and is detailed enough to implement the backend safely.

## What is now covered

- **Backend/test seam:** `OSKeychain() Store` will be a concrete wrapper over an internal adapter/function-field seam, so normal unit tests can fake `go-keyring` without touching the live OS keychain or relying on global mocks.
- **Context behavior:** the plan correctly acknowledges that the keyring backend is not context-aware and commits to checking `ctx.Err()` before and after calls, without using cancellation goroutines that could hide a later `Set` side effect.
- **Linux degradation:** `isKeychainUnavailable(err)` with Linux/non-Linux implementations is the right shape. The plan limits read-time degradation to headless/unavailable keyring cases and keeps permission/unlock/malformed/unexpected errors visible. It also correctly avoids pretending `Set`/`Delete` succeeded when Linux keyring access is unavailable.
- **Logging/redaction:** debug-level action/outcome logs with only service/account fields are acceptable, and the planned slog capture test around `Set` with a known secret directly exercises the important redaction property.
- **Build compatibility:** the one-wrapper plus Linux/non-Linux classifier split, optional `keychain_live` tests, and `CGO_ENABLED=0` cross-compile validation protect the release matrix.

## Implementation notes to keep in mind

- Prefer `errors.Is(err, keyring.ErrNotFound)` for upstream not-found mapping, and preserve `%w` wrapping for all other backend failures.
- Keep Linux unavailable classification narrow and covered by pure unit tests; avoid broad substring matching that could accidentally suppress real security or permission errors.
- If failure logs are added, log an error class/outcome rather than raw error text unless you have verified it cannot contain secret material.
- Restore any process-global slog logger changes in tests with `t.Cleanup` so log-capture tests remain isolated and parallel-safe.

No further plan changes are required before implementation.
