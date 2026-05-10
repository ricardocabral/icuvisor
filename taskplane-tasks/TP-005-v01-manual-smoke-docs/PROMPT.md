# TP-005 — v0.1 manual Claude Desktop smoke path and docs

## Mission

Finish the v0.1 walking skeleton by documenting and validating a manual MCP JSON configuration path for Claude Desktop on macOS. This task should produce the operator-facing instructions and smoke-test checklist that prove the binary → MCP stdio → intervals.icu → `get_athlete_profile` path works.

Roadmap items: **end-to-end via stdio to Claude Desktop on macOS** and **Manual JSON config (no installer yet)**.

Complexity: Blast radius 1, Pattern novelty 1, Security 2, Reversibility 1 = 5 → Review Level 1. Size: S/M.

## Dependencies

- **TP-004** — `get_athlete_profile` tool end-to-end wiring

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` — especially Flow A/B, §7.2.H Configuration, §7.4 #7 MCP schema caching
- `ROADMAP.md` — v0.1 only
- `SECURITY.md`
- `README.md`
- `CONTRIBUTING.md`
- Current binary CLI/config behavior from TP-001 through TP-004
- Claude Desktop MCP config documentation as needed
- Local `.env` file, if present, for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` values used during local smoke validation; never print or commit secrets

## File Scope

Expected files:

- `README.md`
- `docs/clients/claude-desktop.md` or similar v0.1 client guide
- `docs/clients/README.md` if useful
- `cmd/icuvisor/main.go` only for small UX/error-message fixes discovered by smoke testing
- `internal/config/` only for small docs-alignment fixes
- `CHANGELOG.md`
- `taskplane-tasks/TP-005-v01-manual-smoke-docs/STATUS.md`

Do not add installers, onboarding UI, keychain storage, or auto-update behavior.

## Steps

### Step 1: Plan the manual config and smoke test

- [ ] Identify the exact v0.1 config inputs and how users pass them from Claude Desktop
- [ ] Check whether local `.env` provides `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` for maintainer smoke testing; record only availability, not values
- [ ] Identify the macOS Claude Desktop config file path and JSON shape
- [ ] Decide how to show examples without committing real API keys or athlete IDs
- [ ] Write the smoke-test plan in `STATUS.md`

### Step 2: Write manual setup documentation

- [ ] Document how to build/install the local binary for v0.1 (`make build` or equivalent)
- [ ] Document how to get an intervals.icu API key from settings
- [ ] Document how to supply API key, athlete ID, timezone, and optional config path/env vars
- [ ] Document a safe local flow for loading `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` from an untracked `.env` file without committing or displaying secrets
- [ ] Provide a Claude Desktop macOS JSON config example using placeholders only
- [ ] Explain that MCP clients cache tool schemas per conversation, so users should start a new chat after binary/tool changes
- [ ] Include troubleshooting for missing binary path, invalid athlete ID, auth failure, and server startup errors

### Step 3: Add a repeatable local smoke checklist

- [ ] Add a checklist for `icuvisor version`
- [ ] Add a checklist for `make build`
- [ ] Add a checklist for starting Claude Desktop and confirming `get_athlete_profile` is listed/callable
- [ ] Add an expected sample shape for a successful `get_athlete_profile` response without real personal data
- [ ] Add a note that tests never hit the network; manual smoke requires a real intervals.icu account/API key

### Step 4: Align code UX with docs if necessary

- [ ] If docs reveal confusing errors, tighten user-facing errors without leaking secrets
- [ ] Ensure invalid config failures are short and actionable
- [ ] Ensure README quickstart points to the detailed client guide
- [ ] Update `CHANGELOG.md`

### Step 5: Verify v0.1 gate

- [ ] Run `make build`
- [ ] Run `make test`
- [ ] Run `make lint` if available
- [ ] If credentials are available locally, perform manual Claude Desktop macOS smoke test and record result in `STATUS.md`
- [ ] If credentials are not available, record exactly what remains for a human maintainer to verify
- [ ] Confirm every v0.1 roadmap checkbox is represented in completed TP-001 through TP-005 work

## Acceptance Criteria

- A non-installer manual Claude Desktop setup path is documented for macOS.
- Documentation uses placeholders and does not encourage committing API keys.
- v0.1 smoke checklist clearly proves binary → MCP stdio → intervals.icu → `get_athlete_profile`.
- README points users to the guide.
- `make build` and `make test` pass.

## Do NOT

- Do not add macOS `.dmg`, Homebrew tap, DXT bundle, auto-update, tray, onboarding UI, or keychain storage.
- Do not add support guides for every v1 target client; v0.1 only needs manual Claude Desktop macOS.
- Do not commit real credentials, real athlete IDs, or local machine-specific absolute paths except clearly marked placeholders.
- Do not print, log, commit, or copy local `.env` values into docs; `.env` is only for local smoke validation.
- Do not read, copy, paraphrase, transliterate, or port GPL implementation code, including `mvilanova/intervals-mcp-server`.

## Documentation

Must update:

- `STATUS.md`
- `README.md`
- `docs/clients/claude-desktop.md` or equivalent
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-005`, for example: `TP-005 document claude desktop v01 smoke path`.

---

## Amendments

_Add amendments below this line only._
