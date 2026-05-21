# Code Review — Step 3

Verdict: **REQUEST CHANGES**

## Findings

### 1. Coach-mode advanced-capabilities metadata is now hard-coded to safe/core

**File:** `internal/tools/list_advanced_capabilities.go:95-109`, `internal/tools/list_advanced_capabilities.go:120-137`

`NewFilteredAdvancedCapabilitiesHandler` builds `shapeCfg` with `responseShapingOrDefault(nil)`, so the filtered coach-mode handler always uses the default response-shaping metadata (`delete_mode=safe`, `toolset=core`). The response struct initially sets `_meta.toolset` from the active toolset, but `encodeShaped` calls `response.Shape`, whose common-meta pass owns/overwrites `_meta.toolset` (and adds `_meta.delete_mode`) from `shapeCfg`.

That means a coach-mode server running with `ICUVISOR_TOOLSET=full` returns top-level `current_toolset: "full"` while `_meta.toolset` is rewritten to `"core"`; similarly delete mode is reported as `"safe"` even when full delete mode is active. This is a wire regression from the previous in-`mcp` coach handler, which emitted `_meta.toolset` from `r.toolset` and did not invent an incorrect delete-mode field. It also conflicts with the task's requirement not to change MCP wire behavior while lifting the handler.

The current tests miss this because the filtered-handler test checks `CurrentToolset` and count, but not `_meta.toolset`/`_meta.delete_mode`; the protocol helper only validates row names/count.

Suggested fixes:
- either keep the filtered coach handler byte-compatible with the old `tools.TextResult(...)` shape, or
- pass the actual response-shaping options into `NewFilteredAdvancedCapabilitiesHandler` (including active toolset and delete mode) and add regression assertions for `_meta.toolset` in the full/full coach case.

Also note that routing coach mode through `encodeAdvancedCapabilitiesResult` changes the coach-mode `enable_instruction` text from the previous `"enable the full icuvisor toolset"` wording to the base tool's `"enable the full tool catalog"`, which is another visible output change if byte compatibility is required.

## Tests run

```sh
go test ./internal/mcp ./internal/tools ./internal/coach ./internal/toolcatalog
```
