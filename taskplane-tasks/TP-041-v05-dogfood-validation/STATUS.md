# TP-041-v05-dogfood-validation: v0.5 dogfood prep (artifacts only) — Status

**Current Step:** Step 1: Diagnostics subcommand
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
**Size:** M

---

### Step 1: Diagnostics subcommand

**Status:** 🟨 In Progress

Implementation plan (R001):

- Add `internal/app/diagnostics.go` with `runDiagnosticsCommand(ctx, opts, args)` and a typed output model; dispatch `diagnostics` in `Run` before default server startup, route all command/help output through `opts.Stdout`, and accept the existing `--config`, `--env-file`, `--transport`, and `--http-bind` flags without starting the MCP server.
- Load config with the normal loader/credential-store injection, but print only source labels (`env`, `keychain`, `file`, `unset`, or `error`) plus resolved transport and mode env-var values; never print config paths, athlete IDs, config dumps, API keys, or raw load errors that could contain secrets.
- Expose a small reusable MCP helper that builds the registry with the same safety/toolset/coach gates and returns the catalog hash; diagnostics calls that helper only, without creating an intervals.icu client, starting transports, or making network requests.
- Add a redacted recent-tool-call store that persists only `{timestamp_utc, name}` records under a local state path, with MCP middleware recording tool names after registration and diagnostics reading the last N; no arguments, payloads, athlete IDs, or credentials are stored.
- Test success, diagnostics help, config flag passthrough, server bypass/no-network behavior, mode-dependent catalog hash, and secret redaction. The secret matcher will explicitly allow the expected 64-character catalog hash while rejecting fixture API keys, raw and normalized athlete IDs, bearer/key token patterns, and accidental secret-shaped strings elsewhere.
- Update `internal/app/testdata/help.golden` (not the obsolete `help-fixture` path) and `CHANGELOG.md` in Step 1.

- [x] `icuvisor diagnostics` prints version, catalog hash, config source, mode env vars, OS/Go runtime, last-N tool-call names+timestamps
- [x] No-secret-leakage test (fixture-injected key, stdout grep)
- [x] TP-035 `--help` golden file updated
- [x] CHANGELOG.md records `icuvisor diagnostics` under Added

### Step 2: Docs

**Status:** ⏳ Not started

- [ ] `protocol.md` — recruitment + consent + eligibility + 5–10 cap + 14-day time-box
- [ ] `onboarding-playbook.md` — operator terminal recipe + coach variant + troubleshooting
- [ ] `measurement.md` — KR1 / §7.4 #6/#8/#12 measurement procedure + table template
- [ ] `exit-interview.md` — 8–12 question end-of-beta script
- [ ] `findings.md` — empty template only
- [ ] `checklist.md` — single-page operator checklist

### Step 3: Cross-check + verify

**Status:** ⏳ Not started

- [ ] `README.md` index links all five docs in execution order
- [ ] Cold-read of `checklist.md` is self-sufficient
- [ ] `make test` / `make lint` / `make build` clean

---

## Decisions

- **Scope:** artifacts only. Recruitment, cohort execution, synthesis are maintainer-run after this task closes.
- **`findings.md`:** empty template — no fabricated example data.

## Notes

_Add notes as work progresses._

| 2026-05-15 22:40 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 22:40 | Step 1 started | Diagnostics subcommand |
| 2026-05-15 22:43 | Review R001 | plan Step 1: REVISE |

| 2026-05-15 23:29 | Exit intercept reprompt | Supervisor provided instructions (681 chars) — reprompting worker |