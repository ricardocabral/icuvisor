# TP-041-v05-dogfood-validation: v0.5 dogfood prep (artifacts only) — Status

**Current Step:** Step 2: Docs
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 6
**Iteration:** 1
**Size:** M

---

### Step 1: Diagnostics subcommand

**Status:** ✅ Complete

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
- [x] R003: suppress or redirect default config-loader logs during diagnostics and add real-loader leakage coverage for secret/path output

### Step 2: Docs

**Status:** 🟨 In Progress

Implementation plan (R005):
- Create `docs/internal-beta/` with all seven outputs: `README.md`, `protocol.md`, `onboarding-playbook.md`, `measurement.md`, `exit-interview.md`, `findings.md`, and `checklist.md`; keep each under ~150 lines with checklists/tables instead of long prose and no beta execution or fabricated findings.
- `README.md` will be an execution-order index linking the six runnable docs. `protocol.md` will contain recruitment copy, consent/privacy wording, eligibility filters, 5–10 cohort cap, 14-day time-box, and a clear non-execution boundary.
- `onboarding-playbook.md` will cover DMG download/install, `icuvisor setup`, first-call verification, coach-mode variant referencing TP-039 roster/ACL behavior, troubleshooting via `icuvisor diagnostics`, and release/update exercise instructions only. It will link to `docs/install/macos.md`, `docs/clients/claude-desktop.md`, `docs/clients/claude-code.md`, and `docs/coach-mode.md` rather than duplicating JSON snippets.
- `measurement.md` will define manual collection for KR1 and PRD §7.4 #6/#8/#12, including install-to-first-call timing, top tool-call names/timestamps/descriptions, mobile-need answer, willingness-to-recommend/demand signal, surprises, and blockers; it will explicitly avoid opt-in telemetry promises.
- `exit-interview.md` will ask 8–12 questions about coach-mode usability, schema-change notification clarity, willingness to recommend, daily-use blockers, mobile need, and privacy comfort without asking for roster names, athlete IDs, or raw athlete data.
- `findings.md` will contain only an empty header/table skeleton copied from `measurement.md`, and `checklist.md` will be a one-page Recruit → Onboard → Run → Synthesize flow with links and redaction reminders.
- Consent wording will say the maintainer may receive install-to-first-call timing, top tool-call names/timestamps/descriptions, mobile-need answers, qualitative notes, blockers, and voluntarily shared diagnostics output; it will exclude raw training data, API keys, athlete IDs, payloads, tool arguments, screenshots with values, and transcripts, and include revocation instructions.
- After drafting, do a local link/path sanity pass before Step 3; update STATUS immediately after each file is created.

- [x] `README.md` — execution-order index for all seven internal-beta docs
- [x] `protocol.md` — recruitment + consent + eligibility + 5–10 cap + 14-day time-box
- [x] `onboarding-playbook.md` — operator terminal recipe + coach variant + troubleshooting
- [x] `measurement.md` — KR1 / §7.4 #6/#8/#12 measurement procedure + table template
- [x] `exit-interview.md` — 8–12 question end-of-beta script
- [x] `findings.md` — empty template only
- [x] `checklist.md` — single-page operator checklist

### Step 3: Cross-check + verify

**Status:** ⏳ Not started

- [ ] `README.md` index links all seven docs in execution order
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
| 2026-05-15 23:35 | Review R003 | code Step 1: REVISE |
| 2026-05-15 23:41 | Review R004 | code Step 1: APPROVE |
| 2026-05-15 23:43 | Review R005 | plan Step 2: REVISE |
| 2026-05-15 23:45 | Review R006 | plan Step 2: APPROVE |
