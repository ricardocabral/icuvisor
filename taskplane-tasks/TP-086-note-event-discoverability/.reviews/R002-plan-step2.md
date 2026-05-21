# Review R002 — Step 2 plan

Decision: **Approved with clarifications.**

The Step 1 audit captured the important generation boundary: `web/data/tools.json`/the Hugo catalog do not publish `input_examples`, while `internal/tools/add_or_update_event.go` and the schema snapshot are the assistant-facing example surface. The Step 2 checklist is pointed at the right goal, but execution should make the following details explicit so this stays focused and does not leave either user-facing or assistant-facing discoverability incomplete.

## Clarifications to carry into execution

1. **Treat `add_or_update_event` input examples as mandatory, not conditional.**
   Step 1 found that `input_examples` are the assistant-facing surface for complex write tools. Add NOTE examples in `internal/tools/add_or_update_event.go` covering the missing use cases, not only if the website reference happens to consume them. Keep examples schema-valid and synthetic:
   - `category: "NOTE"` uppercase;
   - athlete-local `date` as `YYYY-MM-DD`;
   - non-empty `name` for creates;
   - free-text `description` for the note body;
   - no `type` or `workout_doc` for NOTE examples;
   - no `athlete_id` in base examples, because the base schema is `additionalProperties: false` and coach routing is added outside this tool schema.

2. **Cover all four required use cases across the edited surfaces.**
   The existing travel-day example may satisfy travel logistics if it is made explicit enough, but the final examples/docs must clearly show nutrition plans, travel logistics, daily reminders, and coach annotations. It is fine if some are code examples and some are prose examples, but at least one public/user-facing docs page should show the pattern with `add_or_update_event` and `category: "NOTE"` so this is not only hidden in schema metadata.

3. **Use a stable public docs location, not upstream-gap docs, for discoverability.**
   `docs/upstream-gaps/event-note-payload.md` is useful technical context, but it is not the user-facing discoverability surface. Prefer a small section in `web/content/reference/tools.md` before the generated catalog, or a focused guide/tutorial page under the allowed file scope. The wording should explicitly say NOTE is an event category created with `add_or_update_event`, not a separate `add_note` tool.

4. **Do not hand-edit generated artifacts blindly.**
   Adding `input_examples` will require refreshing `internal/tools/schema_snapshot/add_or_update_event.json` in the validation step. `web/data/tools.json` only needs regeneration if Step 2 changes top-level tool descriptions/catalog metadata; changing examples alone will not affect it. Avoid editing `internal/tools/catalog.go` unless there is a concrete catalog-description change, because there is no separate catalog examples field today.

5. **Keep verification handoff clear for Step 3.**
   If Step 2 changes only input examples and prose docs, Step 3 should at minimum regenerate schema snapshots and run the targeted `internal/tools` example/schema tests plus the site/docs build. If Step 2 also changes `Tool.Description`, include the confusable-name/catalog/docs generation checks in Step 3.

With those constraints, the plan should satisfy the task completion criteria without adding a confusable `add_note` tool or broadening into unrelated event-write behavior changes.
