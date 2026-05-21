# R001 Plan Review — Step 1: Design `ToolDescriptor`

**Verdict:** Request changes before implementation.

I read `PROMPT.md`, `STATUS.md`, and the current tool registry. The step is still at the checklist level: `STATUS.md` explicitly leaves the core Step 1 decisions unchecked (`STATUS.md:17-19`). There is not yet a concrete `ToolDescriptor` design to approve.

## Blocking issues

1. **No exact Go type / JSON contract is recorded.**  
   Step 1 requires deciding the exact fields and JSON shape (`PROMPT.md:63-75`), but the status file does not record those decisions. Before coding, record whether the committed `web/data/tools.json` is a top-level array of descriptors and the precise field names/tags. Also resolve the prompt tension around `Description`: the file-scope text mentions `Description`, while the Step 1 JSON example and acceptance criteria only require `summary`. My recommendation: keep the generated JSON lean with `name`, `group`, `tier`, `safety`, `summary`, and `anchor`; if full MCP descriptions are needed internally to derive summaries, do not emit them to JSON.

2. **`group` has no source-of-truth design.**  
   Existing `tools.Tool` metadata contains name, description, requirement, and toolset only (`internal/tools/registry.go:298-306`); it does not contain `Group`. Step 1 must specify how groups are assigned and validated. Do not infer groups from name prefixes alone: examples like `link_activity_to_event`, `delete_gear`, coach-mode tools, and `icuvisor_list_advanced_capabilities` need explicit grouping. Prefer an allowlisted typed group value plus a required per-tool mapping/metadata check that fails tests when a registered tool has no group.

3. **Safety and tier derivation must be explicit.**  
   The plan should state that `tier` comes from `Tool.EffectiveToolset()` and `safety` comes from `Requirement.effective()` / `RequiresDelete()` rather than tool-name heuristics. This matters because empty toolsets default to `full` (`internal/tools/registry.go:319-327`) and write/delete gating is represented by `Requirement`, not by name (`internal/tools/registry.go:330-338`).

4. **Coach-only and meta tools need an inclusion decision.**  
   `list_athletes` and `select_athlete` are only registered when coach mode is enabled (`internal/tools/registry.go:228-234`), while `icuvisor_list_advanced_capabilities` is always built after collecting the catalog (`internal/tools/registry.go:236-242`). Step 1 should explicitly say whether `Catalog()` includes these public/meta tools. Given the task says every registration site and `toolcatalog.AllToolNames()` already includes them, the design should include them with stable group/tier/safety metadata.

5. **First-sentence extraction must not be a naive `strings.Split(description, ".")`.**  
   Several descriptions contain `intervals.icu`; a naive split would produce broken summaries. The repo already has a safer helper for this use case in `firstDescriptionSentence` (`internal/tools/list_advanced_capabilities.go:162-172`). Step 1 should specify reusing/moving that helper and adding tests for descriptions containing `intervals.icu` and multi-sentence descriptions.

## Suggested Step 1 design to record

A concrete design along these lines would satisfy the step:

```go
type ToolDescriptor struct {
    Name    string `json:"name"`
    Group   string `json:"group"`
    Tier    string `json:"tier"`   // "core" or "full"
    Safety  string `json:"safety"` // "read", "write", or "delete"
    Summary string `json:"summary"`
    Anchor  string `json:"anchor"`
}
```

- JSON output: top-level `[]ToolDescriptor`, sorted by `group` then `name`, marshaled with two-space indentation.
- `anchor`: deterministic slug equal to the snake_case tool name unless a future renderer requires otherwise.
- `summary`: whitespace-normalized first sentence from the registered MCP description, using the existing sentence-boundary logic.
- `group`: explicit typed/allowlisted metadata in code, not inferred by string prefix.
- Include all public registered/meta/coach tools and annotate their gating via `tier`/`safety`; do not omit tools because the current process mode would hide them.

Once these choices are written into `STATUS.md` (or a small design note referenced by it), Step 1 can proceed to implementation.
