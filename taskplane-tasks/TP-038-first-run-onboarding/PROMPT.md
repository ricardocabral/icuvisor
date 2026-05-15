# TP-038 — First-run onboarding flow (paste key, autodetect athlete ID + timezone, test connection)

## Mission

Give a first-run athlete a self-contained way to install their credential into the OS keychain, autodetect their athlete ID and timezone, and confirm the connection — without editing a JSON file or learning `secret-tool` / `cmdkey` / Keychain Access. v0.5 ships a deliberately *basic* version: a CLI subcommand (`icuvisor setup`) that runs in the terminal the athlete already has open after install. The polished GUI / localhost-page version is v1.0 (PRD §7.1 Flow A, "small native onboarding window (or a localhost page)") — not here.

Specifically, `icuvisor setup` does:
1. Prompts for the API key, masking input; pings `https://intervals.icu/api/v1/athlete/0/profile` with Basic Auth to confirm it works.
2. On success, autofills the athlete ID from the profile response (PRD §7.1 step 2 — accepts `i12345`/`12345`; emits canonical form) and the timezone from the OS (PRD §7.1 step 3, athlete can override).
3. Writes the API key to the OS keychain via TP-036's `internal/credstore.Store`.
4. Writes the non-secret fields (`athlete_id`, `timezone`, `api_base_url` if non-default) to `~/.config/icuvisor/config.json` (or platform equivalent) — non-credential config only.
5. Runs a final "Test connection" by calling the same code path `get_athlete_profile` uses, prints the athlete name + FTP, and tells the athlete what to wire next ("now point Claude Desktop at the config snippet in `docs/clients/claude-desktop.md`").

This task does not implement the embedded webview or the localhost HTML+HTMX onboarding page (PRD §7.3 calls that decision "deferred to design spike"). Both stay deferred to v1.0. The CLI form is the v0.5-acceptable surface area for a 5–10-athlete internal beta.

PRD anchors: §7.1 Flow A "First-time install (the golden path)"; §4 KR1 (≥95% install success in <10 minutes); §7.2.D "Athlete ID normalization (`i12345` / `12345`)"; §7.2.D "Timezone normalization to the athlete's configured TZ"; §7.4 #11 (`preferred_units` round-tripping).

ROADMAP positioning: v0.5 — Internal beta, third item. Depends on TP-036 (keychain write target) and TP-004 (`get_athlete_profile` HTTP path for the autodetect + test-connection call).

Complexity: Blast radius 2 (new subcommand + uses existing client), Pattern novelty 2 (interactive prompts + keychain writes are new together), Security 2 (handles credential at first input), Reversibility 2 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-036** — OS keychain credential storage (the keychain write target).
- **TP-004** — `get_athlete_profile` already exists as an MCP tool; this task reuses the underlying `internal/intervals` client method for autodetect + test-connection. Do NOT call the MCP tool — call the typed client directly.
- **TP-035** — CLI help / parser; `setup` is a new subcommand and must appear in `--help`.

## Context to Read First

- `CLAUDE.md` rule #6 (credential handling) and §"MCP-server conventions" — the `setup` subcommand is *not* an MCP tool, just a CLI entrypoint that uses the same HTTP client.
- `docs/prd/PRD-icuvisor.md` §7.1 Flow A in full, §7.2.D.
- `internal/intervals/` — the HTTP client; identify the profile-fetch method.
- `internal/config/config.go` — the file-write path (currently we read but do we write? add only if needed for non-secret fields).
- `internal/app/app.go` — subcommand dispatch.

## File Scope

Expected files:

- `internal/app/setup.go` (new) — `RunSetup(ctx, opts)` that orchestrates the prompts, ping, keychain write, config write, and verify.
- `internal/app/setup_test.go` — table-driven for the prompt flow with a faked stdin/stdout, faked `credstore.Store`, faked profile client.
- `internal/app/app.go` — wire the `setup` subcommand into the parser (extends the surface TP-035 documents).
- `internal/config/config.go` — add `Write(path, Config) error` for the non-secret fields if not already present; ensure it round-trips. **Must NOT write `api_key` to the file** — even if Config struct holds it in memory, the writer drops the field.
- `internal/intervals/` — surface a minimal `Profile(ctx)` method if not already exposed for non-MCP callers; thin wrapper, no new HTTP code.
- `cmd/icuvisor/main.go` — only if exit-code handling needs adjustment.
- `docs/install/macos.md` — point to `icuvisor setup` as the canonical first run.
- `README.md` — Quickstart adds a `./bin/icuvisor setup` step.
- `CHANGELOG.md`.
- `taskplane-tasks/TP-038-first-run-onboarding/STATUS.md`.

