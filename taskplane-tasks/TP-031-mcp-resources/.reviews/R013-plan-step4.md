# Plan Review R013 — Step 4: `icuvisor://custom-item-schemas`

**Verdict: APPROVE**

I read `PROMPT.md`, `STATUS.md`, the prior R012 review, the existing MCP resource pattern, and the current custom-item read/write validation code. The updated Step 4 plan now covers the required decisions for implementing `icuvisor://custom-item-schemas`.

## What is now satisfactory

- The plan identifies a shared single-source boundary by moving custom-item schema inference/validation primitives out of `internal/tools` into a reusable internal package consumed by both write validation and the resource generator.
- It correctly distinguishes the static/golden-locked resource from the runtime live-schema validation used by `create_custom_item` and `update_custom_item`, and explicitly says the resource must not become a validation allow-list.
- It pins the documented item-type families and concrete known upstream values, while preserving pass-through behavior for unknown/custom upstream values.
- It specifies a concrete resource representation: generated Markdown with family sections, item types, representative sample `content` JSON, and inferred path/kind output from shared inference code.
- It records the resource contract: URI, name/title/description, `text/markdown`, no-network static read handler, context cancellation, one text result, and wording about live write validation.
- It includes default registry wiring through `resources.NewRegistry()`.
- It calls out golden-file, coverage/parity, registry/read/cancellation/protocol, and write-validation regression tests.
- It keeps broad inline-description trimming and README catalog updates scoped to Step 6.

## Implementation notes to preserve the approval

1. When adding static documentation samples for item types not currently covered by fixtures (`ACTIVITY_STREAM`, `ACTIVITY_PANEL`, chart variants, etc.), record the clean-room/public or black-box source in comments/tests or `STATUS.md`. Do not let the sample catalog become unexplained hand-authored lore.
2. Make the non-drift tests prove more than Markdown string stability: they should fail if a documented family/item type lacks a sample, if a sample is not run through the shared inference/path renderer, or if the tool validation path stops using the extracted shared inference/validation machinery.
3. Keep runtime behavior unchanged: write tools should continue validating against readable custom items for the target athlete/item and should not reject unknown `item_type` values merely because they are absent from the static resource descriptor.

With those cautions, the plan is specific enough to proceed to implementation.
