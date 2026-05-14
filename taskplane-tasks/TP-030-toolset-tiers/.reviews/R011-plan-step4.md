# R011 plan review — Step 4: `icuvisor_list_advanced_capabilities`

Verdict: **APPROVE**

I reviewed `PROMPT.md`, the revised Step 4 section in `STATUS.md`, the prior R010 feedback, and the current Step 1-3 registration/toolset plumbing. The revised plan now pins the key implementation choices needed before coding and stays within the Step 4 scope.

## What is now sufficiently specified

- **Catalog derivation is sourced from live tool metadata.** The plan uses a wrapping registrar in `internal/tools` to collect existing `Tool` values as `defaultRegistry.Register` adds them, derives rows from `EffectiveToolset()==full`, and avoids a second production name-to-tier or summary table.
- **Registration ordering is clear.** The discoverability tool is added after the existing catalog is collected, marked `core`, uses a no-argument closed schema, excludes core tools/self, and remains visible in both active `core` and `full` tiers under the Step 3 filter semantics.
- **Active toolset flow avoids runtime/env surprises.** Adding `Toolset safety.Toolset` to `tools.RegistryOptions` and passing the already-resolved startup value from `app.defaultStartServer` gives the handler the information it needs without re-reading `ICUVISOR_TOOLSET` or accepting a model-controlled override.
- **The response contract is concrete enough.** `current_toolset`, `status`, `enable_instruction`, `advanced_capabilities: [{name, summary, requirement}]`, and `_meta` with count/source/delete-mode note cover the discoverability requirements, including the exact `ICUVISOR_TOOLSET=full` instruction and the already-full status.
- **Tests cover drift and behavior.** The plan includes updating the tier matrix, handler assertions for inclusion/exclusion and one-line summaries, no-upstream-call protection, protocol visibility in default core/full modes, profile-only registry expectation updates, and schema snapshot coverage.

## Non-blocking implementation notes

- When extracting the “first sentence,” avoid a splitter that truncates on abbreviations or domains such as `intervals.icu`; the tests should catch a distinct one-line summary for at least one known full-only tool.
- Keep the collector as a narrow internal implementation detail of the registry path so Step 3 filtering remains the only registration gate. It should not become a parallel catalog authority.
- Because delete-mode filtering is orthogonal, keep both the per-row `requirement` and the short delete-mode note in the returned content so users do not infer that `ICUVISOR_TOOLSET=full` alone enables destructive tools.

This revised plan addresses the R010 blockers and is ready for Step 4 implementation.
