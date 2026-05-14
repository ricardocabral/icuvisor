# Plan Review R008 — Step 3: `icuvisor://event-categories`

**Verdict: REVISE**

I read `PROMPT.md`, the current `STATUS.md`, the completed resource registry/workout resource code, and the existing event tools/client. The Step 3 section in `STATUS.md` is currently only the checklist, not an implementation plan. Please add a concrete Step 3 plan before coding.

## Required plan additions

1. **Define the source of truth for the category enum.**
   The current event tools do not have a shared event-category enum; they trim and pass through arbitrary category strings (`get_events` filter, `add_or_update_event` write path) while preserving upstream values in responses. The plan needs to say what shared descriptor will be introduced, where it will live, and how event tools/resource generation will consume it. A hand-authored Markdown list only in `internal/resources` would not satisfy “sourced from the same enum the event tools use.”

2. **Name the authoritative upstream evidence and exact enum scope.**
   Step 3 requires the “full event-category enum with one-line descriptions.” The plan should record the public intervals.icu/OpenAPI source used, without consulting GPL code, and clarify whether this is the calendar event category enum only or whether fitness-model categories such as `FITNESS_DAYS`, `SET_FITNESS`, and `SET_EFTP` are out of scope. Also account for the current tool behavior that may accept athlete/account-specific category values; documenting the fixed upstream enum must not accidentally turn into validation that rejects legitimate custom values unless that is explicitly intended and tested.

3. **Pin the resource contract.**
   Record the intended metadata and handler behavior:
   - URI: `icuvisor://event-categories`
   - stable snake_case name, human title, short description
   - MIME type, likely `text/markdown`
   - static/no-network handler that honors context cancellation
   - deterministic ordering and one text result with URI/MIME/text populated

4. **Specify registry wiring.**
   The plan should state that `EventCategoriesResource()` will be added to the default `resources.NewRegistry()` entries alongside `WorkoutSyntaxResource()`, so normal server runs advertise it via `resources/list` and `resources/read`.

5. **Make the tests prove single-source and stability.**
   Add a golden-file test for the generated Markdown plus a parity/coverage test that every category in the shared descriptor appears exactly once with a non-empty one-line description. Add registry/protocol assertions that the default registry exposes and reads `icuvisor://event-categories` with the expected MIME type. If event tool schemas/descriptions are changed to reference the descriptor, include schema snapshot updates/tests as appropriate.

6. **Keep Step 6 boundaries clear.**
   It is fine to defer broader inline-description trimming and README updates to Step 6, but the Step 3 plan should explicitly avoid changing tool wording beyond what is necessary to establish the shared enum source. If descriptions are changed now, make sure the TP-015 schema-stability/confusability guards and snapshots are considered.

## Suggested Step 3 shape

A focused implementation could add an `internal/intervals` or small internal domain descriptor such as `EventCategories()` returning `{Value, Description}` entries, have `internal/resources/event_categories.go` render Markdown from that descriptor, register it in `internal/resources/registry.go`, and add golden/registry tests. The important point is that the descriptor, not the generated Markdown, is the source of truth used by both event-facing code/docs and the MCP resource.
