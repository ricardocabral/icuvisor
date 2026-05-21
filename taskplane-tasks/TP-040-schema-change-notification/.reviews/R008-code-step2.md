# Review R008 — Code review for Step 2: `_meta` injector

Verdict: **REQUEST CHANGES**

I ran the requested diffs and `go test ./...` (passes). I found one correctness issue in the Step 2 implementation.

## Findings

### 1. `icuvisor_list_advanced_capabilities` now emits `_meta.server_version: "dev"` for released builds

`listAdvancedCapabilitiesHandler` calls the common shaper with an empty version:

- `internal/tools/list_advanced_capabilities.go:98`

```go
return encodeShaped(response, false, nil, "", false, listAdvancedCapabilitiesName, "")
```

`addCommonMeta` normalizes an empty `ServerVersion` to `"dev"`:

- `internal/response/shaper.go:347-348`

So this tool's root `_meta.server_version` will be `"dev"` even when the registry/server was created with a real version. The registry currently discards `r.version` when registering this tool:

- `internal/tools/registry.go:259`

```go
newListAdvancedCapabilitiesTool(collector.tools, r.toolset)
```

This violates the existing response contract that every response carries the actual server version, and it also makes the new schema-change metadata internally inconsistent: `_meta.catalog_hash` comes from the process-global runtime catalog set by `internal/mcp.NewServer`, while `_meta.server_version` for this response says `dev`.

Please thread the registry/server version into `newListAdvancedCapabilitiesTool` / `listAdvancedCapabilitiesHandler` and pass it to `encodeShaped`. Add an assertion on the actual response JSON/text so this does not regress.

## Notes

- The other converted bypasses (`update_sport_settings`, `update_wellness`) now route through the common shaper and preserve text/structured-content consistency.
- The process-local first-seen catalog snapshot is mutex-protected and documented as the current fallback, matching the approved Step 2 plan.
