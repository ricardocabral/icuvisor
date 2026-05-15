# TP-043-shaper-remove-global-state — Status

**Current Step:** Step 1: Audit reads
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** S

---

### Step 1: Audit reads

**Status:** 🟨 In Progress

- [x] Grep all readers/writers
- [x] Decide `Options` construction site
- [ ] Add `list_advanced_capabilities` response.Toolset reader to audit/plan
- [ ] Add athlete-profile resource path to `Options` construction decision

### Step 2: Refactor

**Status:** ⏳ Not started

- [ ] Add fields to `Options`
- [ ] Update `addCommonMeta`
- [ ] Delete globals, `init()`, setters
- [ ] Update call sites

### Step 3: Tests

**Status:** ⏳ Not started

- [ ] Existing tests pass without `Set*`
- [ ] Add divergent-`Options` regression test

### Step 4: Verify

**Status:** ⏳ Not started

- [ ] Build / test / race / lint
- [ ] No `init()` left in `internal/response/`
- [ ] `_meta` byte-identical

---

## Decisions

- 2026-05-15: `response.Options` should be assembled at each existing `response.Shape` call from per-tool configuration threaded out of `tools.NewRegistryWithOptions`; `internal/app/app.go` already resolves `deleteMode` and `toolset` and should pass them into `tools.RegistryOptions` instead of writing response globals. Zero-value `response.Options` should preserve safe/core defaults for direct/test callers.

## Notes

- Step 1 grep found globals/setters in `internal/response/shaper.go`, app startup writes in `internal/app/app.go`, and tests in `internal/{response,tools,app}` relying on `SetDeleteMode`/`SetToolset`.
- Current `response.Shape` call sites are limited to `internal/athleteprofile/profile.go` and tools helpers in `get_activity_messages.go`, `get_activity_streams.go`, `get_fitness.go`, `update_wellness.go`, `get_activity_details.go`, and `get_activities.go`.

| 2026-05-15 14:22 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 14:22 | Step 1 started | Audit reads |
| 2026-05-15 14:26 | Review R001 | code Step 1: UNKNOWN |
