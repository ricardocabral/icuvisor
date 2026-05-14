# R008 plan review — Step 3: Registry filtering composition

Verdict: **APPROVE**

I reviewed `PROMPT.md`, the updated `STATUS.md`, and the existing `internal/app`, `internal/mcp`, `internal/safety`, and `internal/tools` registration plumbing. The revised Step 3 plan now addresses the issues from R007 and is specific enough to implement safely.

## What is now ready

- **Resolved toolset propagation is pinned.** The plan explicitly routes the already-loaded `ServerInfo.Toolset` into `mcp.Options.Toolset` and then into `safeRegistrar`, without re-reading `ICUVISOR_TOOLSET` and without any tool-call override.
- **Filtering remains registration-time only.** Extending `safeRegistrar` keeps hidden tools absent from the SDK catalog and therefore absent from `tools/list`, matching the acceptance criteria.
- **Validation before filtering is preserved.** This keeps Step 2's invalid in-code toolset guard meaningful: bad non-empty `Tool.Toolset` values still fail registration instead of being silently normalized or skipped.
- **Tier semantics are clear.** `core` means core-only, `full` means core plus full-only, and empty tool declarations remain effectively `full`, so unmarked future tools cannot expand the default core surface.
- **Gate composition is well-defined.** The plan evaluates toolset and capability gates independently, registers only when both allow the tool, and records count-only `registered_count`, `skipped_toolset_count`, and `skipped_capability_count` without leaking tool names.
- **The test matrix is appropriate.** Synthetic core/full read/write/delete tools crossed with delete mode `none`/`safe`/`full` should catch the important composition failures, including `core + full delete mode` and `full + safe`.
- **Protocol-level absence is covered.** The plan calls for verifying hidden full-only tools are absent from `ListTools` and behave as unknown tools when called, rather than returning a registered-tool error.
- **Existing fixture drift is accounted for.** Calling out `testEchoRegistry`, `capabilityRegistry`, and protocol helpers avoids accidentally weakening the default-core behavior just to keep old tests passing.

## Non-blocking implementation notes

- Add at least one assertion that an omitted/zero `mcp.Options.Toolset` defaults to `core` at the MCP boundary, not only in `safety.ParseToolset`. This can be folded into the protocol or composition tests.
- Because independent skip counters can both increment for the same tool, avoid tests or log wording that imply `registered_count + skipped_*_count == total tools`. If that becomes confusing, an additional `evaluated_count` field would make the log self-explanatory, but it is not required by the task.
- Keep Step 4 and Step 5 out of this implementation except for preserving placeholders: do not implement `icuvisor_list_advanced_capabilities`, `_meta.toolset`, README, or changelog changes in this step unless the task scope is intentionally changed.

Proceed with Step 3 implementation following the plan recorded in `STATUS.md`.
