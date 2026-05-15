# TP-038-first-run-onboarding: First-run onboarding subcommand — Status

**Current Step:** Step 1: UX script
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
**Size:** M

---

### Step 1: UX script

**Status:** ✅ Complete

- [x] Prompt sequence drafted (pasted into Notes below)
- [x] Masking via `golang.org/x/term` ReadPassword

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

| 2026-05-15 18:48 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 18:48 | Step 1 started | UX script |

### Step 1 UX script

Terminal-only v0.5 flow:

```text
Welcome to icuvisor.
This setup stores your intervals.icu API key in the OS keychain and writes non-secret settings to your icuvisor config file.

Paste your intervals.icu API key (from https://intervals.icu/settings):
> ********

Checking intervals.icu… connected as "Jane Doe" (athlete i12345, FTP 245 W).
Detected timezone: Europe/Madrid. Use this? [Y/n]
> 

Save athlete id i12345 and timezone Europe/Madrid to /Users/jane/.config/icuvisor/config.json? [Y/n]
> 

Saved. Your key is in the macOS Keychain; athlete id + timezone are in /Users/jane/.config/icuvisor/config.json.
Test connection OK: Jane Doe, FTP 245 W.

Next: point Claude Desktop at icuvisor — see docs/clients/claude-desktop.md
```

Re-run prompts:

```text
An API key is already stored. Overwrite? [y/N]
A config file already exists at /Users/jane/.config/icuvisor/config.json. Overwrite? [y/N]
```

Failure paths:

```text
API key not accepted by intervals.icu. Double-check the key on https://intervals.icu/settings.
Could not reach intervals.icu. Nothing was written. Re-run setup when online, or use --offline to store settings without verification.
```

Offline mode:

```text
Offline setup skips intervals.icu verification. Your API key will be stored, but icuvisor cannot confirm it works until you run a tool.
Athlete ID (accepts 12345 or i12345):
Timezone (IANA name, for example Europe/Madrid):
```

Masking decision: the implementation will use the standard `golang.org/x/term` `ReadPassword` path for API-key input. `go.mod` does not currently include `golang.org/x/term`; add it during the setup implementation rather than introducing a fancy prompt library.
| 2026-05-15 18:51 | Review R001 | plan Step 1: APPROVE |
| 2026-05-15 18:54 | Review R002 | code Step 1: APPROVE |
