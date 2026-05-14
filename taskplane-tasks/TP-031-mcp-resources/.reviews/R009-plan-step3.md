# Plan Review R009 — Step 3: `icuvisor://event-categories`

**Verdict: APPROVE**

I read `PROMPT.md`, the updated `STATUS.md`, the prior R008 feedback, and the current resource/event code shape. The revised Step 3 plan now covers the missing implementation decisions and is scoped appropriately for this step.

## What looks good

- The plan defines a shared descriptor in `internal/intervals` as the source of truth instead of generating a standalone Markdown list in `internal/resources`.
- It records the upstream public OpenAPI evidence and clearly scopes this to the calendar event category enum, including the fitness-model calendar categories while excluding unrelated `category` schemas.
- It explicitly preserves current pass-through/custom category behavior and avoids turning documentation metadata into validation.
- The resource contract is pinned: URI, name/title/description, `text/markdown`, static/no-network handler, context cancellation, deterministic order, and one populated text content result.
- Registry wiring is concrete: add `EventCategoriesResource()` to the default `resources.NewRegistry()` alongside `WorkoutSyntaxResource()`.
- The planned tests cover golden stability, descriptor coverage/non-empty descriptions, registry/read behavior, and schema-description changes without crossing into Step 6’s broader trimming/README work.

## Implementation notes to keep in mind

- If event tool schema descriptions are updated in this step, keep them terse and make the custom/pass-through behavior explicit, for example by pointing to `icuvisor://event-categories` without adding a JSON Schema `enum`.
- Prefer tests that fail when a descriptor value is added but the Markdown/description rendering is not updated; this will preserve the single-source contract.
- Since the enum values were captured from public OpenAPI during planning, avoid any GPL-derived wording for the one-line descriptions.

This plan is ready to implement.
