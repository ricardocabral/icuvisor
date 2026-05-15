# TP-049-misc-go-hygiene-cleanups — Status

**Current Step:** Step 3: Move env read into config (item 3)
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 2
**Size:** S

---

### Step 1: Recover helper (item 1)

**Status:** ✅ Complete

- [x] Add `withPanicRecovery` helper with doc comment
- [x] Collapse three `defer recover()` blocks in `internal/mcp/server.go`
- [x] `make build` / `test` / `test-race` / `lint`
- [x] Commit `TP-049 extract MCP panic-recovery helper`

### Step 2: Propagate `ctx` through toolchecks (item 2)

**Status:** ✅ Complete

- [x] Add `ctx context.Context` to `toolchecks.Register` (schema_stability + confusable_names)
- [x] Update registry call site to pass through `ctx`
- [x] Update tests
- [x] `make build` / `test` / `test-race` / `lint`
- [x] Commit `TP-049 propagate ctx through toolchecks Register`

### Step 3: Move env read into config (item 3)

**Status:** 🟨 In Progress

- [x] Resolve `DebugMetadata` in `config.Load`; store on `Config`
- [x] `app.Run` reads from `Config`; no env access in `internal/app`
- [x] Remove or relocate `response.DebugMetadataFromEnv`
- [x] `grep -rn "os.Getenv\b" internal/ | grep -v "^internal/config/"` empty
- [x] `make build` / `test` / `test-race` / `lint`
- [ ] Commit `TP-049 move DebugMetadata env read into config`

### Step 4: Reformat long constructor lines (item 4)

**Status:** ⏳ Not started

- [ ] Wrap four `newXxxTool` calls in `internal/tools/get_fitness.go:210-224`
- [ ] `gofmt`/`goimports` clean
- [ ] `make build` / `test` / `lint`
- [ ] Commit `TP-049 wrap long constructor lines in get_fitness`

### Step 5: Fix registry error message (item 5)

**Status:** ⏳ Not started

- [ ] Check if TP-042 already fixed `internal/tools/registry.go:67-70` — note overlap here
- [ ] If not fixed: replace hardcoded `getAthleteProfileName` with failing tool's name
- [ ] If fixed: mark step a no-op
- [ ] `make build` / `test` / `lint`
- [ ] Commit (if not skipped) `TP-049 fix misleading registry error message`

### Step 6: Verify

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint` all green
- [ ] No `os.Getenv` outside `internal/config/` (or justified)
- [ ] No `context.Background()` in `internal/toolchecks/`
- [ ] `git diff --stat` focused; no unrelated churn

---

## Decisions

_Record any deviations from the per-item plan here (e.g. helper name, where relocated `DebugMetadata` reader lives, whether item 5 overlapped with TP-042)._

- Step 2 checks required renaming `apiKeyLocation` local in `internal/config/config.go` to avoid a gosec G101 false positive; behavior unchanged.
- Step 3 checks required increasing `waitForServerRun` timeout from 5s to 10s after the full test suite hit the existing Streamable HTTP shutdown race timeout.

## Notes

_Add notes as work progresses._

| 2026-05-15 15:22 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 15:22 | Step 1 started | Recover helper (item 1) |
| 2026-05-15 15:24 | Review R001 | plan Step 1: APPROVE |

| 2026-05-15 16:33 | Worker iter 1 | done in 4266s, tools: 35 |
| 2026-05-15 17:07 | Exit intercept reprompt | Supervisor provided instructions (683 chars) — reprompting worker |
| 2026-05-15 17:09 | Review R002 | plan Step 2: APPROVE |
| 2026-05-15 17:15 | Review R003 | plan Step 3: APPROVE |
