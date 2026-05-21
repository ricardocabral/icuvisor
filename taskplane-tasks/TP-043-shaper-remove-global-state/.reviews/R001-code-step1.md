# Code Review — Step 1: Audit reads

## Findings

### 1. Missed a non-`Shape` reader of the response global toolset

- **File:** `taskplane-tasks/TP-043-shaper-remove-global-state/STATUS.md:52`
- **Severity:** Medium

The audit notes only globals/setters in `internal/response/shaper.go`, app startup writes, and tests relying on `SetDeleteMode` / `SetToolset`. However `internal/tools/list_advanced_capabilities.go:96` still reads the process-global state via `response.Toolset()` when constructing `_meta.toolset`.

This matters because Step 2 intends to delete `response.Toolset()`. If the audit/plan does not account for this handler, the refactor will either leave a global reader behind or change/break `icuvisor_list_advanced_capabilities` metadata. The handler already captures `activeToolset`, so the planned replacement should explicitly cover using that captured value for `_meta.toolset`.

Suggested action: update the audit notes/checklist to include `internal/tools/list_advanced_capabilities.go` as a reader and update the Step 2 plan/tests accordingly.

### 2. Options construction decision ignores the athlete-profile resource path

- **File:** `taskplane-tasks/TP-043-shaper-remove-global-state/STATUS.md:48`
- **Severity:** Medium

The decision says `response.Options` should be assembled at each `response.Shape` call from per-tool configuration threaded out of `tools.NewRegistryWithOptions`. That is incomplete for current call paths: `internal/athleteprofile/profile.go` calls `response.Shape`, and that shaping function is used by both `internal/tools/get_athlete_profile.go` and `internal/resources/athlete_profile.go`.

`defaultStartServer` currently writes delete-mode/toolset globals once, so the athlete-profile **resource** metadata reflects the configured mode/toolset today. If the refactor only threads these values through `tools.RegistryOptions`, the resource path will likely fall back to zero-value safe/core defaults and violate the acceptance criterion that `_meta` output remains bit-identical for the same logical inputs.

Suggested action: include `resources.ResourceOptions` / `resources.NewRegistryWithOptions` and `athleteprofile.Shape` in the construction-site decision, or otherwise document how the resource path will receive the same `response.Options` delete-mode/toolset values.

## Notes

- I ran the requested `git diff 41ae08b..HEAD --name-only` and full diff. The only changed file is `STATUS.md`.
- I also verified current references with grep; no build/test run was necessary for this audit-only status change.
