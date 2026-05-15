# TP-036-os-keychain-credential-storage: OS keychain credential storage — Status

**Current Step:** Step 1: Backend selection and contract
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 3
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Backend selection and contract

**Status:** ⏳ Not started

- [ ] Pick library (zalando/go-keyring vs 99designs/keyring); record rationale
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

- **Library:** TBD in Step 1. Recommendation in PROMPT.md is `zalando/go-keyring` (MIT, no CGO).

## Notes

_Add notes as work progresses._
