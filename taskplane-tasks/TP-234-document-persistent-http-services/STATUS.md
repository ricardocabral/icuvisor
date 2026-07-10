# TP-234: Add persistent loopback HTTP service recipes — Status

**Current Step:** Step 1: Design secure service recipes
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 2
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

**Status:** 🟨 In Progress

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
- [ ] R002: Foreground HTTP commands and config examples pinned to loopback-only
- [ ] R002: Copy-pasteable absolute-path service creation and recovery mechanics specified
- [ ] R002: Fenced-snippet contract and existing Make/CI invocation specified

---

### Step 2: Write and integrate the guide

**Status:** ⬜ Not Started

- [ ] Cross-platform recipes written
- [ ] Guide navigation and HTTP links added
- [ ] Connector key guidance included
- [ ] LAN and hosted warnings preserved

---

### Step 3: Add documentation contract coverage

**Status:** ⬜ Not Started

- [ ] Three-platform content contract added
- [ ] Loopback, lifecycle, and credential assertions added
- [ ] Insecure executable examples rejected

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
