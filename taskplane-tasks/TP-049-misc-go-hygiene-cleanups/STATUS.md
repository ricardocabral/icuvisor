# TP-049-misc-go-hygiene-cleanups — Status

**Current Step:** Step 1: Recover helper (item 1)
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 0
**Iteration:** 0
**Size:** S

---

### Step 1: Recover helper (item 1)

**Status:** ⏳ Not started

- [ ] Add `withPanicRecovery` helper with doc comment
- [ ] Collapse three `defer recover()` blocks in `internal/mcp/server.go`
- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Commit `TP-049 extract MCP panic-recovery helper`

### Step 2: Propagate `ctx` through toolchecks (item 2)

**Status:** ⏳ Not started

- [ ] Add `ctx context.Context` to `toolchecks.Register` (schema_stability + confusable_names)
- [ ] Update registry call site to pass through `ctx`
- [ ] Update tests
- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Commit `TP-049 propagate ctx through toolchecks Register`

### Step 3: Move env read into config (item 3)

**Status:** ⏳ Not started

- [ ] Resolve `DebugMetadata` in `config.Load`; store on `Config`
- [ ] `app.Run` reads from `Config`; no env access in `internal/app`
- [ ] Remove or relocate `response.DebugMetadataFromEnv`
- [ ] `grep -rn "os.Getenv\b" internal/ | grep -v "^internal/config/"` empty
- [ ] `make build` / `test` / `test-race` / `lint`
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

## Notes

_Add notes as work progresses._
