# TP-234: Add persistent loopback HTTP service recipes — Status

**Current Step:** Step 3: Add documentation contract coverage
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 5
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] TP-232 is complete
- [x] Platform binary/config paths confirmed

---

### Step 1: Design secure service recipes

**Status:** ✅ Complete

**Expanded plan (R001):** The guide will define three per-user, logged-in recipes: `app.icuvisor.http` at `~/Library/LaunchAgents/app.icuvisor.http.plist`, launched in `gui/$(id -u)` with `RunAtLoad`, `KeepAlive.SuccessfulExit=false`, `/` working directory, and `~/Library/Logs/icuvisor/http-service.log`; `icuvisor-http.service` at `~/.config/systemd/user/icuvisor-http.service`, with `Restart=on-failure`, `RestartSec=5`, `/` working directory, and journal inspection; and a current-user interactive-only `icuvisor-http` Task Scheduler task with an unlimited execution limit, a non-secret PowerShell wrapper/log in `%LOCALAPPDATA%\icuvisor\http-service`, and `C:\Windows\System32` as the task working directory. All service launches use an absolute icuvisor binary path with literal `--transport http --http-bind 127.0.0.1:8765`; `http://127.0.0.1:8765/mcp` is client-only.

Each recipe will require interactive `icuvisor setup` under the same OS account first, use the default per-user non-secret config and credential store, and forbid API-key variables, `.env`, `--env-file`, service-manager environment directives, `api_key` config, or Task Scheduler credential arguments. These recipes neither use Linux lingering nor a service account, so the credential store remains in the logged-in user's session boundary.

**R002 implementation detail:** The macOS creation command will use an unquoted heredoc after `mkdir -p "$HOME/Library/Logs/icuvisor"`, causing the generated plist's `StandardOutPath`/`StandardErrorPath` to contain the real absolute home path; it will use `launchctl bootstrap`, `print`, `kickstart -k`, `bootout`, `tail`, and removal of the plist/log. Linux will calculate `ICUVISOR_BINARY="$(command -v icuvisor)"`, reject a non-absolute/non-executable result, expand it into the unit's `ExecStart`, and use `systemctl --user daemon-reload`, `enable --now`, `status`, `restart`, `stop`, `disable`, `journalctl`, and removal. Windows will generate a quoted, non-secret `%LOCALAPPDATA%\icuvisor\http-service\icuvisor-http.ps1` that invokes the expanded absolute executable path with the literal HTTP arguments, redirects all streams with `*>>` to its log, then `exit $LASTEXITCODE`; the task action is the absolute `$PSHOME\powershell.exe`, working in `C:\Windows\System32`, with `-NoProfile -NonInteractive -File`, a current-user `Interactive`/limited principal, current-user logon trigger, `PT0S` equivalent (`[TimeSpan]::Zero`) execution limit, `RestartCount=999`, `RestartInterval=1 minute`, and `IgnoreNew`. Its lifecycle uses `Register-ScheduledTask`, `Start/Stop-ScheduledTask`, `Get-ScheduledTask`, `Get-ScheduledTaskInfo`, `Get-Content` plus Task Scheduler Operational events, then `Unregister-ScheduledTask` and wrapper/log removal. This failure recovery never touches config or Credential Manager.

All foreground HTTP-start commands in `http-transport.md`, including macOS and Windows, will use literal `--transport http --http-bind 127.0.0.1:8765`; its configuration sample remains loopback-only and the prior LAN command becomes warning prose only. `make docs-guidance-test` will invoke both its existing test and `python3 scripts/tests/test_http_service_docs.py`, retaining the current Ubuntu CI path. The new test will parse fenced executable snippets in the new guide and `http-transport.md`, require exact loopback for every HTTP directive and reject credential/environment sources only in executable snippets; it will also require lifecycle/log commands, `/mcp`, hosted URL/OAuth, no-tunnel policy, and the `icuvisor` connector key.

