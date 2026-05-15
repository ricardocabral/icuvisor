# TP-043-shaper-remove-global-state — Status

**Current Step:** Step 2: Refactor
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 4
**Iteration:** 1
**Size:** S

---

### Step 1: Audit reads

**Status:** ✅ Complete

- [x] Grep all readers/writers
- [x] Decide `Options` construction site
- [x] Add `list_advanced_capabilities` response.Toolset reader to audit/plan
- [x] Add athlete-profile resource path to `Options` construction decision

### Step 2: Refactor

**Status:** ✅ Complete

- [x] Add fields to `Options`
- [x] Update `addCommonMeta`
- [x] Replace `list_advanced_capabilities` global toolset reader with captured toolset
- [x] Thread delete-mode/toolset through tool and athlete-profile resource shaping paths
- [x] Delete globals, `init()`, setters
- [x] Update call sites

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

- 2026-05-15: `response.Options` should be assembled at each existing `response.Shape` call from per-tool/resource configuration. `internal/app/app.go` already resolves `deleteMode` and `toolset` and should pass them into both `tools.RegistryOptions` and `resources.ResourceOptions` instead of writing response globals; `athleteprofile.Shape` needs those values because it is shared by `get_athlete_profile` and the `icuvisor://athlete-profile` resource. Zero-value `response.Options` should preserve safe/core defaults for direct/test callers.

## Notes

- Step 1 grep found globals/setters in `internal/response/shaper.go`, app startup writes in `internal/app/app.go`, tests in `internal/{response,tools,app}` relying on `SetDeleteMode`/`SetToolset`, and a non-`Shape` reader in `internal/tools/list_advanced_capabilities.go` via `response.Toolset()` that should use its captured `activeToolset` instead.
- Current `response.Shape` call sites are limited to `internal/athleteprofile/profile.go` (shared by tool and resource) and tools helpers in `get_activity_messages.go`, `get_activity_streams.go`, `get_fitness.go`, `update_wellness.go`, `get_activity_details.go`, and `get_activities.go`.

| 2026-05-15 14:22 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 14:22 | Step 1 started | Audit reads |
| 2026-05-15 14:26 | Review R001 | code Step 1: UNKNOWN |
| 2026-05-15 14:29 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-15 14:32 | Review R003 | plan Step 2: APPROVE |
| 2026-05-15 14:50 | Review R004 | code Step 2: APPROVE |
