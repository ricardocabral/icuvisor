# TP-038-first-run-onboarding: First-run onboarding subcommand — Status

**Current Step:** Step 1: UX script
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: UX script

**Status:** ⏳ Not started

- [ ] Prompt sequence drafted (pasted into Notes below)
- [ ] Masking via `golang.org/x/term` ReadPassword

### Step 2: Subcommand wiring

**Status:** ⏳ Not started

- [ ] `setup` registered in TP-035 parser
- [ ] `--help` golden file updated
- [ ] Overwrite prompts for keychain + config-file
- [ ] Honours `--config`

### Step 3: Autodetect + verify

**Status:** ⏳ Not started

- [ ] Profile fetch → athlete_id normalize + display name + FTP
- [ ] Timezone: `time.Local` → IANA validated
- [ ] 401/403: no writes, named error + fix URL
- [ ] `--offline` override

### Step 4: Write

**Status:** ⏳ Not started

- [ ] Keychain `Set` + immediate `Get` round-trip verify
- [ ] Non-secret fields → config file (no `api_key` ever written)
- [ ] `--force` to clobber existing config

### Step 5: Tests + manual sweep

**Status:** ⏳ Not started

- [ ] Faked stdin + credstore + intervals client
- [ ] Happy / 401 / network / offline / overwrite paths
- [ ] macOS manual sweep documented

---

## Decisions

- **GUI / webview:** explicitly deferred to v1.0; v0.5 ships terminal-only.

## Notes

_Add notes as work progresses._