- [x] Built-in lifecycle mechanism selected per OS
- [x] Credential-store-only design confirmed
- [x] Loopback endpoint fixed in examples
- [x] Lifecycle and hosted fallback coverage planned
- [x] R001: User-scoped service definitions, identities, logs, and lifecycle commands specified
- [x] R001: Session-safe credential boundary and `.env`-safe working directories specified
- [x] R001: Failure recovery and removal behavior specified for each manager
- [x] R001: LAN executable example removal and Make/CI contract scope specified
- [x] R002: Foreground HTTP commands and config examples pinned to loopback-only
- [x] R002: Copy-pasteable absolute-path service creation and recovery mechanics specified
- [x] R002: Fenced-snippet contract and existing Make/CI invocation specified

---

### Step 2: Write and integrate the guide

**Status:** ✅ Complete

**Expanded plan (R004):** `persistent-http-service.md` will begin with the same-account `icuvisor setup` prerequisite, loopback-only client URL, credential/session boundary, and no-credentials-in-service rule. It will then provide named **macOS: LaunchAgent**, **Linux: systemd user service**, and **Windows: Task Scheduler** sections. Each section will have a fenced, copy-pasteable creation sample based on Step 1's approved definition, absolute binary-path validation or expansion, literal `--transport http --http-bind 127.0.0.1:8765`, safe working directory, non-secret log sink, start/status/log/restart/stop/disable-or-unload/removal commands, port/keychain/config recovery, and the same logged-in-user limitation. The Windows section will generate its quoted PowerShell wrapper with all-stream log redirection and child exit propagation; no section will include API-key values, credential environment directives, `.env`, `--env-file`, or plaintext `api_key` examples. A final remote-boundary section will require provider-hosted connector UIs to use hosted HTTPS/OAuth at `https://connect.icuvisor.app/mcp`, reject public tunnels, and recommend the simple connector key `icuvisor`.

`http-transport.md` will link to the guide for persistence, replace the two implicit foreground HTTP starts with literal loopback flags, and replace the LAN code block with warning-only prose while retaining the unauthenticated-LAN warning and hosted URL/no-tunnel guidance. `_index.md` will add a persistence card. `troubleshooting.md` will add a stopped/failed persistent-process row linking to the guide. `CHANGELOG.md` will receive an `[Unreleased]` Added entry in this step. `make web-build` will verify every new `relref` and rendered navigation link after the edits.

- [x] Cross-platform recipes written
- [x] Guide navigation and HTTP links added
- [x] Connector key guidance included
- [x] LAN and hosted warnings preserved
- [x] Targeted website build passes
- [x] R004: New-guide outline and fenced service samples mapped to the approved design
- [x] R004: Foreground, navigation, troubleshooting, and changelog edits mapped file by file
- [x] R004: Link rendering and website-build verification sequence specified

---

### Step 3: Add documentation contract coverage

**Status:** 🟨 In Progress

- [ ] Three-platform content contract added
- [ ] Loopback, lifecycle, and credential assertions added
- [ ] Insecure executable examples rejected
- [ ] Documentation contract wired into existing Make and CI path

---

### Step 4: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Documentation contract test passing
- [ ] Website build passes
- [ ] Lint passing
- [ ] Binary build passes

---

### Step 5: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | APPROVE | inline |
| R004 | Plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | Plan | 2 | APPROVE | inline |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 21:11 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 21:11 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

*Reserved for execution notes*
| 2026-07-10 21:16 | Review R001 | plan Step 1: REVISE |
| 2026-07-10 21:22 | Review R002 | plan Step 1: REVISE |
| 2026-07-10 21:26 | Review R003 | plan Step 1: APPROVE |
| 2026-07-10 21:27 | Review R004 | plan Step 2: REVISE |
| 2026-07-10 21:30 | Review R005 | plan Step 2: APPROVE |
