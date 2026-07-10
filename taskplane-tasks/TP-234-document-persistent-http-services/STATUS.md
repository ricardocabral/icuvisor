# TP-234: Add persistent loopback HTTP service recipes — Status

**Current Step:** Step 1: Design secure service recipes
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 1
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

Each recipe will require interactive `icuvisor setup` under the same OS account first, use the default per-user non-secret config and credential store, and forbid API-key variables, `.env`, `--env-file`, service-manager environment directives, `api_key` config, or Task Scheduler credential arguments. These recipes neither use Linux lingering nor a service account, so the credential store remains in the logged-in user's session boundary. It will specify status/log/restart/stop/disable-or-unload/remove recovery commands and remove the wrapper/log where applicable. The HTTP transport LAN code block will be replaced with warning-only prose. A path-specific contract plus Make/CI target will validate both HTTP guides for OS lifecycle content, loopback-only executable samples, no credential material, `/mcp`, TP-232's hosted URL/no-tunnel policy, remote provider connector boundary/hosted OAuth, and the connector key `icuvisor`.

- [ ] Built-in lifecycle mechanism selected per OS
- [ ] Credential-store-only design confirmed
- [ ] Loopback endpoint fixed in examples
- [ ] Lifecycle and hosted fallback coverage planned
- [ ] R001: User-scoped service definitions, identities, logs, and lifecycle commands specified
- [ ] R001: Session-safe credential boundary and `.env`-safe working directories specified
- [ ] R001: Failure recovery and removal behavior specified for each manager
- [ ] R001: LAN executable example removal and Make/CI contract scope specified

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
