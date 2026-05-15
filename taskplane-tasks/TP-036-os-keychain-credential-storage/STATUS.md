# TP-036-os-keychain-credential-storage: OS keychain credential storage — Status

**Current Step:** Step 1: Backend selection and contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 3
**Review Counter:** 1
**Iteration:** 1
**Size:** M

---

### Step 1: Backend selection and contract

**Status:** 🟨 In Progress

- [x] Pick library (zalando/go-keyring vs 99designs/keyring); record rationale
- [ ] Define `Store` interface + canonical service/account names
- [ ] Decide `ErrNotFound` sentinel + wrapped-error semantics

### Step 2: Backends

**Status:** ⏳ Not started

- [ ] macOS / Windows / Linux backends, no CGO
- [ ] Linux headless D-Bus degradation
- [ ] No secret in logs (slog capture test)

### Step 3: Precedence chain in `internal/config`

**Status:** ⏳ Not started

- [ ] Order: env > keychain > file > error
- [ ] `_source` diagnostic indicator (env|keychain|file)
- [ ] Updated missing-key error message
- [ ] WARN on legacy file `api_key`

### Step 4: Tests + manual sweep

**Status:** ⏳ Not started

- [ ] Table-driven precedence tests
- [ ] Headless-D-Bus degradation test
- [ ] Manual three-OS sweep with platform-native UI
- [ ] `go test -race` clean

### Step 5: Documentation

**Status:** ⏳ Not started

- [ ] README "Getting an API key"
- [ ] CHANGELOG `[Unreleased]`
- [ ] SECURITY.md threat-model note

---

## Decisions

- **Library:** Selected `github.com/zalando/go-keyring` v0.2.8, tagged 2026-03-23T12:00:09Z (`go list -m -json github.com/zalando/go-keyring@v0.2.8`). Rationale: small OS-native backend abstraction for macOS Keychain, Windows Credential Manager, and Linux libsecret over D-Bus; no CGO required for target paths; avoids encrypted-file/pass fallbacks that would reintroduce a key-on-disk default.
- **License review:** `github.com/zalando/go-keyring` MIT; transitive deps from v0.2.8 are `github.com/danieljoos/wincred` v1.2.3 MIT, `github.com/godbus/dbus/v5` v5.2.2 BSD-style, and `golang.org/x/sys` v0.27.0 BSD-style. No GPL/copyleft dependency introduced.
- **Rejected backend:** `github.com/99designs/keyring` is MIT but has broader backend/config surface, including encrypted file/pass-style fallbacks that conflict with the project's "OS keychain by default, no plaintext/encrypted local file fallback" security posture.
- **Contract plan:** `internal/credstore.Store` will use context-aware signatures: `Get(ctx context.Context, account string) (string, error)`, `Set(ctx context.Context, account, secret string) error`, and `Delete(ctx context.Context, account string) error`.
- **Canonical names:** `credstore.ServiceName = "icuvisor"` and `credstore.IntervalsAPIKeyAccount = "intervals-icu-api-key"`; this task stores one API key per host, not per athlete. Coach-mode key layout is deferred to TP-039.
- **Error semantics:** `credstore.ErrNotFound` is the only fall-through sentinel and callers must use `errors.Is`. The wrapper maps `keyring.ErrNotFound` to `credstore.ErrNotFound`; Linux startup `Get` with unavailable D-Bus/session keyring also maps to `ErrNotFound` so config falls through to env/plaintext legacy sources. `Set`/`Delete` on unavailable Linux keyring should return actionable wrapped errors for TP-038 onboarding rather than pretending a write/delete succeeded. Other errors are wrapped with `%w` and must not include the secret value.
- **Credential precedence:** process environment `INTERVALS_ICU_API_KEY` is highest priority; OS keychain is next; plaintext `.env` and JSON `api_key` are legacy file sources below keychain. `.env` remains supported but is not treated as an explicit process-env override.
- **Source indicator:** `Config` will carry an unexported-to-JSON diagnostic source field (`json:"-"`) with values `env`, `keychain`, `file`, or empty. `.env` and JSON both report `file`; `Config.String()` will show the source while still redacting the secret.

## Notes

_Add notes as work progresses._

| 2026-05-15 13:39 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:39 | Step 1 started | Backend selection and contract |
| 2026-05-15 13:43 | Review R001 | plan Step 1: UNKNOWN |
