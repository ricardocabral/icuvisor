# R003 Plan Review — Step 2: Wire `Catalog()` into the registry

**Verdict:** Request changes before implementation.

I read `PROMPT.md`, `STATUS.md`, the existing `internal/tools/registry.go` registration path, and the Step 1 catalog helpers. I also ran:

- `go test ./internal/tools` — passes
- `golangci-lint run ./internal/tools` — fails on the current Step 1 helpers being unused

The Step 2 plan in `STATUS.md:25-26` restates the prompt but does not yet define the implementation approach for the tricky parts of this step. Because this step is the source-of-truth refactor, those details should be settled before code is written.

## Blocking issues

1. **The plan does not say how `Catalog()` will share the registry source of truth without requiring a live client.**  
   `defaultRegistry.Register` currently refuses a nil intervals client (`internal/tools/registry.go:95-101`) and constructs the catalog through a long inline sequence of `add(new...Tool(...))` calls (`internal/tools/registry.go:117-225`). `Catalog()` must be callable with no MCP server, no intervals.icu account, and no network, but it also must not become a hand-maintained duplicate list. Record a concrete approach, e.g. extract the common registration sequence into an unexported helper that both `Register` and `Catalog()` call with different options/callbacks. That helper should let `Catalog()` collect `Tool` metadata without invoking handlers or depending on a real client.

2. **The bypass-gating behavior is underspecified for coach and meta tools.**  
   The runtime registry conditionally adds `list_athletes` and `select_athlete` only when coach mode is enabled (`internal/tools/registry.go:228-234`), and it adds `icuvisor_list_advanced_capabilities` after collecting the other tools (`internal/tools/registry.go:236-242`). The prompt requires enumerating every registration site without applying safety/toolset gates. The plan should explicitly state that `Catalog()` includes coach-mode tools and the meta tool, even when the current environment is not coach/full/delete enabled, and should state how the advanced-capabilities tool is built from the ungated catalog.

3. **Group metadata still has no source-of-truth design for Step 2.**  
   Step 1 created `ToolDescriptor.Group` (`internal/tools/catalog.go:9-15`), but the plan only says “group” and does not specify where the group comes from. Do not infer groups by name prefix; tools such as `link_activity_to_event`, `delete_gear`, coach tools, and `icuvisor_list_advanced_capabilities` need deliberate grouping. The plan should require an explicit allowlisted group mapping/metadata table and a test that fails when any registered tool has no valid group.

4. **Tier and safety derivation must be fixed to code metadata, not names.**  
   The plan should say that `tier` is derived from `Tool.EffectiveToolset()` and `safety` from the effective `Requirement` / `toolRequirement`, not from name prefixes such as `delete_*` or `update_*`. This is important because empty/invalid toolsets default to `full` (`internal/tools/registry.go:319-327`) and capability gating is represented by `Requirement`, with runtime filtering performed later by the MCP registrar.

5. **The registry-parity test needs a precise collection strategy.**  
   “Every tool registered in `internal/tools/registry.go`” is not testable if the test uses only the default registry options: coach tools would be missed, and later environment-dependent gates could hide tools. Specify that the parity test collects the full/coach/delete-unrestricted registration set, or better, compares `Catalog()` against the same extracted helper used by `Register`. The test should also assert `anchor == name` or otherwise document the anchor rule, summaries are non-empty, and returned slices are fresh/deterministically sorted.

6. **The previous review/status inconsistency is still present and lint is currently failing.**  
   `.reviews/R002-code-step1.md` says `Verdict: Request changes`, but `STATUS.md:55` records `Review R002 | code Step 1: APPROVE`. Also, `golangci-lint run ./internal/tools` still fails because `sortToolDescriptors` and `toolSummary` in `internal/tools/catalog.go:18` and `internal/tools/catalog.go:27` are unused. Step 2 can resolve the lint failure by using these helpers, but the plan should make that an explicit cleanup item and correct the audit trail before continuing.

## Suggested plan adjustments

Add enough detail to `STATUS.md` (or a linked design note) before implementation:

- Extract one shared unexported registry-enumeration helper so `Register` and `Catalog()` cannot drift.
- Have `Catalog()` collect all tools in metadata mode, including coach and meta tools, without calling MCP/server code or network clients.
- Add an explicit, allowlisted group mapping and validate every cataloged tool has one.
- Derive `tier` from `EffectiveToolset()` and `safety` from effective `Requirement`.
- Consolidate `toolSummary` with the existing `firstDescriptionSentence` logic or make one helper call the other, so `intervals.icu` sentence handling stays consistent.
- Add table-driven tests for uniqueness, snake_case, stable sort, non-empty summary/group/tier/safety/anchor, PRD analyzer exclusions, and parity with the full ungated registration set including coach/meta tools.
- Correct the `STATUS.md` review log and leave the package lint-clean after this step.
