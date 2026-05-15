# TP-036 — OS keychain credential storage (macOS Keychain, Windows Credential Manager, libsecret)

## Mission

Move the intervals.icu API key out of plaintext config files and environment variables into the platform-native secret store on each target OS: macOS Keychain, Windows Credential Manager, and libsecret (GNOME Keyring / KWallet) on Linux. This is the v0.5 prerequisite that lets first-run onboarding (TP-038) and the future installers (TP-037) deliver on the PRD §7.2.H promise that "API key stored in OS keychain — not in plain text — fixing a recurring concern that `.env` files leak to backups/repos (forum #35 + Marc's security concern in #61)."

Today the binary reads `INTERVALS_ICU_API_KEY` from env or `api_key` from `~/.config/icuvisor/config.json`. Both stay supported (headless / CI / power users), but the **default** for installer-installed athletes becomes "key lives in the OS keychain; config holds only `athlete_id` and `timezone`."

PRD anchors: §7.2.H Configuration; §4 "Privacy: API key and data never leave the athlete's machine"; §6 Pains avoided ("`.env` files leak to backups/repos"); CLAUDE.md hard rule #6 ("API keys live in the OS keychain. Never log them. Never write them to disk in plain text").

ROADMAP positioning: first item under v0.5 — Internal beta. Blocks TP-037 (signed installer) and TP-038 (onboarding "paste API key" step) because the credential write-path must exist before the installer / onboarding can use it.

