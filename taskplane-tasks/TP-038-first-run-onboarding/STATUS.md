# TP-038-first-run-onboarding: First-run onboarding subcommand — Status

**Current Step:** Step 3: Autodetect + verify
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 10
**Iteration:** 1
**Size:** M

---

### Step 1: UX script

**Status:** ✅ Complete

- [x] Prompt sequence drafted (pasted into Notes below)
- [x] Masking via `golang.org/x/term` ReadPassword

### Step 2: Subcommand wiring

**Status:** ✅ Complete

- [x] `setup` registered in TP-035 parser
- [x] `--help` golden file updated
- [x] Overwrite prompts for keychain + config-file
- [x] Honours `--config`
- [x] R006: redact setup unknown-flag values, especially `--api-key=...`, and tidy direct `golang.org/x/term` dependency

### Step 3: Autodetect + verify

**Status:** ✅ Complete

- [x] Profile fetch → athlete_id normalize + display name + FTP
- [x] Timezone: `time.Local` → IANA validated
- [x] 401/403: no writes, named error + fix URL
- [x] `--offline` override
- [x] R009: replace `time.Local == Local` UTC fallback with real IANA timezone detection and regression test

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

### Step 2 implementation plan

Plan review R003 requested a concrete CLI wiring plan before coding. Implement Step 2 as follows:

- Add `internal/app/setup.go` with `RunSetup(ctx, SetupOptions)` and narrow injected dependencies for input/output streams, credential store, setup profile client, and config target path. The real runner can be simple in Step 2, but the interface must let app/parser tests avoid real prompts and keychain access.
- Extend `internal/app.Options` with an injectable setup runner/dependency bundle. `Run` should dispatch `setup` before runtime config loading and before server startup so `icuvisor setup` does not require an existing config and cannot accidentally start the MCP server.
- Support setup flags in command position: `icuvisor setup --config <path>`, `icuvisor setup --config=<path>`, `--offline`, `--force`, and `--help`. Do not add `--api-key` or any command-line secret input. Keep `icuvisor --config <path> setup` unsupported because the current parser documents and implements `icuvisor <command> [flags]`; help/docs for setup must not imply pre-command global flags.
- `icuvisor setup --help` must use setup-specific help, return exit 0, and bypass config loading, keychain access, prompt reads, and server startup. Update the top-level help and golden fixture so `setup` is listed, and add tests for `setup --help`.
- Prompt abstraction: production setup uses `golang.org/x/term.ReadPassword` via a narrow `SecretReader`/`Prompter` dependency instead of reading secrets from args or a generic `io.Reader`; tests fake this dependency without a TTY. Returned errors and logs must not include the API key.
- Existing key handling: call `credstore.Store.Get(ctx, credstore.IntervalsAPIKeyAccount)`; on `credstore.ErrNotFound`, continue; on success, prompt `An API key is already stored. Overwrite? [y/N]` before reading a new key or writing. Default `no` returns nil after printing `Setup canceled; nothing changed.`, preserving exit 0 as a user cancellation rather than a usage/runtime failure. `--force` does not silently overwrite credentials; it is reserved for config-file clobbering in Step 4.
- Config target path precedence: `setup --config` wins; then `ICUVISOR_CONFIG`; then platform default `os.UserConfigDir()/icuvisor/config.json` exposed through a central `internal/config` helper if needed. Existence checks use `os.Stat` only and do not call `config.Load` or require valid JSON. If the file exists, prompt `A config file already exists at <path>. Overwrite? [y/N]`; default no returns nil with `Setup canceled; nothing changed.` and no key/config/server side effects. `--force` bypasses only this config-file prompt.
- Resolve and pass the config target path without requiring the file to exist. If `internal/config` lacks a default path helper suitable for writes, add one before Step 4 uses it.
- Parser/dispatch tests should cover config path propagation, setup bypassing server startup/config load, unknown setup flags returning usage errors, setup-specific help, unsupported `icuvisor --config <path> setup`, existing-key default-no, existing-config default-no, and existing-config with `--force`.
| 2026-05-15 18:51 | Review R001 | plan Step 1: APPROVE |
| 2026-05-15 18:54 | Review R002 | code Step 1: APPROVE |
| 2026-05-15 18:56 | Review R003 | plan Step 2: UNKNOWN |
| 2026-05-15 18:58 | Review R004 | plan Step 2: REVISE |
| 2026-05-15 19:00 | Review R005 | plan Step 2: APPROVE |
| 2026-05-15 19:09 | Review R006 | code Step 2: REVISE |
| 2026-05-15 19:15 | Review R008 | plan Step 3: APPROVE |
| 2026-05-15 19:25 | Review R009 | code Step 3: REVISE |
