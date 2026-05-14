# Plan Review R012 — Step 4: `icuvisor://custom-item-schemas`

**Verdict: REVISE**

I read `PROMPT.md`, the current `STATUS.md`, the completed resource registry/event/workout resource code, and the existing custom-item read/write implementation. The Step 4 section in `STATUS.md` is currently only the task checklist, not an implementation plan. Please add a concrete Step 4 plan before coding.

## Required plan additions

1. **Define the single source of truth for custom-item content schemas.**
   The current write path does not have a static schema catalog: `create_custom_item`/`update_custom_item` infer a `customItemContentSchema` from readable custom items at runtime (`customItemSchemaSamples`, `validateCustomItemContentAgainstReadSchema`). The plan needs to say how the new resource will reuse that validation/schema machinery rather than duplicating it in Markdown. A good direction is to move the schema-inference/validation primitives into a small internal domain package (or otherwise expose a shared descriptor) consumed by both the tools and the resource. A standalone `internal/resources/custom_item_schemas.go` with hand-authored prose would not satisfy the “single source of truth” requirement.

2. **Clarify static resource vs runtime athlete samples.**
   Step 1 recorded `icuvisor://custom-item-schemas` as static, and Step 4 requires golden-file locking. The plan must explain what static samples/descriptors will be used for documentation, and how that relates to the live read-side samples used by writes. It should explicitly avoid implying that the resource replaces live validation: create/update should continue validating against readable schemas for the target athlete/item, while the resource provides general long-form guidance.

3. **Pin the exact `item_type` scope and grouping.**
   The checklist says chart/field/stream/panel/zones, but the existing tools mention more concrete upstream values such as `FITNESS_CHART`, `FITNESS_TABLE`, `TRACE_CHART`, `INPUT_FIELD`, `ACTIVITY_FIELD`, `INTERVAL_FIELD`, `ACTIVITY_STREAM`, `ACTIVITY_CHART`, `ACTIVITY_HISTOGRAM`, `ACTIVITY_HEATMAP`, `ACTIVITY_MAP`, `ACTIVITY_PANEL`, and `ZONES`. The plan should list the item types/families that will be documented, identify the public/black-box source for that list, and state how unknown/custom upstream item types are handled. Do not turn documentation into a JSON Schema enum or new validation allow-list unless that behavior is explicitly intended and tested.

4. **Specify the sample/schema representation.**
   The plan should say whether the resource renders:
   - representative sample `content` JSON per family/item type,
   - inferred object paths and JSON kinds from samples,
   - one-line field notes/constraints,
   - or a combination of those.

   It should also say where those samples live and how tests will prove they are the same samples/descriptors used by custom-item validation tests. Existing fixtures only cover a small subset (`FITNESS_CHART`, `INPUT_FIELD`, `ZONES`), so the plan should identify any additional fixture/sample coverage needed for stream/panel/chart variants.

5. **Pin the resource contract.**
   Record the intended metadata and handler behavior:
   - URI: `icuvisor://custom-item-schemas`
   - stable snake_case name, human title, short description
   - MIME type, likely `text/markdown`
   - static/no-network handler that honors context cancellation
   - one text result with URI/MIME/text populated
   - wording that these are general schema samples and that live writes still validate against readable custom-item schemas.

6. **Specify registry wiring.**
   The plan should state that `CustomItemSchemasResource()` will be added to the default `resources.NewRegistry()` entries alongside `WorkoutSyntaxResource()` and `EventCategoriesResource()`, so normal server runs advertise it via `resources/list` and `resources/read`.

7. **Make tests prove stability and non-drift.**
   Add a golden-file test for the generated Markdown plus coverage/parity tests that fail when a documented item type/family or schema sample is missing from the rendered resource. Add registry/read/cancellation assertions matching the existing workout/event resource tests. If schema inference is moved out of `internal/tools`, add focused tool tests (or update existing ones) proving create/update validation still rejects the same violations and still fetches detail samples when list rows omit content.

8. **Keep Step 6 boundaries clear.**
   Broad trimming of inline custom-item descriptions belongs to Step 6. It is fine for Step 4 to add the resource and possibly introduce shared constants/metadata, but the plan should explicitly defer broad tool-description changes and README catalog updates unless they are needed for compilation/tests. If any tool schema/description text is touched now, account for the TP-015 schema-stability/confusability guards.

## Suggested Step 4 shape

A focused implementation could add a small shared custom-item schema descriptor/inference layer, generate `internal/resources/custom_item_schemas.go` Markdown from that layer, register it in `internal/resources/registry.go`, and add `internal/resources/testdata/custom_item_schemas.md` plus coverage/registry tests. The resource should document known content-shape families without changing the runtime behavior that validates writes against live readable custom items.