Out of scope: webview UI, localhost HTML page, "Set up automatically" buttons that *write* the AI-client JSON (PRD §7.1 step 4); the v0.5 install doc keeps the client wiring manual. Out of scope: coach-mode key onboarding (TP-039 owns the coach flow).

## Steps

### Step 1: UX script

- [ ] Write the prompt sequence down as a doc (paste into `STATUS.md`) before any code. Sample:
  ```
  Welcome to icuvisor.
  Paste your intervals.icu API key (from https://intervals.icu/settings):
  > ********
  Checking… ✓ connected as "Jane Doe" (athlete i12345)
  Detected timezone: Europe/Madrid. Use this? [Y/n]
  Saved. Your key is in the macOS Keychain; athlete id + timezone in /Users/jane/.config/icuvisor/config.json.
  Next: point Claude Desktop at icuvisor — see docs/clients/claude-desktop.md
  ```
- [ ] Confirm with `STATUS.md` notes that masking is done via the standard `golang.org/x/term` ReadPassword path (no new dep on a fancy prompt library).

### Step 2: Subcommand wiring

- [ ] `setup` registers in the parser established by TP-035; updates the golden-file `--help` fixture in lockstep.
- [ ] Re-runnable safely: if a key already exists in the keychain, prompt "An API key is already stored. Overwrite? [y/N]" — default no.
- [ ] Honour `--config` (existing flag) so the test-athlete can write to a non-default config path.

### Step 3: Autodetect + verify

- [ ] On successful profile fetch: extract athlete ID, normalize via the centralized helper in `internal/config` (`i12345` ↔ `12345`); pull display name + FTP for the success message.
- [ ] Timezone: read from `time.Local`; prompt-confirm and accept any IANA zone string the athlete pastes (validate via `time.LoadLocation`).
- [ ] If the profile fetch returns 401/403, print "API key not accepted by intervals.icu. Double-check the key on https://intervals.icu/settings." and exit non-zero **without** writing to keychain or file.
- [ ] If the fetch fails on network: do not write the key. Allow `--offline` to skip the verify and write blindly, behind a clearly-labelled flag.

### Step 4: Write

- [ ] Write API key via `credstore.Store.Set("intervals-icu-api-key", value)`. Confirm by immediate `Get` round-trip before declaring success.
- [ ] Write non-secret fields to the config file via the helper from §"File Scope". Refuse to clobber an existing config file without `--force`.

### Step 5: Tests + manual sweep

- [ ] Table-driven test with a fake stdin (canned replies), a fake credstore (in-memory), and a fake intervals client (returns a profile fixture).
- [ ] Cover: happy path; bad key (401); network error; offline override; "already exists" overwrite prompts (yes / no); config-file already exists with `--force` and without.
- [ ] Manual sweep on macOS first (TP-037's target platform); document the recipe.

## Acceptance Criteria

- `icuvisor setup` walks a first-time athlete through key → autodetect → verify → write in one terminal session.
- On success: API key in keychain (verified by `Get`), `athlete_id` + `timezone` in the config file, **no `api_key` field** in the config file.
- On 401: nothing is written; the error message names the cause and the fix URL.
- Re-running `icuvisor setup` warns before overwriting either the keychain entry or the config file.
- The `setup` subcommand appears in `icuvisor --help`; the help golden file is updated in the same PR.
- Tests run without network access.

## Do NOT

- Do not implement a webview, localhost HTML page, or GUI in this task — PRD defers that decision and v0.5 ships terminal-only.
- Do not write the API key to the config file under any circumstance, even temporarily, even with a comment. Keychain only.
- Do not accept the API key on the command line (`--api-key=...`) — it would land in shell history. Interactive prompt only; `--offline` does not change this.
- Do not write the AI-client JSON (`claude_desktop_config.json`) automatically — v0.5 keeps client wiring manual and documented; auto-write is v1.0.
- Do not call the MCP `get_athlete_profile` *tool* internally; call the typed `internal/intervals` client method directly.
- Do not log the API key, even truncated, even at DEBUG.

## Documentation

Must update:

- `STATUS.md`
- `README.md` Quickstart
- `docs/install/macos.md` (point to `icuvisor setup`)
- `CHANGELOG.md`
- `internal/app/testdata/help-fixture` (the TP-035 golden file) — `setup` must appear

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-038`, for example: `TP-038 add interactive setup subcommand`.

---

## Amendments

_Add amendments below this line only._