Complexity: Blast radius 2 (touches config loader + a new platform package), Pattern novelty 3 (three OS-specific backends with conditional compilation), Security 3 (handles the project's primary credential), Reversibility 2 = 10 → Review Level 3. Size: M.

## Dependencies

- **TP-007** — response-shaping primitives established the config/options shape that `internal/config` exposes. The keychain backend hooks into that loader as a higher-priority credential source than the JSON file but lower priority than explicit env (env wins for headless / CI determinism).

## Context to Read First

- `CLAUDE.md` hard rule #6 — credentials policy is non-negotiable.
- `docs/prd/PRD-icuvisor.md` §7.2.H Configuration, §4, §6.
- `internal/config/config.go` — the existing precedence chain (env, file, defaults). Note the comment markers around `EnvAPIKey` and the `<redacted>` formatting in `String()`.
- `internal/config/config_test.go` — mirror the table-driven style.
- `SECURITY.md` — disclosure and threat-model boundaries.
- ROADMAP.md v0.5.

## Library choice

Pick **one** keyring abstraction and justify in `STATUS.md`:

1. **`github.com/zalando/go-keyring` (MIT)** — small, three-backend (macOS Security framework via `/usr/bin/security`, Windows credential manager via `wincred`, libsecret via `dbus`). No CGO. Lowest-risk default.
2. **`github.com/99designs/keyring` (MIT)** — broader backend set (incl. KWallet, pass, encrypted file fallback) at the cost of more surface area and config knobs.
3. **Hand-rolled per-OS shims** — only if (1) and (2) both fail the license / dependency review. Each backend then needs its own integration test on its host OS.

Default recommendation: **`zalando/go-keyring`**. The 99designs option's encrypted-file fallback would tempt us back into "key on disk" which violates CLAUDE.md rule #6 by spirit.

## File Scope

Expected files:

- `internal/credstore/credstore.go` (new package) — interface `Store { Get/Set/Delete(account string) (string, error) }`, default `OSKeychain()` constructor, and a `NoopStore` for tests. Service name constant: `icuvisor`.
- `internal/credstore/credstore_darwin.go`, `_windows.go`, `_linux.go` — build-tagged backends if any platform-specific code is needed beyond what the chosen library hides.
- `internal/credstore/credstore_test.go` — interface contract tests using `NoopStore` + an in-memory fake. Real keychain hits are gated behind `-tags keychain_live` so CI does not need a desktop session.
- `internal/config/config.go` — add `credstore.Store` injection point; extend the precedence chain.
- `internal/config/config_test.go` — coverage of precedence with the in-memory fake.
- `cmd/icuvisor/main.go` — wire the default `OSKeychain()` store.
- `README.md` — update "Getting an API key" to describe the keychain path and the headless fallback.
- `CHANGELOG.md`.
- `taskplane-tasks/TP-036-os-keychain-credential-storage/STATUS.md`.

Out of scope: a `icuvisor login` subcommand (that's TP-038 onboarding); migration tooling that scrubs the existing plaintext key from `config.json` after a successful keychain write is a stretch goal — design for it, only ship if Step 4 lands cleanly.

## Steps

### Step 1: Backend selection and contract

- [ ] Pick the library; record license, last-release date, and rationale in `STATUS.md`.
- [ ] Define the `Store` interface and the canonical `service`/`account` pair (`service="icuvisor"`, `account="intervals-icu-api-key"`; one record per host, not per athlete — coach-mode keys are handled in TP-039).
- [ ] Decide error semantics: `ErrNotFound` sentinel (used to fall through to env / file), other errors wrapped with `%w` and surfaced.

### Step 2: Backends

- [ ] Implement the OS-keychain backend(s) behind build tags as needed. Do not require CGO.
- [ ] Linux: gracefully degrade when no D-Bus session is reachable (headless servers, SSH-only boxes). Return `ErrNotFound` so callers fall through to env/file rather than crashing.
- [ ] Never log the secret value. Log only the action verb and account name. Confirm in test by capturing the slog output.

### Step 3: Precedence chain in `internal/config`

Order, highest priority first (highest wins on collision):
1. Explicit `INTERVALS_ICU_API_KEY` env var (preserves headless / CI behaviour).
2. Keychain entry for `service=icuvisor, account=intervals-icu-api-key`.
3. `api_key` field in the JSON config file (legacy, still supported with a `slog.Warn` once at load: "api_key found in config file; consider migrating to OS keychain — see <link>").
4. Empty → existing "missing API key" error.

- [ ] Update `Config.String()` so the keychain-sourced key still prints as `<redacted>`; add a `_source` indicator (`env|keychain|file`) so diagnostics make the credential's origin obvious without exposing the value.
- [ ] Update the "missing intervals.icu API key" error message to mention all three sources.

### Step 4: Tests + manual sweep

- [ ] Table-driven precedence tests with the in-memory fake.
- [ ] Linux headless-degradation test (D-Bus unreachable → fall through, no crash).
- [ ] Manual sweep: write a key via the platform's native UI (Keychain Access on macOS, `cmdkey` on Windows, `secret-tool store` on Linux), start `icuvisor`, confirm a successful tool call. Document the recipe in the new "Getting an API key" README section.
- [ ] Confirm `go test -race ./...` is clean and the binary's startup log does not contain the key under any precedence path.

### Step 5: Documentation

- [ ] README "Getting an API key" section, with the three-line `secret-tool store` / `cmdkey /add` / "Keychain Access > +" instructions per OS.
- [ ] CHANGELOG `[Unreleased]` under "Changed" (precedence) and "Added" (keychain backend).
- [ ] SECURITY.md updated if the threat model changes (it should: "API key at rest is now protected by the OS keychain by default; the JSON-file fallback path is documented but discouraged").

## Acceptance Criteria

- A key stored in the OS keychain on macOS/Windows/Linux is read by `icuvisor` at startup without any env var or config-file `api_key` field set.
- Env var still overrides the keychain when both are present (headless CI deterministic).
- A JSON-file `api_key` still works for legacy installs but emits a one-line WARN at load.
- The secret value is never written to logs, never appears in `Config.String()`, and never shows up in `--help` examples.
- No CGO introduced. The binary still cross-compiles via GoReleaser to the existing target matrix.
- The Linux backend degrades to "not found" (not crash) when D-Bus is unreachable.
- `make test`, `make test-race`, `make lint`, `make build` clean.

## Do NOT

- Do not invent a custom encryption-at-rest fallback if the OS keychain is unavailable — CLAUDE.md rule #6 makes "encrypted JSON on disk" unacceptable as a default. Fall through to env / file (with WARN) and let the operator choose.
- Do not introduce a tool argument that lets the LLM read or write the keychain. The MCP surface never sees the key — full stop.
- Do not accept the API key as a CLI flag (TP-035 already locks this out).
- Do not add CGO. Each target OS must build with `CGO_ENABLED=0`.
- Do not delete the existing `api_key` JSON-file path in this task; deprecation is a follow-up after v0.5 dogfooding confirms zero installed-base breakage.
- Do not log secrets even at DEBUG. The redaction rule is unconditional.

## Documentation

Must update:

- `STATUS.md`
- `README.md` "Getting an API key"
- `CHANGELOG.md`
- `SECURITY.md` (threat-model note)

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-036`, for example: `TP-036 add OS keychain backend and precedence`.

---

## Amendments

_Add amendments below this line only._
