# TP-036-os-keychain-credential-storage: OS keychain credential storage — Status

**Current Step:** Step 5: Documentation
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 3
**Review Counter:** 10
**Iteration:** 1
**Size:** M

---

### Step 1: Backend selection and contract

**Status:** ✅ Complete

- [x] Pick library (zalando/go-keyring vs 99designs/keyring); record rationale
- [x] Define `Store` interface + canonical service/account names
- [x] Decide `ErrNotFound` sentinel + wrapped-error semantics

### Step 2: Backends

**Status:** ✅ Complete

- [x] macOS / Windows / Linux backends, no CGO
- [x] Linux headless D-Bus degradation
- [x] No secret in logs (slog capture test)
- [x] Injectable go-keyring adapter seam with context checks and `OSKeychain()` constructor
- [x] Backend unit tests for success, not-found mapping, unexpected errors, context cancellation, and log redaction

### Step 3: Precedence chain in `internal/config`

**Status:** ✅ Complete

- [x] Order: env > keychain > file > error
- [x] `_source` diagnostic indicator (env|keychain|file)
- [x] Updated missing-key error message
- [x] WARN on legacy file `api_key`
- [x] `credstore.Store` injection in config options and production startup wiring
- [x] Keychain error flow: env skips lookup, only `ErrNotFound` falls through, other errors fail load

### Step 4: Tests + manual sweep

**Status:** 🟨 In Progress

- [x] Table-driven precedence tests
- [x] Headless-D-Bus degradation test
- [ ] Manual three-OS sweep with platform-native UI
- [x] `go test -race` clean

### Step 5: Documentation

**Status:** 🟨 In Progress

- [x] README "Getting an API key"
- [x] CHANGELOG `[Unreleased]`
- [x] SECURITY.md threat-model note

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
- **Step 2 backend plan:** implement `OSKeychain() Store` as a concrete wrapper over an internal adapter interface/function fields so normal tests fake `go-keyring` calls without touching the live OS keychain or relying on global mocks. The wrapper uses `ServiceName` internally and accepts only account/secret at the public interface.
- **Step 2 context plan:** because `go-keyring` is not context-aware, each operation checks `ctx.Err()` before calling the adapter and again after it returns. Do not wrap keychain operations in cancellation goroutines because `Set` could still write after cancellation.
- **Step 2 Linux degradation plan:** add an `isKeychainUnavailable(err)` helper with Linux and non-Linux implementations. On `Get` only, map `go-keyring` not-found and Linux headless/unavailable session/Secret Service/collection errors to `ErrNotFound`; permission denials, unlock failures, malformed responses, and unexpected backend errors remain wrapped. `Set` and `Delete` return actionable wrapped errors for unavailable Linux keyrings.
- **Step 2 logging plan:** log `credential get/set/delete` attempts and outcomes at debug level with `service` and `account` fields only. Do not include secret values. Tests capture `slog` output around `Set` with a known secret and assert it is absent from messages and fields.
- **Step 2 build plan:** prefer one wrapper file plus Linux/non-Linux classifier files. Validate `CGO_ENABLED=0 go test`/compile for linux, darwin, and windows targets after adding the dependency; live keychain tests remain behind `//go:build keychain_live`.
- **Step 3 injection plan:** add `CredentialStore credstore.Store` to `config.Options`. A nil store means no keychain lookup so existing tests and explicit headless loads remain deterministic; production startup wires `credstore.OSKeychain()` before invoking `config.Load`; tests pass fakes/`credstore.NoopStore` and never touch a live keychain.
- **Step 3 query plan:** read JSON and `.env` first to preserve existing non-secret settings and relative legacy-file precedence; then if trimmed process env `INTERVALS_ICU_API_KEY` is present, set source `env` and skip keychain entirely. If process env is absent, query `CredentialStore.Get(ctx, credstore.IntervalsAPIKeyAccount)`. A keychain value overrides plaintext `.env`/JSON; `credstore.ErrNotFound` falls through to legacy file values; any other keychain error fails load with a wrapped, actionable error and never falls through to plaintext credentials.
- **Step 3 legacy warning plan:** any plaintext file-sourced API key (`config.json api_key` or `.env INTERVALS_ICU_API_KEY`) emits one WARN at load with source/path metadata only; the credential value is never logged. Process env does not warn.
- **Step 3 source plan:** add a `Config.APIKeySource` diagnostic field with `json:"-"` and values `env`, `keychain`, `file`, or empty. `Config.String()` renders `api_key_source=<source>` beside `api_key=<redacted>`. `.env` and JSON both report `file`. The missing-key error names process env, OS keychain `service=icuvisor` / `account=intervals-icu-api-key`, JSON `api_key`, and `.env`.

## Notes

- **Blocker (Step 4 manual sweep):** Worker environment is a single macOS worktree and cannot access Windows Credential Manager or Linux Secret Service native UI sessions. Implemented and compiled platform backends/tests; README will document the three-OS manual recipes, but the actual Windows/Linux native UI smoke sweep requires operator hardware or CI runners for those OSes.

_Add notes as work progresses._

| 2026-05-15 13:39 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:39 | Step 1 started | Backend selection and contract |
| 2026-05-15 13:43 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 13:47 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 13:50 | Review R003 | code Step 1: APPROVE |
| 2026-05-15 13:53 | Review R004 | plan Step 2: REVISE |
| 2026-05-15 13:55 | Review R005 | plan Step 2: APPROVE |
| 2026-05-15 14:05 | Review R006 | code Step 2: APPROVE |
| 2026-05-15 14:08 | Review R007 | plan Step 3: REVISE |
| 2026-05-15 14:10 | Review R008 | plan Step 3: APPROVE |
| 2026-05-15 14:18 | Review R009 | code Step 3: APPROVE |
| 2026-05-15 14:22 | Review R010 | plan Step 4: APPROVE |
