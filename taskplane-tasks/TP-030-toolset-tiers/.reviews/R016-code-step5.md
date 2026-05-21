# R016 code review — Step 5: `_meta` surfacing + docs

Verdict: **REVISE**

I reviewed the Step 5 diff against `d2632b0888a4cdb6384b162b1280a0bab2874d56`, read the changed files, and ran `go test ./...` successfully. The response-shaper changes, app propagation, README, and changelog are generally aligned with the approved plan, but `_meta.toolset` is not actually present on every tool response yet.

## Findings

### 1. Some direct-return tool responses still omit root `_meta.toolset`

`update_sport_settings` and `update_wellness` bypass the `response.Shape` chokepoint for their top-level response objects, so the new `addCommonMeta` logic never adds `_meta.toolset` to the root response returned to the MCP client.

- `internal/tools/update_sport_settings.go:80-91` defines root `_meta` with `server_version` and `delete_mode`, but no `toolset`; the handler marshals and returns this payload directly at `internal/tools/update_sport_settings.go:128-133`.
- `internal/tools/update_wellness.go:85-98` defines root `_meta` without `server_version`, `delete_mode`, or `toolset`; the handler also marshals and returns the wrapper directly at `internal/tools/update_wellness.go:125-133`. Although the nested `wellness._meta` comes from `response.Shape`, the root response `_meta` visible for the tool call lacks `toolset`.

This violates the Step 5 acceptance criterion that `_meta.toolset` appears on every response. It also leaves the README claim (“The active tier is reported in response metadata as `_meta.toolset`”) false for these tools. Please route the complete wrapper responses through the common shaping path where possible, or explicitly add response-owned `response.Toolset()` to the remaining direct-return root metadata structs and add regression assertions for these direct-return tools.

## Verification

- `go test ./...` passes.
